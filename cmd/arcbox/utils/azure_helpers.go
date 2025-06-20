package utils

import (
	"fmt"
	"os"
	"strings"

	"jumpstartcli/internal/auth"
	"jumpstartcli/internal/azurecli"

	"github.com/spf13/cobra"
)

// This file contains Azure-specific helper functions for ArcBox operations,
// including Azure CLI interactions, resource management, and deployment helpers

// GetSubscriptionID gets the subscription ID from command flags or default using Azure CLI wrapper
func GetSubscriptionID(cmd *cobra.Command, azCLI azurecli.AzureCLI) string {
	// Check Azure CLI authentication first
	if err := auth.CheckAzureAuthentication(azCLI); err != nil {
		return ""
	}

	// Try to get from command flag first
	subscription, _ := cmd.Flags().GetString("subscription")
	if subscription != "" {
		return subscription
	}

	// Try to get from environment variable
	if env := os.Getenv("AZURE_SUBSCRIPTION_ID"); env != "" {
		return env
	}

	// Fallback: use current Azure CLI subscription
	currentSub, err := azCLI.GetCurrentSubscription()
	if err == nil && currentSub != nil {
		return currentSub.ID
	}

	return ""
}

// SetAzureSubscription sets the Azure subscription using Azure CLI wrapper
func SetAzureSubscription(azCLI azurecli.AzureCLI, subscriptionID string) error {
	// Check Azure CLI authentication first
	if err := auth.CheckAzureAuthentication(azCLI); err != nil {
		return err
	}

	if subscriptionID == "" {
		return fmt.Errorf("subscription ID is empty")
	}
	return azCLI.SetSubscription(subscriptionID)
}

// CheckResourceGroupExists checks if a resource group exists using Azure CLI wrapper
func CheckResourceGroupExists(azCLI azurecli.AzureCLI, resourceGroupName, subscriptionID string) (bool, error) {
	// Check Azure CLI authentication first
	if err := auth.CheckAzureAuthentication(azCLI); err != nil {
		return false, err
	}

	if subscriptionID != "" {
		if err := SetAzureSubscription(azCLI, subscriptionID); err != nil {
			return false, fmt.Errorf("failed to set subscription context: %v", err)
		}
	}

	return azCLI.CheckResourceGroupExists(resourceGroupName)
}

// GetRequiredVCPUForSKU returns the number of vCPUs required for a given VM SKU
func GetRequiredVCPUForSKU(sku string) int {
	vcpuMap := map[string]int{
		"Standard_D8s_v5": 8,
		"Standard_D8s_v4": 8,
		"Standard_B2ms":   2,
		"Standard_B4ms":   4,
		"Standard_B8ms":   8,
	}
	if vcpu, ok := vcpuMap[sku]; ok {
		return vcpu
	}
	return 1 // Default fallback
}

// MapSKUToFamilyQuotaName maps a VM SKU to its Azure vCPU family quota name
func MapSKUToFamilyQuotaName(sku string) string {
	sku = strings.TrimPrefix(sku, "Standard_")
	parts := strings.Split(sku, "_")
	if len(parts) < 2 {
		return ""
	}

	main := parts[0]
	ver := parts[1]

	// Remove digits from main part, keep only letters
	letters := ""
	for _, r := range main {
		if r >= '0' && r <= '9' {
			continue
		}
		letters += string(r)
	}

	// If ends with 's', keep it (e.g. D8s → Ds)
	if strings.HasSuffix(main, "s") && !strings.HasSuffix(letters, "s") {
		letters += "s"
	}

	family := "Standard " + strings.ToUpper(letters[:1]) + letters[1:] + ver + " Family vCPUs"
	return family
}
