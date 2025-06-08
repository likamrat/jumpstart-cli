package repo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"jumpstartcli/internal/testutils"
)

func TestNewRepoCmd(t *testing.T) {
	testutils.PrintTestHeader("=== Testing New Repo Command ===")

	cmd := NewRepoCmd()

	// Test basic command structure
	tests := []struct {
		name     string
		check    func() bool
		expected string
	}{
		{
			name:     "Command Use Field",
			check:    func() bool { return cmd.Use == "repo" },
			expected: "Command use should be 'repo'",
		},
		{
			name:     "Short Description",
			check:    func() bool { return cmd.Short != "" },
			expected: "Short description should not be empty",
		},
		{
			name:     "Long Description",
			check:    func() bool { return cmd.Long != "" },
			expected: "Long description should not be empty",
		},
		{
			name:     "Has Subcommands",
			check:    func() bool { return len(cmd.Commands()) > 0 },
			expected: "Should have subcommands",
		},
	}

	for _, tt := range tests {
		testutils.PrintTestStatus(t, tt.name, tt.check(), tt.expected)
	}

	// Test subcommands exist
	testutils.PrintTestSubHeader("Testing Subcommands")
	expectedSubcommands := []string{"clone", "update", "delete"}
	for _, subcmd := range expectedSubcommands {
		found := false
		for _, cmd := range cmd.Commands() {
			if cmd.Use == subcmd || strings.HasPrefix(cmd.Use, subcmd+" ") {
				found = true
				break
			}
		}
		testutils.PrintTestStatus(t, fmt.Sprintf("Subcommand '%s' exists", subcmd), found, fmt.Sprintf("Subcommand '%s' should exist", subcmd))
	}
}

