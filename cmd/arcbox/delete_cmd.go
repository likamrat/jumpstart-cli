package arcbox

import (
	"os"

	"jumpstartcli/cmd/arcbox/services"
	arcboxUtils "jumpstartcli/cmd/arcbox/utils"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/examples"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// createDeleteCommand creates the delete command with the provided deletion service
func createDeleteCommand(deletionService *services.DeletionService, cli azurecli.AzureCLI) *cobra.Command {
	var arcboxDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a Jumpstart ArcBox deployment",
		Long: `Delete an existing Jumpstart ArcBox deployment by deleting its resource group.

This command will delete the specified resource group and all resources within it.
Use --name to specify the resource group containing your ArcBox deployment.
This operation is irreversible and will permanently remove all ArcBox resources.

` + examples.GetExamples("arcbox.delete").FormatExamples(),
		Run: func(cmd *cobra.Command, args []string) {
			requiredArguments := []string{"name"}
			utils.PrintMissingRequiredArgumentsError(cmd, requiredArguments)

			resourceGroupName, _ := cmd.Flags().GetString("name")
			skipConfirmation, _ := cmd.Flags().GetBool("yes")
			subscription, _ := cmd.Flags().GetString("subscription")

			// Validate Azure CLI is logged in
			if !utils.IsAzureLoggedInWithCLI(cli) {
				utils.Error("You are not logged in to Azure. Please run 'az login' and try again.")
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			// Set Azure subscription if provided
			if subscription != "" {
				if err := arcboxUtils.SetAzureSubscription(cli, subscription); err != nil {
					utils.Error("Failed to set subscription: %v", err)
					os.Exit(1)
				}
			}

			// Use the deletion service to handle the deletion process
			if err := deletionService.DeleteDeployment(resourceGroupName, subscription, skipConfirmation); err != nil {
				utils.Error("Deletion failed: %v", err)
				os.Exit(1)
			}
		},
	}

	// Add all the delete command flags
	arcboxDeleteCmd.Flags().StringP("name", "n", "", "Resource group name containing the ArcBox deployment to delete")
	arcboxDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt and proceed with deletion")
	arcboxDeleteCmd.Flags().StringP("subscription", "s", "", "Azure subscription ID to use")

	return arcboxDeleteCmd
}
