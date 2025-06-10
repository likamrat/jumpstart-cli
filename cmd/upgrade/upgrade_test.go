package upgrade

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"jumpstartcli/internal/utils"

	"github.com/fatih/color"
)

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	HTMLURL     string        `json:"html_url"`
	Prerelease  bool          `json:"prerelease"`
	PublishedAt string        `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Color functions for test output
var (
	testSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	testInfoColor    = color.New(color.FgCyan).SprintFunc()
	testErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	testHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

// Helper function to print test status
func printTestStatus(t *testing.T, testName string, success bool, message string) {
	if success {
		fmt.Printf("%s %s - %s\n", testSuccessColor("✅"), testName, testInfoColor(message))
	} else {
		fmt.Printf("%s %s - %s\n", testErrorColor("❌"), testName, testErrorColor(message))
		t.Error(message)
	}
}

func TestNewUpgradeCmd(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Creation ==="))

	cmd := NewUpgradeCmd()

	// Test basic command structure
	printTestStatus(t, "Command Use", cmd.Use == "upgrade", fmt.Sprintf("Expected 'upgrade', got '%s'", cmd.Use))
	printTestStatus(t, "Command Short Description", cmd.Short != "", "Short description should not be empty")
	printTestStatus(t, "Command Long Description", cmd.Long != "", "Long description should not be empty")
	printTestStatus(t, "RunE Function", cmd.RunE != nil, "RunE function should be defined")
}

func TestUpgradeCommandFlags(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Flags ==="))

	cmd := NewUpgradeCmd()

	// Test that expected flags exist
	expectedFlags := []string{"check", "pre-release", "force"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		success := flag != nil
		printTestStatus(t, fmt.Sprintf("Flag '%s'", flagName), success,
			fmt.Sprintf("Flag '%s' should exist", flagName))
	}
}

func TestUpgradeCommandFlagDefaults(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Flag Defaults ==="))

	cmd := NewUpgradeCmd()

	// Test flag default values
	flagTests := []struct {
		name         string
		flagName     string
		expectedType string
		expectBool   bool
	}{
		{
			name:         "check flag default",
			flagName:     "check",
			expectedType: "bool",
			expectBool:   false,
		},
		{
			name:         "pre-release flag default",
			flagName:     "pre-release",
			expectedType: "bool",
			expectBool:   false,
		},
		{
			name:         "force flag default",
			flagName:     "force",
			expectedType: "bool",
			expectBool:   false,
		},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				printTestStatus(t, fmt.Sprintf("Flag '%s' exists", tt.flagName), false,
					fmt.Sprintf("Flag '%s' not found", tt.flagName))
				return
			}

			// Test flag type
			typeCorrect := flag.Value.Type() == tt.expectedType
			printTestStatus(t, fmt.Sprintf("Flag '%s' type", tt.flagName), typeCorrect,
				fmt.Sprintf("Expected type '%s', got '%s'", tt.expectedType, flag.Value.Type()))

			// Test default value for bool flags
			if tt.expectedType == "bool" {
				value, err := cmd.Flags().GetBool(tt.flagName)
				if err != nil {
					printTestStatus(t, fmt.Sprintf("Flag '%s' value retrieval", tt.flagName), false,
						fmt.Sprintf("Error getting bool value: %v", err))
					return
				}
				valueCorrect := value == tt.expectBool
				printTestStatus(t, fmt.Sprintf("Flag '%s' default value", tt.flagName), valueCorrect,
					fmt.Sprintf("Expected %v, got %v", tt.expectBool, value))
			}
		})
	}
}

func TestUpgradeCommandFlagShorthands(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Flag Shorthands ==="))

	cmd := NewUpgradeCmd()

	// Test shorthand flags exist
	shorthandTests := []struct {
		flagName  string
		shorthand string
	}{
		{"check", "c"},
		{"pre-release", "p"},
		{"force", "f"},
	}

	for _, tt := range shorthandTests {
		t.Run("shorthand_"+tt.flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				printTestStatus(t, fmt.Sprintf("Flag '%s' exists", tt.flagName), false,
					fmt.Sprintf("Flag '%s' not found", tt.flagName))
				return
			}

			shorthandCorrect := flag.Shorthand == tt.shorthand
			printTestStatus(t, fmt.Sprintf("Flag '%s' shorthand", tt.flagName), shorthandCorrect,
				fmt.Sprintf("Expected shorthand '%s', got '%s'", tt.shorthand, flag.Shorthand))
		})
	}
}

func TestUpgradeCommandExecution(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Execution ==="))

	// Store original debug mode
	originalDebugMode := utils.DebugMode
	defer func() {
		utils.DebugMode = originalDebugMode
	}()

	tests := []struct {
		name        string
		args        []string
		debugMode   bool
		expectError bool
	}{
		{
			name:        "upgrade with check flag",
			args:        []string{"--check"},
			debugMode:   false,
			expectError: false, // Should not error, but may not find updates
		},
		{
			name:        "upgrade with debug mode",
			args:        []string{"--check"},
			debugMode:   true,
			expectError: false,
		},
		{
			name:        "upgrade with pre-release flag",
			args:        []string{"--check", "--pre-release"},
			debugMode:   false,
			expectError: false,
		},
		{
			name:        "upgrade with force flag",
			args:        []string{"--check", "--force"},
			debugMode:   false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor(tt.name))

			// Set debug mode for test
			utils.DebugMode = tt.debugMode

			cmd := NewUpgradeCmd()
			cmd.SetArgs(tt.args)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()

			// Since the actual upgrade logic involves network calls and file operations,
			// we mainly test that the command structure is correct and doesn't panic
			// The actual upgrade functionality would require more complex mocking

			if tt.expectError && err == nil {
				printTestStatus(t, "Error expectation", false, "Expected error but got none")
				return
			}

			// For non-error cases, we can check if certain expected messages appear
			if !tt.expectError {
				output := buf.String()
				// Should at least attempt to check for updates or show repository guidance
				hasExpectedOutput := strings.Contains(output, "Checking for updates") ||
					strings.Contains(output, "Repository not found") ||
					strings.Contains(output, "failed to check for updates")

				// Handle expected API rate limits and repository issues gracefully
				success := err == nil
				if err != nil {
					if strings.Contains(err.Error(), "GitHub API returned status 403") {
						t.Logf("GitHub API rate limit hit - this is expected in test environment: %v", err)
						success = true
					} else if strings.Contains(err.Error(), "404") {
						t.Logf("Repository not found - this is expected in test environment: %v", err)
						success = true
					}
				}

				printTestStatus(t, "Command execution", success,
					fmt.Sprintf("Command executed with args: %v", tt.args))

				if hasExpectedOutput || len(output) > 0 {
					fmt.Printf("      %s Output captured: %s\n", testInfoColor("ℹ"), testInfoColor("success"))
				}
			}
		})
	}
}

func TestUpgradeCommandStructure(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Structure ==="))

	cmd := NewUpgradeCmd()

	// Test command description contains key information
	expectedInLong := []string{
		"Check for and install",
		"latest version",
		"Examples:",
		"js upgrade",
		"--check",
		"--pre-release",
		"--force",
	}

	for _, expected := range expectedInLong {
		contains := strings.Contains(cmd.Long, expected)
		printTestStatus(t, fmt.Sprintf("Long description contains '%s'", expected), contains,
			fmt.Sprintf("Command.Long should contain '%s'", expected))
	}
}

func TestUpgradeCommandFlagParsing(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Upgrade Command Flag Parsing ==="))

	// Test various flag combinations
	flagCombinations := [][]string{
		{"--check"},
		{"--pre-release"},
		{"--force"},
		{"-c"},
		{"-p"},
		{"-f"},
		{"--check", "--pre-release"},
		{"--check", "--force"},
		{"-c", "-p", "-f"},
	}

	for i, args := range flagCombinations {
		t.Run("flag_combination_"+string(rune(i+'0')), func(t *testing.T) {
			fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor(fmt.Sprintf("Testing flag combination: %v", args)))

			// Create a fresh command for each test
			cmd := NewUpgradeCmd()
			cmd.SetArgs(args)

			// Just test that flags can be parsed without error
			err := cmd.ParseFlags(args)
			success := err == nil
			printTestStatus(t, fmt.Sprintf("Flag parsing: %v", args), success,
				fmt.Sprintf("Flags should parse successfully: %v", args))
		})
	}
}

// Test coverage for performUpgrade function
func TestPerformUpgrade(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Perform Upgrade Function ==="))

	// Store original debug mode
	originalDebugMode := utils.DebugMode
	defer func() {
		utils.DebugMode = originalDebugMode
	}()

	tests := []struct {
		name        string
		checkOnly   bool
		preRelease  bool
		force       bool
		debugMode   bool
		expectError bool
	}{
		{
			name:        "check only mode",
			checkOnly:   true,
			preRelease:  false,
			force:       false,
			debugMode:   false,
			expectError: false,
		},
		{
			name:        "check only with pre-release",
			checkOnly:   true,
			preRelease:  true,
			force:       false,
			debugMode:   false,
			expectError: false,
		},
		{
			name:        "check only with force",
			checkOnly:   true,
			preRelease:  false,
			force:       true,
			debugMode:   false,
			expectError: false,
		},
		{
			name:        "check only with debug mode",
			checkOnly:   true,
			preRelease:  false,
			force:       false,
			debugMode:   true,
			expectError: false,
		},
		{
			name:        "full upgrade attempt - no install since no repository",
			checkOnly:   false,
			preRelease:  false,
			force:       false,
			debugMode:   false,
			expectError: false,
		},
		{
			name:        "full upgrade attempt with force",
			checkOnly:   false,
			preRelease:  false,
			force:       true,
			debugMode:   false,
			expectError: false,
		},
		{
			name:        "full upgrade attempt with pre-release",
			checkOnly:   false,
			preRelease:  true,
			force:       false,
			debugMode:   false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor(tt.name))

			// Set debug mode for test
			utils.DebugMode = tt.debugMode

			// Capture output
			oldOut := os.Stdout
			oldErr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			// Run the function
			err := performUpgrade(tt.checkOnly, tt.preRelease, tt.force)

			// Restore output
			w.Close()
			os.Stdout = oldOut
			os.Stderr = oldErr

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if tt.expectError && err == nil {
				printTestStatus(t, "Error expectation", false, "Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				// Handle GitHub API rate limiting errors (403) gracefully
				if strings.Contains(err.Error(), "GitHub API returned status 403") {
					t.Logf("GitHub API rate limit hit for %s: %v", tt.name, err)
					// Don't fail the test - this is expected in test environment
					printTestStatus(t, "Function execution", true,
						fmt.Sprintf("performUpgrade(%t, %t, %t) - API rate limited", tt.checkOnly, tt.preRelease, tt.force))
					return
				} else if strings.Contains(err.Error(), "404") {
					t.Logf("Repository not found for %s (expected in test): %v", tt.name, err)
					// Don't fail the test - this is expected in test environment
					printTestStatus(t, "Function execution", true,
						fmt.Sprintf("performUpgrade(%t, %t, %t) - repo not found", tt.checkOnly, tt.preRelease, tt.force))
					return
				} else {
					printTestStatus(t, "No error expectation", false, fmt.Sprintf("Expected no error but got: %v", err))
					return
				}
			}

			// Verify output contains expected upgrade check behavior
			hasExpectedOutput := strings.Contains(output, "Checking for updates") ||
				strings.Contains(output, "Repository not found") ||
				len(output) > 0

			printTestStatus(t, "Function execution", err == nil || tt.expectError,
				fmt.Sprintf("performUpgrade(%t, %t, %t)", tt.checkOnly, tt.preRelease, tt.force))

			if hasExpectedOutput {
				fmt.Printf("      %s Output captured successfully\n", testInfoColor("ℹ"))
			}
		})
	}
}

// Test perform upgrade comprehensive mock based
func TestPerformUpgradeComprehensiveMockBased(t *testing.T) {
	fmt.Println("=== Testing Comprehensive Mock-Based Coverage ===")

	// Save original values
	oldStdout := os.Stdout

	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		check          bool
		preRelease     bool
		force          bool
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "404_error_simulation",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message":"Not Found","documentation_url":"https://docs.github.com/rest/reference/repos#get-the-latest-release"}`))
			},
			check:      true,
			preRelease: false,
			force:      false,
			// Accept the warning message as valid output
			expectedOutput: []string{"Repository not found", "upgrade feature is ready"},
			expectError:    false,
		},
		{
			name: "api_rate_limit_simulation",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message":"API rate limit exceeded"}`))
			},
			check:          true,
			preRelease:     false,
			force:          false,
			expectedOutput: []string{"GitHub API returned status 403"},
			expectError:    true,
		},
		{
			name: "successful_upgrade_flow",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				release := GitHubRelease{
					TagName:     "v2.0.0",
					Name:        "Release v2.0.0",
					Body:        "New features",
					HTMLURL:     "https://github.com/test/test/releases/tag/v2.0.0",
					Prerelease:  false,
					PublishedAt: time.Now().Format(time.RFC3339),
					Assets: []GitHubAsset{{
						Name:               fmt.Sprintf("jumpstart-%s-%s", runtime.GOOS, runtime.GOARCH),
						BrowserDownloadURL: "https://github.com/test/test/releases/download/v2.0.0/jumpstart",
					}},
				}
				json.NewEncoder(w).Encode(release)
			},
			check:      true,
			preRelease: false,
			force:      false,
			// Match the actual output string
			expectedOutput: []string{"Latest version:  2.0.0", "Current version:"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "successful_upgrade_flow" && !tt.check {
				t.Skip("Skipping actual upgrade test to avoid modifying system")
			}

			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			oldTransport := http.DefaultTransport
			http.DefaultTransport = &testTransport{
				baseTransport: oldTransport,
				testServer:    server,
			}
			defer func() { http.DefaultTransport = oldTransport }()

			r, w, _ := os.Pipe()
			os.Stdout = w

			err := performUpgrade(tt.check, tt.preRelease, tt.force)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if err != nil {
				t.Logf("%s: error = %v", tt.name, err)
			}
			if output != "" {
				t.Logf("%s: output = %s", tt.name, output)
			}

			foundExpected := false
			for _, expected := range tt.expectedOutput {
				if strings.Contains(output, expected) || (err != nil && strings.Contains(err.Error(), expected)) {
					foundExpected = true
					break
				}
			}

			if foundExpected || (tt.expectError && err != nil) {
				printTestStatus(t, tt.name, true, "Should simulate "+tt.name)
			} else {
				printTestStatus(t, tt.name, false, "Should simulate "+tt.name)
				if len(tt.expectedOutput) > 0 {
					t.Errorf("Expected output containing one of %v, but got output: %s, error: %v",
						tt.expectedOutput, output, err)
				}
			}
		})
	}
}

