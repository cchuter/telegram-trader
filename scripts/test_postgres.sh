#!/bin/bash
set -e

# Script to test PostgreSQL integration
# This script starts a PostgreSQL container, runs tests, and cleans up

CONTAINER_NAME="telegram-trader-postgres-test"
POSTGRES_USER="testuser"
POSTGRES_PASSWORD="testpass"
POSTGRES_DB="testdb"
DATABASE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable"

echo "Starting PostgreSQL container..."
docker run --name $CONTAINER_NAME \
  -e POSTGRES_USER=$POSTGRES_USER \
  -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
  -e POSTGRES_DB=$POSTGRES_DB \
  -p 5432:5432 \
  -d postgres:15-alpine

echo "Waiting for PostgreSQL to be ready..."
sleep 5

# Wait for PostgreSQL to accept connections
for i in {1..30}; do
  if docker exec $CONTAINER_NAME pg_isready -U $POSTGRES_USER > /dev/null 2>&1; then
    echo "PostgreSQL is ready!"
    break
  fi
  echo "Waiting... ($i/30)"
  sleep 1
done

# Cleanup function
cleanup() {
  echo ""
  echo "Cleaning up..."
  docker stop $CONTAINER_NAME > /dev/null 2>&1 || true
  docker rm $CONTAINER_NAME > /dev/null 2>&1 || true
  echo "Cleanup complete"
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Run tests
echo ""
echo "Running PostgreSQL integration tests..."
DATABASE_URL=$DATABASE_URL go test ./internal/storage -v -run TestPostgresIntegration

echo ""
echo "Running migration tests..."
DATABASE_URL=$DATABASE_URL go test ./internal/storage -v -run TestPostgresMigrations

echo ""
echo "All tests passed!"
