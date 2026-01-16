---
spec: telegram-trader
phase: tasks
total_tasks: 72
created: 2026-01-16T15:30:00Z
---

# Tasks: Telegram Trader

## Execution Context

**Testing Approach:** Standard - unit + integration tests (POC first, tests in Phase 3)
**Deployment Focus:** Local development (docker-compose priority, defer Kubernetes)

**Phasing Strategy:**
1. Phase 1: Make It Work (POC) - Validate core functionality
2. Phase 2: Refactoring - Clean up and improve
3. Phase 3: Testing - Add comprehensive tests
4. Phase 4: Quality Gates - Linting, CI/CD

## Task Summary

**Total Tasks:** 72

**By Phase:**
- Phase 1 (POC): 28 tasks
- Phase 2 (Refactoring): 20 tasks
- Phase 3 (Testing): 16 tasks
- Phase 4 (Quality): 8 tasks

**By Priority:**
- P0 (MVP): 35 tasks
- P1 (Enhancement): 25 tasks
- P2 (Advanced): 12 tasks

## Phase 1: Make It Work (POC)

### Task 1: Initialize Go module and project structure

**Do:**
- [x] Run `go mod init github.com/cchuter/telegram-trader`
- [x] Create directory structure: `cmd/telegram-bot`, `internal/{bot,blockchain,dex,wallet,arbitrage,galachain,storage,logging,metrics,config}`
- [x] Create `cmd/telegram-bot/main.go` with basic Go structure (package main, empty main function)
- [x] Add `.gitignore` for Go (standard template + `/logs/`, `*.db`, `.env`)

**Files:**
- `go.mod` - create Go module
- `cmd/telegram-bot/main.go` - create entry point
- `internal/*/` - create package directories
- `.gitignore` - create Git ignore file

**Done when:**
- Go module initialized with correct import path
- Directory structure matches design.md file structure
- main.go compiles without errors (`go build ./cmd/telegram-bot`)

**Verify:**
```bash
go mod verify && go build ./cmd/telegram-bot
```

**Commit:**
```
chore(setup): initialize Go project structure

- Initialize Go module github.com/cchuter/telegram-trader
- Create directory structure for telegram-bot service
- Add basic main.go entry point
- Configure .gitignore for Go project

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-001 (Bot initialization), NFR-038 (Code structure)
**Design:** File Structure section

---

### Task 2: Set up TypeScript GalaChain service project

**Do:**
- [x] Create `galachain-service/` directory
- [x] Run `npm init -y` in galachain-service directory
- [x] Install dependencies: `@grpc/grpc-js @grpc/proto-loader @gala-chain/gswap-sdk typescript @types/node ts-node`
- [x] Create `tsconfig.json` with strict mode enabled
- [x] Create `galachain-service/src/index.ts` with basic structure
- [x] Add `.gitignore` for Node.js (standard template + `dist/`, `*.log`, `.env`)

**Files:**
- `galachain-service/package.json` - create with dependencies
- `galachain-service/tsconfig.json` - create TypeScript config
- `galachain-service/src/index.ts` - create entry point
- `galachain-service/.gitignore` - create Git ignore file

**Done when:**
- npm install completes successfully
- TypeScript compiles without errors (`npm run build`)
- index.ts contains basic console.log "GalaChain service starting"

**Verify:**
```bash
cd galachain-service && npm install && npm run build
```

**Commit:**
```
chore(setup): initialize TypeScript GalaChain service

- Initialize Node.js project with npm
- Add gRPC and GalaChain SDK dependencies
- Configure TypeScript with strict mode
- Add basic service entry point

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-004 (GalaChain operations), NFR-038 (Code structure)
**Design:** File Structure section, TypeScript GalaChain Service component

---

### Task 3: Create gRPC protocol buffer definitions

**Do:**
- [x] Create `proto/` directory at project root
- [x] Create `proto/galachain.proto` with service definitions from design.md
- [x] Add gRPC messages: GetPriceRequest, PriceResponse, BalanceRequest, BalanceResponse, SwapRequest, SwapResponse, WalletSessionRequest, WalletSessionResponse, HealthCheckRequest, HealthCheckResponse
- [x] Create `proto/README.md` documenting proto file and generation commands

**Files:**
- `proto/galachain.proto` - create gRPC service definition
- `proto/README.md` - create documentation

**Done when:**
- Proto file contains all 8 RPC methods from design
- Proto file is valid syntax (will verify in next task when generating code)
- README documents how to generate Go and TypeScript code

**Verify:**
```bash
cat proto/galachain.proto | grep "service GalaChainService"
```

**Commit:**
```
feat(grpc): define gRPC protocol buffer interface

- Add GalaChainService with 6 RPC methods
- Define message types for price, balance, swap, wallet session
- Add health check RPC
- Document code generation process

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-043 (gRPC communication), NFR-038 (Documentation)
**Design:** Interface Definitions - gRPC Protocol Buffers section

---

### Task 4: Generate Go gRPC code and set up client skeleton

**Do:**
- [x] Install protoc-gen-go and protoc-gen-go-grpc: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`
- [x] Generate Go code: `protoc --go_out=. --go-grpc_out=. proto/galachain.proto`
- [x] Create `internal/galachain/client.go` with Client struct wrapping gRPC connection
- [x] Add basic Connect/Close methods (no retry logic yet)

**Files:**
- `internal/galachain/pb/galachain.pb.go` - generate gRPC message types
- `internal/galachain/pb/galachain_grpc.pb.go` - generate gRPC client/server code
- `internal/galachain/client.go` - create gRPC client wrapper

**Done when:**
- Proto generation succeeds without errors
- Generated files are in `internal/galachain/pb/`
- client.go compiles and contains Client struct with grpc.ClientConn field

**Verify:**
```bash
go build ./internal/galachain
```

**Commit:**
```
feat(grpc): generate Go gRPC client code

- Generate protobuf message and gRPC code from proto
- Create gRPC client wrapper in internal/galachain
- Add Connect and Close methods for connection management

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-043 (gRPC communication)
**Design:** Component Design - gRPC Client for GalaChain

---

### Task 5: Generate TypeScript gRPC code and set up server skeleton

**Do:**
- [x] Add proto-loader script to package.json
- [x] Generate TypeScript types: `npx grpc_tools_node_protoc --js_out=import_style=commonjs,binary:./src/types --grpc_out=grpc_js:./src/types --ts_out=service=grpc-node,mode=grpc-js:./src/types -I ../proto ../proto/galachain.proto`
- [x] Create `galachain-service/src/server/grpc.ts` with gRPC server setup
- [x] Add basic server start/stop methods (empty handler implementations)

**Files:**
- `galachain-service/src/types/galachain_pb.js` - generate message types
- `galachain-service/src/types/galachain_grpc_pb.js` - generate service code
- `galachain-service/src/types/galachain_pb.d.ts` - generate TypeScript definitions
- `galachain-service/src/server/grpc.ts` - create gRPC server

**Done when:**
- Proto generation succeeds without errors
- Generated files are in `src/types/`
- grpc.ts compiles and contains server initialization code

**Verify:**
```bash
cd galachain-service && npm run build
```

**Commit:**
```
feat(grpc): generate TypeScript gRPC server code

- Generate protobuf and gRPC code for TypeScript
- Create gRPC server wrapper in src/server
- Add server start/stop methods with basic health check

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-043 (gRPC communication)
**Design:** Component Design - TypeScript GalaChain Service

---

### Task 6: Set up basic configuration loading for Go service

**Do:**
- [x] Create `internal/config/types.go` with Config struct (BotToken, DatabaseURL, EncryptionKey, GalaChainServiceURL)
- [x] Create `internal/config/config.go` with Load() function that reads from environment variables
- [x] Use `os.Getenv()` for now (no validation, no defaults)
- [x] Create `.env.example` with placeholder values

**Files:**
- `internal/config/types.go` - create config structures
- `internal/config/config.go` - create config loader
- `.env.example` - create example environment file

**Done when:**
- Config struct contains all required fields from design.md
- Load() function reads BOT_TOKEN, DATABASE_URL, ENCRYPTION_KEY, GALACHAIN_SERVICE_URL
- .env.example documents all required variables

**Verify:**
```bash
go build ./internal/config
```

**Commit:**
```
feat(config): add basic configuration loading

- Create config types for bot service
- Load configuration from environment variables
- Add .env.example with placeholder values

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-042 (Externalized configuration)
**Design:** Configuration Management section

---

### Task 7: Set up basic configuration loading for TypeScript service

**Do:**
- [x] Install dotenv: `npm install dotenv @types/dotenv`
- [x] Create `galachain-service/src/config/index.ts` with Config interface
- [x] Add loadConfig() function using process.env
- [x] Create `galachain-service/.env.example` with placeholder values

**Files:**
- `galachain-service/src/config/index.ts` - create config loader
- `galachain-service/.env.example` - create example environment file

**Done when:**
- Config interface contains GRPC_PORT, DATABASE_URL, ENCRYPTION_KEY, GSWAP_API_URL
- loadConfig() reads from process.env using dotenv
- .env.example documents all required variables

**Verify:**
```bash
cd galachain-service && npm run build
```

**Commit:**
```
feat(config): add configuration loading for GalaChain service

- Create config interface for TypeScript service
- Load configuration from environment variables
- Add .env.example with placeholder values

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-042 (Externalized configuration)
**Design:** Configuration Management section

---

### Task 8: Implement basic Telegram bot with command routing

**Do:**
- [x] Add dependency: `go get github.com/go-telegram/bot`
- [x] Create `internal/bot/bot.go` with Bot struct
- [x] Implement Start() method with bot initialization
- [x] Register /start and /help command handlers (hardcoded response strings for now)
- [x] Update `cmd/telegram-bot/main.go` to create bot and call Start()

**Files:**
- `internal/bot/bot.go` - create bot instance
- `internal/bot/handlers/start.go` - create /start handler
- `internal/bot/handlers/help.go` - create /help handler
- `cmd/telegram-bot/main.go` - update to initialize and start bot

**Done when:**
- Bot connects to Telegram API successfully
- /start command responds with "Welcome to Telegram Trader!"
- /help command responds with basic command list
- Bot runs without crashing (manual test with real bot token)

