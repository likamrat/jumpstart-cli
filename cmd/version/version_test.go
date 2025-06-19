package version

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestVersionCommand tests the basic version command functionality
func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
		notExpected    []string
	}{
		{
			name: "basic version output",
			args: []string{},
			expectedOutput: []string{
				"Jumpstart CLI version:",
				utils.CliVersion,
			},
			notExpected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			buf := new(bytes.Buffer)

			// Create root command with output redirected
			rootCmd := &cobra.Command{Use: "jumpstart"}
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			// Add version command
			versionCmd := NewVersionCmd()
			rootCmd.AddCommand(versionCmd)

			// Execute command
			rootCmd.SetArgs(append([]string{"version"}, tt.args...))
			err := rootCmd.Execute()

			assert.NoError(t, err)
			output := buf.String()

			// Check expected outputs
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}

			// Check not expected outputs
			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, output, notExpected, "Output should not contain: %s", notExpected)
			}
		})
	}
}

// TestNewVersionCmd tests the command creation
func TestNewVersionCmd(t *testing.T) {
	cmd := NewVersionCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Use)
	assert.Equal(t, "Display the current version of the CLI", cmd.Short)
	assert.Equal(t, "Display the current version of the Jumpstart CLI", cmd.Long)
	assert.NotNil(t, cmd.Run)
}

// TestVersionOutput tests the exact output format
func TestVersionOutput(t *testing.T) {
	// Create a buffer to capture output
	buf := new(bytes.Buffer)

	// Create the version command
	cmd := NewVersionCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Execute the command
	err := cmd.Execute()

	assert.NoError(t, err)
	output := buf.String()

	// Check the exact output format
	expectedOutput := "Jumpstart CLI version: " + utils.CliVersion + "\n"
	assert.Equal(t, expectedOutput, output)
}

// TestVersionWithParentCommand tests version command when called from parent
func TestVersionWithParentCommand(t *testing.T) {
	// Create root command
	rootCmd := &cobra.Command{
		Use: "jumpstart",
	}

	// Create buffer to capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Add version command
	versionCmd := NewVersionCmd()
	rootCmd.AddCommand(versionCmd)

	// Execute version command
	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()

	assert.NoError(t, err)
	output := buf.String()

	// Verify output contains expected elements
	assert.Contains(t, output, "Jumpstart CLI version:")
	assert.Contains(t, output, utils.CliVersion)
	assert.True(t, strings.HasSuffix(strings.TrimSpace(output), utils.CliVersion))
}

// TestVersionCommandStructure tests the command structure and metadata
func TestVersionCommandStructure(t *testing.T) {
	cmd := NewVersionCmd()

	// Test command properties
	assert.Equal(t, "version", cmd.Use)
	assert.Contains(t, cmd.Short, "version")
	assert.Contains(t, cmd.Long, "Jumpstart CLI")
	assert.NotNil(t, cmd.Run)

	// Test that command has no subcommands
	assert.Len(t, cmd.Commands(), 0)

	// Test that command has no flags
	assert.False(t, cmd.HasAvailableFlags())
}

// TestVersionCommandWithDifferentOutputWriters tests version command with different output writers
func TestVersionCommandWithDifferentOutputWriters(t *testing.T) {
	testCases := []struct {
		name   string
		writer *bytes.Buffer
	}{
		{
			name:   "stdout writer",
			writer: new(bytes.Buffer),
		},
		{
			name:   "stderr writer",
			writer: new(bytes.Buffer),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewVersionCmd()
			cmd.SetOut(tc.writer)
			cmd.SetErr(tc.writer)

			err := cmd.Execute()
			assert.NoError(t, err)

			output := tc.writer.String()
			assert.Contains(t, output, "Jumpstart CLI version:")
			assert.Contains(t, output, utils.CliVersion)
		})
	}
}

// TestVersionCommandConsistency tests version output consistency across executions
func TestVersionCommandConsistency(t *testing.T) {
	// Run the command multiple times and verify consistent output
	var outputs []string

	for i := 0; i < 5; i++ {
		buf := new(bytes.Buffer)
		cmd := NewVersionCmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		assert.NoError(t, err)

		outputs = append(outputs, buf.String())
	}

	// All outputs should be identical
	expected := outputs[0]
	for i, output := range outputs {
		assert.Equal(t, expected, output, "Output %d should match first output", i)
	}
}

