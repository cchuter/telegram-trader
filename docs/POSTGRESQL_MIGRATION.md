# PostgreSQL Migration Guide

This document explains how to migrate from SQLite to PostgreSQL and use the PostgreSQL adapter.

## Overview

The telegram-trader bot now supports both SQLite and PostgreSQL databases:
- **SQLite**: Recommended for local development and testing
- **PostgreSQL**: Recommended for production deployments

The database type is automatically detected from the `DATABASE_URL` environment variable.

## Database URL Format

### SQLite
```bash
# Relative path
DATABASE_URL=telegram-trader.db

# Absolute path
DATABASE_URL=/var/lib/telegram-trader.db
```

### PostgreSQL
```bash
# Standard format
DATABASE_URL=postgres://username:password@hostname:port/database

# Example
DATABASE_URL=postgres://telegram_trader:mypassword@localhost:5432/telegram_trader

# With SSL disabled (for local development)
DATABASE_URL=postgres://telegram_trader:mypassword@localhost:5432/telegram_trader?sslmode=disable
```

## Quick Start with PostgreSQL

### 1. Using Docker Compose

The easiest way to run with PostgreSQL is using docker-compose:

```bash
# Set environment variables
export DATABASE_URL=postgres://telegram_trader:postgres@postgres:5432/telegram_trader
export BOT_TOKEN=your_bot_token
export ENCRYPTION_KEY=$(openssl rand -base64 32)

# Start services
cd deployments
docker-compose up -d
```

### 2. Using Local PostgreSQL

If you have PostgreSQL installed locally:

```bash
# Create database and user
psql -U postgres << EOF
CREATE USER telegram_trader WITH PASSWORD 'mypassword';
CREATE DATABASE telegram_trader OWNER telegram_trader;
GRANT ALL PRIVILEGES ON DATABASE telegram_trader TO telegram_trader;
EOF

# Set environment variable
export DATABASE_URL=postgres://telegram_trader:mypassword@localhost:5432/telegram_trader

# Run the bot
go run cmd/telegram-bot/main.go
```

## Migrating from SQLite to PostgreSQL

If you have existing data in SQLite and want to migrate to PostgreSQL:

### Step 1: Set up PostgreSQL database

Follow the "Quick Start" section above to create your PostgreSQL database.

### Step 2: Run migration script

```bash
go run scripts/migrate_sqlite_to_postgres.go \
  telegram-trader.db \
  postgres://telegram_trader:mypassword@localhost:5432/telegram_trader
```

The script will:
1. Connect to both databases
2. Migrate all user sessions
3. Migrate all wallet sessions (including TonConnect fields)
4. Display migration statistics

### Step 3: Update environment variable

```bash
# Update your .env file or environment
DATABASE_URL=postgres://telegram_trader:mypassword@localhost:5432/telegram_trader

# Restart the bot
```

## Testing PostgreSQL Implementation

### Run tests with a live PostgreSQL instance

```bash
# Using docker
./scripts/test_postgres.sh

# Or manually
export DATABASE_URL=postgres://user:pass@localhost:5432/testdb
go test ./internal/storage -v
```

### Run tests without PostgreSQL (unit tests only)

```bash
go test ./internal/storage -v -short
```

## Schema Management

The PostgreSQL adapter automatically manages schema migrations:

1. **schema_migrations table**: Tracks which migrations have been applied
2. **Idempotent migrations**: Can be run multiple times safely
3. **Transactional migrations**: Each migration runs in a transaction

Migrations are applied automatically when the application starts.

## Differences from SQLite

| Feature | SQLite | PostgreSQL |
|---------|--------|------------|
| Connection pool | 1 connection | 25 max connections |
| Concurrent writes | Limited | Full support |
| Data types | INTEGER for booleans | BOOLEAN native type |
| AUTO_INCREMENT | AUTOINCREMENT | SERIAL |
| Parameterized queries | `?` | `$1, $2, ...` |

## Performance Considerations

### SQLite
- Perfect for development
- Single writer limitation
- File-based, no network overhead
- Limited to ~100K queries/second

### PostgreSQL
- Designed for production
- Multiple concurrent writers
- Network overhead (local or remote)
- Scales to millions of queries/second
- Better for high-traffic bots

## Troubleshooting

### Connection errors

```
Error: failed to ping database: dial tcp: connection refused
```

**Solution**: Ensure PostgreSQL is running and accessible:
```bash
psql -h localhost -U telegram_trader -d telegram_trader
```

### Migration errors

```
Error: relation "user_sessions" already exists
```

**Solution**: Migrations are idempotent. This is normal if the schema already exists.

### SSL errors

```
Error: pq: SSL is not enabled on the server
```

**Solution**: Add `?sslmode=disable` to your DATABASE_URL for local development:
```bash
DATABASE_URL=postgres://user:pass@localhost:5432/db?sslmode=disable
```

## Production Checklist

Before deploying to production with PostgreSQL:

- [ ] Enable SSL/TLS connections
- [ ] Use strong passwords
- [ ] Configure connection pooling (already set to 25 max connections)
- [ ] Set up automated backups
- [ ] Monitor connection pool usage
- [ ] Configure PostgreSQL authentication (pg_hba.conf)
- [ ] Set appropriate PostgreSQL performance settings
- [ ] Use a secrets manager for DATABASE_URL

## Backup and Restore

### Backup PostgreSQL database

```bash
pg_dump -U telegram_trader telegram_trader > backup.sql
```

### Restore PostgreSQL database

```bash
psql -U telegram_trader telegram_trader < backup.sql
```

### Backup SQLite database

```bash
cp telegram-trader.db telegram-trader-backup.db
```

## Further Reading

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [lib/pq Driver](https://github.com/lib/pq)
- [Docker Compose PostgreSQL](https://hub.docker.com/_/postgres)
