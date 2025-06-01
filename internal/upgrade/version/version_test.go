package version

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"
	"jumpstartcli/internal/upgrade/installer"
)

var (
	upgradeTestSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	upgradeTestInfoColor    = color.New(color.FgCyan).SprintFunc()
	upgradeTestWarnColor    = color.New(color.FgYellow).SprintFunc()
	upgradeTestErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	upgradeTestHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

func printUpgradeTestStatus(t *testing.T, testName string, success bool, message string) {
	status := upgradeTestSuccessColor("✅")
	if !success {
		status = upgradeTestErrorColor("❌")
	}
	fmt.Printf("  %s %s: %s\n", status, upgradeTestInfoColor(testName), message)
}

func TestCompareVersions(t *testing.T) {
	fmt.Printf("\n%s\n", upgradeTestHeaderColor("=== Testing Version Comparison ==="))

	tests := []struct {
		v1       string
		v2       string
		expected int
		desc     string
	}{
		{"1.0.0", "2.0.0", -1, "v1 < v2"},
		{"2.0.0", "1.0.0", 1, "v1 > v2"},
		{"1.0.0", "1.0.0", 0, "v1 == v2"},
		{"1.0.0", "1.0.1", -1, "patch version difference"},
		{"1.1.0", "1.0.0", 1, "minor version difference"},
		{"v1.0.0", "1.0.0", 0, "with and without v prefix"},
		{"v1.2.3", "v1.2.4", -1, "both with v prefix"},
		{"0.1.0", "1.100.2", -1, "current vs much higher"},
		{"1.100.2", "0.1.0", 1, "higher vs current"},
	}

	successfulTests := 0
	for _, test := range tests {
		result := CompareVersions(test.v1, test.v2)
		if result != test.expected {
			printUpgradeTestStatus(t, fmt.Sprintf("Compare %s vs %s", test.v1, test.v2), false,
				fmt.Sprintf("Got %d, expected %d (%s)", result, test.expected, test.desc))
			t.Errorf("CompareVersions(%s, %s) = %d, expected %d (%s)",
				test.v1, test.v2, result, test.expected, test.desc)
		} else {
			printUpgradeTestStatus(t, fmt.Sprintf("Compare %s vs %s", test.v1, test.v2), true,
				fmt.Sprintf("Correctly returned %d (%s)", result, test.desc))
			successfulTests++
		}
	}

	if successfulTests == len(tests) {
		printUpgradeTestStatus(t, "All version comparisons", true, fmt.Sprintf("All %d version comparison tests passed", successfulTests))
	}
}

func TestCleanVersionTag(t *testing.T) {
	fmt.Printf("\n%s\n", upgradeTestHeaderColor("=== Testing Version Tag Cleaning ==="))

	tests := []struct {
		input    string
		expected string
	}{
		{"v1.0.0", "1.0.0"},
		{"1.0.0", "1.0.0"},
		{"v1.2.3-beta", "1.2.3-beta"},
		{"", ""},
	}

	successfulTests := 0
	for _, test := range tests {
		result := cleanVersionTag(test.input)
		if result != test.expected {
			printUpgradeTestStatus(t, fmt.Sprintf("Clean version '%s'", test.input), false,
				fmt.Sprintf("Got '%s', expected '%s'", result, test.expected))
			t.Errorf("cleanVersionTag(%s) = %s, expected %s",
				test.input, result, test.expected)
		} else {
			printUpgradeTestStatus(t, fmt.Sprintf("Clean version '%s'", test.input), true,
				fmt.Sprintf("Correctly cleaned to '%s'", result))
			successfulTests++
		}
	}

	if successfulTests == len(tests) {
		printUpgradeTestStatus(t, "All version cleanings", true, fmt.Sprintf("All %d version cleaning tests passed", successfulTests))
	}
}

func TestGetPlatformInfo(t *testing.T) {
	fmt.Printf("\n%s\n", upgradeTestHeaderColor("=== Testing Platform Information ==="))

	platform := installer.GetPlatformInfo()

	if platform.OS == "" {
		printUpgradeTestStatus(t, "OS detection", false, "OS should not be empty")
		t.Error("OS should not be empty")
	} else {
		printUpgradeTestStatus(t, "OS detection", true, fmt.Sprintf("OS detected: %s", platform.OS))
	}

	if platform.Architecture == "" {
		printUpgradeTestStatus(t, "Architecture detection", false, "Architecture should not be empty")
		t.Error("Architecture should not be empty")
	} else {
		printUpgradeTestStatus(t, "Architecture detection", true, fmt.Sprintf("Architecture detected: %s", platform.Architecture))
	}

	if platform.BinaryName == "" {
		printUpgradeTestStatus(t, "Binary name generation", false, "BinaryName should not be empty")
		t.Error("BinaryName should not be empty")
	} else {
		printUpgradeTestStatus(t, "Binary name generation", true, fmt.Sprintf("Binary name: %s", platform.BinaryName))
	}

	if platform.AssetPattern == "" {
		printUpgradeTestStatus(t, "Asset pattern generation", false, "AssetPattern should not be empty")
		t.Error("AssetPattern should not be empty")
	} else {
		printUpgradeTestStatus(t, "Asset pattern generation", true, fmt.Sprintf("Asset pattern: %s", platform.AssetPattern))
	}

	// Test that Windows gets .exe extension
	if platform.OS == "windows" && !strings.Contains(platform.BinaryName, ".exe") {
		printUpgradeTestStatus(t, "Windows .exe extension", false, "Windows binary should have .exe extension")
		t.Error("Windows binary should have .exe extension")
	} else if platform.OS == "windows" {
		printUpgradeTestStatus(t, "Windows .exe extension", true, "Windows binary correctly has .exe extension")
	} else {
		printUpgradeTestStatus(t, "Non-Windows binary extension", true, fmt.Sprintf("Non-Windows platform (%s) has appropriate binary name", platform.OS))
	}
}