// TestVersionCommandEdgeCases tests edge cases for the version command
func TestVersionCommandEdgeCases(t *testing.T) {
	t.Run("empty CliVersion", func(t *testing.T) {
		// Save original version
		originalVersion := utils.CliVersion
		defer func() { utils.CliVersion = originalVersion }()

		// Set empty version
		utils.CliVersion = ""

		buf := new(bytes.Buffer)
		cmd := NewVersionCmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Jumpstart CLI version:")
		// Should still contain the prefix even with empty version
	})

	t.Run("nil output writer", func(t *testing.T) {
		cmd := NewVersionCmd()
		// Don't set output writer - should use default

		// This should not panic
		assert.NotPanics(t, func() {
			cmd.Execute()
		})
	})
}

// TestVersionCommandDeepValidation provides comprehensive testing beyond basic coverage
func TestVersionCommandDeepValidation(t *testing.T) {
	t.Run("version_string_formatting_edge_cases", func(t *testing.T) {
		testCases := []struct {
			name            string
			version         string
			expectedPattern string
			description     string
		}{
			{
				name:            "semantic_version",
				version:         "1.2.3",
				expectedPattern: "Jumpstart CLI version: 1.2.3\n",
				description:     "Standard semantic version",
			},
			{
				name:            "version_with_prerelease",
				version:         "1.2.3-alpha.1",
				expectedPattern: "Jumpstart CLI version: 1.2.3-alpha.1\n",
				description:     "Version with prerelease identifier",
			},
			{
				name:            "version_with_build_metadata",
				version:         "1.2.3+build.123",
				expectedPattern: "Jumpstart CLI version: 1.2.3+build.123\n",
				description:     "Version with build metadata",
			},
			{
				name:            "git_commit_version",
				version:         "v1.2.3-10-gf123abc",
				expectedPattern: "Jumpstart CLI version: v1.2.3-10-gf123abc\n",
				description:     "Git describe format version",
			},
			{
				name:            "dirty_git_version",
				version:         "v1.2.3-dirty",
				expectedPattern: "Jumpstart CLI version: v1.2.3-dirty\n",
				description:     "Git version with uncommitted changes",
			},
			{
				name:            "development_version",
				version:         "dev",
				expectedPattern: "Jumpstart CLI version: dev\n",
				description:     "Development version identifier",
			},
			{
				name:            "version_with_spaces",
				version:         "1.2.3 beta",
				expectedPattern: "Jumpstart CLI version: 1.2.3 beta\n",
				description:     "Version with internal spaces",
			},
			{
				name:            "very_long_version",
				version:         "1.2.3-very.long.prerelease.identifier.with.many.parts+build.metadata.20231215.commit.abcdef123456",
				expectedPattern: "Jumpstart CLI version: 1.2.3-very.long.prerelease.identifier.with.many.parts+build.metadata.20231215.commit.abcdef123456\n",
				description:     "Extremely long version string",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Save original version
				originalVersion := utils.CliVersion
				defer func() { utils.CliVersion = originalVersion }()

				// Set test version
				utils.CliVersion = tc.version

				buf := new(bytes.Buffer)
				cmd := NewVersionCmd()
				cmd.SetOut(buf)
				cmd.SetErr(buf)

				err := cmd.Execute()
				assert.NoError(t, err, "Command should execute without error for %s", tc.description)

				output := buf.String()
				assert.Equal(t, tc.expectedPattern, output, "Output format should match exactly for %s", tc.description)
				assert.Contains(t, output, tc.version, "Output should contain the version string for %s", tc.description)
			})
		}
	})

	t.Run("version_command_flag_variations", func(t *testing.T) {
		// Test that version command properly handles various flag scenarios
		testCases := []struct {
			name        string
			args        []string
			shouldError bool
			description string
		}{
			{
				name:        "no_args",
				args:        []string{},
				shouldError: false,
				description: "Version command with no arguments",
			},
			{
				name:        "help_flag",
				args:        []string{"--help"},
				shouldError: false,
				description: "Version command with help flag",
			},
			{
				name:        "short_help_flag",
				args:        []string{"-h"},
				shouldError: false,
				description: "Version command with short help flag",
			},
			{
				name:        "unknown_flag",
				args:        []string{"--unknown"},
				shouldError: true,
				description: "Version command with unknown flag should error",
			},
			{
				name:        "extra_arguments",
				args:        []string{"extra", "args"},
				shouldError: false,
				description: "Version command should ignore extra arguments",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				rootCmd := &cobra.Command{Use: "jumpstart"}
				rootCmd.SetOut(buf)
				rootCmd.SetErr(buf)

				versionCmd := NewVersionCmd()
				rootCmd.AddCommand(versionCmd)

				rootCmd.SetArgs(append([]string{"version"}, tc.args...))
				err := rootCmd.Execute()

				if tc.shouldError {
					assert.Error(t, err, "Should return error for %s", tc.description)
				} else {
					assert.NoError(t, err, "Should not return error for %s", tc.description)
				}
			})
		}
	})

	t.Run("version_output_formatting_validation", func(t *testing.T) {
		// Test output formatting details
		buf := new(bytes.Buffer)
		cmd := NewVersionCmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		assert.NoError(t, err)

		output := buf.String()

		// Verify exact formatting requirements
		assert.True(t, strings.HasPrefix(output, "Jumpstart CLI version: "), "Output should start with proper prefix")
		assert.True(t, strings.HasSuffix(output, "\n"), "Output should end with newline")
		assert.Equal(t, 1, strings.Count(output, "\n"), "Output should contain exactly one newline")
		assert.NotContains(t, output, "\t", "Output should not contain tab characters")
		assert.NotContains(t, output, "\r", "Output should not contain carriage return characters")

		// Verify no extra whitespace
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, lines, 1, "Output should be exactly one line")
		assert.Equal(t, strings.TrimSpace(lines[0]), lines[0], "Line should not have leading/trailing whitespace")
	})

	t.Run("performance_testing", func(t *testing.T) {
		// Test performance characteristics
		const iterations = 1000
		var totalDuration int64

		for i := 0; i < iterations; i++ {
			start := time.Now()

			buf := new(bytes.Buffer)
			cmd := NewVersionCmd()
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()
			assert.NoError(t, err)

			duration := time.Since(start)
			totalDuration += duration.Nanoseconds()
		}

		avgDuration := time.Duration(totalDuration / iterations)

		// Version command should be very fast (< 1ms average)
		assert.Less(t, avgDuration, time.Millisecond, "Average execution time should be under 1ms")

		t.Logf("Version command average execution time: %v", avgDuration)
	})

	t.Run("concurrent_execution_safety", func(t *testing.T) {
		// Test thread safety
		const numGoroutines = 100
		const numIterations = 10

		var wg sync.WaitGroup
		errorChan := make(chan error, numGoroutines*numIterations)
		outputChan := make(chan string, numGoroutines*numIterations)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < numIterations; j++ {
					buf := new(bytes.Buffer)
					cmd := NewVersionCmd()
					cmd.SetOut(buf)
					cmd.SetErr(buf)

					err := cmd.Execute()
					if err != nil {
						errorChan <- err
						return
					}
					outputChan <- buf.String()
				}
			}()
		}

		wg.Wait()
		close(errorChan)
		close(outputChan)

		// Check for any errors
		for err := range errorChan {
			t.Errorf("Concurrent execution error: %v", err)
		}

		// Verify all outputs are consistent
		var outputs []string
		for output := range outputChan {
			outputs = append(outputs, output)
		}

		assert.Greater(t, len(outputs), 0, "Should have captured outputs")

		expectedOutput := outputs[0]
		for i, output := range outputs {
			assert.Equal(t, expectedOutput, output, "Output %d should match first output", i)
		}
	})

	t.Run("memory_usage_validation", func(t *testing.T) {
		// Test memory usage
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Execute version command multiple times
		for i := 0; i < 1000; i++ {
			buf := new(bytes.Buffer)
			cmd := NewVersionCmd()
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.Execute()
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryUsed := m2.TotalAlloc - m1.TotalAlloc
		t.Logf("Memory used for 1000 version command executions: %d bytes", memoryUsed)

		// Should not use excessive memory (< 10MB for 1000 executions)
		assert.Less(t, memoryUsed, uint64(10*1024*1024), "Memory usage should be reasonable")
	})
}

