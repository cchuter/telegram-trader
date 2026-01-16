---
spec: telegram-trader
phase: requirements
created: 2026-01-16T14:00:00Z
---

# Requirements: Telegram Trader

## User Decisions

### Primary User Base
**Decision**: Public bot (any Telegram user)

**Implications**:
- Must implement user authentication (whitelist or registration system)
- Need per-user rate limiting to prevent abuse
- Require multi-tenant data isolation (separate user wallets, trade history)
- Must implement abuse prevention mechanisms (transaction limits, cooldowns)
- Privacy considerations for user data storage and handling
- Need clear terms of service and user consent flows
- Higher security requirements due to public exposure

### Development Priority
**Decision**: Speed of delivery - get working bot quickly, iterate based on usage

**Implications**:
- MVP-first approach with clear feature phasing
- Focus on core functionality (P0) before enhancements (P1/P2)
- Accept technical debt that can be refactored later
- Prioritize working software over perfect architecture
- Plan for iterative releases with user feedback loops
- Start with simpler implementations (e.g., manual arbitrage approval vs full automation)

### Phased Rollout Strategy

**Phase 1 (MVP - P0)**: Basic wallet connection + manual trading
- TonKeeper wallet connection via TonConnect
- Gala wallet connection (manual key input fallback)
- Balance checking for both chains
- Manual swap execution with user confirmation
- Basic error handling and user notifications

**Phase 2 (Enhancement - P1)**: Price monitoring + alerts
- Real-time price checking on ston.fi and gswap
- Price spread monitoring
- User-configurable price alerts
- Basic trade history logging

**Phase 3 (Advanced - P1/P2)**: Arbitrage automation
- Automated spread detection
- Manual arbitrage execution (user approval required)
- Position sizing with safety checks
- Full arbitrage automation (optional, based on user feedback)

## User Stories

### US-001: Wallet Connection - TonKeeper (Priority: P0)

**As a** Telegram user
**I want to** connect my TonKeeper wallet to the bot
**So that** I can trade TON and GALA tokens on ston.fi

**Acceptance Criteria:**
- [ ] AC1: Bot generates TonConnect session with QR code and deep link
- [ ] AC2: User can scan QR code on desktop or click deep link on mobile
- [ ] AC3: TonKeeper app prompts user for connection approval
- [ ] AC4: Bot receives and stores wallet address after approval
- [ ] AC5: Connection status is displayed to user in Telegram
- [ ] AC6: User can disconnect wallet via /disconnect command
- [ ] AC7: Session persists across bot restarts
- [ ] AC8: Failed connections show clear error messages

### US-002: Wallet Connection - Gala Wallet (Priority: P0)

**As a** Telegram user
**I want to** connect my Gala wallet to the bot
**So that** I can trade GALA and GTON tokens on gswap

**Acceptance Criteria:**
- [ ] AC1: Bot supports WalletConnect v2 protocol for Gala wallet
- [ ] AC2: User can alternatively provide wallet address + private key manually
- [ ] AC3: Manual credentials are encrypted before storage
- [ ] AC4: Bot validates wallet address format before accepting
- [ ] AC5: Connection status is displayed to user in Telegram
- [ ] AC6: User can disconnect wallet via /disconnect command
- [ ] AC7: Failed connections show clear error messages with troubleshooting steps

### US-003: Balance Checking (Priority: P0)

**As a** user with connected wallets
**I want to** check my token balances across both chains
**So that** I can see how much I have available to trade

**Acceptance Criteria:**
- [ ] AC1: /balance command shows TON and GALA balances from TonKeeper wallet
- [ ] AC2: /balance command shows GALA and GTON balances from Gala wallet
- [ ] AC3: Balances display with appropriate decimal precision
- [ ] AC4: USD value is shown for each token (via CoinGecko or similar)
- [ ] AC5: Response time is under 5 seconds
- [ ] AC6: Errors (API timeout, wallet not connected) show helpful messages
- [ ] AC7: User can refresh balances on demand

### US-004: Manual Token Swap - ston.fi (Priority: P0)

**As a** user with a connected TonKeeper wallet
**I want to** swap TON for GALA (or vice versa) on ston.fi
**So that** I can manually execute trades based on my analysis

**Acceptance Criteria:**
- [ ] AC1: /swap command accepts parameters: source token, destination token, amount
- [ ] AC2: Bot simulates swap and shows expected output amount, fees, and slippage
- [ ] AC3: User confirms swap via inline button before execution
- [ ] AC4: Bot executes swap via ston.fi smart contracts using tonutils-go
- [ ] AC5: User receives confirmation with transaction hash and actual output amount
- [ ] AC6: Failed swaps show specific error reason (insufficient balance, slippage too high, etc.)
- [ ] AC7: Transaction status can be checked via /orders command
- [ ] AC8: Minimum balance checks (1 TON) are enforced before swap

### US-005: Manual Token Swap - gswap (Priority: P0)

**As a** user with a connected Gala wallet
**I want to** swap GTON for GALA (or vice versa) on gswap
**So that** I can manually execute trades on GalaChain

