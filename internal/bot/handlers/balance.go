package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/cchuter/telegram-trader/internal/errors"
	"github.com/cchuter/telegram-trader/internal/galachain"
	"github.com/cchuter/telegram-trader/internal/logging"
	"github.com/cchuter/telegram-trader/internal/price"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleBalance handles the /balance command
func HandleBalance(ctx context.Context, b *bot.Bot, update *models.Update, galaClient *galachain.Client, priceClient *price.CoinGeckoClient, logger *logging.Logger) {
	// Hardcoded TON balance for POC
	tonBalance := 10.0
	message := fmt.Sprintf("TON: %.2f", tonBalance)

	// Fetch USD prices
	usdPrices := make(map[string]float64)
	if priceClient != nil {
		usdPrices = priceClient.GetMultipleUSDPrices(ctx, []string{"TON", "GALA", "GTON"})
	}

	// Add USD value for TON if available
	if usdPrice, exists := usdPrices["TON"]; exists {
		message += fmt.Sprintf(" ($%.2f)", tonBalance*usdPrice)
	}
	message += "\n"

	// Fetch GalaChain balances via gRPC
	if galaClient != nil {
		userID := update.Message.From.ID
		resp, err := galaClient.GetBalance(ctx, userID)
		if err != nil {
			botErr := errors.ErrBalanceFetchFailed(err)
			logger.LogError(ctx, userID, update.Message.From.Username, err, "Balance fetch error", map[string]interface{}{
				"chain": "galachain",
			})
			log.Printf("Balance fetch error: %v", botErr)
			message += fmt.Sprintf("GALA: %s\nGTON: %s", botErr.GetUserMessage(), botErr.GetUserMessage())
		} else {
			// Parse balances from response
			galaBalance := 0.0
			gtonBalance := 0.0

			for _, balance := range resp.Balances {
				switch balance.Token {
				case "GALA":
					if bal, err := strconv.ParseFloat(balance.Balance, 64); err == nil {
						galaBalance = bal
					}
				case "GTON":
					if bal, err := strconv.ParseFloat(balance.Balance, 64); err == nil {
						gtonBalance = bal
					}
				}
			}

			// Format GALA balance with USD value
			message += fmt.Sprintf("GALA: %.2f", galaBalance)
			if usdPrice, exists := usdPrices["GALA"]; exists && galaBalance > 0 {
				message += fmt.Sprintf(" ($%.2f)", galaBalance*usdPrice)
			}
			message += "\n"

			// Format GTON balance with USD value (use TON price for GTON)
			message += fmt.Sprintf("GTON: %.2f", gtonBalance)
			if usdPrice, exists := usdPrices["GTON"]; exists && gtonBalance > 0 {
				message += fmt.Sprintf(" ($%.2f)", gtonBalance*usdPrice)
			}
		}
	} else {
		message += "GALA: N/A (GalaChain service not connected)\nGTON: N/A (GalaChain service not connected)"
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Error sending balance message", nil)
		log.Printf("Error sending balance message: %v", err)
	}
}
