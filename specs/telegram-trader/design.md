---
spec: telegram-trader
phase: design
created: 2026-01-16T14:30:00Z
---

# Design: Telegram Trader

## Design Inputs

### Interview Responses

**Architecture Style**: Greenfield (new project)
- Starting from scratch with microservices architecture
- No legacy code to maintain or migrate
- Freedom to implement best practices from the start

**Technology Constraints**: Follow research recommendations
- Go Telegram Bot: `go-telegram/bot` library
- TON Blockchain: `tonutils-go` SDK
- GalaChain: `@gala-chain/gswap-sdk` (TypeScript only)
- Inter-service communication: gRPC with Protocol Buffers

**MVP Phasing Strategy**:
- **Phase 1 (P0)**: Basic wallet connection + manual trading
- **Phase 2 (P1)**: Price monitoring + alerts + trade history
- **Phase 3 (P2)**: Arbitrage automation

### Implications

1. **Microservices Required**: GalaChain SDK is TypeScript-only, requiring separate service
2. **gRPC Communication**: Performance-critical for arbitrage timing
3. **Security First**: Public bot handling private keys requires encryption, authentication, rate limiting
4. **Iterative Development**: MVP-first approach allows user feedback before automation
5. **Reference Implementation**: Leverage `cchuter/gswap-bot` patterns for GalaChain service

## Executive Summary

The Telegram Trader is a **dual-chain arbitrage trading bot** that enables users to trade TON/GALA on ston.fi and GTON/GALA on gswap, with automated arbitrage detection and execution. The system uses a **microservices architecture** with two primary services:

1. **Go Telegram Bot Service**: Handles user interaction, TON blockchain operations, and orchestration
2. **TypeScript GalaChain Service**: Manages GalaChain operations via gswap SDK

**Key Design Principles**:
- **Security by Design**: AES-256 encrypted private keys, user authentication, rate limiting
- **Performance Critical**: Sub-second price monitoring for arbitrage opportunities
- **Fault Tolerant**: Graceful degradation, retry logic, circuit breakers
- **Observable**: Structured logging, Prometheus metrics, health checks
- **User-Centric**: Clear error messages, confirmation flows, safety checks

**Technology Stack**:
- **Bot Service**: Go 1.21+, go-telegram/bot, tonutils-go, gRPC client
- **GalaChain Service**: TypeScript 5+, Node.js 20+, gswap-sdk, gRPC server
- **Communication**: gRPC with Protocol Buffers
- **Storage**: SQLite (dev), PostgreSQL (prod), JSONL trade logs
- **Deployment**: Docker Compose (dev), Kubernetes (prod)

## System Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph "External Services"
        TG[Telegram API]
        TONBC[TON Blockchain]
        STONFI[ston.fi API]
        GALABC[GalaChain Network]
        GSWAP[gswap API]
        CG[CoinGecko API]
    end

    subgraph "Telegram Trader System"
        BOT[Go Telegram Bot Service]
        GALA[TypeScript GalaChain Service]
        DB[(Database)]
        LOGS[JSONL Trade Logs]
    end

    subgraph "User Devices"
        USER[Telegram User]
        TK[TonKeeper Wallet]
        GW[Gala Wallet]
    end

    USER -->|Commands| TG
    TG <-->|Bot API| BOT
    BOT <-->|TonConnect| TK
    BOT <-->|Blockchain Ops| TONBC
    BOT <-->|Price/Swap| STONFI
    BOT <-->|gRPC| GALA
    BOT <-->|Read/Write| DB
    BOT -->|Append| LOGS

    GALA <-->|WalletConnect| GW
    GALA <-->|Blockchain Ops| GALABC
    GALA <-->|Swap API| GSWAP
    GALA <-->|Read/Write| DB
    GALA -->|Append| LOGS

    BOT -.->|USD Prices| CG
    GALA -.->|USD Prices| CG

    style BOT fill:#4CAF50
    style GALA fill:#2196F3
    style DB fill:#FF9800
    style LOGS fill:#FF9800
```

### Component Architecture

```mermaid
graph TB
    subgraph "Go Telegram Bot Service"
        CMD[Command Handler]
        AUTH[Auth Middleware]
        RATE[Rate Limiter]
        SESS[Session Manager]
        TONC[TON Client]
        STONC[ston.fi Client]
        WALLETMGR[Wallet Manager]
        ARB[Arbitrage Engine]
        GRPCC[gRPC Client]
        MONITOR[Price Monitor]

        CMD --> AUTH
        AUTH --> RATE
        RATE --> SESS
        SESS --> WALLETMGR
        SESS --> TONC
        SESS --> STONC
        SESS --> ARB
        SESS --> MONITOR
        ARB --> GRPCC
        MONITOR --> GRPCC
    end

    subgraph "TypeScript GalaChain Service"
        GRPCS[gRPC Server]
        GSWC[gswap Client]
        GALAC[GalaChain Client]
        WSIGN[Wallet Signer]
        PRICEP[Price Provider]
        SWAPEX[Swap Executor]

        GRPCS --> GSWC
        GRPCS --> GALAC
        GRPCS --> PRICEP
        GRPCS --> SWAPEX
        SWAPEX --> WSIGN
        PRICEP --> GSWC
    end

    subgraph "Shared Components"
        CONFIG[Config Manager]
        LOGGER[Structured Logger]
        METRICS[Metrics Collector]
        HEALTH[Health Checker]
    end

    GRPCC <-.->|gRPC| GRPCS

    style CMD fill:#4CAF50
    style GRPCS fill:#2196F3
    style CONFIG fill:#FF9800
```

### Deployment Architecture

```mermaid
graph TB
    subgraph "Production Environment"
        subgraph "Kubernetes Cluster"
            subgraph "Bot Service Pod"
                BOTC[Bot Container<br/>Go Binary]
                BOTVOL[Config Volume]
            end

            subgraph "GalaChain Service Pod"
                GALAC[GalaChain Container<br/>Node.js]
                GALAVOL[Config Volume]
            end

            subgraph "Data Tier"
                PG[(PostgreSQL<br/>StatefulSet)]
                PVC[Persistent Volume<br/>Trade Logs]
            end

            subgraph "Monitoring"
                PROM[Prometheus]
                GRAF[Grafana]
            end
        end

        subgraph "Secret Management"
            VAULT[HashiCorp Vault]
        end

        LB[Load Balancer]
    end

    INTERNET[Internet] --> LB
    LB --> BOTC
    BOTC <-->|gRPC| GALAC
    BOTC --> PG
    BOTC --> PVC
    GALAC --> PG
    GALAC --> PVC
    BOTC -.->|Metrics| PROM
    GALAC -.->|Metrics| PROM
    PROM --> GRAF
    BOTC -.->|Secrets| VAULT
    GALAC -.->|Secrets| VAULT

    style BOTC fill:#4CAF50
    style GALAC fill:#2196F3
    style PG fill:#FF9800
    style VAULT fill:#F44336
```

## Data Flow Diagrams

### Wallet Connection Flow (TonConnect)

```mermaid
sequenceDiagram
    participant U as User
    participant T as Telegram
    participant B as Bot Service
    participant TC as TonConnect Bridge
    participant TK as TonKeeper Wallet
    participant DB as Database

    U->>T: /wallet
    T->>B: Command message
    B->>B: Generate TonConnect session
    B->>TC: Create connection request
    TC-->>B: Session ID + Deep Link + QR
    B->>T: Send QR code + deep link
    T->>U: Display options

    alt Mobile (deep link)
        U->>TK: Click deep link
    else Desktop (QR code)
        U->>TK: Scan QR code
    end

    TK->>U: Approve connection?
    U->>TK: Approve
    TK->>TC: Send wallet address + signature
    TC->>B: Wallet connected event
    B->>B: Validate signature
    B->>DB: Store encrypted session
    B->>T: Connection successful
    T->>U: "Wallet connected: UQ..."
```

### Wallet Connection Flow (WalletConnect)

```mermaid
sequenceDiagram
    participant U as User
    participant T as Telegram
    participant B as Bot Service
    participant G as GalaChain Service
    participant WC as WalletConnect Bridge
    participant GW as Gala Wallet
    participant DB as Database

    U->>T: /wallet gala
    T->>B: Command message
    B->>G: CreateWalletSession gRPC
    G->>G: Generate WalletConnect session
    G->>WC: Create connection request
    WC-->>G: Session URI + QR
    G-->>B: SessionResponse
    B->>T: Send QR code + deep link
    T->>U: Display options

    U->>GW: Scan QR or click link
    GW->>U: Approve connection?
    U->>GW: Approve
    GW->>WC: Send wallet address + signature
    WC->>G: Wallet connected event
    G->>G: Validate signature
    G->>DB: Store encrypted session
    G-->>B: WalletConnected event
    B->>T: Connection successful
    T->>U: "Gala wallet connected: 0x..."
```

### Manual Swap Execution Flow

```mermaid
sequenceDiagram
    participant U as User
    participant T as Telegram
    participant B as Bot Service
    participant S as ston.fi API
    participant TON as TON Blockchain
    participant DB as Database
    participant L as JSONL Logs

    U->>T: /swap 1 TON GALA stonfi
    T->>B: Command message
    B->>B: Parse command, validate
    B->>B: Check wallet connected
    B->>TON: Get current balance
    TON-->>B: Balance: 10 TON
    B->>B: Check minimum (1 TON remains)
    B->>S: POST /v1/swap/simulate
    S-->>B: Expected: 850 GALA, fee: 0.02 TON, slippage: 0.5%
    B->>T: Preview with inline buttons
    T->>U: "Swap 1 TON → ~850 GALA<br/>Fee: 0.02 TON<br/>[Confirm] [Cancel]"

    U->>T: Click [Confirm]
    T->>B: Callback query
    B->>B: Build swap transaction
    B->>TON: Submit signed transaction
    TON-->>B: TX hash: abc123
    B->>T: "Executing swap... TX: abc123"
    T->>U: Status message

    loop Poll status (max 30s)
        B->>S: GET /v1/swap/status?tx=abc123
        S-->>B: Status: pending
    end

    S-->>B: Status: success, actual: 848 GALA
    B->>DB: Update user balance
    B->>L: Append trade record (JSONL)
    B->>T: "✅ Swap complete! Received 848 GALA"
    T->>U: Success message
```

### Price Checking Flow (Cross-Chain)

```mermaid
sequenceDiagram
    participant U as User
    participant T as Telegram
    participant B as Bot Service
    participant S as ston.fi API
    participant G as GalaChain Service
    participant GS as gswap API
    participant C as CoinGecko

    U->>T: /price TON/GALA
    T->>B: Command message

    par Fetch ston.fi price
        B->>S: POST /v1/swap/simulate (1 TON → GALA)
        S-->>B: Output: 850 GALA
        B->>B: Calculate: 1 TON = 850 GALA
    and Fetch gswap price
        B->>G: GetPrice gRPC (GTON/GALA)
        G->>GS: Query pool price
        GS-->>G: Output: 1 GTON = 855 GALA
        G-->>B: PriceResponse: 855
    and Fetch USD prices
        B->>C: GET /simple/price?ids=ton,gala
        C-->>B: TON: $5.20, GALA: $0.0061
    end

    B->>B: Calculate spread: (855-850)/850 = 0.59%
    B->>T: Format response
    T->>U: "TON/GALA Prices:<br/>ston.fi: 850 ($5.18)<br/>gswap: 855 ($5.22)<br/>Spread: 0.59% ⚡"
