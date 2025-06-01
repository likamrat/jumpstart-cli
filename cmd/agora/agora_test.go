package agora

import (
	"fmt"
	"os"
	"testing"

	"github.com/fatih/color"
)

// Test color functions for better visual feedback
var (
	testSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	testInfoColor    = color.New(color.FgCyan).SprintFunc()
	testWarnColor    = color.New(color.FgYellow).SprintFunc()
	testErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	testHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

// Helper function to print colored test output
func printTestStatus(t *testing.T, testName string, success bool, message string) {
	if success {
		fmt.Printf("%s ✅ %s: %s\n", testSuccessColor("PASS"), testHeaderColor(testName), testInfoColor(message))
	} else {
		fmt.Printf("%s ❌ %s: %s\n", testErrorColor("FAIL"), testHeaderColor(testName), testErrorColor(message))
		t.Error(message)
	}
}

func TestBuildNormalizedAgoraRegionMap(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Agora Region Map Building ==="))

	// Test the buildNormalizedAgoraRegionMap function
	regionMap := buildNormalizedAgoraRegionMap()

	// The map should not be nil
	testName := "Region Map Not Nil"
	success := regionMap != nil
	message := "buildNormalizedAgoraRegionMap() should not return nil"
	if !success {
		message = "buildNormalizedAgoraRegionMap() returned nil"
	}
	printTestStatus(t, testName, success, message)

	// Note: This test may not have entries if the JSON file doesn't exist
	// but should not panic or return nil
	testName = "Region Map Info"
	success = true // Always pass for info
	message = fmt.Sprintf("Region map has %d entries", len(regionMap))
	printTestStatus(t, testName, success, message)
}

func TestNewAgoraCmd(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Agora Command Creation ==="))

	cmd := NewAgoraCmd()

	// Test basic command structure
	testName := "Command Use"
	success := cmd.Use == "agora"
	message := fmt.Sprintf("Expected 'agora', got '%s'", cmd.Use)
	printTestStatus(t, testName, success, message)

	testName = "Command Short Description"
	success = cmd.Short != ""
	message = "Short description should not be empty"
	if !success {
		message = "Short description is empty"
	}
	printTestStatus(t, testName, success, message)

	testName = "Command Long Description"
	success = cmd.Long != ""
	message = "Long description should not be empty"
	if !success {
		message = "Long description is empty"
	}
	printTestStatus(t, testName, success, message)

	// Test that Run function exists
	testName = "Run Function Exists"
	success = cmd.Run != nil
	message = "Run function should not be nil"
	if !success {
		message = "Run function is nil"
	}
	printTestStatus(t, testName, success, message)
}

func TestAgoraCommandFlags(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Agora Command Flags ==="))

	cmd := NewAgoraCmd()

	// Agora command should not have any custom flags currently
	testName := "No Custom Flags"
	flagCount := cmd.Flags().NFlag()
	success := flagCount == 0
	message := fmt.Sprintf("Expected 0 custom flags, got %d", flagCount)
	printTestStatus(t, testName, success, message)
}

func TestAgoraCommandValidation(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectExitCode bool
	}{
		{
			name:           "no arguments",
			args:           []string{},
			expectError:    false,
			expectExitCode: false,
		},
		{
			name:           "unsupported region argument",
			args:           []string{"unsupported-region"},
			expectError:    false,
			expectExitCode: false, // Should warn but not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewAgoraCmd()
			cmd.SetArgs(tt.args)

			// Redirect stdout/stderr to avoid actual output during tests
			cmd.SetOut(os.Stdout)
			cmd.SetErr(os.Stderr)

			// Note: Since the Run function contains os.Exit calls and user prompts,
			// we can't easily test the actual execution without mocking
			// This test mainly verifies command structure
		})
	}
}

func TestAgoraCommandStructure(t *testing.T) {
	cmd := NewAgoraCmd()

	// Verify command properties
	expectedProperties := map[string]interface{}{
		"Use":   "agora",
		"Short": cmd.Short, // Should not be empty
		"Long":  cmd.Long,  // Should not be empty
	}

	for property, expected := range expectedProperties {
		switch property {
		case "Use":
			if cmd.Use != expected.(string) {
				t.Errorf("Command.%s = %v, want %v", property, cmd.Use, expected)
			}
		case "Short":
			if cmd.Short == "" {
				t.Errorf("Command.%s should not be empty", property)
			}
		case "Long":
			if cmd.Long == "" {
				t.Errorf("Command.%s should not be empty", property)
			}
		}
	}
}

// Test helper functions if any are exported in the future
func TestAgoraHelperFunctions(t *testing.T) {
	// Test buildNormalizedAgoraRegionMap with different scenarios
	t.Run("region map consistency", func(t *testing.T) {
		map1 := buildNormalizedAgoraRegionMap()
		map2 := buildNormalizedAgoraRegionMap()

		if len(map1) != len(map2) {
			t.Errorf("buildNormalizedAgoraRegionMap() not consistent: %d vs %d entries", len(map1), len(map2))
		}

		// Verify all keys in map1 exist in map2
		for key, value := range map1 {
			if map2Value, exists := map2[key]; !exists || map2Value != value {
				t.Errorf("Region map inconsistency for key %s: %s vs %s", key, value, map2Value)
			}
		}
	})
}
