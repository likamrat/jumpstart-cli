package services

import (
	"testing"

	"jumpstartcli/internal/azurecli"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDeleteValidationService_ValidateRequiredArguments(t *testing.T) {
	tests := []struct {
		name          string
		setupCmd      func() *cobra.Command
		expectedValid bool
		expectedError string
	}{
		{
			name: "Required name argument provided",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().StringP("name", "n", "", "Resource group name")
				cmd.Flags().Set("name", "arcbox-rg")
				return cmd
			},
			expectedValid: true,
		},
		{
			name: "Missing name argument",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().StringP("name", "n", "", "Resource group name")
				// Don't set the name flag
				return cmd
			},
			expectedValid: false,
			expectedError: "required argument missing: --name flag must specify a resource group name",
		},
		{
			name: "Empty name argument",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().StringP("name", "n", "", "Resource group name")
				cmd.Flags().Set("name", "")
				return cmd
			},
			expectedValid: false,
			expectedError: "required argument missing: --name flag must specify a resource group name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
			service := NewDeleteValidationService(mockCli)

			cmd := tt.setupCmd()
			result := service.ValidateRequiredArguments(cmd)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestDeleteValidationService_ValidateAzureLogin(t *testing.T) {
	tests := []struct {
		name          string
		loggedIn      bool
		expectedValid bool
		expectedError string
	}{
		{
			name:          "User logged in",
			loggedIn:      true,
			expectedValid: true,
		},
		{
			name:          "User not logged in",
			loggedIn:      false,
			expectedValid: false,
			expectedError: "azure authentication required: please run 'az login' and try again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: tt.loggedIn}
			service := NewDeleteValidationService(mockCli)

			result := service.ValidateAzureLogin()

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestDeleteValidationService_ValidateAndSetSubscription(t *testing.T) {
	tests := []struct {
		name          string
		subscription  string
		shouldFail    bool
		expectedValid bool
		expectedError string
	}{
		{
			name:          "No subscription specified",
			subscription:  "",
			expectedValid: true,
		},
		{
			name:          "Valid subscription set successfully",
			subscription:  "test-subscription-id",
			shouldFail:    false,
			expectedValid: true,
		},
		{
			name:          "Subscription setting fails",
			subscription:  "invalid-subscription-id",
			shouldFail:    true,
			expectedValid: false,
			expectedError: "subscription context setup failed for 'invalid-subscription-id': verify subscription ID and access permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli := &azurecli.MockAzureCLI{
				IsLoggedInResult:     true,
				SetSubscriptionError: nil,
				Subscriptions: []azurecli.SubscriptionInfo{
					{ID: "test-subscription-id", Name: "Test Subscription", IsDefault: true},
				},
			}

			if tt.shouldFail {
				mockCli.SetSubscriptionError = assert.AnError
			}

			service := NewDeleteValidationService(mockCli)
			result := service.ValidateAndSetSubscription(tt.subscription)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestDeleteValidationService_ValidateAllDeleteRequirements(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func() (*cobra.Command, string, *azurecli.MockAzureCLI)
		expectedValid bool
		expectedError string
	}{
		{
			name: "All validations pass",
			setupTest: func() (*cobra.Command, string, *azurecli.MockAzureCLI) {
				cmd := &cobra.Command{}
				cmd.Flags().StringP("name", "n", "", "Resource group name")
				cmd.Flags().Set("name", "arcbox-rg")

				mockCli := &azurecli.MockAzureCLI{
					IsLoggedInResult: true,
					Subscriptions: []azurecli.SubscriptionInfo{
						{ID: "test-subscription", Name: "Test Subscription", IsDefault: true},
					},
				}
				subscription := "test-subscription"

				return cmd, subscription, mockCli
			},
			expectedValid: true,
		},
		{
			name: "Required arguments validation fails",
			setupTest: func() (*cobra.Command, string, *azurecli.MockAzureCLI) {
				cmd := &cobra.Command{}
				cmd.Flags().StringP("name", "n", "", "Resource group name")
				// Don't set name flag

				mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
				subscription := ""

				return cmd, subscription, mockCli
			},
			expectedValid: false,
			expectedError: "required argument missing: --name flag must specify a resource group name",
		},
		{
			name: "Azure login validation fails",
			setupTest: func() (*cobra.Command, string, *azurecli.MockAzureCLI) {
				cmd := &cobra.Command{}
				cmd.Flags().StringP("name", "n", "", "Resource group name")
				cmd.Flags().Set("name", "arcbox-rg")

				mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: false} // Not logged in
				subscription := ""

				return cmd, subscription, mockCli
			},
			expectedValid: false,
			expectedError: "azure authentication required: please run 'az login' and try again",
		},
		{
			name: "Subscription setting fails",
			setupTest: func() (*cobra.Command, string, *azurecli.MockAzureCLI) {
				cmd := &cobra.Command{}
				cmd.Flags().StringP("name", "n", "", "Resource group name")
				cmd.Flags().Set("name", "arcbox-rg")

				mockCli := &azurecli.MockAzureCLI{
					IsLoggedInResult:     true,
					SetSubscriptionError: assert.AnError,
				}
				subscription := "invalid-subscription"

				return cmd, subscription, mockCli
			},
			expectedValid: false,
			expectedError: "subscription context setup failed for 'invalid-subscription': verify subscription ID and access permissions",
		},
		{
			name: "No subscription specified - should pass",
			setupTest: func() (*cobra.Command, string, *azurecli.MockAzureCLI) {
				cmd := &cobra.Command{}
				cmd.Flags().StringP("name", "n", "", "Resource group name")
				cmd.Flags().Set("name", "arcbox-rg")

				mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
				subscription := "" // No subscription

				return cmd, subscription, mockCli
			},
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, subscription, mockCli := tt.setupTest()
			service := NewDeleteValidationService(mockCli)

			result := service.ValidateAllDeleteRequirements(cmd, subscription)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestDeleteValidationService_Creation(t *testing.T) {
	mockCli := &azurecli.MockAzureCLI{}
	service := NewDeleteValidationService(mockCli)

	assert.NotNil(t, service)
	assert.Equal(t, mockCli, service.cli)
}

func TestDeleteValidationService_ArgumentValidation_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		setupCmd      func() *cobra.Command
		expectedValid bool
	}{
		{
			name: "Name with whitespace only",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().StringP("name", "n", "", "Resource group name")
				cmd.Flags().Set("name", "   ")
				return cmd
			},
			expectedValid: false,
		},
		{
			name: "Valid name with leading/trailing spaces",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().StringP("name", "n", "", "Resource group name")
				cmd.Flags().Set("name", "  arcbox-rg  ")
				return cmd
			},
			expectedValid: true, // The validation currently just checks for empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
			service := NewDeleteValidationService(mockCli)

			cmd := tt.setupCmd()
			result := service.ValidateRequiredArguments(cmd)

			assert.Equal(t, tt.expectedValid, result.IsValid)
		})
	}
}
