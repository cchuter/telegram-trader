package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cchuter/telegram-trader/internal/wallet"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleWallet handles the /wallet command
func HandleWallet(ctx context.Context, b *bot.Bot, update *models.Update, walletManager *wallet.Manager) {
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
		errorMsg := fmt.Sprintf("Failed to connect wallet: %v", err)
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   errorMsg,
		})
		if sendErr != nil {
			log.Printf("Error sending error message: %v", sendErr)
		}
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
		log.Printf("Error sending wallet confirmation message: %v", err)
	}
}