**Acceptance Criteria:**
- [ ] AC1: /swap command supports gswap as target exchange
- [ ] AC2: Bot quotes swap via gswap API and shows expected output, fees
- [ ] AC3: User confirms swap via inline button before execution
- [ ] AC4: Bot executes swap via GalaChain TypeScript service
- [ ] AC5: User receives confirmation with transaction ID and actual output amount
- [ ] AC6: Failed swaps show specific error reason
- [ ] AC7: Swap respects gswap rate limits (20 req/10s)
- [ ] AC8: Minimum balance checks (10 GALA) are enforced before swap

### US-006: Price Checking (Priority: P1)

**As a** trader
**I want to** check current prices for TON/GALA and GTON/GALA
**So that** I can make informed trading decisions

**Acceptance Criteria:**
- [ ] AC1: /price command shows TON/GALA price on ston.fi
- [ ] AC2: /price command shows GTON/GALA price on gswap
- [ ] AC3: Prices show bid/ask spread and 24h volume
- [ ] AC4: USD equivalent is shown for each pair
- [ ] AC5: Response includes last update timestamp
- [ ] AC6: Price data is cached with 5-second TTL for performance
- [ ] AC7: Price spread percentage between exchanges is highlighted

### US-007: Price Alerts (Priority: P1)

**As a** trader
**I want to** set price alerts for specific thresholds
**So that** I'm notified when arbitrage opportunities arise

**Acceptance Criteria:**
- [ ] AC1: /alert command creates price alert with threshold (e.g., /alert 0.5% spread)
- [ ] AC2: Bot monitors prices in background and sends notification when threshold met
- [ ] AC3: User can list active alerts via /alert list
- [ ] AC4: User can delete alerts via /alert delete [id]
- [ ] AC5: Maximum 5 active alerts per user to prevent spam
- [ ] AC6: Alerts auto-expire after 24 hours
- [ ] AC7: Notifications include current price, spread, and quick action buttons

### US-008: Portfolio Overview (Priority: P1)

**As a** trader with multiple trades
**I want to** see my portfolio performance across both chains
**So that** I can track my overall profit/loss

**Acceptance Criteria:**
- [ ] AC1: /portfolio command shows total value of all holdings in USD
- [ ] AC2: Shows breakdown by token (TON, GALA, GTON) with quantities and values
- [ ] AC3: Displays unrealized P&L based on entry prices
- [ ] AC4: Shows total fees paid across all trades
- [ ] AC5: Includes simple performance metrics (% gain/loss)
- [ ] AC6: Data persists across bot restarts

### US-009: Trade History (Priority: P1)

**As a** user who has made trades
**I want to** view my past transactions
**So that** I can review my trading activity and calculate taxes

**Acceptance Criteria:**
- [ ] AC1: /orders command shows last 10 transactions by default
- [ ] AC2: Each entry shows: timestamp, type (swap/arbitrage), tokens, amounts, fees, status
- [ ] AC3: User can request more history via /orders [count] (max 100)
- [ ] AC4: Completed, pending, and failed transactions are clearly marked
- [ ] AC5: Transaction IDs are clickable links to blockchain explorers
- [ ] AC6: User can export full history as CSV via /orders export
- [ ] AC7: Trade history is append-only (JSONL format) for audit purposes

### US-010: Manual Arbitrage Execution (Priority: P1)

**As a** trader who monitors spreads
**I want to** execute arbitrage trades with bot assistance
**So that** I can profit from price differences without manual calculations

**Acceptance Criteria:**
- [ ] AC1: /arbitrage command triggers spread analysis and displays opportunity
- [ ] AC2: Bot shows which direction is profitable (ston.fi→gswap or gswap→ston.fi)
- [ ] AC3: Bot calculates exact amounts for both legs to maintain 50% position size
- [ ] AC4: Shows estimated profit after all fees and slippage
- [ ] AC5: User must explicitly approve execution via confirmation button
- [ ] AC6: Bot executes both legs simultaneously (async)
- [ ] AC7: User receives real-time status updates during execution
- [ ] AC8: Minimum balance checks ensure 1 TON and 10 GALA remain after trade
- [ ] AC9: Failed arbitrage shows which leg failed and allows rollback if needed

### US-011: Automated Arbitrage (Priority: P2)

**As a** trader who wants passive income
**I want to** enable automatic arbitrage when spreads exceed my threshold
**So that** I don't miss opportunities while away from my device

**Acceptance Criteria:**
- [ ] AC1: /arbitrage auto [threshold] enables automatic execution
- [ ] AC2: Default threshold is 0.3% (user configurable, minimum 0.2%)
- [ ] AC3: Bot monitors prices continuously and executes when conditions met
- [ ] AC4: Maximum 1 arbitrage execution per hour per user (safety limit)
- [ ] AC5: User receives notification before and after each automatic trade
- [ ] AC6: User can disable automation via /arbitrage stop
- [ ] AC7: Automation status shown in /status command
- [ ] AC8: Circuit breaker stops automation after 3 consecutive failed trades
- [ ] AC9: User can set daily loss limit to prevent runaway losses

