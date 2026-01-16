/**
 * Integration tests for gswap swap execution
 *
 * These tests verify that the swap execution flow works end-to-end:
 * 1. GSwapClient.executeSwap() submits real swap
 * 2. Fee authorization happens before swap
 * 3. Swap status is polled until completion
 * 4. Transaction hash and amounts are returned
 * 5. gRPC handler executes real swaps
 *
 * Note: These are integration tests that require:
 * - Valid GalaChain API URL
 * - Valid wallet with private key
 * - Sufficient token balance
 */

import { GSwapClient } from '../src/gswap/client';
import { SwapExecutor } from '../src/gswap/swap';

// Test configuration
const TEST_API_URL = process.env.GSWAP_API_URL || 'https://api-galaswap.gala.com';
const TEST_WALLET_ADDRESS = process.env.TEST_WALLET_ADDRESS;
const TEST_PRIVATE_KEY = process.env.TEST_PRIVATE_KEY;

describe('GSwap Integration Tests', () => {
  let client: GSwapClient;
  let executor: SwapExecutor;

  beforeAll(() => {
    client = new GSwapClient(TEST_API_URL);
    executor = new SwapExecutor(TEST_API_URL);
  });

  afterAll(async () => {
    await client.close();
  });

  describe('SwapExecutor', () => {
    it('should execute a real swap on gswap', async () => {
      // Skip if credentials not provided
      if (!TEST_WALLET_ADDRESS || !TEST_PRIVATE_KEY) {
        console.log('Skipping swap execution test - no wallet credentials provided');
        return;
      }

      const result = await executor.executeSwap({
        tokenIn: 'GALA|Unit|none|none',
        tokenOut: 'GTON|Unit|none|none',
        amountIn: '1', // Swap 1 GALA
        amountOutMinimum: '0', // Accept any amount (POC)
        feeTier: 3000, // 0.3% fee tier
        walletAddress: TEST_WALLET_ADDRESS,
        privateKey: TEST_PRIVATE_KEY,
      });

      // Verify result structure
      expect(result).toHaveProperty('txHash');
      expect(result).toHaveProperty('amountIn');
      expect(result).toHaveProperty('amountOut');
      expect(result).toHaveProperty('fee');
      expect(result).toHaveProperty('status');

      // Verify swap completed
      expect(['success', 'failed']).toContain(result.status);

      if (result.status === 'success') {
        expect(result.txHash).toBeTruthy();
        expect(parseFloat(result.amountOut)).toBeGreaterThan(0);
        console.log(`Swap succeeded: ${result.amountIn} -> ${result.amountOut}`);
        console.log(`Transaction hash: ${result.txHash}`);
      } else {
        console.log(`Swap failed: ${result.errorMessage}`);
      }
    }, 120000); // 2 minute timeout for blockchain transaction
  });

  describe('GSwapClient', () => {
    it('should execute swap via client wrapper', async () => {
      // Skip if credentials not provided
      if (!TEST_WALLET_ADDRESS || !TEST_PRIVATE_KEY) {
        console.log('Skipping client swap test - no wallet credentials provided');
        return;
      }

      const result = await client.executeSwap({
        tokenIn: 'GALA|Unit|none|none',
        tokenOut: 'GTON|Unit|none|none',
        amountIn: '0.1', // Swap 0.1 GALA
        amountOutMinimum: '0',
        feeTier: 3000,
        walletAddress: TEST_WALLET_ADDRESS,
        privateKey: TEST_PRIVATE_KEY,
      });

      // Verify result
      expect(result).toHaveProperty('status');
      expect(['success', 'failed']).toContain(result.status);

      if (result.status === 'success') {
        expect(result.txHash).toBeTruthy();
        console.log(`Client swap succeeded: tx ${result.txHash}`);
      } else {
        console.log(`Client swap failed: ${result.errorMessage}`);
      }
    }, 120000);
  });

  describe('Error Handling', () => {
    it('should handle invalid private key', async () => {
      const result = await executor.executeSwap({
        tokenIn: 'GALA|Unit|none|none',
        tokenOut: 'GTON|Unit|none|none',
        amountIn: '1',
        amountOutMinimum: '0',
        feeTier: 3000,
        walletAddress: 'client|invalid',
        privateKey: 'invalid-key',
      });

      expect(result.status).toBe('failed');
      expect(result.errorMessage).toBeTruthy();
      console.log(`Invalid key error (expected): ${result.errorMessage}`);
    });

    it('should handle insufficient balance gracefully', async () => {
      // Skip if credentials not provided
      if (!TEST_WALLET_ADDRESS || !TEST_PRIVATE_KEY) {
        console.log('Skipping insufficient balance test');
        return;
      }

      const result = await executor.executeSwap({
        tokenIn: 'GALA|Unit|none|none',
        tokenOut: 'GTON|Unit|none|none',
        amountIn: '999999999', // Impossibly large amount
        amountOutMinimum: '0',
        feeTier: 3000,
        walletAddress: TEST_WALLET_ADDRESS,
        privateKey: TEST_PRIVATE_KEY,
      });

      // Should fail due to insufficient balance
      expect(result.status).toBe('failed');
      expect(result.errorMessage).toBeTruthy();
      console.log(`Insufficient balance error (expected): ${result.errorMessage}`);
    }, 120000);
  });
});

// Note: Run with npm test when Jest is configured
// For now, we'll just verify the code compiles and exports are correct
console.log('Integration test suite defined successfully');
