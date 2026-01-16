package utils

import (
	"context"
	"errors"
	"testing"
	"time"

	botErrors "github.com/cchuter/telegram-trader/internal/errors"
)

func TestRetryWithBackoff_Success(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	fn := func(ctx context.Context) error {
		callCount++
		return nil
	}

	err := RetryWithBackoff(ctx, fn)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestRetryWithBackoff_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	fn := func(ctx context.Context) error {
		callCount++
		if callCount < 3 {
			return botErrors.ErrNetworkError(errors.New("temporary network error"))
		}
		return nil
	}

	err := RetryWithBackoff(ctx, fn)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryWithBackoff_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	fn := func(ctx context.Context) error {
		callCount++
		return botErrors.ErrNetworkError(errors.New("persistent network error"))
	}

	err := RetryWithBackoff(ctx, fn)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Should call 4 times: initial + 3 retries
	if callCount != 4 {
		t.Errorf("expected 4 calls (1 initial + 3 retries), got %d", callCount)
	}
}

func TestRetryWithBackoff_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	fn := func(ctx context.Context) error {
		callCount++
		return botErrors.ErrInsufficientBalance("10", "5", "TON")
	}

	err := RetryWithBackoff(ctx, fn)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Should only call once, no retries for user errors
	if callCount != 1 {
		t.Errorf("expected 1 call (no retries for user error), got %d", callCount)
	}
}

func TestRetryWithBackoff_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0

	fn := func(ctx context.Context) error {
		callCount++
		if callCount == 2 {
			cancel() // Cancel context on second call
		}
		return botErrors.ErrNetworkError(errors.New("network error"))
	}

	err := RetryWithBackoff(ctx, fn)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Should call at least 2 times before context cancellation
	if callCount < 2 {
		t.Errorf("expected at least 2 calls, got %d", callCount)
	}
}

func TestCalculateDelay(t *testing.T) {
	tests := []struct {
		name         string
		attempt      int
		initialDelay time.Duration
		maxDelay     time.Duration
		expected     time.Duration
	}{
		{
			name:         "First retry (attempt 0)",
			attempt:      0,
			initialDelay: 1 * time.Second,
			maxDelay:     10 * time.Second,
			expected:     1 * time.Second, // 1s * 2^0 = 1s
		},
		{
			name:         "Second retry (attempt 1)",
			attempt:      1,
			initialDelay: 1 * time.Second,
			maxDelay:     10 * time.Second,
			expected:     2 * time.Second, // 1s * 2^1 = 2s
		},
		{
			name:         "Third retry (attempt 2)",
			attempt:      2,
			initialDelay: 1 * time.Second,
			maxDelay:     10 * time.Second,
			expected:     4 * time.Second, // 1s * 2^2 = 4s
		},
		{
			name:         "Fourth retry (attempt 3) - capped at max",
			attempt:      3,
			initialDelay: 1 * time.Second,
			maxDelay:     10 * time.Second,
			expected:     8 * time.Second, // 1s * 2^3 = 8s (still under max)
		},
		{
			name:         "Fifth retry (attempt 4) - exceeds max",
			attempt:      4,
			initialDelay: 1 * time.Second,
			maxDelay:     10 * time.Second,
			expected:     10 * time.Second, // 1s * 2^4 = 16s, capped at 10s
		},
		{
			name:         "Large attempt - capped at max",
			attempt:      10,
			initialDelay: 1 * time.Second,
			maxDelay:     10 * time.Second,
			expected:     10 * time.Second, // Would be 1024s, capped at 10s
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := calculateDelay(tt.attempt, tt.initialDelay, tt.maxDelay)
			if delay != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, delay)
			}
		})
	}
}

