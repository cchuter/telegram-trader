package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInitDB(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize database
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}
}

func TestUserSessionOperations(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test saving a user session
	userSession := &UserSession{
		UserID:   123456,
		ChatID:   654321,
		Username: "testuser",
	}

	err = db.SaveUserSession(ctx, userSession)
	if err != nil {
		t.Fatalf("SaveUserSession failed: %v", err)
	}

	// Test retrieving the user session
	retrieved, err := db.GetUserSession(ctx, 123456)
	if err != nil {
		t.Fatalf("GetUserSession failed: %v", err)
	}

	if retrieved.UserID != userSession.UserID {
		t.Errorf("Expected UserID %d, got %d", userSession.UserID, retrieved.UserID)
	}
	if retrieved.ChatID != userSession.ChatID {
		t.Errorf("Expected ChatID %d, got %d", userSession.ChatID, retrieved.ChatID)
	}
	if retrieved.Username != userSession.Username {
		t.Errorf("Expected Username %s, got %s", userSession.Username, retrieved.Username)
	}

	// Test updating a user session
	userSession.Username = "updateduser"
	err = db.SaveUserSession(ctx, userSession)
	if err != nil {
		t.Fatalf("SaveUserSession (update) failed: %v", err)
	}

	retrieved, err = db.GetUserSession(ctx, 123456)
	if err != nil {
		t.Fatalf("GetUserSession (after update) failed: %v", err)
	}

	if retrieved.Username != "updateduser" {
		t.Errorf("Expected updated Username %s, got %s", "updateduser", retrieved.Username)
	}

	// Test retrieving non-existent user session
	_, err = db.GetUserSession(ctx, 999999)
	if err == nil {
		t.Errorf("Expected error when getting non-existent user session, got nil")
	}
}

func TestWalletSessionOperations(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create a user session first (foreign key constraint)
	userSession := &UserSession{
		UserID:   123456,
		ChatID:   654321,
		Username: "testuser",
	}
	err = db.SaveUserSession(ctx, userSession)
	if err != nil {
		t.Fatalf("SaveUserSession failed: %v", err)
	}

	// Test saving a wallet session
	walletSession := &WalletSession{
		UserID:     123456,
		WalletType: "ton",
		Address:    "EQTest123Address",
		IsActive:   true,
	}

	err = db.SaveWalletSession(ctx, walletSession)
	if err != nil {
		t.Fatalf("SaveWalletSession failed: %v", err)
	}

	// Test retrieving the wallet session
	retrieved, err := db.GetWalletSession(ctx, 123456, "ton")
	if err != nil {
		t.Fatalf("GetWalletSession failed: %v", err)
	}

	if retrieved.UserID != walletSession.UserID {
		t.Errorf("Expected UserID %d, got %d", walletSession.UserID, retrieved.UserID)
	}
	if retrieved.WalletType != walletSession.WalletType {
		t.Errorf("Expected WalletType %s, got %s", walletSession.WalletType, retrieved.WalletType)
	}
	if retrieved.Address != walletSession.Address {
		t.Errorf("Expected Address %s, got %s", walletSession.Address, retrieved.Address)
	}
	if retrieved.IsActive != walletSession.IsActive {
		t.Errorf("Expected IsActive %v, got %v", walletSession.IsActive, retrieved.IsActive)
	}

	// Test updating a wallet session
	walletSession.Address = "EQNewAddress456"
	walletSession.IsActive = false
	err = db.SaveWalletSession(ctx, walletSession)
	if err != nil {
		t.Fatalf("SaveWalletSession (update) failed: %v", err)
	}

	retrieved, err = db.GetWalletSession(ctx, 123456, "ton")
	if err != nil {
		t.Fatalf("GetWalletSession (after update) failed: %v", err)
	}

	if retrieved.Address != "EQNewAddress456" {
		t.Errorf("Expected updated Address %s, got %s", "EQNewAddress456", retrieved.Address)
	}
	if retrieved.IsActive != false {
		t.Errorf("Expected updated IsActive %v, got %v", false, retrieved.IsActive)
	}

	// Test saving another wallet type for the same user
	galaWalletSession := &WalletSession{
		UserID:     123456,
		WalletType: "gala",
		Address:    "gala123address",
		IsActive:   true,
	}

	err = db.SaveWalletSession(ctx, galaWalletSession)
	if err != nil {
		t.Fatalf("SaveWalletSession (gala) failed: %v", err)
	}

	// Verify both wallet sessions exist
	tonWallet, err := db.GetWalletSession(ctx, 123456, "ton")
	if err != nil {
		t.Fatalf("GetWalletSession (ton) failed: %v", err)
	}
	if tonWallet.WalletType != "ton" {
		t.Errorf("Expected ton wallet, got %s", tonWallet.WalletType)
	}

	galaWallet, err := db.GetWalletSession(ctx, 123456, "gala")
	if err != nil {
		t.Fatalf("GetWalletSession (gala) failed: %v", err)
	}
	if galaWallet.WalletType != "gala" {
		t.Errorf("Expected gala wallet, got %s", galaWallet.WalletType)
	}

	// Test retrieving non-existent wallet session
	_, err = db.GetWalletSession(ctx, 999999, "ton")
	if err == nil {
		t.Errorf("Expected error when getting non-existent wallet session, got nil")
	}
}

func TestDatabaseClose(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Test closing the database
	err = db.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify operations fail after close
	ctx := context.Background()
	userSession := &UserSession{
		UserID:   123456,
		ChatID:   654321,
		Username: "testuser",
	}

	err = db.SaveUserSession(ctx, userSession)
	if err == nil {
		t.Errorf("Expected error when using closed database, got nil")
	}
}

func TestTimestamps(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Save a user session
	userSession := &UserSession{
		UserID:   123456,
		ChatID:   654321,
		Username: "testuser",
	}

	beforeSave := time.Now()
	err = db.SaveUserSession(ctx, userSession)
	if err != nil {
		t.Fatalf("SaveUserSession failed: %v", err)
	}
	afterSave := time.Now()

	// Verify timestamps were set
	if userSession.CreatedAt.IsZero() {
		t.Errorf("CreatedAt was not set")
	}
	if userSession.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt was not set")
	}

	// Verify timestamps are reasonable
	if userSession.CreatedAt.Before(beforeSave) || userSession.CreatedAt.After(afterSave) {
		t.Errorf("CreatedAt timestamp is not within expected range")
	}

	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)
	originalCreatedAt := userSession.CreatedAt

	userSession.Username = "updateduser"
	beforeUpdate := time.Now()
	err = db.SaveUserSession(ctx, userSession)
	if err != nil {
		t.Fatalf("SaveUserSession (update) failed: %v", err)
	}
	afterUpdate := time.Now()

	// Verify CreatedAt didn't change but UpdatedAt did
	retrieved, err := db.GetUserSession(ctx, 123456)
	if err != nil {
		t.Fatalf("GetUserSession failed: %v", err)
	}

	if !retrieved.CreatedAt.Equal(originalCreatedAt) {
		t.Errorf("CreatedAt changed on update: original=%v, retrieved=%v", originalCreatedAt, retrieved.CreatedAt)
	}

	if retrieved.UpdatedAt.Before(beforeUpdate) || retrieved.UpdatedAt.After(afterUpdate) {
		t.Errorf("UpdatedAt timestamp is not within expected range")
	}
}
