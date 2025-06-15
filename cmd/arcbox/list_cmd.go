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
			// Validate Azure CLI is logged in
			if !utils.IsAzureLoggedInWithCLI(cli) {
				utils.Error("You are not logged in to Azure. Please run 'az login' and try again.")
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			allSubscriptions, _ := cmd.Flags().GetBool("all-subscriptions")
			currentSubscription, _ := cmd.Flags().GetBool("current-subscription")
			subscription, _ := cmd.Flags().GetString("subscription")

			// Check if any subscription selection flag is provided
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

			// If no subscription selection flag is provided, show help (like arcbox deploy does for required args)
			if flagCount == 0 {
				fmt.Fprintf(os.Stderr, "%s\n\n", utils.ErrorColor("please specify a subscription selection flag: --current-subscription, --all-subscriptions, or --subscription <id>"))
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			// Validate flag combinations - prevent contradictory flags
			if flagCount > 1 {
				utils.Error("Cannot use multiple subscription selection flags together. Choose one of: --all-subscriptions, --current-subscription, or --subscription")
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			// Validate output format
			if !utils.ValidateOutputFormat(utils.OutputFormat) {
				utils.Error("Invalid output format '%s'. Valid formats are: table, json, yaml, tsv", utils.OutputFormat)
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			// Validate subscription access if specific subscription is provided
			if subscription != "" {
				if _, err := listingService.GetSubscription(subscription); err != nil {
					utils.Error("Cannot access subscription '%s'. Please verify the subscription ID and your permissions.", subscription)
					os.Exit(1)
				}
			}

			// Use the listing service to handle the list operation
			if err := listingService.ListDeployments(allSubscriptions, currentSubscription, subscription, utils.OutputFormat); err != nil {
				utils.Error("Failed to list ArcBox deployments: %v", err)
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
