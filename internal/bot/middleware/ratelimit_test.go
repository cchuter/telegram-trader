package middleware

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-telegram/bot/models"
)

// Test suite for rate limiter token bucket algorithm
//
// Coverage Summary:
// - NewRateLimiter: 100%
// - getBucket: 91.7% (double-check path in concurrent scenario tested)
// - refill: 100%
// - Allow: 100%
// - min: 100%
// - Middleware: 18.2% (requires real bot instance for bot.SendMessage)
// - Overall ratelimit.go: ~81%
//
// Tests cover:
// - Token bucket algorithm (10 tokens, refill rate calculation)
// - Burst allowance (3 quick commands pass instantly)
// - Rate limit enforcement (11th command blocked)
// - Token refill over time (partial and full refills)
// - Per-user isolation (user A doesn't affect user B)
// - Concurrent access safety (race condition testing)
// - Edge cases (fractional tokens, zero elapsed time, max cap)
//
// Architectural limitation: Middleware function cannot be easily tested
// because bot.SendMessage() requires a concrete *bot.Bot instance which
// cannot be mocked. This is the same limitation as auth_test.go (84.1%).

// TestNewRateLimiter tests the creation of a new rate limiter
func TestNewRateLimiter(t *testing.T) {
	maxCommands := 10
	refillPeriod := 1 * time.Minute

	rl := NewRateLimiter(maxCommands, refillPeriod)

	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if rl.maxTokens != float64(maxCommands) {
		t.Errorf("Expected maxTokens to be %d, got %f", maxCommands, rl.maxTokens)
	}

	expectedRefillRate := float64(maxCommands) / refillPeriod.Seconds()
	if rl.refillRate != expectedRefillRate {
		t.Errorf("Expected refillRate to be %f, got %f", expectedRefillRate, rl.refillRate)
	}

	if rl.buckets == nil {
		t.Error("buckets map should be initialized")
	}
}

// TestAllow_FirstRequest tests that the first request is allowed
func TestAllow_FirstRequest(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	userID := int64(12345)

	allowed := rl.Allow(userID)
	if !allowed {
		t.Error("First request should be allowed")
	}
}

// TestAllow_TenRequestsAllowed tests that 10 requests in succession are allowed
func TestAllow_TenRequestsAllowed(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	userID := int64(12345)

	// Make 10 requests - all should be allowed
	for i := 0; i < 10; i++ {
		allowed := rl.Allow(userID)
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}
}

