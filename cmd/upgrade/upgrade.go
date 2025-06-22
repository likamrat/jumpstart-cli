package upgrade

import (
	"fmt"
	"runtime"
	"strings"

	"jumpstartcli/internal/examples"
	"jumpstartcli/internal/upgrade/installer"
	"jumpstartcli/internal/upgrade/version"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// NewUpgradeCmd creates and returns the upgrade command
func NewUpgradeCmd() *cobra.Command {
	var upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the Jumpstart CLI to the latest version",
		Long: `Check for and install the latest version of the Jumpstart CLI.

Subcommands:
  • check      Check for available updates without installing
  • install    Download and install the latest version
  • rollback   Rollback to a previous version
  • list       List available backup versions

Use 'js upgrade <subcommand> --help' for more details.`,
		// Disable suggestions to use our custom handling
		DisableSuggestions: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		// Use RunE instead of Args for better control over suggestion handling
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// Check if the first argument matches any subcommand
				validCommands := []string{"check", "install", "rollback", "list"}
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
				return fmt.Errorf("unknown subcommand '%s' for 'js upgrade'", invalidCommand)
			}
			// If no args, show help
			utils.ShowHelpWithoutTypes(cmd)
			return nil
		},
	}

	// upgrade check
	var upgradeCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Check for available updates without installing",
		Long: `This command will check GitHub releases for newer versions and display
version information without making any changes to your installation.

Use --yes/-y to execute the check operation.

` + examples.GetExamples("upgrade.check").FormatExamples(),
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")

			// If --yes not provided, use centralized error handling
			if !yes {
				utils.HandleMissingRequiredArguments(cmd, []string{"--yes"})
				return nil
			}

			preRelease, _ := cmd.Flags().GetBool("pre-release")
			return performUpgrade(true, preRelease, false) // checkOnly=true
		},
	}
	upgradeCheckCmd.Flags().BoolP("yes", "y", false, "Execute the check operation")
	upgradeCheckCmd.Flags().BoolP("pre-release", "p", false, "Include pre-release versions in check")

	// upgrade install
	var upgradeInstallCmd = &cobra.Command{
		Use:   "install",
		Short: "Download and install the latest version",
		Long: `This command will check for updates, download the latest version,
and replace your current installation with the newer version.

Use --yes/-y to execute the installation operation.

` + examples.GetExamples("upgrade.install").FormatExamples(),
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")

			// If --yes not provided, use centralized error handling
			if !yes {
				utils.HandleMissingRequiredArguments(cmd, []string{"--yes"})
				return nil
			}

			preRelease, _ := cmd.Flags().GetBool("pre-release")
			force, _ := cmd.Flags().GetBool("force")
			cleanupDays, _ := cmd.Flags().GetInt("cleanup-days")

			// Log cleanup days for future implementation
			if cleanupDays > 0 {
				fmt.Printf("Cleanup configured for %d days (feature coming soon)\n", cleanupDays)
			}

			return performUpgrade(false, preRelease, force) // checkOnly=false
		},
	}
	upgradeInstallCmd.Flags().BoolP("yes", "y", false, "Execute the installation operation")
	upgradeInstallCmd.Flags().BoolP("pre-release", "p", false, "Include pre-release versions")
	upgradeInstallCmd.Flags().BoolP("force", "f", false, "Force upgrade even if already latest version")
	upgradeInstallCmd.Flags().Int("cleanup-days", 7, "Days to keep old versions for rollback")

	// upgrade rollback (placeholder for future implementation)
	var upgradeRollbackCmd = &cobra.Command{
		Use:   "rollback",
		Short: "Rollback to a previous version",
		Long: `This command will restore a previous version from your backup installations.
You can specify a version or select from available backups interactively.

Use --yes/-y to execute the rollback operation.

` + examples.GetExamples("upgrade.rollback").FormatExamples(),
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")

			// If --yes not provided, use centralized error handling
			if !yes {
				utils.HandleMissingRequiredArguments(cmd, []string{"--yes"})
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Rollback functionality is not yet implemented")
			fmt.Fprintln(cmd.OutOrStdout(), "This feature will allow rolling back to previous versions.")
			return nil
		},
	}
	upgradeRollbackCmd.Flags().BoolP("yes", "y", false, "Execute the rollback operation")
	upgradeRollbackCmd.Flags().String("version", "", "Specific version to rollback to")

	// upgrade list (placeholder for future implementation)
	var upgradeListCmd = &cobra.Command{
		Use:   "list",
		Short: "List available backup versions",
		Long: `This command will display all backup versions available for rollback,
including version numbers, installation dates, and current status.

Use --yes/-y to execute the list operation.

` + examples.GetExamples("upgrade.list").FormatExamples(),
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")

			// If --yes not provided, use centralized error handling
			if !yes {
				utils.HandleMissingRequiredArguments(cmd, []string{"--yes"})
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "List backups functionality is not yet implemented")
			fmt.Fprintln(cmd.OutOrStdout(), "This feature will show available backup versions.")
			return nil
		},
	}
	upgradeListCmd.Flags().BoolP("yes", "y", false, "Execute the list operation")

	// Add subcommands
	upgradeCmd.AddCommand(upgradeCheckCmd)
	upgradeCmd.AddCommand(upgradeInstallCmd)
	upgradeCmd.AddCommand(upgradeRollbackCmd)
	upgradeCmd.AddCommand(upgradeListCmd)

	return upgradeCmd
}

