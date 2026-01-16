package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/storage"
)

// MockDatabase is a mock implementation of storage.Database for testing
type MockDatabase struct {
	trades []*storage.TradeHistory
	err    error
}

func (m *MockDatabase) GetUserSession(ctx context.Context, userID int64) (*storage.UserSession, error) {
	return nil, nil
}

func (m *MockDatabase) SaveUserSession(ctx context.Context, session *storage.UserSession) error {
	return nil
}

func (m *MockDatabase) GetWalletSession(ctx context.Context, userID int64, walletType string) (*storage.WalletSession, error) {
	return nil, nil
}

func (m *MockDatabase) SaveWalletSession(ctx context.Context, session *storage.WalletSession) error {
	return nil
}

func (m *MockDatabase) GetTradeHistory(ctx context.Context, userID int64, limit int) ([]*storage.TradeHistory, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.trades, nil
}

func (m *MockDatabase) Close() error {
	return nil
}

func TestGetTradeHistory(t *testing.T) {
	ctx := context.Background()

	// Test with mock trades
	now := time.Now()
	mockTrades := []*storage.TradeHistory{
		{
			ID:              1,
			UserID:          123,
			TradeType:       "swap",
			Chain:           "ton",
			FromToken:       "TON",
			ToToken:         "GALA",
			AmountIn:        "1.0",
			AmountOut:       "850.0",
			Fee:             "0.01",
			TxHashTon:       "abc123def456",
			Status:          "success",
			ExecutionTimeMs: 1500,
			CreatedAt:       now,
			CompletedAt:     &now,
		},
		{
			ID:              2,
			UserID:          123,
			TradeType:       "arbitrage",
			Chain:           "both",
			FromToken:       "TON",
			ToToken:         "GALA",
			AmountIn:        "2.0",
			AmountOut:       "1700.0",
			Fee:             "0.05",
			TxHashTon:       "def456ghi789",
			TxHashGala:      "ghi789jkl012",
			Status:          "success",
			ExecutionTimeMs: 2500,
			ProfitUSD:       "5.50",
			CreatedAt:       now.Add(-1 * time.Hour),
			CompletedAt:     &now,
		},
	}

	mockDB := &MockDatabase{trades: mockTrades}

	// Test fetching trade history
	trades, err := mockDB.GetTradeHistory(ctx, 123, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(trades) != 2 {
		t.Fatalf("Expected 2 trades, got %d", len(trades))
	}

	// Verify first trade
	if trades[0].TradeType != "swap" {
		t.Errorf("Expected trade type 'swap', got '%s'", trades[0].TradeType)
	}
	if trades[0].FromToken != "TON" {
		t.Errorf("Expected from token 'TON', got '%s'", trades[0].FromToken)
	}
	if trades[0].ToToken != "GALA" {
		t.Errorf("Expected to token 'GALA', got '%s'", trades[0].ToToken)
	}
	if trades[0].Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", trades[0].Status)
	}

	// Verify second trade (arbitrage)
	if trades[1].TradeType != "arbitrage" {
		t.Errorf("Expected trade type 'arbitrage', got '%s'", trades[1].TradeType)
	}
	if trades[1].Chain != "both" {
		t.Errorf("Expected chain 'both', got '%s'", trades[1].Chain)
	}
	if trades[1].TxHashGala != "ghi789jkl012" {
		t.Errorf("Expected gala tx hash 'ghi789jkl012', got '%s'", trades[1].TxHashGala)
	}
	if trades[1].ProfitUSD != "5.50" {
		t.Errorf("Expected profit USD '5.50', got '%s'", trades[1].ProfitUSD)
	}
}

func TestGetStatusEmoji(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"success", "✅"},
		{"pending", "⏳"},
		{"failed", "❌"},
		{"partial", "⚠️"},
		{"unknown", "❓"},
	}

	for _, tt := range tests {
		result := getStatusEmoji(tt.status)
		if result != tt.expected {
			t.Errorf("getStatusEmoji(%s) = %s; expected %s", tt.status, result, tt.expected)
		}
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		amount   string
		expected string
	}{
		{"1.00000000", "1"},
		{"1.50000000", "1.5"},
		{"0.12345678", "0.12345678"},
		{"123.45600000", "123.456"},
		{"", "N/A"},
		{"invalid", "invalid"},
	}

	for _, tt := range tests {
		result := formatAmount(tt.amount)
		if result != tt.expected {
			t.Errorf("formatAmount(%s) = %s; expected %s", tt.amount, result, tt.expected)
		}
	}
}

func TestTruncateHash(t *testing.T) {
	tests := []struct {
		hash     string
		expected string
	}{
		{"abc123def456ghi789jkl012mno345pqr678", "abc123de...pqr678"},
		{"short", "short"},
		{"", ""},
		{"abc123def456ghi789jk", "abc123def456ghi789jk"}, // 20 chars exactly, not truncated
		{"abc123def456ghi789jkl", "abc123de...789jkl"},   // 21 chars, truncated
	}

	for _, tt := range tests {
		result := truncateHash(tt.hash)
		if result != tt.expected {
			t.Errorf("truncateHash(%s) = %s; expected %s", tt.hash, result, tt.expected)
		}
	}
}
