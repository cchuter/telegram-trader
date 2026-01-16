package bot

import (
	"context"
	"log"

	"github.com/cchuter/telegram-trader/internal/bot/handlers"
	"github.com/cchuter/telegram-trader/internal/bot/middleware"
	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Bot represents the Telegram bot instance
type Bot struct {
	token          string
	bot            *bot.Bot
	db             storage.Database
	authMiddleware *middleware.AuthMiddleware
}

// New creates a new Bot instance
func New(token string, db storage.Database, adminUserIDs string) *Bot {
	return &Bot{
		token:          token,
		db:             db,
		authMiddleware: middleware.NewAuthMiddleware(db, adminUserIDs),
	}
}

// Start initializes the bot and starts polling for updates
func (b *Bot) Start(ctx context.Context) error {
	opts := []bot.Option{
		bot.WithDefaultHandler(b.defaultHandler),
	}

	botInstance, err := bot.New(b.token, opts...)
	if err != nil {
		return err
	}

	b.bot = botInstance

	// Register command handlers
	b.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, b.handleStart)
	b.bot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, b.handleHelp)
	b.bot.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeExact, b.handleBalance)

	log.Println("Bot started successfully")
	b.bot.Start(ctx)

	return nil
}

// defaultHandler handles messages that don't match any specific handler
func (b *Bot) defaultHandler(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Default handler - does nothing for now
}

// handleStart handles the /start command
func (b *Bot) handleStart(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Authenticate user
	if err := b.authMiddleware.Authenticate(ctx, botInstance, update); err != nil {
		log.Printf("Authentication failed for user: %v", err)
		return
	}

	message := `🚀 Welcome to GalaSwap & STON.fi Trading Bot!

Trade tokens on GalaChain and TON blockchain with ease.

Available Commands:
/wallet - Connect your TonKeeper wallet
/balance - View your balances
/swap - Execute a token swap
/price - Check token prices
/alert - Set price alerts
/portfolio - View your portfolio
/orders - Manage automated orders
/help - Show all commands

Get started by connecting your wallet with /wallet`

	_, err := botInstance.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		log.Printf("Error sending start message: %v", err)
	}
}

// handleHelp handles the /help command
func (b *Bot) handleHelp(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Authenticate user
	if err := b.authMiddleware.Authenticate(ctx, botInstance, update); err != nil {
		log.Printf("Authentication failed for user: %v", err)
		return
	}

	message := `Available Commands:

/wallet - Connect your TonKeeper wallet
/balance - View your balances
/swap - Execute a token swap
/price - Check token prices
/alert - Set price alerts
/portfolio - View your portfolio
/orders - Manage automated orders
/help - Show all commands

For more information about a command, simply type it in the chat.`

	_, err := botInstance.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		log.Printf("Error sending help message: %v", err)
	}
}

// handleBalance handles the /balance command
func (b *Bot) handleBalance(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Authenticate user
	if err := b.authMiddleware.Authenticate(ctx, botInstance, update); err != nil {
		log.Printf("Authentication failed for user: %v", err)
		return
	}

	// Call the handler
	handlers.HandleBalance(ctx, botInstance, update)
}
