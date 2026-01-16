package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/galachain"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	// OneTON is 1 TON in nanotons (1 TON = 1,000,000,000 nanotons)
	OneTON = "1000000000"
)

// HandlePrice handles the /price command
func HandlePrice(ctx context.Context, b *bot.Bot, update *models.Update, dexClient dex.Client, galaClient *galachain.Client) {
	var message string

	// Fetch TON/GALA price from ston.fi
	var stonfiPrice float64
	if dexClient != nil {
		sim, err := dexClient.SimulateSwap(ctx, "TON", GALATokenAddress, OneTON)
		if err != nil {
			log.Printf("Error fetching ston.fi price: %v", err)
			message = "Error fetching TON/GALA price from ston.fi\n"
		} else {
			// Convert output amount to GALA (assuming 9 decimals for GALA)
			galaAmount, err := strconv.ParseFloat(sim.OutputAmount, 64)
			if err != nil {
				log.Printf("Error parsing ston.fi output amount: %v", err)
				message = "Error parsing TON/GALA price from ston.fi\n"
			} else {
				stonfiPrice = galaAmount / 1e9 // Convert from smallest units to GALA
				message = fmt.Sprintf("TON/GALA on ston.fi: %.2f GALA\n", stonfiPrice)
			}
		}
	} else {
		message = "TON/GALA on ston.fi: N/A (DEX client not available)\n"
	}

	// Fetch GTON/GALA price from GalaChain service
	var gswapPrice float64
	if galaClient != nil {
		price, err := galaClient.GetPrice(ctx, "GTON/GALA")
		if err != nil {
			log.Printf("Error fetching GalaChain price: %v", err)
			message += "Error fetching GTON/GALA price from gswap\n"
		} else {
			// Parse the price string to float
			priceFloat, err := strconv.ParseFloat(price.Price, 64)
			if err != nil {
				log.Printf("Error parsing GalaChain price: %v", err)
				message += "Error parsing GTON/GALA price from gswap\n"
			} else {
				gswapPrice = priceFloat
				message += fmt.Sprintf("GTON/GALA on gswap: %.2f GALA\n", gswapPrice)
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
		log.Printf("Error sending price message: %v", err)
	}
}
