// arcbox.go - ArcBox command and deployment logic for Jumpstart CLI
package arcbox

import (
	"fmt"

	"jumpstartcli/cmd/arcbox/display"
	"jumpstartcli/cmd/arcbox/services"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// Azure CLI instance for dependency injection
var defaultAzureCLI azurecli.AzureCLI = &azurecli.RealAzureCLI{}

// SetAzureCLI allows overriding the Azure CLI implementation for testing
func SetAzureCLI(cli azurecli.AzureCLI) {
	defaultAzureCLI = cli
}

// NewArcboxCmd creates the main arcbox command with default Azure CLI
func NewArcboxCmd() *cobra.Command {
	return NewArcboxCmdWithCLI(defaultAzureCLI)
}

// NewArcboxCmdWithCLI creates the arcbox command with injectable Azure CLI for testing
func NewArcboxCmdWithCLI(cli azurecli.AzureCLI) *cobra.Command {
	// Create services
	deployDisplay := display.NewDeploymentDisplay(cli)
	deploymentService := services.NewDeploymentService(cli, deployDisplay)
	deletionService := services.NewDeletionService(cli)
	listingService := services.NewListingService(cli)
	quotaService := services.NewQuotaService(cli)
	validationService := services.NewValidationService(cli)

	// Create main command
	arcboxCmd := &cobra.Command{
		Use:   "arcbox",
		Short: "Manage Jumpstart ArcBox automation",
		Long: `Manage Jumpstart ArcBox automation resources.

Subcommands:
  • deploy     Deploy a new Jumpstart ArcBox deployment
  • delete     Delete a Jumpstart ArcBox deployment
  • list       List all Jumpstart ArcBox deployments
  • preflight  Run preflight checks for ArcBox deployment

Use 'js arcbox <subcommand> --help' for more details.`,
		// Disable suggestions to use our custom handling
		DisableSuggestions: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		// Use RunE instead of Args for better control over suggestion handling
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// Check if the first argument matches any subcommand
				validCommands := []string{"deploy", "delete", "list", "preflight"}
				invalidCommand := args[0]

				for _, validCmd := range validCommands {
					if invalidCommand == validCmd {
						return nil // Valid command, continue normal processing
					}
				}

				// Special case for "create" -> "deploy" since they're conceptually the same
				if invalidCommand == "create" {
					utils.PrintDidYouMean(invalidCommand, "deploy")
					return nil
				}

				// General similarity checking with higher threshold
				if suggestion := utils.SuggestSimilarCommand(invalidCommand, validCommands, 3); suggestion != "" {
					utils.PrintDidYouMean(invalidCommand, suggestion)
					return nil
				}

				// No suggestion found, show normal error
				return fmt.Errorf("unknown subcommand '%s' for 'js arcbox'", invalidCommand)
			}
			// If no args, show help
			utils.ShowHelpWithoutTypes(cmd)
			return nil
		},
	}

	// Add subcommands using extracted functions
	arcboxCmd.AddCommand(createDeployCommand(deploymentService, validationService, cli))
	arcboxCmd.AddCommand(createDeleteCommand(deletionService, cli))
	arcboxCmd.AddCommand(createListCommand(listingService, cli))
	arcboxCmd.AddCommand(createPreflightCommand(quotaService, validationService, cli))

	return arcboxCmd
}
