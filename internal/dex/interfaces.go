package dex

import "context"

// SwapSimulation contains the result of a swap simulation
type SwapSimulation struct {
	OutputAmount string  // Expected output amount in token units
	Fee          string  // Transaction fee
	Slippage     float64 // Slippage percentage
	PriceImpact  float64 // Price impact percentage
}

// Client defines the interface for DEX operations
type Client interface {
	// SimulateSwap simulates a token swap and returns expected output
	// fromToken: source token address (e.g., "TON" or token address)
	// toToken: destination token address
	// amount: input amount in token units
	SimulateSwap(ctx context.Context, fromToken, toToken, amount string) (*SwapSimulation, error)

	// Close closes any open connections
	Close() error
}
