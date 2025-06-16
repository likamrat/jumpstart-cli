package utils

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"jumpstartcli/internal/azurecli"

	"github.com/spf13/cobra"
)

func TestGetSubscriptionID(t *testing.T) {
	tests := []struct {
		name           string
		flagValue      string
		envValue       string
		currentSubID   string
		currentSubErr  error
		expectedResult string
	}{
		{
			name:           "flag_takes_priority",
			flagValue:      "flag-sub-id",
			envValue:       "env-sub-id",
			currentSubID:   "current-sub-id",
			expectedResult: "flag-sub-id",
		},
		{
			name:           "env_when_no_flag",
			flagValue:      "",
			envValue:       "env-sub-id",
			currentSubID:   "current-sub-id",
			expectedResult: "env-sub-id",
		},
		{
			name:           "current_when_no_flag_or_env",
			flagValue:      "",
			envValue:       "",
			currentSubID:   "current-sub-id",
			expectedResult: "current-sub-id",
		},
		{
			name:           "empty_when_all_fail",
			flagValue:      "",
			envValue:       "",
			currentSubErr:  errors.New("no current subscription"),
			expectedResult: "",
		},
		{
			name:           "empty_when_current_sub_nil",
			flagValue:      "",
			envValue:       "",
			currentSubID:   "",
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock command with subscription flag
			cmd := &cobra.Command{}
			cmd.Flags().String("subscription", "", "subscription flag")
			if tt.flagValue != "" {
				cmd.Flags().Set("subscription", tt.flagValue)
			}

			// Set environment variable
			originalEnv := os.Getenv("AZURE_SUBSCRIPTION_ID")
			defer func() {
				if originalEnv != "" {
					os.Setenv("AZURE_SUBSCRIPTION_ID", originalEnv)
				} else {
					os.Unsetenv("AZURE_SUBSCRIPTION_ID")
				}
			}()

			if tt.envValue != "" {
				os.Setenv("AZURE_SUBSCRIPTION_ID", tt.envValue)
			} else {
				os.Unsetenv("AZURE_SUBSCRIPTION_ID")
			}

			// Create mock Azure CLI
			mockCLI := &azurecli.MockAzureCLI{
				GetCurrentSubscriptionError: tt.currentSubErr,
			}
			if tt.currentSubID != "" {
				mockCLI.CurrentSubscription = &azurecli.SubscriptionInfo{
					ID: tt.currentSubID,
				}
			}

			result := GetSubscriptionID(cmd, mockCLI)
			if result != tt.expectedResult {
				t.Errorf("GetSubscriptionID() = %q, expected %q", result, tt.expectedResult)
			}
		})
	}
}

func TestSetAzureSubscription(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID string
		setupMock      func(*azurecli.MockAzureCLI)
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:           "successful_set",
			subscriptionID: "valid-sub-id",
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Add the subscription to the mock's subscriptions list
				mock.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "valid-sub-id", Name: "Test Subscription", IsDefault: false},
				}
			},
			wantErr: false,
		},
		{
			name:           "empty_subscription_id",
			subscriptionID: "",
			setupMock:      func(mock *azurecli.MockAzureCLI) {},
			wantErr:        true,
			expectedErrMsg: "subscription ID is empty",
		},
		{
			name:           "azure_cli_error",
			subscriptionID: "test-sub",
			setupMock: func(mock *azurecli.MockAzureCLI) {
				mock.SetErrorForSetSubscription(fmt.Errorf("az cli error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			tt.setupMock(mockCLI)

			err := SetAzureSubscription(mockCLI, tt.subscriptionID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SetAzureSubscription() expected error, got nil")
					return
				}
				if tt.expectedErrMsg != "" && err.Error() != tt.expectedErrMsg {
					t.Errorf("SetAzureSubscription() error = %q, expected %q", err.Error(), tt.expectedErrMsg)
				}
			} else {
				if err != nil {
					t.Errorf("SetAzureSubscription() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCheckResourceGroupExists(t *testing.T) {
	tests := []struct {
		name                string
		resourceGroupName   string
		subscriptionID      string
		setupMock           func(*azurecli.MockAzureCLI)
		expectedResult      bool
		wantErr             bool
		expectedErrContains string
	}{
		{
			name:              "successful_check_exists",
			resourceGroupName: "test-rg",
			subscriptionID:    "test-sub",
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Add subscription for successful SetSubscription
				mock.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Subscription", IsDefault: false},
				}
				mock.ResourceGroupExists["test-rg"] = true
			},
			expectedResult: true,
			wantErr:        false,
		},
		{
			name:              "successful_check_not_exists",
			resourceGroupName: "test-rg",
			subscriptionID:    "test-sub",
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Add subscription for successful SetSubscription
				mock.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Subscription", IsDefault: false},
				}
				mock.ResourceGroupExists["test-rg"] = false
			},
			expectedResult: false,
			wantErr:        false,
		},
		{
			name:              "no_subscription_provided",
			resourceGroupName: "test-rg",
			subscriptionID:    "",
			setupMock: func(mock *azurecli.MockAzureCLI) {
				mock.ResourceGroupExists["test-rg"] = true
			},
			expectedResult: true,
			wantErr:        false,
		},
		{
			name:              "set_subscription_fails",
			resourceGroupName: "test-rg",
			subscriptionID:    "test-sub",
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Don't add subscription to list, causing SetSubscription to fail
			},
			wantErr:             true,
			expectedErrContains: "failed to set subscription context",
		},
		{
			name:              "check_resource_group_fails",
			resourceGroupName: "test-rg",
			subscriptionID:    "test-sub",
			setupMock: func(mock *azurecli.MockAzureCLI) {
				// Add subscription for successful SetSubscription
				mock.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub", Name: "Test Subscription", IsDefault: false},
				}
				mock.CheckResourceGroupExistsError = errors.New("access denied")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			tt.setupMock(mockCLI)

			result, err := CheckResourceGroupExists(mockCLI, tt.resourceGroupName, tt.subscriptionID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CheckResourceGroupExists() expected error, got nil")
					return
				}
				if tt.expectedErrContains != "" && !containsString(err.Error(), tt.expectedErrContains) {
					t.Errorf("CheckResourceGroupExists() error = %q, expected to contain %q", err.Error(), tt.expectedErrContains)
				}
			} else {
				if err != nil {
					t.Errorf("CheckResourceGroupExists() unexpected error: %v", err)
					return
				}
				if result != tt.expectedResult {
					t.Errorf("CheckResourceGroupExists() = %v, expected %v", result, tt.expectedResult)
				}
			}
		})
	}
}

