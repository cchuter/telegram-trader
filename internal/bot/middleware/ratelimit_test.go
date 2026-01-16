package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/go-telegram/bot/models"
)

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
