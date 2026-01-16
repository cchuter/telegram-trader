package middleware

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/go-telegram/bot/models"
)

// mockDatabase is a mock implementation of storage.Database for testing
type mockDatabase struct {
	sessions  map[int64]*storage.UserSession
	saveError error
	getError  error
}

func newMockDatabase() *mockDatabase {
	return &mockDatabase{
		sessions: make(map[int64]*storage.UserSession),
	}
}

func (m *mockDatabase) GetUserSession(ctx context.Context, userID int64) (*storage.UserSession, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	session, ok := m.sessions[userID]
	if !ok {
		return nil, nil
	}
	return session, nil
}

func (m *mockDatabase) SaveUserSession(ctx context.Context, session *storage.UserSession) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.sessions[session.UserID] = session
	return nil
}

func (m *mockDatabase) GetWalletSession(ctx context.Context, userID int64, walletType string) (*storage.WalletSession, error) {
	return nil, nil
}

func (m *mockDatabase) SaveWalletSession(ctx context.Context, session *storage.WalletSession) error {
	return nil
}

func (m *mockDatabase) GetTradeHistory(ctx context.Context, userID int64, limit int) ([]*storage.TradeHistory, error) {
	return []*storage.TradeHistory{}, nil
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
		{
			name:     "empty entries ignored",
			input:    "123456,  ,789012",
			expected: []int64{123456, 789012},
		},
		{
			name:     "whitespace only entries ignored",
			input:    "123456,   ,  ,789012",
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
				} else if session.ExpiresAt.IsZero() {
					t.Error("Expected ExpiresAt to be set, but it was zero")
				}
			}
		})
	}
}

func TestAuthenticateInvalidUpdate(t *testing.T) {
	db := newMockDatabase()
	middleware := NewAuthMiddleware(db, "123456")
	ctx := context.Background()

	tests := []struct {
		name        string
		update      *models.Update
		errContains string
	}{
		{
			name: "nil message",
			update: &models.Update{
				Message: nil,
			},
			errContains: "invalid update",
		},
		{
			name: "nil user",
			update: &models.Update{
				Message: &models.Message{
					From: nil,
					Chat: models.Chat{
						ID: 123,
					},
				},
			},
			errContains: "invalid update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := middleware.Authenticate(ctx, nil, tt.update)
			if err == nil {
				t.Error("Expected error for invalid update, got nil")
				return
			}
			if !contains(err.Error(), tt.errContains) {
				t.Errorf("Error %q does not contain %q", err.Error(), tt.errContains)
			}
		})
	}
}

func TestNewAuthMiddleware(t *testing.T) {
	db := newMockDatabase()

	middleware := NewAuthMiddleware(db, "123456,789012")

	if middleware == nil {
		t.Error("NewAuthMiddleware() returned nil")
		return
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

func TestSessionExpiry(t *testing.T) {
	db := newMockDatabase()
	middleware := NewAuthMiddleware(db, "123456")
	ctx := context.Background()

	tests := []struct {
		name        string
		setupFunc   func()
		userID      int64
		shouldErr   bool
		errContains string
	}{
		{
			name: "expired session blocks access",
			setupFunc: func() {
				// Create an expired session
				expiredSession := &storage.UserSession{
					UserID:    123456,
					ChatID:    123,
					Username:  "testuser",
					ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
				}
				db.SaveUserSession(ctx, expiredSession)
			},
			userID:      123456,
			shouldErr:   true,
			errContains: "session expired",
		},
		{
			name: "valid session allows access",
			setupFunc: func() {
				// Create a valid session
				validSession := &storage.UserSession{
					UserID:    123456,
					ChatID:    123,
					Username:  "testuser",
					ExpiresAt: time.Now().Add(12 * time.Hour), // Valid for 12 more hours
				}
				db.SaveUserSession(ctx, validSession)
			},
			userID:    123456,
			shouldErr: false,
		},
		{
			name: "new user creates session with expiry",
			setupFunc: func() {
				// Clean up any existing session
				delete(db.sessions, int64(123456))
			},
			userID:    123456,
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Create update
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

			// Authenticate
			err := middleware.Authenticate(ctx, nil, update)

			// Verify error expectation
			if (err != nil) != tt.shouldErr {
				t.Errorf("Authenticate() error = %v, shouldErr %v", err, tt.shouldErr)
				return
			}

			if tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Error message %q does not contain %q", err.Error(), tt.errContains)
				}
			}

			// If no error, verify session was renewed
			if !tt.shouldErr {
				session, _ := db.GetUserSession(ctx, tt.userID)
				if session == nil {
					t.Error("Expected session to exist")
				} else {
					// Verify expiry is set and in the future
					if session.ExpiresAt.IsZero() {
						t.Error("Expected ExpiresAt to be set")
					} else if session.ExpiresAt.Before(time.Now()) {
						t.Error("Expected ExpiresAt to be in the future")
					}
					// Verify expiry is approximately 24 hours from now (within 1 minute tolerance)
					expectedExpiry := time.Now().Add(24 * time.Hour)
					diff := session.ExpiresAt.Sub(expectedExpiry)
					if diff < -1*time.Minute || diff > 1*time.Minute {
						t.Errorf("ExpiresAt = %v, want approximately %v (diff: %v)", session.ExpiresAt, expectedExpiry, diff)
					}
				}
			}
		})
	}
}

