package installer

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"jumpstartcli/internal/upgrade/config"
	"jumpstartcli/internal/utils"
)

// PlatformInfo contains platform-specific information
type PlatformInfo struct {
	OS           string
	Architecture string
	BinaryName   string
	AssetPattern string
}

// DownloadInfo contains download progress information
type DownloadInfo struct {
	URL        string
	Filename   string
	Size       int64
	Downloaded int64
	TempPath   string
	FinalPath  string
}

// BackupInfo contains backup information for rollback capability
type BackupInfo struct {
	BackupPath   string
	OriginalPath string
	BackupTime   time.Time
	Version      string
	ChecksumMD5  string
}

// GetPlatformInfo detects the current platform and returns appropriate binary info
func GetPlatformInfo() *PlatformInfo {
	platform := &PlatformInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	// Set binary name based on OS
	if platform.OS == "windows" {
		platform.BinaryName = "js.exe"
	} else {
		platform.BinaryName = "js"
	}

	// Create asset pattern for matching releases
	if platform.OS == "windows" {
		platform.AssetPattern = config.WindowsBinaryPattern
	} else if platform.OS == "darwin" {
		platform.AssetPattern = config.DarwinBinaryPattern
	} else {
		platform.AssetPattern = config.LinuxBinaryPattern
	}

	return platform
}

