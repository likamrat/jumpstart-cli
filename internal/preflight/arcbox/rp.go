// rp.go - ArcBox-specific resource provider checking functionality
package arcbox

import (
	"fmt"
	"os"

	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/resourceproviders"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// CreateResourceProviderCommands creates the resource provider subcommands for ArcBox preflight
func CreateResourceProviderCommands(cli azurecli.AzureCLI) *cobra.Command {
	// Main rp command
	var rpCmd = &cobra.Command{
		Use:   "rp",
		Short: "Check and manage Azure resource provider registration",
		Long:  `Check, list, and register required Azure resource providers for ArcBox deployment`,
		// Disable Cobra's built-in suggestions and errors to use our custom ones
		DisableSuggestions: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// List of valid subcommands for arcbox preflight rp
				validSubcommands := []string{"show", "list", "register"}

				// Check if the provided argument is a valid subcommand
				invalidSubcommand := args[0]
				for _, validCmd := range validSubcommands {
					if invalidSubcommand == validCmd {
						return nil // Valid subcommand, continue normal processing
					}
				}

				// If we reach here, it's an invalid subcommand - suggest similar ones
				if suggestion := utils.SuggestSimilarCommand(invalidSubcommand, validSubcommands, 3); suggestion != "" {
					utils.PrintDidYouMean(invalidSubcommand, suggestion)
					return nil
				}

				// No suggestion found, show normal error
				return fmt.Errorf("unknown subcommand '%s' for 'js arcbox preflight rp'", invalidSubcommand)
			}
			// If no args, show help
			utils.ShowHelpWithoutTypes(cmd)
			return nil
		},
	}

	// rp show: check registration status
	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show registration status of required Azure resource providers",
		Long:  `Check and display the registration status of all required Azure resource providers for ArcBox deployment.`,
		Run: func(cmd *cobra.Command, args []string) {
			ShowResourceProviderStatus(cli)
		},
	}

	// rp list: list required providers
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List required Azure resource providers for ArcBox",
		Long:  `List the Azure resource providers required for ArcBox deployment (names only).`,
		Run: func(cmd *cobra.Command, args []string) {
			ListRequiredResourceProviders()
		},
	}

	// rp register: register a provider
	var registerCmd = &cobra.Command{
		Use:   "register",
		Short: "Register a required Azure resource provider",
		Long:  `Register a required Azure resource provider for ArcBox deployment.`,
		Run: func(cmd *cobra.Command, args []string) {
			provider, _ := cmd.Flags().GetString("name")

			// Check if required argument is missing using centralized error handling
			if provider == "" {
				utils.PrintRequiredArgumentsError([]string{"--name/-n"})
				os.Exit(1)
			}

			if err := RegisterResourceProvider(cli, provider); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	registerCmd.Flags().StringP("name", "n", "", "Azure resource provider name to register")

	rpCmd.AddCommand(listCmd)
	rpCmd.AddCommand(registerCmd)
	rpCmd.AddCommand(showCmd)

	return rpCmd
}

// ShowResourceProviderStatus checks and displays the registration status of all required Azure resource providers
func ShowResourceProviderStatus(cli azurecli.AzureCLI) {
	config := resourceproviders.GetArcBoxProviders()
	resourceproviders.CheckAllProviders(cli, config)
}

// ListRequiredResourceProviders lists the Azure resource providers required for ArcBox deployment
func ListRequiredResourceProviders() {
	config := resourceproviders.GetArcBoxProviders()
	resourceproviders.ListProviders(config)
}

// RegisterResourceProvider registers a required Azure resource provider for ArcBox deployment
func RegisterResourceProvider(cli azurecli.AzureCLI, provider string) error {
	if err := resourceproviders.RegisterProvider(cli, provider); err != nil {
		return fmt.Errorf("failed to register resource provider '%s': %w", provider, err)
	}
	fmt.Printf(utils.SuccessColor("✅ [SUCCESS] Successfully registered resource provider '%s'\n"), provider)
	return nil
}

// CheckAllResourceProviders performs comprehensive resource provider validation
func CheckAllResourceProviders(cli azurecli.AzureCLI) bool {
	config := resourceproviders.GetArcBoxProviders()
	allRegistered, missingProviders := resourceproviders.CheckAllProviders(cli, config)

	if !allRegistered {
		fmt.Fprintf(os.Stderr, "%s\n", utils.ErrorColor(fmt.Sprintf("%d resource provider(s) not registered: %v", len(missingProviders), missingProviders)))
		return false
	}

	fmt.Println(utils.SuccessColor("✅ [SUCCESS] All required resource providers are registered"))
	return true
}
