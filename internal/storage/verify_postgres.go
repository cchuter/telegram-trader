// +build manual

package storage

import (
	"context"
	"fmt"
	"log"
	"os"
)

// VerifyPostgresImplementation is a manual verification script
// Run with: DATABASE_URL=postgres://user:pass@localhost:5432/testdb go run internal/storage/verify_postgres.go
func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Parse database URL
	db, err := ParseDatabaseURL(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse database URL: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test 1: User session operations
	fmt.Println("Test 1: User session operations...")
	userSession := &UserSession{
		UserID:   99999,
		ChatID:   88888,
		Username: "verify_test_user",
	}

	if err := db.SaveUserSession(ctx, userSession); err != nil {
		log.Fatalf("Failed to save user session: %v", err)
	}
	fmt.Println("✓ User session saved")

	retrieved, err := db.GetUserSession(ctx, userSession.UserID)
	if err != nil {
		log.Fatalf("Failed to get user session: %v", err)
	}
	if retrieved.UserID != userSession.UserID {
		log.Fatalf("User ID mismatch: expected %d, got %d", userSession.UserID, retrieved.UserID)
	}
	fmt.Println("✓ User session retrieved")

	// Test 2: Wallet session operations
	fmt.Println("\nTest 2: Wallet session operations...")
	walletSession := &WalletSession{
		UserID:     99999,
		WalletType: "ton",
		Address:    "UQVerifyTest123",
		IsActive:   true,
	}

	if err := db.SaveWalletSession(ctx, walletSession); err != nil {
		log.Fatalf("Failed to save wallet session: %v", err)
	}
	fmt.Println("✓ Wallet session saved")

	retrievedWallet, err := db.GetWalletSession(ctx, walletSession.UserID, walletSession.WalletType)
	if err != nil {
		log.Fatalf("Failed to get wallet session: %v", err)
	}
	if retrievedWallet.Address != walletSession.Address {
		log.Fatalf("Address mismatch: expected %s, got %s", walletSession.Address, retrievedWallet.Address)
	}
	fmt.Println("✓ Wallet session retrieved")

	// Test 3: TonConnect fields
	fmt.Println("\nTest 3: TonConnect fields...")
	walletSession.TonConnectClientID = "client_verify_123"
	walletSession.TonConnectPrivateKey = "encrypted_key_verify"
	walletSession.TonConnectWalletID = "wallet_verify_456"

	if err := db.SaveWalletSession(ctx, walletSession); err != nil {
		log.Fatalf("Failed to update wallet session: %v", err)
	}
	fmt.Println("✓ TonConnect fields saved")

	retrievedWallet, err = db.GetWalletSession(ctx, walletSession.UserID, walletSession.WalletType)
	if err != nil {
		log.Fatalf("Failed to get wallet session: %v", err)
	}
	if retrievedWallet.TonConnectClientID != walletSession.TonConnectClientID {
		log.Fatalf("TonConnect client ID mismatch")
	}
	fmt.Println("✓ TonConnect fields retrieved")

	fmt.Println("\n✅ All PostgreSQL implementation tests passed!")
}
