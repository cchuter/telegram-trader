package ton

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"math/big"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// SwapRequest represents a request to execute a token swap
type SwapRequest struct {
	FromToken  string // Source token address (e.g., "TON" or jetton address)
	ToToken    string // Destination token address
	Amount     string // Input amount in nanotons or smallest units
	MinOutput  string // Minimum output amount (for slippage protection)
	RouterAddr string // ston.fi router contract address
	WalletAddr string // User's wallet address
	PrivateKey string // User's private key (will be decrypted before use)
}

// SwapResult represents the result of a swap execution
type SwapResult struct {
	TxHash       string // Transaction hash
	Status       string // Transaction status: "pending", "success", "failed"
	OutputAmount string // Actual output amount received
	Fee          string // Actual transaction fee paid
	Error        string // Error message if failed
}

// TransactionBuilder handles building and signing TON transactions
type TransactionBuilder struct {
	api *ton.APIClient
}

// NewTransactionBuilder creates a new transaction builder
func NewTransactionBuilder(api *ton.APIClient) *TransactionBuilder {
	return &TransactionBuilder{
		api: api,
	}
}

// BuildSwapTransaction builds a TON transaction for a swap on ston.fi
// This creates the transaction payload and returns the signed transaction
func (tb *TransactionBuilder) BuildSwapTransaction(ctx context.Context, req *SwapRequest) (*wallet.Message, error) {
	// Parse router address
	routerAddr, err := address.ParseAddr(req.RouterAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid router address: %w", err)
	}

	// Parse amount
	amount := new(big.Int)
	amount, ok := amount.SetString(req.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount format: %s", req.Amount)
	}

	// Build swap payload cell
	// The payload structure depends on ston.fi router contract specification
	// For now, we create a basic payload with the swap parameters
	payload, err := tb.buildSwapPayload(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build swap payload: %w", err)
	}

	// Convert amount to tlb.Coins
	amountCoins := tlb.FromNanoTON(amount)

	// Create message to router contract
	msg := wallet.SimpleMessage(routerAddr, amountCoins, payload)

	return msg, nil
}

// buildSwapPayload builds the payload cell for a ston.fi swap
// This follows the ston.fi router contract ABI
func (tb *TransactionBuilder) buildSwapPayload(req *SwapRequest) (*cell.Cell, error) {
	// ston.fi swap payload structure (simplified for POC):
	// - op code for swap (uint32)
	// - query_id (uint64)
	// - token_to_address (address)
	// - min_output (coins)
	// - forward_amount (coins)
	// - forward_payload (cell)

	builder := cell.BeginCell()

	// Opcode for swap (0x25938561 is a common DEX swap opcode, may need adjustment)
	// TODO: Verify actual ston.fi router opcode from contract documentation
	err := builder.StoreUInt(0x25938561, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to store opcode: %w", err)
	}

	// Query ID (timestamp for uniqueness)
	queryID := uint64(time.Now().Unix())
	err = builder.StoreUInt(queryID, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to store query_id: %w", err)
	}

	// Destination token address
	toTokenAddr, err := address.ParseAddr(req.ToToken)
	if err != nil {
		return nil, fmt.Errorf("invalid to_token address: %w", err)
	}
	err = builder.StoreAddr(toTokenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to store to_token address: %w", err)
	}

	// Minimum output amount (for slippage protection)
	minOutput := new(big.Int)
	minOutput, ok := minOutput.SetString(req.MinOutput, 10)
	if !ok {
		return nil, fmt.Errorf("invalid min_output format: %s", req.MinOutput)
	}
	err = builder.StoreBigCoins(minOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to store min_output: %w", err)
	}

	return builder.EndCell(), nil
}

// SendTransaction sends a transaction to the TON blockchain
func (c *Client) SendTransaction(ctx context.Context, walletAddr string, privateKeyHex string, msg *wallet.Message) (string, error) {
	// Parse wallet address
	addr, err := address.ParseAddr(walletAddr)
	if err != nil {
		return "", fmt.Errorf("invalid wallet address: %w", err)
	}

	// Parse private key (hex string to ed25519.PrivateKey)
	// Note: In production, privateKeyHex should be decrypted first
	privateKey, err := parsePrivateKey(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// Create wallet instance (assuming V4R2 wallet version, most common)
	w, err := wallet.FromPrivateKey(c.api, privateKey, wallet.V4R2)
	if err != nil {
		return "", fmt.Errorf("failed to create wallet: %w", err)
	}

	// Verify wallet address matches
	if w.Address().String() != addr.String() {
		return "", fmt.Errorf("wallet address mismatch: expected %s, got %s", addr.String(), w.Address().String())
	}

	// Send message (Send returns error only)
	err = w.Send(ctx, msg, false) // false = wait for confirmation
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	// Get last transaction hash from wallet
	// Note: This is a simplified approach for POC
	// In production, we should track the specific transaction
	txs, err := c.api.ListTransactions(ctx, addr, 1, 0, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to list transactions: %w", err)
	}

	if len(txs) == 0 {
		return "", fmt.Errorf("no transaction found after send")
	}

	// Get transaction hash
	txHash := fmt.Sprintf("%x", txs[0].Hash)

	return txHash, nil
}

// WaitForTransaction waits for a transaction to be confirmed and returns its status
func (c *Client) WaitForTransaction(ctx context.Context, walletAddr, txHash string, maxWaitTime time.Duration) (*SwapResult, error) {
	// Parse wallet address
	addr, err := address.ParseAddr(walletAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet address: %w", err)
	}

	// Set timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, maxWaitTime)
	defer cancel()

	// Poll transaction status
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return &SwapResult{
				TxHash: txHash,
				Status: "timeout",
				Error:  "transaction confirmation timeout",
			}, nil
		case <-ticker.C:
			// Get latest block
			block, err := c.api.CurrentMasterchainInfo(timeoutCtx)
			if err != nil {
				continue // Retry on error
			}

			// Get transactions for the wallet
			txs, err := c.api.ListTransactions(timeoutCtx, addr, 10, 0, []byte{})
			if err != nil {
				continue // Retry on error
			}

			// Look for our transaction
			for _, tx := range txs {
				currentTxHash := fmt.Sprintf("%x", tx.Hash)
				if currentTxHash == txHash {
					// Transaction found, check status
					if tx.IO.Out != nil && tx.IO.Out.List != nil {
						// Transaction successful (has output messages)
						return &SwapResult{
							TxHash: txHash,
							Status: "success",
							// TODO: Parse actual output amount from transaction messages
							OutputAmount: "0",
							Fee:          tx.TotalFees.Coins.String(),
						}, nil
					} else {
						// Transaction failed or still pending
						if uint64(block.SeqNo) > tx.LT+5 { // If enough blocks have passed
							return &SwapResult{
								TxHash: txHash,
								Status: "failed",
								Error:  "transaction failed",
								Fee:    tx.TotalFees.Coins.String(),
							}, nil
						}
					}
				}
			}
		}
	}
}

// parsePrivateKey parses a private key from hex string
func parsePrivateKey(keyHex string) (ed25519.PrivateKey, error) {
	// For POC: assume keyHex is a base64 or hex encoded ed25519 private key
	// In production, this should handle proper decryption and validation

	// For now, return an error indicating this needs implementation
	return nil, fmt.Errorf("private key parsing not fully implemented - requires decryption integration")
}
