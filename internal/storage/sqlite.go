package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/001_initial_schema.up.sql
var initialSchemaMigration string

// SQLiteDB implements the Database interface using SQLite
type SQLiteDB struct {
	db *sql.DB
}

// InitDB initializes the SQLite database and runs migrations
func InitDB(dbPath string) (Database, error) {
	// Create directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(1) // SQLite only supports one writer at a time
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &SQLiteDB{db: db}, nil
}

// runMigrations executes the SQL migration files
func runMigrations(db *sql.DB) error {
	// Execute embedded migration
	if _, err := db.Exec(initialSchemaMigration); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// GetUserSession retrieves a user session by user ID
func (s *SQLiteDB) GetUserSession(ctx context.Context, userID int64) (*UserSession, error) {
	query := `
		SELECT user_id, chat_id, username, created_at, updated_at
		FROM user_sessions
		WHERE user_id = ?
	`

	var session UserSession
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
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
func (s *SQLiteDB) SaveUserSession(ctx context.Context, session *UserSession) error {
	query := `
		INSERT INTO user_sessions (user_id, chat_id, username, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			chat_id = excluded.chat_id,
			username = excluded.username,
			updated_at = excluded.updated_at
	`

	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, query,
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
func (s *SQLiteDB) GetWalletSession(ctx context.Context, userID int64, walletType string) (*WalletSession, error) {
	query := `
		SELECT user_id, wallet_type, address, connected_at, updated_at, is_active
		FROM wallet_sessions
		WHERE user_id = ? AND wallet_type = ?
	`

	var session WalletSession
	var isActive int
	err := s.db.QueryRowContext(ctx, query, userID, walletType).Scan(
		&session.UserID,
		&session.WalletType,
		&session.Address,
		&session.ConnectedAt,
		&session.UpdatedAt,
		&isActive,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("wallet session not found for user_id %d and wallet_type %s", userID, walletType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet session: %w", err)
	}

	session.IsActive = isActive == 1

	return &session, nil
}

// SaveWalletSession saves or updates a wallet session
func (s *SQLiteDB) SaveWalletSession(ctx context.Context, session *WalletSession) error {
	query := `
		INSERT INTO wallet_sessions (user_id, wallet_type, address, connected_at, updated_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, wallet_type) DO UPDATE SET
			address = excluded.address,
			updated_at = excluded.updated_at,
			is_active = excluded.is_active
	`

	now := time.Now()
	if session.ConnectedAt.IsZero() {
		session.ConnectedAt = now
	}
	session.UpdatedAt = now

	isActive := 0
	if session.IsActive {
		isActive = 1
	}

	_, err := s.db.ExecContext(ctx, query,
		session.UserID,
		session.WalletType,
		session.Address,
		session.ConnectedAt,
		session.UpdatedAt,
		isActive,
	)

	if err != nil {
		return fmt.Errorf("failed to save wallet session: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
