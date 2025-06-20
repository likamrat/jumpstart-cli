package auth

import (
	"fmt"

	"jumpstartcli/internal/azurecli"
)

// CheckAzureAuthentication validates that the user is logged in to Azure CLI.
// It returns nil if the user is authenticated, or an error with a helpful message if not.
func CheckAzureAuthentication(azCLI azurecli.AzureCLI) error {
	if !azCLI.IsLoggedIn() {
		return fmt.Errorf("Azure CLI authentication required. Please run 'az login' to setup your account")
	}
	return nil
}
