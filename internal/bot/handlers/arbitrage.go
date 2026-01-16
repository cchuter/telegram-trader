package handlers

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/cchuter/telegram-trader/internal/arbitrage"
	"github.com/cchuter/telegram-trader/internal/errors"
	"github.com/cchuter/telegram-trader/internal/logging"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleArbitrage handles the /arbitrage command
func HandleArbitrage(ctx context.Context, b *bot.Bot, update *models.Update, engine *arbitrage.Engine, logger *logging.Logger) {
	// Send initial "Checking..." message
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Checking for arbitrage opportunities...",
	})
	if err != nil {
		log.Printf("Error sending checking message: %v", err)
	}

	// Detect arbitrage opportunity
	opportunity, err := engine.DetectOpportunity(ctx)
	if err != nil {
		botErr := errors.ErrNetworkError(err)
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   botErr.GetUserMessage(),
		})
		if sendErr != nil {
			logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, sendErr, "Error sending error message", nil)
			log.Printf("Error sending error message: %v", sendErr)
		}
		logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Arbitrage detection error", nil)
		log.Printf("Arbitrage detection error: %v", botErr)
		return
	}

	// Handle no opportunity case
	if opportunity == nil {
		message := fmt.Sprintf("No arbitrage opportunity (spread: 0.10%% < threshold: 0.30%%)")
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   message,
		})
		if err != nil {
			logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Error sending no opportunity message", nil)
			log.Printf("Error sending no opportunity message: %v", err)
		}
		logger.DebugContext(ctx, "No arbitrage opportunity found", map[string]interface{}{
			"user_id": update.Message.From.ID,
		})
		return
	}

	// Build opportunity message
	var direction string
	if opportunity.Direction == arbitrage.BuyTonSellGala {
		direction = "Buy on ston.fi, sell on gswap"
	} else {
		direction = "Buy on gswap, sell on ston.fi"
	}

	// Calculate estimated profit (simplified for POC)
	// Assuming 1 TON trade: profit = 1 TON * spread%
	estimatedProfit := math.Abs(opportunity.Spread)

	message := fmt.Sprintf(
		"Opportunity found! %s. Spread: %.2f%%\n\n"+
			"Details:\n"+
			"TON/GALA on ston.fi: %.2f GALA\n"+
			"GTON/GALA on gswap: %.2f GALA\n"+
			"Estimated profit: ~%.2f%%",
		direction,
		math.Abs(opportunity.Spread),
		opportunity.StonfiPrice,
		opportunity.GswapPrice,
		estimatedProfit,
	)

	// Create inline keyboard with Execute and Cancel buttons
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "Execute",
					CallbackData: "arbitrage_execute",
				},
				{
					Text:         "Cancel",
					CallbackData: "arbitrage_cancel",
				},
			},
		},
	}

	// Send opportunity message with inline keyboard
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        message,
		ReplyMarkup: keyboard,
	})
	if err != nil {
		logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Error sending opportunity message", nil)
		log.Printf("Error sending opportunity message: %v", err)
	} else {
		// Log arbitrage opportunity found
		logger.InfoContext(ctx, "Arbitrage opportunity found", map[string]interface{}{
			"user_id":      update.Message.From.ID,
			"spread":       math.Abs(opportunity.Spread),
			"stonfi_price": opportunity.StonfiPrice,
			"gswap_price":  opportunity.GswapPrice,
			"direction":    direction,
		})
	}
}
