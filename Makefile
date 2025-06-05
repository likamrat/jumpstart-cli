.PHONY: test test-unit test-integration test-coverage test-azure clean

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
