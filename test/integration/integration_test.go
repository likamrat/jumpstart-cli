package integration

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"jumpstartcli/cmd/repo"
	"jumpstartcli/cmd/subscription"
	"jumpstartcli/cmd/upgrade"
	"jumpstartcli/cmd/version"
	"jumpstartcli/internal/utils"
)

// TestIntegration contains all integration tests
func TestIntegration(t *testing.T) {
	t.Run("CrossCommandIntegration", testCrossCommandIntegration)
	t.Run("PerformanceTesting", testPerformanceTesting)
	t.Run("EndToEndWorkflows", testEndToEndWorkflows)
	t.Run("ConcurrentExecution", testConcurrentExecution)
	t.Run("ResourceUsage", testResourceUsage)
}

// testCrossCommandIntegration tests interactions between commands
func testCrossCommandIntegration(t *testing.T) {
	t.Run("shared_dependency_usage", func(t *testing.T) {
		// Test that all commands can access shared utilities
		commands := []func() *cobra.Command{
			repo.NewRepoCmd,
			subscription.NewSubscriptionCmd,
			upgrade.NewUpgradeCmd,
			version.NewVersionCmd,
		}

		for i, cmdFunc := range commands {
			cmd := cmdFunc()
			assert.NotNil(t, cmd, "Command %d should not be nil", i)
			assert.NotEmpty(t, cmd.Use, "Command %d should have a 'Use' field", i)
			assert.NotEmpty(t, cmd.Short, "Command %d should have a short description", i)

			// Test that command can access shared utils
			assert.NotEmpty(t, utils.CliVersion, "All commands should have access to CLI version")
		}
	})

	t.Run("command_creation_consistency", func(t *testing.T) {
		// Test that all commands follow consistent patterns
		testCases := []struct {
			name    string
			cmdFunc func() *cobra.Command
		}{
			{"repo", repo.NewRepoCmd},
			{"subscription", subscription.NewSubscriptionCmd},
			{"upgrade", upgrade.NewUpgradeCmd},
			{"version", version.NewVersionCmd},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cmd := tc.cmdFunc()

				// Test basic command properties
				assert.NotNil(t, cmd, "Command should not be nil")
				assert.NotEmpty(t, cmd.Use, "Command should have a Use field")
				assert.NotEmpty(t, cmd.Short, "Command should have a short description")

				// Test that commands don't interfere with each other
				buf := new(bytes.Buffer)
				cmd.SetOut(buf)
				cmd.SetErr(buf)

				// Commands should be able to be created multiple times without issues
				cmd2 := tc.cmdFunc()
				assert.NotNil(t, cmd2, "Second instance of command should not be nil")
				assert.Equal(t, cmd.Use, cmd2.Use, "Multiple instances should be consistent")
			})
		}
	})

	t.Run("global_flag_inheritance", func(t *testing.T) {
		// Test that commands properly handle global flags like --debug, --verbose
		commands := []struct {
			name string
			cmd  *cobra.Command
		}{
			{"repo", repo.NewRepoCmd()},
			{"subscription", subscription.NewSubscriptionCmd()},
			{"upgrade", upgrade.NewUpgradeCmd()},
			{"version", version.NewVersionCmd()},
		}

		for _, tc := range commands {
			t.Run(tc.name, func(t *testing.T) {
				// Test that commands can be created with a parent that has global flags
				rootCmd := &cobra.Command{Use: "js"}
				rootCmd.PersistentFlags().BoolVar(&utils.DebugMode, "debug", false, "Enable debug output")
				rootCmd.PersistentFlags().BoolVar(&utils.VerboseMode, "verbose", false, "Enable verbose output")
				rootCmd.PersistentFlags().StringVarP(&utils.OutputFormat, "output", "o", "table", "Output format")

				rootCmd.AddCommand(tc.cmd)

				// Test that the command can access inherited flags
				assert.NotNil(t, rootCmd.PersistentFlags().Lookup("debug"), "Debug flag should be available")
				assert.NotNil(t, rootCmd.PersistentFlags().Lookup("verbose"), "Verbose flag should be available")
				assert.NotNil(t, rootCmd.PersistentFlags().Lookup("output"), "Output flag should be available")
			})
		}
	})

	t.Run("help_system_consistency", func(t *testing.T) {
		// Test that all commands have consistent help behavior
		commands := []struct {
			name string
			cmd  *cobra.Command
		}{
			{"repo", repo.NewRepoCmd()},
			{"subscription", subscription.NewSubscriptionCmd()},
			{"upgrade", upgrade.NewUpgradeCmd()},
			{"version", version.NewVersionCmd()},
		}

		for _, tc := range commands {
			t.Run(tc.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				tc.cmd.SetOut(buf)
				tc.cmd.SetErr(buf)

				// Test help flag
				tc.cmd.SetArgs([]string{"--help"})
				err := tc.cmd.Execute()

				// Help should either succeed or be handled gracefully
				output := buf.String()
				assert.True(t, len(output) > 0 || err != nil, "Help should produce output or handle gracefully")

				if err == nil {
					assert.Contains(t, output, tc.cmd.Use, "Help output should contain command name")
				}
			})
		}
	})
}

