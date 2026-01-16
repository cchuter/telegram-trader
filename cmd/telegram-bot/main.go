package main

import (
	"context"
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
	// Initialize logger first (before config validation)
	logger := logging.New("telegram-bot", logging.LogLevelInfo)

	// Load configuration
	cfg := config.Load()

	// Validate required configuration
	if cfg.BotToken == "" {
		logger.Fatal("BOT_TOKEN environment variable is required", nil, nil)
	}
	if cfg.DatabaseURL == "" {
		logger.Fatal("DATABASE_URL environment variable is required", nil, nil)
	}
	if cfg.EncryptionKey == "" {
		logger.Fatal("ENCRYPTION_KEY environment variable is required", nil, nil)
	}

	// Initialize database (auto-detect SQLite vs PostgreSQL from URL)
	db, err := storage.ParseDatabaseURL(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to initialize database", err, nil)
	}
	defer db.Close()

	logger.Info("Database initialized successfully", nil)

	// Initialize trade logger
	tradeLogger, err := logging.NewTradeLogger("./logs")
	if err != nil {
		logger.Fatal("Failed to initialize trade logger", err, nil)
	}
	defer tradeLogger.Close()

	logger.Info("Trade logger initialized", map[string]interface{}{
		"log_dir": "./logs",
	})

	// Initialize TON blockchain client
	tonClient := ton.NewClient()
	if err := tonClient.Connect(context.Background()); err != nil {
		logger.Fatal("Failed to connect to TON blockchain", err, nil)
	}
	defer tonClient.Close()

	logger.Info("TON blockchain client initialized", nil)

	// Initialize DEX client (ston.fi)
	dexClient := stonfi.NewClient()
	defer dexClient.Close()

	logger.Info("DEX client initialized", nil)

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
		} else {
			defer galaClient.Close()
			logger.Info("GalaChain client initialized", map[string]interface{}{
				"service_url": cfg.GalaChainServiceURL,
			})
		}
	} else {
		logger.Info("GALACHAIN_SERVICE_URL not set, continuing without GalaChain service", nil)
	}

	// Initialize health checker
	healthChecker := bot.NewHealthChecker(db, tonClient, galaClient)

	// Start health check HTTP server in a goroutine
	go func() {
		logger.Info("Starting health check server on port 8080", nil)
		if err := healthChecker.StartHealthServer("8080"); err != nil {
			logger.LogError(ctx, 0, "system", err, "Health check server failed", nil)
		}
	}()

	// Create and start the bot
	b, err := bot.New(cfg.BotToken, db, cfg.BotAdminUserIDs, cfg.EncryptionKey, dexClient, galaClient, logger, tradeLogger)
	if err != nil {
		logger.Fatal("Failed to create bot", err, nil)
	}
	logger.Info("Starting Telegram bot", map[string]interface{}{
		"admin_user_ids": cfg.BotAdminUserIDs,
	})
	if err := b.Start(ctx); err != nil {
		logger.Fatal("Failed to start bot", err, nil)
	}
}
