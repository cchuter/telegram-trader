package handlers

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleHelp handles the /help command
func HandleHelp(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := `Available Commands:

/wallet - Connect your TonKeeper wallet
/balance - View your balances
/swap - Execute a token swap
/price - Check token prices
/alert - Set price alerts
/portfolio - View your portfolio
/orders - Manage automated orders
/help - Show all commands

For more information about a command, simply type it in the chat.`

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		log.Printf("Error sending help message: %v", err)
	}
}
