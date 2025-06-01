package completion

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
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

func TestNewCompletionCmd(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Completion Command Creation ==="))
	
	cmd := NewCompletionCmd()
	
	// Test basic command structure
	testName := "Command Use"
	success := cmd.Use == "completion"
	message := fmt.Sprintf("Expected 'completion', got '%s'", cmd.Use)
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

func TestCompletionCommandArgs(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Completion Command Arguments ==="))
	
	// Test argument validation (should accept 0-1 args)
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "valid bash argument",
			args:        []string{"bash"},
			expectError: false,
		},
		{
			name:        "valid zsh argument", 
			args:        []string{"zsh"},
			expectError: false,
		},
		{
			name:        "valid fish argument",
			args:        []string{"fish"},
			expectError: false,
		},
		{
			name:        "valid powershell argument",
			args:        []string{"powershell"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("Argument Test: %s", tt.name)
		
		// Create a fresh command for each test to avoid state issues
		testCmd := NewCompletionCmd()
		testCmd.SetArgs(tt.args)
		
		// Capture output
		var buf bytes.Buffer
		testCmd.SetOut(&buf)
		testCmd.SetErr(&buf)
		
		err := testCmd.Execute()
		
		success := true
		var message string
		
		if tt.expectError && err == nil {
			success = false
			message = "Expected error but got none"
		} else if !tt.expectError && err != nil {
			success = false
			message = fmt.Sprintf("Expected no error but got: %v", err)
		} else {
			message = fmt.Sprintf("Args %v processed correctly", tt.args)
		}
				printTestStatus(t, testName, success, message)
	}
}

func TestCompletionCommandInvalidArgs(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Completion Command Invalid Arguments ==="))
	
	tests := []struct {
		name          string
		args          []string
		expectContain string
	}{
		{
			name:          "invalid argument",
			args:          []string{"invalid"},
			expectContain: "[ERROR] invalid argument",
		},
		{
			name:          "too many arguments",
			args:          []string{"bash", "zsh"},
			expectContain: "accepts between 0 and 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("Invalid Args: %s", tt.name)
		
		cmd := NewCompletionCmd()
		cmd.SetArgs(tt.args)
		
		// Capture both stdout and stderr
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stdout = w
		os.Stderr = w
		
		// Execute command
		err := cmd.Execute()
		
		// Restore stdout/stderr and capture output
		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		
		var captured bytes.Buffer
		io.Copy(&captured, r)
		output := captured.String()
		
		success := true
		var message string
		
		// For the "too many arguments" case, expect an error from cobra
		if tt.name == "too many arguments" && err == nil {
			success = false
			message = "Expected error for too many arguments but got none"
		} else if !strings.Contains(output, tt.expectContain) {
			success = false
			message = fmt.Sprintf("Expected output to contain %q, got %q", tt.expectContain, output)
		} else {
			message = fmt.Sprintf("Invalid args %v handled correctly", tt.args)
		}
		
		printTestStatus(t, testName, success, message)
	}
}

func TestCompletionCommandValidShells(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Completion Command Valid Shells ==="))
	
	validShells := []string{"bash", "zsh", "fish", "powershell"}
	
	for _, shell := range validShells {
		t.Run("completion_"+shell, func(t *testing.T) {
			testName := fmt.Sprintf("Shell Completion: %s", shell)
			
			// Create a root command to attach completion to
			rootCmd := &cobra.Command{Use: "js"}
			cmd := NewCompletionCmd()
			rootCmd.AddCommand(cmd)
			
			// Set args
			cmd.SetArgs([]string{shell})
			
			// Capture stdout since completion generation writes there
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			
			err := cmd.Execute()
			
			// Restore stdout and capture output
			w.Close()
			os.Stdout = oldStdout
			
			var captured bytes.Buffer
			io.Copy(&captured, r)
			output := captured.String()
			
			success := true
			var message string
			
			if err != nil {
				success = false
				message = fmt.Sprintf("Unexpected error for shell %s: %v", shell, err)
			} else if len(output) == 0 {
				success = false
				message = fmt.Sprintf("Expected completion output for shell %s, got empty string", shell)
			} else {
				message = fmt.Sprintf("Shell %s completion generated successfully", shell)
			}
			
			printTestStatus(t, testName, success, message)
		})
	}
}

func TestCompletionCommandFlags(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Completion Command Flags ==="))
	
	cmd := NewCompletionCmd()
	
	testName := "Custom Flags Count"
	success := cmd.Flags().NFlag() == 0
	var message string
	if success {
		message = "Completion command correctly has no custom flags"
	} else {
		message = fmt.Sprintf("Completion command should not have custom flags, but has %d", cmd.Flags().NFlag())
	}
	printTestStatus(t, testName, success, message)
}

func TestCompletionCommandHelpOutput(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Completion Command Help Output ==="))
	
	cmd := NewCompletionCmd()
	cmd.SetArgs([]string{})
	
	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	
	// Execute with no arguments should show help
	err := cmd.Execute()
	
	testName := "Execute Without Error"
	success := err == nil
	var message string
	if success {
		message = "Command executed without error"
	} else {
		message = fmt.Sprintf("Unexpected error: %v", err)
	}
	printTestStatus(t, testName, success, message)
	
	if !success {
		return // Don't continue if execution failed
	}
	
	output := buf.String()
	
	// Help output should contain usage information
	expectedContents := []string{
		"completion",
		"Generate shell completion scripts",
		"Bash:",
		"Zsh:",
		"Fish:",
		"PowerShell:",
	}
	
	for _, expected := range expectedContents {
		testName = fmt.Sprintf("Help Contains: %s", expected)
		success = strings.Contains(output, expected)
		if success {
			message = fmt.Sprintf("Help output correctly contains %q", expected)
		} else {
			message = fmt.Sprintf("Help output should contain %q", expected)
		}
		printTestStatus(t, testName, success, message)
	}
}

func TestCompletionCommandStructure(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Completion Command Structure ==="))
	
	cmd := NewCompletionCmd()
	
	// Test Args field
	testName := "Args Field Not Nil"
	success := cmd.Args != nil
	var message string
	if success {
		message = "Args field is properly initialized"
	} else {
		message = "Completion command Args field should not be nil"
	}
	printTestStatus(t, testName, success, message)
	
	// Test that it accepts the right number of arguments
	// RangeArgs(0, 1) should accept 0 or 1 arguments
	testArgs := [][]string{
		{},           // 0 args - should be valid
		{"bash"},     // 1 arg - should be valid
		{"bash", "zsh"}, // 2 args - should be invalid for RangeArgs(0,1)
	}
	
	for i, args := range testArgs {
		expectError := i == 2 // Only the third case should error
		testName = fmt.Sprintf("Args Validation: %v", args)
		
		var err error
		if cmd.Args != nil {
			err = cmd.Args(cmd, args)
		}
		
		success = true
		if expectError && err == nil {
			success = false
			message = fmt.Sprintf("Expected error for args %v but got none", args)
		} else if !expectError && err != nil {
			success = false
			message = fmt.Sprintf("Unexpected error for args %v: %v", args, err)
		} else {
			if expectError {
				message = fmt.Sprintf("Args %v correctly rejected", args)
			} else {
				message = fmt.Sprintf("Args %v correctly accepted", args)
			}
		}
		
		printTestStatus(t, testName, success, message)
	}
}