**Verify:**
```bash
BOT_TOKEN=test_token go run ./cmd/telegram-bot
```

**Commit:**
```
feat(bot): implement basic Telegram bot with command routing

- Initialize Telegram bot using go-telegram/bot library
- Add /start and /help command handlers
- Set up command routing in main.go
- Add basic welcome and help messages

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-001 (Bot initialization), US-012 (Help command)
**Design:** Component Design - Bot Command Handler

---

### Task 9: Set up SQLite database for development

**Do:**
- [x] Add dependency: `go get github.com/mattn/go-sqlite3`
- [x] Create `internal/storage/db.go` with Database interface (GetUserSession, SaveUserSession, GetWalletSession, SaveWalletSession methods)
- [x] Create `internal/storage/sqlite.go` implementing Database interface
- [x] Create `migrations/001_initial_schema.up.sql` with user_sessions and wallet_sessions tables
- [x] Add InitDB() function that creates database file and runs migrations

**Files:**
- `internal/storage/db.go` - create database interface
- `internal/storage/sqlite.go` - create SQLite implementation
- `internal/storage/models.go` - create data models
- `migrations/001_initial_schema.up.sql` - create initial schema

**Done when:**
- Database interface has methods for user and wallet sessions
- SQLite implementation creates database file
- Migrations create user_sessions and wallet_sessions tables
- InitDB() runs successfully and creates telegram-trader.db

**Verify:**
```bash
go test ./internal/storage -v
```

**Commit:**
```
feat(storage): implement SQLite database for local development

- Define database interface for user and wallet sessions
- Implement SQLite adapter
- Create initial database schema with migrations
- Add user_sessions and wallet_sessions tables

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-048 (Session persistence), NFR-009 (Storage)
**Design:** Data Models section

---

### Task 10: Implement basic authentication middleware

**Do:**
- [x] Create `internal/bot/middleware/auth.go` with AuthMiddleware struct
- [x] Implement Authenticate() method that checks user_id against hardcoded whitelist (BOT_ADMIN_USER_IDS env var)
- [x] Create user session on first /start command
- [x] Return error for non-whitelisted users with "Access denied" message
- [x] Wire middleware into bot command handlers

**Files:**
- `internal/bot/middleware/auth.go` - create authentication middleware
- `internal/bot/bot.go` - update to use auth middleware

**Done when:**
- Whitelisted users can execute commands
- Non-whitelisted users receive "Access denied" message
- User session created in database on /start
- Middleware runs before all command handlers

**Verify:**
```bash
go test ./internal/bot/middleware -v
```

**Commit:**
```
feat(auth): implement basic authentication middleware

- Add user authentication with whitelist check
- Create user sessions on first access
- Block non-whitelisted users with error message
- Wire middleware into command routing

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-038 (User whitelist), US-013 (User authentication)
**Design:** Component Design - Auth Middleware

---

### Task 11: Implement TON blockchain client skeleton

**Do:**
- [x] Add dependency: `go get github.com/xssnick/tonutils-go`
- [x] Create `internal/blockchain/ton/client.go` with Client struct
- [x] Implement Connect() method to initialize TON lite client
- [x] Implement GetBalance() method (returns hardcoded "10.0" for POC)
- [x] Create `internal/blockchain/interfaces.go` with Client interface

**Files:**
- `internal/blockchain/interfaces.go` - create blockchain client interface
- `internal/blockchain/ton/client.go` - create TON client
- `internal/blockchain/ton/wallet.go` - create wallet operations

**Done when:**
- TON client connects to mainnet or testnet successfully
- GetBalance() returns string balance (hardcoded for POC)
- Client implements blockchain.Client interface
- Code compiles without errors

**Verify:**
```bash
go build ./internal/blockchain/ton
```

**Commit:**
```
feat(blockchain): implement TON blockchain client skeleton

- Add tonutils-go client initialization
- Create blockchain client interface
- Implement GetBalance method (hardcoded for POC)
- Set up connection to TON network

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-003 (TON blockchain operations)
**Design:** Component Design - TON Blockchain Client

---

### Task 12: Implement ston.fi DEX client skeleton

**Do:**
- [x] Create `internal/dex/stonfi/client.go` with Client struct
- [x] Add HTTP client using net/http
- [x] Implement SimulateSwap() method calling POST /v1/swap/simulate endpoint
- [x] Return simulated output amount, fee, slippage from API response
- [x] Create `internal/dex/interfaces.go` with DEX Client interface

**Files:**
- `internal/dex/interfaces.go` - create DEX client interface
- `internal/dex/stonfi/client.go` - create ston.fi client
- `internal/dex/stonfi/swap.go` - create swap operations

**Done when:**
- HTTP client configured with base URL https://api.ston.fi
- SimulateSwap() makes real API call and parses response
- Client implements dex.Client interface
- Can simulate 1 TON -> GALA swap successfully

**Verify:**
```bash
go test ./internal/dex/stonfi -v
```

**Commit:**
```
feat(dex): implement ston.fi DEX client

- Add HTTP client for ston.fi API
- Implement swap simulation endpoint
- Parse simulation response for amount, fee, slippage
- Create DEX client interface

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-005 (Swap simulation), FR-019 (Price checking)
**Design:** Component Design - ston.fi DEX Client

---

### Task 13: Implement /balance command handler

**Do:**
- [x] Create `internal/bot/handlers/balance.go`
- [x] Call TON client GetBalance() for TON and GALA tokens
- [x] Call GalaChain gRPC service GetBalance() for GALA and GTON tokens (stub for now)
- [x] Format response showing both chain balances
- [x] Register handler in bot.go

**Files:**
- `internal/bot/handlers/balance.go` - create balance command handler

**Done when:**
- /balance command returns TON balance (hardcoded "10.0 TON")
- /balance command shows "GALA balance: N/A (GalaChain service not connected)" for POC
- Response formatted as "TON: 10.0\nGALA: N/A"
- Command works without crashing

**Verify:**
```bash
go run ./cmd/telegram-bot
```

**Commit:**
```
feat(bot): implement /balance command handler

- Add balance checking for TON chain
- Format balance response for user
- Register balance command handler
- Placeholder for GalaChain balances

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-017 (Balance checking), US-003 (Balance checking)
**Design:** Data Flow Diagrams - Balance checking

---

### Task 14: Implement basic wallet connection UI for TonConnect

**Do:**
- [x] Create `internal/wallet/manager.go` with Manager struct
- [x] Implement ConnectTonWallet() that generates placeholder TonConnect URL
- [x] Create /wallet command handler that sends TonConnect deep link
- [x] For POC: Accept any wallet address via /wallet <address> command instead of real TonConnect
- [x] Save wallet address to database in wallet_sessions table

**Files:**
- `internal/wallet/manager.go` - create wallet manager
- `internal/bot/handlers/wallet.go` - create wallet command handler

**Done when:**
- /wallet command sends message "Send wallet address using: /wallet <address>"
- /wallet UQabc123... saves address to database
- wallet_sessions table populated with user_id, chain='ton', wallet_address
- User sees "TON wallet connected: UQ...abc" confirmation

**Verify:**
```bash
go run ./cmd/telegram-bot
```

**Commit:**
```
feat(wallet): implement basic TON wallet connection

- Add wallet manager for connection handling
- Create /wallet command handler
- Store wallet address in database
- Add manual wallet address input for POC

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-011 (Wallet validation), US-001 (Wallet connection)
**Design:** Data Flow Diagrams - Wallet Connection Flow

---

### Task 15: Implement basic swap simulation on ston.fi

**Do:**
- [x] Create `internal/bot/handlers/swap.go`
- [x] Parse /swap command: `/swap <amount> <from_token> <to_token> stonfi`
- [x] Call ston.fi SimulateSwap() to get expected output
- [x] Display swap preview with inline keyboard [Confirm] [Cancel] buttons
- [x] For POC: Don't execute real swap, just show simulation

**Files:**
- `internal/bot/handlers/swap.go` - create swap command handler

**Done when:**
- /swap 1 TON GALA stonfi shows swap preview
- Preview displays: "Swap 1 TON → ~850 GALA (estimated), Fee: 0.02 TON"
- Inline keyboard with Confirm and Cancel buttons displayed
- No actual swap execution yet (POC phase)

**Verify:**
```bash
go run ./cmd/telegram-bot
```

**Commit:**
```
feat(swap): implement swap simulation on ston.fi

- Add /swap command parser
- Call ston.fi swap simulation endpoint
- Display swap preview with confirmation buttons
- No execution yet (POC phase)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-005 (Swap simulation), US-004 (Manual swap)
**Design:** Data Flow Diagrams - Manual Swap Execution Flow

---

### Task 16: Implement GalaChain gRPC GetBalance handler

**Do:**
- [x] Create `galachain-service/src/server/handlers.ts` with GrpcHandlers class
- [x] Implement getBalance() method that returns hardcoded balance for POC
- [x] Wire handler into gRPC server in grpc.ts
- [x] Return mock data: GALA: "5000.0", GTON: "2.5"

**Files:**
- `galachain-service/src/server/handlers.ts` - create gRPC handlers
- `galachain-service/src/server/grpc.ts` - update to wire handlers

**Done when:**
- gRPC server starts on port 50051
- GetBalance RPC returns BalanceResponse with hardcoded values
- TypeScript compiles without errors
- Server runs without crashing

**Verify:**
```bash
cd galachain-service && npm run dev
```

**Commit:**
```
feat(grpc): implement GetBalance gRPC handler

- Add gRPC handlers class
- Implement getBalance method with mock data
- Wire handler into gRPC server
- Return hardcoded GALA and GTON balances

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-017 (Balance checking)
**Design:** Component Design - gRPC Server

---

### Task 17: Connect Go bot to GalaChain gRPC service

**Do:**
- [x] Update `internal/galachain/client.go` to implement GetBalance() method
- [x] Call pb.GalaChainServiceClient.GetBalance() over gRPC
- [x] Update /balance command handler to call GalaChain client
- [x] Display GALA and GTON balances from GalaChain service

**Files:**
- `internal/galachain/client.go` - update with GetBalance method
- `internal/bot/handlers/balance.go` - update to call GalaChain service

**Done when:**
- Bot successfully connects to GalaChain service on localhost:50051
- /balance command shows TON: 10.0, GALA: 5000.0, GTON: 2.5
- gRPC call succeeds without errors
- Both services run together successfully

**Verify:**
```bash
# Terminal 1: Start GalaChain service
cd galachain-service && npm run dev
# Terminal 2: Start bot
GALACHAIN_SERVICE_URL=localhost:50051 go run ./cmd/telegram-bot
```

**Commit:**
```
feat(grpc): connect bot to GalaChain service for balance checking

