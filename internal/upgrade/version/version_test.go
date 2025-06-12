package version

import (
	"strings"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	result := CompareVersions("1.0.0", "2.0.0")
	if result != -1 {
		t.Errorf("CompareVersions(1.0.0, 2.0.0) = %d, expected -1", result)
	}

	result = CompareVersions("2.0.0", "1.0.0")
	if result != 1 {
		t.Errorf("CompareVersions(2.0.0, 1.0.0) = %d, expected 1", result)
	}

	result = CompareVersions("1.0.0", "1.0.0")
	if result != 0 {
		t.Errorf("CompareVersions(1.0.0, 1.0.0) = %d, expected 0", result)
	}

	result = CompareVersions("1.0.0-alpha", "1.0.0")
	if result != -1 {
		t.Errorf("CompareVersions(1.0.0-alpha, 1.0.0) = %d, expected -1", result)
	}

	result = CompareVersions("v1.0.0", "1.0.0")
	if result != 0 {
		t.Errorf("CompareVersions(v1.0.0, 1.0.0) = %d, expected 0", result)
	}

	result = CompareVersions("", "1.0.0")
	if result != -1 {
		t.Errorf("CompareVersions('', 1.0.0) = %d, expected -1", result)
	}

	result = CompareVersions("invalid", "1.0.0")
	if result != -1 {
		t.Errorf("CompareVersions(invalid, 1.0.0) = %d, expected -1", result)
	}

	// Additional comprehensive test cases
	result = CompareVersions("1.0.0", "1.0.0-alpha")
	if result != 1 {
		t.Errorf("CompareVersions(1.0.0, 1.0.0-alpha) = %d, expected 1", result)
	}

	result = CompareVersions("1.0.0-alpha", "1.0.0-beta")
	if result != -1 {
		t.Errorf("CompareVersions(1.0.0-alpha, 1.0.0-beta) = %d, expected -1", result)
	}

	result = CompareVersions("1.0.0-alpha.1", "1.0.0-alpha.2")
	if result != -1 {
		t.Errorf("CompareVersions(1.0.0-alpha.1, 1.0.0-alpha.2) = %d, expected -1", result)
	}

	result = CompareVersions("2.0.0-alpha", "1.9.9")
	if result != 1 {
		t.Errorf("CompareVersions(2.0.0-alpha, 1.9.9) = %d, expected 1", result)
	}

	result = CompareVersions("1.0.0", "")
	if result != 1 {
		t.Errorf("CompareVersions(1.0.0, '') = %d, expected 1", result)
	}

	result = CompareVersions("", "")
	if result != 0 {
		t.Errorf("CompareVersions('', '') = %d, expected 0", result)
	}

	result = CompareVersions("invalid", "invalid")
	if result != 0 {
		t.Errorf("CompareVersions(invalid, invalid) = %d, expected 0", result)
	}

	result = CompareVersions("1", "1.0")
	if result != 0 {
		t.Errorf("CompareVersions(1, 1.0) = %d, expected 0", result)
	}

	result = CompareVersions("1.0", "1.0.0")
	if result != 0 {
		t.Errorf("CompareVersions(1.0, 1.0.0) = %d, expected 0", result)
	}

	result = CompareVersions("1.0.0.0", "1.0.0")
	if result != 0 {
		t.Errorf("CompareVersions(1.0.0.0, 1.0.0) = %d, expected 0", result)
	}
}

