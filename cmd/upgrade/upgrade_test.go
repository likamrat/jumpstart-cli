package upgrade

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

// Helper function to print test status
func printTestStatus(t *testing.T, testName string, success bool, message string) {
	if success {
		fmt.Printf("%s %s - %s\n", testSuccessColor("✅"), testName, testInfoColor(message))
	} else {
		fmt.Printf("%s %s - %s\n", testErrorColor("❌"), testName, testErrorColor(message))
		t.Error(message)
	}
}

func TestNewUpgradeCmd(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Creation ==="))
	
	cmd := NewUpgradeCmd()
	
	// Test basic command structure
	printTestStatus(t, "Command Use", cmd.Use == "upgrade", fmt.Sprintf("Expected 'upgrade', got '%s'", cmd.Use))
	printTestStatus(t, "Command Short Description", cmd.Short != "", "Short description should not be empty")
	printTestStatus(t, "Command Long Description", cmd.Long != "", "Long description should not be empty")
	printTestStatus(t, "RunE Function", cmd.RunE != nil, "RunE function should be defined")
}

func TestUpgradeCommandFlags(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Flags ==="))
	
	cmd := NewUpgradeCmd()
	
	// Test that expected flags exist
	expectedFlags := []string{"check", "pre-release", "force"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		success := flag != nil
		printTestStatus(t, fmt.Sprintf("Flag '%s'", flagName), success, 
			fmt.Sprintf("Flag '%s' should exist", flagName))
	}
}

func TestUpgradeCommandFlagDefaults(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Flag Defaults ==="))
	
	cmd := NewUpgradeCmd()
	
	// Test flag default values
	flagTests := []struct {
		name         string
		flagName     string
		expectedType string
		expectBool   bool
	}{
		{
			name:         "check flag default",
			flagName:     "check",
			expectedType: "bool",
			expectBool:   false,
		},
		{
			name:         "pre-release flag default", 
			flagName:     "pre-release",
			expectedType: "bool",
			expectBool:   false,
		},
		{
			name:         "force flag default",
			flagName:     "force",
			expectedType: "bool",
			expectBool:   false,
		},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				printTestStatus(t, fmt.Sprintf("Flag '%s' exists", tt.flagName), false, 
					fmt.Sprintf("Flag '%s' not found", tt.flagName))
				return
			}
			
			// Test flag type
			typeCorrect := flag.Value.Type() == tt.expectedType
			printTestStatus(t, fmt.Sprintf("Flag '%s' type", tt.flagName), typeCorrect,
				fmt.Sprintf("Expected type '%s', got '%s'", tt.expectedType, flag.Value.Type()))
			
			// Test default value for bool flags
			if tt.expectedType == "bool" {
				value, err := cmd.Flags().GetBool(tt.flagName)
				if err != nil {
					printTestStatus(t, fmt.Sprintf("Flag '%s' value retrieval", tt.flagName), false,
						fmt.Sprintf("Error getting bool value: %v", err))
					return
				}
				valueCorrect := value == tt.expectBool
				printTestStatus(t, fmt.Sprintf("Flag '%s' default value", tt.flagName), valueCorrect,
					fmt.Sprintf("Expected %v, got %v", tt.expectBool, value))
			}
		})
	}
}

func TestUpgradeCommandFlagShorthands(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Flag Shorthands ==="))
	
	cmd := NewUpgradeCmd()
	
	// Test shorthand flags exist
	shorthandTests := []struct {
		flagName  string
		shorthand string
	}{
		{"check", "c"},
		{"pre-release", "p"},
		{"force", "f"},
	}

	for _, tt := range shorthandTests {
		t.Run("shorthand_"+tt.flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				printTestStatus(t, fmt.Sprintf("Flag '%s' exists", tt.flagName), false,
					fmt.Sprintf("Flag '%s' not found", tt.flagName))
				return
			}
			
			shorthandCorrect := flag.Shorthand == tt.shorthand
			printTestStatus(t, fmt.Sprintf("Flag '%s' shorthand", tt.flagName), shorthandCorrect,
				fmt.Sprintf("Expected shorthand '%s', got '%s'", tt.shorthand, flag.Shorthand))
		})
	}
}

