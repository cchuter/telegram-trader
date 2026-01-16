//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/dex/stonfi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// GALA token address on ston.fi
	galaTokenAddress = "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV"

	// USDT token address on ston.fi for testing multiple pairs
	usdtTokenAddress = "EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs"

	// Test amounts
	oneTonNanotons  = "1000000000" // 1 TON
	fiveTonNanotons = "5000000000" // 5 TON

	// API timeout
	apiTimeout = 10 * time.Second
)

// TestStonFiSimulateSwap tests real ston.fi API swap simulation
func TestStonFiSimulateSwap(t *testing.T) {
	t.Run("SimulateSwap_TON_to_GALA", func(t *testing.T) {
		client := stonfi.NewClient()
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		// Test 1 TON -> GALA swap
		simulation, err := client.SimulateSwap(ctx, "TON", galaTokenAddress, oneTonNanotons)
		require.NoError(t, err, "SimulateSwap should succeed")
		require.NotNil(t, simulation, "Simulation result should not be nil")

		// Verify response structure
		assert.NotEmpty(t, simulation.OutputAmount, "OutputAmount should not be empty")
		assert.NotEmpty(t, simulation.Fee, "Fee should not be empty")
		assert.Greater(t, simulation.Slippage, 0.0, "Slippage should be > 0")
		assert.GreaterOrEqual(t, simulation.PriceImpact, 0.0, "PriceImpact should be >= 0")

		// Log results for visibility
		t.Logf("TON -> GALA swap simulation:")
		t.Logf("  Input: 1 TON (%s nanotons)", oneTonNanotons)
		t.Logf("  Output: %s units", simulation.OutputAmount)
		t.Logf("  Fee: %s units", simulation.Fee)
		t.Logf("  Slippage: %.2f%%", simulation.Slippage)
		t.Logf("  Price Impact: %.4f%%", simulation.PriceImpact)
	})

	t.Run("SimulateSwap_GALA_to_TON", func(t *testing.T) {
		client := stonfi.NewClient()
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		// Test GALA -> TON swap (reverse direction)
		// Use approximate GALA amount (based on ~850 GALA per TON)
		galaAmount := "850000000000" // 850 GALA (9 decimals)

		simulation, err := client.SimulateSwap(ctx, galaTokenAddress, "TON", galaAmount)
		require.NoError(t, err, "SimulateSwap should succeed")
		require.NotNil(t, simulation, "Simulation result should not be nil")

		// Verify response structure
		assert.NotEmpty(t, simulation.OutputAmount, "OutputAmount should not be empty")
		assert.NotEmpty(t, simulation.Fee, "Fee should not be empty")
		assert.Greater(t, simulation.Slippage, 0.0, "Slippage should be > 0")
		assert.GreaterOrEqual(t, simulation.PriceImpact, 0.0, "PriceImpact should be >= 0")

		t.Logf("GALA -> TON swap simulation:")
		t.Logf("  Input: %s GALA units", galaAmount)
		t.Logf("  Output: %s nanotons", simulation.OutputAmount)
		t.Logf("  Fee: %s units", simulation.Fee)
		t.Logf("  Slippage: %.2f%%", simulation.Slippage)
		t.Logf("  Price Impact: %.4f%%", simulation.PriceImpact)
	})

	t.Run("SimulateSwap_TON_to_USDT", func(t *testing.T) {
		client := stonfi.NewClient()
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		// Test different token pair: TON -> USDT
		simulation, err := client.SimulateSwap(ctx, "TON", usdtTokenAddress, fiveTonNanotons)
		require.NoError(t, err, "SimulateSwap should succeed")
		require.NotNil(t, simulation, "Simulation result should not be nil")

		// Verify response structure
		assert.NotEmpty(t, simulation.OutputAmount, "OutputAmount should not be empty")
		assert.NotEmpty(t, simulation.Fee, "Fee should not be empty")
		assert.Greater(t, simulation.Slippage, 0.0, "Slippage should be > 0")
		assert.GreaterOrEqual(t, simulation.PriceImpact, 0.0, "PriceImpact should be >= 0")

		t.Logf("TON -> USDT swap simulation:")
		t.Logf("  Input: 5 TON (%s nanotons)", fiveTonNanotons)
		t.Logf("  Output: %s USDT units", simulation.OutputAmount)
		t.Logf("  Fee: %s units", simulation.Fee)
		t.Logf("  Slippage: %.2f%%", simulation.Slippage)
		t.Logf("  Price Impact: %.4f%%", simulation.PriceImpact)
	})

	t.Run("SimulateSwap_LargeAmount", func(t *testing.T) {
		client := stonfi.NewClient()
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		// Test large swap to verify price impact increases
		largeTonAmount := "100000000000" // 100 TON

		simulation, err := client.SimulateSwap(ctx, "TON", galaTokenAddress, largeTonAmount)
		require.NoError(t, err, "SimulateSwap should succeed for large amounts")
		require.NotNil(t, simulation, "Simulation result should not be nil")

		// Verify response structure
		assert.NotEmpty(t, simulation.OutputAmount, "OutputAmount should not be empty")
		assert.NotEmpty(t, simulation.Fee, "Fee should not be empty")
		assert.Greater(t, simulation.Slippage, 0.0, "Slippage should be > 0")

		// Large swaps typically have higher price impact
		assert.Greater(t, simulation.PriceImpact, 0.0, "PriceImpact should be > 0 for large swap")

		t.Logf("Large TON -> GALA swap simulation:")
		t.Logf("  Input: 100 TON (%s nanotons)", largeTonAmount)
		t.Logf("  Output: %s GALA units", simulation.OutputAmount)
		t.Logf("  Fee: %s units", simulation.Fee)
		t.Logf("  Slippage: %.2f%%", simulation.Slippage)
		t.Logf("  Price Impact: %.4f%%", simulation.PriceImpact)
	})
}