```

### Arbitrage Execution Flow

```mermaid
sequenceDiagram
    participant U as User
    participant T as Telegram
    participant B as Bot Service
    participant A as Arbitrage Engine
    participant S as ston.fi API
    participant G as GalaChain Service
    participant TON as TON Blockchain
    participant GALA as GalaChain
    participant L as JSONL Logs

    U->>T: /arbitrage
    T->>B: Command message
    B->>A: Trigger arbitrage check

    par Check balances
        A->>TON: Get TON balance
        TON-->>A: 10 TON
        A->>G: GetBalance gRPC
        G->>GALA: Query balance
        GALA-->>G: 5000 GALA
        G-->>A: BalanceResponse
    end

    par Get prices
        A->>S: Simulate 1 TON → GALA
        S-->>A: 850 GALA
        A->>G: GetPrice gRPC (GTON/GALA)
        G-->>A: 855 GALA/GTON
    end

    A->>A: Detect: gswap price higher (0.59% spread)
    A->>A: Calculate: Buy on ston.fi, sell on gswap
    A->>A: Position size: 50% of 10 TON = 5 TON (leaves 5 TON > 1 min)
    A->>A: Expected: Buy 4250 GALA, sell for ~4.97 GTON
    A->>A: Estimate profit: 0.025 GTON (~$0.13)

    A->>T: Preview arbitrage
    T->>U: "Arbitrage Opportunity:<br/>Buy 4250 GALA (5 TON on ston.fi)<br/>Sell 4250 GALA (4.97 GTON on gswap)<br/>Est. profit: 0.025 GTON ($0.13)<br/>[Execute] [Cancel]"

    U->>T: Click [Execute]
    T->>B: Callback query
    B->>A: Execute arbitrage

    par Execute both legs
        A->>S: Submit buy on ston.fi
        S->>TON: Execute swap
        TON-->>S: TX: abc123
        S-->>A: Buy complete: 4248 GALA (actual)
    and
        A->>G: ExecuteSwap gRPC (4250 GALA → GTON)
        G->>GALA: Submit swap transaction
        GALA-->>G: TX: def456
        G-->>A: Sell complete: 4.96 GTON (actual)
    end

    A->>A: Calculate actual profit: 4.96 GTON - 5 TON equivalent
    A->>L: Log arbitrage execution (JSONL)
    A->>T: "✅ Arbitrage complete!<br/>Bought: 4248 GALA<br/>Sold: 4.96 GTON<br/>Profit: 0.024 GTON ($0.12)"
    T->>U: Success message
```

### User Authentication Flow

```mermaid
sequenceDiagram
    participant U as User
    participant T as Telegram
    participant B as Bot Service
    participant A as Auth Middleware
    participant DB as Database

    U->>T: /start
    T->>B: Message (user_id: 12345)
    B->>A: Authenticate request
    A->>DB: Check whitelist (user_id: 12345)

    alt User whitelisted
        DB-->>A: User found, active
        A->>DB: Check/create session
        DB-->>A: Session valid (expires: 24h)
        A-->>B: Auth successful
        B->>T: Welcome message + commands
        T->>U: "Welcome to Telegram Trader!"
    else User not whitelisted
        DB-->>A: User not found
        A-->>B: Auth failed
        B->>T: Access denied message
        T->>U: "Access denied. Contact @admin for access."
    else Session expired
        DB-->>A: Session expired
        A->>DB: Create new session
        DB-->>A: New session created
        A-->>B: Auth successful
        B->>T: Re-auth message + commands
        T->>U: "Session renewed. Welcome back!"
    end
```

## Technical Decisions

| Decision | Options Considered | Choice | Rationale |
|----------|-------------------|--------|-----------|
| **Bot Framework** | go-telegram-bot-api, telebot, go-telegram/bot, telego | go-telegram/bot | Active maintenance, clean API, community feedback positive |
| **TON SDK** | tonutils-go, tongo, tonlib-go | tonutils-go | Complete protocol implementation, concurrent-safe, well-maintained |
| **GalaChain Language** | Rewrite SDK in Go, Use TypeScript, Use REST wrapper | TypeScript microservice | Only SDK available, proven reference implementation (gswap-bot) |
| **Inter-Service Protocol** | gRPC, REST API, Message Queue | gRPC | Type-safe, high performance, supports streaming for real-time prices |
| **Database** | SQLite, PostgreSQL, MongoDB | SQLite (dev), PostgreSQL (prod) | Simple for MVP, PostgreSQL for production scale and reliability |
| **Trade Logging** | Database only, JSONL files, Both | JSONL + Database | Append-only audit trail (JSONL), queryable data (DB) |
| **Private Key Encryption** | AES-256-GCM, RSA, HSM-only | AES-256-GCM (+ HSM prod) | Standard symmetric encryption, HSM for production security |
| **Session Management** | JWT, Database sessions, In-memory | Database sessions | Persistent across restarts, easy revocation, 24h expiry |
| **Rate Limiting** | Token bucket, Leaky bucket, Fixed window | Token bucket | Smooth burst handling, fair allocation, standard algorithm |
| **Price Caching** | No cache, Redis, In-memory TTL | In-memory TTL (5s) | Low latency, simple, sufficient for MVP |
| **Error Handling** | Panic/recover, Return errors, Both | Return errors + structured logging | Idiomatic Go, traceable errors, no silent failures |
| **Configuration** | YAML files, ENV vars, Both | ENV vars (secrets) + YAML (config) | 12-factor app, separate secrets from config |
| **Logging Format** | Plain text, JSON structured | JSON structured | Machine-parseable, correlation IDs, log aggregation ready |
| **Metrics** | Custom, Prometheus, StatsD | Prometheus | Industry standard, Kubernetes native, Grafana integration |
| **Deployment (Dev)** | Docker Compose, Kubernetes, Local binaries | Docker Compose | Simple setup, matches production topology, reproducible |
| **Deployment (Prod)** | VMs, Docker Swarm, Kubernetes | Kubernetes | Auto-scaling, self-healing, secrets management, monitoring |
| **Arbitrage Threshold** | 0.1%, 0.3%, User-configurable | 0.3% default (configurable) | Account for slippage/timing, safer than 0.1%, user can adjust |

## Component Design

### Go Telegram Bot Service

#### Package Structure

```
cmd/telegram-bot/
  └── main.go                    # Entry point, service initialization

internal/
  ├── bot/
  │   ├── bot.go                 # Bot instance, Telegram API wrapper
  │   ├── commands.go            # Command registry
  │   ├── handlers/              # Command handlers
  │   │   ├── start.go           # /start, /help
  │   │   ├── wallet.go          # /wallet, /disconnect
  │   │   ├── balance.go         # /balance
  │   │   ├── swap.go            # /swap
  │   │   ├── price.go           # /price
  │   │   ├── alert.go           # /alert
  │   │   ├── portfolio.go       # /portfolio
  │   │   ├── orders.go          # /orders
  │   │   └── arbitrage.go       # /arbitrage
  │   └── middleware/
  │       ├── auth.go            # Authentication middleware
  │       ├── ratelimit.go       # Rate limiting
  │       └── logging.go         # Request logging
  │
  ├── blockchain/
  │   ├── ton/
  │   │   ├── client.go          # TON blockchain client
  │   │   ├── wallet.go          # Wallet operations
  │   │   └── transaction.go     # Transaction builder/signer
  │   └── interfaces.go          # Blockchain client interface
  │
  ├── dex/
  │   ├── stonfi/
  │   │   ├── client.go          # ston.fi API client
  │   │   ├── swap.go            # Swap operations
  │   │   └── price.go           # Price fetching
  │   └── interfaces.go          # DEX client interface
  │
  ├── wallet/
  │   ├── manager.go             # Wallet connection manager
  │   ├── tonconnect.go          # TonConnect implementation
  │   ├── session.go             # Session persistence
  │   └── encryption.go          # Key encryption/decryption
  │
  ├── arbitrage/
  │   ├── engine.go              # Arbitrage detection engine
  │   ├── executor.go            # Trade execution coordinator
  │   ├── monitor.go             # Price monitoring loop
  │   └── calculator.go          # Profit/position calculation
  │
  ├── galachain/
  │   ├── client.go              # gRPC client for GalaChain service
  │   ├── types.go               # Request/response types
  │   └── retry.go               # Retry logic with backoff
  │
  ├── storage/
  │   ├── db.go                  # Database interface
  │   ├── sqlite.go              # SQLite implementation
  │   ├── postgres.go            # PostgreSQL implementation
  │   ├── models.go              # Data models
  │   └── migrations/            # SQL migration files
  │
  ├── logging/
  │   ├── logger.go              # Structured logger
  │   ├── trade_log.go           # JSONL trade logger
  │   └── correlation.go         # Correlation ID tracking
  │
  ├── metrics/
  │   ├── collector.go           # Prometheus metrics
  │   └── metrics.go             # Metric definitions
  │
  └── config/
      ├── config.go              # Configuration loader
      └── types.go               # Config structures
```

#### Core Components

**1. Bot Command Handler** (`internal/bot/bot.go`)

```go
type Bot struct {
    api           *bot.Bot
    db            storage.Database
    tonClient     blockchain.Client
    stonfiClient  dex.Client
    galaClient    *galachain.Client
    arbEngine     *arbitrage.Engine
    walletMgr     *wallet.Manager
    metrics       *metrics.Collector
    logger        *logging.Logger
}

func (b *Bot) Start(ctx context.Context) error {
    // Register command handlers
    b.api.RegisterHandler(bot.HandlerTypeMessageText, "/start", b.handleStart)
    b.api.RegisterHandler(bot.HandlerTypeMessageText, "/wallet", b.handleWallet)
    b.api.RegisterHandler(bot.HandlerTypeMessageText, "/balance", b.handleBalance)
    // ... more handlers

    // Start background workers
    go b.arbEngine.StartMonitoring(ctx)

    // Start bot
    return b.api.Start(ctx)
}
```

**2. TON Blockchain Client** (`internal/blockchain/ton/client.go`)

```go
type Client struct {
    liteClient *liteclient.ConnectionPool
    api        *ton.APIClient
    logger     *logging.Logger
}

func (c *Client) GetBalance(ctx context.Context, address string) (*big.Int, error)
func (c *Client) ExecuteSwap(ctx context.Context, req *SwapRequest) (*SwapResult, error)
func (c *Client) GetTransactionStatus(ctx context.Context, txHash string) (*TxStatus, error)
func (c *Client) EstimateGas(ctx context.Context, tx *Transaction) (uint64, error)
```

**3. ston.fi DEX Client** (`internal/dex/stonfi/client.go`)

```go
type Client struct {
    baseURL    string
    httpClient *http.Client
    cache      *cache.TTLCache
    logger     *logging.Logger
}

func (c *Client) SimulateSwap(ctx context.Context, req *SimulateRequest) (*SimulateResponse, error)
func (c *Client) GetSwapStatus(ctx context.Context, txHash string) (*SwapStatus, error)
func (c *Client) GetPoolInfo(ctx context.Context, poolAddress string) (*PoolInfo, error)
```

**4. Wallet Manager** (`internal/wallet/manager.go`)

```go
type Manager struct {
    db         storage.Database
    encryption *encryption.Service
    logger     *logging.Logger
}

func (m *Manager) ConnectTonWallet(ctx context.Context, userID int64) (*TonConnectSession, error)
func (m *Manager) ConnectGalaWallet(ctx context.Context, userID int64, privateKey string) error
func (m *Manager) GetWalletSession(ctx context.Context, userID int64, chain string) (*WalletSession, error)
func (m *Manager) DisconnectWallet(ctx context.Context, userID int64, chain string) error
```

**5. Arbitrage Engine** (`internal/arbitrage/engine.go`)

```go
type Engine struct {
    stonfiClient  dex.Client
    galaClient    *galachain.Client
    tonClient     blockchain.Client
    executor      *Executor
    config        *Config
    logger        *logging.Logger
}