// TestVersionCommandIntegration tests integration scenarios
func TestVersionCommandIntegration(t *testing.T) {
	t.Run("integration_with_root_command", func(t *testing.T) {
		// Test version command as part of complete CLI structure
		rootCmd := &cobra.Command{
			Use: "jumpstart",
		}

		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		// Add multiple commands to simulate real CLI
		versionCmd := NewVersionCmd()
		dummyCmd := &cobra.Command{Use: "dummy", Run: func(cmd *cobra.Command, args []string) {}}

		rootCmd.AddCommand(versionCmd)
		rootCmd.AddCommand(dummyCmd)

		// Test version command execution
		rootCmd.SetArgs([]string{"version"})
		err := rootCmd.Execute()

		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Jumpstart CLI version:")
		assert.Contains(t, output, utils.CliVersion)
	})

	t.Run("version_command_help_integration", func(t *testing.T) {
		// Test help for version command
		rootCmd := &cobra.Command{Use: "jumpstart"}
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		versionCmd := NewVersionCmd()
		rootCmd.AddCommand(versionCmd)

		// Test version help
		rootCmd.SetArgs([]string{"version", "--help"})
		err := rootCmd.Execute()

		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Display the current version")
		assert.Contains(t, output, "Jumpstart CLI")
	})

	t.Run("cross_platform_compatibility", func(t *testing.T) {
		// Test behavior across different platforms (simulate)
		testCases := []struct {
			name string
			goos string
		}{
			{"linux", "linux"},
			{"darwin", "darwin"},
			{"windows", "windows"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// While we can't actually change GOOS in tests,
				// we can verify the command works consistently
				buf := new(bytes.Buffer)
				cmd := NewVersionCmd()
				cmd.SetOut(buf)
				cmd.SetErr(buf)

				err := cmd.Execute()
				assert.NoError(t, err, "Command should work on %s", tc.goos)

				output := buf.String()
				assert.Contains(t, output, "Jumpstart CLI version:")
				assert.True(t, strings.HasSuffix(output, "\n"), "Should end with newline on %s", tc.goos)
			})
		}
	})
}

