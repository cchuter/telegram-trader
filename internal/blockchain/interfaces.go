package blockchain

import "context"

// Client defines the interface for blockchain operations
type Client interface {
	// Connect initializes the blockchain client connection
	Connect(ctx context.Context) error

	// GetBalance returns the wallet balance as a string
	// For POC phase, this may return a hardcoded value
	GetBalance(ctx context.Context, address string) (string, error)

	// Close closes the blockchain client connection
	Close() error
}
