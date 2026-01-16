package handlers

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/cchuter/telegram-trader/internal/arbitrage"
	"github.com/cchuter/telegram-trader/internal/blockchain/ton"
	"github.com/cchuter/telegram-trader/internal/errors"
	"github.com/cchuter/telegram-trader/internal/galachain/pb"
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
		_, errSend := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   message,
		})
		if errSend != nil {
			logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, errSend, "Error sending no opportunity message", nil)
			log.Printf("Error sending no opportunity message: %v", errSend)
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

// HandleArbitrageCallback handles the inline keyboard callbacks for arbitrage execution
func HandleArbitrageCallback(ctx context.Context, b *bot.Bot, update *models.Update, executor *arbitrage.Executor, engine *arbitrage.Engine, logger *logging.Logger, tradeLogger *logging.TradeLogger) {
	callback := update.CallbackQuery
	if callback == nil {
		return
	}

	// Handle cancel button
	if callback.Data == "arbitrage_cancel" {
		_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    callback.Message.Message.Chat.ID,
			MessageID: callback.Message.Message.ID,
			Text:      "Arbitrage execution cancelled.",
		})
		return
	}

	// Handle execute button
	if callback.Data == "arbitrage_execute" {
		// Update message to show execution in progress
		_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    callback.Message.Message.Chat.ID,
			MessageID: callback.Message.Message.ID,
			Text:      "Executing arbitrage... (both legs running concurrently)",
		})

		// For POC: Re-detect opportunity (in production, store from previous detection)
		opportunity, err := engine.DetectOpportunity(ctx)
		if err != nil || opportunity == nil {
			_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    callback.Message.Message.Chat.ID,
				MessageID: callback.Message.Message.ID,
				Text:      "Arbitrage execution failed: opportunity no longer available.",
			})
			return
		}

		// Calculate position size (using mock balances for POC)
		// In production, fetch actual balances from blockchain
		tonBalance := 10.0    // Mock TON balance
		galaBalance := 5000.0 // Mock GALA balance
		position := engine.GetPositionSize(tonBalance, galaBalance, opportunity.Direction)

		if !position.Valid {
			_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    callback.Message.Message.Chat.ID,
				MessageID: callback.Message.Message.ID,
				Text:      fmt.Sprintf("Arbitrage execution failed: %s", position.Reason),
			})
			return
		}

		// Execute arbitrage (using mock wallet credentials for POC)
		// In production, retrieve actual wallet credentials from database
		result := executor.ExecuteArbitrage(
			ctx,
			opportunity,
			position,
			callback.From.ID,
			"mock_wallet_address", // In production: fetch from wallet manager
			"mock_private_key",    // In production: decrypt from wallet manager
		)

		// Format result message
		var resultMsg string
		if result.Success {
			resultMsg = fmt.Sprintf(
				"✅ Arbitrage executed successfully!\n\n"+
					"Ston.fi: %s (status: %s)\n"+
					"Gswap: %s (status: %s)\n\n"+
					"Profit: %.2f%%\n"+
					"Execution time: %s",
				result.StonfiResult.TxHash,
				result.StonfiResult.Status,
				result.GswapResult.TxHash,
				result.GswapResult.Status,
				result.Profit,
				result.ExecutionTime.String(),
			)
		} else if result.PartialSuccess {
			resultMsg = fmt.Sprintf(
				"⚠️ Arbitrage partially executed (one leg failed)\n\n"+
					"Ston.fi: %s (status: %s)\n"+
					"Gswap: %s (status: %s)\n\n"+
					"Error: %v\n"+
					"Execution time: %s",
				getResultHash(result.StonfiResult),
				getResultStatus(result.StonfiResult),
				getGswapHash(result.GswapResult),
				getGswapStatus(result.GswapResult),
				result.Error,
				result.ExecutionTime.String(),
			)
		} else {
			resultMsg = fmt.Sprintf(
				"❌ Arbitrage execution failed\n\n"+
					"Error: %v\n"+
					"Execution time: %s",
				result.Error,
				result.ExecutionTime.String(),
			)
		}

		// Update message with final result
		_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    callback.Message.Message.Chat.ID,
			MessageID: callback.Message.Message.ID,
			Text:      resultMsg,
		})

		// Log arbitrage execution to structured log
		logger.InfoContext(ctx, "Arbitrage execution completed", map[string]interface{}{
			"user_id":        callback.From.ID,
			"success":        result.Success,
			"partial":        result.PartialSuccess,
			"profit":         result.Profit,
			"execution_time": result.ExecutionTime.String(),
		})

		// Log to trade logger (JSONL)
		if tradeLogger != nil {
			status := logging.TradeStatusFailed
			if result.Success {
				status = logging.TradeStatusSuccess
			} else if result.PartialSuccess {
				status = logging.TradeStatusPending // Partial success treated as pending for review
			}

			stonfiTxHash := getResultHash(result.StonfiResult)
			gswapTxHash := getGswapHash(result.GswapResult)

			// Determine trade direction based on opportunity
			fromToken := "TON"
			toToken := "GALA"
			amountIn := "N/A"  // TODO: Extract from position size
			amountOut := "N/A" // TODO: Extract from result
			fee := "N/A"       // TODO: Sum fees from both legs

			errorMsg := ""
			if result.Error != nil {
				errorMsg = result.Error.Error()
			}

			_ = tradeLogger.LogArbitrage(
				callback.From.ID,
				fromToken,
				toToken,
				amountIn,
				amountOut,
				fee,
				stonfiTxHash,
				gswapTxHash,
				status,
				result.ExecutionTime.Milliseconds(),
				errorMsg,
			)
		}
	}
}

// Helper functions to safely access result fields
func getResultHash(result *ton.SwapResult) string {
	if result == nil {
		return "N/A"
	}
	return result.TxHash
}

func getResultStatus(result *ton.SwapResult) string {
	if result == nil {
		return "N/A"
	}
	return result.Status
}

func getGswapHash(result *pb.SwapResponse) string {
	if result == nil {
		return "N/A"
	}
	return result.TxHash
}

func getGswapStatus(result *pb.SwapResponse) string {
	if result == nil {
		return "N/A"
	}
	return result.Status
}
