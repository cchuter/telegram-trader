import * as grpc from '@grpc/grpc-js';
import * as messages from '../types/galachain_pb';

/**
 * GrpcHandlers class implements the business logic for all gRPC service methods.
 * This separates handler logic from server setup for better maintainability.
 */
export class GrpcHandlers {
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
   * TODO: Implement actual price fetching logic
   */
  public getPrice(
    call: grpc.ServerUnaryCall<messages.GetPriceRequest, messages.PriceResponse>,
    callback: grpc.sendUnaryData<messages.PriceResponse>
  ): void {
    console.log(`GetPrice called for pair: ${call.request.getPair()}`);

    const response = new messages.PriceResponse();
    response.setPair(call.request.getPair());
    response.setPrice('0.0');
    response.setTimestamp(Date.now());
    response.setBid('0.0');
    response.setAsk('0.0');
    response.setVolume24h('0.0');

    callback(null, response);
  }

  /**
   * ExecuteSwap handler - executes a token swap
   * TODO: Implement actual swap execution logic
   */
  public executeSwap(
    call: grpc.ServerUnaryCall<messages.SwapRequest, messages.SwapResponse>,
    callback: grpc.sendUnaryData<messages.SwapResponse>
  ): void {
    console.log(`ExecuteSwap called for user: ${call.request.getUserId()}`);

    const response = new messages.SwapResponse();
    response.setTxHash('');
    response.setAmountIn('0.0');
    response.setAmountOut('0.0');
    response.setFee('0.0');
    response.setStatus('pending');
    response.setErrorMessage('');

    callback(null, response);
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
   * TODO: Implement actual wallet session creation logic
   */
  public createWalletSession(
    call: grpc.ServerUnaryCall<messages.WalletSessionRequest, messages.WalletSessionResponse>,
    callback: grpc.sendUnaryData<messages.WalletSessionResponse>
  ): void {
    console.log(`CreateWalletSession called for user: ${call.request.getUserId()}`);

    const response = new messages.WalletSessionResponse();
    response.setSessionId('');
    response.setMethod(call.request.getMethod());
    response.setSuccess(false);
    response.setErrorMessage('Not implemented');

    callback(null, response);
  }

  /**
   * HealthCheck handler - returns service health status
   */
  public healthCheck(
    _call: grpc.ServerUnaryCall<messages.HealthCheckRequest, messages.HealthCheckResponse>,
    callback: grpc.sendUnaryData<messages.HealthCheckResponse>,
    startTime: Date
  ): void {
    console.log('HealthCheck called');

    const response = new messages.HealthCheckResponse();
    response.setStatus('healthy');

    const dependencies = response.getDependenciesMap();
    dependencies.set('galachain', 'connected');
    dependencies.set('database', 'connected');

    const uptimeSeconds = Math.floor((Date.now() - startTime.getTime()) / 1000);
    response.setUptimeSeconds(uptimeSeconds);

    callback(null, response);
  }
}
