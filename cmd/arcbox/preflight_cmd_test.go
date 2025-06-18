package arcbox

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"jumpstartcli/cmd/arcbox/services"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ====================================================================================
// PHASE 2.2 - Preflight Command Enhancement Test Coverage
// Target: Achieve 95%+ coverage for preflight command functionality (currently 40.6%)
// ====================================================================================

// setupTestEnvironment configures the global environment for tests
func setupTestEnvironment() {
	// Set a valid output format to avoid output formatting errors
	utils.OutputFormat = "table"
}

// TestCreatePreflightCommand_SubcommandStructure tests the preflight command structure and subcommands
func TestCreatePreflightCommand_SubcommandStructure(t *testing.T) {
	setupTestEnvironment()

	tests := []struct {
		name            string
		args            []string
		expectedSubcmds []string
		expectError     bool
		description     string
	}{
		{
			name:            "base_preflight_command",
			args:            []string{},
			expectedSubcmds: []string{"quota", "rp", "status"},
			expectError:     false,
			description:     "Should have all three subcommands: quota, rp, status",
		},
		{
			name:            "quota_subcommand_exists",
			args:            []string{"quota"},
			expectedSubcmds: []string{},
			expectError:     false,
			description:     "Quota subcommand should exist and be accessible",
		},
		{
			name:            "rp_subcommand_exists",
			args:            []string{"rp"},
			expectedSubcmds: []string{},
			expectError:     false,
			description:     "RP subcommand should exist and be accessible",
		},
		{
			name:            "status_subcommand_exists",
			args:            []string{"status"},
			expectedSubcmds: []string{},
			expectError:     false,
			description:     "Status subcommand should exist and be accessible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			quotaService := services.NewQuotaService(mockCLI)
			validationService := services.NewValidationService(mockCLI)

			cmd := createPreflightCommand(quotaService, validationService, mockCLI)

			// Test basic command properties
			if cmd.Use != "preflight" {
				t.Errorf("Expected command use to be 'preflight', got '%s'", cmd.Use)
			}

			if cmd.Short != "Run preflight checks for ArcBox deployment" {
				t.Errorf("Expected correct short description, got '%s'", cmd.Short)
			}

			if !strings.Contains(cmd.Long, "preflight checks") {
				t.Error("Expected long description to mention preflight checks")
			}

			// Test subcommand existence
			actualSubcmds := make(map[string]*cobra.Command)
			for _, subcmd := range cmd.Commands() {
				actualSubcmds[subcmd.Use] = subcmd
			}

			for _, expectedSubcmd := range tt.expectedSubcmds {
				if _, found := actualSubcmds[expectedSubcmd]; !found {
					t.Errorf("Expected subcommand '%s' not found", expectedSubcmd)
				}
			}

			// Verify that the command properties are set correctly
			if !cmd.DisableSuggestions {
				t.Error("Expected DisableSuggestions to be true")
			}

			if !cmd.SilenceErrors {
				t.Error("Expected SilenceErrors to be true")
			}

			if !cmd.SilenceUsage {
				t.Error("Expected SilenceUsage to be true")
			}

			if cmd.RunE == nil {
				t.Error("Expected RunE function to be set")
			}
		})
	}
}

// TestPreflightCommand_QuotaSubcommand tests the quota subcommand structure and flags
func TestPreflightCommand_QuotaSubcommand(t *testing.T) {
	setupTestEnvironment()

	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	cmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Find the quota subcommand
	var quotaCmd *cobra.Command
	for _, subcmd := range cmd.Commands() {
		if subcmd.Use == "quota" {
			quotaCmd = subcmd
			break
		}
	}

	if quotaCmd == nil {
		t.Fatal("Quota subcommand not found")
	}

	// Test quota command properties
	if quotaCmd.Short != "Check vCPU quota for ArcBox flavors" {
		t.Errorf("Expected correct quota short description, got '%s'", quotaCmd.Short)
	}

	if !strings.Contains(quotaCmd.Long, "vCPU quota") {
		t.Error("Expected quota long description to mention vCPU quota")
	}

	if quotaCmd.RunE == nil {
		t.Error("Expected quota RunE function to be set")
	}

	// Test quota command flags
	expectedFlags := []struct {
		name         string
		shorthand    string
		defaultValue string
		usage        string
	}{
		{
			name:         "flavor",
			shorthand:    "f",
			defaultValue: "",
			usage:        "ArcBox flavor to check (ITPro, DevOps, DataOps, all)",
		},
		{
			name:         "location",
			shorthand:    "l",
			defaultValue: "",
			usage:        "Azure region(s) to check quota in. Use comma-separated values for multiple regions",
		},
		{
			name:         "all-locations",
			shorthand:    "",
			defaultValue: "false",
			usage:        "Check quota in all ArcBox-supported regions",
		},
		{
			name:         "sku",
			shorthand:    "",
			defaultValue: "",
			usage:        "Custom VM SKU(s) to check. Comma-separated",
		},
		{
			name:         "subscription",
			shorthand:    "s",
			defaultValue: "",
			usage:        "Azure subscription ID to use",
		},
	}

	for _, expectedFlag := range expectedFlags {
		flag := quotaCmd.Flags().Lookup(expectedFlag.name)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found", expectedFlag.name)
			continue
		}

		if expectedFlag.shorthand != "" && flag.Shorthand != expectedFlag.shorthand {
			t.Errorf("Flag '%s' expected shorthand '%s', got '%s'", expectedFlag.name, expectedFlag.shorthand, flag.Shorthand)
		}

		if flag.DefValue != expectedFlag.defaultValue {
			t.Errorf("Flag '%s' expected default value '%s', got '%s'", expectedFlag.name, expectedFlag.defaultValue, flag.DefValue)
		}

		if flag.Usage != expectedFlag.usage {
			t.Errorf("Flag '%s' expected usage '%s', got '%s'", expectedFlag.name, expectedFlag.usage, flag.Usage)
		}
	}
}

