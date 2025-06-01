package version

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"
	"jumpstartcli/internal/utils"
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

func TestNewVersionCmd(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing New Version Command ==="))
	
	tests := []struct {
		name string
		want string
	}{
		{
			name: "version command creation",
			want: "version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewVersionCmd()
			
			testName := "Command Use Field"
			success := cmd.Use == tt.want
			message := fmt.Sprintf("Expected '%s', got '%s'", tt.want, cmd.Use)
			printTestStatus(t, testName, success, message)
			
			testName = "Short Description"
			success = cmd.Short != ""
			if success {
				message = "Short description is properly set"
			} else {
				message = "Short description should not be empty"
			}
			printTestStatus(t, testName, success, message)
			
			testName = "Long Description"
			success = cmd.Long != ""
			if success {
				message = "Long description is properly set"
			} else {
				message = "Long description should not be empty"
			}
			printTestStatus(t, testName, success, message)
		})
	}
}

func TestVersionCommandExecution(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Version Command Execution ==="))
	
	// Store original version
	originalVersion := utils.CliVersion
	
	tests := []struct {
		name        string
		version     string
		wantContains string
	}{
		{
			name:        "version output with test version",
			version:     "test-version-1.0.0",
			wantContains: "Jumpstart CLI version: test-version-1.0.0",
		},
		{
			name:        "version output with empty version",
			version:     "",
			wantContains: "Jumpstart CLI version:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testName := fmt.Sprintf("Version Output: %s", tt.name)
			
			// Set test version
			utils.CliVersion = tt.version
			
			// Capture output
			var buf bytes.Buffer
			cmd := NewVersionCmd()
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			
			// Execute command
			err := cmd.Execute()
			success := err == nil
			var message string
			
			if !success {
				message = fmt.Sprintf("Version command execution failed: %v", err)
			} else {
				output := buf.String()
				contains := strings.Contains(output, tt.wantContains)
				if contains {
					message = fmt.Sprintf("Output correctly contains '%s'", tt.wantContains)
				} else {
					message = fmt.Sprintf("Output = %q, want to contain %q", output, tt.wantContains)
					success = false
				}
			}
			
			printTestStatus(t, testName, success, message)
		})
	}
	
	// Restore original version
	utils.CliVersion = originalVersion
}

func TestVersionCommandFlags(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Version Command Flags ==="))
	
	cmd := NewVersionCmd()
	
	// Version command should not have any custom flags
	testName := "Custom Flags Count"
	flagCount := cmd.Flags().NFlag()
	success := flagCount == 0
	var message string
	if success {
		message = "Version command correctly has no custom flags"
	} else {
		message = fmt.Sprintf("Version command should not have custom flags, but has %d", flagCount)
	}
	printTestStatus(t, testName, success, message)
}
