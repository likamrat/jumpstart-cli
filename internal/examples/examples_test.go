package examples

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"
)

var (
	exampleTestSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	exampleTestInfoColor    = color.New(color.FgCyan).SprintFunc()
	exampleTestErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	exampleTestHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

func printExampleTestStatus(t *testing.T, testName string, success bool, message string) {
	status := exampleTestSuccessColor("✅")
	if !success {
		status = exampleTestErrorColor("❌")
	}
	fmt.Printf("  %s %s: %s\n", status, exampleTestInfoColor(testName), message)
}

func TestExample(t *testing.T) {
	fmt.Printf("\n%s\n", exampleTestHeaderColor("=== Testing Example Struct ==="))

	// Test Example struct creation and basic functionality
	example := Example{
		Description: "Test example description",
		Command:     "js test command",
	}

	if example.Description != "Test example description" {
		printExampleTestStatus(t, "Description assignment", false, fmt.Sprintf("Expected description 'Test example description', got '%s'", example.Description))
		t.Errorf("Expected description 'Test example description', got '%s'", example.Description)
	} else {
		printExampleTestStatus(t, "Description assignment", true, "Description correctly assigned")
	}

	if example.Command != "js test command" {
		printExampleTestStatus(t, "Command assignment", false, fmt.Sprintf("Expected command 'js test command', got '%s'", example.Command))
		t.Errorf("Expected command 'js test command', got '%s'", example.Command)
	} else {
		printExampleTestStatus(t, "Command assignment", true, "Command correctly assigned")
	}
}

func TestExampleSet(t *testing.T) {
	fmt.Printf("\n%s\n", exampleTestHeaderColor("=== Testing ExampleSet ==="))

	// Test ExampleSet with multiple examples
	examples := []Example{
		{
			Description: "First example",
			Command:     "js first",
		},
		{
			Description: "Second example",
			Command:     "js second --flag",
		},
	}

	exampleSet := ExampleSet{Examples: examples}

	if len(exampleSet.Examples) != 2 {
		printExampleTestStatus(t, "Example count", false, fmt.Sprintf("Expected 2 examples, got %d", len(exampleSet.Examples)))
		t.Errorf("Expected 2 examples, got %d", len(exampleSet.Examples))
	} else {
		printExampleTestStatus(t, "Example count", true, fmt.Sprintf("ExampleSet contains %d examples", len(exampleSet.Examples)))
	}

	if exampleSet.Examples[0].Description != "First example" {
		printExampleTestStatus(t, "First example description", false, fmt.Sprintf("Expected first example description 'First example', got '%s'", exampleSet.Examples[0].Description))
		t.Errorf("Expected first example description 'First example', got '%s'", exampleSet.Examples[0].Description)
	} else {
		printExampleTestStatus(t, "First example description", true, "First example description correctly set")
	}

	if exampleSet.Examples[1].Command != "js second --flag" {
		printExampleTestStatus(t, "Second example command", false, fmt.Sprintf("Expected second example command 'js second --flag', got '%s'", exampleSet.Examples[1].Command))
		t.Errorf("Expected second example command 'js second --flag', got '%s'", exampleSet.Examples[1].Command)
	} else {
		printExampleTestStatus(t, "Second example command", true, "Second example command correctly set")
	}
}

func TestFormatExamples(t *testing.T) {
	fmt.Printf("\n%s\n", exampleTestHeaderColor("=== Testing Example Formatting ==="))

	t.Run("empty_examples", func(t *testing.T) {
		fmt.Printf("  %s Testing empty example set\n", exampleTestInfoColor("Testing:"))

		emptySet := ExampleSet{}
		result := emptySet.FormatExamples()

		if result != "" {
			printExampleTestStatus(t, "Empty set formatting", false, fmt.Sprintf("Expected empty string for empty example set, got '%s'", result))
			t.Errorf("Expected empty string for empty example set, got '%s'", result)
		} else {
			printExampleTestStatus(t, "Empty set formatting", true, "Empty example set returns empty string")
		}
	})

	t.Run("single_example", func(t *testing.T) {
		fmt.Printf("  %s Testing single example formatting\n", exampleTestInfoColor("Testing:"))

		singleSet := ExampleSet{
			Examples: []Example{
				{
					Description: "Deploy ArcBox",
					Command:     "js arcbox deploy --flavor ITPro",
				},
			},
		}

		result := singleSet.FormatExamples()

		// Should contain "Examples" header
		if !strings.Contains(result, "Examples") {
			printExampleTestStatus(t, "Examples header", false, "Formatted output should contain 'Examples' header")
			t.Error("Formatted output should contain 'Examples' header")
		} else {
			printExampleTestStatus(t, "Examples header", true, "Examples header found in output")
		}

		// Should contain the description
		if !strings.Contains(result, "Deploy ArcBox") {
			printExampleTestStatus(t, "Description content", false, "Formatted output should contain the example description")
			t.Error("Formatted output should contain the example description")
		} else {
			printExampleTestStatus(t, "Description content", true, "Example description found in output")
		}

		// Should contain the command
		if !strings.Contains(result, "js arcbox deploy --flavor ITPro") {
			printExampleTestStatus(t, "Command content", false, "Formatted output should contain the example command")
			t.Error("Formatted output should contain the example command")
		} else {
			printExampleTestStatus(t, "Command content", true, "Example command found in output")
		}

		// Should have proper indentation (command more indented than description)
		lines := strings.Split(result, "\n")
		var descLine, cmdLine string
		for _, line := range lines {
			if strings.Contains(line, "Deploy ArcBox") {
				descLine = line
			}
			if strings.Contains(line, "js arcbox deploy") {
				cmdLine = line
			}
		}
		if descLine == "" || cmdLine == "" {
			printExampleTestStatus(t, "Line parsing", false, "Could not find description or command lines in formatted output")
			t.Fatal("Could not find description or command lines in formatted output")
		} else {
			printExampleTestStatus(t, "Line parsing", true, "Description and command lines found")
		}

		// Count leading spaces for indentation
		descIndent := len(descLine) - len(strings.TrimLeft(descLine, " "))
		cmdIndent := len(cmdLine) - len(strings.TrimLeft(cmdLine, " "))

		if cmdIndent <= descIndent {
			printExampleTestStatus(t, "Indentation", false, fmt.Sprintf("Command should be more indented than description. Desc: %d spaces, Cmd: %d spaces", descIndent, cmdIndent))
			t.Errorf("Command should be more indented than description. Desc: %d spaces, Cmd: %d spaces", descIndent, cmdIndent)
		} else {
			printExampleTestStatus(t, "Indentation", true, fmt.Sprintf("Proper indentation: description (%d), command (%d)", descIndent, cmdIndent))
		}
	})
	t.Run("multiple_examples", func(t *testing.T) {
		fmt.Printf("  %s Testing multiple examples formatting\n", exampleTestInfoColor("Testing:"))

		multiSet := ExampleSet{
			Examples: []Example{
				{
					Description: "Deploy ArcBox ITPro",
					Command:     "js arcbox deploy --flavor ITPro --location eastus",
				},
				{
					Description: "List ArcBox deployments",
					Command:     "js arcbox list",
				},
				{
					Description: "Delete ArcBox deployment",
					Command:     "js arcbox delete --resource-group arcbox-rg",
				},
			},
		}

		result := multiSet.FormatExamples()

		// Should contain all descriptions
		descriptions := []string{"Deploy ArcBox ITPro", "List ArcBox deployments", "Delete ArcBox deployment"}
		missingDescriptions := 0
		for _, desc := range descriptions {
			if !strings.Contains(result, desc) {
				printExampleTestStatus(t, fmt.Sprintf("Description %s", desc), false, fmt.Sprintf("Formatted output should contain description '%s'", desc))
				t.Errorf("Formatted output should contain description '%s'", desc)
				missingDescriptions++
			}
		}

		if missingDescriptions == 0 {
			printExampleTestStatus(t, "All descriptions", true, fmt.Sprintf("All %d descriptions found in output", len(descriptions)))
		}

		// Should contain all commands
		commands := []string{
			"js arcbox deploy --flavor ITPro --location eastus",
			"js arcbox list",
			"js arcbox delete --resource-group arcbox-rg",
		}
		missingCommands := 0
		for _, cmd := range commands {
			if !strings.Contains(result, cmd) {
				printExampleTestStatus(t, "Command validation", false, fmt.Sprintf("Formatted output should contain command '%s'", cmd))
				t.Errorf("Formatted output should contain command '%s'", cmd)
				missingCommands++
			}
		}

		if missingCommands == 0 {
			printExampleTestStatus(t, "All commands", true, fmt.Sprintf("All %d commands found in output", len(commands)))
		}

		// Should have blank lines between examples
		lines := strings.Split(result, "\n")
		blankLineCount := 0
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				blankLineCount++
			}
		}

		// Should have at least some blank lines for separation
		if blankLineCount < 2 {
			printExampleTestStatus(t, "Blank line separation", false, fmt.Sprintf("Expected at least 2 blank lines between examples, found %d", blankLineCount))
			t.Errorf("Expected at least 2 blank lines between examples, found %d", blankLineCount)
		} else {
			printExampleTestStatus(t, "Blank line separation", true, fmt.Sprintf("Found %d blank lines for proper separation", blankLineCount))
		}
	})
}

