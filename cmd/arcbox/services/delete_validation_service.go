package services

import (
	"fmt"
	"strings"

	arcboxUtils "jumpstartcli/cmd/arcbox/utils"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// DeleteValidationService handles all validation logic for delete commands
type DeleteValidationService struct {
	cli azurecli.AzureCLI
}

// NewDeleteValidationService creates a new delete validation service
func NewDeleteValidationService(cli azurecli.AzureCLI) *DeleteValidationService {
	return &DeleteValidationService{
		cli: cli,
	}
}

// ValidateRequiredArguments validates that all required arguments are provided
func (dvs *DeleteValidationService) ValidateRequiredArguments(cmd *cobra.Command) ValidationResult {
	if err := utils.ValidateAllFlags(cmd); err != nil {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("failed to validate command flags: %w", err),
		}
	}

	// Check if the name flag is provided and not empty
	resourceGroupName, _ := cmd.Flags().GetString("name")
	if strings.TrimSpace(resourceGroupName) == "" {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("required argument missing: --name flag must specify a resource group name"),
		}
	}

	return ValidationResult{IsValid: true}
}

// ValidateAzureLogin validates that the user is logged in to Azure CLI
func (dvs *DeleteValidationService) ValidateAzureLogin() ValidationResult {
	if !utils.IsAzureLoggedInWithCLI(dvs.cli) {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("azure authentication required: please run 'az login' and try again"),
		}
	}
	return ValidationResult{IsValid: true}
}

// ValidateAndSetSubscription validates and sets the Azure subscription if provided
func (dvs *DeleteValidationService) ValidateAndSetSubscription(subscription string) ValidationResult {
	if subscription == "" {
		return ValidationResult{IsValid: true} // No subscription specified, use current
	}

	if err := arcboxUtils.SetAzureSubscription(dvs.cli, subscription); err != nil {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("subscription context setup failed for '%s': verify subscription ID and access permissions: %w", subscription, err),
		}
	}
	return ValidationResult{IsValid: true}
}

// ValidateAllDeleteRequirements runs all validation steps for delete command
func (dvs *DeleteValidationService) ValidateAllDeleteRequirements(cmd *cobra.Command, subscription string) ValidationResult {
	// Step 1: Validate required arguments
	if result := dvs.ValidateRequiredArguments(cmd); !result.IsValid {
		return result
	}

	// Step 2: Validate Azure CLI login
	if result := dvs.ValidateAzureLogin(); !result.IsValid {
		return result
	}

	// Step 3: Validate and set subscription if provided
	if result := dvs.ValidateAndSetSubscription(subscription); !result.IsValid {
		return result
	}

	return ValidationResult{IsValid: true}
}
