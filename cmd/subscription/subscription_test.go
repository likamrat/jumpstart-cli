package subscription

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/jumpstart-cli/internal/testutils"
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

func TestIsValidGUID(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing GUID Validation ==="))

	tests := []struct {
		name string
		guid string
		want bool
	}{
		{
			name: "valid GUID",
			guid: "12345678-1234-1234-1234-123456789012",
			want: true,
		},
		{
			name: "valid GUID with uppercase",
			guid: "12345678-1234-1234-1234-123456789ABC",
			want: true,
		},
		{
			name: "valid GUID with mixed case",
			guid: "12345678-1234-1234-1234-123456789AbC",
			want: true,
		},
		{
			name: "invalid GUID - too short",
			guid: "12345678-1234-1234-1234-12345678901",
			want: false,
		},
		{
			name: "invalid GUID - too long",
			guid: "12345678-1234-1234-1234-1234567890123",
			want: false,
		},
		{
			name: "invalid GUID - missing dashes",
			guid: "12345678123412341234123456789012",
			want: false,
		},
		{
			name: "invalid GUID - wrong dash positions",
			guid: "123456781-234-1234-1234-123456789012",
			want: false,
		},
		{
			name: "invalid GUID - invalid characters",
			guid: "12345678-1234-1234-1234-12345678901G",
			want: false,
		},
		{
			name: "invalid GUID - empty string",
			guid: "",
			want: false,
		},
		{
			name: "invalid GUID - contains spaces",
			guid: "12345678-1234-1234-1234-12345678901 ",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testName := fmt.Sprintf("GUID Validation: %s", tt.name)
			got := isValidGUID(tt.guid)
			success := got == tt.want
			var message string
			if success {
				message = fmt.Sprintf("GUID '%s' correctly validated as %t", tt.guid, tt.want)
			} else {
				message = fmt.Sprintf("isValidGUID(%q) = %v, want %v", tt.guid, got, tt.want)
			}
			printTestStatus(t, testName, success, message)
		})
	}
}

func TestNewSubscriptionCmd(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing New Subscription Command ==="))

	cmd := NewSubscriptionCmd()

	// Test basic command structure
	testName := "Command Use Field"
	success := cmd.Use == "subscription"
	message := fmt.Sprintf("Expected 'subscription', got '%s'", cmd.Use)
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

	// Test that subcommands are registered
	expectedSubcommands := []string{"list", "set", "show"}
	actualSubcommands := make([]string, 0)

	for _, subCmd := range cmd.Commands() {
		actualSubcommands = append(actualSubcommands, subCmd.Use)
	}

	for _, expected := range expectedSubcommands {
		testName = fmt.Sprintf("Subcommand: %s", expected)
		found := false
		for _, actual := range actualSubcommands {
			if actual == expected {
				found = true
				break
			}
		}
		if found {
			message = fmt.Sprintf("Subcommand '%s' found correctly", expected)
		} else {
			message = fmt.Sprintf("Expected subcommand '%s' not found", expected)
		}
		printTestStatus(t, testName, found, message)
	}
}

func TestSubscriptionSetCommandFlags(t *testing.T) {
	cmd := NewSubscriptionCmd()

	// Find the set subcommand
	var setCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "set" {
			setCmd = subCmd
			break
		}
	}

	if setCmd == nil {
		t.Fatal("set subcommand not found")
	}

	// Test that required flags exist
	requiredFlags := []string{"subscription", "name"}
	for _, flagName := range requiredFlags {
		flag := setCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Required flag %q not found in set subcommand", flagName)
		}
	}
}

func TestSubscriptionShowCommandFlags(t *testing.T) {
	cmd := NewSubscriptionCmd()

	// Find the show subcommand
	var showCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "show" {
			showCmd = subCmd
			break
		}
	}

	if showCmd == nil {
		t.Fatal("show subcommand not found")
	}

	// Test that optional flags exist
	optionalFlags := []string{"id", "name"}
	for _, flagName := range optionalFlags {
		flag := showCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Optional flag %q not found in show subcommand", flagName)
		}
	}
}

// Add tests for command execution
func TestSubscriptionCommandExecution(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Subscription Command Execution ===")

	t.Run("list_command_output", func(t *testing.T) {
		cmd := NewSubscriptionCmd()
		listCmd := findSubcommand(cmd, "list")

		// Capture output
		buf := new(bytes.Buffer)
		listCmd.SetOut(buf)
		listCmd.SetErr(buf)

		// Execute (this might fail without Azure CLI, but shouldn't panic)
		err := listCmd.Execute()
		output := buf.String()

		// Check that it attempted to run
		if err != nil {
			testutils.PrintTestStatus(t, "List execution",
				strings.Contains(err.Error(), "az") || strings.Contains(output, "az"),
				"Should attempt to run Azure CLI")
		} else {
			testutils.PrintTestStatus(t, "List execution", true,
				"List command executed successfully")
		}
	})
}

// Add tests for GUID validation edge cases
func TestGUIDValidationEdgeCases(t *testing.T) {
	testutils.PrintTestHeader("=== Testing GUID Validation Edge Cases ===")

	edgeCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"all zeros", "00000000-0000-0000-0000-000000000000", true},
		{"all Fs", "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF", true},
		{"mixed case", "AbCdEfGh-1234-5678-9012-aBcDeFgHiJkL", false}, // too long
		{"unicode", "12345678-1234-1234-1234-12345678901中", false},
		{"with newline", "12345678-1234-1234-1234-123456789012\n", false},
		{"with tab", "12345678-1234-1234-1234-123456789012\t", false},
		{"URL encoded", "12345678%2D1234%2D1234%2D1234%2D123456789012", false},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidGUID(tc.input)
			testutils.PrintTestStatus(t, tc.name, result == tc.expected,
				fmt.Sprintf("GUID '%s' validation: expected %v, got %v", tc.input, tc.expected, result))
		})
	}
}

// Add tests for command structure validation
func TestCommandStructureValidation(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Command Structure ===")

	cmd := NewSubscriptionCmd()

	t.Run("command_aliases", func(t *testing.T) {
		// Check if aliases are set correctly
		testutils.PrintTestStatus(t, "Command aliases",
			len(cmd.Aliases) > 0 && contains(cmd.Aliases, "sub"),
			"Should have 'sub' as an alias")
	})

	t.Run("command_examples", func(t *testing.T) {
		// Check if examples are provided
		hasExamples := cmd.Example != "" ||
			(findSubcommand(cmd, "list") != nil && findSubcommand(cmd, "list").Example != "") ||
			(findSubcommand(cmd, "set") != nil && findSubcommand(cmd, "set").Example != "") ||
			(findSubcommand(cmd, "show") != nil && findSubcommand(cmd, "show").Example != "")

		testutils.PrintTestStatus(t, "Command examples", hasExamples,
			"Should have examples for at least one command")
	})

	t.Run("required_flags_marked", func(t *testing.T) {
		setCmd := findSubcommand(cmd, "set")
		if setCmd != nil {
			// Check if required flags are marked
			subFlag := setCmd.Flags().Lookup("subscription")
			nameFlag := setCmd.Flags().Lookup("name")

			testutils.PrintTestStatus(t, "Required flags",
				subFlag != nil && nameFlag != nil,
				"Set command should have required flags")
		}
	})
}

// Helper functions
func findSubcommand(cmd *cobra.Command, use string) *cobra.Command {
	for _, sub := range cmd.Commands() {
		if sub.Use == use {
			return sub
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkGUIDValidation(b *testing.B) {
	validGUID := "12345678-1234-1234-1234-123456789012"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isValidGUID(validGUID)
	}
}
