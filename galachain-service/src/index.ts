import { GalaChainGrpcServer } from './server/grpc';
import { HealthChecker } from './server/health';
import { GSwapClient } from './gswap/client';
import { loadConfig } from './config';
import { Logger, LogLevel } from './logging/logger';

const logger = new Logger('galachain-service', LogLevel.INFO);

logger.info('GalaChain service starting');

async function main() {
  // Load configuration
  const config = loadConfig();

  // Initialize gRPC server
  const grpcServer = new GalaChainGrpcServer(
    config.GRPC_PORT,
    config.GSWAP_API_URL,
    config.WALLETCONNECT_PROJECT_ID
  );

  // Start the gRPC server
  await grpcServer.start();

  // Initialize health checker with GSwap client
  const gswapClient = new GSwapClient(config.GSWAP_API_URL);
  const healthChecker = new HealthChecker(config.GSWAP_API_URL, gswapClient);

  // Start health check HTTP server on port 8081
  const healthServer = await healthChecker.startHealthServer(8081);

  logger.info('GalaChain service initialized and running', {
    grpc_port: config.GRPC_PORT,
    health_port: 8081,
  });

  // Handle graceful shutdown
  const shutdown = async () => {
    logger.info('Shutting down GalaChain service...');
    await grpcServer.stop();
    healthServer.close();
    process.exit(0);
  };

  process.on('SIGINT', shutdown);
  process.on('SIGTERM', shutdown);
}

main().catch((error) => {
  logger.fatal('Failed to start GalaChain service', error);
});