### US-012: Help and Command Discovery (Priority: P0)

**As a** new user
**I want to** learn available commands and how to use them
**So that** I can get started quickly without external documentation

**Acceptance Criteria:**
- [ ] AC1: /help command shows list of all available commands with brief descriptions
- [ ] AC2: /help [command] shows detailed usage with examples
- [ ] AC3: /start command shows welcome message and setup instructions
- [ ] AC4: Invalid commands trigger helpful error with suggested alternatives
- [ ] AC5: Help text matches user's current state (e.g., wallet not connected shows setup help)
- [ ] AC6: Command examples use realistic values and token symbols

### US-013: User Authentication (Priority: P0)

**As a** bot operator
**I want to** control who can use the bot
**So that** I prevent abuse and unauthorized access to trading features

**Acceptance Criteria:**
- [ ] AC1: Bot maintains whitelist of authorized Telegram user IDs
- [ ] AC2: Unauthorized users receive "access denied" message with contact info
- [ ] AC3: Admin can add/remove users via configuration file or admin commands
- [ ] AC4: First-time authorized users go through onboarding flow
- [ ] AC5: User session expires after 24 hours of inactivity
- [ ] AC6: Re-authentication required after session expiry

### US-014: Security and Key Management (Priority: P0)

**As a** user entrusting my wallet to the bot
**I want to** be confident my private keys are secure
**So that** my funds are protected from theft or unauthorized access

**Acceptance Criteria:**
- [ ] AC1: Private keys are encrypted at rest using AES-256
- [ ] AC2: Encryption password is never stored with encrypted keys
- [ ] AC3: Private keys are never logged or displayed in plain text
- [ ] AC4: User can view encrypted key but not decrypt without password
- [ ] AC5: Bot warns users about security risks during wallet connection
- [ ] AC6: Session tokens rotate every 6 hours
- [ ] AC7: Failed authentication attempts are logged and rate-limited

## Functional Requirements

### Core Trading Features

**FR-001** (P0): Bot MUST support TonConnect protocol for TonKeeper wallet integration with both QR code and deep link flows

**FR-002** (P0): Bot MUST support WalletConnect v2 protocol for Gala wallet integration with fallback to manual private key input

**FR-003** (P0): Bot MUST execute token swaps on ston.fi using tonutils-go SDK and ston.fi smart contracts

**FR-004** (P0): Bot MUST execute token swaps on gswap via TypeScript GalaChain microservice using gswap SDK

**FR-005** (P0): Bot MUST simulate swaps before execution to calculate expected output, fees, and slippage

**FR-006** (P0): Bot MUST require explicit user confirmation before executing any swap or arbitrage trade

**FR-007** (P1): Bot MUST track all trades in append-only JSONL format with timestamp, user ID, type, amounts, fees, and status

**FR-008** (P1): Bot MUST calculate and display portfolio value including unrealized P&L and total fees paid

**FR-009** (P1): Bot MUST support arbitrage execution with simultaneous execution of both legs (ston.fi and gswap)

**FR-010** (P2): Bot SHOULD support automated arbitrage when enabled by user with configurable spread threshold

### Wallet Management

**FR-011** (P0): Bot MUST validate wallet addresses before accepting connection

**FR-012** (P0): Bot MUST encrypt private keys using AES-256 before storing

**FR-013** (P0): Bot MUST persist wallet connections across restarts using encrypted session storage

**FR-014** (P0): Bot MUST support wallet disconnection with secure deletion of session data

**FR-015** (P1): Bot SHOULD support multiple wallet connections per user (one per blockchain)

**FR-016** (P1): Bot SHOULD notify users when wallet sessions expire or become invalid

### Balance and Price Monitoring

**FR-017** (P0): Bot MUST fetch and display token balances from both TonKeeper and Gala wallets

**FR-018** (P0): Bot MUST show USD values for token balances using CoinGecko or equivalent price feed

**FR-019** (P1): Bot MUST check TON/GALA prices on ston.fi via swap simulation API

**FR-020** (P1): Bot MUST check GTON/GALA prices on gswap via GalaChain service

**FR-021** (P1): Bot MUST calculate and display price spread percentage between exchanges

**FR-022** (P1): Bot SHOULD cache price data with 5-second TTL to reduce API calls

**FR-023** (P1): Bot SHOULD support user-configurable price alerts with notifications

### Safety and Risk Management

**FR-024** (P0): Bot MUST enforce minimum balance checks (1 TON, 10 GALA) before executing trades

**FR-025** (P0): Bot MUST validate trade amounts are within user's available balance minus minimums

**FR-026** (P0): Bot MUST reject swaps if simulated slippage exceeds 5% (configurable)

**FR-027** (P1): Bot MUST enforce position sizing for arbitrage: 50% of relevant balance while maintaining minimums

**FR-028** (P1): Bot SHOULD implement transaction timeout (30 seconds) and show pending status