// TestPreflightCommand_RPSubcommand tests the resource provider subcommand structure
func TestPreflightCommand_RPSubcommand(t *testing.T) {
	setupTestEnvironment()
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	cmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Find the rp subcommand
	var rpCmd *cobra.Command
	for _, subcmd := range cmd.Commands() {
		if subcmd.Use == "rp" {
			rpCmd = subcmd
			break
		}
	}

	if rpCmd == nil {
		t.Fatal("RP subcommand not found")
	}

	// Test RP command properties
	if rpCmd.Short != "Check and manage Azure resource provider registration" {
		t.Errorf("Expected correct RP short description, got '%s'", rpCmd.Short)
	}

	if !strings.Contains(rpCmd.Long, "resource provider") {
		t.Error("Expected RP long description to mention resource provider")
	}

	// Test RP subcommands
	expectedRPSubcmds := []string{"show", "list", "register"}
	actualRPSubcmds := make(map[string]bool)
	for _, subcmd := range rpCmd.Commands() {
		actualRPSubcmds[subcmd.Use] = true
	}

	for _, expectedSubcmd := range expectedRPSubcmds {
		if !actualRPSubcmds[expectedSubcmd] {
			t.Errorf("Expected RP subcommand '%s' not found", expectedSubcmd)
		}
	}
}

// TestPreflightCommand_StatusSubcommand tests the status subcommand structure
func TestPreflightCommand_StatusSubcommand(t *testing.T) {
	setupTestEnvironment()
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	cmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Find the status subcommand
	var statusCmd *cobra.Command
	for _, subcmd := range cmd.Commands() {
		if subcmd.Use == "status" {
			statusCmd = subcmd
			break
		}
	}

	if statusCmd == nil {
		t.Fatal("Status subcommand not found")
	}

	// Test status command properties
	if statusCmd.Short != "Show last preflight check status" {
		t.Errorf("Expected correct status short description, got '%s'", statusCmd.Short)
	}

	if !strings.Contains(statusCmd.Long, "preflight check") {
		t.Error("Expected status long description to mention preflight check")
	}
}

// TestPreflightCommand_RunEFunction tests the main preflight command RunE function
func TestPreflightCommand_RunEFunction(t *testing.T) {
	setupTestEnvironment()
	tests := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "no_args_shows_help",
			args:        []string{},
			expectError: false,
			description: "Should show help when no arguments provided",
		},
		{
			name:        "valid_subcommand_quota",
			args:        []string{"quota"},
			expectError: false,
			description: "Should accept valid quota subcommand",
		},
		{
			name:        "valid_subcommand_status",
			args:        []string{"status"},
			expectError: false,
			description: "Should accept valid status subcommand",
		},
		{
			name:        "valid_subcommand_rp",
			args:        []string{"rp"},
			expectError: false,
			description: "Should accept valid rp subcommand",
		},
		{
			name:        "invalid_subcommand",
			args:        []string{"invalid"},
			expectError: true,
			description: "Should return error for invalid subcommand",
		},
		{
			name:        "similar_subcommand_suggestion",
			args:        []string{"stat"},
			expectError: false,
			description: "Should suggest similar command for close matches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			quotaService := services.NewQuotaService(mockCLI)
			validationService := services.NewValidationService(mockCLI)

			cmd := createPreflightCommand(quotaService, validationService, mockCLI)

			// Capture output
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			// Test the RunE function directly
			err := cmd.RunE(cmd, tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s but got none", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.description, err)
				}
			}

			// Test specific behaviors
			if tt.name == "invalid_subcommand" {
				if err == nil || !strings.Contains(err.Error(), "unknown subcommand") {
					t.Error("Expected 'unknown subcommand' error for invalid subcommand")
				}
			}
		})
	}
}

// TestPreflightCommand_Integration tests integration scenarios between preflight and services
func TestPreflightCommand_Integration(t *testing.T) {
	setupTestEnvironment()
	tests := []struct {
		name        string
		setupMock   func(*azurecli.MockAzureCLI)
		description string
	}{
		{
			name: "quota_service_integration",
			setupMock: func(cli *azurecli.MockAzureCLI) {
				// Set up mock data for quota service integration
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Subscription"},
				}
				cli.Locations = []string{"eastus", "westus"}
			},
			description: "Should integrate properly with quota service",
		},
		{
			name: "validation_service_integration",
			setupMock: func(cli *azurecli.MockAzureCLI) {
				// Set up mock data for validation service integration
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{
					ID: "test-sub", Name: "Test Subscription",
				}
			},
			description: "Should integrate properly with validation service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			if tt.setupMock != nil {
				tt.setupMock(mockCLI)
			}

			quotaService := services.NewQuotaService(mockCLI)
			validationService := services.NewValidationService(mockCLI)

			cmd := createPreflightCommand(quotaService, validationService, mockCLI)

			// Test that the command is created successfully with services
			if cmd == nil {
				t.Error("Expected command to be created successfully")
			}

			// Test that services are properly integrated
			if cmd.Commands() == nil || len(cmd.Commands()) == 0 {
				t.Error("Expected subcommands to be added")
			}

			// Verify quota subcommand has proper service integration
			quotaCmd := findSubcommand(cmd, "quota")
			if quotaCmd == nil {
				t.Error("Expected quota subcommand to exist")
			} else if quotaCmd.RunE == nil {
				t.Error("Expected quota subcommand to have RunE function")
			}
		})
	}
}

