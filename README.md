# Telegram Trader

[![CI](https://github.com/YOUR_USERNAME/telegram-trader/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_USERNAME/telegram-trader/actions/workflows/ci.yml)
[![Security](https://github.com/YOUR_USERNAME/telegram-trader/actions/workflows/security.yml/badge.svg)](https://github.com/YOUR_USERNAME/telegram-trader/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/YOUR_USERNAME/telegram-trader/branch/main/graph/badge.svg)](https://codecov.io/gh/YOUR_USERNAME/telegram-trader)

A Telegram bot for trading tokens on GalaChain (GSwap) and TON blockchain (ston.fi). Execute trades, check prices, and detect arbitrage opportunities across chains.

## Features

- **Multi-Chain Trading**: Trade on both GalaChain (GSwap) and TON blockchain (ston.fi)
- **Wallet Management**: Connect TonKeeper and Gala wallet
- **Price Checking**: Real-time token prices from both DEXs
- **Arbitrage Detection**: Automatically detect profitable arbitrage opportunities between chains
- **Token Swaps**: Execute token swaps with simulation and confirmation
- **Balance Tracking**: View your TON and GALA token balances
- **Security**: AES-256-GCM encryption for private keys, rate limiting, and admin whitelist

## Architecture

This project uses a microservices architecture:

- **Go Bot Service** (`cmd/telegram-bot`): Telegram bot interface with command handlers
- **TypeScript GalaChain Service** (`galachain-service/`): gRPC service for GalaChain/GSwap integration
- **gRPC Communication**: Services communicate via Protocol Buffers over gRPC
- **SQLite Database**: Local storage for user sessions and wallet data (development)
- **PostgreSQL**: Production database for multi-instance deployments

### Key Token Pairs

- **TON/GALA** on ston.fi: [View on ston.fi](https://app.ston.fi/swap?chartVisible=false&ft=TON&tt=EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV)
- **GTON/GALA** on GSwap: [View on GSwap](https://swap.gala.com/explore-balance/2954bd1e38c2dff11a2ee8798e2f59b99c206e64bef7bdf5c4ac9a91315e0d34)

## Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Node.js 20+** - [Install Node.js](https://nodejs.org/)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Telegram Bot Token** - Create a bot via [@BotFather](https://t.me/botfather)

## Quick Start

Get the bot running in under 5 minutes:

### 1. Clone and Configure

```bash
git clone <repository-url>
cd telegram-trader
```

### 2. Set Up Environment Variables

Create `.env` file in the project root:

```bash
# Copy from example
cp .env.example .env
```

Edit `.env` with your values:

```env
BOT_TOKEN=your_telegram_bot_token_here
BOT_ADMIN_USER_IDS=your_telegram_user_id
ENCRYPTION_KEY=$(openssl rand -base64 32)
```

Create `.env` file in `galachain-service/`:

```bash
cd galachain-service
cp .env.example .env
cd ..
```

### 3. Start with Docker Compose

```bash
docker-compose -f deployments/docker-compose.yml up --build
```

The bot will start and connect to Telegram. Services will be available at:
- Bot service: Running in container
- GalaChain gRPC service: `localhost:50051`
- PostgreSQL: `localhost:5432`

### 4. Test the Bot

Open Telegram and message your bot with `/start`

## Development Setup

For local development without Docker:

### 1. Install Dependencies

**Go dependencies:**
```bash
go mod download
```

**TypeScript dependencies:**
```bash
cd galachain-service
npm install
cd ..
```

### 2. Generate gRPC Code

**Go:**
```bash
protoc --go_out=. --go-grpc_out=. proto/galachain.proto
```

**TypeScript:**
```bash
cd galachain-service
npm run proto
cd ..
```

### 3. Set Up Database

The SQLite database will be created automatically on first run at `telegram-trader.db`.

Migrations are embedded in the Go binary and run automatically.

### 4. Start Services

**Terminal 1 - GalaChain Service:**
```bash
cd galachain-service
npm run build
npm start
```

**Terminal 2 - Telegram Bot:**
```bash
go run ./cmd/telegram-bot
```

## Environment Variables

### Bot Service (Root `.env`)

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `BOT_TOKEN` | Yes | Telegram bot token from @BotFather | `1234567890:ABCdef...` |
| `BOT_ADMIN_USER_IDS` | Yes | Comma-separated Telegram user IDs with access | `123456789,987654321` |
| `DATABASE_URL` | Yes | SQLite database path (dev) or PostgreSQL URL (prod) | `telegram-trader.db` |
| `ENCRYPTION_KEY` | Yes | Base64-encoded 32-byte key for AES-256-GCM | Generate: `openssl rand -base64 32` |
| `GALACHAIN_SERVICE_URL` | Yes | gRPC address of GalaChain service | `localhost:50051` |

### GalaChain Service (`galachain-service/.env`)

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `GRPC_PORT` | Yes | Port for gRPC server | `50051` |
| `DATABASE_URL` | Yes | PostgreSQL connection string | `postgresql://user:pass@localhost:5432/telegram_trader` |
| `ENCRYPTION_KEY` | Yes | 64-character hex-encoded encryption key | `0123456789abcdef...` |
| `GSWAP_API_URL` | Yes | GalaChain GSwap API URL | `https://swap.gala.com/api` |

### Docker Compose Variables

When using Docker Compose, also set in your `.env`:

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_USER` | `telegram_trader` | PostgreSQL username |
| `POSTGRES_PASSWORD` | `postgres` | PostgreSQL password |
| `POSTGRES_DB` | `telegram_trader` | PostgreSQL database name |

## Bot Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/start` | Welcome message and command overview | `/start` |
| `/help` | Show all available commands | `/help` |
| `/wallet` | Connect TON wallet address | `/wallet UQAbc123...` |
| `/balance` | View your TON and GALA balances | `/balance` |
| `/swap` | Execute token swap with preview | `/swap 1 TON GALA stonfi` |
| `/price` | Check token prices on both DEXs | `/price` |
| `/arbitrage` | Detect arbitrage opportunities | `/arbitrage` |

### Command Details

#### `/wallet <address>`
Connect your TonKeeper wallet address to the bot.

```
/wallet UQAbc123def456...
```

#### `/swap <amount> <from_token> <to_token> <dex>`
Simulate and execute token swaps.

```
/swap 1 TON GALA stonfi
/swap 100 GALA TON stonfi
```

Supported tokens: `TON`, `GALA`
Supported DEX: `stonfi` (GSwap support coming soon)

#### `/price`
Get current prices for TON/GALA pair on both exchanges plus spread.

```
TON/GALA on ston.fi: 850.00 GALA
GTON/GALA on gswap: 855.00 GALA
Spread: 0.59%
```

#### `/arbitrage`
Check for profitable arbitrage opportunities between ston.fi and GSwap.

```
Arbitrage opportunity detected!

Direction: Buy on ston.fi, sell on gswap
Spread: 0.59%
ston.fi price: 850.00 GALA
gswap price: 855.00 GALA
Estimated profit: 50% of TON balance

[Execute] [Cancel]
```

The bot will automatically execute arbitrage when:
- Spread ≥ 0.3% (configurable)
- Sufficient balance available
- Maintains minimum 10 GALA or 1 TON

## Security

### Rate Limiting
- **10 commands per minute** per user
- Token bucket algorithm with automatic refill
- Prevents abuse and API overload

### Authentication
- **Admin whitelist**: Only users in `BOT_ADMIN_USER_IDS` can access the bot
- User sessions tracked in database
- Automatic session creation on first interaction

### Encryption
- **AES-256-GCM** encryption for private keys
- Master key stored in `ENCRYPTION_KEY` environment variable
- Authenticated encryption with additional data (AEAD)
- **⚠️ Production**: Use proper key management system (AWS KMS, HashiCorp Vault)

### Best Practices
1. Never commit `.env` files to version control
2. Generate strong encryption keys: `openssl rand -base64 32`
3. Rotate encryption keys regularly
4. Use separate credentials for dev/staging/production
5. Enable HTTPS for production webhooks

## Project Structure

```
telegram-trader/
├── cmd/
│   └── telegram-bot/          # Bot service entry point
├── internal/
│   ├── arbitrage/             # Arbitrage detection engine
│   ├── blockchain/            # Blockchain clients (TON)
│   │   └── ton/
│   ├── bot/                   # Telegram bot logic
│   │   ├── handlers/          # Command handlers
│   │   └── middleware/        # Auth & rate limiting
│   ├── config/                # Configuration loading
│   ├── dex/                   # DEX clients
│   │   └── stonfi/            # ston.fi integration
│   ├── errors/                # Error handling
│   ├── galachain/             # GalaChain gRPC client
│   ├── logging/               # Structured logging
│   ├── storage/               # Database layer
│   └── wallet/                # Wallet management & encryption
├── galachain-service/         # TypeScript GalaChain service
│   ├── src/
│   │   ├── config/            # Configuration
│   │   ├── gswap/             # GSwap client
│   │   ├── server/            # gRPC server
│   │   └── types/             # Generated gRPC types
│   └── package.json
├── proto/                     # Protocol Buffer definitions
│   └── galachain.proto
├── deployments/
│   └── docker-compose.yml     # Docker Compose configuration
├── migrations/                # Database migrations
├── .env.example               # Environment template (Go)
└── README.md                  # This file
```

## Logging

### Structured Logs
All services use structured JSON logging to stdout:

```json
{
  "timestamp": "2026-01-16T10:30:00Z",
  "level": "INFO",
  "service": "telegram-bot",
  "message": "User command executed",
  "user_id": 123456789,
  "command": "/balance"
}
```

### Trade Logs
Trade executions are logged to `logs/trades-YYYY-MM-DD.jsonl`:

```json
{
  "timestamp": "2026-01-16T10:30:00Z",
  "user_id": 123456789,
  "type": "swap",
  "from_token": "TON",
  "to_token": "GALA",
  "from_amount": "1.0",
  "to_amount": "850.0",
  "fee": "0.003",
  "tx_hash": "abc123...",
  "status": "success",
  "execution_time_ms": 1234
}
```

## Testing

### Go Tests

Run all tests:
```bash
go test ./... -v
```

Run specific package:
```bash
go test ./internal/arbitrage -v
go test ./internal/bot/middleware -v
go test ./internal/wallet -v
```

### TypeScript Tests

```bash
cd galachain-service
npm test
```

### Integration Tests

Start services with Docker Compose and test commands:
```bash
docker-compose -f deployments/docker-compose.yml up -d
# Test via Telegram bot interface
docker-compose -f deployments/docker-compose.yml logs -f bot-service
```

## Troubleshooting

### Bot not responding
1. Check `BOT_TOKEN` is correct
2. Verify your user ID is in `BOT_ADMIN_USER_IDS`
3. Check bot logs: `docker-compose logs bot-service`

### GalaChain service connection failed
1. Verify `GALACHAIN_SERVICE_URL` is correct (`galachain-service:50051` in Docker)
2. Check gRPC service is running: `docker-compose ps`
3. Check service logs: `docker-compose logs galachain-service`

### Database errors
1. Check write permissions for `telegram-trader.db`
2. Verify `DATABASE_URL` is correct
3. Delete database and restart to recreate: `rm telegram-trader.db`

### Rate limit errors
Wait 60 seconds or restart bot to reset rate limits (dev only).

### Encryption errors
1. Verify `ENCRYPTION_KEY` is base64-encoded 32-byte key
2. Regenerate key: `openssl rand -base64 32`
3. **Warning**: Changing key will invalidate existing encrypted data

## Documentation

- **[Requirements](specs/telegram-trader/requirements.md)**: User stories, functional & non-functional requirements
- **[Design](specs/telegram-trader/design.md)**: Architecture, component design, API interfaces
- **[Tasks](specs/telegram-trader/tasks.md)**: Implementation task breakdown (72 tasks across 4 phases)
- **[Research](specs/telegram-trader/research.md)**: Technology research and decisions

## Roadmap

### Phase 1: POC (Current - 28/28 tasks complete ✅)
- ✅ Basic Telegram bot with command routing
- ✅ TON/ston.fi integration
- ✅ GalaChain/GSwap integration via gRPC
- ✅ Price checking and spread calculation
- ✅ Arbitrage detection (0.3% threshold)
- ✅ Wallet connection (manual address)
- ✅ Token swaps with simulation
- ✅ Authentication and rate limiting
- ✅ AES-256 encryption

### Phase 2: Refactoring (Tasks 29-43)
- Separate concerns into clean packages
- Improve error handling
- Add retry logic with exponential backoff
- Implement connection pooling
- Add metrics and monitoring

### Phase 3: Testing (Tasks 44-60)
- Unit tests (70% coverage)
- Integration tests for each component
- End-to-end tests
- Load testing

### Phase 4: Quality (Tasks 61-72)
- Code linting and formatting
- Security audit
- Performance optimization
- Documentation improvements
- Production deployment guides

## Contributing

This is a POC project. Contributions welcome after Phase 1 completion.

## License

MIT License - see LICENSE file for details

## Support

- **Issues**: Open an issue on GitHub
- **Telegram**: Contact bot administrator
- **Documentation**: See `specs/telegram-trader/` for detailed docs

---

**⚠️ Disclaimer**: This is a POC (Proof of Concept) bot for development and testing. Do not use with large amounts of funds. Always test on testnets first. Trading cryptocurrencies carries risk.
