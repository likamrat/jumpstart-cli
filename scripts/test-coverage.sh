#!/bin/bash

# Test Coverage Analysis Script for Jumpstart CLI
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
COVERAGE_DIR="coverage"
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"
MIN_COVERAGE=80

echo -e "${PURPLE}🧪 Jumpstart CLI Test Coverage Analysis${NC}"
echo -e "${CYAN}======================================${NC}"

# Create coverage directory
mkdir -p "$COVERAGE_DIR"

echo -e "\n${BLUE}📊 Running test suite with coverage...${NC}"

# Run tests with coverage
go test -v -coverprofile="$COVERAGE_DIR/$COVERAGE_FILE" -covermode=atomic ./... 2>&1 | tee "$COVERAGE_DIR/test_output.log"

if [ ${PIPESTATUS[0]} -ne 0 ]; then
    echo -e "${RED}❌ Tests failed. Coverage analysis aborted.${NC}"
    exit 1
fi

echo -e "\n${BLUE}📈 Generating coverage reports...${NC}"

# Generate HTML coverage report
go tool cover -html="$COVERAGE_DIR/$COVERAGE_FILE" -o "$COVERAGE_DIR/$COVERAGE_HTML"
echo -e "${GREEN}✅ HTML report: $COVERAGE_DIR/$COVERAGE_HTML${NC}"

# Generate function-level coverage report
go tool cover -func="$COVERAGE_DIR/$COVERAGE_FILE" > "$COVERAGE_DIR/coverage_func.txt"
echo -e "${GREEN}✅ Function report: $COVERAGE_DIR/coverage_func.txt${NC}"

# Extract overall coverage percentage
TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_DIR/$COVERAGE_FILE" | grep "total:" | awk '{print $3}' | sed 's/%//')

echo -e "\n${PURPLE}📋 Coverage Summary${NC}"
echo -e "${CYAN}==================${NC}"

echo -e "${GREEN}✅ Total Coverage: $TOTAL_COVERAGE%${NC}"

echo -e "\n${GREEN}🚀 Coverage analysis completed successfully!${NC}"
echo -e "${CYAN}View HTML report: open $COVERAGE_DIR/$COVERAGE_HTML${NC}"
