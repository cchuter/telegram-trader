package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// PostgresDB implements the Database interface using PostgreSQL
type PostgresDB struct {
	db *sql.DB
}

// InitPostgresDB initializes the PostgreSQL database and runs migrations
func InitPostgresDB(connectionString string) (Database, error) {
	// Open database connection
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runPostgresMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

// runPostgresMigrations executes PostgreSQL migration scripts
func runPostgresMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Migration 1: Initial schema
	if err := applyMigration(db, 1, func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			-- User sessions table
			CREATE TABLE IF NOT EXISTS user_sessions (
				user_id BIGINT PRIMARY KEY,
				chat_id BIGINT NOT NULL,
				username TEXT NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);

			-- Wallet sessions table
			CREATE TABLE IF NOT EXISTS wallet_sessions (
				id SERIAL PRIMARY KEY,
				user_id BIGINT NOT NULL,
				wallet_type TEXT NOT NULL CHECK(wallet_type IN ('ton', 'gala')),
				address TEXT NOT NULL,
				connected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN NOT NULL DEFAULT true,
				UNIQUE(user_id, wallet_type),
				FOREIGN KEY (user_id) REFERENCES user_sessions(user_id) ON DELETE CASCADE
			);

			-- Create indexes for faster lookups
			CREATE INDEX IF NOT EXISTS idx_wallet_sessions_user_id ON wallet_sessions(user_id);
			CREATE INDEX IF NOT EXISTS idx_wallet_sessions_wallet_type ON wallet_sessions(wallet_type);
			CREATE INDEX IF NOT EXISTS idx_wallet_sessions_is_active ON wallet_sessions(is_active);
		`)
		return err
	}); err != nil {
		return err
	}

	// Migration 2: TonConnect fields
	if err := applyMigration(db, 2, func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			-- Add TonConnect fields to wallet_sessions table
			ALTER TABLE wallet_sessions
			ADD COLUMN IF NOT EXISTS tonconnect_client_id TEXT,
			ADD COLUMN IF NOT EXISTS tonconnect_private_key TEXT,
			ADD COLUMN IF NOT EXISTS tonconnect_wallet_id TEXT;
		`)
		return err
	}); err != nil {
		return err
	}

	return nil
}

// applyMigration applies a migration if it hasn't been applied yet
func applyMigration(db *sql.DB, version int, migrationFunc func(*sql.Tx) error) error {
	// Check if migration already applied
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		return nil // Already applied
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Run migration
	if err := migrationFunc(tx); err != nil {
		return fmt.Errorf("migration %d failed: %w", version, err)
	}

	// Record migration
	_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// GetUserSession retrieves a user session by user ID
func (p *PostgresDB) GetUserSession(ctx context.Context, userID int64) (*UserSession, error) {
	query := `
		SELECT user_id, chat_id, username, created_at, updated_at
		FROM user_sessions
		WHERE user_id = $1
	`

	var session UserSession
	err := p.db.QueryRowContext(ctx, query, userID).Scan(
		&session.UserID,
		&session.ChatID,
		&session.Username,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user session not found for user_id %d", userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user session: %w", err)
	}

	return &session, nil
}

// SaveUserSession saves or updates a user session
func (p *PostgresDB) SaveUserSession(ctx context.Context, session *UserSession) error {
	query := `
		INSERT INTO user_sessions (user_id, chat_id, username, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT(user_id) DO UPDATE SET
			chat_id = EXCLUDED.chat_id,
			username = EXCLUDED.username,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	_, err := p.db.ExecContext(ctx, query,
		session.UserID,
		session.ChatID,
		session.Username,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save user session: %w", err)
	}

	return nil
}

// GetWalletSession retrieves a wallet session by user ID and wallet type
func (p *PostgresDB) GetWalletSession(ctx context.Context, userID int64, walletType string) (*WalletSession, error) {
	query := `
		SELECT user_id, wallet_type, address, connected_at, updated_at, is_active,
		       tonconnect_client_id, tonconnect_private_key, tonconnect_wallet_id
		FROM wallet_sessions
		WHERE user_id = $1 AND wallet_type = $2
	`

	var session WalletSession
	var isActive bool
	var clientID, privateKey, walletID sql.NullString
	err := p.db.QueryRowContext(ctx, query, userID, walletType).Scan(
		&session.UserID,
		&session.WalletType,
		&session.Address,
		&session.ConnectedAt,
		&session.UpdatedAt,
		&isActive,
		&clientID,
		&privateKey,
		&walletID,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("wallet session not found for user_id %d and wallet_type %s", userID, walletType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet session: %w", err)
	}

	session.IsActive = isActive
	if clientID.Valid {
		session.TonConnectClientID = clientID.String
	}
	if privateKey.Valid {
		session.TonConnectPrivateKey = privateKey.String
	}
	if walletID.Valid {
		session.TonConnectWalletID = walletID.String
	}

	return &session, nil
}

// SaveWalletSession saves or updates a wallet session
func (p *PostgresDB) SaveWalletSession(ctx context.Context, session *WalletSession) error {
	query := `
		INSERT INTO wallet_sessions (user_id, wallet_type, address, connected_at, updated_at, is_active,
		                             tonconnect_client_id, tonconnect_private_key, tonconnect_wallet_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT(user_id, wallet_type) DO UPDATE SET
			address = EXCLUDED.address,
			updated_at = EXCLUDED.updated_at,
			is_active = EXCLUDED.is_active,
			tonconnect_client_id = EXCLUDED.tonconnect_client_id,
			tonconnect_private_key = EXCLUDED.tonconnect_private_key,
			tonconnect_wallet_id = EXCLUDED.tonconnect_wallet_id
	`

	now := time.Now()
	if session.ConnectedAt.IsZero() {
		session.ConnectedAt = now
	}
	session.UpdatedAt = now

	// Handle NULL values for optional TonConnect fields
	var clientID, privateKey, walletID interface{}
	if session.TonConnectClientID != "" {
		clientID = session.TonConnectClientID
	}
	if session.TonConnectPrivateKey != "" {
		privateKey = session.TonConnectPrivateKey
	}
	if session.TonConnectWalletID != "" {
		walletID = session.TonConnectWalletID
	}

	_, err := p.db.ExecContext(ctx, query,
		session.UserID,
		session.WalletType,
		session.Address,
		session.ConnectedAt,
		session.UpdatedAt,
		session.IsActive,
		clientID,
		privateKey,
		walletID,
	)

	if err != nil {
		return fmt.Errorf("failed to save wallet session: %w", err)
	}

	return nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// QueryContext executes a query that returns rows (exposed for migration script)
func (p *PostgresDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

// ParseDatabaseURL detects whether DATABASE_URL is SQLite or PostgreSQL and returns the appropriate Database instance
func ParseDatabaseURL(databaseURL string) (Database, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is empty")
	}

	// Detect PostgreSQL URLs (postgres://, postgresql://)
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		return InitPostgresDB(databaseURL)
	}

	// Default to SQLite for file paths
	return InitDB(databaseURL)
}
