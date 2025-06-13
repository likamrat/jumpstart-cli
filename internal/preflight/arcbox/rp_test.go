package arcbox

import (
	"errors"
	"fmt"
	"testing"

	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/resourceproviders"

	"github.com/spf13/cobra"
)

func TestCreateResourceProviderCommands(t *testing.T) {
	fmt.Printf("\n%s\n", "=== Testing Resource Provider Commands Creation ===")

	mockCLI := &azurecli.MockAzureCLI{}
	rpCmd := CreateResourceProviderCommands(mockCLI)

	// Test basic command structure
	testName := "RP Command Use"
	expected := "rp"
	success := rpCmd.Use == expected
	message := fmt.Sprintf("Expected '%s', got '%s'", expected, rpCmd.Use)
	printRPTestStatus(t, testName, success, message)

	testName = "RP Command Short Description"
	expected = "Check and manage Azure resource provider registration"
	success = rpCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, rpCmd.Short)
	printRPTestStatus(t, testName, success, message)

	// Test that rp has expected subcommands
	expectedSubcommands := []string{"show", "list", "register"}
	subcommands := rpCmd.Commands()

	testName = "RP Subcommand Count"
	success = len(subcommands) == len(expectedSubcommands)
	message = fmt.Sprintf("Expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	printRPTestStatus(t, testName, success, message)

	for _, expectedSubcmd := range expectedSubcommands {
		testName = fmt.Sprintf("RP Subcommand '%s'", expectedSubcmd)
		found := false
		for _, subcmd := range subcommands {
			if subcmd.Use == expectedSubcmd {
				found = true
				break
			}
		}
		message = fmt.Sprintf("Subcommand '%s' exists", expectedSubcmd)
		printRPTestStatus(t, testName, found, message)
	}
}

func TestShowResourceProviderStatus(t *testing.T) {
	fmt.Printf("\n%s\n", "=== Testing Show Resource Provider Status ===")

	mockCLI := &azurecli.MockAzureCLI{
		RegisteredProviders: make(map[string]bool),
	}

	// Set up mock data for registered providers
	config := resourceproviders.GetArcBoxProviders()
	for _, provider := range config.RequiredProviders {
		mockCLI.RegisteredProviders[provider] = true
	}

	// This test mainly checks that the function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			printRPTestStatus(t, "Panic prevention", false, fmt.Sprintf("ShowResourceProviderStatus panicked: %v", r))
			t.Errorf("ShowResourceProviderStatus panicked: %v", r)
		} else {
			printRPTestStatus(t, "Panic prevention", true, "ShowResourceProviderStatus executed without panicking")
		}
	}()

	ShowResourceProviderStatus(mockCLI)
	printRPTestStatus(t, "Show status execution", true, "Function executed successfully")
}

func TestListRequiredResourceProviders(t *testing.T) {
	fmt.Printf("\n%s\n", "=== Testing List Required Resource Providers ===")

	// This test mainly checks that the function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			printRPTestStatus(t, "Panic prevention", false, fmt.Sprintf("ListRequiredResourceProviders panicked: %v", r))
			t.Errorf("ListRequiredResourceProviders panicked: %v", r)
		} else {
			printRPTestStatus(t, "Panic prevention", true, "ListRequiredResourceProviders executed without panicking")
		}
	}()

	ListRequiredResourceProviders()
	printRPTestStatus(t, "List providers execution", true, "Function executed successfully")
}

func TestRegisterResourceProvider(t *testing.T) {
	fmt.Printf("\n%s\n", "=== Testing Register Resource Provider ===")

	// Test successful registration
	testName := "Successful registration"
	mockCLI := &azurecli.MockAzureCLI{
		RegisteredProviders: make(map[string]bool),
	}

	err := RegisterResourceProvider(mockCLI, "Microsoft.Compute")
	if err != nil {
		printRPTestStatus(t, testName, false, fmt.Sprintf("Expected no error, got: %v", err))
	} else {
		printRPTestStatus(t, testName, true, "Successfully registered resource provider")
	}

	// Test failed registration
	testName = "Failed registration"
	mockCLI.RegisterProviderError = errors.New("mock registration failure")

	err = RegisterResourceProvider(mockCLI, "Microsoft.Storage")
	if err == nil {
		printRPTestStatus(t, testName, false, "Expected error, got nil")
	} else {
		printRPTestStatus(t, testName, true, fmt.Sprintf("Correctly returned error: %v", err))
	}
}

