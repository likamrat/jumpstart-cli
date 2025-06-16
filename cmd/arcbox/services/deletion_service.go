package services

import (
	"fmt"
	"strings"

	arcboxUtils "jumpstartcli/cmd/arcbox/utils"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"
)

// DeletionService handles deletion functionality for ArcBox deployments
type DeletionService struct {
	cli azurecli.AzureCLI
}

// NewDeletionService creates a new DeletionService with the provided Azure CLI
func NewDeletionService(cli azurecli.AzureCLI) *DeletionService {
	return &DeletionService{
		cli: cli,
	}
}

// DeleteDeployment deletes an ArcBox deployment by deleting its resource group
func (d *DeletionService) DeleteDeployment(resourceGroupName, subscription string, skipConfirmation bool) error {
	// Check if resource group exists
	rgExists, err := arcboxUtils.CheckResourceGroupExists(d.cli, resourceGroupName, subscription)
	if err != nil {
		return fmt.Errorf("resource group existence check failed for '%s': please verify Azure CLI authentication and subscription access: %w", resourceGroupName, err)
	}

	if !rgExists {
		return fmt.Errorf("resource group '%s' does not exist or is not accessible: please check the resource group name and your Azure permissions", resourceGroupName)
	}

	// Confirmation prompt (unless --yes is specified)
	if !skipConfirmation {
		fmt.Printf(utils.WarnColor("⚠️  WARNING: This will permanently delete resource group '%s' and all its resources.\n"), resourceGroupName)
		fmt.Print("Are you sure you want to continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion cancelled by user.")
			return nil // This is not an error, user chose to cancel
		}
	}

	// Perform deletion
	fmt.Printf(utils.InfoColor("[INFO] Deleting ArcBox resource group '%s'...\n"), resourceGroupName)

	// Use Azure CLI wrapper to delete the resource group
	err = d.cli.DeleteResourceGroup(resourceGroupName, true)
	if err != nil {
		return fmt.Errorf("resource group deletion failed for '%s': check the Azure Portal for more details: %w", resourceGroupName, err)
	}

	fmt.Printf(utils.SuccessColor("✅ Successfully initiated deletion of resource group '%s'.\n"), resourceGroupName)
	fmt.Println(utils.InfoColor("[INFO] Deletion is running in the background. Check the Azure Portal to monitor progress."))

	// Provide Azure Portal link for monitoring
	if subscription != "" {
		portalUrl := fmt.Sprintf("https://portal.azure.com/#view/HubsExtension/BrowseResource/resourceType/Microsoft.Resources%%2Fresourcegroups")
		fmt.Printf(utils.InfoColor("🔗 [INFO] Monitor deletion progress in the Azure Portal: %s\n"), portalUrl)
	}

	return nil
}
