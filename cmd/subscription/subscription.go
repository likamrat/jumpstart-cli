package subscription

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"jumpstartcli/internal/auth"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/table"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// Default Azure CLI instance - can be overridden for testing
var defaultAzureCLI azurecli.AzureCLI = azurecli.NewAzureCLI()

// SetAzureCLI allows overriding the Azure CLI implementation for testing
func SetAzureCLI(cli azurecli.AzureCLI) {
	defaultAzureCLI = cli
}

// NewSubscriptionCmd creates the subscription command
func NewSubscriptionCmd() *cobra.Command {
	return NewSubscriptionCmdWithCLI(defaultAzureCLI)
}

// NewSubscriptionCmdWithCLI creates the subscription command with a specific Azure CLI implementation
func NewSubscriptionCmdWithCLI(azCLI azurecli.AzureCLI) *cobra.Command {
	var subscriptionCmd = &cobra.Command{
		Use:   "subscription",
		Short: "Manage Azure subscriptions",
		Long: `Manage Azure subscriptions (show, list, set).

Subcommands:
  • show     Show the current Azure subscription details
  • list     List all available Azure subscriptions  
  • set      Set the current Azure subscription

Use 'js subscription <subcommand> --help' for more details.`,
		DisableSuggestions: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				validSubcommands := []string{"list", "set", "show"}

				invalidSubcommand := args[0]
				for _, validCmd := range validSubcommands {
					if invalidSubcommand == validCmd {
						return nil
					}
				}

				if suggestion := utils.SuggestSimilarCommand(invalidSubcommand, validSubcommands, 3); suggestion != "" {
					utils.PrintDidYouMean(invalidSubcommand, suggestion)
					return nil
				}

				return fmt.Errorf("unknown subcommand '%s' for 'js subscription'", invalidSubcommand)
			}
			utils.ShowHelpWithoutTypes(cmd)
			return nil
		},
	}

	var subscriptionShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show the current Azure subscription details",
		Long: `Show details about the current Azure subscription including subscription ID, name, and tenant information.

The command displays the subscription that is currently set as the default for Azure CLI operations.
Use different output formats to integrate with scripts or automation tools.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Check Azure CLI authentication first
			if err := auth.CheckAzureAuthentication(azCLI); err != nil {
				utils.Error(err.Error())
				return
			}

			idOnly, _ := cmd.Flags().GetBool("id")
			nameOnly, _ := cmd.Flags().GetBool("name")

			// Check for mutually exclusive flags
			if idOnly && nameOnly {
				utils.HandleValidationError(
					fmt.Errorf("argument --name: not allowed with argument --id"),
					cmd,
					false,
				)
				return
			}

			sub, err := azCLI.GetCurrentSubscription()
			if err != nil {
				utils.Error("Could not get current subscription. Please ensure you are logged in with 'az login'.")
				utils.Debug("Azure CLI error: %v", err)
				return
			}

			if idOnly {
				fmt.Fprintln(cmd.OutOrStdout(), sub.ID)
				return
			}
			if nameOnly {
				fmt.Fprintln(cmd.OutOrStdout(), sub.Name)
				return
			}

			// Default behavior - show subscription details
			switch strings.ToLower(utils.OutputFormat) {
			case "json":
				prettyJSON, err := json.MarshalIndent(sub, "", "  ")
				if err != nil {
					utils.Error("Failed to format JSON output.")
					return
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(prettyJSON))

			case "yaml":
				fmt.Fprintf(cmd.OutOrStdout(), "id: %s\n", sub.ID)
				fmt.Fprintf(cmd.OutOrStdout(), "name: %s\n", sub.Name)
				if sub.TenantID != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "tenantId: %s\n", sub.TenantID)
				}
				if sub.State != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "state: %s\n", sub.State)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "isDefault: %t\n", sub.IsDefault)
				if utils.VerboseMode && sub.User != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "user:\n")
					fmt.Fprintf(cmd.OutOrStdout(), "  name: %s\n", sub.User.Name)
					fmt.Fprintf(cmd.OutOrStdout(), "  type: %s\n", sub.User.Type)
				}

			case "tsv":
				if utils.VerboseMode {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s", sub.ID, sub.Name)
					if sub.TenantID != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "\t%s", sub.TenantID)
					}
					if sub.State != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "\t%s", sub.State)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "\t%t", sub.IsDefault)
					if sub.User != nil {
						fmt.Fprintf(cmd.OutOrStdout(), "\t%s\t%s", sub.User.Name, sub.User.Type)
					}
					fmt.Fprintln(cmd.OutOrStdout())
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", sub.ID, sub.Name)
				}

			case "table":
				fallthrough
			default:
				if utils.VerboseMode {
					headers := []string{"Property", "Value"}
					rows := [][]string{
						{"Subscription ID", sub.ID},
						{"Subscription Name", sub.Name},
					}
					if sub.TenantID != "" {
						rows = append(rows, []string{"Tenant ID", sub.TenantID})
					}
					if sub.State != "" {
						rows = append(rows, []string{"State", sub.State})
					}
					rows = append(rows, []string{"Is Default", fmt.Sprintf("%t", sub.IsDefault)})
					if sub.User != nil {
						rows = append(rows, []string{"User Name", sub.User.Name})
						rows = append(rows, []string{"User Type", sub.User.Type})
					}
					table.PrintASCIITable(headers, rows)
				} else {
					defaultIndicator := ""
					if sub.IsDefault {
						defaultIndicator = " " + utils.SuccessColor("(default)")
					}
					utils.Success("Current subscription: %s (%s)%s", sub.Name, sub.ID, defaultIndicator)

					if sub.State != "" && sub.State != "Enabled" {
						utils.Warn("Subscription state: %s", sub.State)
					}
				}
			}
		},
	}

	var subscriptionListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available Azure subscriptions",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				utils.HandleValidationError(
					fmt.Errorf("unrecognized arguments: %s", strings.Join(args, " ")),
					cmd,
					false,
				)
				os.Exit(1)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Check Azure CLI authentication first
			if err := auth.CheckAzureAuthentication(azCLI); err != nil {
				utils.Error(err.Error())
				return
			}

			if utils.DebugMode {
				utils.Debug("Current output format: '%s'", utils.OutputFormat)
			}

			subs, err := azCLI.ListSubscriptions()
			if err != nil {
				utils.Error("Could not list subscriptions. Are you logged in with 'az login'?")
				utils.Debug("Azure CLI error: %v", err)
				return
			}

			// Use global output formatting
			switch strings.ToLower(utils.OutputFormat) {
			case "json":
				jsonData, err := json.MarshalIndent(subs, "", "  ")
				if err != nil {
					utils.Error("Failed to output JSON: %v", err)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
				}
			case "yaml":
				yamlData, err := yaml.Marshal(subs)
				if err != nil {
					utils.Error("Failed to output YAML: %v", err)
				} else {
					fmt.Fprint(cmd.OutOrStdout(), string(yamlData))
				}
			case "tsv":
				fmt.Fprintf(cmd.OutOrStdout(), "Name\tSubscription ID\tState\n")
				for _, sub := range subs {
					marker := ""
					if sub.IsDefault {
						marker = "*"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", sub.Name, sub.ID, marker)
				}
			case "table":
				fallthrough
			default:
				headers := []string{"Name", "Subscription ID", "State"}
				rows := [][]string{}
				for _, sub := range subs {
					marker := ""
					if sub.IsDefault {
						marker = "*"
					}
					rows = append(rows, []string{sub.Name, sub.ID, marker})
				}
				table.PrintASCIITable(headers, rows)
				fmt.Println("* = current/default subscription")
			}
		},
	}

	var subscriptionSetCmd = &cobra.Command{
		Use:   "set",
		Short: "Set the current Azure subscription",
		Long: `Set the current Azure subscription using either subscription ID or subscription name.

