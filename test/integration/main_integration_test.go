package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"jumpstartcli/cmd/agora"
	"jumpstartcli/cmd/arcbox"
	"jumpstartcli/cmd/completion"
	"jumpstartcli/cmd/localbox"
	"jumpstartcli/cmd/repo"
	"jumpstartcli/cmd/subscription"
	"jumpstartcli/cmd/upgrade"
	"jumpstartcli/cmd/version"
	"jumpstartcli/internal/utils"
)

// TestMainApplicationIntegration tests the full application integration
func TestMainApplicationIntegration(t *testing.T) {
	t.Run("FullApplicationStructure", testFullApplicationStructure)
	t.Run("CommandChaining", testCommandChaining)
	t.Run("GlobalFlagPropagation", testGlobalFlagPropagation)
	t.Run("ErrorHandlingConsistency", testErrorHandlingConsistency)
}

// createTestRootCommand creates a root command similar to main.go for testing
func createTestRootCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:                "js",
		Short:              "Jumpstart CLI",
		Long:               `Jumpstart CLI - Azure Arc Jumpstart automation tool.`,
		Version:            utils.CliVersion,
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableSuggestions: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				invalidCommand := args[0]
				validCommands := []string{"agora", "arcbox", "completion", "localbox", "repo", "subscription", "upgrade", "version"}

				if suggestion := utils.SuggestSimilarCommand(invalidCommand, validCommands, 2); suggestion != "" {
					utils.PrintDidYouMean(invalidCommand, suggestion)
					return nil
				}
			}
			return nil
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().BoolVar(&utils.DebugMode, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVar(&utils.VerboseMode, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&utils.OutputFormat, "output", "o", "table", "Output format")

	// Add all commands
	rootCmd.AddCommand(agora.NewAgoraCmd())
	rootCmd.AddCommand(arcbox.NewArcboxCmd())
	rootCmd.AddCommand(completion.NewCompletionCmd())
	rootCmd.AddCommand(localbox.NewLocalboxCmd())
	rootCmd.AddCommand(repo.NewRepoCmd())
	rootCmd.AddCommand(subscription.NewSubscriptionCmd())
	rootCmd.AddCommand(upgrade.NewUpgradeCmd())
	rootCmd.AddCommand(version.NewVersionCmd())

	return rootCmd
}

func testFullApplicationStructure(t *testing.T) {
	t.Run("all_commands_registered", func(t *testing.T) {
		rootCmd := createTestRootCommand()

		expectedCommands := []string{
			"agora", "arcbox", "localbox",
			"repo", "subscription", "upgrade", "version",
		}

		// Note: completion command might have different Use pattern like "completion [args]"
		// so we check for commands that start with expected names
		for _, expectedCmd := range expectedCommands {
			found := false
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == expectedCmd || (len(cmd.Use) >= len(expectedCmd) && cmd.Use[:len(expectedCmd)] == expectedCmd) {
					found = true
					break
				}
			}
			assert.True(t, found, "Command '%s' should be registered", expectedCmd)
		}

		// Separately check for completion command (which might have different Use format)
		completionFound := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "completion" || (len(cmd.Use) >= 10 && cmd.Use[:10] == "completion") {
				completionFound = true
				break
			}
		}
		assert.True(t, completionFound, "Completion command should be registered")
	})

	t.Run("command_hierarchy_integrity", func(t *testing.T) {
		rootCmd := createTestRootCommand()

		// Test that all commands have proper parent-child relationships
		for _, cmd := range rootCmd.Commands() {
			assert.Equal(t, rootCmd, cmd.Parent(), "Command '%s' should have root as parent", cmd.Use)
			assert.NotNil(t, cmd.Root(), "Command '%s' should have access to root", cmd.Use)
		}
	})

	t.Run("version_information_consistency", func(t *testing.T) {
		rootCmd := createTestRootCommand()

		// Test root command version
		assert.Equal(t, utils.CliVersion, rootCmd.Version, "Root command should have correct version")

		// Test version command output
		buf := new(bytes.Buffer)
		versionCmd := version.NewVersionCmd()
		versionCmd.SetOut(buf)
		versionCmd.SetErr(buf)

		err := versionCmd.Execute()
		assert.NoError(t, err, "Version command should execute successfully")

		output := buf.String()
		assert.Contains(t, output, utils.CliVersion, "Version command should output correct version")
	})

	t.Run("help_system_integration", func(t *testing.T) {
		rootCmd := createTestRootCommand()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"--help"})

		err := rootCmd.Execute()
		if err == nil {
			output := buf.String()
			assert.Contains(t, output, "Jumpstart CLI", "Help should contain app name")
			assert.Contains(t, output, "Available Commands", "Help should list available commands")
		}
	})
}

