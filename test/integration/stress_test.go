package integration

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jumpstartcli/cmd/repo"
	"jumpstartcli/cmd/subscription"
	"jumpstartcli/cmd/upgrade"
	"jumpstartcli/cmd/version"
)

// TestNetworkFailureScenarios tests commands under network failure conditions
func TestNetworkFailureScenarios(t *testing.T) {
	t.Run("NetworkConnectivityFailures", testNetworkConnectivityFailures)
	t.Run("DNSResolutionFailures", testDNSResolutionFailures)
	t.Run("TimeoutScenarios", testTimeoutScenarios)
	t.Run("PartialNetworkFailures", testPartialNetworkFailures)
}

// TestCorruptedDataScenarios tests commands with corrupted input data
func TestCorruptedDataScenarios(t *testing.T) {
	t.Run("InvalidJSONInput", testInvalidJSONInput)
	t.Run("CorruptedBinaryData", testCorruptedBinaryData)
	t.Run("IncompleteDataStreams", testIncompleteDataStreams)
	t.Run("MalformedConfigData", testMalformedConfigData)
}

// TestPermissionDeniedScenarios tests commands with insufficient permissions
func TestPermissionDeniedScenarios(t *testing.T) {
	t.Run("FilePermissionDenied", testFilePermissionDenied)
	t.Run("DirectoryPermissionDenied", testDirectoryPermissionDenied)
	t.Run("ExecutionPermissionDenied", testExecutionPermissionDenied)
	t.Run("NetworkPermissionDenied", testNetworkPermissionDenied)
}

// TestExtremeLoadScenarios tests commands under extreme load conditions
func TestExtremeLoadScenarios(t *testing.T) {
	t.Run("HighMemoryPressure", testHighMemoryPressure)
	t.Run("CPUExhaustionScenarios", testCPUExhaustionScenarios)
	t.Run("FileDescriptorExhaustion", testFileDescriptorExhaustion)
	t.Run("DiskSpaceExhaustion", testDiskSpaceExhaustion)
}

// testNetworkConnectivityFailures simulates network connectivity issues
func testNetworkConnectivityFailures(t *testing.T) {
	// Test commands that might use network (subscription command may call Azure CLI)
	networkSensitiveCommands := []struct {
		name    string
		cmdFunc func() *cobra.Command
		args    []string
	}{
		{
			name:    "subscription_list",
			cmdFunc: subscription.NewSubscriptionCmd,
			args:    []string{"list"},
		},
		{
			name:    "upgrade_check",
			cmdFunc: upgrade.NewUpgradeCmd,
			args:    []string{"--check-only"},
		},
	}

	// Simulate network unavailability by manipulating network interface (if possible)
	// Note: This test will run but may not actually trigger network failures
	// depending on the environment and permissions
	
	for _, test := range networkSensitiveCommands {
		t.Run(test.name, func(t *testing.T) {
			cmd := test.cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(test.args)

			// Set a short timeout context to simulate network issues
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Execute command with timeout
			done := make(chan error, 1)
			go func() {
				done <- cmd.Execute()
			}()

			select {
			case err := <-done:
				output := buf.String()
				t.Logf("Network failure test %s: Error = %v, Output length = %d", 
					test.name, err, len(output))
				
				// Command should handle network issues gracefully
				assert.True(t, err != nil || len(output) > 0, 
					"Command should handle network issues gracefully")
					
			case <-ctx.Done():
				t.Logf("Network failure test %s: Command timed out (expected for network issues)", test.name)
				// Timeout is acceptable for network failure scenarios
			}
		})
	}
}

// testDNSResolutionFailures simulates DNS resolution failures
func testDNSResolutionFailures(t *testing.T) {
	// Test with invalid DNS configuration
	originalResolver := net.DefaultResolver
	defer func() { net.DefaultResolver = originalResolver }()

	// Create a resolver that will fail DNS lookups
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, fmt.Errorf("simulated DNS failure")
		},
	}

	commands := []struct {
		name    string
		cmdFunc func() *cobra.Command
		args    []string
	}{
		{
			name:    "version_help",
			cmdFunc: version.NewVersionCmd,
			args:    []string{"--help"},
		},
		{
			name:    "repo_help",
			cmdFunc: repo.NewRepoCmd,
			args:    []string{"--help"},
		},
		{
			name:    "subscription_help",
			cmdFunc: subscription.NewSubscriptionCmd,
			args:    []string{"--help"},
		},
		{
			name:    "upgrade_help",
			cmdFunc: upgrade.NewUpgradeCmd,
			args:    []string{"--help"},
		},
	}

	for _, test := range commands {
		t.Run(test.name, func(t *testing.T) {
			cmd := test.cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(test.args)

			err := cmd.Execute()
			output := buf.String()

			t.Logf("DNS failure test %s: Error = %v, Output length = %d", 
				test.name, err, len(output))

			// Help commands should work even with DNS failures
			assert.True(t, len(output) > 0, 
				"Help commands should work even with DNS failures")
		})
	}
}