// FindDownloadURL finds the appropriate download URL from release assets
func FindDownloadURL(release *config.GitHubRelease, platform *PlatformInfo) (string, error) {
	if utils.DebugMode {
		utils.Debug("Looking for asset pattern: %s", platform.AssetPattern)
		utils.Debug("Available assets: %d", len(release.Assets))
	}

	for _, asset := range release.Assets {
		if utils.DebugMode {
			utils.Debug("Checking asset: %s", asset.Name)
		}

		// Check for exact match first
		if asset.Name == platform.AssetPattern {
			return asset.BrowserDownloadURL, nil
		}

		// Check for partial match (case-insensitive)
		assetLower := strings.ToLower(asset.Name)

		if strings.Contains(assetLower, strings.ToLower(platform.OS)) &&
			strings.Contains(assetLower, strings.ToLower(platform.Architecture)) {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("no suitable binary found for %s/%s", platform.OS, platform.Architecture)
}

// DownloadBinary downloads the binary from the given URL with progress tracking
func DownloadBinary(url string, platform *PlatformInfo) (*DownloadInfo, error) {
	if utils.DebugMode {
		utils.Debug("Starting download from: %s", url)
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "jscli-upgrade-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Prepare download info
	downloadInfo := &DownloadInfo{
		URL:      url,
		Filename: platform.BinaryName,
		TempPath: filepath.Join(tempDir, platform.BinaryName),
	}

	// Get current executable path for replacement
	currentExe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get current executable path: %v", err)
	}
	downloadInfo.FinalPath = currentExe

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute, // Longer timeout for downloads
	}

	// Make request
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to start download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Get content length for progress tracking
	downloadInfo.Size = resp.ContentLength

	// Create temp file
	tempFile, err := os.Create(downloadInfo.TempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	// Download with progress tracking
	fmt.Printf("📥 Downloading %s...\n", platform.BinaryName)

	progressReader := &progressReader{
		Reader:     resp.Body,
		total:      downloadInfo.Size,
		downloaded: &downloadInfo.Downloaded,
	}

	_, err = io.Copy(tempFile, progressReader)
	if err != nil {
		return nil, fmt.Errorf("failed to download binary: %v", err)
	}

	// Make the downloaded binary executable (Unix systems)
	if runtime.GOOS != "windows" {
		err = os.Chmod(downloadInfo.TempPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to make binary executable: %v", err)
		}
	}

	fmt.Printf("\n✅ Download completed: %s\n", downloadInfo.TempPath)
	return downloadInfo, nil
}

// InstallBinary replaces the current binary with the downloaded one
func InstallBinary(downloadInfo *DownloadInfo) error {
	if utils.DebugMode {
		utils.Debug("Installing binary from %s to %s", downloadInfo.TempPath, downloadInfo.FinalPath)
	}

	// Check if we need elevated permissions
	needsSudo, err := checkPermissions(downloadInfo.FinalPath)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %v", err)
	}

	if needsSudo {
		// Ask user what they want to do
		choice, err := promptUserForInstallChoice(downloadInfo.FinalPath)
		if err != nil {
			return fmt.Errorf("failed to get user choice: %v", err)
		}

		switch choice {
		case "sudo":
			return installWithSudo(downloadInfo)
		case "user":
			return installToUserLocation(downloadInfo)
		case "manual":
			return provideManualInstructions(downloadInfo)
		case "cancel":
			fmt.Println("❌ Upgrade cancelled by user.")
			return fmt.Errorf("upgrade cancelled")
		default:
			return fmt.Errorf("invalid choice")
		}
	}

	// Normal installation when no elevated permissions needed
	return performInstallation(downloadInfo)
}

// checkPermissions checks if we need elevated permissions to write to the target path
func checkPermissions(targetPath string) (bool, error) {
	// Always check the parent directory permissions since we can't reliably
	// test write access to a file that might be currently executing
	parentDir := filepath.Dir(targetPath)

	// Try to create a temp file in the parent directory
	tempFile, err := os.CreateTemp(parentDir, "perm-test-")
	if err != nil {
		if os.IsPermission(err) {
			return true, nil
		}
		return false, err
	}
	tempFile.Close()
	os.Remove(tempFile.Name())

	// For system directories like /usr/local/bin, we should ask for user choice
	// even if we can write (since it's a system-wide installation)
	if strings.HasPrefix(targetPath, "/usr/") ||
		strings.HasPrefix(targetPath, "/bin/") ||
		strings.HasPrefix(targetPath, "/sbin/") ||
		strings.HasPrefix(targetPath, "/opt/") {
		return true, nil
	}

	return false, nil
}

// promptUserForInstallChoice asks the user how they want to handle the installation
func promptUserForInstallChoice(targetPath string) (string, error) {
	fmt.Printf("\n⚠️  Permission required to update binary at: %s\n", targetPath)
	fmt.Println("\nChoose how to proceed:")
	fmt.Println("1. [sudo] Use sudo to update system-wide binary (requires admin password)")
	fmt.Println("2. [user] Install to user directory (~/.local/bin) - no admin required")
	fmt.Println("3. [manual] Show manual installation instructions")
	fmt.Println("4. [cancel] Cancel the upgrade")
	fmt.Print("\nEnter your choice [1-4 or sudo/user/manual/cancel]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	choice := strings.TrimSpace(strings.ToLower(input))

	// Map numeric choices to text
	switch choice {
	case "1", "sudo":
		return "sudo", nil
	case "2", "user":
		return "user", nil
	case "3", "manual":
		return "manual", nil
	case "4", "cancel":
		return "cancel", nil
	default:
		fmt.Println("❌ Invalid choice. Please try again.")
		return promptUserForInstallChoice(targetPath)
	}
}

// installWithSudo performs installation using sudo
func installWithSudo(downloadInfo *DownloadInfo) error {
	fmt.Println("\n🔐 Using sudo to install system-wide binary...")
	fmt.Println("⚠️  You may be prompted for your password.")

	// On Unix systems, we need to use a replacement script even with sudo
	// because of the "text file busy" error when replacing a running executable
	if runtime.GOOS != "windows" {
		return installWithSudoUnix(downloadInfo)
	}

	// Windows implementation (simpler, no "text file busy" issue)
	return installWithSudoWindows(downloadInfo)
}

// installWithSudoUnix handles sudo installation on Unix systems using a replacement script
func installWithSudoUnix(downloadInfo *DownloadInfo) error {
	// Create a replacement script that uses sudo
	scriptPath := filepath.Join(filepath.Dir(downloadInfo.TempPath), "sudo-replace.sh")
	scriptContent := fmt.Sprintf(`#!/bin/bash
# Auto-generated script for sudo binary replacement
set -e

# Wait a moment for the parent process to exit
sleep 1

# Create backup
sudo cp "%s" "%s.backup" 2>/dev/null || true

# Copy new binary to final location with sudo
sudo cp "%s" "%s"

# Make it executable
sudo chmod +x "%s"

# Clean up
rm -rf "%s"
rm -f "%s"

echo "✅ System-wide installation completed successfully!"
echo "🎉 You can now run the command again to use the new version."
`, downloadInfo.FinalPath, downloadInfo.FinalPath, downloadInfo.TempPath, downloadInfo.FinalPath, downloadInfo.FinalPath, filepath.Dir(downloadInfo.TempPath), scriptPath)

	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		return fmt.Errorf("failed to create sudo replacement script: %v", err)
	}

	if utils.DebugMode {
		utils.Debug("Created sudo replacement script: %s", scriptPath)
	}

	// Execute the script in the background
	cmd := exec.Command("bash", scriptPath)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start sudo replacement script: %v", err)
	}

	fmt.Println("🔄 System-wide binary replacement initiated...")
	fmt.Println("✅ Process will complete after this command exits.")

	return nil
}

