package config

// Config holds the configuration for the telegram bot service
type Config struct {
	// Secrets (loaded from environment variables)
	BotToken      string
	EncryptionKey string

	// Non-secret configuration (loaded from YAML + env overrides)
	DatabaseURL         string
	GalaChainServiceURL string
	BotAdminUserIDs     string

	// Bot configuration (from YAML)
	Bot struct {
		PollingTimeoutSeconds int `mapstructure:"polling_timeout_seconds"`
		MaxConcurrentUpdates  int `mapstructure:"max_concurrent_updates"`
	}

	// Rate limiting configuration (from YAML)
	RateLimiting struct {
		CommandsPerMinute int `mapstructure:"commands_per_minute"`
		BurstSize         int `mapstructure:"burst_size"`
	} `mapstructure:"rate_limiting"`

	// Arbitrage configuration (from YAML)
	Arbitrage struct {
		SpreadThresholdPercent float64 `mapstructure:"spread_threshold_percent"`
		PositionSizePercent    float64 `mapstructure:"position_size_percent"`
		MinBalanceTon          float64 `mapstructure:"min_balance_ton"`
		MinBalanceGala         float64 `mapstructure:"min_balance_gala"`
	}

	// Wallet configuration (from YAML)
	Wallet struct {
		SessionExpiryHours int `mapstructure:"session_expiry_hours"`
	}

	// Trading configuration (from YAML)
	Trading struct {
		DefaultSlippageBps int `mapstructure:"default_slippage_bps"`
		MaxSlippageBps     int `mapstructure:"max_slippage_bps"`
	}

	// Monitoring configuration (from YAML)
	Monitoring struct {
		HealthCheckIntervalSeconds int  `mapstructure:"health_check_interval_seconds"`
		MetricsEnabled             bool `mapstructure:"metrics_enabled"`
	}

	// Logging configuration (from YAML)
	Logging struct {
		Level  string
		Format string
	}
}
