package ton

import (
	"context"
	"math/big"
	"testing"
)

// TestClient_GetBalance tests the GetBalance method
// Note: This requires -integration flag and network connectivity
func TestClient_GetBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create and connect client
	client := NewClient()
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to TON blockchain: %v", err)
	}
	defer client.Close()

	// Test with a known TON address (TON Foundation wallet as example)
	// This address should have a balance
	testAddress := "EQCD39VS5jcptHL8vMjEXrzGaRcCVYto7HUn4bpAOg8xqB2N"

	balance, err := client.GetBalance(ctx, testAddress)
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}

	// Balance should be a valid decimal string
	if balance == "" {
		t.Error("Balance should not be empty")
	}

	t.Logf("Balance for address %s: %s TON", testAddress, balance)
}

// TestClient_GetBalance_InvalidAddress tests error handling for invalid addresses
func TestClient_GetBalance_InvalidAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := NewClient()
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to TON blockchain: %v", err)
	}
	defer client.Close()

	// Test with invalid address
	_, err = client.GetBalance(ctx, "invalid-address")
	if err == nil {
		t.Error("Expected error for invalid address, got nil")
	}
}

// TestClient_GetBalance_NotConnected tests that GetBalance fails when not connected
func TestClient_GetBalance_NotConnected(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Try to get balance without connecting
	_, err := client.GetBalance(ctx, "EQCD39VS5jcptHL8vMjEXrzGaRcCVYto7HUn4bpAOg8xqB2N")
	if err == nil {
		t.Error("Expected error when client not connected, got nil")
	}
}

// TestClient_GetJettonBalance tests the GetJettonBalance method
// Note: This requires -integration flag and network connectivity
func TestClient_GetJettonBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := NewClient()
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to TON blockchain: %v", err)
	}
	defer client.Close()

	// GALA token master contract address on TON
	galaTokenAddress := "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV"

	// Test with a wallet address (using TON Foundation as test)
	testWalletAddress := "EQCD39VS5jcptHL8vMjEXrzGaRcCVYto7HUn4bpAOg8xqB2N"

	balance, err := client.GetJettonBalance(ctx, testWalletAddress, galaTokenAddress)
	if err != nil {
		t.Fatalf("Failed to get jetton balance: %v", err)
	}

	// Balance should be a valid decimal string (may be "0.0" if no GALA tokens)
	if balance == "" {
		t.Error("Balance should not be empty")
	}

	t.Logf("GALA balance for address %s: %s GALA", testWalletAddress, balance)
}

// TestFormatBalance tests the formatBalance helper function
func TestFormatBalance(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "whole number",
			input:    10.0,
			expected: "10.0",
		},
		{
			name:     "with decimals",
			input:    10.5,
			expected: "10.5",
		},
		{
			name:     "many decimals",
			input:    10.123456789,
			expected: "10.123456789",
		},
		{
			name:     "trailing zeros",
			input:    10.100000000,
			expected: "10.1",
		},
		{
			name:     "zero",
			input:    0.0,
			expected: "0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance := big.NewFloat(tt.input)
			result := formatBalance(balance)
			if result != tt.expected {
				t.Errorf("formatBalance(%f) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}
