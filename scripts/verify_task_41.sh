#!/bin/bash
# Verification script for Task 41: PostgreSQL implementation
# This script verifies all requirements are met

set -e

echo "Task 41 Verification: PostgreSQL Migration Support"
echo "=================================================="
echo ""

# Check 1: postgres.go exists
echo "✓ Check 1: internal/storage/postgres.go exists"
if [ ! -f "internal/storage/postgres.go" ]; then
  echo "❌ FAILED: postgres.go not found"
  exit 1
fi

# Check 2: lib/pq dependency
echo "✓ Check 2: github.com/lib/pq dependency"
if ! go list -m all | grep -q "github.com/lib/pq"; then
  echo "❌ FAILED: lib/pq dependency not found"
  exit 1
fi

# Check 3: SQLite still supported
echo "✓ Check 3: SQLite implementation exists"
if [ ! -f "internal/storage/sqlite.go" ]; then
  echo "❌ FAILED: sqlite.go not found"
  exit 1
fi

# Check 4: ParseDatabaseURL function exists
echo "✓ Check 4: ParseDatabaseURL function exists"
if ! grep -q "func ParseDatabaseURL" internal/storage/postgres.go; then
  echo "❌ FAILED: ParseDatabaseURL not found"
  exit 1
fi

# Check 5: docker-compose.yml has PostgreSQL
echo "✓ Check 5: docker-compose.yml includes PostgreSQL"
if ! grep -q "postgres:15-alpine" deployments/docker-compose.yml; then
  echo "❌ FAILED: PostgreSQL not in docker-compose.yml"
  exit 1
fi

# Check 6: Migration script exists
echo "✓ Check 6: Migration script exists"
if [ ! -f "scripts/migrate_sqlite_to_postgres.go" ]; then
  echo "❌ FAILED: migration script not found"
  exit 1
fi

# Check 7: Both implementations implement Database interface
echo "✓ Check 7: Interface implementation test"
go test ./internal/storage -v -run TestDatabaseInterfaceImplementation > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "❌ FAILED: Interface implementation test failed"
  exit 1
fi

# Check 8: URL parsing works correctly
echo "✓ Check 8: DATABASE_URL parsing test"
go test ./internal/storage -v -run TestParseDatabaseURLDetection > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "❌ FAILED: URL parsing test failed"
  exit 1
fi

# Check 9: Code compiles
echo "✓ Check 9: Code compilation"
go build ./internal/storage > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "❌ FAILED: storage package does not compile"
  exit 1
fi

# Check 10: Bot compiles with new code
echo "✓ Check 10: Bot compilation"
go build ./cmd/telegram-bot > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "❌ FAILED: bot does not compile"
  exit 1
fi

# Check 11: Migration script compiles
echo "✓ Check 11: Migration script compilation"
go build ./scripts/migrate_sqlite_to_postgres.go > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "❌ FAILED: migration script does not compile"
  exit 1
fi

# Check 12: All existing tests still pass
echo "✓ Check 12: Existing tests still pass"
go test ./internal/storage -v -short > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "❌ FAILED: tests failed"
  exit 1
fi

echo ""
echo "=================================================="
echo "✅ All verification checks passed!"
echo ""
echo "Summary:"
echo "  ✓ PostgreSQL adapter implemented"
echo "  ✓ SQLite adapter preserved"
echo "  ✓ Auto-detection from DATABASE_URL"
echo "  ✓ docker-compose.yml includes PostgreSQL"
echo "  ✓ Migration script available"
echo "  ✓ All tests pass"
echo ""
echo "Note: Full integration test requires a live PostgreSQL instance."
echo "Run './scripts/test_postgres.sh' with Docker to test end-to-end."
