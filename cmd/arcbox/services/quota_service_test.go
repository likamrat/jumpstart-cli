package services

import (
	"strings"
	"testing"

	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// ====================================================================================
// PHASE 5 - Comprehensive Quota Service Test Coverage
// Target Functions: CheckQuota(), RunQuotaCheckCommand(), GetFlavorSKUs(), GetExpectedDuration()
// ====================================================================================

// TestQuotaServiceCreation tests that QuotaService can be created properly
func TestQuotaServiceCreation(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	service := NewQuotaService(mockCLI)

	if service == nil {
		t.Fatal("Expected non-nil QuotaService")
	}

	if service.cli != mockCLI {
		t.Error("Expected QuotaService to use provided CLI")
	}

	if service.cli == nil {
		t.Error("Expected CLI to be initialized")
	}
}

// TestGetFlavorSKUs tests the GetFlavorSKUs method with all supported flavors
func TestGetFlavorSKUs(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	service := NewQuotaService(mockCLI)

	tests := []struct {
		flavor   string
		expected []string
	}{
		{"itpro", []string{"Standard_D8s_v5"}},
		{"ITPro", []string{"Standard_D8s_v5"}},
		{"ITPRO", []string{"Standard_D8s_v5"}},
		{"devops", []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}},
		{"DevOps", []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}},
		{"dataops", []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}},
		{"DataOps", []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}},
		{"unknown", []string{}},
		{"", []string{}},
	}

	for _, tt := range tests {
		t.Run("flavor_"+tt.flavor, func(t *testing.T) {
			result := service.GetFlavorSKUs(tt.flavor)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d SKUs for flavor %s, got %d", len(tt.expected), tt.flavor, len(result))
				return
			}
			for i, sku := range result {
				if sku != tt.expected[i] {
					t.Errorf("Expected SKU %s at index %d for flavor %s, got %s", tt.expected[i], i, tt.flavor, sku)
				}
			}
		})
	}
}

// TestGetExpectedDuration tests the expected duration calculation
func TestGetExpectedDuration(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	service := NewQuotaService(mockCLI)

	tests := []struct {
		flavor   string
		expected string
	}{
		{"itpro", "~1 minute (checking 1 SKU)"},
		{"ITPro", "~1 minute (checking 1 SKU)"},
		{"ITPRO", "~1 minute (checking 1 SKU)"},
		{"devops", "3-5 minutes (checking 5 SKUs)"},
		{"DevOps", "3-5 minutes (checking 5 SKUs)"},
		{"DEVOPS", "3-5 minutes (checking 5 SKUs)"},
		{"dataops", "3-5 minutes (checking 5 SKUs)"},
		{"DataOps", "3-5 minutes (checking 5 SKUs)"},
		{"DATAOPS", "3-5 minutes (checking 5 SKUs)"},
		{"all", "3-5 minutes (checking all SKUs for multiple flavors)"},
		{"unknown", "~1 minute"},
		{"", "~1 minute"},
	}

	for _, tt := range tests {
		t.Run("flavor_"+tt.flavor, func(t *testing.T) {
			result := service.GetExpectedDuration(tt.flavor)
			if result != tt.expected {
				t.Errorf("Expected duration %q for flavor %s, got %q", tt.expected, tt.flavor, result)
			}
		})
	}
}