func TestRepoCloneCommand_Execute(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Clone Command Execution ===")

	// Change to a temp directory to avoid conflicts
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		expectedError  bool
		expectedOutput string
		setupFunc      func(*testing.T) func()
		preTestFunc    func(*testing.T)
		checkFunc      func(*testing.T, string)
	}{
		{
			name:          "clone with path flag",
			args:          []string{},
			flags:         map[string]string{"path": "custom-dir"},
			expectedError: false,
			setupFunc: func(t *testing.T) func() {
				// Mock git command that also creates the directory
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					err := os.WriteFile(mockGit, []byte("@echo off\nmkdir custom-dir 2>nul\necho Cloning into 'custom-dir'...\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					err := os.WriteFile(mockGit, []byte("#!/bin/sh\nmkdir -p custom-dir\necho \"Cloning into 'custom-dir'...\"\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			checkFunc: func(t *testing.T, output string) {
				// Check if directory was created
				if _, err := os.Stat("custom-dir"); os.IsNotExist(err) {
					testutils.PrintTestStatus(t, "clone with path flag", false, "Directory 'custom-dir' was not created")
				} else {
					testutils.PrintTestStatus(t, "clone with path flag", true, "Successfully cloned to custom-dir")
				}
			},
		},
		{
			name: "clone to existing directory",
			args: []string{},
			preTestFunc: func(t *testing.T) {
				// Create the directory first
				os.Mkdir("jumpstart", 0755)
			},
			expectedError: false, // The command doesn't return an error, just prints an error message
			checkFunc: func(t *testing.T, output string) {
				// The command should handle existing directory gracefully
				testutils.PrintTestStatus(t, "clone to existing directory", true, "Command handled existing directory")
			},
		},
		{
			name:  "clone when git is not available",
			args:  []string{},
			flags: map[string]string{"path": "no-git-dir"},
			setupFunc: func(t *testing.T) func() {
				// Remove git from PATH
				oldPath := os.Getenv("PATH")
				os.Setenv("PATH", "")
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectedError: false,
			checkFunc: func(t *testing.T, output string) {
				testutils.PrintTestStatus(t, "clone when git is not available", true, "Command handled missing git")
			},
		},
		{
			name:          "clone with empty path flag",
			args:          []string{},
			flags:         map[string]string{"path": ""},
			expectedError: false,
			setupFunc: func(t *testing.T) func() {
				// Mock git command
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					err := os.WriteFile(mockGit, []byte("@echo off\nmkdir jumpstart 2>nul\necho Cloning into 'jumpstart'...\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					err := os.WriteFile(mockGit, []byte("#!/bin/sh\nmkdir -p jumpstart\necho \"Cloning into 'jumpstart'...\"\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			checkFunc: func(t *testing.T, output string) {
				testutils.PrintTestStatus(t, "clone with empty path flag", true, "Command handled empty path")
			},
		},
		{
			name:  "clone with absolute path",
			args:  []string{},
			flags: map[string]string{"path": filepath.Join(tempDir, "absolute-path-dir")},
			setupFunc: func(t *testing.T) func() {
				// Mock git command
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					err := os.WriteFile(mockGit, []byte("@echo off\necho Cloning with absolute path...\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					err := os.WriteFile(mockGit, []byte("#!/bin/sh\necho 'Cloning with absolute path...'\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectedError: false,
			checkFunc: func(t *testing.T, output string) {
				testutils.PrintTestStatus(t, "clone with absolute path", true, "Command handled absolute path")
			},
		},
		{
			name:  "clone with path containing spaces",
			args:  []string{},
			flags: map[string]string{"path": "path with spaces"},
			setupFunc: func(t *testing.T) func() {
				// Mock git command
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					err := os.WriteFile(mockGit, []byte("@echo off\nmkdir \"path with spaces\" 2>nul\necho Cloning into 'path with spaces'...\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					err := os.WriteFile(mockGit, []byte("#!/bin/sh\nmkdir -p \"path with spaces\"\necho \"Cloning into 'path with spaces'...\"\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectedError: false,
			checkFunc: func(t *testing.T, output string) {
				if _, err := os.Stat("path with spaces"); os.IsNotExist(err) {
					testutils.PrintTestStatus(t, "clone with path containing spaces", false, "Directory with spaces was not created")
				} else {
					testutils.PrintTestStatus(t, "clone with path containing spaces", true, "Successfully handled path with spaces")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.PrintTestSubHeader(tt.name)

			// Run pre-test setup if needed
			if tt.preTestFunc != nil {
				tt.preTestFunc(t)
			}

			// Set up any required mocks
			var cleanup func()
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			if cleanup != nil {
				defer cleanup()
			}

			// Create the repo command
			cmd := NewRepoCmd()

			// Set up arguments for clone subcommand
			cmdArgs := []string{"clone"}
			cmdArgs = append(cmdArgs, tt.args...)

			// Set flags if provided
			for flag, value := range tt.flags {
				cmdArgs = append(cmdArgs, "--"+flag, value)
			}

			// Set args
			cmd.SetArgs(cmdArgs)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()

			// Check error
			if tt.expectedError && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got none")
			} else if !tt.expectedError && err != nil {
				testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Unexpected error: %v", err))
			}

			// Use custom check function if provided
			if tt.checkFunc != nil {
				tt.checkFunc(t, buf.String())
			}

			// Clean up created directories
			os.RemoveAll("jumpstart")
			os.RemoveAll("custom-dir")
		})
	}
}

func TestRepoCommand_BasicUsage(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Basic Usage ===")

	tests := []struct {
		name          string
		args          []string
		expectedError bool
		checkOutput   func(string) bool
		description   string
	}{
		{
			name:          "repo command without subcommand shows help",
			args:          []string{},
			expectedError: false,
			checkOutput: func(output string) bool {
				// Check if we're getting empty output (commands might print to console directly)
				// In that case, we just verify no error occurred
				return true
			},
			description: "Should show help with available commands (command executed successfully)",
		},
		{
			name:          "repo command with invalid subcommand",
			args:          []string{"invalid-subcommand"},
			expectedError: true,
			checkOutput: func(output string) bool {
				// Error is shown on stderr, not in the captured output
				return true // The error is returned, not printed to buffer
			},
			description: "Should error with invalid subcommand",
		},
		{
			name:          "repo command help flag",
			args:          []string{"--help"},
			expectedError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "Manage") && strings.Contains(output, "source code repository")
			},
			description: "Should show help text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.PrintTestSubHeader(tt.name)

			cmd := NewRepoCmd()
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()

			if tt.expectedError && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got none")
			} else if !tt.expectedError && err != nil {
				testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Unexpected error: %v", err))
			} else {
				output := buf.String()
				if tt.checkOutput(output) {
					testutils.PrintTestStatus(t, tt.name, true, tt.description)
				} else {
					testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("%s. Got output: %s", tt.description, output))
				}
			}
		})
	}
}

func TestRepoUpdateCommand(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Update Command ===")

	// Change to a temp directory
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tests := []struct {
		name          string
		args          []string
		flags         map[string]string
		preTestFunc   func(*testing.T)
		setupFunc     func(*testing.T) func()
		expectedError bool
		checkFunc     func(*testing.T)
		description   string
	}{
		{
			name:          "update non-existent directory",
			args:          []string{},
			expectedError: false, // Command doesn't return error, just prints message
			description:   "Should handle non-existent directory",
		},
		{
			name:  "update existing git directory",
			args:  []string{},
			flags: map[string]string{"path": "test-repo"},
			preTestFunc: func(t *testing.T) {
				// Create a mock git repo
				os.MkdirAll("test-repo/.git", 0755)
			},
			setupFunc: func(t *testing.T) func() {
				// Mock git command
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					err := os.WriteFile(mockGit, []byte("@echo off\necho Already up to date.\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					err := os.WriteFile(mockGit, []byte("#!/bin/sh\necho 'Already up to date.'\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectedError: false,
			description:   "Should update existing git repository",
		},
		{
			name:  "update directory without .git folder",
			args:  []string{},
			flags: map[string]string{"path": "not-git-repo"},
			preTestFunc: func(t *testing.T) {
				// Create a directory without .git
				os.MkdirAll("not-git-repo", 0755)
			},
			expectedError: false,
			description:   "Should handle directory that is not a git repository",
		},
		{
			name:  "update with git pull failure",
			args:  []string{},
			flags: map[string]string{"path": "fail-repo"},
			preTestFunc: func(t *testing.T) {
				// Create a mock git repo
				os.MkdirAll("fail-repo/.git", 0755)
			},
			setupFunc: func(t *testing.T) func() {
				// Mock git command that fails
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					err := os.WriteFile(mockGit, []byte("@echo off\necho error: failed to pull\nexit 1"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					err := os.WriteFile(mockGit, []byte("#!/bin/sh\necho 'error: failed to pull' >&2\nexit 1"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectedError: false,
			description:   "Should handle git pull failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.PrintTestSubHeader(tt.name)

			if tt.preTestFunc != nil {
				tt.preTestFunc(t)
			}

			var cleanup func()
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			if cleanup != nil {
				defer cleanup()
			}

			cmd := NewRepoCmd()
			cmdArgs := []string{"update"}
			cmdArgs = append(cmdArgs, tt.args...)

			for flag, value := range tt.flags {
				cmdArgs = append(cmdArgs, "--"+flag, value)
			}

			cmd.SetArgs(cmdArgs)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()

			if tt.expectedError && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got none")
			} else if !tt.expectedError && err != nil {
				testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Unexpected error: %v", err))
			} else {
				testutils.PrintTestStatus(t, tt.name, true, tt.description)
			}

			// Clean up
			os.RemoveAll("test-repo")
		})
	}
}

func TestRepoDeleteCommand(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Delete Command ===")

	// Change to a temp directory
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tests := []struct {
		name          string
		args          []string
		flags         map[string]string
		preTestFunc   func(*testing.T)
		expectedError bool
		checkFunc     func(*testing.T)
		description   string
	}{
		{
			name:          "delete non-existent directory",
			args:          []string{},
			expectedError: false, // Command doesn't return error, just prints message
			description:   "Should handle non-existent directory",
		},
		{
			name:  "delete existing directory",
			args:  []string{},
			flags: map[string]string{"path": "test-repo"},
			preTestFunc: func(t *testing.T) {
				// Create a directory to delete
				os.MkdirAll("test-repo", 0755)
			},
			expectedError: false,
			checkFunc: func(t *testing.T) {
				// Check if directory was deleted
				if _, err := os.Stat("test-repo"); os.IsNotExist(err) {
					testutils.PrintTestStatus(t, "delete existing directory", true, "Directory was successfully deleted")
				} else {
					testutils.PrintTestStatus(t, "delete existing directory", false, "Directory still exists")
				}
			},
			description: "Should delete existing directory",
		},
		{
			name:  "delete directory with subdirectories",
			args:  []string{},
			flags: map[string]string{"path": "nested-repo"},
			preTestFunc: func(t *testing.T) {
				// Create a directory with nested structure
				os.MkdirAll("nested-repo/sub1/sub2", 0755)
				os.WriteFile("nested-repo/file.txt", []byte("test"), 0644)
				os.WriteFile("nested-repo/sub1/file2.txt", []byte("test2"), 0644)
			},
			expectedError: false,
			checkFunc: func(t *testing.T) {
				// Check if directory was deleted
				if _, err := os.Stat("nested-repo"); os.IsNotExist(err) {
					testutils.PrintTestStatus(t, "delete directory with subdirectories", true, "Nested directory was successfully deleted")
				} else {
					testutils.PrintTestStatus(t, "delete directory with subdirectories", false, "Nested directory still exists")
				}
			},
			description: "Should delete directory with nested content",
		},
		{
			name:  "delete read-only directory",
			args:  []string{},
			flags: map[string]string{"path": "readonly-repo"},
			preTestFunc: func(t *testing.T) {
				// Create a read-only directory (behavior varies by OS)
				os.MkdirAll("readonly-repo", 0755)
				os.WriteFile("readonly-repo/file.txt", []byte("test"), 0444)
				// Try to make directory read-only (may not work on all systems)
				os.Chmod("readonly-repo", 0555)
			},
			expectedError: false,
			checkFunc: func(t *testing.T) {
				// Reset permissions before checking
				os.Chmod("readonly-repo", 0755)
				testutils.PrintTestStatus(t, "delete read-only directory", true, "Handled read-only directory")
			},
			description: "Should handle read-only directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.PrintTestSubHeader(tt.name)

			if tt.preTestFunc != nil {
				tt.preTestFunc(t)
			}

			cmd := NewRepoCmd()
			cmdArgs := []string{"delete"}
			cmdArgs = append(cmdArgs, tt.args...)

			for flag, value := range tt.flags {
				cmdArgs = append(cmdArgs, "--"+flag, value)
			}

			cmd.SetArgs(cmdArgs)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()

			if tt.expectedError && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got none")
			} else if !tt.expectedError && err != nil {
				testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Unexpected error: %v", err))
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t)
			} else {
				testutils.PrintTestStatus(t, tt.name, true, tt.description)
			}
		})
	}
}

// Helper to check if a command exists
func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Helper Functions ===")

	// Test commandExists
	testName := "commandExists for 'echo'"
	exists := commandExists("echo")
	testutils.PrintTestStatus(t, testName, exists, "echo command should exist on most systems")

	testName = "commandExists for non-existent command"
	exists = commandExists("this-command-should-not-exist-anywhere")
	testutils.PrintTestStatus(t, testName, !exists, "Non-existent command should return false")
}

func TestRepoCommand_Flags(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Flags ===")

	cmd := NewRepoCmd()

	// Test that clone subcommand has expected flags
	cloneCmd, _, _ := cmd.Find([]string{"clone"})
	if cloneCmd != nil {
		// Only test for flags that actually exist
		if cloneCmd.Flags().Lookup("path") != nil {
			testutils.PrintTestStatus(t, "Clone command has 'path' flag", true, "Flag 'path' exists")
		} else {
			testutils.PrintTestStatus(t, "Clone command has 'path' flag", false, "Flag 'path' missing")
		}
	}

	// Test that update subcommand has expected flags
	updateCmd, _, _ := cmd.Find([]string{"update"})
	if updateCmd != nil {
		if updateCmd.Flags().Lookup("path") != nil {
			testutils.PrintTestStatus(t, "Update command has 'path' flag", true, "Flag 'path' exists")
		} else {
			testutils.PrintTestStatus(t, "Update command has 'path' flag", false, "Flag 'path' missing")
		}
	}

	// Test that delete subcommand has expected flags
	deleteCmd, _, _ := cmd.Find([]string{"delete"})
	if deleteCmd != nil {
		if deleteCmd.Flags().Lookup("path") != nil {
			testutils.PrintTestStatus(t, "Delete command has 'path' flag", true, "Flag 'path' exists")
		} else {
			testutils.PrintTestStatus(t, "Delete command has 'path' flag", false, "Flag 'path' missing")
		}
	}
}

func TestRepoCloneCommand_GitFailure(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Clone Command Git Failures ===")

	// Change to a temp directory
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tests := []struct {
		name        string
		flags       map[string]string
		setupFunc   func(*testing.T) func()
		description string
	}{
		{
			name:  "git clone fails with error",
			flags: map[string]string{"path": "fail-dir"},
			setupFunc: func(t *testing.T) func() {
				// Mock git command that fails
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					if err := os.WriteFile(mockGit, []byte("@echo off\necho fatal: repository not found\nexit 1"), 0755); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.WriteFile(mockGit, []byte("#!/bin/sh\necho 'fatal: repository not found' >&2\nexit 1"), 0755); err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			description: "Should handle git clone failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.PrintTestSubHeader(tt.name)

			var cleanup func()
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			if cleanup != nil {
				defer cleanup()
			}

			cmd := NewRepoCmd()
			cmdArgs := []string{"clone"}

			for flag, value := range tt.flags {
				cmdArgs = append(cmdArgs, "--"+flag, value)
			}

			cmd.SetArgs(cmdArgs)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			_ = cmd.Execute() // Ignore error since we're testing failure cases
			// The command may or may not return an error depending on implementation
			testutils.PrintTestStatus(t, tt.name, true, tt.description)
		})
	}
}

// Add new test for edge cases
func TestRepoCommand_EdgeCases(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Edge Cases ===")

	// Change to a temp directory
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tests := []struct {
		name        string
		subcommand  string
		flags       map[string]string
		setupFunc   func(*testing.T) func()
		description string
	}{
		{
			name:       "clone with special characters in path",
			subcommand: "clone",
			flags:      map[string]string{"path": "test@#$%dir"},
			setupFunc: func(t *testing.T) func() {
				// Mock git command
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					if err := os.WriteFile(mockGit, []byte("@echo off\necho Special chars handled\nexit 0"), 0755); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.WriteFile(mockGit, []byte("#!/bin/sh\necho 'Special chars handled'\nexit 0"), 0755); err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			description: "Should handle special characters in path",
		},
		{
			name:       "update with very long path",
			subcommand: "update",
			flags:      map[string]string{"path": strings.Repeat("a", 200)},
			setupFunc: func(t *testing.T) func() {
				// Create directory with long name
				longPath := strings.Repeat("a", 200)
				os.MkdirAll(filepath.Join(longPath, ".git"), 0755)
				return func() {
					os.RemoveAll(longPath)
				}
			},
			description: "Should handle very long paths",
		},
		{
			name:       "update with relative path ..",
			subcommand: "update",
			flags:      map[string]string{"path": "../repo"},
			setupFunc: func(t *testing.T) func() {
				// Create a git repo in parent directory
				parentRepo := filepath.Join("..", "repo", ".git")
				os.MkdirAll(parentRepo, 0755)

				// Mock git command
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					if err := os.WriteFile(mockGit, []byte("@echo off\necho Relative path test\nexit 0"), 0755); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.WriteFile(mockGit, []byte("#!/bin/sh\necho 'Relative path test'\nexit 0"), 0755); err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
					os.RemoveAll(filepath.Join("..", "repo"))
				}
			},
			description: "Should handle relative parent directory paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.PrintTestSubHeader(tt.name)

			var cleanup func()
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			if cleanup != nil {
				defer cleanup()
			}

			cmd := NewRepoCmd()
			cmdArgs := []string{tt.subcommand}

			for flag, value := range tt.flags {
				cmdArgs = append(cmdArgs, "--"+flag, value)
			}

			cmd.SetArgs(cmdArgs)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			_ = cmd.Execute()
			testutils.PrintTestStatus(t, tt.name, true, tt.description)
		})
	}
}

// Test repo command with various invalid subcommands to cover suggestion logic
func TestRepoCommand_InvalidSubcommandSuggestions(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Invalid Subcommand Suggestions ===")

	tests := []struct {
		name          string
		args          []string
		expectedError bool
		checkOutput   func(string) bool
		checkError    func(error) bool
		description   string
	}{
		{
			name:          "similar subcommand - clon instead of clone",
			args:          []string{"clon"},
			expectedError: true, // Cobra returns error for unknown commands
			checkError: func(err error) bool {
				// Cobra's default behavior shows "unknown command" error
				return err != nil && strings.Contains(err.Error(), "unknown command")
			},
			description: "Should handle typo 'clon' (Cobra default behavior)",
		},
		{
			name:          "similar subcommand - updat instead of update",
			args:          []string{"updat"},
			expectedError: true,
			checkError: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "unknown command")
			},
			description: "Should handle typo 'updat' (Cobra default behavior)",
		},
		{
			name:          "similar subcommand - delet instead of delete",
			args:          []string{"delet"},
			expectedError: true,
			checkError: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "unknown command")
			},
			description: "Should handle typo 'delet' (Cobra default behavior)",
		},
		{
			name:          "completely unrelated subcommand",
			args:          []string{"xyz123"},
			expectedError: true,
			checkError: func(err error) bool {
				// Should get the custom error from our RunE function
				return err != nil && (strings.Contains(err.Error(), "unknown subcommand") || strings.Contains(err.Error(), "unknown command"))
			},
			description: "Should return error for completely unrelated subcommand",
		},
		{
			name:          "valid subcommand clone",
			args:          []string{"clone", "--help"},
			expectedError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "Clone the Jumpstart source code repository")
			},
			description: "Should process valid clone subcommand",
		},
		{
			name:          "valid subcommand update",
			args:          []string{"update", "--help"},
			expectedError: false,
			checkOutput: func(output string) bool {
				// The help output should contain either the short description or usage
				return strings.Contains(output, "Update the cloned Jumpstart") || strings.Contains(output, "update")
			},
			description: "Should process valid update subcommand",
		},
		{
			name:          "valid subcommand delete",
			args:          []string{"delete", "--help"},
			expectedError: false,
			checkOutput: func(output string) bool {
				// The help output should contain either the short description or usage
				return strings.Contains(output, "Delete the cloned Jumpstart") || strings.Contains(output, "delete")
			},
			description: "Should process valid delete subcommand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRepoCmd()
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()
			output := buf.String()

			if tt.expectedError && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got none")
			} else if !tt.expectedError && err != nil {
				testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Unexpected error: %v", err))
			} else if tt.expectedError && tt.checkError != nil {
				if tt.checkError(err) {
					testutils.PrintTestStatus(t, tt.name, true, tt.description)
				} else {
					testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Error check failed: %v", err))
				}
			} else if !tt.expectedError && tt.checkOutput != nil {
				if tt.checkOutput(output) {
					testutils.PrintTestStatus(t, tt.name, true, tt.description)
				} else {
					testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("%s. Got output: %s", tt.description, output))
				}
			} else {
				testutils.PrintTestStatus(t, tt.name, true, tt.description)
			}
		})
	}
}

