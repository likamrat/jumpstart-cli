package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"jumpstartcli/internal/upgrade/config"
	"jumpstartcli/internal/upgrade/installer"
	"jumpstartcli/internal/utils"
)

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
	var apiURL string
	var release config.GitHubRelease

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	if includePrereleases {
		// When including prereleases, get all releases and find the latest
		apiURL = config.GetRepositoryAllReleasesURL()

		if utils.DebugMode {
			utils.Debug("Checking for updates (including prereleases) from: %s", apiURL)
		}

		resp, err := client.Get(apiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to check for updates: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		}

		// Parse response as array of releases
		var releases []config.GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, fmt.Errorf("failed to parse release information: %v", err)
		}

		if len(releases) == 0 {
			return nil, fmt.Errorf("no releases found")
		}

		// Find the latest release (first in the list is latest)
		release = releases[0]
	} else {
		// For stable releases only, use the /releases/latest endpoint
		apiURL = config.GetRepositoryURL()

		if utils.DebugMode {
			utils.Debug("Checking for updates (stable only) from: %s", apiURL)
		}

		resp, err := client.Get(apiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to check for updates: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		}

		// Parse response as single release
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil, fmt.Errorf("failed to parse release information: %v", err)
		}

		// Skip prereleases unless explicitly requested
		if release.Prerelease && !includePrereleases {
			return nil, fmt.Errorf("latest release is a prerelease, use --pre-release flag to include")
		}
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
	platform := installer.GetPlatformInfo()
	downloadURL, err := installer.FindDownloadURL(&release, platform)
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
// Supports semantic versioning with pre-release and build metadata
func CompareVersions(v1, v2 string) int {
	// Clean version strings (remove 'v' prefix if present)
	v1 = cleanVersionTag(v1)
	v2 = cleanVersionTag(v2)

	if v1 == v2 {
		return 0
	}

	// Parse versions to handle pre-release and build metadata
	ver1 := parseVersion(v1)
	ver2 := parseVersion(v2)

	// Compare core version parts (major.minor.patch)
	result := compareCoreVersion(ver1.core, ver2.core)
	if result != 0 {
		return result
	}

	// If core versions are equal, compare pre-release versions
	return comparePreRelease(ver1.preRelease, ver2.preRelease)
}

// versionParts represents parsed version components
type versionParts struct {
	core       []string
	preRelease string
	build      string
}

// parseVersion parses a version string into its components
func parseVersion(version string) versionParts {
	var parts versionParts

	// Split on '+' to separate build metadata
	buildSplit := strings.Split(version, "+")
	if len(buildSplit) > 1 {
		parts.build = buildSplit[1]
	}

	// Split on '-' to separate pre-release
	preReleaseSplit := strings.Split(buildSplit[0], "-")
	parts.core = strings.Split(preReleaseSplit[0], ".")

	if len(preReleaseSplit) > 1 {
		parts.preRelease = strings.Join(preReleaseSplit[1:], "-")
	}

	return parts
}

// compareCoreVersion compares core version numbers (major.minor.patch)
func compareCoreVersion(parts1, parts2 []string) int {
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

// comparePreRelease compares pre-release versions
// Pre-release versions have lower precedence than normal versions
func comparePreRelease(pre1, pre2 string) int {
	// If neither has pre-release, they're equal
	if pre1 == "" && pre2 == "" {
		return 0
	}

	// Version without pre-release has higher precedence
	if pre1 == "" && pre2 != "" {
		return 1
	}
	if pre1 != "" && pre2 == "" {
		return -1
	}

	// Both have pre-release, compare them lexically
	if pre1 < pre2 {
		return -1
	}
	if pre1 > pre2 {
		return 1
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
	return config.GetRepositoryWebURL()
}
