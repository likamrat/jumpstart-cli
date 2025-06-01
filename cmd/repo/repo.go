package repo

import (
	"fmt"
	"os"
	"os/exec"

	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// NewRepoCmd creates and returns the repo command
func NewRepoCmd() *cobra.Command {
	var repoCmd = &cobra.Command{
		Use:   "repo",
		Short: "Manage Jumpstart user local source code repository",		Long: `Manage the Jumpstart source code repository.

Subcommands:
  • clone    Clone the Jumpstart source code repository
  • update   Update the cloned Jumpstart repo directory
  • delete   Delete the cloned Jumpstart repo directory

Use 'js repo <subcommand> --help' for more details.`,
		// Disable Cobra's built-in suggestions and errors to use our custom ones
		DisableSuggestions: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {				// List of valid subcommands for repo
				validSubcommands := []string{"clone", "update", "delete"}

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
				return fmt.Errorf("unknown subcommand '%s' for 'js repo'", invalidSubcommand)
			}
			// If no args, show help
			utils.ShowHelpWithoutTypes(cmd)
			return nil
		},
	}
	var repoCloneCmd = &cobra.Command{
		Use:   "clone",
		Short: "Clone the Jumpstart source code repository",
		Long: `Clone the Jumpstart source code repository from GitHub to the working directory or a custom path.

Optional argument:
  • --path / -p   Custom path to clone the repo (default: ./jumpstart)`,
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := cmd.Flags().GetString("path")
			repoURL := "https://github.com/microsoft/azure_arc"
			if path == "" {
				path = "jumpstart"
			}
			if _, err := os.Stat(path); err == nil {
				utils.Error("Directory '%s' already exists.", path)
				return
			}
			utils.Info("Cloning Jumpstart repo to '%s'...", path)
			cmdExec := exec.Command("git", "clone", repoURL, path)
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			if err := cmdExec.Run(); err != nil {
				utils.Error("Failed to clone repo: %v", err)
				return
			}
			utils.Info("Jumpstart repo cloned to '%s'", path)
		},
	}

	var repoDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete the cloned Jumpstart repo directory",
		Long: `Delete the cloned Jumpstart source code repository directory.

Optional argument:
  • --path / -p   Custom path of the repo directory (default: ./jumpstart)`,
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := cmd.Flags().GetString("path")
			if path == "" {
				path = "jumpstart"
			}
			if _, err := os.Stat(path); os.IsNotExist(err) {
				utils.Error("Directory '%s' does not exist.", path)
				return
			}
			utils.Info("Deleting directory '%s'...", path)
			if err := os.RemoveAll(path); err != nil {
				utils.Error("Failed to delete directory: %v", err)
				return
			}
			utils.Info("Deleted directory '%s'", path)
		},
	}

	var repoUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the cloned Jumpstart repo directory",
		Long: `Update the cloned Jumpstart source code repository directory (git pull).

Optional argument:
  • --path / -p   Custom path of the repo directory (default: ./jumpstart)`,
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := cmd.Flags().GetString("path")
			if path == "" {
				path = "jumpstart"
			}
			gitDir := path + "/.git"
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				utils.Error("No git repo found at '%s'", path)
				return
			}
			utils.Info("Updating repo at '%s'...", path)
			cmdExec := exec.Command("git", "-C", path, "pull")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			if err := cmdExec.Run(); err != nil {
				utils.Error("Failed to update repo: %v", err)
				return
			}
			utils.Info("Repo at '%s' updated", path)
		},
	}
	// Add flags for repo commands
	repoCloneCmd.Flags().StringP("path", "p", "", "Custom path to clone the repo (default: ./jumpstart)")
	repoDeleteCmd.Flags().StringP("path", "p", "", "Custom path of the repo directory (default: ./jumpstart)")
	repoUpdateCmd.Flags().StringP("path", "p", "", "Custom path of the repo directory (default: ./jumpstart)")

	// Add subcommands to repo command
	repoCmd.AddCommand(repoDeleteCmd, repoCloneCmd, repoUpdateCmd)

	return repoCmd
}
