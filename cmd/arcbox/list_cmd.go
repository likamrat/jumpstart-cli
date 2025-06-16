package arcbox

import (
	"fmt"
	"os"

	"jumpstartcli/cmd/arcbox/services"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/examples"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

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

` + examples.GetExamples("arcbox.list").FormatExamples(),
		Run: func(cmd *cobra.Command, args []string) {
			// Create validation service
			validationService := services.NewListValidationService(cli)

			// Run all validations
			result := validationService.ValidateAllListRequirements(cmd, listingService)
			if !result.IsValid {
				fmt.Printf(utils.ErrorColor("❌ [ERROR] %v\n"), result.Error)
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			// Extract command flags
			allSubscriptions, _ := cmd.Flags().GetBool("all-subscriptions")
			currentSubscription, _ := cmd.Flags().GetBool("current-subscription")
			subscription, _ := cmd.Flags().GetString("subscription")

			// Use the listing service to handle the list operation
			if err := listingService.ListDeployments(allSubscriptions, currentSubscription, subscription, utils.OutputFormat); err != nil {
				fmt.Printf(utils.ErrorColor("❌ [ERROR] %v\n"), err)
				os.Exit(1)
			}
		},
	}

	// Add all the list command flags
	arcboxListCmd.Flags().Bool("all-subscriptions", false, "Search for ArcBox deployments across all accessible subscriptions")
	arcboxListCmd.Flags().Bool("current-subscription", false, "Search for ArcBox deployments in the current subscription")
	arcboxListCmd.Flags().StringP("subscription", "s", "", "Azure subscription ID to search")

	return arcboxListCmd
}