func TestGetExamples(t *testing.T) {
	fmt.Printf("\n%s\n", exampleTestHeaderColor("=== Testing GetExamples Function ==="))

	t.Run("existing_command_paths", func(t *testing.T) {
		fmt.Printf("  %s Testing example retrieval for existing command paths\n", exampleTestInfoColor("Testing:"))

		// Test with actual command paths that exist in the registry
		existingPaths := []string{
			"js.arcbox.deploy",
			"js.arcbox.delete",
			"js.subscription.set",
			"js.subscription.show",
			"js.completion",
			"arcbox.delete",
			"js.arcbox.preflight.quota",
			"arcbox.preflight.rp.register",
			"arcbox.preflight.rp.show",
			"arcbox.preflight.rp.list",
			"js.arcbox.preflight.rp",
			"js.arcbox.preflight.rp.list",
			"js.arcbox.preflight.rp.register",
			"js.arcbox.preflight.rp.show",
			"js.arcbox.list",
		}

		for _, path := range existingPaths {
			result := GetExamples(path)

			if result == nil {
				printExampleTestStatus(t, fmt.Sprintf("Non-nil result for %s", path), false, "GetExamples should never return nil")
				t.Errorf("GetExamples should never return nil for path '%s'", path)
				continue
			}

			// For existing paths, we should get examples
			if len(result.Examples) == 0 {
				printExampleTestStatus(t, fmt.Sprintf("Examples for %s", path), false, fmt.Sprintf("Expected examples for existing path '%s', got 0", path))
				t.Errorf("Expected examples for existing path '%s', got 0 examples", path)
			} else {
				printExampleTestStatus(t, fmt.Sprintf("Examples for %s", path), true, fmt.Sprintf("Found %d examples", len(result.Examples)))

				// Validate that each example has both description and command
				for i, example := range result.Examples {
					if example.Description == "" {
						printExampleTestStatus(t, fmt.Sprintf("Example %d description", i), false, "Example description should not be empty")
						t.Errorf("Example %d for path '%s' has empty description", i, path)
					}
					if example.Command == "" {
						printExampleTestStatus(t, fmt.Sprintf("Example %d command", i), false, "Example command should not be empty")
						t.Errorf("Example %d for path '%s' has empty command", i, path)
					}
				}
			}
		}
	})

	t.Run("non_existing_command", func(t *testing.T) {
		fmt.Printf("  %s Testing example retrieval for non-existing command\n", exampleTestInfoColor("Testing:"))

		result := GetExamples("nonexistent-command")

		// Should return empty ExampleSet for non-existent commands
		if result == nil {
			printExampleTestStatus(t, "Non-nil result", false, "GetExamples should never return nil")
			t.Error("GetExamples should never return nil")
		} else {
			printExampleTestStatus(t, "Non-nil result", true, "GetExamples returned valid ExampleSet")
		}

		if result != nil && len(result.Examples) != 0 {
			printExampleTestStatus(t, "Empty example set", false, fmt.Sprintf("Expected empty example set for non-existent command, got %d examples", len(result.Examples)))
			t.Errorf("Expected empty example set for non-existent command, got %d examples", len(result.Examples))
		} else {
			printExampleTestStatus(t, "Empty example set", true, "Non-existent command returns empty example set")
		}
	})

	t.Run("empty_command_path", func(t *testing.T) {
		fmt.Printf("  %s Testing example retrieval for empty command path\n", exampleTestInfoColor("Testing:"))

		result := GetExamples("")

		if result == nil {
			printExampleTestStatus(t, "Non-nil result", false, "GetExamples should never return nil")
			t.Error("GetExamples should never return nil")
		} else {
			printExampleTestStatus(t, "Non-nil result", true, "GetExamples returned valid ExampleSet")
		}

		if result != nil && len(result.Examples) != 0 {
			printExampleTestStatus(t, "Empty example set", false, fmt.Sprintf("Expected empty example set for empty command path, got %d examples", len(result.Examples)))
			t.Errorf("Expected empty example set for empty command path, got %d examples", len(result.Examples))
		} else {
			printExampleTestStatus(t, "Empty example set", true, "Empty command path returns empty example set")
		}
	})

	t.Run("existing_command_paths", func(t *testing.T) {
		fmt.Printf("  %s Testing example retrieval for existing command paths\n", exampleTestInfoColor("Testing:"))

		// Test with actual command paths that exist in the registry
		existingPaths := []string{
			"js.arcbox.deploy",
			"js.arcbox.delete",
			"js.subscription.set",
			"js.subscription.show",
			"js.completion",
			"arcbox.delete",
			"js.arcbox.preflight.quota",
			"arcbox.preflight.rp.register",
			"arcbox.preflight.rp.show",
			"arcbox.preflight.rp.list",
			"js.arcbox.preflight.rp",
			"js.arcbox.preflight.rp.list",
			"js.arcbox.preflight.rp.register",
			"js.arcbox.preflight.rp.show",
			"js.arcbox.list",
		}

		for _, path := range existingPaths {
			result := GetExamples(path)

			if result == nil {
				printExampleTestStatus(t, fmt.Sprintf("Non-nil result for %s", path), false, "GetExamples should never return nil")
				t.Errorf("GetExamples should never return nil for path '%s'", path)
				continue
			}

			// For existing paths, we should get examples
			if len(result.Examples) == 0 {
				printExampleTestStatus(t, fmt.Sprintf("Examples for %s", path), false, fmt.Sprintf("Expected examples for existing path '%s', got 0", path))
				t.Errorf("Expected examples for existing path '%s', got 0 examples", path)
			} else {
				printExampleTestStatus(t, fmt.Sprintf("Examples for %s", path), true, fmt.Sprintf("Found %d examples", len(result.Examples)))

				// Validate that each example has both description and command
				for i, example := range result.Examples {
					if example.Description == "" {
						printExampleTestStatus(t, fmt.Sprintf("Example %d description", i), false, "Example description should not be empty")
						t.Errorf("Example %d for path '%s' has empty description", i, path)
					}
					if example.Command == "" {
						printExampleTestStatus(t, fmt.Sprintf("Example %d command", i), false, "Example command should not be empty")
						t.Errorf("Example %d for path '%s' has empty command", i, path)
					}
				}
			}
		}
	})
}

