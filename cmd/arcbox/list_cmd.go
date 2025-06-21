package arcbox

import (
	"fmt"
	"os"
	"strings"

	"jumpstartcli/cmd/arcbox/services"
	"jumpstartcli/internal/auth"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/examples"
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
		errorMessage := result.Error.Error()

		// Special case: if validation failed but output was already shown, exit silently
		if errorMessage == "validation_failed_with_output_already_shown" {
			utils.PrintMissingRequiredArgumentsTip(cmd)
			os.Exit(1)
		}

		return result.Error
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
	// For subscription selection errors, don't add ERROR prefix - keep it clean
	errorMessage := err.Error()
	if strings.Contains(errorMessage, "subscription selection required") || strings.Contains(errorMessage, "conflicting subscription flags") {
		fmt.Fprintln(cmd.ErrOrStderr(), errorMessage)
		utils.ShowHelpWithoutTypes(cmd)
		fmt.Fprintln(cmd.ErrOrStderr(), utils.InfoColor("💡 [TIP] Use 'js arcbox list --help' to see all required arguments."))
	} else {
		// For other errors, use the standard error format
		fmt.Printf(utils.ErrorColor("❌ [ERROR] %v\n"), err)
		if result := services.NewListValidationService(cli).ValidateAllListRequirements(cmd, listingService); !result.IsValid {
			utils.ShowHelpWithoutTypes(cmd)
		}
	}
	os.Exit(1)
}

// createListCommand creates the list command with the provided listing service
func createListCommand(listingService *services.ListingService, cli azurecli.AzureCLI) *cobra.Command {
	var arcboxListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Jumpstart ArcBox deployments",
		Long: `List all Jumpstart ArcBox deployments across your Azure subscriptions.

Discovers ArcBox deployments by identifying resource groups containing resources with:
- Solution tag "jumpstart_arcbox" (default identification method)
- ArcBox naming prefix (configurable, default: "ArcBox")
- Specific ArcBox resource types (VMs, Key Vaults, etc.)

Requires explicit subscription selection: --current-subscription, --all-subscriptions, or --subscription <id>.

` + examples.GetExamples("js.arcbox.list").FormatExamples(),
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
