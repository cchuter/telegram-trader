# Build stage
FROM golang:1.25.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o telegram-bot ./cmd/telegram-bot

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/telegram-bot .

# Create logs directory
RUN mkdir -p /root/logs

# Expose any necessary ports (if needed in future)
# EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD pgrep -f telegram-bot || exit 1

# Run the bot
CMD ["./telegram-bot"]
