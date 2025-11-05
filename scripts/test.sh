#!/bin/bash

# Test runner script for Movie DB API
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Set required environment variables
export JWT_SECRET=${JWT_SECRET:-test-secret-key-for-testing}

echo -e "${GREEN}Running Movie DB API Tests${NC}"
echo "======================================"

# Check if Redis is running (for rate limiter tests)
if redis-cli ping >/dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Redis is available - all tests will run"
else
    echo -e "${YELLOW}⚠${NC} Redis is not available - rate limiter tests may be skipped"
fi

echo ""

# Run tests based on argument
case "${1:-all}" in
    all)
        echo "Running all tests..."
        go test ./internal/... -v -count=1
        ;;

    unit)
        echo "Running unit tests..."
        go test ./internal/movies -v -run TestValidateMovie -count=1
        go test ./internal/auth -v -run TestJWT -count=1
        go test ./internal/middleware -v -run TestGetRateLimiterConfig -count=1
        ;;

    integration)
        echo "Running integration tests..."
        go test ./internal/movies -v -run "TestCreate|TestGet|TestUpdate|TestDelete|TestList" -count=1
        go test ./internal/auth -v -run "TestLogin|TestAuthFlow" -count=1
        go test ./internal/middleware -v -run TestRateLimiter -count=1
        ;;

    coverage)
        echo "Running tests with coverage..."
        go test ./internal/... -coverprofile=coverage.out -covermode=atomic -count=1
        echo ""
        echo "Coverage Summary:"
        go tool cover -func=coverage.out | grep total
        echo ""
        echo "HTML coverage report: coverage.html"
        go tool cover -html=coverage.out -o coverage.html
        ;;

    quick)
        echo "Running quick test (no verbose)..."
        go test ./internal/... -count=1
        ;;

    *)
        echo "Usage: $0 {all|unit|integration|coverage|quick}"
        echo ""
        echo "  all         - Run all tests (default)"
        echo "  unit        - Run only unit tests"
        echo "  integration - Run only integration tests"
        echo "  coverage    - Run tests with coverage report"
        echo "  quick       - Run all tests without verbose output"
        exit 1
        ;;
esac

# Check exit code
if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Tests failed!${NC}"
    exit 1
fi