func TestIsRetryable_BotErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		// Retryable errors
		{
			name:      "Network error - retryable",
			err:       botErrors.ErrNetworkError(errors.New("network error")),
			retryable: true,
		},
		{
			name:      "Service timeout - retryable",
			err:       botErrors.ErrServiceTimeout(errors.New("timeout")),
			retryable: true,
		},
		{
			name:      "Balance fetch failed - retryable",
			err:       botErrors.ErrBalanceFetchFailed(errors.New("api error")),
			retryable: true,
		},
		{
			name:      "Swap simulation failed - retryable",
			err:       botErrors.ErrSwapSimulationFailed(errors.New("api error")),
			retryable: true,
		},
		{
			name:      "Transaction failed - retryable",
			err:       botErrors.ErrTransactionFailed("blockchain error", errors.New("error")),
			retryable: true,
		},
		{
			name:      "Price fetch failed - retryable",
			err:       botErrors.ErrPriceFetchFailed("ston.fi", errors.New("api error")),
			retryable: true,
		},
		{
			name:      "GalaChain unavailable - retryable",
			err:       botErrors.ErrGalaChainUnavailable(errors.New("connection error")),
			retryable: true,
		},
		{
			name:      "Wallet timeout - retryable",
			err:       botErrors.ErrWalletTimeout(errors.New("timeout")),
			retryable: true,
		},

		// Non-retryable errors
		{
			name:      "Access denied - not retryable",
			err:       botErrors.ErrAccessDenied(),
			retryable: false,
		},
		{
			name:      "Insufficient balance - not retryable",
			err:       botErrors.ErrInsufficientBalance("10", "5", "TON"),
			retryable: false,
		},
		{
			name:      "Invalid amount - not retryable",
			err:       botErrors.ErrInvalidAmount(),
			retryable: false,
		},
		{
			name:      "Invalid wallet address - not retryable",
			err:       botErrors.ErrInvalidWalletAddress(),
			retryable: false,
		},
		{
			name:      "Rate limit exceeded - not retryable",
			err:       botErrors.ErrRateLimitExceeded(),
			retryable: false,
		},
		{
			name:      "Slippage too high - not retryable",
			err:       botErrors.ErrSlippageTooHigh("5.0", "1.0"),
			retryable: false,
		},
		{
			name:      "Invalid command format - not retryable",
			err:       botErrors.ErrInvalidCommandFormat(),
			retryable: false,
		},
		{
			name:      "Wallet not connected - not retryable",
			err:       botErrors.ErrWalletNotConnected(),
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryable := isRetryable(tt.err)
			if retryable != tt.retryable {
				t.Errorf("expected retryable=%v, got %v", tt.retryable, retryable)
			}
		})
	}
}

func TestIsRetryable_GenericError(t *testing.T) {
	// Generic errors (non-BotError) should be retryable by default
	err := errors.New("generic error")
	if !isRetryable(err) {
		t.Error("expected generic error to be retryable")
	}
}

func TestRetryWithBackoffConfig_CustomConfig(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	fn := func(ctx context.Context) error {
		callCount++
		return botErrors.ErrNetworkError(errors.New("network error"))
	}

	config := &RetryConfig{
		MaxRetries:   2, // Only 2 retries instead of 3
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		JitterMax:    50 * time.Millisecond,
	}

	err := RetryWithBackoffConfig(ctx, fn, config)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Should call 3 times: initial + 2 retries
	if callCount != 3 {
		t.Errorf("expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}
}

func TestRetryWithBackoff_TimingWithJitter(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	var delays []time.Duration
	lastCall := time.Now()

	fn := func(ctx context.Context) error {
		now := time.Now()
		if callCount > 0 {
			delay := now.Sub(lastCall)
			delays = append(delays, delay)
		}
		lastCall = now
		callCount++
		if callCount < 3 {
			return botErrors.ErrNetworkError(errors.New("network error"))
		}
		return nil
	}

	start := time.Now()
	err := RetryWithBackoff(ctx, fn)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}

	// Check that delays are approximately correct (with jitter)
	// First retry: ~1s + jitter (0-500ms) = 1s-1.5s
	// Second retry: ~2s + jitter (0-500ms) = 2s-2.5s
	// Total: ~3s-4s

	if elapsed < 3*time.Second || elapsed > 5*time.Second {
		t.Errorf("expected total elapsed time between 3-5s, got %v", elapsed)
	}

	// Verify first delay is roughly 1s + jitter
	if len(delays) >= 1 {
		if delays[0] < 1*time.Second || delays[0] > 1500*time.Millisecond {
			t.Errorf("expected first delay between 1s-1.5s, got %v", delays[0])
		}
	}

	// Verify second delay is roughly 2s + jitter
	if len(delays) >= 2 {
		if delays[1] < 2*time.Second || delays[1] > 2500*time.Millisecond {
			t.Errorf("expected second delay between 2s-2.5s, got %v", delays[1])
		}
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", config.MaxRetries)
	}

	if config.InitialDelay != 1*time.Second {
		t.Errorf("expected InitialDelay=1s, got %v", config.InitialDelay)
	}

	if config.MaxDelay != 10*time.Second {
		t.Errorf("expected MaxDelay=10s, got %v", config.MaxDelay)
	}

	if config.JitterMax != 500*time.Millisecond {
		t.Errorf("expected JitterMax=500ms, got %v", config.JitterMax)
	}
}
