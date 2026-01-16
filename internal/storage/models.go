package storage

import "time"

// UserSession represents a Telegram user session
type UserSession struct {
	UserID    int64     `json:"user_id"`
	ChatID    int64     `json:"chat_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"` // Session expires 24 hours after last activity
}

// WalletSession represents a connected wallet session
type WalletSession struct {
	UserID      int64     `json:"user_id"`
	WalletType  string    `json:"wallet_type"` // "ton" or "gala"
	Address     string    `json:"address"`
	ConnectedAt time.Time `json:"connected_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`

	// TonConnect session data (for TON wallets)
	TonConnectClientID  string `json:"tonconnect_client_id,omitempty"`  // Client public key (hex)
	TonConnectPrivateKey string `json:"tonconnect_private_key,omitempty"` // Encrypted client private key
	TonConnectWalletID   string `json:"tonconnect_wallet_id,omitempty"`   // Wallet public key (hex)
}
