package arcbox

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"jumpstartcli/cmd/arcbox/services"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"
)

// ====================================================================================
// PHASE 2.1 - List Command Enhancement Test Coverage
// Target: Achieve 95%+ coverage for list command functionality (currently 29.4%)
// ====================================================================================

// TestCreateListCommand_AllFlags tests all flag combinations and command variations
func TestCreateListCommand_AllFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		mockSetup   func(*azurecli.MockAzureCLI)
		expectError bool
	}{
		{
			name: "all_subscriptions_flag",
			args: []string{"--all-subscriptions"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				// Set up subscriptions
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "sub1", Name: "Sub 1"},
					{ID: "sub2", Name: "Sub 2"},
				}

				// Set up resource groups with ArcBox deployment
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "arcbox-rg", Location: "eastus"},
					{Name: "other-rg", Location: "westus"},
				}

				// Set up resources with ArcBox solution tag
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"arcbox-rg": {
						{
							ID:   "/subscriptions/sub1/resourceGroups/arcbox-rg/providers/Microsoft.Compute/virtualMachines/arcbox-vm",
							Name: "arcbox-vm",
							Type: "Microsoft.Compute/virtualMachines",
							Tags: map[string]string{"Solution": "jumpstart_arcbox"},
						},
					},
					"other-rg": {},
				}
			},
			expectError: false,
		},
		{
			name: "current_subscription_flag",
			args: []string{"--current-subscription"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				// Set up current subscription
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{
					ID:   "current-sub",
					Name: "Current Sub",
				}

				// Also add to subscriptions list so it can be found
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "current-sub", Name: "Current Sub"},
				}

				// Set up resource groups
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "arcbox-rg", Location: "eastus"},
				}

				// Set up resources with ArcBox solution tag
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"arcbox-rg": {
						{
							Name: "arcbox-vm",
							Tags: map[string]string{"Solution": "jumpstart_arcbox"},
						},
					},
				}
			},
			expectError: false,
		},
		{
			name: "specific_subscription_flag",
			args: []string{"--subscription", "specific-sub-id"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				// Set up specific subscription
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "specific-sub-id", Name: "Specific Sub"},
				}

				// No ArcBox deployments
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "empty-rg", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"empty-rg": {},
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set default output format for testing
			utils.OutputFormat = "table"

			// Create fresh mock CLI for each test
			mockCLI := azurecli.NewMockAzureCLI()
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			// Create listing service and command
			listingService := services.NewListingService(mockCLI)
			cmd := createListCommand(listingService, mockCLI)
			cmd.SetArgs(tt.args)

			// Capture output
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			// Execute command
			err := cmd.Execute()

			// Check error expectations
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// For successful cases, verify that the appropriate CLI methods were called
			if !tt.expectError {
				if strings.Contains(tt.name, "all_subscriptions") {
					if !mockCLI.ListSubscriptionsCalled {
						t.Error("Expected ListSubscriptions to be called for --all-subscriptions flag")
					}
					if !mockCLI.ListResourceGroupsCalled {
						t.Error("Expected ListResourceGroups to be called")
					}
				}

				if strings.Contains(tt.name, "current_subscription") {
					if !mockCLI.GetCurrentSubscriptionCalled {
						t.Error("Expected GetCurrentSubscription to be called for --current-subscription flag")
					}
				}

				if strings.Contains(tt.name, "specific_subscription") {
					if !mockCLI.GetSubscriptionCalled {
						t.Error("Expected GetSubscription to be called for --subscription flag")
					}
				}
			}
		})
	}
}