func (e *Engine) StartMonitoring(ctx context.Context) error
func (e *Engine) DetectOpportunity(ctx context.Context) (*Opportunity, error)
func (e *Engine) ExecuteArbitrage(ctx context.Context, opp *Opportunity, userID int64) (*Result, error)
func (e *Engine) CalculatePositionSize(ctx context.Context, userID int64, direction Direction) (*PositionSize, error)
```

**6. gRPC Client for GalaChain** (`internal/galachain/client.go`)

```go
type Client struct {
    conn   *grpc.ClientConn
    client pb.GalaChainServiceClient
    retry  *retry.Config
    logger *logging.Logger
}

func (c *Client) GetPrice(ctx context.Context, pair string) (*Price, error)
func (c *Client) GetBalance(ctx context.Context, userID int64) (*Balance, error)
func (c *Client) ExecuteSwap(ctx context.Context, req *SwapRequest) (*SwapResult, error)
func (c *Client) WatchPrices(ctx context.Context) (<-chan *PriceUpdate, error)
```

### TypeScript GalaChain Service

#### Module Structure

```
galachain-service/
  ├── src/
  │   ├── index.ts               # Entry point, gRPC server initialization
  │   ├── server/
  │   │   ├── grpc.ts            # gRPC server setup
  │   │   ├── handlers.ts        # gRPC method handlers
  │   │   └── middleware.ts      # Logging, error handling
  │   │
  │   ├── gswap/
  │   │   ├── client.ts          # GSwap SDK wrapper
  │   │   ├── swap.ts            # Swap execution
  │   │   ├── price.ts           # Price fetching
  │   │   └── pool.ts            # Pool monitoring
  │   │
  │   ├── wallet/
  │   │   ├── manager.ts         # Wallet session manager
  │   │   ├── signer.ts          # Transaction signer
  │   │   ├── walletconnect.ts   # WalletConnect implementation
  │   │   └── encryption.ts      # Key encryption
  │   │
  │   ├── blockchain/
  │   │   ├── client.ts          # GalaChain client
  │   │   └── transaction.ts     # Transaction builder
  │   │
  │   ├── storage/
  │   │   ├── db.ts              # Database client
  │   │   └── models.ts          # TypeORM models
  │   │
  │   ├── logging/
  │   │   ├── logger.ts          # Structured logger
  │   │   └── trade-log.ts       # JSONL trade logger
  │   │
  │   ├── metrics/
  │   │   └── collector.ts       # Prometheus metrics
  │   │
  │   ├── config/
  │   │   └── index.ts           # Configuration loader
  │   │
  │   └── types/
  │       ├── grpc.ts            # Generated gRPC types
  │       ├── gswap.ts           # GSwap SDK types
  │       └── models.ts          # Domain models
  │
  ├── proto/                     # Symlink to ../../proto
  ├── package.json
  ├── tsconfig.json
  └── Dockerfile
```

#### Core Modules

**1. gRPC Server** (`src/server/grpc.ts`)

```typescript
import * as grpc from '@grpc/grpc-js';
import { GalaChainServiceService } from './types/grpc';
import { GrpcHandlers } from './server/handlers';

export class GrpcServer {
  private server: grpc.Server;
  private handlers: GrpcHandlers;

  constructor(
    private gswapClient: GSwapClient,
    private walletManager: WalletManager,
    private logger: Logger,
  ) {
    this.server = new grpc.Server();
    this.handlers = new GrpcHandlers(gswapClient, walletManager, logger);
  }

  async start(port: number): Promise<void> {
    this.server.addService(GalaChainServiceService, {
      GetPrice: this.handlers.getPrice.bind(this.handlers),
      GetBalance: this.handlers.getBalance.bind(this.handlers),
      ExecuteSwap: this.handlers.executeSwap.bind(this.handlers),
      WatchPrices: this.handlers.watchPrices.bind(this.handlers),
      CreateWalletSession: this.handlers.createWalletSession.bind(this.handlers),
    });

    await this.listen(port);
  }

  private listen(port: number): Promise<void> {
    return new Promise((resolve, reject) => {
      this.server.bindAsync(
        `0.0.0.0:${port}`,
        grpc.ServerCredentials.createInsecure(),
        (err, boundPort) => {
          if (err) {
            reject(err);
          } else {
            this.server.start();
            this.logger.info(`gRPC server listening on port ${boundPort}`);
            resolve();
          }
        },
      );
    });
  }
}
```

**2. GSwap Client** (`src/gswap/client.ts`)

```typescript
import { GSwapSDK } from '@gala-chain/gswap-sdk';
import { RateLimiter } from '../utils/rate-limiter';

export class GSwapClient {
  private sdk: GSwapSDK;
  private rateLimiter: RateLimiter;
  private priceCache: Map<string, CachedPrice>;

  constructor(
    private config: GSwapConfig,
    private logger: Logger,
  ) {
    this.sdk = new GSwapSDK(config);
    // 20 requests per 10 seconds
    this.rateLimiter = new RateLimiter(20, 10000);
    this.priceCache = new Map();
  }

  async getPrice(pair: string): Promise<Price> {
    // Check cache first (5s TTL)
    const cached = this.priceCache.get(pair);
    if (cached && Date.now() - cached.timestamp < 5000) {
      return cached.price;
    }

    // Rate limit
    await this.rateLimiter.acquire();

    // Fetch from gswap
    const price = await this.sdk.getPoolPrice(pair);
    this.priceCache.set(pair, { price, timestamp: Date.now() });
    return price;
  }

  async executeSwap(request: SwapRequest): Promise<SwapResult> {
    await this.rateLimiter.acquire();

    // Authorize fee credit
    await this.sdk.authorizeFee(request.userWallet);

    // Submit swap
    const result = await this.sdk.requestTokenSwap({
      from: request.fromToken,
      to: request.toToken,
      amount: request.amount,
      slippageBps: request.slippageBps,
    });

    return result;
  }

  async getBalance(walletAddress: string): Promise<Balance> {
    await this.rateLimiter.acquire();
    return this.sdk.fetchBalances(walletAddress);
  }
}
```

**3. Wallet Manager** (`src/wallet/manager.ts`)

```typescript
export class WalletManager {
  constructor(
    private db: Database,
    private encryption: EncryptionService,
    private logger: Logger,
  ) {}

  async createWalletSession(
    userId: number,
    method: 'walletconnect' | 'manual',
    credentials?: { privateKey: string; publicKey: string; address: string },
  ): Promise<WalletSession> {
    if (method === 'walletconnect') {
      return this.createWalletConnectSession(userId);
    } else if (method === 'manual' && credentials) {
      return this.createManualSession(userId, credentials);
    } else {
      throw new Error('Invalid wallet connection method');
    }
  }

  private async createWalletConnectSession(userId: number): Promise<WalletSession> {
    // Generate WalletConnect session
    const connector = new WalletConnect({
      bridge: 'https://bridge.walletconnect.org',
      qrcodeModal: QRCodeModal,
    });

    // Create session
    await connector.createSession();

    return {
      userId,
      type: 'walletconnect',
      connector,
      qrCodeUri: connector.uri,
      deepLink: `galawallet://wc?uri=${encodeURIComponent(connector.uri)}`,
    };
  }

  private async createManualSession(
    userId: number,
    credentials: WalletCredentials,
  ): Promise<WalletSession> {
    // Encrypt private key
    const encryptedKey = await this.encryption.encrypt(credentials.privateKey);

    // Store in database
    await this.db.saveWalletSession({
      userId,
      address: credentials.address,
      publicKey: credentials.publicKey,
      encryptedPrivateKey: encryptedKey,
      createdAt: new Date(),
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000), // 24h
    });

    return {
      userId,
      type: 'manual',
      address: credentials.address,
    };
  }

  async getWalletSigner(userId: number): Promise<WalletSigner> {
    const session = await this.db.getWalletSession(userId);
    if (!session) {
      throw new Error('Wallet not connected');
    }

    const privateKey = await this.encryption.decrypt(session.encryptedPrivateKey);
    return new WalletSigner(privateKey);
  }
}
```

**4. Swap Executor** (`src/gswap/swap.ts`)

```typescript
export class SwapExecutor {
  constructor(
    private gswapClient: GSwapClient,
    private walletManager: WalletManager,
    private tradeLog: TradeLogger,
    private logger: Logger,
  ) {}

  async executeSwap(request: SwapRequest): Promise<SwapResult> {
    const startTime = Date.now();

    try {
      // Get wallet signer
      const signer = await this.walletManager.getWalletSigner(request.userId);

      // Estimate fee
      const feeEstimate = await this.gswapClient.estimateFee(request);

      // Execute swap
      this.logger.info('Executing swap on gswap', {
        userId: request.userId,
        from: request.fromToken,
        to: request.toToken,
        amount: request.amount,
        estimatedFee: feeEstimate,
      });

      const result = await this.gswapClient.executeSwap({
        ...request,
        signer,
      });

      // Log trade
      await this.tradeLog.append({
        timestamp: new Date().toISOString(),
        userId: request.userId,
        type: 'swap',
        chain: 'galachain',
        fromToken: request.fromToken,
        toToken: request.toToken,
        amountIn: request.amount,
        amountOut: result.amountOut,
        fee: result.fee,
        txHash: result.txHash,
        status: 'success',
        executionTimeMs: Date.now() - startTime,
      });

      return result;
    } catch (error) {
      // Log failed trade
      await this.tradeLog.append({
        timestamp: new Date().toISOString(),
        userId: request.userId,
        type: 'swap',
        chain: 'galachain',
        fromToken: request.fromToken,
        toToken: request.toToken,
        amountIn: request.amount,
        status: 'failed',
        error: error.message,
        executionTimeMs: Date.now() - startTime,
      });

      throw error;
    }
  }
}
```

## Interface Definitions

### gRPC Protocol Buffers

```protobuf
// proto/galachain.proto
syntax = "proto3";

package galachain;

option go_package = "github.com/cchuter/telegram-trader/internal/galachain/pb";

// GalaChain service for managing GalaChain operations
service GalaChainService {
  // Get current price for a trading pair
  rpc GetPrice(GetPriceRequest) returns (PriceResponse);

  // Get wallet balance
  rpc GetBalance(BalanceRequest) returns (BalanceResponse);

  // Execute a token swap
  rpc ExecuteSwap(SwapRequest) returns (SwapResponse);

  // Watch price updates (streaming)
  rpc WatchPrices(WatchPricesRequest) returns (stream PriceUpdate);

  // Create wallet connection session
  rpc CreateWalletSession(WalletSessionRequest) returns (WalletSessionResponse);

  // Health check
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}

// Price request for a specific trading pair
message GetPriceRequest {
  string pair = 1;  // e.g., "GTON/GALA"
}

message PriceResponse {
  string pair = 1;
  string price = 2;  // Decimal string to avoid precision loss
  int64 timestamp = 3;  // Unix timestamp
  string bid = 4;  // Bid price
  string ask = 5;  // Ask price
  string volume_24h = 6;  // 24h volume
}

// Balance request for user wallet
message BalanceRequest {
  int64 user_id = 1;
}

message BalanceResponse {
  repeated TokenBalance balances = 1;
}

message TokenBalance {
  string token = 1;  // Token symbol
  string balance = 2;  // Balance as decimal string
  string usd_value = 3;  // USD value
}

// Swap execution request
message SwapRequest {
  int64 user_id = 1;
  string from_token = 2;
  string to_token = 3;
  string amount = 4;  // Amount as decimal string
  int32 slippage_bps = 5;  // Slippage in basis points (100 bps = 1%)
  int32 fee_tier = 6;  // Optional fee tier override
}

