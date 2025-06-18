package arcbox

import (
	"fmt"

	"jumpstartcli/cmd/arcbox/services"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			subscription, _ := cmd.Flags().GetString("subscription")
			skipConfirmation, _ := cmd.Flags().GetBool("yes")
			resourceGroupName, _ := cmd.Flags().GetString("name")

			// Create validation service
			validationService := services.NewDeleteValidationService(cli)

			// Perform all validation
			if validationResult := validationService.ValidateAllDeleteRequirements(cmd, subscription); !validationResult.IsValid {
				fmt.Printf(utils.ErrorColor("❌ [ERROR] %v\n"), validationResult.Error)
				utils.ShowHelpWithoutTypes(cmd)
				return validationResult.Error
			}

			// Use the deletion service to handle the deletion process
			if err := deletionService.DeleteDeployment(resourceGroupName, subscription, skipConfirmation); err != nil {
				fmt.Printf(utils.ErrorColor("❌ [ERROR] %v\n"), err)
				return err
			}
			return nil
		},
	}

	// Add all the delete command flags
	arcboxDeleteCmd.Flags().StringP("name", "n", "", "Resource group name containing the ArcBox deployment to delete")
	arcboxDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt and proceed with deletion")
	arcboxDeleteCmd.Flags().StringP("subscription", "s", "", "Azure subscription ID to use")

	return arcboxDeleteCmd
}