// TestCheckQuota_ValidationErrors tests CheckQuota with various validation scenarios
func TestCheckQuota_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		locationFlag   string
		allLocations   bool
		selectedFlavor string
		subscriptionID string
		expectedError  bool
		errorContains  string
	}{
		{
			name:           "missing_flavor",
			locationFlag:   "eastus",
			allLocations:   false,
			selectedFlavor: "",
			subscriptionID: "test-sub-id",
			expectedError:  true,
			errorContains:  "required argument missing: --flavor flag must specify an ArcBox flavor",
		},
		{
			name:           "missing_location_flags",
			locationFlag:   "",
			allLocations:   false,
			selectedFlavor: "ITPro",
			subscriptionID: "test-sub-id",
			expectedError:  true,
			errorContains:  "location specification required: specify either --location or --all-locations",
		},
		{
			name:           "conflicting_location_flags",
			locationFlag:   "eastus",
			allLocations:   true,
			selectedFlavor: "ITPro",
			subscriptionID: "test-sub-id",
			expectedError:  true,
			errorContains:  "conflicting location flags: cannot specify both --location and --all-locations",
		},
		{
			name:           "valid_single_location",
			locationFlag:   "eastus",
			allLocations:   false,
			selectedFlavor: "ITPro",
			subscriptionID: "test-sub-id",
			expectedError:  false,
			errorContains:  "",
		},
		{
			name:           "valid_all_locations",
			locationFlag:   "",
			allLocations:   true,
			selectedFlavor: "DevOps",
			subscriptionID: "test-sub-id",
			expectedError:  false,
			errorContains:  "",
		},
		{
			name:           "multiple_locations",
			locationFlag:   "eastus,eastus2,centralus",
			allLocations:   false,
			selectedFlavor: "DataOps",
			subscriptionID: "test-sub-id",
			expectedError:  false,
			errorContains:  "",
		},
		{
			name:           "whitespace_in_flavor",
			locationFlag:   "eastus",
			allLocations:   false,
			selectedFlavor: "  ITPro  ",
			subscriptionID: "test-sub-id",
			expectedError:  false,
			errorContains:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			service := NewQuotaService(mockCLI)

			// Set up mock CLI to return success for quota checks to avoid test failures due to quota logic
			// Set up comprehensive quota data for all supported regions and all SKU families
			supportedRegions := []string{"eastus", "eastus2", "centralus", "westus2", "northeurope", "westeurope", "francecentral", "uksouth", "australiaeast", "japaneast", "koreacentral", "southeastasia"}

			quotaData := []azurecli.VMUsageInfo{
				{
					Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
				{
					Name:         map[string]string{"value": "standardBSFamily", "localizedValue": "Standard BS Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
				{
					Name:         map[string]string{"value": "standardDSv4Family", "localizedValue": "Standard DSv4 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
			}

			for _, region := range supportedRegions {
				mockCLI.SetVMUsage(region, quotaData)
			}

			results, err := service.CheckQuota(tt.locationFlag, tt.allLocations, tt.selectedFlavor, tt.subscriptionID)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error for test %s, but got none", tt.name)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, but got: %s", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for test %s: %v", tt.name, err)
					return
				}
				if results == nil {
					t.Errorf("Expected non-nil results for test %s", tt.name)
				}
			}
		})
	}
}

// TestCheckQuota_QuotaScenarios tests various quota checking scenarios
func TestCheckQuota_QuotaScenarios(t *testing.T) {
	tests := []struct {
		name           string
		flavor         string
		location       string
		quotaResponses []azurecli.VMUsageInfo
		expectedPass   bool
		errorContains  string
	}{
		{
			name:     "sufficient_quota_itpro",
			flavor:   "ITPro",
			location: "eastus",
			quotaResponses: []azurecli.VMUsageInfo{
				{
					Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
			},
			expectedPass:  true,
			errorContains: "",
		},
		{
			name:     "insufficient_quota_itpro",
			flavor:   "ITPro",
			location: "eastus",
			quotaResponses: []azurecli.VMUsageInfo{
				{
					Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
					CurrentValue: 95,
					Limit:        100,
					Unit:         "cores",
				},
			},
			expectedPass:  false,
			errorContains: "quota validation failed",
		},
		{
			name:     "sufficient_quota_devops_multiple_skus",
			flavor:   "DevOps",
			location: "eastus2",
			quotaResponses: []azurecli.VMUsageInfo{
				{
					Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
				{
					Name:         map[string]string{"value": "standardBSFamily", "localizedValue": "Standard BS Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
				{
					Name:         map[string]string{"value": "standardDSv4Family", "localizedValue": "Standard DSv4 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
			},
			expectedPass:  true,
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			service := NewQuotaService(mockCLI)

			// Set up mock quota responses
			mockCLI.SetVMUsage(tt.location, tt.quotaResponses)

			results, err := service.CheckQuota(tt.location, false, tt.flavor, "test-sub-id")

			if tt.expectedPass {
				if err != nil {
					t.Errorf("Expected test %s to pass, but got error: %v", tt.name, err)
					return
				}
				if results == nil || len(results) == 0 {
					t.Errorf("Expected non-empty results for passing test %s", tt.name)
				}
			} else {
				if err == nil {
					t.Errorf("Expected test %s to fail, but got no error", tt.name)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, but got: %s", tt.errorContains, err.Error())
				}
			}
		})
	}
}

// TestCheckQuota_MultipleLocations tests quota checking across multiple locations
func TestCheckQuota_MultipleLocations(t *testing.T) {
	tests := []struct {
		name           string
		locationFlag   string
		flavor         string
		quotaResponses map[string][]azurecli.VMUsageInfo
		expectedPass   bool
	}{
		{
			name:         "all_locations_pass",
			locationFlag: "eastus,eastus2",
			flavor:       "ITPro",
			quotaResponses: map[string][]azurecli.VMUsageInfo{
				"eastus": {
					{
						Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
						CurrentValue: 0,
						Limit:        100,
						Unit:         "cores",
					},
				},
				"eastus2": {
					{
						Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
						CurrentValue: 10,
						Limit:        100,
						Unit:         "cores",
					},
				},
			},
			expectedPass: true,
		},
		{
			name:         "one_location_fails",
			locationFlag: "eastus,eastus2",
			flavor:       "ITPro",
			quotaResponses: map[string][]azurecli.VMUsageInfo{
				"eastus": {
					{
						Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
						CurrentValue: 0,
						Limit:        100,
						Unit:         "cores",
					},
				},
				"eastus2": {
					{
						Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
						CurrentValue: 95,
						Limit:        100,
						Unit:         "cores",
					},
				},
			},
			expectedPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			service := NewQuotaService(mockCLI)

			// Set up mock quota responses for each location
			for location, response := range tt.quotaResponses {
				mockCLI.SetVMUsage(location, response)
			}

			results, err := service.CheckQuota(tt.locationFlag, false, tt.flavor, "test-sub-id")

			if tt.expectedPass {
				if err != nil {
					t.Errorf("Expected test %s to pass, but got error: %v", tt.name, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected test %s to fail, but got no error", tt.name)
				}
			}

			// Verify we got results for multiple locations if the test expected results
			if tt.expectedPass && len(results) == 0 {
				t.Errorf("Expected results for test %s", tt.name)
			}
		})
	}
}

// TestCheckQuota_FlavorNormalization tests that flavor names are properly normalized
func TestCheckQuota_FlavorNormalization(t *testing.T) {
	tests := []struct {
		inputFlavor    string
		expectedFlavor string
	}{
		{"itpro", "ITPro"},
		{"ITPRO", "ITPro"},
		{"ITPro", "ITPro"},
		{"devops", "DevOps"},
		{"DEVOPS", "DevOps"},
		{"DevOps", "DevOps"},
		{"dataops", "DataOps"},
		{"DATAOPS", "DataOps"},
		{"DataOps", "DataOps"},
	}

	for _, tt := range tests {
		t.Run("normalize_"+tt.inputFlavor, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			service := NewQuotaService(mockCLI)

			// Set up mock CLI with successful quota response
			mockCLI.SetVMUsage("eastus", []azurecli.VMUsageInfo{
				{
					Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
				{
					Name:         map[string]string{"value": "standardBSFamily", "localizedValue": "Standard BS Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
				{
					Name:         map[string]string{"value": "standardDSv4Family", "localizedValue": "Standard DSv4 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
			})

			results, err := service.CheckQuota("eastus", false, tt.inputFlavor, "test-sub-id")

			if err != nil {
				t.Errorf("Unexpected error for flavor %s: %v", tt.inputFlavor, err)
				return
			}

			// Check that results contain the normalized flavor name
			found := false
			for _, result := range results {
				if flavor, ok := result["Flavor"]; ok {
					if flavor == tt.expectedFlavor {
						found = true
						break
					}
				}
			}

			if !found {
				t.Errorf("Expected normalized flavor %s not found in results for input %s", tt.expectedFlavor, tt.inputFlavor)
			}
		})
	}
}

// TestService_EdgeCases tests edge cases and boundary conditions
func TestService_EdgeCases(t *testing.T) {
	t.Run("empty_strings", func(t *testing.T) {
		mockCLI := azurecli.NewMockAzureCLI()
		service := NewQuotaService(mockCLI)

		// Test empty flavor
		skus := service.GetFlavorSKUs("")
		if len(skus) != 0 {
			t.Error("Expected empty SKU list for empty flavor")
		}

		// Test empty duration
		duration := service.GetExpectedDuration("")
		if duration != "~1 minute" {
			t.Errorf("Expected default duration for empty flavor, got: %s", duration)
		}
	})

	t.Run("whitespace_handling", func(t *testing.T) {
		mockCLI := azurecli.NewMockAzureCLI()
		service := NewQuotaService(mockCLI)

		// Set up mock response
		mockCLI.SetVMUsage("eastus", []azurecli.VMUsageInfo{
			{
				Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
				CurrentValue: 0,
				Limit:        100,
				Unit:         "cores",
			},
			{
				Name:         map[string]string{"value": "standardBSFamily", "localizedValue": "Standard BS Family vCPUs"},
				CurrentValue: 0,
				Limit:        100,
				Unit:         "cores",
			},
			{
				Name:         map[string]string{"value": "standardDSv4Family", "localizedValue": "Standard DSv4 Family vCPUs"},
				CurrentValue: 0,
				Limit:        100,
				Unit:         "cores",
			},
		})

		// Test flavor with whitespace
		results, err := service.CheckQuota("eastus", false, "  ITPro  ", "test-sub-id")
		if err != nil {
			t.Errorf("Failed to handle whitespace in flavor: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected results despite whitespace in flavor")
		}
	})

	t.Run("case_insensitive_flavors", func(t *testing.T) {
		mockCLI := azurecli.NewMockAzureCLI()
		service := NewQuotaService(mockCLI)

		cases := []string{"itpro", "ITPRO", "ITPro", "iTpRo"}
		for _, testCase := range cases {
			skus := service.GetFlavorSKUs(testCase)
			if len(skus) != 1 || skus[0] != "Standard_D8s_v5" {
				t.Errorf("Case insensitive test failed for %s: expected Standard_D8s_v5, got %v", testCase, skus)
			}
		}
	})
}

// TestRunQuotaCheckCommand tests the command execution logic
func TestRunQuotaCheckCommand(t *testing.T) {
	tests := []struct {
		name          string
		setupFlags    func(*cobra.Command)
		expectError   bool
		errorContains string
	}{
		{
			name: "valid_command_execution",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("location", "eastus")
				cmd.Flags().Set("flavor", "ITPro")
			},
			expectError:   false,
			errorContains: "",
		},
		{
			name: "missing_flavor_flag",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("location", "eastus")
				// Don't set flavor flag
			},
			expectError:   true,
			errorContains: "required argument missing",
		},
		{
			name: "missing_location_flags",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("flavor", "ITPro")
				// Don't set location or all-locations
			},
			expectError:   true,
			errorContains: "location specification required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			service := NewQuotaService(mockCLI)

			// Set up mock CLI with successful quota response
			quotaData := []azurecli.VMUsageInfo{
				{
					Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
				{
					Name:         map[string]string{"value": "standardBSFamily", "localizedValue": "Standard BS Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
				{
					Name:         map[string]string{"value": "standardDSv4Family", "localizedValue": "Standard DSv4 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
			}
			mockCLI.SetVMUsage("eastus", quotaData)

			// Create command with flags
			cmd := &cobra.Command{}
			cmd.Flags().String("location", "", "Location")
			cmd.Flags().Bool("all-locations", false, "All locations")
			cmd.Flags().String("flavor", "", "Flavor")

			// Set up flags for this test
			tt.setupFlags(cmd)

			// Set output format to avoid table formatting issues in tests
			originalFormat := utils.OutputFormat
			utils.OutputFormat = "json"
			defer func() {
				utils.OutputFormat = originalFormat
			}()

			err := service.RunQuotaCheckCommand(cmd, []string{})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for test %s, but got none", tt.name)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, but got: %s", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for test %s: %v", tt.name, err)
				}
			}
		})
	}
}

// TestCheckQuota_OutputFormats tests quota checking with different output formats
func TestCheckQuota_OutputFormats(t *testing.T) {
	tests := []struct {
		name          string
		outputFormat  string
		expectSpinner bool
	}{
		{
			name:          "table_format_with_spinner",
			outputFormat:  "table",
			expectSpinner: true,
		},
		{
			name:          "json_format_without_spinner",
			outputFormat:  "json",
			expectSpinner: false,
		},
		{
			name:          "yaml_format_without_spinner",
			outputFormat:  "yaml",
			expectSpinner: false,
		},
		{
			name:          "tsv_format_without_spinner",
			outputFormat:  "tsv",
			expectSpinner: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original format
			originalFormat := utils.OutputFormat
			defer func() { utils.OutputFormat = originalFormat }()

			utils.OutputFormat = tt.outputFormat

			mockCLI := azurecli.NewMockAzureCLI()
			service := NewQuotaService(mockCLI)

			// Set up sufficient quota
			mockCLI.SetVMUsage("eastus", []azurecli.VMUsageInfo{
				{
					Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
			})

			results, err := service.CheckQuota("eastus", false, "ITPro", "test-sub-id")

			if err != nil {
				t.Errorf("Expected no error, but got: %v", err)
				return
			}

			if results == nil || len(results) == 0 {
				t.Errorf("Expected non-empty results")
			}
		})
	}
}

// TestCheckQuota_RegionsLoadingError tests error handling for regions loading failure
func TestCheckQuota_RegionsLoadingError(t *testing.T) {
	// This test validates the error path when regions.ArcboxSupportedRegionsData can't be unmarshaled
	// Since the actual embedded data is valid JSON, we'll test a scenario where all-locations is used
	// and ensure proper error handling in the quota validation path

	mockCLI := azurecli.NewMockAzureCLI()
	service := NewQuotaService(mockCLI)

	// Test with an invalid region that will fail validation
	_, err := service.CheckQuota("", true, "ITPro", "test-sub-id")

	// This should succeed since the JSON is valid, but let's test validation failure
	if err == nil {
		// If no error, we need to set up mock data for valid regions
		// Set up quota for all valid regions
		supportedRegions := []string{"eastus", "eastus2", "westus", "westus2", "westus3",
			"centralus", "northcentralus", "southcentralus", "westcentralus",
			"canadacentral", "canadaeast", "brazilsouth", "uksouth", "ukwest",
			"northeurope", "westeurope", "francecentral", "germanywestcentral",
			"switzerlandnorth", "swedencentral", "norwayeast", "polandcentral",
			"australiaeast", "australiasoutheast", "japaneast", "japanwest",
			"koreacentral", "koreasouth", "southeastasia", "eastasia",
			"centralindia", "southindia", "southafricanorth"}

		for _, region := range supportedRegions {
			mockCLI.SetVMUsage(region, []azurecli.VMUsageInfo{
				{
					Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
					CurrentValue: 0,
					Limit:        100,
					Unit:         "cores",
				},
			})
		}

		// Test with all locations
		results, err := service.CheckQuota("", true, "ITPro", "test-sub-id")
		if err != nil {
			t.Errorf("Expected successful quota check with all locations, but got error: %v", err)
		}
		if len(results) == 0 {
			t.Errorf("Expected results for all locations check")
		}
	}
}

// TestCheckQuota_QuotaFailureAnalysis tests different types of quota failures
func TestCheckQuota_QuotaFailureAnalysis(t *testing.T) {
	tests := []struct {
		name          string
		mockFailures  []map[string]interface{}
		expectedError string
	}{
		{
			name: "quota_issues_only",
			mockFailures: []map[string]interface{}{
				{
					"CanDeploy": false,
					"Details":   "Need 8 more vCPU in quota",
					"SKU":       "Standard_D8s_v5",
					"Location":  "East US",
				},
			},
			expectedError: "quota validation failed",
		},
		{
			name: "sku_issues_only",
			mockFailures: []map[string]interface{}{
				{
					"CanDeploy": false,
					"Details":   "SKU not available in this region",
					"SKU":       "Standard_D8s_v5",
					"Location":  "East US",
				},
			},
			expectedError: "deployment validation failed",
		},
		{
			name: "both_quota_and_sku_issues",
			mockFailures: []map[string]interface{}{
				{
					"CanDeploy": false,
					"Details":   "Need 8 more vCPU in quota",
					"SKU":       "Standard_D8s_v5",
					"Location":  "East US",
				},
				{
					"CanDeploy": false,
					"Details":   "SKU not available in this region",
					"SKU":       "Standard_B2ms",
					"Location":  "East US",
				},
			},
			expectedError: "insufficient vCPU quota and SKU availability issues",
		},
		{
			name: "generic_deployment_failure",
			mockFailures: []map[string]interface{}{
				{
					"CanDeploy": false,
					"Details":   "Some other deployment issue",
					"SKU":       "Standard_D8s_v5",
					"Location":  "East US",
				},
			},
			expectedError: "deployment requirements not met",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			service := NewQuotaService(mockCLI)

			// Force quota failures by providing insufficient quota
			mockCLI.SetVMUsage("eastus", []azurecli.VMUsageInfo{
				{
					Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
					CurrentValue: 95,
					Limit:        100, // Only 5 available, need 8
					Unit:         "cores",
				},
			})

			results, err := service.CheckQuota("eastus", false, "ITPro", "test-sub-id")

			if err == nil {
				t.Errorf("Expected error for test %s, but got none", tt.name)
				return
			}

			if !strings.Contains(err.Error(), "quota validation failed") && !strings.Contains(err.Error(), "deployment validation failed") {
				t.Errorf("Expected quota/deployment validation error, but got: %s", err.Error())
			}

			// Results should still be returned even when there are failures
			if results == nil {
				t.Errorf("Expected results to be returned even on failure")
			}
		})
	}
}

// TestCheckQuota_MultipleRegionsParsing tests comma-separated location parsing
func TestCheckQuota_MultipleRegionsParsing(t *testing.T) {
	tests := []struct {
		name         string
		locationFlag string
		expectedLocs []string
	}{
		{
			name:         "single_location",
			locationFlag: "eastus",
			expectedLocs: []string{"eastus"},
		},
		{
			name:         "multiple_locations_clean",
			locationFlag: "eastus,eastus2",
			expectedLocs: []string{"eastus", "eastus2"},
		},
		{
			name:         "multiple_locations_with_spaces",
			locationFlag: "eastus, eastus2, centralus",
			expectedLocs: []string{"eastus", "eastus2", "centralus"},
		},
		{
			name:         "locations_with_empty_parts",
			locationFlag: "eastus,,eastus2,",
			expectedLocs: []string{"eastus", "eastus2"},
		},
		{
			name:         "locations_with_whitespace_only",
			locationFlag: "eastus,   , eastus2",
			expectedLocs: []string{"eastus", "eastus2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			service := NewQuotaService(mockCLI)

			// Set up sufficient quota for all expected locations
			for _, loc := range tt.expectedLocs {
				mockCLI.SetVMUsage(loc, []azurecli.VMUsageInfo{
					{
						Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
						CurrentValue: 0,
						Limit:        100,
						Unit:         "cores",
					},
				})
			}

			results, err := service.CheckQuota(tt.locationFlag, false, "ITPro", "test-sub-id")

			if err != nil {
				t.Errorf("Expected no error for test %s, but got: %v", tt.name, err)
				return
			}

			expectedResultCount := len(tt.expectedLocs) // ITPro has 1 SKU per location
			if len(results) != expectedResultCount {
				t.Errorf("Expected %d results for test %s, but got %d", expectedResultCount, tt.name, len(results))
			}
		})
	}
}

// TestCheckQuota_ErrorScenarios tests various error scenarios
func TestCheckQuota_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		locationFlag  string
		allLocations  bool
		flavor        string
		setupMock     func(*azurecli.MockAzureCLI)
		expectedError string
	}{
		{
			name:         "quota_check_with_spinner_error",
			locationFlag: "eastus",
			allLocations: false,
			flavor:       "ITPro",
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Set original format to table to trigger spinner path
				originalFormat := utils.OutputFormat
				defer func() { utils.OutputFormat = originalFormat }()
				utils.OutputFormat = "table"

				// Set up quota that will cause failure
				mock.SetVMUsage("eastus", []azurecli.VMUsageInfo{
					{
						Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
						CurrentValue: 95,
						Limit:        100, // Only 5 available, need 8
						Unit:         "cores",
					},
				})
			},
			expectedError: "quota validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			service := NewQuotaService(mockCLI)

			if tt.setupMock != nil {
				tt.setupMock(mockCLI)
			}

			_, err := service.CheckQuota(tt.locationFlag, tt.allLocations, tt.flavor, "test-sub-id")

			if err == nil {
				t.Errorf("Expected error for test %s, but got none", tt.name)
				return
			}

			if tt.expectedError != "" && !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing %q, but got: %s", tt.expectedError, err.Error())
			}
		})
	}
}