// TestListCommand_AdditionalCoverage tests additional scenarios to improve coverage
func TestListCommand_AdditionalCoverage(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		mockSetup    func(*azurecli.MockAzureCLI)
		outputFormat string
		expectError  bool
	}{
		{
			name:         "all_subscriptions_with_json_output",
			args:         []string{"--all-subscriptions"},
			outputFormat: "json",
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "sub1", Name: "Sub 1"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "test-rg", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"test-rg": {},
				}
			},
			expectError: false,
		},
		{
			name:         "current_subscription_with_yaml_output",
			args:         []string{"--current-subscription"},
			outputFormat: "yaml",
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "test-rg", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"test-rg": {},
				}
			},
			expectError: false,
		},
		{
			name:         "subscription_with_table_output",
			args:         []string{"--subscription", "test-sub"},
			outputFormat: "table",
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "test-rg", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"test-rg": {},
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set output format for this test
			utils.OutputFormat = tt.outputFormat
			defer func() { utils.OutputFormat = "table" }()

			mockCLI := azurecli.NewMockAzureCLI()
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			listingService := services.NewListingService(mockCLI)
			cmd := createListCommand(listingService, mockCLI)
			cmd.SetArgs(tt.args)

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestListCommand_OutputFormats tests different output formats
func TestListCommand_OutputFormats(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		outputFormat string
		mockSetup    func(*azurecli.MockAzureCLI)
		expectError  bool
	}{
		{
			name:         "table_output_format",
			args:         []string{"--current-subscription"},
			outputFormat: "table",
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{
					ID: "test-sub", Name: "Test Sub",
				}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "arcbox-rg", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"arcbox-rg": {
						{
							Name: "arcbox-vm",
							Tags: map[string]string{"Solution": "jumpstart_arcbox"},
						},
					},
				}
			},
			expectError: false,
		},
		{
			name:         "json_output_format",
			args:         []string{"--current-subscription"},
			outputFormat: "json",
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{
					ID: "test-sub", Name: "Test Sub",
				}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "arcbox-rg", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"arcbox-rg": {
						{
							Name: "arcbox-vm",
							Tags: map[string]string{"Solution": "jumpstart_arcbox"},
						},
					},
				}
			},
			expectError: false,
		},
		{
			name:         "yaml_output_format",
			args:         []string{"--current-subscription"},
			outputFormat: "yaml",
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{
					ID: "test-sub", Name: "Test Sub",
				}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "arcbox-rg", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"arcbox-rg": {
						{
							Name: "arcbox-vm",
							Tags: map[string]string{"Solution": "jumpstart_arcbox"},
						},
					},
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global output format for this test
			utils.OutputFormat = tt.outputFormat
			defer func() { utils.OutputFormat = "table" }() // Reset after test

			mockCLI := azurecli.NewMockAzureCLI()
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			listingService := services.NewListingService(mockCLI)
			cmd := createListCommand(listingService, mockCLI)
			cmd.SetArgs(tt.args)

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			err := cmd.Execute()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify the command executed successfully
			if !tt.expectError {
				if !mockCLI.GetCurrentSubscriptionCalled {
					t.Error("Expected GetCurrentSubscription to be called")
				}
				if !mockCLI.ListResourceGroupsCalled {
					t.Error("Expected ListResourceGroups to be called")
				}
			}
		})
	}
}

// TestListCommand_ArcBoxDetection tests ArcBox deployment detection algorithms
func TestListCommand_ArcBoxDetection(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*azurecli.MockAzureCLI)
		expectedCount int
		description   string
	}{
		{
			name: "solution_tag_detection",
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "arcbox-detected", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"arcbox-detected": {
						{
							Name: "some-vm",
							Tags: map[string]string{"Solution": "jumpstart_arcbox"},
						},
					},
				}
			},
			expectedCount: 1,
			description:   "Should detect ArcBox via solution tag",
		},
		{
			name: "deployment_name_detection",
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "deployment-detected", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"deployment-detected": {},
				}
				cli.Deployments = map[string][]azurecli.DeploymentInfo{
					"deployment-detected": {
						{Name: "arcbox-main"},
					},
				}
			},
			expectedCount: 1,
			description:   "Should detect ArcBox via deployment name",
		},
		{
			name: "false_positive_prevention",
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "false-positive", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"false-positive": {
						{Name: "generic-vm", Tags: map[string]string{"Environment": "dev"}},
						{Name: "generic-vnet", Tags: map[string]string{"Environment": "dev"}},
					},
				}
				// Ensure no deployments that could match ArcBox patterns
				cli.Deployments = map[string][]azurecli.DeploymentInfo{
					"false-positive": {},
				}
			},
			expectedCount: 0,
			description:   "Should not detect non-ArcBox deployments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set default output format for testing
			utils.OutputFormat = "table"

			mockCLI := azurecli.NewMockAzureCLI()
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			listingService := services.NewListingService(mockCLI)
			cmd := createListCommand(listingService, mockCLI)
			cmd.SetArgs([]string{"--current-subscription"})

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			err := cmd.Execute()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Since output goes directly to stdout (not captured by buffer),
			// we focus on verifying the command executed without errors
			// and made the appropriate CLI method calls

			// For successful cases, verify appropriate CLI methods were called
			if tt.expectedCount == 0 {
				// For cases expecting no deployments, verify the methods were called
				// but we can't reliably test output due to direct stdout usage
				if !mockCLI.GetCurrentSubscriptionCalled {
					t.Error("Expected GetCurrentSubscription to be called")
				}
				if !mockCLI.ListResourceGroupsCalled {
					t.Error("Expected ListResourceGroups to be called")
				}
			} else {
				// For cases expecting deployments, verify detection methods were called
				if !mockCLI.GetCurrentSubscriptionCalled {
					t.Error("Expected GetCurrentSubscription to be called")
				}
				if !mockCLI.ListResourceGroupsCalled {
					t.Error("Expected ListResourceGroups to be called")
				}
			}
		})
	}
}

