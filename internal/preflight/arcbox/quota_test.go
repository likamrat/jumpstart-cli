package arcbox

import (
	"fmt"
	"testing"

	"jumpstartcli/internal/azurecli"
)

func TestCheckQuotaForSKU(t *testing.T) {
	tests := []struct {
		name         string
		sku          string
		required     int
		region       string
		flavor       string
		mockSetup    func(*azurecli.MockAzureCLI)
		expectQuota  bool
		expectAvail  bool
		expectDeploy bool
		expectError  bool
	}{
		{
			name:     "sufficient quota and available SKU",
			sku:      "Standard_D8s_v5",
			required: 8,
			region:   "eastus",
			flavor:   "ITPro",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
					{
						Name: map[string]string{
							"value":          "Standard DSv5 Family vCPUs",
							"localizedValue": "Standard DSv5 Family vCPUs",
						},
						CurrentValue: 2,
						Limit:        64,
					},
				})
				mock.SetAvailableSKUsForRegion("eastus", []string{"Standard_D8s_v5"})
			},
			expectQuota:  true,
			expectAvail:  true,
			expectDeploy: true,
			expectError:  false,
		},
		{
			name:     "insufficient quota",
			sku:      "Standard_D8s_v5",
			required: 8,
			region:   "eastus",
			flavor:   "ITPro",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
					{
						Name: map[string]string{
							"value":          "Standard DSv5 Family vCPUs",
							"localizedValue": "Standard DSv5 Family vCPUs",
						},
						CurrentValue: 60,
						Limit:        64,
					},
				})
				mock.SetAvailableSKUsForRegion("eastus", []string{"Standard_D8s_v5"})
			},
			expectQuota:  false,
			expectAvail:  true,
			expectDeploy: false,
			expectError:  false,
		},
		{
			name:     "SKU not available",
			sku:      "Standard_D8s_v5",
			required: 8,
			region:   "eastus",
			flavor:   "ITPro",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
					{
						Name: map[string]string{
							"value":          "Standard DSv5 Family vCPUs",
							"localizedValue": "Standard DSv5 Family vCPUs",
						},
						CurrentValue: 2,
						Limit:        64,
					},
				})
				mock.SetAvailableSKUsForRegion("eastus", []string{})
			},
			expectQuota:  true,
			expectAvail:  false,
			expectDeploy: false,
			expectError:  false,
		},
		{
			name:     "VM usage error",
			sku:      "Standard_D8s_v5",
			required: 8,
			region:   "eastus",
			flavor:   "ITPro",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetErrorForListVMUsage(fmt.Errorf("failed to get VM usage"))
			},
			expectError: true,
		},
		{
			name:     "SKU availability error",
			sku:      "Standard_D8s_v5",
			required: 8,
			region:   "eastus",
			flavor:   "ITPro",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{})
				mock.SetErrorForCheckSKUAvailability(fmt.Errorf("failed to check SKU"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := azurecli.NewMockAzureCLI()
			tt.mockSetup(mock)

			result := CheckQuotaForSKU(mock, tt.sku, tt.required, tt.region, tt.flavor)

			if tt.expectError && result.Details == "" {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectError && result.Details != "" {
				t.Errorf("Expected no result.Detailsor, got %v", result.Details)
			}

			if !tt.expectError {
				if result.QuotaOK != tt.expectQuota {
					t.Errorf("Expected quota OK %v, got %v", tt.expectQuota, result.QuotaOK)
				}

				if result.CanDeploy != tt.expectDeploy {
					t.Errorf("Expected deploy OK %v, got %v", tt.expectDeploy, result.CanDeploy)
				}

				if tt.expectQuota {
					if result.Available < tt.required {
						t.Errorf("Expected result.Available (%d) >= required (%d)", result.Available, tt.required)
					}
					if result.Current+tt.required > result.Limit {
						t.Errorf("Expected result.Current (%d) + required (%d) <= result.Limit (%d)", result.Current, tt.required, result.Limit)
					}
				}
			}
		})
	}
}

