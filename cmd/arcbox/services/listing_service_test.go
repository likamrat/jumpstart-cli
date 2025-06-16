package services

import (
	"fmt"
	"strings"
	"testing"

	"jumpstartcli/cmd/arcbox/models"
	"jumpstartcli/internal/azurecli"
)

// TestListingServiceCreation tests that ListingService can be created properly
func TestListingServiceCreation(t *testing.T) {
	mockCLI := &azurecli.MockAzureCLI{}
	service := NewListingService(mockCLI)

	if service == nil {
		t.Fatal("Expected non-nil ListingService")
	}

	if service.cli != mockCLI {
		t.Error("Expected ListingService to use provided CLI")
	}

	if service.formatter == nil {
		t.Error("Expected formatter to be initialized")
	}
}

// TestListingService_GetSubscription tests getting a specific subscription
func TestListingService_GetSubscription(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID string
		mockSetup      func(*azurecli.MockAzureCLI)
		expectedError  string
		expectedSub    *models.AzureSubscription
	}{
		{
			name:           "valid_subscription",
			subscriptionID: "test-sub-123",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "test-sub-123", Name: "Test Subscription"},
				}
			},
			expectedError: "",
			expectedSub: &models.AzureSubscription{
				ID:   "test-sub-123",
				Name: "Test Subscription",
			},
		},
		{
			name:           "subscription_not_found",
			subscriptionID: "nonexistent-sub",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "other-sub", Name: "Other Subscription"},
				}
				mock.GetSubscriptionError = fmt.Errorf("subscription not found")
			},
			expectedError: "subscription not found",
			expectedSub:   nil,
		},
		{
			name:           "cli_error",
			subscriptionID: "test-sub",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.GetSubscriptionError = fmt.Errorf("CLI error")
			},
			expectedError: "CLI error",
			expectedSub:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := &azurecli.MockAzureCLI{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			service := NewListingService(mockCLI)
			result, err := service.GetSubscription(tt.subscriptionID)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if tt.expectedSub != nil {
					if result.ID != tt.expectedSub.ID {
						t.Errorf("Expected subscription ID %s, got %s", tt.expectedSub.ID, result.ID)
					}
					if result.Name != tt.expectedSub.Name {
						t.Errorf("Expected subscription name %s, got %s", tt.expectedSub.Name, result.Name)
					}
				}
			}
		})
	}
}

// TestListingService_GetCurrentSubscription tests getting the current subscription
func TestListingService_GetCurrentSubscription(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*azurecli.MockAzureCLI)
		expectedError string
		expectedSub   *models.AzureSubscription
	}{
		{
			name: "valid_current_subscription",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.CurrentSubscription = &azurecli.SubscriptionInfo{
					ID:   "current-sub",
					Name: "Current Subscription",
				}
			},
			expectedError: "",
			expectedSub: &models.AzureSubscription{
				ID:   "current-sub",
				Name: "Current Subscription",
			},
		},
		{
			name: "cli_error",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.GetCurrentSubscriptionError = fmt.Errorf("CLI error")
			},
			expectedError: "CLI error",
			expectedSub:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := &azurecli.MockAzureCLI{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			service := NewListingService(mockCLI)
			result, err := service.getCurrentSubscription()

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if tt.expectedSub != nil {
					if result.ID != tt.expectedSub.ID {
						t.Errorf("Expected subscription ID %s, got %s", tt.expectedSub.ID, result.ID)
					}
					if result.Name != tt.expectedSub.Name {
						t.Errorf("Expected subscription name %s, got %s", tt.expectedSub.Name, result.Name)
					}
				}
			}
		})
	}
}

// TestListingService_ListDeployments tests the main list deployments functionality
func TestListingService_ListDeployments(t *testing.T) {
	tests := []struct {
		name                string
		allSubscriptions    bool
		currentSubscription bool
		subscriptionID      string
		outputFormat        string
		mockSetup           func(*azurecli.MockAzureCLI)
		shouldError         bool
		expectedError       string
	}{
		{
			name:                "current_subscription_valid",
			allSubscriptions:    false,
			currentSubscription: true,
			subscriptionID:      "",
			outputFormat:        "table",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
				mock.CurrentSubscription = &azurecli.SubscriptionInfo{
					ID:   "current-sub",
					Name: "Current Subscription",
				}
				mock.ResourceGroups = []azurecli.ResourceGroupInfo{
					{
						Name:     "arcbox-rg",
						Location: "eastus",
					},
				}
			},
			shouldError: false,
		},
		{
			name:                "specific_subscription_valid",
			allSubscriptions:    false,
			currentSubscription: false,
			subscriptionID:      "specific-sub",
			outputFormat:        "json",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
				mock.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "specific-sub", Name: "Specific Subscription"},
				}
				mock.ResourceGroups = []azurecli.ResourceGroupInfo{
					{
						Name:     "arcbox-rg",
						Location: "westus",
					},
				}
			},
			shouldError: false,
		},
		{
			name:                "all_subscriptions_valid",
			allSubscriptions:    true,
			currentSubscription: false,
			subscriptionID:      "",
			outputFormat:        "yaml",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
				mock.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "sub-1", Name: "Subscription 1"},
					{ID: "sub-2", Name: "Subscription 2"},
				}
				mock.ResourceGroups = []azurecli.ResourceGroupInfo{
					{
						Name:     "arcbox-rg-1",
						Location: "eastus",
					},
				}
			},
			shouldError: false,
		},
		{
			name:                "no_selection_error",
			allSubscriptions:    false,
			currentSubscription: false,
			subscriptionID:      "",
			outputFormat:        "table",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
			},
			shouldError:   false, // Service doesn't error, it just finds no deployments
			expectedError: "",
		},
		{
			name:                "cli_error",
			allSubscriptions:    false,
			currentSubscription: true,
			subscriptionID:      "",
			outputFormat:        "table",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
				mock.GetCurrentSubscriptionError = fmt.Errorf("CLI connection error")
			},
			shouldError:   true,
			expectedError: "CLI connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := &azurecli.MockAzureCLI{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			service := NewListingService(mockCLI)
			err := service.ListDeployments(tt.allSubscriptions, tt.currentSubscription, tt.subscriptionID, tt.outputFormat)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if tt.expectedError != "" && !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestListingService_DetectArcBoxFlavor tests flavor detection functionality
