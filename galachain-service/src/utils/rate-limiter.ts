/**
 * RateLimiter - Enforces request rate limits
 *
 * Implements a sliding window rate limiter to prevent
 * exceeding API rate limits.
 */

export interface RateLimiterConfig {
  maxRequests: number;
  windowMs: number;
}

export class RateLimiter {
  private readonly maxRequests: number;
  private readonly windowMs: number;
  private requests: number[] = [];

  constructor(config: RateLimiterConfig) {
    this.maxRequests = config.maxRequests;
    this.windowMs = config.windowMs;
  }

  /**
   * Attempt to acquire a rate limit token
   * @returns Promise that resolves when request is allowed
   */
  async acquire(): Promise<void> {
    const now = Date.now();

    // Remove requests outside the current window
    this.requests = this.requests.filter((timestamp) => now - timestamp < this.windowMs);

    // If we're at the limit, wait until the oldest request expires
    if (this.requests.length >= this.maxRequests) {
      const oldestRequest = this.requests[0];
      const waitTime = this.windowMs - (now - oldestRequest);

      if (waitTime > 0) {
        await new Promise((resolve) => setTimeout(resolve, waitTime));
        // Recursively try again after waiting
        return this.acquire();
      }
    }

    // Record this request
    this.requests.push(Date.now());
  }

  /**
   * Get current request count in the window
   */
  getCurrentCount(): number {
    const now = Date.now();
    this.requests = this.requests.filter((timestamp) => now - timestamp < this.windowMs);
    return this.requests.length;
  }

  /**
   * Reset the rate limiter
   */
  reset(): void {
    this.requests = [];
  }
}