**FR-029** (P2): Bot SHOULD implement circuit breaker that pauses automation after 3 consecutive failures

**FR-030** (P2): Bot SHOULD support user-configurable daily loss limits for automated trading

### User Experience

**FR-031** (P0): Bot MUST provide clear error messages with actionable guidance for all failure cases

**FR-032** (P0): Bot MUST show transaction confirmation with hash/ID after successful trades

**FR-033** (P0): Bot MUST respond to commands within 5 seconds or show "processing" indicator

**FR-034** (P0): Bot MUST provide /help command with usage documentation for all commands

**FR-035** (P1): Bot SHOULD use inline keyboards for confirmations and multi-step flows

**FR-036** (P1): Bot SHOULD support command history and recent action quick access

**FR-037** (P1): Bot SHOULD provide /status command showing wallet connections, automation state, and recent activity

### Rate Limiting and Abuse Prevention

**FR-038** (P0): Bot MUST enforce user whitelist for access control (configurable)

**FR-039** (P0): Bot MUST implement per-user rate limiting: 10 commands per minute

**FR-040** (P1): Bot MUST respect gswap API rate limit (20 req/10s) with request queuing

**FR-041** (P1): Bot SHOULD log all user commands and bot responses for audit purposes

**FR-042** (P2): Bot SHOULD implement progressive cooldowns for users hitting rate limits repeatedly

### Multi-Chain Coordination

**FR-043** (P0): Bot MUST communicate between Go Telegram service and TypeScript GalaChain service via gRPC

**FR-044** (P0): Bot MUST handle inter-service communication failures gracefully with retries

**FR-045** (P1): Bot MUST monitor health of both services and notify user if GalaChain service unavailable

**FR-046** (P1): Bot SHOULD continue TON operations even if GalaChain service is down

**FR-047** (P2): Bot SHOULD support streaming price updates from GalaChain service via gRPC streams

### Data Persistence

**FR-048** (P0): Bot MUST persist user sessions with encrypted wallet credentials

**FR-049** (P1): Bot MUST persist trade history in JSONL format for audit and tax reporting

**FR-050** (P1): Bot SHOULD persist user preferences (alerts, automation settings, thresholds)

**FR-051** (P2): Bot SHOULD support data export in CSV format for user convenience

## Non-Functional Requirements

### Performance

**NFR-001** (P0): Bot response time for simple commands (/help, /balance) MUST be under 3 seconds 95% of the time

**NFR-002** (P0): Price check operations MUST complete within 5 seconds including external API calls

**NFR-003** (P0): Swap simulation MUST complete within 5 seconds to ensure price accuracy

**NFR-004** (P1): Arbitrage detection and execution MUST complete within 10 seconds from trigger to final confirmation

**NFR-005** (P1): GalaChain service RPC calls MUST complete within 2 seconds for local network deployment

**NFR-006** (P1): Bot SHOULD support minimum 10 concurrent users without performance degradation

**NFR-007** (P2): Background price monitoring SHOULD check ston.fi every 1 second and gswap every 3 seconds when arbitrage automation enabled

### Scalability

**NFR-008** (P1): Bot architecture MUST support horizontal scaling of GalaChain service for increased load

**NFR-009** (P1): Bot MUST handle at least 50 active users without memory leaks or resource exhaustion

**NFR-010** (P2): System SHOULD support adding new DEXs or blockchains without major architectural changes

**NFR-011** (P2): Trade history storage SHOULD support at least 100,000 transactions per user without query degradation

### Security

**NFR-012** (P0): All private keys MUST be encrypted at rest using AES-256-GCM or stronger

**NFR-013** (P0): Encryption keys MUST be stored separately from encrypted data (environment variable, secret manager)

**NFR-014** (P0): Private keys MUST NEVER be logged, displayed in plain text, or transmitted unencrypted

**NFR-015** (P0): Bot MUST use TLS 1.3 for all external API communications

**NFR-016** (P0): Session tokens MUST expire after 24 hours of inactivity

**NFR-017** (P0): Failed authentication attempts MUST be rate-limited (max 5 per hour per user)

**NFR-018** (P1): Production deployment SHOULD use HashiCorp Vault or Azure Managed HSM for key storage

**NFR-019** (P1): Bot SHOULD implement hot/cold wallet split (minimal funds in hot wallet for trading)

**NFR-020** (P1): All user commands MUST be logged with timestamp and user ID for audit trail

**NFR-021** (P2): Bot SHOULD support 2FA for high-value transactions (configurable threshold)

**NFR-022** (P2): System SHOULD implement IP whitelisting for admin commands

### Reliability

**NFR-023** (P0): Bot MUST handle API timeouts gracefully without crashing

**NFR-024** (P0): Bot MUST implement retry logic with exponential backoff for transient failures

**NFR-025** (P0): Critical operations (swap execution) MUST be idempotent to prevent duplicate transactions

**NFR-026** (P1): Bot uptime SHOULD be 99.5% or higher (excluding planned maintenance)

