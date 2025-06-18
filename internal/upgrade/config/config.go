package config

import (
	"os"
	"time"
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

// Configuration for the upgrade system
// Updated to point to the fork for testing and development

const (
	// Repository configuration
	DefaultOwner = "likamrat"
	DefaultRepo  = "jumpstart-cli"

	// Binary naming patterns
	// These match the release workflow binary names
	LinuxBinaryPattern   = "js-linux-amd64"
	LinuxARM64Pattern    = "js-linux-arm64"
	WindowsBinaryPattern = "js-windows-amd64.exe"
	DarwinBinaryPattern  = "js-darwin-amd64"
	DarwinARM64Pattern   = "js-darwin-arm64"

	// Release configuration
	IncludePrereleases = false
	CheckInterval      = 24 // hours
)

// GetRepositoryURL returns the GitHub API URL for releases
func GetRepositoryURL() string {
	// Allow environment variable override for testing
	if testURL := os.Getenv("GITHUB_API_TEST_URL_LATEST"); testURL != "" {
		return testURL
	}
	return "https://api.github.com/repos/" + DefaultOwner + "/" + DefaultRepo + "/releases/latest"
}

// GetRepositoryAllReleasesURL returns the GitHub API URL for all releases
func GetRepositoryAllReleasesURL() string {
	// Allow environment variable override for testing
	if testURL := os.Getenv("GITHUB_API_TEST_URL_ALL"); testURL != "" {
		return testURL
	}
	return "https://api.github.com/repos/" + DefaultOwner + "/" + DefaultRepo + "/releases"
}

// GetRepositoryWebURL returns the web URL for manual downloads
func GetRepositoryWebURL() string {
	return "https://github.com/" + DefaultOwner + "/" + DefaultRepo + "/releases"
}

// GetEnv gets the environment variable or returns the default value
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetRepositoryURLFromEnv returns the GitHub API URL for releases from environment variable
func GetRepositoryURLFromEnv() string {
	return GetEnv("GITHUB_API_URL", GetRepositoryURL())
}

// GetRepositoryAllReleasesURLFromEnv returns the GitHub API URL for all releases from environment variable
func GetRepositoryAllReleasesURLFromEnv() string {
	return GetEnv("GITHUB_API_ALL_URL", GetRepositoryAllReleasesURL())
}

// GetRepositoryWebURLFromEnv returns the web URL for manual downloads from environment variable
func GetRepositoryWebURLFromEnv() string {
	return GetEnv("GITHUB_WEB_URL", GetRepositoryWebURL())
}
