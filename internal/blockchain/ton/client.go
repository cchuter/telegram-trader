package ton

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/cchuter/telegram-trader/internal/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// Client implements blockchain.Client for TON blockchain
type Client struct {
	connection *liteclient.ConnectionPool
	api        *ton.APIClient
}

// NewClient creates a new TON blockchain client
func NewClient() *Client {
	return &Client{}
}

// Connect initializes the connection to TON blockchain
// Connects to TON mainnet by default
func (c *Client) Connect(ctx context.Context) error {
	// Create connection pool for TON blockchain
	client := liteclient.NewConnectionPool()

	// Connect to mainnet (using global config)
	configUrl := "https://ton.org/global-config.json"
	err := client.AddConnectionsFromConfigUrl(ctx, configUrl)
	if err != nil {
		return fmt.Errorf("failed to add connections from config: %w", err)
	}

	// Initialize API client
	c.connection = client
	c.api = ton.NewAPIClient(client)

	return nil
}

// GetBalance returns the balance for the given address
// Fetches real balance from TON blockchain via tonutils-go
// Returns balance in decimal format (e.g., "10.5" for 10.5 TON)
// Uses retry logic with exponential backoff for network errors
func (c *Client) GetBalance(ctx context.Context, walletAddr string) (string, error) {
	// Check if client is connected
	if c.api == nil {
		return "", fmt.Errorf("client not connected: call Connect() first")
	}

	// Parse the TON address (don't retry on validation errors)
	addr, err := address.ParseAddr(walletAddr)
	if err != nil {
		return "", fmt.Errorf("invalid address format: %w", err)
	}

	var result string
	err = utils.RetryWithBackoff(ctx, func(ctx context.Context) error {
		// Get the current blockchain block
		block, errBlock := c.api.CurrentMasterchainInfo(ctx)
		if errBlock != nil {
			return fmt.Errorf("failed to get masterchain info: %w", errBlock)
		}

		// Get account state from blockchain
		account, errAcct := c.api.GetAccount(ctx, block, addr)
		if errAcct != nil {
			return fmt.Errorf("failed to get account: %w", errAcct)
		}

		// Check if account exists
		if !account.IsActive {
			// Account doesn't exist yet or has zero balance
			result = "0.0"
			return nil
		}

		// Convert nanotons to TON (1 TON = 1e9 nanotons)
		balance := account.State.Balance
		balanceFloat := new(big.Float).SetInt(balance.Nano())
		balanceFloat.Quo(balanceFloat, big.NewFloat(1e9))

		// Format to string with proper precision
		result = formatBalance(balanceFloat)
		return nil
	})

	return result, err
}

// GetJettonBalance returns the jetton balance for the given wallet address and jetton master address
// GALA token address on TON: EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV
// Uses retry logic with exponential backoff for network errors
func (c *Client) GetJettonBalance(ctx context.Context, walletAddr string, jettonMasterAddr string) (string, error) {
	// Check if client is connected
	if c.api == nil {
		return "", fmt.Errorf("client not connected: call Connect() first")
	}

	// Parse addresses (don't retry on validation errors)
	ownerAddr, err := address.ParseAddr(walletAddr)
	if err != nil {
		return "", fmt.Errorf("invalid wallet address format: %w", err)
	}

	masterAddr, err := address.ParseAddr(jettonMasterAddr)
	if err != nil {
		return "", fmt.Errorf("invalid jetton master address format: %w", err)
	}

	var result string
	err = utils.RetryWithBackoff(ctx, func(ctx context.Context) error {
		// Get the current blockchain block
		block, errBlock := c.api.CurrentMasterchainInfo(ctx)
		if errBlock != nil {
			return fmt.Errorf("failed to get masterchain info: %w", errBlock)
		}

		// Get jetton wallet address for the owner
		// This calls the get_wallet_address method on the jetton master contract
		jettonWallet, errWallet := c.getJettonWalletAddress(ctx, block, masterAddr, ownerAddr)
		if errWallet != nil {
			return fmt.Errorf("failed to get jetton wallet address: %w", errWallet)
		}

		// Get the jetton wallet account
		account, errAcct := c.api.GetAccount(ctx, block, jettonWallet)
		if errAcct != nil {
			return fmt.Errorf("failed to get jetton wallet account: %w", errAcct)
		}

		// Check if jetton wallet exists
		if !account.IsActive {
			// No jetton balance for this wallet
			result = "0.0"
			return nil
		}

		// Call get_wallet_data method to get the balance
		balance, errBal := c.getJettonWalletBalance(ctx, block, jettonWallet)
		if errBal != nil {
			return fmt.Errorf("failed to get jetton balance: %w", errBal)
		}

		// Convert balance to float (assuming 9 decimals for GALA, same as TON)
		balanceFloat := new(big.Float).SetInt(balance)
		balanceFloat.Quo(balanceFloat, big.NewFloat(1e9))

		result = formatBalance(balanceFloat)
		return nil
	})

	return result, err
}

