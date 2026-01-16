import { Metadata } from '@grpc/grpc-js';
import { v4 as uuidv4 } from 'uuid';

/**
 * Metadata key for correlation ID in gRPC calls
 */
export const CORRELATION_ID_KEY = 'x-correlation-id';

/**
 * Generates a new UUID v4 correlation ID
 */
export function generateCorrelationID(): string {
  return uuidv4();
}

/**
 * Extracts correlation ID from gRPC metadata
 * Returns empty string if not found
 */
export function extractCorrelationID(metadata: Metadata): string {
  const values = metadata.get(CORRELATION_ID_KEY);
  if (values && values.length > 0) {
    return values[0] as string;
  }
  return '';
}

/**
 * Adds correlation ID to gRPC metadata
 */
export function addCorrelationID(metadata: Metadata, correlationID: string): void {
  metadata.set(CORRELATION_ID_KEY, correlationID);
}

/**
 * Context object with correlation ID for logging
 */
export interface LogContext {
  correlationId: string;
  [key: string]: any;
}

/**
 * Creates a log context with correlation ID
 */
export function createLogContext(correlationId: string, additionalContext: Record<string, any> = {}): LogContext {
  return {
    correlationId,
    ...additionalContext,
  };
}