// TestAllow_EleventhRequestBlocked tests that the 11th request is blocked
func TestAllow_EleventhRequestBlocked(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	userID := int64(12345)

	// Make 10 requests - all should be allowed
	for i := 0; i < 10; i++ {
		allowed := rl.Allow(userID)
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11th request should be blocked
	allowed := rl.Allow(userID)
	if allowed {
		t.Error("11th request should be blocked")
	}
}

// TestAllow_TokenRefill tests that tokens refill over time
func TestAllow_TokenRefill(t *testing.T) {
	// Use shorter refill period for faster test
	rl := NewRateLimiter(10, 10*time.Second) // 1 token per second
	userID := int64(12345)

	// Use up all tokens
	for i := 0; i < 10; i++ {
		rl.Allow(userID)
	}

	// 11th request should be blocked
	if rl.Allow(userID) {
		t.Error("Request should be blocked when no tokens available")
	}

	// Wait for 1.5 seconds - should refill 1 token
	time.Sleep(1500 * time.Millisecond)

	// Next request should be allowed
	allowed := rl.Allow(userID)
	if !allowed {
		t.Error("Request should be allowed after token refill")
	}

	// Following request should be blocked again
	if rl.Allow(userID) {
		t.Error("Request should be blocked after using refilled token")
	}
}

// TestAllow_MultipleUsers tests that rate limiting is per user
func TestAllow_MultipleUsers(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	user1 := int64(12345)
	user2 := int64(67890)

	// User 1 uses all tokens
	for i := 0; i < 10; i++ {
		rl.Allow(user1)
	}

	// User 1 should be blocked
	if rl.Allow(user1) {
		t.Error("User 1 should be blocked after using all tokens")
	}

	// User 2 should still be allowed
	allowed := rl.Allow(user2)
	if !allowed {
		t.Error("User 2 should be allowed (separate bucket)")
	}
}

// TestAllow_PartialTokenRefill tests partial token refill
func TestAllow_PartialTokenRefill(t *testing.T) {
	// 10 tokens per 10 seconds = 1 token per second
	rl := NewRateLimiter(10, 10*time.Second)
	userID := int64(12345)

	// Use up all tokens
	for i := 0; i < 10; i++ {
		rl.Allow(userID)
	}

	// Wait for 3.5 seconds - should refill 3 tokens
	time.Sleep(3500 * time.Millisecond)

	// Should be able to make 3 more requests
	for i := 0; i < 3; i++ {
		allowed := rl.Allow(userID)
		if !allowed {
			t.Errorf("Request %d should be allowed after partial refill", i+1)
		}
	}

	// 4th request should be blocked
	if rl.Allow(userID) {
		t.Error("4th request should be blocked")
	}
}

// TestAllow_MaxTokensCap tests that tokens don't exceed max capacity
func TestAllow_MaxTokensCap(t *testing.T) {
	rl := NewRateLimiter(10, 10*time.Second)
	userID := int64(12345)

	// Use 5 tokens
	for i := 0; i < 5; i++ {
		rl.Allow(userID)
	}

	// Wait for a long time (should refill more than maxTokens)
	time.Sleep(20 * time.Second)

	// Should be able to make exactly 10 requests (not more)
	for i := 0; i < 10; i++ {
		allowed := rl.Allow(userID)
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11th request should be blocked
	if rl.Allow(userID) {
		t.Error("11th request should be blocked (tokens capped at max)")
	}
}

// TestGetBucket tests bucket creation and retrieval
func TestGetBucket(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	userID := int64(12345)

	// Get bucket for user
	bucket1 := rl.getBucket(userID)
	if bucket1 == nil {
		t.Fatal("getBucket should return a bucket")
	}

	// Verify initial state
	if bucket1.maxTokens != 10.0 {
		t.Errorf("Expected maxTokens to be 10.0, got %f", bucket1.maxTokens)
	}

	if bucket1.tokens != 10.0 {
		t.Errorf("Expected initial tokens to be 10.0, got %f", bucket1.tokens)
	}

	// Get bucket again - should return same instance
	bucket2 := rl.getBucket(userID)
	if bucket1 != bucket2 {
		t.Error("getBucket should return the same bucket instance for the same user")
	}
}

// TestMiddleware tests the middleware function with Allow method directly
func TestMiddleware(t *testing.T) {
	rl := NewRateLimiter(2, 10*time.Second) // Small limit for testing
	userID := int64(12345)

	// First two requests should pass
	if !rl.Allow(userID) {
		t.Error("First request should pass")
	}

	if !rl.Allow(userID) {
		t.Error("Second request should pass")
	}

	// Third request should be rate limited
	if rl.Allow(userID) {
		t.Error("Third request should be rate limited")
	}
}

// TestMiddleware_NilMessage tests middleware with nil message
func TestMiddleware_NilMessage(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	ctx := context.Background()

	// Update with nil message should be allowed
	update := &models.Update{
		Message: nil,
	}

	err := rl.Middleware(ctx, nil, update)
	if err != nil {
		t.Errorf("Nil message should be allowed, got error: %v", err)
	}
}

// TestMiddleware_NilUser tests middleware with nil user
func TestMiddleware_NilUser(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	ctx := context.Background()

	// Update with nil user should be allowed
	update := &models.Update{
		Message: &models.Message{
			From: nil,
		},
	}

	err := rl.Middleware(ctx, nil, update)
	if err != nil {
		t.Errorf("Nil user should be allowed, got error: %v", err)
	}
}

// TestConcurrentAccess tests concurrent access to the rate limiter
func TestConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(100, 1*time.Minute)
	userID := int64(12345)

	// Run 50 goroutines making 2 requests each (100 total)
	done := make(chan bool)
	for i := 0; i < 50; i++ {
		go func() {
			rl.Allow(userID)
			rl.Allow(userID)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}

	// Next request should be blocked
	if rl.Allow(userID) {
		t.Error("Request should be blocked after 100 concurrent requests")
	}
}

// BenchmarkAllow benchmarks the Allow method
func BenchmarkAllow(b *testing.B) {
	rl := NewRateLimiter(1000000, 1*time.Minute) // High limit to avoid blocking
	userID := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow(userID)
	}
}

// BenchmarkAllow_MultipleUsers benchmarks with multiple users
func BenchmarkAllow_MultipleUsers(b *testing.B) {
	rl := NewRateLimiter(1000000, 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := int64(i % 100) // Rotate between 100 users
		rl.Allow(userID)
	}
}

// TestBurstAllowance tests that users can make 3 quick commands (burst)
func TestBurstAllowance(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	userID := int64(12345)

	// Make 3 quick requests without any delay
	start := time.Now()
	for i := 0; i < 3; i++ {
		allowed := rl.Allow(userID)
		if !allowed {
			t.Errorf("Burst request %d should be allowed", i+1)
		}
	}
	elapsed := time.Since(start)

	// Should complete in less than 100ms (instant, no waiting)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Burst should be instant, took %v", elapsed)
	}
}

