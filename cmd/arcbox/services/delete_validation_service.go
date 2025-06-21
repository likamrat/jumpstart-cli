package services

import (
	"fmt"
	"os"

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
	// Use standard missing required flags validation - this prints the error message
	requiredFlags := []string{"name"}
	if !utils.PrintMissingRequiredFlagsError(cmd, requiredFlags) {
		// PrintMissingRequiredFlagsError already printed the error message using centralized handling
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("missing required arguments"),
		}
	}

	return ValidationResult{IsValid: true}
}

// ValidateAzureLogin validates that the user is logged in to Azure CLI
func (dvs *DeleteValidationService) ValidateAzureLogin() ValidationResult {
	if !utils.IsAzureLoggedInWithCLI(dvs.cli) {
		utils.PrintAuthenticationError()
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("authentication required"),
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
		fmt.Fprintf(os.Stderr, "The subscription '%s' doesn't exist or you don't have access.\n", subscription)
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("subscription validation failed"),
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
