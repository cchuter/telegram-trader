package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cchuter/telegram-trader/internal/blockchain/ton"
	"github.com/cchuter/telegram-trader/internal/bot"
	"github.com/cchuter/telegram-trader/internal/config"
	"github.com/cchuter/telegram-trader/internal/dex/stonfi"
	"github.com/cchuter/telegram-trader/internal/galachain"
	"github.com/cchuter/telegram-trader/internal/logging"
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
	if cfg.EncryptionKey == "" {
		log.Fatal("ENCRYPTION_KEY environment variable is required")
	}

	// Initialize database
	db, err := storage.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Database initialized successfully")

	// Initialize logger
	logger := logging.New("telegram-bot", logging.LogLevelInfo)
	logger.Info("Logger initialized", nil)

	// Initialize trade logger
	tradeLogger, err := logging.NewTradeLogger("./logs")
	if err != nil {
		log.Fatalf("Failed to initialize trade logger: %v", err)
	}
	defer tradeLogger.Close()

	logger.Info("Trade logger initialized", map[string]interface{}{
		"log_dir": "./logs",
	})

	// Initialize TON blockchain client
	tonClient := ton.NewClient()
	if err := tonClient.Connect(context.Background()); err != nil {
		log.Fatalf("Failed to connect to TON blockchain: %v", err)
	}
	defer tonClient.Close()

	logger.Info("TON blockchain client initialized", nil)
	log.Println("TON blockchain client initialized successfully")

	// Initialize DEX client (ston.fi)
	dexClient := stonfi.NewClient()
	defer dexClient.Close()

	logger.Info("DEX client initialized", nil)
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
			logger.Warn("Failed to connect to GalaChain service, continuing without it", map[string]interface{}{
				"error": err.Error(),
			})
			log.Printf("Warning: Failed to connect to GalaChain service: %v", err)
			log.Println("Continuing without GalaChain service...")
		} else {
			defer galaClient.Close()
			logger.Info("GalaChain client initialized", map[string]interface{}{
				"service_url": cfg.GalaChainServiceURL,
			})
			log.Println("GalaChain client initialized successfully")
		}
	} else {
		logger.Info("GALACHAIN_SERVICE_URL not set, continuing without GalaChain service", nil)
		log.Println("GALACHAIN_SERVICE_URL not set, continuing without GalaChain service")
	}

	// Initialize health checker
	healthChecker := bot.NewHealthChecker(db, tonClient, galaClient)

	// Start health check HTTP server in a goroutine
	go func() {
		logger.Info("Starting health check server on port 8080", nil)
		log.Println("Health check server starting on port 8080")
		if err := healthChecker.StartHealthServer("8080"); err != nil {
			logger.LogError(ctx, 0, "system", err, "Health check server failed", nil)
			log.Printf("Health check server error: %v", err)
		}
	}()

	// Create and start the bot
	b, err := bot.New(cfg.BotToken, db, cfg.BotAdminUserIDs, cfg.EncryptionKey, dexClient, galaClient, logger, tradeLogger)
	if err != nil {
		logger.Fatal("Failed to create bot", err, nil)
		log.Fatalf("Failed to create bot: %v", err)
	}
	logger.Info("Starting Telegram bot", map[string]interface{}{
		"admin_user_ids": cfg.BotAdminUserIDs,
	})
	if err := b.Start(ctx); err != nil {
		logger.Fatal("Failed to start bot", err, nil)
		log.Fatalf("Failed to start bot: %v", err)
	}
}
