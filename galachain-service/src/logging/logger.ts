export enum LogLevel {
  DEBUG = 'DEBUG',
  INFO = 'INFO',
  WARN = 'WARN',
  ERROR = 'ERROR',
  FATAL = 'FATAL',
}

export enum LogFormat {
  JSON = 'json',
  TEXT = 'text',
}

export interface LogEntry {
  timestamp: string;
  level: string;
  service: string;
  event_type?: string;
  user_id?: number;
  correlation_id?: string;
  message: string;
  username?: string;
  command?: string;
  error?: string;
  details?: Record<string, any>;
}

export class Logger {
  private serviceName: string;
  private level: LogLevel;
  private format: LogFormat;

  constructor(serviceName: string, level: LogLevel = LogLevel.INFO) {
    this.serviceName = serviceName;
    this.level = level;

    // Check LOG_FORMAT env var, default to json
    const envFormat = process.env.LOG_FORMAT || 'json';
    this.format = envFormat === 'text' ? LogFormat.TEXT : LogFormat.JSON;
  }

  debug(message: string, details?: Record<string, any>): void {
    if (this.shouldLog(LogLevel.DEBUG)) {
      this.log(LogLevel.DEBUG, message, undefined, details);
    }
  }

  info(message: string, details?: Record<string, any>): void {
    if (this.shouldLog(LogLevel.INFO)) {
      this.log(LogLevel.INFO, message, undefined, details);
    }
  }

  warn(message: string, details?: Record<string, any>): void {
    if (this.shouldLog(LogLevel.WARN)) {
      this.log(LogLevel.WARN, message, undefined, details);
    }
  }

  error(message: string, error?: Error | string, details?: Record<string, any>): void {
    if (this.shouldLog(LogLevel.ERROR)) {
      const errorMsg = error instanceof Error ? error.message : error;
      this.log(LogLevel.ERROR, message, errorMsg, details);
    }
  }

  fatal(message: string, error?: Error | string, details?: Record<string, any>): void {
    const errorMsg = error instanceof Error ? error.message : error;
    this.log(LogLevel.FATAL, message, errorMsg, details);
    process.exit(1);
  }

  logCommand(
    correlationId: string,
    userId: number,
    username: string,
    command: string,
    details?: Record<string, any>
  ): void {
    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level: LogLevel.INFO,
      service: this.serviceName,
      correlation_id: correlationId,
      user_id: userId,
      username,
      event_type: 'command_executed',
      command,
      message: `User ${userId} executed command: ${command}`,
      details,
    };
    this.output(entry);
  }

  logError(
    correlationId: string,
    userId: number,
    username: string,
    error: Error | string,
    message: string,
    details?: Record<string, any>
  ): void {
    const errorMsg = error instanceof Error ? error.message : error;
    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level: LogLevel.ERROR,
      service: this.serviceName,
      correlation_id: correlationId,
      user_id: userId,
      username,
      event_type: 'error',
      message,
      error: errorMsg,
      details,
    };
    this.output(entry);
  }

  logWithContext(
    level: LogLevel,
    message: string,
    correlationId?: string,
    userId?: number,
    eventType?: string,
    error?: string,
    details?: Record<string, any>
  ): void {
    if (this.shouldLog(level)) {
      const entry: LogEntry = {
        timestamp: new Date().toISOString(),
        level,
        service: this.serviceName,
        message,
      };

      if (correlationId) entry.correlation_id = correlationId;
      if (userId) entry.user_id = userId;
      if (eventType) entry.event_type = eventType;
      if (error) entry.error = error;
      if (details) entry.details = details;

      this.output(entry);
    }
  }

  private log(
    level: LogLevel,
    message: string,
    error?: string,
    details?: Record<string, any>
  ): void {
    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level,
      service: this.serviceName,
      message,
    };

    if (error) entry.error = error;
    if (details) entry.details = details;

    this.output(entry);
  }

  private output(entry: LogEntry): void {
    if (this.format === LogFormat.JSON) {
      console.log(JSON.stringify(entry));
    } else {
      // Text format for local development
      let msg = `[${entry.timestamp}] ${entry.level} ${entry.service}: ${entry.message}`;
      if (entry.error) msg += ` | error=${entry.error}`;
      if (entry.correlation_id) msg += ` | correlation_id=${entry.correlation_id}`;
      if (entry.user_id) msg += ` | user_id=${entry.user_id}`;
      console.log(msg);
    }
  }

  private shouldLog(level: LogLevel): boolean {
    const levels: Record<LogLevel, number> = {
      [LogLevel.DEBUG]: 0,
      [LogLevel.INFO]: 1,
      [LogLevel.WARN]: 2,
      [LogLevel.ERROR]: 3,
      [LogLevel.FATAL]: 4,
    };
    return levels[level] >= levels[this.level];
  }
}

export function sanitizeAddress(address: string): string {
  if (address.length <= 5) {
    return address;
  }
  return address.substring(0, 2) + '...' + address.substring(address.length - 3);
}
