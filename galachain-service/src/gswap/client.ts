/**
 * GSwapClient - Wrapper for GalaChain GSwap SDK
 *
 * This client provides a simplified interface to the GSwap SDK
 * for querying token prices and swap data on GalaChain.
 */

import { GSwap, parseTokenClassKey } from '@gala-chain/gswap-sdk';
import { RateLimiter } from '../utils/rate-limiter';
import { SwapExecutor, SwapParams, SwapResult } from './swap';

export interface PriceData {
  token0: string;
  token1: string;
  price: string;
  timestamp: number;
}

interface CachedPrice {
  data: PriceData;
  expiresAt: number;
}

export class GSwapClient {
  private readonly apiUrl: string;
  private readonly gswap: GSwap;
  private readonly rateLimiter: RateLimiter;
  private readonly swapExecutor: SwapExecutor;
  private readonly priceCache: Map<string, CachedPrice> = new Map();
  private readonly CACHE_TTL_MS = 5000; // 5 seconds
  private readonly FEE_TIER = 3000; // 0.3% fee tier

  constructor(apiUrl: string) {
    this.apiUrl = apiUrl;

    // Initialize GSwap SDK (read-only mode, no signer needed for price queries)
    this.gswap = new GSwap({
      gatewayBaseUrl: apiUrl,
    });

    // Initialize rate limiter: 20 requests per 10 seconds
    this.rateLimiter = new RateLimiter({
      maxRequests: 20,
      windowMs: 10000, // 10 seconds
    });

    // Initialize swap executor
    this.swapExecutor = new SwapExecutor(apiUrl);
  }

  /**
   * Get the configured API URL
   * Used for debugging and future SDK initialization
   */
  getApiUrl(): string {
    return this.apiUrl;
  }

  /**
   * Get the price for a token pair on GSwap
   *
   * Fetches real-time price from GSwap API with caching and rate limiting.
   * Price is calculated from pool reserves (sqrtPrice).
   *
   * @param token0 - First token symbol (e.g., "GTON")
   * @param token1 - Second token symbol (e.g., "GALA")
   * @returns PriceData with current pool price
   */
  async getPrice(token0: string, token1: string): Promise<PriceData> {
    const cacheKey = `${token0}/${token1}`;

    // Check cache first
    const cached = this.priceCache.get(cacheKey);
    if (cached && Date.now() < cached.expiresAt) {
      return cached.data;
    }

    // Apply rate limiting
    await this.rateLimiter.acquire();

    try {
      // Parse token class keys
      const token0ClassKey = parseTokenClassKey(`${token0}|Unit|none|none`);
      const token1ClassKey = parseTokenClassKey(`${token1}|Unit|none|none`);

      // Fetch pool data from GSwap API
      const poolData = await this.gswap.pools.getPoolData(
        token0ClassKey,
        token1ClassKey,
        this.FEE_TIER
      );

      // Calculate spot price from pool's sqrt price
      const spotPrice = this.gswap.pools.calculateSpotPrice(
        token0ClassKey,
        token1ClassKey,
        poolData.sqrtPrice
      );

      const priceData: PriceData = {
        token0,
        token1,
        price: spotPrice.toFixed(0), // Convert to string, rounded to integer
        timestamp: Date.now(),
      };

      // Cache the result
      this.priceCache.set(cacheKey, {
        data: priceData,
        expiresAt: Date.now() + this.CACHE_TTL_MS,
      });

      return priceData;
    } catch (error) {
      throw new Error(
        `Failed to fetch price for ${token0}/${token1}: ${error}`
      );
    }
  }

  /**
   * Execute a token swap on gswap
   *
   * Submits a real swap transaction to the gswap API.
   * The swap executor handles:
   * - Fee credit authorization
   * - Swap submission via RequestTokenSwap
   * - Status polling until completion
   *
   * @param params - Swap parameters
   * @returns Swap result with tx hash and amounts
   */
  async executeSwap(params: SwapParams): Promise<SwapResult> {
    // Apply rate limiting
    await this.rateLimiter.acquire();

    // Execute swap via swap executor
    return this.swapExecutor.executeSwap(params);
  }

  /**
   * Close any open connections and clear cache
   */
  async close(): Promise<void> {
    this.priceCache.clear();
    this.rateLimiter.reset();
    return Promise.resolve();
  }
}