message SwapResponse {
  string tx_hash = 1;
  string amount_in = 2;
  string amount_out = 3;
  string fee = 4;
  string status = 5;  // "pending", "success", "failed"
  string error_message = 6;  // If status = "failed"
}

// Watch prices request
message WatchPricesRequest {
  repeated string pairs = 1;  // Pairs to watch
  int32 interval_ms = 2;  // Update interval in milliseconds
}

message PriceUpdate {
  string pair = 1;
  string price = 2;
  int64 timestamp = 3;
}

// Wallet session creation
message WalletSessionRequest {
  int64 user_id = 1;
  string method = 2;  // "walletconnect" or "manual"

  // For manual method
  optional string private_key = 3;
  optional string public_key = 4;
  optional string address = 5;
}

message WalletSessionResponse {
  string session_id = 1;
  string method = 2;

  // For WalletConnect
  optional string qr_code_uri = 3;
  optional string deep_link = 4;

  // For manual
  optional string address = 5;

  bool success = 6;
  string error_message = 7;
}

// Health check
message HealthCheckRequest {}

message HealthCheckResponse {
  string status = 1;  // "healthy", "degraded", "unhealthy"
  map<string, string> dependencies = 2;  // Dependency status
  int64 uptime_seconds = 3;
}

// Error codes
enum ErrorCode {
  UNKNOWN = 0;
  INVALID_REQUEST = 1;
  WALLET_NOT_CONNECTED = 2;
  INSUFFICIENT_BALANCE = 3;
  SLIPPAGE_EXCEEDED = 4;
  RATE_LIMIT_EXCEEDED = 5;
  NETWORK_ERROR = 6;
  TRANSACTION_FAILED = 7;
}
```

### Go Interfaces

```go
// internal/blockchain/interfaces.go
package blockchain

import (
    "context"
    "math/big"
)

// Client represents a blockchain client
type Client interface {
    // GetBalance returns the balance for an address
    GetBalance(ctx context.Context, address string, token string) (*big.Int, error)

    // ExecuteSwap executes a token swap
    ExecuteSwap(ctx context.Context, req *SwapRequest) (*SwapResult, error)

    // GetTransactionStatus checks transaction status
    GetTransactionStatus(ctx context.Context, txHash string) (*TxStatus, error)

    // EstimateGas estimates gas cost for a transaction
    EstimateGas(ctx context.Context, tx *Transaction) (uint64, error)
}

// SwapRequest represents a swap request
type SwapRequest struct {
    FromAddress  string
    FromToken    string
    ToToken      string
    Amount       *big.Int
    SlippageBps  int32
    MaxGas       uint64
}

// SwapResult represents a swap result
type SwapResult struct {
    TxHash      string
    AmountIn    *big.Int
    AmountOut   *big.Int
    Fee         *big.Int
    Status      TxStatus
}

// TxStatus represents transaction status
type TxStatus string

const (
    TxStatusPending TxStatus = "pending"
    TxStatusSuccess TxStatus = "success"
    TxStatusFailed  TxStatus = "failed"
)
```

```go
// internal/dex/interfaces.go
package dex

import (
    "context"
    "math/big"
)

// Client represents a DEX client
type Client interface {
    // SimulateSwap simulates a swap and returns expected output
    SimulateSwap(ctx context.Context, req *SimulateRequest) (*SimulateResponse, error)

    // GetPrice returns current price for a pair
    GetPrice(ctx context.Context, pair string) (*Price, error)

    // GetPoolInfo returns pool information
    GetPoolInfo(ctx context.Context, poolAddress string) (*PoolInfo, error)
}

// SimulateRequest represents a swap simulation request
type SimulateRequest struct {
    FromToken string
    ToToken   string
    Amount    *big.Int
}

// SimulateResponse represents simulation result
type SimulateResponse struct {
    AmountOut   *big.Int
    Fee         *big.Int
    SlippageBps int32
    PriceImpact float64
}

// Price represents a token pair price
type Price struct {
    Pair        string
    Price       *big.Float
    Bid         *big.Float
    Ask         *big.Float
    Volume24h   *big.Float
    Timestamp   int64
}

// PoolInfo represents DEX pool information
type PoolInfo struct {
    Address     string
    Token0      string
    Token1      string
    Reserve0    *big.Int
    Reserve1    *big.Int
    TotalLiquidity *big.Int
}
```

```go
// internal/wallet/interfaces.go
package wallet

import (
    "context"
)

// Manager manages wallet connections and sessions
type Manager interface {
    // ConnectTonWallet initiates TonConnect flow
    ConnectTonWallet(ctx context.Context, userID int64) (*TonConnectSession, error)

    // ConnectGalaWallet connects Gala wallet
    ConnectGalaWallet(ctx context.Context, userID int64, privateKey string) error

    // GetWalletSession retrieves active session
    GetWalletSession(ctx context.Context, userID int64, chain string) (*WalletSession, error)

    // DisconnectWallet disconnects and clears session
    DisconnectWallet(ctx context.Context, userID int64, chain string) error
}

// TonConnectSession represents a TonConnect session
type TonConnectSession struct {
    SessionID string
    QRCodeURL string
    DeepLink  string
    ExpiresAt time.Time
}

// WalletSession represents an active wallet session
type WalletSession struct {
    UserID    int64
    Chain     string
    Address   string
    PublicKey string
    CreatedAt time.Time
    ExpiresAt time.Time
}
```

```go
// internal/arbitrage/interfaces.go
package arbitrage

import (
    "context"
)

// Engine detects and executes arbitrage opportunities
type Engine interface {
    // StartMonitoring starts background price monitoring
    StartMonitoring(ctx context.Context) error

    // DetectOpportunity checks for arbitrage opportunity
    DetectOpportunity(ctx context.Context) (*Opportunity, error)

    // ExecuteArbitrage executes an arbitrage trade
    ExecuteArbitrage(ctx context.Context, opp *Opportunity, userID int64) (*Result, error)
}

// Opportunity represents an arbitrage opportunity
type Opportunity struct {
    Direction   Direction  // BuyTonSellGala or BuyGalaSellTon
    Spread      float64    // Spread percentage
    ProfitUSD   float64    // Estimated profit in USD
    BuyAmount   *big.Int   // Amount to buy
    SellAmount  *big.Int   // Amount to sell
    Timestamp   time.Time
}

// Direction represents arbitrage direction
type Direction int

const (
    BuyTonSellGala Direction = iota  // Buy on ston.fi, sell on gswap
    BuyGalaSellTon                    // Buy on gswap, sell on ston.fi
)

// Result represents arbitrage execution result
type Result struct {
    BuyTxHash   string
    SellTxHash  string
    ActualProfit float64
    TotalFees   *big.Int
    Status      Status
}

// Status represents execution status
type Status int

const (
    StatusSuccess Status = iota
    StatusPartialSuccess  // One leg succeeded
    StatusFailed
)
```

### TypeScript Interfaces

```typescript
// galachain-service/src/types/models.ts

export interface GalaChainClient {
  /**
   * Get price for a trading pair
   */
  getPrice(pair: string): Promise<Price>;

  /**
   * Get wallet balance
   */
  getBalance(walletAddress: string): Promise<Balance>;

  /**
   * Execute token swap
   */
  executeSwap(request: SwapRequest): Promise<SwapResult>;

  /**
   * Estimate swap fee
   */
  estimateFee(request: SwapRequest): Promise<string>;
}

export interface Price {
  pair: string;
  price: string;  // Decimal string
  bid?: string;
  ask?: string;
  volume24h?: string;
  timestamp: number;
}

export interface Balance {
  tokens: TokenBalance[];
}

export interface TokenBalance {
  token: string;
  balance: string;
  usdValue?: string;
}

export interface SwapRequest {
  userId: number;
  fromToken: string;
  toToken: string;
  amount: string;
  slippageBps: number;
  feeTier?: number;
  signer?: WalletSigner;
}

export interface SwapResult {
  txHash: string;
  amountIn: string;
  amountOut: string;
  fee: string;
  status: 'pending' | 'success' | 'failed';
  errorMessage?: string;
}

export interface WalletSigner {
  /**
   * Sign a transaction
   */
  signTransaction(tx: Transaction): Promise<SignedTransaction>;

  /**
   * Get wallet address
   */
  getAddress(): string;

  /**
   * Get public key
   */
  getPublicKey(): string;
}

export interface WalletSession {
  userId: number;
  type: 'walletconnect' | 'manual';
  address?: string;
  connector?: any;  // WalletConnect connector
  qrCodeUri?: string;
  deepLink?: string;
}

export interface TradeLogEntry {
  timestamp: string;
  userId: number;
  type: 'swap' | 'arbitrage';
  chain: 'ton' | 'galachain';
  fromToken: string;
  toToken: string;
  amountIn: string;
  amountOut?: string;
  fee?: string;
  txHash?: string;
  status: 'success' | 'failed' | 'pending';
  error?: string;
  executionTimeMs: number;
}
```

## File Structure

```
telegram-trader/
├── cmd/
│   ├── telegram-bot/
│   │   └── main.go                        # Bot service entry point
│   └── galachain-service/
│       └── main.ts                        # GalaChain service entry point (symlink to ../galachain-service/src/index.ts)
│
├── internal/                              # Go packages (private)
│   ├── bot/                               # Telegram bot logic
│   ├── blockchain/                        # Blockchain clients (TON)
│   ├── dex/                               # DEX integrations (ston.fi)
│   ├── wallet/                            # Wallet management
│   ├── arbitrage/                         # Arbitrage engine
│   ├── galachain/                         # gRPC client for GalaChain service
│   ├── storage/                           # Database and persistence
│   ├── logging/                           # Logging utilities
│   ├── metrics/                           # Metrics collection
│   └── config/                            # Configuration
│
├── galachain-service/                     # TypeScript service
│   ├── src/
│   │   ├── index.ts                       # Entry point
│   │   ├── server/                        # gRPC server
│   │   ├── gswap/                         # GSwap SDK integration
│   │   ├── wallet/                        # Wallet management
│   │   ├── blockchain/                    # GalaChain client
│   │   ├── storage/                       # Database
│   │   ├── logging/                       # Logging
│   │   ├── metrics/                       # Metrics
│   │   ├── config/                        # Configuration
│   │   └── types/                         # TypeScript types
│   ├── package.json
│   ├── tsconfig.json
│   ├── Dockerfile
│   └── .env.example
│
├── proto/                                 # gRPC Protocol Buffer definitions
│   ├── galachain.proto                    # Main service definition
│   └── README.md                          # Proto documentation
│
├── configs/                               # Configuration files
│   ├── bot-service.yaml                   # Bot service config
│   ├── galachain-service.yaml             # GalaChain service config
│   ├── development.yaml                   # Dev environment overrides
│   └── production.yaml                    # Prod environment overrides
│
├── scripts/                               # Utility scripts
│   ├── setup.sh                           # Initial setup
│   ├── generate-proto.sh                  # Generate gRPC code
│   ├── migrate.sh                         # Run database migrations
│   ├── deploy.sh                          # Deployment script
│   └── backup.sh                          # Backup trade logs and DB
│
├── migrations/                            # SQL migrations
│   ├── 001_initial_schema.up.sql
│   ├── 001_initial_schema.down.sql
│   ├── 002_add_wallet_sessions.up.sql
│   └── 002_add_wallet_sessions.down.sql
│
├── deployments/                           # Deployment configurations
│   ├── docker-compose.yml                 # Local development
│   ├── docker-compose.prod.yml            # Production compose
│   └── kubernetes/                        # Kubernetes manifests
│       ├── bot-service.yaml
│       ├── galachain-service.yaml
│       ├── postgres.yaml
│       ├── prometheus.yaml
│       └── grafana.yaml
│
├── monitoring/                            # Monitoring configuration
│   ├── prometheus.yml                     # Prometheus config
│   └── grafana/                           # Grafana dashboards
│       ├── dashboards/
│       │   ├── bot-overview.json
│       │   ├── arbitrage-performance.json
│       │   └── system-health.json
│       └── provisioning/
│
├── logs/                                  # Log directory (gitignored)
│   ├── bot-service.log
│   ├── galachain-service.log
│   └── trades.jsonl                       # Append-only trade log
│
├── test/                                  # Integration tests
│   ├── e2e/                               # End-to-end tests
│   ├── integration/                       # Integration tests
│   └── fixtures/                          # Test fixtures
│
├── docs/                                  # Documentation
│   ├── API.md                             # API documentation
│   ├── DEPLOYMENT.md                      # Deployment guide
│   ├── SECURITY.md                        # Security guidelines
│   └── DEVELOPMENT.md                     # Development setup
│
├── .github/                               # GitHub configuration
│   └── workflows/
│       ├── ci.yml                         # CI pipeline
│       ├── security.yml                   # Security scanning
│       └── deploy.yml                     # Deployment pipeline
│
├── go.mod                                 # Go dependencies
├── go.sum
├── Makefile                               # Common commands
├── .env.example                           # Example environment variables
├── .gitignore
├── LICENSE
└── README.md                              # Project overview