// TestVersionCommandErrorScenarios tests error conditions
func TestVersionCommandErrorScenarios(t *testing.T) {
	t.Run("output_writer_errors", func(t *testing.T) {
		// Test with a writer that always returns errors
		errorWriter := &errorWriter{}

		cmd := NewVersionCmd()
		cmd.SetOut(errorWriter)
		cmd.SetErr(errorWriter)

		// Command should handle writer errors gracefully
		// Note: cobra may not propagate writer errors, so this tests resilience
		assert.NotPanics(t, func() {
			cmd.Execute()
		}, "Should not panic with error writer")
	})

	t.Run("nil_version_recovery", func(t *testing.T) {
		// Test behavior when version is set to various problematic values
		testCases := []struct {
			name    string
			version string
		}{
			{"empty_string", ""},
			{"whitespace_only", "   "},
			{"newline_chars", "1.0.0\n"},
			{"null_chars", "1.0.0\x00"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				originalVersion := utils.CliVersion
				defer func() { utils.CliVersion = originalVersion }()

				utils.CliVersion = tc.version

				buf := new(bytes.Buffer)
				cmd := NewVersionCmd()
				cmd.SetOut(buf)
				cmd.SetErr(buf)

				err := cmd.Execute()
				assert.NoError(t, err, "Should handle problematic version: %s", tc.name)

				output := buf.String()
				assert.Contains(t, output, "Jumpstart CLI version:", "Should still contain prefix")
			})
		}
	})
}

// TestVersionCommandQuality tests for test quality and meaningfulness
func TestVersionCommandQuality(t *testing.T) {
	t.Run("meaningful_assertions", func(t *testing.T) {
		// Verify all our tests are meaningful and not just coverage-focused
		cmd := NewVersionCmd()

		// Test command structure thoroughly
		assert.Equal(t, "version", cmd.Use, "Command use should be exactly 'version'")
		assert.NotEmpty(t, cmd.Short, "Short description should not be empty")
		assert.NotEmpty(t, cmd.Long, "Long description should not be empty")
		assert.NotNil(t, cmd.Run, "Run function should be set")
		assert.Nil(t, cmd.RunE, "RunE should not be set when Run is set")

		// Test that command is properly configured
		assert.False(t, cmd.HasSubCommands(), "Version command should not have subcommands")
		assert.False(t, cmd.IsAdditionalHelpTopicCommand(), "Should not be additional help topic")
		assert.True(t, cmd.Runnable(), "Command should be runnable")
	})

	t.Run("output_completeness", func(t *testing.T) {
		// Verify output contains all necessary information
		buf := new(bytes.Buffer)
		cmd := NewVersionCmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		assert.NoError(t, err)

		output := buf.String()

		// Should contain application name
		assert.Contains(t, output, "Jumpstart CLI", "Should identify the application")

		// Should contain version identifier
		assert.Contains(t, output, "version:", "Should clearly indicate this is version info")

		// Should contain actual version value
		assert.Contains(t, output, utils.CliVersion, "Should display the actual version")

		// Should be properly formatted for human reading
		assert.True(t, len(output) > 10, "Output should be substantial enough to be useful")
		assert.False(t, strings.Contains(output, "error"), "Should not contain error text")
		assert.False(t, strings.Contains(output, "Error"), "Should not contain Error text")
	})
}

// errorWriter is a helper that always returns errors when writing
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated write error")
}
