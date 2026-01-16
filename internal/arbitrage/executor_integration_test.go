// +build integration

package arbitrage

import (
	"context"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/blockchain/ton"
	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/galachain/pb"
)

// TestExecutorIntegration tests the executor with real-like scenarios
func TestExecutorIntegration(t *testing.T) {
	t.Skip("Integration test - requires real clients")

	// This test would use real DEX and blockchain clients
	// For now, we skip it and use mocks in unit tests

	// Example structure:
	// 1. Create real TON client
	// 2. Create real GalaChain client
	// 3. Create real DEX client
	// 4. Detect opportunity
	// 5. Execute arbitrage
	// 6. Verify transaction on blockchain
}

// TestConcurrentExecution_Integration verifies that both legs execute in parallel
func TestConcurrentExecution_Integration(t *testing.T) {
	// Create mock clients with artificial delays
	tonClient := &delayedTonClient{delay: 500 * time.Millisecond}
	galaSwapClient := &delayedGalaSwapClient{delay: 500 * time.Millisecond}

	dexClient := &mockDexClient{
		simulateSwapFunc: func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
			return &dex.SwapSimulation{
				OutputAmount: "850000000000",
				Fee:          "50000000",
				Slippage:     1.0,
				PriceImpact:  0.1,
			}, nil
		},
	}

	galaClient := &mockGalaClient{
		getPriceFunc: func(ctx context.Context, pair string) (*pb.PriceResponse, error) {
			return &pb.PriceResponse{
				Price:     "855",
				Timestamp: time.Now().Unix(),
			}, nil
		},
	}

	// Create engine and executor
	engine := NewEngine(dexClient, galaClient)
	executor := NewExecutor(tonClient, galaSwapClient, engine)

	// Create opportunity
	opportunity := &Opportunity{
		Direction:   BuyTonSellGala,
		StonfiPrice: 850.0,
		GswapPrice:  855.0,
		Spread:      0.59,
	}

	// Create position
	position := &PositionSize{
		TONAmount:  5.0,
		GALAAmount: 2500.0,
		Valid:      true,
		Reason:     "",
	}

	// Execute arbitrage
	ctx := context.Background()
	startTime := time.Now()
	result := executor.ExecuteArbitrage(ctx, opportunity, position, 123456, "mock_wallet", "mock_key")
	executionTime := time.Since(startTime)

	// Verify result
	if !result.Success {
		t.Errorf("Expected success, got failure: %v", result.Error)
	}

	// If both legs ran sequentially, it would take ~1000ms (500ms + 500ms)
	// If running concurrently, it should take ~500ms
	if executionTime > 700*time.Millisecond {
		t.Errorf("Execution took %s, likely not running concurrently (expected ~500ms)", executionTime)
	}

	t.Logf("Concurrent execution completed in %s (both legs 500ms each)", executionTime)
}

// delayedTonClient simulates network latency
type delayedTonClient struct {
	delay time.Duration
}

func (d *delayedTonClient) SendTransaction(ctx context.Context, walletAddr string, privateKeyHex string, msg *ton.TransactionBuilder) (string, error) {
	time.Sleep(d.delay)
	return "delayed_tx_hash", nil
}

func (d *delayedTonClient) WaitForTransaction(ctx context.Context, walletAddr, txHash string, maxWaitTime time.Duration) (*ton.SwapResult, error) {
	return &ton.SwapResult{
		TxHash:       txHash,
		Status:       "success",
		OutputAmount: "850000000000",
		Fee:          "50000000",
	}, nil
}

// delayedGalaSwapClient simulates network latency
type delayedGalaSwapClient struct {
	delay time.Duration
}

func (d *delayedGalaSwapClient) ExecuteSwap(ctx context.Context, req *pb.SwapRequest) (*pb.SwapResponse, error) {
	time.Sleep(d.delay)
	return &pb.SwapResponse{
		TxHash:    "delayed_gswap_tx",
		AmountIn:  req.Amount,
		AmountOut: "855000000000",
		Fee:       "30000000",
		Status:    "success",
	}, nil
}