```

**Key Directories Explained**:

- **`cmd/`**: Application entry points (Go and TypeScript binaries)
- **`internal/`**: Go packages (not importable by external projects)
- **`galachain-service/`**: Complete TypeScript service for GalaChain
- **`proto/`**: gRPC definitions shared between services
- **`configs/`**: YAML configuration files (non-sensitive)
- **`scripts/`**: Shell scripts for development and deployment
- **`migrations/`**: SQL migration files for database schema
- **`deployments/`**: Docker and Kubernetes deployment configs
- **`monitoring/`**: Prometheus and Grafana configurations
- **`logs/`**: Runtime logs and trade history (gitignored, backed up)
- **`test/`**: Test files (integration and end-to-end)
- **`docs/`**: Project documentation

## Data Models

### User Data

**User Session** (SQLite/PostgreSQL)

```sql
CREATE TABLE user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    telegram_username VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    whitelist_status VARCHAR(50) NOT NULL DEFAULT 'pending'  -- 'pending', 'approved', 'denied'
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
```

**Wallet Connections** (SQLite/PostgreSQL)

```sql
CREATE TABLE wallet_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    chain VARCHAR(50) NOT NULL,  -- 'ton' or 'galachain'
    wallet_address VARCHAR(255) NOT NULL,
    public_key TEXT,
    encrypted_private_key TEXT,  -- AES-256-GCM encrypted (for manual connections)
    connection_method VARCHAR(50) NOT NULL,  -- 'tonconnect', 'walletconnect', 'manual'
    session_data JSONB,  -- TonConnect/WalletConnect session data
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE(user_id, chain)
);

