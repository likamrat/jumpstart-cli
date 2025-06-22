// repo.go - Repository management command for Jumpstart CLI
package repo

import (
	"fmt"
	"os"

	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// NewRepoCmd creates the repo command
func NewRepoCmd() *cobra.Command {
	var repoCmd = &cobra.Command{
		Use:   "repo",
		Short: "Manage Jumpstart user local source code repository",
		Long: `This command helps you initialize, update, and manage your local Jumpstart repository
containing automation scripts, templates, and resources.

Subcommands:
  • init     Initialize a new local Jumpstart repository
  • update   Update existing repository with latest templates
  • delete   Delete local repository`,
		// Disable suggestions to use our custom handling
		DisableSuggestions: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// Check if the first argument matches any subcommand
				validCommands := []string{"init", "update", "delete"}
				invalidCommand := args[0]

				for _, validCmd := range validCommands {
					if invalidCommand == validCmd {
						return nil // Valid command, continue normal processing
					}
				}

				// General similarity checking
				if suggestion := utils.SuggestSimilarCommand(invalidCommand, validCommands, 3); suggestion != "" {
					utils.PrintDidYouMean(invalidCommand, suggestion)
					return nil
				}

				// No suggestion found, show normal error
				return fmt.Errorf("unknown subcommand '%s' for 'js repo'", invalidCommand)
			}
			// If no args, show help
			utils.ShowHelpWithoutTypes(cmd)
			return nil
		},
	}

	// repo init
	var repoInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new local Jumpstart repository",
		Long: `This command will:
- Clone or download the latest Jumpstart repository templates
- Set up the local directory structure
- Configure initial settings

The repository will contain all the necessary templates, scripts, and resources
for deploying Jumpstart scenarios like ArcBox, LocalBox, and Agora.`,
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := cmd.Flags().GetString("path")
			branch, _ := cmd.Flags().GetString("branch")
			force, _ := cmd.Flags().GetBool("force")

			fmt.Println(utils.InfoColor("🚀 Initializing Jumpstart repository..."))

			if path == "" {
				path = "./jumpstart"
			}

			fmt.Printf("Repository path: %s\n", path)
			fmt.Printf("Branch: %s\n", branch)
			if force {
				fmt.Println("Force mode: enabled")
			}

			// TODO: Implement actual repository initialization logic
			fmt.Println(utils.WarnColor("⚠️  Repository initialization is not yet implemented."))
			fmt.Println("This feature will be available in a future release.")
			fmt.Println("\nFor now, you can manually clone the repository:")
			fmt.Println("  git clone https://github.com/microsoft/azure_arc.git")
		},
	}

	// repo update
	var repoUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update existing repository with latest templates",
		Long: `This command will:
- Pull the latest changes from the remote repository
- Update templates and automation scripts
- Preserve your local customizations

This ensures you have access to the latest features, bug fixes, and improvements.`,
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := cmd.Flags().GetString("path")
			branch, _ := cmd.Flags().GetString("branch")

			fmt.Println(utils.InfoColor("🔄 Updating Jumpstart repository..."))

			if path == "" {
				path = "./jumpstart"
			}

			fmt.Printf("Repository path: %s\n", path)
			fmt.Printf("Branch: %s\n", branch)

			// TODO: Implement actual repository update logic
			fmt.Println(utils.WarnColor("⚠️  Repository update is not yet implemented."))
			fmt.Println("This feature will be available in a future release.")
			fmt.Println("\nFor now, you can manually update the repository:")
			fmt.Println("  cd <your-repo-path> && git pull origin main")
		},
	}

	// repo delete
	var repoDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete local repository",
		Long: `This command will remove the entire local repository directory and all its contents.
Use with caution as this action cannot be undone.

Any local customizations or modifications will be permanently lost unless
they have been backed up or committed to a remote repository.`,
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := cmd.Flags().GetString("path")
			force, _ := cmd.Flags().GetBool("force")

			if path == "" {
				path = "./jumpstart"
			}

			if !force {
				utils.PrintRequiredArgumentsError([]string{"--force"})
				os.Exit(1)
			}

			fmt.Printf("Repository path: %s\n", path)
			fmt.Println(utils.InfoColor("🗑️  Deleting Jumpstart repository..."))

			// TODO: Implement actual repository deletion logic
			fmt.Println(utils.WarnColor("⚠️  Repository deletion is not yet implemented."))
			fmt.Println("This feature will be available in a future release.")
			fmt.Println("\nFor now, you can manually delete the repository:")
			fmt.Printf("  rm -rf %s\n", path)
		},
	}

	// Add flags for init command
	repoInitCmd.Flags().StringP("path", "p", "", "Path where to initialize the repository (default: ./jumpstart)")
	repoInitCmd.Flags().StringP("branch", "b", "main", "Git branch to use")
	repoInitCmd.Flags().BoolP("force", "f", false, "Force initialization even if directory exists")

	// Add flags for update command
	repoUpdateCmd.Flags().StringP("path", "p", "", "Path to the repository to update (default: ./jumpstart)")
	repoUpdateCmd.Flags().StringP("branch", "b", "main", "Git branch to use")

	// Add flags for delete command
	repoDeleteCmd.Flags().StringP("path", "p", "", "Path to the repository to delete (default: ./jumpstart)")
	repoDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")

	// Add subcommands
	repoCmd.AddCommand(repoInitCmd)
	repoCmd.AddCommand(repoUpdateCmd)
	repoCmd.AddCommand(repoDeleteCmd)

	return repoCmd
}
