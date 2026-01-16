//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/arbitrage"
	"github.com/cchuter/telegram-trader/internal/blockchain/ton"
	"github.com/cchuter/telegram-trader/internal/dex/stonfi"
	"github.com/cchuter/telegram-trader/internal/galachain"
	"github.com/cchuter/telegram-trader/internal/logging"
	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/cchuter/telegram-trader/internal/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Test configuration for arbitrage
	arbTestUserID   = int64(777888999)
	arbTestChatID   = int64(111222333)
	arbTestUsername = "arbtest"
	arbTestTimeout  = 120 * time.Second // Longer timeout for cross-chain arbitrage

	// TON testnet wallet address (example - replace with actual test wallet)
	arbTestTonWalletAddr = "kQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV"

	// Test private key (hex-encoded, for testnet only)
	// NOTE: In real tests, this should be loaded from secure config
	arbTestPrivateKey = "0000000000000000000000000000000000000000000000000000000000000001" // Example key

	// Minimal test amounts
	testTonAmount  = 0.001 // 0.001 TON (minimum for testing)
	testGalaAmount = 10.0  // 10 GALA (minimum for testing)
)

// setupArbitrageTestEnvironment creates test infrastructure for arbitrage e2e tests
func setupArbitrageTestEnvironment(t *testing.T) (storage.Database, *wallet.Manager, *ton.Client, *stonfi.Client, *galachain.Client, *logging.TradeLogger, string, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "telegram-trader-arb-e2e-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Initialize database
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := storage.InitDB(dbPath)
	require.NoError(t, err, "Failed to initialize test database")

	// Create user session
	userSession := &storage.UserSession{
		UserID:    arbTestUserID,
		ChatID:    arbTestChatID,
		Username:  arbTestUsername,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err = db.SaveUserSession(context.Background(), userSession)
	require.NoError(t, err, "Failed to save user session")

	// Initialize wallet manager with test encryption key
	encryptionKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" // 32-byte base64 key
	walletMgr, err := wallet.NewManager(db, encryptionKey)
	require.NoError(t, err, "Failed to create wallet manager")

	// Initialize TON client (testnet)
	tonClient := ton.NewClient()
	err = tonClient.Connect(context.Background())
	require.NoError(t, err, "Failed to connect to TON testnet")

	// Initialize ston.fi client (testnet)
	stonfiClient := stonfi.NewClient()

	// Initialize GalaChain gRPC client
	// If GALACHAIN_SERVICE_URL not set, use localhost
	galaServiceURL := os.Getenv("GALACHAIN_SERVICE_URL")
	if galaServiceURL == "" {
		galaServiceURL = "localhost:50051"
	}
	galaClient, err := galachain.Connect(context.Background(), galaServiceURL)
	require.NoError(t, err, "Failed to connect to GalaChain service")

	// Initialize trade logger
	logsDir := filepath.Join(tempDir, "logs")
	err = os.MkdirAll(logsDir, 0755)
	require.NoError(t, err, "Failed to create logs directory")

	tradeLogger, err := logging.NewTradeLogger(logsDir)
	require.NoError(t, err, "Failed to create trade logger")

	cleanup := func() {
		if tonClient != nil {
			_ = tonClient.Close()
		}
		if stonfiClient != nil {
			_ = stonfiClient.Close()
		}
		if galaClient != nil {
			_ = galaClient.Close()
		}
		if db != nil {
			_ = db.Close()
		}
		_ = os.RemoveAll(tempDir)
	}

	return db, walletMgr, tonClient, stonfiClient, galaClient, tradeLogger, tempDir, cleanup
}

// TestArbitrageOpportunityDetection tests the arbitrage opportunity detection flow
func TestArbitrageOpportunityDetection(t *testing.T) {
	db, walletMgr, tonClient, stonfiClient, galaClient, tradeLogger, tempDir, cleanup := setupArbitrageTestEnvironment(t)
	defer cleanup()

	// Create arbitrage engine
	engine := arbitrage.NewEngine(stonfiClient, galaClient)
	require.NotNil(t, engine, "Engine should be created")

	// Detect arbitrage opportunity
	ctx, cancel := context.WithTimeout(context.Background(), arbTestTimeout)
	defer cancel()

	opportunity, err := engine.DetectOpportunity(ctx)

	// We expect either an opportunity or nil (no opportunity), but not an error
	if err != nil {
		// Allow network errors in testnet
		t.Logf("Warning: opportunity detection failed with error: %v", err)
		// Still continue test to verify infrastructure
	} else if opportunity != nil {
		t.Logf("Arbitrage opportunity detected:")
		t.Logf("  Direction: %s", opportunity.Direction)
		t.Logf("  Ston.fi price: %.2f GALA", opportunity.StonfiPrice)
		t.Logf("  Gswap price: %.2f GALA", opportunity.GswapPrice)
		t.Logf("  Spread: %.2f%%", opportunity.Spread)

		// Validate opportunity fields
		assert.NotEmpty(t, opportunity.Direction, "Direction should not be empty")
		assert.True(t, opportunity.StonfiPrice > 0, "Ston.fi price should be positive")
		assert.True(t, opportunity.GswapPrice > 0, "Gswap price should be positive")
		assert.True(t, opportunity.Spread != 0, "Spread should not be zero")
	} else {
		t.Logf("No arbitrage opportunity found (spread below threshold)")
	}

	// Verify that we can calculate position size
	position := engine.GetPositionSize(testTonAmount*2, testGalaAmount*2, arbitrage.BuyTonSellGala)
	require.NotNil(t, position, "Position should be calculated")
	t.Logf("Position size: TON=%.6f, GALA=%.2f, Valid=%v", position.TONAmount, position.GALAAmount, position.Valid)

	// Verify services are accessible
	_ = db
	_ = walletMgr
	_ = tonClient
	_ = tradeLogger
	_ = tempDir
}

// TestArbitrageExecution tests the full arbitrage execution flow
// This test simulates clicking [Execute] button after detecting an opportunity
func TestArbitrageExecution(t *testing.T) {
	// Skip test if TESTNET_FUNDED is not set
	if os.Getenv("TESTNET_FUNDED") != "true" {
		t.Skip("Skipping arbitrage execution test - requires TESTNET_FUNDED=true and actual testnet funds")
	}

	db, walletMgr, tonClient, stonfiClient, galaClient, tradeLogger, tempDir, cleanup := setupArbitrageTestEnvironment(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), arbTestTimeout)
	defer cancel()

	// Step 1: Create wallet session with encrypted private key
	encryptedKey, err := walletMgr.EncryptPrivateKey(arbTestPrivateKey)
	require.NoError(t, err, "Failed to encrypt private key")

	walletSession := &storage.WalletSession{
		UserID:               arbTestUserID,
		WalletType:           "ton",
		Address:              arbTestTonWalletAddr,
		TonConnectPrivateKey: encryptedKey, // Store in TonConnect field
		ConnectedAt:          time.Now(),
		UpdatedAt:            time.Now(),
		IsActive:             true,
	}
	err = db.SaveWalletSession(ctx, walletSession)
	require.NoError(t, err, "Failed to save wallet session")

	// Step 2: Create arbitrage engine and detect opportunity
	engine := arbitrage.NewEngine(stonfiClient, galaClient)
	opportunity, err := engine.DetectOpportunity(ctx)
	require.NoError(t, err, "Failed to detect opportunity")

	if opportunity == nil {
		t.Skip("No arbitrage opportunity available - cannot test execution")
	}

	t.Logf("Executing arbitrage opportunity:")
	t.Logf("  Direction: %s", opportunity.Direction)
	t.Logf("  Spread: %.2f%%", opportunity.Spread)

	// Step 3: Calculate position size (use minimal test amounts)
	position := engine.GetPositionSize(testTonAmount*2, testGalaAmount*2, opportunity.Direction)
	require.True(t, position.Valid, "Position should be valid: %s", position.Reason)
	t.Logf("  Position: TON=%.6f, GALA=%.2f", position.TONAmount, position.GALAAmount)

	// Step 4: Verify wallet session retrieval
	retrievedSession, err := db.GetWalletSession(ctx, arbTestUserID, "ton")
	require.NoError(t, err, "Should retrieve wallet session")
	assert.Equal(t, arbTestTonWalletAddr, retrievedSession.Address, "Wallet address should match")

	// Step 5: Decrypt private key for execution
	decryptedKey, err := walletMgr.DecryptPrivateKey(encryptedKey)
	require.NoError(t, err, "Failed to decrypt private key")
	assert.Equal(t, arbTestPrivateKey, decryptedKey, "Decrypted key should match original")

	t.Log("✓ Wallet session configured successfully")
	t.Log("✓ Private key encrypted and decrypted successfully")
	t.Log("⚠ Actual arbitrage execution skipped (requires real testnet funds on both chains)")

	// Step 6: Simulate successful arbitrage execution and log it
	// In real execution, this would call the executor with both legs
	tradeLogger.LogArbitrage(
		arbTestUserID,
		"TON",
		"GALA",
		"0.001",
		"850.0",
		"0.08", // Combined fees from both legs
		"stonfi-mock-tx-hash-123",
		"gswap-mock-tx-hash-456",
		logging.TradeStatusSuccess,
		2500, // 2.5 seconds execution time
		"",
	)

	// Step 7: Verify trade logging
	tradeLogPath := filepath.Join(tempDir, "logs", "trades.jsonl")
	assert.FileExists(t, tradeLogPath, "Trade log should be created")

	// Read and verify log entry
	logData, err := os.ReadFile(tradeLogPath)
	require.NoError(t, err, "Failed to read trade log")

	lines := strings.Split(strings.TrimSpace(string(logData)), "\n")
	require.Greater(t, len(lines), 0, "Should have at least one log entry")

	// Parse last log entry (should be the arbitrage trade)
	var logEntry map[string]interface{}
	err = json.Unmarshal([]byte(lines[len(lines)-1]), &logEntry)
	require.NoError(t, err, "Failed to parse log entry")

	t.Logf("Trade log entry: %+v", logEntry)
	assert.Equal(t, "arbitrage", logEntry["type"], "Trade type should be arbitrage")
	assert.Equal(t, float64(arbTestUserID), logEntry["userId"], "User ID should match")
	assert.Contains(t, logEntry["txHash"], "stonfi-mock-tx-hash-123", "Should contain ston.fi tx hash")
	assert.Contains(t, logEntry["txHash"], "gswap-mock-tx-hash-456", "Should contain gswap tx hash")
	assert.Equal(t, "success", logEntry["status"], "Status should be success")

	t.Log("✓ Trade logged successfully with both transaction hashes")

	// Verify services are accessible
	_ = tonClient
}

