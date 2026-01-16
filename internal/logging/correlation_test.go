package logging

import (
	"context"
	"testing"
)

func TestGenerateCorrelationID(t *testing.T) {
	id1 := GenerateCorrelationID()
	id2 := GenerateCorrelationID()

	if id1 == "" {
		t.Error("Expected non-empty correlation ID")
	}

	if id1 == id2 {
		t.Error("Expected unique correlation IDs")
	}

	// UUID v4 format: 8-4-4-4-12 characters
	if len(id1) != 36 {
		t.Errorf("Expected UUID format with 36 characters, got %d", len(id1))
	}
}

func TestContextWithCorrelationID(t *testing.T) {
	ctx := context.Background()
	correlationID := "test-correlation-id-123"

	ctx = ContextWithCorrelationID(ctx, correlationID)

	retrievedID := GetCorrelationIDFromContext(ctx)
	if retrievedID != correlationID {
		t.Errorf("Expected correlation ID %s, got %s", correlationID, retrievedID)
	}
}

func TestGetCorrelationIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()

	retrievedID := GetCorrelationIDFromContext(ctx)
	if retrievedID != "" {
		t.Errorf("Expected empty correlation ID, got %s", retrievedID)
	}
}

func TestGetCorrelationIDFromContext_Nil(t *testing.T) {
	retrievedID := GetCorrelationIDFromContext(context.TODO())
	if retrievedID != "" {
		t.Errorf("Expected empty correlation ID for empty context, got %s", retrievedID)
	}
}
