package galachain

import (
	"context"
	"fmt"

	"github.com/cchuter/telegram-trader/internal/galachain/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the gRPC connection to the GalaChain service
type Client struct {
	conn   *grpc.ClientConn
	client pb.GalaChainServiceClient
}

// Connect establishes a connection to the GalaChain gRPC service
func Connect(ctx context.Context, address string) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to galachain service: %w", err)
	}

	client := pb.NewGalaChainServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetClient returns the underlying gRPC client for making service calls
func (c *Client) GetClient() pb.GalaChainServiceClient {
	return c.client
}
