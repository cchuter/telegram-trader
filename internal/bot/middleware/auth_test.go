package middleware

import (
	"context"
	"testing"

	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/go-telegram/bot/models"
)

// mockDatabase is a mock implementation of storage.Database for testing
type mockDatabase struct {
	sessions map[int64]*storage.UserSession
}

func newMockDatabase() *mockDatabase {
	return &mockDatabase{
		sessions: make(map[int64]*storage.UserSession),
	}
}

func (m *mockDatabase) GetUserSession(ctx context.Context, userID int64) (*storage.UserSession, error) {
	session, ok := m.sessions[userID]
	if !ok {
		return nil, nil
	}
	return session, nil
}

func (m *mockDatabase) SaveUserSession(ctx context.Context, session *storage.UserSession) error {
	m.sessions[session.UserID] = session
	return nil
}

func (m *mockDatabase) GetWalletSession(ctx context.Context, userID int64, walletType string) (*storage.WalletSession, error) {
	return nil, nil
}

func (m *mockDatabase) SaveWalletSession(ctx context.Context, session *storage.WalletSession) error {
	return nil
}

func (m *mockDatabase) Close() error {
	return nil
}

func TestParseWhitelist(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int64
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []int64{},
		},
		{
			name:     "single user ID",
			input:    "123456",
			expected: []int64{123456},
		},
		{
			name:     "multiple user IDs",
			input:    "123456,789012,345678",
			expected: []int64{123456, 789012, 345678},
		},
		{
			name:     "user IDs with spaces",
			input:    "123456, 789012, 345678",
			expected: []int64{123456, 789012, 345678},
		},
		{
			name:     "invalid user ID ignored",
			input:    "123456,invalid,789012",
			expected: []int64{123456, 789012},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseWhitelist(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseWhitelist() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, id := range result {
				if id != tt.expected[i] {
					t.Errorf("parseWhitelist()[%d] = %v, want %v", i, id, tt.expected[i])
				}
			}
		})
	}
}

func TestIsWhitelisted(t *testing.T) {
	db := newMockDatabase()
	middleware := NewAuthMiddleware(db, "123456,789012")

	tests := []struct {
		name     string
		userID   int64
		expected bool
	}{
		{
			name:     "whitelisted user",
			userID:   123456,
			expected: true,
		},
		{
			name:     "another whitelisted user",
			userID:   789012,
			expected: true,
		},
		{
			name:     "non-whitelisted user",
			userID:   999999,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.isWhitelisted(tt.userID)
			if result != tt.expected {
				t.Errorf("isWhitelisted(%d) = %v, want %v", tt.userID, result, tt.expected)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	db := newMockDatabase()
	middleware := NewAuthMiddleware(db, "123456")

	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		shouldErr bool
	}{
		{
			name:      "whitelisted user authenticates successfully",
			userID:    123456,
			shouldErr: false,
		},
		{
			name:      "non-whitelisted user fails authentication",
			userID:    999999,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock update
			update := &models.Update{
				Message: &models.Message{
					From: &models.User{
						ID:       tt.userID,
						Username: "testuser",
					},
					Chat: models.Chat{
						ID: 123,
					},
				},
			}

			// Note: We can't fully test this without a real bot instance,
			// but we can test that it returns the correct error
			err := middleware.Authenticate(ctx, nil, update)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Authenticate() error = %v, shouldErr %v", err, tt.shouldErr)
				return
			}

			// If authentication succeeded, verify session was created
			if !tt.shouldErr {
				session, _ := db.GetUserSession(ctx, tt.userID)
				if session == nil {
					t.Error("Expected user session to be created, but it wasn't")
				} else if session.UserID != tt.userID {
					t.Errorf("Session UserID = %d, want %d", session.UserID, tt.userID)
				}
			}
		})
	}
}

func TestNewAuthMiddleware(t *testing.T) {
	db := newMockDatabase()

	middleware := NewAuthMiddleware(db, "123456,789012")

	if middleware == nil {
		t.Error("NewAuthMiddleware() returned nil")
	}

	if middleware.db != db {
		t.Error("NewAuthMiddleware() did not set db correctly")
	}

	if len(middleware.whitelistIDs) != 2 {
		t.Errorf("NewAuthMiddleware() whitelist length = %d, want 2", len(middleware.whitelistIDs))
	}

	expectedIDs := map[int64]bool{123456: true, 789012: true}
	for _, id := range middleware.whitelistIDs {
		if !expectedIDs[id] {
			t.Errorf("Unexpected ID in whitelist: %d", id)
		}
	}
}
