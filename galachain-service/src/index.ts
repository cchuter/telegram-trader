import { GalaChainGrpcServer } from './server/grpc';
import { loadConfig } from './config';

console.log('GalaChain service starting');

async function main() {
  // Load configuration
  const config = loadConfig();

  // Initialize gRPC server
  const grpcServer = new GalaChainGrpcServer(
    config.GRPC_PORT,
    config.GSWAP_API_URL,
    config.WALLETCONNECT_PROJECT_ID
  );

  // Start the server
  await grpcServer.start();

  console.log('GalaChain service initialized and running');

  // Handle graceful shutdown
  const shutdown = async () => {
    console.log('Shutting down GalaChain service...');
    await grpcServer.stop();
    process.exit(0);
  };

  process.on('SIGINT', shutdown);
  process.on('SIGTERM', shutdown);
}

main().catch((error) => {
  console.error('Failed to start GalaChain service:', error);
  process.exit(1);
});