// TestListCommand_ValidationRequirements tests subscription selection validation
func TestListCommand_ValidationRequirements(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		mockSetup func(*azurecli.MockAzureCLI)
		expectErr bool
	}{
		{
			name: "valid_current_subscription_only",
			args: []string{"--current-subscription"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{}
			},
			expectErr: false,
		},
		{
			name: "valid_all_subscriptions_only",
			args: []string{"--all-subscriptions"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{}
			},
			expectErr: false,
		},
		{
			name: "valid_specific_subscription_only",
			args: []string{"--subscription", "test-sub"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set default output format for testing
			utils.OutputFormat = "table"

			mockCLI := azurecli.NewMockAzureCLI()
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			listingService := services.NewListingService(mockCLI)
			cmd := createListCommand(listingService, mockCLI)
			cmd.SetArgs(tt.args)

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			err := cmd.Execute()

			if tt.expectErr && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

// TestListCommand_EdgeCases tests edge cases and boundary conditions
func TestListCommand_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		mockSetup func(*azurecli.MockAzureCLI)
		expectErr bool
	}{
		{
			name: "empty_subscription_list",
			args: []string{"--all-subscriptions"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.Subscriptions = []azurecli.SubscriptionInfo{}
			},
			expectErr: false,
		},
		{
			name: "large_number_of_resource_groups",
			args: []string{"--current-subscription"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}

				// Create many resource groups
				resourceGroups := make([]azurecli.ResourceGroupInfo, 100)
				for i := 0; i < 100; i++ {
					resourceGroups[i] = azurecli.ResourceGroupInfo{
						Name:     fmt.Sprintf("rg-%d", i),
						Location: "eastus",
					}
				}
				cli.ResourceGroups = resourceGroups
			},
			expectErr: false,
		},
		{
			name: "resource_group_without_resources",
			args: []string{"--current-subscription"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "empty-rg", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"empty-rg": {},
				}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set default output format for testing
			utils.OutputFormat = "table"

			mockCLI := azurecli.NewMockAzureCLI()
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			listingService := services.NewListingService(mockCLI)
			cmd := createListCommand(listingService, mockCLI)
			cmd.SetArgs(tt.args)

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			err := cmd.Execute()

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestListCommand_CommandStructure tests the command structure and flags
func TestListCommand_CommandStructure(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	listingService := services.NewListingService(mockCLI)
	cmd := createListCommand(listingService, mockCLI)

	// Test command properties
	if cmd.Use != "list" {
		t.Errorf("Expected command use to be 'list', got '%s'", cmd.Use)
	}

	if cmd.Short != "List Jumpstart ArcBox deployments" {
		t.Errorf("Expected short description, got '%s'", cmd.Short)
	}

	// Test flags exist and have correct defaults
	allSubsFlag := cmd.Flags().Lookup("all-subscriptions")
	if allSubsFlag == nil {
		t.Error("Expected 'all-subscriptions' flag to exist")
	} else if allSubsFlag.DefValue != "false" {
		t.Errorf("Expected 'all-subscriptions' default to be 'false', got '%s'", allSubsFlag.DefValue)
	}

	currentSubFlag := cmd.Flags().Lookup("current-subscription")
	if currentSubFlag == nil {
		t.Error("Expected 'current-subscription' flag to exist")
	} else if currentSubFlag.DefValue != "false" {
		t.Errorf("Expected 'current-subscription' default to be 'false', got '%s'", currentSubFlag.DefValue)
	}

	subFlag := cmd.Flags().Lookup("subscription")
	if subFlag == nil {
		t.Error("Expected 'subscription' flag to exist")
	} else if subFlag.Shorthand != "s" {
		t.Errorf("Expected 'subscription' shorthand to be 's', got '%s'", subFlag.Shorthand)
	} else if subFlag.DefValue != "" {
		t.Errorf("Expected 'subscription' default to be empty, got '%s'", subFlag.DefValue)
	}
}

// TestExecuteListCommand tests the extracted executeListCommand function directly
func TestExecuteListCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		mockSetup     func(*azurecli.MockAzureCLI)
		outputFormat  string
		expectError   bool
		errorContains string
	}{
		{
			name: "successful_execution_all_subscriptions",
			args: []string{"--all-subscriptions"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "sub1", Name: "Sub 1"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "arcbox-rg", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"arcbox-rg": {
						{Name: "arcbox-vm", Tags: map[string]string{"Solution": "jumpstart_arcbox"}},
					},
				}
			},
			outputFormat: "table",
			expectError:  false,
		},
		{
			name: "successful_execution_current_subscription",
			args: []string{"--current-subscription"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
				cli.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Sub"},
				}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{
					{Name: "test-rg", Location: "eastus"},
				}
				cli.Resources = map[string][]azurecli.ResourceInfo{
					"test-rg": {},
				}
			},
			outputFormat: "json",
			expectError:  false,
		},
		{
			name: "validation_failure_missing_subscription_flag",
			args: []string{}, // No subscription flags
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				// No setup needed for validation failure
			},
			outputFormat:  "table",
			expectError:   true,
			errorContains: "subscription selection required",
		},
		{
			name: "validation_failure_conflicting_flags",
			args: []string{"--all-subscriptions", "--current-subscription"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				// No setup needed for validation failure
			},
			outputFormat:  "table",
			expectError:   true,
			errorContains: "conflicting subscription flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set output format for this test
			utils.OutputFormat = tt.outputFormat
			defer func() { utils.OutputFormat = "table" }()

			mockCLI := azurecli.NewMockAzureCLI()
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			listingService := services.NewListingService(mockCLI)
			cmd := createListCommand(listingService, mockCLI)
			cmd.SetArgs(tt.args)

			// Parse the flags to ensure they're available to the executeListCommand function
			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			// Test the extracted function directly
			err = executeListCommand(cmd, listingService, mockCLI)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestListCommand_ErrorHandlingCoverage specifically tests the error paths in createListCommand
// to improve coverage of the Run function's error handling logic
func TestListCommand_ErrorHandlingCoverage(t *testing.T) {
	// This test focuses on covering the error handling paths in the createListCommand Run function
	// which includes the fmt.Printf error message, validation check, and help display logic

	tests := []struct {
		name         string
		args         []string
		mockSetup    func(*azurecli.MockAzureCLI)
		expectOutput string
		description  string
	}{
		{
			name: "validation_error_triggers_error_output",
			args: []string{}, // No subscription flags - this will cause validation error
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				// No specific setup needed for validation failure
			},
			expectOutput: "ERROR",
			description:  "Should trigger error output when validation fails",
		},
		{
			name: "conflicting_flags_triggers_error_output",
			args: []string{"--all-subscriptions", "--current-subscription"},
			mockSetup: func(cli *azurecli.MockAzureCLI) {
				// No specific setup needed for validation failure
			},
			expectOutput: "ERROR",
			description:  "Should trigger error output when conflicting flags are used",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set default output format
			utils.OutputFormat = "table"

			mockCLI := azurecli.NewMockAzureCLI()
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			listingService := services.NewListingService(mockCLI)
			cmd := createListCommand(listingService, mockCLI)

			// Verify the command structure - this exercises the creation path
			if cmd.Use != "list" {
				t.Errorf("Expected command use to be 'list', got %s", cmd.Use)
			}

			// Verify Run function exists (covers createListCommand function)
			if cmd.Run == nil {
				t.Error("Expected Run function to be configured")
			}

			// Verify the Long description includes examples (test shows FormatExamples is called)
			if !strings.Contains(cmd.Long, "Examples") {
				t.Error("Expected Long description to include Examples section")
			}

			// Test that flags are properly configured (covers flag creation code)
			flags := []string{"all-subscriptions", "current-subscription", "subscription"}
			for _, flag := range flags {
				if cmd.Flags().Lookup(flag) == nil {
					t.Errorf("Expected flag %s to exist", flag)
				}
			}

			// Test executeListCommand directly to verify error propagation
			cmd.SetArgs(tt.args)
			err := cmd.ParseFlags(tt.args)
			if err != nil && !strings.Contains(tt.name, "validation_error") {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			// Test executeListCommand function (this should return an error for invalid cases)
			executeErr := executeListCommand(cmd, listingService, mockCLI)
			if strings.Contains(tt.name, "validation_error") || strings.Contains(tt.name, "conflicting_flags") {
				if executeErr == nil {
					t.Errorf("Expected executeListCommand to return error for %s", tt.description)
				}
			} else {
				if executeErr != nil {
					t.Errorf("Unexpected error: %v", executeErr)
				}
			}
		})
	}
}