**NFR-027** (P1): Bot MUST implement health checks for all critical dependencies (Telegram API, ston.fi, gswap, GalaChain service)

**NFR-028** (P1): Failed transactions MUST be logged with full error context for debugging

**NFR-029** (P1): Bot SHOULD automatically restart GalaChain service on crash with max 3 restart attempts

**NFR-030** (P2): System SHOULD support graceful shutdown with pending transaction completion

### Monitoring and Observability

**NFR-031** (P0): All services MUST implement structured logging in JSON format

**NFR-032** (P0): Logs MUST include correlation IDs for tracing requests across services

**NFR-033** (P1): System MUST expose health check endpoints for monitoring tools

**NFR-034** (P1): Critical metrics SHOULD be tracked: command count, error rate, swap success rate, arbitrage P&L

**NFR-035** (P1): System SHOULD integrate with Prometheus for metrics collection

**NFR-036** (P2): System SHOULD provide Grafana dashboards for real-time monitoring

**NFR-037** (P2): System SHOULD send alerts for critical failures (service down, transaction failure rate >10%)

### Maintainability

**NFR-038** (P0): Code MUST follow language-specific style guides (gofmt for Go, Prettier for TypeScript)

**NFR-039** (P0): All public functions MUST have documentation comments

**NFR-040** (P1): Code MUST maintain minimum 70% test coverage

**NFR-041** (P1): System SHOULD use dependency injection for testability

**NFR-042** (P1): Configuration SHOULD be externalized (environment variables, config files) not hardcoded

**NFR-043** (P2): System SHOULD provide setup scripts for local development environment

### Usability

**NFR-044** (P0): Error messages MUST be user-friendly with actionable guidance, not technical stack traces

**NFR-045** (P0): All monetary amounts MUST be displayed with appropriate decimal precision (8 decimals for crypto)

**NFR-046** (P0): Command syntax MUST be intuitive and follow Telegram bot conventions

**NFR-047** (P1): Bot SHOULD provide confirmation prompts for destructive operations (disconnect wallet, delete alerts)

**NFR-048** (P1): Bot SHOULD support command abbreviations (e.g., /bal for /balance)

**NFR-049** (P2): Bot SHOULD remember user preferences between sessions

### Compliance and Legal

**NFR-050** (P0): Bot MUST display disclaimer about trading risks before first use

**NFR-051** (P1): Bot SHOULD maintain audit logs for minimum 12 months

**NFR-052** (P1): System SHOULD support data export for regulatory compliance (GDPR, tax reporting)

**NFR-053** (P2): Bot SHOULD implement terms of service acceptance flow

## Glossary

**Arbitrage**: Trading strategy that exploits price differences for the same asset across different markets to generate profit

**GALA**: Native token of GalaChain blockchain and Gala Games ecosystem

**GTON**: Wrapped version of TON token that exists on GalaChain for cross-chain trading

**gswap**: Decentralized exchange (DEX) on GalaChain for swapping tokens, particularly GALA and GTON

**Slippage**: Difference between expected price and actual execution price, caused by market movement or low liquidity

**Spread**: Price difference between two markets for the same asset, expressed as percentage

**ston.fi**: Decentralized exchange (DEX) on TON blockchain for swapping TON and other tokens including GALA

**TON**: The Open Network blockchain and its native cryptocurrency

**TonConnect**: Standard protocol for connecting Telegram bots and dApps to TON wallets like TonKeeper

**TonKeeper**: Popular mobile wallet for TON blockchain, supports iOS and Android

**WalletConnect**: Open protocol for connecting cryptocurrency wallets to dApps across multiple blockchains

**Hot Wallet**: Cryptocurrency wallet connected to the internet and actively used for trading

**Cold Wallet**: Cryptocurrency wallet kept offline for secure long-term storage

**gRPC**: High-performance RPC framework using Protocol Buffers for service-to-service communication

**JSONL**: JSON Lines format where each line is a valid JSON object, used for append-only logging

**P0/P1/P2**: Priority levels - P0 (must-have for MVP), P1 (should-have for complete product), P2 (nice-to-have enhancement)

**Basis Points (bps)**: Unit for measuring percentages, where 100 bps = 1%

**Circuit Breaker**: Safety mechanism that automatically stops operations after repeated failures to prevent cascading issues

**HSM**: Hardware Security Module - specialized hardware for secure key storage and cryptographic operations

## Out of Scope

**Multi-DEX Support**: Support for additional DEXs beyond ston.fi and gswap is NOT included in v1
- **Reason**: MVP focuses on single pair arbitrage, additional DEXs add significant complexity

**Advanced Trading Features**: Limit orders, stop-loss orders, trailing stops are NOT included
- **Reason**: Complexity doesn't align with "speed of delivery" priority; manual execution sufficient for MVP

**Mobile Application**: Dedicated iOS/Android app is NOT included
- **Reason**: Telegram bot provides mobile access; dedicated app requires separate development effort

