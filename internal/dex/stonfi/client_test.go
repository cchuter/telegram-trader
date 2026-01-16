package stonfi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSimulateSwap(t *testing.T) {
	// Mock API response based on real ston.fi API
	mockResponse := SwapSimulateResponse{
		OfferAddress: TONAddress,
		AskAddress:   "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV",
		OfferUnits:   "1000000000",  // 1 TON in nanotons
		AskUnits:     "26185872472", // Actual response from API
		FeeAddress:   "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV",
		FeeUnits:     "78938549", // Actual fee from API
		FeePercent:   "0.003005487",
		MinAskUnits:  "25924013747", // After slippage
		PriceImpact:  "0.000744987",
		SwapRate:     "261.858724720",
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/swap/simulate" {
			t.Errorf("Expected path /v1/swap/simulate, got %s", r.URL.Path)
		}

		// Verify query parameters
		query := r.URL.Query()
		if query.Get("offer_address") != TONAddress {
			t.Errorf("Expected offer_address: %s, got %s", TONAddress, query.Get("offer_address"))
		}
		if query.Get("units") != "1000000000" {
			t.Errorf("Expected units: 1000000000, got %s", query.Get("units"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClientWithURL(server.URL)
	defer client.Close()

	// Test simulation
	ctx := context.Background()
	galaAddress := "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV"
	simulation, err := client.SimulateSwap(ctx, "TON", galaAddress, "1000000000")
	if err != nil {
		t.Fatalf("SimulateSwap failed: %v", err)
	}

	// Verify results
	if simulation.OutputAmount != "26185872472" {
		t.Errorf("Expected OutputAmount: 26185872472, got %s", simulation.OutputAmount)
	}
	if simulation.Fee != "78938549" {
		t.Errorf("Expected Fee: 78938549, got %s", simulation.Fee)
	}
	if simulation.Slippage == 0 {
		t.Errorf("Expected non-zero slippage")
	}
	if simulation.PriceImpact == 0 {
		t.Errorf("Expected non-zero price impact")
	}
}

func TestSimulateSwap_APIError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid token address"}`))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	defer client.Close()

	ctx := context.Background()
	_, err := client.SimulateSwap(ctx, "INVALID", "INVALID", "1000000000")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestSimulateSwap_RealAPI(t *testing.T) {
	// Skip in CI or if API is unavailable
	if testing.Short() {
		t.Skip("Skipping real API test in short mode")
	}

	client := NewClient()
	defer client.Close()

	ctx := context.Background()
	galaAddress := "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV"

	// Test 1 TON -> GALA swap
	simulation, err := client.SimulateSwap(ctx, "TON", galaAddress, "1000000000")
	if err != nil {
		t.Logf("Real API test failed (this may be expected if API is unavailable): %v", err)
		t.Skip("Skipping real API test - API unavailable")
		return
	}

	// Basic validation
	if simulation.OutputAmount == "" {
		t.Error("Expected non-empty OutputAmount")
	}
	if simulation.Fee == "" {
		t.Error("Expected non-empty Fee")
	}

	t.Logf("Swap simulation results:")
	t.Logf("  Output Amount: %s", simulation.OutputAmount)
	t.Logf("  Fee: %s", simulation.Fee)
	t.Logf("  Slippage: %.2f%%", simulation.Slippage)
	t.Logf("  Price Impact: %.2f%%", simulation.PriceImpact)
}

func TestSwapSimulateResponse_SlippagePercentage(t *testing.T) {
	resp := SwapSimulateResponse{
		AskUnits:    "1000",
		MinAskUnits: "995", // 0.5% slippage
	}

	slippage := resp.SlippagePercentage()
	expected := 0.5
	if slippage < expected-0.01 || slippage > expected+0.01 {
		t.Errorf("Expected slippage ~%.2f%%, got %.2f%%", expected, slippage)
	}
}

func TestSwapSimulateResponse_PriceImpactPercentage(t *testing.T) {
	resp := SwapSimulateResponse{
		PriceImpact: "0.005", // 0.5%
	}

	impact := resp.PriceImpactPercentage()
	expected := 0.5
	if impact < expected-0.01 || impact > expected+0.01 {
		t.Errorf("Expected price impact ~%.2f%%, got %.2f%%", expected, impact)
	}
}
