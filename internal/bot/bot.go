package bot

import (
	"context"
	"log"
	"time"

	"github.com/cchuter/telegram-trader/internal/arbitrage"
	"github.com/cchuter/telegram-trader/internal/bot/handlers"
	"github.com/cchuter/telegram-trader/internal/bot/middleware"
	"github.com/cchuter/telegram-trader/internal/dex"
	"github.com/cchuter/telegram-trader/internal/galachain"
	"github.com/cchuter/telegram-trader/internal/logging"
	"github.com/cchuter/telegram-trader/internal/storage"
	"github.com/cchuter/telegram-trader/internal/wallet"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Bot represents the Telegram bot instance
type Bot struct {
	token             string
	bot               *bot.Bot
	db                storage.Database
	authMiddleware    *middleware.AuthMiddleware
	rateLimiter       *middleware.RateLimiter
	walletManager     *wallet.Manager
	dexClient         dex.Client
	galaClient        *galachain.Client
	arbitrageEngine   *arbitrage.Engine
	arbitrageExecutor *arbitrage.Executor
	logger            *logging.Logger
	tradeLogger       *logging.TradeLogger
}

// New creates a new Bot instance
func New(token string, db storage.Database, adminUserIDs string, encryptionKey string, dexClient dex.Client, galaClient *galachain.Client, logger *logging.Logger, tradeLogger *logging.TradeLogger) (*Bot, error) {
	// Initialize wallet manager with encryption
	walletManager, err := wallet.NewManager(db, encryptionKey)
	if err != nil {
		return nil, err
	}

	engine := arbitrage.NewEngine(dexClient, galaClient)

	return &Bot{
		token:             token,
		db:                db,
		authMiddleware:    middleware.NewAuthMiddleware(db, adminUserIDs),
		rateLimiter:       middleware.NewRateLimiter(10, 1*time.Minute),
		walletManager:     walletManager,
		dexClient:         dexClient,
		galaClient:        galaClient,
		arbitrageEngine:   engine,
		arbitrageExecutor: arbitrage.NewExecutor(nil, nil, engine), // TODO: Pass actual clients in production
		logger:            logger,
		tradeLogger:       tradeLogger,
	}, nil
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
	b.bot.RegisterHandler(bot.HandlerTypeMessageText, "/wallet", bot.MatchTypePrefix, b.handleWallet)
	b.bot.RegisterHandler(bot.HandlerTypeMessageText, "/swap", bot.MatchTypePrefix, b.handleSwap)
	b.bot.RegisterHandler(bot.HandlerTypeMessageText, "/price", bot.MatchTypeExact, b.handlePrice)
	b.bot.RegisterHandler(bot.HandlerTypeMessageText, "/arbitrage", bot.MatchTypeExact, b.handleArbitrage)

	// Register callback handlers
	b.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "swap_", bot.MatchTypePrefix, b.handleSwapCallback)
	b.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "arbitrage_", bot.MatchTypePrefix, b.handleArbitrageCallback)

	b.logger.Info("Bot started successfully", nil)
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
	// Generate correlation ID for this command
	correlationID := logging.GenerateCorrelationID()
	ctx = logging.ContextWithCorrelationID(ctx, correlationID)

	// Apply rate limiting
	if err := b.rateLimiter.Middleware(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Rate limit exceeded", nil)
		log.Printf("Rate limit exceeded for user: %v", err)
		return
	}

	// Authenticate user
	if err := b.authMiddleware.Authenticate(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Authentication failed", nil)
		log.Printf("Authentication failed for user: %v", err)
		return
	}

	// Log command execution
	b.logger.LogCommand(ctx, update.Message.From.ID, update.Message.From.Username, "/start", nil)

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
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Error sending start message", nil)
		log.Printf("Error sending start message: %v", err)
	}
}

// handleHelp handles the /help command
func (b *Bot) handleHelp(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Generate correlation ID for this command
	correlationID := logging.GenerateCorrelationID()
	ctx = logging.ContextWithCorrelationID(ctx, correlationID)

	// Apply rate limiting
	if err := b.rateLimiter.Middleware(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Rate limit exceeded", nil)
		log.Printf("Rate limit exceeded for user: %v", err)
		return
	}

	// Authenticate user
	if err := b.authMiddleware.Authenticate(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Authentication failed", nil)
		log.Printf("Authentication failed for user: %v", err)
		return
	}

	// Log command execution
	b.logger.LogCommand(ctx, update.Message.From.ID, update.Message.From.Username, "/help", nil)

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
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Error sending help message", nil)
		log.Printf("Error sending help message: %v", err)
	}
}

// handleBalance handles the /balance command
func (b *Bot) handleBalance(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Generate correlation ID for this command
	correlationID := logging.GenerateCorrelationID()
	ctx = logging.ContextWithCorrelationID(ctx, correlationID)

	// Apply rate limiting
	if err := b.rateLimiter.Middleware(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Rate limit exceeded", nil)
		log.Printf("Rate limit exceeded for user: %v", err)
		return
	}

	// Authenticate user
	if err := b.authMiddleware.Authenticate(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Authentication failed", nil)
		log.Printf("Authentication failed for user: %v", err)
		return
	}

	// Log command execution
	b.logger.LogCommand(ctx, update.Message.From.ID, update.Message.From.Username, "/balance", nil)

	// Call the handler
	handlers.HandleBalance(ctx, botInstance, update, b.galaClient, b.logger)
}