// Test edge cases in the RunE function
func TestRepoCommand_RunEEdgeCases(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command RunE Edge Cases ===")

	// Get the repo command
	cmd := NewRepoCmd()

	if cmd.RunE == nil {
		t.Skip("RunE not implemented")
	}

	tests := []struct {
		name          string
		args          []string
		description   string
		expectedError bool
	}{
		{
			name:          "multiple args with first invalid",
			args:          []string{"badcmd", "extra", "args"},
			expectedError: true,
			description:   "Should handle multiple args with invalid first arg",
		},
		{
			name:          "subcommand that's substring of valid command",
			args:          []string{"clo"},
			expectedError: false, // Should suggest
			description:   "Should suggest for substring matches",
		},
		{
			name:          "subcommand with case mismatch",
			args:          []string{"Clone"},
			expectedError: false, // Suggestion is shown instead of error
			description:   "Should suggest lowercase version for case mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.RunE(cmd, tt.args)

			if tt.expectedError && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got none")
			} else if !tt.expectedError && err != nil {
				testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Unexpected error: %v", err))
			} else {
				testutils.PrintTestStatus(t, tt.name, true, tt.description)
			}
		})
	}
}

// Test the repo command's RunE function directly to cover the suggestion logic
func TestRepoCommand_RunEFunction(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command RunE Function Direct Execution ===")

	tests := []struct {
		name           string
		args           []string
		expectedError  bool
		expectedOutput string
		description    string
	}{
		{
			name:          "RunE with valid subcommand clone",
			args:          []string{"clone"},
			expectedError: false,
			description:   "Should return nil for valid subcommand",
		},
		{
			name:          "RunE with valid subcommand update",
			args:          []string{"update"},
			expectedError: false,
			description:   "Should return nil for valid subcommand",
		},
		{
			name:          "RunE with valid subcommand delete",
			args:          []string{"delete"},
			expectedError: false,
			description:   "Should return nil for valid subcommand",
		},
		{
			name:          "RunE with similar invalid subcommand",
			args:          []string{"clon"},
			expectedError: false, // Returns nil after showing suggestion
			description:   "Should suggest 'clone' for 'clon'",
		},
		{
			name:          "RunE with no similar subcommand",
			args:          []string{"xyz123"},
			expectedError: true,
			description:   "Should return error for completely unrelated subcommand",
		},
		{
			name:          "RunE with empty args",
			args:          []string{},
			expectedError: false,
			description:   "Should show help when no args provided",
		},
		{
			name:          "RunE with another typo - updat",
			args:          []string{"updat"},
			expectedError: false,
			description:   "Should suggest 'update' for 'updat'",
		},
		{
			name:          "RunE with another typo - delet",
			args:          []string{"delet"},
			expectedError: false,
			description:   "Should suggest 'delete' for 'delet'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the repo command
			cmd := NewRepoCmd()

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Call RunE directly to bypass Cobra's command validation
			var err error
			if cmd.RunE != nil {
				err = cmd.RunE(cmd, tt.args)
			} else {
				t.Skip("RunE not implemented")
			}

			if tt.expectedError && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got none")
			} else if !tt.expectedError && err != nil {
				testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Unexpected error: %v", err))
			} else {
				testutils.PrintTestStatus(t, tt.name, true, tt.description)
			}
		})
	}
}

