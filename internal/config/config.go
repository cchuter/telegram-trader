package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Load reads configuration from YAML files with environment variable overrides
// Priority: 1. Default values (YAML), 2. Environment variables (override)
func Load() (*Config, error) {
	return LoadWithConfigPath("configs/bot-service.yaml")
}

// LoadWithConfigPath loads configuration from a specific YAML file path
func LoadWithConfigPath(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read YAML configuration file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Read the config file (ignore error if file doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		// If config file doesn't exist, use defaults + env vars
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Enable environment variable overrides
	v.SetEnvPrefix("")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Always load secrets from environment variables (never from YAML)
	cfg.BotToken = os.Getenv("BOT_TOKEN")
	cfg.EncryptionKey = os.Getenv("ENCRYPTION_KEY")

	// Load other configs with env var overrides
	if envDB := os.Getenv("DATABASE_URL"); envDB != "" {
		cfg.DatabaseURL = envDB
	}
	if envGala := os.Getenv("GALACHAIN_SERVICE_URL"); envGala != "" {
		cfg.GalaChainServiceURL = envGala
	}
	if envAdmins := os.Getenv("BOT_ADMIN_USER_IDS"); envAdmins != "" {
		cfg.BotAdminUserIDs = envAdmins
	}

	// Load logging config with env overrides
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.Logging.Level = envLogLevel
	}
	if envLogFormat := os.Getenv("LOG_FORMAT"); envLogFormat != "" {
		cfg.Logging.Format = envLogFormat
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Bot defaults
	v.SetDefault("bot.polling_timeout_seconds", 30)
	v.SetDefault("bot.max_concurrent_updates", 100)

	// Rate limiting defaults
	v.SetDefault("rate_limiting.commands_per_minute", 10)
	v.SetDefault("rate_limiting.burst_size", 5)

	// Arbitrage defaults
	v.SetDefault("arbitrage.spread_threshold_percent", 0.3)
	v.SetDefault("arbitrage.position_size_percent", 0.5)
	v.SetDefault("arbitrage.min_balance_ton", 1.0)
	v.SetDefault("arbitrage.min_balance_gala", 10.0)

	// Wallet defaults
	v.SetDefault("wallet.session_expiry_hours", 24)

	// Trading defaults
	v.SetDefault("trading.default_slippage_bps", 500)
	v.SetDefault("trading.max_slippage_bps", 1000)

	// Monitoring defaults
	v.SetDefault("monitoring.health_check_interval_seconds", 30)
	v.SetDefault("monitoring.metrics_enabled", true)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Non-secret config defaults (can be overridden by YAML or env)
	v.SetDefault("database_url", "telegram-trader.db")
	v.SetDefault("galachain_service_url", "localhost:50051")
	v.SetDefault("bot_admin_user_ids", "")
}
