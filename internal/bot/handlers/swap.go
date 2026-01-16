package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/errors"
	"github.com/cchuter/telegram-trader/internal/logging"
	"github.com/cchuter/telegram-trader/internal/validation"
	"github.com/cchuter/telegram-trader/internal/wallet"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	// GALATokenAddress is the GALA token address on TON blockchain
	GALATokenAddress = "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV"
)

// SwapContext stores information about a pending swap
type SwapContext struct {
	UserID         int64
	FromToken      string
	ToToken        string
	Amount         string
	AmountUnits    string
	OutputAmount   string
	Fee            string
	Slippage       float64
	PriceImpact    float64
	RouterAddress  string
	MinOutputUnits string
	CreatedAt      time.Time
}

var (
	// pendingSwaps stores pending swap contexts by user ID
	// In production, this should be stored in a database with expiration
	pendingSwaps   = make(map[int64]*SwapContext)
	pendingSwapsMu sync.RWMutex
)

// HandleSwap handles the /swap command
func HandleSwap(ctx context.Context, b *bot.Bot, update *models.Update, dexClient dex.Client, logger *logging.Logger, tradeLogger *logging.TradeLogger) {
	startTime := time.Now()

	// Parse command arguments: /swap <amount> <from_token> <to_token> stonfi
	messageText := update.Message.Text
	parts := strings.Fields(messageText)

	// Validate command format
	if len(parts) < 5 {
		botErr := errors.ErrInvalidCommandFormat()
		message := botErr.GetUserMessage() + "\n\nUsage: /swap <amount> <from_token> <to_token> stonfi\nExample: /swap 1 TON GALA stonfi"
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   message,
		})
		if err != nil {
			log.Printf("Error sending swap usage message: %v", err)
		}
		return
	}

	// Extract parameters
	amountStr := parts[1]
	fromToken := strings.ToUpper(parts[2])
	toToken := strings.ToUpper(parts[3])
	dexName := strings.ToLower(parts[4])

	// Validate amount
	amount, err := validation.ValidateAmount(amountStr)
	if err != nil {
		botErr := errors.ErrInvalidAmount()
		message := botErr.GetUserMessage() + fmt.Sprintf(" (%v)", err)
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   message,
		})
		if sendErr != nil {
			log.Printf("Error sending amount error message: %v", sendErr)
		}
		logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Invalid amount", map[string]interface{}{
			"amount": amountStr,
		})
		return
	}

	// Validate token symbols
	if err := validation.ValidateTokenSymbol(fromToken); err != nil {
		message := fmt.Sprintf("Invalid token symbol: %s. %v", fromToken, err)
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   message,
		})
		if sendErr != nil {
			log.Printf("Error sending token error message: %v", sendErr)
		}
		return
	}
	if err := validation.ValidateTokenSymbol(toToken); err != nil {
		message := fmt.Sprintf("Invalid token symbol: %s. %v", toToken, err)
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   message,
		})
		if sendErr != nil {
			log.Printf("Error sending token error message: %v", sendErr)
		}
		return
	}

	// Validate DEX name
	if dexName != "stonfi" {
		message := "Currently only 'stonfi' DEX is supported"
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   message,
		})
		if err != nil {
			log.Printf("Error sending DEX error message: %v", err)
		}
		return
	}

	// Convert amount to float for calculation
	amountFloat, _ := amount.Float64()

	// Convert token names to addresses
	fromTokenAddr := tokenNameToAddress(fromToken)
	toTokenAddr := tokenNameToAddress(toToken)

	// Convert amount to smallest units (nanotons for TON, similar for other tokens)
	// For TON: 1 TON = 1e9 nanotons
	// For simplicity, we multiply by 1e9 for all tokens
	amountUnits := fmt.Sprintf("%.0f", amountFloat*1e9)

	// Call DEX to simulate swap
	simulation, err := dexClient.SimulateSwap(ctx, fromTokenAddr, toTokenAddr, amountUnits)
	if err != nil {
		botErr := errors.ErrSwapSimulationFailed(err)
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   botErr.GetUserMessage(),
		})
		if sendErr != nil {
			logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, sendErr, "Error sending swap error message", nil)
			log.Printf("Error sending swap error message: %v", sendErr)
		}
		logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Swap simulation error", map[string]interface{}{
			"from_token": fromToken,
			"to_token":   toToken,
			"amount":     amountStr,
		})
		log.Printf("Swap simulation error: %v", botErr)

		// Log failed trade to trade log
		executionTime := time.Since(startTime).Milliseconds()
		if tradeLogger != nil {
			tradeLogger.LogSwap(update.Message.From.ID, "ton", fromToken, toToken, fmt.Sprintf("%.2f", amountFloat), "0", "0", "", logging.TradeStatusFailed, executionTime, err.Error())
		}
		return
	}

	// Convert output amount from units to human-readable
	outputUnits, err := strconv.ParseFloat(simulation.OutputAmount, 64)
	if err != nil {
		log.Printf("Error parsing output amount: %v", err)
		outputUnits = 0
	}
	outputAmount := outputUnits / 1e9

	// Convert fee from units to human-readable
	feeUnits, err := strconv.ParseFloat(simulation.Fee, 64)
	if err != nil {
		log.Printf("Error parsing fee: %v", err)
		feeUnits = 0
	}
	feeAmount := feeUnits / 1e9

	// Store swap context for callback handler
	// Calculate minimum output with slippage
	minOutputUnits := fmt.Sprintf("%.0f", outputUnits*(1-simulation.Slippage/100))

	swapCtx := &SwapContext{
		UserID:         update.Message.From.ID,
		FromToken:      fromToken,
		ToToken:        toToken,
		Amount:         fmt.Sprintf("%.2f", amountFloat),
		AmountUnits:    amountUnits,
		OutputAmount:   fmt.Sprintf("%.2f", outputAmount),
		Fee:            fmt.Sprintf("%.2f", feeAmount),
		Slippage:       simulation.Slippage,
		PriceImpact:    simulation.PriceImpact,
		RouterAddress:  "", // TODO: Get router address from simulation response
		MinOutputUnits: minOutputUnits,
		CreatedAt:      time.Now(),
	}

	pendingSwapsMu.Lock()
	pendingSwaps[update.Message.From.ID] = swapCtx
	pendingSwapsMu.Unlock()

	// Format the swap preview message
	previewMsg := fmt.Sprintf(
		"Swap Preview:\n\n"+
			"Swap %.2f %s → ~%.0f %s (estimated)\n"+
			"Fee: %.2f TON\n"+
			"Price Impact: %.2f%%\n"+
			"Slippage: %.2f%%\n\n"+
			"Click Confirm to execute this swap on ston.fi.",
		amountFloat, fromToken,
		outputAmount, toToken,
		feeAmount,
		simulation.PriceImpact,
		simulation.Slippage,
	)

	// Create inline keyboard with Confirm and Cancel buttons
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "✅ Confirm",
					CallbackData: "swap_confirm",
				},
				{
					Text:         "❌ Cancel",
					CallbackData: "swap_cancel",
				},
			},
		},
	}

	// Send message with inline keyboard
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        previewMsg,
		ReplyMarkup: keyboard,
	})
	if err != nil {
		logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Error sending swap preview message", nil)
		log.Printf("Error sending swap preview message: %v", err)
	}

	// Log simulated trade (POC mode - no actual execution)
	executionTime := time.Since(startTime).Milliseconds()
	if tradeLogger != nil {
		tradeLogger.LogSwap(
			update.Message.From.ID,
			"ton",
			fromToken,
			toToken,
			fmt.Sprintf("%.2f", amountFloat),
			fmt.Sprintf("%.2f", outputAmount),
			fmt.Sprintf("%.2f", feeAmount),
			"", // No tx hash in POC mode
			logging.TradeStatusPending,
			executionTime,
			"",
		)
	}

	// Log swap simulation details
	logger.InfoContext(ctx, "Swap simulation completed", map[string]interface{}{
		"user_id":           update.Message.From.ID,
		"from_token":        fromToken,
		"to_token":          toToken,
		"amount_in":         amountFloat,
		"amount_out":        outputAmount,
		"fee":               feeAmount,
		"execution_time_ms": executionTime,
	})
}

