package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cchuter/telegram-trader/internal/bot"
)

func main() {
	// Get bot token from environment variable
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}

	// Create context that listens for interrupt signals
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create and start the bot
	b := bot.New(token)
	if err := b.Start(ctx); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