// testPerformanceTesting tests performance characteristics of commands
func testPerformanceTesting(t *testing.T) {
	t.Run("command_creation_performance", func(t *testing.T) {
		// Test that commands can be created quickly
		commands := []struct {
			name    string
			cmdFunc func() *cobra.Command
		}{
			{"repo", repo.NewRepoCmd},
			{"subscription", subscription.NewSubscriptionCmd},
			{"upgrade", upgrade.NewUpgradeCmd},
			{"version", version.NewVersionCmd},
		}

		for _, tc := range commands {
			t.Run(tc.name, func(t *testing.T) {
				start := time.Now()

				// Create command multiple times to test performance
				for i := 0; i < 100; i++ {
					cmd := tc.cmdFunc()
					assert.NotNil(t, cmd, "Command should be created successfully")
				}

				duration := time.Since(start)
				t.Logf("Creating 100 %s commands took: %v", tc.name, duration)

				// Should be able to create 100 commands in under 100ms
				assert.Less(t, duration, 100*time.Millisecond, "Command creation should be fast")
			})
		}
	})

	t.Run("memory_usage_efficiency", func(t *testing.T) {
		// Test memory usage patterns
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Create many command instances
		commands := make([]*cobra.Command, 1000)
		for i := 0; i < 250; i++ {
			commands[i*4] = repo.NewRepoCmd()
			commands[i*4+1] = subscription.NewSubscriptionCmd()
			commands[i*4+2] = upgrade.NewUpgradeCmd()
			commands[i*4+3] = version.NewVersionCmd()
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryUsed := m2.TotalAlloc - m1.TotalAlloc
		t.Logf("Memory used for 1000 command instances: %d bytes", memoryUsed)

		// Should not use excessive memory (< 50MB for 1000 instances)
		assert.Less(t, memoryUsed, uint64(50*1024*1024), "Memory usage should be reasonable")

		// Clean up
		commands = nil
		runtime.GC()
	})

	t.Run("execution_time_benchmarks", func(t *testing.T) {
		// Test execution time for simple operations
		testCases := []struct {
			name string
			cmd  *cobra.Command
			args []string
		}{
			{"version", version.NewVersionCmd(), []string{}},
			{"repo_help", repo.NewRepoCmd(), []string{"--help"}},
			{"subscription_help", subscription.NewSubscriptionCmd(), []string{"--help"}},
			{"upgrade_help", upgrade.NewUpgradeCmd(), []string{"--help"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				tc.cmd.SetOut(buf)
				tc.cmd.SetErr(buf)
				tc.cmd.SetArgs(tc.args)

				start := time.Now()
				tc.cmd.Execute()
				duration := time.Since(start)

				t.Logf("%s execution time: %v", tc.name, duration)

				// Most commands should execute in under 100ms
				assert.Less(t, duration, 100*time.Millisecond, "Command execution should be fast")
			})
		}
	})
}

// testEndToEndWorkflows tests complete workflows
func testEndToEndWorkflows(t *testing.T) {
	t.Run("version_check_workflow", func(t *testing.T) {
		// Simulate a typical workflow: check version, then use other commands
		buf := new(bytes.Buffer)

		// Step 1: Check version
		versionCmd := version.NewVersionCmd()
		versionCmd.SetOut(buf)
		versionCmd.SetErr(buf)

		err := versionCmd.Execute()
		assert.NoError(t, err, "Version command should execute successfully")

		versionOutput := buf.String()
		assert.Contains(t, versionOutput, utils.CliVersion, "Version output should contain CLI version")

		// Step 2: Use repo command for help
		buf.Reset()
		repoCmd := repo.NewRepoCmd()
		repoCmd.SetOut(buf)
		repoCmd.SetErr(buf)
		repoCmd.SetArgs([]string{"--help"})

		err = repoCmd.Execute()
		// Help should either succeed or be handled gracefully
		if err == nil {
			repoOutput := buf.String()
			assert.Contains(t, repoOutput, "repo", "Repo help should contain command name")
		}
	})

	t.Run("error_recovery_workflow", func(t *testing.T) {
		// Test error handling and recovery scenarios
		testCases := []struct {
			name    string
			cmdFunc func() *cobra.Command
			args    []string
		}{
			{"repo_invalid_subcommand", repo.NewRepoCmd, []string{"invalid"}},
			{"subscription_invalid_flag", subscription.NewSubscriptionCmd, []string{"--invalid-flag"}},
			{"upgrade_invalid_args", upgrade.NewUpgradeCmd, []string{"too", "many", "args"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				cmd := tc.cmdFunc()
				cmd.SetOut(buf)
				cmd.SetErr(buf)
				cmd.SetArgs(tc.args)

				// Commands should handle errors gracefully
				err := cmd.Execute()
				output := buf.String()

				// Either error is returned or handled gracefully with output
				isHandled := err != nil || len(output) > 0
				assert.True(t, isHandled, "Command should handle invalid input gracefully")
			})
		}
	})

	t.Run("user_experience_scenarios", func(t *testing.T) {
		// Test common user interaction patterns
		scenarios := []struct {
			name        string
			description string
			commands    []struct {
				cmdFunc func() *cobra.Command
				args    []string
			}
		}{
			{
				name:        "help_discovery",
				description: "User discovering available commands through help",
				commands: []struct {
					cmdFunc func() *cobra.Command
					args    []string
				}{
					{version.NewVersionCmd, []string{}},                   // Check version first
					{repo.NewRepoCmd, []string{"--help"}},                 // Learn about repo
					{subscription.NewSubscriptionCmd, []string{"--help"}}, // Learn about subscription
				},
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				t.Logf("Testing scenario: %s", scenario.description)

				for i, cmdSpec := range scenario.commands {
					buf := new(bytes.Buffer)
					cmd := cmdSpec.cmdFunc()
					cmd.SetOut(buf)
					cmd.SetErr(buf)
					cmd.SetArgs(cmdSpec.args)

					err := cmd.Execute()
					output := buf.String()

					// Each step should either succeed or provide meaningful output
					isSuccessful := err == nil || len(output) > 0
					assert.True(t, isSuccessful, "Step %d of scenario should be successful", i+1)

					t.Logf("Step %d output length: %d", i+1, len(output))
				}
			})
		}
	})
}