func TestParseVersion(t *testing.T) {
	result := parseVersion("1.2.3")
	if len(result.core) != 3 || result.core[0] != "1" || result.core[1] != "2" || result.core[2] != "3" {
		t.Errorf("parseVersion(1.2.3) = %+v, expected core [1 2 3]", result)
	}

	// parseVersion doesn't clean the v prefix, so v1.2.3 becomes [v1 2 3]
	result = parseVersion("v1.2.3")
	if len(result.core) != 3 || result.core[0] != "v1" || result.core[1] != "2" || result.core[2] != "3" {
		t.Errorf("parseVersion(v1.2.3) = %+v, expected core [v1 2 3]", result)
	}

	result = parseVersion("1.2.3-alpha")
	if result.preRelease != "alpha" {
		t.Errorf("parseVersion(1.2.3-alpha) preRelease = %s, expected alpha", result.preRelease)
	}

	result = parseVersion("1.2.3+build")
	if result.build != "build" {
		t.Errorf("parseVersion(1.2.3+build) build = %s, expected build", result.build)
	}

	result = parseVersion("1.2.3-alpha+build")
	if result.preRelease != "alpha" || result.build != "build" {
		t.Errorf("parseVersion(1.2.3-alpha+build) = %+v, expected preRelease=alpha, build=build", result)
	}

	// parseVersion splits empty string on "." and gets [""]
	result = parseVersion("")
	if len(result.core) != 1 || result.core[0] != "" {
		t.Errorf("parseVersion('') = %+v, expected core ['']", result)
	}

	result = parseVersion("invalid")
	if len(result.core) != 1 || result.core[0] != "invalid" {
		t.Errorf("parseVersion(invalid) = %+v, expected core [invalid]", result)
	}

	result = parseVersion("1.2.3-alpha.1")
	if result.preRelease != "alpha.1" {
		t.Errorf("parseVersion(1.2.3-alpha.1) preRelease = %s, expected alpha.1", result.preRelease)
	}

	result = parseVersion("1.2.3-alpha-beta")
	if result.preRelease != "alpha-beta" {
		t.Errorf("parseVersion(1.2.3-alpha-beta) preRelease = %s, expected alpha-beta", result.preRelease)
	}

	result = parseVersion("1")
	if len(result.core) != 1 || result.core[0] != "1" {
		t.Errorf("parseVersion(1) = %+v, expected core [1]", result)
	}

	result = parseVersion("1.2")
	if len(result.core) != 2 || result.core[0] != "1" || result.core[1] != "2" {
		t.Errorf("parseVersion(1.2) = %+v, expected core [1 2]", result)
	}
}

func TestCompareCoreVersion(t *testing.T) {
	result := compareCoreVersion([]string{"1", "0", "0"}, []string{"2", "0", "0"})
	if result != -1 {
		t.Errorf("compareCoreVersion([1 0 0], [2 0 0]) = %d, expected -1", result)
	}

	result = compareCoreVersion([]string{"2", "0", "0"}, []string{"1", "0", "0"})
	if result != 1 {
		t.Errorf("compareCoreVersion([2 0 0], [1 0 0]) = %d, expected 1", result)
	}

	result = compareCoreVersion([]string{"1", "0", "0"}, []string{"1", "0", "0"})
	if result != 0 {
		t.Errorf("compareCoreVersion([1 0 0], [1 0 0]) = %d, expected 0", result)
	}

	result = compareCoreVersion([]string{"1", "0"}, []string{"1", "0", "0"})
	if result != 0 {
		t.Errorf("compareCoreVersion([1 0], [1 0 0]) = %d, expected 0", result)
	}

	result = compareCoreVersion([]string{}, []string{"1", "0", "0"})
	if result != -1 {
		t.Errorf("compareCoreVersion([], [1 0 0]) = %d, expected -1", result)
	}

	result = compareCoreVersion([]string{"10"}, []string{"2"})
	if result != 1 {
		t.Errorf("compareCoreVersion([10], [2]) = %d, expected 1", result)
	}

	result = compareCoreVersion([]string{"invalid"}, []string{"1"})
	if result != -1 {
		t.Errorf("compareCoreVersion([invalid], [1]) = %d, expected -1", result)
	}

	// Additional test cases for better coverage
	result = compareCoreVersion([]string{"1", "0", "0"}, []string{"1", "0"})
	if result != 0 {
		t.Errorf("compareCoreVersion([1 0 0], [1 0]) = %d, expected 0", result)
	}

	result = compareCoreVersion([]string{"1", "0", "0"}, []string{})
	if result != 1 {
		t.Errorf("compareCoreVersion([1 0 0], []) = %d, expected 1", result)
	}

	result = compareCoreVersion([]string{}, []string{})
	if result != 0 {
		t.Errorf("compareCoreVersion([], []) = %d, expected 0", result)
	}

	result = compareCoreVersion([]string{"1", "0", "0"}, []string{"1", "0", "0", "1"})
	if result != -1 {
		t.Errorf("compareCoreVersion([1 0 0], [1 0 0 1]) = %d, expected -1", result)
	}

	result = compareCoreVersion([]string{"1", "0", "0", "1"}, []string{"1", "0", "0"})
	if result != 1 {
		t.Errorf("compareCoreVersion([1 0 0 1], [1 0 0]) = %d, expected 1", result)
	}

	result = compareCoreVersion([]string{"2"}, []string{"10"})
	if result != -1 {
		t.Errorf("compareCoreVersion([2], [10]) = %d, expected -1", result)
	}

	result = compareCoreVersion([]string{"1"}, []string{"invalid"})
	if result != 1 {
		t.Errorf("compareCoreVersion([1], [invalid]) = %d, expected 1", result)
	}

	result = compareCoreVersion([]string{"invalid"}, []string{"invalid"})
	if result != 0 {
		t.Errorf("compareCoreVersion([invalid], [invalid]) = %d, expected 0", result)
	}

	result = compareCoreVersion([]string{"1", "2"}, []string{"1", "1", "9"})
	if result != 1 {
		t.Errorf("compareCoreVersion([1 2], [1 1 9]) = %d, expected 1", result)
	}

	result = compareCoreVersion([]string{"1", "1", "9"}, []string{"1", "2"})
	if result != -1 {
		t.Errorf("compareCoreVersion([1 1 9], [1 2]) = %d, expected -1", result)
	}
}

