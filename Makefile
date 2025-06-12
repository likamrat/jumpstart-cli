.PHONY: build build-all install clean lint fmt vet deps audit check dev info help
.PHONY: test test-unit test-integration test-coverage test-azure test-clean all

# Variables
BINARY_NAME=jumpstart-cli
BUILD_DIR=bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X jumpstartcli/internal/utils.CliVersion=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Default target
all: clean fmt vet test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Install binary to GOPATH/bin or /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@if [ -n "$(GOPATH)" ] && [ -d "$(GOPATH)/bin" ]; then \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/; \
		echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"; \
	else \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/; \
		echo "Installed to /usr/local/bin/$(BINARY_NAME)"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Update dependencies
deps:
	@echo "Updating dependencies..."
	go mod tidy
	go mod download

# Security audit
audit:
	@echo "Running security audit..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Run quick development checks
check: fmt vet lint
	@echo "All checks passed!"

# Development build (faster, no version info)
dev:
	@echo "Building development version..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Show build info
info:
	@echo "Build Information:"
	@echo "  Binary Name: $(BINARY_NAME)"
	@echo "  Version:     $(VERSION)"
	@echo "  Build Time:  $(BUILD_TIME)"
	@echo "  Git Commit:  $(GIT_COMMIT)"
	@echo "  Build Dir:   $(BUILD_DIR)"

# Run unit tests only
test-unit:
	go test -v ./... -short

# Run integration tests (requires Azure CLI and login)
test-integration:
	@echo "Running Azure CLI integration tests..."
	@echo "Prerequisites:"
	@echo "  - Azure CLI installed"
	@echo "  - Logged in (az login)"
	@echo "  - Active subscription"
	go test -v ./internal/utils -tags=integration -timeout=10m

# Run all tests including integration
test:
	go test -v ./... -tags=integration -timeout=10m

# Generate test coverage
test-coverage:
	go test -v ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.out (text) and coverage.html (HTML)"
	@echo "Total coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"

# View coverage report in terminal
test-coverage-report:
	go tool cover -func=coverage.out

# Open coverage in browser
test-coverage-html:
	go tool cover -html=coverage.out

# Test specifically Azure CLI functions with detailed output
test-azure:
	@echo "Testing Azure CLI integration..."
	go test -v ./internal/utils -tags=integration -run=".*Azure.*|.*Resource.*|.*Region.*" -timeout=10m

# Clean test artifacts
test-clean:
	rm -f coverage.out coverage.html
	go clean -testcache

# Clean build artifacts and test cache
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean
	go clean -testcache

# Show help
help:
	@echo "Jumpstart CLI Makefile Commands:"
	@echo ""
	@echo "Development:"
	@echo "  build          - Build the binary with version info"
	@echo "  dev            - Quick development build (no version info)"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  install        - Install binary to system PATH"
	@echo "  clean          - Clean build artifacts and test cache"
	@echo "  check          - Run fmt, vet, and lint"
	@echo "  fmt            - Format code"
	@echo "  vet            - Vet code for issues"
	@echo "  lint           - Lint code (requires golangci-lint)"
	@echo "  deps           - Update and download dependencies"
	@echo "  audit          - Run security audit (requires govulncheck)"
	@echo "  info           - Show build information"
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run all tests including integration"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests"
	@echo "  test-coverage  - Generate test coverage report"
	@echo "  test-azure     - Test Azure CLI specific functions"
	@echo "  test-clean     - Clean test artifacts"
	@echo ""
	@echo "Meta:"
	@echo "  all            - Clean, format, vet, test, and build"
	@echo "  help           - Show this help message"
