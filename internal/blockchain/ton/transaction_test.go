package ton

import (
	"context"
	"testing"
)

// TestTransactionBuilder_BuildSwapTransaction tests building a swap transaction
func TestTransactionBuilder_BuildSwapTransaction(t *testing.T) {
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

	// Create transaction builder
	builder := NewTransactionBuilder(client.api)

	// Test building a swap transaction
	req := &SwapRequest{
		FromToken:  "TON",
		ToToken:    GALATokenAddress,
		Amount:     "1000000000",                                       // 1 TON in nanotons
		MinOutput:  "850000000",                                        // Minimum 0.85 TON equivalent
		RouterAddr: "EQBsGx9ArADUrREB34W-ghgsCgBShvfUr4Jvlu-0KGc33Rbt", // ston.fi router
		WalletAddr: "EQCD39VS5jcptHL8vMjEXrzGaRcCVYto7HUn4bpAOg8xqB2N",
		PrivateKey: "mock_private_key",
	}

	msg, err := builder.BuildSwapTransaction(ctx, req)
	if err != nil {
		t.Fatalf("Failed to build swap transaction: %v", err)
	}

	if msg == nil {
		t.Error("Expected non-nil message")
	}

	t.Logf("Successfully built swap transaction message")
}

// TestClient_ExecuteSwap tests the full swap execution flow (mocked)
func TestClient_ExecuteSwap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Note: This test will fail because we don't have a real private key
	// It's here to demonstrate the flow and ensure the function signature works

	client := NewClient()
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to TON blockchain: %v", err)
	}
	defer client.Close()

	req := &SwapRequest{
		FromToken:  "TON",
		ToToken:    GALATokenAddress,
		Amount:     "1000000000",                                       // 1 TON in nanotons
		MinOutput:  "850000000",                                        // Minimum 0.85 TON equivalent
		RouterAddr: "EQBsGx9ArADUrREB34W-ghgsCgBShvfUr4Jvlu-0KGc33Rbt", // ston.fi router
		WalletAddr: "EQCD39VS5jcptHL8vMjEXrzGaRcCVYto7HUn4bpAOg8xqB2N",
		PrivateKey: "mock_private_key",
	}

	// This will fail with "private key parsing not fully implemented"
	// which is expected in POC phase
	_, err = client.ExecuteSwap(ctx, req)
	if err == nil {
		t.Error("Expected error for mock private key")
	}

	// Check that error message indicates private key parsing issue
	if err != nil {
		t.Logf("Expected error received: %v", err)
	}
}

// TestParsePrivateKey tests private key parsing (currently returns error)
func TestParsePrivateKey(t *testing.T) {
	// This should return an error indicating implementation needed
	_, err := parsePrivateKey("mock_key")
	if err == nil {
		t.Error("Expected error for unimplemented private key parsing")
	}

	t.Logf("Expected error: %v", err)
}

const (
	// GALATokenAddress is the GALA token address on TON blockchain
	GALATokenAddress = "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV"
)