// performUpgrade handles the main upgrade logic
func performUpgrade(checkOnly, preRelease, force bool) error {
	// Check for updates
	fmt.Println("Checking for updates...")
	versionInfo, err := version.CheckForUpdates(preRelease)
	if err != nil {
		// If the repository doesn't exist yet, provide helpful guidance
		if strings.Contains(err.Error(), "404") {
			fmt.Println("Repository not found - jscli releases not yet published")
			fmt.Println("The upgrade feature is ready, but requires:")
			fmt.Println("   1. A GitHub repository with binary releases")
			fmt.Println("   2. Release assets named like: js-linux-amd64, js-windows-amd64.exe, js-darwin-arm64")
			fmt.Println("   3. Update the GitHubReleasesAPI constant in internal/upgrade/version.go")
			fmt.Printf("For manual installation, visit: %s\n", version.GetManualDownloadURL())
			return nil
		}
		return fmt.Errorf("failed to check for updates: %v", err)
	}

	// Display version information
	fmt.Print(versionInfo.FormatVersionInfo())

	// If only checking, stop here
	if checkOnly {
		return nil
	}

	// If no newer version and not forcing, stop here
	if !versionInfo.IsNewer && !force {
		fmt.Println("You are already running the latest version!")
		return nil
	}

	// Proceed with upgrade
	if versionInfo.IsNewer || force {
		// Show appropriate message for force installation
		if force && !versionInfo.IsNewer {
			fmt.Println("Force installation requested - reinstalling current version")
		}

		// Check if we have a download URL
		if versionInfo.DownloadURL == "" {
			fmt.Println("No binary available for automatic download")
			fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Println("To enable automatic upgrades:")
			fmt.Println("   1. Create binary releases in your GitHub repository")
			fmt.Println("   2. Name assets like: js-linux-amd64, js-windows-amd64.exe, js-darwin-arm64")
			fmt.Println("   3. Update GitHubReleasesAPI in internal/upgrade/version.go")
			fmt.Printf("For manual installation, visit: %s\n", version.GetManualDownloadURL())
			return nil
		}

		fmt.Printf("Found binary for %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Download URL: %s\n", versionInfo.DownloadURL)

		// Download the binary
		platform := installer.GetPlatformInfo()
		downloadInfo, err := installer.DownloadBinary(versionInfo.DownloadURL, platform)
		if err != nil {
			return fmt.Errorf("download failed: %v", err)
		}

		// Install the binary
		err = installer.InstallBinary(downloadInfo)
		if err != nil {
			return fmt.Errorf("installation failed: %v", err)
		}

		fmt.Printf("Successfully upgraded from %s to %s!\n",
			versionInfo.Current, versionInfo.Latest)
	}

	return nil
}
