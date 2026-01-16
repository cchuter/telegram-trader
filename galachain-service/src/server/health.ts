import * as http from 'http';
import { GSwapClient } from '../gswap/client';

/**
 * Health status types
 */
export type HealthStatus = 'healthy' | 'degraded' | 'unhealthy';

/**
 * Dependency status
 */
export interface DependencyStatus {
  name: string;
  status: string;
  message?: string;
  latency?: string;
}

/**
 * Health check response
 */
export interface HealthCheckResponse {
  status: HealthStatus;
  timestamp: string;
  dependencies: DependencyStatus[];
  version: string;
  uptime: string;
}

/**
 * HealthChecker provides health check functionality for GalaChain service
 */
export class HealthChecker {
  private gswapClient: GSwapClient | null;
  private gswapApiUrl: string;
  private startTime: Date;

  constructor(gswapApiUrl: string, gswapClient?: GSwapClient) {
    this.gswapApiUrl = gswapApiUrl;
    this.gswapClient = gswapClient || null;
    this.startTime = new Date();
  }

  /**
   * Perform health checks on all dependencies
   */
  public async check(): Promise<HealthCheckResponse> {
    const dependencies: DependencyStatus[] = [];

    // Check gswap API connectivity
    const gswapStatus = await this.checkGswapAPI();
    dependencies.push(gswapStatus);

    // Check database connection (placeholder - not implemented in POC)
    const dbStatus = this.checkDatabase();
    dependencies.push(dbStatus);

    // Determine overall status
    const overallStatus = this.calculateOverallStatus(dependencies);

    // Calculate uptime
    const uptimeMs = Date.now() - this.startTime.getTime();
    const uptimeSeconds = Math.floor(uptimeMs / 1000);
    const uptime = this.formatUptime(uptimeSeconds);

    return {
      status: overallStatus,
      timestamp: new Date().toISOString(),
      dependencies,
      version: '1.0.0',
      uptime,
    };
  }

  /**
   * Check gswap API connectivity
   */
  private async checkGswapAPI(): Promise<DependencyStatus> {
    const start = Date.now();

    try {
      // Try to fetch a simple price to test API connectivity
      // Use a well-known pair: GTON/GALA
      if (this.gswapClient) {
        await this.gswapClient.getPrice('GTON', 'GALA');
      } else {
        // If no client provided, just check if the URL is reachable
        const response = await fetch(this.gswapApiUrl, {
          method: 'HEAD',
          signal: AbortSignal.timeout(5000),
        });

        if (!response.ok && response.status !== 404) {
          throw new Error(`API returned status ${response.status}`);
        }
      }

      const latency = `${Date.now() - start}ms`;

      return {
        name: 'gswap_api',
        status: 'healthy',
        message: 'Connected to GSwap API',
        latency,
      };
    } catch (error) {
      const latency = `${Date.now() - start}ms`;

      return {
        name: 'gswap_api',
        status: 'unhealthy',
        message: `GSwap API connection failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
        latency,
      };
    }
  }

  /**
   * Check database connection
   * Placeholder - not implemented in POC
   */
  private checkDatabase(): DependencyStatus {
    // In production, this would check actual database connectivity
    // For POC, we don't have a database in the GalaChain service yet
    return {
      name: 'database',
      status: 'healthy',
      message: 'No database configured (POC)',
    };
  }

  /**
   * Calculate overall health status based on dependencies
   */
  private calculateOverallStatus(dependencies: DependencyStatus[]): HealthStatus {
    let unhealthyCount = 0;
    let degradedCount = 0;

    for (const dep of dependencies) {
      if (dep.status === 'unhealthy') {
        unhealthyCount++;
      } else if (dep.status === 'degraded') {
        degradedCount++;
      }
    }

    // If multiple dependencies are unhealthy, overall is unhealthy
    if (unhealthyCount >= 2) {
      return 'unhealthy';
    }

    // If one dependency is unhealthy or any are degraded, overall is degraded
    if (unhealthyCount > 0 || degradedCount > 0) {
      return 'degraded';
    }

    // All dependencies are healthy
    return 'healthy';
  }

  /**
   * Format uptime in human-readable format
   */
  private formatUptime(seconds: number): string {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;

    const parts: string[] = [];
    if (days > 0) parts.push(`${days}d`);
    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    parts.push(`${secs}s`);

    return parts.join(' ');
  }

  /**
   * HTTP handler for health check endpoint
   */
  public getHTTPHandler(): (_req: http.IncomingMessage, res: http.ServerResponse) => Promise<void> {
    return async (_req: http.IncomingMessage, res: http.ServerResponse) => {
      try {
        // Perform health check
        const health = await this.check();

        // Set status code based on health
        let statusCode = 200;
        if (health.status === 'degraded') {
          statusCode = 200; // Still 200, but degraded
        } else if (health.status === 'unhealthy') {
          statusCode = 503; // Service Unavailable
        }

        // Write JSON response
        res.writeHead(statusCode, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify(health, null, 2));
      } catch (error) {
        console.error('Health check failed:', error);
        res.writeHead(500, { 'Content-Type': 'application/json' });
        res.end(
          JSON.stringify({
            status: 'unhealthy',
            error: error instanceof Error ? error.message : 'Unknown error',
          })
        );
      }
    };
  }

  /**
   * Start HTTP server for health checks
   */
  public async startHealthServer(port: number = 8081): Promise<http.Server> {
    const server = http.createServer(this.getHTTPHandler());

    return new Promise((resolve, reject) => {
      server.listen(port, () => {
        console.log(`Health check server listening on port ${port}`);
        resolve(server);
      });

      server.on('error', reject);
    });
  }
}
