-- Trade history table for tracking swaps and arbitrage trades
CREATE TABLE IF NOT EXISTS trade_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    trade_type TEXT NOT NULL CHECK(trade_type IN ('swap', 'arbitrage')),
    chain TEXT NOT NULL,  -- 'ton', 'galachain', 'both'
    from_token TEXT NOT NULL,
    to_token TEXT NOT NULL,
    amount_in TEXT NOT NULL,  -- Store as string to preserve precision
    amount_out TEXT,  -- Store as string to preserve precision
    fee TEXT,  -- Store as string to preserve precision
    tx_hash_ton TEXT,
    tx_hash_gala TEXT,
    status TEXT NOT NULL CHECK(status IN ('pending', 'success', 'failed', 'partial')),
    error_message TEXT,
    execution_time_ms INTEGER,
    profit_usd TEXT,  -- For arbitrage trades, store as string
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_sessions(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_trade_history_user_id ON trade_history(user_id);
CREATE INDEX IF NOT EXISTS idx_trade_history_created_at ON trade_history(created_at);
CREATE INDEX IF NOT EXISTS idx_trade_history_status ON trade_history(status);
