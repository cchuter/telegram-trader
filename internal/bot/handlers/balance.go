package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/cchuter/telegram-trader/internal/errors"
	"github.com/cchuter/telegram-trader/internal/galachain"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleBalance handles the /balance command
func HandleBalance(ctx context.Context, b *bot.Bot, update *models.Update, galaClient *galachain.Client) {
	// Hardcoded TON balance for POC
	message := "TON: 10.0\n"

	// Fetch GalaChain balances via gRPC
	if galaClient != nil {
		userID := update.Message.From.ID
		resp, err := galaClient.GetBalance(ctx, userID)
		if err != nil {
			botErr := errors.ErrBalanceFetchFailed(err)
			log.Printf("Balance fetch error: %v", botErr)
			message += fmt.Sprintf("GALA: %s\nGTON: %s", botErr.GetUserMessage(), botErr.GetUserMessage())
		} else {
			// Parse balances from response
			galaBalance := "0.0"
			gtonBalance := "0.0"

			for _, balance := range resp.Balances {
				switch balance.Token {
				case "GALA":
					galaBalance = balance.Balance
				case "GTON":
					gtonBalance = balance.Balance
				}
			}

			message += fmt.Sprintf("GALA: %s\nGTON: %s", galaBalance, gtonBalance)
		}
	} else {
		message += "GALA: N/A (GalaChain service not connected)\nGTON: N/A (GalaChain service not connected)"
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		log.Printf("Error sending balance message: %v", err)
	}
}