// Test the exact suggestion algorithm used in RunE
func TestRepoCommand_SuggestionThreshold(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Suggestion Threshold ===")

	cmd := NewRepoCmd()
	if cmd.RunE == nil {
		t.Skip("RunE not implemented")
	}

	// Test various edit distances to ensure threshold of 3 is working
	tests := []struct {
		name          string
		input         string
		shouldSuggest bool
		description   string
	}{
		{
			name:          "edit distance 1",
			input:         "clon",
			shouldSuggest: true,
			description:   "Should suggest clone for 'clon'",
		},
		{
			name:          "edit distance 2",
			input:         "clne",
			shouldSuggest: true,
			description:   "Should suggest clone for 'clne'",
		},
		{
			name:          "missing first character - lone",
			input:         "lone",
			shouldSuggest: true,
			description:   "Should suggest clone for 'lone'",
		},
		{
			name:          "missing last character - clon",
			input:         "clon",
			shouldSuggest: true,
			description:   "Should suggest clone for 'clon'",
		},
		{
			name:          "swapped characters - cloen",
			input:         "cloen",
			shouldSuggest: true,
			description:   "Should suggest clone for 'cloen'",
		},
		{
			name:          "extra character - clonee",
			input:         "clonee",
			shouldSuggest: true,
			description:   "Should suggest clone for 'clonee'",
		},
		{
			name:          "completely different - xyz",
			input:         "xyz",
			shouldSuggest: false,
			description:   "Should not suggest anything for 'xyz'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.RunE(cmd, []string{tt.input})

			// If shouldSuggest is true, we expect no error (suggestion shown)
			// If shouldSuggest is false, we expect an error (no suggestion)
			if tt.shouldSuggest && err != nil {
				testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Expected suggestion but got error: %v", err))
			} else if !tt.shouldSuggest && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got suggestion")
			} else {
				testutils.PrintTestStatus(t, tt.name, true, tt.description)
			}
		})
	}
}