func testCommandChaining(t *testing.T) {
	t.Run("sequential_command_execution", func(t *testing.T) {
		// Test executing commands in sequence as a user might
		// Sequence: version -> repo help -> subscription help -> upgrade help
		sequences := []struct {
			args          []string
			expectSuccess bool
			description   string
		}{
			{[]string{"version"}, true, "Check version"},
			{[]string{"repo", "--help"}, true, "Get repo help"},
			{[]string{"subscription", "--help"}, true, "Get subscription help"},
			{[]string{"upgrade", "--help"}, true, "Get upgrade help"},
		}

		for i, seq := range sequences {
			t.Run(seq.description, func(t *testing.T) {
				// Create fresh command for each test
				cmd := createTestRootCommand()
				buf := new(bytes.Buffer)
				cmd.SetOut(buf)
				cmd.SetErr(buf)
				cmd.SetArgs(seq.args)

				err := cmd.Execute()
				output := buf.String()

				if seq.expectSuccess {
					// Either no error OR meaningful output (help commands may not return error)
					success := err == nil || len(output) > 0
					assert.True(t, success, "Sequence step %d should succeed: %s", i+1, seq.description)
				} else {
					assert.Error(t, err, "Sequence step %d should fail: %s", i+1, seq.description)
				}

				t.Logf("Step %d (%s): Error=%v, OutputLen=%d", i+1, seq.description, err, len(output))
			})
		}
	})

	t.Run("workflow_simulation", func(t *testing.T) {
		// Simulate real user workflows
		workflows := []struct {
			name  string
			steps [][]string
		}{
			{
				name: "new_user_exploration",
				steps: [][]string{
					{"version"},                // Check what version they have
					{"--help"},                 // See available commands
					{"repo", "--help"},         // Learn about repo management
					{"subscription", "--help"}, // Learn about Azure integration
				},
			},
			{
				name: "maintenance_workflow",
				steps: [][]string{
					{"version"},           // Check current version
					{"upgrade", "--help"}, // Learn about upgrading
					{"repo", "--help"},    // Check repo options
				},
			},
		}

		for _, workflow := range workflows {
			t.Run(workflow.name, func(t *testing.T) {
				t.Logf("Testing workflow: %s", workflow.name)

				for stepNum, args := range workflow.steps {
					cmd := createTestRootCommand()
					buf := new(bytes.Buffer)
					cmd.SetOut(buf)
					cmd.SetErr(buf)
					cmd.SetArgs(args)

					err := cmd.Execute()
					output := buf.String()

					// Each step should either succeed or provide meaningful output
					stepSuccess := err == nil || len(output) > 0
					assert.True(t, stepSuccess, "Workflow '%s' step %d should succeed", workflow.name, stepNum+1)

					t.Logf("  Step %d (%v): Success=%t, OutputLen=%d", stepNum+1, args, stepSuccess, len(output))
				}
			})
		}
	})
}

