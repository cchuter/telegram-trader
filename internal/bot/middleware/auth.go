package middleware

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// AuthMiddleware handles user authentication and authorization
type AuthMiddleware struct {
	db           storage.Database
	whitelistIDs []int64
}

// NewAuthMiddleware creates a new authentication middleware instance
func NewAuthMiddleware(db storage.Database, adminUserIDs string) *AuthMiddleware {
	// Parse comma-separated list of admin user IDs
	whitelist := parseWhitelist(adminUserIDs)

	return &AuthMiddleware{
		db:           db,
		whitelistIDs: whitelist,
	}
}

// parseWhitelist converts a comma-separated string of user IDs to a slice of int64
func parseWhitelist(adminUserIDs string) []int64 {
	if adminUserIDs == "" {
		return []int64{}
	}

	parts := strings.Split(adminUserIDs, ",")
	whitelist := make([]int64, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		id, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			log.Printf("Warning: invalid user ID in whitelist: %s", trimmed)
			continue
		}

		whitelist = append(whitelist, id)
	}

	return whitelist
}

// Authenticate checks if a user is whitelisted and creates/updates their session
func (m *AuthMiddleware) Authenticate(ctx context.Context, b *bot.Bot, update *models.Update) error {
	// Extract user ID from update
	if update.Message == nil || update.Message.From == nil {
		return fmt.Errorf("invalid update: missing message or user")
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	username := update.Message.From.Username

	// Check if user is whitelisted
	if !m.isWhitelisted(userID) {
		// Send access denied message if bot instance is available
		if b != nil {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Access denied. You are not authorized to use this bot.",
			})
			if err != nil {
				log.Printf("Error sending access denied message: %v", err)
			}
		}
		return fmt.Errorf("user %d is not whitelisted", userID)
	}

	// Check existing session for expiry
	existingSession, err := m.db.GetUserSession(ctx, userID)
	if err == nil && existingSession != nil {
		// Session exists - check if expired
		if !existingSession.ExpiresAt.IsZero() && existingSession.ExpiresAt.Before(time.Now()) {
			// Session expired
			if b != nil {
				_, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "Your session has expired. Please use /start to begin a new session.",
				})
				if err != nil {
					log.Printf("Error sending session expired message: %v", err)
				}
			}
			return fmt.Errorf("session expired for user %d", userID)
		}
	}

	// Create or update user session with renewed expiry (sliding window)
	now := time.Now()
	session := &storage.UserSession{
		UserID:    userID,
		ChatID:    chatID,
		Username:  username,
		ExpiresAt: now.Add(24 * time.Hour), // Renew for 24 hours from now
	}

	if err := m.db.SaveUserSession(ctx, session); err != nil {
		log.Printf("Error saving user session: %v", err)
		// Continue execution even if session save fails
	}

	return nil
}

// isWhitelisted checks if a user ID is in the whitelist
func (m *AuthMiddleware) isWhitelisted(userID int64) bool {
	for _, id := range m.whitelistIDs {
		if id == userID {
			return true
		}
	}
	return false
}
