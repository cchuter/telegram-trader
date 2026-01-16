package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cchuter/telegram-trader/internal/bot"
	"github.com/cchuter/telegram-trader/internal/config"
	"github.com/cchuter/telegram-trader/internal/dex/stonfi"
	"github.com/cchuter/telegram-trader/internal/galachain"
	"github.com/cchuter/telegram-trader/internal/storage"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate required configuration
	if cfg.BotToken == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Initialize database
	db, err := storage.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Database initialized successfully")

	// Initialize DEX client (ston.fi)
	dexClient := stonfi.NewClient()
	defer dexClient.Close()

	log.Println("DEX client initialized successfully")

	// Create context that listens for interrupt signals
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Initialize GalaChain client if service URL is configured
	var galaClient *galachain.Client
	if cfg.GalaChainServiceURL != "" {
		var err error
		galaClient, err = galachain.Connect(ctx, cfg.GalaChainServiceURL)
		if err != nil {
			log.Printf("Warning: Failed to connect to GalaChain service: %v", err)
			log.Println("Continuing without GalaChain service...")
		} else {
			defer galaClient.Close()
			log.Println("GalaChain client initialized successfully")
		}
	} else {
		log.Println("GALACHAIN_SERVICE_URL not set, continuing without GalaChain service")
	}

	// Create and start the bot
	b := bot.New(cfg.BotToken, db, cfg.BotAdminUserIDs, dexClient, galaClient)
	if err := b.Start(ctx); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
