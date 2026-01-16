package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cchuter/telegram-trader/internal/bot"
	"github.com/cchuter/telegram-trader/internal/config"
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

	// Create context that listens for interrupt signals
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create and start the bot
	b := bot.New(cfg.BotToken, db, cfg.BotAdminUserIDs)
	if err := b.Start(ctx); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
