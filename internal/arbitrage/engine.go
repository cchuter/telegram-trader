package arbitrage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/galachain/pb"
)

const (
	// OneTON is 1 TON in nanotons (1 TON = 1,000,000,000 nanotons)
	OneTON = "1000000000"
	// GALATokenAddress is the GALA token address on TON blockchain
	GALATokenAddress = "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV"
	// MinSpreadThreshold is the minimum spread required for an opportunity (0.3%)
	MinSpreadThreshold = 0.3
)

// GalaClient defines the interface for GalaChain operations needed by the engine
type GalaClient interface {
	GetPrice(ctx context.Context, pair string) (*pb.PriceResponse, error)
}

// Engine handles arbitrage opportunity detection
type Engine struct {
	dexClient  dex.Client
	galaClient GalaClient
}

// NewEngine creates a new arbitrage engine
func NewEngine(dexClient dex.Client, galaClient GalaClient) *Engine {
	return &Engine{
		dexClient:  dexClient,
		galaClient: galaClient,
	}
}

// DetectOpportunity checks for arbitrage opportunities between ston.fi and gswap
// Returns an Opportunity if spread >= 0.3%, otherwise returns nil
func (e *Engine) DetectOpportunity(ctx context.Context) (*Opportunity, error) {
	// Fetch TON/GALA price from ston.fi
	stonfiPrice, err := e.getStonfiPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ston.fi price: %w", err)
	}

	// Fetch GTON/GALA price from GalaChain service
	gswapPrice, err := e.getGswapPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gswap price: %w", err)
	}

	// Calculate spread: (price_gswap - price_stonfi) / price_stonfi * 100
	spread := (gswapPrice - stonfiPrice) / stonfiPrice * 100

	// Check if spread meets minimum threshold
	if spread >= MinSpreadThreshold {
		// gswap is more expensive, buy on ston.fi, sell on gswap
		return &Opportunity{
			Direction:   BuyTonSellGala,
			StonfiPrice: stonfiPrice,
			GswapPrice:  gswapPrice,
			Spread:      spread,
		}, nil
	} else if spread <= -MinSpreadThreshold {
		// ston.fi is more expensive, buy on gswap, sell on ston.fi
		return &Opportunity{
			Direction:   BuyGalaSellTon,
			StonfiPrice: stonfiPrice,
			GswapPrice:  gswapPrice,
			Spread:      spread,
		}, nil
	}

	// No profitable opportunity
	return nil, nil
}

// getStonfiPrice fetches the TON/GALA price from ston.fi
func (e *Engine) getStonfiPrice(ctx context.Context) (float64, error) {
	if e.dexClient == nil {
		return 0, fmt.Errorf("dex client not available")
	}

	// Simulate swapping 1 TON to GALA
	sim, err := e.dexClient.SimulateSwap(ctx, "TON", GALATokenAddress, OneTON)
	if err != nil {
		return 0, fmt.Errorf("failed to simulate swap: %w", err)
	}

	// Convert output amount to GALA (assuming 9 decimals for GALA)
	galaAmount, err := strconv.ParseFloat(sim.OutputAmount, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse output amount: %w", err)
	}

	// Convert from smallest units to GALA
	price := galaAmount / 1e9
	return price, nil
}

// getGswapPrice fetches the GTON/GALA price from GalaChain service
func (e *Engine) getGswapPrice(ctx context.Context) (float64, error) {
	if e.galaClient == nil {
		return 0, fmt.Errorf("gala client not available")
	}

	// Fetch GTON/GALA price
	priceResp, err := e.galaClient.GetPrice(ctx, "GTON/GALA")
	if err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}

	// Parse the price string to float
	price, err := strconv.ParseFloat(priceResp.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	return price, nil
}
