package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	// LogLevelDebug for detailed debugging information
	LogLevelDebug LogLevel = "DEBUG"
	// LogLevelInfo for general informational messages
	LogLevelInfo LogLevel = "INFO"
	// LogLevelWarn for warning messages
	LogLevelWarn LogLevel = "WARN"
	// LogLevelError for error messages
	LogLevelError LogLevel = "ERROR"
	// LogLevelFatal for fatal errors that cause service termination
	LogLevelFatal LogLevel = "FATAL"
)

// Logger provides structured logging capabilities
type Logger struct {
	serviceName string
	level       LogLevel
	logger      *log.Logger
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp     string                 `json:"timestamp"`
	Level         string                 `json:"level"`
	Service       string                 `json:"service"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	UserID        int64                  `json:"user_id,omitempty"`
	Username      string                 `json:"username,omitempty"`
	EventType     string                 `json:"event_type,omitempty"`
	Command       string                 `json:"command,omitempty"`
	Message       string                 `json:"message"`
	Error         string                 `json:"error,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

// New creates a new Logger instance
func New(serviceName string, level LogLevel) *Logger {
	return &Logger{
		serviceName: serviceName,
		level:       level,
		logger:      log.New(os.Stdout, "", 0),
	}
}

// Debug logs a debug-level message
func (l *Logger) Debug(message string, details map[string]interface{}) {
	if l.shouldLog(LogLevelDebug) {
		l.log(LogLevelDebug, message, "", details)
	}
}

// DebugContext logs a debug-level message with context
func (l *Logger) DebugContext(ctx context.Context, message string, details map[string]interface{}) {
	if l.shouldLog(LogLevelDebug) {
		correlationID := getCorrelationID(ctx)
		l.logWithContext(LogLevelDebug, message, "", details, correlationID, 0, "")
	}
}

// Info logs an info-level message
func (l *Logger) Info(message string, details map[string]interface{}) {
	if l.shouldLog(LogLevelInfo) {
		l.log(LogLevelInfo, message, "", details)
	}
}

// InfoContext logs an info-level message with context
func (l *Logger) InfoContext(ctx context.Context, message string, details map[string]interface{}) {
	if l.shouldLog(LogLevelInfo) {
		correlationID := getCorrelationID(ctx)
		l.logWithContext(LogLevelInfo, message, "", details, correlationID, 0, "")
	}
}

// Warn logs a warning-level message
func (l *Logger) Warn(message string, details map[string]interface{}) {
	if l.shouldLog(LogLevelWarn) {
		l.log(LogLevelWarn, message, "", details)
	}
}

// WarnContext logs a warning-level message with context
func (l *Logger) WarnContext(ctx context.Context, message string, details map[string]interface{}) {
	if l.shouldLog(LogLevelWarn) {
		correlationID := getCorrelationID(ctx)
		l.logWithContext(LogLevelWarn, message, "", details, correlationID, 0, "")
	}
}

// Error logs an error-level message
func (l *Logger) Error(message string, err error, details map[string]interface{}) {
	if l.shouldLog(LogLevelError) {
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		l.log(LogLevelError, message, errMsg, details)
	}
}

// ErrorContext logs an error-level message with context
func (l *Logger) ErrorContext(ctx context.Context, message string, err error, details map[string]interface{}) {
	if l.shouldLog(LogLevelError) {
		correlationID := getCorrelationID(ctx)
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		l.logWithContext(LogLevelError, message, errMsg, details, correlationID, 0, "")
	}
}

// Fatal logs a fatal-level message and exits the program
func (l *Logger) Fatal(message string, err error, details map[string]interface{}) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	l.log(LogLevelFatal, message, errMsg, details)
	os.Exit(1)
}

// LogCommand logs a user command execution
func (l *Logger) LogCommand(ctx context.Context, userID int64, username, command string, details map[string]interface{}) {
	correlationID := getCorrelationID(ctx)
	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         string(LogLevelInfo),
		Service:       l.serviceName,
		CorrelationID: correlationID,
		UserID:        userID,
		Username:      username,
		EventType:     "command_executed",
		Command:       command,
		Message:       fmt.Sprintf("User %d executed command: %s", userID, command),
		Details:       details,
	}
	l.output(entry)
}

// LogError logs an error with full context
func (l *Logger) LogError(ctx context.Context, userID int64, username string, err error, message string, details map[string]interface{}) {
	correlationID := getCorrelationID(ctx)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         string(LogLevelError),
		Service:       l.serviceName,
		CorrelationID: correlationID,
		UserID:        userID,
		Username:      username,
		EventType:     "error",
		Message:       message,
		Error:         errMsg,
		Details:       details,
	}
	l.output(entry)
}

// log creates and outputs a log entry
func (l *Logger) log(level LogLevel, message, err string, details map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     string(level),
		Service:   l.serviceName,
		Message:   message,
		Error:     err,
		Details:   details,
	}
	l.output(entry)
}

// logWithContext creates and outputs a log entry with context information
func (l *Logger) logWithContext(level LogLevel, message, err string, details map[string]interface{}, correlationID string, userID int64, username string) {
	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         string(level),
		Service:       l.serviceName,
		CorrelationID: correlationID,
		UserID:        userID,
		Username:      username,
		Message:       message,
		Error:         err,
		Details:       details,
	}
	l.output(entry)
}

// output writes the log entry to stdout in JSON format
func (l *Logger) output(entry LogEntry) {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to standard logger if JSON marshaling fails
		l.logger.Printf("Failed to marshal log entry: %v", err)
		return
	}
	l.logger.Println(string(jsonBytes))
}

// shouldLog determines if a log level should be output based on configured level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}
	return levels[level] >= levels[l.level]
}

// getCorrelationID extracts correlation ID from context
func getCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value("correlation_id").(string); ok {
		return id
	}
	return ""
}

// SanitizeAddress truncates wallet address for logging (security measure)
// Example: UQabcdefghijklmnop... -> UQ...nop
func SanitizeAddress(address string) string {
	if len(address) <= 5 {
		return address
	}
	return address[:2] + "..." + address[len(address)-3:]
}
