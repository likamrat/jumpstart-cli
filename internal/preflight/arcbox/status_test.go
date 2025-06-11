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

func TestStatusCommandValidation(t *testing.T) {
	cmd := CreateStatusCommand()

	// Test basic command properties
	if cmd.Use != "status" {
		t.Errorf("Expected Use to be 'status', got '%s'", cmd.Use)
	}

	// Test that RunE function exists
	if cmd.RunE == nil {
		t.Error("Expected RunE function to be defined")
	}

	// Test command silence settings
	if !cmd.DisableSuggestions || !cmd.SilenceErrors || !cmd.SilenceUsage {
		t.Error("Expected command to have proper silence settings")
	}
}

func TestStatusCommandFlags(t *testing.T) {
	cmd := CreateStatusCommand()

	// Test all expected flags
	expectedFlags := map[string]struct {
		shorthand    string
		defaultValue string
	}{
		"verbose": {"v", "false"},
		"json":    {"j", "false"},
		"format":  {"f", "table"},
	}

	for flagName, expected := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be defined", flagName)
			continue
		}

		if flag.Shorthand != expected.shorthand {
			t.Errorf("Flag '%s': expected shorthand '%s', got '%s'",
				flagName, expected.shorthand, flag.Shorthand)
		}

		if flag.DefValue != expected.defaultValue {
			t.Errorf("Flag '%s': expected default value '%s', got '%s'",
				flagName, expected.defaultValue, flag.DefValue)
		}
	}
}

func TestStatusCheckResultValidation(t *testing.T) {
	testCases := []struct {
		name     string
		result   StatusCheckResult
		expected string
	}{
		{
			name: "Success status",
			result: StatusCheckResult{
				CheckType: "Test",
				Status:    "Success",
				Success:   true,
			},
			expected: "Success",
		},
		{
			name: "Failed status",
			result: StatusCheckResult{
				CheckType:    "Test",
				Status:       "Failed",
				Success:      false,
				ErrorMessage: "Test error",
			},
			expected: "Failed",
		},
		{
			name: "Warning status",
			result: StatusCheckResult{
				CheckType: "Test",
				Status:    "Warning",
				Success:   false,
			},
			expected: "Warning",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.result.Status != tc.expected {
				t.Errorf("Expected status '%s', got '%s'", tc.expected, tc.result.Status)
			}
		})
	}
}

func TestStatusErrorScenarios(t *testing.T) {
	// Test with nil results
	var nilResults []StatusCheckResult

	// This should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("DisplayStatusResults panicked with nil results: %v", r)
			}
		}()
		DisplayStatusResults(nilResults)
	}()

	// Test JSON output with empty results
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("DisplayStatusAsJSON panicked with empty results: %v", r)
			}
		}()
		DisplayStatusAsJSON([]StatusCheckResult{})
	}()

	// Test YAML output with empty results
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("DisplayStatusAsYAML panicked with empty results: %v", r)
			}
		}()
		DisplayStatusAsYAML([]StatusCheckResult{})
	}()
}

func TestShowPreflightStatus(t *testing.T) {
	// Test that ShowPreflightStatus doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowPreflightStatus panicked: %v", r)
		}
	}()

	ShowPreflightStatus()
}

func TestStatusCommandWithDifferentFormats(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "Default format",
			args: []string{},
		},
		{
			name: "JSON format",
			args: []string{"--json"},
		},
		{
			name: "YAML format",
			args: []string{"--format", "yaml"},
		},
		{
			name: "Verbose mode",
			args: []string{"--verbose"},
		},
		{
			name: "JSON with verbose",
			args: []string{"--json", "--verbose"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := CreateStatusCommand()
			err := runCommandSilently(cmd, tc.args)
			if err != nil {
				t.Errorf("Command failed with args %v: %v", tc.args, err)
			}
		})
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
