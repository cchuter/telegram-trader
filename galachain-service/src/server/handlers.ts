import * as grpc from '@grpc/grpc-js';
import * as messages from '../types/galachain_pb';
import { GSwapClient } from '../gswap/client';
import { WalletConnectManager, ManualWalletManager } from '../wallet/walletconnect';

/**
 * GrpcHandlers class implements the business logic for all gRPC service methods.
 * This separates handler logic from server setup for better maintainability.
 */
export class GrpcHandlers {
  private gswapClient: GSwapClient;
  private walletConnectManager: WalletConnectManager;
  private manualWalletManager: ManualWalletManager;

  constructor(gswapApiUrl: string, walletConnectProjectId: string) {
    this.gswapClient = new GSwapClient(gswapApiUrl);
    this.walletConnectManager = new WalletConnectManager(walletConnectProjectId);
    this.manualWalletManager = new ManualWalletManager();
  }
  /**
   * GetBalance handler - returns wallet balances for a user
   * POC implementation: returns hardcoded balances
   */
  public getBalance(
    call: grpc.ServerUnaryCall<messages.BalanceRequest, messages.BalanceResponse>,
    callback: grpc.sendUnaryData<messages.BalanceResponse>
  ): void {
    const userId = call.request.getUserId();
    console.log(`GetBalance called for user: ${userId}`);

    // Create response with hardcoded balances for POC
    const response = new messages.BalanceResponse();

    // Create GALA balance
    const galaBalance = new messages.TokenBalance();
    galaBalance.setToken('GALA');
    galaBalance.setBalance('5000.0');
    galaBalance.setUsdValue('0.0'); // USD value not implemented in POC

    // Create GTON balance
    const gtonBalance = new messages.TokenBalance();
    gtonBalance.setToken('GTON');
    gtonBalance.setBalance('2.5');
    gtonBalance.setUsdValue('0.0'); // USD value not implemented in POC

    // Set balances list
    response.setBalancesList([galaBalance, gtonBalance]);

    callback(null, response);
  }

  /**
   * GetPrice handler - returns price for a trading pair
   * POC implementation: calls GSwapClient.getPrice() and returns mock price "855"
   */
  public async getPrice(
    call: grpc.ServerUnaryCall<messages.GetPriceRequest, messages.PriceResponse>,
    callback: grpc.sendUnaryData<messages.PriceResponse>
  ): Promise<void> {
    const pair = call.request.getPair();
    console.log(`GetPrice called for pair: ${pair}`);

    try {
      // Parse pair format (e.g., "GTON/GALA" -> token0: "GTON", token1: "GALA")
      const [token0, token1] = pair.split('/');
      if (!token0 || !token1) {
        callback(
          {
            code: grpc.status.INVALID_ARGUMENT,
            message: 'Invalid pair format. Expected format: TOKEN0/TOKEN1',
          },
          null
        );
        return;
      }

      // Call GSwapClient to get price data
      const priceData = await this.gswapClient.getPrice(token0, token1);

      // Build and return PriceResponse
      const response = new messages.PriceResponse();
      response.setPair(pair);
      response.setPrice(priceData.price);
      response.setTimestamp(priceData.timestamp);
      response.setBid('0.0'); // Not implemented in POC
      response.setAsk('0.0'); // Not implemented in POC
      response.setVolume24h('0.0'); // Not implemented in POC

      callback(null, response);
    } catch (error) {
      console.error('Error in getPrice handler:', error);
      callback(
        {
          code: grpc.status.INTERNAL,
          message: error instanceof Error ? error.message : 'Unknown error',
        },
        null
      );
    }
  }

  /**
   * ExecuteSwap handler - executes a token swap
   * Calls real gswap API to execute swap
   */
  public async executeSwap(
    call: grpc.ServerUnaryCall<messages.SwapRequest, messages.SwapResponse>,
    callback: grpc.sendUnaryData<messages.SwapResponse>
  ): Promise<void> {
    const userId = call.request.getUserId();
    const fromToken = call.request.getFromToken();
    const toToken = call.request.getToToken();
    const amount = call.request.getAmount();
    const slippageBps = call.request.getSlippageBps();
    const feeTier = call.request.getFeeTier() || 3000; // Default to 0.3% fee tier

    console.log(
      `ExecuteSwap called for user ${userId}: ${amount} ${fromToken} -> ${toToken}, slippage: ${slippageBps}bps, fee tier: ${feeTier}`
    );

    const response = new messages.SwapResponse();

    try {
      // Get wallet session for the user
      const wallet = this.manualWalletManager.getWallet(userId);
      if (!wallet) {
        response.setTxHash('');
        response.setAmountIn('0.0');
        response.setAmountOut('0.0');
        response.setFee('0.0');
        response.setStatus('failed');
        response.setErrorMessage('Wallet not connected for user');
        callback(null, response);
        return;
      }

      // Convert token symbols to GalaChain token class keys
      // Format: "TOKEN|Unit|none|none"
      const tokenIn = `${fromToken}|Unit|none|none`;
      const tokenOut = `${toToken}|Unit|none|none`;

      // Calculate minimum output amount based on slippage
      // For simplicity, we'll use a basic calculation
      // In production, this should be based on a price quote
      const amountOutMinimum = '0'; // Accept any amount (POC - should use slippage calculation)

      // Execute swap via GSwapClient
      const swapResult = await this.gswapClient.executeSwap({
        tokenIn,
        tokenOut,
        amountIn: amount,
        amountOutMinimum,
        feeTier,
        walletAddress: wallet.address,
        privateKey: wallet.privateKey,
      });

      // Build response
      response.setTxHash(swapResult.txHash);
      response.setAmountIn(swapResult.amountIn);
      response.setAmountOut(swapResult.amountOut);
      response.setFee(swapResult.fee);
      response.setStatus(swapResult.status);
      if (swapResult.errorMessage) {
        response.setErrorMessage(swapResult.errorMessage);
      }

      console.log(
        `Swap ${swapResult.status}: tx ${swapResult.txHash}, out: ${swapResult.amountOut}`
      );

      callback(null, response);
    } catch (error) {
      console.error('Error in executeSwap handler:', error);
      response.setTxHash('');
      response.setAmountIn('0.0');
      response.setAmountOut('0.0');
      response.setFee('0.0');
      response.setStatus('failed');
      response.setErrorMessage(
        error instanceof Error ? error.message : 'Unknown error'
      );
      callback(null, response);
    }
  }