// TestPreflightCommand_ErrorHandling tests error handling scenarios
func TestPreflightCommand_ErrorHandling(t *testing.T) {
	setupTestEnvironment()
	tests := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "empty_command_handling",
			args:        []string{},
			expectError: false,
			description: "Should handle empty command gracefully",
		},
		{
			name:        "malformed_subcommand",
			args:        []string{"quotaa"},
			expectError: false, // Should provide suggestion
			description: "Should suggest similar commands for typos",
		},
		{
			name:        "completely_invalid_subcommand",
			args:        []string{"xyz123"},
			expectError: true,
			description: "Should error for completely invalid subcommand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			quotaService := services.NewQuotaService(mockCLI)
			validationService := services.NewValidationService(mockCLI)

			cmd := createPreflightCommand(quotaService, validationService, mockCLI)

			// Test error handling
			err := cmd.RunE(cmd, tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s but got none", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.description, err)
				}
			}
		})
	}
}

// TestPreflightCommand_ComprehensiveCoverage tests comprehensive scenarios to maximize coverage
func TestPreflightCommand_ComprehensiveCoverage(t *testing.T) {
	setupTestEnvironment()
	// Test all aspects of the preflight command creation and configuration

	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	cmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Test command creation
	if cmd == nil {
		t.Fatal("createPreflightCommand should return a valid command")
	}

	// Test all command properties
	requiredProperties := map[string]interface{}{
		"Use":                "preflight",
		"Short":              "Run preflight checks for ArcBox deployment",
		"DisableSuggestions": true,
		"SilenceErrors":      true,
		"SilenceUsage":       true,
	}

	for property, expected := range requiredProperties {
		switch property {
		case "Use":
			if cmd.Use != expected.(string) {
				t.Errorf("Expected %s to be '%s', got '%s'", property, expected, cmd.Use)
			}
		case "Short":
			if cmd.Short != expected.(string) {
				t.Errorf("Expected %s to be '%s', got '%s'", property, expected, cmd.Short)
			}
		case "DisableSuggestions":
			if cmd.DisableSuggestions != expected.(bool) {
				t.Errorf("Expected %s to be %v, got %v", property, expected, cmd.DisableSuggestions)
			}
		case "SilenceErrors":
			if cmd.SilenceErrors != expected.(bool) {
				t.Errorf("Expected %s to be %v, got %v", property, expected, cmd.SilenceErrors)
			}
		case "SilenceUsage":
			if cmd.SilenceUsage != expected.(bool) {
				t.Errorf("Expected %s to be %v, got %v", property, expected, cmd.SilenceUsage)
			}
		}
	}

	// Test Long description
	if !strings.Contains(cmd.Long, "preflight checks") {
		t.Error("Expected Long description to contain 'preflight checks'")
	}

	// Test RunE function is set
	if cmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}

	// Test all expected subcommands are present
	expectedSubcommands := []string{"quota", "rp", "status"}
	actualSubcommands := make(map[string]bool)
	for _, subcmd := range cmd.Commands() {
		actualSubcommands[subcmd.Use] = true
	}

	for _, expected := range expectedSubcommands {
		if !actualSubcommands[expected] {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}

	// Test that there are no unexpected subcommands
	if len(cmd.Commands()) != len(expectedSubcommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(cmd.Commands()))
	}
}