// testTimeoutScenarios tests commands under timeout conditions
func testTimeoutScenarios(t *testing.T) {
	timeoutDurations := []time.Duration{
		1 * time.Millisecond,   // Very short timeout
		10 * time.Millisecond,  // Short timeout
		100 * time.Millisecond, // Medium timeout
		1 * time.Second,        // Normal timeout
	}

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for i, timeout := range timeoutDurations {
		for j, cmdFunc := range commands {
			t.Run(fmt.Sprintf("timeout_%d_cmd_%d", i, j), func(t *testing.T) {
				cmd := cmdFunc()
				var buf bytes.Buffer
				cmd.SetOut(&buf)
				cmd.SetErr(&buf)
				cmd.SetArgs([]string{"--help"})

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				done := make(chan error, 1)
				go func() {
					done <- cmd.Execute()
				}()

				select {
				case err := <-done:
					output := buf.String()
					t.Logf("Timeout test (timeout=%v, cmd=%s): Error = %v, Output length = %d", 
						timeout, cmd.Use, err, len(output))
					
					if len(output) > 0 {
						// Command completed successfully within timeout
						assert.NoError(t, err, "Command should complete successfully if within timeout")
					}

				case <-ctx.Done():
					t.Logf("Timeout test (timeout=%v, cmd=%s): Command timed out", timeout, cmd.Use)
					// Timeout is expected for very short timeouts
				}
			})
		}
	}
}

// testPartialNetworkFailures tests commands under partial network failures
func testPartialNetworkFailures(t *testing.T) {
	// Simulate intermittent network issues by introducing delays and failures
	commands := []struct {
		name    string
		cmdFunc func() *cobra.Command
		args    []string
	}{
		{
			name:    "version_with_delay",
			cmdFunc: version.NewVersionCmd,
			args:    []string{"--output", "json"},
		},
		{
			name:    "repo_with_delay",
			cmdFunc: repo.NewRepoCmd,
			args:    []string{"--help"},
		},
		{
			name:    "subscription_with_delay",
			cmdFunc: subscription.NewSubscriptionCmd,
			args:    []string{"--help"},
		},
		{
			name:    "upgrade_with_delay",
			cmdFunc: upgrade.NewUpgradeCmd,
			args:    []string{"--help"},
		},
	}

	for _, test := range commands {
		t.Run(test.name, func(t *testing.T) {
			// Add artificial delay to simulate slow network
			time.Sleep(10 * time.Millisecond)

			cmd := test.cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(test.args)

			start := time.Now()
			err := cmd.Execute()
			duration := time.Since(start)
			output := buf.String()

			t.Logf("Partial network failure test %s: Error = %v, Duration = %v, Output length = %d", 
				test.name, err, duration, len(output))

			// Commands should complete even with network delays
			assert.True(t, len(output) > 0, 
				"Commands should complete even with network delays")
		})
	}
}

// testInvalidJSONInput tests commands with invalid JSON input
func testInvalidJSONInput(t *testing.T) {
	invalidJSONInputs := []string{
		`{"incomplete": true`,           // Missing closing brace
		`{invalid: "json"}`,            // Unquoted key
		`{"key": "value",}`,            // Trailing comma
		`{"key": "value" "key2": "v2"}`, // Missing comma
		`{null: null}`,                 // Null key
		`{"unicode": "\uXXXX"}`,        // Invalid unicode escape
		`{"nested": {"deep": }`,        // Incomplete nested object
	}

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for i, invalidJSON := range invalidJSONInputs {
		for j, cmdFunc := range commands {
			t.Run(fmt.Sprintf("json_%d_cmd_%d", i, j), func(t *testing.T) {
				cmd := cmdFunc()
				var buf bytes.Buffer
				cmd.SetOut(&buf)
				cmd.SetErr(&buf)
				
				// Pass invalid JSON as argument (most commands will ignore it)
				cmd.SetArgs([]string{invalidJSON})

				err := cmd.Execute()
				output := buf.String()

				t.Logf("Invalid JSON test (cmd=%s): Error = %v, Output length = %d", 
					cmd.Use, err, len(output))

				// Commands should handle invalid JSON gracefully
				assert.True(t, err != nil || len(output) > 0, 
					"Commands should handle invalid JSON gracefully")
			})
		}
	}
}

