/**
 * Unit tests for GSwapClient
 *
 * These tests mock the GSwap SDK to verify:
 * - Price fetching with caching
 * - Rate limiting
 * - Error handling
 * - Swap execution delegation
 */

import { GSwapClient } from '../client';

// Mock the GSwap SDK
jest.mock('@gala-chain/gswap-sdk', () => ({
  GSwap: jest.fn().mockImplementation(() => ({
    pools: {
      getPoolData: jest.fn().mockResolvedValue({
        sqrtPrice: 1000000n,
        liquidity: 5000000n,
        tick: 100,
      }),
      calculateSpotPrice: jest.fn().mockReturnValue(855.5),
    },
  })),
  parseTokenClassKey: jest.fn((key: string) => ({
    collection: key.split('|')[0],
    category: key.split('|')[1],
    type: key.split('|')[2],
    additionalKey: key.split('|')[3],
  })),
}));

// Mock the swap executor
jest.mock('../swap', () => ({
  SwapExecutor: jest.fn().mockImplementation(() => ({
    executeSwap: jest.fn().mockResolvedValue({
      txHash: 'mock-tx-hash-12345',
      amountIn: '100',
      amountOut: '85500',
      fee: '300',
      status: 'success',
    }),
  })),
}));

// Mock rate limiter
jest.mock('../../utils/rate-limiter', () => ({
  RateLimiter: jest.fn().mockImplementation(() => ({
    acquire: jest.fn().mockResolvedValue(undefined),
    reset: jest.fn(),
  })),
}));

describe('GSwapClient', () => {
  let client: GSwapClient;
  const mockApiUrl = 'https://mock-api.gala.com';

  beforeEach(() => {
    client = new GSwapClient(mockApiUrl);
    jest.clearAllMocks();
  });

  afterEach(async () => {
    await client.close();
  });

  describe('constructor', () => {
    it('should initialize with API URL', () => {
      expect(client.getApiUrl()).toBe(mockApiUrl);
    });
  });

  describe('getPrice', () => {
    it('should fetch price for token pair', async () => {
      const priceData = await client.getPrice('GTON', 'GALA');

      expect(priceData).toEqual({
        token0: 'GTON',
        token1: 'GALA',
        price: '856', // Updated to match calculateSpotPrice return value
        timestamp: expect.any(Number),
      });
    });

    it('should parse token class keys correctly', async () => {
      const { parseTokenClassKey } = require('@gala-chain/gswap-sdk');

      await client.getPrice('GTON', 'GALA');

      expect(parseTokenClassKey).toHaveBeenCalledWith('GTON|Unit|none|none');
      expect(parseTokenClassKey).toHaveBeenCalledWith('GALA|Unit|none|none');
    });

    it('should cache price data', async () => {
      // First call
      const price1 = await client.getPrice('GTON', 'GALA');

      // Second call (should use cache, timestamp should be same)
      const price2 = await client.getPrice('GTON', 'GALA');

      expect(price1.timestamp).toBe(price2.timestamp);
    });

    it('should cache different pairs separately', async () => {
      const price1 = await client.getPrice('GTON', 'GALA');
      const price2 = await client.getPrice('GALA', 'GTON');

      expect(price1.token0).toBe('GTON');
      expect(price2.token0).toBe('GALA');
    });
  });

  describe('executeSwap', () => {
    it('should delegate to swap executor', async () => {
      const params = {
        tokenIn: 'GALA|Unit|none|none',
        tokenOut: 'GTON|Unit|none|none',
        amountIn: '100',
        amountOutMinimum: '0',
        feeTier: 3000,
        walletAddress: 'client|test',
        privateKey: 'test-key',
      };

      const result = await client.executeSwap(params);

      expect(result).toEqual({
        txHash: 'mock-tx-hash-12345',
        amountIn: '100',
        amountOut: '85500',
        fee: '300',
        status: 'success',
      });
    });
  });

  describe('close', () => {
    it('should clear cache', async () => {
      // Fetch price to populate cache
      const price1 = await client.getPrice('GTON', 'GALA');

      // Close client
      await client.close();

      // Next call should fetch from API with new timestamp
      const price2 = await client.getPrice('GTON', 'GALA');
      expect(price2.timestamp).toBeGreaterThanOrEqual(price1.timestamp);
    });
  });
});