CREATE INDEX idx_wallet_sessions_user_id ON wallet_sessions(user_id);
CREATE INDEX idx_wallet_sessions_chain ON wallet_sessions(chain);
```

**User Preferences** (SQLite/PostgreSQL)

```sql
CREATE TABLE user_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    arbitrage_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    arbitrage_threshold_bps INT NOT NULL DEFAULT 30,  -- 0.3%
    max_position_size_usd DECIMAL(20, 2),
    daily_loss_limit_usd DECIMAL(20, 2),
    notification_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    slippage_tolerance_bps INT NOT NULL DEFAULT 500,  -- 5%
    preferences_json JSONB,  -- Additional preferences
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_sessions(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
```

### Trading Data

**Trade History** (SQLite/PostgreSQL)

```sql
CREATE TABLE trade_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    trade_type VARCHAR(50) NOT NULL,  -- 'swap', 'arbitrage'
    chain VARCHAR(50) NOT NULL,  -- 'ton', 'galachain', 'both'
    from_token VARCHAR(50) NOT NULL,
    to_token VARCHAR(50) NOT NULL,
    amount_in DECIMAL(40, 18) NOT NULL,
    amount_out DECIMAL(40, 18),
    fee DECIMAL(40, 18),
    tx_hash_ton VARCHAR(255),
    tx_hash_gala VARCHAR(255),
    status VARCHAR(50) NOT NULL,  -- 'pending', 'success', 'failed', 'partial'
    error_message TEXT,
    execution_time_ms INT,
    profit_usd DECIMAL(20, 2),  -- For arbitrage trades
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_sessions(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_trade_history_user_id ON trade_history(user_id);
CREATE INDEX idx_trade_history_created_at ON trade_history(created_at);
CREATE INDEX idx_trade_history_status ON trade_history(status);
```

**Price Alerts** (SQLite/PostgreSQL)

```sql
CREATE TABLE price_alerts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    pair VARCHAR(50) NOT NULL,  -- 'TON/GALA', 'GTON/GALA'
    alert_type VARCHAR(50) NOT NULL,  -- 'spread', 'price_above', 'price_below'
    threshold_value DECIMAL(20, 8) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    triggered_count INT NOT NULL DEFAULT 0,
    last_triggered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,  -- Auto-expire after 24h
    FOREIGN KEY (user_id) REFERENCES user_sessions(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_price_alerts_user_id ON price_alerts(user_id);
CREATE INDEX idx_price_alerts_is_active ON price_alerts(is_active);
```

**JSONL Trade Logs** (Append-Only File)

```jsonl
{"timestamp":"2026-01-16T14:30:00.123Z","userId":12345,"type":"swap","chain":"ton","fromToken":"TON","toToken":"GALA","amountIn":"1.0","amountOut":"850.5","fee":"0.02","txHash":"abc123","status":"success","executionTimeMs":3450}
{"timestamp":"2026-01-16T14:35:00.456Z","userId":12345,"type":"arbitrage","chain":"both","direction":"buy_ton_sell_gala","buyTxHash":"def456","sellTxHash":"ghi789","buyAmount":"5.0","sellAmount":"4250.0","profitUSD":"0.12","totalFees":"0.15","status":"success","executionTimeMs":8920}
{"timestamp":"2026-01-16T14:40:00.789Z","userId":67890,"type":"swap","chain":"galachain","fromToken":"GTON","toToken":"GALA","amountIn":"2.0","status":"failed","error":"Insufficient balance","executionTimeMs":1200}
```

### Configuration

**Environment Variables** (`.env`)

```bash
# Bot Service
BOT_TOKEN=your_telegram_bot_token
BOT_WEBHOOK_URL=https://your-domain.com/webhook  # Optional, for webhook mode
BOT_ADMIN_USER_IDS=12345,67890  # Comma-separated admin user IDs

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/telegram_trader
# or for SQLite
# DATABASE_URL=sqlite://./data/telegram-trader.db

# Encryption
ENCRYPTION_KEY=your-32-byte-base64-encoded-key  # AES-256 key
ENCRYPTION_ALGORITHM=AES-256-GCM

# TON Blockchain
TON_NETWORK=mainnet  # or testnet
TON_RPC_ENDPOINTS=https://toncenter.com/api/v2/jsonRPC,https://tonapi.io/v2
TONCONNECT_BRIDGE_URL=https://bridge.tonapi.io/bridge

# ston.fi
STONFI_API_URL=https://api.ston.fi
STONFI_GALA_TOKEN_ADDRESS=EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV

# GalaChain Service
GALACHAIN_SERVICE_URL=localhost:50051
GALACHAIN_SERVICE_TIMEOUT_MS=5000

# CoinGecko
COINGECKO_API_URL=https://api.coingecko.com/api/v3
COINGECKO_API_KEY=your_api_key  # Optional, for higher rate limits

# Monitoring
PROMETHEUS_PORT=9090
METRICS_ENABLED=true
LOG_LEVEL=info  # debug, info, warn, error
LOG_FORMAT=json  # json or text

# Arbitrage
ARBITRAGE_ENABLED=false  # Default disabled, users opt-in
ARBITRAGE_DEFAULT_THRESHOLD_BPS=30  # 0.3%
ARBITRAGE_MIN_PROFIT_USD=0.10
ARBITRAGE_MAX_POSITION_SIZE_USD=1000
```

**GalaChain Service Environment** (`.env`)

```bash
# GalaChain
GALACHAIN_NETWORK=mainnet  # or testnet
GALA_GTON_TOKEN_ADDRESS=2954bd1e38c2dff11a2ee8798e2f59b99c206e64bef7bdf5c4ac9a91315e0d34

# gswap
GSWAP_API_URL=https://api-galaswap.gala.com
GSWAP_RATE_LIMIT_REQUESTS=20
GSWAP_RATE_LIMIT_WINDOW_MS=10000

# WalletConnect
WALLETCONNECT_BRIDGE_URL=https://bridge.walletconnect.org
WALLETCONNECT_PROJECT_ID=your_project_id

# Database (shared with bot service)
DATABASE_URL=postgres://user:pass@localhost:5432/telegram_trader

# Encryption (same key as bot service)
ENCRYPTION_KEY=your-32-byte-base64-encoded-key

# gRPC
GRPC_PORT=50051
GRPC_MAX_CONNECTIONS=100

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Metrics
PROMETHEUS_PORT=9091
```

**YAML Configuration** (`configs/bot-service.yaml`)

```yaml
service:
  name: telegram-bot
  version: 1.0.0
  environment: production

bot:
  polling_timeout_seconds: 30
  max_concurrent_updates: 100
  allowed_updates:
    - message
    - callback_query
    - inline_query

rate_limiting:
  enabled: true
  commands_per_minute: 10
  burst_size: 3
  cleanup_interval_minutes: 5

arbitrage:
  monitor_interval_seconds: 3  # Check prices every 3 seconds
  execution_timeout_seconds: 10
  min_balance_ton: 1.0
  min_balance_gala: 10.0
  safety_margin_multiplier: 1.5  # Keep 1.5x minimum balance
  max_retries: 3
  circuit_breaker:
    enabled: true
    failure_threshold: 3
    timeout_seconds: 300  # 5 minutes

wallet:
  session_expiry_hours: 24
  tonconnect_timeout_seconds: 300  # 5 minutes for connection
  walletconnect_timeout_seconds: 300

trading:
  default_slippage_bps: 500  # 5%
  max_slippage_bps: 1000  # 10%
  transaction_timeout_seconds: 30
  price_cache_ttl_seconds: 5

monitoring:
  health_check_interval_seconds: 30
  metrics_path: /metrics
  log_all_commands: true
  log_all_trades: true
```

## Error Handling

| Category | Error Code | User Message | Retry Logic | Logging |
|----------|-----------|--------------|-------------|---------|
| **Authentication** | `AUTH_001` | "Access denied. Contact @admin for access." | No retry | Log user_id, timestamp |
| **Authentication** | `AUTH_002` | "Session expired. Please use /start to reconnect." | No retry | Log session expiry |
| **Rate Limiting** | `RATE_001` | "Too many requests. Please wait 60 seconds." | No retry, show cooldown | Log user_id, count |
| **Wallet Connection** | `WALLET_001` | "Wallet not connected. Use /wallet to connect." | No retry | Log chain, user_id |
| **Wallet Connection** | `WALLET_002` | "Wallet connection timeout. Please try again." | Retry 3x with backoff | Log timeout details |
| **Wallet Connection** | `WALLET_003` | "Invalid wallet address format." | No retry | Log address (sanitized) |
| **Balance Check** | `BALANCE_001` | "Insufficient balance. Required: {X} {TOKEN}, Available: {Y}" | No retry | Log amounts |
| **Balance Check** | `BALANCE_002` | "Failed to fetch balance. Please try again." | Retry 3x | Log error, chain |
| **Swap Simulation** | `SWAP_001` | "Failed to simulate swap. Exchange may be unavailable." | Retry 2x | Log API error |
| **Swap Execution** | `SWAP_002` | "Slippage too high ({X}%). Max allowed: {Y}%. Adjust amount or slippage." | No retry | Log slippage values |
| **Swap Execution** | `SWAP_003` | "Transaction failed. Reason: {REASON}" | No retry | Log full error context |
| **Swap Execution** | `SWAP_004` | "Transaction timeout. Check /orders for status." | No retry (manual check) | Log tx hash |
| **Swap Execution** | `SWAP_005` | "Minimum balance not met. Must keep {MIN} {TOKEN}." | No retry | Log min balance check |
| **Price Check** | `PRICE_001` | "Failed to fetch price from {EXCHANGE}. Retrying..." | Retry 3x | Log exchange, error |
| **Price Check** | `PRICE_002` | "Price data unavailable. Please try again later." | No retry after retries exhausted | Log final error |
| **Arbitrage** | `ARB_001` | "No arbitrage opportunity found (spread: {X}% < threshold: {Y}%)." | N/A (informational) | Debug log only |
| **Arbitrage** | `ARB_002` | "Arbitrage failed: {LEG} leg unsuccessful. Reason: {REASON}" | No retry | Log both legs status |
| **Arbitrage** | `ARB_003` | "Arbitrage partially complete. Buy succeeded, sell failed. Manual intervention needed." | No retry (alert admin) | Log critical, alert |
| **GalaChain Service** | `GALA_001` | "GalaChain service unavailable. TON operations still available." | Retry 5x with exp backoff | Log gRPC error |
| **GalaChain Service** | `GALA_002` | "gswap rate limit reached. Please wait {X} seconds." | Wait and retry | Log rate limit hit |
| **Database** | `DB_001` | "Failed to save data. Please try again." | Retry 3x | Log query, error |
| **Database** | `DB_002` | "Data retrieval failed. Please try again later." | Retry 3x | Log query, error |
| **Network** | `NET_001` | "Network error. Retrying..." | Retry 3x with exp backoff | Log endpoint, error |
| **Network** | `NET_002` | "Service timeout. Please try again." | Retry 2x | Log timeout duration |
| **Validation** | `VAL_001` | "Invalid command format. Use /help for examples." | No retry | Log command |
| **Validation** | `VAL_002` | "Invalid amount. Must be positive number." | No retry | Log input |
| **Validation** | `VAL_003` | "Invalid token symbol. Supported: TON, GALA, GTON." | No retry | Log input |
| **Encryption** | `ENC_001` | "Failed to decrypt wallet key. Please reconnect wallet." | No retry (critical) | Log error, alert admin |
| **Unknown** | `UNK_001` | "An unexpected error occurred. Support has been notified. Error ID: {CORRELATION_ID}" | No retry | Log full stack trace |

**Retry Logic Strategy**:
- **Exponential Backoff**: Initial delay 1s, multiply by 2 each retry, max 10s
- **Jitter**: Add random 0-500ms to prevent thundering herd
- **Circuit Breaker**: After 3 consecutive failures, stop retries for 5 minutes
- **User Notification**: Show "Retrying..." for user-facing operations

**Logging Standards**:
- All errors include: `correlation_id`, `user_id`, `timestamp`, `error_code`, `message`, `stack_trace` (for errors)
- Trade failures log: `trade_id`, `amounts`, `tokens`, `tx_hashes`, `failure_stage`
- Sensitive data (private keys, full addresses) NEVER logged
- Wallet addresses logged truncated: `UQ...abc` (first 2 + last 3 chars)

## Security Design

### Private Key Encryption

**Algorithm**: AES-256-GCM (Galois/Counter Mode)

**Implementation**:

```go
// internal/wallet/encryption.go
package wallet

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
)

type EncryptionService struct {
    masterKey []byte  // 32 bytes for AES-256
}

// Encrypt encrypts plaintext with AES-256-GCM
func (e *EncryptionService) Encrypt(plaintext string) (string, error) {
    block, err := aes.NewCipher(e.masterKey)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    // Generate random nonce
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    // Encrypt and append nonce
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

    // Base64 encode for storage
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext with AES-256-GCM
func (e *EncryptionService) Decrypt(ciphertext string) (string, error) {
    // Base64 decode
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(e.masterKey)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    // Extract nonce
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", fmt.Errorf("ciphertext too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]

    // Decrypt
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}
```

**Key Management**:
- **Development**: Master key stored in environment variable (`ENCRYPTION_KEY`)
- **Production**: Master key stored in HashiCorp Vault or Azure Key Vault
- **Rotation**: Support key rotation with versioned keys (`key_v1`, `key_v2`)
- **Access**: Only bot service and GalaChain service have access to master key
- **Never Stored**: Encryption key never written to database or logs

### Session Management

**Session Lifecycle**:
1. User sends `/start` command
2. Auth middleware checks whitelist
3. Create session with 24-hour expiry
4. Store session in database with encrypted token
5. Session renewed on each command (sliding window)
6. Session expires after 24 hours of inactivity
7. User must `/start` again to create new session

**Session Token**: UUID v4 (128-bit random)

**Implementation**:

```go
// internal/bot/middleware/auth.go
type SessionManager struct {
    db storage.Database
}

func (sm *SessionManager) Authenticate(ctx context.Context, userID int64) (*Session, error) {
    session, err := sm.db.GetUserSession(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("session not found: %w", err)
    }

    // Check expiry
    if time.Now().After(session.ExpiresAt) {
        return nil, fmt.Errorf("session expired")
    }

    // Check whitelist
    if session.WhitelistStatus != "approved" {
        return nil, fmt.Errorf("user not whitelisted")
    }

    // Renew session (sliding window)
    session.LastActiveAt = time.Now()
    session.ExpiresAt = time.Now().Add(24 * time.Hour)
    sm.db.UpdateUserSession(ctx, session)

    return session, nil
}
```

### Rate Limiting Implementation

**Algorithm**: Token Bucket

**Limits**:
- **Commands**: 10 per minute per user
- **Burst**: 3 commands (instant)
- **Cooldown**: 60 seconds after hitting limit

**Implementation**:

```go
// internal/bot/middleware/ratelimit.go
type RateLimiter struct {
    buckets sync.Map  // map[int64]*TokenBucket
}

type TokenBucket struct {
    tokens     int
    maxTokens  int
    refillRate time.Duration
    lastRefill time.Time
    mu         sync.Mutex
}

func (rl *RateLimiter) Allow(userID int64) bool {
    bucket := rl.getBucket(userID)
    return bucket.Take()
}

func (tb *TokenBucket) Take() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()

    // Refill tokens based on time elapsed
    now := time.Now()
    elapsed := now.Sub(tb.lastRefill)
    tokensToAdd := int(elapsed / tb.refillRate)

    if tokensToAdd > 0 {
        tb.tokens = min(tb.tokens+tokensToAdd, tb.maxTokens)
        tb.lastRefill = now
    }

    // Check if token available
    if tb.tokens > 0 {
        tb.tokens--
        return true
    }

    return false
}
```

### Input Validation

**Wallet Addresses**:
- TON: Validate EQ/UQ prefix, base64 encoding, checksum
- GalaChain: Validate hex format, length (40 chars)

**Amounts**:
- Positive numbers only
- Max 18 decimal places
- Range check (min: 0.000001, max: 1,000,000)

**Token Symbols**:
- Whitelist: TON, GALA, GTON
- Reject all other inputs

**Implementation**:

```go
// internal/bot/validation.go
func ValidateTonAddress(address string) error {
    if !strings.HasPrefix(address, "EQ") && !strings.HasPrefix(address, "UQ") {
        return fmt.Errorf("invalid TON address prefix")
    }

    // Decode base64
    decoded, err := base64.StdEncoding.DecodeString(address[2:])
    if err != nil {
        return fmt.Errorf("invalid base64 encoding")
    }

    // Verify checksum (CRC16)
    if !verifyTonChecksum(decoded) {
        return fmt.Errorf("invalid checksum")
    }

    return nil
}

func ValidateGalaAddress(address string) error {
    if len(address) != 64 {
        return fmt.Errorf("invalid GalaChain address length")
    }

    if !isHex(address) {
        return fmt.Errorf("address must be hexadecimal")
    }

    return nil
}

func ValidateAmount(amount string) (*big.Float, error) {
    value, ok := new(big.Float).SetString(amount)
    if !ok {
        return nil, fmt.Errorf("invalid number format")
    }

    if value.Sign() <= 0 {
        return nil, fmt.Errorf("amount must be positive")
    }

    // Check decimal places
    if countDecimals(amount) > 18 {
        return nil, fmt.Errorf("max 18 decimal places")
    }

    return value, nil
}
```

### Audit Logging

**What to Log**:
- All user commands with parameters (sanitized)
- All wallet connections/disconnections
- All trade executions (swap, arbitrage)
- All authentication attempts (success and failure)
- All errors with full context
- All admin actions

**Log Format** (JSON):

```json
{
  "timestamp": "2026-01-16T14:30:00.123Z",
  "level": "info",
  "correlation_id": "abc-123-def-456",
  "service": "telegram-bot",
  "event_type": "command_executed",
  "user_id": 12345,
  "username": "alice",
  "command": "/swap",
  "parameters": {
    "amount": "1.0",
    "from_token": "TON",
    "to_token": "GALA",
    "exchange": "stonfi"
  },
  "result": "success",
  "execution_time_ms": 3450
}
```

**Retention**:
- Development: 7 days
- Production: 12 months (compliance)
- Critical events (trades): Archive indefinitely (JSONL)

## Testing Strategy

### Unit Tests

**Go Testing Approach**:

```go
// internal/arbitrage/calculator_test.go
package arbitrage_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/cchuter/telegram-trader/internal/arbitrage"
)

func TestCalculatePositionSize(t *testing.T) {
    tests := []struct {
        name           string
        balance        float64
        minBalance     float64
        safetyMargin   float64
        expectedSize   float64
        expectedError  error
    }{
        {
            name:         "sufficient balance",
            balance:      100,
            minBalance:   10,
            safetyMargin: 1.5,
            expectedSize: 50,  // (100 - 10*1.5) / 2 = 42.5, rounds to 50% of usable
            expectedError: nil,
        },
        {
            name:         "insufficient balance",
            balance:      10,
            minBalance:   10,
            safetyMargin: 1.5,
            expectedSize: 0,
            expectedError: arbitrage.ErrInsufficientBalance,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            calc := arbitrage.NewCalculator(tt.minBalance, tt.safetyMargin)
            size, err := calc.CalculatePositionSize(tt.balance)

            assert.Equal(t, tt.expectedError, err)
            assert.InDelta(t, tt.expectedSize, size, 0.01)
        })
    }
}
```

**TypeScript Testing (Jest)**:

```typescript
// galachain-service/src/gswap/__tests__/client.test.ts
import { GSwapClient } from '../client';
import { RateLimiter } from '../../utils/rate-limiter';

jest.mock('../../utils/rate-limiter');

