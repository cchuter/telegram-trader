# Protocol Buffers - gRPC Interface Definitions

This directory contains the Protocol Buffer definitions for the gRPC interface between the Go Telegram bot service and the TypeScript GalaChain service.

## Files

- `galachain.proto` - Service and message definitions for GalaChain operations

## Service Definition

The `GalaChainService` provides 6 RPC methods:

1. **GetPrice** - Get current price for a trading pair
2. **GetBalance** - Get wallet balance for a user
3. **ExecuteSwap** - Execute a token swap on GalaChain
4. **WatchPrices** - Stream price updates (server streaming)
5. **CreateWalletSession** - Create wallet connection session
6. **HealthCheck** - Check service health status

## Message Types

### Price Operations
- `GetPriceRequest` / `PriceResponse`
- `WatchPricesRequest` / `PriceUpdate`

### Balance Operations
- `BalanceRequest` / `BalanceResponse`
- `TokenBalance` (nested message)

### Swap Operations
- `SwapRequest` / `SwapResponse`

### Wallet Operations
- `WalletSessionRequest` / `WalletSessionResponse`

### Health Check
- `HealthCheckRequest` / `HealthCheckResponse`

## Code Generation

### Prerequisites

#### For Go
Install the Protocol Buffers compiler and Go plugins:
```bash
# Install protoc (macOS)
brew install protobuf

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Ensure $GOPATH/bin is in your PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

#### For TypeScript
Install the required npm packages in the galachain-service directory:
```bash
cd galachain-service
npm install --save-dev @grpc/grpc-js @grpc/proto-loader
npm install --save-dev grpc-tools grpc_tools_node_protoc_ts
```

### Generate Go Code

From the project root directory:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/galachain.proto
```

This generates:
- `internal/galachain/pb/galachain.pb.go` - Message type definitions
- `internal/galachain/pb/galachain_grpc.pb.go` - gRPC service client and server code

### Generate TypeScript Code

From the project root directory:

```bash
# Create output directory
mkdir -p galachain-service/src/types/grpc

# Generate TypeScript definitions
grpc_tools_node_protoc \
  --plugin=protoc-gen-ts=./galachain-service/node_modules/.bin/protoc-gen-ts \
  --ts_out=grpc_js:./galachain-service/src/types/grpc \
  --js_out=import_style=commonjs:./galachain-service/src/types/grpc \
  --grpc_out=grpc_js:./galachain-service/src/types/grpc \
  -I ./proto \
  proto/galachain.proto
```

This generates:
- `galachain-service/src/types/grpc/galachain_pb.js` - Message type implementations
- `galachain-service/src/types/grpc/galachain_pb.d.ts` - TypeScript type definitions
- `galachain-service/src/types/grpc/galachain_grpc_pb.js` - gRPC service code
- `galachain-service/src/types/grpc/galachain_grpc_pb.d.ts` - gRPC service types

### Alternative: Using @grpc/proto-loader (Dynamic)

For TypeScript, you can also use dynamic proto loading instead of code generation:

```typescript
import * as protoLoader from '@grpc/proto-loader';
import * as grpc from '@grpc/grpc-js';

const packageDefinition = protoLoader.loadSync(
  './proto/galachain.proto',
  {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
  }
);

const protoDescriptor = grpc.loadPackageDefinition(packageDefinition);
const galachainProto = protoDescriptor.galachain as any;
```

## Updating Definitions

After modifying `galachain.proto`:

1. Regenerate Go code: `make proto-go` (or run the Go generation command above)
2. Regenerate TypeScript code: `make proto-ts` (or run the TypeScript generation command above)
3. Update both services to use the new types
4. Test the changes thoroughly

## Version Compatibility

- Proto3 syntax
- Go package: `github.com/cchuter/telegram-trader/internal/galachain/pb`
- TypeScript package: `galachain` (in generated code)

## Notes

- All monetary amounts use `string` type to avoid floating-point precision issues
- Timestamps use Unix timestamp (int64) for consistency
- Optional fields use `optional` keyword for proto3 semantics
- Error handling uses status codes and error messages in responses
