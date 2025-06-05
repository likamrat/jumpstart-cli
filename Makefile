.PHONY: test test-coverage test-coverage-report test-coverage-html clean

# Run all tests
test:
	go test -v ./...

# Generate test coverage
test-coverage:
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.out (text) and coverage.html (HTML)"
	@echo "Total coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"

# View coverage report in terminal
test-coverage-report:
	go tool cover -func=coverage.out

# Open coverage in browser
test-coverage-html:
	go tool cover -html=coverage.out

# Clean coverage files
clean:
	rm -f coverage.out coverage.html
