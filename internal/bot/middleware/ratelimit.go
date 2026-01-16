package middleware

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/cchuter/telegram-trader/internal/errors"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// TokenBucket represents a token bucket for a single user
type TokenBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64 // tokens per second
	lastRefillTime time.Time
	mu             sync.Mutex
}

// RateLimiter implements token bucket rate limiting per user
type RateLimiter struct {
	buckets    map[int64]*TokenBucket
	maxTokens  float64
	refillRate float64 // tokens per second
	mu         sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with token bucket algorithm
// maxCommands: maximum number of commands allowed in the time window
// refillPeriod: time period for refilling all tokens (e.g., 1 minute)
func NewRateLimiter(maxCommands int, refillPeriod time.Duration) *RateLimiter {
	// Calculate refill rate: tokens per second
	// For 10 commands per minute: 10 tokens / 60 seconds = 1 token per 6 seconds
	refillRate := float64(maxCommands) / refillPeriod.Seconds()

	return &RateLimiter{
		buckets:    make(map[int64]*TokenBucket),
		maxTokens:  float64(maxCommands),
		refillRate: refillRate,
	}
}

// getBucket retrieves or creates a token bucket for a user
func (rl *RateLimiter) getBucket(userID int64) *TokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[userID]
	rl.mu.RUnlock()

	if exists {
		return bucket
	}

	// Create new bucket with full tokens
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if bucketExisting, existsAgain := rl.buckets[userID]; existsAgain {
		return bucketExisting
	}

	bucket = &TokenBucket{
		tokens:         rl.maxTokens,
		maxTokens:      rl.maxTokens,
		refillRate:     rl.refillRate,
		lastRefillTime: time.Now(),
	}
	rl.buckets[userID] = bucket

	return bucket
}

// refill adds tokens to the bucket based on elapsed time
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()

	// Calculate tokens to add
	tokensToAdd := elapsed * tb.refillRate

	// Add tokens, capped at maxTokens
	tb.tokens = min(tb.tokens+tokensToAdd, tb.maxTokens)
	tb.lastRefillTime = now
}

// Allow checks if a command is allowed (has tokens available)
// Returns true if allowed, false if rate limited
func (rl *RateLimiter) Allow(userID int64) bool {
	bucket := rl.getBucket(userID)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens based on elapsed time
	bucket.refill()

	// Check if we have at least 1 token
	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}

	return false
}

// Middleware wraps the rate limiter as a bot middleware
func (rl *RateLimiter) Middleware(ctx context.Context, b *bot.Bot, update *models.Update) error {
	// Extract user ID from update
	if update.Message == nil || update.Message.From == nil {
		// If we can't extract user ID, allow the request
		return nil
	}

	userID := update.Message.From.ID

	// Check rate limit
	if !rl.Allow(userID) {
		// Rate limit exceeded - send error message
		chatID := update.Message.Chat.ID
		botErr := errors.ErrRateLimitExceeded()

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   botErr.GetUserMessage(),
		})
		if err != nil {
			log.Printf("Error sending rate limit message: %v", err)
		}

		// Return error to stop further processing
		return botErr
	}

	// Allow the request to proceed
	return nil
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
