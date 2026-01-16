package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cchuter/telegram-trader/internal/storage"
)

// MockDatabase is an in-memory implementation of storage.Database for testing
type MockDatabase struct {
	mu sync.RWMutex

	// In-memory storage
	UserSessions   map[int64]*storage.UserSession
	WalletSessions map[string]*storage.WalletSession // key: "userID:walletType"
	TradeHistory   map[int64][]*storage.TradeHistory // key: userID

	// Error simulation
	GetUserSessionErr    error
	SaveUserSessionErr   error
	GetWalletSessionErr  error
	SaveWalletSessionErr error
	GetTradeHistoryErr   error
	CloseErr             error
}

// NewMockDatabase creates a new in-memory mock database
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		UserSessions:   make(map[int64]*storage.UserSession),
		WalletSessions: make(map[string]*storage.WalletSession),
		TradeHistory:   make(map[int64][]*storage.TradeHistory),
	}
}

// GetUserSession retrieves a user session
func (m *MockDatabase) GetUserSession(ctx context.Context, userID int64) (*storage.UserSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetUserSessionErr != nil {
		return nil, m.GetUserSessionErr
	}

	session, exists := m.UserSessions[userID]
	if !exists {
		return nil, fmt.Errorf("user session not found")
	}

	return session, nil
}

// SaveUserSession saves a user session
func (m *MockDatabase) SaveUserSession(ctx context.Context, session *storage.UserSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SaveUserSessionErr != nil {
		return m.SaveUserSessionErr
	}

	// Copy the session to avoid external mutations
	sessionCopy := &storage.UserSession{
		UserID:    session.UserID,
		ChatID:    session.ChatID,
		Username:  session.Username,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		ExpiresAt: session.ExpiresAt,
	}

	m.UserSessions[session.UserID] = sessionCopy
	return nil
}

// GetWalletSession retrieves a wallet session
func (m *MockDatabase) GetWalletSession(ctx context.Context, userID int64, walletType string) (*storage.WalletSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetWalletSessionErr != nil {
		return nil, m.GetWalletSessionErr
	}

	key := fmt.Sprintf("%d:%s", userID, walletType)
	session, exists := m.WalletSessions[key]
	if !exists {
		return nil, fmt.Errorf("wallet session not found")
	}

	return session, nil
}

// SaveWalletSession saves a wallet session
func (m *MockDatabase) SaveWalletSession(ctx context.Context, session *storage.WalletSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SaveWalletSessionErr != nil {
		return m.SaveWalletSessionErr
	}

	key := fmt.Sprintf("%d:%s", session.UserID, session.WalletType)

	// Copy the session to avoid external mutations
	sessionCopy := &storage.WalletSession{
		UserID:                 session.UserID,
		WalletType:             session.WalletType,
		Address:                session.Address,
		ConnectedAt:            session.ConnectedAt,
		UpdatedAt:              session.UpdatedAt,
		IsActive:               session.IsActive,
		TonConnectClientID:     session.TonConnectClientID,
		TonConnectPrivateKey:   session.TonConnectPrivateKey,
		TonConnectWalletID:     session.TonConnectWalletID,
	}

	m.WalletSessions[key] = sessionCopy
	return nil
}

// GetTradeHistory retrieves trade history for a user
func (m *MockDatabase) GetTradeHistory(ctx context.Context, userID int64, limit int) ([]*storage.TradeHistory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetTradeHistoryErr != nil {
		return nil, m.GetTradeHistoryErr
	}

	history, exists := m.TradeHistory[userID]
	if !exists {
		return []*storage.TradeHistory{}, nil
	}

	// Apply limit
	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	// Return copies to prevent external mutation
	result := make([]*storage.TradeHistory, len(history))
	for i, trade := range history {
		tradeCopy := *trade
		result[i] = &tradeCopy
	}

	return result, nil
}

// Close closes the database connection (no-op for mock)
func (m *MockDatabase) Close() error {
	if m.CloseErr != nil {
		return m.CloseErr
	}
	return nil
}

// Reset clears all data from the mock database
func (m *MockDatabase) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UserSessions = make(map[int64]*storage.UserSession)
	m.WalletSessions = make(map[string]*storage.WalletSession)
	m.TradeHistory = make(map[int64][]*storage.TradeHistory)

	// Clear error simulations
	m.GetUserSessionErr = nil
	m.SaveUserSessionErr = nil
	m.GetWalletSessionErr = nil
	m.SaveWalletSessionErr = nil
	m.GetTradeHistoryErr = nil
	m.CloseErr = nil
}

// AddTradeHistory adds a trade record to the mock database
func (m *MockDatabase) AddTradeHistory(trade *storage.TradeHistory) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.TradeHistory[trade.UserID] == nil {
		m.TradeHistory[trade.UserID] = []*storage.TradeHistory{}
	}

	tradeCopy := *trade
	if tradeCopy.CreatedAt.IsZero() {
		tradeCopy.CreatedAt = time.Now()
	}

	m.TradeHistory[trade.UserID] = append(m.TradeHistory[trade.UserID], &tradeCopy)
}
