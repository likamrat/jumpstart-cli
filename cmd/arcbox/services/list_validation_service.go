package services

import (
	"fmt"
	"os"

	"jumpstartcli/cmd/arcbox/models"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// ListValidationService handles all validation logic for list commands
type ListValidationService struct {
	cli azurecli.AzureCLI
}

// NewListValidationService creates a new list validation service
func NewListValidationService(cli azurecli.AzureCLI) *ListValidationService {
	return &ListValidationService{
		cli: cli,
	}
}

// ValidateAzureLogin validates that the user is logged in to Azure CLI
func (lvs *ListValidationService) ValidateAzureLogin() ValidationResult {
	if !utils.IsAzureLoggedInWithCLI(lvs.cli) {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("azure authentication required: please run 'az login' and try again"),
		}
	}
	return ValidationResult{IsValid: true}
}

// ValidateSubscriptionSelection validates that exactly one subscription selection flag is provided
func (lvs *ListValidationService) ValidateSubscriptionSelection(cmd *cobra.Command) ValidationResult {
	allSubscriptions, _ := cmd.Flags().GetBool("all-subscriptions")
	currentSubscription, _ := cmd.Flags().GetBool("current-subscription")
	subscription, _ := cmd.Flags().GetString("subscription")

	// Count subscription selection flags
	flagCount := 0
	if allSubscriptions {
		flagCount++
	}
	if currentSubscription {
		flagCount++
	}
	if subscription != "" {
		flagCount++
	}

	// If no subscription selection flag is provided
	if flagCount == 0 {
		// Print the standardized error message format with red color
		fmt.Fprintf(os.Stderr, "%s\n\n", utils.ErrorColor("the following arguments are required (choose one): --current-subscription, --all-subscriptions, or --subscription"))
		utils.ShowHelpWithoutTypes(cmd)
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("validation_failed_with_output_already_shown"),
		}
	}

	// Validate flag combinations - prevent contradictory flags
	if flagCount > 1 {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("conflicting subscription flags: choose only one of --all-subscriptions, --current-subscription, or --subscription"),
		}
	}

	return ValidationResult{IsValid: true}
}

// ValidateOutputFormat validates the output format
func (lvs *ListValidationService) ValidateOutputFormat() ValidationResult {
	if !utils.ValidateOutputFormat(utils.OutputFormat) {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("invalid output format '%s': supported formats are table, json, yaml, tsv", utils.OutputFormat),
		}
	}
	return ValidationResult{IsValid: true}
}

// ListingServiceInterface defines the interface needed for subscription access validation
type ListingServiceInterface interface {
	GetSubscription(subscriptionID string) (models.AzureSubscription, error)
}

// ValidateSubscriptionAccess validates that the specified subscription is accessible
func (lvs *ListValidationService) ValidateSubscriptionAccess(listingService ListingServiceInterface, subscription string) ValidationResult {
	if subscription == "" {
		return ValidationResult{IsValid: true} // No specific subscription to validate
	}

	if _, err := listingService.GetSubscription(subscription); err != nil {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("subscription access failed for '%s': verify subscription ID and permissions: %w", subscription, err),
		}
	}
	return ValidationResult{IsValid: true}
}

// ValidateAllListRequirements runs all validation steps for list command
func (lvs *ListValidationService) ValidateAllListRequirements(cmd *cobra.Command, listingService ListingServiceInterface) ValidationResult {
	// Step 1: Validate Azure CLI login
	if result := lvs.ValidateAzureLogin(); !result.IsValid {
		return result
	}

	// Step 2: Validate subscription selection flags
	if result := lvs.ValidateSubscriptionSelection(cmd); !result.IsValid {
		return result
	}

	// Step 3: Validate output format
	if result := lvs.ValidateOutputFormat(); !result.IsValid {
		return result
	}

	// Step 4: Validate subscription access if specific subscription is provided
	subscription, _ := cmd.Flags().GetString("subscription")
	if result := lvs.ValidateSubscriptionAccess(listingService, subscription); !result.IsValid {
		return result
	}

	return ValidationResult{IsValid: true}
}
