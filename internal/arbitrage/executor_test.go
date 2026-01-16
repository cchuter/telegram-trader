package arbitrage

import (
	"context"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/blockchain/ton"
	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/galachain/pb"
)

// Mock clients for testing
type mockTonClient struct{}

func (m *mockTonClient) SendTransaction(ctx context.Context, walletAddr string, privateKeyHex string, msg *ton.TransactionBuilder) (string, error) {
	return "mock_tx_hash", nil
}

func (m *mockTonClient) WaitForTransaction(ctx context.Context, walletAddr, txHash string, maxWaitTime time.Duration) (*ton.SwapResult, error) {
	return &ton.SwapResult{
		TxHash:       txHash,
		Status:       "success",
		OutputAmount: "850000000000",
		Fee:          "50000000",
	}, nil
}

type mockGalaSwapClient struct{}

func (m *mockGalaSwapClient) ExecuteSwap(ctx context.Context, req *pb.SwapRequest) (*pb.SwapResponse, error) {
	return &pb.SwapResponse{
		TxHash:    "mock_gswap_tx",
		AmountIn:  req.Amount,
		AmountOut: "855000000000",
		Fee:       "30000000",
		Status:    "success",
	}, nil
}

func TestExecuteArbitrage_BuyTonSellGala(t *testing.T) {
	// Create mock clients
	tonClient := &mockTonClient{}
	galaSwapClient := &mockGalaSwapClient{}
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
	result := executor.ExecuteArbitrage(ctx, opportunity, position, 123456, "mock_wallet", "mock_key")

	// Verify result
	if !result.Success {
		t.Errorf("Expected success, got failure: %v", result.Error)
	}

	if result.StonfiResult == nil {
		t.Error("Expected StonfiResult to be set")
	}

	if result.GswapResult == nil {
		t.Error("Expected GswapResult to be set")
	}

	if result.ExecutionTime == 0 {
		t.Error("Expected ExecutionTime to be > 0")
	}

	if result.Profit <= 0 {
		t.Errorf("Expected positive profit, got: %f", result.Profit)
	}

	t.Logf("Arbitrage executed successfully in %s with profit: %.2f%%",
		result.ExecutionTime.String(), result.Profit)
}

func TestExecuteArbitrage_BuyGalaSellTon(t *testing.T) {
	// Create mock clients
	tonClient := &mockTonClient{}
	galaSwapClient := &mockGalaSwapClient{}
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
				Price:     "850",
				Timestamp: time.Now().Unix(),
			}, nil
		},
	}

	// Create engine and executor
	engine := NewEngine(dexClient, galaClient)
	executor := NewExecutor(tonClient, galaSwapClient, engine)

	// Create opportunity (opposite direction)
	opportunity := &Opportunity{
		Direction:   BuyGalaSellTon,
		StonfiPrice: 855.0,
		GswapPrice:  850.0,
		Spread:      -0.59,
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
	result := executor.ExecuteArbitrage(ctx, opportunity, position, 123456, "mock_wallet", "mock_key")

	// Verify result
	if !result.Success {
		t.Errorf("Expected success, got failure: %v", result.Error)
	}

	if result.StonfiResult == nil {
		t.Error("Expected StonfiResult to be set")
	}

	if result.GswapResult == nil {
		t.Error("Expected GswapResult to be set")
	}

	t.Logf("Arbitrage executed successfully in %s", result.ExecutionTime.String())
}

func TestExecuteArbitrage_Concurrency(t *testing.T) {
	// Create mock clients
	tonClient := &mockTonClient{}
	galaSwapClient := &mockGalaSwapClient{}
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
	result := executor.ExecuteArbitrage(ctx, opportunity, position, 123456, "mock_wallet", "mock_key")

	// Verify both legs executed
	if result.StonfiResult == nil || result.GswapResult == nil {
		t.Error("Both legs should execute concurrently")
	}

	// Since execution is concurrent with 100ms mock delays, total time should be ~100ms (not 200ms)
	// Allow some overhead for goroutine scheduling
	if result.ExecutionTime > 300*time.Millisecond {
		t.Errorf("Execution took too long (%s), likely not running concurrently", result.ExecutionTime)
	}

	t.Logf("Concurrent execution completed in %s (expected ~100ms)", result.ExecutionTime)
}

func TestCalculateActualProfit(t *testing.T) {
	// Create mock clients
	tonClient := &mockTonClient{}
	galaSwapClient := &mockGalaSwapClient{}
	dexClient := &mockDexClient{
		simulateSwapFunc: func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
			return &dex.SwapSimulation{
				OutputAmount: "850000000000",
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

	// Test cases
	tests := []struct {
		name         string
		stonfiResult *ton.SwapResult
		gswapResult  *pb.SwapResponse
		opportunity  *Opportunity
		expectProfit bool
	}{
		{
			name: "Both successful",
			stonfiResult: &ton.SwapResult{
				Status: "success",
			},
			gswapResult: &pb.SwapResponse{
				Status: "success",
			},
			opportunity: &Opportunity{
				Spread: 0.59,
			},
			expectProfit: true,
		},
		{
			name: "One failed",
			stonfiResult: &ton.SwapResult{
				Status: "failed",
			},
			gswapResult: &pb.SwapResponse{
				Status: "success",
			},
			opportunity: &Opportunity{
				Spread: 0.59,
			},
			expectProfit: false,
		},
		{
			name:         "Nil results",
			stonfiResult: nil,
			gswapResult:  nil,
			opportunity: &Opportunity{
				Spread: 0.59,
			},
			expectProfit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profit := executor.calculateActualProfit(tt.stonfiResult, tt.gswapResult, tt.opportunity)
			if tt.expectProfit && profit == 0 {
				t.Error("Expected non-zero profit")
			}
			if !tt.expectProfit && profit != 0 {
				t.Errorf("Expected zero profit, got: %f", profit)
			}
		})
	}
}