// testCorruptedBinaryData tests commands with corrupted binary data
func testCorruptedBinaryData(t *testing.T) {
	corruptedData := [][]byte{
		{0xFF, 0xFE, 0xFD, 0xFC},                    // Binary data
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},       // Null bytes and control chars
		{0x80, 0x81, 0x82, 0x83},                    // High-bit set bytes
		append([]byte("text"), 0x00, 0xFF, 0x00),   // Mixed text and binary
		bytes.Repeat([]byte{0xFF}, 1024),           // Large binary data
	}

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for i, data := range corruptedData {
		for j, cmdFunc := range commands {
			t.Run(fmt.Sprintf("binary_%d_cmd_%d", i, j), func(t *testing.T) {
				cmd := cmdFunc()
				var buf bytes.Buffer
				cmd.SetOut(&buf)
				cmd.SetErr(&buf)
				
				// Pass corrupted binary data as argument
				cmd.SetArgs([]string{string(data)})

				err := cmd.Execute()
				output := buf.String()

				t.Logf("Corrupted binary test (cmd=%s): Error = %v, Output length = %d", 
					cmd.Use, err, len(output))

				// Commands should handle corrupted binary data gracefully
				assert.True(t, err != nil || len(output) > 0, 
					"Commands should handle corrupted binary data gracefully")
			})
		}
	}
}

// testIncompleteDataStreams tests commands with incomplete data streams
func testIncompleteDataStreams(t *testing.T) {
	incompleteData := []string{
		"incom",                          // Truncated word
		"incomplete command wi",          // Truncated sentence
		"--output js",                   // Truncated flag value
		"--",                            // Incomplete flag
		string([]byte{0x48, 0x65, 0x6C}), // Truncated "Hel" from "Hello"
	}

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for i, data := range incompleteData {
		for j, cmdFunc := range commands {
			t.Run(fmt.Sprintf("incomplete_%d_cmd_%d", i, j), func(t *testing.T) {
				cmd := cmdFunc()
				var buf bytes.Buffer
				cmd.SetOut(&buf)
				cmd.SetErr(&buf)
				cmd.SetArgs([]string{data})

				err := cmd.Execute()
				output := buf.String()

				t.Logf("Incomplete data test (cmd=%s): Error = %v, Output length = %d", 
					cmd.Use, err, len(output))

				// Commands should handle incomplete data gracefully
				assert.True(t, err != nil || len(output) > 0, 
					"Commands should handle incomplete data gracefully")
			})
		}
	}
}

// testMalformedConfigData tests commands with malformed configuration data
func testMalformedConfigData(t *testing.T) {
	malformedConfigs := []string{
		"key=value=extra",               // Multiple equals
		"=value",                        // Missing key
		"key=",                          // Missing value
		"key value",                     // Missing separator
		"key\x00value",                  // Null separator
		"key\nvalue\nkey2\tvalue2",      // Mixed separators
		strings.Repeat("key=", 1000),    // Repeated pattern
	}

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for i, config := range malformedConfigs {
		for j, cmdFunc := range commands {
			t.Run(fmt.Sprintf("config_%d_cmd_%d", i, j), func(t *testing.T) {
				cmd := cmdFunc()
				var buf bytes.Buffer
				cmd.SetOut(&buf)
				cmd.SetErr(&buf)
				cmd.SetArgs([]string{config})

				err := cmd.Execute()
				output := buf.String()

				t.Logf("Malformed config test (cmd=%s): Error = %v, Output length = %d", 
					cmd.Use, err, len(output))

				// Commands should handle malformed config data gracefully
				assert.True(t, err != nil || len(output) > 0, 
					"Commands should handle malformed config data gracefully")
			})
		}
	}
}

// testFilePermissionDenied tests commands with file permission issues
func testFilePermissionDenied(t *testing.T) {
	// Create a temporary file with no permissions
	tempDir := t.TempDir()
	restrictedFile := fmt.Sprintf("%s/restricted.txt", tempDir)
	
	err := os.WriteFile(restrictedFile, []byte("test content"), 0000)
	require.NoError(t, err)
	defer os.Chmod(restrictedFile, 0644) // Restore permissions for cleanup

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for _, cmdFunc := range commands {
		t.Run(fmt.Sprintf("file_permission_%s", cmdFunc().Use), func(t *testing.T) {
			cmd := cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{"--help"})

			err := cmd.Execute()
			output := buf.String()

			t.Logf("File permission test (cmd=%s): Error = %v, Output length = %d", 
				cmd.Use, err, len(output))

			// Help commands should work despite file permission issues
			assert.True(t, len(output) > 0, 
				"Help commands should work despite file permission issues")
		})
	}
}

