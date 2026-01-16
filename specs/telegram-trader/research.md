---
spec: telegram-trader
phase: research
created: 2026-01-16T13:42:19Z
---

# Research: Telegram Trader

## Executive Summary

The Telegram Trading Bot for GalaChain and TON blockchain integration is **technically feasible** with a hybrid Go+TypeScript architecture. Key findings:

- **Go Telegram Bot Frameworks**: Multiple mature options available (go-telegram/bot, telebot, telego)
- **TON Blockchain**: Strong Go SDK support via tonutils-go and tongo
- **GalaChain/gswap**: TypeScript-only SDK, requiring inter-process communication pattern
- **Wallet Integration**: TonConnect protocol for TonKeeper, WalletConnect for Gala wallet (both support deep linking)
- **Reference Implementation**: cchuter/gswap-bot provides working TypeScript implementation patterns
- **Primary Risk**: Cross-chain arbitrage timing and slippage management in volatile conditions

**Recommended Architecture**: Go microservice for Telegram bot + TypeScript microservice for GalaChain operations, communicating via gRPC or REST API.

## Technology Stack

### Telegram Bot Framework

**Recommended: go-telegram/bot** (https://github.com/go-telegram/bot)

- Actively maintained (as of 2026)
- Clean architecture and code structure
- Good community feedback from developers migrating from older libraries
- Full Telegram Bot API coverage

**Alternatives:**

- **telebot** (https://github.com/tucnak/telebot) - Mature, highload-ready, excellent API design for command routing and callbacks
- **telego** (https://github.com/mymmrac/telego) - Full 1:1 API implementation, uses fasthttp for performance

**NOT Recommended:**

- go-telegram-bot-api - Last updated 2021, maintenance concerns

### Blockchain Integration

#### TON (ston.fi)

**Primary SDK: tonutils-go** (https://github.com/xssnick/tonutils-go)

- Native Go implementation of ADNL and lite protocol
- Implements all main TON protocols: ADNL, DHT, RLDP, Overlays
- Concurrent-safe for high workload scenarios
- Active maintenance

**Alternative: tongo** (https://github.com/tonkeeper/tongo)

- Official SDK from Tonkeeper team
- Well-integrated with TON ecosystem tooling

**Installation:**

```bash
go get github.com/xssnick/tonutils-go@latest
```

**STON.fi DEX API:**

- Base URL: https://api.ston.fi
- **Rate Limits**: Currently NO rate limits on DEX API
- **Key Endpoints**:
  - `POST /v1/swap/simulate` - Simulate swap before execution (calculate output, fees, gas)
  - `GET /v1/swap/status` - Check swap operation status
  - `POST /v1/reverse_swap/simulate` - Calculate required input for desired output
- Interactive API docs: https://api.ston.fi/swagger-ui/
- Full documentation: https://docs.ston.fi/developer-section/dex/api/reference

#### GalaChain (gswap)

**SDK: TypeScript only** - No native Go implementation

- Must use TypeScript/Node.js for GalaChain operations
- Reference implementation: https://github.com/cchuter/gswap-bot

**GalaSwap API:**

- Base URL: https://api-galaswap.gala.com
- **Rate Limits**: 20 requests per 10 seconds (global), wallet creation limited to 4 requests/min
- **Key Endpoints**:
  - `CreateHeadlessWallet` - Wallet creation
  - `FetchBalances` - Check account balances
  - `TransferToken` - Transfer tokens between accounts
  - Token swap endpoints via GalaSwap API
- API docs: https://galaswap.gala.com/info/api.html
- Explorer API: https://explorer-api.galachain.com/docs/
- galachain docs: https://docs.galachain.com/v2.7.1/
- gswap npm: https://www.npmjs.com/package/@gala-chain/gswap-sdk

**Authentication Requirements:**

- Gala account (create at games.gala.com)
- GalaChain wallet address, private key, and public key

### Architecture Pattern

**Recommended: Microservices with gRPC/REST Communication**

```
┌─────────────────────────────────────────────┐
│         Telegram Bot (Go)                   │
│  - User interface & command routing         │
│  - Session management                       │
│  - TonKeeper wallet integration (TonConnect)│
│  - TON blockchain operations (tonutils-go)  │
│  - STON.fi DEX integration                  │
└─────────────────┬───────────────────────────┘
                  │ gRPC/REST API
┌─────────────────▼───────────────────────────┐
│    GalaChain Service (TypeScript/Node.js)   │
│  - GalaChain SDK operations                 │
│  - gswap integration                        │
│  - Gala wallet connection (WalletConnect)   │
│  - Transaction signing & submission         │
└─────────────────────────────────────────────┘
```

**Why This Pattern:**

1. **Language Optimization**: Use Go for Telegram bot (performance, concurrency) and TypeScript for GalaChain (SDK requirement)
2. **Separation of Concerns**: Each service handles one blockchain ecosystem
3. **Independent Scaling**: Can scale TON and GALA services separately based on load
4. **Easier Testing**: Mock gRPC/REST interfaces for unit tests
5. **Reference Alignment**: Leverages existing gswap-bot TypeScript codebase

**Inter-Process Communication Options:**

- **gRPC** (Recommended): Type-safe, efficient, supports streaming for price updates
- **REST API**: Simpler, easier debugging, HTTP/JSON standard
- **Subprocess**: Possible but less robust for production (no example found in research)

**Alternative: Monorepo with Shared Types**

- Use TypeScript for entire stack with Telegram bot in TS
- Simpler deployment but loses Go performance advantages
- NOT recommended given user preference for Go Telegram API

## API Analysis

### ston.fi

**Capabilities:**

- Price data retrieval via swap simulation
- Token swap execution (requires smart contract interaction via TON SDK)
- Swap status monitoring
- Reverse swap calculations (input from desired output)
- Pool data and analytics

**Rate Limits:**

- **None currently** - Makes it ideal for high-frequency price checking for arbitrage

**Documentation Quality:**

- Excellent - Swagger UI, Redoc, comprehensive examples
- Current version: v1.168.0

**Integration Approach:**

1. Use tonutils-go for wallet operations and smart contract calls
2. Use ston.fi REST API for price data and swap simulation
3. Execute swaps by calling STON.fi router smart contracts via TON SDK

**Token Address (from requirements):**

- GALA on TON: `EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV`
- Reference: https://app.ston.fi/swap?chartVisible=false&ft=TON&tt=EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV

### gswap

**Capabilities (from gswap-bot analysis):**

- Full swap execution via GSwap SDK
- Pool discovery via Gala Explore API
- Real-time pool monitoring
- Fee credit minting (`AuthorizeFee` endpoint)
- Fee estimation (`RequestTokenSwap/fee` endpoint)
- Swap submission (`RequestTokenSwap`)
- Swap cancellation (`TerminateTokenSwap`)
- USD pricing integration (CoinGecko)

**Rate Limits:**

- **Global**: 20 requests per 10 seconds
- **Wallet Creation**: 4 requests per minute
- **Implication**: Must implement request queuing/throttling for arbitrage bot

**SDK Pattern (from gswap-bot):**

```typescript
// Signature-based authentication using wallet private keys
// GALA ↔ GTON token pair (per requirements)
// Full-position execution support
// Append-only JSONL transaction logging (swap-history.log)
```

**Token Address (from requirements):**

- GTON on GalaChain: `2954bd1e38c2dff11a2ee8798e2f59b99c206e64bef7bdf5c4ac9a91315e0d34`
- Reference: https://swap.gala.com/explore-balance/2954bd1e38c2dff11a2ee8798e2f59b99c206e64bef7bdf5c4ac9a91315e0d34

**Configuration (from gswap-bot .env.example):**

```
WALLET_ADDRESS - Active Gala wallet
PRIVATE_KEY or PRIVATE_KEY_ENCRYPTED - Signing credentials
GALA_TUI_REFRESH_MS - Dashboard update interval
GALA_WBTC_SLIPPAGE_BPS - Price tolerance (basis points)
GALA_WBTC_FEE - Pool fee tier override
```

## Wallet Integration

### TonKeeper

**Connection Mechanism: TonConnect Protocol**

- TON's equivalent to WalletConnect
- Universal Link: `https://app.tonkeeper.com/ton-connect`
- Deep Link: `tonkeeper-tc://`
- Bridge URL: `https://bridge.tonapi.io/bridge`

**Integration for Telegram Bot:**

1. Generate TonConnect session request
2. Present QR code or deep link to user
3. User approves in TonKeeper app (iPhone)
4. Bot receives wallet address + session key via bridge
5. Subsequent transactions signed via TonConnect protocol

**Session Protocol:**

- All communications except initial request are encrypted
- Persistent session for multiple transactions
- Supports transaction signing without re-authentication

**Documentation:**

- TonConnect SDK: https://ton-connect.github.io/sdk/modules/_tonconnect_sdk.html
- Deep linking guide: https://docs.ton.org/ecosystem/wallet-apps/deep-links
- Integration manual: https://docs.ton.org/develop/dapps/ton-connect/integration

**Mobile-Specific:**

- Universal link works for QR code display (desktop → mobile)
- Deep link for in-Telegram flow (mobile only)
- Both iOS and Android supported

### Gala Wallet

**Connection Mechanism: WalletConnect + Custom Integration**

- Gala wallet supports WalletConnect improvements (confirmed in app updates)
- GalaChain Connect library provides dApp connectivity
- Deep linking support for iOS (custom URL scheme)

**Integration Approach:**

1. **Option A - WalletConnect**: Standard WalletConnect v2 protocol

   - Generate WalletConnect session URI
   - User scans QR or opens deep link
   - Gala wallet app handles approval

2. **Option B - GalaChain Connect**: Custom library (similar to ethers.js)
   - Allows MetaMask and other Web3 wallets to connect to GalaChain
   - May not be optimal for native Gala wallet

**Challenges:**

- Less documentation than TonConnect
- iOS deep linking caveats (single app per URL scheme)
- May require testing to determine best approach

**Recommendation:**

- Start with WalletConnect standard approach
- Fall back to direct wallet address + private key input for power users
- Reference gswap-bot pattern: uses wallet address + private key directly

**Documentation:**

- Gala wallet support: https://support.gala.com/hc/en-us/articles/22441079637275-Use-a-Web3-Wallet-on-a-Mobile-Device
- WalletConnect mobile linking: https://docs.walletconnect.network/wallet-sdk/ios/mobile-linking
- GalaChain Connect: https://connect.gala.com/info/api.html

## Arbitrage Implementation

**Goal**: Monitor 0.1% price differences between TON/GALA on ston.fi and GTON/GALA on gswap, execute 50% trades while maintaining minimums (10 GALA or 1 TON).

### Technical Approach

**1. Price Monitoring Loop**

```
Every N seconds (configurable):
├─ Fetch TON/GALA price from ston.fi (simulate swap)
├─ Fetch GTON/GALA price from gswap (via TS service)
├─ Calculate spread: |price_A - price_B| / min(price_A, price_B)
└─ If spread ≥ 0.1%: Trigger arbitrage execution
```

**2. Spread Detection**

- Use swap simulation endpoints (no cost, no rate limit on ston.fi)
- Account for fees in spread calculation
- Calculate net profit after gas + swap fees

**3. Trade Execution**

```
If ston.fi price < gswap price:
  ├─ Buy GALA with TON on ston.fi
  └─ Sell GALA for GTON on gswap

If gswap price < ston.fi price:
  ├─ Buy GALA with GTON on gswap
  └─ Sell GALA for TON on ston.fi
```

**4. Position Sizing**

- Calculate 50% of relevant asset (TON or GALA)
- Check minimums: Must leave ≥10 GALA or ≥1 TON
- If insufficient balance, use maximum safe amount

**5. Execution Coordination**

- Execute both legs simultaneously (async)
- Use transaction IDs to track completion
- Implement rollback logic if one leg fails

### Timing Considerations

**Critical Success Factors:**

- **Speed**: Arbitrage opportunities last seconds to milliseconds (per research)
- **Latency**: Even slight delays can turn profitable trades into losses
- **Slippage**: Check order book depth before execution

**Optimization Strategies:**

1. **Pre-simulation**: Always call swap simulation before real execution
2. **Order Book Analysis**: Check liquidity depth (thick book = less slippage)
3. **Smart Position Sizing**: For liquid exchanges, $50k trade ≈ 0.05% slippage
4. **Minimum Profit Threshold**: Set higher than 0.1% to account for slippage
5. **Guardrails**: Cancel if slippage or fees breach safety limits

### Research-Based Best Practices

**From 2026 Arbitrage Research:**

- Net profit margins are razor-thin: 0.01-0.1% typical
- Advanced systems analyze order book health + historical data to predict slippage
- AI systems size trades based on liquidity conditions
- Only traders with minimal fees (maker rebates) and fast execution can profit consistently

**Risk Mitigation:**

1. Start with conservative minimums (higher than 0.1%)
2. Implement dry-run mode for testing
3. Log all trades to JSONL (like gswap-bot pattern)
4. Set maximum trade size limits
5. Implement circuit breaker for excessive losses

### Rate Limit Management

**ston.fi**: No limits - can poll aggressively
**gswap**: 20 req/10s - must throttle price checks

- Recommendation: Poll ston.fi every 1s, gswap every 3s
- Stagger requests to stay under 20/10s limit

## Existing Code Analysis

**Repository State**: Fresh repository with minimal code

- `README.md`: Basic project description
- `requirements.md`: Detailed feature requirements (source of truth)
- `.gitignore`: Standard spec tracking configuration

**No Existing Trading Code**: Greenfield project

**Reference Repository: cchuter/gswap-bot** (TypeScript)

**Key Files from gswap-bot to Reference:**

1. **`gswap.ts`**: Main swap execution engine

   - Terminal UI dashboard
   - PnL calculations
   - Swap execution logic
   - Pattern to adapt for arbitrage execution

2. **`pool-monitor.ts`**: Pool analysis & monitoring

   - Real-time price tracking
   - Can snapshot or continuously monitor
   - Useful pattern for arbitrage price monitoring

3. **`decrypt.ts`**: Key decryption utility

   - Converts encrypted Gala keys using transfer codes
   - Security pattern for key management

4. **`scripts/quote-swap.ts`**: Price quotation

   - Fetches live pricing
   - CoinGecko USD conversion
   - Direct pattern for price checking

5. **`scripts/request-token-swap.ts`**: Swap submission
   - Signs and submits authenticated requests
   - Pattern for GalaChain transaction signing

**Key Patterns to Adopt:**

- JSONL append-only logging (`swap-history.log`)
- Environment variable configuration (`.env` pattern)
- Elliptic cryptography for signing
- Fee credit minting workflow
- Error handling and retry logic

**Technology Alignment:**

- gswap-bot uses TypeScript → directly usable in GalaChain microservice
- Can extract common utilities and API interaction code
- Should maintain similar logging and monitoring patterns

## Risk Assessment

### Technical Risks

**1. Cross-Chain Timing Risk** (HIGH)

- **Issue**: Arbitrage spreads disappear in seconds/milliseconds
- **Impact**: Unprofitable trades if one leg executes but other fails
- **Mitigation**:
  - Implement pre-flight checks (swap simulation)
  - Set higher minimum spread threshold (e.g., 0.3% vs 0.1%)
  - Use transaction timeouts and rollback logic
  - Start with manual execution, automate after proven

**2. Slippage Risk** (HIGH)

- **Issue**: Price impact from trade size can exceed profit margin
- **Impact**: Net loss even with apparent arbitrage opportunity
- **Mitigation**:
  - Always check order book depth before execution
  - Implement dynamic position sizing based on liquidity
  - Set slippage tolerance limits (e.g., 0.5% max)
  - Use swap simulation endpoints to predict slippage

**3. GalaChain Rate Limiting** (MEDIUM)

- **Issue**: 20 req/10s limit on gswap API
- **Impact**: Delayed price updates, missed arbitrage opportunities
- **Mitigation**:
  - Implement request queue with rate limiter
  - Poll ston.fi more frequently (no limits), gswap less frequently
  - Cache results with short TTL
  - Consider WebSocket connections if available

**4. Inter-Process Communication Complexity** (MEDIUM)

- **Issue**: Go ↔ TypeScript communication adds latency and failure points
- **Impact**: Slower execution, potential for communication failures
- **Mitigation**:
  - Use gRPC for performance (binary protocol, streaming)
  - Implement health checks and automatic restart
  - Add comprehensive logging at service boundaries
  - Consider deploying services on same machine to reduce network latency

**5. Wallet Connection UX** (MEDIUM)

- **Issue**: TonConnect well-documented, Gala wallet less so
- **Impact**: Poor user experience, lower adoption
- **Mitigation**:
  - Implement TonConnect first (proven, documented)
  - For Gala wallet, support both WalletConnect and manual key input
  - Provide clear setup instructions with screenshots
  - Test thoroughly on iPhone (target platform)

### Security Risks

**1. Private Key Management** (CRITICAL)

- **Issue**: Bot needs access to private keys for signing transactions
- **Impact**: Compromised keys = total loss of funds
- **Mitigation**:
  - Use encrypted private keys (like gswap-bot `PRIVATE_KEY_ENCRYPTED`)
  - Implement HSM support for production (separate signing service)
  - Hot/cold wallet split (small amount in hot wallet for trading)
  - Environment variable isolation (never log keys)
  - Consider hardware security modules (Azure Managed HSM, HashiCorp Vault)

**2. Transaction Signing Security** (HIGH)

- **Issue**: Malicious transactions could drain wallets
- **Mitigation**:
  - Implement transaction amount limits
  - Require confirmation for large trades
  - Log all transactions before execution
  - Implement whitelist of allowed token addresses
  - Add manual approval mode for testing

**3. API Credential Exposure** (HIGH)

- **Issue**: Telegram bot token, API keys in environment
- **Impact**: Unauthorized access, impersonation, fund theft
- **Mitigation**:
  - Never commit `.env` files
  - Use secret management service (HashiCorp Vault, AWS Secrets Manager)
  - Rotate credentials regularly
  - Implement IP whitelisting where possible

**4. Telegram Bot Security** (MEDIUM)

- **Issue**: Public bot accessible to anyone
- **Impact**: Spam, abuse, unauthorized trading attempts
- **Mitigation**:
  - Implement user authentication (whitelist of Telegram user IDs)
  - Rate limiting per user
  - Session management with timeouts
  - Log all commands for audit trail

### Financial Risks

**1. Price Volatility** (HIGH)

- **Issue**: Crypto prices can move dramatically mid-execution
- **Impact**: Arbitrage profit disappears or turns into loss
- **Mitigation**:
  - Set maximum execution time windows
  - Implement stop-loss mechanisms
  - Use limit orders where possible vs market orders
  - Start with small trade sizes

**2. Fee Accumulation** (MEDIUM)

- **Issue**: Network fees + swap fees eat into thin margins
- **Impact**: Unprofitable arbitrage even with price spreads
- **Mitigation**:
  - Calculate fees before execution (use fee estimation endpoints)
  - Set minimum profit threshold above total fees
  - Track cumulative fees in PnL calculations
  - Consider fee rebates on high-volume exchanges

**3. Minimum Balance Risk** (MEDIUM)

- **Issue**: 50% trades could leave account below minimums
- **Impact**: Unable to execute future trades, locked funds
- **Mitigation**:
  - Enforce minimum balance checks (10 GALA, 1 TON) before trades
  - Implement safety margin above minimums
  - Alert user when balances approach minimums
  - Provide manual override for emergency withdrawals

### Operational Risks

**1. Service Downtime** (MEDIUM)

- **Issue**: ston.fi, gswap, or bot service outages
- **Impact**: Missed arbitrage opportunities, incomplete transactions
- **Mitigation**:
  - Implement comprehensive health checks
  - Automatic service restart on failure
  - Graceful degradation (disable arbitrage if one service down)
  - Status monitoring and alerting

**2. Network Congestion** (MEDIUM)

- **Issue**: TON or GalaChain network congestion delays transactions
- **Impact**: Failed arbitrage, stuck transactions
- **Mitigation**:
  - Monitor pending transaction status
  - Implement transaction timeout and resubmit logic
  - Allow manual transaction cancellation
  - Adjust gas/fee settings based on network conditions

### Compliance Risks

**1. Regulatory Uncertainty** (LOW-MEDIUM)

- **Issue**: Trading bot regulations vary by jurisdiction
- **Impact**: Potential legal issues depending on user location
- **Mitigation**:
  - Add disclaimer about automated trading risks
  - Recommend users consult local regulations
  - Implement opt-in for arbitrage feature
  - Keep audit logs of all trades
  - Consider geographic restrictions if needed

**2. Terms of Service** (LOW)

- **Issue**: Exchange ToS may prohibit automated trading
- **Impact**: Account suspension, loss of access
- **Mitigation**:
  - Review ston.fi and gswap ToS for bot usage policies
  - Implement reasonable rate limiting (good API citizen)
  - Avoid aggressive practices (flash loans, MEV extraction)

## Recommendations

### Technology Choices

**1. Telegram Bot Framework**

- **Use**: `github.com/go-telegram/bot`
- **Rationale**: Active maintenance, clean architecture, positive community feedback

**2. TON Blockchain SDK**

- **Use**: `github.com/xssnick/tonutils-go`
- **Rationale**: Complete protocol implementation, concurrent-safe, well-maintained

**3. GalaChain Integration**

- **Use**: TypeScript microservice based on `cchuter/gswap-bot`
- **Rationale**: Only SDK available, proven reference implementation

**4. Inter-Service Communication**

- **Use**: gRPC with Protocol Buffers
- **Rationale**: Type-safe, high performance, supports streaming for real-time price updates

**5. Private Key Storage**

- **Development**: Encrypted environment variables (like gswap-bot)
- **Production**: HashiCorp Vault or Azure Managed HSM
- **Rationale**: Balance ease of use with security best practices

**6. Wallet Connection**

- **TonKeeper**: TonConnect protocol (universal link + deep link)
- **Gala Wallet**: WalletConnect v2 + fallback to manual key input
- **Rationale**: Standard protocols with proven mobile support

### Architecture Decisions

**1. Microservices Architecture**

```
telegram-trader/
├── cmd/
│   ├── telegram-bot/          # Go binary
│   └── galachain-service/     # TypeScript/Node.js
├── internal/                  # Go shared code
│   ├── ton/                   # TON blockchain client
│   ├── stonfi/                # STON.fi DEX integration
│   ├── wallet/                # TonConnect implementation
│   └── arbitrage/             # Arbitrage logic
├── galachain-service/         # TypeScript service
│   ├── src/
│   │   ├── gswap/            # GSwap SDK integration
│   │   ├── api/              # gRPC server
│   │   └── wallet/           # Gala wallet integration
│   └── package.json
├── proto/                     # gRPC definitions
│   └── galachain.proto
├── docker-compose.yml         # Local development
└── README.md
```

**2. Communication Protocol**

```protobuf
// proto/galachain.proto
service GalaChainService {
  rpc GetPrice(GetPriceRequest) returns (PriceResponse);
  rpc ExecuteSwap(SwapRequest) returns (SwapResponse);
  rpc GetBalance(BalanceRequest) returns (BalanceResponse);
  rpc WatchPrices(WatchPricesRequest) returns (stream PriceUpdate);
}
```

**3. Configuration Management**

- Development: `.env` files (separate for each service)
- Production: Secret manager (HashiCorp Vault)
- Config schema validation on startup
- Separate configs for mainnet/testnet

**4. Logging & Monitoring**

- Structured logging (JSON format)
- Separate log levels per service
- JSONL append-only trade log (like gswap-bot)
- Metrics: Prometheus + Grafana
- Alerting: PagerDuty or similar for production

### Development Phasing

**Phase 1: Foundation (Weeks 1-2)**

- Set up Go project structure
- Implement basic Telegram bot with command routing
- Set up TypeScript GalaChain service
- Establish gRPC communication
- Implement basic health checks

**Phase 2: Wallet Integration (Weeks 3-4)**

- Implement TonConnect for TonKeeper
- Implement WalletConnect for Gala wallet
- Add wallet balance checking
- Secure key storage implementation

**Phase 3: Trading Features (Weeks 5-6)**

- Integrate ston.fi API (price checking)
- Integrate gswap via TypeScript service
- Implement manual swap commands
- Add swap history tracking

**Phase 4: Arbitrage (Weeks 7-8)**

- Implement price monitoring loop
- Add spread detection logic
- Implement dry-run arbitrage mode
- Add position sizing with minimum checks
- Implement actual arbitrage execution

**Phase 5: Production Hardening (Weeks 9-10)**

- Comprehensive error handling
- Rate limiting and retry logic
- Security audit of key management
- Performance optimization
- Deployment automation

### Security Best Practices

**1. Key Management**

- Use encrypted private keys with separate decryption password
- Implement key rotation capability
- Never log private keys or decryption passwords
- Use HSM for production environments
- Implement hot/cold wallet split

**2. Transaction Security**

- Always simulate swaps before execution
- Implement transaction amount limits
- Require 2FA for large transactions (optional feature)
- Log all transactions with timestamp and user ID
- Implement manual approval mode

**3. Access Control**

- Whitelist Telegram user IDs in production
- Implement per-user rate limiting
- Session timeout after inactivity
- Admin commands require separate authentication

**4. Network Security**

- Use HTTPS/TLS for all API calls
- Implement certificate pinning for critical APIs
- Use VPN for production deployment
- Firewall rules to restrict service access

**5. Monitoring & Auditing**

- Log all user commands and bot responses
- Monitor for suspicious activity patterns
- Alert on unusual trade sizes or frequencies
- Regular security audits of dependencies
- Automated vulnerability scanning

## References

### Official Documentation

**TON Blockchain:**

- TON Docs: https://docs.ton.org
- TON SDK Documentation: https://docs.ton.org/v3/guidelines/dapps/apis-sdks/sdk
- TonConnect Integration: https://docs.ton.org/develop/dapps/ton-connect/integration
- TON Deep Linking: https://docs.ton.org/ecosystem/wallet-apps/deep-links

**STON.fi:**

- Main Site: https://ston.fi
- Developer Documentation: https://docs.ston.fi/
- API Reference: https://docs.ston.fi/developer-section/dex/api/reference
- Swagger UI: https://api.ston.fi/swagger-ui/
- SDK Documentation: https://docs.ston.fi/docs/developer-section/sdk

**GalaChain:**

- GalaChain Developer Site: https://www.galachain.com/
- SDK Documentation: https://docs.galachain.com/v2.04/
- GalaSwap API: https://galaswap.gala.com/info/api.html
- Explorer API: https://explorer-api.galachain.com/docs/
- Gala Support: https://support.gala.com/

**Telegram:**

- Bot API Samples: https://core.telegram.org/bots/samples

**WalletConnect:**

- iOS Mobile Linking: https://docs.walletconnect.network/wallet-sdk/ios/mobile-linking

### GitHub Repositories

**Reference Implementation:**

- gswap-bot (TypeScript): https://github.com/cchuter/gswap-bot
- GalaChain SDK: https://github.com/GalaChain/sdk
- GalaChain gswap-bot: https://github.com/GalaChain/galaswap-bot

**Go Telegram Bot Libraries:**

- go-telegram/bot: https://github.com/go-telegram/bot
- telebot: https://github.com/tucnak/telebot
- telego: https://github.com/mymmrac/telego
- go-telegram-bot-api: https://github.com/go-telegram-bot-api/telegram-bot-api

**TON Go SDKs:**

- tonutils-go: https://github.com/xssnick/tonutils-go
- tongo (Tonkeeper): https://github.com/tonkeeper/tongo
- tonlib-go (official): https://github.com/ton-blockchain/tonlib-go
- Awesome TON: https://github.com/ton-community/awesome-ton

**TonConnect:**

- TonConnect SDK: https://ton-connect.github.io/sdk/modules/_tonconnect_sdk.html
- Tonkeeper Wallet API: https://github.com/tonkeeper/wallet-api
- Tonkeeper TonConnect Guide: https://tonkeeper.gitbook.io/ton-connect-2.0-guide-for-sdk/ton-connect-in-tonkeeper

### Technical Articles

**Architecture:**

- "Why Golang Is Ideal for Microservices Architecture (2026)": https://medium.com/@amin-softtech/why-golang-is-ideal-for-microservices-architecture-2026-and-beyond-fd0775c97d4c
- "TonConnect Deep Dive": https://medium.com/tonx-lab/tonconnect-deep-dive-how-tons-wallet-connect-protocol-powers-dapps-f9ee67381c2a
- "Building a multi-protocol arbitrage bot": https://medium.com/@suyashnyn1/building-a-multi-protocol-arbitrage-bot-c8af44f2bfb9

**Arbitrage Trading:**

- "Crypto Arbitrage in 2025": https://wundertrading.com/journal/en/learn/article/crypto-arbitrage
- "High-Frequency Arbitrage and Profit Maximization": https://medium.com/@gwrx2005/high-frequency-arbitrage-and-profit-maximization-across-cryptocurrency-exchanges-4842d7b7d4d9
- "Best Crypto Arbitrage Bots in 2026": https://99bitcoins.com/analysis/crypto-arbitrage-bots/
- "Crypto Arbitrage: Complete Guide": https://www.kucoin.com/learn/trading/crypto-arbitrage-complete-guide-to-making-low-risk-gains
- Kraken Arbitrage Guide: https://www.kraken.com/learn/trading/crypto-arbitrage

**Security:**

- "Private Key Security": https://www.alphaexcapital.com/cryptocurrencies/buying-selling-and-storing-crypto/crypto-wallets/private-key-explained/
- Azure Managed HSM Best Practices: https://learn.microsoft.com/en-us/azure/key-vault/managed-hsm/best-practices
- HashiCorp Vault HSM: https://developer.hashicorp.com/vault/docs/enterprise/hsm

**DEX Integration:**

- "Cross-Chain DEX Arbitrage Bot": https://www.progressiverobot.com/2025/08/06/cross‑chain-dex-arbitrage-bot/
- Multi-Exchange Arbitrage Development: https://pixelplex.io/crypto-arbitrage-bot/
- TON APIs and Libraries Guide: https://chainstack.com/ton-ultimate-guide-to-apis-and-interaction-libraries/

### Tools & Resources

**Package Repositories:**

- tonutils-go package: https://pkg.go.dev/github.com/xssnick/tonutils-go
- go-telegram/bot package: https://pkg.go.dev/github.com/go-telegram/bot

**Community:**

- TON Developers Chat: (find in TON Docs)
- Gala Games Developer Discord: (find in GalaChain Docs)
- Telegram Bot Developers: https://core.telegram.org/bots

## Next Steps

### For Requirements Phase

**1. Functional Requirements Deep Dive**

- Define exact user flows for each command (/wallet, /balance, /swap, etc.)
- Specify wallet connection UX (QR codes, deep links, fallbacks)
- Detail arbitrage trigger conditions and execution logic
- Define error handling and user notification patterns
- Specify data persistence requirements (trade history, user preferences)

**2. Non-Functional Requirements**

- Performance targets (response time, arbitrage execution speed)
- Scalability requirements (concurrent users, trade volume)
- Security requirements (authentication, authorization, encryption)
- Availability targets (uptime, disaster recovery)
- Monitoring and observability requirements

**3. Data Modeling**

- User data (Telegram ID, wallet addresses, preferences)
- Transaction history (trades, swaps, arbitrage executions)
- Price data (historical for analytics, real-time for arbitrage)
- Configuration data (minimums, thresholds, limits)

**4. API Contracts**

- gRPC service definitions (messages, RPCs)
- Internal API boundaries
- Error codes and handling
- Versioning strategy

**5. Security Requirements**

- Key storage and encryption schemes
- Authentication mechanisms
- Authorization policies
- Audit logging requirements
- Compliance considerations

**6. Testing Strategy**

- Unit testing approach
- Integration testing with live APIs (testnet)
- End-to-end testing scenarios
- Security testing (penetration, vulnerability scanning)
- Performance testing (load, stress)

**7. Deployment Requirements**

- Infrastructure needs (servers, databases, message queues)
- CI/CD pipeline
- Monitoring and alerting
- Backup and disaster recovery
- Rollback procedures

**8. Open Questions to Resolve**

- Should arbitrage be fully automatic or require user approval?
- What level of trading history should be stored?
- Should the bot support multiple users or single user?
- What's the target environment (cloud, on-premise, hybrid)?
- Should testnet mode be supported for safe testing?
- What's the acceptable latency for arbitrage execution?
- Should the bot support stop-loss or take-profit orders?

### Immediate Action Items

1. Set up development environment (Go, Node.js, Docker)
2. Create project repository structure
3. Initialize Go module and npm project
4. Set up gRPC tooling (protoc, plugins)
5. Clone and study gswap-bot reference implementation
6. Test TonConnect flow with sample Telegram bot
7. Test GalaChain API with sample TypeScript code
8. Create testnet wallets on both TON and GalaChain
9. Document setup process for team members
10. Schedule requirements review session

### Key Decisions Needed

1. **Arbitrage Automation Level**: Fully automatic vs. user-approved vs. semi-automatic?
2. **Multi-User Support**: Single user bot vs. multi-tenant service?
3. **Deployment Target**: Cloud provider (AWS, GCP, Azure) vs. on-premise vs. hybrid?
4. **Testing Environment**: Mainnet from day 1 vs. testnet first?
5. **Key Management**: Simple encryption vs. HSM from start?
6. **Monitoring Solution**: Self-hosted (Prometheus) vs. managed (Datadog)?
7. **Database**: Needed for user data? SQLite, PostgreSQL, or document DB?
8. **Price Data Storage**: In-memory only vs. persistent for analytics?

---

**Research Phase Complete**: All necessary technology evaluation and feasibility analysis completed. Ready to proceed to requirements phase with clear technology choices and architecture direction.
