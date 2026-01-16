package price

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetUSDPrice_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("ids") == "the-open-network" {
			response := map[string]map[string]float64{
				"the-open-network": {"usd": 5.25},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client := NewCoinGeckoClient()
	// Use mock server by replacing base URL in the client
	client.httpClient.Transport = &mockTransport{server: server}

	ctx := context.Background()
	price, err := client.GetUSDPrice(ctx, "TON")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if price != 5.25 {
		t.Errorf("Expected price 5.25, got: %f", price)
	}

	// Verify cache works
	cachedPrice, err := client.GetUSDPrice(ctx, "TON")
	if err != nil {
		t.Fatalf("Expected no error on cached fetch, got: %v", err)
	}

	if cachedPrice != 5.25 {
		t.Errorf("Expected cached price 5.25, got: %f", cachedPrice)
	}
}

func TestGetUSDPrice_UnsupportedToken(t *testing.T) {
	client := NewCoinGeckoClient()
	ctx := context.Background()

	_, err := client.GetUSDPrice(ctx, "UNKNOWN")

	if err == nil {
		t.Fatal("Expected error for unsupported token")
	}
}

func TestGetUSDPrice_CacheExpiration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]map[string]float64{
			"the-open-network": {"usd": 6.0},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewCoinGeckoClient()
	client.httpClient.Transport = &mockTransport{server: server}

	// Manually set a cached price that's expired
	client.cache["TON"] = &cachedPrice{
		price:     1.0,
		expiresAt: time.Now().Add(-1 * time.Second),
	}

	ctx := context.Background()
	// This should trigger a fetch and get new price
	price, err := client.GetUSDPrice(ctx, "TON")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should get the new price, not the cached one
	if price != 6.0 {
		t.Errorf("Expected fresh price 6.0, got: %f", price)
	}
}

func TestGetMultipleUSDPrices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenID := r.URL.Query().Get("ids")
		var response map[string]map[string]float64

		switch tokenID {
		case "the-open-network":
			response = map[string]map[string]float64{
				"the-open-network": {"usd": 5.25},
			}
		case "gala":
			response = map[string]map[string]float64{
				"gala": {"usd": 0.045},
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewCoinGeckoClient()
	client.httpClient.Transport = &mockTransport{server: server}

	ctx := context.Background()
	prices := client.GetMultipleUSDPrices(ctx, []string{"TON", "GALA", "GTON"})

	if len(prices) < 1 {
		t.Error("Expected at least one price to be fetched")
	}

	// Check that valid tokens are present (some may fail due to mock limitations)
	if _, exists := prices["TON"]; !exists {
		t.Error("Expected TON price in results")
	}
}

// mockTransport redirects requests to the test server
type mockTransport struct {
	server *httptest.Server
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect all requests to our test server
	req.URL.Scheme = "http"
	req.URL.Host = t.server.URL[7:] // Strip "http://"
	return http.DefaultTransport.RoundTrip(req)
}

// Integration test - skipped in short mode
func TestGetUSDPrice_RealAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := NewCoinGeckoClient()
	ctx := context.Background()

	tokens := []string{"TON", "GALA"}
	for _, token := range tokens {
		price, err := client.GetUSDPrice(ctx, token)
		if err != nil {
			t.Errorf("Failed to fetch price for %s: %v", token, err)
			continue
		}

		if price <= 0 {
			t.Errorf("Expected positive price for %s, got: %f", token, price)
		}

		t.Logf("%s price: $%.4f", token, price)
	}
}