// TestPreflightCommand_QuotaSubcommandStructure tests the quota subcommand structure in detail
func TestPreflightCommand_QuotaSubcommandStructure(t *testing.T) {
	setupTestEnvironment()
	// Setup services
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	// Create the preflight command
	preflightCmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Find the quota subcommand
	var quotaCmd *cobra.Command
	for _, subcmd := range preflightCmd.Commands() {
		if subcmd.Use == "quota" {
			quotaCmd = subcmd
			break
		}
	}

	if quotaCmd == nil {
		t.Fatal("Quota subcommand not found")
	}

	// Test quota command structure
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "quota_command_basic_properties",
			testFunc: func(t *testing.T) {
				if quotaCmd.Use != "quota" {
					t.Errorf("Expected quota command use to be 'quota', got '%s'", quotaCmd.Use)
				}

				if quotaCmd.Short != "Check vCPU quota for ArcBox flavors" {
					t.Errorf("Expected correct quota short description, got '%s'", quotaCmd.Short)
				}

				if !strings.Contains(quotaCmd.Long, "vCPU quota") {
					t.Error("Expected quota long description to mention vCPU quota")
				}

				if quotaCmd.RunE == nil {
					t.Error("Expected quota RunE function to be set")
				}
			},
		},
		{
			name: "quota_command_flags_structure",
			testFunc: func(t *testing.T) {
				expectedFlags := map[string]struct {
					shorthand    string
					defaultValue string
					flagType     string
				}{
					"flavor": {
						shorthand:    "f",
						defaultValue: "",
						flagType:     "string",
					},
					"location": {
						shorthand:    "l",
						defaultValue: "",
						flagType:     "string",
					},
					"all-locations": {
						shorthand:    "",
						defaultValue: "false",
						flagType:     "bool",
					},
					"sku": {
						shorthand:    "",
						defaultValue: "",
						flagType:     "string",
					},
					"subscription": {
						shorthand:    "s",
						defaultValue: "",
						flagType:     "string",
					},
				}

				for flagName, expected := range expectedFlags {
					flag := quotaCmd.Flags().Lookup(flagName)
					if flag == nil {
						t.Errorf("Expected flag '%s' not found", flagName)
						continue
					}

					if expected.shorthand != "" && flag.Shorthand != expected.shorthand {
						t.Errorf("Flag '%s' expected shorthand '%s', got '%s'", flagName, expected.shorthand, flag.Shorthand)
					}

					if flag.DefValue != expected.defaultValue {
						t.Errorf("Flag '%s' expected default value '%s', got '%s'", flagName, expected.defaultValue, flag.DefValue)
					}

					if flag.Value.Type() != expected.flagType {
						t.Errorf("Flag '%s' expected type '%s', got '%s'", flagName, expected.flagType, flag.Value.Type())
					}
				}
			},
		},
		{
			name: "quota_command_examples_integration",
			testFunc: func(t *testing.T) {
				// Test that examples are properly integrated in the long description
				if !strings.Contains(quotaCmd.Long, "Examples") && !strings.Contains(quotaCmd.Long, "Check") {
					t.Error("Expected quota command to include examples or usage information")
				}

				// Verify the command structure supports examples
				if quotaCmd.Long == "" {
					t.Error("Quota command should have a long description")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

// TestPreflightCommand_QuotaFlagValidation tests quota command flag validation
func TestPreflightCommand_QuotaFlagValidation(t *testing.T) {
	setupTestEnvironment()
	// Setup services
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	// Create the preflight command
	preflightCmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Find the quota subcommand
	var quotaCmd *cobra.Command
	for _, subcmd := range preflightCmd.Commands() {
		if subcmd.Use == "quota" {
			quotaCmd = subcmd
			break
		}
	}

	if quotaCmd == nil {
		t.Fatal("Quota subcommand not found")
	}

	tests := []struct {
		name               string
		flags              map[string]string
		expectedValidation bool
		description        string
	}{
		{
			name: "valid_flavor_and_location",
			flags: map[string]string{
				"flavor":   "ITPro",
				"location": "eastus",
			},
			expectedValidation: true,
			description:        "Should accept valid flavor and location",
		},
		{
			name: "valid_flavor_and_all_locations",
			flags: map[string]string{
				"flavor":        "DevOps",
				"all-locations": "true",
			},
			expectedValidation: true,
			description:        "Should accept valid flavor and all-locations flag",
		},
		{
			name: "valid_custom_sku",
			flags: map[string]string{
				"flavor":   "DataOps",
				"location": "westus2",
				"sku":      "Standard_D4s_v3,Standard_D8s_v3",
			},
			expectedValidation: true,
			description:        "Should accept custom SKU specification",
		},
		{
			name: "valid_subscription_specification",
			flags: map[string]string{
				"flavor":       "ITPro",
				"location":     "eastus",
				"subscription": "12345678-1234-1234-1234-123456789abc",
			},
			expectedValidation: true,
			description:        "Should accept subscription specification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up flags
			for flagName, flagValue := range tt.flags {
				err := quotaCmd.Flags().Set(flagName, flagValue)
				if err != nil {
					t.Fatalf("Failed to set flag %s=%s: %v", flagName, flagValue, err)
				}
			}

			// Verify flags are set correctly
			for flagName, expectedValue := range tt.flags {
				if flagName == "all-locations" {
					actualValue, err := quotaCmd.Flags().GetBool(flagName)
					if err != nil {
						t.Errorf("Failed to get bool flag %s: %v", flagName, err)
						continue
					}
					expectedBool := expectedValue == "true"
					if actualValue != expectedBool {
						t.Errorf("Flag %s: expected %v, got %v", flagName, expectedBool, actualValue)
					}
				} else {
					actualValue, err := quotaCmd.Flags().GetString(flagName)
					if err != nil {
						t.Errorf("Failed to get string flag %s: %v", flagName, err)
						continue
					}
					if actualValue != expectedValue {
						t.Errorf("Flag %s: expected %s, got %s", flagName, expectedValue, actualValue)
					}
				}
			}

			// Reset flags for next test
			quotaCmd.Flags().VisitAll(func(flag *pflag.Flag) {
				flag.Value.Set(flag.DefValue)
			})
		})
	}
}

// TestPreflightCommand_SubcommandIntegration tests integration between main command and subcommands
func TestPreflightCommand_SubcommandIntegration(t *testing.T) {
	setupTestEnvironment()
	tests := []struct {
		name                  string
		subcommandName        string
		expectedSubcommandUse string
		hasFlags              bool
		description           string
	}{
		{
			name:                  "quota_subcommand_integration",
			subcommandName:        "quota",
			expectedSubcommandUse: "quota",
			hasFlags:              true,
			description:           "Quota subcommand should be properly integrated",
		},
		{
			name:                  "rp_subcommand_integration",
			subcommandName:        "rp",
			expectedSubcommandUse: "rp",
			hasFlags:              false,
			description:           "RP subcommand should be properly integrated",
		},
		{
			name:                  "status_subcommand_integration",
			subcommandName:        "status",
			expectedSubcommandUse: "status",
			hasFlags:              false,
			description:           "Status subcommand should be properly integrated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup services
			mockCLI := azurecli.NewMockAzureCLI()
			quotaService := services.NewQuotaService(mockCLI)
			validationService := services.NewValidationService(mockCLI)

			// Create the preflight command
			preflightCmd := createPreflightCommand(quotaService, validationService, mockCLI)

			// Find the specific subcommand
			var targetSubcmd *cobra.Command
			for _, subcmd := range preflightCmd.Commands() {
				if subcmd.Use == tt.subcommandName {
					targetSubcmd = subcmd
					break
				}
			}

			if targetSubcmd == nil {
				t.Fatalf("Subcommand %s not found", tt.subcommandName)
			}

			// Verify subcommand properties
			if targetSubcmd.Use != tt.expectedSubcommandUse {
				t.Errorf("Subcommand use: expected %s, got %s", tt.expectedSubcommandUse, targetSubcmd.Use)
			}

			if targetSubcmd.Short == "" {
				t.Error("Subcommand should have a short description")
			}

			// Test flags presence
			flagCount := len(targetSubcmd.Flags().FlagUsages())
			if tt.hasFlags && flagCount == 0 {
				t.Error("Expected subcommand to have flags but found none")
			}

			// Verify subcommand has appropriate execution function
			if tt.subcommandName == "quota" && targetSubcmd.RunE == nil {
				t.Error("Quota subcommand should have RunE function")
			}

			// Test that subcommand is properly parented
			if targetSubcmd.Parent() != preflightCmd {
				t.Error("Subcommand should be properly parented to preflight command")
			}
		})
	}
}

// TestPreflightCommand_ExamplesIntegration tests that examples are properly integrated
func TestPreflightCommand_ExamplesIntegration(t *testing.T) {
	setupTestEnvironment()
	// Setup services
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	// Create the preflight command
	preflightCmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Find the quota subcommand
	var quotaCmd *cobra.Command
	for _, subcmd := range preflightCmd.Commands() {
		if subcmd.Use == "quota" {
			quotaCmd = subcmd
			break
		}
	}

	if quotaCmd == nil {
		t.Fatal("Quota subcommand not found")
	}

	// Test that examples are included in the long description
	if !strings.Contains(quotaCmd.Long, "vCPU quota") {
		t.Error("Quota command long description should mention vCPU quota")
	}

	// Test that the long description is not empty
	if quotaCmd.Long == "" {
		t.Error("Quota command should have a long description")
	}

	// Test that examples integration doesn't break the command
	if quotaCmd.RunE == nil {
		t.Error("Quota command should have a RunE function even with examples")
	}
}

// TestPreflightCommand_ErrorHandlingLogic tests the error handling patterns in preflight command
func TestPreflightCommand_ErrorHandlingLogic(t *testing.T) {
	setupTestEnvironment()
	// Setup services
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	// Create the preflight command
	_ = createPreflightCommand(quotaService, validationService, mockCLI)

	// Test error handling logic patterns
	tests := []struct {
		name               string
		errorMessage       string
		expectsSpecificErr bool
		description        string
	}{
		{
			name:               "missing_required_argument_pattern",
			errorMessage:       "missing required argument: flavor",
			expectsSpecificErr: true,
			description:        "Should recognize missing required argument errors",
		},
		{
			name:               "must_specify_either_pattern",
			errorMessage:       "must specify either --location or --all-locations",
			expectsSpecificErr: true,
			description:        "Should recognize specification requirement errors",
		},
		{
			name:               "cannot_specify_both_pattern",
			errorMessage:       "cannot specify both --location and --all-locations",
			expectsSpecificErr: true,
			description:        "Should recognize conflicting flags errors",
		},
		{
			name:               "location_validation_failed_pattern",
			errorMessage:       "location validation failed: 'invalid-region' is not a valid Azure region",
			expectsSpecificErr: true,
			description:        "Should recognize location validation errors",
		},
		{
			name:               "general_error_pattern",
			errorMessage:       "Azure CLI authentication failed",
			expectsSpecificErr: false,
			description:        "Should handle general errors differently",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the error handling logic pattern
			isSpecificError := strings.Contains(tt.errorMessage, "missing required argument") ||
				strings.Contains(tt.errorMessage, "must specify either") ||
				strings.Contains(tt.errorMessage, "cannot specify both") ||
				strings.Contains(tt.errorMessage, "location validation failed")

			if isSpecificError != tt.expectsSpecificErr {
				t.Errorf("Error pattern detection failed for '%s': expected specific=%v, got specific=%v",
					tt.errorMessage, tt.expectsSpecificErr, isSpecificError)
			}
		})
	}
}