func TestCheckBatchSKUAvailability(t *testing.T) {
	tests := []struct {
		name            string
		skus            []string
		region          string
		mockSetup       func(*azurecli.MockAzureCLI)
		expectedUnavail []string
		expectError     bool
	}{
		{
			name:   "all SKUs available",
			skus:   []string{"Standard_D8s_v5", "Standard_B2ms"},
			region: "eastus",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetAvailableSKUsForRegion("eastus", []string{"Standard_D8s_v5", "Standard_B2ms"})
			},
			expectedUnavail: []string{},
			expectError:     false,
		},
		{
			name:   "some SKUs unavailable",
			skus:   []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_Unavailable"},
			region: "eastus",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetAvailableSKUsForRegion("eastus", []string{"Standard_D8s_v5"})
			},
			expectedUnavail: []string{"Standard_B2ms", "Standard_Unavailable"},
			expectError:     false,
		},
		{
			name:   "empty SKU list",
			skus:   []string{},
			region: "eastus",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				// No setup needed
			},
			expectedUnavail: []string{},
			expectError:     false,
		},
		{
			name:   "SKU availability check error",
			skus:   []string{"Standard_D8s_v5"},
			region: "eastus",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetErrorForCheckSKUAvailability(fmt.Errorf("check failed"))
			},
			expectedUnavail: []string{"Standard_D8s_v5"},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := azurecli.NewMockAzureCLI()
			tt.mockSetup(mock)

			unavailable, err := CheckBatchSKUAvailability(mock, tt.skus, tt.region)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(unavailable) != len(tt.expectedUnavail) {
				t.Errorf("Expected %d unavailable SKUs, got %d", len(tt.expectedUnavail), len(unavailable))
			}

			// Check that all expected unavailable SKUs are in the result
			for _, expected := range tt.expectedUnavail {
				found := false
				for _, actual := range unavailable {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected unavailable SKU %s not found in result", expected)
				}
			}
		})
	}
}

func TestRunQuotaChecks(t *testing.T) {
	tests := []struct {
		name         string
		flavor       string
		region       string
		subscription string
		mockSetup    func(*azurecli.MockAzureCLI)
		expectError  bool
		expectedSKUs []string
	}{
		{
			name:         "ITPro flavor successful",
			flavor:       "ITPro",
			region:       "eastus",
			subscription: "test-sub",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
					{
						Name: map[string]string{
							"value":          "Standard DSv5 Family vCPUs",
							"localizedValue": "Standard DSv5 Family vCPUs",
						},
						CurrentValue: 0,
						Limit:        64,
					},
				})
				mock.SetAvailableSKUsForRegion("eastus", []string{"Standard_D8s_v5"})
			},
			expectError:  false,
			expectedSKUs: []string{"Standard_D8s_v5"},
		},
		{
			name:         "DevOps flavor successful",
			flavor:       "DevOps",
			region:       "eastus",
			subscription: "test-sub",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
					{
						Name: map[string]string{
							"value":          "Standard DSv5 Family vCPUs",
							"localizedValue": "Standard DSv5 Family vCPUs",
						},
						CurrentValue: 0,
						Limit:        64,
					},
					{
						Name: map[string]string{
							"value":          "Standard BS Family vCPUs",
							"localizedValue": "Standard BS Family vCPUs",
						},
						CurrentValue: 0,
						Limit:        32,
					},
					{
						Name: map[string]string{
							"value":          "Standard DSv4 Family vCPUs",
							"localizedValue": "Standard DSv4 Family vCPUs",
						},
						CurrentValue: 0,
						Limit:        32,
					},
				})
				mock.SetAvailableSKUsForRegion("eastus", []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"})
			},
			expectError:  false,
			expectedSKUs: []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"},
		},
		{
			name:         "unknown flavor error",
			flavor:       "UnknownFlavor",
			region:       "eastus",
			subscription: "test-sub",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				// No setup needed
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := azurecli.NewMockAzureCLI()
			tt.mockSetup(mock)

			results, err := RunQuotaChecks(mock, tt.region, tt.flavor, tt.subscription)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if !tt.expectError {
				if len(results) != len(tt.expectedSKUs) {
					t.Errorf("Expected %d results, got %d", len(tt.expectedSKUs), len(results))
				}

				// Check that all expected SKUs are in the results
				for _, expectedSKU := range tt.expectedSKUs {
					found := false
					for _, result := range results {
						if result.SKU == expectedSKU {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected SKU %s not found in results", expectedSKU)
					}
				}
			}
		})
	}
}

