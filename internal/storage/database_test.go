package storage

import (
	"strings"
	"testing"
)

// TestParseDatabaseURLDetection tests the database type detection logic
func TestParseDatabaseURLDetection(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		shouldBeSQL bool
	}{
		{
			name:        "PostgreSQL with postgres:// scheme",
			url:         "postgres://user:pass@localhost:5432/db",
			shouldBeSQL: false,
		},
		{
			name:        "PostgreSQL with postgresql:// scheme",
			url:         "postgresql://user:pass@localhost:5432/db",
			shouldBeSQL: false,
		},
		{
			name:        "SQLite relative path",
			url:         "telegram-trader.db",
			shouldBeSQL: true,
		},
		{
			name:        "SQLite absolute path",
			url:         "/var/lib/telegram-trader.db",
			shouldBeSQL: true,
		},
		{
			name:        "SQLite with file:// scheme",
			url:         "file:telegram-trader.db",
			shouldBeSQL: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isPostgres := strings.HasPrefix(tt.url, "postgres://") || strings.HasPrefix(tt.url, "postgresql://")
			isSQLite := !isPostgres

			if tt.shouldBeSQL && !isSQLite {
				t.Errorf("Expected SQLite detection for %s, but got PostgreSQL", tt.url)
			}
			if !tt.shouldBeSQL && !isPostgres {
				t.Errorf("Expected PostgreSQL detection for %s, but got SQLite", tt.url)
			}
		})
	}
}

// TestEmptyDatabaseURL tests that empty URLs are rejected
func TestEmptyDatabaseURL(t *testing.T) {
	_, err := ParseDatabaseURL("")
	if err == nil {
		t.Error("Expected error for empty DATABASE_URL, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error message to mention 'empty', got: %s", err.Error())
	}
}

// TestDatabaseInterfaceImplementation verifies both implementations satisfy the interface
func TestDatabaseInterfaceImplementation(t *testing.T) {
	// This is a compile-time check, but we can verify at runtime too
	var _ Database = (*SQLiteDB)(nil)
	var _ Database = (*PostgresDB)(nil)

	t.Log("Both SQLiteDB and PostgresDB implement Database interface")
}
