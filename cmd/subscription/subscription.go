package subscription

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"jumpstartcli/internal/table"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// SubscriptionInfo represents subscription details
type SubscriptionInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewSubscriptionCmd creates the subscription command
func NewSubscriptionCmd() *cobra.Command {
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
			idOnly, _ := cmd.Flags().GetBool("id")
			nameOnly, _ := cmd.Flags().GetBool("name")

			// Check for mutually exclusive flags
			if idOnly && nameOnly {
				utils.Error("Cannot specify both --id and --name flags. Please use only one.")
				return
			}

			out, err := exec.Command("az", "account", "show", "--output", "json").Output()
			if err != nil {
				utils.Error("Could not get current subscription. Please ensure you are logged in with 'az login'.")
				utils.Debug("Azure CLI error: %v", err)
				return
			}

			var sub struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				TenantID  string `json:"tenantId"`
				State     string `json:"state"`
				IsDefault bool   `json:"isDefault"`
				User      *struct {
					Name string `json:"name"`
					Type string `json:"type"`
				} `json:"user,omitempty"`
			}
			if err := json.Unmarshal(out, &sub); err != nil {
				utils.Error("Failed to parse subscription information.")
				utils.Debug("JSON parsing error: %v", err)
				return
			}

			if idOnly {
				fmt.Println(sub.ID)
				return
			}
			if nameOnly {
				fmt.Println(sub.Name)
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
				fmt.Println(string(prettyJSON))

			case "yaml":
				fmt.Printf("id: %s\n", sub.ID)
				fmt.Printf("name: %s\n", sub.Name)
				fmt.Printf("tenantId: %s\n", sub.TenantID)
				fmt.Printf("state: %s\n", sub.State)
				fmt.Printf("isDefault: %t\n", sub.IsDefault)
				if utils.VerboseMode && sub.User != nil {
					fmt.Printf("user:\n")
					fmt.Printf("  name: %s\n", sub.User.Name)
					fmt.Printf("  type: %s\n", sub.User.Type)
				}

			case "tsv":
				if utils.VerboseMode {
					fmt.Printf("%s\t%s\t%s\t%s\t%t", sub.ID, sub.Name, sub.TenantID, sub.State, sub.IsDefault)
					if sub.User != nil {
						fmt.Printf("\t%s\t%s", sub.User.Name, sub.User.Type)
					}
					fmt.Println()
				} else {
					fmt.Printf("%s\t%s\n", sub.ID, sub.Name)
				}

			case "table":
				fallthrough
			default:
				if utils.VerboseMode {
					headers := []string{"Property", "Value"}
					rows := [][]string{
						{"Subscription ID", sub.ID},
						{"Subscription Name", sub.Name},
						{"Tenant ID", sub.TenantID},
						{"State", sub.State},
						{"Is Default", fmt.Sprintf("%t", sub.IsDefault)},
					}
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

					if sub.State != "Enabled" {
						utils.Warn("Subscription state: %s", sub.State)
					}
				}
			}
		},
	}

	var subscriptionListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available Azure subscriptions",
		Run: func(cmd *cobra.Command, args []string) {
			if utils.DebugMode {
				utils.Debug("Current output format: '%s'", utils.OutputFormat)
			}

			out, err := exec.Command("az", "account", "list", "--output", "json").Output()
			if err != nil {
				utils.Error("Could not list subscriptions. Are you logged in with 'az login'?")
				utils.Debug("Azure CLI error: %v", err)
				return
			}

			var subs []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				IsDefault bool   `json:"isDefault"`
			}
			if err := json.Unmarshal(out, &subs); err != nil {
				utils.Error("Failed to parse subscriptions: %v", err)
				return
			}

			// Use global output formatting
			switch strings.ToLower(utils.OutputFormat) {
			case "json":
				if err := utils.PrintJSON(subs); err != nil {
					utils.Error("Failed to output JSON: %v", err)
				}
			case "yaml":
				if err := utils.PrintYAML(subs); err != nil {
					utils.Error("Failed to output YAML: %v", err)
				}
			case "tsv":
				headers := []string{"Name", "Subscription ID", "State"}
				rows := [][]string{}
				for _, sub := range subs {
					marker := ""
					if sub.IsDefault {
						marker = "*"
					}
					rows = append(rows, []string{sub.Name, sub.ID, marker})
				}
				utils.PrintTSV(headers, rows)
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
				fmt.Fprintln(cmd.ErrOrStderr(), utils.ErrorColor("[ERROR] Cannot specify multiple subscription selection methods. Use only one of: --subscription/-s, --name/-n, or positional argument."))
				return
			} else if flagCount == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), utils.ErrorColor("[ERROR] Must specify subscription using --subscription/-s, --name/-n, or as a positional argument."))
				utils.ShowHelpWithoutTypes(cmd)
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
				fmt.Fprintln(cmd.ErrOrStderr(), utils.ErrorColor("[ERROR] Invalid subscription ID format. Expected GUID format (e.g., 12345678-1234-1234-1234-123456789012)"))
				return
			}

			sub, err := validateSubscriptionAccess(targetSubscription)
			if err != nil {
				utils.Error("Subscription validation failed: %v", err)
				utils.Info("Make sure the subscription exists and you have access to it.")
				return
			}

			utils.Success("Subscription validation passed: %s (%s)", sub.Name, sub.ID)

			targetSubscription = sub.ID

			currentSub, err := getCurrentSubscriptionSafe()
			if err == nil && currentSub.ID == targetSubscription {
				utils.Warn("Subscription '%s' is already the current subscription.", currentSub.Name)
				return
			}

			utils.Info("Setting Azure subscription...")
			cmdOut := exec.Command("az", "account", "set", "--subscription", targetSubscription)
			if err := cmdOut.Run(); err != nil {
				utils.Error("Failed to set subscription: %v", err)
				return
			}

			newSub, err := getCurrentSubscriptionSafe()
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

	if guid[8] != '-' || guid[13] != '-' || guid[18] != '-' || guid[23] != '-' {
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

// validateSubscriptionAccess validates access to a subscription
func validateSubscriptionAccess(subscription string) (SubscriptionInfo, error) {
	// Validate input
	if subscription == "" {
		return SubscriptionInfo{}, fmt.Errorf("subscription cannot be empty")
	}

	cmd := exec.Command("az", "account", "show", "--subscription", subscription, "--query", "{id:id,name:name}", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return SubscriptionInfo{}, fmt.Errorf("subscription '%s' not found or inaccessible", subscription)
	}

	var sub SubscriptionInfo
	if err := json.Unmarshal(output, &sub); err != nil {
		return SubscriptionInfo{}, fmt.Errorf("failed to parse subscription information")
	}

	return sub, nil
}

// getCurrentSubscriptionSafe gets current subscription safely
func getCurrentSubscriptionSafe() (SubscriptionInfo, error) {
	cmd := exec.Command("az", "account", "show", "--query", "{id:id,name:name}", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return SubscriptionInfo{}, fmt.Errorf("failed to get current subscription")
	}

	var sub SubscriptionInfo
	if err := json.Unmarshal(output, &sub); err != nil {
		return SubscriptionInfo{}, fmt.Errorf("failed to parse current subscription")
	}

	return sub, nil
}
