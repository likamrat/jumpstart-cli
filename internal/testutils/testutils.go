package testutils

import (
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Color functions for consistent test output across all packages
var (
	TestSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	TestInfoColor    = color.New(color.FgCyan).SprintFunc()
	TestWarnColor    = color.New(color.FgYellow).SprintFunc()
	TestErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	TestHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

// PrintTestStatus provides consistent test result formatting across all packages
func PrintTestStatus(t *testing.T, testName string, success bool, message string) {
	status := TestSuccessColor("✅")
	if !success {
		status = TestErrorColor("❌")
		t.Error(message)
	}
	fmt.Printf("  %s %s: %s\n", status, TestInfoColor(testName), message)
}

// PrintTestHeader provides consistent header formatting
func PrintTestHeader(title string) {
	fmt.Printf("\n%s\n", TestHeaderColor(title))
}

// PrintTestSubHeader provides consistent sub-section formatting
func PrintTestSubHeader(title string) {
	fmt.Printf("  %s %s\n", TestInfoColor("→"), TestInfoColor(title))
}

// TableTest represents a table-driven test case with common fields
type TableTest struct {
	Name      string
	Input     interface{}
	Expected  interface{}
	ShouldErr bool
	Desc      string
}

// RunTableTests executes a slice of table tests with consistent formatting
func RunTableTests[T any](t *testing.T, tests []TableTest, testFunc func(input interface{}) (T, error)) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result, err := testFunc(tt.Input)
			
			if tt.ShouldErr && err == nil {
				PrintTestStatus(t, tt.Name, false, "Expected error but got none")
				return
			}
			
			if !tt.ShouldErr && err != nil {
				PrintTestStatus(t, tt.Name, false, fmt.Sprintf("Unexpected error: %v", err))
				return
			}
			
			// Type-safe comparison would require reflection or generics constraints
			// For now, use interface{} comparison
			if fmt.Sprintf("%v", result) == fmt.Sprintf("%v", tt.Expected) {
				PrintTestStatus(t, tt.Name, true, tt.Desc)
			} else {
				PrintTestStatus(t, tt.Name, false, fmt.Sprintf("Expected %v, got %v", tt.Expected, result))
			}
		})
	}
}

// MockCommand creates a mock cobra.Command for testing
func MockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}
	return cmd
}