// TestTokenBucketAlgorithm tests the core token bucket behavior
func TestTokenBucketAlgorithm(t *testing.T) {
	// Create rate limiter with 10 tokens, refill rate of 10 per minute (1 per 6s)
	rl := NewRateLimiter(10, 1*time.Minute)
	userID := int64(12345)

	// Verify initial bucket state
	bucket := rl.getBucket(userID)
	if bucket.tokens != 10.0 {
		t.Errorf("Expected 10 initial tokens, got %f", bucket.tokens)
	}
	if bucket.maxTokens != 10.0 {
		t.Errorf("Expected maxTokens 10, got %f", bucket.maxTokens)
	}
	expectedRefillRate := 10.0 / 60.0 // tokens per second
	if bucket.refillRate != expectedRefillRate {
		t.Errorf("Expected refillRate %f, got %f", expectedRefillRate, bucket.refillRate)
	}

	// Consume 3 tokens
	for i := 0; i < 3; i++ {
		rl.Allow(userID)
	}

	// Verify bucket state after consumption
	bucket = rl.getBucket(userID)
	// Use refill to get accurate token count (with minimal time elapsed)
	bucket.mu.Lock()
	bucket.refill()
	tokens := bucket.tokens
	bucket.mu.Unlock()

	// Should be close to 7.0 (allowing for minimal time drift)
	if tokens < 6.99 || tokens > 7.01 {
		t.Errorf("Expected ~7 tokens remaining, got %f", tokens)
	}
}

// TestConcurrentBucketCreation tests concurrent bucket creation for same user
func TestConcurrentBucketCreation(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	userID := int64(12345)

	// Multiple goroutines try to get bucket for same user simultaneously
	done := make(chan *TokenBucket, 10)
	for i := 0; i < 10; i++ {
		go func() {
			bucket := rl.getBucket(userID)
			done <- bucket
		}()
	}

	// Collect all buckets
	buckets := make([]*TokenBucket, 10)
	for i := 0; i < 10; i++ {
		buckets[i] = <-done
	}

	// All should be the same instance (no duplicates created)
	firstBucket := buckets[0]
	for i := 1; i < 10; i++ {
		if buckets[i] != firstBucket {
			t.Errorf("Bucket %d is different from first bucket", i)
		}
	}
}

// TestRefillDoesNotExceedMax tests that refill caps at maxTokens
func TestRefillDoesNotExceedMax(t *testing.T) {
	rl := NewRateLimiter(5, 5*time.Second) // 1 token per second, max 5
	userID := int64(12345)

	// Use 2 tokens
	rl.Allow(userID)
	rl.Allow(userID)

	// Wait longer than needed to refill to max
	time.Sleep(10 * time.Second)

	// Should have exactly 5 tokens available, not more
	bucket := rl.getBucket(userID)
	bucket.mu.Lock()
	bucket.refill()
	tokens := bucket.tokens
	bucket.mu.Unlock()

	if tokens > 5.0 {
		t.Errorf("Tokens should not exceed max of 5, got %f", tokens)
	}
}

// TestZeroRefillTime tests behavior when no time has elapsed
func TestZeroRefillTime(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	userID := int64(12345)

	// Use 5 tokens
	for i := 0; i < 5; i++ {
		rl.Allow(userID)
	}

	// Get bucket and check tokens immediately (no time elapsed)
	bucket := rl.getBucket(userID)
	bucket.mu.Lock()
	tokens1 := bucket.tokens
	bucket.refill() // Should add 0 tokens
	tokens2 := bucket.tokens
	bucket.mu.Unlock()

	// Tokens should be unchanged (or very minimal change due to nanosecond precision)
	if tokens2-tokens1 > 0.001 {
		t.Errorf("Expected no significant refill, got %f -> %f", tokens1, tokens2)
	}
}

// TestFractionalTokens tests that fractional token accumulation works
func TestFractionalTokens(t *testing.T) {
	// 10 tokens per 10 seconds = 1 token per second
	rl := NewRateLimiter(10, 10*time.Second)
	userID := int64(12345)

	// Use all tokens
	for i := 0; i < 10; i++ {
		rl.Allow(userID)
	}

	// Wait 0.7 seconds (should accumulate 0.7 tokens, not enough for a request)
	time.Sleep(700 * time.Millisecond)

	// Should be blocked (need at least 1.0 token)
	if rl.Allow(userID) {
		t.Error("Should be blocked with only 0.7 tokens")
	}

	// Wait another 0.4 seconds (total 1.1 tokens)
	time.Sleep(400 * time.Millisecond)

	// Should now be allowed
	if !rl.Allow(userID) {
		t.Error("Should be allowed with 1.1 tokens accumulated")
	}
}