**Web Dashboard**: Browser-based dashboard for trade monitoring is NOT included
- **Reason**: Telegram provides sufficient UI for MVP; web dashboard can be future enhancement

**Social Trading**: Copy trading, signal sharing, leaderboards are NOT included
- **Reason**: Not core to arbitrage functionality; adds moderation and liability complexity

**Backtesting**: Historical data analysis and strategy backtesting tools are NOT included
- **Reason**: Focus on live trading for MVP; backtesting requires historical data infrastructure

**Advanced Analytics**: Machine learning price prediction, sentiment analysis are NOT included
- **Reason**: Significant R&D effort; simple spread detection sufficient for initial version

**Multi-User Collaboration**: Shared wallets or collaborative trading features are NOT included
- **Reason**: Security complexity and unclear use case for arbitrage bot

**Fiat On/Off Ramps**: Integration with fiat payment processors is NOT included
- **Reason**: Regulatory complexity; users expected to acquire crypto elsewhere

**Tax Reporting**: Automated tax form generation (1099, etc.) is NOT included
- **Reason**: Varies by jurisdiction; CSV export sufficient for users to provide to accountants

**Multi-Language Support**: Interface is English-only in v1
- **Reason**: Internationalization adds complexity; can be added based on user demand

**Cross-Chain Bridges**: Native bridging between TON and GalaChain is NOT included
- **Reason**: Relies on third-party bridges with their own risks and complexity

**Liquidity Providing**: Adding liquidity to pools or earning LP rewards is NOT included
- **Reason**: Different use case from arbitrage trading; adds DEX-specific complexity

**NFT Trading**: Support for NFT swaps or marketplace integration is NOT included
- **Reason**: Fundamentally different asset type requiring separate architecture

**Margin/Leverage Trading**: Borrowing funds to increase position size is NOT included
- **Reason**: Extreme risk for public bot; requires sophisticated risk management

## Dependencies

### External Services

**Telegram Bot API** (Critical)
- Purpose: Core interface for user interaction
- Rate Limits: 30 messages per second per bot
- Failure Impact: Complete service outage
- Mitigation: Implement message queuing, cache bot token securely

**ston.fi API** (Critical)
- Purpose: Price data, swap simulation, pool information for TON/GALA trading
- Rate Limits: None currently
- Failure Impact: Cannot execute TON-side swaps or check prices
- Mitigation: Implement retry logic, cache price data briefly, graceful degradation

**TON Blockchain Network** (Critical)
- Purpose: Execute transactions, read wallet balances
- Failure Impact: Cannot interact with TON blockchain
- Mitigation: Use multiple RPC endpoints, implement fallback nodes

**GalaSwap API** (Critical)
- Purpose: Execute swaps, check balances, authenticate transactions on GalaChain
- Rate Limits: 20 requests per 10 seconds globally, 4/min for wallet creation
- Failure Impact: Cannot execute GALA-side swaps
- Mitigation: Request queuing with rate limiter, cache balances, retry with backoff

**GalaChain Network** (Critical)
- Purpose: Submit transactions, read state from GalaChain blockchain
- Failure Impact: Cannot interact with GalaChain
- Mitigation: Monitor network status, implement transaction timeout and resubmit

**TonConnect Bridge** (High Priority)
- Purpose: Facilitate wallet connection protocol for TonKeeper
- URL: https://bridge.tonapi.io/bridge
- Failure Impact: Cannot connect new TonKeeper wallets (existing sessions work)
- Mitigation: Implement connection retry, support manual wallet address input as fallback

**WalletConnect Bridge** (Medium Priority)
- Purpose: Facilitate wallet connection for Gala wallet
- Failure Impact: Cannot connect new Gala wallets via WalletConnect
- Mitigation: Support manual private key input as primary method

**CoinGecko API** (Low Priority)
- Purpose: Fetch USD prices for portfolio valuation
- Rate Limits: 10-50 calls/min (free tier)
- Failure Impact: Cannot show USD values, portfolio still functions
- Mitigation: Cache prices aggressively (5 min TTL), graceful degradation to crypto-only display

### Internal Components

**Go Telegram Bot Service** (Critical)
- Purpose: Main bot process, command routing, TON blockchain integration, user session management
- Language: Go
- Dependencies: go-telegram/bot, tonutils-go
- Deployment: Single binary, containerized

**TypeScript GalaChain Service** (Critical)
- Purpose: GalaChain operations, gswap integration, wallet signing
- Language: TypeScript/Node.js
- Dependencies: @gala-chain/gswap-sdk, gRPC Node.js libraries
- Deployment: Node.js container, separate from bot service

**gRPC Communication Layer** (Critical)
- Purpose: Inter-service communication between Go bot and TypeScript service
- Protocol: gRPC with Protocol Buffers
- Failure Impact: Cannot execute GalaChain operations
- Mitigation: Health checks, automatic service restart, circuit breaker pattern

**Database/Storage** (Medium Priority)
- Purpose: Persist user sessions, trade history, preferences
- Options: SQLite (simple), PostgreSQL (production)
- Failure Impact: Cannot persist state across restarts, lose trade history
- Mitigation: Regular backups, append-only JSONL for critical trade data