describe('GSwapClient', () => {
  let client: GSwapClient;
  let mockRateLimiter: jest.Mocked<RateLimiter>;

  beforeEach(() => {
    mockRateLimiter = new RateLimiter(20, 10000) as jest.Mocked<RateLimiter>;
    client = new GSwapClient(config, logger);
  });

  describe('getPrice', () => {
    it('should return cached price if available', async () => {
      // Set up cache
      client['priceCache'].set('GTON/GALA', {
        price: { pair: 'GTON/GALA', price: '855', timestamp: Date.now() },
        timestamp: Date.now(),
      });

      const price = await client.getPrice('GTON/GALA');

      expect(price.price).toBe('855');
      expect(mockRateLimiter.acquire).not.toHaveBeenCalled();
    });

    it('should fetch from API if cache expired', async () => {
      // Set up expired cache
      client['priceCache'].set('GTON/GALA', {
        price: { pair: 'GTON/GALA', price: '855', timestamp: Date.now() - 10000 },
        timestamp: Date.now() - 10000,
      });

      mockRateLimiter.acquire.mockResolvedValue(undefined);
      const mockSdk = client['sdk'];
      jest.spyOn(mockSdk, 'getPoolPrice').mockResolvedValue({ price: '860' });

      const price = await client.getPrice('GTON/GALA');

      expect(price.price).toBe('860');
      expect(mockRateLimiter.acquire).toHaveBeenCalled();
    });
  });
});
```

**Mock Strategies**:
- **Blockchain Clients**: Mock with test fixtures (balance, transaction responses)
- **gRPC**: Use `grpc-mock` library for TypeScript, custom mocks for Go
- **External APIs**: Use `httptest` (Go) and `nock` (TypeScript) for HTTP mocking
- **Database**: Use in-memory SQLite for tests, reset between tests
- **Time**: Inject time interface for testable time-dependent logic

### Integration Tests

**gRPC Communication Tests** (`test/integration/grpc_test.go`):

```go
func TestGrpcPriceFlow(t *testing.T) {
    // Start real gRPC server (GalaChain service)
    server := startTestGrpcServer(t)
    defer server.Stop()

    // Create gRPC client
    client := galachain.NewClient(server.Address())

    // Test GetPrice
    ctx := context.Background()
    price, err := client.GetPrice(ctx, "GTON/GALA")

    assert.NoError(t, err)
    assert.NotNil(t, price)
    assert.Equal(t, "GTON/GALA", price.Pair)
    assert.NotEmpty(t, price.Price)
}
```

**Blockchain Interaction Tests** (Testnet):

```go
// Uses TON testnet and GalaChain testnet
func TestTonSwapExecution(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping testnet integration test")
    }

    // Create testnet client
    client := ton.NewClient(ton.ConfigTestnet)

    // Create test wallet with test funds
    wallet := createTestWallet(t)

    // Execute small swap
    result, err := client.ExecuteSwap(ctx, &blockchain.SwapRequest{
        FromAddress: wallet.Address,
        FromToken:   "TON",
        ToToken:     "GALA",
        Amount:      big.NewInt(100000), // 0.0001 TON
        SlippageBps: 500,
    })

    assert.NoError(t, err)
    assert.NotEmpty(t, result.TxHash)
    assert.Equal(t, blockchain.TxStatusSuccess, result.Status)
}
```

**End-to-End Command Flows** (`test/e2e/bot_test.go`):

```go
func TestWalletConnectionFlow(t *testing.T) {
    // Start bot in test mode
    bot := startTestBot(t)
    defer bot.Stop()

    // Simulate user sending /wallet
    update := createTestUpdate(userID, "/wallet")
    bot.ProcessUpdate(update)

    // Verify response contains TonConnect link
    response := bot.LastResponse(userID)
    assert.Contains(t, response.Text, "tonkeeper-tc://")
    assert.NotEmpty(t, response.InlineKeyboard)

    // Simulate wallet connection callback
    callback := createTestCallback(userID, "wallet_connected", "UQabc123...")
    bot.ProcessCallback(callback)

    // Verify wallet stored
    session, err := bot.GetWalletSession(userID, "ton")
    assert.NoError(t, err)
    assert.Equal(t, "UQabc123...", session.Address)
}
```

### Test Coverage Targets

**Per Requirements**: 70% code coverage

**Breakdown**:
- **Critical Path**: 90% coverage (arbitrage, trading, wallet encryption)
- **Command Handlers**: 80% coverage (all user-facing commands)
- **Utilities**: 70% coverage (logging, metrics, config)
- **Generated Code**: Excluded (gRPC stubs, database models)

**CI/CD Integration**:
- Run tests on every PR
- Block merge if coverage drops below 70%
- Generate coverage report and upload to Codecov

## Configuration Management

### Environment Variables

**Bot Service** (`.env`):

```bash
# Required
BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
ENCRYPTION_KEY=base64-encoded-32-byte-key
DATABASE_URL=postgres://user:pass@localhost:5432/telegram_trader

# Optional (with defaults)
BOT_ADMIN_USER_IDS=12345,67890
TON_NETWORK=mainnet
STONFI_API_URL=https://api.ston.fi
GALACHAIN_SERVICE_URL=localhost:50051
LOG_LEVEL=info
PROMETHEUS_PORT=9090
```

**GalaChain Service** (`.env`):

```bash
# Required
ENCRYPTION_KEY=base64-encoded-32-byte-key
DATABASE_URL=postgres://user:pass@localhost:5432/telegram_trader
GSWAP_API_URL=https://api-galaswap.gala.com

# Optional (with defaults)
GALACHAIN_NETWORK=mainnet
GRPC_PORT=50051
LOG_LEVEL=info
PROMETHEUS_PORT=9091
```

**Secrets Handling**:
- **Development**: Use `.env` file (gitignored)
- **Production**: Use HashiCorp Vault or Kubernetes Secrets
- **Never commit**: `.env.example` has placeholder values only

### Config Files

**Bot Service Config** (`configs/bot-service.yaml`):

```yaml
service:
  name: telegram-bot
  version: 1.0.0

bot:
  polling_timeout_seconds: 30
  max_concurrent_updates: 100

rate_limiting:
  commands_per_minute: 10
  burst_size: 3

arbitrage:
  monitor_interval_seconds: 3
  min_balance_ton: 1.0
  min_balance_gala: 10.0

wallet:
  session_expiry_hours: 24

trading:
  default_slippage_bps: 500
  max_slippage_bps: 1000

# Override in production.yaml
monitoring:
  health_check_interval_seconds: 30
```

**Development vs Production**:
- `development.yaml`: Testnet endpoints, verbose logging, relaxed limits
- `production.yaml`: Mainnet endpoints, structured logging, strict limits

**Loading Priority**:
1. Default values (hardcoded)
2. YAML config file
3. Environment variables (override)
4. Command-line flags (override all)

## Monitoring & Observability

### Logging

**Structured Logging Format** (JSON):

```json
{
  "timestamp": "2026-01-16T14:30:00.123Z",
  "level": "info",
  "service": "telegram-bot",
  "correlation_id": "abc-123-def-456",
  "user_id": 12345,
  "event": "swap_executed",
  "details": {
    "from_token": "TON",
    "to_token": "GALA",
    "amount": "1.0",
    "tx_hash": "abc123",
    "execution_time_ms": 3450
  }
}
```

**Log Levels**:
- **DEBUG**: Detailed flow for debugging (dev only)
- **INFO**: Normal operations, command execution
- **WARN**: Recoverable errors, degraded performance
- **ERROR**: Unrecoverable errors, failed operations
- **FATAL**: Service crash, critical failures

**Trade Logging** (JSONL - `logs/trades.jsonl`):

```jsonl
{"timestamp":"2026-01-16T14:30:00.123Z","userId":12345,"type":"swap","chain":"ton","fromToken":"TON","toToken":"GALA","amountIn":"1.0","amountOut":"850.5","fee":"0.02","txHash":"abc123","status":"success","executionTimeMs":3450}
```

**Correlation IDs**:
- Generated for each user request
- Propagated across services (Go → TypeScript)
- Included in all related log entries
- Returned to user on errors for support

### Metrics

**Prometheus Metrics to Collect**:

**Bot Service**:
```
# Command metrics
telegram_bot_commands_total{command="/swap", status="success"} 150
telegram_bot_command_duration_seconds{command="/swap"} 3.45

# Arbitrage metrics
arbitrage_opportunities_detected_total 45
arbitrage_executions_total{status="success"} 30
arbitrage_profit_usd_total 125.50

# Wallet metrics
wallet_connections_active{chain="ton"} 25
wallet_connections_active{chain="galachain"} 18

# Rate limiting
rate_limit_hits_total{user_id="12345"} 5

# Error metrics
errors_total{category="swap", code="SWAP_003"} 3
```

**GalaChain Service**:
```
# gRPC metrics
grpc_server_requests_total{method="GetPrice", status="success"} 1000
grpc_server_duration_seconds{method="ExecuteSwap"} 2.1

# gswap metrics
gswap_rate_limit_hits_total 12
gswap_api_requests_total{endpoint="/swap", status="success"} 85

# Cache metrics
price_cache_hits_total 450
price_cache_misses_total 50
```

**Grafana Dashboards**:

1. **Bot Overview Dashboard**:
   - Active users (gauge)
   - Commands per minute (graph)
   - Error rate (graph)
   - Active wallet connections (gauge)

2. **Arbitrage Performance Dashboard**:
   - Opportunities detected (graph)
   - Executions success rate (gauge)
   - Total profit (graph)
   - Average execution time (graph)

3. **System Health Dashboard**:
   - Service uptime (gauge)
   - gRPC latency (graph)
   - Database query time (graph)
   - API response times (graph)

### Health Checks

**Bot Service** (`/health`):

```json
{
  "status": "healthy",
  "uptime_seconds": 86400,
  "dependencies": {
    "database": "healthy",
    "galachain_service": "healthy",
    "ton_rpc": "healthy",
    "stonfi_api": "healthy",
    "telegram_api": "healthy"
  },
  "metrics": {
    "active_users": 25,
    "wallet_connections": 43,
    "pending_trades": 2
  }
}
```

**GalaChain Service** (gRPC `HealthCheck`):

```json
{
  "status": "healthy",
  "uptime_seconds": 86400,
  "dependencies": {
    "database": "healthy",
    "galachain_network": "healthy",
    "gswap_api": "healthy"
  }
}
```

**Health Check Logic**:
- Check database connection (simple query)
- Check gRPC service (ping)
- Check external APIs (with timeout)
- Return "degraded" if non-critical dependency down
- Return "unhealthy" if critical dependency down

## Deployment Architecture

### Development (Docker Compose)

**`docker-compose.yml`**:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: telegram_trader
      POSTGRES_PASSWORD: dev_password
      POSTGRES_DB: telegram_trader
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U telegram_trader"]
      interval: 10s
      timeout: 5s
      retries: 5

  galachain-service:
    build:
      context: .
      dockerfile: galachain-service/Dockerfile
    environment:
      DATABASE_URL: postgres://telegram_trader:dev_password@postgres:5432/telegram_trader
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      GSWAP_API_URL: https://api-galaswap.gala.com
      GRPC_PORT: 50051
      LOG_LEVEL: debug
    ports:
      - "50051:50051"
      - "9091:9091"  # Prometheus metrics
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./logs:/app/logs

  telegram-bot:
    build:
      context: .
      dockerfile: cmd/telegram-bot/Dockerfile
    environment:
      BOT_TOKEN: ${BOT_TOKEN}
      DATABASE_URL: postgres://telegram_trader:dev_password@postgres:5432/telegram_trader
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      GALACHAIN_SERVICE_URL: galachain-service:50051
      LOG_LEVEL: debug
    ports:
      - "9090:9090"  # Prometheus metrics
    depends_on:
      - postgres
      - galachain-service
    volumes:
      - ./logs:/app/logs

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    ports:
      - "9000:9090"
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'

  grafana:
    image: grafana/grafana:latest
    volumes:
      - ./monitoring/grafana:/etc/grafana/provisioning
      - grafana_data:/var/lib/grafana
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
    depends_on:
      - prometheus

volumes:
  postgres_data:
  prometheus_data:
  grafana_data:
```