func TestSessionRenewal(t *testing.T) {
	db := newMockDatabase()
	middleware := NewAuthMiddleware(db, "123456")
	ctx := context.Background()

	// Create initial session
	initialExpiry := time.Now().Add(12 * time.Hour)
	initialSession := &storage.UserSession{
		UserID:    123456,
		ChatID:    123,
		Username:  "testuser",
		ExpiresAt: initialExpiry,
	}
	db.SaveUserSession(ctx, initialSession)

	// Simulate time passing
	time.Sleep(10 * time.Millisecond)

	// Authenticate again (should renew)
	update := &models.Update{
		Message: &models.Message{
			From: &models.User{
				ID:       123456,
				Username: "testuser",
			},
			Chat: models.Chat{
				ID: 123,
			},
		},
	}

	err := middleware.Authenticate(ctx, nil, update)
	if err != nil {
		t.Errorf("Authenticate() unexpected error: %v", err)
	}

	// Verify session was renewed
	renewedSession, _ := db.GetUserSession(ctx, 123456)
	if renewedSession == nil {
		t.Fatal("Expected session to exist")
	}

	// The renewed expiry should be later than the initial expiry
	if !renewedSession.ExpiresAt.After(initialExpiry) {
		t.Errorf("Expected ExpiresAt to be renewed (initial: %v, renewed: %v)", initialExpiry, renewedSession.ExpiresAt)
	}
}

func TestAuthenticateDatabaseErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("database get error is handled gracefully", func(t *testing.T) {
		db := newMockDatabase()
		db.getError = fmt.Errorf("database connection error")
		middleware := NewAuthMiddleware(db, "123456")

		update := &models.Update{
			Message: &models.Message{
				From: &models.User{
					ID:       123456,
					Username: "testuser",
				},
				Chat: models.Chat{
					ID: 123,
				},
			},
		}

		// Should not error - GetSession failure is non-fatal
		err := middleware.Authenticate(ctx, nil, update)
		if err != nil {
			t.Errorf("Expected no error when GetSession fails, got: %v", err)
		}
	})

	t.Run("database save error is logged but doesn't prevent authentication", func(t *testing.T) {
		db := newMockDatabase()
		db.saveError = fmt.Errorf("database save error")
		middleware := NewAuthMiddleware(db, "123456")

		update := &models.Update{
			Message: &models.Message{
				From: &models.User{
					ID:       123456,
					Username: "testuser",
				},
				Chat: models.Chat{
					ID: 123,
				},
			},
		}

		// Should not error - SaveSession failure is non-fatal
		err := middleware.Authenticate(ctx, nil, update)
		if err != nil {
			t.Errorf("Expected no error when SaveSession fails, got: %v", err)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
