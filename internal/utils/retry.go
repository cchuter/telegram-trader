package utils

import (
	"context"
	"errors"
	"math/rand"
	"time"

	botErrors "github.com/cchuter/telegram-trader/internal/errors"
)

const (
	// DefaultMaxRetries is the default number of retry attempts
	DefaultMaxRetries = 3

	// DefaultInitialDelay is the initial delay before first retry
	DefaultInitialDelay = 1 * time.Second

	// DefaultMaxDelay is the maximum delay between retries
	DefaultMaxDelay = 10 * time.Second

	// DefaultJitterMax is the maximum jitter to add (random 0-500ms)
	DefaultJitterMax = 500 * time.Millisecond
)

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	JitterMax    time.Duration
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   DefaultMaxRetries,
		InitialDelay: DefaultInitialDelay,
		MaxDelay:     DefaultMaxDelay,
		JitterMax:    DefaultJitterMax,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// RetryWithBackoff executes a function with exponential backoff retry logic
// It retries network and transient errors, but not user/validation errors
// Exponential backoff: 1s, 2s, 4s (capped at 10s)
// Adds random jitter 0-500ms to each delay to prevent thundering herd
func RetryWithBackoff(ctx context.Context, fn RetryableFunc) error {
	return RetryWithBackoffConfig(ctx, fn, DefaultRetryConfig())
}

// RetryWithBackoffConfig executes a function with exponential backoff using custom config
func RetryWithBackoffConfig(ctx context.Context, fn RetryableFunc, config *RetryConfig) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err) {
			return err // Don't retry user errors
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate exponential backoff delay
		delay := calculateDelay(attempt, config.InitialDelay, config.MaxDelay)

		// Add random jitter to prevent thundering herd
		jitter := time.Duration(rand.Int63n(int64(config.JitterMax)))
		totalDelay := delay + jitter

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(totalDelay):
			// Continue to next retry
		}
	}

	return lastErr
}

// calculateDelay calculates the exponential backoff delay
// delay = initialDelay * 2^attempt, capped at maxDelay
func calculateDelay(attempt int, initialDelay, maxDelay time.Duration) time.Duration {
	// Calculate 2^attempt
	multiplier := 1 << uint(attempt) // Bit shift for 2^attempt

	delay := initialDelay * time.Duration(multiplier)

	// Cap at max delay
	if delay > maxDelay {
		return maxDelay
	}

	return delay
}

// isRetryable determines if an error should be retried
// Returns true for network/transient errors, false for user/validation errors
func isRetryable(err error) bool {
	// Unwrap error if it's a BotError
	var botErr *botErrors.BotError
	if errors.As(err, &botErr) {
		return isRetryableErrorCode(botErr.Code)
	}

	// For non-BotError errors, retry by default (e.g., network errors)
	return true
}

// isRetryableErrorCode determines if a BotError code should be retried
func isRetryableErrorCode(code botErrors.ErrorCode) bool {
	// DO NOT retry these user/validation errors:
	nonRetryableCodes := map[botErrors.ErrorCode]bool{
		// Authentication errors - user needs to take action
		botErrors.AUTH_001: false, // Access denied
		botErrors.AUTH_002: false, // Session expired

		// Rate limiting - already handled by middleware
		botErrors.RATE_001: false, // Too many requests

		// Wallet errors - require user action
		botErrors.WALLET_001: false, // Wallet not connected
		botErrors.WALLET_003: false, // Invalid wallet address

		// Balance errors - user needs more funds
		botErrors.BALANCE_001: false, // Insufficient balance

		// Swap validation errors
		botErrors.SWAP_002: false, // Slippage too high
		botErrors.SWAP_005: false, // Minimum balance not met

		// Validation errors - bad input
		botErrors.VAL_001: false, // Invalid command format
		botErrors.VAL_002: false, // Invalid amount
		botErrors.VAL_003: false, // Invalid token symbol

		// Encryption errors - need to reconnect wallet
		botErrors.ENC_001: false, // Failed to decrypt

		// Arbitrage errors - no opportunity (not a failure)
		botErrors.ARB_001: false, // No opportunity found
	}

	// Check if explicitly non-retryable
	if shouldRetry, exists := nonRetryableCodes[code]; exists {
		return shouldRetry
	}

	// RETRY these transient/network errors:
	retryableCodes := map[botErrors.ErrorCode]bool{
		// Network errors - temporary
		botErrors.NET_001: true, // Network error
		botErrors.NET_002: true, // Service timeout

		// Balance fetch - might be temporary
		botErrors.BALANCE_002: true, // Failed to fetch balance

		// Swap errors - might be temporary
		botErrors.SWAP_001: true, // Failed to simulate swap
		botErrors.SWAP_003: true, // Transaction failed
		botErrors.SWAP_004: true, // Transaction timeout

		// Price errors - temporary
		botErrors.PRICE_001: true, // Failed to fetch price
		botErrors.PRICE_002: true, // Price data unavailable

		// Arbitrage execution errors - might be temporary
		botErrors.ARB_002: true, // Arbitrage failed
		botErrors.ARB_003: true, // Arbitrage partially complete

		// GalaChain service errors - might be temporary
		botErrors.GALA_001: true, // GalaChain service unavailable
		botErrors.GALA_002: true, // gswap rate limit

		// Database errors - might be temporary
		botErrors.DB_001: true, // Failed to save
		botErrors.DB_002: true, // Data retrieval failed

		// Wallet timeout - might succeed on retry
		botErrors.WALLET_002: true, // Wallet connection timeout
	}

	// Default to retryable if not in non-retryable list
	if shouldRetry, exists := retryableCodes[code]; exists {
		return shouldRetry
	}

	// Unknown error codes - retry by default
	return true
}
