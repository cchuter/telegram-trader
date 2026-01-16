/**
 * Unit tests for Logger
 *
 * These tests verify:
 * - Log formatting (JSON)
 * - Log levels
 * - Context logging with correlation ID
 */

import { Logger, LogLevel } from '../logger';

describe('Logger', () => {
  let consoleLogSpy: jest.SpyInstance;

  beforeEach(() => {
    consoleLogSpy = jest.spyOn(console, 'log').mockImplementation();
    // Ensure JSON format for tests
    process.env.LOG_FORMAT = 'json';
  });

  afterEach(() => {
    consoleLogSpy.mockRestore();
    delete process.env.LOG_FORMAT;
  });

  describe('constructor', () => {
    it('should create logger with service name', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      expect(logger).toBeDefined();
    });

    it('should default to INFO level', () => {
      const logger = new Logger('test-service');
      logger.debug('debug message');

      // Debug should not log at INFO level
      expect(consoleLogSpy).not.toHaveBeenCalled();
    });
  });

  describe('log levels', () => {
    it('should log INFO messages', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.info('test message');

      expect(consoleLogSpy).toHaveBeenCalled();
      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.level).toBe('INFO');
      expect(logOutput.message).toBe('test message');
    });

    it('should log ERROR messages', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.error('error message', new Error('test error'));

      expect(consoleLogSpy).toHaveBeenCalled();
      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.level).toBe('ERROR');
      expect(logOutput.message).toBe('error message');
      expect(logOutput.error).toBe('test error');
    });

    it('should log DEBUG messages when level is DEBUG', () => {
      const logger = new Logger('test-service', LogLevel.DEBUG);
      logger.debug('debug message');

      expect(consoleLogSpy).toHaveBeenCalled();
      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.level).toBe('DEBUG');
    });

    it('should not log DEBUG when level is INFO', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.debug('debug message');

      expect(consoleLogSpy).not.toHaveBeenCalled();
    });
  });

  describe('logWithContext', () => {
    it('should log with correlation ID', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.logWithContext(
        LogLevel.INFO,
        'test message',
        'corr-123',
        undefined,
        'test_event'
      );

      expect(consoleLogSpy).toHaveBeenCalled();
      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.correlation_id).toBe('corr-123');
      expect(logOutput.event_type).toBe('test_event');
    });

    it('should log with user ID', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.logWithContext(
        LogLevel.INFO,
        'test message',
        'corr-123',
        12345,
        'test_event'
      );

      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.user_id).toBe(12345);
    });

    it('should log with additional data', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.logWithContext(
        LogLevel.INFO,
        'test message',
        'corr-123',
        undefined,
        'test_event',
        undefined,
        { key: 'value', count: 42 }
      );

      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.details).toEqual({ key: 'value', count: 42 });
    });

    it('should log errors with context', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.logWithContext(
        LogLevel.ERROR,
        'error occurred',
        'corr-123',
        12345,
        'error_event',
        'Error details here'
      );

      expect(consoleLogSpy).toHaveBeenCalled();
      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.level).toBe('ERROR');
      expect(logOutput.error).toBe('Error details here');
    });
  });

  describe('log format', () => {
    it('should include timestamp', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.info('test');

      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.timestamp).toBeDefined();
      expect(new Date(logOutput.timestamp)).toBeInstanceOf(Date);
    });

    it('should include service name', () => {
      const logger = new Logger('galachain-service', LogLevel.INFO);
      logger.info('test');

      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.service).toBe('galachain-service');
    });

    it('should produce valid JSON', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.info('test message');

      expect(() => {
        JSON.parse(consoleLogSpy.mock.calls[0][0]);
      }).not.toThrow();
    });
  });

  describe('error handling', () => {
    it('should handle string errors', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.error('error message', 'string error');

      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.error).toBe('string error');
    });

    it('should handle Error objects', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      const error = new Error('test error');
      logger.error('error message', error);

      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.error).toBe('test error');
    });

    it('should handle undefined errors', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.error('error message');

      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.error).toBeUndefined();
    });
  });

  describe('info with details', () => {
    it('should log details object', () => {
      const logger = new Logger('test-service', LogLevel.INFO);
      logger.info('test message', { foo: 'bar', num: 123 });

      const logOutput = JSON.parse(consoleLogSpy.mock.calls[0][0]);
      expect(logOutput.details).toEqual({ foo: 'bar', num: 123 });
    });
  });
});