func TestExampleRegistryIntegration(t *testing.T) {
	fmt.Printf("\n%s\n", exampleTestHeaderColor("=== Testing Example Registry Integration ==="))

	// Test that the example registry is properly initialized and accessible
	t.Run("registry_initialization", func(t *testing.T) {
		fmt.Printf("  %s Testing registry initialization with various command paths\n", exampleTestInfoColor("Testing:"))

		// Try to get examples for various command paths that might exist
		commandPaths := []string{
			"arcbox",
			"arcbox deploy",
			"localbox",
			"agora",
			"repo",
			"subscription",
			"upgrade",
			"version",
		}

		validResponses := 0
		for _, path := range commandPaths {
			result := GetExamples(path)
			if result == nil {
				printExampleTestStatus(t, fmt.Sprintf("Registry access for %s", path), false, fmt.Sprintf("GetExamples should never return nil for path '%s'", path))
				t.Errorf("GetExamples should never return nil for path '%s'", path)
			} else {
				validResponses++
			}

			// Log the number of examples found for each path
			exampleCount := 0
			if result != nil {
				exampleCount = len(result.Examples)
			}
			printExampleTestStatus(t, fmt.Sprintf("Path %s", path), true, fmt.Sprintf("%d examples found", exampleCount))
			t.Logf("Command path '%s': %d examples", path, exampleCount)
		}

		if validResponses == len(commandPaths) {
			printExampleTestStatus(t, "Registry integrity", true, fmt.Sprintf("All %d command paths returned valid responses", validResponses))
		}
	})
}

