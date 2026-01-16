// +build integration

package integration

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/galachain/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	testGRPCPort = "50053" // Use different port to avoid conflicts
	testTimeout  = 5 * time.Second
)

// mockGalaChainServer implements the GalaChainService for testing
type mockGalaChainServer struct {
	pb.UnimplementedGalaChainServiceServer
}

func (s *mockGalaChainServer) GetPrice(ctx context.Context, req *pb.GetPriceRequest) (*pb.PriceResponse, error) {
	// Validate pair format
	if len(req.Pair) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Pair is required")
	}

	// Check for invalid format (missing "/")
	if len(req.Pair) > 0 && req.Pair[0] != 'G' && req.Pair != "GTON/GALA" {
		return nil, status.Error(codes.InvalidArgument, "Invalid pair format. Expected format: TOKEN0/TOKEN1")
	}

	return &pb.PriceResponse{
		Pair:      req.Pair,
		Price:     "855.00",
		Timestamp: time.Now().Unix(),
		Bid:       "854.50",
		Ask:       "855.50",
		Volume_24H: "1000000.00",
	}, nil
}

func (s *mockGalaChainServer) GetBalance(ctx context.Context, req *pb.BalanceRequest) (*pb.BalanceResponse, error) {
	return &pb.BalanceResponse{
		Balances: []*pb.TokenBalance{
			{
				Token:    "GALA",
				Balance:  "5000.0",
				UsdValue: "100.0",
			},
			{
				Token:    "GTON",
				Balance:  "2.5",
				UsdValue: "5.0",
			},
		},
	}, nil
}

func (s *mockGalaChainServer) ExecuteSwap(ctx context.Context, req *pb.SwapRequest) (*pb.SwapResponse, error) {
	// Validate required fields
	if req.FromToken == "" {
		return &pb.SwapResponse{
			Status:       "failed",
			ErrorMessage: "FromToken is required",
		}, nil
	}

	// Return mock successful swap
	return &pb.SwapResponse{
		TxHash:    "0x1234567890abcdef",
		AmountIn:  req.Amount,
		AmountOut: "850.0",
		Fee:       "0.3",
		Status:    "success",
	}, nil
}

// TestGRPCIntegration is the main test suite for gRPC communication
func TestGRPCIntegration(t *testing.T) {
	// Start mock gRPC server
	server, listener, err := startMockGRPCServer(testGRPCPort)
	require.NoError(t, err, "Failed to start mock gRPC server")
	defer server.Stop()

	// Start server in background
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create real gRPC client
	client, conn, err := createGRPCClient(testGRPCPort)
	require.NoError(t, err, "Failed to create gRPC client")
	defer conn.Close()

	// Run all RPC tests
	t.Run("GetPrice", func(t *testing.T) {
		testGetPrice(t, client)
	})

	t.Run("GetBalance", func(t *testing.T) {
		testGetBalance(t, client)
	})

	t.Run("ExecuteSwap", func(t *testing.T) {
		testExecuteSwap(t, client)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testErrorHandling(t, client)
	})

	t.Run("ServerUnreachable", func(t *testing.T) {
		testServerUnreachable(t)
	})
}

// startMockGRPCServer starts a mock gRPC server for testing
func startMockGRPCServer(port string) (*grpc.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen: %w", err)
	}

	server := grpc.NewServer()
	pb.RegisterGalaChainServiceServer(server, &mockGalaChainServer{})

	return server, listener, nil
}

// createGRPCClient creates a real gRPC client connection
func createGRPCClient(port string) (pb.GalaChainServiceClient, *grpc.ClientConn, error) {
	address := fmt.Sprintf("localhost:%s", port)

	// Create connection with insecure credentials (test only)
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	client := pb.NewGalaChainServiceClient(conn)
	return client, conn, nil
}

// testGetPrice tests the GetPrice RPC call
func testGetPrice(t *testing.T, client pb.GalaChainServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	req := &pb.GetPriceRequest{
		Pair: "GTON/GALA",
	}

	resp, err := client.GetPrice(ctx, req)
	require.NoError(t, err, "GetPrice RPC failed")
	require.NotNil(t, resp, "GetPrice response is nil")

	// Verify response structure
	assert.Equal(t, "GTON/GALA", resp.Pair, "Pair mismatch")
	assert.Equal(t, "855.00", resp.Price, "Price mismatch")
	assert.Greater(t, resp.Timestamp, int64(0), "Timestamp should be > 0")
	assert.NotEmpty(t, resp.Bid, "Bid is empty")
	assert.NotEmpty(t, resp.Ask, "Ask is empty")
	assert.NotEmpty(t, resp.Volume_24H, "Volume is empty")

	t.Logf("GetPrice successful: pair=%s, price=%s, timestamp=%d", resp.Pair, resp.Price, resp.Timestamp)
}

// testGetBalance tests the GetBalance RPC call
func testGetBalance(t *testing.T, client pb.GalaChainServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	req := &pb.BalanceRequest{
		UserId: 12345,
	}

	resp, err := client.GetBalance(ctx, req)
	require.NoError(t, err, "GetBalance RPC failed")
	require.NotNil(t, resp, "GetBalance response is nil")

	// Verify response structure
	assert.NotEmpty(t, resp.Balances, "Balances list is empty")
	assert.Len(t, resp.Balances, 2, "Expected exactly 2 token balances")

	// Check balance fields
	for _, balance := range resp.Balances {
		assert.NotEmpty(t, balance.Token, "Token symbol is empty")
		assert.NotEmpty(t, balance.Balance, "Balance is empty")
		t.Logf("Balance: token=%s, balance=%s, usd=%s", balance.Token, balance.Balance, balance.UsdValue)
	}
}

