package upgrade

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
		platform.AssetPattern = WindowsBinaryPattern
	} else if platform.OS == "darwin" {
		platform.AssetPattern = DarwinBinaryPattern
	} else {
		platform.AssetPattern = LinuxBinaryPattern
	}

	return platform
}

// FindDownloadURL finds the appropriate download URL from release assets
func FindDownloadURL(release *GitHubRelease, platform *PlatformInfo) (string, error) {
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

	// Create backup of current binary
	backupPath := downloadInfo.FinalPath + ".backup"
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
		// Unix systems: direct replacement
		err = os.Rename(downloadInfo.TempPath, downloadInfo.FinalPath)
		if err != nil {
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