// installWithSudoWindows handles sudo installation on Windows (if applicable)
func installWithSudoWindows(downloadInfo *DownloadInfo) error {
	// Create backup first
	tempDir := filepath.Dir(downloadInfo.TempPath)
	backupPath := filepath.Join(tempDir, filepath.Base(downloadInfo.FinalPath)+".backup")

	// Use runas or equivalent for Windows elevation (simplified for now)
	cmd := exec.Command("copy", downloadInfo.FinalPath, backupPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Install new binary
	cmd = exec.Command("copy", downloadInfo.TempPath, downloadInfo.FinalPath)
	if err := cmd.Run(); err != nil {
		// Try to restore backup
		exec.Command("copy", backupPath, downloadInfo.FinalPath).Run()
		return fmt.Errorf("failed to install binary: %v", err)
	}

	// Clean up
	os.RemoveAll(filepath.Dir(downloadInfo.TempPath))
	os.Remove(backupPath)

	fmt.Println("✅ System-wide installation completed successfully!")
	return nil
}

// installToUserLocation installs the binary to user's local bin directory
func installToUserLocation(downloadInfo *DownloadInfo) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	userBinDir := filepath.Join(homeDir, ".local", "bin")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(userBinDir, 0755); err != nil {
		return fmt.Errorf("failed to create user bin directory: %v", err)
	}

	platform := GetPlatformInfo()
	userBinaryPath := filepath.Join(userBinDir, platform.BinaryName)

	fmt.Printf("\n📁 Installing to user directory: %s\n", userBinaryPath)

	// Create backup if binary exists
	if _, err := os.Stat(userBinaryPath); err == nil {
		tempDir := filepath.Dir(downloadInfo.TempPath)
		backupPath := filepath.Join(tempDir, filepath.Base(userBinaryPath)+".backup")
		if err := copyFile(userBinaryPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %v", err)
		}
	}

	// Copy new binary
	if err := copyFile(downloadInfo.TempPath, userBinaryPath); err != nil {
		return fmt.Errorf("failed to copy binary to user directory: %v", err)
	}

	// Make executable
	if err := os.Chmod(userBinaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %v", err)
	}

	// Clean up temp files
	os.RemoveAll(filepath.Dir(downloadInfo.TempPath))

	fmt.Println("✅ User installation completed successfully!")
	fmt.Printf("\n📋 To use the new binary, make sure %s is in your PATH:\n", userBinDir)
	fmt.Printf("   export PATH=\"%s:$PATH\"\n", userBinDir)
	fmt.Println("\n💡 Add this line to your ~/.bashrc or ~/.zshrc to make it permanent.")

	return nil
}

