/**
 * Swap Executor - Handles real swap execution on gswap
 *
 * This module executes token swaps via the gswap API:
 * 1. Authorize fee credit before swap
 * 2. Submit swap via RequestTokenSwap
 * 3. Poll swap status until completion
 */

import { GSwap, PrivateKeySigner, PendingTransaction } from '@gala-chain/gswap-sdk';

export interface SwapParams {
  tokenIn: string;
  tokenOut: string;
  amountIn: string;
  amountOutMinimum: string;
  feeTier: number;
  walletAddress: string;
  privateKey: string;
}

export interface SwapResult {
  txHash: string;
  amountIn: string;
  amountOut: string;
  fee: string;
  status: 'success' | 'failed';
  errorMessage?: string;
}

export class SwapExecutor {
  private readonly apiUrl: string;

  constructor(apiUrl: string) {
    this.apiUrl = apiUrl;
  }

  /**
   * Execute a token swap on gswap
   *
   * Flow:
   * 1. Initialize GSwap SDK with signer
   * 2. Connect to event socket for transaction updates
   * 3. Execute swap via SDK
   * 4. Wait for transaction confirmation
   * 5. Parse result and return
   */
  async executeSwap(params: SwapParams): Promise<SwapResult> {
    try {
      // Initialize GSwap SDK with private key signer
      const signer = new PrivateKeySigner(params.privateKey);
      const gswap = new GSwap({
        signer,
        gatewayBaseUrl: this.apiUrl,
        walletAddress: params.walletAddress,
      });

      // Connect event socket for transaction status updates
      // This is required for the PendingTransaction.wait() to work
      await GSwap.events.connectEventSocket();

      console.log(
        `Executing swap: ${params.amountIn} ${params.tokenIn} -> ${params.tokenOut} (fee tier: ${params.feeTier})`
      );

      // Submit swap transaction
      // The SDK handles both fee authorization and swap submission internally
      const pendingTx: PendingTransaction = await gswap.swaps.swap(
        params.tokenIn,
        params.tokenOut,
        params.feeTier,
        {
          exactIn: params.amountIn,
          amountOutMinimum: params.amountOutMinimum,
        },
        params.walletAddress
      );

      // Check if submission failed immediately
      if (pendingTx.error) {
        console.error(`Swap submission failed: ${pendingTx.message}`);
        return {
          txHash: pendingTx.transactionId,
          amountIn: params.amountIn,
          amountOut: '0',
          fee: '0',
          status: 'failed',
          errorMessage: pendingTx.message,
        };
      }

      console.log(
        `Swap submitted successfully, tx ID: ${pendingTx.transactionId}`
      );

      // Wait for transaction confirmation
      // This polls the event socket until the transaction completes
      const result = await pendingTx.wait();

      console.log(
        `Swap confirmed: tx hash ${result.transactionHash}, tx ID: ${result.txId}`
      );

      // Parse transaction data to extract swap details
      const swapData = result.Data as {
        amountIn?: string;
        amountOut?: string;
        fee?: string;
      };

      // Disconnect event socket to cleanup resources
      GSwap.events.disconnectEventSocket();

      return {
        txHash: result.transactionHash,
        amountIn: swapData.amountIn || params.amountIn,
        amountOut: swapData.amountOut || '0',
        fee: swapData.fee || '0',
        status: 'success',
      };
    } catch (error) {
      // Ensure event socket is disconnected even on error
      try {
        GSwap.events.disconnectEventSocket();
      } catch {
        // Ignore disconnect errors
      }

      console.error('Swap execution failed:', error);

      return {
        txHash: '',
        amountIn: params.amountIn,
        amountOut: '0',
        fee: '0',
        status: 'failed',
        errorMessage:
          error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }
}
