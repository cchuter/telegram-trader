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
	db         storage.Database
	encryption *EncryptionService
}

// NewManager creates a new wallet Manager instance
// encryptionKey should be a base64-encoded 32-byte key for AES-256
func NewManager(db storage.Database, encryptionKey string) (*Manager, error) {
	// Initialize encryption service
	encryption, err := NewEncryptionService(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryption service: %w", err)
	}

	return &Manager{
		db:         db,
		encryption: encryption,
	}, nil
}

// EncryptPrivateKey encrypts a private key for secure storage
// This method can be used when storing wallet credentials
func (m *Manager) EncryptPrivateKey(privateKey string) (string, error) {
	return m.encryption.Encrypt(privateKey)
}

// DecryptPrivateKey decrypts a private key for use
// This method can be used when retrieving wallet credentials
func (m *Manager) DecryptPrivateKey(encryptedKey string) (string, error) {
	return m.encryption.Decrypt(encryptedKey)
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