// provideManualInstructions shows manual installation instructions
func provideManualInstructions(downloadInfo *DownloadInfo) error {
	fmt.Println("\n📋 Manual Installation Instructions:")
	fmt.Println("=" + strings.Repeat("=", 40))
	fmt.Printf("1. Downloaded binary location: %s\n", downloadInfo.TempPath)
	fmt.Printf("2. Target installation location: %s\n", downloadInfo.FinalPath)
	fmt.Println("\n🔧 To install manually, run:")
	fmt.Printf("   sudo cp \"%s\" \"%s\"\n", downloadInfo.TempPath, downloadInfo.FinalPath)
	fmt.Printf("   sudo chmod +x \"%s\"\n", downloadInfo.FinalPath)

	fmt.Println("\n💡 Alternative - Install to user directory:")
	homeDir, _ := os.UserHomeDir()
	userBinDir := filepath.Join(homeDir, ".local", "bin")
	platform := GetPlatformInfo()
	userBinaryPath := filepath.Join(userBinDir, platform.BinaryName)

	fmt.Printf("   mkdir -p \"%s\"\n", userBinDir)
	fmt.Printf("   cp \"%s\" \"%s\"\n", downloadInfo.TempPath, userBinaryPath)
	fmt.Printf("   chmod +x \"%s\"\n", userBinaryPath)
	fmt.Printf("   export PATH=\"%s:$PATH\"\n", userBinDir)

	fmt.Printf("\n⚠️  Remember to clean up: rm -rf \"%s\"\n", filepath.Dir(downloadInfo.TempPath))
	fmt.Println("\n❌ Upgrade process stopped. Please complete the installation manually.")

	return fmt.Errorf("manual installation required")
}

// performInstallation performs the actual installation when no special permissions needed
func performInstallation(downloadInfo *DownloadInfo) error {
	// Create backup of current binary in temp directory (to avoid permission issues)
	tempDir := filepath.Dir(downloadInfo.TempPath)
	backupPath := filepath.Join(tempDir, filepath.Base(downloadInfo.FinalPath)+".backup")
	err := copyFile(downloadInfo.FinalPath, backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	if utils.DebugMode {
		utils.Debug("Created backup at: %s", backupPath)
	}

	// Replace current binary
	fmt.Println("🔄 Installing new binary...")

	// On Windows, we might need to handle file locking differently
	if runtime.GOOS == "windows" {
		// Try to rename current binary first
		tempOldPath := downloadInfo.FinalPath + ".old"
		err = os.Rename(downloadInfo.FinalPath, tempOldPath)
		if err != nil {
			return fmt.Errorf("failed to move current binary: %v", err)
		}

		// Move new binary into place
		err = os.Rename(downloadInfo.TempPath, downloadInfo.FinalPath)
		if err != nil {
			// Try to restore original
			os.Rename(tempOldPath, downloadInfo.FinalPath)
			return fmt.Errorf("failed to install new binary: %v", err)
		}

		// Clean up old binary
		os.Remove(tempOldPath)
	} else {
		// Unix systems: Use a replacement script for self-updating
		// This is necessary because a running process cannot replace its own executable
		err = replaceBinaryUnix(downloadInfo)
		if err != nil {
			// Try to restore from backup
			copyFile(backupPath, downloadInfo.FinalPath)
			return fmt.Errorf("failed to install new binary: %v", err)
		}
	}

	// Clean up temp directory
	os.RemoveAll(filepath.Dir(downloadInfo.TempPath))

	// Remove backup if installation was successful
	os.Remove(backupPath)

	fmt.Println("✅ Installation completed successfully!")
	fmt.Println("🎉 Restart your terminal or run the command again to use the new version.")

	return nil
}

// replaceBinaryUnix handles binary replacement on Unix systems
// Since a running process cannot replace its own executable, we create a script
// that will perform the replacement after the current process exits
func replaceBinaryUnix(downloadInfo *DownloadInfo) error {
	if utils.DebugMode {
		utils.Debug("Using Unix binary replacement strategy")
	}

	// Try a different approach: rename current binary and move new one in place
	// This should work even if the current binary is running
	oldBinaryPath := downloadInfo.FinalPath + ".old"

	// Step 1: Rename current binary
	fmt.Printf("🔄 Renaming current binary: %s -> %s\n", downloadInfo.FinalPath, oldBinaryPath)
	err := os.Rename(downloadInfo.FinalPath, oldBinaryPath)
	if err != nil {
		return fmt.Errorf("failed to rename current binary: %v", err)
	}

	// Step 2: Copy new binary into place
	fmt.Printf("🔄 Installing new binary: %s -> %s\n", downloadInfo.TempPath, downloadInfo.FinalPath)
	err = copyFile(downloadInfo.TempPath, downloadInfo.FinalPath)
	if err != nil {
		// Try to restore the original binary
		os.Rename(oldBinaryPath, downloadInfo.FinalPath)
		return fmt.Errorf("failed to install new binary: %v", err)
	}

	// Step 3: Make new binary executable
	err = os.Chmod(downloadInfo.FinalPath, 0755)
	if err != nil {
		// Try to restore the original binary
		os.Remove(downloadInfo.FinalPath)
		os.Rename(oldBinaryPath, downloadInfo.FinalPath)
		return fmt.Errorf("failed to make new binary executable: %v", err)
	}

	// Step 4: Create a cleanup script to remove old binary and temp files
	scriptPath := filepath.Join(filepath.Dir(downloadInfo.TempPath), "cleanup.sh")
	cleanupScript := fmt.Sprintf(`#!/bin/bash
# Cleanup script for upgrade process
sleep 3
rm -f "%s" 2>/dev/null || true
rm -rf "%s" 2>/dev/null || true
rm -f "%s" 2>/dev/null || true
`, oldBinaryPath, filepath.Dir(downloadInfo.TempPath), scriptPath)

	err = os.WriteFile(scriptPath, []byte(cleanupScript), 0755)
	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("⚠️  Warning: failed to create cleanup script: %v\n", err)
	} else {
		// Run cleanup script in background
		cmd := exec.Command("bash", scriptPath)
		cmd.Start()
	}

	fmt.Println("✅ Binary replacement completed successfully!")
	fmt.Println("🎉 You can now run the command again to use the new version.")

	return nil
}