func TestExampleFormatting(t *testing.T) {
	// Test various formatting scenarios
	t.Run("long_commands", func(t *testing.T) {
		longExample := ExampleSet{
			Examples: []Example{
				{
					Description: "Deploy ArcBox with all optional parameters",
					Command:     "js arcbox deploy --flavor ITPro --location eastus --resource-group my-rg --windows-user admin --windows-password MyPassword123! --admin-username azureuser --ssh-rsa-public-key ~/.ssh/id_rsa.pub",
				},
			},
		}

		result := longExample.FormatExamples()

		// Should handle long commands without issues
		if !strings.Contains(result, "Deploy ArcBox with all optional parameters") {
			t.Error("Should contain long description")
		}

		if !strings.Contains(result, "js arcbox deploy") {
			t.Error("Should contain long command")
		}
	})

	t.Run("special_characters", func(t *testing.T) {
		specialExample := ExampleSet{
			Examples: []Example{
				{
					Description: "Example with special characters: !@#$%^&*()",
					Command:     "js test --param 'value with spaces' --flag",
				},
			},
		}

		result := specialExample.FormatExamples()

		// Should handle special characters without issues
		if !strings.Contains(result, "!@#$%^&*()") {
			t.Error("Should contain special characters in description")
		}

		if !strings.Contains(result, "'value with spaces'") {
			t.Error("Should contain quoted parameters in command")
		}
	})

	t.Run("multiline_handling", func(t *testing.T) {
		// Test that the formatting creates proper line breaks
		multiExample := ExampleSet{
			Examples: []Example{
				{Description: "First", Command: "js first"},
				{Description: "Second", Command: "js second"},
			},
		}

		result := multiExample.FormatExamples()
		lines := strings.Split(result, "\n")

		// Should have proper structure: header, desc1, cmd1, blank, desc2, cmd2
		if len(lines) < 6 {
			t.Errorf("Expected at least 6 lines in output, got %d", len(lines))
		}

		// First line should be "Examples"
		if !strings.Contains(lines[0], "Examples") {
			t.Errorf("First line should contain 'Examples', got '%s'", lines[0])
		}
	})
}

