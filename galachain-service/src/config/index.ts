import * as dotenv from 'dotenv';

// Load environment variables from .env file
dotenv.config();

export interface Config {
  GRPC_PORT: number;
  DATABASE_URL: string;
  ENCRYPTION_KEY: string;
  GSWAP_API_URL: string;
}

export function loadConfig(): Config {
  const grpcPort = process.env.GRPC_PORT;
  const databaseUrl = process.env.DATABASE_URL;
  const encryptionKey = process.env.ENCRYPTION_KEY;
  const gswapApiUrl = process.env.GSWAP_API_URL;

  // Validate required environment variables
  if (!grpcPort) {
    throw new Error('GRPC_PORT environment variable is required');
  }
  if (!databaseUrl) {
    throw new Error('DATABASE_URL environment variable is required');
  }
  if (!encryptionKey) {
    throw new Error('ENCRYPTION_KEY environment variable is required');
  }
  if (!gswapApiUrl) {
    throw new Error('GSWAP_API_URL environment variable is required');
  }

  const port = parseInt(grpcPort, 10);
  if (isNaN(port)) {
    throw new Error('GRPC_PORT must be a valid number');
  }

  return {
    GRPC_PORT: port,
    DATABASE_URL: databaseUrl,
    ENCRYPTION_KEY: encryptionKey,
    GSWAP_API_URL: gswapApiUrl,
  };
}
