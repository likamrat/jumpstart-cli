package main

import (
"bytes"
"io"
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

// printTestStatus prints a colored status message for test operations
func printTestStatus(success bool, message string) {
	if success {
		color.Green("✅ %s", message)
	} else {
		color.Red("❌ %s", message)
	}
}

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	
	// Save original stdout
	originalStdout := os.Stdout
	
	// Replace stdout with our writer
	os.Stdout = w
	
	// Create a channel to receive the captured output
	outputChan := make(chan string)
	
	// Start a goroutine to read from the pipe
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outputChan <- buf.String()
	}()
	
	// Execute the function
	f()
	
	// Restore original stdout
	w.Close()
	os.Stdout = originalStdout
	
	// Get the captured output
	output := <-outputChan
	r.Close()
	
	return output
}

func TestPrintWelcome(t *testing.T) {
	color.Cyan("\n" + testHeaderColor("=== Testing main.go printWelcome Function ==="))
	
	t.Run("PrintWelcome", func(t *testing.T) {
color.Cyan(testInfoColor("Testing printWelcome output content and format..."))

// Capture the output of printWelcome function
output := captureOutput(func() {
			printWelcome()
		})
		
		// Verify the output is not empty
		if len(strings.TrimSpace(output)) == 0 {
			t.Error(testErrorColor("printWelcome produced no output"))
			printTestStatus(false, "Output length validation")
			return
		}
		printTestStatus(true, "Output length validation")
		
		// Test cases for expected content
		testCases := []struct {
			name     string
			expected string
			desc     string
		}{
			{
				name:     "ASCII_Art",
				expected: "jumpstart",
				desc:     "ASCII art contains 'jumpstart'",
			},
			{
				name:     "Help_Command",
				expected: "js --help",
				desc:     "Contains help command reference",
			},
			{
				name:     "GitHub_URL",
				expected: "https://github.com/Azure/jumpstart-sdk",
				desc:     "Contains GitHub repository URL",
			},
			{
				name:     "Azure_CLI_Note",
				expected: "Azure CLI",
				desc:     "Contains Azure CLI requirement note",
			},
			{
				name:     "Base_Commands_Header",
				expected: "Here are the base commands:",
				desc:     "Contains base commands header",
			},
			{
				name:     "Agora_Command",
				expected: "agora",
				desc:     "Lists agora command",
			},
			{
				name:     "ArcBox_Command",
				expected: "arcbox",
				desc:     "Lists arcbox command",
			},
			{
				name:     "Completion_Command",
				expected: "completion",
				desc:     "Lists completion command",
			},
			{
				name:     "LocalBox_Command",
				expected: "localbox",
				desc:     "Lists localbox command",
			},
			{
				name:     "Repo_Command",
				expected: "repo",
				desc:     "Lists repo command",
			},
			{
				name:     "Subscription_Command",
				expected: "subscription",
				desc:     "Lists subscription command",
			},
			{
				name:     "Upgrade_Command",
				expected: "upgrade",
				desc:     "Lists upgrade command",
			},
			{
				name:     "Version_Command",
				expected: "version",
				desc:     "Lists version command",
			},
			{
				name:     "Help_Usage",
				expected: "js <command> --help",
				desc:     "Contains command help usage instructions",
			},
		}
		
		// Check each test case
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
if !strings.Contains(output, tc.expected) {
t.Errorf(testErrorColor("Expected output to contain '%s' (%s), but it didn't"), tc.expected, tc.desc)
printTestStatus(false, tc.desc)
return
}
printTestStatus(true, tc.desc)
})
		}
		
		// Verify structure: ASCII art should come before commands
		asciiIndex := strings.Index(output, "jumpstart")
		commandsIndex := strings.Index(output, "Here are the base commands:")
		
		if asciiIndex == -1 {
			t.Error(testErrorColor("ASCII art not found in output"))
			printTestStatus(false, "ASCII art presence")
		} else if commandsIndex == -1 {
			t.Error(testErrorColor("Commands section not found in output"))
			printTestStatus(false, "Commands section presence")
		} else if asciiIndex >= commandsIndex {
			t.Error(testErrorColor("ASCII art should appear before commands section"))
			printTestStatus(false, "Content structure validation")
		} else {
			printTestStatus(true, "Content structure validation")
		}
		
		// Verify all expected commands are present in the correct format
		expectedCommands := []string{
			"agora         : Manage Jumpstart Agora automation",
			"arcbox        : Manage Jumpstart ArcBox automation",
			"completion    : Generate shell completion scripts",
			"localbox      : Manage Jumpstart LocalBox automation",
			"repo          : Manage Jumpstart user local source code repository",
			"subscription  : Manage Azure subscriptions",
			"upgrade       : Upgrade the Jumpstart CLI to the latest version",
			"version       : Display the current version of the CLI",
		}
		
		commandsFound := 0
		for _, cmd := range expectedCommands {
			if strings.Contains(output, cmd) {
				commandsFound++
			}
		}
		
		if commandsFound == len(expectedCommands) {
			printTestStatus(true, "All commands properly formatted and present")
		} else {
			t.Errorf(testErrorColor("Expected %d commands in proper format, found %d"), len(expectedCommands), commandsFound)
			printTestStatus(false, "Command formatting validation")
		}
		
		// Verify the output contains newlines for proper formatting
		if !strings.Contains(output, "\n") {
			t.Error(testErrorColor("Output should contain newlines for proper formatting"))
			printTestStatus(false, "Output formatting")
		} else {
			printTestStatus(true, "Output formatting validation")
		}
		
		color.Green(testInfoColor("✅ printWelcome function test completed successfully"))
	})
}

func TestPrintWelcomeOutputFormat(t *testing.T) {
	color.Cyan("\n" + testHeaderColor("=== Testing printWelcome Output Format ==="))
	
	t.Run("OutputFormat", func(t *testing.T) {
color.Cyan(testInfoColor("Testing printWelcome output formatting and structure..."))

output := captureOutput(func() {
			printWelcome()
		})
		
		// Count lines to ensure we have substantial output
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) < 15 {
			t.Errorf(testErrorColor("Expected at least 15 lines of output, got %d"), len(lines))
			printTestStatus(false, "Minimum output lines check")
		} else {
			printTestStatus(true, "Minimum output lines check")
		}
		
		// Verify it starts with newline and ASCII art
		if !strings.HasPrefix(output, "\n") {
			t.Error(testErrorColor("Output should start with a newline"))
			printTestStatus(false, "Output starts with newline")
		} else {
			printTestStatus(true, "Output starts with newline")
		}
		
		// Verify it contains proper spacing and formatting
		if !strings.Contains(output, "    agora") {
			t.Error(testErrorColor("Commands should be properly indented"))
			printTestStatus(false, "Command indentation")
		} else {
			printTestStatus(true, "Command indentation")
		}
		
		color.Green(testInfoColor("✅ printWelcome output format test completed successfully"))
	})
}

// Benchmark the printWelcome function
func BenchmarkPrintWelcome(b *testing.B) {
	color.Cyan(testHeaderColor("=== Benchmarking printWelcome Function ==="))
	
	// Capture output during benchmark to avoid polluting benchmark results
	originalStdout := os.Stdout
	devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	defer devNull.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os.Stdout = devNull
		printWelcome()
		os.Stdout = originalStdout
	}
	
	color.Green(testInfoColor("✅ printWelcome benchmark completed"))
}