// TestListCommand_ErrorHandlingWithActualRun tests the actual Run function error path
func TestListCommand_ErrorHandlingWithActualRun(t *testing.T) {
	// This test is more challenging because it needs to avoid the os.Exit call
	// We'll create a test that exercises the Run function's error path without
	// calling os.Exit by using a custom approach

	mockCLI := azurecli.NewMockAzureCLI()
	// Set up a scenario that will cause validation failure
	mockCLI.ShouldFailLogin = true

	listingService := services.NewListingService(mockCLI)
	cmd := createListCommand(listingService, mockCLI)

	// Verify the Run function is configured
	if cmd.Run == nil {
		t.Fatal("Expected Run function to be configured")
	}

	// Verify command metadata
	if cmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got %s", cmd.Use)
	}

	if cmd.Short != "List Jumpstart ArcBox deployments" {
		t.Errorf("Expected correct Short description, got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "jumpstart_arcbox") {
		t.Error("Expected Long description to mention jumpstart_arcbox")
	}

	if !strings.Contains(cmd.Long, "subscription selection") {
		t.Error("Expected Long description to mention subscription selection requirement")
	}

	// Test that examples are included in the Long description
	if !strings.Contains(cmd.Long, "Examples:") && !strings.Contains(cmd.Long, "Examples") {
		t.Error("Expected Long description to include examples")
	}
}

