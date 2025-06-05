package upgrade

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

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

Examples:
  js upgrade                    # Check and upgrade if newer version available
  js upgrade --check           # Only check for updates, don't install
  js upgrade --pre-release     # Include pre-release versions
  js upgrade --force           # Force upgrade even if already latest
  js upgrade --rollback        # Rollback to a previous version
  js upgrade --list-backups    # List available backup versions

The upgrade command checks GitHub releases for the latest version and can
automatically download and install updates. It also supports rolling back
to previous versions if needed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flag values
			checkOnly, _ := cmd.Flags().GetBool("check")
			preRelease, _ := cmd.Flags().GetBool("pre-release")
			force, _ := cmd.Flags().GetBool("force")
			rollback, _ := cmd.Flags().GetBool("rollback")
			listBackups, _ := cmd.Flags().GetBool("list-backups")
			cleanupDays, _ := cmd.Flags().GetInt("cleanup-days")

			if utils.DebugMode {
				utils.Debug("Upgrade flags: check=%t, pre-release=%t, force=%t, rollback=%t, list-backups=%t",
					checkOnly, preRelease, force, rollback, listBackups)
			}

			// Handle backup-related commands first
			if listBackups {
				return handleListBackups()
			}

			if rollback {
				return handleRollback()
			}

			if cleanupDays > 0 {
				return installer.CleanupOldBackups(cleanupDays)
			}

			// Regular upgrade process starts here
			return performUpgrade(checkOnly, preRelease, force)
		},
	}

	// Add flags
	upgradeCmd.Flags().BoolP("check", "c", false, "Only check for updates, don't install")
	upgradeCmd.Flags().BoolP("force", "f", false, "Force upgrade even if already latest version")
	upgradeCmd.Flags().BoolP("pre-release", "p", false, "Include pre-release versions in check")
	upgradeCmd.Flags().BoolP("rollback", "r", false, "Rollback to a previous version")
	upgradeCmd.Flags().BoolP("list-backups", "l", false, "List available backup versions")
	upgradeCmd.Flags().IntP("cleanup-days", "", 0, "Clean up backups older than specified days (e.g., --cleanup-days=30)")

	return upgradeCmd
}

// performUpgrade handles the main upgrade logic
func performUpgrade(checkOnly, preRelease, force bool) error {
	// Check for updates
	fmt.Println(utils.InfoColor("🔍 Checking for updates..."))
	versionInfo, err := version.CheckForUpdates(preRelease)
	if err != nil {
		// If the repository doesn't exist yet, provide helpful guidance
		if strings.Contains(err.Error(), "404") {
			fmt.Println(utils.WarnColor("⚠️  Repository not found - jscli releases not yet published"))
			fmt.Println("📝 The upgrade feature is ready, but requires:")
			fmt.Println("   1. A GitHub repository with binary releases")
			fmt.Println("   2. Release assets named like: js-linux-amd64, js-windows-amd64.exe, js-darwin-arm64")
			fmt.Println("   3. Update the GitHubReleasesAPI constant in internal/upgrade/version.go")
			fmt.Printf("📖 For manual installation, visit: %s\n", version.GetManualDownloadURL())
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
		if force {
			fmt.Println(utils.WarnColor("⚠️  No newer version available, but --force specified"))
		}
		return nil
	}

	// Create backup before upgrade
	currentExe, err := os.Executable()
	if err != nil {
		fmt.Printf("⚠️  Warning: failed to get current executable path: %v\n", err)
	} else {
		backup, err := installer.CreateBackup(currentExe, versionInfo.Current)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to create backup: %v\n", err)
		} else {
			fmt.Printf("📦 Created backup of current version (%s)\n", backup.Version)
		}
	}

	// Proceed with upgrade
	if versionInfo.IsNewer || force {
		// Check if we have a download URL
		if versionInfo.DownloadURL == "" {
			fmt.Println(utils.WarnColor("🚧 No binary available for automatic download"))
			fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Println("📝 To enable automatic upgrades:")
			fmt.Println("   1. Create binary releases in your GitHub repository")
			fmt.Println("   2. Name assets like: js-linux-amd64, js-windows-amd64.exe, js-darwin-arm64")
			fmt.Println("   3. Update GitHubReleasesAPI in internal/upgrade/version.go")
			fmt.Printf("📖 For manual installation, visit: %s\n", version.GetManualDownloadURL())
			return nil
		}

		fmt.Printf("📦 Found binary for %s/%s\n", runtime.GOOS, runtime.GOARCH)
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

		fmt.Printf("🎉 Successfully upgraded from %s to %s!\n",
			versionInfo.Current, versionInfo.Latest)
		
		// Only show rollback message if current binary supports it
		if supportsRollback() {
			fmt.Println("💡 If you encounter issues, you can rollback using: js upgrade --rollback")
		}
	}

	return nil
}

