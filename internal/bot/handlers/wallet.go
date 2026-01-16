package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cchuter/telegram-trader/internal/errors"
	"github.com/cchuter/telegram-trader/internal/logging"
	"github.com/cchuter/telegram-trader/internal/wallet"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleWallet handles the /wallet command
func HandleWallet(ctx context.Context, b *bot.Bot, update *models.Update, walletManager *wallet.Manager, logger *logging.Logger) {
	userID := update.Message.From.ID

	// Check if wallet is already connected
	existingWallet, err := walletManager.GetTonWallet(ctx, userID)
	if err == nil && existingWallet != nil && existingWallet.IsActive {
		// Wallet already connected
		displayAddress := existingWallet.Address
		if len(displayAddress) > 12 {
			displayAddress = displayAddress[:6] + "..." + displayAddress[len(displayAddress)-3:]
		}
		message := fmt.Sprintf("TON wallet already connected: %s\n\nTo connect a different wallet, disconnect first.", displayAddress)
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   message,
		})
		if sendErr != nil {
			log.Printf("Error sending wallet status message: %v", sendErr)
		}
		return
	}

	// Initiate TonConnect session
	session, qrURL, tonkeeperURL, err := walletManager.InitiateTonConnect(ctx, userID)
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
		logger.LogError(ctx, userID, update.Message.From.Username, err, "TonConnect initiation error", nil)
		log.Printf("TonConnect initiation error: %v", botErr)
		return
	}

	// Send TonConnect instructions with links
	message := fmt.Sprintf(`Connect your TON wallet using TonConnect:

📱 Mobile: Open this link in your Telegram app
%s

💼 TonKeeper: Tap here to connect
%s

Waiting for wallet approval... (expires in 5 minutes)`, qrURL, tonkeeperURL)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		logger.LogError(ctx, userID, update.Message.From.Username, err, "Error sending TonConnect message", nil)
		log.Printf("Error sending TonConnect message: %v", err)
		return
	}

	// Wait for connection in background (session has callback that saves to DB)
	// Monitor for connection with timeout
	go func() {
		timeout := time.After(5 * time.Minute)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				// Timeout - stop listening
				walletManager.StopTonConnect(userID)
				_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Wallet connection timeout. Please try /wallet again.",
				})
				logger.InfoContext(ctx, "TonConnect session timeout", map[string]interface{}{
					"user_id": userID,
				})
				return

			case <-ticker.C:
				// Check if wallet connected
				if session.IsConnected() {
					address := session.GetWalletAddress()
					displayAddress := address
					if len(address) > 12 {
						displayAddress = address[:6] + "..." + address[len(address)-3:]
					}

					_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("✅ TON wallet connected: %s", displayAddress),
					})

					logger.InfoContext(ctx, "Wallet connected successfully via TonConnect", map[string]interface{}{
						"user_id":        userID,
						"wallet_address": logging.SanitizeAddress(address),
						"chain":          "ton",
					})

					// Stop listening
					walletManager.StopTonConnect(userID)
					return
				}
			}
		}
	}()
}

// HandleDisconnect handles the /disconnect command to remove wallet connection
func HandleDisconnect(ctx context.Context, b *bot.Bot, update *models.Update, walletManager *wallet.Manager, logger *logging.Logger) {
	userID := update.Message.From.ID

	// Check if wallet is connected
	existingWallet, err := walletManager.GetTonWallet(ctx, userID)
	if err != nil || existingWallet == nil || !existingWallet.IsActive {
		// No wallet connected
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "No wallet is currently connected. Use /wallet to connect.",
		})
		if sendErr != nil {
			log.Printf("Error sending no wallet message: %v", sendErr)
		}
		return
	}

	// Show confirmation prompt with inline keyboard
	displayAddress := existingWallet.Address
	if len(displayAddress) > 12 {
		displayAddress = displayAddress[:6] + "..." + displayAddress[len(displayAddress)-3:]
	}

	message := fmt.Sprintf("Are you sure you want to disconnect your TON wallet?\n\nAddress: %s\n\n⚠️ This will remove your wallet connection and clear all encrypted keys.", displayAddress)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "Confirm Disconnect",
					CallbackData: "disconnect_confirm",
				},
				{
					Text:         "Cancel",
					CallbackData: "disconnect_cancel",
				},
			},
		},
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        message,
		ReplyMarkup: keyboard,
	})
	if err != nil {
		logger.LogError(ctx, userID, update.Message.From.Username, err, "Error sending disconnect confirmation", nil)
		log.Printf("Error sending disconnect confirmation: %v", err)
	}
}

// HandleDisconnectCallback handles the disconnect confirmation button callback
func HandleDisconnectCallback(ctx context.Context, b *bot.Bot, update *models.Update, walletManager *wallet.Manager, logger *logging.Logger) {
	userID := update.CallbackQuery.From.ID
	callbackData := update.CallbackQuery.Data

	// Handle cancel
	if callbackData == "disconnect_cancel" {
		_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
			Text:      "Wallet disconnect cancelled.",
		})
		if err != nil {
			log.Printf("Error updating message: %v", err)
		}
		return
	}

	// Handle confirm disconnect
	if callbackData == "disconnect_confirm" {
		// Disconnect the wallet
		err := walletManager.DisconnectWallet(ctx, userID, "ton")
		if err != nil {
			_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
				MessageID: update.CallbackQuery.Message.Message.ID,
				Text:      fmt.Sprintf("Error disconnecting wallet: %v", err),
			})
			logger.LogError(ctx, userID, update.CallbackQuery.From.Username, err, "Failed to disconnect wallet", nil)
			return
		}

		// Success message
		_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
			Text:      "✅ Wallet disconnected successfully.\n\nYour wallet session has been removed and all encrypted keys have been cleared.",
		})
		if err != nil {
			log.Printf("Error updating message: %v", err)
		}

		logger.InfoContext(ctx, "Wallet disconnected", map[string]interface{}{
			"user_id": userID,
			"chain":   "ton",
		})
	}
}