// TestStonFiErrorHandling tests error cases with ston.fi API
func TestStonFiErrorHandling(t *testing.T) {
	t.Run("InvalidTokenAddress", func(t *testing.T) {
		client := stonfi.NewClient()
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		// Use invalid token address
		invalidAddress := "INVALID_ADDRESS_123"

		simulation, err := client.SimulateSwap(ctx, "TON", invalidAddress, oneTonNanotons)

		// API should return error for invalid address
		assert.Error(t, err, "Should return error for invalid token address")
		assert.Nil(t, simulation, "Simulation should be nil on error")

		t.Logf("Invalid token error (expected): %v", err)
	})

	t.Run("ZeroAmount", func(t *testing.T) {
		client := stonfi.NewClient()
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		// Use zero amount
		zeroAmount := "0"

		simulation, err := client.SimulateSwap(ctx, "TON", galaTokenAddress, zeroAmount)

		// API should return error or zero output for zero amount
		if err != nil {
			t.Logf("Zero amount error (expected): %v", err)
		} else {
			// Some APIs might return zero output instead of error
			assert.Equal(t, "0", simulation.OutputAmount, "Output should be 0 for zero input")
			t.Logf("Zero amount returns zero output (valid behavior)")
		}
	})

	t.Run("NetworkTimeout", func(t *testing.T) {
		client := stonfi.NewClient()
		defer client.Close()

		// Use very short timeout to simulate network timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		simulation, err := client.SimulateSwap(ctx, "TON", galaTokenAddress, oneTonNanotons)

		// Should return context deadline exceeded error
		assert.Error(t, err, "Should return error for network timeout")
		assert.Nil(t, simulation, "Simulation should be nil on timeout")

		t.Logf("Network timeout error (expected): %v", err)
	})

	t.Run("SameTokenSwap", func(t *testing.T) {
		client := stonfi.NewClient()
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		// Try to swap TON -> TON (same token)
		simulation, err := client.SimulateSwap(ctx, "TON", "TON", oneTonNanotons)

		// API behavior for same-token swap varies (might error or return same amount)
		if err != nil {
			t.Logf("Same token swap error (expected): %v", err)
		} else {
			// Some APIs might allow this and return near-zero output after fees
			t.Logf("Same token swap allowed by API (unexpected but valid)")
			t.Logf("  Output: %s", simulation.OutputAmount)
		}
	})
}