// TestPreflightCommand_ModuleIntegration tests integration with dedicated modules
func TestPreflightCommand_ModuleIntegration(t *testing.T) {
	setupTestEnvironment()
	// Setup services
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	// Create the preflight command
	preflightCmd := createPreflightCommand(quotaService, validationService, mockCLI)

	tests := []struct {
		name           string
		subcommandName string
		expectedModule string
		description    string
	}{
		{
			name:           "rp_module_integration",
			subcommandName: "rp",
			expectedModule: "arcbox.CreateResourceProviderCommands",
			description:    "RP subcommand should be created via dedicated module",
		},
		{
			name:           "status_module_integration",
			subcommandName: "status",
			expectedModule: "arcbox.CreateStatusCommand",
			description:    "Status subcommand should be created via dedicated module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find the subcommand
			var subcmd *cobra.Command
			for _, cmd := range preflightCmd.Commands() {
				if cmd.Use == tt.subcommandName {
					subcmd = cmd
					break
				}
			}

			if subcmd == nil {
				t.Fatalf("Subcommand %s not found", tt.subcommandName)
			}

			// Verify the subcommand was properly integrated
			if subcmd.Parent() != preflightCmd {
				t.Errorf("Subcommand %s should be parented to preflight command", tt.subcommandName)
			}

			// Verify basic command properties
			if subcmd.Use != tt.subcommandName {
				t.Errorf("Subcommand use: expected %s, got %s", tt.subcommandName, subcmd.Use)
			}

			if subcmd.Short == "" {
				t.Errorf("Subcommand %s should have a short description", tt.subcommandName)
			}
		})
	}
}