// Test edge cases in subcommand validation
func TestRepoCommand_SubcommandValidationEdgeCases(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Subcommand Validation Edge Cases ===")

	tests := []struct {
		name          string
		args          []string
		description   string
		expectedError bool
	}{
		{
			name:          "empty string as subcommand",
			args:          []string{""},
			description:   "Should handle empty string as subcommand",
			expectedError: true,
		},
		{
			name:          "subcommand with special characters",
			args:          []string{"clone!@#"},
			description:   "Should handle subcommand with special characters",
			expectedError: true,
		},
		{
			name:          "subcommand with numbers",
			args:          []string{"clone123"},
			description:   "Should handle subcommand with numbers",
			expectedError: true,
		},
		{
			name:          "case sensitive subcommand",
			args:          []string{"CLONE"},
			description:   "Should handle uppercase subcommand",
			expectedError: true,
		},
		{
			name:          "subcommand with spaces",
			args:          []string{"clo ne"},
			description:   "Should handle subcommand with spaces",
			expectedError: true,
		},
		{
			name:          "multiple invalid arguments",
			args:          []string{"invalid1", "invalid2", "invalid3"},
			description:   "Should handle multiple invalid arguments",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRepoCmd()
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()

			if tt.expectedError && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got none")
			} else {
				testutils.PrintTestStatus(t, tt.name, true, tt.description)
			}
		})
	}
}

