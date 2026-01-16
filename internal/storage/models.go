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
	TonConnectClientID   string `json:"tonconnect_client_id,omitempty"`   // Client public key (hex)
	TonConnectPrivateKey string `json:"tonconnect_private_key,omitempty"` // Encrypted client private key
	TonConnectWalletID   string `json:"tonconnect_wallet_id,omitempty"`   // Wallet public key (hex)
}

// TradeHistory represents a completed or pending trade
type TradeHistory struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	TradeType       string     `json:"trade_type"` // "swap" or "arbitrage"
	Chain           string     `json:"chain"`      // "ton", "galachain", "both"
	FromToken       string     `json:"from_token"`
	ToToken         string     `json:"to_token"`
	AmountIn        string     `json:"amount_in"`  // Decimal as string
	AmountOut       string     `json:"amount_out"` // Decimal as string
	Fee             string     `json:"fee"`        // Decimal as string
	TxHashTon       string     `json:"tx_hash_ton"`
	TxHashGala      string     `json:"tx_hash_gala"`
	Status          string     `json:"status"` // "pending", "success", "failed", "partial"
	ErrorMessage    string     `json:"error_message"`
	ExecutionTimeMs int        `json:"execution_time_ms"`
	ProfitUSD       string     `json:"profit_usd"` // For arbitrage
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at"`
}
