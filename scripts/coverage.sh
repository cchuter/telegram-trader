#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Coverage threshold
THRESHOLD=70

echo "======================================="
echo "Running test coverage analysis"
echo "======================================="

# Create coverage directory if it doesn't exist
mkdir -p coverage

# Go coverage
echo ""
echo "${YELLOW}Running Go tests with coverage...${NC}"
cd "$(dirname "$0")/.."
# Run tests and capture coverage even if some tests fail
go test -coverprofile=coverage/coverage.out ./... -v || true

echo ""
echo "${YELLOW}Generating Go HTML coverage report...${NC}"
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# Extract Go coverage percentage
GO_COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "${GREEN}Go coverage: ${GO_COVERAGE}%${NC}"

# TypeScript coverage
echo ""
echo "${YELLOW}Running TypeScript tests with coverage...${NC}"
cd galachain-service

# Run tests and capture output
TS_OUTPUT=$(npm run test -- --coverage --coverageDirectory=../coverage/typescript 2>&1)
echo "$TS_OUTPUT"

# Extract TypeScript coverage percentage from Jest summary output
cd ..
TS_COVERAGE=$(echo "$TS_OUTPUT" | grep "All files" | awk '{print $4}' || echo "0")

if [ -z "$TS_COVERAGE" ]; then
    echo "${RED}Could not extract TypeScript coverage${NC}"
    TS_COVERAGE=0
fi

echo "${GREEN}TypeScript coverage: ${TS_COVERAGE}%${NC}"

# Calculate combined coverage (weighted average based on lines of code)
# Simple approach: average the two
COMBINED_COVERAGE=$(echo "scale=2; ($GO_COVERAGE + $TS_COVERAGE) / 2" | bc)
echo ""
echo "======================================="
echo "${GREEN}Combined coverage: ${COMBINED_COVERAGE}%${NC}"
echo "======================================="

# Check threshold for Go
if (( $(echo "$GO_COVERAGE < $THRESHOLD" | bc -l) )); then
    echo ""
    echo "${RED}ERROR: Go coverage ${GO_COVERAGE}% is below threshold ${THRESHOLD}%${NC}"
    GO_FAIL=1
else
    echo "${GREEN}Go coverage meets threshold${NC}"
    GO_FAIL=0
fi

# Check threshold for TypeScript
if (( $(echo "$TS_COVERAGE < $THRESHOLD" | bc -l) )); then
    echo ""
    echo "${RED}ERROR: TypeScript coverage ${TS_COVERAGE}% is below threshold ${THRESHOLD}%${NC}"
    TS_FAIL=1
else
    echo "${GREEN}TypeScript coverage meets threshold${NC}"
    TS_FAIL=0
fi

echo ""
echo "Coverage reports generated:"
echo "  - Go: coverage/coverage.html"
echo "  - TypeScript: coverage/typescript/lcov-report/index.html"

# Exit with error if either failed
if [ $GO_FAIL -eq 1 ] || [ $TS_FAIL -eq 1 ]; then
    exit 1
fi

echo ""
echo "${GREEN}All coverage checks passed!${NC}"
exit 0