- Implement GetBalance gRPC call in Go client
- Update /balance handler to fetch GalaChain balances
- Display balances from both TON and GalaChain
- Verify end-to-end gRPC communication

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-043 (gRPC communication), FR-017 (Balance checking)
**Design:** Data Flow Diagrams - Cross-chain balance flow

---

### Task 18: Implement GswapClient skeleton in TypeScript

**Do:**
- [x] Install gswap SDK: `npm install @gala-chain/gswap-sdk`
- [x] Create `galachain-service/src/gswap/client.ts` with GSwapClient class
- [x] Add getPrice() method (returns hardcoded "855" for GTON/GALA for POC)
- [x] Add rate limiter placeholder (no actual rate limiting yet)

**Files:**
- `galachain-service/src/gswap/client.ts` - create GSwap client wrapper

**Done when:**
- GSwap SDK installed successfully
- GSwapClient class created with getPrice() method
- Returns hardcoded price for POC
- TypeScript compiles without errors

**Verify:**
```bash
cd galachain-service && npm run build
```

**Commit:**
```
feat(gswap): implement GSwap client skeleton

- Add @gala-chain/gswap-sdk dependency
- Create GSwapClient wrapper class
- Implement getPrice method (hardcoded for POC)
- Add rate limiter placeholder

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-020 (GalaChain price checking)
**Design:** Component Design - GSwap Client

---

### Task 19: Implement GetPrice gRPC handler

**Do:**
- [x] Add getPrice() method to GrpcHandlers class
- [x] Call GSwapClient.getPrice() internally
- [x] Return PriceResponse with price, timestamp
- [x] Wire into gRPC server

**Files:**
- `galachain-service/src/server/handlers.ts` - update with getPrice handler

**Done when:**
- GetPrice RPC returns PriceResponse with mock price "855"
- Response includes timestamp
- gRPC call succeeds from Go client
- No errors in TypeScript compilation

**Verify:**
```bash
cd galachain-service && npm run dev
```

**Commit:**
```
feat(grpc): implement GetPrice gRPC handler

- Add getPrice handler method
- Call GSwapClient for price data
- Return formatted PriceResponse
- Wire into gRPC service

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-020 (Price checking)
**Design:** Data Flow Diagrams - Price Checking Flow

---

### Task 20: Implement /price command handler

**Do:**
- [x] Create `internal/bot/handlers/price.go`
- [x] Fetch TON/GALA price from ston.fi (simulate 1 TON -> GALA)
- [x] Fetch GTON/GALA price from GalaChain service via gRPC
- [x] Calculate spread percentage: (price_gswap - price_stonfi) / price_stonfi * 100
- [x] Format response showing both prices and spread

**Files:**
- `internal/bot/handlers/price.go` - create price command handler

**Done when:**
- /price shows "TON/GALA on ston.fi: 850 GALA"
- /price shows "GTON/GALA on gswap: 855 GALA"
- /price shows "Spread: 0.59%"
- Command completes in under 5 seconds

**Verify:**
```bash
go run ./cmd/telegram-bot
```

**Commit:**
```
feat(bot): implement /price command handler

- Add price checking for both DEXs
- Calculate spread percentage
- Format response with prices and spread
- Fetch data from ston.fi and GalaChain service

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-021 (Spread calculation), US-006 (Price checking)
**Design:** Data Flow Diagrams - Price Checking Flow

---

### Task 21: Implement basic arbitrage detection logic

**Do:**
- [x] Create `internal/arbitrage/engine.go` with Engine struct
- [x] Implement DetectOpportunity() method that checks spread between prices
- [x] Return Opportunity struct if spread >= 0.3% (hardcoded threshold for POC)
- [x] Calculate buy/sell direction (buy on cheaper, sell on expensive)
- [x] Don't implement execution yet

**Files:**
- `internal/arbitrage/engine.go` - create arbitrage engine
- `internal/arbitrage/types.go` - create opportunity types

**Done when:**
- DetectOpportunity() fetches prices from both DEXs
- Returns Opportunity if spread >= 0.3%
- Opportunity contains direction (BuyTonSellGala or BuyGalaSellTon)
- Returns nil if spread < 0.3%

**Verify:**
```bash
go test ./internal/arbitrage -v
```

**Commit:**
```
feat(arbitrage): implement arbitrage opportunity detection

- Create arbitrage engine with detection logic
- Calculate spread between ston.fi and gswap
- Identify profitable direction
- Return opportunity if spread >= 0.3%

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-009 (Arbitrage detection), US-010 (Manual arbitrage)
**Design:** Component Design - Arbitrage Engine

---

### Task 22: Implement /arbitrage command handler

**Do:**
- [x] Create `internal/bot/handlers/arbitrage.go`
- [x] Call arbitrage engine DetectOpportunity()
- [x] Display opportunity details: direction, spread, estimated profit
- [x] Show [Execute] [Cancel] inline keyboard
- [x] Don't implement execution callback yet (POC phase)

**Files:**
- `internal/bot/handlers/arbitrage.go` - create arbitrage command handler

**Done when:**
- /arbitrage command shows "Checking for arbitrage opportunities..."
- If spread >= 0.3%: Shows "Opportunity found! Buy on ston.fi, sell on gswap. Spread: 0.59%"
- If spread < 0.3%: Shows "No arbitrage opportunity (spread: 0.10% < threshold: 0.30%)"
- Inline keyboard displayed for opportunities

**Verify:**
```bash
go run ./cmd/telegram-bot
```

**Commit:**
```
feat(bot): implement /arbitrage command handler

- Add arbitrage opportunity checking
- Display opportunity details to user
- Show execution confirmation UI
- Handle no-opportunity case

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** US-010 (Manual arbitrage execution)
**Design:** Data Flow Diagrams - Arbitrage Execution Flow

---

### Task 23: Implement basic error handling and user-friendly messages

**Do:**
- [x] Create `internal/bot/errors.go` with error types and user message mapping
- [x] Add error codes from design.md (AUTH_001, WALLET_001, SWAP_001, etc.)
- [x] Update command handlers to return errors instead of panicking
- [x] Format error messages for users with actionable guidance

**Files:**
- `internal/errors/errors.go` - create error types and messages (moved to avoid import cycle)

**Done when:**
- All error codes from design.md Error Handling section defined
- Each error has user-friendly message (no technical jargon)
- Command handlers return errors properly
- Bot sends formatted error messages to user

**Verify:**
```bash
go test ./internal/bot -v
```

**Commit:**
```
feat(errors): implement error handling with user-friendly messages

- Define error codes and types
- Map errors to user-friendly messages
- Update handlers to return errors gracefully
- Add actionable guidance in error messages

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-031 (Clear error messages), NFR-044 (User-friendly errors)
**Design:** Error Handling section

---

### Task 24: Implement basic logging infrastructure

**Do:**
- [x] Create `internal/logging/logger.go` with Logger struct
- [x] Use standard library log for POC (console output)
- [x] Add Info(), Error(), Debug() methods
- [x] Log all commands, errors, and trades to console
- [x] Add logger to bot, clients, and handlers

**Files:**
- `internal/logging/logger.go` - create logger
- `internal/logging/trade_log.go` - create JSONL trade logger

**Done when:**
- Logger outputs to stdout in readable format
- All commands logged with user_id and command text
- All errors logged with error message and context
- Trade operations logged with amounts and tokens

**Verify:**
```bash
go run ./cmd/telegram-bot 2>&1 | grep "INFO"
```

**Commit:**
```
feat(logging): implement basic logging infrastructure

- Add logger with Info, Error, Debug methods
- Log all user commands to console
- Log errors with context
- Add JSONL trade log file for audit

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-031 (Structured logging), FR-041 (Audit logging)
**Design:** Monitoring & Observability - Logging section

---

### Task 25: Set up Docker Compose for local development

**Do:**
- [x] Create `deployments/docker-compose.yml`
- [x] Add service definitions: bot-service, galachain-service, postgres
- [x] Configure environment variables from .env file
- [x] Set up volume mounts for logs/ directory
- [x] Add health checks for each service

**Files:**
- `deployments/docker-compose.yml` - create docker compose config
- `Dockerfile` (bot service) - create at project root
- `galachain-service/Dockerfile` - create for TS service

**Done when:**
- docker-compose up starts all services
- Services can communicate via docker network
- Logs persisted to logs/ directory
- Health checks pass for all services

**Verify:**
```bash
docker-compose -f deployments/docker-compose.yml up
```

**Commit:**
```
feat(deploy): add Docker Compose for local development

- Create docker-compose.yml with all services
- Add Dockerfiles for bot and GalaChain services
- Configure inter-service networking
- Add volume mounts for logs

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-043 (Setup scripts), Deployment focus: docker-compose
**Design:** Deployment Architecture section

---

### Task 26: Implement basic rate limiting

**Do:**
- [x] Create `internal/bot/middleware/ratelimit.go` with RateLimiter struct
- [x] Implement token bucket algorithm (10 commands per minute per user)
- [x] Add Allow() method that checks token availability
- [x] Wire into bot middleware before command handlers
- [x] Return "Too many requests, please wait" error when limit exceeded

**Files:**
- `internal/bot/middleware/ratelimit.go` - create rate limiter

**Done when:**
- Users can send 10 commands in quick succession
- 11th command returns rate limit error
- Tokens refill over time (1 token per 6 seconds)
- Rate limiting per user_id (not global)

**Verify:**
```bash
go test ./internal/bot/middleware -v
```