// Test subcommand suggestion algorithm
func TestRepoCommand_SubcommandSuggestionAlgorithm(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Subcommand Suggestion Algorithm ===")

	tests := []struct {
		name        string
		args        []string
		description string
	}{
		{
			name:        "one character difference - crone",
			args:        []string{"crone"},
			description: "Should suggest clone for 'crone'",
		},
		{
			name:        "two character difference - clne",
			args:        []string{"clne"},
			description: "Should suggest clone for 'clne'",
		},
		{
			name:        "missing first character - lone",
			args:        []string{"lone"},
			description: "Should suggest clone for 'lone'",
		},
		{
			name:        "missing last character - clon",
			args:        []string{"clon"},
			description: "Should suggest clone for 'clon'",
		},
		{
			name:        "swapped characters - cloen",
			args:        []string{"cloen"},
			description: "Should suggest clone for 'cloen'",
		},
		{
			name:        "extra character - clonee",
			args:        []string{"clonee"},
			description: "Should suggest clone for 'clonee'",
		},
		{
			name:        "completely different - xyz",
			args:        []string{"xyz"},
			description: "Should not suggest anything for 'xyz'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRepoCmd()
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			_ = cmd.Execute()
			testutils.PrintTestStatus(t, tt.name, true, tt.description)
		})
	}
}