// Additional comprehensive test cases for 100% coverage

func TestQuotaEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		sku       string
		required  int
		region    string
		flavor    string
		mockSetup func(*azurecli.MockAzureCLI)
	}{
		{
			name:     "quota family not found fallback to total regional",
			sku:      "Standard_X99_v1", // Unknown SKU that falls back to total regional
			required: 4,
			region:   "eastus",
			flavor:   "ITPro",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
{
Name: map[string]string{
"value":          "Total Regional vCPUs",
"localizedValue": "Total Regional vCPUs",
},
CurrentValue: 10,
Limit:        100,
},
})
				mock.SetAvailableSKUsForRegion("eastus", []string{"Standard_X99_v1"})
			},
		},
		{
			name:     "no quota family and no total regional found",
			sku:      "Standard_Unknown",
			required: 4,
			region:   "eastus",
			flavor:   "ITPro",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
{
Name: map[string]string{
"value":          "Some Other Quota",
"localizedValue": "Some Other Quota",
},
CurrentValue: 10,
Limit:        100,
},
})
				mock.SetAvailableSKUsForRegion("eastus", []string{"Standard_Unknown"})
			},
		},
		{
			name:     "usage info with missing name fields",
			sku:      "Standard_D8s_v5",
			required: 8,
			region:   "eastus",
			flavor:   "ITPro",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
{
Name: map[string]string{
// Missing "value" field
"localizedValue": "Standard DSv5 Family vCPUs",
},
CurrentValue: 2,
Limit:        64,
},
{
Name: map[string]string{
"value": "Standard DSv5 Family vCPUs",
// Missing "localizedValue" field
},
CurrentValue: 2,
Limit:        64,
},
{
Name: map[string]string{
"value":          "Standard DSv5 Family vCPUs",
"localizedValue": "Standard DSv5 Family vCPUs",
},
CurrentValue: 2,
Limit:        64,
},
})
				mock.SetAvailableSKUsForRegion("eastus", []string{"Standard_D8s_v5"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
mock := azurecli.NewMockAzureCLI()
			tt.mockSetup(mock)

			result := CheckQuotaForSKU(mock, tt.sku, tt.required, tt.region, tt.flavor)
			
			// Just ensure it doesn't panic and returns a result
if result.SKU != tt.sku {
t.Errorf("Expected SKU %s, got %s", tt.sku, result.SKU)
}
})
}
}

func TestAllFlavorsScenario(t *testing.T) {
tests := []struct {
name         string
flavor       string
region       string
subscription string
mockSetup    func(*azurecli.MockAzureCLI)
expectError  bool
}{
{
name:         "all flavors successful",
flavor:       "all",
region:       "eastus",
subscription: "test-sub",
mockSetup: func(mock *azurecli.MockAzureCLI) {
// Set up quota for all possible SKUs
mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
{
Name: map[string]string{
"value":          "Standard DSv5 Family vCPUs",
"localizedValue": "Standard DSv5 Family vCPUs",
},
CurrentValue: 0,
Limit:        64,
},
{
Name: map[string]string{
"value":          "Standard BS Family vCPUs",
"localizedValue": "Standard BS Family vCPUs",
},
CurrentValue: 0,
Limit:        32,
},
{
Name: map[string]string{
"value":          "Standard DSv4 Family vCPUs",
"localizedValue": "Standard DSv4 Family vCPUs",
},
CurrentValue: 0,
Limit:        32,
},
})
// All SKUs from ITPro, DevOps, and DataOps flavors
allSKUs := []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}
mock.SetAvailableSKUsForRegion("eastus", allSKUs)
},
expectError: false,
},
{
name:         "DataOps flavor",
flavor:       "DataOps",
region:       "eastus",
subscription: "test-sub",
mockSetup: func(mock *azurecli.MockAzureCLI) {
mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
{
Name: map[string]string{
"value":          "Standard DSv5 Family vCPUs",
"localizedValue": "Standard DSv5 Family vCPUs",
},
CurrentValue: 0,
Limit:        64,
},
{
Name: map[string]string{
"value":          "Standard BS Family vCPUs",
"localizedValue": "Standard BS Family vCPUs",
},
CurrentValue: 0,
Limit:        32,
},
{
Name: map[string]string{
"value":          "Standard DSv4 Family vCPUs",
"localizedValue": "Standard DSv4 Family vCPUs",
},
CurrentValue: 0,
Limit:        32,
},
})
dataOpsSKUs := []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}
mock.SetAvailableSKUsForRegion("eastus", dataOpsSKUs)
},
expectError: false,
},
{
name:         "batch SKU availability check error",
flavor:       "ITPro",
region:       "eastus",
subscription: "test-sub",
mockSetup: func(mock *azurecli.MockAzureCLI) {
				// Set up quota data so RunQuotaChecks has the data it needs
				mock.SetVMUsageForRegion("eastus", []azurecli.VMUsageInfo{
					{
						Name: map[string]string{
							"value":          "Standard DSv5 Family vCPUs",
							"localizedValue": "Standard DSv5 Family vCPUs",
						},
						CurrentValue: 0,
						Limit:        64,
					},
				})
mock.SetErrorForCheckSKUAvailability(fmt.Errorf("batch check failed"))
},
expectError: false,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
mock := azurecli.NewMockAzureCLI()
tt.mockSetup(mock)

results, err := RunQuotaChecks(mock, tt.region, tt.flavor, tt.subscription)

if tt.expectError && err == nil {
t.Errorf("Expected error, got nil")
}

if !tt.expectError && err != nil {
t.Errorf("Expected no error, got %v", err)
}

if !tt.expectError && len(results) == 0 {
t.Errorf("Expected results, got empty slice")
}
})
}
}

