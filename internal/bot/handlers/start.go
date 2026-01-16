package handlers

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleStart handles the /start command
func HandleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := `🚀 Welcome to GalaSwap & STON.fi Trading Bot!

Trade tokens on GalaChain and TON blockchain with ease.

Available Commands:
/wallet - Connect your TonKeeper wallet
/balance - View your balances
/swap - Execute a token swap
/price - Check token prices
/alert - Set price alerts
/portfolio - View your portfolio
/orders - Manage automated orders
/help - Show all commands

Get started by connecting your wallet with /wallet`

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		log.Printf("Error sending start message: %v", err)
	}
}