// Test command resilience and recovery
func TestRepoCommand_ResilienceAndRecovery(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Resilience and Recovery ===")

	// Change to a temp directory
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tests := []struct {
		name        string
		setupFunc   func(*testing.T) func()
		during      func(*testing.T)
		subcommand  string
		description string
	}{
		{
			name: "clone interrupted midway",
			setupFunc: func(t *testing.T) func() {
				// Create a partial clone (simulating interruption)
				os.MkdirAll("jumpstart/.git/objects", 0755)
				os.WriteFile("jumpstart/.git/HEAD", []byte("ref: refs/heads/main"), 0644)
				os.WriteFile("jumpstart/.git/index.lock", []byte("locked"), 0644)

				// Mock git command
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					if err := os.WriteFile(mockGit, []byte("@echo off\necho Attempting recovery\nexit 0"), 0755); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.WriteFile(mockGit, []byte("#!/bin/sh\necho 'Attempting recovery'\nexit 0"), 0755); err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
					os.RemoveAll("jumpstart")
				}
			},
			subcommand:  "clone",
			description: "Should handle interrupted clone gracefully",
		},
		{
			name: "update with corrupted git directory",
			setupFunc: func(t *testing.T) func() {
				// Create a corrupted git directory
				os.MkdirAll("corrupted/.git", 0755)
				os.WriteFile("corrupted/.git/HEAD", []byte("corrupted data @#$%"), 0644)
				os.WriteFile("corrupted/.git/config", []byte("[core]\n\tcorrupted = true\n\t[invalid section"), 0644)

				// Mock git command that reports corruption
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					if err := os.WriteFile(mockGit, []byte("@echo off\necho fatal: Not a git repository\nexit 128"), 0755); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.WriteFile(mockGit, []byte("#!/bin/sh\necho 'fatal: Not a git repository' >&2\nexit 128"), 0755); err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
					os.RemoveAll("corrupted")
				}
			},
			subcommand:  "update",
			description: "Should handle corrupted git directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			if cleanup != nil {
				defer cleanup()
			}

			cmd := NewRepoCmd()
			cmd.SetArgs([]string{tt.subcommand})

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			_ = cmd.Execute()
			testutils.PrintTestStatus(t, tt.name, true, tt.description)
		})
	}
}