func TestGetRequiredVCPUForSKU(t *testing.T) {
	tests := []struct {
		name     string
		sku      string
		expected int
	}{
		// Known SKUs
		{
			name:     "standard_d8s_v5",
			sku:      "Standard_D8s_v5",
			expected: 8,
		},
		{
			name:     "standard_d8s_v4",
			sku:      "Standard_D8s_v4",
			expected: 8,
		},
		{
			name:     "standard_b2ms",
			sku:      "Standard_B2ms",
			expected: 2,
		},
		{
			name:     "standard_b4ms",
			sku:      "Standard_B4ms",
			expected: 4,
		},
		{
			name:     "standard_b8ms",
			sku:      "Standard_B8ms",
			expected: 8,
		},

		// Case sensitivity
		{
			name:     "lowercase_sku",
			sku:      "standard_d8s_v5",
			expected: 1, // Should fallback to default
		},
		{
			name:     "uppercase_sku",
			sku:      "STANDARD_D8S_V5",
			expected: 1, // Should fallback to default
		},

		// Unknown SKUs - should return default (1)
		{
			name:     "unknown_sku",
			sku:      "Standard_Unknown",
			expected: 1,
		},
		{
			name:     "empty_sku",
			sku:      "",
			expected: 1,
		},
		{
			name:     "invalid_format",
			sku:      "NotAValidSKU",
			expected: 1,
		},

		// Edge cases
		{
			name:     "sku_with_spaces",
			sku:      " Standard_D8s_v5 ",
			expected: 1, // Exact match required
		},
		{
			name:     "partial_sku_name",
			sku:      "D8s_v5",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRequiredVCPUForSKU(tt.sku)
			if result != tt.expected {
				t.Errorf("GetRequiredVCPUForSKU(%q) = %d, expected %d", tt.sku, result, tt.expected)
			}
		})
	}
}

func TestMapSKUToFamilyQuotaName(t *testing.T) {
	tests := []struct {
		name     string
		sku      string
		expected string
	}{
		// Valid D-series SKUs
		{
			name:     "d8s_v5",
			sku:      "Standard_D8s_v5",
			expected: "Standard Dsv5 Family vCPUs",
		},
		{
			name:     "d4s_v4",
			sku:      "Standard_D4s_v4",
			expected: "Standard Dsv4 Family vCPUs",
		},
		{
			name:     "d2s_v3",
			sku:      "Standard_D2s_v3",
			expected: "Standard Dsv3 Family vCPUs",
		},

		// B-series SKUs (these don't follow the standard naming pattern)
		{
			name:     "b2ms",
			sku:      "Standard_B2ms",
			expected: "", // B-series doesn't have version suffix, so function returns empty
		},
		{
			name:     "b4ms",
			sku:      "Standard_B4ms",
			expected: "", // B-series doesn't have version suffix, so function returns empty
		},

		// E-series SKUs
		{
			name:     "e4s_v5",
			sku:      "Standard_E4s_v5",
			expected: "Standard Esv5 Family vCPUs",
		},

		// F-series SKUs
		{
			name:     "f8s_v2",
			sku:      "Standard_F8s_v2",
			expected: "Standard Fsv2 Family vCPUs",
		},

		// SKUs without 's' suffix
		{
			name:     "d8_v5",
			sku:      "Standard_D8_v5",
			expected: "Standard Dv5 Family vCPUs",
		},

		// Edge cases
		{
			name:     "sku_without_standard_prefix",
			sku:      "D8s_v5",
			expected: "Standard Dsv5 Family vCPUs",
		},
		{
			name:     "sku_with_only_one_part",
			sku:      "Standard_D8s",
			expected: "", // Should return empty for invalid format
		},
		{
			name:     "empty_sku",
			sku:      "",
			expected: "", // Should return empty
		},
		{
			name:     "invalid_format",
			sku:      "InvalidSKU",
			expected: "", // Should return empty
		},

		// Complex cases
		{
			name:     "sku_with_multiple_numbers",
			sku:      "Standard_D16s_v5",
			expected: "Standard Dsv5 Family vCPUs",
		},
		{
			name:     "sku_with_mixed_case",
			sku:      "Standard_D8S_V5",            // Mixed case handling
			expected: "Standard DSV5 Family vCPUs", // Function doesn't preserve lowercase 'v'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapSKUToFamilyQuotaName(tt.sku)
			if result != tt.expected {
				t.Errorf("MapSKUToFamilyQuotaName(%q) = %q, expected %q", tt.sku, result, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// Benchmark tests to ensure functions are performant
func BenchmarkGetRequiredVCPUForSKU(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetRequiredVCPUForSKU("Standard_D8s_v5")
	}
}

func BenchmarkMapSKUToFamilyQuotaName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MapSKUToFamilyQuotaName("Standard_D8s_v5")
	}
}
