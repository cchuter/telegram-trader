package config

// Config holds the configuration for the telegram bot service
type Config struct {
	BotToken            string
	DatabaseURL         string
	EncryptionKey       string
	GalaChainServiceURL string
	BotAdminUserIDs     string
}
