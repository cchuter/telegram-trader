.PHONY: lint test build clean run help

# Default target
help:
	@echo "Available targets:"
	@echo "  lint       - Run golangci-lint on all Go code"
	@echo "  test       - Run all tests"
	@echo "  build      - Build the telegram bot binary"
	@echo "  clean      - Remove built binaries"
	@echo "  run        - Run the telegram bot locally"

# Run golangci-lint
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found in PATH, trying ~/go/bin/golangci-lint"; }
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	elif [ -x ~/go/bin/golangci-lint ]; then \
		~/go/bin/golangci-lint run ./...; \
	else \
		echo "Error: golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Run tests
test:
	go test -v ./...

# Build the bot
build:
	go build -o telegram-bot ./cmd/telegram-bot

# Clean built binaries
clean:
	rm -f telegram-bot

# Run the bot locally
run: build
	./telegram-bot
