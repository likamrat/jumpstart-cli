// status_test.go - Tests for ArcBox preflight status functionality
package arcbox

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestCreateStatusCommand(t *testing.T) {
	cmd := CreateStatusCommand()

	// Test basic command properties
	if cmd.Use != "status" {
		t.Errorf("Expected Use to be 'status', got '%s'", cmd.Use)
	}

	if cmd.Short != "Show last preflight check status" {
		t.Errorf("Expected Short description to match, got '%s'", cmd.Short)
	}

	// Test that Long description contains expected content
	expectedKeywords := []string{"preflight check", "Resource provider", "quota", "ArcBox deployment"}
	for _, keyword := range expectedKeywords {
		if !strings.Contains(cmd.Long, keyword) {
			t.Errorf("Expected Long description to contain '%s'", keyword)
		}
	}

	// Test flags
	verboseFlag := cmd.Flags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Expected 'verbose' flag to be defined")
	}
	if verboseFlag.Shorthand != "v" {
		t.Errorf("Expected verbose flag shorthand to be 'v', got '%s'", verboseFlag.Shorthand)
	}

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("Expected 'json' flag to be defined")
	}
	if jsonFlag.Shorthand != "j" {
		t.Errorf("Expected json flag shorthand to be 'j', got '%s'", jsonFlag.Shorthand)
	}

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Error("Expected 'format' flag to be defined")
	}
	if formatFlag.Shorthand != "f" {
		t.Errorf("Expected format flag shorthand to be 'f', got '%s'", formatFlag.Shorthand)
	}
	if formatFlag.DefValue != "table" {
		t.Errorf("Expected format flag default value to be 'table', got '%s'", formatFlag.DefValue)
	}

	// Test command configuration
	if !cmd.DisableSuggestions {
		t.Error("Expected DisableSuggestions to be true")
	}
	if !cmd.SilenceErrors {
		t.Error("Expected SilenceErrors to be true")
	}
	if !cmd.SilenceUsage {
		t.Error("Expected SilenceUsage to be true")
	}
}

func TestStatusCheckResult(t *testing.T) {
	now := time.Now()
	result := StatusCheckResult{
		Timestamp:    now,
		CheckType:    "Test Check",
		Status:       "Success",
		Details:      "Test details",
		Success:      true,
		ErrorMessage: "",
	}

	if result.Timestamp != now {
		t.Error("Timestamp should match assigned value")
	}
	if result.CheckType != "Test Check" {
		t.Error("CheckType should match assigned value")
	}
	if result.Status != "Success" {
		t.Error("Status should match assigned value")
	}
	if result.Details != "Test details" {
		t.Error("Details should match assigned value")
	}
	if !result.Success {
		t.Error("Success should be true")
	}
	if result.ErrorMessage != "" {
		t.Error("ErrorMessage should be empty")
	}
}

func TestGetLastPreflightStatus(t *testing.T) {
	results := GetLastPreflightStatus()

	if len(results) == 0 {
		t.Error("Expected at least one status result")
	}

	// Test that we get expected check types
	expectedCheckTypes := map[string]bool{
		"Resource Providers": false,
		"Azure Quota":        false,
		"Configuration":      false,
	}

	for _, result := range results {
		if _, exists := expectedCheckTypes[result.CheckType]; exists {
			expectedCheckTypes[result.CheckType] = true
		}

		// Test that timestamps are reasonable (within last day)
		if time.Since(result.Timestamp) > 24*time.Hour {
			t.Errorf("Timestamp for %s seems too old: %v", result.CheckType, result.Timestamp)
		}

		// Test that all required fields are populated
		if result.CheckType == "" {
			t.Error("CheckType should not be empty")
		}
		if result.Status == "" {
			t.Error("Status should not be empty")
		}
	}

	// Verify we got all expected check types
	for checkType, found := range expectedCheckTypes {
		if !found {
			t.Errorf("Expected to find check type '%s' in results", checkType)
		}
	}
}

func TestDisplayStatusResults_EmptyResults(t *testing.T) {
	// Test behavior with empty results slice
	var results []StatusCheckResult

	// This should not panic and should handle empty slice gracefully
	DisplayStatusResults(results)
	// If we reach here without panic, the test passes
}

func TestDisplayStatusResults_WithResults(t *testing.T) {
	results := []StatusCheckResult{
		{
			Timestamp: time.Now(),
			CheckType: "Test Check 1",
			Status:    "Success",
			Details:   "Test passed",
			Success:   true,
		},
		{
			Timestamp:    time.Now(),
			CheckType:    "Test Check 2",
			Status:       "Failed",
			Details:      "Test failed",
			Success:      false,
			ErrorMessage: "Something went wrong",
		},
		{
			Timestamp: time.Now(),
			CheckType: "Test Check 3",
			Status:    "Warning",
			Details:   "Test has warnings",
			Success:   false,
		},
	}

	// This should not panic and should handle various status types
	DisplayStatusResults(results)
	// If we reach here without panic, the test passes
}

func TestValidatePreflightEnvironment(t *testing.T) {
	// Test that the function returns a boolean
	result := ValidatePreflightEnvironment()

	// Currently returns true as placeholder
	if result != true {
		t.Error("Expected ValidatePreflightEnvironment to return true")
	}
}

func TestDisplayStatusAsJSON(t *testing.T) {
	results := []StatusCheckResult{
		{
			Timestamp: time.Now(),
			CheckType: "Test Check",
			Status:    "Success",
			Details:   "Test passed",
			Success:   true,
		},
	}

	// This should not panic
	DisplayStatusAsJSON(results)
	// If we reach here without panic, the test passes
}

func TestDisplayStatusAsYAML(t *testing.T) {
	results := []StatusCheckResult{
		{
			Timestamp: time.Now(),
			CheckType: "Test Check",
			Status:    "Success",
			Details:   "Test passed",
			Success:   true,
		},
	}

	// This should not panic
	DisplayStatusAsYAML(results)
	// If we reach here without panic, the test passes
}

// Test helper to run command without output
func runCommandSilently(cmd *cobra.Command, args []string) error {
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestStatusCommandExecution(t *testing.T) {
	cmd := CreateStatusCommand()

	// Test that command can be executed without arguments
	err := runCommandSilently(cmd, []string{})
	if err != nil {
		t.Errorf("Expected command to execute successfully, got error: %v", err)
	}

	// Test with verbose flag
	cmd = CreateStatusCommand() // Reset command
	err = runCommandSilently(cmd, []string{"--verbose"})
	if err != nil {
		t.Errorf("Expected command with --verbose to execute successfully, got error: %v", err)
	}

	// Test with json flag
	cmd = CreateStatusCommand() // Reset command
	err = runCommandSilently(cmd, []string{"--json"})
	if err != nil {
		t.Errorf("Expected command with --json to execute successfully, got error: %v", err)
	}

	// Test with format flag
	cmd = CreateStatusCommand() // Reset command
	err = runCommandSilently(cmd, []string{"--format", "yaml"})
	if err != nil {
		t.Errorf("Expected command with --format yaml to execute successfully, got error: %v", err)
	}
}

// Benchmark tests for performance
func BenchmarkGetLastPreflightStatus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetLastPreflightStatus()
	}
}

func BenchmarkCreateStatusCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CreateStatusCommand()
	}
}
