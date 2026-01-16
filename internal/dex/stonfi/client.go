package stonfi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/utils"
)

const (
	// BaseURL is the base URL for ston.fi API
	BaseURL = "https://api.ston.fi"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 10 * time.Second

	// TONAddress is the native TON token address on ston.fi
	// This represents native TON (not a jetton)
	TONAddress = "EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM9c"
)

// Client implements the dex.Client interface for ston.fi
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new ston.fi client
func NewClient() *Client {
	return &Client{
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// NewClientWithURL creates a new ston.fi client with a custom base URL
// Useful for testing with mock servers
func NewClientWithURL(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// SimulateSwap simulates a token swap on ston.fi
// fromToken: "TON" or token address
// toToken: token address (e.g., EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV for GALA)
// amount: input amount in nanotons or smallest token unit
// Uses retry logic with exponential backoff for network errors
func (c *Client) SimulateSwap(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
	// Normalize "TON" to the native TON address
	if fromToken == "TON" {
		fromToken = TONAddress
	}
	if toToken == "TON" {
		toToken = TONAddress
	}

	var result *dex.SwapSimulation
	err := utils.RetryWithBackoff(ctx, func(ctx context.Context) error {
		// Build URL with query parameters
		apiURL := fmt.Sprintf("%s/v1/swap/simulate", c.baseURL)
		params := url.Values{}
		params.Add("offer_address", fromToken)
		params.Add("ask_address", toToken)
		params.Add("units", amount)
		params.Add("slippage_tolerance", "0.01") // 1% default slippage tolerance

		fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

		httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("failed to execute request: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		var simResp SwapSimulateResponse
		if err := json.Unmarshal(body, &simResp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		result = &dex.SwapSimulation{
			OutputAmount: simResp.AskUnits,
			Fee:          simResp.FeeUnits,
			Slippage:     simResp.SlippagePercentage(),
			PriceImpact:  simResp.PriceImpactPercentage(),
		}
		return nil
	})

	return result, err
}

// Close closes the HTTP client
func (c *Client) Close() error {
	// HTTP client doesn't require explicit cleanup
	// but we implement this to satisfy the dex.Client interface
	return nil
}
