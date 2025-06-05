#!/bin/bash

# Enhanced Test Runner with Configuration Management
set -euo pipefail

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m'

# Test configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
readonly TEST_CONFIG_FILE="$PROJECT_ROOT/.test-config.yaml"

# Default test settings
declare -A TEST_CONFIG=(
    [unit_tests]=true
    [integration_tests]=false
    [e2e_tests]=false
    [benchmark_tests]=false
    [coverage_threshold]=80
    [timeout]=300
    [parallel]=true
    [verbose]=false
    [fail_fast]=false
)

print_header() {
    echo -e "\n${PURPLE}$1${NC}"
    echo -e "${PURPLE}$(printf '=%.0s' $(seq 1 ${#1}))${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Load test configuration
load_test_config() {
    if [[ -f "$TEST_CONFIG_FILE" ]]; then
        print_info "Loading test configuration from $TEST_CONFIG_FILE"
        # Simple YAML parsing for our use case
        while IFS=': ' read -r key value; do
            if [[ -n "$key" && ! "$key" =~ ^[[:space:]]*# ]]; then
                key=$(echo "$key" | tr -d ' ')
                value=$(echo "$value" | tr -d ' ')
                TEST_CONFIG["$key"]="$value"
            fi
        done < "$TEST_CONFIG_FILE"
    else
        print_info "Using default test configuration"
        create_default_config
    fi
}

# Create default test configuration file
create_default_config() {
    cat > "$TEST_CONFIG_FILE" << EOF
# Jumpstart CLI Test Configuration
# Edit this file to customize test execution

# Test Types
unit_tests: true
integration_tests: false
e2e_tests: false
benchmark_tests: false

# Coverage
coverage_threshold: 80

# Execution
timeout: 300
parallel: true
verbose: false
fail_fast: false

# Test Tags (space-separated)
# tags: "unit integration"

# Environment Variables
# AZURE_SUBSCRIPTION_ID: "your-subscription-id"
# INTEGRATION_TESTS: "true"
EOF
    print_success "Created default test configuration: $TEST_CONFIG_FILE"
}

# Run tests based on configuration
run_tests() {
    local test_args=()
    local exit_code=0
    
    print_header "🧪 Jumpstart CLI Test Suite"
    
    # Build test arguments based on configuration
    if [[ "${TEST_CONFIG[verbose]}" == "true" ]]; then
        test_args+=("-v")
    fi
    
    if [[ "${TEST_CONFIG[parallel]}" == "true" ]]; then
        test_args+=("-parallel" "4")
    fi
    
    if [[ "${TEST_CONFIG[fail_fast]}" == "true" ]]; then
        test_args+=("-failfast")
    fi
    
    # Set timeout
    test_args+=("-timeout" "${TEST_CONFIG[timeout]}s")
    
    # Export environment variables for integration tests
    if [[ "${TEST_CONFIG[integration_tests]}" == "true" ]]; then
        export INTEGRATION_TESTS=true
        print_info "Integration tests enabled"
    fi
    
    if [[ "${TEST_CONFIG[e2e_tests]}" == "true" ]]; then
        export E2E_TESTS=true
        print_info "E2E tests enabled"
    fi
    
    # Run unit tests
    if [[ "${TEST_CONFIG[unit_tests]}" == "true" ]]; then
        print_header "Unit Tests with Coverage"
        # Create coverage directory
        mkdir -p coverage
        
        if ! go test "${test_args[@]}" -coverprofile=coverage/coverage.out -covermode=atomic ./...; then
            exit_code=1
        else
            # Generate coverage reports only if tests passed
            print_header "Coverage Analysis"
            
            # Generate HTML coverage report
            go tool cover -html=coverage/coverage.out -o coverage/coverage.html
            echo -e "${GREEN}✅ HTML report: coverage/coverage.html${NC}"
            
            # Generate function-level coverage report
            go tool cover -func=coverage/coverage.out > coverage/coverage_func.txt
            echo -e "${GREEN}✅ Function report: coverage/coverage_func.txt${NC}"
            
            # Extract overall coverage percentage
            TOTAL_COVERAGE=$(go tool cover -func=coverage/coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//')
            echo -e "\n${GREEN}✅ Total Coverage: $TOTAL_COVERAGE%${NC}"
        fi
    fi
    
    # Run benchmarks
    if [[ "${TEST_CONFIG[benchmark_tests]}" == "true" ]]; then
        print_header "Benchmark Tests"
        if ! go test -bench=. -benchmem ./...; then
            exit_code=1
        fi
    fi
    
    return $exit_code
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --unit)
                TEST_CONFIG[unit_tests]=true
                shift
                ;;
            --no-unit)
                TEST_CONFIG[unit_tests]=false
                shift
                ;;
            --integration)
                TEST_CONFIG[integration_tests]=true
                shift
                ;;
            --e2e)
                TEST_CONFIG[e2e_tests]=true
                shift
                ;;
            --benchmark)
                TEST_CONFIG[benchmark_tests]=true
                shift
                ;;
            --coverage-threshold)
                TEST_CONFIG[coverage_threshold]="$2"
                shift 2
                ;;
            --timeout)
                TEST_CONFIG[timeout]="$2"
                shift 2
                ;;
            --verbose|-v)
                TEST_CONFIG[verbose]=true
                shift
                ;;
            --fail-fast)
                TEST_CONFIG[fail_fast]=true
                shift
                ;;
            --config)
                TEST_CONFIG_FILE="$2"
                shift 2
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

show_help() {
    cat << EOF
Jumpstart CLI Test Runner

USAGE:
    $(basename "$0") [OPTIONS]

OPTIONS:
    --unit                Run unit tests (default: true)
    --no-unit            Skip unit tests
    --integration        Run integration tests
    --e2e                Run end-to-end tests
    --benchmark          Run benchmark tests
    --coverage-threshold Set coverage threshold (default: 80)
    --timeout SECONDS    Set test timeout (default: 300)
    --verbose, -v        Verbose output
    --fail-fast          Stop on first failure
    --config FILE        Use custom config file
    --help, -h           Show this help

EXAMPLES:
    $(basename "$0")                          # Run default tests
    $(basename "$0") --integration --e2e      # Run integration and E2E tests
    $(basename "$0") --verbose --fail-fast    # Verbose output, stop on failure
    $(basename "$0") --config .test-ci.yaml  # Use custom config

CONFIG FILE:
    Configuration is loaded from $TEST_CONFIG_FILE
    Use --config to specify a different file.
EOF
}

# Main execution
main() {
    cd "$PROJECT_ROOT"
    
    load_test_config
    parse_args "$@"
    
    print_info "Test configuration:"
    for key in "${!TEST_CONFIG[@]}"; do
        echo "  $key: ${TEST_CONFIG[$key]}"
    done
    
    if run_tests; then
        print_success "All tests completed successfully!"
        exit 0
    else
        print_error "Some tests failed!"
        exit 1
    fi
}

main "$@"
