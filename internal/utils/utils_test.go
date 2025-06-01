package utils

import (
	"fmt"
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

// Helper function to print test status
func printTestStatus(t *testing.T, testName string, success bool, message string) {
	if success {
		fmt.Printf("%s %s - %s\n", testSuccessColor("✅"), testName, testInfoColor(message))
	} else {
		fmt.Printf("%s %s - %s\n", testErrorColor("❌"), testName, testErrorColor(message))
		t.Error(message)
	}
}

func TestIsAzureLoggedIn(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Azure CLI Login Status ==="))

	// Test the IsAzureLoggedIn function
	// Note: This test may fail if Azure CLI is not installed or configured
	t.Run("azure_cli_check", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Checking Azure CLI login status"))
		result := IsAzureLoggedIn()
		// We can't assert a specific value since it depends on environment
		// Just ensure the function doesn't panic
		printTestStatus(t, "Azure CLI Check", true, fmt.Sprintf("Function executed, result: %v", result))
	})
}

func TestResourceGroupExists(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Resource Group Existence ==="))

	// Test with a resource group that likely doesn't exist
	t.Run("nonexistent_rg", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing non-existent resource group"))
		result := ResourceGroupExists("jumpstart-test-nonexistent-rg-12345")
		// Should return false for non-existent resource group
		printTestStatus(t, "Non-existent RG", !result, "Non-existent resource group should return false")
		if result {
			fmt.Printf("      %s %s\n", testWarnColor("⚠️"), testWarnColor("Warning: Test resource group unexpectedly exists"))
		}
	})

	// Test with empty name
	t.Run("empty_name", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing empty resource group name"))
		result := ResourceGroupExists("")
		printTestStatus(t, "Empty RG Name", !result, "Empty resource group name should return false")
	})

	// Test with invalid characters
	t.Run("invalid_name", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing invalid resource group name"))
		result := ResourceGroupExists("invalid@name!")
		printTestStatus(t, "Invalid RG Name", !result, "Invalid resource group name should return false")
	})
}

func TestCreateResourceGroup(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Resource Group Creation ==="))

	// Test with invalid parameters to ensure error handling
	t.Run("invalid_location", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing invalid location"))
		err := CreateResourceGroup("test-rg", "invalid-location-12345")
		// Should return an error for invalid location
		printTestStatus(t, "Invalid Location", err != nil, "Should return error for invalid location")
		if err == nil {
			fmt.Printf("      %s %s\n", testWarnColor("⚠️"), testWarnColor("Warning: CreateResourceGroup with invalid location didn't return error"))
		}
	})

	// Test with empty parameters
	t.Run("empty_parameters", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing empty parameters"))
		err := CreateResourceGroup("", "")
		printTestStatus(t, "Empty Parameters", err != nil, "Should return error for empty resource group parameters")
	})
}

func TestColoredLogging(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Colored Logging Functions ==="))

	// Save original debug mode
	originalDebugMode := DebugMode
	defer func() { DebugMode = originalDebugMode }()

	// Capture output for testing logging functions
	testCases := []struct {
		name     string
		function func(string, ...interface{})
		message  string
	}{
		{"Info", Info, "Test info message"},
		{"Warn", Warn, "Test warning message"},
		{"Success", Success, "Test success message"},
		{"Debug", Debug, "Test debug message"},
		{"Prompt", Prompt, "Test prompt message"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor(fmt.Sprintf("Testing %s logging", tc.name)))

			// Enable debug mode for Debug function
			if tc.name == "Debug" {
				DebugMode = true
			}

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Test the logging function
			tc.function(tc.message)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read the output
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			// Verify the message was logged
			success := strings.Contains(outputStr, tc.message)
			printTestStatus(t, fmt.Sprintf("%s Function", tc.name), success,
				fmt.Sprintf("Output should contain '%s'", tc.message))
		})
	}
}

func TestErrorLogging(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Error Logging ==="))

	// Test Error function which writes to stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Error("Test error message")

	w.Close()
	os.Stderr = old

	output, _ := io.ReadAll(r)
	outputStr := string(output)

	printTestStatus(t, "Error Message Content", strings.Contains(outputStr, "Test error message"),
		"Stderr output should contain 'Test error message'")
	printTestStatus(t, "Error Prefix", strings.Contains(outputStr, "[ERROR]"),
		"Stderr output should contain '[ERROR]' prefix")
}

func TestDebugMode(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Debug Mode ==="))

	// Save original debug mode
	originalDebugMode := DebugMode

	t.Run("debug_enabled", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing debug mode enabled"))
		DebugMode = true

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		Debug("Test debug message")

		w.Close()
		os.Stdout = old

		output, _ := io.ReadAll(r)
		outputStr := string(output)

		success := strings.Contains(outputStr, "Test debug message")
		printTestStatus(t, "Debug Output When Enabled", success,
			"Should show debug output when DebugMode=true")
	})

	t.Run("debug_disabled", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing debug mode disabled"))
		DebugMode = false

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		Debug("Test debug message")

		w.Close()
		os.Stdout = old

		output, _ := io.ReadAll(r)
		outputStr := string(output)

		success := !strings.Contains(outputStr, "Test debug message")
		printTestStatus(t, "Debug Output When Disabled", success,
			"Should not show debug output when DebugMode=false")
	})

	// Restore original debug mode
	DebugMode = originalDebugMode
}