**Local Development Workflow**:
1. Clone repository
2. Copy `.env.example` to `.env` and fill in values
3. Run `make setup` (install dependencies, generate proto)
4. Run `docker-compose up`
5. Bot starts polling Telegram API
6. Access Grafana at `http://localhost:3000`

### Production (Kubernetes)

**Bot Service Deployment** (`deployments/kubernetes/bot-service.yaml`):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: telegram-bot
  labels:
    app: telegram-bot
spec:
  replicas: 2
  selector:
    matchLabels:
      app: telegram-bot
  template:
    metadata:
      labels:
        app: telegram-bot
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: telegram-bot
        image: telegram-trader/bot:latest
        ports:
        - containerPort: 9090
          name: metrics
        env:
        - name: BOT_TOKEN
          valueFrom:
            secretKeyRef:
              name: telegram-bot-secrets
              key: bot-token
        - name: ENCRYPTION_KEY
          valueFrom:
            secretKeyRef:
              name: telegram-bot-secrets
              key: encryption-key
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: telegram-bot-secrets
              key: database-url
        - name: GALACHAIN_SERVICE_URL
          value: "galachain-service:50051"
        - name: LOG_LEVEL
          value: "info"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 10
          periodSeconds: 5
        volumeMounts:
        - name: logs
          mountPath: /app/logs
      volumes:
      - name: logs
        persistentVolumeClaim:
          claimName: telegram-bot-logs
---
apiVersion: v1
kind: Service
metadata:
  name: telegram-bot-metrics
spec:
  selector:
    app: telegram-bot
  ports:
  - port: 9090
    targetPort: 9090
    name: metrics
```

**GalaChain Service Deployment** (`deployments/kubernetes/galachain-service.yaml`):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: galachain-service
spec:
  replicas: 3  # Horizontal scaling for gRPC load
  selector:
    matchLabels:
      app: galachain-service
  template:
    metadata:
      labels:
        app: galachain-service
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9091"
    spec:
      containers:
      - name: galachain-service
        image: telegram-trader/galachain:latest
        ports:
        - containerPort: 50051
          name: grpc
        - containerPort: 9091
          name: metrics
        env:
        - name: ENCRYPTION_KEY
          valueFrom:
            secretKeyRef:
              name: galachain-secrets
              key: encryption-key
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: galachain-secrets
              key: database-url
        - name: GRPC_PORT
          value: "50051"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          exec:
            command: ["/bin/grpc_health_probe", "-addr=:50051"]
          initialDelaySeconds: 30
          periodSeconds: 10
        volumeMounts:
        - name: logs
          mountPath: /app/logs
      volumes:
      - name: logs
        persistentVolumeClaim:
          claimName: galachain-logs
---
apiVersion: v1
kind: Service
metadata:
  name: galachain-service
spec:
  selector:
    app: galachain-service
  ports:
  - port: 50051
    targetPort: 50051
    name: grpc
  - port: 9091
    targetPort: 9091
    name: metrics
  type: ClusterIP  # Internal service
```

**Scaling Strategy**:
- **Bot Service**: 2 replicas (active-active for redundancy)
- **GalaChain Service**: 3 replicas (higher load from price monitoring)
- **Horizontal Pod Autoscaler**: Scale based on CPU (>70%) or gRPC request rate
- **Database**: Managed PostgreSQL (AWS RDS, GCP Cloud SQL) with read replicas

## Migration & Rollout Plan

### Phase 1: MVP (P0 Features)

**Timeline**: Weeks 1-6

**Files to Create**:
```
cmd/telegram-bot/main.go
internal/bot/bot.go
internal/bot/handlers/start.go
internal/bot/handlers/wallet.go
internal/bot/handlers/balance.go
internal/bot/handlers/swap.go
internal/wallet/manager.go
internal/wallet/tonconnect.go
internal/wallet/encryption.go
internal/blockchain/ton/client.go
internal/dex/stonfi/client.go
galachain-service/src/index.ts
galachain-service/src/server/grpc.ts
galachain-service/src/gswap/client.ts
galachain-service/src/wallet/manager.ts
proto/galachain.proto
```

**Features to Implement**:
1. **Week 1-2**: Project setup, gRPC communication
   - Initialize Go and TypeScript projects
   - Set up gRPC protocol definitions
   - Implement basic gRPC server/client
   - Set up Docker Compose for local dev

2. **Week 3**: TonKeeper wallet connection
   - Implement TonConnect protocol
   - Generate QR codes and deep links
   - Store encrypted wallet sessions
   - Test on iPhone with TonKeeper app

3. **Week 4**: Gala wallet connection
   - Implement WalletConnect v2
   - Add manual private key input fallback
   - Encrypt and store credentials
   - Test wallet signing

4. **Week 5**: Balance checking and manual swaps
   - Fetch balances from TON and GalaChain
   - Implement ston.fi swap simulation
   - Implement gswap swap execution
   - Add confirmation flows

5. **Week 6**: Testing and polish
   - Integration tests with testnets
   - Error handling and user messages
   - Help documentation
   - Security audit of key handling

**Testing Approach**:
- Unit tests for all core logic (70% coverage)
- Integration tests with testnet (TON testnet, GalaChain testnet)
- Manual testing with real wallets (small amounts)
- Security review of encryption implementation

**Success Criteria**:
- [ ] User can connect TonKeeper wallet via QR/deep link
- [ ] User can connect Gala wallet (WalletConnect or manual)
- [ ] User can check balances on both chains
- [ ] User can execute manual swap on ston.fi
- [ ] User can execute manual swap on gswap
- [ ] All private keys encrypted and stored securely
- [ ] 70% test coverage achieved

### Phase 2: Enhancements (P1 Features)

**Timeline**: Weeks 7-10

**New Files/Modules**:
```
internal/bot/handlers/price.go
internal/bot/handlers/alert.go
internal/bot/handlers/portfolio.go
internal/bot/handlers/orders.go
internal/arbitrage/engine.go
internal/arbitrage/monitor.go
internal/arbitrage/calculator.go
galachain-service/src/gswap/price.ts
```

**Features to Add**:
1. **Week 7**: Price monitoring
   - Implement price fetching from ston.fi and gswap
   - Add price caching (5s TTL)
   - Display spread between exchanges
   - Add USD value integration (CoinGecko)

2. **Week 8**: Price alerts
   - Implement background price monitoring
   - Store user alert preferences
   - Send Telegram notifications on threshold
   - Add alert management commands

3. **Week 9**: Trade history and portfolio
   - Implement JSONL trade logging
   - Store trades in database
   - Calculate portfolio P&L
   - Add CSV export for trade history

4. **Week 10**: Manual arbitrage
   - Implement spread detection logic
   - Calculate position sizing with safety checks
   - Implement dual-leg execution
   - Add arbitrage preview and confirmation

**Testing**:
- Price monitoring accuracy tests
- Alert triggering tests
- Arbitrage calculation tests (position sizing, profit estimation)
- End-to-end arbitrage flow on testnet

**Success Criteria**:
- [ ] Price checking shows real-time data from both chains
- [ ] Price alerts trigger notifications correctly
- [ ] Trade history persists and displays accurately
- [ ] Portfolio shows correct P&L calculations
- [ ] Manual arbitrage executes both legs successfully

### Phase 3: Advanced (P2 Features)

**Timeline**: Weeks 11-12

**New Files/Modules**:
```
internal/arbitrage/automation.go
internal/arbitrage/circuit_breaker.go
internal/bot/handlers/arbitrage_auto.go
```

**Features to Add**:
1. **Week 11**: Arbitrage automation
   - Implement continuous price monitoring
   - Auto-execute when spread exceeds threshold
   - Add circuit breaker (3 failures → pause)
   - Implement daily loss limits

2. **Week 12**: Production hardening
   - Comprehensive error handling
   - Performance optimization
   - Security hardening (HSM integration)
   - Production deployment automation

**Testing**:
- Automated arbitrage simulation
- Circuit breaker tests
- Load testing (concurrent users)
- Security penetration testing

**Success Criteria**:
- [ ] Automated arbitrage executes reliably
- [ ] Circuit breaker stops automation on failures
- [ ] Daily loss limits enforced
- [ ] Production deployment successful
- [ ] Zero critical security vulnerabilities

## Open Questions

1. **Database Choice**: Should we use PostgreSQL from day 1, or start with SQLite and migrate later?
   - **Recommendation**: Start with SQLite for simplicity, plan PostgreSQL migration for production

2. **Arbitrage Threshold**: Is 0.3% spread too conservative? Should it be user-configurable from start?
   - **Recommendation**: Default to 0.3%, make configurable in Phase 2

3. **Multi-User vs Single-User**: Should MVP support multiple users or focus on single user first?
   - **Recommendation**: Build multi-user from start (minimal extra work, better testing)

4. **Testnet vs Mainnet**: Should we deploy to mainnet immediately or test on testnet first?
   - **Recommendation**: Phase 1 on testnet, Phase 2 on mainnet with small amounts

5. **Wallet Connection Fallback**: If TonConnect fails, should we support manual private key input?
   - **Recommendation**: TonConnect only for TON (standard), manual fallback for Gala (less mature)

6. **Rate Limiting Scope**: Should rate limiting be per-user or global?
   - **Recommendation**: Per-user (10/min) + global (100/min) to prevent single user abuse

7. **Price Data Source**: Should we use CoinGecko or integrate directly with exchanges?
   - **Recommendation**: CoinGecko for USD prices (simpler), direct API for trading pairs

8. **HSM Integration**: Should we integrate HSM in Phase 1 or Phase 3?
   - **Recommendation**: Phase 1 uses AES-256, Phase 3 adds HSM for production hardening

## Appendix

### References

**Technology Documentation**:
- Go Telegram Bot: https://github.com/go-telegram/bot
- tonutils-go: https://github.com/xssnick/tonutils-go
- gswap SDK: https://www.npmjs.com/package/@gala-chain/gswap-sdk
- TonConnect: https://docs.ton.org/develop/dapps/ton-connect/integration
- WalletConnect: https://docs.walletconnect.network/
- gRPC: https://grpc.io/docs/languages/go/ and https://grpc.io/docs/languages/node/
- ston.fi API: https://docs.ston.fi/developer-section/dex/api/reference
- gswap API: https://galaswap.gala.com/info/api.html

**Reference Implementation**:
- gswap-bot: https://github.com/cchuter/gswap-bot

**Security Best Practices**:
- OWASP Cryptographic Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html
- HashiCorp Vault: https://www.vaultproject.io/docs
- Azure Key Vault: https://learn.microsoft.com/en-us/azure/key-vault/

### Glossary

**AES-256-GCM**: Advanced Encryption Standard with 256-bit key in Galois/Counter Mode (authenticated encryption)

**Arbitrage**: Trading strategy exploiting price differences for the same asset across exchanges

**Basis Points (bps)**: Unit for percentages where 100 bps = 1%

**Circuit Breaker**: Safety mechanism that stops operations after repeated failures

**Correlation ID**: Unique identifier for tracking a request across multiple services

**GALA**: Native token of GalaChain blockchain

**GTON**: Wrapped TON token on GalaChain for cross-chain trading

**gRPC**: High-performance RPC framework using Protocol Buffers

**gswap**: Decentralized exchange on GalaChain

**JSONL**: JSON Lines format (one JSON object per line)

**P0/P1/P2**: Priority levels (P0 = must-have MVP, P1 = important, P2 = nice-to-have)

**Slippage**: Difference between expected and actual execution price

**Spread**: Price difference between two markets, expressed as percentage

**ston.fi**: Decentralized exchange on TON blockchain

**TON**: The Open Network blockchain

**TonConnect**: Standard protocol for connecting dApps to TON wallets

**TonKeeper**: Popular mobile wallet for TON blockchain

**WalletConnect**: Open protocol for connecting wallets to dApps