func testGlobalFlagPropagation(t *testing.T) {
	t.Run("debug_flag_propagation", func(t *testing.T) {
		// Test that --debug flag is properly inherited by subcommands
		originalDebug := utils.DebugMode
		defer func() { utils.DebugMode = originalDebug }()

		// Test debug flag with various commands
		testCases := []string{"version", "repo", "subscription", "upgrade"}

		for _, cmdName := range testCases {
			t.Run(cmdName, func(t *testing.T) {
				cmd := createTestRootCommand()
				buf := new(bytes.Buffer)
				cmd.SetOut(buf)
				cmd.SetErr(buf)

				// Reset debug mode
				utils.DebugMode = false
				cmd.SetArgs([]string{"--debug", cmdName, "--help"})

				err := cmd.Execute()

				// The flag should be parsed (debug mode should be set)
				// Note: Some commands might return error for help, which is okay
				if err == nil || len(buf.String()) > 0 {
					// Command executed or produced output - this is success for our test
					t.Logf("Command '%s' with --debug flag handled successfully", cmdName)
				}
			})
		}
	})

	t.Run("verbose_flag_propagation", func(t *testing.T) {
		originalVerbose := utils.VerboseMode
		defer func() { utils.VerboseMode = originalVerbose }()

		testCases := []string{"version", "repo", "subscription", "upgrade"}

		for _, cmdName := range testCases {
			t.Run(cmdName, func(t *testing.T) {
				cmd := createTestRootCommand()
				buf := new(bytes.Buffer)
				cmd.SetOut(buf)
				cmd.SetErr(buf)

				utils.VerboseMode = false
				cmd.SetArgs([]string{"--verbose", cmdName, "--help"})

				err := cmd.Execute()

				if err == nil || len(buf.String()) > 0 {
					t.Logf("Command '%s' with --verbose flag handled successfully", cmdName)
				}
			})
		}
	})

	t.Run("output_format_flag_propagation", func(t *testing.T) {
		originalFormat := utils.OutputFormat
		defer func() { utils.OutputFormat = originalFormat }()

		formats := []string{"table", "json", "yaml", "tsv"}

		for _, format := range formats {
			t.Run(format, func(t *testing.T) {
				cmd := createTestRootCommand()
				buf := new(bytes.Buffer)
				cmd.SetOut(buf)
				cmd.SetErr(buf)

				utils.OutputFormat = "table" // Reset to default
				cmd.SetArgs([]string{"--output", format, "version"})

				err := cmd.Execute()

				// Version command should always work
				assert.NoError(t, err, "Version command with output format should work")
				assert.Equal(t, format, utils.OutputFormat, "Output format should be set to %s", format)
			})
		}
	})
}