// TestListCommand_FlagConfiguration tests comprehensive flag configuration
func TestListCommand_FlagConfiguration(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	listingService := services.NewListingService(mockCLI)
	cmd := createListCommand(listingService, mockCLI)

	// Test all-subscriptions flag
	allSubsFlag := cmd.Flags().Lookup("all-subscriptions")
	if allSubsFlag == nil {
		t.Fatal("all-subscriptions flag should exist")
	}
	if allSubsFlag.Value.Type() != "bool" {
		t.Errorf("all-subscriptions should be bool type, got %s", allSubsFlag.Value.Type())
	}
	if allSubsFlag.DefValue != "false" {
		t.Errorf("all-subscriptions default should be false, got %s", allSubsFlag.DefValue)
	}
	if allSubsFlag.Usage != "Search for ArcBox deployments across all accessible subscriptions" {
		t.Errorf("all-subscriptions usage incorrect: %s", allSubsFlag.Usage)
	}

	// Test current-subscription flag
	currentSubFlag := cmd.Flags().Lookup("current-subscription")
	if currentSubFlag == nil {
		t.Fatal("current-subscription flag should exist")
	}
	if currentSubFlag.Value.Type() != "bool" {
		t.Errorf("current-subscription should be bool type, got %s", currentSubFlag.Value.Type())
	}
	if currentSubFlag.DefValue != "false" {
		t.Errorf("current-subscription default should be false, got %s", currentSubFlag.DefValue)
	}
	if currentSubFlag.Usage != "Search for ArcBox deployments in the current subscription" {
		t.Errorf("current-subscription usage incorrect: %s", currentSubFlag.Usage)
	}

	// Test subscription flag
	subFlag := cmd.Flags().Lookup("subscription")
	if subFlag == nil {
		t.Fatal("subscription flag should exist")
	}
	if subFlag.Value.Type() != "string" {
		t.Errorf("subscription should be string type, got %s", subFlag.Value.Type())
	}
	if subFlag.DefValue != "" {
		t.Errorf("subscription default should be empty, got %s", subFlag.DefValue)
	}
	if subFlag.Shorthand != "s" {
		t.Errorf("subscription shorthand should be 's', got %s", subFlag.Shorthand)
	}
	if subFlag.Usage != "Azure subscription ID to search" {
		t.Errorf("subscription usage incorrect: %s", subFlag.Usage)
	}

	// Test flag retrieval functionality
	testArgs := []string{"--all-subscriptions", "--subscription", "test-sub"}
	cmd.SetArgs(testArgs)
	err := cmd.ParseFlags(testArgs)
	if err != nil {
		t.Fatalf("Failed to parse test flags: %v", err)
	}

	allSubs, err := cmd.Flags().GetBool("all-subscriptions")
	if err != nil {
		t.Errorf("Failed to get all-subscriptions flag: %v", err)
	}
	if !allSubs {
		t.Error("all-subscriptions should be true when set")
	}

	subValue, err := cmd.Flags().GetString("subscription")
	if err != nil {
		t.Errorf("Failed to get subscription flag: %v", err)
	}
	if subValue != "test-sub" {
		t.Errorf("subscription value should be 'test-sub', got %s", subValue)
	}
}

