package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"jumpstartcli/cmd/agora"
	"jumpstartcli/cmd/arcbox"
	"jumpstartcli/cmd/completion"
	"jumpstartcli/cmd/localbox"
	"jumpstartcli/cmd/repo"
	"jumpstartcli/cmd/subscription"
	"jumpstartcli/cmd/upgrade"
	"jumpstartcli/cmd/version"
	"jumpstartcli/internal/utils"
)

func main() {
	// Set debugMode from utils
	var rootCmd = &cobra.Command{
		Use:     "js",
		Short:   "Jumpstart CLI",
		Long:    `Jumpstart CLI - Azure Arc Jumpstart automation tool.`,
		Version: utils.CliVersion,
		// Disable Cobra's built-in suggestions and errors to use our custom ones
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableSuggestions: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				invalidCommand := args[0]
				validCommands := []string{"agora", "arcbox", "completion", "localbox", "repo", "subscription", "upgrade", "version"}

				if suggestion := utils.SuggestSimilarCommand(invalidCommand, validCommands, 2); suggestion != "" {
					utils.PrintDidYouMean(invalidCommand, suggestion)
					return nil
				}
			}
			// If no suggestion or no args, show welcome message
			printWelcome()
			return nil
		},
	}

	// Add persistent arguments (will show as "Global Arguments" in help)
	rootCmd.PersistentFlags().BoolVar(&utils.DebugMode, "debug", false, "Enable debug output. Show detailed information for troubleshooting")
	rootCmd.PersistentFlags().BoolVar(&utils.VerboseMode, "verbose", false, "Enable verbose output. Show detailed information about operations")
	rootCmd.PersistentFlags().StringVarP(&utils.OutputFormat, "output", "o", "table", "Output format: table, json, yaml, tsv")

	// Add commands in alphabetical order
	rootCmd.AddCommand(agora.NewAgoraCmd())               // agora
	rootCmd.AddCommand(arcbox.NewArcboxCmd())             // ArcBox parent command
	rootCmd.AddCommand(completion.NewCompletionCmd())     // completion
	rootCmd.AddCommand(localbox.NewLocalboxCmd())         // localbox
	rootCmd.AddCommand(subscription.NewSubscriptionCmd()) // subscription command

	// Add remaining commands in alphabetical order
	rootCmd.AddCommand(repo.NewRepoCmd())
	rootCmd.AddCommand(upgrade.NewUpgradeCmd())
	rootCmd.AddCommand(version.NewVersionCmd())

	// Add help as a persistent argument to match Azure CLI behavior
	// This will make it appear in "Global Arguments" section
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Show help message and exit. Display command usage information")

	// Note: Cobra automatically provides --version/-v flag when Version is set on root command
	// No need to manually add version flag

	// Set custom version template to match our subcommand format
	rootCmd.SetVersionTemplate("Jumpstart CLI version: {{.Version}}\n")

	// Hide the default help command
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Use consistent help function that removes type annotations
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		utils.ShowHelpWithoutTypes(cmd)
	})

	if err := rootCmd.Execute(); err != nil {
		// Check if it's an "unknown command" error and try to suggest
		errorStr := err.Error()
		if strings.Contains(errorStr, "unknown command") {
			// Extract the invalid command from error message
			// Error format is usually: "unknown command \"invalidcmd\" for \"js\""
			if strings.Contains(errorStr, "\"") {
				parts := strings.Split(errorStr, "\"")
				if len(parts) >= 2 {
					invalidCommand := parts[1]
					validCommands := []string{"arcbox", "agora", "localbox", "subscription", "repo", "version", "completion", "upgrade"}

					if suggestion := utils.SuggestSimilarCommand(invalidCommand, validCommands, 2); suggestion != "" {
						utils.PrintDidYouMean(invalidCommand, suggestion)
						return
					}
				}
			}
		}
		// For other errors, show them normally
		if !strings.Contains(errorStr, "unknown command") {
			utils.Error("%v", err)
		}
		os.Exit(1)
	}
}

// printWelcome displays the welcome message with ASCII art and basic usage information
func printWelcome() {
	ascii := utils.InfoColor(`
       _                           _             _   
      | |                         | |           | |  
      | |_   _ _ __ ___  _ __  ___| |_ __ _ _ __| |_ 
  _   | | | | | '_ ' _ \| '_ \/ __| __/ _' | '__| __|
 | |__| | |_| | | | | | | |_) \__ \ || (_| | |  | |_ 
  \____/ \__,_|_| |_| |_| .__/|___/\__\__,_|_|   \__|
                        | |                          
                        |_|                          
`)
	fmt.Println("\n" + ascii)
	fmt.Println(utils.InfoColor("Use `js --help` to see available commands or visit https://github.com/Azure/jumpstart-sdk."))
	fmt.Println(utils.WarnColor("\nNote: The Jumpstart CLI is based on the Azure CLI and requires Azure CLI to be installed and available in your PATH."))
	fmt.Println("\nHere are the base commands:")
	fmt.Println("    agora         : Manage Jumpstart Agora automation")
	fmt.Println("    arcbox        : Manage Jumpstart ArcBox automation (deploy, delete, list)")
	fmt.Println("    completion    : Generate shell completion scripts")
	fmt.Println("    localbox      : Manage Jumpstart LocalBox automation")
	fmt.Println("    repo          : Manage Jumpstart user local source code repository (init, update, delete)")
	fmt.Println("    subscription  : Manage Azure subscriptions (show, list, set)")
	fmt.Println("    upgrade       : Upgrade the Jumpstart CLI to the latest version")
	fmt.Println("    version       : Display the current version of the CLI")
	fmt.Println("\nUse 'js <command> --help' for more information on a command.")
}