**Secret Management** (High Priority)
- Purpose: Store encryption keys, API credentials, encrypted private keys
- Development: Environment variables
- Production: HashiCorp Vault or Azure Key Vault
- Failure Impact: Cannot decrypt user wallets or authenticate to services
- Mitigation: Redundant secret storage, regular key rotation procedures

## Success Criteria

**User Adoption**
- Minimum 5 active users within first month (beta testers)
- At least 10 successful arbitrage trades executed (manual or automated)
- User retention: 60% of users execute at least 2 trades

**Technical Performance**
- Bot uptime: 99% or higher during beta period
- Average command response time: <3 seconds
- Swap success rate: >95% (excluding user-cancelled transactions)
- Zero critical security incidents (private key exposure, unauthorized transactions)

**Functional Completeness**
- All P0 user stories implemented and tested
- TonKeeper wallet connection working on iOS
- Gala wallet connection working (manual or WalletConnect)
- Successful TON/GALA swaps on ston.fi
- Successful GTON/GALA swaps on gswap
- Manual arbitrage execution with profit calculation

**Financial Performance**
- At least 1 profitable arbitrage trade demonstrating end-to-end flow
- Fee calculations accurate within 1% of actual costs
- No trades executed that violate minimum balance requirements
- Trade history and P&L calculations match blockchain records

**Documentation and Usability**
- /help command provides clear usage instructions
- Setup process documented and tested by non-technical user
- All error messages provide actionable guidance
- User can complete wallet connection within 5 minutes

**Code Quality**
- Test coverage: >70% for core trading logic
- No high or critical security vulnerabilities (per Snyk/Dependabot scan)
- Code review completed for all P0 features
- CI/CD pipeline successfully deploys to test environment

## Risks

### Technical Risks

#### Cross-Chain Timing Risk
**Impact**: High
**Likelihood**: High
**Description**: Arbitrage spreads disappear in seconds; if one leg executes but other fails, user loses money
**Mitigation**:
- Implement pre-flight swap simulation on both chains before execution
- Set higher minimum spread threshold (0.3% vs 0.1%) to account for timing delays
- Use transaction timeouts and automatic rollback on single-leg failure
- Start with manual arbitrage requiring user approval before automation
- Log all arbitrage attempts with timing data to tune thresholds

#### Slippage Exceeds Profit Margin
**Impact**: High
**Likelihood**: Medium
**Description**: Price impact from trade size can eliminate arbitrage profit or cause net loss
**Mitigation**:
- Always check simulated slippage before execution
- Implement dynamic position sizing based on liquidity depth
- Set hard slippage limit (5% max) that triggers transaction cancellation
- Warn user if order book appears thin
- Track actual vs expected slippage to improve models

#### GalaChain Rate Limiting
**Impact**: Medium
**Likelihood**: High
**Description**: 20 req/10s limit on gswap API can delay price updates and miss arbitrage windows
**Mitigation**:
- Implement request queue with token bucket rate limiter
- Poll ston.fi more frequently (1s) and gswap less frequently (3s)
- Cache gswap prices with short TTL (2-3 seconds)
- Monitor rate limit headers and adjust polling frequency dynamically
- Consider WebSocket connections if available for streaming price data

#### Inter-Service Communication Latency
**Impact**: Medium
**Likelihood**: Medium
**Description**: Go ↔ TypeScript gRPC calls add latency that can turn profitable arbitrage unprofitable
**Mitigation**:
- Deploy both services on same machine to minimize network latency
- Use gRPC for performance (binary protocol vs JSON)
- Implement streaming RPCs for price updates to reduce connection overhead
- Add comprehensive logging at service boundaries to measure actual latency
- Set up health checks and automatic restart on communication failures

#### Wallet Connection UX Complexity
**Impact**: Medium
**Likelihood**: Medium
**Description**: TonConnect is well-documented, but Gala wallet integration less proven; poor UX reduces adoption
**Mitigation**:
- Implement TonConnect first (proven protocol with good docs)
- For Gala wallet, support both WalletConnect and manual key input as fallback
- Provide step-by-step instructions with screenshots in /help
- Test extensively on target platform (iPhone)
- Gather user feedback during beta and iterate on flow

### Security Risks

#### Private Key Compromise
**Impact**: Critical
**Likelihood**: Low
**Description**: If bot's storage or memory is compromised, attacker gains access to user funds
**Mitigation**:
- Encrypt all private keys at rest using AES-256-GCM
- Store encryption keys in separate secret manager (Vault, Azure Key Vault)
- Never log private keys or encryption passwords
- Implement hot/cold wallet split (minimal funds in hot wallet)
- For production, use HSM (Hardware Security Module) for key operations
- Regular security audits of key handling code

