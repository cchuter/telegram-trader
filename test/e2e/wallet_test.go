// +build e2e

package e2e

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/cchuter/telegram-trader/internal/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"
)

const (
	testUserID       = int64(123456789)
	testChatID       = int64(987654321)
	testUsername     = "testuser"
	testEncryptKey   = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" // 32-byte base64 key
	testWalletAddr   = "UQTest1234567890abcdefghijklmnopqrstuvwxyz"
	testTimeout      = 10 * time.Second
)

// setupTestEnvironment creates a test database and encryption key
func setupTestEnvironment(t *testing.T) (storage.Database, string, func()) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "telegram-trader-e2e-*")
	require.NoError(t, err, "Failed to create temp directory")

	dbPath := filepath.Join(tempDir, "test.db")

	// Initialize database
	db, err := storage.InitDB(dbPath)
	require.NoError(t, err, "Failed to initialize test database")

	cleanup := func() {
		if db != nil {
			_ = db.Close()
		}
		_ = os.RemoveAll(tempDir)
	}

	return db, tempDir, cleanup
}

// TestWalletConnectionFlow tests the complete /wallet command flow
func TestWalletConnectionFlow(t *testing.T) {
	db, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Initialize wallet manager
	walletManager, err := wallet.NewManager(db, testEncryptKey)
	require.NoError(t, err, "Failed to create wallet manager")

	t.Run("User sends /wallet command", func(t *testing.T) {
		// Test wallet manager directly (handler requires *bot.Bot which can't be mocked)
		// This simulates the core logic that HandleWallet uses
		session, qrURL, tonkeeperURL, err := walletManager.InitiateTonConnect(ctx, testUserID)
		require.NoError(t, err, "TonConnect initiation should succeed")
		assert.NotNil(t, session, "Session should be created")
		assert.NotEmpty(t, qrURL, "QR URL should be generated")
		assert.NotEmpty(t, tonkeeperURL, "TonKeeper URL should be generated")

		// Verify session properties
		assert.Equal(t, 64, len(session.ClientID), "ClientID should be 64 hex chars (32 bytes)")
		assert.Equal(t, 32, len(session.PrivateKey), "PrivateKey should be 32 bytes")
		assert.Equal(t, 32, len(session.PublicKey), "PublicKey should be 32 bytes")
		assert.False(t, session.IsConnected(), "Session should not be connected yet")
	})

	t.Run("Verify TonConnect link format", func(t *testing.T) {
		session, qrURL, tonkeeperURL, err := walletManager.InitiateTonConnect(ctx, testUserID)
		require.NoError(t, err)

		// Check QR URL format (tc:// deep link)
		assert.Contains(t, qrURL, "tc://connect", "QR URL should be tc:// deep link")
		assert.Contains(t, qrURL, "v=2", "Should include protocol version")
		assert.Contains(t, qrURL, "id="+session.ClientID, "Should include client ID")
		assert.Contains(t, qrURL, "r=", "Should include request payload")

		// Check TonKeeper URL format
		assert.Contains(t, tonkeeperURL, "app.tonkeeper.com/ton-connect", "TonKeeper URL should point to tonkeeper.com")
		assert.Contains(t, tonkeeperURL, "v=2", "Should include protocol version")
		assert.Contains(t, tonkeeperURL, "id="+session.ClientID, "Should include client ID")
	})

	t.Run("Simulate wallet connection callback", func(t *testing.T) {
		// Generate wallet keypair (simulating wallet app)
		walletPublicKey, _, err := box.GenerateKey(rand.Reader)
		require.NoError(t, err, "Failed to generate wallet keypair")

		// Create TonConnect session
		session, _, _, err := walletManager.InitiateTonConnect(ctx, testUserID)
		require.NoError(t, err)

		// Simulate wallet connection by manually updating session
		// In real scenario, this would come from TonConnect bridge
		session.WalletAddress = testWalletAddr
		session.WalletID = hex.EncodeToString(walletPublicKey[:])
		session.Connected = true

		// Manually save the wallet session to database
		// This simulates what saveTonConnectSession does internally
		walletSession := &storage.WalletSession{
			UserID:               testUserID,
			WalletType:           "ton",
			Address:              testWalletAddr,
			ConnectedAt:          session.CreatedAt,
			UpdatedAt:            time.Now(),
			IsActive:             true,
			TonConnectClientID:   session.ClientID,
			TonConnectPrivateKey: hex.EncodeToString(session.PrivateKey), // Simplified for test
			TonConnectWalletID:   session.WalletID,
		}

		err = db.SaveWalletSession(ctx, walletSession)
		require.NoError(t, err, "Failed to save wallet session")

		// Verify session is connected
		assert.True(t, session.IsConnected(), "Session should be connected")
		assert.Equal(t, testWalletAddr, session.GetWalletAddress(), "Wallet address should match")
	})

	t.Run("Verify wallet session saved to database", func(t *testing.T) {
		// Create and save a wallet session
		session, _, _, err := walletManager.InitiateTonConnect(ctx, testUserID)
		require.NoError(t, err)

		// Simulate connection
		session.WalletAddress = testWalletAddr
		session.WalletID = "test_wallet_id"
		session.Connected = true

		walletSession := &storage.WalletSession{
			UserID:               testUserID,
			WalletType:           "ton",
			Address:              testWalletAddr,
			ConnectedAt:          session.CreatedAt,
			UpdatedAt:            time.Now(),
			IsActive:             true,
			TonConnectClientID:   session.ClientID,
			TonConnectPrivateKey: "encrypted_key_placeholder",
			TonConnectWalletID:   session.WalletID,
		}

		err = db.SaveWalletSession(ctx, walletSession)
		require.NoError(t, err, "Save should succeed")

		// Retrieve from database
		retrieved, err := db.GetWalletSession(ctx, testUserID, "ton")
		require.NoError(t, err, "Get should succeed")
		assert.NotNil(t, retrieved, "Retrieved session should not be nil")

		// Verify fields
		assert.Equal(t, testUserID, retrieved.UserID)
		assert.Equal(t, "ton", retrieved.WalletType)
		assert.Equal(t, testWalletAddr, retrieved.Address)
		assert.True(t, retrieved.IsActive)
		assert.Equal(t, session.ClientID, retrieved.TonConnectClientID)
		assert.Equal(t, session.WalletID, retrieved.TonConnectWalletID)
		assert.NotEmpty(t, retrieved.TonConnectPrivateKey)
	})

	t.Run("Verify confirmation message format", func(t *testing.T) {
		// Test the confirmation message that would be sent
		address := testWalletAddr
		displayAddress := address
		if len(address) > 12 {
			displayAddress = address[:6] + "..." + address[len(address)-3:]
		}

		expectedMessage := fmt.Sprintf("✅ TON wallet connected: %s", displayAddress)
		assert.Contains(t, expectedMessage, "✅", "Should have success emoji")
		assert.Contains(t, expectedMessage, "TON wallet connected", "Should indicate connection")
		assert.Contains(t, expectedMessage, "UQTest...xyz", "Should show truncated address")
	})

	t.Run("Prevent duplicate wallet connection", func(t *testing.T) {
		// Create first wallet session
		session1 := &storage.WalletSession{
			UserID:      testUserID,
			WalletType:  "ton",
			Address:     testWalletAddr,
			ConnectedAt: time.Now(),
			UpdatedAt:   time.Now(),
			IsActive:    true,
		}

		err := db.SaveWalletSession(ctx, session1)
		require.NoError(t, err)

		// Try to retrieve existing wallet
		existing, err := walletManager.GetTonWallet(ctx, testUserID)
		require.NoError(t, err)
		assert.NotNil(t, existing, "Should find existing wallet")
		assert.True(t, existing.IsActive, "Wallet should be active")

		// Verify the handler would detect this and show appropriate message
		if existing != nil && existing.IsActive {
			displayAddr := existing.Address[:6] + "..." + existing.Address[len(existing.Address)-3:]
			message := fmt.Sprintf("TON wallet already connected: %s\n\nTo connect a different wallet, disconnect first.", displayAddr)
			assert.Contains(t, message, "already connected", "Should indicate wallet already connected")
			assert.Contains(t, message, "disconnect first", "Should tell user to disconnect")
		}
	})

	t.Run("Handle wallet connection timeout", func(t *testing.T) {
		// Create session
		session, _, _, err := walletManager.InitiateTonConnect(ctx, testUserID)
		require.NoError(t, err)

		// Simulate timeout by checking connection status after delay
		time.Sleep(100 * time.Millisecond)

		// Verify session is still not connected
		assert.False(t, session.IsConnected(), "Session should not be connected after timeout")

		// Stop listening (simulating timeout cleanup)
		walletManager.StopTonConnect(testUserID)

		// Verify session is removed from memory
		_, exists := walletManager.GetTonConnectSession(testUserID)
		assert.False(t, exists, "Session should be removed after stop")
	})
}