func TestUpgradeCommandExecution(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Execution ==="))
	
	// Store original debug mode
	originalDebugMode := utils.DebugMode
	defer func() {
		utils.DebugMode = originalDebugMode
	}()

	tests := []struct {
		name        string
		args        []string
		debugMode   bool
		expectError bool
	}{
		{
			name:        "upgrade with check flag",
			args:        []string{"--check"},
			debugMode:   false,
			expectError: false, // Should not error, but may not find updates
		},
		{
			name:        "upgrade with debug mode",
			args:        []string{"--check"},
			debugMode:   true,
			expectError: false,
		},
		{
			name:        "upgrade with pre-release flag",
			args:        []string{"--check", "--pre-release"},
			debugMode:   false,
			expectError: false,
		},
		{
			name:        "upgrade with force flag",
			args:        []string{"--check", "--force"},
			debugMode:   false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor(tt.name))
			
			// Set debug mode for test
			utils.DebugMode = tt.debugMode
			
			cmd := NewUpgradeCmd()
			cmd.SetArgs(tt.args)
			
			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			
			err := cmd.Execute()
			
			// Since the actual upgrade logic involves network calls and file operations,
			// we mainly test that the command structure is correct and doesn't panic
			// The actual upgrade functionality would require more complex mocking
			
			if tt.expectError && err == nil {
				printTestStatus(t, "Error expectation", false, "Expected error but got none")
				return
			}
			
			// For non-error cases, we can check if certain expected messages appear
			if !tt.expectError {
				output := buf.String()
				// Should at least attempt to check for updates or show repository guidance
				hasExpectedOutput := strings.Contains(output, "Checking for updates") || 
				                   strings.Contains(output, "Repository not found") ||
				                   strings.Contains(output, "failed to check for updates")
				                   
				printTestStatus(t, "Command execution", err == nil, 
					fmt.Sprintf("Command executed with args: %v", tt.args))
				
				if hasExpectedOutput || len(output) > 0 {
					fmt.Printf("      %s Output captured: %s\n", testInfoColor("ℹ"), testInfoColor("success"))
				}
			}
		})
	}
}

func TestUpgradeCommandStructure(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Structure ==="))
	
	cmd := NewUpgradeCmd()
	
	// Test command description contains key information
	expectedInLong := []string{
		"Check for and install",
		"latest version",
		"Examples:",
		"js upgrade",
		"--check",
		"--pre-release",
		"--force",
	}
	
	for _, expected := range expectedInLong {
		contains := strings.Contains(cmd.Long, expected)
		printTestStatus(t, fmt.Sprintf("Long description contains '%s'", expected), contains,
			fmt.Sprintf("Command.Long should contain '%s'", expected))
	}
}

func TestUpgradeCommandFlagParsing(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Flag Parsing ==="))
	
	// Test various flag combinations
	flagCombinations := [][]string{
		{"--check"},
		{"--pre-release"},
		{"--force"},
		{"-c"},
		{"-p"},
		{"-f"},
		{"--check", "--pre-release"},
		{"--check", "--force"},
		{"-c", "-p", "-f"},
	}

	for i, args := range flagCombinations {
		t.Run("flag_combination_"+string(rune(i+'0')), func(t *testing.T) {
			fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor(fmt.Sprintf("Testing flag combination: %v", args)))
			
			// Create a fresh command for each test
			cmd := NewUpgradeCmd()
			cmd.SetArgs(args)
			
			// Just test that flags can be parsed without error
			err := cmd.ParseFlags(args)
			success := err == nil
			printTestStatus(t, fmt.Sprintf("Flag parsing: %v", args), success,
				fmt.Sprintf("Flags should parse successfully: %v", args))
		})
	}
}