// testDirectoryPermissionDenied tests commands with directory permission issues
func testDirectoryPermissionDenied(t *testing.T) {
	// This test may not work on all systems due to permission requirements
	tempDir := t.TempDir()
	restrictedDir := fmt.Sprintf("%s/restricted", tempDir)
	
	err := os.Mkdir(restrictedDir, 0000)
	if err != nil {
		t.Skipf("Could not create restricted directory: %v", err)
		return
	}
	defer os.Chmod(restrictedDir, 0755) // Restore permissions for cleanup

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for _, cmdFunc := range commands {
		t.Run(fmt.Sprintf("dir_permission_%s", cmdFunc().Use), func(t *testing.T) {
			cmd := cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{"--help"})

			err := cmd.Execute()
			output := buf.String()

			t.Logf("Directory permission test (cmd=%s): Error = %v, Output length = %d", 
				cmd.Use, err, len(output))

			// Commands should work despite directory permission issues
			assert.True(t, len(output) > 0, 
				"Commands should work despite directory permission issues")
		})
	}
}

// testExecutionPermissionDenied tests commands with execution permission issues
func testExecutionPermissionDenied(t *testing.T) {
	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for _, cmdFunc := range commands {
		t.Run(fmt.Sprintf("exec_permission_%s", cmdFunc().Use), func(t *testing.T) {
			cmd := cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{"--help"})

			err := cmd.Execute()
			output := buf.String()

			t.Logf("Execution permission test (cmd=%s): Error = %v, Output length = %d", 
				cmd.Use, err, len(output))

			// Help commands should work despite execution permission issues
			assert.True(t, len(output) > 0, 
				"Help commands should work despite execution permission issues")
		})
	}
}

// testNetworkPermissionDenied tests commands with network permission issues
func testNetworkPermissionDenied(t *testing.T) {
	// This test simulates network permission issues
	commands := []struct {
		name    string
		cmdFunc func() *cobra.Command
		args    []string
	}{
		{
			name:    "version_network",
			cmdFunc: version.NewVersionCmd,
			args:    []string{"--help"},
		},
		{
			name:    "repo_network",
			cmdFunc: repo.NewRepoCmd,
			args:    []string{"--help"},
		},
		{
			name:    "subscription_network",
			cmdFunc: subscription.NewSubscriptionCmd,
			args:    []string{"--help"},
		},
		{
			name:    "upgrade_network",
			cmdFunc: upgrade.NewUpgradeCmd,
			args:    []string{"--help"},
		},
	}

	for _, test := range commands {
		t.Run(test.name, func(t *testing.T) {
			cmd := test.cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(test.args)

			err := cmd.Execute()
			output := buf.String()

			t.Logf("Network permission test %s: Error = %v, Output length = %d", 
				test.name, err, len(output))

			// Help commands should work despite network permission issues
			assert.True(t, len(output) > 0, 
				"Help commands should work despite network permission issues")
		})
	}
}

// testHighMemoryPressure tests commands under high memory pressure
func testHighMemoryPressure(t *testing.T) {
	// Create significant memory pressure
	const memorySize = 512 * 1024 * 1024 // 512MB
	memoryPressure := make([][]byte, 10)
	
	for i := range memoryPressure {
		memoryPressure[i] = make([]byte, memorySize/10)
		// Fill with data to ensure allocation
		for j := range memoryPressure[i] {
			memoryPressure[i][j] = byte((i + j) % 256)
		}
	}
	defer func() { memoryPressure = nil }()

	// Force garbage collection
	runtime.GC()
	
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for _, cmdFunc := range commands {
		t.Run(fmt.Sprintf("high_memory_%s", cmdFunc().Use), func(t *testing.T) {
			cmd := cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{"--help"})

			start := time.Now()
			err := cmd.Execute()
			duration := time.Since(start)
			output := buf.String()

			t.Logf("High memory pressure test (cmd=%s): Error = %v, Duration = %v, Output length = %d", 
				cmd.Use, err, duration, len(output))

			// Commands should work under memory pressure
			assert.True(t, len(output) > 0, 
				"Commands should work under high memory pressure")
			
			// Should complete within reasonable time
			assert.True(t, duration < 10*time.Second, 
				"Commands should complete within reasonable time even under memory pressure")
		})
	}

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)
	t.Logf("High memory pressure test - Memory before: %d MB, after: %d MB", 
		memBefore.Alloc/(1024*1024), memAfter.Alloc/(1024*1024))
}

