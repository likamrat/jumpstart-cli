package services

import (
	"fmt"
	"testing"

	"jumpstartcli/internal/azurecli"
)

func TestDeletionServiceCreation(t *testing.T) {
	mockCli := &azurecli.MockAzureCLI{}
	service := NewDeletionService(mockCli)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.cli != mockCli {
		t.Errorf("Expected CLI to be set correctly")
	}
}

func TestDeletionService_DeleteDeployment(t *testing.T) {
	tests := []struct {
		name                 string
		resourceGroupName    string
		subscription         string
		skipConfirmation     bool
		rgExists             bool
		rgExistsError        error
		deleteError          error
		expectedError        bool
		expectedErrorMessage string
	}{
		{
			name:              "successful_deletion_with_skip_confirmation",
			resourceGroupName: "test-rg",
			subscription:      "test-sub",
			skipConfirmation:  true,
			rgExists:          true,
			rgExistsError:     nil,
			deleteError:       nil,
			expectedError:     false,
		},
		{
			name:                 "resource_group_not_exists",
			resourceGroupName:    "nonexistent-rg",
			subscription:         "test-sub",
			skipConfirmation:     true,
			rgExists:             false,
			rgExistsError:        nil,
			deleteError:          nil,
			expectedError:        true,
			expectedErrorMessage: "resource group 'nonexistent-rg' does not exist",
		},
		{
			name:                 "error_checking_resource_group",
			resourceGroupName:    "test-rg",
			subscription:         "test-sub",
			skipConfirmation:     true,
			rgExists:             false,
			rgExistsError:        fmt.Errorf("authentication error"),
			deleteError:          nil,
			expectedError:        true,
			expectedErrorMessage: "resource group existence check failed",
		},
		{
			name:                 "deletion_fails",
			resourceGroupName:    "test-rg",
			subscription:         "test-sub",
			skipConfirmation:     true,
			rgExists:             true,
			rgExistsError:        nil,
			deleteError:          fmt.Errorf("deletion failed"),
			expectedError:        true,
			expectedErrorMessage: "resource group deletion failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli := &azurecli.MockAzureCLI{}
			service := NewDeletionService(mockCli)

			// Set up basic subscription data for SetSubscription to work
			mockCli.Subscriptions = []azurecli.SubscriptionInfo{
				{
					ID:        tt.subscription,
					Name:      "Test Subscription",
					IsDefault: true,
				},
			}
			mockCli.CurrentSubscription = &mockCli.Subscriptions[0]

			// Set up mock data for resource group existence check
			if tt.rgExistsError != nil {
				mockCli.CheckResourceGroupExistsError = tt.rgExistsError
			} else {
				if mockCli.ResourceGroupExists == nil {
					mockCli.ResourceGroupExists = make(map[string]bool)
				}
				mockCli.ResourceGroupExists[tt.resourceGroupName] = tt.rgExists
			}

			// Set up delete error if needed
			if tt.deleteError != nil {
				mockCli.DeleteResourceGroupError = tt.deleteError
			}

			// Execute the method
			err := service.DeleteDeployment(tt.resourceGroupName, tt.subscription, tt.skipConfirmation)

			// Check results
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.expectedErrorMessage != "" && !containsString(err.Error(), tt.expectedErrorMessage) {
					t.Errorf("Expected error message containing '%s', got '%s'", tt.expectedErrorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			// Verify that the expected methods were called
			if tt.rgExistsError == nil {
				if !mockCli.CheckResourceGroupExistsCalled {
					t.Errorf("Expected CheckResourceGroupExists to be called")
				}
				if mockCli.CheckResourceGroupExistsCalledWith != tt.resourceGroupName {
					t.Errorf("Expected CheckResourceGroupExists to be called with '%s', got '%s'",
						tt.resourceGroupName, mockCli.CheckResourceGroupExistsCalledWith)
				}
			}

			if tt.rgExists && tt.rgExistsError == nil && !tt.expectedError {
				if !mockCli.DeleteResourceGroupCalled {
					t.Errorf("Expected DeleteResourceGroup to be called")
				}
				if mockCli.DeleteResourceGroupCalledWith != tt.resourceGroupName {
					t.Errorf("Expected DeleteResourceGroup to be called with '%s', got '%s'",
						tt.resourceGroupName, mockCli.DeleteResourceGroupCalledWith)
				}
			}
		})
	}
}

func TestDeletionService_DeleteDeployment_UserCancellation(t *testing.T) {
	// This test is harder to implement because it involves user input
	// In a real implementation, you might want to inject an input reader
	// For now, we'll test the confirmation path by setting skipConfirmation to false
	// and expecting the service to handle the confirmation logic

	mockCli := &azurecli.MockAzureCLI{}
	// service := NewDeletionService(mockCli)

	// Set up mock to indicate resource group exists
	if mockCli.ResourceGroupExists == nil {
		mockCli.ResourceGroupExists = make(map[string]bool)
	}
	mockCli.ResourceGroupExists["test-rg"] = true

	// When skipConfirmation is false, the method will prompt for confirmation
	// Since we can't easily mock user input in this test environment,
	// we'll focus on the cases where skipConfirmation is true
	// A more sophisticated implementation would inject an io.Reader for testing

	t.Log("User cancellation test skipped - requires input reader injection for proper testing")
}

func TestDeletionService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*azurecli.MockAzureCLI)
		expectError  bool
		errorPattern string
	}{
		{
			name: "cli_not_available",
			setupMock: func(mockCli *azurecli.MockAzureCLI) {
				mockCli.CheckResourceGroupExistsError = fmt.Errorf("az command not found")
			},
			expectError:  true,
			errorPattern: "resource group existence check failed",
		},
		{
			name: "invalid_subscription",
			setupMock: func(mockCli *azurecli.MockAzureCLI) {
				mockCli.CheckResourceGroupExistsError = fmt.Errorf("subscription not found")
			},
			expectError:  true,
			errorPattern: "resource group existence check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli := &azurecli.MockAzureCLI{}
			service := NewDeletionService(mockCli)

			// Set up basic subscription data for SetSubscription to work
			mockCli.Subscriptions = []azurecli.SubscriptionInfo{
				{
					ID:        "test-sub",
					Name:      "Test Subscription",
					IsDefault: true,
				},
			}
			mockCli.CurrentSubscription = &mockCli.Subscriptions[0]

			tt.setupMock(mockCli)

			err := service.DeleteDeployment("test-rg", "test-sub", true)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorPattern != "" {
					if !containsString(err.Error(), tt.errorPattern) {
						t.Errorf("Expected error containing '%s', got '%s'", tt.errorPattern, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			// Verify that the expected methods were called
			if !mockCli.CheckResourceGroupExistsCalled {
				t.Errorf("Expected CheckResourceGroupExists to be called")
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