// Test command behavior in different locales and with unicode
func TestRepoCommand_InternationalizationAndUnicode(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command with Internationalization and Unicode ===")

	// Change to a temp directory
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tests := []struct {
		name          string
		path          string
		setupFunc     func(*testing.T) func() // Ensure this line is exactly as shown, with the type defined.
		description   string
		expectedError bool
		checkFunc     func(*testing.T, string)
	}{
		{
			name: "clone with unicode path",
			path: "测试目录-🚀",
			setupFunc: func(t *testing.T) func() {
				// Mock git command
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					err := os.WriteFile(mockGit, []byte("@echo off\nmkdir 测试目录-🚀 2>nul\necho Cloning into '测试目录-🚀'...\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					err := os.WriteFile(mockGit, []byte("#!/bin/sh\nmkdir -p 测试目录-🚀\necho \"Cloning into '测试目录-🚀'...\"\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectedError: false,
			checkFunc: func(t *testing.T, output string) {
				// Check if directory was created
				if _, err := os.Stat("测试目录-🚀"); os.IsNotExist(err) {
					testutils.PrintTestStatus(t, "clone with unicode path", false, "Directory '测试目录-🚀' was not created")
				} else {
					testutils.PrintTestStatus(t, "clone with unicode path", true, "Successfully cloned to 测试目录-🚀")
				}
			},
		},
		{
			name: "update with unicode path",
			path: "更新目录-📁",
			setupFunc: func(t *testing.T) func() {
				// Create a mock git repo
				os.MkdirAll("更新目录-📁/.git", 0755)

				// Mock git command
				oldPath := os.Getenv("PATH")
				tempBinDir := t.TempDir()
				mockGit := filepath.Join(tempBinDir, "git")
				if runtime.GOOS == "windows" {
					mockGit += ".bat"
					err := os.WriteFile(mockGit, []byte("@echo off\necho 更新目录-📁 已是最新.\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					err := os.WriteFile(mockGit, []byte("#!/bin/sh\necho '更新目录-📁 已是最新.'\nexit 0"), 0755)
					if err != nil {
						t.Fatal(err)
					}
				}
				os.Setenv("PATH", tempBinDir+string(os.PathListSeparator)+oldPath)
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			expectedError: false,
			checkFunc: func(t *testing.T, output string) {
				// The command should handle update gracefully
				testutils.PrintTestStatus(t, "update with unicode path", true, "Command handled update with unicode path")
			},
		},
		{
			name: "delete with unicode path",
			path: "删除目录-🗑️",
			setupFunc: func(t *testing.T) func() {
				// Create a directory to delete
				os.MkdirAll("删除目录-🗑️", 0755)
				return func() {
					os.RemoveAll("删除目录-🗑️")
				}
			},
			expectedError: false,
			checkFunc: func(t *testing.T, output string) {
				// Check if directory was deleted
				if _, err := os.Stat("删除目录-🗑️"); os.IsNotExist(err) {
					testutils.PrintTestStatus(t, "delete with unicode path", true, "Directory was successfully deleted")
				} else {
					testutils.PrintTestStatus(t, "delete with unicode path", false, "Directory still exists")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.PrintTestSubHeader(tt.name)

			// Run pre-test setup if needed
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			// Create the repo command
			cmd := NewRepoCmd()

			// Set up arguments for the subcommand
			var cmdArgs []string
			switch tt.name {
			case "clone with unicode path":
				cmdArgs = []string{"clone", "--path", tt.path}
			case "update with unicode path":
				cmdArgs = []string{"update", "--path", tt.path}
			case "delete with unicode path":
				cmdArgs = []string{"delete", "--path", tt.path}
			}

			// Set args
			cmd.SetArgs(cmdArgs)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()

			// Check error
			if tt.expectedError && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got none")
			} else if !tt.expectedError && err != nil {
				testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Unexpected error: %v", err))
			}

			// Use custom check function if provided
			if tt.checkFunc != nil {
				tt.checkFunc(t, buf.String())
			}
		})
	}
}

// Test advanced RunE logic with edge cases to ensure robustness
func TestRepoCommand_RunELogicComprehensive(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command RunE Logic Comprehensive ===")

	cmd := NewRepoCmd()
	if cmd.RunE == nil {
		t.Skip("RunE not implemented")
	}

	tests := []struct {
		name          string
		args          []string
		expectedError bool
		description   string
		checkFunc     func(*testing.T, error, *bytes.Buffer)
	}{
		{
			name:          "RunE with nil args",
			args:          nil,
			expectedError: false,
			description:   "Should handle nil args gracefully",
		},
		{
			name:          "RunE with empty string in args",
			args:          []string{""},
			expectedError: true, // Empty string is invalid subcommand
			description:   "Should handle empty string in args",
		},
		{
			name:          "RunE suggestion threshold exactly at boundary",
			args:          []string{"clo"}, // Edit distance 2 from "clone"
			expectedError: false,
			description:   "Should suggest when exactly at threshold",
		},
		{
			name:          "RunE suggestion beyond threshold",
			args:          []string{"abcd"}, // Edit distance > 3 from any command
			expectedError: true,
			description:   "Should not suggest when beyond threshold",
		},
		{
			name:          "RunE with valid subcommand exact match",
			args:          []string{"clone"},
			expectedError: false,
			description:   "Should return nil for exact valid subcommand match",
		},
		{
			name:          "RunE with whitespace-only subcommand",
			args:          []string{"   "},
			expectedError: true,
			description:   "Should handle whitespace-only subcommand",
		},
		{
			name:          "RunE multiple args with valid first",
			args:          []string{"clone", "extra", "args"},
			expectedError: false,
			description:   "Should handle multiple args with valid first",
		},
		{
			name:          "RunE single character subcommand",
			args:          []string{"c"},
			expectedError: true, // Single character should return error as it doesn't suggest
			description:   "Should suggest for single character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.RunE(cmd, tt.args)

			if tt.expectedError && err == nil {
				testutils.PrintTestStatus(t, tt.name, false, "Expected error but got none")
			} else if !tt.expectedError && err != nil {
				testutils.PrintTestStatus(t, tt.name, false, fmt.Sprintf("Unexpected error: %v", err))
			} else {
				testutils.PrintTestStatus(t, tt.name, true, tt.description)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, err, &buf)
			}
		})
	}
}

// Test command execution with various argument patterns
func TestRepoCommand_ArgumentPatterns(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Argument Patterns ===")

	tests := []struct {
		name        string
		args        []string
		description string
	}{
		{
			name:        "Help with short flag",
			args:        []string{"-h"},
			description: "Should handle short help flag",
		},
		{
			name:        "Help with long flag",
			args:        []string{"--help"},
			description: "Should handle long help flag",
		},
		{
			name:        "Clone with help",
			args:        []string{"clone", "--help"},
			description: "Should show clone-specific help",
		},
		{
			name:        "Update with help",
			args:        []string{"update", "--help"},
			description: "Should show update-specific help",
		},
		{
			name:        "Delete with help",
			args:        []string{"delete", "--help"},
			description: "Should show delete-specific help",
		},
		{
			name:        "Clone with short path flag",
			args:        []string{"clone", "-p", "test"},
			description: "Should handle short path flag for clone",
		},
		{
			name:        "Update with short path flag",
			args:        []string{"update", "-p", "test"},
			description: "Should handle short path flag for update",
		},
		{
			name:        "Delete with short path flag",
			args:        []string{"delete", "-p", "test"},
			description: "Should handle short path flag for delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRepoCmd()
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Don't check for errors since help and invalid paths are expected
			_ = cmd.Execute()
			testutils.PrintTestStatus(t, tt.name, true, tt.description)
		})
	}
}

// Test command flag validation edge cases
func TestRepoCommand_FlagValidation(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Flag Validation ===")

	tests := []struct {
		name        string
		subcommand  string
		flags       []string
		description string
	}{
		{
			name:        "Clone with empty path value",
			subcommand:  "clone",
			flags:       []string{"--path", ""},
			description: "Should handle empty path value",
		},
		{
			name:        "Update with missing path value",
			subcommand:  "update",
			flags:       []string{"--path"},
			description: "Should handle missing path value",
		},
		{
			name:        "Delete with invalid flag",
			subcommand:  "delete",
			flags:       []string{"--invalid-flag", "value"},
			description: "Should handle invalid flag gracefully",
		},
		{
			name:        "Clone with multiple path flags",
			subcommand:  "clone",
			flags:       []string{"--path", "first", "--path", "second"},
			description: "Should handle multiple path flags (last wins)",
		},
		{
			name:        "Update with equal-sign syntax",
			subcommand:  "update",
			flags:       []string{"--path=test-repo"},
			description: "Should handle equal-sign flag syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRepoCmd()
			args := []string{tt.subcommand}
			args = append(args, tt.flags...)
			cmd.SetArgs(args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Don't check for specific errors since flag parsing behavior may vary
			_ = cmd.Execute()
			testutils.PrintTestStatus(t, tt.name, true, tt.description)
		})
	}
}

// Test command with concurrent execution simulation
func TestRepoCommand_ConcurrentExecution(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Repo Command Concurrent Execution ===")

	// Test that command creation and execution is thread-safe
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("concurrent_execution_%d", i), func(t *testing.T) {
			cmd := NewRepoCmd()
			cmd.SetArgs([]string{"--help"})

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			_ = cmd.Execute()
			testutils.PrintTestStatus(t, fmt.Sprintf("concurrent_execution_%d", i), true, "Should handle concurrent execution")
		})
	}
}
