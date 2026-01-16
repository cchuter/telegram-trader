//go:build integration
// +build integration

package wallet

import (
	"context"
	"testing"
	"time"
)

// TestTonConnectSessionCreation tests creating a TonConnect session
func TestTonConnectSessionCreation(t *testing.T) {
	connector := NewTonConnector("", "")

	session, err := connector.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session has required fields
	if session.ClientID == "" {
		t.Error("ClientID should not be empty")
	}

	if len(session.PrivateKey) != 32 {
		t.Errorf("PrivateKey should be 32 bytes, got %d", len(session.PrivateKey))
	}

	if len(session.PublicKey) != 32 {
		t.Errorf("PublicKey should be 32 bytes, got %d", len(session.PublicKey))
	}

	if session.BridgeURL == "" {
		t.Error("BridgeURL should not be empty")
	}

	if session.ManifestURL == "" {
		t.Error("ManifestURL should not be empty")
	}

	if session.Connected {
		t.Error("Session should not be connected initially")
	}
}

// TestGenerateQRCodeURL tests generating a TonConnect QR code URL
func TestGenerateQRCodeURL(t *testing.T) {
	connector := NewTonConnector("", "")

	session, err := connector.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	qrURL, err := connector.GenerateQRCodeURL(session)
	if err != nil {
		t.Fatalf("Failed to generate QR code URL: %v", err)
	}

	// Verify URL format
	if qrURL == "" {
		t.Error("QR URL should not be empty")
	}

	// Should start with tc:// protocol
	if len(qrURL) < 5 || qrURL[:5] != "tc://" {
		t.Errorf("QR URL should start with 'tc://', got: %s", qrURL[:min(20, len(qrURL))])
	}

	// Should contain client ID
	if len(qrURL) > 0 && !contains(qrURL, session.ClientID) {
		t.Error("QR URL should contain client ID")
	}
}

// TestGenerateTonKeeperURL tests generating a TonKeeper connection URL
func TestGenerateTonKeeperURL(t *testing.T) {
	connector := NewTonConnector("", "")

	session, err := connector.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	tonkeeperURL, err := connector.GenerateTonKeeperURL(session)
	if err != nil {
		t.Fatalf("Failed to generate TonKeeper URL: %v", err)
	}

	// Verify URL format
	if tonkeeperURL == "" {
		t.Error("TonKeeper URL should not be empty")
	}

	// Should start with https://
	if len(tonkeeperURL) < 8 || tonkeeperURL[:8] != "https://" {
		t.Errorf("TonKeeper URL should start with 'https://', got: %s", tonkeeperURL[:min(20, len(tonkeeperURL))])
	}

	// Should contain client ID
	if !contains(tonkeeperURL, session.ClientID) {
		t.Error("TonKeeper URL should contain client ID")
	}
}

// TestListenForConnection tests starting the connection listener
// Note: This is a basic test - full integration would require a real wallet connection
func TestListenForConnection(t *testing.T) {
	connector := NewTonConnector("", "")

	session, err := connector.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = connector.ListenForConnection(ctx, session, func(address string, err error) {
		if err != nil {
			t.Logf("Connection callback received error: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	// Wait a bit to see if connection starts
	time.Sleep(2 * time.Second)

	// Stop listening
	connector.StopListening(session)

	// Note: In a real integration test with a wallet, we would verify:
	// - Bridge connection is established
	// - Callback is called on wallet approval
	// - Session is marked as connected
	// - Wallet address is extracted correctly
}

// TestSessionStateMethods tests session state getter methods
func TestSessionStateMethods(t *testing.T) {
	connector := NewTonConnector("", "")

	session, err := connector.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test IsConnected (should be false initially)
	if session.IsConnected() {
		t.Error("Session should not be connected initially")
	}

	// Test GetWalletAddress (should be empty initially)
	address := session.GetWalletAddress()
	if address != "" {
		t.Errorf("Wallet address should be empty initially, got: %s", address)
	}

	// Simulate connection (for testing purposes)
	session.mu.Lock()
	session.Connected = true
	session.WalletAddress = "EQTest1234567890"
	session.mu.Unlock()

	// Test after simulated connection
	if !session.IsConnected() {
		t.Error("Session should be connected after setting Connected=true")
	}

	address = session.GetWalletAddress()
	if address != "EQTest1234567890" {
		t.Errorf("Expected wallet address 'EQTest1234567890', got: %s", address)
	}
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
