package galachain

import (
	"context"
	"fmt"

	"github.com/cchuter/telegram-trader/internal/galachain/pb"
	"github.com/cchuter/telegram-trader/internal/utils"
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

// GetBalance retrieves the balance for a user from GalaChain service
// Uses retry logic with exponential backoff for network errors
func (c *Client) GetBalance(ctx context.Context, userID int64) (*pb.BalanceResponse, error) {
	req := &pb.BalanceRequest{
		UserId: userID,
	}

	var result *pb.BalanceResponse
	err := utils.RetryWithBackoff(ctx, func(ctx context.Context) error {
		resp, err := c.client.GetBalance(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to get balance from galachain service: %w", err)
		}
		result = resp
		return nil
	})

	return result, err
}

// GetPrice retrieves the price for a trading pair from GalaChain service
// Uses retry logic with exponential backoff for network errors
func (c *Client) GetPrice(ctx context.Context, pair string) (*pb.PriceResponse, error) {
	req := &pb.GetPriceRequest{
		Pair: pair,
	}

	var result *pb.PriceResponse
	err := utils.RetryWithBackoff(ctx, func(ctx context.Context) error {
		resp, err := c.client.GetPrice(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to get price from galachain service: %w", err)
		}
		result = resp
		return nil
	})

	return result, err
}
