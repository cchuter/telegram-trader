.PHONY: help build test lint run docker-up docker-down proto-gen migrate clean

# Default target
help:
	@echo "Telegram Trader - Available Make Targets"
	@echo ""
	@echo "Build & Run:"
	@echo "  build         - Build both Go bot and TypeScript GalaChain service"
	@echo "  run           - Run the telegram bot locally (requires build)"
	@echo "  test          - Run all tests (Go + TypeScript)"
	@echo "  lint          - Run linters on all code (golangci-lint + ESLint)"
	@echo ""
	@echo "Docker:"
	@echo "  docker-up     - Start all services (postgres, galachain-service, bot)"
	@echo "  docker-down   - Stop all services and remove containers"
	@echo ""
	@echo "Development:"
	@echo "  proto-gen     - Regenerate protobuf code for Go and TypeScript"
	@echo "  migrate       - Run database migrations"
	@echo "  clean         - Remove build artifacts and binaries"
	@echo ""
	@echo "Examples:"
	@echo "  make build && make run    # Build and run locally"
	@echo "  make docker-up            # Start all services with Docker"
	@echo "  make test                 # Run full test suite"

# Build both services
build:
	@echo "Building Go bot service..."
	@go build -o telegram-bot ./cmd/telegram-bot
	@echo "Building TypeScript GalaChain service..."
	@cd galachain-service && npm run build

# Run all tests
test:
	@echo "Running Go tests..."
	@go test -v ./...
	@echo "Running TypeScript tests..."
	@cd galachain-service && npm test

# Run linters
lint:
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found in PATH, trying ~/go/bin/golangci-lint"; }
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	elif [ -x ~/go/bin/golangci-lint ]; then \
		~/go/bin/golangci-lint run ./...; \
	else \
		echo "Error: golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	@echo "Running ESLint..."
	@cd galachain-service && npm run lint

# Run the bot locally
run: build
	@./telegram-bot

# Start all services with Docker Compose
docker-up:
	@echo "Starting all services with Docker Compose..."
	@cd deployments && docker-compose up -d
	@echo "Services started. Use 'docker-compose logs -f' to view logs."

# Stop all services
docker-down:
	@echo "Stopping all services..."
	@cd deployments && docker-compose down
	@echo "Services stopped."

# Regenerate protobuf code
proto-gen:
	@echo "Generating Go protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/galachain.proto
	@echo "Generating TypeScript protobuf code..."
	@cd galachain-service && npm run proto
	@echo "Protobuf code generation complete."

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@if [ -f telegram-trader.db ]; then \
		echo "SQLite database exists at telegram-trader.db"; \
		echo "Migrations are applied automatically on startup."; \
	else \
		echo "No database found. Migrations will run on first startup."; \
	fi
	@echo "For PostgreSQL migrations, use DATABASE_URL environment variable."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f telegram-bot
	@rm -f migrate_sqlite_to_postgres
	@rm -rf galachain-service/dist
	@rm -f coverage.out coverage.html
	@rm -rf coverage/
	@echo "Clean complete."