You can specify the subscription using either:
  • --subscription/-s flag with subscription ID
  • --name/-n flag with subscription name
  • positional argument (for backward compatibility)

The subscription will be set as the default for all subsequent Azure CLI commands.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Check Azure CLI authentication first
			if err := auth.CheckAzureAuthentication(azCLI); err != nil {
				utils.Error(err.Error())
				return
			}

			subscriptionID, _ := cmd.Flags().GetString("subscription")
			subscriptionName, _ := cmd.Flags().GetString("name")

			var targetSubscription string
			var isSubscriptionID bool

			flagCount := 0
			if subscriptionID != "" {
				flagCount++
			}
			if subscriptionName != "" {
				flagCount++
			}
			if len(args) > 0 {
				flagCount++
			}

			if flagCount > 1 {
				// Multiple selection methods specified - this is a validation error, not missing arguments
				errorMsg := "Cannot specify multiple subscription selection methods. Use only one of: --subscription/-s, --name/-n, or positional argument."
				utils.HandleSubscriptionSelectionError(cmd, errorMsg)
				return
			} else if flagCount == 0 {
				// Use the centralized error handling for missing required arguments
				// Azure CLI format: show mutual exclusion options with forward slashes
				// This matches Azure CLI: "az account set" shows "--name --subscription -n -s [Required]"
				missingArgs := []string{"--subscription/-s/--name/-n"}
				utils.HandleMissingRequiredArguments(cmd, missingArgs)
				return
			}

			if subscriptionID != "" {
				targetSubscription = subscriptionID
				isSubscriptionID = true
			} else if subscriptionName != "" {
				targetSubscription = subscriptionName
				isSubscriptionID = false
			} else if len(args) == 1 {
				targetSubscription = args[0]
				isSubscriptionID = isValidGUID(targetSubscription)
			}

			utils.Info("Validating subscription access...")

			if isSubscriptionID && !isValidGUID(targetSubscription) {
				utils.HandleValidationError(
					fmt.Errorf("invalid subscription ID format. Expected GUID format (e.g., 12345678-1234-1234-1234-123456789012)"),
					cmd,
					false,
				)
				return
			}

			sub, err := validateSubscriptionAccessWithCLI(azCLI, targetSubscription)
			if err != nil {
				utils.Error("Subscription validation failed: %v", err)
				utils.Info("Make sure the subscription exists and you have access to it.")
				return
			}

			utils.Success("Subscription validation passed: %s (%s)", sub.Name, sub.ID)

			targetSubscription = sub.ID

			currentSub, err := azCLI.GetCurrentSubscription()
			if err == nil && currentSub.ID == targetSubscription {
				utils.Warn("Subscription '%s' is already the current subscription.", currentSub.Name)
				return
			}

			utils.Info("Setting Azure subscription...")
			if err := azCLI.SetSubscription(targetSubscription); err != nil {
				utils.Error("Failed to set subscription: %v", err)
				return
			}

			newSub, err := azCLI.GetCurrentSubscription()
			if err != nil {
				utils.Warn("Subscription was set, but verification failed: %v", err)
			} else {
				utils.Success("Successfully set subscription to '%s' (%s)", newSub.Name, newSub.ID)
				return
			}

			utils.Success("Subscription set to '%s'", targetSubscription)
		},
	}

	subscriptionCmd.AddCommand(subscriptionListCmd)
	subscriptionCmd.AddCommand(subscriptionSetCmd)
	subscriptionCmd.AddCommand(subscriptionShowCmd)

	subscriptionSetCmd.Flags().StringP("subscription", "s", "", "Azure subscription ID to set as current")
	subscriptionSetCmd.Flags().StringP("name", "n", "", "Azure subscription name to set as current")

	subscriptionShowCmd.Flags().Bool("id", false, "Show only the subscription ID")
	subscriptionShowCmd.Flags().Bool("name", false, "Show only the subscription name")

	return subscriptionCmd
}

