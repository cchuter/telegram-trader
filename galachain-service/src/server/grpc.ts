import * as grpc from '@grpc/grpc-js';
import { GalaChainServiceService, IGalaChainServiceServer } from '../types/galachain_grpc_pb';
import * as messages from '../types/galachain_pb';

export class GalaChainGrpcServer {
  private server: grpc.Server;
  private port: number;
  private startTime: Date;

  constructor(port: number = 50051) {
    this.server = new grpc.Server();
    this.port = port;
    this.startTime = new Date();
    this.setupHandlers();
  }

  private setupHandlers(): void {
    const serviceImplementation: IGalaChainServiceServer = {
      getPrice: this.handleGetPrice.bind(this),
      getBalance: this.handleGetBalance.bind(this),
      executeSwap: this.handleExecuteSwap.bind(this),
      watchPrices: this.handleWatchPrices.bind(this),
      createWalletSession: this.handleCreateWalletSession.bind(this),
      healthCheck: this.handleHealthCheck.bind(this),
    };

    this.server.addService(GalaChainServiceService, serviceImplementation);
  }

  // Unary RPC: Get price for a trading pair
  private handleGetPrice(
    call: grpc.ServerUnaryCall<messages.GetPriceRequest, messages.PriceResponse>,
    callback: grpc.sendUnaryData<messages.PriceResponse>
  ): void {
    console.log(`GetPrice called for pair: ${call.request.getPair()}`);

    // TODO: Implement actual price fetching logic
    const response = new messages.PriceResponse();
    response.setPair(call.request.getPair());
    response.setPrice('0.0');
    response.setTimestamp(Date.now());
    response.setBid('0.0');
    response.setAsk('0.0');
    response.setVolume24h('0.0');

    callback(null, response);
  }

  // Unary RPC: Get wallet balance
  private handleGetBalance(
    call: grpc.ServerUnaryCall<messages.BalanceRequest, messages.BalanceResponse>,
    callback: grpc.sendUnaryData<messages.BalanceResponse>
  ): void {
    console.log(`GetBalance called for user: ${call.request.getUserId()}`);

    // TODO: Implement actual balance fetching logic
    const response = new messages.BalanceResponse();
    response.setBalancesList([]);

    callback(null, response);
  }

  // Unary RPC: Execute token swap
  private handleExecuteSwap(
    call: grpc.ServerUnaryCall<messages.SwapRequest, messages.SwapResponse>,
    callback: grpc.sendUnaryData<messages.SwapResponse>
  ): void {
    console.log(`ExecuteSwap called for user: ${call.request.getUserId()}`);

    // TODO: Implement actual swap execution logic
    const response = new messages.SwapResponse();
    response.setTxHash('');
    response.setAmountIn('0.0');
    response.setAmountOut('0.0');
    response.setFee('0.0');
    response.setStatus('pending');
    response.setErrorMessage('');

    callback(null, response);
  }

  // Server streaming RPC: Watch price updates
  private handleWatchPrices(
    call: grpc.ServerWritableStream<messages.WatchPricesRequest, messages.PriceUpdate>
  ): void {
    console.log(`WatchPrices called for pairs: ${call.request.getPairsList()}`);

    // TODO: Implement actual price watching logic with streaming
    // For now, just end the stream
    call.end();
  }

  // Unary RPC: Create wallet connection session
  private handleCreateWalletSession(
    call: grpc.ServerUnaryCall<messages.WalletSessionRequest, messages.WalletSessionResponse>,
    callback: grpc.sendUnaryData<messages.WalletSessionResponse>
  ): void {
    console.log(`CreateWalletSession called for user: ${call.request.getUserId()}`);

    // TODO: Implement actual wallet session creation logic
    const response = new messages.WalletSessionResponse();
    response.setSessionId('');
    response.setMethod(call.request.getMethod());
    response.setSuccess(false);
    response.setErrorMessage('Not implemented');

    callback(null, response);
  }

  // Unary RPC: Health check
  private handleHealthCheck(
    _call: grpc.ServerUnaryCall<messages.HealthCheckRequest, messages.HealthCheckResponse>,
    callback: grpc.sendUnaryData<messages.HealthCheckResponse>
  ): void {
    console.log('HealthCheck called');

    const response = new messages.HealthCheckResponse();
    response.setStatus('healthy');

    const dependencies = response.getDependenciesMap();
    dependencies.set('galachain', 'connected');
    dependencies.set('database', 'connected');

    const uptimeSeconds = Math.floor((Date.now() - this.startTime.getTime()) / 1000);
    response.setUptimeSeconds(uptimeSeconds);

    callback(null, response);
  }

  // Start the gRPC server
  public async start(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.server.bindAsync(
        `0.0.0.0:${this.port}`,
        grpc.ServerCredentials.createInsecure(),
        (error, port) => {
          if (error) {
            reject(error);
            return;
          }

          console.log(`gRPC server started on port ${port}`);
          resolve();
        }
      );
    });
  }

  // Stop the gRPC server gracefully
  public async stop(): Promise<void> {
    return new Promise((resolve) => {
      this.server.tryShutdown(() => {
        console.log('gRPC server stopped');
        resolve();
      });
    });
  }

  // Force stop the gRPC server
  public forceStop(): void {
    this.server.forceShutdown();
    console.log('gRPC server force stopped');
  }
}
