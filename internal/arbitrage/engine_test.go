package arbitrage

import (
	"context"
	"errors"
	"testing"

	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/galachain/pb"
)

// mockDexClient is a mock implementation of dex.Client for testing
type mockDexClient struct {
	simulateSwapFunc func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error)
}

func (m *mockDexClient) SimulateSwap(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
	if m.simulateSwapFunc != nil {
		return m.simulateSwapFunc(ctx, fromToken, toToken, amount)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDexClient) Close() error {
	return nil
}

// mockGalaClient is a mock implementation of galachain.Client for testing
type mockGalaClient struct {
	getPriceFunc func(ctx context.Context, pair string) (*pb.PriceResponse, error)
}

func (m *mockGalaClient) GetPrice(ctx context.Context, pair string) (*pb.PriceResponse, error) {
	if m.getPriceFunc != nil {
		return m.getPriceFunc(ctx, pair)
	}
	return nil, errors.New("not implemented")
}

func TestDetectOpportunity_ProfitableBuyTonSellGala(t *testing.T) {
	// Setup: gswap price is higher than ston.fi (spread > 0.3%)
	// ston.fi: 100 GALA per TON
	// gswap: 101 GALA per GTON
	// spread = (101 - 100) / 100 * 100 = 1%
	dexClient := &mockDexClient{
		simulateSwapFunc: func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
			return &dex.SwapSimulation{
				OutputAmount: "100000000000", // 100 GALA (with 9 decimals)
			}, nil
		},
	}

	galaClient := &mockGalaClient{
		getPriceFunc: func(ctx context.Context, pair string) (*pb.PriceResponse, error) {
			return &pb.PriceResponse{
				Price: "101",
			}, nil
		},
	}

	engine := NewEngine(dexClient, galaClient)
	opp, err := engine.DetectOpportunity(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if opp == nil {
		t.Fatal("Expected opportunity, got nil")
	}
	if opp.Direction != BuyTonSellGala {
		t.Errorf("Expected direction %s, got %s", BuyTonSellGala, opp.Direction)
	}
	if opp.StonfiPrice != 100 {
		t.Errorf("Expected ston.fi price 100, got %f", opp.StonfiPrice)
	}
	if opp.GswapPrice != 101 {
		t.Errorf("Expected gswap price 101, got %f", opp.GswapPrice)
	}
	if opp.Spread != 1.0 {
		t.Errorf("Expected spread 1.0%%, got %f%%", opp.Spread)
	}
}

func TestDetectOpportunity_ProfitableBuyGalaSellTon(t *testing.T) {
	// Setup: ston.fi price is higher than gswap (spread < -0.3%)
	// ston.fi: 100 GALA per TON
	// gswap: 99 GALA per GTON
	// spread = (99 - 100) / 100 * 100 = -1%
	dexClient := &mockDexClient{
		simulateSwapFunc: func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
			return &dex.SwapSimulation{
				OutputAmount: "100000000000", // 100 GALA (with 9 decimals)
			}, nil
		},
	}

	galaClient := &mockGalaClient{
		getPriceFunc: func(ctx context.Context, pair string) (*pb.PriceResponse, error) {
			return &pb.PriceResponse{
				Price: "99",
			}, nil
		},
	}

	engine := NewEngine(dexClient, galaClient)
	opp, err := engine.DetectOpportunity(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if opp == nil {
		t.Fatal("Expected opportunity, got nil")
	}
	if opp.Direction != BuyGalaSellTon {
		t.Errorf("Expected direction %s, got %s", BuyGalaSellTon, opp.Direction)
	}
	if opp.StonfiPrice != 100 {
		t.Errorf("Expected ston.fi price 100, got %f", opp.StonfiPrice)
	}
	if opp.GswapPrice != 99 {
		t.Errorf("Expected gswap price 99, got %f", opp.GswapPrice)
	}
	if opp.Spread != -1.0 {
		t.Errorf("Expected spread -1.0%%, got %f%%", opp.Spread)
	}
}

func TestDetectOpportunity_NoOpportunityBelowThreshold(t *testing.T) {
	// Setup: spread is 0.2% (below 0.3% threshold)
	// ston.fi: 100 GALA per TON
	// gswap: 100.2 GALA per GTON
	// spread = (100.2 - 100) / 100 * 100 = 0.2%
	dexClient := &mockDexClient{
		simulateSwapFunc: func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
			return &dex.SwapSimulation{
				OutputAmount: "100000000000", // 100 GALA (with 9 decimals)
			}, nil
		},
	}

	galaClient := &mockGalaClient{
		getPriceFunc: func(ctx context.Context, pair string) (*pb.PriceResponse, error) {
			return &pb.PriceResponse{
				Price: "100.2",
			}, nil
		},
	}

	engine := NewEngine(dexClient, galaClient)
	opp, err := engine.DetectOpportunity(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if opp != nil {
		t.Errorf("Expected no opportunity, got: %+v", opp)
	}
}

func TestDetectOpportunity_NoOpportunityEqualPrices(t *testing.T) {
	// Setup: prices are equal (spread = 0%)
	dexClient := &mockDexClient{
		simulateSwapFunc: func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
			return &dex.SwapSimulation{
				OutputAmount: "100000000000", // 100 GALA (with 9 decimals)
			}, nil
		},
	}

	galaClient := &mockGalaClient{
		getPriceFunc: func(ctx context.Context, pair string) (*pb.PriceResponse, error) {
			return &pb.PriceResponse{
				Price: "100",
			}, nil
		},
	}

	engine := NewEngine(dexClient, galaClient)
	opp, err := engine.DetectOpportunity(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if opp != nil {
		t.Errorf("Expected no opportunity, got: %+v", opp)
	}
}

func TestDetectOpportunity_DexClientError(t *testing.T) {
	// Setup: dex client returns error
	dexClient := &mockDexClient{
		simulateSwapFunc: func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
			return nil, errors.New("dex error")
		},
	}

	galaClient := &mockGalaClient{
		getPriceFunc: func(ctx context.Context, pair string) (*pb.PriceResponse, error) {
			return &pb.PriceResponse{
				Price: "100",
			}, nil
		},
	}

	engine := NewEngine(dexClient, galaClient)
	opp, err := engine.DetectOpportunity(context.Background())

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if opp != nil {
		t.Errorf("Expected no opportunity on error, got: %+v", opp)
	}
}