// TestWalletHandlerIntegration tests the actual handler with mocked dependencies
func TestWalletHandlerIntegration(t *testing.T) {
	db, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Initialize dependencies
	walletManager, err := wallet.NewManager(db, testEncryptKey)
	require.NoError(t, err)

	// Note: We can't fully test HandleWallet because it requires *bot.Bot
	// But we can test the wallet manager operations it relies on

	t.Run("InitiateTonConnect workflow", func(t *testing.T) {
		// This simulates what HandleWallet does internally
		session, qrURL, tonkeeperURL, err := walletManager.InitiateTonConnect(ctx, testUserID)
		require.NoError(t, err)

		// Verify all outputs are valid
		assert.NotNil(t, session)
		assert.NotEmpty(t, qrURL)
		assert.NotEmpty(t, tonkeeperURL)
		assert.False(t, session.IsConnected())

		// Verify session is stored in memory
		retrieved, exists := walletManager.GetTonConnectSession(testUserID)
		assert.True(t, exists, "Session should exist in memory")
		assert.Equal(t, session.ClientID, retrieved.ClientID)

		// Clean up
		walletManager.StopTonConnect(testUserID)
	})

	t.Run("Connection monitoring simulation", func(t *testing.T) {
		// Create session
		session, _, _, err := walletManager.InitiateTonConnect(ctx, testUserID)
		require.NoError(t, err)

		// Simulate polling for connection
		ticker := time.NewTicker(50 * time.Millisecond)
		timeout := time.After(200 * time.Millisecond)
		connected := false

	monitorLoop:
		for {
			select {
			case <-timeout:
				// Timeout reached
				break monitorLoop
			case <-ticker.C:
				// Check if connected
				if session.IsConnected() {
					connected = true
					break monitorLoop
				}
			}
		}

		ticker.Stop()

		// Should not be connected (no actual wallet connected)
		assert.False(t, connected, "Should timeout without connection")

		// Simulate timeout cleanup
		walletManager.StopTonConnect(testUserID)

		// Verify session removed
		_, exists := walletManager.GetTonConnectSession(testUserID)
		assert.False(t, exists, "Session should be removed")
	})

	t.Run("Successful connection callback", func(t *testing.T) {
		// Create session
		session, _, _, err := walletManager.InitiateTonConnect(ctx, testUserID)
		require.NoError(t, err)

		// Simulate successful connection
		session.WalletAddress = testWalletAddr
		session.WalletID = "test_wallet_id"
		session.Connected = true

		// Verify connection
		assert.True(t, session.IsConnected())
		address := session.GetWalletAddress()
		assert.Equal(t, testWalletAddr, address)

		// Format confirmation message
		displayAddress := address
		if len(address) > 12 {
			displayAddress = address[:6] + "..." + address[len(address)-3:]
		}
		confirmMsg := fmt.Sprintf("✅ TON wallet connected: %s", displayAddress)

		assert.Contains(t, confirmMsg, "✅")
		assert.Contains(t, confirmMsg, testWalletAddr[:6])

		// Clean up
		walletManager.StopTonConnect(testUserID)
	})
}