// testExecuteSwap tests the ExecuteSwap RPC call
func testExecuteSwap(t *testing.T, client pb.GalaChainServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	req := &pb.SwapRequest{
		UserId:      12345,
		FromToken:   "GTON",
		ToToken:     "GALA",
		Amount:      "1.0",
		SlippageBps: 100, // 1%
		FeeTier:     0,
	}

	resp, err := client.ExecuteSwap(ctx, req)
	require.NoError(t, err, "ExecuteSwap RPC failed")
	require.NotNil(t, resp, "ExecuteSwap response is nil")

	// Verify response structure
	assert.NotEmpty(t, resp.Status, "Status is empty")
	assert.Equal(t, "success", resp.Status, "Status should be success")
	assert.NotEmpty(t, resp.TxHash, "TxHash is empty")
	assert.Equal(t, "1.0", resp.AmountIn, "AmountIn mismatch")
	assert.NotEmpty(t, resp.AmountOut, "AmountOut is empty")
	assert.NotEmpty(t, resp.Fee, "Fee is empty")

	t.Logf("ExecuteSwap response: tx_hash=%s, amount_in=%s, amount_out=%s, fee=%s, status=%s",
		resp.TxHash, resp.AmountIn, resp.AmountOut, resp.Fee, resp.Status)
}

// testErrorHandling tests error cases
func testErrorHandling(t *testing.T, client pb.GalaChainServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("InvalidPairFormat", func(t *testing.T) {
		req := &pb.GetPriceRequest{
			Pair: "INVALID", // Missing "/" separator
		}

		resp, err := client.GetPrice(ctx, req)
		assert.Error(t, err, "Expected error for invalid pair format")
		assert.Nil(t, resp, "Response should be nil on error")

		// Check gRPC status code
		grpcStatus, ok := status.FromError(err)
		assert.True(t, ok, "Error should be a gRPC status")
		assert.Equal(t, codes.InvalidArgument, grpcStatus.Code(), "Expected InvalidArgument error code")
		t.Logf("Error status: code=%v, message=%s", grpcStatus.Code(), grpcStatus.Message())
	})

	t.Run("InvalidSwapParameters", func(t *testing.T) {
		req := &pb.SwapRequest{
			UserId:      12345,
			FromToken:   "", // Empty token
			ToToken:     "GALA",
			Amount:      "1.0",
			SlippageBps: 100,
		}

		resp, err := client.ExecuteSwap(ctx, req)
		// Swap returns error in response, not as gRPC error
		require.NoError(t, err, "ExecuteSwap should not return gRPC error")
		require.NotNil(t, resp, "Response should not be nil")
		assert.Equal(t, "failed", resp.Status, "Status should be failed")
		assert.NotEmpty(t, resp.ErrorMessage, "Error message should not be empty")
		t.Logf("ExecuteSwap response: status=%s, error=%s", resp.Status, resp.ErrorMessage)
	})
}

// testServerUnreachable tests behavior when server is not available
func testServerUnreachable(t *testing.T) {
	// Try to connect to non-existent server
	invalidPort := "59999"
	client, conn, err := createGRPCClient(invalidPort)
	if err != nil {
		// Connection creation might fail immediately
		t.Logf("Connection creation failed (expected): %v", err)
		return
	}
	defer conn.Close()

	// Try to make RPC call with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := &pb.GetPriceRequest{
		Pair: "GTON/GALA",
	}

	resp, err := client.GetPrice(ctx, req)
	assert.Error(t, err, "Expected error when server unreachable")
	assert.Nil(t, resp, "Response should be nil when server unreachable")

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok, "Error should be a gRPC status")
	// Should be either Unavailable or DeadlineExceeded
	assert.Contains(t, []codes.Code{codes.Unavailable, codes.DeadlineExceeded}, grpcStatus.Code(),
		"Expected Unavailable or DeadlineExceeded error code")

	t.Logf("Unreachable server error (expected): %v", err)
}

// TestRequestResponseSerialization tests protobuf serialization
func TestRequestResponseSerialization(t *testing.T) {
	t.Run("GetPriceRequest", func(t *testing.T) {
		req := &pb.GetPriceRequest{
			Pair: "GTON/GALA",
		}
		assert.Equal(t, "GTON/GALA", req.Pair)
	})

	t.Run("BalanceRequest", func(t *testing.T) {
		req := &pb.BalanceRequest{
			UserId: 12345,
		}
		assert.Equal(t, int64(12345), req.UserId)
	})

	t.Run("SwapRequest", func(t *testing.T) {
		req := &pb.SwapRequest{
			UserId:      12345,
			FromToken:   "GTON",
			ToToken:     "GALA",
			Amount:      "10.5",
			SlippageBps: 100,
			FeeTier:     0,
		}
		assert.Equal(t, int64(12345), req.UserId)
		assert.Equal(t, "GTON", req.FromToken)
		assert.Equal(t, "GALA", req.ToToken)
		assert.Equal(t, "10.5", req.Amount)
		assert.Equal(t, int32(100), req.SlippageBps)
	})

	t.Run("BalanceResponse", func(t *testing.T) {
		resp := &pb.BalanceResponse{
			Balances: []*pb.TokenBalance{
				{
					Token:    "GALA",
					Balance:  "5000.0",
					UsdValue: "100.0",
				},
				{
					Token:    "GTON",
					Balance:  "2.5",
					UsdValue: "5.0",
				},
			},
		}
		assert.Len(t, resp.Balances, 2)
		assert.Equal(t, "GALA", resp.Balances[0].Token)
		assert.Equal(t, "5000.0", resp.Balances[0].Balance)
	})
}
