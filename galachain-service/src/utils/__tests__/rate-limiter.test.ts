/**
 * Unit tests for RateLimiter
 *
 * These tests verify:
 * - Request counting and limiting
 * - Window-based rate limiting
 * - Reset functionality
 */

import { RateLimiter } from '../rate-limiter';

describe('RateLimiter', () => {
  describe('basic rate limiting', () => {
    it('should allow requests within limit', async () => {
      const limiter = new RateLimiter({
        maxRequests: 3,
        windowMs: 1000,
      });

      // Should allow 3 requests
      await limiter.acquire();
      await limiter.acquire();
      await limiter.acquire();

      expect(limiter.getCurrentCount()).toBe(3);
    });

    it('should wait when exceeding limit', async () => {
      const limiter = new RateLimiter({
        maxRequests: 2,
        windowMs: 100, // Short window for testing
      });

      const start = Date.now();

      // First 2 should complete quickly
      await limiter.acquire();
      await limiter.acquire();

      // Third should wait for window to expire
      await limiter.acquire();

      const elapsed = Date.now() - start;
      // Should have waited approximately the window duration
      expect(elapsed).toBeGreaterThan(80); // Allow some timing variance
    });

    it('should reset after window expires', async () => {
      const limiter = new RateLimiter({
        maxRequests: 2,
        windowMs: 100, // 100ms window
      });

      // Use up the limit
      await limiter.acquire();
      await limiter.acquire();
      expect(limiter.getCurrentCount()).toBe(2);

      // Wait for window to expire
      await new Promise((resolve) => setTimeout(resolve, 110));

      // Count should be 0 after window expires
      expect(limiter.getCurrentCount()).toBe(0);

      // Should allow requests again immediately
      await limiter.acquire();
      expect(limiter.getCurrentCount()).toBe(1);
    });
  });

  describe('reset', () => {
    it('should clear all request counts', async () => {
      const limiter = new RateLimiter({
        maxRequests: 2,
        windowMs: 1000,
      });

      // Use up the limit
      await limiter.acquire();
      await limiter.acquire();
      expect(limiter.getCurrentCount()).toBe(2);

      // Reset
      limiter.reset();

      // Count should be 0
      expect(limiter.getCurrentCount()).toBe(0);

      // Should allow requests again immediately
      await limiter.acquire();
      await limiter.acquire();
      expect(limiter.getCurrentCount()).toBe(2);
    });
  });

  describe('configuration', () => {
    it('should respect custom max requests', async () => {
      const limiter = new RateLimiter({
        maxRequests: 5,
        windowMs: 1000,
      });

      // Should allow 5 requests
      for (let i = 0; i < 5; i++) {
        await limiter.acquire();
      }

      expect(limiter.getCurrentCount()).toBe(5);
    });

    it('should respect custom window size', async () => {
      const limiter = new RateLimiter({
        maxRequests: 1,
        windowMs: 50, // 50ms window
      });

      await limiter.acquire();
      expect(limiter.getCurrentCount()).toBe(1);

      // Second request should wait ~50ms
      const start = Date.now();
      await limiter.acquire();
      const elapsed = Date.now() - start;

      expect(elapsed).toBeGreaterThan(40); // Allow timing variance
    });
  });

  describe('concurrent requests', () => {
    it('should handle concurrent acquire calls', async () => {
      const limiter = new RateLimiter({
        maxRequests: 3,
        windowMs: 1000,
      });

      // Fire 3 requests concurrently
      const promises = [limiter.acquire(), limiter.acquire(), limiter.acquire()];

      await Promise.all(promises);
      expect(limiter.getCurrentCount()).toBe(3);
    });

    it('should serialize excess concurrent requests', async () => {
      const limiter = new RateLimiter({
        maxRequests: 2,
        windowMs: 100, // Short window
      });

      const start = Date.now();

      // Fire 3 requests concurrently
      const promises = [limiter.acquire(), limiter.acquire(), limiter.acquire()];

      // All should complete, but third should wait
      await Promise.all(promises);

      const elapsed = Date.now() - start;
      // Should have waited for window to expire
      expect(elapsed).toBeGreaterThan(80);
    });
  });

  describe('getCurrentCount', () => {
    it('should return current request count', async () => {
      const limiter = new RateLimiter({
        maxRequests: 5,
        windowMs: 1000,
      });

      expect(limiter.getCurrentCount()).toBe(0);

      await limiter.acquire();
      expect(limiter.getCurrentCount()).toBe(1);

      await limiter.acquire();
      expect(limiter.getCurrentCount()).toBe(2);
    });

    it('should filter expired requests', async () => {
      const limiter = new RateLimiter({
        maxRequests: 5,
        windowMs: 100,
      });

      await limiter.acquire();
      await limiter.acquire();
      expect(limiter.getCurrentCount()).toBe(2);

      // Wait for window to expire
      await new Promise((resolve) => setTimeout(resolve, 110));

      // Expired requests should be filtered out
      expect(limiter.getCurrentCount()).toBe(0);
    });
  });
});
