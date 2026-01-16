package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/galachain/pb"
	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/cchuter/telegram-trader/test/mocks"
)

// SetupMockDatabase creates a fresh mock database for testing
func SetupMockDatabase() *mocks.MockDatabase {
	return mocks.NewMockDatabase()
}

// SetupMockBlockchain creates a mock blockchain client with default test data
func SetupMockBlockchain() *mocks.MockBlockchainClient {
	client := mocks.NewMockBlockchainClient()

	// Add some default test balances
	client.SetBalance("UQTest1234567890abcdefghijklmnopqrstuvwxyz", "10.5")
	client.SetBalance("UQTest9876543210zyxwvutsrqponmlkjihgfedcba", "25.0")

	return client
}

// SetupMockDEX creates a mock DEX client with default test data
func SetupMockDEX() *mocks.MockDEXClient {
	client := mocks.NewMockDEXClient()

	// Add some default test simulations
	// TON -> GALA swap
	client.SetSimulation("TON", "GALA", "1000000000", &dex.SwapSimulation{
		OutputAmount: "850000000000",
		Fee:          "5000000",
		Slippage:     1.0,
		PriceImpact:  0.05,
	})

	return client
}

// SetupMockGRPC creates a mock gRPC client with default test data
func SetupMockGRPC() *mocks.MockGRPCClient {
	client := mocks.NewMockGRPCClient()

	// Add default test balances
	client.SetBalance(123456, []*pb.TokenBalance{
		{Token: "GALA", Balance: "5000.0"},
		{Token: "GTON", Balance: "2.5"},
	})

	// Add default test prices
	client.SetPrice("GTON/GALA", &pb.PriceResponse{
		Price:      "855.0",
		Timestamp:  time.Now().Unix(),
		Bid:        "854.5",
		Ask:        "855.5",
		Volume_24H: "1000000.0",
	})

	return client
}

// CreateTestUserSession creates a user session for testing
func CreateTestUserSession(userID int64) *storage.UserSession {
	return &storage.UserSession{
		UserID:    userID,
		ChatID:    userID,
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

// CreateTestWalletSession creates a wallet session for testing
func CreateTestWalletSession(userID int64, walletType string, address string) *storage.WalletSession {
	return &storage.WalletSession{
		UserID:      userID,
		WalletType:  walletType,
		Address:     address,
		IsActive:    true,
		ConnectedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CreateTestTradeHistory creates a trade history record for testing
func CreateTestTradeHistory(userID int64, tradeType string) *storage.TradeHistory {
	return &storage.TradeHistory{
		UserID:          userID,
		TradeType:       tradeType,
		Chain:           "ton",
		FromToken:       "TON",
		ToToken:         "GALA",
		AmountIn:        "1.0",
		AmountOut:       "850.0",
		Fee:             "0.005",
		TxHashTon:       "abc123def456",
		TxHashGala:      "",
		Status:          "completed",
		ErrorMessage:    "",
		ExecutionTimeMs: 1500,
		ProfitUSD:       "0.0",
		CreatedAt:       time.Now(),
		CompletedAt:     timePtr(time.Now()),
	}
}

// timePtr returns a pointer to the given time
func timePtr(t time.Time) *time.Time {
	return &t
}

// AssertNoError is a helper to fail tests on unexpected errors
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError is a helper to fail tests when an error is expected but not received
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error but got nil", message)
	}
}

// AssertEqual is a generic equality assertion helper
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotNil is a helper to check if a value is not nil
func AssertNotNil(t *testing.T, value interface{}, message string) {
	t.Helper()
	if value == nil {
		t.Fatalf("%s: expected non-nil value", message)
	}
}

// ContextWithTimeout creates a context with a standard test timeout
func ContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// WaitForCondition polls a condition function until it returns true or times out
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return
		}

		<-ticker.C
		if time.Now().After(deadline) {
			t.Fatalf("%s: condition not met within timeout", message)
		}
	}
}