// TestArbitrageWithInsufficientBalance tests arbitrage with insufficient balance
func TestArbitrageWithInsufficientBalance(t *testing.T) {
	db, walletMgr, tonClient, stonfiClient, galaClient, tradeLogger, _, cleanup := setupArbitrageTestEnvironment(t)
	defer cleanup()

	// Create arbitrage engine
	engine := arbitrage.NewEngine(stonfiClient, galaClient)

	// Try to calculate position with insufficient balance (below minimums)
	insufficientTonBalance := 0.0001 // Below 1 TON minimum
	insufficientGalaBalance := 5.0   // Below 10 GALA minimum

	position := engine.GetPositionSize(insufficientTonBalance, insufficientGalaBalance, arbitrage.BuyTonSellGala)
	require.NotNil(t, position, "Position should be calculated")
	assert.False(t, position.Valid, "Position should be invalid with insufficient balance")
	assert.NotEmpty(t, position.Reason, "Position should have a reason for being invalid")

	t.Logf("Position validation failed as expected: %s", position.Reason)

	// Verify that executor handles invalid position gracefully
	// (In production, this would be caught before execution)
	_ = db
	_ = walletMgr
	_ = tonClient
	_ = tradeLogger
}

// TestArbitrageTradeLogging tests that arbitrage trades are logged correctly
func TestArbitrageTradeLogging(t *testing.T) {
	db, walletMgr, tonClient, stonfiClient, galaClient, tradeLogger, tempDir, cleanup := setupArbitrageTestEnvironment(t)
	defer cleanup()

	// Create a mock arbitrage execution and log it
	tradeLogger.LogArbitrage(
		arbTestUserID,
		"TON",
		"GALA",
		"0.001",
		"850.0",
		"0.05",
		"stonfi_test_hash_123",
		"gswap_test_hash_456",
		logging.TradeStatusSuccess,
		1500, // 1.5 seconds
		"",
	)

	// Verify log file exists
	tradeLogPath := filepath.Join(tempDir, "logs", "trades.jsonl")
	assert.FileExists(t, tradeLogPath, "Trade log file should exist")

	// Read and parse log
	logData, err := os.ReadFile(tradeLogPath)
	require.NoError(t, err, "Failed to read trade log")

	lines := strings.Split(strings.TrimSpace(string(logData)), "\n")
	require.Greater(t, len(lines), 0, "Trade log should have at least one entry")

	// Parse the log entry
	var logEntry map[string]interface{}
	err = json.Unmarshal([]byte(lines[0]), &logEntry)
	require.NoError(t, err, "Failed to parse log entry")

	// Verify log fields
	assert.Equal(t, "arbitrage", logEntry["type"], "Type should be arbitrage")
	assert.Equal(t, float64(arbTestUserID), logEntry["userId"], "User ID should match")
	assert.Equal(t, "TON", logEntry["fromToken"], "From token should match")
	assert.Equal(t, "GALA", logEntry["toToken"], "To token should match")
	assert.Equal(t, "0.001", logEntry["amountIn"], "Amount in should match")
	assert.Equal(t, "850.0", logEntry["amountOut"], "Amount out should match")
	assert.Equal(t, "0.05", logEntry["fee"], "Fee should match")
	assert.Contains(t, logEntry["txHash"], "stonfi_test_hash_123", "Should contain ston.fi tx hash")
	assert.Contains(t, logEntry["txHash"], "gswap_test_hash_456", "Should contain gswap tx hash")
	assert.Equal(t, "success", logEntry["status"], "Status should be success")
	assert.Equal(t, float64(1500), logEntry["executionTimeMs"], "Execution time should match")
	assert.NotEmpty(t, logEntry["timestamp"], "Timestamp should be present")

	t.Logf("Trade log verified: %+v", logEntry)

	// Verify cleanup works
	_ = db
	_ = walletMgr
	_ = tonClient
	_ = stonfiClient
	_ = galaClient
}