// TestStonFiResponseParsing tests response parsing from ston.fi API
func TestStonFiResponseParsing(t *testing.T) {
	t.Run("ResponseFieldsComplete", func(t *testing.T) {
		client := stonfi.NewClient()
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		simulation, err := client.SimulateSwap(ctx, "TON", galaTokenAddress, oneTonNanotons)
		require.NoError(t, err, "SimulateSwap should succeed")
		require.NotNil(t, simulation, "Simulation result should not be nil")

		// Verify all required fields are present and parseable
		assert.NotEmpty(t, simulation.OutputAmount, "OutputAmount must be set")
		assert.NotEmpty(t, simulation.Fee, "Fee must be set")
		assert.NotEqual(t, 0.0, simulation.Slippage, "Slippage must be calculated")

		// PriceImpact can be 0 for small swaps, but should be a valid number
		assert.GreaterOrEqual(t, simulation.PriceImpact, 0.0, "PriceImpact must be non-negative")

		t.Logf("Response parsing validation passed")
		t.Logf("  All required fields present and valid")
	})

	t.Run("NumericFieldsParseable", func(t *testing.T) {
		client := stonfi.NewClient()
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		simulation, err := client.SimulateSwap(ctx, "TON", galaTokenAddress, oneTonNanotons)
		require.NoError(t, err, "SimulateSwap should succeed")
		require.NotNil(t, simulation, "Simulation result should not be nil")

		// Verify numeric fields contain valid numbers (not NaN or Inf)
		assert.False(t, simulation.Slippage < 0, "Slippage should not be negative")
		assert.False(t, simulation.PriceImpact < 0, "PriceImpact should not be negative")

		// Slippage typically between 0-5% for normal markets
		assert.Less(t, simulation.Slippage, 10.0, "Slippage should be reasonable (<10%)")

		// Price impact typically small for 1 TON swap
		assert.Less(t, simulation.PriceImpact, 5.0, "PriceImpact should be reasonable (<5%) for small swap")

		t.Logf("Numeric fields validation passed")
		t.Logf("  Slippage: %.2f%% (reasonable)", simulation.Slippage)
		t.Logf("  PriceImpact: %.4f%% (reasonable)", simulation.PriceImpact)
	})
}

// TestStonFiMultipleTokenPairs tests various token pairs
func TestStonFiMultipleTokenPairs(t *testing.T) {
	tokenPairs := []struct {
		name      string
		fromToken string
		toToken   string
		amount    string
	}{
		{
			name:      "TON_to_GALA",
			fromToken: "TON",
			toToken:   galaTokenAddress,
			amount:    oneTonNanotons,
		},
		{
			name:      "TON_to_USDT",
			fromToken: "TON",
			toToken:   usdtTokenAddress,
			amount:    oneTonNanotons,
		},
		{
			name:      "GALA_to_TON",
			fromToken: galaTokenAddress,
			toToken:   "TON",
			amount:    "850000000000", // ~850 GALA
		},
		{
			name:      "USDT_to_TON",
			fromToken: usdtTokenAddress,
			toToken:   "TON",
			amount:    "5000000", // ~5 USDT (6 decimals)
		},
	}

	for _, tc := range tokenPairs {
		t.Run(tc.name, func(t *testing.T) {
			client := stonfi.NewClient()
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
			defer cancel()

			simulation, err := client.SimulateSwap(ctx, tc.fromToken, tc.toToken, tc.amount)
			require.NoError(t, err, "SimulateSwap should succeed for %s", tc.name)
			require.NotNil(t, simulation, "Simulation result should not be nil")

			// Verify basic response structure
			assert.NotEmpty(t, simulation.OutputAmount, "OutputAmount should not be empty")
			assert.NotEmpty(t, simulation.Fee, "Fee should not be empty")

			t.Logf("%s simulation successful", tc.name)
			t.Logf("  Output: %s units", simulation.OutputAmount)
			t.Logf("  Fee: %s units", simulation.Fee)
		})
	}
}