func TestDetectOpportunity_GalaClientError(t *testing.T) {
	// Setup: gala client returns error
	dexClient := &mockDexClient{
		simulateSwapFunc: func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
			return &dex.SwapSimulation{
				OutputAmount: "100000000000", // 100 GALA (with 9 decimals)
			}, nil
		},
	}

	galaClient := &mockGalaClient{
		getPriceFunc: func(ctx context.Context, pair string) (*pb.PriceResponse, error) {
			return nil, errors.New("gala error")
		},
	}

	engine := NewEngine(dexClient, galaClient)
	opp, err := engine.DetectOpportunity(context.Background())

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if opp != nil {
		t.Errorf("Expected no opportunity on error, got: %+v", opp)
	}
}

func TestDetectOpportunity_AtThreshold(t *testing.T) {
	// Setup: spread is just above 0.3% threshold
	// ston.fi: 100 GALA per TON
	// gswap: 100.31 GALA per GTON
	// spread = (100.31 - 100) / 100 * 100 = 0.31%
	dexClient := &mockDexClient{
		simulateSwapFunc: func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
			return &dex.SwapSimulation{
				OutputAmount: "100000000000", // 100 GALA (with 9 decimals)
			}, nil
		},
	}

	galaClient := &mockGalaClient{
		getPriceFunc: func(ctx context.Context, pair string) (*pb.PriceResponse, error) {
			return &pb.PriceResponse{
				Price: "100.31",
			}, nil
		},
	}

	engine := NewEngine(dexClient, galaClient)
	opp, err := engine.DetectOpportunity(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if opp == nil {
		t.Fatal("Expected opportunity just above threshold, got nil")
	}
	if opp.Direction != BuyTonSellGala {
		t.Errorf("Expected direction %s, got %s", BuyTonSellGala, opp.Direction)
	}
	if opp.Spread < 0.3 {
		t.Errorf("Expected spread >= 0.3%%, got %f%%", opp.Spread)
	}
}

func TestDetectOpportunity_NilDexClient(t *testing.T) {
	// Setup: dex client is nil
	galaClient := &mockGalaClient{
		getPriceFunc: func(ctx context.Context, pair string) (*pb.PriceResponse, error) {
			return &pb.PriceResponse{
				Price: "100",
			}, nil
		},
	}

	engine := NewEngine(nil, galaClient)
	opp, err := engine.DetectOpportunity(context.Background())

	if err == nil {
		t.Fatal("Expected error for nil dex client, got nil")
	}
	if opp != nil {
		t.Errorf("Expected no opportunity with nil client, got: %+v", opp)
	}
}

func TestDetectOpportunity_NilGalaClient(t *testing.T) {
	// Setup: gala client is nil
	dexClient := &mockDexClient{
		simulateSwapFunc: func(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
			return &dex.SwapSimulation{
				OutputAmount: "100000000000", // 100 GALA (with 9 decimals)
			}, nil
		},
	}

	engine := NewEngine(dexClient, nil)
	opp, err := engine.DetectOpportunity(context.Background())

	if err == nil {
		t.Fatal("Expected error for nil gala client, got nil")
	}
	if opp != nil {
		t.Errorf("Expected no opportunity with nil client, got: %+v", opp)
	}
}
