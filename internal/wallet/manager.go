package wallet

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cchuter/telegram-trader/internal/storage"
)

// Manager handles wallet connection operations
type Manager struct {
	db         storage.Database
	encryption *EncryptionService
	connector  *TonConnector

	// Active TonConnect sessions (in-memory, indexed by user ID)
	sessions map[int64]*TonConnectSession
	mu       sync.RWMutex
}

// NewManager creates a new wallet Manager instance
// encryptionKey should be a base64-encoded 32-byte key for AES-256
func NewManager(db storage.Database, encryptionKey string) (*Manager, error) {
	// Initialize encryption service
	encryption, err := NewEncryptionService(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryption service: %w", err)
	}

	// Initialize TonConnect connector
	connector := NewTonConnector("", "") // Use default URLs

	return &Manager{
		db:         db,
		encryption: encryption,
		connector:  connector,
		sessions:   make(map[int64]*TonConnectSession),
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

// GetWalletSession retrieves a wallet session for a user
func (m *Manager) GetWalletSession(ctx context.Context, userID int64, walletType string) (*storage.WalletSession, error) {
	return m.db.GetWalletSession(ctx, userID, walletType)
}

// InitiateTonConnect initiates a TonConnect wallet connection session
// Returns the session and connection URLs (QR code deep link and TonKeeper URL)
func (m *Manager) InitiateTonConnect(ctx context.Context, userID int64) (*TonConnectSession, string, string, error) {
	// Create new TonConnect session
	session, err := m.connector.CreateSession()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create TonConnect session: %w", err)
	}

	// Generate connection URLs
	qrURL, err := m.connector.GenerateQRCodeURL(session)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate QR code URL: %w", err)
	}

	tonkeeperURL, err := m.connector.GenerateTonKeeperURL(session)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate TonKeeper URL: %w", err)
	}

	// Store session in memory
	m.mu.Lock()
	m.sessions[userID] = session
	m.mu.Unlock()

	// Start listening for wallet connection
	err = m.connector.ListenForConnection(ctx, session, func(address string, err error) {
		if err != nil {
			// Log error but don't fail the connection
			fmt.Printf("TonConnect error for user %d: %v\n", userID, err)
			return
		}

		// Save wallet session to database
		if address != "" {
			_ = m.saveTonConnectSession(ctx, userID, session, address)
		}
	})

	if err != nil {
		return nil, "", "", fmt.Errorf("failed to start listening for connection: %w", err)
	}

	return session, qrURL, tonkeeperURL, nil
}

// saveTonConnectSession saves a successful TonConnect session to database
func (m *Manager) saveTonConnectSession(ctx context.Context, userID int64, session *TonConnectSession, address string) error {
	// Encrypt the private key before storage
	encryptedPrivateKey, err := m.encryption.Encrypt(string(session.PrivateKey))
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Create wallet session
	walletSession := &storage.WalletSession{
		UserID:               userID,
		WalletType:           "ton",
		Address:              address,
		ConnectedAt:          session.CreatedAt,
		UpdatedAt:            time.Now(),
		IsActive:             true,
		TonConnectClientID:   session.ClientID,
		TonConnectPrivateKey: encryptedPrivateKey,
		TonConnectWalletID:   session.WalletID,
	}

	// Save to database
	if err := m.db.SaveWalletSession(ctx, walletSession); err != nil {
		return fmt.Errorf("failed to save wallet session: %w", err)
	}

	return nil
}

// GetTonConnectSession retrieves an active TonConnect session for a user
func (m *Manager) GetTonConnectSession(userID int64) (*TonConnectSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, exists := m.sessions[userID]
	return session, exists
}

// StopTonConnect stops listening for TonConnect events for a user
func (m *Manager) StopTonConnect(userID int64) {
	m.mu.Lock()
	session, exists := m.sessions[userID]
	if exists {
		m.connector.StopListening(session)
		delete(m.sessions, userID)
	}
	m.mu.Unlock()
}

// DisconnectWallet disconnects a wallet for a user
// Deletes the wallet session from database and clears encrypted keys from memory
func (m *Manager) DisconnectWallet(ctx context.Context, userID int64, walletType string) error {
	// Get the existing wallet session
	session, err := m.db.GetWalletSession(ctx, userID, walletType)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	// Mark session as inactive and clear sensitive data
	session.IsActive = false
	session.TonConnectPrivateKey = "" // Clear encrypted private key
	session.UpdatedAt = time.Now()

	// Update database to mark inactive and clear keys
	if err := m.db.SaveWalletSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update wallet session: %w", err)
	}

	// Stop any active TonConnect sessions in memory
	if walletType == "ton" {
		m.StopTonConnect(userID)
	}

	return nil
}

// GenerateTonConnectURL generates a placeholder TonConnect URL
// Deprecated: Use InitiateTonConnect instead
func (m *Manager) GenerateTonConnectURL() string {
	// Placeholder for backward compatibility
	return "ton://connect?placeholder=true"
}