// CreateBackup creates a backup of the current binary with metadata
func CreateBackup(binaryPath string, version string) (*BackupInfo, error) {
	if utils.DebugMode {
		utils.Debug("Creating backup of %s (version %s)", binaryPath, version)
	}

	// Create backup directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".jscli", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupFileName := fmt.Sprintf("js-backup-%s-%s", version, timestamp)
	if runtime.GOOS == "windows" {
		backupFileName += ".exe"
	}
	backupPath := filepath.Join(backupDir, backupFileName)

	// Copy binary to backup location
	if err := copyFile(binaryPath, backupPath); err != nil {
		return nil, fmt.Errorf("failed to create backup: %v", err)
	}

	// Calculate checksum for integrity verification
	checksum, err := calculateMD5(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate backup checksum: %v", err)
	}

	backup := &BackupInfo{
		BackupPath:   backupPath,
		OriginalPath: binaryPath,
		BackupTime:   time.Now(),
		Version:      version,
		ChecksumMD5:  checksum,
	}

	// Save backup metadata
	if err := saveBackupMetadata(backup); err != nil {
		return nil, fmt.Errorf("failed to save backup metadata: %v", err)
	}

	fmt.Printf("📦 Backup created: %s\n", backupPath)
	return backup, nil
}

// ListBackups returns a list of available backups
func ListBackups() ([]*BackupInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".jscli", "backups")
	metadataFile := filepath.Join(backupDir, "metadata.json")

	// Check if metadata file exists
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		return []*BackupInfo{}, nil
	}

	// Read metadata
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup metadata: %v", err)
	}

	var backups []*BackupInfo
	if err := json.Unmarshal(data, &backups); err != nil {
		return nil, fmt.Errorf("failed to parse backup metadata: %v", err)
	}

	// Filter out backups where files no longer exist
	validBackups := make([]*BackupInfo, 0)
	for _, backup := range backups {
		if _, err := os.Stat(backup.BackupPath); err == nil {
			validBackups = append(validBackups, backup)
		}
	}

	return validBackups, nil
}