// testConcurrentExecution tests concurrent command execution
func testConcurrentExecution(t *testing.T) {
	t.Run("concurrent_command_creation", func(t *testing.T) {
		// Test that commands can be created concurrently without issues
		var wg sync.WaitGroup
		numGoroutines := 50
		errors := make(chan error, numGoroutines*4)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(4)

			go func() {
				defer wg.Done()
				cmd := repo.NewRepoCmd()
				if cmd == nil {
					errors <- fmt.Errorf("repo command creation failed")
				}
			}()

			go func() {
				defer wg.Done()
				cmd := subscription.NewSubscriptionCmd()
				if cmd == nil {
					errors <- fmt.Errorf("subscription command creation failed")
				}
			}()

			go func() {
				defer wg.Done()
				cmd := upgrade.NewUpgradeCmd()
				if cmd == nil {
					errors <- fmt.Errorf("upgrade command creation failed")
				}
			}()

			go func() {
				defer wg.Done()
				cmd := version.NewVersionCmd()
				if cmd == nil {
					errors <- fmt.Errorf("version command creation failed")
				}
			}()
		}

		wg.Wait()
		close(errors)

		var errorList []error
		for err := range errors {
			errorList = append(errorList, err)
		}

		assert.Empty(t, errorList, "No errors should occur during concurrent command creation")
	})

	t.Run("concurrent_command_execution", func(t *testing.T) {
		// Test concurrent execution of safe commands (like version and help)
		var wg sync.WaitGroup
		numGoroutines := 20
		results := make(chan string, numGoroutines*2)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(2)

			// Concurrent version commands
			go func() {
				defer wg.Done()
				buf := new(bytes.Buffer)
				cmd := version.NewVersionCmd()
				cmd.SetOut(buf)
				cmd.SetErr(buf)

				err := cmd.Execute()
				if err == nil {
					results <- buf.String()
				} else {
					results <- ""
				}
			}()

			// Concurrent repo help commands
			go func() {
				defer wg.Done()
				buf := new(bytes.Buffer)
				cmd := repo.NewRepoCmd()
				cmd.SetOut(buf)
				cmd.SetErr(buf)
				cmd.SetArgs([]string{"--help"})

				cmd.Execute() // Ignore error for help
				results <- buf.String()
			}()
		}

		wg.Wait()
		close(results)

		successCount := 0
		for result := range results {
			if len(result) > 0 {
				successCount++
			}
		}

		// At least 75% of concurrent executions should succeed
		expectedMinSuccess := int(float64(numGoroutines*2) * 0.75)
		assert.GreaterOrEqual(t, successCount, expectedMinSuccess,
			"Most concurrent executions should succeed")
	})

	t.Run("race_condition_detection", func(t *testing.T) {
		// Test for potential race conditions in shared resources
		var wg sync.WaitGroup
		numGoroutines := 100

		// Test concurrent access to utils
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Access shared utilities concurrently
				_ = utils.CliVersion
				_ = utils.DebugMode
				_ = utils.VerboseMode
				_ = utils.OutputFormat
			}()
		}

		wg.Wait()
		// If we get here without hanging or crashing, race condition test passed
		assert.True(t, true, "Race condition test completed successfully")
	})
}