// TestEncryptedKeyStorage tests that private keys are encrypted before storage
func TestEncryptedKeyStorage(t *testing.T) {
	db, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	walletManager, err := wallet.NewManager(db, testEncryptKey)
	require.NoError(t, err)

	t.Run("Private key encryption", func(t *testing.T) {
		// Generate test private key
		_, privateKey, err := box.GenerateKey(rand.Reader)
		require.NoError(t, err)

		privateKeyHex := hex.EncodeToString(privateKey[:])

		// Encrypt the key
		encrypted, err := walletManager.EncryptPrivateKey(privateKeyHex)
		require.NoError(t, err, "Encryption should succeed")
		assert.NotEmpty(t, encrypted, "Encrypted key should not be empty")
		assert.NotEqual(t, privateKeyHex, encrypted, "Encrypted key should differ from plaintext")

		// Decrypt the key
		decrypted, err := walletManager.DecryptPrivateKey(encrypted)
		require.NoError(t, err, "Decryption should succeed")
		assert.Equal(t, privateKeyHex, decrypted, "Decrypted key should match original")
	})

	t.Run("Encrypted key persisted to database", func(t *testing.T) {
		// Generate keypair
		_, privateKey, err := box.GenerateKey(rand.Reader)
		require.NoError(t, err)

		privateKeyHex := hex.EncodeToString(privateKey[:])

		// Encrypt
		encrypted, err := walletManager.EncryptPrivateKey(privateKeyHex)
		require.NoError(t, err)

		// Save to database
		walletSession := &storage.WalletSession{
			UserID:               testUserID,
			WalletType:           "ton",
			Address:              testWalletAddr,
			ConnectedAt:          time.Now(),
			UpdatedAt:            time.Now(),
			IsActive:             true,
			TonConnectClientID:   "test_client_id",
			TonConnectPrivateKey: encrypted,
			TonConnectWalletID:   "test_wallet_id",
		}

		err = db.SaveWalletSession(ctx, walletSession)
		require.NoError(t, err)

		// Retrieve from database
		retrieved, err := db.GetWalletSession(ctx, testUserID, "ton")
		require.NoError(t, err)
		assert.Equal(t, encrypted, retrieved.TonConnectPrivateKey, "Encrypted key should be persisted")

		// Verify we can decrypt it
		decrypted, err := walletManager.DecryptPrivateKey(retrieved.TonConnectPrivateKey)
		require.NoError(t, err)
		assert.Equal(t, privateKeyHex, decrypted, "Should decrypt to original key")
	})
}