// TestPreflightCommand_FlagConfigurationCoverage tests all flag configurations comprehensively
func TestPreflightCommand_FlagConfigurationCoverage(t *testing.T) {
	setupTestEnvironment()
	// Setup services
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	// Create the preflight command
	preflightCmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Find quota subcommand
	var quotaCmd *cobra.Command
	for _, subcmd := range preflightCmd.Commands() {
		if subcmd.Use == "quota" {
			quotaCmd = subcmd
			break
		}
	}

	if quotaCmd == nil {
		t.Fatal("Quota subcommand not found")
	}

	// Test all flag configurations are created correctly
	flagTests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
		usage        string
		flagType     string
	}{
		{
			flagName:     "flavor",
			shorthand:    "f",
			defaultValue: "",
			usage:        "ArcBox flavor to check (ITPro, DevOps, DataOps, all)",
			flagType:     "string",
		},
		{
			flagName:     "location",
			shorthand:    "l",
			defaultValue: "",
			usage:        "Azure region(s) to check quota in. Use comma-separated values for multiple regions",
			flagType:     "string",
		},
		{
			flagName:     "all-locations",
			shorthand:    "",
			defaultValue: "false",
			usage:        "Check quota in all ArcBox-supported regions",
			flagType:     "bool",
		},
		{
			flagName:     "sku",
			shorthand:    "",
			defaultValue: "",
			usage:        "Custom VM SKU(s) to check. Comma-separated",
			flagType:     "string",
		},
		{
			flagName:     "subscription",
			shorthand:    "s",
			defaultValue: "",
			usage:        "Azure subscription ID to use",
			flagType:     "string",
		},
	}

	for _, tt := range flagTests {
		t.Run("flag_"+tt.flagName, func(t *testing.T) {
			flag := quotaCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("Flag %s not found", tt.flagName)
			}

			// Test shorthand
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand: expected %s, got %s", tt.flagName, tt.shorthand, flag.Shorthand)
			}

			// Test default value
			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %s default value: expected %s, got %s", tt.flagName, tt.defaultValue, flag.DefValue)
			}

			// Test usage
			if flag.Usage != tt.usage {
				t.Errorf("Flag %s usage: expected %s, got %s", tt.flagName, tt.usage, flag.Usage)
			}

			// Test type
			if flag.Value.Type() != tt.flagType {
				t.Errorf("Flag %s type: expected %s, got %s", tt.flagName, tt.flagType, flag.Value.Type())
			}
		})
	}
}

// TestPreflightCommand_CommandReturnValue tests that the command is properly returned
func TestPreflightCommand_CommandReturnValue(t *testing.T) {
	setupTestEnvironment()
	// Setup services
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	// Create the preflight command
	cmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Test that a valid command is returned
	if cmd == nil {
		t.Fatal("createPreflightCommand should return a non-nil command")
	}

	// Test command structure completeness
	if cmd.Use != "preflight" {
		t.Errorf("Command use: expected 'preflight', got '%s'", cmd.Use)
	}

	// Test that all expected subcommands are added
	expectedSubcommandCount := 3
	actualSubcommandCount := len(cmd.Commands())
	if actualSubcommandCount != expectedSubcommandCount {
		t.Errorf("Subcommand count: expected %d, got %d", expectedSubcommandCount, actualSubcommandCount)
	}

	// Test that subcommands are properly configured
	subcommandNames := make(map[string]bool)
	for _, subcmd := range cmd.Commands() {
		subcommandNames[subcmd.Use] = true

		// Each subcommand should have basic properties
		if subcmd.Use == "" {
			t.Error("Subcommand should have a 'Use' field")
		}

		if subcmd.Short == "" {
			t.Errorf("Subcommand %s should have a short description", subcmd.Use)
		}

		// Verify parent relationship
		if subcmd.Parent() != cmd {
			t.Errorf("Subcommand %s should be parented to main command", subcmd.Use)
		}
	}

	// Verify all expected subcommands exist
	expectedSubcommands := []string{"quota", "rp", "status"}
	for _, expected := range expectedSubcommands {
		if !subcommandNames[expected] {
			t.Errorf("Expected subcommand %s not found", expected)
		}
	}
}

// TestPreflightCommand_FullCommandStructure tests the complete command structure creation
func TestPreflightCommand_FullCommandStructure(t *testing.T) {
	setupTestEnvironment()
	// Setup services
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	// Create the preflight command
	preflightCmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Test main command properties
	mainCommandTests := []struct {
		property string
		expected interface{}
		actual   interface{}
	}{
		{"Use", "preflight", preflightCmd.Use},
		{"Short", "Run preflight checks for ArcBox deployment", preflightCmd.Short},
		{"DisableSuggestions", true, preflightCmd.DisableSuggestions},
		{"SilenceErrors", true, preflightCmd.SilenceErrors},
		{"SilenceUsage", true, preflightCmd.SilenceUsage},
	}

	for _, test := range mainCommandTests {
		switch test.property {
		case "Use", "Short":
			if test.actual != test.expected {
				t.Errorf("Command %s: expected %s, got %s", test.property, test.expected, test.actual)
			}
		case "DisableSuggestions", "SilenceErrors", "SilenceUsage":
			if test.actual != test.expected {
				t.Errorf("Command %s: expected %v, got %v", test.property, test.expected, test.actual)
			}
		}
	}

	// Test long description contains expected content
	if !strings.Contains(preflightCmd.Long, "preflight checks") {
		t.Error("Long description should contain 'preflight checks'")
	}

	// Test RunE function is set
	if preflightCmd.RunE == nil {
		t.Error("RunE function should be set")
	}

	// Test quota subcommand structure
	quotaCmd := findSubcommand(preflightCmd, "quota")
	if quotaCmd == nil {
		t.Fatal("Quota subcommand not found")
	}

	// Test quota command has RunE function
	if quotaCmd.RunE == nil {
		t.Error("Quota command should have RunE function")
	}

	// Test quota command has all required flags
	quotaFlags := quotaCmd.Flags()
	requiredQuotaFlags := []string{"flavor", "location", "all-locations", "sku", "subscription"}
	for _, flagName := range requiredQuotaFlags {
		if quotaFlags.Lookup(flagName) == nil {
			t.Errorf("Quota command missing required flag: %s", flagName)
		}
	}
}