#### Malicious Transaction Injection
**Impact**: High
**Likelihood**: Low
**Description**: Attacker exploits vulnerability to sign unauthorized transactions
**Mitigation**:
- Implement strict transaction amount limits
- Require user confirmation for all trades via Telegram callback
- Log all transactions before execution with user ID and timestamp
- Whitelist allowed token contract addresses
- Add manual approval mode for testing new features
- Implement anomaly detection for unusual trade patterns

#### Telegram Bot Token Exposure
**Impact**: High
**Likelihood**: Low
**Description**: Bot token leak allows attacker to impersonate bot and phish users
**Mitigation**:
- Never commit bot token to git repository
- Store token in environment variable or secret manager
- Rotate token regularly (monthly)
- Implement IP whitelisting for bot API access if possible
- Monitor bot API usage for suspicious patterns
- Revoke and rotate immediately if exposure suspected

#### Public Bot Abuse
**Impact**: Medium
**Likelihood**: High
**Description**: Open access allows spam, DOS attacks, resource exhaustion
**Mitigation**:
- Implement user whitelist (Telegram user IDs) for access control
- Rate limiting: 10 commands per minute per user
- Progressive cooldowns for users repeatedly hitting limits
- Log all commands for audit trail and abuse detection
- Implement CAPTCHA or verification for first-time users
- Admin commands for banning abusive users

### Financial Risks

#### Price Volatility During Execution
**Impact**: High
**Likelihood**: Medium
**Description**: Crypto prices can swing dramatically in seconds, turning profit into loss mid-execution
**Mitigation**:
- Set maximum execution time window (10 seconds)
- Implement stop-loss mechanism that cancels if price moves unfavorably
- Use limit orders where supported vs market orders
- Start with small trade sizes during beta
- Monitor historical volatility and adjust spread thresholds accordingly

#### Fee Accumulation Eliminates Profit
**Impact**: Medium
**Likelihood**: High
**Description**: Network fees (gas) + swap fees can exceed thin arbitrage margins
**Mitigation**:
- Always calculate fees before execution using fee estimation endpoints
- Set minimum profit threshold above total fees (e.g., 0.3% vs 0.1%)
- Track cumulative fees in P&L calculations and display to user
- Consider fee rebates or reduced fees for high-volume traders
- Optimize gas usage (batch transactions where possible)

#### Insufficient Balance After Trade
**Impact**: Medium
**Likelihood**: Medium
**Description**: 50% trades could drain account below minimums, preventing future trades
**Mitigation**:
- Enforce minimum balance checks (10 GALA, 1 TON) BEFORE trade execution
- Implement safety margin (15 GALA, 1.5 TON) above minimums
- Alert user when balances approach minimums (warning threshold)
- Provide manual override for emergency withdrawals
- Add balance projection to arbitrage preview ("After trade: X GALA remaining")

### Operational Risks

#### Service Downtime
**Impact**: Medium
**Likelihood**: Medium
**Description**: ston.fi, gswap, or internal service outages cause missed opportunities or incomplete trades
**Mitigation**:
- Implement comprehensive health checks for all dependencies
- Automatic service restart on failure (max 3 attempts)
- Graceful degradation (disable arbitrage if one chain unavailable, but keep other features working)
- Status monitoring with alerts (PagerDuty, SMS, Telegram)
- Display service status in bot via /status command

#### Network Congestion
**Impact**: Medium
**Likelihood**: Medium
**Description**: TON or GalaChain network congestion delays transactions, causing failures or stuck funds
**Mitigation**:
- Monitor pending transaction status and show to user
- Implement transaction timeout (30s) and resubmit logic
- Allow manual transaction cancellation if stuck
- Adjust gas/fee settings dynamically based on network conditions
- Use multiple RPC endpoints with automatic fallback

#### Data Loss
**Impact**: Medium
**Likelihood**: Low
**Description**: Server failure, corruption, or accidental deletion loses trade history or user sessions
**Mitigation**:
- Append-only JSONL trade log (never modified, only appended)
- Regular automated backups of database (daily)
- Store backups in separate location (cloud storage)
- Test restore procedures quarterly
- Implement write-ahead logging for critical state changes

### Compliance and Regulatory Risks

#### Automated Trading Regulations
**Impact**: Low
**Likelihood**: Low
**Description**: Jurisdictions may have unclear or restrictive rules about automated trading bots
**Mitigation**:
- Display clear disclaimer about trading risks before first use
- Recommend users consult local regulations before automated trading
- Implement opt-in for arbitrage feature (not enabled by default)
- Keep detailed audit logs for 12+ months
- Consider geographic restrictions if specific countries prohibit automated trading
- Include terms of service acknowledgment

#### Exchange Terms of Service Violations
**Impact**: Low
**Likelihood**: Low
**Description**: ston.fi or gswap ToS may prohibit automated trading, leading to account bans
**Mitigation**:
- Review both exchanges' ToS for bot usage policies before launch
- Implement reasonable rate limiting (be good API citizen)
- Avoid aggressive practices (flash loans, MEV extraction, sandwich attacks)
- Contact exchange support to disclose bot usage and confirm compliance
- Monitor for ToS changes and adapt accordingly