  /**
   * WatchPrices handler - streams price updates
   * TODO: Implement actual price watching logic with streaming
   */
  public watchPrices(
    call: grpc.ServerWritableStream<messages.WatchPricesRequest, messages.PriceUpdate>
  ): void {
    console.log(`WatchPrices called for pairs: ${call.request.getPairsList()}`);

    // For now, just end the stream
    call.end();
  }

  /**
   * CreateWalletSession handler - creates wallet connection session
   * Supports both WalletConnect and manual private key input
   */
  public async createWalletSession(
    call: grpc.ServerUnaryCall<messages.WalletSessionRequest, messages.WalletSessionResponse>,
    callback: grpc.sendUnaryData<messages.WalletSessionResponse>
  ): Promise<void> {
    const userId = call.request.getUserId();
    const method = call.request.getMethod();

    console.log(`CreateWalletSession called for user: ${userId}, method: ${method}`);

    const response = new messages.WalletSessionResponse();
    response.setMethod(method);

    try {
      if (method === 'walletconnect') {
        // WalletConnect method
        const session = await this.walletConnectManager.createSession(userId);

        response.setSessionId(session.sessionId);
        response.setQrCodeUri(session.uri);
        response.setDeepLink(session.deepLink);
        response.setSuccess(true);

        console.log(`WalletConnect session created for user ${userId}`);
      } else if (method === 'manual') {
        // Manual private key method
        const privateKey = call.request.getPrivateKey();
        const publicKey = call.request.getPublicKey();
        const address = call.request.getAddress();

        if (!privateKey || !publicKey || !address) {
          response.setSuccess(false);
          response.setErrorMessage('Private key, public key, and address are required for manual method');
          callback(null, response);
          return;
        }

        // Store the manual wallet (in production, encrypt the private key!)
        const wallet = this.manualWalletManager.storeWallet(
          userId,
          privateKey,
          publicKey,
          address
        );

        response.setSessionId(`manual_${userId}`);
        response.setAddress(wallet.address);
        response.setSuccess(true);

        console.log(`Manual wallet stored for user ${userId}`);
      } else {
        response.setSuccess(false);
        response.setErrorMessage(`Unknown method: ${method}. Use "walletconnect" or "manual"`);
        callback(null, response);
        return;
      }

      callback(null, response);
    } catch (error) {
      console.error('Error in createWalletSession handler:', error);
      response.setSuccess(false);
      response.setErrorMessage(error instanceof Error ? error.message : 'Unknown error');
      callback(null, response);
    }
  }

  /**
   * HealthCheck handler - returns service health status
   */
  public async healthCheck(
    _call: grpc.ServerUnaryCall<messages.HealthCheckRequest, messages.HealthCheckResponse>,
    callback: grpc.sendUnaryData<messages.HealthCheckResponse>,
    startTime: Date
  ): Promise<void> {
    console.log('HealthCheck called');

    const response = new messages.HealthCheckResponse();

    try {
      // Check gswap API connectivity
      const gswapHealthy = await this.checkGswapAPIHealth();

      // Set overall status based on dependencies
      if (gswapHealthy) {
        response.setStatus('healthy');
      } else {
        response.setStatus('degraded');
      }

      // Set dependencies status
      const dependencies = response.getDependenciesMap();
      dependencies.set('gswap_api', gswapHealthy ? 'connected' : 'disconnected');
      dependencies.set('database', 'not_configured'); // POC - no database yet

      // Calculate uptime
      const uptimeSeconds = Math.floor((Date.now() - startTime.getTime()) / 1000);
      response.setUptimeSeconds(uptimeSeconds);

      callback(null, response);
    } catch (error) {
      console.error('Health check error:', error);
      response.setStatus('unhealthy');
      const dependencies = response.getDependenciesMap();
      dependencies.set('gswap_api', 'error');
      dependencies.set('database', 'not_configured');
      response.setUptimeSeconds(0);
      callback(null, response);
    }
  }

  /**
   * Check gswap API health
   */
  private async checkGswapAPIHealth(): Promise<boolean> {
    try {
      // Try to fetch a price to test connectivity
      await this.gswapClient.getPrice('GTON', 'GALA');
      return true;
    } catch (error) {
      console.error('GSwap API health check failed:', error);
      return false;
    }
  }
}