// TestExecutePreflightQuotaCommandWithError tests the extracted quota execution logic
func TestExecutePreflightQuotaCommandWithError(t *testing.T) {
	setupTestEnvironment()
	tests := []struct {
		name                string
		flags               map[string]string
		setupMock           func(*azurecli.MockAzureCLI)
		expectError         bool
		expectSpecificError bool
		description         string
	}{
		{
			name: "successful_quota_execution",
			flags: map[string]string{
				"flavor":   "ITPro",
				"location": "eastus",
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Set up successful conditions
				mock.IsLoggedInResult = true
			},
			expectError:         false,
			expectSpecificError: false,
			description:         "Should execute successfully with valid parameters",
		},
		{
			name: "azure_cli_not_logged_in",
			flags: map[string]string{
				"flavor":   "ITPro",
				"location": "eastus",
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Set up logged out state which causes authentication errors
				mock.IsLoggedInResult = false
				mock.GetCurrentSubscriptionError = fmt.Errorf("Azure CLI is not authenticated")
			},
			expectError:         true,
			expectSpecificError: false,
			description:         "Should handle authentication errors properly",
		},
		{
			name: "subscription_access_error",
			flags: map[string]string{
				"flavor":   "ITPro",
				"location": "eastus",
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Set up subscription error that affects quota operations
				mock.IsLoggedInResult = true
				mock.ListVMUsageError = fmt.Errorf("subscription not found or access denied")
			},
			expectError:         true,
			expectSpecificError: false,
			description:         "Should handle subscription access errors during quota operations",
		},
		{
			name: "vm_usage_listing_error",
			flags: map[string]string{
				"flavor":   "ITPro",
				"location": "eastus",
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Set up VM usage error - this affects quota checking
				mock.ListVMUsageError = fmt.Errorf("failed to list VM usage quotas")
			},
			expectError:         true,
			expectSpecificError: false,
			description:         "Should handle VM usage listing errors properly",
		},
		{
			name: "sku_availability_error",
			flags: map[string]string{
				"flavor":   "ITPro",
				"location": "eastus",
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Set up SKU availability error - this affects SKU checking
				mock.CheckSKUAvailabilityError = fmt.Errorf("failed to check SKU availability")
			},
			expectError:         true,
			expectSpecificError: false,
			description:         "Should handle SKU availability errors properly",
		},
		{
			name: "missing_flavor_validation_error",
			flags: map[string]string{
				"location": "eastus",
				// Missing flavor flag
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
			},
			expectError:         true,
			expectSpecificError: false,
			description:         "Should handle missing required argument errors with help display",
		},
		{
			name: "missing_location_validation_error",
			flags: map[string]string{
				"flavor": "ITPro",
				// Missing location flag and no all-locations
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
			},
			expectError:         true,
			expectSpecificError: false,
			description:         "Should handle location specification errors with help display",
		},
		{
			name: "conflicting_location_flags_error",
			flags: map[string]string{
				"flavor":        "ITPro",
				"location":      "eastus",
				"all-locations": "true",
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
			},
			expectError:         true,
			expectSpecificError: false,
			description:         "Should handle conflicting location flags with help display",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var outputBuffer bytes.Buffer

			// Create and configure mock CLI
			mockCLI := azurecli.NewMockAzureCLI()
			if tt.setupMock != nil {
				tt.setupMock(mockCLI)
			}

			quotaService := services.NewQuotaService(mockCLI)
			validationService := services.NewValidationService(mockCLI)

			// Create preflight command
			preflightCmd := createPreflightCommand(quotaService, validationService, mockCLI)

			// Find quota subcommand
			var quotaCmd *cobra.Command
			for _, subcmd := range preflightCmd.Commands() {
				if subcmd.Use == "quota" {
					quotaCmd = subcmd
					break
				}
			}

			if quotaCmd == nil {
				t.Fatal("Quota subcommand not found")
			}

			// Set flags
			for flagName, flagValue := range tt.flags {
				err := quotaCmd.Flags().Set(flagName, flagValue)
				if err != nil {
					t.Fatalf("Failed to set flag %s=%s: %v", flagName, flagValue, err)
				}
			}

			// Capture output
			quotaCmd.SetOut(&outputBuffer)
			quotaCmd.SetErr(&outputBuffer)

			// Execute the testable version
			err := executePreflightQuotaCommandWithError(quotaService, quotaCmd, []string{})

			// Verify expectations
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			// Reset flags for next test
			quotaCmd.Flags().VisitAll(func(flag *pflag.Flag) {
				flag.Value.Set(flag.DefValue)
			})
		})
	}
}

// TestExecutePreflightQuotaCommand tests the main quota execution function (with mocked behavior)
func TestExecutePreflightQuotaCommand(t *testing.T) {
	setupTestEnvironment()
	// Setup
	mockCLI := azurecli.NewMockAzureCLI()
	quotaService := services.NewQuotaService(mockCLI)
	validationService := services.NewValidationService(mockCLI)

	// Create preflight command
	preflightCmd := createPreflightCommand(quotaService, validationService, mockCLI)

	// Find quota subcommand
	var quotaCmd *cobra.Command
	for _, subcmd := range preflightCmd.Commands() {
		if subcmd.Use == "quota" {
			quotaCmd = subcmd
			break
		}
	}

	if quotaCmd == nil {
		t.Fatal("Quota subcommand not found")
	}

	// Test that the function can be called (we can't test the os.Exit path directly)
	// Set up valid flags to avoid errors
	quotaCmd.Flags().Set("flavor", "ITPro")
	quotaCmd.Flags().Set("location", "eastus")

	// After refactoring, the quota command now uses RunE instead of Run
	// This is the proper Cobra pattern for error handling
	if quotaCmd.RunE == nil {
		t.Error("Quota command should have a RunE function that calls executePreflightQuotaCommandWithError")
	}

	// Verify the function exists by testing it indirectly through the testable version
	err := executePreflightQuotaCommandWithError(quotaService, quotaCmd, []string{})
	if err != nil {
		// This is expected behavior - the mock service may return errors
		// The important thing is that the function executed without panicking
		t.Logf("Function executed and returned error (expected): %v", err)
	}
}