func TestFriendlyResourceName(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Friendly Resource Name Mapping ==="))

	testCases := []struct {
		resourceType     string
		resourceName     string
		expectedContains string
	}{
		{"Microsoft.OperationalInsights/workspaces", "test-workspace", "Log Analytics workspace"},
		{"Microsoft.Network/networkSecurityGroups", "test-nsg", "Network Security Group"},
		{"Microsoft.KeyVault/vaults", "test-vault", "Azure Key Vault"},
		{"Microsoft.Network/virtualNetworks", "test-vnet", "Virtual Network"},
		{"Microsoft.Compute/disks", "test-disk", "Disk"},
		{"Microsoft.Network/publicIPAddresses", "test-ip", "Public IP Address"},
		{"Microsoft.Network/networkInterfaces", "test-nic", "Network Interface"},
		{"Microsoft.Compute/virtualMachines", "test-vm", "Virtual Machine"},
		{"Microsoft.Resources/deployments", "test-deployment", "Nested Deployment"},
		{"Microsoft.DevTestLab/schedules", "test-schedule", "Schedule"},
		{"Microsoft.Compute/virtualMachines/extensions", "vm1/Bootstrap", "Bootstrap"},
		{"Microsoft.Compute/virtualMachines/extensions", "vm1/Microsoft.Azure.Geneva.GenevaMonitoring", "Azure Geneva Monitoring"},
		{"Microsoft.Compute/virtualMachines/extensions", "vm1/CustomExtension", "CustomExtension"},
		{"Microsoft.Unknown/resources", "test-resource", "Resource"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.resourceType, tc.resourceName), func(t *testing.T) {
			result := FriendlyResourceName(tc.resourceType, tc.resourceName)
			success := strings.Contains(result, tc.expectedContains)
			printTestStatus(t, fmt.Sprintf("Resource mapping: %s", tc.resourceType), success,
				fmt.Sprintf("Should contain '%s', got: %s", tc.expectedContains, result))
		})
	}
}

func TestGlobalVariables(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Global Variables ==="))

	// Test that global variables have reasonable defaults
	t.Run("cli_version", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing CLI version"))
		printTestStatus(t, "CLI Version Not Empty", CliVersion != "", "CliVersion should not be empty")
		if CliVersion != "0.1.0" {
			fmt.Printf("      %s %s: %s\n", testInfoColor("ℹ"), testInfoColor("CliVersion is set to"), testInfoColor(CliVersion))
		}
	})

	t.Run("color_functions", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing color functions"))
		// Test that color functions don't panic
		testStr := "test"

		printTestStatus(t, "InfoColor Function", InfoColor(testStr) != "", "InfoColor should not return empty string")
		printTestStatus(t, "WarnColor Function", WarnColor(testStr) != "", "WarnColor should not return empty string")
		printTestStatus(t, "ErrorColor Function", ErrorColor(testStr) != "", "ErrorColor should not return empty string")
		printTestStatus(t, "SuccessColor Function", SuccessColor(testStr) != "", "SuccessColor should not return empty string")
		printTestStatus(t, "DebugColor Function", DebugColor(testStr) != "", "DebugColor should not return empty string")
		printTestStatus(t, "FatalColor Function", FatalColor(testStr) != "", "FatalColor should not return empty string")
		printTestStatus(t, "PromptColor Function", PromptColor(testStr) != "", "PromptColor should not return empty string")
	})

	t.Run("mode_flags", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing mode flags"))
		// Test that mode flags can be set
		originalDebug := DebugMode
		originalVerbose := VerboseMode
		originalOutput := OutputFormat

		DebugMode = true
		VerboseMode = true
		OutputFormat = "json"

		printTestStatus(t, "DebugMode Setting", DebugMode, "DebugMode should be settable to true")
		printTestStatus(t, "VerboseMode Setting", VerboseMode, "VerboseMode should be settable to true")
		printTestStatus(t, "OutputFormat Setting", OutputFormat == "json", "OutputFormat should be settable")

		// Restore original values
		DebugMode = originalDebug
		VerboseMode = originalVerbose
		OutputFormat = originalOutput
	})
}

func TestLoggingFormats(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Logging Formats ==="))

	// Test logging with format strings
	t.Run("formatted_logging", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing formatted logging"))

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		Info("Test %s with %d parameters", "message", 2)

		w.Close()
		os.Stdout = old

		output, _ := io.ReadAll(r)
		outputStr := string(output)

		success := strings.Contains(outputStr, "Test message with 2 parameters")
		printTestStatus(t, "Formatted Output", success, "Should format strings correctly")
	})
}

// Test helper function to check if all required functions are exported
func TestExportedFunctions(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Exported Functions ==="))

	// Test that main utility functions are accessible
	t.Run("main_functions_accessible", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing function accessibility"))

		// These should not panic when called with valid parameters
		defer func() {
			if r := recover(); r != nil {
				printTestStatus(t, "Function Panic Check", false, fmt.Sprintf("Function call panicked: %v", r))
				return
			}
			printTestStatus(t, "Function Accessibility", true, "All main functions are accessible")
		}()

		// Test that functions can be called (even if they might fail due to environment)
		_ = IsAzureLoggedIn()
		_ = ResourceGroupExists("test")
		_ = FriendlyResourceName("Microsoft.Test/test", "test")
	})
}
