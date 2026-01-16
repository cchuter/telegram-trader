package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cchuter/telegram-trader/internal/errors"
	"github.com/cchuter/telegram-trader/internal/logging"
	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleOrders handles the /orders command for showing trade history
func HandleOrders(ctx context.Context, b *bot.Bot, update *models.Update, db storage.Database, logger *logging.Logger) {
	userID := update.Message.From.ID
	username := update.Message.From.Username

	// Parse limit from command (default 10, max 100)
	limit := 10
	parts := strings.Fields(update.Message.Text)
	if len(parts) == 2 {
		if parsedLimit, err := strconv.Atoi(parts[1]); err == nil {
			limit = parsedLimit
		} else {
			logger.LogError(ctx, userID, username, err, "Invalid count parameter", map[string]interface{}{
				"parameter": parts[1],
			})
			_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Invalid count. Usage: /orders [count]\nExample: /orders 20",
			})
			if sendErr != nil {
				logger.LogError(ctx, userID, username, sendErr, "Error sending error message", nil)
			}
			return
		}
	}

	// Fetch trade history from database
	trades, err := db.GetTradeHistory(ctx, userID, limit)
	if err != nil {
		botErr := errors.ErrDatabaseRetrievalFailed(err)
		logger.LogError(ctx, userID, username, err, "Failed to fetch trade history", map[string]interface{}{
			"limit": limit,
		})
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   botErr.GetUserMessage(),
		})
		if sendErr != nil {
			logger.LogError(ctx, userID, username, sendErr, "Error sending error message", nil)
		}
		return
	}

	// Build response message
	var message strings.Builder
	if len(trades) == 0 {
		message.WriteString("No trade history found.\n\nExecute your first trade with /swap or /arbitrage!")
	} else {
		message.WriteString(fmt.Sprintf("📊 Trade History (last %d trades)\n\n", len(trades)))

		for i, trade := range trades {
			// Format timestamp
			timestamp := trade.CreatedAt.Format("Jan 2, 15:04")

			// Format trade type and status
			typeEmoji := "🔄"
			if trade.TradeType == "arbitrage" {
				typeEmoji = "⚡"
			}
			statusEmoji := getStatusEmoji(trade.Status)

			// Build trade header
			message.WriteString(fmt.Sprintf("%d. %s %s → %s %s\n", i+1, typeEmoji, trade.FromToken, trade.ToToken, statusEmoji))
			message.WriteString(fmt.Sprintf("   📅 %s\n", timestamp))

			// Amount in/out
			message.WriteString(fmt.Sprintf("   💰 In: %s %s\n", formatAmount(trade.AmountIn), trade.FromToken))
			if trade.AmountOut != "" {
				message.WriteString(fmt.Sprintf("   💸 Out: %s %s\n", formatAmount(trade.AmountOut), trade.ToToken))
			}

			// Fee
			if trade.Fee != "" {
				message.WriteString(fmt.Sprintf("   💵 Fee: %s\n", formatAmount(trade.Fee)))
			}

			// Chain
			message.WriteString(fmt.Sprintf("   🔗 Chain: %s\n", trade.Chain))

			// TX hash with clickable link
			if trade.TxHashTon != "" {
				explorerLink := fmt.Sprintf("https://tonviewer.com/transaction/%s", trade.TxHashTon)
				message.WriteString(fmt.Sprintf("   🔍 TX: [%s](%s)\n", truncateHash(trade.TxHashTon), explorerLink))
			}
			if trade.TxHashGala != "" {
				explorerLink := fmt.Sprintf("https://explorer.gala.com/tx/%s", trade.TxHashGala)
				message.WriteString(fmt.Sprintf("   🔍 TX (Gala): [%s](%s)\n", truncateHash(trade.TxHashGala), explorerLink))
			}

			// Error message if failed
			if trade.Status == "failed" && trade.ErrorMessage != "" {
				message.WriteString(fmt.Sprintf("   ❌ Error: %s\n", trade.ErrorMessage))
			}

			// Execution time
			if trade.ExecutionTimeMs > 0 {
				message.WriteString(fmt.Sprintf("   ⏱️ Time: %dms\n", trade.ExecutionTimeMs))
			}

			// Separator between trades
			if i < len(trades)-1 {
				message.WriteString("\n")
			}
		}
	}

	// Create inline keyboard with [Refresh] [Export CSV] buttons
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🔄 Refresh", CallbackData: fmt.Sprintf("orders_refresh_%d", limit)},
				{Text: "📥 Export CSV", CallbackData: "orders_export"},
			},
		},
	}

	// Send message with inline keyboard
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        message.String(),
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: keyboard,
	})
	if err != nil {
		logger.LogError(ctx, userID, username, err, "Error sending orders message", nil)
		log.Printf("Error sending orders message: %v", err)
	}

	logger.Info("Orders displayed", map[string]interface{}{
		"user_id": userID,
		"count":   len(trades),
		"limit":   limit,
	})
}

