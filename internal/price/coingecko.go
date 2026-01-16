package price

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	// CoinGeckoAPIURL is the base URL for CoinGecko API
	CoinGeckoAPIURL = "https://api.coingecko.com/api/v3"

	// CacheDuration is how long to cache USD prices (5 minutes)
	CacheDuration = 5 * time.Minute
)

// Token ID mappings for CoinGecko API
var tokenIDMap = map[string]string{
	"TON":  "the-open-network",
	"GALA": "gala",
	"GTON": "the-open-network", // GTON is TON on GalaChain
}

// CoinGeckoClient fetches USD prices from CoinGecko API
type CoinGeckoClient struct {
	httpClient *http.Client
	cache      map[string]*cachedPrice
	mu         sync.RWMutex
}

// cachedPrice stores a price with expiration
type cachedPrice struct {
	price     float64
	expiresAt time.Time
}

// NewCoinGeckoClient creates a new CoinGecko client
func NewCoinGeckoClient() *CoinGeckoClient {
	return &CoinGeckoClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]*cachedPrice),
	}
}

// GetUSDPrice fetches the USD price for a token
// Returns the price in USD or error if unavailable
func (c *CoinGeckoClient) GetUSDPrice(ctx context.Context, token string) (float64, error) {
	// Check cache first
	c.mu.RLock()
	cached, exists := c.cache[token]
	c.mu.RUnlock()

	if exists && time.Now().Before(cached.expiresAt) {
		return cached.price, nil
	}

	// Get CoinGecko token ID
	tokenID, exists := tokenIDMap[token]
	if !exists {
		return 0, fmt.Errorf("unsupported token: %s", token)
	}

	// Fetch from CoinGecko API
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", CoinGeckoAPIURL, tokenID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetching price from CoinGecko: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("CoinGecko API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}

	priceData, exists := result[tokenID]
	if !exists {
		return 0, fmt.Errorf("token not found in response: %s", tokenID)
	}

	price, exists := priceData["usd"]
	if !exists {
		return 0, fmt.Errorf("USD price not found for token: %s", token)
	}

	// Cache the price
	c.mu.Lock()
	c.cache[token] = &cachedPrice{
		price:     price,
		expiresAt: time.Now().Add(CacheDuration),
	}
	c.mu.Unlock()

	return price, nil
}

// GetMultipleUSDPrices fetches USD prices for multiple tokens in parallel
// Returns a map of token -> price, ignoring any errors for individual tokens
func (c *CoinGeckoClient) GetMultipleUSDPrices(ctx context.Context, tokens []string) map[string]float64 {
	prices := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, token := range tokens {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			price, err := c.GetUSDPrice(ctx, t)
			if err == nil {
				mu.Lock()
				prices[t] = price
				mu.Unlock()
			}
		}(token)
	}

	wg.Wait()
	return prices
}
