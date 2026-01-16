package ton

import (
	"context"
	"fmt"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

// Client implements blockchain.Client for TON blockchain
type Client struct {
	connection *liteclient.ConnectionPool
	api        *ton.APIClient
}

// NewClient creates a new TON blockchain client
func NewClient() *Client {
	return &Client{}
}

// Connect initializes the connection to TON blockchain
// Connects to TON mainnet by default
func (c *Client) Connect(ctx context.Context) error {
	// Create connection pool for TON blockchain
	client := liteclient.NewConnectionPool()

	// Connect to mainnet (using global config)
	configUrl := "https://ton.org/global-config.json"
	err := client.AddConnectionsFromConfigUrl(ctx, configUrl)
	if err != nil {
		return fmt.Errorf("failed to add connections from config: %w", err)
	}

	// Initialize API client
	c.connection = client
	c.api = ton.NewAPIClient(client)

	return nil
}

// GetBalance returns the balance for the given address
// POC implementation: returns hardcoded value
func (c *Client) GetBalance(ctx context.Context, address string) (string, error) {
	// POC: Return hardcoded balance as specified in task
	// TODO: Implement actual balance fetching from blockchain
	return "10.0", nil
}

// Close closes the connection to TON blockchain
func (c *Client) Close() error {
	if c.connection != nil {
		c.connection.Stop()
		c.connection = nil
		c.api = nil
	}
	return nil
}
