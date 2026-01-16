package config

import "os"

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		BotToken:            os.Getenv("BOT_TOKEN"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		EncryptionKey:       os.Getenv("ENCRYPTION_KEY"),
		GalaChainServiceURL: os.Getenv("GALACHAIN_SERVICE_URL"),
		BotAdminUserIDs:     os.Getenv("BOT_ADMIN_USER_IDS"),
	}
}