// Test performance with large example sets
func BenchmarkFormatExamples(b *testing.B) {
	// Create a large example set
	examples := make([]Example, 100)
	for i := 0; i < 100; i++ {
		examples[i] = Example{
			Description: "Example description number " + string(rune(i+'0')),
			Command:     "js command number " + string(rune(i+'0')) + " --flag value",
		}
	}

	exampleSet := ExampleSet{Examples: examples}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exampleSet.FormatExamples()
	}
}

func BenchmarkGetExamples(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetExamples("arcbox")
	}
}

func TestExampleRegistryComprehensive(t *testing.T) {
	fmt.Printf("\n%s\n", exampleTestHeaderColor("=== Testing Comprehensive Registry Coverage ==="))

	t.Run("all_registry_paths", func(t *testing.T) {
		fmt.Printf("  %s Testing all command paths in the registry\n", exampleTestInfoColor("Testing:"))

		// Test that all paths in the registry are accessible and return valid examples
		registryPaths := []string{
			"js.arcbox.deploy",
			"js.arcbox.delete",
			"arcbox.delete",
			"js.arcbox.preflight.quota",
			"arcbox.preflight.rp.register",
			"arcbox.preflight.rp.show",
			"arcbox.preflight.rp.list",
			"js.arcbox.preflight.rp",
			"js.arcbox.preflight.rp.list",
			"js.arcbox.preflight.rp.register",
			"js.arcbox.preflight.rp.show",
			"js.arcbox.list",
			"js.subscription.set",
			"js.subscription.show",
			"js.completion",
		}

		totalExamples := 0
		for _, path := range registryPaths {
			result := GetExamples(path)
			if result == nil {
				printExampleTestStatus(t, fmt.Sprintf("Registry path %s", path), false, "Should return non-nil ExampleSet")
				t.Errorf("Registry path '%s' returned nil", path)
				continue
			}

			exampleCount := len(result.Examples)
			totalExamples += exampleCount

			if exampleCount == 0 {
				printExampleTestStatus(t, fmt.Sprintf("Registry path %s", path), false, "Should contain examples")
				t.Errorf("Registry path '%s' has no examples", path)
			} else {
				printExampleTestStatus(t, fmt.Sprintf("Registry path %s", path), true, fmt.Sprintf("%d examples", exampleCount))

				// Validate each example in the path
				for i, example := range result.Examples {
					if example.Description == "" {
						t.Errorf("Example %d in path '%s' has empty description", i, path)
					}
					if example.Command == "" {
						t.Errorf("Example %d in path '%s' has empty command", i, path)
					}
					// Ensure commands have valid structure (allow various command patterns)
					if !strings.HasPrefix(example.Command, "js ") &&
						!strings.HasPrefix(example.Command, "source ") &&
						!strings.HasPrefix(example.Command, "echo ") &&
						!strings.Contains(example.Command, "js ") {
						t.Errorf("Example %d in path '%s' has invalid command format: '%s'", i, path, example.Command)
					}
				}
			}
		}

		printExampleTestStatus(t, "Total examples", true, fmt.Sprintf("Found %d total examples across all paths", totalExamples))
		t.Logf("Total examples across all registry paths: %d", totalExamples)
	})

	t.Run("registry_consistency", func(t *testing.T) {
		fmt.Printf("  %s Testing registry consistency\n", exampleTestInfoColor("Testing:"))

		// Test that getting the same path multiple times returns consistent results
		testPath := "js.completion"
		result1 := GetExamples(testPath)
		result2 := GetExamples(testPath)

		if result1 == nil || result2 == nil {
			printExampleTestStatus(t, "Consistency non-nil", false, "Both results should be non-nil")
			t.Error("Both results should be non-nil")
			return
		}

		if len(result1.Examples) != len(result2.Examples) {
			printExampleTestStatus(t, "Consistency length", false, "Both results should have same length")
			t.Errorf("Inconsistent example count: %d vs %d", len(result1.Examples), len(result2.Examples))
		} else {
			printExampleTestStatus(t, "Consistency length", true, fmt.Sprintf("Both calls returned %d examples", len(result1.Examples)))
		}

		// Compare content
		for i := 0; i < len(result1.Examples) && i < len(result2.Examples); i++ {
			if result1.Examples[i].Description != result2.Examples[i].Description {
				t.Errorf("Inconsistent description at index %d", i)
			}
			if result1.Examples[i].Command != result2.Examples[i].Command {
				t.Errorf("Inconsistent command at index %d", i)
			}
		}
	})
}

