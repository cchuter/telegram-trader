import * as grpc from '@grpc/grpc-js';
import { GalaChainServiceService, IGalaChainServiceServer } from '../types/galachain_grpc_pb';
import { GrpcHandlers } from './handlers';

export class GalaChainGrpcServer {
  private server: grpc.Server;
  private port: number;
  private startTime: Date;
  private handlers: GrpcHandlers;

  constructor(port: number = 50051, gswapApiUrl: string, walletConnectProjectId: string) {
    this.server = new grpc.Server();
    this.port = port;
    this.startTime = new Date();
    this.handlers = new GrpcHandlers(gswapApiUrl, walletConnectProjectId);
    this.setupHandlers();
  }

  private setupHandlers(): void {
    const serviceImplementation: IGalaChainServiceServer = {
      getPrice: this.handlers.getPrice.bind(this.handlers),
      getBalance: this.handlers.getBalance.bind(this.handlers),
      executeSwap: this.handlers.executeSwap.bind(this.handlers),
      watchPrices: this.handlers.watchPrices.bind(this.handlers),
      createWalletSession: this.handlers.createWalletSession.bind(this.handlers),
      healthCheck: (call, callback) => this.handlers.healthCheck(call, callback, this.startTime),
    };

    this.server.addService(GalaChainServiceService, serviceImplementation);
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
