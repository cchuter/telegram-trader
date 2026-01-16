package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cchuter/telegram-trader/internal/errors"
	"github.com/cchuter/telegram-trader/internal/logging"
	"github.com/cchuter/telegram-trader/internal/wallet"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleWallet handles the /wallet command
func HandleWallet(ctx context.Context, b *bot.Bot, update *models.Update, walletManager *wallet.Manager, logger *logging.Logger) {
	// Parse command arguments
	messageText := update.Message.Text
	parts := strings.Fields(messageText)

	// If no wallet address provided, show usage instructions
	if len(parts) == 1 {
		message := "Send wallet address using: /wallet <address>"
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   message,
		})
		if err != nil {
			log.Printf("Error sending wallet usage message: %v", err)
		}
		return
	}

	// Extract wallet address
	walletAddress := parts[1]

	// Connect wallet
	userID := update.Message.From.ID
	err := walletManager.ConnectTonWallet(ctx, userID, walletAddress)
	if err != nil {
		// Create user-friendly error message
		botErr := errors.ErrWalletTimeout(err)
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   botErr.GetUserMessage(),
		})
		if sendErr != nil {
			logger.LogError(ctx, userID, update.Message.From.Username, sendErr, "Error sending error message", nil)
			log.Printf("Error sending error message: %v", sendErr)
		}
		logger.LogError(ctx, userID, update.Message.From.Username, err, "Wallet connection error", map[string]interface{}{
			"wallet_address": logging.SanitizeAddress(walletAddress),
		})
		log.Printf("Wallet connection error: %v", botErr)
		return
	}

	// Send confirmation
	// Truncate address for display: show first 6 and last 3 characters
	displayAddress := walletAddress
	if len(walletAddress) > 12 {
		displayAddress = walletAddress[:6] + "..." + walletAddress[len(walletAddress)-3:]
	}

	confirmationMsg := fmt.Sprintf("TON wallet connected: %s", displayAddress)
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   confirmationMsg,
	})
	if err != nil {
		logger.LogError(ctx, userID, update.Message.From.Username, err, "Error sending wallet confirmation message", nil)
		log.Printf("Error sending wallet confirmation message: %v", err)
	} else {
		// Log successful wallet connection
		logger.InfoContext(ctx, "Wallet connected successfully", map[string]interface{}{
			"user_id":        userID,
			"wallet_address": logging.SanitizeAddress(walletAddress),
			"chain":          "ton",
		})
	}
}
