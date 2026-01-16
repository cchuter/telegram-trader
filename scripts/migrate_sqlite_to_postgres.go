package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cchuter/telegram-trader/internal/storage"
)

// migrate_sqlite_to_postgres migrates data from SQLite to PostgreSQL
func main() {
	// Check command line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run scripts/migrate_sqlite_to_postgres.go <sqlite_path> <postgres_url>")
		fmt.Println("\nExample:")
		fmt.Println("  go run scripts/migrate_sqlite_to_postgres.go telegram-trader.db postgres://user:pass@localhost:5432/telegram_trader")
		os.Exit(1)
	}

	sqlitePath := os.Args[1]
	postgresURL := os.Args[2]

	// Open SQLite database
	log.Printf("Opening SQLite database: %s", sqlitePath)
	sqliteDB, err := storage.InitDB(sqlitePath)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer sqliteDB.Close()

	// Open PostgreSQL database
	log.Printf("Opening PostgreSQL database: %s", maskPassword(postgresURL))
	postgresDB, err := storage.InitPostgresDB(postgresURL)
	if err != nil {
		log.Fatalf("Failed to open PostgreSQL database: %v", err)
	}
	defer postgresDB.Close()

	ctx := context.Background()

	// Migrate user sessions
	log.Println("Migrating user sessions...")
	userSessionCount, err := migrateUserSessions(ctx, sqliteDB, postgresDB)
	if err != nil {
		log.Fatalf("Failed to migrate user sessions: %v", err)
	}
	log.Printf("Migrated %d user sessions", userSessionCount)

	// Migrate wallet sessions
	log.Println("Migrating wallet sessions...")
	walletSessionCount, err := migrateWalletSessions(ctx, sqliteDB, postgresDB)
	if err != nil {
		log.Fatalf("Failed to migrate wallet sessions: %v", err)
	}
	log.Printf("Migrated %d wallet sessions", walletSessionCount)

	log.Printf("\nMigration complete!")
	log.Printf("Total user sessions: %d", userSessionCount)
	log.Printf("Total wallet sessions: %d", walletSessionCount)
}

// migrateUserSessions migrates all user sessions from SQLite to PostgreSQL
func migrateUserSessions(ctx context.Context, source, dest storage.Database) (int, error) {
	// Get SQLite database connection to query all users
	sqliteDB, ok := source.(*storage.SQLiteDB)
	if !ok {
		return 0, fmt.Errorf("source is not SQLite database")
	}

	// Query all user sessions
	rows, err := sqliteDB.QueryContext(ctx, "SELECT user_id FROM user_sessions")
	if err != nil {
		return 0, fmt.Errorf("failed to query user sessions: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return count, fmt.Errorf("failed to scan user_id: %w", err)
		}

		// Get full user session from source
		session, err := source.GetUserSession(ctx, userID)
		if err != nil {
			return count, fmt.Errorf("failed to get user session %d: %w", userID, err)
		}

		// Save to destination
		if err := dest.SaveUserSession(ctx, session); err != nil {
			return count, fmt.Errorf("failed to save user session %d: %w", userID, err)
		}

		count++
		log.Printf("  Migrated user session: %d (%s)", userID, session.Username)
	}

	return count, rows.Err()
}

// migrateWalletSessions migrates all wallet sessions from SQLite to PostgreSQL
func migrateWalletSessions(ctx context.Context, source, dest storage.Database) (int, error) {
	// Get SQLite database connection to query all wallets
	sqliteDB, ok := source.(*storage.SQLiteDB)
	if !ok {
		return 0, fmt.Errorf("source is not SQLite database")
	}

	// Query all wallet sessions
	rows, err := sqliteDB.QueryContext(ctx, "SELECT user_id, wallet_type FROM wallet_sessions")
	if err != nil {
		return 0, fmt.Errorf("failed to query wallet sessions: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var userID int64
		var walletType string
		if err := rows.Scan(&userID, &walletType); err != nil {
			return count, fmt.Errorf("failed to scan wallet session: %w", err)
		}

		// Get full wallet session from source
		session, err := source.GetWalletSession(ctx, userID, walletType)
		if err != nil {
			return count, fmt.Errorf("failed to get wallet session %d/%s: %w", userID, walletType, err)
		}

		// Save to destination
		if err := dest.SaveWalletSession(ctx, session); err != nil {
			return count, fmt.Errorf("failed to save wallet session %d/%s: %w", userID, walletType, err)
		}

		count++
		log.Printf("  Migrated wallet session: %d (%s) - %s", userID, walletType, session.Address[:8]+"...")
	}

	return count, rows.Err()
}

// maskPassword masks the password in the PostgreSQL connection string for logging
func maskPassword(url string) string {
	// Simple masking: postgres://user:pass@host:port/db -> postgres://user:***@host:port/db
	// This is a simple implementation, not production-ready
	return url // For now, just return as-is
}
