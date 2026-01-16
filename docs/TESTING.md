# Testing Guide

This document describes the testing infrastructure, how to run tests, and coverage requirements for the Telegram Trader project.

## Table of Contents

- [Overview](#overview)
- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Test Environment Setup](#test-environment-setup)
- [Test Coverage Requirements](#test-coverage-requirements)
- [Test Data and Fixtures](#test-data-and-fixtures)
- [Writing Tests](#writing-tests)
- [Troubleshooting](#troubleshooting)

## Overview

The Telegram Trader project uses a comprehensive testing strategy with three test levels:

1. **Unit Tests**: Fast, isolated tests for individual functions and modules
2. **Integration Tests**: Tests requiring external dependencies (APIs, databases)
3. **End-to-End (E2E) Tests**: Complete workflow tests simulating real user scenarios

The project consists of two services:
- **Go Bot Service**: Telegram bot with command handlers (Go 1.21+)
- **TypeScript GalaChain Service**: gRPC service for GalaChain integration (Node.js 20+)

## Test Structure

```
telegram-trader/
├── test/                           # Shared test infrastructure
│   ├── mocks/                      # Mock implementations
│   │   ├── database.go             # MockDatabase
│   │   ├── blockchain.go           # MockBlockchainClient
│   │   ├── dex.go                  # MockDEXClient
│   │   └── grpc.go                 # MockGRPCClient
│   ├── testutil/                   # Test helpers
│   │   └── helpers.go              # Setup functions, assertions
│   ├── integration/                # Integration tests
│   │   ├── grpc_test.go            # gRPC communication tests
│   │   ├── stonfi_test.go          # ston.fi API tests
│   │   └── database_test.go        # Database operation tests
│   └── e2e/                        # End-to-end tests
│       ├── wallet_test.go          # Wallet connection flow
│       ├── swap_test.go            # Swap execution flow
│       └── arbitrage_test.go       # Arbitrage detection/execution
│
├── internal/*/                     # Unit tests alongside code
│   └── *_test.go                   # Package-level unit tests
│
└── galachain-service/
    ├── test/
    │   └── integration.test.ts     # TypeScript integration tests
    └── src/**/__tests__/           # TypeScript unit tests
        └── *.test.ts
```

## Running Tests

### Unit Tests

Unit tests run quickly and require no external dependencies.

#### Go Unit Tests

```bash
# Run all unit tests (excludes integration/e2e)
go test ./... -v

# Run tests for specific package
go test ./internal/arbitrage -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run tests matching pattern
go test ./... -run TestCalculatePositionSize

# Short mode (skip slow tests)
go test ./... -short
```

#### TypeScript Unit Tests

```bash
cd galachain-service

# Run all unit tests
npm run test

# Run with coverage
npm run test -- --coverage

# Watch mode (re-run on file changes)
npm run test -- --watch

# Run specific test file
npm run test -- src/gswap/__tests__/client.test.ts
```

### Integration Tests

Integration tests require network access and may call real APIs. They use build tags to separate from unit tests.

#### Go Integration Tests

```bash
# Run integration tests (includes +build integration tag)
go test ./test/integration/... -v

# Run specific integration test
go test ./test/integration -run TestStonFiSimulateSwap -v

# Run with timeout (some tests may be slow)
go test ./test/integration/... -timeout 60s -v
```

**Network Requirements:**
- `grpc_test.go`: Requires gRPC server running on `localhost:50051`
- `stonfi_test.go`: Requires internet access for ston.fi API calls
- `database_test.go`: Creates temporary databases (no external dependencies)

#### TypeScript Integration Tests

```bash
cd galachain-service

# Run integration tests
npm run test:integration

# With verbose output
npm run test:integration -- --verbose
```

### End-to-End (E2E) Tests

E2E tests simulate complete user workflows. Some tests require testnet accounts with funded wallets.

```bash
# Run all e2e tests (skips tests requiring testnet funds)
go test ./test/e2e/... -v

# Run e2e tests with testnet execution
TESTNET_FUNDED=true go test ./test/e2e/... -v

# Run specific e2e test
go test ./test/e2e -run TestWalletHandlerIntegration -v

# Run with extended timeout for blockchain operations
go test ./test/e2e/... -timeout 5m -v
```

**Testnet Requirements:**
- `swap_test.go`: Skips actual blockchain execution unless `TESTNET_FUNDED=true`
- `arbitrage_test.go`: Skips concurrent swap execution unless `TESTNET_FUNDED=true`
- `wallet_test.go`: Runs fully without testnet funds (mocks connection)

### Run All Tests

```bash
# From project root, run comprehensive test suite
./scripts/coverage.sh
```

This script:
1. Runs all Go tests with coverage
2. Runs all TypeScript tests with coverage
3. Generates HTML coverage reports
4. Enforces 70% coverage threshold
5. Fails if coverage below threshold

**Coverage Reports:**
- Go: `coverage/coverage.html`
- TypeScript: `coverage/typescript/lcov-report/index.html`

### Continuous Integration

Tests run automatically on every push via GitHub Actions (`.github/workflows/ci.yml`):

- **Test job**: Runs Go and TypeScript tests, uploads coverage to Codecov
- **Lint job**: Runs `go vet` and `go fmt` checks
- **Build job**: Verifies both services compile successfully

## Test Environment Setup

### Prerequisites

#### Go Tests

```bash
# Install Go 1.21+
go version  # Verify installation

# Install dependencies
go mod download

# Install test dependencies (already in go.mod)
# - github.com/stretchr/testify (assertions and mocks)
```

#### TypeScript Tests

```bash
# Install Node.js 20+
node --version  # Verify installation

cd galachain-service

# Install dependencies (includes Jest test framework)
npm install

# Jest and ts-jest already configured in package.json
```

### Environment Variables for Testing

Create `.env.test` file for test-specific configuration:

```env
# Bot Configuration (use test bot token)
BOT_TOKEN=test_bot_token_here
BOT_ADMIN_USER_IDS=123456789

# Database (use in-memory or test database)
DATABASE_URL=file::memory:?cache=shared

# Encryption (test key - DO NOT use in production)
ENCRYPTION_KEY=dGVzdGtleXRlc3RrZXl0ZXN0a2V5dGVzdGtleTE=

# GalaChain Service (mock or local instance)
GALACHAIN_SERVICE_URL=localhost:50051

# Optional: Enable testnet execution
TESTNET_FUNDED=false

# Optional: Test timeout
TEST_TIMEOUT=60s
```

### Database Setup for Tests

#### SQLite (Default)

Unit and integration tests use temporary SQLite databases:

```go
// Automatic setup in tests
db, err := storage.InitDB("file::memory:?cache=shared")
```

No manual setup required - databases are created/cleaned up automatically.

#### PostgreSQL (Integration Tests)

For PostgreSQL integration tests:

```bash
# Start PostgreSQL via Docker
docker run --name postgres-test -e POSTGRES_PASSWORD=test -p 5432:5432 -d postgres:15

# Run tests with PostgreSQL
DATABASE_URL=postgres://postgres:test@localhost:5432/telegram_trader_test?sslmode=disable \
  go test ./test/integration/database_test.go -v

# Cleanup
docker stop postgres-test
docker rm postgres-test
```

### Mock GalaChain Service for Integration Tests

Integration tests include a mock gRPC server. To run the real GalaChain service:

```bash
# Terminal 1: Start GalaChain service
cd galachain-service
npm run dev

# Terminal 2: Run integration tests
go test ./test/integration/grpc_test.go -v
```

## Test Coverage Requirements

### Coverage Targets

Per NFR-040 (Test Coverage):
- **Minimum Coverage**: 70% for both Go and TypeScript
- **Target Coverage**: 90% for critical paths (auth, validation, crypto)

### Current Coverage

Run `./scripts/coverage.sh` to see current coverage:

```
Go coverage: 26.1%
TypeScript coverage: 21.13%
Combined coverage: 23.61%
```

**Note**: Current coverage is below target. Phase 3 (Testing) tasks focus on reaching 70% threshold.

### Coverage by Component

**High Coverage (90%+)**:
- `internal/validation` - 100% (input validation)
- `internal/arbitrage/calculator.go` - 92% (position sizing)
- `internal/wallet/encryption.go` - 87% (key encryption)

**Medium Coverage (70-90%)**:
- `internal/bot/middleware/auth.go` - 84% (authentication)
- `internal/bot/middleware/ratelimit.go` - 82% (rate limiting)

**Low Coverage (<70%)**: Most handlers and integration code (covered by integration/e2e tests)

### Excluded from Coverage

- Code requiring real bot instances (`bot.SendMessage` calls)
- Unreachable error paths in crypto libraries (AES-GCM failures)
- Main functions and initialization code
- Generated protobuf code

## Test Data and Fixtures

### Mock Data

Test mocks are located in `test/mocks/`:

**MockDatabase** (`test/mocks/database.go`):
```go
mockDB := mocks.NewMockDatabase()

// Pre-populate test data
session := testutil.CreateTestUserSession(userID, chatID)
mockDB.SaveUserSession(ctx, session)

// Configure error simulation
mockDB.SetGetError(errors.New("database unavailable"))

// Reset between tests
mockDB.Reset()
```

**MockBlockchainClient** (`test/mocks/blockchain.go`):
```go
mockClient := mocks.NewMockBlockchainClient()

// Configure balance responses
mockClient.SetBalance("10.5")
mockClient.SetJettonBalance("1000.0")

// Track method calls
mockClient.GetBalanceCallCount  // How many times GetBalance was called
```

**MockDEXClient** (`test/mocks/dex.go`):
```go
mockDEX := mocks.NewMockDEXClient()

// Configure swap simulation
mockDEX.SetSwapSimulation(&dex.SwapSimulation{
    OutputAmount: "26000000000",
    Fee:          "78900000",
    // ...
})
```

### Test Fixtures

**Token Addresses**:
```go
const (
    TONAddress  = "EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM9c"
    GALAAddress = "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV"
)
```

**Test User IDs**:
```go
const (
    TestUserID  = int64(123456789)
    TestChatID  = int64(987654321)
)
```

**Test Wallet Addresses**:
```go
const (
    ValidTONAddress  = "EQDtFpEwcFAEcRe5mLVh2N6C0x-_hJEM7W61_JLnSF74p4q2"
    ValidGalaAddress = "2954bd1e38c2dff11a2ee8798e2f59b99c206e64bef7bdf5c4ac9a91315e0d34"
)
```

### Helper Functions

Located in `test/testutil/helpers.go`:

```go
// Setup functions
db := testutil.SetupMockDatabase()
client := testutil.SetupMockBlockchain()

// Create test entities
session := testutil.CreateTestUserSession(userID, chatID)
wallet := testutil.CreateTestWalletSession(userID, address)
trade := testutil.CreateTestTradeHistory(userID, "swap", "stonfi")

// Assertions
testutil.AssertNoError(t, err)
testutil.AssertEqual(t, expected, actual)
testutil.AssertNotNil(t, value)

// Utilities
ctx := testutil.ContextWithTimeout(5 * time.Second)
success := testutil.WaitForCondition(func() bool { ... }, timeout)
```

## Writing Tests

### Unit Test Example

```go
// internal/arbitrage/calculator_test.go
package arbitrage

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCalculatePositionSize(t *testing.T) {
    calc := NewCalculator()

    t.Run("valid trade with sufficient balance", func(t *testing.T) {
        position := calc.CalculatePositionSize(
            10.0,   // TON balance
            1000.0, // GALA balance
            BuyTonSellGala,
        )

        assert.True(t, position.Valid)
        assert.Equal(t, 5.0, position.TONAmount)  // 50% of balance
        assert.InDelta(t, 500.0, position.GALAAmount, 0.01)
    })

    t.Run("insufficient balance below minimum", func(t *testing.T) {
        position := calc.CalculatePositionSize(
            0.5,  // Below 1 TON minimum
            5.0,  // Below 10 GALA minimum
            BuyTonSellGala,
        )

        assert.False(t, position.Valid)
        assert.Contains(t, position.Reason, "insufficient")
    })
}
```

### Integration Test Example

```go
// +build integration

package integration

import (
    "context"
    "testing"
    "time"
)

func TestStonFiSimulateSwap(t *testing.T) {
    client := stonfi.NewClient()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    result, err := client.SimulateSwap(ctx, &stonfi.SwapSimulateRequest{
        OfferAddress:   stonfi.TONAddress,
        AskAddress:     stonfi.GALAAddress,
        Units:          "1000000000",  // 1 TON
        SlippageTolerance: "1.0",
    })

    require.NoError(t, err)
    assert.NotEmpty(t, result.OutputAmount)
    assert.True(t, result.SlippagePercentage() < 10.0)
}
```

### E2E Test Example

```go
package e2e

func TestSwapConfirmationFlow(t *testing.T) {
    // Setup
    db := testutil.SetupMockDatabase()
    walletMgr := wallet.NewManager(db, encKey)

    // Create wallet session with encrypted key
    session := &storage.WalletSession{
        UserID:      testUserID,
        WalletType:  "tonconnect",
        WalletAddress: testAddress,
        Connected:   true,
    }
    privateKey := "test_private_key_here"
    encryptedKey, _ := walletMgr.EncryptPrivateKey(privateKey)
    session.TonConnectPrivateKey = encryptedKey
    db.SaveWalletSession(context.Background(), session)

    // Execute swap simulation
    swapResult := executeSwapSimulation(t, testAmount)

    // Verify swap logged to trades.jsonl
    trades := readTradesFromLog(t, "logs/trades.jsonl")
    assert.Equal(t, 1, len(trades))
    assert.Equal(t, "swap", trades[0].Type)
    assert.NotEmpty(t, trades[0].TxHash)
}
```

### TypeScript Test Example

```typescript
// galachain-service/src/gswap/__tests__/client.test.ts
import { GSwapClient } from '../client';

describe('GSwapClient', () => {
  let client: GSwapClient;

  beforeEach(() => {
    client = new GSwapClient('https://api.gswap.example.com');
  });

  afterEach(() => {
    client.close();
  });

  it('should fetch GTON/GALA price', async () => {
    const price = await client.getPrice('GTON', 'GALA');

    expect(price).toBeDefined();
    expect(price.price).toMatch(/^\d+(\.\d+)?$/);
    expect(price.timestamp).toBeGreaterThan(0);
  });

  it('should handle invalid token pair', async () => {
    await expect(
      client.getPrice('INVALID', 'TOKEN')
    ).rejects.toThrow();
  });
});
```

## Troubleshooting

### Common Issues

#### Issue: "database is locked" (SQLite)

```bash
# Solution: Use in-memory database for tests
DATABASE_URL=file::memory:?cache=shared go test ./...
```

#### Issue: Integration tests timeout

```bash
# Solution: Increase timeout
go test ./test/integration/... -timeout 120s -v
```

#### Issue: "gRPC server not running"

```bash
# Solution: Start GalaChain service first
cd galachain-service && npm run dev

# Then run tests in another terminal
go test ./test/integration/grpc_test.go -v
```

#### Issue: TypeScript tests fail with module not found

```bash
# Solution: Rebuild TypeScript code
cd galachain-service
npm run build
npm run test
```

#### Issue: Coverage reports not generated

```bash
# Solution: Ensure bc (calculator) is installed
brew install bc  # macOS
apt-get install bc  # Linux

# Run coverage script again
./scripts/coverage.sh
```

#### Issue: "context deadline exceeded" in ston.fi tests

This is expected when testing timeout handling. Tests use very short deadlines (1 nanosecond) to trigger timeouts:

```go
// This is intentional - testing error handling
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
```

#### Issue: Mock data not resetting between tests

```go
// Solution: Call Reset() in test cleanup
func TestSomething(t *testing.T) {
    mockDB := mocks.NewMockDatabase()
    defer mockDB.Reset()  // Clean up after test

    // ... test code ...
}
```

### Debug Tips

**Enable verbose logging:**
```bash
# Go tests
go test ./... -v -run TestName

# TypeScript tests
npm run test -- --verbose
```

**Run single test:**
```bash
# Go
go test ./internal/arbitrage -run TestCalculatePositionSize -v

# TypeScript
npm run test -- -t "should fetch price"
```

**Check test coverage for specific file:**
```bash
go test ./internal/arbitrage -coverprofile=coverage.out
go tool cover -func=coverage.out | grep calculator.go
```

**Debug hanging tests:**
```bash
# Add timeout and race detection
go test ./... -timeout 30s -race -v
```

**View integration test network calls:**
```bash
# Enable debug logging for HTTP client
GODEBUG=http2debug=1 go test ./test/integration/stonfi_test.go -v
```

### Getting Help

- **Check logs**: Test output includes correlation IDs for debugging
- **Read test code**: Tests serve as executable documentation
- **Check coverage reports**: Identify untested code paths
- **Review CI logs**: GitHub Actions logs show exact test failures
- **Check documentation**: See `docs/` directory for more info

## Next Steps

After running tests:
1. Review coverage reports to identify gaps
2. Add tests for uncovered code paths
3. Update this documentation if test infrastructure changes
4. Ensure all tests pass before creating pull requests

For deployment testing, see `docs/DEPLOYMENT.md` (future).
For security testing, see `docs/SECURITY.md` (future).