// RestoreBackup restores a previous backup
func RestoreBackup(backup *BackupInfo) error {
	if utils.DebugMode {
		utils.Debug("Restoring backup from %s to %s", backup.BackupPath, backup.OriginalPath)
	}

	// Verify backup integrity
	currentChecksum, err := calculateMD5(backup.BackupPath)
	if err != nil {
		return fmt.Errorf("failed to verify backup integrity: %v", err)
	}

	if currentChecksum != backup.ChecksumMD5 {
		return fmt.Errorf("backup integrity check failed - backup may be corrupted")
	}

	// Create backup of current version before restore
	currentVersion := utils.CliVersion
	tempBackup, err := CreateBackup(backup.OriginalPath, currentVersion)
	if err != nil {
		fmt.Printf("⚠️  Warning: failed to backup current version: %v\n", err)
	}

	// Perform restore
	fmt.Printf("🔄 Restoring version %s...\n", backup.Version)

	if runtime.GOOS == "windows" {
		// Windows: rename current and restore
		tempPath := backup.OriginalPath + ".temp"
		if err := os.Rename(backup.OriginalPath, tempPath); err != nil {
			return fmt.Errorf("failed to move current binary: %v", err)
		}

		if err := copyFile(backup.BackupPath, backup.OriginalPath); err != nil {
			// Try to restore original
			os.Rename(tempPath, backup.OriginalPath)
			return fmt.Errorf("failed to restore backup: %v", err)
		}

		os.Remove(tempPath)
	} else {
		// Unix: use replacement strategy
		oldPath := backup.OriginalPath + ".old"
		if err := os.Rename(backup.OriginalPath, oldPath); err != nil {
			return fmt.Errorf("failed to rename current binary: %v", err)
		}

		if err := copyFile(backup.BackupPath, backup.OriginalPath); err != nil {
			os.Rename(oldPath, backup.OriginalPath)
			return fmt.Errorf("failed to restore backup: %v", err)
		}

		if err := os.Chmod(backup.OriginalPath, 0755); err != nil {
			return fmt.Errorf("failed to make restored binary executable: %v", err)
		}

		os.Remove(oldPath)
	}

	fmt.Printf("✅ Successfully restored to version %s\n", backup.Version)
	if tempBackup != nil {
		fmt.Printf("💡 Previous version backed up to: %s\n", tempBackup.BackupPath)
	}

	return nil
}

// CleanupOldBackups removes backups older than the specified number of days
func CleanupOldBackups(maxAgeDays int) error {
	backups, err := ListBackups()
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -maxAgeDays)
	removed := 0

	for _, backup := range backups {
		if backup.BackupTime.Before(cutoffTime) {
			if err := os.Remove(backup.BackupPath); err != nil {
				fmt.Printf("⚠️  Warning: failed to remove old backup %s: %v\n", backup.BackupPath, err)
			} else {
				removed++
			}
		}
	}

	if removed > 0 {
		fmt.Printf("🧹 Cleaned up %d old backup(s)\n", removed)
		// Update metadata to remove deleted backups
		validBackups, _ := ListBackups()
		saveBackupMetadata(validBackups...)
	}

	return nil
}

// saveBackupMetadata saves backup metadata to disk
func saveBackupMetadata(backups ...*BackupInfo) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	backupDir := filepath.Join(homeDir, ".jscli", "backups")
	metadataFile := filepath.Join(backupDir, "metadata.json")

	// Read existing metadata
	var existingBackups []*BackupInfo
	if data, err := os.ReadFile(metadataFile); err == nil {
		json.Unmarshal(data, &existingBackups)
	}

	// Add new backups
	allBackups := append(existingBackups, backups...)

	// Write updated metadata
	data, err := json.MarshalIndent(allBackups, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataFile, data, 0644)
}

// calculateMD5 calculates MD5 checksum of a file
func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// progressReader wraps an io.Reader to track download progress
type progressReader struct {
	io.Reader
	total      int64
	downloaded *int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if n > 0 {
		*pr.downloaded += int64(n)
		pr.printProgress()
	}
	return n, err
}

func (pr *progressReader) printProgress() {
	if pr.total <= 0 {
		fmt.Printf("\rDownloaded: %s", formatBytes(*pr.downloaded))
	} else {
		percentage := float64(*pr.downloaded) / float64(pr.total) * 100
		fmt.Printf("\rProgress: %.1f%% (%s / %s)",
			percentage, formatBytes(*pr.downloaded), formatBytes(pr.total))
	}
}

// formatBytes formats byte count as human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
