package services

import (
	"fmt"
	"os"
	"testing"

	"jumpstartcli/cmd/arcbox/models"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// MockListingService for testing
type MockListingService struct {
	shouldFailGetSubscription bool
}

func (m *MockListingService) GetSubscription(subscriptionID string) (models.AzureSubscription, error) {
	if m.shouldFailGetSubscription {
		return models.AzureSubscription{}, fmt.Errorf("subscription not found")
	}
	return models.AzureSubscription{
		ID:   subscriptionID,
		Name: "Test Subscription",
	}, nil
}

func TestListValidationService_ValidateAzureLogin(t *testing.T) {
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
			expectedError: "azure authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: tt.loggedIn}
			service := NewListValidationService(mockCli)

			result := service.ValidateAzureLogin()

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestListValidationService_ValidateSubscriptionSelection(t *testing.T) {
	tests := []struct {
		name          string
		setupCmd      func() *cobra.Command
		expectedValid bool
		expectedError string
	}{
		{
			name: "All subscriptions flag set",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				cmd.Flags().Set("all-subscriptions", "true")

				return cmd
			},
			expectedValid: true,
		},
		{
			name: "Current subscription flag set",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				cmd.Flags().Set("current-subscription", "true")

				return cmd
			},
			expectedValid: true,
		},
		{
			name: "Specific subscription flag set",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				cmd.Flags().Set("subscription", "test-sub-id")

				return cmd
			},
			expectedValid: true,
		},
		{
			name: "No subscription flags set",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				// No flags set

				return cmd
			},
			expectedValid: false,
			expectedError: "subscription selection required",
		},
		{
			name: "Multiple subscription flags set",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				cmd.Flags().Set("all-subscriptions", "true")
				cmd.Flags().Set("current-subscription", "true")

				return cmd
			},
			expectedValid: false,
			expectedError: "conflicting subscription flags",
		},
		{
			name: "All flags and specific subscription set",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				cmd.Flags().Set("all-subscriptions", "true")
				cmd.Flags().Set("subscription", "test-sub-id")

				return cmd
			},
			expectedValid: false,
			expectedError: "conflicting subscription flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
			service := NewListValidationService(mockCli)

			cmd := tt.setupCmd()
			result := service.ValidateSubscriptionSelection(cmd)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestListValidationService_ValidateOutputFormat(t *testing.T) {
	tests := []struct {
		name          string
		outputFormat  string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Valid table format",
			outputFormat:  "table",
			expectedValid: true,
		},
		{
			name:          "Valid json format",
			outputFormat:  "json",
			expectedValid: true,
		},
		{
			name:          "Valid yaml format",
			outputFormat:  "yaml",
			expectedValid: true,
		},
		{
			name:          "Valid tsv format",
			outputFormat:  "tsv",
			expectedValid: true,
		},
		{
			name:          "Invalid format",
			outputFormat:  "invalid",
			expectedValid: false,
			expectedError: "invalid output format 'invalid'",
		},
		{
			name:          "Empty format is invalid",
			outputFormat:  "",
			expectedValid: false,
			expectedError: "invalid output format ''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original output format
			originalFormat := utils.OutputFormat
			defer func() { utils.OutputFormat = originalFormat }()

			// Set test output format
			utils.OutputFormat = tt.outputFormat

			mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
			service := NewListValidationService(mockCli)

			result := service.ValidateOutputFormat()

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestListValidationService_ValidateSubscriptionAccess(t *testing.T) {
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
			shouldFail:    false,
			expectedValid: true,
		},
		{
			name:          "Valid subscription access",
			subscription:  "valid-sub-id",
			shouldFail:    false,
			expectedValid: true,
		},
		{
			name:          "Invalid subscription access",
			subscription:  "invalid-sub-id",
			shouldFail:    true,
			expectedValid: false,
			expectedError: "subscription access failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
			service := NewListValidationService(mockCli)

			mockListingService := &MockListingService{
				shouldFailGetSubscription: tt.shouldFail,
			}

			result := service.ValidateSubscriptionAccess(mockListingService, tt.subscription)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestListValidationService_ValidateAllListRequirements(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func() (*cobra.Command, *MockListingService, *azurecli.MockAzureCLI)
		expectedValid bool
		expectedError string
	}{
		{
			name: "All validations pass",
			setupTest: func() (*cobra.Command, *MockListingService, *azurecli.MockAzureCLI) {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				cmd.Flags().Set("current-subscription", "true")

				mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
				mockListingService := &MockListingService{shouldFailGetSubscription: false}

				// Set valid output format
				utils.OutputFormat = "table"

				return cmd, mockListingService, mockCli
			},
			expectedValid: true,
		},
		{
			name: "Azure login fails",
			setupTest: func() (*cobra.Command, *MockListingService, *azurecli.MockAzureCLI) {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				cmd.Flags().Set("current-subscription", "true")

				mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: false} // Not logged in
				mockListingService := &MockListingService{shouldFailGetSubscription: false}

				utils.OutputFormat = "table"

				return cmd, mockListingService, mockCli
			},
			expectedValid: false,
			expectedError: "azure authentication required",
		},
		{
			name: "Subscription selection fails",
			setupTest: func() (*cobra.Command, *MockListingService, *azurecli.MockAzureCLI) {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				// No subscription flags set

				mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
				mockListingService := &MockListingService{shouldFailGetSubscription: false}

				utils.OutputFormat = "table"

				return cmd, mockListingService, mockCli
			},
			expectedValid: false,
			expectedError: "subscription selection required",
		},
		{
			name: "Output format fails",
			setupTest: func() (*cobra.Command, *MockListingService, *azurecli.MockAzureCLI) {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				cmd.Flags().Set("current-subscription", "true")

				mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
				mockListingService := &MockListingService{shouldFailGetSubscription: false}

				utils.OutputFormat = "invalid-format"

				return cmd, mockListingService, mockCli
			},
			expectedValid: false,
			expectedError: "invalid output format 'invalid-format'",
		},
		{
			name: "Subscription access fails",
			setupTest: func() (*cobra.Command, *MockListingService, *azurecli.MockAzureCLI) {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("all-subscriptions", false, "All subscriptions")
				cmd.Flags().Bool("current-subscription", false, "Current subscription")
				cmd.Flags().String("subscription", "", "Subscription")

				cmd.Flags().Set("subscription", "invalid-sub-id")

				mockCli := &azurecli.MockAzureCLI{IsLoggedInResult: true}
				mockListingService := &MockListingService{shouldFailGetSubscription: true}

				utils.OutputFormat = "table"

				return cmd, mockListingService, mockCli
			},
			expectedValid: false,
			expectedError: "subscription access failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original output format
			originalFormat := utils.OutputFormat
			defer func() { utils.OutputFormat = originalFormat }()

			// Redirect stdout/stderr to avoid cluttering test output
			originalStdout := os.Stdout
			originalStderr := os.Stderr
			defer func() {
				os.Stdout = originalStdout
				os.Stderr = originalStderr
			}()

			cmd, mockListingService, mockCli := tt.setupTest()
			service := NewListValidationService(mockCli)

			result := service.ValidateAllListRequirements(cmd, mockListingService)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}