**Commit:**
```
feat(ratelimit): implement token bucket rate limiting

- Add rate limiter middleware
- Limit to 10 commands per minute per user
- Return user-friendly error when exceeded
- Implement token refill logic

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-039 (Rate limiting), NFR-017 (Rate limiting)
**Design:** Security Design - Rate Limiting Implementation

---

### Task 27: Implement basic key encryption for wallet credentials

**Do:**
- [x] Create `internal/wallet/encryption.go` with EncryptionService
- [x] Implement Encrypt() and Decrypt() methods using AES-256-GCM
- [x] Read master key from ENCRYPTION_KEY environment variable
- [x] Update wallet manager to encrypt private keys before saving to database
- [x] For POC: Store master key in .env (warn in comment this is dev-only)

**Files:**
- `internal/wallet/encryption.go` - create encryption service

**Done when:**
- Encryption/decryption works with AES-256-GCM
- Master key loaded from environment variable
- Private keys encrypted before database storage
- Decryption successful when retrieving keys

**Verify:**
```bash
go test ./internal/wallet -v
```

**Commit:**
```
feat(security): implement AES-256-GCM encryption for private keys

- Add encryption service with AES-256-GCM
- Encrypt wallet credentials before storage
- Load master key from environment
- Add decrypt method for key retrieval

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-012 (Key encryption), NFR-012 (AES-256 encryption)
**Design:** Security Design - Private Key Encryption

---

### Task 28: Create README with setup instructions

**Do:**
- [x] Create comprehensive README.md at project root
- [x] Add project description and features
- [x] Document prerequisites (Go 1.21+, Node.js 20+, Docker)
- [x] Add quick start guide with docker-compose
- [x] Document environment variables needed
- [x] Add command reference for all bot commands
- [x] Link to other documentation in docs/

**Files:**
- `README.md` - create project documentation

**Done when:**
- README includes project description
- Quick start section allows new user to run bot in 5 minutes
- All environment variables documented
- Command reference complete
- Links to detailed docs in docs/ folder

**Verify:**
```bash
cat README.md | grep "Quick Start"
```

**Commit:**
```
docs: add comprehensive README with setup instructions

- Add project description and features
- Document prerequisites and dependencies
- Add quick start guide with docker-compose
- Document all environment variables
- Include bot command reference

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-043 (Documentation), US-012 (Help and discovery)
**Design:** File Structure - docs/ section

---

## Phase 2: Refactoring

### Task 29: Implement real TON balance checking via blockchain

**Do:**
- [x] Update `internal/blockchain/ton/client.go` GetBalance() to call real TON blockchain
- [x] Use tonutils-go to fetch balance for wallet address
- [x] Handle both TON and GALA token balances on TON chain
- [x] Remove hardcoded balance values
- [x] Add error handling for network failures

**Files:**
- `internal/blockchain/ton/client.go` - update GetBalance implementation

**Done when:**
- GetBalance() returns real balance from TON blockchain
- Supports both native TON and jetton (GALA) balances
- Handles network errors gracefully
- Returns balance in proper decimal format

**Verify:**
```bash
go test ./internal/blockchain/ton -v -integration
```

**Commit:**
```
refactor(blockchain): implement real TON balance checking

- Connect to TON blockchain via tonutils-go
- Fetch native TON and jetton balances
- Remove hardcoded test values
- Add network error handling

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-017 (Balance checking)
**Design:** Component Design - TON Blockchain Client

---

### Task 30: Implement real TonConnect wallet connection protocol

**Do:**
- [x] Install TonConnect SDK: `go get github.com/ton-connect/sdk-go`
- [x] Update `internal/wallet/tonconnect.go` to use real TonConnect protocol
- [x] Generate proper TonConnect session with QR code and deep link
- [x] Listen for wallet connection events from TonConnect bridge
- [x] Store wallet address after successful connection
- [x] Remove manual address input workaround

**Files:**
- `internal/wallet/tonconnect.go` - implement TonConnect protocol

**Done when:**
- /wallet generates real TonConnect QR code
- Deep link works on mobile Telegram
- TonKeeper app connection succeeds
- Wallet address automatically saved after approval
- Session persists in database

**Verify:**
```bash
go test ./internal/wallet -v -integration
```

**Commit:**
```
refactor(wallet): implement real TonConnect protocol

- Add TonConnect SDK integration
- Generate proper connection session
- Listen for wallet connection events
- Store address after successful connection
- Remove manual address workaround

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-001 (TonConnect protocol), US-001 (Wallet connection)
**Design:** Data Flow Diagrams - Wallet Connection Flow (TonConnect)

---

### Task 31: Implement WalletConnect for Gala wallet

**Do:**
- [x] Install WalletConnect: `npm install @walletconnect/client`
- [x] Update `galachain-service/src/wallet/walletconnect.ts` with WalletConnect v2
- [x] Generate WalletConnect session URI with QR code
- [x] Implement CreateWalletSession gRPC handler properly
- [x] Listen for wallet approval events
- [x] Add manual private key fallback (from POC)

**Files:**
- `galachain-service/src/wallet/walletconnect.ts` - implement WalletConnect
- `galachain-service/src/server/handlers.ts` - update wallet session handler

**Done when:**
- WalletConnect session generates valid QR code
- Gala wallet app can scan and connect
- Session stored after approval
- Manual key input still works as fallback
- Both methods save to database

**Verify:**
```bash
cd galachain-service && npm test
```

**Commit:**
```
refactor(wallet): implement WalletConnect for Gala wallet

- Add WalletConnect v2 integration
- Generate connection session with QR code
- Handle wallet approval events
- Keep manual key input as fallback
- Update gRPC handler for wallet session creation

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-002 (WalletConnect protocol), US-002 (Gala wallet connection)
**Design:** Data Flow Diagrams - Wallet Connection Flow (WalletConnect)

---

### Task 32: Implement real gswap price fetching

**Do:**
- [x] Update `galachain-service/src/gswap/client.ts` getPrice() to call real gswap API
- [x] Remove hardcoded price value
- [x] Parse pool data and calculate GTON/GALA price
- [x] Add 5-second TTL caching for prices
- [x] Implement proper rate limiting (20 requests per 10 seconds)

**Files:**
- `galachain-service/src/gswap/client.ts` - update getPrice implementation
- `galachain-service/src/utils/rate-limiter.ts` - create rate limiter

**Done when:**
- getPrice() fetches real price from gswap API
- Price cached for 5 seconds
- Rate limiter enforces 20 req/10s limit
- Returns accurate GTON/GALA pool price

**Verify:**
```bash
cd galachain-service && npm test
```

**Commit:**
```
refactor(gswap): implement real price fetching with caching

- Call gswap API for real pool prices
- Remove hardcoded test values
- Add 5-second TTL cache
- Implement rate limiter for 20 req/10s
- Parse pool data to calculate price

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-020 (GalaChain price checking), FR-022 (Price caching)
**Design:** Component Design - GSwap Client

---

### Task 33: Implement real swap execution on ston.fi

**Do:**
- [x] Create `internal/blockchain/ton/transaction.go` for transaction building
- [x] Implement ExecuteSwap() method that builds and signs TON transaction
- [x] Call ston.fi router smart contract via tonutils-go
- [x] Wait for transaction confirmation (poll transaction status)
- [x] Return transaction hash and actual output amount
- [x] Update /swap handler to execute real swaps after confirmation

**Files:**
- `internal/blockchain/ton/transaction.go` - create transaction builder
- `internal/blockchain/ton/client.go` - update with ExecuteSwap
- `internal/bot/handlers/swap.go` - update to execute after user confirmation

**Done when:**
- ExecuteSwap() builds valid TON transaction
- Transaction submitted to ston.fi smart contract
- Polls status until success or failure
- Returns tx hash and actual swap output
- /swap command executes real swaps on confirmation

**Verify:**
```bash
go test ./internal/blockchain/ton -v -integration
```

**Commit:**
```
refactor(swap): implement real swap execution on ston.fi

- Build and sign TON transactions
- Call ston.fi router smart contract
- Poll transaction status until completion
- Return tx hash and actual amounts
- Execute swaps after user confirmation

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-003 (Swap execution), US-004 (Manual token swap)
**Design:** Data Flow Diagrams - Manual Swap Execution Flow

---

### Task 34: Implement real swap execution on gswap

**Do:**
- [x] Create `galachain-service/src/gswap/swap.ts` for swap execution
- [x] Update GSwapClient.executeSwap() to call real gswap API
- [x] Authorize fee credit (AuthorizeFee endpoint)
- [x] Submit swap via RequestTokenSwap endpoint
- [x] Poll swap status until completion
- [x] Update ExecuteSwap gRPC handler to call real swap

**Files:**
- `galachain-service/src/gswap/swap.ts` - create swap executor
- `galachain-service/src/gswap/client.ts` - update executeSwap method
- `galachain-service/src/server/handlers.ts` - update ExecuteSwap handler

**Done when:**
- executeSwap() submits real swap to gswap API
- Fee credit authorized before swap
- Swap status polled until success/failure
- Returns tx hash and actual output amount
- gRPC handler executes real swaps

**Verify:**
```bash
cd galachain-service && npm run test:integration
```

**Commit:**
```
refactor(swap): implement real swap execution on gswap

- Submit real swaps via gswap API
- Authorize fee credits before swap
- Poll swap status until completion
- Return transaction hash and amounts
- Update gRPC handler for real execution

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-004 (gswap swap execution), US-005 (Manual swap on gswap)
**Design:** Component Design - Swap Executor

---

### Task 35: Implement position sizing calculator for arbitrage

**Do:**
- [x] Create `internal/arbitrage/calculator.go` with position sizing logic
- [x] Implement CalculatePositionSize() that uses 50% of balance
- [x] Check minimum balances (1 TON, 10 GALA)
- [x] Apply 1.5x safety margin above minimums
- [x] Calculate exact amounts for both legs of arbitrage
- [x] Update arbitrage engine to use calculator

**Files:**
- `internal/arbitrage/calculator.go` - create position size calculator

**Done when:**
- CalculatePositionSize() returns 50% of available balance
- Ensures 1 TON minimum remains after trade
- Ensures 10 GALA minimum remains after trade
- Safety margin of 1.5x applied
- Returns exact amounts for buy and sell legs

**Verify:**
```bash
go test ./internal/arbitrage -v
```

**Commit:**
```
refactor(arbitrage): implement position sizing calculator