func TestHelperFunctionsComprehensive(t *testing.T) {
// Test getFlavorSKUs with all cases including unknown
tests := []struct {
flavor       string
expectedSKUs int
}{
{"itpro", 1},
{"ITPRO", 1}, // Test case insensitivity
{"devops", 5},
{"DEVOPS", 5},
{"dataops", 5},
{"DATAOPS", 5},
{"unknown", 0},
{"", 0},
}

for _, tt := range tests {
t.Run(fmt.Sprintf("getFlavorSKUs_%s", tt.flavor), func(t *testing.T) {
skus := getFlavorSKUs(tt.flavor)
if len(skus) != tt.expectedSKUs {
t.Errorf("Expected %d SKUs for flavor %s, got %d", tt.expectedSKUs, tt.flavor, len(skus))
}
})
}

// Test getRequiredVCPUForSKU with known and unknown SKUs
vcpuTests := []struct {
sku      string
expected int
}{
{"Standard_D8s_v5", 8},
{"Standard_D8s_v4", 8},
{"Standard_B2ms", 2},
{"Standard_B4ms", 4},
{"Standard_B8ms", 8},
{"Standard_Unknown", 1}, // Fallback case
{"", 1},                 // Empty case
}

for _, tt := range vcpuTests {
t.Run(fmt.Sprintf("getRequiredVCPUForSKU_%s", tt.sku), func(t *testing.T) {
vcpu := getRequiredVCPUForSKU(tt.sku)
if vcpu != tt.expected {
t.Errorf("Expected %d vCPU for SKU %s, got %d", tt.expected, tt.sku, vcpu)
}
})
}

// Test mapSKUToFamilyQuotaName with edge cases
familyTests := []struct {
sku      string
expected string
}{
{"Standard_D8s_v5", "Standard Dsv5 Family vCPUs"},
{"Standard_B2ms", "Standard BS Family vCPUs"},
{"Standard_B4ms", "Standard BS Family vCPUs"},
{"Standard_D8s_v4", "Standard Dsv4 Family vCPUs"},
{"Standard_F4s_v2", "Standard Fsv2 Family vCPUs"},
{"Standard_X", "Standard X Family vCPUs"}, // Single part, not B-series
{"", ""},           // Empty string
{"InvalidFormat", "Standard InvalidFormat Family vCPUs"}, // No underscore
{"Standard_", ""},     // Just prefix
}

for _, tt := range familyTests {
t.Run(fmt.Sprintf("mapSKUToFamilyQuotaName_%s", tt.sku), func(t *testing.T) {
family := mapSKUToFamilyQuotaName(tt.sku)
if family != tt.expected {
t.Errorf("Expected family name '%s' for SKU %s, got '%s'", tt.expected, tt.sku, family)
}
})
}
}