// handleWallet handles the /wallet command
func (b *Bot) handleWallet(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Generate correlation ID for this command
	correlationID := logging.GenerateCorrelationID()
	ctx = logging.ContextWithCorrelationID(ctx, correlationID)

	// Apply rate limiting
	if err := b.rateLimiter.Middleware(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Rate limit exceeded", nil)
		log.Printf("Rate limit exceeded for user: %v", err)
		return
	}

	// Authenticate user
	if err := b.authMiddleware.Authenticate(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Authentication failed", nil)
		log.Printf("Authentication failed for user: %v", err)
		return
	}

	// Log command execution
	b.logger.LogCommand(ctx, update.Message.From.ID, update.Message.From.Username, "/wallet", nil)

	// Call the handler
	handlers.HandleWallet(ctx, botInstance, update, b.walletManager, b.logger)
}

// handleSwap handles the /swap command
func (b *Bot) handleSwap(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Generate correlation ID for this command
	correlationID := logging.GenerateCorrelationID()
	ctx = logging.ContextWithCorrelationID(ctx, correlationID)

	// Apply rate limiting
	if err := b.rateLimiter.Middleware(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Rate limit exceeded", nil)
		log.Printf("Rate limit exceeded for user: %v", err)
		return
	}

	// Authenticate user
	if err := b.authMiddleware.Authenticate(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Authentication failed", nil)
		log.Printf("Authentication failed for user: %v", err)
		return
	}

	// Log command execution
	b.logger.LogCommand(ctx, update.Message.From.ID, update.Message.From.Username, "/swap", map[string]interface{}{
		"full_command": update.Message.Text,
	})

	// Call the handler
	handlers.HandleSwap(ctx, botInstance, update, b.dexClient, b.logger, b.tradeLogger)
}

// handlePrice handles the /price command
func (b *Bot) handlePrice(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Generate correlation ID for this command
	correlationID := logging.GenerateCorrelationID()
	ctx = logging.ContextWithCorrelationID(ctx, correlationID)

	// Apply rate limiting
	if err := b.rateLimiter.Middleware(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Rate limit exceeded", nil)
		log.Printf("Rate limit exceeded for user: %v", err)
		return
	}

	// Authenticate user
	if err := b.authMiddleware.Authenticate(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Authentication failed", nil)
		log.Printf("Authentication failed for user: %v", err)
		return
	}

	// Log command execution
	b.logger.LogCommand(ctx, update.Message.From.ID, update.Message.From.Username, "/price", nil)

	// Call the handler
	handlers.HandlePrice(ctx, botInstance, update, b.dexClient, b.galaClient, b.logger)
}

// handleArbitrage handles the /arbitrage command
func (b *Bot) handleArbitrage(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Generate correlation ID for this command
	correlationID := logging.GenerateCorrelationID()
	ctx = logging.ContextWithCorrelationID(ctx, correlationID)

	// Apply rate limiting
	if err := b.rateLimiter.Middleware(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Rate limit exceeded", nil)
		log.Printf("Rate limit exceeded for user: %v", err)
		return
	}

	// Authenticate user
	if err := b.authMiddleware.Authenticate(ctx, botInstance, update); err != nil {
		b.logger.LogError(ctx, update.Message.From.ID, update.Message.From.Username, err, "Authentication failed", nil)
		log.Printf("Authentication failed for user: %v", err)
		return
	}

	// Log command execution
	b.logger.LogCommand(ctx, update.Message.From.ID, update.Message.From.Username, "/arbitrage", nil)

	// Call the handler
	handlers.HandleArbitrage(ctx, botInstance, update, b.arbitrageEngine, b.logger)
}

// handleSwapCallback handles callback queries from swap confirmation buttons
func (b *Bot) handleSwapCallback(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Callback queries don't have Message, they have CallbackQuery
	if update.CallbackQuery == nil {
		return
	}

	// Generate correlation ID for this command
	correlationID := logging.GenerateCorrelationID()
	ctx = logging.ContextWithCorrelationID(ctx, correlationID)

	// Apply rate limiting (use CallbackQuery.From instead of Message.From)
	// For callbacks, we need to construct a temporary update for middleware
	// For now, skip middleware for callbacks and implement direct handling

	// Log callback execution
	b.logger.LogCommand(ctx, update.CallbackQuery.From.ID, update.CallbackQuery.From.Username, "swap_callback", map[string]interface{}{
		"callback_data": update.CallbackQuery.Data,
	})

	// Call the handler
	handlers.HandleSwapCallback(ctx, botInstance, update, b.dexClient, b.walletManager, b.logger, b.tradeLogger)
}

// handleArbitrageCallback handles callback queries from arbitrage confirmation buttons
func (b *Bot) handleArbitrageCallback(ctx context.Context, botInstance *bot.Bot, update *models.Update) {
	// Callback queries don't have Message, they have CallbackQuery
	if update.CallbackQuery == nil {
		return
	}

	// Generate correlation ID for this command
	correlationID := logging.GenerateCorrelationID()
	ctx = logging.ContextWithCorrelationID(ctx, correlationID)

	// Log callback execution
	b.logger.LogCommand(ctx, update.CallbackQuery.From.ID, update.CallbackQuery.From.Username, "arbitrage_callback", map[string]interface{}{
		"callback_data": update.CallbackQuery.Data,
	})

	// Call the handler
	handlers.HandleArbitrageCallback(ctx, botInstance, update, b.arbitrageExecutor, b.arbitrageEngine, b.logger, b.tradeLogger)
}