- Calculate 50% position size
- Check minimum balance requirements
- Apply 1.5x safety margin
- Calculate amounts for both legs
- Validate sufficient balance before trade

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-027 (Position sizing), FR-024 (Minimum balance checks)
**Design:** Component Design - Arbitrage Engine

---

### Task 36: Implement arbitrage execution coordinator

**Do:**
- [x] Create `internal/arbitrage/executor.go` with Executor struct
- [x] Implement ExecuteArbitrage() that runs both legs concurrently
- [x] Execute buy on ston.fi and sell on gswap in parallel
- [x] Wait for both transactions to complete
- [x] Handle partial success (one leg fails)
- [x] Calculate actual profit from completed transactions
- [x] Update /arbitrage handler to use executor

**Files:**
- `internal/arbitrage/executor.go` - create arbitrage executor

**Done when:**
- ExecuteArbitrage() runs both swaps concurrently
- Uses goroutines for parallel execution
- Waits for both to complete before returning
- Handles partial success gracefully
- Calculates actual profit after fees
- Returns detailed execution result

**Verify:**
```bash
go test ./internal/arbitrage -v -integration
```

**Commit:**
```
refactor(arbitrage): implement concurrent arbitrage execution

- Execute both legs in parallel
- Use goroutines for concurrent swaps
- Handle partial success scenarios
- Calculate actual profit after completion
- Return detailed execution result

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-009 (Arbitrage execution), US-010 (Manual arbitrage)
**Design:** Data Flow Diagrams - Arbitrage Execution Flow

---

### Task 37: Implement JSONL trade logging

**Do:**
- [x] Update `internal/logging/trade_log.go` to append to logs/trades.jsonl
- [x] Log every swap and arbitrage execution
- [x] Include: timestamp, user_id, type, amounts, tokens, fees, tx_hashes, status
- [x] Ensure atomic writes (use file locking)
- [x] Never modify existing lines (append-only)
- [x] Update all swap and arbitrage handlers to log trades

**Files:**
- `internal/logging/trade_log.go` - update trade logger
- `logs/trades.jsonl` - create log file

**Done when:**
- Every swap logged to trades.jsonl
- Every arbitrage logged with both tx hashes
- Log format matches design.md TradeLogEntry
- File is append-only (never overwrites)
- Handlers call trade logger after execution

**Verify:**
```bash
cat logs/trades.jsonl | jq .
```

**Commit:**
```
refactor(logging): implement JSONL trade logging

- Append all trades to trades.jsonl file
- Log swaps and arbitrage executions
- Include tx hashes, amounts, fees, status
- Use append-only file writes
- Update handlers to log all trades

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-007 (Trade tracking), FR-049 (JSONL logging)
**Design:** Data Models - JSONL Trade Logs

---

### Task 38: Improve error handling with retry logic

**Do:**
- [x] Create `internal/utils/retry.go` with exponential backoff retry logic
- [x] Implement RetryWithBackoff() function (initial 1s, max 10s, 3 retries)
- [x] Add jitter to prevent thundering herd (random 0-500ms)
- [x] Update blockchain clients to use retry logic for network calls
- [x] Update DEX clients to use retry for API calls
- [x] Don't retry on user errors (invalid input, insufficient balance)

**Files:**
- `internal/utils/retry.go` - create retry utility

**Done when:**
- RetryWithBackoff() implements exponential backoff
- Retries 3 times with increasing delays
- Adds random jitter to each retry
- Blockchain and DEX clients use retry logic
- User errors not retried (return immediately)

**Verify:**
```bash
go test ./internal/utils -v
```

**Commit:**
```
refactor(errors): add exponential backoff retry logic

- Implement retry with exponential backoff
- Add random jitter to prevent thundering herd
- Retry network errors up to 3 times
- Skip retry for user/validation errors
- Update clients to use retry logic

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-024 (Retry logic), NFR-023 (Graceful failure handling)
**Design:** Error Handling - Retry Logic Strategy

---

### Task 39: Add input validation for all commands

**Do:**
- [x] Create `internal/bot/validation.go` with validation functions
- [x] Implement ValidateTonAddress() with checksum verification
- [x] Implement ValidateGalaAddress() with hex and length checks
- [x] Implement ValidateAmount() with positive check and decimal limit
- [x] Implement ValidateTokenSymbol() with whitelist (TON, GALA, GTON)
- [x] Update all command handlers to validate inputs before processing

**Files:**
- `internal/bot/validation.go` - create validation utilities

**Done when:**
- All validation functions implemented per design.md
- TON addresses validated with checksum
- Amounts validated as positive with max 18 decimals
- Token symbols validated against whitelist
- Command handlers reject invalid inputs with clear messages

**Verify:**
```bash
go test ./internal/bot -v
```

**Commit:**
```
refactor(validation): add comprehensive input validation

- Validate TON addresses with checksum
- Validate GalaChain addresses (hex, length)
- Validate amounts (positive, max 18 decimals)
- Whitelist token symbols
- Update handlers to validate before processing

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-011 (Wallet validation), NFR-044 (Input validation)
**Design:** Security Design - Input Validation

---

### Task 40: Implement health checks for all services

**Do:**
- [x] Add HealthCheck() method to bot service checking database, TON client, gRPC connection
- [x] Implement HealthCheck gRPC handler in GalaChain service
- [x] Check gswap API connectivity, database connection
- [x] Add /health HTTP endpoint to bot service (for monitoring)
- [x] Add /health HTTP endpoint to GalaChain service
- [x] Return status: healthy/degraded/unhealthy with dependency details

**Files:**
- `internal/bot/health.go` - create health check for bot
- `galachain-service/src/server/health.ts` - create health check handler

**Done when:**
- Bot service /health endpoint returns JSON status
- GalaChain service HealthCheck gRPC works
- Both check all critical dependencies
- Returns degraded if one dependency fails
- Returns unhealthy if multiple dependencies fail

**Verify:**
```bash
curl http://localhost:8080/health
```

**Commit:**
```
refactor(monitoring): implement health checks for all services

- Add health check endpoints to both services
- Check critical dependencies (DB, APIs, gRPC)
- Return status with dependency details
- Support degraded and unhealthy states
- Enable monitoring tool integration

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-027 (Health checks), NFR-033 (Health endpoints)
**Design:** Monitoring & Observability section

---

### Task 41: Migrate from SQLite to PostgreSQL option

**Do:**
- [x] Create `internal/storage/postgres.go` implementing Database interface
- [x] Use `github.com/lib/pq` driver for PostgreSQL
- [x] Keep SQLite implementation for local dev
- [x] Add DATABASE_URL parsing to detect SQLite vs PostgreSQL
- [x] Update docker-compose.yml to include PostgreSQL service
- [x] Create migration script to copy SQLite data to PostgreSQL

**Files:**
- `internal/storage/postgres.go` - create PostgreSQL adapter
- `deployments/docker-compose.yml` - add PostgreSQL service

**Done when:**
- PostgreSQL implementation works identically to SQLite
- Both adapters implement same Database interface
- docker-compose includes PostgreSQL container
- Can switch via DATABASE_URL environment variable
- Migration script successfully copies data

**Verify:**
```bash
DATABASE_URL=postgres://user:pass@localhost:5432/testdb go test ./internal/storage -v
```

**Commit:**
```
refactor(storage): add PostgreSQL support alongside SQLite

- Implement PostgreSQL adapter
- Keep SQLite for local development
- Auto-detect database type from URL
- Add PostgreSQL to docker-compose
- Create data migration script

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-048 (Data persistence), NFR-009 (Database scalability)
**Design:** Technical Decisions - Database choice

---

### Task 42: Implement correlation ID tracking across services

**Do:**
- [x] Create `internal/logging/correlation.go` for correlation ID generation
- [x] Generate UUID v4 correlation ID for each command
- [x] Pass correlation ID in gRPC metadata to GalaChain service
- [x] Extract correlation ID in TypeScript service
- [x] Include correlation ID in all log entries
- [x] Return correlation ID in error messages to users

**Files:**
- `internal/logging/correlation.go` - create correlation ID utilities
- `galachain-service/src/logging/correlation.ts` - create correlation ID handler

**Done when:**
- Each command assigned unique correlation ID
- Correlation ID passed via gRPC metadata
- TypeScript service extracts and uses correlation ID
- All logs include correlation_id field
- Error messages show correlation ID for support

**Verify:**
```bash
go run ./cmd/telegram-bot 2>&1 | jq .correlation_id
```

**Commit:**
```
refactor(logging): implement correlation ID tracking

- Generate unique correlation ID per command
- Pass correlation ID via gRPC metadata
- Include in all log entries across services
- Return correlation ID in error messages
- Enable request tracing across service boundary

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-032 (Correlation IDs), NFR-031 (Structured logging)
**Design:** Monitoring & Observability section

---

### Task 43: Implement structured JSON logging

**Do:**
- [x] Update `internal/logging/logger.go` to output JSON format
- [x] Include fields: timestamp, level, service, event_type, user_id, correlation_id, message
- [x] Replace fmt.Println with proper logger calls throughout codebase
- [x] Add LOG_FORMAT env var (json or text) for switching
- [x] Update TypeScript service logger to match JSON format

**Files:**
- `internal/logging/logger.go` - update to JSON format
- `galachain-service/src/logging/logger.ts` - update to JSON format

**Done when:**
- All logs output as valid JSON (when LOG_FORMAT=json)
- JSON includes all required fields from design.md
- No more fmt.Println in codebase
- TypeScript logs match Go log format
- Can switch to text format for local dev

**Verify:**
```bash
LOG_FORMAT=json go run ./cmd/telegram-bot 2>&1 | jq .
```

**Commit:**
```
refactor(logging): migrate to structured JSON logging

- Output logs in JSON format
- Include standard fields (timestamp, level, correlation_id)
- Remove fmt.Println calls
- Add LOG_FORMAT toggle for dev/prod
- Align TypeScript logs with Go format

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-031 (Structured logging)
**Design:** Monitoring & Observability - Logging section

---

### Task 44: Add CoinGecko integration for USD pricing

