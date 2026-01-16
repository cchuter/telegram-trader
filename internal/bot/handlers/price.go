package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/errors"
	"github.com/cchuter/telegram-trader/internal/galachain"
	"github.com/cchuter/telegram-trader/internal/logging"
	"github.com/cchuter/telegram-trader/internal/price"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	// OneTON is 1 TON in nanotons (1 TON = 1,000,000,000 nanotons)
	OneTON = "1000000000"
)

// HandlePrice handles the /price command
func HandlePrice(ctx context.Context, b *bot.Bot, update *models.Update, dexClient dex.Client, galaClient *galachain.Client, priceClient *price.CoinGeckoClient, logger *logging.Logger) {
	var message string

	// Fetch USD prices
	usdPrices := make(map[string]float64)
	if priceClient != nil {
		usdPrices = priceClient.GetMultipleUSDPrices(ctx, []string{"TON", "GALA"})
	}

	// Fetch TON/GALA price from ston.fi
	var stonfiPrice float64
	if dexClient != nil {
		sim, err := dexClient.SimulateSwap(ctx, "TON", GALATokenAddress, OneTON)
		if err != nil {
			botErr := errors.ErrPriceFetchFailed("ston.fi", err)
			log.Printf("Price fetch error: %v", botErr)
			message = fmt.Sprintf("%s\n", botErr.GetUserMessage())
		} else {
			// Convert output amount to GALA (assuming 9 decimals for GALA)
			galaAmount, err := strconv.ParseFloat(sim.OutputAmount, 64)
			if err != nil {
				log.Printf("Error parsing ston.fi output amount: %v", err)
				message = "Error parsing TON/GALA price from ston.fi\n"
			} else {
				stonfiPrice = galaAmount / 1e9 // Convert from smallest units to GALA
				message = fmt.Sprintf("TON/GALA on ston.fi: %.2f GALA", stonfiPrice)

				// Add USD equivalent if available
				if tonUSD, tonExists := usdPrices["TON"]; tonExists {
					message += fmt.Sprintf(" (~$%.2f per TON)", tonUSD)
				}
				message += "\n"
			}
		}
	} else {
		message = "TON/GALA on ston.fi: N/A (DEX client not available)\n"
	}

	// Fetch GTON/GALA price from GalaChain service
	var gswapPrice float64
	if galaClient != nil {
		priceResp, err := galaClient.GetPrice(ctx, "GTON/GALA")
		if err != nil {
			botErr := errors.ErrPriceFetchFailed("gswap", err)
			log.Printf("Price fetch error: %v", botErr)
			message += fmt.Sprintf("%s\n", botErr.GetUserMessage())
		} else {
			// Parse the price string to float
			priceFloat, err := strconv.ParseFloat(priceResp.Price, 64)
			if err != nil {
				log.Printf("Error parsing GalaChain price: %v", err)
				message += "Error parsing GTON/GALA price from gswap\n"
			} else {
				gswapPrice = priceFloat
				message += fmt.Sprintf("GTON/GALA on gswap: %.2f GALA", gswapPrice)

				// Add USD equivalent if available (GTON = TON)
				if tonUSD, tonExists := usdPrices["TON"]; tonExists {
					message += fmt.Sprintf(" (~$%.2f per GTON)", tonUSD)
				}
				message += "\n"
			}
		}
	} else {
		message += "GTON/GALA on gswap: N/A (GalaChain service not connected)\n"
	}

	// Calculate spread if both prices are available
	if stonfiPrice > 0 && gswapPrice > 0 {
		spread := (gswapPrice - stonfiPrice) / stonfiPrice * 100
		message += fmt.Sprintf("Spread: %.2f%%", spread)
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Error sending price message", nil)
		log.Printf("Error sending price message: %v", err)
	} else {
		// Log price check
		logger.InfoContext(ctx, "Price check completed", map[string]interface{}{
			"user_id":      update.Message.From.ID,
			"stonfi_price": stonfiPrice,
			"gswap_price":  gswapPrice,
			"ton_usd":      usdPrices["TON"],
			"gala_usd":     usdPrices["GALA"],
		})
	}
}