func TestGetExamplesEdgeCases(t *testing.T) {
	fmt.Printf("\n%s\n", exampleTestHeaderColor("=== Testing GetExamples Edge Cases ==="))

	t.Run("whitespace_handling", func(t *testing.T) {
		fmt.Printf("  %s Testing whitespace handling in command paths\n", exampleTestInfoColor("Testing:"))

		// Test paths with various whitespace patterns
		testCases := []struct {
			name string
			path string
		}{
			{"leading_space", " js.completion"},
			{"trailing_space", "js.completion "},
			{"multiple_spaces", "js.completion  "},
			{"tab_character", "js.completion\t"},
			{"newline_character", "js.completion\n"},
		}

		for _, tc := range testCases {
			result := GetExamples(tc.path)
			if result == nil {
				printExampleTestStatus(t, fmt.Sprintf("Whitespace test %s", tc.name), false, "Should return non-nil ExampleSet")
				t.Errorf("GetExamples should never return nil for path '%s'", tc.path)
			} else {
				// These should all return empty since they don't match exact keys
				if len(result.Examples) != 0 {
					printExampleTestStatus(t, fmt.Sprintf("Whitespace test %s", tc.name), false, "Should return empty for whitespace-modified paths")
					t.Errorf("Expected empty examples for whitespace path '%s', got %d", tc.path, len(result.Examples))
				} else {
					printExampleTestStatus(t, fmt.Sprintf("Whitespace test %s", tc.name), true, "Correctly returned empty for whitespace-modified path")
				}
			}
		}
	})

	t.Run("case_sensitivity", func(t *testing.T) {
		fmt.Printf("  %s Testing case sensitivity in command paths\n", exampleTestInfoColor("Testing:"))

		// Test case variations of existing paths
		basePath := "js.completion"
		casePaths := []string{
			"JS.COMPLETION",
			"Js.Completion",
			"js.COMPLETION",
			"JS.completion",
		}

		// First verify the base path works
		baseResult := GetExamples(basePath)
		if baseResult == nil || len(baseResult.Examples) == 0 {
			t.Skip("Base path has no examples, skipping case sensitivity test")
		}

		for _, casePath := range casePaths {
			result := GetExamples(casePath)
			if result == nil {
				printExampleTestStatus(t, fmt.Sprintf("Case sensitivity %s", casePath), false, "Should return non-nil ExampleSet")
				t.Errorf("GetExamples should never return nil for path '%s'", casePath)
			} else {
				// These should all return empty since Go maps are case-sensitive
				if len(result.Examples) != 0 {
					printExampleTestStatus(t, fmt.Sprintf("Case sensitivity %s", casePath), false, "Should return empty for case-modified paths")
					t.Errorf("Expected empty examples for case-modified path '%s', got %d", casePath, len(result.Examples))
				} else {
					printExampleTestStatus(t, fmt.Sprintf("Case sensitivity %s", casePath), true, "Correctly returned empty for case-modified path")
				}
			}
		}
	})

	t.Run("special_characters_in_path", func(t *testing.T) {
		fmt.Printf("  %s Testing special characters in command paths\n", exampleTestInfoColor("Testing:"))

		specialPaths := []string{
			"js.completion@#$",
			"js/completion",
			"js\\completion",
			"js.completion!",
			"js.completion?",
			"js.completion*",
			"js.completion[",
			"js.completion]",
			"js.completion{",
			"js.completion}",
		}

		for _, path := range specialPaths {
			result := GetExamples(path)
			if result == nil {
				printExampleTestStatus(t, fmt.Sprintf("Special char %s", path), false, "Should return non-nil ExampleSet")
				t.Errorf("GetExamples should never return nil for path '%s'", path)
			} else {
				// These should all return empty since they don't match exact keys
				if len(result.Examples) != 0 {
					printExampleTestStatus(t, fmt.Sprintf("Special char %s", path), false, "Should return empty for paths with special characters")
					t.Errorf("Expected empty examples for special path '%s', got %d", path, len(result.Examples))
				} else {
					printExampleTestStatus(t, fmt.Sprintf("Special char %s", path), true, "Correctly returned empty for special character path")
				}
			}
		}
	})
}