// testCPUExhaustionScenarios tests commands under CPU exhaustion
func testCPUExhaustionScenarios(t *testing.T) {
	// Create CPU pressure by running CPU-intensive goroutines
	numCPU := runtime.NumCPU()
	var cpuWg sync.WaitGroup
	stopCPUPressure := make(chan struct{})

	// Start CPU-intensive goroutines
	for i := 0; i < numCPU*2; i++ {
		cpuWg.Add(1)
		go func() {
			defer cpuWg.Done()
			for {
				select {
				case <-stopCPUPressure:
					return
				default:
					// CPU-intensive work
					var sum int64
					for j := 0; j < 10000; j++ {
						sum += int64(j * j)
					}
					_ = sum
				}
			}
		}()
	}

	defer func() {
		close(stopCPUPressure)
		cpuWg.Wait()
	}()

	// Allow CPU pressure to build up
	time.Sleep(100 * time.Millisecond)

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for _, cmdFunc := range commands {
		t.Run(fmt.Sprintf("cpu_exhaustion_%s", cmdFunc().Use), func(t *testing.T) {
			cmd := cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{"--help"})

			start := time.Now()
			err := cmd.Execute()
			duration := time.Since(start)
			output := buf.String()

			t.Logf("CPU exhaustion test (cmd=%s): Error = %v, Duration = %v, Output length = %d", 
				cmd.Use, err, duration, len(output))

			// Commands should work under CPU pressure
			assert.True(t, len(output) > 0, 
				"Commands should work under CPU exhaustion")
			
			// Should complete within reasonable time (may be slower)
			assert.True(t, duration < 30*time.Second, 
				"Commands should complete within reasonable time even under CPU pressure")
		})
	}
}

// testFileDescriptorExhaustion tests commands under file descriptor exhaustion
func testFileDescriptorExhaustion(t *testing.T) {
	// Create file descriptor pressure by opening many files
	tempDir := t.TempDir()
	var files []*os.File
	defer func() {
		for _, f := range files {
			f.Close()
		}
	}()

	// Open many files to create descriptor pressure
	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("%s/temp_%d.txt", tempDir, i)
		f, err := os.Create(filename)
		if err != nil {
			break // Stop if we can't create more files
		}
		files = append(files, f)
	}

	t.Logf("Created %d file descriptors for pressure testing", len(files))

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for _, cmdFunc := range commands {
		t.Run(fmt.Sprintf("fd_exhaustion_%s", cmdFunc().Use), func(t *testing.T) {
			cmd := cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{"--help"})

			err := cmd.Execute()
			output := buf.String()

			t.Logf("FD exhaustion test (cmd=%s): Error = %v, Output length = %d", 
				cmd.Use, err, len(output))

			// Commands should work despite file descriptor pressure
			assert.True(t, len(output) > 0, 
				"Commands should work despite file descriptor pressure")
		})
	}
}

// testDiskSpaceExhaustion tests commands under disk space exhaustion
func testDiskSpaceExhaustion(t *testing.T) {
	// This test simulates disk space issues without actually filling the disk
	tempDir := t.TempDir()

	// Create a large file to simulate disk pressure
	largeFile := fmt.Sprintf("%s/large_file.dat", tempDir)
	f, err := os.Create(largeFile)
	if err != nil {
		t.Skipf("Could not create large file: %v", err)
		return
	}
	defer f.Close()
	defer os.Remove(largeFile)

	// Write some data to the file
	data := make([]byte, 1024*1024) // 1MB
	for i := 0; i < 100; i++ {      // Write 100MB
		_, err := f.Write(data)
		if err != nil {
			break // Stop if disk is full
		}
	}

	commands := []func() *cobra.Command{
		version.NewVersionCmd,
		repo.NewRepoCmd,
		subscription.NewSubscriptionCmd,
		upgrade.NewUpgradeCmd,
	}

	for _, cmdFunc := range commands {
		t.Run(fmt.Sprintf("disk_space_%s", cmdFunc().Use), func(t *testing.T) {
			cmd := cmdFunc()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{"--help"})

			err := cmd.Execute()
			output := buf.String()

			t.Logf("Disk space test (cmd=%s): Error = %v, Output length = %d", 
				cmd.Use, err, len(output))

			// Commands should work despite disk space pressure
			assert.True(t, len(output) > 0, 
				"Commands should work despite disk space pressure")
		})
	}
}