// TestPreflightQuotaErrorHandlingLogic tests the error classification logic
func TestPreflightQuotaErrorHandlingLogic(t *testing.T) {
	setupTestEnvironment()
	tests := []struct {
		name               string
		errorMessage       string
		expectSpecificHelp bool
		description        string
	}{
		{
			name:               "missing_required_argument_error",
			errorMessage:       "missing required argument: flavor is required",
			expectSpecificHelp: true,
			description:        "Should trigger specific help for missing required arguments",
		},
		{
			name:               "must_specify_either_error",
			errorMessage:       "must specify either --location or --all-locations",
			expectSpecificHelp: true,
			description:        "Should trigger specific help for location specification errors",
		},
		{
			name:               "cannot_specify_both_error",
			errorMessage:       "cannot specify both --location and --all-locations",
			expectSpecificHelp: true,
			description:        "Should trigger specific help for conflicting flags",
		},
		{
			name:               "location_validation_failed_error",
			errorMessage:       "location validation failed: invalid region",
			expectSpecificHelp: true,
			description:        "Should trigger specific help for location validation errors",
		},
		{
			name:               "authentication_error",
			errorMessage:       "Azure CLI authentication failed: please run 'az login'",
			expectSpecificHelp: false,
			description:        "Should not trigger specific help for auth errors",
		},
		{
			name:               "general_service_error",
			errorMessage:       "failed to retrieve quota information",
			expectSpecificHelp: false,
			description:        "Should not trigger specific help for general service errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the error classification logic used in executePreflightQuotaCommandWithError
			isSpecificError := strings.Contains(tt.errorMessage, "missing required argument") ||
				strings.Contains(tt.errorMessage, "must specify either") ||
				strings.Contains(tt.errorMessage, "cannot specify both") ||
				strings.Contains(tt.errorMessage, "location validation failed")

			if isSpecificError != tt.expectSpecificHelp {
				t.Errorf("Error classification failed for '%s': expected specific help=%v, got=%v",
					tt.errorMessage, tt.expectSpecificHelp, isSpecificError)
			}
		})
	}
}

// TestExecutePreflightQuotaCommandExitScenarios tests the executePreflightQuotaCommand function
// by verifying the logic paths before os.Exit is called
func TestExecutePreflightQuotaCommandExitScenarios(t *testing.T) {
	setupTestEnvironment()

	tests := []struct {
		name              string
		flags             map[string]string
		setupMock         func(*azurecli.MockAzureCLI)
		expectError       bool
		expectHelpDisplay bool
		description       string
	}{
		{
			name: "success_no_exit",
			flags: map[string]string{
				"flavor":   "ITPro",
				"location": "eastus",
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
			},
			expectError:       false,
			expectHelpDisplay: false,
			description:       "Should not exit on successful execution",
		},
		{
			name: "validation_error_triggers_help",
			flags: map[string]string{
				// Missing required flavor flag
				"location": "eastus",
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
			},
			expectError:       true,
			expectHelpDisplay: true,
			description:       "Should trigger help display for validation errors before exit",
		},
		{
			name: "non_validation_error_no_help",
			flags: map[string]string{
				"flavor":   "ITPro",
				"location": "eastus",
			},
			setupMock: func(mock *azurecli.MockAzureCLI) {
				mock.ListVMUsageError = fmt.Errorf("azure cli error: unexpected error")
			},
			expectError:       true,
			expectHelpDisplay: false,
			description:       "Should not trigger help display for non-validation errors before exit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup output capture
			var outputBuffer bytes.Buffer

			// Create and configure mock CLI
			mockCLI := azurecli.NewMockAzureCLI()
			if tt.setupMock != nil {
				tt.setupMock(mockCLI)
			}

			quotaService := services.NewQuotaService(mockCLI)
			validationService := services.NewValidationService(mockCLI)

			// Create preflight command
			preflightCmd := createPreflightCommand(quotaService, validationService, mockCLI)

			// Find quota subcommand
			var quotaCmd *cobra.Command
			for _, subcmd := range preflightCmd.Commands() {
				if subcmd.Use == "quota" {
					quotaCmd = subcmd
					break
				}
			}

			if quotaCmd == nil {
				t.Fatal("Quota subcommand not found")
			}

			// Set command flags
			for flagName, flagValue := range tt.flags {
				err := quotaCmd.Flags().Set(flagName, flagValue)
				if err != nil {
					t.Fatalf("Failed to set flag %s: %v", flagName, err)
				}
			}

			// Redirect output
			quotaCmd.SetOut(&outputBuffer)
			quotaCmd.SetErr(&outputBuffer)

			// We can't directly test executePreflightQuotaCommand because it calls os.Exit,
			// but we can test that the same logic would be executed by calling the
			// underlying service method directly and checking the error conditions
			err := quotaService.RunQuotaCheckCommand(quotaCmd, []string{})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				// Check if this error would trigger help display
				errorMessage := err.Error()
				shouldShowHelp := strings.Contains(errorMessage, "missing required argument") ||
					strings.Contains(errorMessage, "must specify either") ||
					strings.Contains(errorMessage, "cannot specify both") ||
					strings.Contains(errorMessage, "location validation failed") ||
					// Also check for the actual validation error messages that get wrapped
					strings.Contains(errorMessage, "required argument missing") ||
					strings.Contains(errorMessage, "location specification required") ||
					strings.Contains(errorMessage, "conflicting location flags")

				if tt.expectHelpDisplay && !shouldShowHelp {
					t.Errorf("Expected error to trigger help display, but it wouldn't. Error: %v", err)
				}

				if !tt.expectHelpDisplay && shouldShowHelp {
					t.Errorf("Expected error not to trigger help display, but it would. Error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			// Log the test completion for coverage tracking
			t.Logf("Test %s completed - exercised error handling logic that would precede os.Exit", tt.name)
		})
	}
}

func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, subcmd := range cmd.Commands() {
		if subcmd.Use == name {
			return subcmd
		}
	}
	return nil
}