**Do:**
- [x] Create `internal/price/coingecko.go` with CoinGecko client
- [x] Implement GetUSDPrice() for TON, GALA, GTON tokens
- [x] Cache prices for 5 minutes (longer than trading prices)
- [x] Update /balance handler to show USD values
- [x] Update /price handler to show USD equivalents
- [x] Handle CoinGecko API failures gracefully (show crypto amounts only)

**Files:**
- `internal/price/coingecko.go` - create CoinGecko client

**Done when:**
- CoinGecko client fetches USD prices successfully
- Prices cached for 5 minutes
- /balance shows USD values next to crypto amounts
- /price shows USD equivalent for trades
- Falls back gracefully if CoinGecko unavailable

**Verify:**
```bash
go test ./internal/price -v
```

**Commit:**
```
refactor(price): add CoinGecko USD pricing integration

- Create CoinGecko API client
- Fetch USD prices for TON, GALA, GTON
- Cache prices for 5 minutes
- Update balance and price commands with USD values
- Graceful degradation if API unavailable

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-018 (USD values)
**Design:** Architecture diagram shows CoinGecko integration

---

### Task 45: Refactor configuration to use YAML files

**Do:**
- [x] Create `configs/bot-service.yaml` with config from design.md
- [x] Create `configs/galachain-service.yaml` for TypeScript service
- [x] Install viper: `go get github.com/spf13/viper`
- [x] Update config loader to read YAML files
- [x] Keep environment variables for secrets (BOT_TOKEN, ENCRYPTION_KEY)
- [x] Support environment variable overrides of YAML values

**Files:**
- `configs/bot-service.yaml` - create bot config file
- `configs/galachain-service.yaml` - create GalaChain config file
- `internal/config/config.go` - update to load YAML

**Done when:**
- YAML files contain non-secret configuration
- Config loader reads YAML files with viper
- Environment variables still work for secrets
- Environment variables can override YAML values
- Both services use YAML + env var pattern

**Verify:**
```bash
go run ./cmd/telegram-bot --config=configs/bot-service.yaml
```

**Commit:**
```
refactor(config): migrate to YAML configuration files

- Create YAML config files for both services
- Add viper for config loading
- Keep env vars for secrets
- Support env var overrides of YAML
- Separate config from secrets

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-042 (Externalized configuration)
**Design:** Configuration Management section

---

### Task 46: Implement session expiry and renewal

**Do:**
- [x] Update SessionManager to check session expiry (24 hours)
- [x] Implement sliding window: renew expiry on each command
- [x] Set session.ExpiresAt = now + 24h on each authenticated request
- [x] Return session expired error if session > 24h old
- [x] Require /start to create new session after expiry

**Files:**
- `internal/bot/middleware/auth.go` - update session checking

**Done when:**
- Sessions expire after 24 hours of inactivity
- Active users' sessions renewed automatically
- Expired sessions return clear error message
- /start command creates new session
- Session expiry stored and checked in database

**Verify:**
```bash
go test ./internal/bot/middleware -v
```

**Commit:**
```
refactor(auth): implement session expiry and renewal

- Check session age on each command
- Implement 24-hour sliding window
- Renew session expiry on activity
- Return expired error after 24h inactivity
- Require /start to recreate expired session

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-016 (Session expiry)
**Design:** Security Design - Session Management

---

### Task 47: Add /disconnect command for wallet

**Do:**
- [x] Create disconnect handler in `internal/bot/handlers/wallet.go`
- [x] Implement DisconnectWallet() in wallet manager
- [x] Delete wallet_sessions row from database
- [x] Clear encrypted private key from memory
- [x] Show confirmation message to user
- [x] Add inline keyboard [Confirm Disconnect] for destructive action

**Files:**
- `internal/bot/handlers/wallet.go` - add disconnect handler

**Done when:**
- /disconnect shows confirmation prompt
- User must click [Confirm Disconnect] button
- Wallet session deleted from database
- Encrypted keys removed
- User sees "Wallet disconnected successfully"

**Verify:**
```bash
go test ./internal/bot/handlers -v
```

**Commit:**
```
refactor(wallet): add /disconnect command for wallet removal

- Implement wallet disconnection
- Delete session from database
- Clear encrypted keys
- Add confirmation prompt for safety
- Update user after successful disconnect

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** FR-014 (Wallet disconnection)
**Design:** Data Flow Diagrams - Wallet operations

---

### Task 48: Implement /orders command for trade history

**Do:**
- [x] Create `internal/bot/handlers/orders.go` with orders handler
- [x] Query trade_history table for user's trades
- [x] Show last 10 trades by default
- [x] Format each trade: timestamp, type, tokens, amounts, status, tx hash link
- [x] Support /orders <count> to show more (max 100)
- [x] Add inline buttons [Refresh] [Export CSV]

**Files:**
- `internal/bot/handlers/orders.go` - create orders handler
- `internal/storage/db.go` - add GetTradeHistory method

**Done when:**
- /orders shows last 10 trades
- /orders 50 shows last 50 trades
- Each trade formatted clearly with all details
- TX hashes are clickable links to explorers
- Refresh button reloads trade list

**Verify:**
```bash
go test ./internal/bot/handlers -v
```

**Commit:**
```
refactor(bot): implement /orders command for trade history

- Add orders command handler
- Query trade_history from database
- Show last N trades with formatting
- Make tx hashes clickable links
- Add refresh and export buttons

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** US-009 (Trade history)
**Design:** Data Models - Trade History

---

## Phase 3: Testing

### Task 49: Set up test infrastructure and mocks

**Do:**
- [x] Install testify: `go get github.com/stretchr/testify`
- [x] Create `test/mocks/` directory with mock implementations
- [x] Create mock Database (in-memory)
- [x] Create mock TON client (test fixtures)
- [x] Create mock ston.fi client (hardcoded responses)
- [x] Create mock gRPC client
- [x] Create test helpers in `test/testutil/` for common setup

**Files:**
- `test/mocks/database.go` - mock database
- `test/mocks/blockchain.go` - mock blockchain client
- `test/mocks/dex.go` - mock DEX client
- `test/mocks/grpc.go` - mock gRPC client
- `test/testutil/helpers.go` - test utilities

**Done when:**
- All mock implementations follow interfaces
- Mocks can be configured with test data
- Test helpers simplify test setup
- In-memory database resets between tests

**Verify:**
```bash
go test ./test/mocks/... -v
```

**Commit:**
```
test(setup): add test infrastructure and mocks

- Create mock implementations for all interfaces
- Add in-memory database for testing
- Create test fixtures for blockchain responses
- Add test helpers for common setup
- Enable isolated unit testing

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage 70%)
**Design:** Testing Strategy - Unit Tests

---

### Task 50: Add unit tests for authentication middleware

**Do:**
- [x] Create `internal/bot/middleware/auth_test.go`
- [x] Test whitelisted user authentication (success)
- [x] Test non-whitelisted user authentication (failure)
- [x] Test session creation on first access
- [x] Test session expiry after 24 hours
- [x] Test session renewal on activity
- [x] Aim for 90% coverage of auth.go

**Files:**
- `internal/bot/middleware/auth_test.go` - create auth tests

**Done when:**
- All authentication scenarios tested
- Tests use mock database
- Coverage >= 90% for auth middleware
- Tests run in <1 second

**Verify:**
```bash
go test ./internal/bot/middleware -v -cover
```

**Commit:**
```
test(auth): add unit tests for authentication middleware

- Test whitelisted and non-whitelisted users
- Test session creation and expiry
- Test session renewal logic
- Achieve 90% code coverage
- Use mock database for isolation

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage)
**Design:** Testing Strategy section

---

### Task 51: Add unit tests for rate limiting

**Do:**
- [x] Create `internal/bot/middleware/ratelimit_test.go`
- [x] Test token bucket algorithm (10 tokens, refill rate)
- [x] Test burst allowance (3 quick commands)
- [x] Test rate limit enforcement (11th command blocked)
- [x] Test token refill over time
- [x] Test per-user isolation (user A doesn't affect user B)
- [x] Aim for 90% coverage

**Files:**
- `internal/bot/middleware/ratelimit_test.go` - create rate limit tests

**Done when:**
- Token bucket behavior fully tested
- Burst and refill tested with time mocking
- Per-user isolation verified
- Coverage >= 90%

**Verify:**
```bash
go test ./internal/bot/middleware -v -cover
```

**Commit:**
```
test(ratelimit): add unit tests for rate limiting

- Test token bucket algorithm
- Test burst allowance and refill
- Test rate limit enforcement
- Test per-user isolation
- Achieve 90% coverage

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage)
**Design:** Testing Strategy section

---

### Task 52: Add unit tests for arbitrage calculator

**Do:**
- [x] Create `internal/arbitrage/calculator_test.go`
- [x] Test position sizing (50% of balance)
- [x] Test minimum balance enforcement (1 TON, 10 GALA)
- [x] Test safety margin calculation (1.5x)
- [x] Test insufficient balance scenarios
- [x] Test edge cases (exactly at minimum, slightly above minimum)
- [x] Aim for 90% coverage

**Files:**
- `internal/arbitrage/calculator_test.go` - create calculator tests

**Done when:**
- Position sizing logic fully tested
- All minimum balance scenarios covered
- Edge cases tested
- Coverage >= 90%

**Verify:**
```bash
go test ./internal/arbitrage -v -cover
```

**Commit:**
```
test(arbitrage): add unit tests for position calculator

- Test 50% position sizing
- Test minimum balance enforcement
- Test safety margin calculation
- Test insufficient balance scenarios
- Achieve 90% coverage

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage)
**Design:** Testing Strategy section

---

### Task 53: Add unit tests for encryption service

**Do:**
- [x] Create `internal/wallet/encryption_test.go`
- [x] Test AES-256-GCM encryption/decryption
- [x] Test round-trip (encrypt then decrypt returns original)
- [x] Test different master keys produce different ciphertext
- [x] Test invalid ciphertext returns error on decrypt
- [x] Test nonce randomness (same plaintext -> different ciphertext)
- [x] Aim for 100% coverage (critical security component)

**Files:**
- `internal/wallet/encryption_test.go` - create encryption tests

**Done when:**
- All encryption scenarios tested
- Security properties verified (nonce randomness, different keys)
- Round-trip successful
- Coverage = 100%

**Verify:**
```bash
go test ./internal/wallet -v -cover
```

**Commit:**
```
test(security): add unit tests for encryption service