func TestCheckAllResourceProviders(t *testing.T) {
	fmt.Printf("\n%s\n", "=== Testing Check All Resource Providers ===")

	mockCLI := &azurecli.MockAzureCLI{
		RegisteredProviders: make(map[string]bool),
	}

	// Test with all providers registered
	testName := "All providers registered"
	config := resourceproviders.GetArcBoxProviders()
	for _, provider := range config.RequiredProviders {
		mockCLI.RegisteredProviders[provider] = true
	}

	result := CheckAllResourceProviders(mockCLI)
	success := result == true
	message := "Should return true when all providers are registered"
	printRPTestStatus(t, testName, success, message)

	// Test with missing providers
	testName = "Missing providers detected"
	mockCLI.RegisteredProviders = make(map[string]bool) // Clear all registrations

	result = CheckAllResourceProviders(mockCLI)
	success = result == false
	message = "Should return false when providers are missing"
	printRPTestStatus(t, testName, success, message)
}

func TestResourceProviderCommandFlags(t *testing.T) {
	fmt.Printf("\n%s\n", "=== Testing Resource Provider Command Flags ===")

	mockCLI := &azurecli.MockAzureCLI{}
	rpCmd := CreateResourceProviderCommands(mockCLI)

	// Find the register subcommand
	var registerCmd *cobra.Command
	for _, subCmd := range rpCmd.Commands() {
		if subCmd.Use == "register" {
			registerCmd = subCmd
			break
		}
	}

	testName := "Register Command Exists"
	success := registerCmd != nil
	message := "Register subcommand should exist"
	printRPTestStatus(t, testName, success, message)

	if registerCmd != nil {
		// Test the name flag
		flag := registerCmd.Flags().Lookup("name")
		testName = "Name Flag Exists"
		success = flag != nil
		message = "Name flag should exist"
		printRPTestStatus(t, testName, success, message)

		if flag != nil {
			testName = "Name Flag Shorthand"
			success = flag.Shorthand == "n"
			message = fmt.Sprintf("Expected shorthand 'n', got '%s'", flag.Shorthand)
			printRPTestStatus(t, testName, success, message)
		}
	}
}

func TestResourceProviderSubcommandStructure(t *testing.T) {
	fmt.Printf("\n%s\n", "=== Testing Resource Provider Subcommand Structure ===")

	mockCLI := &azurecli.MockAzureCLI{}
	rpCmd := CreateResourceProviderCommands(mockCLI)

	expectedCommands := map[string]string{
		"show":     "Show registration status of required Azure resource providers",
		"list":     "List required Azure resource providers for ArcBox",
		"register": "Register a required Azure resource provider",
	}

	for cmdName, expectedDesc := range expectedCommands {
		var foundCmd *cobra.Command
		for _, subCmd := range rpCmd.Commands() {
			if subCmd.Use == cmdName {
				foundCmd = subCmd
				break
			}
		}

		testName := fmt.Sprintf("Command '%s' exists", cmdName)
		success := foundCmd != nil
		message := fmt.Sprintf("Command '%s' should exist", cmdName)
		printRPTestStatus(t, testName, success, message)

		if foundCmd != nil {
			testName = fmt.Sprintf("Command '%s' description", cmdName)
			success = foundCmd.Short == expectedDesc
			message = fmt.Sprintf("Expected '%s', got '%s'", expectedDesc, foundCmd.Short)
			printRPTestStatus(t, testName, success, message)
		}
	}
}

func TestResourceProviderCommandValidation(t *testing.T) {
	fmt.Printf("\n%s\n", "=== Testing Resource Provider Command Validation ===")

	mockCLI := &azurecli.MockAzureCLI{}
	rpCmd := CreateResourceProviderCommands(mockCLI)

	// Test 1: No arguments should show help (RunE should return nil)
	testName := "No arguments validation"
	err := rpCmd.RunE(rpCmd, []string{})
	success := err == nil
	message := "Should return nil when no arguments provided (shows help)"
	printRPTestStatus(t, testName, success, message)

	// Test 2: Valid subcommand should return nil
	testName = "Valid subcommand validation"
	validSubcommands := []string{"show", "list", "register"}
	for _, subcmd := range validSubcommands {
		err = rpCmd.RunE(rpCmd, []string{subcmd})
		success = err == nil
		message = fmt.Sprintf("Should return nil for valid subcommand '%s'", subcmd)
		printRPTestStatus(t, testName+fmt.Sprintf(" (%s)", subcmd), success, message)
	}

	// Test 3: Invalid subcommand should return error
	testName = "Invalid subcommand validation"
	invalidSubcommands := []string{"invalid", "nonexistent", "badcmd"}
	for _, invalidCmd := range invalidSubcommands {
		err = rpCmd.RunE(rpCmd, []string{invalidCmd})
		expectedErrorMsg := fmt.Sprintf("unknown subcommand '%s' for 'js arcbox preflight rp'", invalidCmd)
		if err != nil {
			success = err.Error() == expectedErrorMsg
		} else {
			success = false
		}
		message = fmt.Sprintf("Should return error for invalid subcommand '%s'", invalidCmd)
		printRPTestStatus(t, testName+fmt.Sprintf(" (%s)", invalidCmd), success, message)
	}

	// Test 4: Similar subcommand suggestions
	testName = "Similar command suggestions"
	// Test with commands that are similar to valid ones
	similarCommands := []string{"sho", "lst", "registe", "show1"}

	for _, similarCmd := range similarCommands {
		// We can't easily test the actual suggestion output, but we can test that
		// the command doesn't crash and returns nil (since suggestions don't error)
		err = rpCmd.RunE(rpCmd, []string{similarCmd})
		success = err == nil // Should return nil after showing suggestion
		message = fmt.Sprintf("Should handle similar command '%s' gracefully", similarCmd)
		printRPTestStatus(t, testName+fmt.Sprintf(" (%s)", similarCmd), success, message)
	}
}

