package repo

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Color functions for test output
var (
	testSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	testInfoColor    = color.New(color.FgCyan).SprintFunc()
	testWarnColor    = color.New(color.FgYellow).SprintFunc()
	testErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	testHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

// Helper function to print colored test status
func printTestStatus(t *testing.T, testName string, success bool, message string) {
	var icon string
	var colorFunc func(a ...interface{}) string

	if success {
		icon = "✅"
		colorFunc = testSuccessColor
	} else {
		icon = "❌"
		colorFunc = testErrorColor
		t.Errorf("Test failed: %s", message)
	}

	fmt.Printf("%s %s: %s\n", colorFunc("PASS"), icon, testInfoColor(fmt.Sprintf("%s: %s", testName, message)))
}

func TestNewRepoCmd(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing New Repo Command ==="))

	cmd := NewRepoCmd()

	// Test basic command structure
	testName := "Command Use Field"
	success := cmd.Use == "repo"
	message := fmt.Sprintf("Expected 'repo', got '%s'", cmd.Use)
	printTestStatus(t, testName, success, message)

	// Test Short description
	testName = "Short Description"
	success = cmd.Short != ""
	if success {
		message = "Short description is properly set"
	} else {
		message = "Short description should not be empty"
	}
	printTestStatus(t, testName, success, message)

	// Test Long description
	testName = "Long Description"
	success = cmd.Long != ""
	if success {
		message = "Long description is properly set"
	} else {
		message = "Long description should not be empty"
	}
	printTestStatus(t, testName, success, message)

	// Test that RunE function exists
	testName = "RunE Function"
	success = cmd.RunE != nil
	if success {
		message = "RunE function is properly set"
	} else {
		message = "RunE function should not be nil"
	}
	printTestStatus(t, testName, success, message)

	// Test command configuration
	testName = "Disable Suggestions"
	success = cmd.DisableSuggestions
	if success {
		message = "DisableSuggestions correctly set to true"
	} else {
		message = "DisableSuggestions should be true"
	}
	printTestStatus(t, testName, success, message)

	testName = "Silence Errors"
	success = cmd.SilenceErrors
	if success {
		message = "SilenceErrors correctly set to true"
	} else {
		message = "SilenceErrors should be true"
	}
	printTestStatus(t, testName, success, message)

	testName = "Silence Usage"
	success = cmd.SilenceUsage
	if success {
		message = "SilenceUsage correctly set to true"
	} else {
		message = "SilenceUsage should be true"
	}
	printTestStatus(t, testName, success, message)
}

func TestRepoCommandSubcommands(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Repo Command Subcommands ==="))

	cmd := NewRepoCmd()
	// Test that subcommands are registered
	expectedSubcommands := []string{"clone", "update", "delete"}
	actualSubcommands := make([]string, 0)

	for _, subCmd := range cmd.Commands() {
		actualSubcommands = append(actualSubcommands, subCmd.Use)
	}

	for _, expected := range expectedSubcommands {
		testName := fmt.Sprintf("Subcommand: %s", expected)
		found := false
		for _, actual := range actualSubcommands {
			if actual == expected {
				found = true
				break
			}
		}
		var message string
		if found {
			message = fmt.Sprintf("Subcommand '%s' found correctly", expected)
		} else {
			message = fmt.Sprintf("Expected subcommand '%s' not found", expected)
		}
		printTestStatus(t, testName, found, message)
	}
}

func TestRepoCommandInvalidSubcommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "invalid subcommand",
			args:        []string{"invalid"},
			expectError: true,
		},
		{
			name:        "no arguments",
			args:        []string{},
			expectError: false, // Should show help
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRepoCmd()
			cmd.SetArgs(tt.args)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRepoCloneCommandFlags(t *testing.T) {
	cmd := NewRepoCmd()

	// Find the clone subcommand
	var cloneCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "clone" {
			cloneCmd = subCmd
			break
		}
	}

	if cloneCmd == nil {
		t.Fatal("clone subcommand not found")
	}

	// Test that path flag exists
	pathFlag := cloneCmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Error("clone subcommand should have 'path' flag")
	} // Test shorthand flag
	pFlag := cloneCmd.Flags().ShorthandLookup("p")
	if pFlag == nil {
		t.Error("clone subcommand should have 'p' shorthand flag")
	}
}

