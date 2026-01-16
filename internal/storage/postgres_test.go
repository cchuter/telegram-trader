package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestPostgresIntegration tests PostgreSQL implementation
// Skip in short mode as it requires a real PostgreSQL instance
func TestPostgresIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	// Get PostgreSQL URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" || !isPostgresURL(dbURL) {
		t.Skip("DATABASE_URL not set or not PostgreSQL, skipping integration test")
	}

	// Initialize PostgreSQL database
	db, err := InitPostgresDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test user session operations
	t.Run("UserSession", func(t *testing.T) {
		session := &UserSession{
			UserID:   12345,
			ChatID:   67890,
			Username: "testuser",
		}

		// Save user session
		if err := db.SaveUserSession(ctx, session); err != nil {
			t.Fatalf("Failed to save user session: %v", err)
		}

		// Retrieve user session
		retrieved, err := db.GetUserSession(ctx, session.UserID)
		if err != nil {
			t.Fatalf("Failed to get user session: %v", err)
		}

		if retrieved.UserID != session.UserID {
			t.Errorf("Expected UserID %d, got %d", session.UserID, retrieved.UserID)
		}
		if retrieved.ChatID != session.ChatID {
			t.Errorf("Expected ChatID %d, got %d", session.ChatID, retrieved.ChatID)
		}
		if retrieved.Username != session.Username {
			t.Errorf("Expected Username %s, got %s", session.Username, retrieved.Username)
		}
	})

	// Test wallet session operations
	t.Run("WalletSession", func(t *testing.T) {
		// First create a user session (required for foreign key)
		userSession := &UserSession{
			UserID:   98765,
			ChatID:   11111,
			Username: "walletuser",
		}
		if err := db.SaveUserSession(ctx, userSession); err != nil {
			t.Fatalf("Failed to save user session: %v", err)
		}

		walletSession := &WalletSession{
			UserID:     98765,
			WalletType: "ton",
			Address:    "UQTest123",
			IsActive:   true,
		}

		// Save wallet session
		if err := db.SaveWalletSession(ctx, walletSession); err != nil {
			t.Fatalf("Failed to save wallet session: %v", err)
		}

		// Retrieve wallet session
		retrieved, err := db.GetWalletSession(ctx, walletSession.UserID, walletSession.WalletType)
		if err != nil {
			t.Fatalf("Failed to get wallet session: %v", err)
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
	})

	// Test TonConnect fields
	t.Run("TonConnectFields", func(t *testing.T) {
		// First create a user session
		userSession := &UserSession{
			UserID:   88888,
			ChatID:   22222,
			Username: "tonconnectuser",
		}
		if err := db.SaveUserSession(ctx, userSession); err != nil {
			t.Fatalf("Failed to save user session: %v", err)
		}

		walletSession := &WalletSession{
			UserID:               88888,
			WalletType:           "ton",
			Address:              "UQTonConnect123",
			IsActive:             true,
			TonConnectClientID:   "client123",
			TonConnectPrivateKey: "encryptedprivkey",
			TonConnectWalletID:   "wallet456",
		}

		// Save wallet session with TonConnect fields
		if err := db.SaveWalletSession(ctx, walletSession); err != nil {
			t.Fatalf("Failed to save wallet session: %v", err)
		}

		// Retrieve and verify TonConnect fields
		retrieved, err := db.GetWalletSession(ctx, walletSession.UserID, walletSession.WalletType)
		if err != nil {
			t.Fatalf("Failed to get wallet session: %v", err)
		}

		if retrieved.TonConnectClientID != walletSession.TonConnectClientID {
			t.Errorf("Expected TonConnectClientID %s, got %s", walletSession.TonConnectClientID, retrieved.TonConnectClientID)
		}
		if retrieved.TonConnectPrivateKey != walletSession.TonConnectPrivateKey {
			t.Errorf("Expected TonConnectPrivateKey %s, got %s", walletSession.TonConnectPrivateKey, retrieved.TonConnectPrivateKey)
		}
		if retrieved.TonConnectWalletID != walletSession.TonConnectWalletID {
			t.Errorf("Expected TonConnectWalletID %s, got %s", walletSession.TonConnectWalletID, retrieved.TonConnectWalletID)
		}
	})
}

// TestParseDatabaseURL tests the URL parsing function
func TestParseDatabaseURL(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		expectType string
		shouldFail bool
	}{
		{
			name:       "SQLite file path",
			url:        "telegram-trader.db",
			expectType: "sqlite",
			shouldFail: false,
		},
		{
			name:       "PostgreSQL URL with postgres://",
			url:        "postgres://user:pass@localhost:5432/testdb",
			expectType: "postgres",
			shouldFail: false,
		},
		{
			name:       "PostgreSQL URL with postgresql://",
			url:        "postgresql://user:pass@localhost:5432/testdb",
			expectType: "postgres",
			shouldFail: false,
		},
		{
			name:       "Empty URL",
			url:        "",
			expectType: "",
			shouldFail: true,
		},
		{
			name:       "Absolute SQLite path",
			url:        "/tmp/test.db",
			expectType: "sqlite",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For empty URL test, just verify it returns error
			if tt.url == "" {
				_, err := ParseDatabaseURL(tt.url)
				if err == nil {
					t.Error("Expected error for empty URL, got nil")
				}
				return
			}

			// For valid URLs, we can't fully test without a real database
			// Just verify the function doesn't panic
			if tt.expectType == "postgres" {
				// Skip actual connection test for PostgreSQL URLs
				t.Skip("Skipping PostgreSQL connection test")
			}

			// For SQLite, test with a temporary file
			if tt.expectType == "sqlite" {
				tmpFile := "/tmp/test_" + time.Now().Format("20060102150405") + ".db"
				defer os.Remove(tmpFile)

				db, err := ParseDatabaseURL(tmpFile)
				if err != nil {
					if !tt.shouldFail {
						t.Errorf("Unexpected error: %v", err)
					}
					return
				}
				defer db.Close()

				if tt.shouldFail {
					t.Error("Expected error but got success")
				}
			}
		})
	}
}

// isPostgresURL checks if a URL is a PostgreSQL connection string
func isPostgresURL(url string) bool {
	return len(url) > 0 && (url[:10] == "postgres://" || (len(url) > 13 && url[:13] == "postgresql://"))
}

// TestPostgresMigrations tests that migrations run correctly
func TestPostgresMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL migration test in short mode")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" || !isPostgresURL(dbURL) {
		t.Skip("DATABASE_URL not set or not PostgreSQL, skipping migration test")
	}

	// Initialize database twice to ensure migrations are idempotent
	db1, err := InitPostgresDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to initialize PostgreSQL (first time): %v", err)
	}
	db1.Close()

	db2, err := InitPostgresDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to initialize PostgreSQL (second time): %v", err)
	}
	defer db2.Close()

	// Verify schema_migrations table exists
	postgresDB, ok := db2.(*PostgresDB)
	if !ok {
		t.Fatal("Expected PostgresDB type")
	}

	var count int
	err = postgresDB.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query schema_migrations: %v", err)
	}

	if count < 2 {
		t.Errorf("Expected at least 2 migrations, got %d", count)
	}
}