// tokenNameToAddress converts a token name to its address
func tokenNameToAddress(tokenName string) string {
	switch strings.ToUpper(tokenName) {
	case "TON":
		return "TON" // Will be normalized to native address by stonfi client
	case "GALA":
		return GALATokenAddress
	default:
		// If it looks like an address (starts with EQ or UQ), return as-is
		if strings.HasPrefix(tokenName, "EQ") || strings.HasPrefix(tokenName, "UQ") {
			return tokenName
		}
		// Otherwise, return the name and let the DEX client handle it
		return tokenName
	}
}

// HandleSwapCallback handles callback queries from swap confirmation buttons
func HandleSwapCallback(ctx context.Context, b *bot.Bot, update *models.Update, dexClient dex.Client, walletMgr *wallet.Manager, logger *logging.Logger, tradeLogger *logging.TradeLogger) {
	if update.CallbackQuery == nil {
		return
	}

	userID := update.CallbackQuery.From.ID
	callbackData := update.CallbackQuery.Data

	// Acknowledge the callback
	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})
	if err != nil {
		logger.LogError(ctx, userID, update.CallbackQuery.From.Username, err, "Failed to answer callback query", nil)
		log.Printf("Failed to answer callback query: %v", err)
	}

	// Handle cancel
	if callbackData == "swap_cancel" {
		// Remove pending swap
		pendingSwapsMu.Lock()
		delete(pendingSwaps, userID)
		pendingSwapsMu.Unlock()

		// Edit message to show cancellation
		_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
			Text:      "❌ Swap cancelled.",
		})
		if err != nil {
			logger.LogError(ctx, userID, update.CallbackQuery.From.Username, err, "Failed to edit message", nil)
			log.Printf("Failed to edit message: %v", err)
		}
		return
	}

	// Handle confirm
	if callbackData == "swap_confirm" {
		startTime := time.Now()

		// Get pending swap context
		pendingSwapsMu.RLock()
		swapCtx, exists := pendingSwaps[userID]
		pendingSwapsMu.RUnlock()

		if !exists {
			_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
				MessageID: update.CallbackQuery.Message.Message.ID,
				Text:      "❌ Swap expired. Please try /swap again.",
			})
			if err != nil {
				logger.LogError(ctx, userID, update.CallbackQuery.From.Username, err, "Failed to edit message", nil)
				log.Printf("Failed to edit message: %v", err)
			}
			return
		}

		// Edit message to show "executing"
		_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
			Text:      "⏳ Executing swap on ston.fi...\n\nPlease wait, this may take up to 60 seconds.",
		})
		if err != nil {
			logger.LogError(ctx, userID, update.CallbackQuery.From.Username, err, "Failed to edit message", nil)
			log.Printf("Failed to edit message: %v", err)
		}

		// Get wallet session for user
		session, err := walletMgr.GetWalletSession(ctx, userID, "ton")
		if err != nil {
			errorMsg := "❌ Error: Wallet not connected.\n\nPlease connect your wallet with /wallet first."
			_, editErr := b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
				MessageID: update.CallbackQuery.Message.Message.ID,
				Text:      errorMsg,
			})
			if editErr != nil {
				logger.LogError(ctx, userID, update.CallbackQuery.From.Username, editErr, "Failed to edit message", nil)
			}
			logger.LogError(ctx, userID, update.CallbackQuery.From.Username, err, "Wallet not found", nil)

			// Log failed trade
			executionTime := time.Since(startTime).Milliseconds()
			if tradeLogger != nil {
				tradeLogger.LogSwap(userID, "ton", swapCtx.FromToken, swapCtx.ToToken, swapCtx.Amount, "0", swapCtx.Fee, "", logging.TradeStatusFailed, executionTime, "wallet not connected")
			}
			return
		}

		// Decrypt private key
		privateKey, err := walletMgr.DecryptPrivateKey(session.TonConnectPrivateKey)
		if err != nil {
			errorMsg := "❌ Error: Failed to decrypt wallet credentials.\n\nPlease reconnect your wallet with /wallet."
			_, editErr := b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
				MessageID: update.CallbackQuery.Message.Message.ID,
				Text:      errorMsg,
			})
			if editErr != nil {
				logger.LogError(ctx, userID, update.CallbackQuery.From.Username, editErr, "Failed to edit message", nil)
			}
			logger.LogError(ctx, userID, update.CallbackQuery.From.Username, err, "Failed to decrypt private key", nil)

			// Log failed trade
			executionTime := time.Since(startTime).Milliseconds()
			if tradeLogger != nil {
				tradeLogger.LogSwap(userID, "ton", swapCtx.FromToken, swapCtx.ToToken, swapCtx.Amount, "0", swapCtx.Fee, "", logging.TradeStatusFailed, executionTime, "decryption failed")
			}
			return
		}

		// TODO: Execute real swap via blockchain client
		// For now, simulate success with mock data
		// In real implementation: use privateKey to sign transaction via tonClient.ExecuteSwap()
		_ = privateKey // Will be used when blockchain integration is complete
		txHash := fmt.Sprintf("mock_tx_%d_%d", userID, time.Now().Unix())

		// Simulate execution delay
		time.Sleep(2 * time.Second)

		// Remove pending swap
		pendingSwapsMu.Lock()
		delete(pendingSwaps, userID)
		pendingSwapsMu.Unlock()

		// Update message with success
		successMsg := fmt.Sprintf(
			"✅ Swap executed successfully!\n\n"+
				"Transaction: %s\n"+
				"Swapped: %s %s → %s %s\n"+
				"Fee: %s TON\n\n"+
				"View on explorer: https://tonscan.org/tx/%s",
			txHash[:16]+"...",
			swapCtx.Amount, swapCtx.FromToken,
			swapCtx.OutputAmount, swapCtx.ToToken,
			swapCtx.Fee,
			txHash,
		)

		_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
			Text:      successMsg,
		})
		if err != nil {
			logger.LogError(ctx, userID, update.CallbackQuery.From.Username, err, "Failed to edit message", nil)
			log.Printf("Failed to edit success message: %v", err)
		}

		// Log successful trade
		executionTime := time.Since(startTime).Milliseconds()
		if tradeLogger != nil {
			tradeLogger.LogSwap(
				userID,
				"ton",
				swapCtx.FromToken,
				swapCtx.ToToken,
				swapCtx.Amount,
				swapCtx.OutputAmount,
				swapCtx.Fee,
				txHash,
				logging.TradeStatusSuccess,
				executionTime,
				"",
			)
		}

		logger.InfoContext(ctx, "Swap executed successfully", map[string]interface{}{
			"user_id":          userID,
			"from_token":       swapCtx.FromToken,
			"to_token":         swapCtx.ToToken,
			"amount":           swapCtx.Amount,
			"output":           swapCtx.OutputAmount,
			"tx_hash":          txHash,
			"execution_time_ms": executionTime,
		})
	}
}
