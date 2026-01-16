# Deployment Guide

Complete guide for deploying the Telegram Trader bot in various environments.

## Table of Contents

- [Quick Start](#quick-start)
- [Environment Variables](#environment-variables)
- [Docker Compose Deployment](#docker-compose-deployment)
- [Database Setup](#database-setup)
- [Backup and Restore](#backup-and-restore)
- [Monitoring and Health Checks](#monitoring-and-health-checks)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)

## Quick Start

Get the bot running in under 5 minutes with Docker Compose:

```bash
# 1. Clone repository
git clone <repository-url>
cd telegram-trader

# 2. Configure environment
cp .env.example .env
# Edit .env with your values (see Environment Variables section)

# 3. Start services
docker-compose -f deployments/docker-compose.yml up -d

# 4. Verify deployment
docker-compose -f deployments/docker-compose.yml ps
docker-compose -f deployments/docker-compose.yml logs -f bot-service
```

The bot is now running and connected to Telegram!

## Environment Variables

### Required Variables

All environments require these variables configured in `.env` file at project root:

#### Bot Service

| Variable | Description | Example | Generation |
|----------|-------------|---------|------------|
| `BOT_TOKEN` | Telegram bot token from @BotFather | `1234567890:ABCdef...` | Get from [@BotFather](https://t.me/botfather) |
| `BOT_ADMIN_USER_IDS` | Comma-separated list of authorized Telegram user IDs | `123456789,987654321` | Get your ID from [@userinfobot](https://t.me/userinfobot) |
| `ENCRYPTION_KEY` | Base64-encoded 32-byte AES-256 key | `YourBase64Key...` | Generate: `openssl rand -base64 32` |

#### Database Configuration

| Variable | Description | Default | Notes |
|----------|-------------|---------|-------|
| `DATABASE_URL` | Database connection string | `telegram-trader.db` | SQLite: `file.db`<br>PostgreSQL: `postgres://user:pass@host:5432/dbname` |
| `POSTGRES_USER` | PostgreSQL username | `telegram_trader` | Required for Docker Compose |
| `POSTGRES_PASSWORD` | PostgreSQL password | `postgres` | **Change in production!** |
| `POSTGRES_DB` | PostgreSQL database name | `telegram_trader` | Database will be created automatically |

#### Service URLs

| Variable | Description | Default | Docker Value |
|----------|-------------|---------|--------------|
| `GALACHAIN_SERVICE_URL` | gRPC address of GalaChain service | `localhost:50051` | `galachain-service:50051` |
| `GSWAP_API_URL` | GalaChain GSwap API endpoint | `https://swap.gala.com/api` | Leave as default |

#### Optional Variables

| Variable | Description | Default | Options |
|----------|-------------|---------|---------|
| `LOG_FORMAT` | Logging output format | `json` | `json` (production), `text` (development) |

### GalaChain Service Variables

Create `galachain-service/.env` with these variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server port | `50051` |
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://user:pass@postgres:5432/telegram_trader` |
| `ENCRYPTION_KEY` | 64-character hex encryption key | `0123456789abcdef...` |
| `GSWAP_API_URL` | GSwap API URL | `https://swap.gala.com/api` |

### Example .env File

```env
# Telegram Configuration
BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
BOT_ADMIN_USER_IDS=123456789

# Database
DATABASE_URL=postgres://telegram_trader:your_secure_password@postgres:5432/telegram_trader
POSTGRES_USER=telegram_trader
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=telegram_trader

# Security
ENCRYPTION_KEY=$(openssl rand -base64 32)

# Service URLs
GALACHAIN_SERVICE_URL=galachain-service:50051
GSWAP_API_URL=https://swap.gala.com/api

# Logging
LOG_FORMAT=json
```

## Docker Compose Deployment

### Architecture Overview

The Docker Compose setup includes three services:

1. **PostgreSQL Database** - Persistent data storage
2. **GalaChain Service** - TypeScript gRPC service for GalaChain/GSwap
3. **Bot Service** - Go Telegram bot

```
┌─────────────────┐
│  Bot Service    │◄──── Telegram API
│   (Go)          │
└────────┬────────┘
         │ gRPC
         ▼
┌─────────────────┐
│ GalaChain Svc   │◄──── GSwap API
│  (TypeScript)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   PostgreSQL    │
└─────────────────┘
```

### Starting Services

```bash
# Start all services in background
docker-compose -f deployments/docker-compose.yml up -d

# View logs
docker-compose -f deployments/docker-compose.yml logs -f

# View specific service logs
docker-compose -f deployments/docker-compose.yml logs -f bot-service
docker-compose -f deployments/docker-compose.yml logs -f galachain-service
```

### Stopping Services

```bash
# Stop all services (keeps data)
docker-compose -f deployments/docker-compose.yml stop

# Stop and remove containers (keeps volumes)
docker-compose -f deployments/docker-compose.yml down

# Stop and remove everything including volumes (⚠️ deletes data!)
docker-compose -f deployments/docker-compose.yml down -v
```

### Rebuilding Services

After code changes:

```bash
# Rebuild and restart specific service
docker-compose -f deployments/docker-compose.yml up -d --build bot-service

# Rebuild all services
docker-compose -f deployments/docker-compose.yml up -d --build
```

### Viewing Service Status

```bash
# Check service health
docker-compose -f deployments/docker-compose.yml ps

# Expected output:
# NAME                      STATUS
# telegram-trader-bot       Up (healthy)
# telegram-trader-galachain Up (healthy)
# telegram-trader-postgres  Up (healthy)
```

### Scaling (Future)

Currently, services are stateful and single-instance. For production scaling:

```bash
# Scale bot service (requires shared database and session management)
docker-compose -f deployments/docker-compose.yml up -d --scale bot-service=3
```

**Note**: Bot service scaling requires webhook mode instead of polling. See [Telegram Bot API documentation](https://core.telegram.org/bots/api#setwebhook).

## Database Setup

### SQLite (Development)

SQLite is automatically initialized on first run:

```bash
# File-based database
DATABASE_URL=telegram-trader.db

# Bot creates database and runs migrations automatically
go run ./cmd/telegram-bot
```

**Pros**: Zero configuration, perfect for development
**Cons**: Single-writer, not suitable for scaled deployments

### PostgreSQL (Staging/Production)

#### Automatic Setup (Docker Compose)

PostgreSQL is automatically initialized via Docker Compose:

```yaml
# deployments/docker-compose.yml
postgres:
  image: postgres:15-alpine
  environment:
    POSTGRES_USER: telegram_trader
    POSTGRES_PASSWORD: postgres
    POSTGRES_DB: telegram_trader
  volumes:
    - postgres-data:/var/lib/postgresql/data
    - ../migrations:/docker-entrypoint-initdb.d:ro  # Auto-run migrations
```

Migrations in `migrations/*.sql` are automatically executed on first startup.

#### Manual Setup

If running PostgreSQL separately:

```bash
# 1. Create database
createdb telegram_trader

# 2. Create user
psql postgres -c "CREATE USER telegram_trader WITH PASSWORD 'your_password';"

# 3. Grant permissions
psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE telegram_trader TO telegram_trader;"

# 4. Run migrations
psql -U telegram_trader -d telegram_trader -f migrations/001_initial_schema.up.sql

# 5. Configure connection
export DATABASE_URL="postgres://telegram_trader:your_password@localhost:5432/telegram_trader"
```

### Database Migrations

Migrations are located in `migrations/` directory and versioned:

```
migrations/
├── 001_initial_schema.up.sql     # User and wallet sessions tables
├── 002_session_expiry.up.sql     # Add expires_at column
├── 003_postgres_expiry.up.sql    # PostgreSQL-specific expiry
└── 004_trade_history.up.sql      # Trade history table
```

#### SQLite Migrations

SQLite migrations are embedded in the Go binary (`internal/storage/sqlite.go`) and run automatically via `embed` package.

#### PostgreSQL Migrations

PostgreSQL tracks applied migrations in `schema_migrations` table:

```sql
-- View applied migrations
SELECT * FROM schema_migrations ORDER BY version;

-- Manually mark migration as applied (if needed)
INSERT INTO schema_migrations (version, applied_at) VALUES (4, NOW());
```

#### Migration Tool

Use the provided migration script for SQLite → PostgreSQL migration:

```bash
# Migrate data from SQLite to PostgreSQL
go run scripts/migrate_sqlite_to_postgres.go \
  --source telegram-trader.db \
  --target "postgres://telegram_trader:password@localhost:5432/telegram_trader"
```

### Database Schema

#### user_sessions

Stores authenticated user sessions with 24-hour expiry:

```sql
CREATE TABLE user_sessions (
    user_id INTEGER PRIMARY KEY,
    chat_id INTEGER NOT NULL,
    username TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP DEFAULT (datetime(CURRENT_TIMESTAMP, '+24 hours'))
);
```

#### wallet_sessions

Stores connected wallet addresses (TON and GALA):

```sql
CREATE TABLE wallet_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    wallet_type TEXT CHECK(wallet_type IN ('ton', 'gala')),
    address TEXT NOT NULL,
    connected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active INTEGER DEFAULT 1,
    UNIQUE(user_id, wallet_type),
    FOREIGN KEY (user_id) REFERENCES user_sessions(user_id)
);
```

#### trade_history

Stores all swap and arbitrage executions:

```sql
CREATE TABLE trade_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    trade_type TEXT CHECK(trade_type IN ('swap', 'arbitrage')),
    chain TEXT NOT NULL,
    from_token TEXT NOT NULL,
    to_token TEXT NOT NULL,
    amount_in TEXT NOT NULL,
    amount_out TEXT,
    fee TEXT,
    tx_hash_ton TEXT,
    tx_hash_gala TEXT,
    status TEXT DEFAULT 'pending',
    execution_time_ms INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Backup and Restore

### PostgreSQL Backup

#### Full Database Backup

```bash
# Using docker-compose
docker-compose -f deployments/docker-compose.yml exec postgres \
  pg_dump -U telegram_trader telegram_trader > backup-$(date +%Y%m%d-%H%M%S).sql

# Direct PostgreSQL connection
pg_dump -U telegram_trader -d telegram_trader -h localhost > backup.sql
```

#### Automated Daily Backups

Add to crontab:

```bash
# Daily backup at 2 AM
0 2 * * * docker-compose -f /path/to/deployments/docker-compose.yml exec -T postgres \
  pg_dump -U telegram_trader telegram_trader | gzip > /backups/backup-$(date +\%Y\%m\%d).sql.gz
```

#### Table-Specific Backup

```bash
# Backup only user sessions
docker-compose -f deployments/docker-compose.yml exec postgres \
  pg_dump -U telegram_trader -t user_sessions telegram_trader > user_sessions_backup.sql

# Backup only trade history
docker-compose -f deployments/docker-compose.yml exec postgres \
  pg_dump -U telegram_trader -t trade_history telegram_trader > trade_history_backup.sql
```

### PostgreSQL Restore

```bash
# Stop bot service first
docker-compose -f deployments/docker-compose.yml stop bot-service galachain-service

# Restore from backup
docker-compose -f deployments/docker-compose.yml exec -T postgres \
  psql -U telegram_trader telegram_trader < backup.sql

# Or restore compressed backup
gunzip -c backup-20260116.sql.gz | \
  docker-compose -f deployments/docker-compose.yml exec -T postgres \
  psql -U telegram_trader telegram_trader

# Restart services
docker-compose -f deployments/docker-compose.yml start bot-service galachain-service
```

### SQLite Backup

```bash
# Simple file copy (stop bot first)
docker-compose -f deployments/docker-compose.yml stop bot-service
cp telegram-trader.db telegram-trader.db.backup-$(date +%Y%m%d)

# Or use SQLite backup command
sqlite3 telegram-trader.db ".backup 'telegram-trader.db.backup'"
```

### SQLite Restore

```bash
# Stop bot
docker-compose -f deployments/docker-compose.yml stop bot-service

# Restore from backup
cp telegram-trader.db.backup telegram-trader.db

# Restart bot
docker-compose -f deployments/docker-compose.yml start bot-service
```

### Trade Logs Backup

Trade logs are stored in JSONL format in `logs/trades.jsonl`:

```bash
# Backup trade logs
cp logs/trades.jsonl logs/trades.jsonl.backup-$(date +%Y%m%d)

# Compress old logs
gzip logs/trades.jsonl.backup-20260115

# View recent trades
tail -n 100 logs/trades.jsonl | jq .
```

### Backup Strategy Recommendations

**Development**:
- Manual backups before major changes
- Keep last 7 days of backups

**Staging**:
- Daily automated backups
- Retain 30 days of backups
- Test restore procedures weekly

**Production**:
- Hourly incremental backups (WAL archiving)
- Daily full backups
- Retain 90 days of backups
- Store backups off-site (S3, Google Cloud Storage)
- Monthly restore drills
- Monitor backup success via alerting

## Monitoring and Health Checks

### Health Check Endpoints

Both services expose HTTP health check endpoints:

#### Bot Service Health Check

```bash
# Check bot service health
curl http://localhost:8080/health

# Response format:
{
  "status": "healthy",           # healthy | degraded | unhealthy
  "timestamp": "2026-01-16T12:00:00Z",
  "dependencies": [
    {
      "name": "database",
      "status": "healthy",
      "latency_ms": 2
    },
    {
      "name": "ton_client",
      "status": "healthy",
      "message": "Connected to mainnet",
      "latency_ms": 45
    },
    {
      "name": "galachain_grpc",
      "status": "healthy",
      "latency_ms": 8
    }
  ],
  "version": "1.0.0",
  "uptime_seconds": 3600
}
```

#### GalaChain Service Health Check

```bash
# Check GalaChain service health
curl http://localhost:8081/health

# Response format:
{
  "status": "healthy",
  "timestamp": "2026-01-16T12:00:00Z",
  "dependencies": [
    {
      "name": "gswap_api",
      "status": "healthy",
      "message": "API reachable",
      "latency_ms": 120
    }
  ],
  "version": "1.0.0",
  "uptime_seconds": 3600
}
```

### Health Status Definitions

- **healthy**: All dependencies operational
- **degraded**: 1 dependency unhealthy OR any dependency degraded
- **unhealthy**: 2+ dependencies unhealthy

### Docker Health Checks

Docker Compose includes built-in health checks:

```bash
# View health status
docker-compose -f deployments/docker-compose.yml ps

# Detailed health status
docker inspect telegram-trader-bot | jq '.[0].State.Health'
```

Health checks run every 30 seconds with 3 retries before marking unhealthy.

### Logging

#### Structured JSON Logs

All services output structured JSON logs to stdout:

```json
{
  "timestamp": "2026-01-16T12:00:00Z",
  "level": "INFO",
  "service": "telegram-bot",
  "event_type": "command_executed",
  "user_id": 123456789,
  "correlation_id": "abc-123-def-456",
  "message": "User executed /balance command"
}
```

#### Viewing Logs

```bash
# Live logs from all services
docker-compose -f deployments/docker-compose.yml logs -f

# Filter by service
docker-compose -f deployments/docker-compose.yml logs -f bot-service

# Filter by log level (requires jq)
docker-compose -f deployments/docker-compose.yml logs bot-service | jq 'select(.level == "ERROR")'

# Follow errors only
docker-compose -f deployments/docker-compose.yml logs -f bot-service | grep ERROR
```

#### Log Levels

- **DEBUG**: Detailed information for debugging
- **INFO**: Normal operations (commands, connections)
- **WARN**: Warning conditions (rate limits, retries)
- **ERROR**: Error conditions (failed operations, external API errors)
- **FATAL**: Critical errors causing service shutdown

#### Trade Audit Logs

Separate audit trail for all trades in `logs/trades.jsonl`:

```bash
# View all trades
cat logs/trades.jsonl | jq .

# View trades for specific user
grep '"user_id":123456789' logs/trades.jsonl | jq .

# View failed trades
jq 'select(.status == "failed")' logs/trades.jsonl

# Count trades by type
jq -r '.type' logs/trades.jsonl | sort | uniq -c

# Daily trade volume
jq -r 'select(.type == "swap") | .from_amount' logs/trades.jsonl | \
  awk '{sum += $1} END {print sum}'
```

### Monitoring Recommendations

**Basic Monitoring** (Development):
- Check health endpoints manually
- Review logs for errors
- Monitor Docker container status

**Advanced Monitoring** (Production):
- **Prometheus** + **Grafana**: Metrics dashboards
- **ELK Stack**: Centralized log aggregation
- **Alerting**: PagerDuty, Slack webhooks
- **Uptime Monitoring**: UptimeRobot, Pingdom

#### Key Metrics to Monitor

1. **Service Health**
   - Health check status (healthy/degraded/unhealthy)
   - Service uptime
   - Restart count

2. **Database**
   - Connection pool usage
   - Query latency
   - Active sessions count

3. **External APIs**
   - TON blockchain RPC latency
   - ston.fi API response time
   - GSwap API availability
   - gRPC latency

4. **Business Metrics**
   - Commands per minute
   - Active users (last 24h)
   - Successful swaps vs failed
   - Arbitrage opportunities detected
   - Average trade volume

5. **Errors**
   - Error rate by type
   - Rate limit violations
   - Authentication failures

## Troubleshooting

### Bot Not Responding

**Symptoms**: Bot doesn't reply to commands in Telegram

**Diagnosis**:
```bash
# 1. Check bot service is running
docker-compose -f deployments/docker-compose.yml ps bot-service

# 2. Check logs for errors
docker-compose -f deployments/docker-compose.yml logs --tail=100 bot-service

# 3. Verify BOT_TOKEN
docker-compose -f deployments/docker-compose.yml exec bot-service env | grep BOT_TOKEN
```

**Solutions**:
- **Invalid BOT_TOKEN**: Get new token from @BotFather and update `.env`
- **User not authorized**: Add your Telegram user ID to `BOT_ADMIN_USER_IDS`
- **Network issues**: Check Docker network connectivity
- **Service crashed**: Restart with `docker-compose restart bot-service`

### Database Connection Failed

**Symptoms**: `error: failed to connect to database` in logs

**Diagnosis**:
```bash
# 1. Check PostgreSQL is running
docker-compose -f deployments/docker-compose.yml ps postgres

# 2. Test database connection
docker-compose -f deployments/docker-compose.yml exec postgres \
  psql -U telegram_trader -d telegram_trader -c "SELECT 1;"

# 3. Check connection string
docker-compose -f deployments/docker-compose.yml exec bot-service env | grep DATABASE_URL
```

**Solutions**:
- **PostgreSQL not ready**: Wait for health check, PostgreSQL takes ~10s to initialize
- **Wrong credentials**: Verify `POSTGRES_USER` and `POSTGRES_PASSWORD` match in `.env`
- **Network issue**: Check `telegram-trader-network` exists: `docker network ls`
- **Migration failed**: Check PostgreSQL logs for migration errors

### GalaChain Service Unreachable

**Symptoms**: `error: failed to connect to galachain service` in bot logs

**Diagnosis**:
```bash
# 1. Check GalaChain service is running
docker-compose -f deployments/docker-compose.yml ps galachain-service

# 2. Check service logs
docker-compose -f deployments/docker-compose.yml logs --tail=100 galachain-service

# 3. Test gRPC connectivity
grpcurl -plaintext localhost:50051 list
```

**Solutions**:
- **Service not started**: Ensure PostgreSQL is healthy first (dependency)
- **Wrong URL**: Verify `GALACHAIN_SERVICE_URL=galachain-service:50051` (Docker network name)
- **Port conflict**: Change `GRPC_PORT` if 50051 is in use
- **Build failed**: Rebuild with `docker-compose up -d --build galachain-service`

### Rate Limit Errors

**Symptoms**: "Too many requests. Please wait 60 seconds."

**Diagnosis**:
```bash
# Check rate limit middleware logs
docker-compose -f deployments/docker-compose.yml logs bot-service | grep "rate_limit"
```

**Solutions**:
- **Expected behavior**: User exceeded 10 commands/minute, wait and retry
- **Development override**: Temporarily increase limit in `internal/bot/middleware/ratelimit.go` (not recommended for production)
- **Reset rate limits**: Restart bot service (dev only)

### Swap Execution Failed

**Symptoms**: "Swap failed" error after confirmation

**Diagnosis**:
```bash
# 1. Check trade logs for details
tail -n 50 logs/trades.jsonl | jq 'select(.status == "failed")'

# 2. Check wallet connection
docker-compose -f deployments/docker-compose.yml logs bot-service | grep "wallet"

# 3. Check external API logs
docker-compose -f deployments/docker-compose.yml logs bot-service | grep "stonfi"
```

**Solutions**:
- **Wallet not connected**: Run `/wallet <address>` first
- **Insufficient balance**: Check balance with `/balance`
- **ston.fi API down**: Check [ston.fi status](https://status.ston.fi)
- **Invalid private key**: Reconnect wallet with valid key
- **Network congestion**: Retry after a few minutes

### High Memory Usage

**Symptoms**: Container using excessive memory (>500MB)

**Diagnosis**:
```bash
# Check container stats
docker stats telegram-trader-bot telegram-trader-galachain

# Check for memory leaks in logs
docker-compose -f deployments/docker-compose.yml logs bot-service | grep -i "memory\|oom"
```

**Solutions**:
- **PostgreSQL connection leak**: Check connection pool size in code
- **Log buffer buildup**: Rotate logs more frequently
- **Restart services**: `docker-compose restart bot-service galachain-service`
- **Set memory limits**: Add `mem_limit: 512m` to docker-compose.yml

### Encryption Key Errors

**Symptoms**: "Encryption failed" or "Invalid key" errors

**Diagnosis**:
```bash
# 1. Verify key format (must be base64, 44 characters for 32 bytes)
echo $ENCRYPTION_KEY | base64 -d | wc -c  # Should output: 32

# 2. Check key is set
docker-compose -f deployments/docker-compose.yml exec bot-service env | grep ENCRYPTION_KEY
```

**Solutions**:
- **Invalid key**: Generate new key: `openssl rand -base64 32`
- **Key changed**: ⚠️ Changing key invalidates existing encrypted data, restore from backup
- **Key not set**: Add `ENCRYPTION_KEY` to `.env` file

### Disk Space Full

**Symptoms**: Database writes failing, logs not being written

**Diagnosis**:
```bash
# Check disk usage
docker system df

# Check volume usage
docker volume ls
docker inspect telegram-trader_postgres-data | jq '.[0].Mountpoint' | xargs du -sh
```

**Solutions**:
- **Clean old logs**: `rm logs/trades.jsonl.backup-*`
- **Prune unused Docker data**: `docker system prune -a --volumes` (⚠️ careful!)
- **Compress trade logs**: `gzip logs/trades.jsonl`
- **Archive old trades**: Move old data to cold storage

### Logs Full of Warnings

**Symptoms**: High volume of WARN level logs

**Common Warnings**:
- `Rate limit approached for user X`: User sending many commands (expected)
- `TON client reconnecting`: Temporary network issue (retries automatically)
- `GSwap API slow response`: External API latency (no action needed if < 5s)
- `Session expires soon`: User session expiring within 1 hour (expected)

**When to Act**:
- **ERROR logs**: Investigate immediately
- **Repeated warnings**: May indicate underlying issue (rate limit abuse, network problems)
- **FATAL logs**: Service crash, requires immediate attention

## Security Best Practices

### Environment Security

1. **Never commit `.env` files**
   ```bash
   # Verify .env is in .gitignore
   cat .gitignore | grep .env
   ```

2. **Use strong passwords**
   ```bash
   # Generate secure PostgreSQL password
   openssl rand -base64 32
   ```

3. **Rotate encryption keys** (every 90 days)
   ```bash
   # Generate new key
   openssl rand -base64 32

   # Update .env with new key
   # Deploy new key
   # Migrate encrypted data (see scripts/rotate_encryption_key.sh)
   ```

4. **Separate credentials per environment**
   - Development: `ENCRYPTION_KEY_DEV`
   - Staging: `ENCRYPTION_KEY_STAGING`
   - Production: `ENCRYPTION_KEY_PROD`

### Network Security

1. **Firewall configuration**
   ```bash
   # Only expose necessary ports
   # Block: 5432 (PostgreSQL), 50051 (gRPC)
   # Allow: 8080 (bot health), 8081 (galachain health)
   ```

2. **HTTPS for webhooks** (production)
   - Use Telegram webhooks instead of polling
   - Require TLS 1.2+ with valid certificates
   - Validate Telegram's webhook IPs

3. **Internal network isolation**
   - Keep PostgreSQL on private network
   - Use Docker network for inter-service communication
   - Don't expose gRPC port publicly

### Access Control

1. **Admin whitelist**
   - Maintain minimal `BOT_ADMIN_USER_IDS` list
   - Review access quarterly
   - Remove inactive users

2. **Database permissions**
   ```sql
   -- Create read-only user for analytics
   CREATE USER analyst WITH PASSWORD 'secure_password';
   GRANT CONNECT ON DATABASE telegram_trader TO analyst;
   GRANT SELECT ON ALL TABLES IN SCHEMA public TO analyst;
   ```

3. **Rate limiting**
   - Keep rate limits enabled (10 commands/minute)
   - Monitor for abuse patterns
   - Implement IP-based rate limiting for webhooks

### Data Protection

1. **Encryption at rest**
   - Enable PostgreSQL encryption
   - Use encrypted Docker volumes
   - Encrypt backup files before off-site storage

2. **Encryption in transit**
   - Use TLS for PostgreSQL connections
   - Use TLS for gRPC (production)
   - Use HTTPS for external APIs

3. **Key management**
   - ⚠️ **Development**: `.env` files (acceptable)
   - ✅ **Production**: AWS Secrets Manager, HashiCorp Vault, GCP Secret Manager
   - Never log encryption keys
   - Rotate keys regularly

### Monitoring and Auditing

1. **Enable audit logging**
   ```bash
   # All trades logged to logs/trades.jsonl
   # Review regularly for suspicious activity
   tail -f logs/trades.jsonl | jq .
   ```

2. **Monitor authentication failures**
   ```bash
   # Check for unauthorized access attempts
   docker-compose logs bot-service | grep "Access denied" | jq .
   ```

3. **Alert on anomalies**
   - High error rates
   - Unusual trading patterns
   - Failed authentication attempts
   - Database connection failures

### Compliance

1. **Data retention**
   - Define retention policy (e.g., 90 days for logs, 1 year for trades)
   - Implement automated cleanup
   - Document in privacy policy

2. **GDPR considerations** (if applicable)
   - Implement user data export
   - Implement user data deletion
   - Document data processing

3. **Financial regulations**
   - Consult legal counsel for your jurisdiction
   - Implement KYC/AML if required
   - Maintain audit trails for trades

### Incident Response

1. **Prepare incident runbook**
   - Contact list (engineers, management)
   - Escalation procedures
   - Communication templates

2. **Backup and recovery tested**
   - Monthly restore drills
   - Document restore procedures
   - Measure RTO/RPO (Recovery Time/Point Objectives)

3. **Security incident procedures**
   - Isolate compromised systems
   - Rotate all credentials
   - Review audit logs
   - Notify affected users if required

---

## Additional Resources

- **[README.md](../README.md)** - Project overview and quick start
- **[TESTING.md](./TESTING.md)** - Testing guide and procedures
- **[Design Document](../specs/telegram-trader/design.md)** - Architecture and technical design
- **[Requirements](../specs/telegram-trader/requirements.md)** - Functional and non-functional requirements

## Support

For deployment issues:
1. Check this guide's [Troubleshooting](#troubleshooting) section
2. Review service logs for errors
3. Check [GitHub Issues](https://github.com/YOUR_USERNAME/telegram-trader/issues)
4. Contact: your-email@example.com

---

**Last Updated**: 2026-01-16
**Document Version**: 1.0.0
