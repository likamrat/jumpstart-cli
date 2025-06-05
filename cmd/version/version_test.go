package version

import (
	"bytes"
	"strings"
	"testing"

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