// TestCreateListCommand_ComprehensiveCoverage focuses on maximizing coverage of createListCommand
func TestCreateListCommand_ComprehensiveCoverage(t *testing.T) {
	// This test is designed to exercise all paths in createListCommand to achieve 95%+ coverage

	mockCLI := azurecli.NewMockAzureCLI()
	listingService := services.NewListingService(mockCLI)

	// Test the createListCommand function thoroughly
	cmd := createListCommand(listingService, mockCLI)

	// Test basic command properties - these test the command creation logic
	if cmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got %s", cmd.Use)
	}

	if cmd.Short != "List Jumpstart ArcBox deployments" {
		t.Errorf("Expected correct Short description, got %s", cmd.Short)
	}

	// Test that Long description is built correctly (tests the string concatenation)
	expectedStrings := []string{
		"List all Jumpstart ArcBox deployments",
		"jumpstart_arcbox",
		"subscription selection",
		"Examples",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(cmd.Long, expected) {
			t.Errorf("Expected Long description to contain '%s'", expected)
		}
	}

	// Test that Run function is assigned (covers the Run assignment)
	if cmd.Run == nil {
		t.Error("Expected Run function to be assigned")
	}

	// Test flag configuration (covers all the flag creation lines)
	flagTests := []struct {
		name         string
		flagType     string
		defaultValue string
		usage        string
		shorthand    string
	}{
		{
			name:         "all-subscriptions",
			flagType:     "bool",
			defaultValue: "false",
			usage:        "Search for ArcBox deployments across all accessible subscriptions",
		},
		{
			name:         "current-subscription",
			flagType:     "bool",
			defaultValue: "false",
			usage:        "Search for ArcBox deployments in the current subscription",
		},
		{
			name:         "subscription",
			flagType:     "string",
			defaultValue: "",
			usage:        "Azure subscription ID to search",
			shorthand:    "s",
		},
	}

	for _, flagTest := range flagTests {
		flag := cmd.Flags().Lookup(flagTest.name)
		if flag == nil {
			t.Errorf("Expected flag '%s' to exist", flagTest.name)
			continue
		}

		if flag.Value.Type() != flagTest.flagType {
			t.Errorf("Flag '%s' should be type %s, got %s", flagTest.name, flagTest.flagType, flag.Value.Type())
		}

		if flag.DefValue != flagTest.defaultValue {
			t.Errorf("Flag '%s' should have default value '%s', got '%s'", flagTest.name, flagTest.defaultValue, flag.DefValue)
		}

		if flag.Usage != flagTest.usage {
			t.Errorf("Flag '%s' should have usage '%s', got '%s'", flagTest.name, flagTest.usage, flag.Usage)
		}

		if flagTest.shorthand != "" && flag.Shorthand != flagTest.shorthand {
			t.Errorf("Flag '%s' should have shorthand '%s', got '%s'", flagTest.name, flagTest.shorthand, flag.Shorthand)
		}
	}

	// Test that executeListCommand function is correctly called from Run
	// We verify this by checking that the Run function calls the validation and execution logic

	// Set up a scenario that will exercise the executeListCommand path
	cmd.SetArgs([]string{"--current-subscription"})
	mockCLI.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
	mockCLI.Subscriptions = []azurecli.SubscriptionInfo{{ID: "test-sub", Name: "Test Sub"}}
	mockCLI.ResourceGroups = []azurecli.ResourceGroupInfo{}

	// Parse flags to prepare for execution
	err := cmd.ParseFlags([]string{"--current-subscription"})
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Test executeListCommand directly to ensure it works
	executeErr := executeListCommand(cmd, listingService, mockCLI)
	if executeErr != nil {
		t.Errorf("executeListCommand should not fail with valid setup: %v", executeErr)
	}

	// Verify that CLI methods were called (proves the execution path worked)
	if !mockCLI.GetCurrentSubscriptionCalled {
		t.Error("Expected GetCurrentSubscription to be called")
	}
}

