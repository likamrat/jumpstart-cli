package upgrade

import (
	"fmt"
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

			if utils.DebugMode {
				utils.Debug("Upgrade flags: check=%t, pre-release=%t, force=%t",
					checkOnly, preRelease, force)
			}

			// Regular upgrade process starts here
			return performUpgrade(checkOnly, preRelease, force)
		},
	}

	// Add flags
	upgradeCmd.Flags().BoolP("check", "c", false, "Only check for updates, don't install")
	upgradeCmd.Flags().BoolP("force", "f", false, "Force upgrade even if already latest version")
	upgradeCmd.Flags().BoolP("pre-release", "p", false, "Include pre-release versions in check")

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
	}

	return nil
}