// isValidGUID validates if a string is a valid GUID format
func isValidGUID(guid string) bool {
	if len(guid) != 36 {
		return false
	}

	guidWithoutDashes := strings.ReplaceAll(guid, "-", "")
	if len(guidWithoutDashes) != 32 {
		return false
	}

	for _, char := range guidWithoutDashes {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}

// validateSubscriptionAccessWithCLI validates access to a subscription using the provided Azure CLI
func validateSubscriptionAccessWithCLI(azCLI azurecli.AzureCLI, subscription string) (*azurecli.SubscriptionInfo, error) {
	// Validate input
	if subscription == "" {
		return nil, fmt.Errorf("subscription cannot be empty")
	}

	sub, err := azCLI.GetSubscription(subscription)
	if err != nil {
		return nil, fmt.Errorf("subscription '%s' not found or inaccessible", subscription)
	}

	return sub, nil
}

// Legacy functions for backward compatibility - these use the default Azure CLI instance

// validateSubscriptionAccess validates access to a subscription using the default Azure CLI
func validateSubscriptionAccess(subscription string) (SubscriptionInfo, error) {
	sub, err := validateSubscriptionAccessWithCLI(defaultAzureCLI, subscription)
	if err != nil {
		return SubscriptionInfo{}, err
	}

	// Convert to legacy format
	return SubscriptionInfo{
		ID:   sub.ID,
		Name: sub.Name,
	}, nil
}

// getCurrentSubscriptionSafe gets current subscription safely using the default Azure CLI
func getCurrentSubscriptionSafe() (SubscriptionInfo, error) {
	sub, err := defaultAzureCLI.GetCurrentSubscription()
	if err != nil {
		return SubscriptionInfo{}, fmt.Errorf("failed to get current subscription")
	}

	// Convert to legacy format
	return SubscriptionInfo{
		ID:   sub.ID,
		Name: sub.Name,
	}, nil
}

// Legacy SubscriptionInfo type for backward compatibility
type SubscriptionInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
