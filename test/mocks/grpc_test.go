package mocks

import (
	"context"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/galachain/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockGRPCClient_GetBalance(t *testing.T) {
	client := NewMockGRPCClient()
	ctx := context.Background()

	userID := int64(123456)

	// Set balance
	balances := []*pb.TokenBalance{
		{Token: "GALA", Balance: "5000.0"},
		{Token: "GTON", Balance: "2.5"},
	}
	client.SetBalance(userID, balances)

	// Get balance
	response, err := client.GetBalance(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, response.Balances, 2)
	assert.Equal(t, "GALA", response.Balances[0].Token)
	assert.Equal(t, "5000.0", response.Balances[0].Balance)
	assert.Equal(t, "GTON", response.Balances[1].Token)
	assert.Equal(t, "2.5", response.Balances[1].Balance)
}

func TestMockGRPCClient_GetBalance_Default(t *testing.T) {
	client := NewMockGRPCClient()
	ctx := context.Background()

	// Get balance without setting it first
	response, err := client.GetBalance(ctx, 999999)
	require.NoError(t, err)

	// Should return default balance
	assert.Len(t, response.Balances, 2)
	assert.Equal(t, "GALA", response.Balances[0].Token)
	assert.Equal(t, "0.0", response.Balances[0].Balance)
}

func TestMockGRPCClient_GetPrice(t *testing.T) {
	client := NewMockGRPCClient()
	ctx := context.Background()

	pair := "GTON/GALA"

	// Set price
	timestamp := time.Now().Unix()
	priceResponse := &pb.PriceResponse{
		Price:     "855.0",
		Timestamp: timestamp,
		Bid:       "854.5",
		Ask:       "855.5",
		Volume_24H: "1000000.0",
	}
	client.SetPrice(pair, priceResponse)

	// Get price
	response, err := client.GetPrice(ctx, pair)
	require.NoError(t, err)
	assert.Equal(t, "855.0", response.Price)
	assert.Equal(t, timestamp, response.Timestamp)
	assert.Equal(t, "854.5", response.Bid)
	assert.Equal(t, "855.5", response.Ask)
	assert.Equal(t, "1000000.0", response.Volume_24H)
}

func TestMockGRPCClient_GetPrice_Default(t *testing.T) {
	client := NewMockGRPCClient()
	ctx := context.Background()

	// Get price without setting it first
	response, err := client.GetPrice(ctx, "UNKNOWN/PAIR")
	require.NoError(t, err)

	// Should return default price
	assert.Equal(t, "0.0", response.Price)
	assert.Equal(t, int64(0), response.Timestamp)
}

func TestMockGRPCClient_Close(t *testing.T) {
	client := NewMockGRPCClient()

	// Close
	err := client.Close()
	require.NoError(t, err)
	assert.Equal(t, 1, client.CloseCalls)
}

func TestMockGRPCClient_ErrorSimulation(t *testing.T) {
	client := NewMockGRPCClient()
	ctx := context.Background()

	// Simulate GetBalance error
	expectedErr := assert.AnError
	client.SetGetBalanceErr(expectedErr)

	_, err := client.GetBalance(ctx, 123456)
	assert.Equal(t, expectedErr, err)

	// Simulate GetPrice error
	client.SetGetPriceErr(expectedErr)

	_, err = client.GetPrice(ctx, "GTON/GALA")
	assert.Equal(t, expectedErr, err)
}

func TestMockGRPCClient_CallTracking(t *testing.T) {
	client := NewMockGRPCClient()
	ctx := context.Background()

	userID1 := int64(123456)
	userID2 := int64(789012)
	pair1 := "GTON/GALA"
	pair2 := "TON/GALA"

	// Make multiple calls
	_, _ = client.GetBalance(ctx, userID1)
	_, _ = client.GetBalance(ctx, userID1)
	_, _ = client.GetBalance(ctx, userID2)

	_, _ = client.GetPrice(ctx, pair1)
	_, _ = client.GetPrice(ctx, pair1)
	_, _ = client.GetPrice(ctx, pair2)

	// Verify call counts
	assert.Equal(t, 2, client.GetBalanceCallCount(userID1))
	assert.Equal(t, 1, client.GetBalanceCallCount(userID2))
	assert.Equal(t, 0, client.GetBalanceCallCount(999999))

	assert.Equal(t, 2, client.GetPriceCallCount(pair1))
	assert.Equal(t, 1, client.GetPriceCallCount(pair2))
	assert.Equal(t, 0, client.GetPriceCallCount("UNKNOWN/PAIR"))
}

func TestMockGRPCClient_Reset(t *testing.T) {
	client := NewMockGRPCClient()
	ctx := context.Background()

	userID := int64(123456)
	pair := "GTON/GALA"

	// Set data and make calls
	client.SetBalance(userID, []*pb.TokenBalance{
		{Token: "GALA", Balance: "5000.0"},
	})
	client.SetPrice(pair, &pb.PriceResponse{
		Price: "855.0",
	})

	_, _ = client.GetBalance(ctx, userID)
	_, _ = client.GetPrice(ctx, pair)

	// Reset
	client.Reset()

	// Verify reset - should return defaults
	balanceResp, err := client.GetBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "0.0", balanceResp.Balances[0].Balance) // Default value

	priceResp, err := client.GetPrice(ctx, pair)
	require.NoError(t, err)
	assert.Equal(t, "0.0", priceResp.Price) // Default value

	// Call counts should start from 1 after reset
	assert.Equal(t, 1, client.GetBalanceCallCount(userID))
	assert.Equal(t, 1, client.GetPriceCallCount(pair))
}
