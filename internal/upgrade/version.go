package upgrade

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"jumpstartcli/internal/utils"
)

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// VersionInfo contains version comparison information
type VersionInfo struct {
	Current     string
	Latest      string
	DownloadURL string
	ReleaseDate time.Time
	Changelog   string
	IsNewer     bool
}

const (
	// Manual download URL for when binary releases aren't available
	ManualDownloadURL = "https://github.com/microsoft/azure_arc/releases"
)

// CheckForUpdates checks GitHub for the latest release version
func CheckForUpdates(includePrereleases bool) (*VersionInfo, error) {
	apiURL := GetRepositoryURL()

	if utils.DebugMode {
		utils.Debug("Checking for updates from: %s", apiURL)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request to GitHub API
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Parse response
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release information: %v", err)
	}

	// Skip prereleases unless explicitly requested
	if release.Prerelease && !includePrereleases {
		return nil, fmt.Errorf("latest release is a prerelease, use --pre-release flag to include")
	}

	// Create version info
	versionInfo := &VersionInfo{
		Current:     utils.CliVersion,
		Latest:      cleanVersionTag(release.TagName),
		ReleaseDate: release.PublishedAt,
		Changelog:   release.Body,
		IsNewer:     false,
	}

	// Find download URL for current platform
	platform := GetPlatformInfo()
	downloadURL, err := FindDownloadURL(&release, platform)
	if err != nil {
		if utils.DebugMode {
			utils.Debug("Could not find download URL: %v", err)
		}
		// Don't fail the whole operation, just leave download URL empty
	} else {
		versionInfo.DownloadURL = downloadURL
	}

	// Compare versions
	comparison := CompareVersions(versionInfo.Current, versionInfo.Latest)
	versionInfo.IsNewer = comparison < 0

	if utils.DebugMode {
		utils.Debug("Current: %s, Latest: %s, IsNewer: %t",
			versionInfo.Current, versionInfo.Latest, versionInfo.IsNewer)
	}

	return versionInfo, nil
}

// CompareVersions compares two semantic version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Clean version strings (remove 'v' prefix if present)
	v1 = cleanVersionTag(v1)
	v2 = cleanVersionTag(v2)

	if v1 == v2 {
		return 0
	}

	// Split versions into parts
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Pad shorter version with zeros
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for len(parts1) < maxLen {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < maxLen {
		parts2 = append(parts2, "0")
	}

	// Compare each part
	for i := 0; i < maxLen; i++ {
		// Convert to integers for proper numeric comparison
		var num1, num2 int
		fmt.Sscanf(parts1[i], "%d", &num1)
		fmt.Sscanf(parts2[i], "%d", &num2)

		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
	}

	return 0
}

// cleanVersionTag removes 'v' prefix from version tags
func cleanVersionTag(version string) string {
	if strings.HasPrefix(version, "v") {
		return version[1:]
	}
	return version
}

// FormatVersionInfo returns a formatted string with version information
func (v *VersionInfo) FormatVersionInfo() string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Current version: %s\n", v.Current))
	result.WriteString(fmt.Sprintf("Latest version:  %s\n", v.Latest))

	if v.IsNewer {
		result.WriteString(utils.InfoColor("📦 A newer version is available!\n"))
	} else {
		result.WriteString(utils.SuccessColor("✅ You have the latest version!\n"))
	}

	if v.ReleaseDate.Year() > 1 { // Check if date is set
		result.WriteString(fmt.Sprintf("Released: %s\n", v.ReleaseDate.Format("January 2, 2006")))
	}

	return result.String()
}

// GetManualDownloadURL returns the manual download URL
func GetManualDownloadURL() string {
	return GetRepositoryWebURL()
}