// getJettonWalletAddress gets the jetton wallet address for an owner from the master contract
func (c *Client) getJettonWalletAddress(ctx context.Context, block *ton.BlockIDExt, masterAddr, ownerAddr *address.Address) (*address.Address, error) {
	// Build the get_wallet_address call parameter
	b := cell.BeginCell()
	err := b.StoreAddr(ownerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to store owner address: %w", err)
	}

	// Execute get_wallet_address method
	result, err := c.api.RunGetMethod(ctx, block, masterAddr, "get_wallet_address", b.EndCell())
	if err != nil {
		return nil, fmt.Errorf("failed to run get_wallet_address: %w", err)
	}

	// Parse the result as a slice
	slice := result.MustSlice(0)
	jettonWalletAddr, err := slice.LoadAddr()
	if err != nil {
		return nil, fmt.Errorf("failed to parse jetton wallet address: %w", err)
	}

	return jettonWalletAddr, nil
}

// getJettonWalletBalance gets the balance from a jetton wallet contract
func (c *Client) getJettonWalletBalance(ctx context.Context, block *ton.BlockIDExt, jettonWalletAddr *address.Address) (*big.Int, error) {
	// Execute get_wallet_data method on jetton wallet
	result, err := c.api.RunGetMethod(ctx, block, jettonWalletAddr, "get_wallet_data")
	if err != nil {
		return nil, fmt.Errorf("failed to run get_wallet_data: %w", err)
	}

	// First element is the balance (as int)
	balance := result.MustInt(0)
	return balance, nil
}

// formatBalance formats a balance with trailing zeros removed
func formatBalance(balance *big.Float) string {
	// Format with 9 decimal places
	str := balance.Text('f', 9)

	// Remove trailing zeros
	for len(str) > 0 && str[len(str)-1] == '0' && str[len(str)-2] != '.' {
		str = str[:len(str)-1]
	}

	// Ensure at least one decimal place
	if str[len(str)-1] == '.' {
		str += "0"
	}

	return str
}

// ExecuteSwap executes a token swap on ston.fi
// This builds, signs, and submits a transaction to the ston.fi router contract
// Returns transaction hash and waits for confirmation
func (c *Client) ExecuteSwap(ctx context.Context, req *SwapRequest) (*SwapResult, error) {
	// Check if client is connected
	if c.api == nil {
		return nil, fmt.Errorf("client not connected: call Connect() first")
	}

	// Create transaction builder
	builder := NewTransactionBuilder(c.api)

	// Build swap transaction
	msg, err := builder.BuildSwapTransaction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build swap transaction: %w", err)
	}

	// Send transaction
	txHash, err := c.SendTransaction(ctx, req.WalletAddr, req.PrivateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Wait for transaction confirmation (with 60 second timeout)
	result, err := c.WaitForTransaction(ctx, req.WalletAddr, txHash, 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	return result, nil
}

// Close closes the connection to TON blockchain
func (c *Client) Close() error {
	if c.connection != nil {
		c.connection.Stop()
		c.connection = nil
		c.api = nil
	}
	return nil
}