func TestListingService_DetectArcBoxFlavor(t *testing.T) {
	tests := []struct {
		name              string
		resourceGroupName string
		mockSetup         func(*azurecli.MockAzureCLI)
		expectedFlavor    string
		expectedPrefix    string
	}{
		{
			name:              "itpro_flavor_detected",
			resourceGroupName: "arcbox-itpro-rg",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				// Setup deployments list error to force fallback method
				mock.ListDeploymentsError = fmt.Errorf("no deployments found")

				mock.Resources = map[string][]azurecli.ResourceInfo{
					"arcbox-itpro-rg": {
						{
							Name: "arcbox-vm",
							Type: "Microsoft.Compute/virtualMachines",
						},
					},
				}
			},
			expectedFlavor: "ITPro",
			expectedPrefix: "ArcBox",
		},
		{
			name:              "empty_resource_group",
			resourceGroupName: "empty-rg",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				// Setup deployments list error to force fallback method
				mock.ListDeploymentsError = fmt.Errorf("no deployments found")

				mock.Resources = map[string][]azurecli.ResourceInfo{
					"empty-rg": {},
				}
			},
			expectedFlavor: "ITPro",
			expectedPrefix: "ArcBox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := &azurecli.MockAzureCLI{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			service := NewListingService(mockCLI)
			flavor, prefix := service.DetectArcBoxFlavor(tt.resourceGroupName)

			if flavor != tt.expectedFlavor {
				t.Errorf("Expected flavor %s, got %s", tt.expectedFlavor, flavor)
			}

			if prefix != tt.expectedPrefix {
				t.Errorf("Expected prefix %s, got %s", tt.expectedPrefix, prefix)
			}
		})
	}
}

// TestListingService_OutputFormats tests different output format handling
func TestListingService_OutputFormats(t *testing.T) {
	mockCLI := &azurecli.MockAzureCLI{
		IsLoggedInResult: true,
		CurrentSubscription: &azurecli.SubscriptionInfo{
			ID:   "test-sub",
			Name: "Test Subscription",
		},
		ResourceGroups: []azurecli.ResourceGroupInfo{
			{
				Name:     "arcbox-rg",
				Location: "eastus",
			},
		},
	}

	service := NewListingService(mockCLI)

	formats := []string{"table", "json", "yaml", "tsv"}
	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			err := service.ListDeployments(false, true, "", format)
			if err != nil {
				t.Errorf("Expected no error for format %s, got %v", format, err)
			}
		})
	}
}

// TestListingService_GetAllSubscriptions tests getting all subscriptions
func TestListingService_GetAllSubscriptions(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*azurecli.MockAzureCLI)
		expectedCount int
		expectedError string
	}{
		{
			name: "multiple_subscriptions",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.Subscriptions = []azurecli.SubscriptionInfo{
					{ID: "sub-1", Name: "Subscription 1"},
					{ID: "sub-2", Name: "Subscription 2"},
					{ID: "sub-3", Name: "Subscription 3"},
				}
			},
			expectedCount: 3,
			expectedError: "",
		},
		{
			name: "no_subscriptions",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.Subscriptions = []azurecli.SubscriptionInfo{}
			},
			expectedCount: 0,
			expectedError: "",
		},
		{
			name: "cli_error",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.ListSubscriptionsError = fmt.Errorf("CLI error")
			},
			expectedCount: 0,
			expectedError: "CLI error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := &azurecli.MockAzureCLI{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			service := NewListingService(mockCLI)
			result, err := service.getAllSubscriptions()

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(result) != tt.expectedCount {
					t.Errorf("Expected %d subscriptions, got %d", tt.expectedCount, len(result))
				}
			}
		})
	}
}
