package subscription

import (
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
