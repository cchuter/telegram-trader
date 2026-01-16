package handlers

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleBalance handles the /balance command
func HandleBalance(ctx context.Context, b *bot.Bot, update *models.Update) {
	// POC phase: hardcoded TON balance, GalaChain service stub
	message := `TON: 10.0
GALA: N/A (GalaChain service not connected)`

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		log.Printf("Error sending balance message: %v", err)
	}
}
