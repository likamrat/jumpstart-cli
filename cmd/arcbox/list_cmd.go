package arcbox

import (
	"fmt"
	"os"

	"jumpstartcli/cmd/arcbox/services"
	"jumpstartcli/internal/auth"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// executeListCommand contains the core list command logic, extracted for testability
func executeListCommand(cmd *cobra.Command, listingService *services.ListingService, cli azurecli.AzureCLI) error {
	// Check Azure CLI authentication first
	if err := auth.CheckAzureAuthentication(cli); err != nil {
		return err
	}

	// Create validation service
	validationService := services.NewListValidationService(cli)

	// Run all validations
	result := validationService.ValidateAllListRequirements(cmd, listingService)
	if !result.IsValid {
		// For validation errors, the validation service already showed
		// the clean Azure CLI-style error message, so we just exit
		os.Exit(1)
	}

	// Extract command flags
	allSubscriptions, _ := cmd.Flags().GetBool("all-subscriptions")
	currentSubscription, _ := cmd.Flags().GetBool("current-subscription")
	subscription, _ := cmd.Flags().GetString("subscription")

	// Use the listing service to handle the list operation
	if err := listingService.ListDeployments(allSubscriptions, currentSubscription, subscription, utils.OutputFormat); err != nil {
		return err
	}

	return nil
}

// handleListCommandError handles errors from executeListCommand, including display and exit
func handleListCommandError(err error, cmd *cobra.Command, listingService *services.ListingService, cli azurecli.AzureCLI) {
	// Use standard error format without ERROR prefix for consistency
	fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
	os.Exit(1)
}

// createListCommand creates the list command with the provided listing service
func createListCommand(listingService *services.ListingService, cli azurecli.AzureCLI) *cobra.Command {
	var arcboxListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Jumpstart ArcBox deployments",
		Long: `Discovers ArcBox deployments by identifying resource groups containing resources with:
- Solution tag "jumpstart_arcbox" (default identification method)
- ArcBox naming prefix (configurable, default: "ArcBox")
- Specific ArcBox resource types (VMs, Key Vaults, etc.)

Requires explicit subscription selection: --current-subscription, --all-subscriptions, or --subscription <id>.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Execute the core command logic
			if err := executeListCommand(cmd, listingService, cli); err != nil {
				handleListCommandError(err, cmd, listingService, cli)
			}
		},
	}

	// Add all the list command flags
	arcboxListCmd.Flags().Bool("all-subscriptions", false, "Search for ArcBox deployments across all accessible subscriptions")
	arcboxListCmd.Flags().Bool("current-subscription", false, "Search for ArcBox deployments in the current subscription")
	arcboxListCmd.Flags().StringP("subscription", "s", "", "Azure subscription ID to search")

	return arcboxListCmd
}