func TestComparePreRelease(t *testing.T) {
	result := comparePreRelease("alpha", "beta")
	if result != -1 {
		t.Errorf("comparePreRelease(alpha, beta) = %d, expected -1", result)
	}

	result = comparePreRelease("beta", "alpha")
	if result != 1 {
		t.Errorf("comparePreRelease(beta, alpha) = %d, expected 1", result)
	}

	result = comparePreRelease("alpha", "alpha")
	if result != 0 {
		t.Errorf("comparePreRelease(alpha, alpha) = %d, expected 0", result)
	}

	result = comparePreRelease("", "alpha")
	if result != 1 {
		t.Errorf("comparePreRelease('', alpha) = %d, expected 1", result)
	}

	result = comparePreRelease("alpha", "")
	if result != -1 {
		t.Errorf("comparePreRelease(alpha, '') = %d, expected -1", result)
	}

	result = comparePreRelease("alpha.1", "alpha.2")
	if result != -1 {
		t.Errorf("comparePreRelease(alpha.1, alpha.2) = %d, expected -1", result)
	}

	result = comparePreRelease("1", "2")
	if result != -1 {
		t.Errorf("comparePreRelease(1, 2) = %d, expected -1", result)
	}

	result = comparePreRelease("10", "2")
	if result != -1 {
		t.Errorf("comparePreRelease(10, 2) = %d, expected -1", result)
	}

	// Additional comprehensive test cases
	result = comparePreRelease("", "")
	if result != 0 {
		t.Errorf("comparePreRelease('', '') = %d, expected 0", result)
	}

	result = comparePreRelease("alpha.2", "alpha.1")
	if result != 1 {
		t.Errorf("comparePreRelease(alpha.2, alpha.1) = %d, expected 1", result)
	}

	result = comparePreRelease("alpha.1", "alpha.1")
	if result != 0 {
		t.Errorf("comparePreRelease(alpha.1, alpha.1) = %d, expected 0", result)
	}

	result = comparePreRelease("alpha", "alpha.1")
	if result != -1 {
		t.Errorf("comparePreRelease(alpha, alpha.1) = %d, expected -1", result)
	}

	result = comparePreRelease("alpha.1", "alpha")
	if result != 1 {
		t.Errorf("comparePreRelease(alpha.1, alpha) = %d, expected 1", result)
	}

	result = comparePreRelease("beta.1", "alpha.2")
	if result != 1 {
		t.Errorf("comparePreRelease(beta.1, alpha.2) = %d, expected 1", result)
	}

	result = comparePreRelease("alpha.2", "beta.1")
	if result != -1 {
		t.Errorf("comparePreRelease(alpha.2, beta.1) = %d, expected -1", result)
	}

	result = comparePreRelease("rc", "beta")
	if result != 1 {
		t.Errorf("comparePreRelease(rc, beta) = %d, expected 1", result)
	}

	result = comparePreRelease("beta", "rc")
	if result != -1 {
		t.Errorf("comparePreRelease(beta, rc) = %d, expected -1", result)
	}

	result = comparePreRelease("2", "1")
	if result != 1 {
		t.Errorf("comparePreRelease(2, 1) = %d, expected 1", result)
	}

	result = comparePreRelease("2", "10")
	if result != 1 {
		t.Errorf("comparePreRelease(2, 10) = %d, expected 1", result)
	}

	result = comparePreRelease("1.0", "1.1")
	if result != -1 {
		t.Errorf("comparePreRelease(1.0, 1.1) = %d, expected -1", result)
	}

	result = comparePreRelease("1.1", "1.0")
	if result != 1 {
		t.Errorf("comparePreRelease(1.1, 1.0) = %d, expected 1", result)
	}
}