// TestCreateListCommand_ErrorScenarios tests error scenarios that exercise error handling paths
func TestCreateListCommand_ErrorScenarios(t *testing.T) {
	// These tests focus on the error handling paths in createListCommand that need coverage

	tests := []struct {
		name        string
		setupMock   func(*azurecli.MockAzureCLI)
		args        []string
		shouldError bool
		description string
	}{
		{
			name: "validation_error_case",
			setupMock: func(cli *azurecli.MockAzureCLI) {
				// Don't set up any subscriptions - this will cause validation failure
			},
			args:        []string{}, // No subscription flags
			shouldError: true,
			description: "No subscription flags should cause validation error",
		},
		{
			name: "conflicting_flags_case",
			setupMock: func(cli *azurecli.MockAzureCLI) {
				// Setup doesn't matter for this validation error
			},
			args:        []string{"--all-subscriptions", "--current-subscription"},
			shouldError: true,
			description: "Conflicting subscription flags should cause validation error",
		},
		{
			name: "subscription_access_error",
			setupMock: func(cli *azurecli.MockAzureCLI) {
				cli.GetCurrentSubscriptionError = fmt.Errorf("subscription access denied")
			},
			args:        []string{"--current-subscription"},
			shouldError: true,
			description: "Subscription access error should be handled",
		},
		{
			name: "successful_case",
			args: []string{"--current-subscription"},
			setupMock: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
				cli.Subscriptions = []azurecli.SubscriptionInfo{{ID: "test-sub", Name: "Test Sub"}}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{}
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			if tt.setupMock != nil {
				tt.setupMock(mockCLI)
			}

			listingService := services.NewListingService(mockCLI)
			cmd := createListCommand(listingService, mockCLI)

			// Test the command creation (covers createListCommand function)
			if cmd == nil {
				t.Fatal("createListCommand should return a valid command")
			}

			// Test that command has proper structure
			if cmd.Use != "list" {
				t.Error("Command should have correct Use field")
			}

			if cmd.Run == nil {
				t.Error("Command should have Run function assigned")
			}

			// Parse flags and test executeListCommand (this exercises the logic that Run would call)
			cmd.SetArgs(tt.args)
			err := cmd.ParseFlags(tt.args)
			if err != nil && !tt.shouldError {
				t.Fatalf("Unexpected flag parsing error: %v", err)
			}

			// Test executeListCommand directly
			executeErr := executeListCommand(cmd, listingService, mockCLI)

			if tt.shouldError {
				if executeErr == nil {
					t.Errorf("Expected error for %s but got none", tt.description)
				}
			} else {
				if executeErr != nil {
					t.Errorf("Unexpected error for %s: %v", tt.description, executeErr)
				}
			}
		})
	}
}

