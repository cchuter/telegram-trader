package mocks

import (
	"context"
	"testing"

	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockDEXClient_SimulateSwap(t *testing.T) {
	client := NewMockDEXClient()
	ctx := context.Background()

	fromToken := "TON"
	toToken := "GALA"
	amount := "1000000000"

	// Set simulation
	expectedSimulation := &dex.SwapSimulation{
		OutputAmount: "850000000000",
		Fee:          "5000000",
		Slippage:     1.0,
		PriceImpact:  0.05,
	}
	client.SetSimulation(fromToken, toToken, amount, expectedSimulation)

	// Simulate swap
	simulation, err := client.SimulateSwap(ctx, fromToken, toToken, amount)
	require.NoError(t, err)
	assert.Equal(t, expectedSimulation.OutputAmount, simulation.OutputAmount)
	assert.Equal(t, expectedSimulation.Fee, simulation.Fee)
	assert.Equal(t, expectedSimulation.Slippage, simulation.Slippage)
	assert.Equal(t, expectedSimulation.PriceImpact, simulation.PriceImpact)
}

func TestMockDEXClient_SimulateSwap_Default(t *testing.T) {
	client := NewMockDEXClient()
	ctx := context.Background()

	// Simulate swap without setting specific simulation
	simulation, err := client.SimulateSwap(ctx, "TON", "GALA", "1000000000")
	require.NoError(t, err)

	// Should return default simulation
	assert.NotEmpty(t, simulation.OutputAmount)
	assert.NotEmpty(t, simulation.Fee)
}

func TestMockDEXClient_Close(t *testing.T) {
	client := NewMockDEXClient()

	// Close
	err := client.Close()
	require.NoError(t, err)
	assert.Equal(t, 1, client.CloseCalls)
}

func TestMockDEXClient_ErrorSimulation(t *testing.T) {
	client := NewMockDEXClient()
	ctx := context.Background()

	// Simulate SimulateSwap error
	expectedErr := assert.AnError
	client.SetSimulateSwapErr(expectedErr)

	_, err := client.SimulateSwap(ctx, "TON", "GALA", "1000000000")
	assert.Equal(t, expectedErr, err)

	// Clear error
	client.SetSimulateSwapErr(nil)
	_, err = client.SimulateSwap(ctx, "TON", "GALA", "1000000000")
	assert.NoError(t, err)
}

func TestMockDEXClient_CallTracking(t *testing.T) {
	client := NewMockDEXClient()
	ctx := context.Background()

	fromToken := "TON"
	toToken := "GALA"
	amount1 := "1000000000"
	amount2 := "2000000000"

	// Make multiple calls
	_, _ = client.SimulateSwap(ctx, fromToken, toToken, amount1)
	_, _ = client.SimulateSwap(ctx, fromToken, toToken, amount1)
	_, _ = client.SimulateSwap(ctx, fromToken, toToken, amount2)

	// Verify call counts
	assert.Equal(t, 2, client.GetCallCount(fromToken, toToken, amount1))
	assert.Equal(t, 1, client.GetCallCount(fromToken, toToken, amount2))
	assert.Equal(t, 0, client.GetCallCount("GALA", "TON", amount1))
}

func TestMockDEXClient_Reset(t *testing.T) {
	client := NewMockDEXClient()
	ctx := context.Background()

	fromToken := "TON"
	toToken := "GALA"
	amount := "1000000000"

	// Set simulation and make calls
	client.SetSimulation(fromToken, toToken, amount, &dex.SwapSimulation{
		OutputAmount: "850000000000",
		Fee:          "5000000",
		Slippage:     1.0,
		PriceImpact:  0.05,
	})
	_, _ = client.SimulateSwap(ctx, fromToken, toToken, amount)

	// Reset
	client.Reset()

	// Verify reset - should return default simulation
	simulation, err := client.SimulateSwap(ctx, fromToken, toToken, amount)
	require.NoError(t, err)
	assert.Equal(t, "1000.0", simulation.OutputAmount)                  // Default value
	assert.Equal(t, 1, client.GetCallCount(fromToken, toToken, amount)) // Call count should start from 1
}
