-- User sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    user_id INTEGER PRIMARY KEY,
    chat_id INTEGER NOT NULL,
    username TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT (datetime(CURRENT_TIMESTAMP, '+24 hours'))
);

-- Wallet sessions table
CREATE TABLE IF NOT EXISTS wallet_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    wallet_type TEXT NOT NULL CHECK(wallet_type IN ('ton', 'gala')),
    address TEXT NOT NULL,
    connected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active INTEGER NOT NULL DEFAULT 1,
    UNIQUE(user_id, wallet_type),
    FOREIGN KEY (user_id) REFERENCES user_sessions(user_id) ON DELETE CASCADE
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_wallet_sessions_user_id ON wallet_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_sessions_wallet_type ON wallet_sessions(wallet_type);
CREATE INDEX IF NOT EXISTS idx_wallet_sessions_is_active ON wallet_sessions(is_active);
