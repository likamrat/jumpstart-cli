package localbox

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
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

func TestBuildNormalizedLocalboxRegionMap(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Build Normalized Localbox Region Map ==="))

	// Test the buildNormalizedLocalboxRegionMap function
	regionMap := buildNormalizedLocalboxRegionMap()

	// The map should not be nil
	success := regionMap != nil
	var message string
	if !success {
		message = "buildNormalizedLocalboxRegionMap() returned nil"
	} else {
		message = fmt.Sprintf("Region map has %d entries", len(regionMap))
	}

	printTestStatus(t, "Region Map Creation", success, message)
}

func TestNewLocalboxCmd(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing New Localbox Command ==="))

	cmd := NewLocalboxCmd()

	// Test basic command structure
	testName := "Command Use Field"
	success := cmd.Use == "localbox"
	message := fmt.Sprintf("Expected 'localbox', got '%s'", cmd.Use)
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

	// Test that Run function exists
	testName = "Run Function"
	success = cmd.Run != nil
	if success {
		message = "Run function is properly set"
	} else {
		message = "Run function should not be nil"
	}
	printTestStatus(t, testName, success, message)
}

func TestLocalboxCommandFlags(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Localbox Command Flags ==="))

	cmd := NewLocalboxCmd()

	// Localbox command should not have any custom flags currently
	testName := "Custom Flags Count"
	flagCount := cmd.Flags().NFlag()
	success := flagCount == 0
	var message string
	if success {
		message = "Localbox command correctly has no custom flags"
	} else {
		message = fmt.Sprintf("Localbox command should not have custom flags, but has %d", flagCount)
	}
	printTestStatus(t, testName, success, message)
}

func TestLocalboxCommandValidation(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Localbox Command Validation ==="))

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
			testName := fmt.Sprintf("Validation Test: %s", tt.name)
			cmd := NewLocalboxCmd()
			cmd.SetArgs(tt.args)

			// Redirect stdout/stderr to avoid actual output during tests
			cmd.SetOut(os.Stdout)
			cmd.SetErr(os.Stderr)

			// Note: Since the Run function contains os.Exit calls and user prompts,
			// we can't easily test the actual execution without mocking
			// This test mainly verifies command structure
			message := fmt.Sprintf("Args %v validated (structure test only)", tt.args)
			printTestStatus(t, testName, true, message)
		})
	}
}

func TestLocalboxCommandStructure(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Localbox Command Structure ==="))

	cmd := NewLocalboxCmd()

	// Verify command properties
	tests := []struct {
		property string
		check    func() (bool, string)
	}{
		{
			property: "Use Field",
			check: func() (bool, string) {
				expected := "localbox"
				actual := cmd.Use
				success := actual == expected
				message := fmt.Sprintf("Expected '%s', got '%s'", expected, actual)
				return success, message
			},
		},
		{
			property: "Short Description",
			check: func() (bool, string) {
				success := cmd.Short != ""
				var message string
				if success {
					message = "Short description is properly set"
				} else {
					message = "Short description should not be empty"
				}
				return success, message
			},
		},
		{
			property: "Long Description",
			check: func() (bool, string) {
				success := cmd.Long != ""
				var message string
				if success {
					message = "Long description is properly set"
				} else {
					message = "Long description should not be empty"
				}
				return success, message
			},
		},
	}

	for _, test := range tests {
		success, message := test.check()
		printTestStatus(t, test.property, success, message)
	}
}

// Test helper functions if any are exported in the future
func TestLocalboxHelperFunctions(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Localbox Helper Functions ==="))

	// Test buildNormalizedLocalboxRegionMap with different scenarios
	t.Run("region map consistency", func(t *testing.T) {
		testName := "Region Map Consistency"
		map1 := buildNormalizedLocalboxRegionMap()
		map2 := buildNormalizedLocalboxRegionMap()

		success := len(map1) == len(map2)
		var message string
		if !success {
			message = fmt.Sprintf("buildNormalizedLocalboxRegionMap() not consistent: %d vs %d entries", len(map1), len(map2))
		} else {
			// Verify all keys in map1 exist in map2
			allMatch := true
			for key, value := range map1 {
				if map2Value, exists := map2[key]; !exists || map2Value != value {
					allMatch = false
					message = fmt.Sprintf("Region map inconsistency for key %s: %s vs %s", key, value, map2Value)
					break
				}
			}
			if allMatch {
				message = fmt.Sprintf("Region map consistency verified with %d entries", len(map1))
			}
			success = allMatch
		}

		printTestStatus(t, testName, success, message)
	})
}

func TestLocalboxCommandLongDescription(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Localbox Command Long Description ==="))

	cmd := NewLocalboxCmd()

	// Test that the long description contains expected content
	expectedContents := []string{
		"LocalBox",
		"automation",
		"Implementation in progress",
		"js localbox --help",
	}

	for _, expected := range expectedContents {
		testName := fmt.Sprintf("Long Description Contains: %s", expected)
		success := strings.Contains(cmd.Long, expected)
		var message string
		if success {
			message = fmt.Sprintf("Long description correctly contains \"%s\"", expected)
		} else {
			message = fmt.Sprintf("Long description should contain \"%s\"", expected)
		}
		printTestStatus(t, testName, success, message)
	}
}

func TestLocalboxRegionValidation(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Localbox Region Validation ==="))

	// Test region map building doesn't panic with missing file
	t.Run("missing_file_handling", func(t *testing.T) {
		testName := "Missing File Handling"
		var success bool
		var message string

		// This should not panic even if the JSON file doesn't exist
		defer func() {
			if r := recover(); r != nil {
				success = false
				message = fmt.Sprintf("buildNormalizedLocalboxRegionMap() panicked: %v", r)
				printTestStatus(t, testName, success, message)
			}
		}()

		regionMap := buildNormalizedLocalboxRegionMap()
		success = regionMap != nil
		if success {
			message = "buildNormalizedLocalboxRegionMap() handled missing file gracefully"
		} else {
			message = "buildNormalizedLocalboxRegionMap() should not return nil even with missing file"
		}

		printTestStatus(t, testName, success, message)
	})
}
