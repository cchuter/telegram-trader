package logging

import (
	"context"

	"github.com/google/uuid"
)

// CorrelationIDKey is the context key for correlation IDs
type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
)

// GenerateCorrelationID generates a new UUID v4 correlation ID
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// ContextWithCorrelationID adds a correlation ID to the context
func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// GetCorrelationIDFromContext extracts correlation ID from context
func GetCorrelationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}
