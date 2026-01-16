package mocks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockBlockchainClient_GetBalance(t *testing.T) {
	client := NewMockBlockchainClient()
	ctx := context.Background()

	address := "UQTest1234567890"

	// Set balance
	client.SetBalance(address, "10.5")

	// Get balance
	balance, err := client.GetBalance(ctx, address)
	require.NoError(t, err)
	assert.Equal(t, "10.5", balance)

	// Non-existent address should return "0.0"
	balance, err = client.GetBalance(ctx, "UQNonExistent")
	require.NoError(t, err)
	assert.Equal(t, "0.0", balance)
}

func TestMockBlockchainClient_Connect(t *testing.T) {
	client := NewMockBlockchainClient()
	ctx := context.Background()

	// Successful connect
	err := client.Connect(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, client.ConnectCalls)

	// Connect again
	err = client.Connect(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, client.ConnectCalls)
}

func TestMockBlockchainClient_Close(t *testing.T) {
	client := NewMockBlockchainClient()

	// Close
	err := client.Close()
	require.NoError(t, err)
	assert.Equal(t, 1, client.CloseCalls)
}

func TestMockBlockchainClient_ErrorSimulation(t *testing.T) {
	client := NewMockBlockchainClient()
	ctx := context.Background()

	address := "UQTest1234567890"

	// Simulate GetBalance error
	expectedErr := assert.AnError
	client.SetBalanceErr(expectedErr)

	_, err := client.GetBalance(ctx, address)
	assert.Equal(t, expectedErr, err)

	// Clear error
	client.SetBalanceErr(nil)
	_, err = client.GetBalance(ctx, address)
	assert.NoError(t, err)
}

func TestMockBlockchainClient_CallTracking(t *testing.T) {
	client := NewMockBlockchainClient()
	ctx := context.Background()

	address1 := "UQTest1234567890"
	address2 := "UQTest9876543210"

	// Make multiple calls
	_, _ = client.GetBalance(ctx, address1)
	_, _ = client.GetBalance(ctx, address1)
	_, _ = client.GetBalance(ctx, address2)

	// Verify call counts
	assert.Equal(t, 2, client.GetCallCount(address1))
	assert.Equal(t, 1, client.GetCallCount(address2))
	assert.Equal(t, 0, client.GetCallCount("UQNonExistent"))
}

func TestMockBlockchainClient_Reset(t *testing.T) {
	client := NewMockBlockchainClient()
	ctx := context.Background()

	address := "UQTest1234567890"

	// Set balance and make calls
	client.SetBalance(address, "10.5")
	_, _ = client.GetBalance(ctx, address)

	// Reset
	client.Reset()

	// Verify reset
	balance, err := client.GetBalance(ctx, address)
	require.NoError(t, err)
	assert.Equal(t, "0.0", balance)                  // Should return default "0.0"
	assert.Equal(t, 1, client.GetCallCount(address)) // Call count should start from 1 after reset
}