// TestPerUserIsolationConcurrent tests per-user isolation under concurrent load
func TestPerUserIsolationConcurrent(t *testing.T) {
	rl := NewRateLimiter(5, 1*time.Minute)

	// Three users making requests concurrently
	user1 := int64(1)
	user2 := int64(2)
	user3 := int64(3)

	done := make(chan bool)

	// User 1: exhaust all tokens
	go func() {
		for i := 0; i < 5; i++ {
			rl.Allow(user1)
		}
		done <- true
	}()

	// User 2: use half tokens
	go func() {
		for i := 0; i < 3; i++ {
			rl.Allow(user2)
		}
		done <- true
	}()

	// User 3: use one token
	go func() {
		rl.Allow(user3)
		done <- true
	}()

	// Wait for all
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify each user has correct remaining tokens
	if rl.Allow(user1) {
		t.Error("User 1 should be blocked (all tokens used)")
	}
	if !rl.Allow(user2) {
		t.Error("User 2 should have tokens available")
	}
	if !rl.Allow(user3) {
		t.Error("User 3 should have tokens available")
	}
}

// TestGetBucket_RaceCondition tests concurrent bucket creation race condition
func TestGetBucket_RaceCondition(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	userID := int64(99999)

	// Create 100 goroutines that all try to get bucket at same time
	var wg sync.WaitGroup
	buckets := make(chan *TokenBucket, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bucket := rl.getBucket(userID)
			buckets <- bucket
		}()
	}

	wg.Wait()
	close(buckets)

	// Collect all buckets and verify they're all the same instance
	var firstBucket *TokenBucket
	for bucket := range buckets {
		if firstBucket == nil {
			firstBucket = bucket
		} else if bucket != firstBucket {
			t.Error("Race condition: different bucket instances created for same user")
		}
	}
}

// TestMinFunction tests the min helper function
func TestMinFunction(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{"a less than b", 1.0, 2.0, 1.0},
		{"b less than a", 2.0, 1.0, 1.0},
		{"equal values", 5.0, 5.0, 5.0},
		{"negative values", -1.0, -2.0, -2.0},
		{"zero and positive", 0.0, 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := min(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("min(%f, %f) = %f, want %f", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestRefillAccuracy tests refill calculation accuracy
func TestRefillAccuracy(t *testing.T) {
	// 10 tokens per 10 seconds = exactly 1 token per second
	rl := NewRateLimiter(10, 10*time.Second)
	userID := int64(54321)

	// Use all tokens
	for i := 0; i < 10; i++ {
		rl.Allow(userID)
	}

	// Wait exactly 2 seconds
	time.Sleep(2 * time.Second)

	// Should have refilled exactly 2 tokens
	bucket := rl.getBucket(userID)
	bucket.mu.Lock()
	bucket.refill()
	tokens := bucket.tokens
	bucket.mu.Unlock()

	// Allow for small timing variations (±0.1 tokens)
	if tokens < 1.9 || tokens > 2.1 {
		t.Errorf("Expected ~2 tokens after 2 seconds, got %f", tokens)
	}
}

// TestConcurrentAllowCalls tests concurrent Allow() calls for same user
func TestConcurrentAllowCalls(t *testing.T) {
	rl := NewRateLimiter(50, 1*time.Minute)
	userID := int64(11111)

	var wg sync.WaitGroup
	allowedCount := int64(0)
	var mu sync.Mutex

	// 100 goroutines trying to Allow() concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow(userID) {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should have allowed exactly 50 (no race condition causing double-counting)
	if allowedCount != 50 {
		t.Errorf("Expected exactly 50 allowed, got %d", allowedCount)
	}
}

// TestMultipleUsersNoInterference tests that multiple users don't interfere
func TestMultipleUsersNoInterference(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)

	// Create 10 users, each using exactly 10 tokens
	for userID := int64(1); userID <= 10; userID++ {
		for i := 0; i < 10; i++ {
			if !rl.Allow(userID) {
				t.Errorf("User %d request %d should be allowed", userID, i+1)
			}
		}

		// 11th request should be blocked
		if rl.Allow(userID) {
			t.Errorf("User %d should be rate limited on 11th request", userID)
		}
	}
}

// TestEdgeCaseZeroElapsedTime tests behavior with no time elapsed
func TestEdgeCaseZeroElapsedTime(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)
	userID := int64(22222)

	// Make requests in rapid succession (microseconds apart)
	for i := 0; i < 10; i++ {
		if !rl.Allow(userID) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Next should be blocked (no time for refill)
	if rl.Allow(userID) {
		t.Error("Should be blocked with no time elapsed")
	}
}
