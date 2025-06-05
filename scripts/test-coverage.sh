#!/bin/bash

# Jumpstart CLI Test Coverage Script

set -e

echo "🧪 Running Jumpstart CLI Test Suite with Coverage"
echo "================================================"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_THRESHOLD=80
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"

# Clean previous coverage files
rm -f ${COVERAGE_FILE} ${COVERAGE_HTML}

# Run tests with coverage
echo -e "\n📊 Running tests with coverage..."
go test -v -race -coverprofile=${COVERAGE_FILE} -covermode=atomic ./...

# Generate HTML coverage report
echo -e "\n📄 Generating HTML coverage report..."
go tool cover -html=${COVERAGE_FILE} -o ${COVERAGE_HTML}

# Check coverage threshold
COVERAGE=$(go tool cover -func=${COVERAGE_FILE} | grep total | awk '{print $3}' | sed 's/%//')
COVERAGE_INT=${COVERAGE%.*}

echo -e "\n📈 Coverage Summary:"
echo "===================="
go tool cover -func=${COVERAGE_FILE} | grep -E "(total:|utils|resourceproviders|validator|arcbox|subscription)"

echo -e "\n🎯 Total Coverage: ${COVERAGE}%"

if [ ${COVERAGE_INT} -ge ${COVERAGE_THRESHOLD} ]; then
    echo -e "${GREEN}✅ Coverage threshold (${COVERAGE_THRESHOLD}%) met!${NC}"
    exit 0
else
    echo -e "${RED}❌ Coverage (${COVERAGE}%) is below threshold (${COVERAGE_THRESHOLD}%)${NC}"
    echo -e "${YELLOW}📝 Open ${COVERAGE_HTML} to see detailed coverage report${NC}"
    exit 1
fi

# Generate coverage report
echo "Running tests with coverage..."
go test -v -coverprofile=coverage.out -covermode=atomic ./...

# Generate HTML report
echo "Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

# Show summary
echo ""
echo "Coverage files generated:"
echo "  - coverage.out (text format)"
echo "  - coverage.html (HTML format)"
echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out | grep total

# Optional: Show uncovered lines
echo ""
echo "Files with low coverage:"
go tool cover -func=coverage.out | awk '$3 < 80.0 && $3 != "0.0%" {print $1 " - " $3}' | sort -t'-' -k2 -n