// HandleOrdersCallback handles callback queries from orders buttons (Refresh, Export CSV)
func HandleOrdersCallback(ctx context.Context, b *bot.Bot, update *models.Update, db storage.Database, logger *logging.Logger) {
	if update.CallbackQuery == nil {
		return
	}

	userID := update.CallbackQuery.From.ID
	username := update.CallbackQuery.From.Username
	callbackData := update.CallbackQuery.Data

	// Answer callback query to stop loading indicator
	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})
	if err != nil {
		logger.LogError(ctx, userID, username, err, "Error answering callback query", nil)
	}

	// Handle refresh button
	if strings.HasPrefix(callbackData, "orders_refresh_") {
		// Parse limit from callback data
		limitStr := strings.TrimPrefix(callbackData, "orders_refresh_")
		limit := 10
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}

		// Fetch trade history
		trades, err := db.GetTradeHistory(ctx, userID, limit)
		if err != nil {
			botErr := errors.ErrDatabaseRetrievalFailed(err)
			logger.LogError(ctx, userID, username, err, "Failed to refresh trade history", nil)
			_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.CallbackQuery.Message.Message.Chat.ID,
				Text:   botErr.GetUserMessage(),
			})
			if sendErr != nil {
				logger.LogError(ctx, userID, username, sendErr, "Error sending error message", nil)
			}
			return
		}

		// Build refreshed message (same format as HandleOrders)
		var message strings.Builder
		if len(trades) == 0 {
			message.WriteString("No trade history found.\n\nExecute your first trade with /swap or /arbitrage!")
		} else {
			message.WriteString(fmt.Sprintf("📊 Trade History (last %d trades) 🔄\n\n", len(trades)))

			for i, trade := range trades {
				timestamp := trade.CreatedAt.Format("Jan 2, 15:04")
				typeEmoji := "🔄"
				if trade.TradeType == "arbitrage" {
					typeEmoji = "⚡"
				}
				statusEmoji := getStatusEmoji(trade.Status)

				message.WriteString(fmt.Sprintf("%d. %s %s → %s %s\n", i+1, typeEmoji, trade.FromToken, trade.ToToken, statusEmoji))
				message.WriteString(fmt.Sprintf("   📅 %s\n", timestamp))
				message.WriteString(fmt.Sprintf("   💰 In: %s %s\n", formatAmount(trade.AmountIn), trade.FromToken))
				if trade.AmountOut != "" {
					message.WriteString(fmt.Sprintf("   💸 Out: %s %s\n", formatAmount(trade.AmountOut), trade.ToToken))
				}
				if trade.Fee != "" {
					message.WriteString(fmt.Sprintf("   💵 Fee: %s\n", formatAmount(trade.Fee)))
				}
				message.WriteString(fmt.Sprintf("   🔗 Chain: %s\n", trade.Chain))
				if trade.TxHashTon != "" {
					explorerLink := fmt.Sprintf("https://tonviewer.com/transaction/%s", trade.TxHashTon)
					message.WriteString(fmt.Sprintf("   🔍 TX: [%s](%s)\n", truncateHash(trade.TxHashTon), explorerLink))
				}
				if trade.TxHashGala != "" {
					explorerLink := fmt.Sprintf("https://explorer.gala.com/tx/%s", trade.TxHashGala)
					message.WriteString(fmt.Sprintf("   🔍 TX (Gala): [%s](%s)\n", truncateHash(trade.TxHashGala), explorerLink))
				}
				if trade.Status == "failed" && trade.ErrorMessage != "" {
					message.WriteString(fmt.Sprintf("   ❌ Error: %s\n", trade.ErrorMessage))
				}
				if trade.ExecutionTimeMs > 0 {
					message.WriteString(fmt.Sprintf("   ⏱️ Time: %dms\n", trade.ExecutionTimeMs))
				}
				if i < len(trades)-1 {
					message.WriteString("\n")
				}
			}
		}

		// Create keyboard again
		keyboard := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "🔄 Refresh", CallbackData: fmt.Sprintf("orders_refresh_%d", limit)},
					{Text: "📥 Export CSV", CallbackData: "orders_export"},
				},
			},
		}

		// Edit message with refreshed data
		_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
			MessageID:   update.CallbackQuery.Message.Message.ID,
			Text:        message.String(),
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: keyboard,
		})
		if err != nil {
			logger.LogError(ctx, userID, username, err, "Error refreshing orders message", nil)
		}

		logger.Info("Orders refreshed", map[string]interface{}{
			"user_id": userID,
			"count":   len(trades),
		})
	}

	// Handle export CSV button
	if callbackData == "orders_export" {
		// TODO: Implement CSV export in future task
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   "CSV export coming soon!",
		})
		if err != nil {
			logger.LogError(ctx, userID, username, err, "Error sending export message", nil)
		}
	}
}

// getStatusEmoji returns an emoji for the trade status
func getStatusEmoji(status string) string {
	switch status {
	case "success":
		return "✅"
	case "pending":
		return "⏳"
	case "failed":
		return "❌"
	case "partial":
		return "⚠️"
	default:
		return "❓"
	}
}

// formatAmount formats a decimal amount string for display (max 8 decimal places)
func formatAmount(amount string) string {
	if amount == "" {
		return "N/A"
	}
	// Try to parse as float
	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return amount
	}
	// Format with max 8 decimals, removing trailing zeros
	formatted := fmt.Sprintf("%.8f", val)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	return formatted
}

// truncateHash truncates a transaction hash for display (first 8 chars ... last 6 chars)
func truncateHash(hash string) string {
	if len(hash) <= 20 {
		return hash
	}
	return hash[:8] + "..." + hash[len(hash)-6:]
}
