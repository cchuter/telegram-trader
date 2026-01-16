-- Add expires_at column to user_sessions table
ALTER TABLE user_sessions ADD COLUMN expires_at TIMESTAMP NOT NULL DEFAULT (datetime('now', '+24 hours'));