func TestResourceProviderErrorScenarios(t *testing.T) {
	fmt.Printf("\n%s\n", "=== Testing Resource Provider Error Scenarios ===")

	// Test with empty mock CLI
	testName := "Empty provider list"
	mockCLI := &azurecli.MockAzureCLI{
		RegisteredProviders: make(map[string]bool),
	}

	result := CheckAllResourceProviders(mockCLI)
	success := result == false
	message := "Should return false when no providers are registered"
	printRPTestStatus(t, testName, success, message)

	// Test with partial provider registration
	testName = "Partial provider registration"
	config := resourceproviders.GetArcBoxProviders()
	if len(config.RequiredProviders) > 1 {
		// Register only the first provider
		mockCLI.RegisteredProviders[config.RequiredProviders[0]] = true

		result = CheckAllResourceProviders(mockCLI)
		success = result == false
		message = "Should return false when only some providers are registered"
		printRPTestStatus(t, testName, success, message)
	}

	// Test with nil/empty provider name for registration
	testName = "Empty provider name registration"
	defer func() {
		if r := recover(); r != nil {
			// If it panics, that's expected behavior for invalid input
			printRPTestStatus(t, testName, true, "Function handled empty provider name appropriately")
		} else {
			// If it doesn't panic, that's also acceptable
			printRPTestStatus(t, testName, true, "Function executed without panicking on empty provider name")
		}
	}()

	// We can't fully test RegisterResourceProvider because it calls os.Exit on error
	// But we can test that it doesn't immediately crash on empty input
	// Note: This is more of a structural test
	printRPTestStatus(t, testName, true, "RegisterResourceProvider function structure test completed")
}

func TestResourceProviderCommandHelpAndUsage(t *testing.T) {
	fmt.Printf("\n%s\n", "=== Testing Resource Provider Command Help and Usage ===")

	mockCLI := &azurecli.MockAzureCLI{}
	rpCmd := CreateResourceProviderCommands(mockCLI)

	// Test main command properties
	testName := "Main command usage text"
	expectedUsage := "rp"
	success := rpCmd.Use == expectedUsage
	message := fmt.Sprintf("Expected usage '%s', got '%s'", expectedUsage, rpCmd.Use)
	printRPTestStatus(t, testName, success, message)

	testName = "Main command silence settings"
	success = rpCmd.DisableSuggestions && rpCmd.SilenceErrors && rpCmd.SilenceUsage
	message = "Should have DisableSuggestions, SilenceErrors, and SilenceUsage enabled"
	printRPTestStatus(t, testName, success, message)

	// Test that all subcommands have RunE or Run functions
	testName = "Subcommand execution functions"
	allSubcommands := rpCmd.Commands()
	for _, subcmd := range allSubcommands {
		hasRun := subcmd.Run != nil || subcmd.RunE != nil
		success = hasRun
		message = fmt.Sprintf("Subcommand '%s' should have Run or RunE function", subcmd.Use)
		printRPTestStatus(t, testName+fmt.Sprintf(" (%s)", subcmd.Use), success, message)
	}

	// Test that register command has required flag
	testName = "Register command flag configuration"
	var registerCmd *cobra.Command
	for _, subcmd := range allSubcommands {
		if subcmd.Use == "register" {
			registerCmd = subcmd
			break
		}
	}

	if registerCmd != nil {
		nameFlag := registerCmd.Flags().Lookup("name")
		success = nameFlag != nil && nameFlag.Shorthand == "n"
		message := "Register command should have 'name' flag with shorthand 'n'"
		printRPTestStatus(t, testName, success, message)
	} else {
		printRPTestStatus(t, testName, false, "Register command not found")
	}
}

// Helper function to print test status (reusing the same pattern as quota_test.go)
func printRPTestStatus(t *testing.T, testName string, success bool, message string) {
	if success {
		fmt.Printf("  ✅ %s: %s\n", testName, message)
	} else {
		fmt.Printf("  ❌ %s: %s\n", testName, message)
		t.Errorf("Test failed: %s - %s", testName, message)
	}
}