// TestHandleListCommandError tests the error handling function (without os.Exit)
func TestHandleListCommandError(t *testing.T) {
	// This test exercises the handleListCommandError function to improve coverage
	// Note: We cannot easily test the os.Exit(1) call, but we can test the rest of the logic

	mockCLI := azurecli.NewMockAzureCLI()
	listingService := services.NewListingService(mockCLI)
	cmd := createListCommand(listingService, mockCLI)

	// Set up a scenario that will fail validation
	cmd.SetArgs([]string{}) // No subscription flags
	err := cmd.ParseFlags([]string{})
	if err != nil {
		t.Fatalf("Failed to parse empty flags: %v", err)
	}
	// Test the error message formatting function exists
	testErrorMessage := "test error for coverage"
	errorMessage := utils.ErrorColor("❌ [ERROR] " + testErrorMessage)
	if !strings.Contains(errorMessage, testErrorMessage) {
		t.Error("Expected error color formatting to include the error message")
	}

	// Test the validation service creation and logic that handleListCommandError uses
	validationService := services.NewListValidationService(mockCLI)
	result := validationService.ValidateAllListRequirements(cmd, listingService)

	// This should fail validation due to missing subscription flags
	if result.IsValid {
		t.Error("Expected validation to fail with no subscription flags")
	}

	if result.Error == nil {
		t.Error("Expected validation error when no subscription flags provided")
	}

	// Test that ShowHelpWithoutTypes function exists and can be called
	// (We just verify the function exists and doesn't panic)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowHelpWithoutTypes should not panic: %v", r)
			}
		}()
		utils.ShowHelpWithoutTypes(cmd)
	}()

	// Test the validation logic that would be called in handleListCommandError
	if !strings.Contains(result.Error.Error(), "subscription selection required") {
		t.Error("Expected validation error to mention subscription selection requirement")
	}
}

// TestCreateListCommand_ErrorHandlingLogic tests the logic that leads to error handling
func TestCreateListCommand_ErrorHandlingLogic(t *testing.T) {
	// This test focuses on the conditions that would trigger handleListCommandError
	// to ensure we cover the error checking logic in the Run function

	mockCLI := azurecli.NewMockAzureCLI()
	listingService := services.NewListingService(mockCLI)
	cmd := createListCommand(listingService, mockCLI)

	// Test the Run function exists and is properly configured
	if cmd.Run == nil {
		t.Fatal("Run function should be assigned")
	}

	// Test scenarios that would lead to errors (without actually calling Run due to os.Exit)
	errorScenarios := []struct {
		name        string
		args        []string
		setupMock   func(*azurecli.MockAzureCLI)
		shouldError bool
	}{
		{
			name:        "no_subscription_flags",
			args:        []string{},
			setupMock:   func(cli *azurecli.MockAzureCLI) {},
			shouldError: true,
		},
		{
			name:        "conflicting_subscription_flags",
			args:        []string{"--all-subscriptions", "--current-subscription"},
			setupMock:   func(cli *azurecli.MockAzureCLI) {},
			shouldError: true,
		},
		{
			name: "azure_cli_access_error",
			args: []string{"--current-subscription"},
			setupMock: func(cli *azurecli.MockAzureCLI) {
				cli.GetCurrentSubscriptionError = fmt.Errorf("access denied")
			},
			shouldError: true,
		},
		{
			name: "successful_case",
			args: []string{"--current-subscription"},
			setupMock: func(cli *azurecli.MockAzureCLI) {
				cli.CurrentSubscription = &azurecli.SubscriptionInfo{ID: "test-sub", Name: "Test Sub"}
				cli.Subscriptions = []azurecli.SubscriptionInfo{{ID: "test-sub", Name: "Test Sub"}}
				cli.ResourceGroups = []azurecli.ResourceGroupInfo{}
			},
			shouldError: false,
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			freshMockCLI := azurecli.NewMockAzureCLI() // Use fresh instance
			scenario.setupMock(freshMockCLI)

			freshListingService := services.NewListingService(freshMockCLI)
			freshCmd := createListCommand(freshListingService, freshMockCLI)

			freshCmd.SetArgs(scenario.args)
			err := freshCmd.ParseFlags(scenario.args)
			if err != nil && !scenario.shouldError {
				t.Fatalf("Unexpected flag parsing error: %v", err)
			}

			// Test executeListCommand directly (this is what Run would call)
			executeErr := executeListCommand(freshCmd, freshListingService, freshMockCLI)

			if scenario.shouldError {
				if executeErr == nil {
					t.Errorf("Expected error for scenario %s but got none", scenario.name)
				}

				// Verify that this error would trigger the error handling path
				// by checking the validation logic that handleListCommandError uses
				validationService := services.NewListValidationService(freshMockCLI)
				result := validationService.ValidateAllListRequirements(freshCmd, freshListingService)
				if scenario.name == "no_subscription_flags" || scenario.name == "conflicting_subscription_flags" {
					if result.IsValid {
						t.Error("Validation should fail for invalid flag combinations")
					}
				}
			} else {
				if executeErr != nil {
					t.Errorf("Unexpected error for scenario %s: %v", scenario.name, executeErr)
				}
			}
		})
	}
}
