package arcbox

import (
	"fmt"

	"jumpstartcli/cmd/arcbox/services"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/examples"
	"jumpstartcli/internal/preflight/arcbox"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// createPreflightCommand creates the preflight command with all its subcommands
func createPreflightCommand(quotaService *services.QuotaService, cli azurecli.AzureCLI) *cobra.Command {
	var arcboxPreflightCmd = &cobra.Command{
		Use:   "preflight",
		Short: "Run preflight checks for ArcBox deployment",
		Long:  `Run preflight checks to ensure your Azure environment is ready for ArcBox deployment.`,
		// Disable Cobra's built-in suggestions and errors to use our custom ones
		DisableSuggestions: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// List of valid subcommands for arcbox preflight
				validSubcommands := []string{"quota", "status", "rp"}

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
				return fmt.Errorf("unknown subcommand '%s' for 'js arcbox preflight'", invalidSubcommand)
			}
			// If no args, show help
			utils.ShowHelpWithoutTypes(cmd)
			return nil
		},
	}

	// arcbox preflight quota
	var arcboxPreflightQuotaCmd = &cobra.Command{
		Use:   "quota",
		Short: "Check vCPU quota for ArcBox flavors",
		Long: `Check if your Azure subscription and region have sufficient vCPU quota for ArcBox ITPro, DevOps, and DataOps flavors.

` + examples.GetExamples("arcbox.preflight.quota").FormatExamples(),
		Run: func(cmd *cobra.Command, args []string) {
			// Use the quota service to run the command
			quotaService.RunQuotaCheckCommand(cmd, args)
		},
	}
	arcboxPreflightQuotaCmd.Flags().StringP("flavor", "f", "", "ArcBox flavor to check (ITPro, DevOps, DataOps, all)")
	arcboxPreflightQuotaCmd.Flags().StringP("location", "l", "", "Azure region(s) to check quota in. Use comma-separated values for multiple regions")
	arcboxPreflightQuotaCmd.Flags().Bool("all-locations", false, "Check quota in all ArcBox-supported regions")
	arcboxPreflightQuotaCmd.Flags().String("sku", "", "Custom VM SKU(s) to check. Comma-separated")
	arcboxPreflightQuotaCmd.Flags().StringP("subscription", "s", "", "Azure subscription ID to use")
	arcboxPreflightCmd.AddCommand(arcboxPreflightQuotaCmd)

	// arcbox preflight rp (using dedicated module)
	arcboxPreflightRPCmd := arcbox.CreateResourceProviderCommands(cli)
	arcboxPreflightCmd.AddCommand(arcboxPreflightRPCmd)

	// arcbox preflight status (using dedicated module)
	arcboxPreflightStatusCmd := arcbox.CreateStatusCommand()
	arcboxPreflightCmd.AddCommand(arcboxPreflightStatusCmd)

	return arcboxPreflightCmd
}
