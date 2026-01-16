package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	// GALATokenAddress is the GALA token address on TON blockchain
	GALATokenAddress = "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV"
)

// HandleSwap handles the /swap command
func HandleSwap(ctx context.Context, b *bot.Bot, update *models.Update, dexClient dex.Client) {
	// Parse command arguments: /swap <amount> <from_token> <to_token> stonfi
	messageText := update.Message.Text
	parts := strings.Fields(messageText)

	// Validate command format
	if len(parts) < 5 {
		message := "Usage: /swap <amount> <from_token> <to_token> stonfi\nExample: /swap 1 TON GALA stonfi"
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

	// Parse amount
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		message := "Invalid amount. Please provide a positive number."
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   message,
		})
		if sendErr != nil {
			log.Printf("Error sending amount error message: %v", sendErr)
		}
		return
	}

	// Convert token names to addresses
	fromTokenAddr := tokenNameToAddress(fromToken)
	toTokenAddr := tokenNameToAddress(toToken)

	// Convert amount to smallest units (nanotons for TON, similar for other tokens)
	// For TON: 1 TON = 1e9 nanotons
	// For simplicity, we multiply by 1e9 for all tokens
	amountUnits := fmt.Sprintf("%.0f", amount*1e9)

	// Call DEX to simulate swap
	simulation, err := dexClient.SimulateSwap(ctx, fromTokenAddr, toTokenAddr, amountUnits)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to simulate swap: %v", err)
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   errorMsg,
		})
		if sendErr != nil {
			log.Printf("Error sending swap error message: %v", sendErr)
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

	// Format the swap preview message
	previewMsg := fmt.Sprintf(
		"Swap Preview:\n\n"+
			"Swap %.2f %s → ~%.0f %s (estimated)\n"+
			"Fee: %.2f TON\n"+
			"Price Impact: %.2f%%\n"+
			"Slippage: %.2f%%\n\n"+
			"⚠️ POC Mode: This is a simulation only. No actual swap will be executed.",
		amount, fromToken,
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
					Text:         "Confirm",
					CallbackData: "swap_confirm",
				},
				{
					Text:         "Cancel",
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
		log.Printf("Error sending swap preview message: %v", err)
	}
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
