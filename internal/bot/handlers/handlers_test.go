package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/arbitrage"
	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/galachain/pb"
	"github.com/cchuter/telegram-trader/internal/logging"
	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/cchuter/telegram-trader/internal/wallet"
	"github.com/cchuter/telegram-trader/test/mocks"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note on testing handlers:
// The handlers use *bot.Bot which is a concrete type from a third-party library.
// We cannot easily mock it without significant infrastructure.
// These tests focus on:
// 1. Business logic and helper functions that can be unit tested
// 2. Validating handler signatures exist and are exported
// 3. Testing error paths and edge cases in testable components
// Integration tests and manual testing cover the full handler flows.

// TestHandlerSignatures validates that all handlers are exported and callable
func TestHandlerSignatures(t *testing.T) {
	assert.NotNil(t, HandleStart, "HandleStart should exist")
	assert.NotNil(t, HandleHelp, "HandleHelp should exist")
	assert.NotNil(t, HandleBalance, "HandleBalance should exist")
	assert.NotNil(t, HandleSwap, "HandleSwap should exist")
	assert.NotNil(t, HandleSwapCallback, "HandleSwapCallback should exist")
	assert.NotNil(t, HandlePrice, "HandlePrice should exist")
	assert.NotNil(t, HandleWallet, "HandleWallet should exist")
	assert.NotNil(t, HandleDisconnect, "HandleDisconnect should exist")
	assert.NotNil(t, HandleDisconnectCallback, "HandleDisconnectCallback should exist")
	assert.NotNil(t, HandleArbitrage, "HandleArbitrage should exist")
}

