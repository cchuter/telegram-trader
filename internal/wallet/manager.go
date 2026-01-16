package wallet

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cchuter/telegram-trader/internal/storage"
)

// Manager handles wallet connection operations
type Manager struct {
	db storage.Database
}

// NewManager creates a new wallet Manager instance
func NewManager(db storage.Database) *Manager {
	return &Manager{
		db: db,
	}
}

// ConnectTonWallet connects a TON wallet for the user
// For POC: accepts wallet address directly instead of real TonConnect flow
func (m *Manager) ConnectTonWallet(ctx context.Context, userID int64, walletAddress string) error {
	// Basic validation - TON addresses typically start with UQ or EQ
	if !strings.HasPrefix(walletAddress, "UQ") && !strings.HasPrefix(walletAddress, "EQ") {
		return fmt.Errorf("invalid TON wallet address format")
	}

	// Create wallet session
	session := &storage.WalletSession{
		UserID:      userID,
		WalletType:  "ton",
		Address:     walletAddress,
		ConnectedAt: time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	// Save to database
	if err := m.db.SaveWalletSession(ctx, session); err != nil {
		return fmt.Errorf("failed to save wallet session: %w", err)
	}

	return nil
}

// GetTonWallet retrieves the connected TON wallet for a user
func (m *Manager) GetTonWallet(ctx context.Context, userID int64) (*storage.WalletSession, error) {
	return m.db.GetWalletSession(ctx, userID, "ton")
}

// GenerateTonConnectURL generates a placeholder TonConnect URL
// For POC: this is not fully implemented, just returns a placeholder
func (m *Manager) GenerateTonConnectURL() string {
	// Placeholder for future TonConnect integration
	return "ton://connect?placeholder=true"
}
