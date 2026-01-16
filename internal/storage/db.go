package storage

import "context"

// Database defines the interface for storage operations
type Database interface {
	// User session methods
	GetUserSession(ctx context.Context, userID int64) (*UserSession, error)
	SaveUserSession(ctx context.Context, session *UserSession) error

	// Wallet session methods
	GetWalletSession(ctx context.Context, userID int64, walletType string) (*WalletSession, error)
	SaveWalletSession(ctx context.Context, session *WalletSession) error

	// Close closes the database connection
	Close() error
}