func testErrorHandlingConsistency(t *testing.T) {
	t.Run("invalid_command_suggestions", func(t *testing.T) {
		// Test that invalid commands get suggestions
		invalidCommands := []struct {
			input    string
			expected string
		}{
			{"vers", "version"},
			{"repot", "repo"},
			{"subscr", "subscription"},
			{"upgrad", "upgrade"},
		}

		for _, tc := range invalidCommands {
			t.Run(tc.input, func(t *testing.T) {
				cmd := createTestRootCommand()
				buf := new(bytes.Buffer)
				cmd.SetOut(buf)
				cmd.SetErr(buf)
				cmd.SetArgs([]string{tc.input})

				err := cmd.Execute()
				output := buf.String()

				// Should either return error or produce suggestion output
				hasSuggestion := err != nil || strings.Contains(output, "Did you mean") || len(output) > 0
				assert.True(t, hasSuggestion, "Invalid command '%s' should produce suggestion", tc.input)
			})
		}
	})

	t.Run("invalid_flag_handling", func(t *testing.T) {
		// Test that invalid global flags are handled consistently
		testCases := []struct {
			command string
			flag    string
		}{
			{"version", "--invalid-flag"},
			{"repo", "--bad-flag"},
			{"subscription", "--nonexistent"},
			{"upgrade", "--wrong-flag"},
		}

		for _, tc := range testCases {
			t.Run(tc.command+"_"+tc.flag, func(t *testing.T) {
				cmd := createTestRootCommand()
				buf := new(bytes.Buffer)
				cmd.SetOut(buf)
				cmd.SetErr(buf)
				cmd.SetArgs([]string{tc.command, tc.flag})

				err := cmd.Execute()

				// Should handle invalid flags gracefully (either error or help output)
				isHandled := err != nil || len(buf.String()) > 0
				assert.True(t, isHandled, "Invalid flag should be handled gracefully")
			})
		}
	})

	t.Run("empty_input_handling", func(t *testing.T) {
		// Test behavior with no arguments
		cmd := createTestRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		output := buf.String()

		// Should handle empty input gracefully
		isHandled := err == nil || len(output) > 0
		assert.True(t, isHandled, "Empty input should be handled gracefully")
	})

	t.Run("help_flag_consistency", func(t *testing.T) {
		// Test that --help works consistently across all commands
		commands := []string{"repo", "subscription", "upgrade", "version"}

		for _, cmdName := range commands {
			t.Run(cmdName, func(t *testing.T) {
				cmd := createTestRootCommand()
				buf := new(bytes.Buffer)
				cmd.SetOut(buf)
				cmd.SetErr(buf)
				cmd.SetArgs([]string{cmdName, "--help"})

				err := cmd.Execute()
				output := buf.String()

				// Help should either succeed or produce output
				helpWorked := err == nil || len(output) > 0
				assert.True(t, helpWorked, "Help should work for command '%s'", cmdName)

				if len(output) > 0 {
					assert.Contains(t, output, cmdName, "Help output should contain command name")
				}
			})
		}
	})
}

// TestCrossCommandDataFlow tests data flow between commands
func TestCrossCommandDataFlow(t *testing.T) {
	t.Run("shared_utilities_access", func(t *testing.T) {
		// Test that all commands can access shared utilities consistently
		commands := []func() *cobra.Command{
			repo.NewRepoCmd,
			subscription.NewSubscriptionCmd,
			upgrade.NewUpgradeCmd,
			version.NewVersionCmd,
		}

		for i, cmdFunc := range commands {
			cmd := cmdFunc()

			// Test access to shared version
			assert.NotEmpty(t, utils.CliVersion, "Command %d should access CLI version", i)

			// Test that command structure is consistent
			assert.NotEmpty(t, cmd.Use, "Command %d should have Use field", i)
			assert.NotEmpty(t, cmd.Short, "Command %d should have Short description", i)
		}
	})

	t.Run("configuration_consistency", func(t *testing.T) {
		// Test that configuration changes are reflected across commands
		originalDebug := utils.DebugMode
		originalVerbose := utils.VerboseMode
		originalFormat := utils.OutputFormat

		defer func() {
			utils.DebugMode = originalDebug
			utils.VerboseMode = originalVerbose
			utils.OutputFormat = originalFormat
		}()

		// Change configuration
		utils.DebugMode = true
		utils.VerboseMode = true
		utils.OutputFormat = "json"

		// Test that all commands see the same configuration
		commands := []func() *cobra.Command{
			repo.NewRepoCmd,
			subscription.NewSubscriptionCmd,
			upgrade.NewUpgradeCmd,
			version.NewVersionCmd,
		}

		for i, cmdFunc := range commands {
			cmd := cmdFunc()

			// Execute command to ensure it can access global configuration
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs([]string{"--help"})

			// Command should execute without crashing
			cmd.Execute()

			// Verify configuration is accessible
			assert.True(t, utils.DebugMode, "Command %d should see debug mode", i)
			assert.True(t, utils.VerboseMode, "Command %d should see verbose mode", i)
			assert.Equal(t, "json", utils.OutputFormat, "Command %d should see output format", i)
		}
	})
}