func TestRepoUpdateCommandFlags(t *testing.T) {
	cmd := NewRepoCmd()

	// Find the update subcommand
	var updateCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "update" {
			updateCmd = subCmd
			break
		}
	}

	if updateCmd == nil {
		t.Fatal("update subcommand not found")
	}

	// Test that path flag exists
	pathFlag := updateCmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Error("update subcommand should have 'path' flag")
	}
	// Test shorthand flag
	pFlag := updateCmd.Flags().ShorthandLookup("p")
	if pFlag == nil {
		t.Error("update subcommand should have 'p' shorthand flag")
	}
}

func TestRepoDeleteCommandFlags(t *testing.T) {
	cmd := NewRepoCmd()

	// Find the delete subcommand
	var deleteCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "delete" {
			deleteCmd = subCmd
			break
		}
	}

	if deleteCmd == nil {
		t.Fatal("delete subcommand not found")
	}

	// Test that path flag exists
	pathFlag := deleteCmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Error("delete subcommand should have 'path' flag")
	}
	// Test shorthand flag
	pFlag := deleteCmd.Flags().ShorthandLookup("p")
	if pFlag == nil {
		t.Error("delete subcommand should have 'p' shorthand flag")
	}
}

func TestRepoCommandStructure(t *testing.T) {
	cmd := NewRepoCmd()

	// Verify command properties
	expectedProperties := map[string]bool{
		"DisableSuggestions": true,
		"SilenceErrors":      true,
		"SilenceUsage":       true,
	}

	actualProperties := map[string]bool{
		"DisableSuggestions": cmd.DisableSuggestions,
		"SilenceErrors":      cmd.SilenceErrors,
		"SilenceUsage":       cmd.SilenceUsage,
	}

	for property, expected := range expectedProperties {
		if actual := actualProperties[property]; actual != expected {
			t.Errorf("Command.%s = %v, want %v", property, actual, expected)
		}
	}
}

func TestRepoSubcommandStructure(t *testing.T) {
	cmd := NewRepoCmd()
	subcommands := cmd.Commands()

	// Test each subcommand has required properties
	for _, subCmd := range subcommands {
		t.Run("subcommand_"+subCmd.Use, func(t *testing.T) {
			if subCmd.Use == "" {
				t.Error("Subcommand Use should not be empty")
			}

			if subCmd.Short == "" {
				t.Error("Subcommand Short should not be empty")
			}

			if subCmd.Long == "" {
				t.Error("Subcommand Long should not be empty")
			}

			if subCmd.Run == nil {
				t.Error("Subcommand Run should not be nil")
			}
		})
	}
}

func TestRepoCommandBasicFunctionality(t *testing.T) {
	cmd := NewRepoCmd()

	// Test that command has the expected structure
	if len(cmd.Commands()) != 3 {
		t.Errorf("Expected 3 subcommands, got %d", len(cmd.Commands()))
	}

	// Test that each subcommand exists and has basic properties
	subcommandNames := make(map[string]bool)
	for _, subCmd := range cmd.Commands() {
		subcommandNames[subCmd.Use] = true

		// Each subcommand should have basic properties
		if subCmd.Short == "" {
			t.Errorf("Subcommand %s should have a Short description", subCmd.Use)
		}
		if subCmd.Long == "" {
			t.Errorf("Subcommand %s should have a Long description", subCmd.Use)
		}
	}

	// Verify expected subcommands exist
	expectedSubcommands := []string{"clone", "update", "delete"}
	for _, expected := range expectedSubcommands {
		if !subcommandNames[expected] {
			t.Errorf("Expected subcommand %s not found", expected)
		}
	}
}