func TestCheckForUpdates(t *testing.T) {
	vInfo, err := CheckForUpdates(false)
	if err != nil {
		t.Logf("CheckForUpdates without pre-release failed (expected in test environment): %v", err)
	}
	_ = vInfo

	vInfo, err = CheckForUpdates(true)
	if err != nil {
		t.Logf("CheckForUpdates with pre-release failed (expected in test environment): %v", err)
	}
	_ = vInfo
}

func TestFormatVersionInfo(t *testing.T) {
	vInfo := &VersionInfo{
		Current:     "1.0.0",
		Latest:      "1.0.1",
		IsNewer:     true,
		ReleaseDate: time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC),
		Changelog:   "Bug fixes",
		DownloadURL: "https://example.com/download",
	}

	result := vInfo.FormatVersionInfo()
	if result == "" {
		t.Error("FormatVersionInfo should return a non-empty string")
	}

	if !strings.Contains(result, "1.0.0") {
		t.Error("FormatVersionInfo should contain current version")
	}

	if !strings.Contains(result, "1.0.1") {
		t.Error("FormatVersionInfo should contain latest version")
	}

	vInfo2 := &VersionInfo{
		Current:     "1.0.1",
		Latest:      "1.0.1",
		IsNewer:     false,
		ReleaseDate: time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC),
		Changelog:   "",
		DownloadURL: "",
	}

	result2 := vInfo2.FormatVersionInfo()
	if result2 == "" {
		t.Error("FormatVersionInfo should return a non-empty string for up-to-date version")
	}
}

func TestGetManualDownloadURL(t *testing.T) {
	result := GetManualDownloadURL()
	if result == "" {
		t.Error("GetManualDownloadURL should return a non-empty string")
	}

	if !strings.HasPrefix(result, "http") {
		t.Errorf("GetManualDownloadURL should return a URL, got: %s", result)
	}
}

func TestCleanVersionTag(t *testing.T) {
	result := cleanVersionTag("v1.2.3")
	if result != "1.2.3" {
		t.Errorf("cleanVersionTag(v1.2.3) = %s, expected 1.2.3", result)
	}

	result = cleanVersionTag("1.2.3")
	if result != "1.2.3" {
		t.Errorf("cleanVersionTag(1.2.3) = %s, expected 1.2.3", result)
	}

	result = cleanVersionTag("")
	if result != "" {
		t.Errorf("cleanVersionTag('') = %s, expected ''", result)
	}

	result = cleanVersionTag("V1.2.3")
	if result != "V1.2.3" {
		t.Errorf("cleanVersionTag(V1.2.3) = %s, expected V1.2.3", result)
	}

	// Additional test cases for edge cases
	result = cleanVersionTag("v")
	if result != "" {
		t.Errorf("cleanVersionTag(v) = %s, expected ''", result)
	}

	result = cleanVersionTag("V")
	if result != "V" {
		t.Errorf("cleanVersionTag(V) = %s, expected V", result)
	}

	result = cleanVersionTag("vv1.2.3")
	if result != "v1.2.3" {
		t.Errorf("cleanVersionTag(vv1.2.3) = %s, expected v1.2.3", result)
	}

	result = cleanVersionTag("version1.2.3")
	if result != "ersion1.2.3" {
		t.Errorf("cleanVersionTag(version1.2.3) = %s, expected ersion1.2.3", result)
	}
}