// handleListBackups lists all available backup versions
func handleListBackups() error {
	fmt.Println(utils.InfoColor("📋 Available Backup Versions:"))
	fmt.Println(strings.Repeat("=", 50))

	backups, err := installer.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %v", err)
	}

	if len(backups) == 0 {
		fmt.Println("No backups found.")
		fmt.Println("💡 Backups are created automatically when you upgrade.")
		return nil
	}

	for i, backup := range backups {
		fmt.Printf("%d. Version: %s\n", i+1, backup.Version)
		fmt.Printf("   Date: %s\n", backup.BackupTime.Format("January 2, 2006 at 15:04"))
		fmt.Printf("   Location: %s\n", backup.BackupPath)
		if i < len(backups)-1 {
			fmt.Println()
		}
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("💡 To rollback to a version, use: js upgrade --rollback")

	return nil
}

// handleRollback handles the rollback process
func handleRollback() error {
	fmt.Println(utils.InfoColor("🔄 Rollback to Previous Version"))
	fmt.Println(strings.Repeat("=", 40))

	backups, err := installer.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %v", err)
	}

	if len(backups) == 0 {
		fmt.Println("❌ No backups available for rollback.")
		fmt.Println("💡 Backups are created automatically when you upgrade.")
		return nil
	}

	// Display available backups
	fmt.Println("Available versions to rollback to:")
	for i, backup := range backups {
		fmt.Printf("%d. Version %s (backed up on %s)\n",
			i+1, backup.Version, backup.BackupTime.Format("Jan 2, 2006"))
	}

	// Get user choice
	fmt.Print("\nEnter the number of the version to rollback to (or 'cancel'): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}

	choice := strings.TrimSpace(strings.ToLower(input))
	if choice == "cancel" || choice == "c" {
		fmt.Println("❌ Rollback cancelled.")
		return nil
	}

	// Parse choice
	var selectedBackup *installer.BackupInfo
	if choiceNum := parseInt(choice); choiceNum > 0 && choiceNum <= len(backups) {
		selectedBackup = backups[choiceNum-1]
	} else {
		return fmt.Errorf("invalid choice: %s", choice)
	}

	// Confirm rollback
	fmt.Printf("\n⚠️  This will rollback from version %s to %s.\n", utils.CliVersion, selectedBackup.Version)
	fmt.Print("Are you sure? (y/N): ")

	confirm, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %v", err)
	}

	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
		fmt.Println("❌ Rollback cancelled.")
		return nil
	}

	// Perform rollback
	return installer.RestoreBackup(selectedBackup)
}

// parseInt safely parses an integer from a string
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// supportsRollback checks if the current binary supports rollback functionality
// This prevents showing rollback messages in older versions that don't have the feature
func supportsRollback() bool {
	// Create a temporary upgrade command to check if rollback flag exists
	tempCmd := NewUpgradeCmd()
	flag := tempCmd.Flags().Lookup("rollback")
	return flag != nil
}