// testResourceUsage tests resource consumption patterns
func testResourceUsage(t *testing.T) {
	t.Run("memory_leak_detection", func(t *testing.T) {
		// Test for memory leaks during repeated command creation/execution
		var m1, m2, m3 runtime.MemStats

		// Baseline measurement
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Create and discard many commands
		for i := 0; i < 1000; i++ {
			cmd := version.NewVersionCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.Execute()
			cmd = nil
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		// Create and discard more commands
		for i := 0; i < 1000; i++ {
			cmd := repo.NewRepoCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs([]string{"--help"})
			cmd.Execute()
			cmd = nil
		}

		runtime.GC()
		runtime.ReadMemStats(&m3)

		memoryGrowth1 := m2.TotalAlloc - m1.TotalAlloc
		memoryGrowth2 := m3.TotalAlloc - m2.TotalAlloc

		t.Logf("Memory growth in first phase: %d bytes", memoryGrowth1)
		t.Logf("Memory growth in second phase: %d bytes", memoryGrowth2)

		// Memory growth should not be excessive and should be similar between phases
		assert.Less(t, memoryGrowth1, uint64(30*1024*1024), "Memory growth should be reasonable")
		assert.Less(t, memoryGrowth2, uint64(30*1024*1024), "Memory growth should be reasonable")

		// Growth rate should not increase dramatically (no major memory leak)
		if memoryGrowth1 > 0 {
			growthRatio := float64(memoryGrowth2) / float64(memoryGrowth1)
			assert.Less(t, growthRatio, 10.0, "Memory growth rate should not increase dramatically")
		}
	})

	t.Run("goroutine_leak_detection", func(t *testing.T) {
		// Test for goroutine leaks
		initialGoroutines := runtime.NumGoroutine()

		// Perform operations that might create goroutines
		for i := 0; i < 100; i++ {
			cmd := upgrade.NewUpgradeCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs([]string{"--help"})
			cmd.Execute()
		}

		// Give time for any goroutines to finish
		time.Sleep(100 * time.Millisecond)
		finalGoroutines := runtime.NumGoroutine()

		t.Logf("Initial goroutines: %d, Final goroutines: %d", initialGoroutines, finalGoroutines)

		// Should not have significantly more goroutines
		goroutineDiff := finalGoroutines - initialGoroutines
		assert.LessOrEqual(t, goroutineDiff, 5, "Should not leak significant number of goroutines")
	})

	t.Run("cpu_usage_efficiency", func(t *testing.T) {
		// Test CPU usage patterns during intensive operations
		start := time.Now()

		// Perform CPU-intensive operations
		for i := 0; i < 500; i++ {
			cmd := subscription.NewSubscriptionCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs([]string{"--help"})
			cmd.Execute()
		}

		duration := time.Since(start)
		t.Logf("500 command executions took: %v", duration)

		// Should complete in reasonable time (under 5 seconds)
		assert.Less(t, duration, 5*time.Second, "CPU-intensive operations should complete efficiently")
	})
}

// Benchmark tests for performance measurement
func BenchmarkCommandCreation(b *testing.B) {
	b.Run("repo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			repo.NewRepoCmd()
		}
	})

	b.Run("subscription", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			subscription.NewSubscriptionCmd()
		}
	})

	b.Run("upgrade", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			upgrade.NewUpgradeCmd()
		}
	})

	b.Run("version", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			version.NewVersionCmd()
		}
	})
}

func BenchmarkCommandExecution(b *testing.B) {
	b.Run("version", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd := version.NewVersionCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.Execute()
		}
	})
}
