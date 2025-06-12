package testutils

import (
	"context"
	"os"
	"testing"
	"time"
)

// IntegrationTest represents a complete integration test scenario
type IntegrationTest struct {
	Name       string
	Setup      func(t *testing.T) error
	Execute    func(t *testing.T) error
	Cleanup    func(t *testing.T) error
	Timeout    time.Duration
	SkipReason string
}

// RunIntegrationTests executes integration tests with proper setup/cleanup
func RunIntegrationTests(t *testing.T, tests []IntegrationTest) {
	PrintTestHeader("=== Running Integration Tests ===")

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.SkipReason != "" {
				t.Skipf("Skipping %s: %s", test.Name, test.SkipReason)
				return
			}

			// Set timeout
			timeout := test.Timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Setup phase
			if test.Setup != nil {
				PrintTestSubHeader("Setup: " + test.Name)
				if err := test.Setup(t); err != nil {
					PrintTestStatus(t, "Setup", false, err.Error())
					return
				}
			}

			// Ensure cleanup runs even if test fails
			defer func() {
				if test.Cleanup != nil {
					PrintTestSubHeader("Cleanup: " + test.Name)
					if err := test.Cleanup(t); err != nil {
						PrintTestStatus(t, "Cleanup", false, err.Error())
					}
				}
			}()
			// Execute test
			PrintTestSubHeader("Execute: " + test.Name)
			done := make(chan error, 1)
			go func() {
				done <- test.Execute(t)
			}()

			select {
			case err := <-done:
				if err != nil {
					PrintTestStatus(t, test.Name, false, err.Error())
				} else {
					PrintTestStatus(t, test.Name, true, "Integration test passed")
				}
			case <-ctx.Done():
				PrintTestStatus(t, test.Name, false, "Test timed out")
				t.Fatal("Test timed out")
			}
		})
	}
}

// Example integration test for CLI commands
func TestCLIIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests disabled. Set INTEGRATION_TESTS=true to run.")
	}

	tests := []IntegrationTest{
		{
			Name: "version_command",
			Execute: func(t *testing.T) error {
				// Example: test version command
				// This would use actual CLI execution
				return nil
			},
			Timeout: 10 * time.Second,
		},
		{
			Name: "upgrade_dry_run",
			Execute: func(t *testing.T) error {
				// Example: test upgrade --dry-run
				return nil
			},
			Timeout: 30 * time.Second,
		},
	}

	RunIntegrationTests(t, tests)
}