- Test AES-256-GCM encryption/decryption
- Test round-trip correctness
- Test nonce randomness
- Test error handling for invalid ciphertext
- Achieve 100% coverage for critical security code

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage), NFR-012 (Encryption)
**Design:** Security Design - Private Key Encryption

---

### Task 54: Add unit tests for input validation

**Do:**
- [x] Create `internal/bot/validation_test.go`
- [x] Test TON address validation (valid EQ/UQ, invalid prefix, bad checksum)
- [x] Test GalaChain address validation (valid hex 64-char, invalid length, non-hex)
- [x] Test amount validation (positive, negative, zero, too many decimals, non-numeric)
- [x] Test token symbol validation (whitelist, invalid symbols)
- [x] Aim for 90% coverage

**Files:**
- `internal/bot/validation_test.go` - create validation tests

**Done when:**
- All validation functions tested
- Valid and invalid cases covered
- Edge cases tested (empty strings, max length, etc.)
- Coverage >= 90%

**Verify:**
```bash
go test ./internal/bot -v -cover
```

**Commit:**
```
test(validation): add unit tests for input validation

- Test TON address validation with checksums
- Test GalaChain address format validation
- Test amount validation (positive, decimals)
- Test token symbol whitelist
- Achieve 90% coverage

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage)
**Design:** Security Design - Input Validation

---

### Task 55: Add unit tests for command handlers

**Do:**
- [x] Create test files for each handler: start_test.go, balance_test.go, swap_test.go, etc.
- [x] Use mock bot API, mock clients
- [x] Test happy path for each command
- [x] Test error cases (wallet not connected, insufficient balance, etc.)
- [x] Test response formatting
- [x] Aim for 80% coverage of handlers package

**Files:**
- `internal/bot/handlers/*_test.go` - create handler tests

**Done when:**
- All major commands have tests
- Happy path and error cases covered
- Mock dependencies used
- Coverage >= 80%

**Verify:**
```bash
go test ./internal/bot/handlers -v -cover
```

**Commit:**
```
test(handlers): add unit tests for command handlers

- Test all command handlers with mocks
- Cover happy path and error scenarios
- Test response formatting
- Verify error messages match design
- Achieve 80% coverage

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage)
**Design:** Testing Strategy section

---

### Task 56: Add unit tests for TypeScript GalaChain service

**Do:**
- [x] Set up Jest for TypeScript: `npm install --save-dev jest ts-jest @types/jest`
- [x] Create `galachain-service/src/**/__tests__/` directories
- [x] Add tests for GSwapClient (getPrice, executeSwap)
- [x] Add tests for gRPC handlers
- [x] Add tests for wallet manager
- [x] Mock gswap SDK responses
- [x] Aim for 70% coverage

**Files:**
- `galachain-service/src/**/__tests__/*.test.ts` - create TypeScript tests
- `galachain-service/jest.config.js` - create Jest config

**Done when:**
- Jest configured for TypeScript
- All major modules have tests
- gswap SDK calls mocked
- Coverage >= 70%

**Verify:**
```bash
cd galachain-service && npm test
```

**Commit:**
```
test(galachain): add unit tests for TypeScript service

- Set up Jest testing framework
- Test GSwapClient with mocked SDK
- Test gRPC handlers
- Test wallet manager
- Achieve 70% coverage

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage)
**Design:** Testing Strategy - TypeScript Testing

---

### Task 57: Add integration test for gRPC communication

**Do:**
- [x] Create `test/integration/grpc_test.go`
- [x] Start real GalaChain gRPC server in test
- [x] Create real gRPC client
- [x] Test GetPrice RPC call
- [x] Test GetBalance RPC call
- [x] Test ExecuteSwap RPC call (mock execution, verify request/response)
- [x] Test error handling (server unreachable)

**Files:**
- `test/integration/grpc_test.go` - create integration test

**Done when:**
- Real gRPC server started in test
- All RPC methods called successfully
- Request/response serialization verified
- Error cases tested
- Tests pass with real gRPC communication

**Verify:**
```bash
go test ./test/integration -v -tags=integration
```

**Commit:**
```
test(integration): add gRPC communication tests

- Start real gRPC server for testing
- Test all RPC methods end-to-end
- Verify request/response serialization
- Test error handling and retries
- Validate cross-service communication

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage)
**Design:** Testing Strategy - Integration Tests

---

### Task 58: Add integration test for ston.fi API

**Do:**
- [x] Create `test/integration/stonfi_test.go`
- [x] Test real SimulateSwap API call (1 TON -> GALA)
- [x] Verify response parsing (amount, fee, slippage)
- [x] Test with different token pairs
- [x] Test error handling (invalid token, network error)
- [x] Mark as integration test (requires network)

**Files:**
- `test/integration/stonfi_test.go` - create ston.fi integration test

**Done when:**
- Real ston.fi API called successfully
- Response parsed correctly
- Different token pairs tested
- Error cases handled
- Tests pass with live API

**Verify:**
```bash
go test ./test/integration -v -tags=integration -run TestStonFi
```

**Commit:**
```
test(integration): add ston.fi API integration tests

- Test real swap simulation API
- Verify response parsing
- Test multiple token pairs
- Test error handling
- Validate API integration

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage)
**Design:** Testing Strategy - Integration Tests

---

### Task 59: Add integration test for database operations

**Do:**
- [x] Create `test/integration/database_test.go`
- [x] Test with real SQLite database (temp file)
- [x] Test user session CRUD operations
- [x] Test wallet session CRUD operations
- [x] Test trade history logging
- [x] Test concurrent writes (goroutines)
- [x] Clean up temp database after test

**Files:**
- `test/integration/database_test.go` - create database integration test

**Done when:**
- Real SQLite database used in test
- All CRUD operations tested
- Concurrent writes verified
- Temp database cleaned up
- Tests pass with real database

**Verify:**
```bash
go test ./test/integration -v -tags=integration -run TestDatabase
```

**Commit:**
```
test(integration): add database integration tests

- Test real SQLite database operations
- Test user and wallet session CRUD
- Test trade history logging
- Test concurrent write safety
- Verify data persistence

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage)
**Design:** Testing Strategy - Integration Tests

---

### Task 60: Add end-to-end wallet connection test

**Do:**
- [x] Create `test/e2e/wallet_test.go`
- [x] Simulate user sending /wallet command
- [x] Verify bot response with TonConnect link
- [x] Simulate wallet connection callback (mock TonConnect bridge)
- [x] Verify wallet session saved to database
- [x] Verify confirmation message sent to user

**Files:**
- `test/e2e/wallet_test.go` - create e2e wallet test

**Done when:**
- Full /wallet command flow tested
- TonConnect session generation verified
- Wallet callback handled correctly
- Database updated with session
- User receives confirmation

**Verify:**
```bash
go test ./test/e2e -v -tags=e2e
```

**Commit:**
```
test(e2e): add end-to-end wallet connection test

- Test complete /wallet command flow
- Simulate user interaction
- Mock TonConnect callback
- Verify database persistence
- Validate user confirmation

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage)
**Design:** Testing Strategy - End-to-End Tests

---

### Task 61: Add end-to-end swap execution test

**Do:**
- [x] Create `test/e2e/swap_test.go`
- [x] Use testnet for real blockchain testing (TON testnet, GalaChain testnet)
- [x] Create test wallet with test funds
- [x] Execute /swap command with small amount (0.001 TON)
- [x] Verify swap simulation response
- [x] Simulate user clicking [Confirm]
- [x] Verify swap execution on testnet
- [x] Verify transaction logged to trades.jsonl

**Files:**
- `test/e2e/swap_test.go` - create e2e swap test

**Done when:**
- Full /swap flow tested on testnet
- Real blockchain transaction executed
- Transaction confirmed on chain
- Trade logged correctly
- Test completes without manual intervention

**Verify:**
```bash
go test ./test/e2e -v -tags=e2e -run TestSwap
```

**Commit:**
```
test(e2e): add end-to-end swap execution test

- Test complete /swap command flow on testnet
- Execute real blockchain transaction
- Verify transaction confirmation
- Validate trade logging
- Test with minimal test funds

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage), Testing approach: integration tests
**Design:** Testing Strategy - End-to-End Tests

---

### Task 62: Add end-to-end arbitrage test

**Do:**
- [x] Create `test/e2e/arbitrage_test.go`
- [x] Use testnet with test wallets on both chains
- [x] Execute /arbitrage command
- [x] Verify opportunity detection
- [x] Simulate user clicking [Execute]
- [x] Verify both legs execute on testnet
- [x] Verify trade logging for arbitrage
- [x] Calculate actual profit/loss

**Files:**
- `test/e2e/arbitrage_test.go` - create e2e arbitrage test

**Done when:**
- Full /arbitrage flow tested on testnet
- Both swaps execute successfully
- Arbitrage result calculated
- Trade logged with both tx hashes
- Test uses minimal test funds

**Verify:**
```bash
go test ./test/e2e -v -tags=e2e -run TestArbitrage
```

**Commit:**
```
test(e2e): add end-to-end arbitrage test

- Test complete /arbitrage flow on testnet
- Execute real cross-chain arbitrage
- Verify both legs complete
- Calculate actual profit
- Validate comprehensive trade logging

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage), Testing approach: integration tests
**Design:** Testing Strategy - End-to-End Tests

---

### Task 63: Set up test coverage reporting

**Do:**
- [x] Create `scripts/coverage.sh` script
- [x] Run all unit tests with coverage: `go test -coverprofile=coverage.out ./...`
- [x] Generate HTML coverage report: `go tool cover -html=coverage.out -o coverage.html`
- [x] Add coverage badge to README
- [x] Set coverage threshold: fail CI if <70%
- [x] Add coverage for TypeScript with Istanbul

**Files:**
- `scripts/coverage.sh` - create coverage script
- `.github/workflows/ci.yml` - update with coverage check

**Done when:**
- Coverage script runs all tests
- HTML report generated
- Coverage threshold enforced
- TypeScript coverage included
- CI fails if coverage <70%

**Verify:**
```bash
bash scripts/coverage.sh
```

**Commit:**
```
test(coverage): add test coverage reporting

- Create coverage generation script
- Generate HTML coverage reports
- Add coverage threshold enforcement (70%)
- Include TypeScript coverage
- Update CI to check coverage

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-040 (Test coverage 70%)
**Design:** Testing Strategy - Test Coverage Targets

---

### Task 64: Create testing documentation

**Do:**
- [ ] Create `docs/TESTING.md`
- [ ] Document how to run unit tests
- [ ] Document how to run integration tests (requires network)
- [ ] Document how to run e2e tests (requires testnet accounts)
- [ ] Document test coverage requirements
- [ ] Add instructions for setting up test environment
- [ ] Document test data and fixtures

**Files:**
- `docs/TESTING.md` - create testing documentation

**Done when:**
- Testing guide complete and accurate
- All test types documented
- Setup instructions clear
- Coverage requirements documented
- Examples provided

**Verify:**
```bash
cat docs/TESTING.md | grep "Running Tests"
```

**Commit:**
```
docs: add comprehensive testing documentation

- Document all test types and how to run them
- Add test environment setup instructions
- Document coverage requirements
- Provide test data and fixture info
- Add troubleshooting section

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-043 (Documentation)
**Design:** docs/ section

---

## Phase 4: Quality Gates

### Task 65: Set up golangci-lint

**Do:**
- [ ] Install golangci-lint: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- [ ] Create `.golangci.yml` configuration
- [ ] Enable linters: govet, errcheck, staticcheck, gosimple, ineffassign, unused, gofmt
- [ ] Fix all linting errors in codebase
- [ ] Add lint command to Makefile

**Files:**
- `.golangci.yml` - create linter configuration
- `Makefile` - add lint target

**Done when:**
- golangci-lint configuration created
- All enabled linters pass
- No linting errors in codebase
- Makefile includes `make lint` command

**Verify:**
```bash
golangci-lint run ./...
```

**Commit:**
```
chore(lint): set up golangci-lint for Go code

- Add golangci-lint configuration
- Enable recommended linters
- Fix all linting errors
- Add lint target to Makefile

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-038 (Code style), NFR-039 (Documentation)
**Design:** Phase 4 guidance - Go linting

---

### Task 66: Set up ESLint and Prettier for TypeScript

**Do:**
- [ ] Install ESLint: `npm install --save-dev eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin`
- [ ] Install Prettier: `npm install --save-dev prettier eslint-config-prettier`
- [ ] Create `.eslintrc.json` configuration
- [ ] Create `.prettierrc` configuration
- [ ] Fix all linting errors in TypeScript codebase
- [ ] Add lint and format scripts to package.json

**Files:**
- `galachain-service/.eslintrc.json` - create ESLint config
- `galachain-service/.prettierrc` - create Prettier config
- `galachain-service/package.json` - add lint/format scripts

**Done when:**
- ESLint and Prettier configured
- All TypeScript files pass linting
- Code formatted consistently
- npm run lint passes
- npm run format succeeds

**Verify:**
```bash
cd galachain-service && npm run lint && npm run format
```

**Commit:**
```
chore(lint): set up ESLint and Prettier for TypeScript

- Add ESLint with TypeScript support
- Add Prettier for code formatting
- Fix all linting errors
- Add lint and format npm scripts

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-038 (Code style)
**Design:** Phase 4 guidance - TypeScript linting

---

### Task 67: Enable strict TypeScript mode

**Do:**
- [ ] Update `tsconfig.json` with strict mode options
- [ ] Enable: strict, noImplicitAny, strictNullChecks, strictFunctionTypes
- [ ] Fix all type errors introduced by strict mode
- [ ] Ensure all functions have return types
- [ ] Ensure all variables have explicit types where needed

**Files:**
- `galachain-service/tsconfig.json` - update with strict mode

**Done when:**
- tsconfig.json has strict: true
- All strict mode errors fixed
- Code compiles without type errors
- Type safety improved

**Verify:**
```bash
cd galachain-service && npm run build
```

**Commit:**
```
chore(typescript): enable strict mode and fix type errors

- Enable TypeScript strict mode
- Fix all type errors
- Add explicit return types
- Improve type safety throughout codebase

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-038 (Code quality)
**Design:** Phase 4 guidance - Type checking

---

### Task 68: Set up pre-commit hooks

**Do:**
- [ ] Install pre-commit: `pip install pre-commit` or use Go equivalent
- [ ] Create `.pre-commit-config.yaml`
- [ ] Add hooks: gofmt, golangci-lint, go test, eslint, prettier
- [ ] Install hooks: `pre-commit install`
- [ ] Test hooks work on commit

**Files:**
- `.pre-commit-config.yaml` - create pre-commit configuration

**Done when:**
- Pre-commit hooks configured
- Hooks run automatically on git commit
- Formatting, linting, and tests run before commit
- Bad commits rejected automatically

**Verify:**
```bash
pre-commit run --all-files
```

**Commit:**
```
chore(git): add pre-commit hooks for quality gates

- Set up pre-commit framework
- Add formatting, linting, test hooks
- Run checks before each commit
- Prevent bad code from being committed

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-038 (Code quality)
**Design:** Phase 4 guidance - Pre-commit hooks

---

### Task 69: Create CI pipeline with GitHub Actions

**Do:**
- [ ] Create `.github/workflows/ci.yml`
- [ ] Add jobs: test-go, test-typescript, lint-go, lint-typescript
- [ ] Run on: push to main, pull requests
- [ ] Check test coverage (fail if <70%)
- [ ] Upload coverage reports to Codecov
- [ ] Add status badge to README

**Files:**
- `.github/workflows/ci.yml` - create CI workflow

**Done when:**
- CI workflow created
- All jobs run on PR and push
- Tests and linting automated
- Coverage enforced
- Status badge in README

**Verify:**
```bash
git push origin feature-branch
```

**Commit:**
```
ci: add GitHub Actions CI pipeline

- Create CI workflow for automated testing
- Add jobs for Go and TypeScript
- Run tests and linting on every PR
- Enforce test coverage threshold
- Add coverage reporting

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-038 (Code quality), CI/CD guidance
**Design:** Phase 4 guidance - CI pipeline

---

### Task 70: Add security scanning with Snyk

**Do:**
- [ ] Create `.github/workflows/security.yml`
- [ ] Add Snyk scanning for Go dependencies
- [ ] Add Snyk scanning for npm dependencies
- [ ] Run on: schedule (weekly), push to main
- [ ] Fail on high/critical vulnerabilities
- [ ] Add security badge to README

**Files:**
- `.github/workflows/security.yml` - create security workflow

**Done when:**
- Security workflow created
- Scans both Go and npm dependencies
- Runs weekly automatically
- High/critical vulnerabilities block merge
- Badge shows security status

**Verify:**
```bash
# Manually trigger workflow in GitHub UI
```

**Commit:**
```
ci: add security scanning with Snyk

- Create security scan workflow
- Scan Go and npm dependencies
- Run weekly and on push to main
- Block high/critical vulnerabilities
- Add security status badge

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-022 (Security scanning)
**Design:** Phase 4 guidance - Security scanning

---

### Task 71: Create deployment documentation

**Do:**
- [ ] Create `docs/DEPLOYMENT.md`
- [ ] Document docker-compose deployment for local/staging
- [ ] Document environment variable configuration
- [ ] Document database setup and migrations
- [ ] Document backup and restore procedures
- [ ] Add troubleshooting section
- [ ] Document monitoring and health checks

**Files:**
- `docs/DEPLOYMENT.md` - create deployment guide

**Done when:**
- Deployment guide complete
- Docker compose instructions clear
- All configuration options documented
- Backup/restore procedures included
- Troubleshooting tips provided

**Verify:**
```bash
cat docs/DEPLOYMENT.md | grep "Quick Start"
```

**Commit:**
```
docs: add comprehensive deployment documentation

- Document docker-compose deployment
- Add environment variable reference
- Document database setup
- Include backup/restore procedures
- Add troubleshooting guide

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-043 (Documentation), Deployment focus: docker-compose
**Design:** docs/ section

---

### Task 72: Create Makefile with common commands

**Do:**
- [ ] Create `Makefile` at project root
- [ ] Add targets: build, test, lint, run, docker-up, docker-down
- [ ] Add target: proto-gen (regenerate protobuf code)
- [ ] Add target: migrate (run database migrations)
- [ ] Add target: clean (remove build artifacts)
- [ ] Add help target that lists all commands

**Files:**
- `Makefile` - create with all common commands

**Done when:**
- Makefile includes all major operations
- `make help` lists all targets
- `make build` builds both services
- `make test` runs all tests
- `make docker-up` starts all services

**Verify:**
```bash
make help
```

**Commit:**
```
chore: add Makefile with common development commands

- Create Makefile for project automation
- Add build, test, lint, run targets
- Add docker-compose shortcuts
- Add proto generation and migration targets
- Include help documentation

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

**Requirements:** NFR-043 (Development setup)
**Design:** File Structure section

---

## Appendix

### Verification Commands Reference

Quick reference for all verify commands:

```bash
# Build
go build ./cmd/telegram-bot
cd galachain-service && npm run build

# Test
go test ./... -v
cd galachain-service && npm test
go test ./test/integration -v -tags=integration
go test ./test/e2e -v -tags=e2e

# Lint
golangci-lint run ./...
cd galachain-service && npm run lint

# Coverage
bash scripts/coverage.sh

# Docker
docker-compose -f deployments/docker-compose.yml up

# Proto generation
protoc --go_out=. --go-grpc_out=. proto/galachain.proto

# Database
DATABASE_URL=sqlite://./data/test.db go test ./internal/storage -v
```

### Task Dependencies

Key task dependencies (must complete before starting dependent task):

- Task 3 (proto) → Task 4 (Go gRPC gen)
- Task 3 (proto) → Task 5 (TS gRPC gen)
- Task 9 (database) → Task 10 (auth middleware)
- Task 11 (TON client) → Task 13 (balance handler)
- Task 16 (gRPC handler) → Task 17 (gRPC connection)
- Task 21 (arbitrage engine) → Task 22 (arbitrage handler)
- Task 49 (test infrastructure) → All testing tasks (50-64)

Most tasks within a phase can be done in parallel if dependencies are met.