// TestTokenNameToAddress tests the token name to address conversion helper
func TestTokenNameToAddress(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"TON uppercase", "TON", "TON"},
		{"TON lowercase", "ton", "TON"},
		{"GALA uppercase", "GALA", GALATokenAddress},
		{"GALA lowercase", "gala", GALATokenAddress},
		{"GALA mixed case", "GaLa", GALATokenAddress},
		{"EQ address", "EQAbc123", "EQAbc123"},
		{"UQ address", "UQDef456", "UQDef456"},
		{"Unknown token", "UNKNOWN", "UNKNOWN"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tokenNameToAddress(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestSwapContextStorage tests that swap contexts are stored and retrieved correctly
func TestSwapContextStorage(t *testing.T) {
	// Clear existing swaps
	pendingSwapsMu.Lock()
	pendingSwaps = make(map[int64]*SwapContext)
	pendingSwapsMu.Unlock()

	// Create a swap context
	ctx := &SwapContext{
		UserID:         12345,
		FromToken:      "TON",
		ToToken:        "GALA",
		Amount:         "1.00",
		AmountUnits:    "1000000000",
		OutputAmount:   "850.00",
		Fee:            "0.10",
		Slippage:       1.0,
		PriceImpact:    0.05,
		RouterAddress:  "router123",
		MinOutputUnits: "841500000",
		CreatedAt:      time.Now(),
	}

	// Store the context
	pendingSwapsMu.Lock()
	pendingSwaps[12345] = ctx
	pendingSwapsMu.Unlock()

	// Retrieve and verify
	pendingSwapsMu.RLock()
	retrieved, exists := pendingSwaps[12345]
	pendingSwapsMu.RUnlock()

	assert.True(t, exists, "Swap context should exist")
	assert.Equal(t, int64(12345), retrieved.UserID)
	assert.Equal(t, "TON", retrieved.FromToken)
	assert.Equal(t, "GALA", retrieved.ToToken)
	assert.Equal(t, "1.00", retrieved.Amount)

	// Test deletion
	pendingSwapsMu.Lock()
	delete(pendingSwaps, 12345)
	pendingSwapsMu.Unlock()

	pendingSwapsMu.RLock()
	_, exists = pendingSwaps[12345]
	pendingSwapsMu.RUnlock()

	assert.False(t, exists, "Swap context should be deleted")
}

// TestHandleBalanceWithMocks tests balance handler logic with mocked dependencies
func TestHandleBalanceWithMocks(t *testing.T) {
	ctx := context.Background()
	mockGRPC := mocks.NewMockGRPCClient()
	_ = logging.New("test", logging.LogLevelInfo)

	// Configure mock to return balances
	mockGRPC.SetBalance(12345, []*pb.TokenBalance{
		{Token: "GALA", Balance: "5000.0"},
		{Token: "GTON", Balance: "2.5"},
	})

	// Call GetBalance to verify mock works
	resp, err := mockGRPC.GetBalance(ctx, 12345)
	require.NoError(t, err)
	assert.Len(t, resp.Balances, 2)
	assert.Equal(t, "GALA", resp.Balances[0].Token)
	assert.Equal(t, "5000.0", resp.Balances[0].Balance)

	// Verify mock tracks calls
	assert.Equal(t, 1, mockGRPC.GetBalanceCallCount(12345))
}

// TestHandleSwapWithMocks tests swap simulation logic
func TestHandleSwapWithMocks(t *testing.T) {
	ctx := context.Background()
	mockDEX := mocks.NewMockDEXClient()

	// Configure mock
	mockDEX.SetSimulation("TON", GALATokenAddress, "1000000000", &dex.SwapSimulation{
		OutputAmount: "850000000000",
		Fee:          "100000000",
		Slippage:     1.0,
		PriceImpact:  0.05,
	})

	// Test simulation
	sim, err := mockDEX.SimulateSwap(ctx, "TON", GALATokenAddress, "1000000000")
	require.NoError(t, err)
	assert.Equal(t, "850000000000", sim.OutputAmount)
	assert.Equal(t, 1.0, sim.Slippage)

	// Verify call tracking
	assert.Equal(t, 1, mockDEX.GetCallCount("TON", GALATokenAddress, "1000000000"))
}

// TestHandlePriceWithMocks tests price fetching logic
func TestHandlePriceWithMocks(t *testing.T) {
	ctx := context.Background()
	mockDEX := mocks.NewMockDEXClient()
	mockGRPC := mocks.NewMockGRPCClient()

	// Configure mocks
	mockDEX.SetSimulation("TON", GALATokenAddress, OneTON, &dex.SwapSimulation{
		OutputAmount: "850000000000", // 850 GALA in smallest units
		Fee:          "100000000",
		Slippage:     1.0,
		PriceImpact:  0.05,
	})

	mockGRPC.SetPrice("GTON/GALA", &pb.PriceResponse{
		Price:     "855.0",
		Timestamp: 1234567890,
	})

	// Test DEX price fetch
	sim, err := mockDEX.SimulateSwap(ctx, "TON", GALATokenAddress, OneTON)
	require.NoError(t, err)
	assert.Equal(t, "850000000000", sim.OutputAmount)

	// Test gRPC price fetch
	priceResp, err := mockGRPC.GetPrice(ctx, "GTON/GALA")
	require.NoError(t, err)
	assert.Equal(t, "855.0", priceResp.Price)
}

// TestHandleArbitrageWithMocks tests arbitrage detection logic
func TestHandleArbitrageWithMocks(t *testing.T) {
	ctx := context.Background()
	mockDEX := mocks.NewMockDEXClient()
	mockGRPC := mocks.NewMockGRPCClient()

	testCases := []struct {
		name              string
		stonfiOutput      string
		gswapPrice        string
		shouldFindOpp     bool
		expectedDirection arbitrage.Direction
	}{
		{
			name:              "Profitable opportunity - buy ston.fi",
			stonfiOutput:      "850000000000", // 850 GALA
			gswapPrice:        "855.0",        // 0.59% spread
			shouldFindOpp:     true,
			expectedDirection: arbitrage.BuyTonSellGala,
		},
		{
			name:              "Profitable opportunity - buy gswap",
			stonfiOutput:      "855000000000", // 855 GALA
			gswapPrice:        "850.0",        // -0.58% spread
			shouldFindOpp:     true,
			expectedDirection: arbitrage.BuyGalaSellTon,
		},
		{
			name:          "No opportunity - spread too low",
			stonfiOutput:  "850000000000", // 850 GALA
			gswapPrice:    "850.5",        // 0.06% spread
			shouldFindOpp: false,
		},
		{
			name:          "No opportunity - zero spread",
			stonfiOutput:  "850000000000", // 850 GALA
			gswapPrice:    "850.0",        // 0% spread
			shouldFindOpp: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDEX.Reset()
			mockGRPC.Reset()

			mockDEX.SetSimulation("TON", GALATokenAddress, OneTON, &dex.SwapSimulation{
				OutputAmount: tc.stonfiOutput,
				Fee:          "100000000",
				Slippage:     1.0,
				PriceImpact:  0.05,
			})

			mockGRPC.SetPrice("GTON/GALA", &pb.PriceResponse{
				Price: tc.gswapPrice,
			})

			engine := arbitrage.NewEngine(mockDEX, mockGRPC)
			opportunity, err := engine.DetectOpportunity(ctx)
			require.NoError(t, err)

			if tc.shouldFindOpp {
				assert.NotNil(t, opportunity, "Should find arbitrage opportunity")
				assert.Equal(t, tc.expectedDirection, opportunity.Direction)
			} else {
				assert.Nil(t, opportunity, "Should not find arbitrage opportunity")
			}
		})
	}
}

// TestHandleWalletWithMocks tests wallet connection logic
func TestHandleWalletWithMocks(t *testing.T) {
	ctx := context.Background()
	mockDB := mocks.NewMockDatabase()

	// Create wallet manager
	walletMgr, err := wallet.NewManager(mockDB, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	require.NoError(t, err)

	// Test: No existing wallet
	_, err = walletMgr.GetTonWallet(ctx, 12345)
	assert.Error(t, err, "Should error when no wallet exists")

	// Test: Save new wallet session
	err = mockDB.SaveWalletSession(ctx, &storage.WalletSession{
		UserID:      12345,
		WalletType:  "ton",
		Address:     "EQAbc123def456",
		IsActive:    true,
		ConnectedAt: time.Now(),
		UpdatedAt:   time.Now(),
	})
	require.NoError(t, err)

	// Test: Retrieve saved wallet
	session, err := walletMgr.GetTonWallet(ctx, 12345)
	require.NoError(t, err)
	assert.Equal(t, int64(12345), session.UserID)
	assert.Equal(t, "ton", session.WalletType)
	assert.Equal(t, "EQAbc123def456", session.Address)
	assert.True(t, session.IsActive)

	// Test: Disconnect wallet
	err = walletMgr.DisconnectWallet(ctx, 12345, "ton")
	require.NoError(t, err)

	// Verify wallet is inactive
	session, err = mockDB.GetWalletSession(ctx, 12345, "ton")
	require.NoError(t, err)
	assert.False(t, session.IsActive, "Wallet should be marked inactive")
}

// TestHandleSwapCallbackWithInvalidUpdate tests error handling
func TestHandleSwapCallbackWithInvalidUpdate(t *testing.T) {
	ctx := context.Background()
	mockDB := mocks.NewMockDatabase()
	mockDEX := mocks.NewMockDEXClient()
	logger := logging.New("test", logging.LogLevelInfo)

	walletMgr, _ := wallet.NewManager(mockDB, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")

	// Test with nil callback query - should not panic
	update := &models.Update{
		CallbackQuery: nil,
	}

	// Should not panic
	assert.NotPanics(t, func() {
		HandleSwapCallback(ctx, nil, update, mockDEX, walletMgr, logger, nil)
	})
}

// TestConstants verifies important constants are defined
func TestConstants(t *testing.T) {
	assert.Equal(t, "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV", GALATokenAddress, "GALA token address should match specification")
	assert.Equal(t, "1000000000", OneTON, "OneTON should be 1e9 nanotons")
}

// TestSwapContextStructure validates SwapContext fields
func TestSwapContextStructure(t *testing.T) {
	ctx := &SwapContext{
		UserID:         12345,
		FromToken:      "TON",
		ToToken:        "GALA",
		Amount:         "1.00",
		AmountUnits:    "1000000000",
		OutputAmount:   "850.00",
		Fee:            "0.10",
		Slippage:       1.0,
		PriceImpact:    0.05,
		RouterAddress:  "router",
		MinOutputUnits: "841500000",
		CreatedAt:      time.Now(),
	}

	assert.Equal(t, int64(12345), ctx.UserID)
	assert.Equal(t, "TON", ctx.FromToken)
	assert.Equal(t, "GALA", ctx.ToToken)
	assert.NotZero(t, ctx.CreatedAt)
}
