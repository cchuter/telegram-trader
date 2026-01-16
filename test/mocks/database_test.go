package mocks

import (
	"context"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockDatabase_UserSession(t *testing.T) {
	db := NewMockDatabase()
	ctx := context.Background()

	session := &storage.UserSession{
		UserID:    123456,
		ChatID:    123456,
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Save session
	err := db.SaveUserSession(ctx, session)
	require.NoError(t, err)

	// Retrieve session
	retrieved, err := db.GetUserSession(ctx, 123456)
	require.NoError(t, err)
	assert.Equal(t, session.UserID, retrieved.UserID)
	assert.Equal(t, session.Username, retrieved.Username)

	// Non-existent session
	_, err = db.GetUserSession(ctx, 999999)
	assert.Error(t, err)
}

func TestMockDatabase_WalletSession(t *testing.T) {
	db := NewMockDatabase()
	ctx := context.Background()

	session := &storage.WalletSession{
		UserID:      123456,
		WalletType:  "ton",
		Address:     "UQTest1234567890",
		IsActive:    true,
		ConnectedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save session
	err := db.SaveWalletSession(ctx, session)
	require.NoError(t, err)

	// Retrieve session
	retrieved, err := db.GetWalletSession(ctx, 123456, "ton")
	require.NoError(t, err)
	assert.Equal(t, session.UserID, retrieved.UserID)
	assert.Equal(t, session.WalletType, retrieved.WalletType)
	assert.Equal(t, session.Address, retrieved.Address)

	// Non-existent session
	_, err = db.GetWalletSession(ctx, 999999, "ton")
	assert.Error(t, err)
}

func TestMockDatabase_TradeHistory(t *testing.T) {
	db := NewMockDatabase()
	ctx := context.Background()

	userID := int64(123456)

	// Add trade records
	trade1 := &storage.TradeHistory{
		UserID:          userID,
		TradeType:       "swap",
		Chain:           "ton",
		FromToken:       "TON",
		ToToken:         "GALA",
		AmountIn:        "1.0",
		AmountOut:       "850.0",
		Fee:             "0.005",
		TxHashTon:       "tx1",
		Status:          "completed",
		ExecutionTimeMs: 1500,
		CreatedAt:       time.Now().Add(-2 * time.Hour),
	}

	trade2 := &storage.TradeHistory{
		UserID:          userID,
		TradeType:       "arbitrage",
		Chain:           "ton,gala",
		FromToken:       "TON",
		ToToken:         "GALA",
		AmountIn:        "2.0",
		AmountOut:       "1700.0",
		Fee:             "0.01",
		TxHashTon:       "tx2",
		TxHashGala:      "tx3",
		Status:          "completed",
		ExecutionTimeMs: 2500,
		ProfitUSD:       "5.0",
		CreatedAt:       time.Now().Add(-1 * time.Hour),
	}

	db.AddTradeHistory(trade1)
	db.AddTradeHistory(trade2)

	// Retrieve all trades
	trades, err := db.GetTradeHistory(ctx, userID, 0)
	require.NoError(t, err)
	assert.Len(t, trades, 2)
	assert.Equal(t, "swap", trades[0].TradeType)
	assert.Equal(t, "arbitrage", trades[1].TradeType)

	// Retrieve with limit
	trades, err = db.GetTradeHistory(ctx, userID, 1)
	require.NoError(t, err)
	assert.Len(t, trades, 1)

	// Non-existent user
	trades, err = db.GetTradeHistory(ctx, 999999, 0)
	require.NoError(t, err)
	assert.Empty(t, trades)
}

func TestMockDatabase_ErrorSimulation(t *testing.T) {
	db := NewMockDatabase()
	ctx := context.Background()

	// Simulate GetUserSession error
	expectedErr := assert.AnError
	db.GetUserSessionErr = expectedErr
	_, err := db.GetUserSession(ctx, 123456)
	assert.Equal(t, expectedErr, err)

	// Simulate SaveUserSession error
	db.SaveUserSessionErr = expectedErr
	session := &storage.UserSession{UserID: 123456}
	err = db.SaveUserSession(ctx, session)
	assert.Equal(t, expectedErr, err)
}

func TestMockDatabase_Reset(t *testing.T) {
	db := NewMockDatabase()
	ctx := context.Background()

	// Add data
	session := &storage.UserSession{
		UserID:    123456,
		ChatID:    123456,
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err := db.SaveUserSession(ctx, session)
	require.NoError(t, err)

	// Reset database
	db.Reset()

	// Verify data is cleared
	_, err = db.GetUserSession(ctx, 123456)
	assert.Error(t, err)
}

func TestMockDatabase_ConcurrentAccess(t *testing.T) {
	db := NewMockDatabase()
	ctx := context.Background()

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(userID int64) {
			session := &storage.UserSession{
				UserID:    userID,
				ChatID:    userID,
				Username:  "testuser",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}
			_ = db.SaveUserSession(ctx, session)
			done <- true
		}(int64(i))
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all sessions were saved
	for i := 0; i < 10; i++ {
		_, err := db.GetUserSession(ctx, int64(i))
		assert.NoError(t, err)
	}
}
