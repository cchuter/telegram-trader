-- Add TonConnect session fields to wallet_sessions table
ALTER TABLE wallet_sessions ADD COLUMN tonconnect_client_id TEXT;
ALTER TABLE wallet_sessions ADD COLUMN tonconnect_private_key TEXT;
ALTER TABLE wallet_sessions ADD COLUMN tonconnect_wallet_id TEXT;