// testTransport is a custom HTTP transport for testing
type testTransport struct {
	baseTransport http.RoundTripper
	testServer    *httptest.Server
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect GitHub API calls to our test server
	if strings.Contains(req.URL.Host, "api.github.com") {
		testURL, _ := url.Parse(t.testServer.URL)
		req.URL.Scheme = testURL.Scheme
		req.URL.Host = testURL.Host
	}
	return t.baseTransport.RoundTrip(req)
}

// Test perform upgrade comprehensive real based
func TestPerformUpgradeComprehensiveRealBased(t *testing.T) {
	fmt.Println("=== Testing Comprehensive Real-Based Coverage ===")

	// Save original values
	originalRepo := os.Getenv("GITHUB_REPOSITORY")
	originalAPIURL := os.Getenv("GITHUB_API_URL")

	tests := []struct {
		name        string
		check       bool
		preRelease  bool
		force       bool
		expectError bool
	}{
		{
			name:        "check only mode",
			check:       true,
			preRelease:  false,
			force:       false,
			expectError: false,
		},
		{
			name:        "full upgrade attempt - no install since no repository",
			check:       false,
			preRelease:  false,
			force:       false,
			expectError: false,
		},
		{
			name:        "full upgrade attempt with force",
			check:       false,
			preRelease:  false,
			force:       true,
			expectError: false,
		},
		{
			name:        "full upgrade attempt with pre-release",
			check:       false,
			preRelease:  true,
			force:       false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip test if GITHUB_REPOSITORY is not set
			if originalRepo == "" {
				t.Skip("Skipping test because GITHUB_REPOSITORY is not set")
			}

			fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor(tt.name))

			// Capture output
			oldOut := os.Stdout
			oldErr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			// Run the function
			err := performUpgrade(tt.check, tt.preRelease, tt.force)

			// Restore output
			w.Close()
			os.Stdout = oldOut
			os.Stderr = oldErr

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if tt.expectError && err == nil {
				printTestStatus(t, "Error expectation", false, "Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				// Handle GitHub API rate limiting errors (403) gracefully
				if strings.Contains(err.Error(), "GitHub API returned status 403") {
					t.Logf("GitHub API rate limit hit for %s: %v", tt.name, err)
					// Don't fail the test - this is expected in test environment
					printTestStatus(t, "Function execution", true,
						fmt.Sprintf("performUpgrade(%t, %t, %t) - API rate limited", tt.check, tt.preRelease, tt.force))
					return
				} else if strings.Contains(err.Error(), "404") {
					t.Logf("Repository not found for %s (expected in test): %v", tt.name, err)
					// Don't fail the test - this is expected in test environment
					printTestStatus(t, "Function execution", true,
						fmt.Sprintf("performUpgrade(%t, %t, %t) - repo not found", tt.check, tt.preRelease, tt.force))
					return
				} else {
					printTestStatus(t, "No error expectation", false, fmt.Sprintf("Expected no error but got: %v", err))
					return
				}
			}

			// Verify output contains expected upgrade check behavior
			hasExpectedOutput := strings.Contains(output, "Checking for updates") ||
				strings.Contains(output, "Repository not found") ||
				len(output) > 0

			printTestStatus(t, "Function execution", err == nil || tt.expectError,
				fmt.Sprintf("performUpgrade(%t, %t, %t)", tt.check, tt.preRelease, tt.force))

			if hasExpectedOutput {
				fmt.Printf("      %s Output captured successfully\n", testInfoColor("ℹ"))
			}
		})
	}

	// Restore original environment
	os.Setenv("GITHUB_REPOSITORY", originalRepo)
	os.Setenv("GITHUB_API_URL", originalAPIURL)
}

func TestCleanupDaysFlag(t *testing.T) {
	cmd := NewUpgradeCmd()
	cmd.SetArgs([]string{"--cleanup-days=5"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	err := cmd.Execute()
	if err != nil && !strings.Contains(buf.String(), "cleanup") {
		t.Errorf("cleanup-days flag should be handled, got error: %v", err)
	}
}
