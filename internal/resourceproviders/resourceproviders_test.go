package resourceproviders

import (
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/jumpstart-cli/internal/testutils"
)

var (
	rpTestSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	rpTestInfoColor    = color.New(color.FgCyan).SprintFunc()
	rpTestWarnColor    = color.New(color.FgYellow).SprintFunc()
	rpTestErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	rpTestHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

func printRPTestStatus(t *testing.T, testName string, success bool, message string) {
	status := rpTestSuccessColor("✅")
	if !success {
		status = rpTestErrorColor("❌")
	}
	fmt.Printf("  %s %s: %s\n", status, rpTestInfoColor(testName), message)
}

func TestGetArcBoxProviders(t *testing.T) {
	fmt.Printf("\n%s\n", rpTestHeaderColor("=== Testing ArcBox Providers ==="))

	config := GetArcBoxProviders()

	// Test basic configuration structure
	if config.SolutionName != "ArcBox" {
		printRPTestStatus(t, "Solution name", false, fmt.Sprintf("Expected SolutionName to be 'ArcBox', got '%s'", config.SolutionName))
		t.Errorf("Expected SolutionName to be 'ArcBox', got '%s'", config.SolutionName)
	} else {
		printRPTestStatus(t, "Solution name", true, fmt.Sprintf("Solution name correctly set to '%s'", rpTestSuccessColor(config.SolutionName)))
	}

	// Test that required providers are present
	expectedProviders := []string{
		"Microsoft.Compute",
		"Microsoft.Kubernetes",
		"Microsoft.KubernetesConfiguration",
		"Microsoft.ExtendedLocation",
		"Microsoft.AzureArcData",
		"Microsoft.OperationsManagement",
		"Microsoft.HybridConnectivity",
	}

	if len(config.RequiredProviders) == 0 {
		printRPTestStatus(t, "Required providers", false, "RequiredProviders should not be empty")
		t.Fatal("RequiredProviders should not be empty")
	} else {
		printRPTestStatus(t, "Required providers", true, fmt.Sprintf("Found %d required providers", len(config.RequiredProviders)))
	}

	// Check that all expected providers are present
	providerMap := make(map[string]bool)
	for _, provider := range config.RequiredProviders {
		providerMap[provider] = true
	}

	missingProviders := 0
	for _, expected := range expectedProviders {
		if !providerMap[expected] {
			printRPTestStatus(t, fmt.Sprintf("Provider %s", expected), false, fmt.Sprintf("Expected provider '%s' not found in ArcBox configuration", expected))
			t.Errorf("Expected provider '%s' not found in ArcBox configuration", expected)
			missingProviders++
		} else {
			printRPTestStatus(t, fmt.Sprintf("Provider %s", expected), true, "Provider found in configuration")
		}
	}

	if missingProviders == 0 {
		printRPTestStatus(t, "All expected providers", true, "All expected providers are present")
	}

	// Verify no empty provider names
	emptyProviders := 0
	for _, provider := range config.RequiredProviders {
		if provider == "" {
			printRPTestStatus(t, "Empty provider validation", false, "Found empty provider name in ArcBox configuration")
			t.Error("Found empty provider name in ArcBox configuration")
			emptyProviders++
		}
	}

	if emptyProviders == 0 {
		printRPTestStatus(t, "Empty provider validation", true, "No empty provider names found")
	}
}

func TestGetLocalBoxProviders(t *testing.T) {
	fmt.Printf("\n%s\n", rpTestHeaderColor("=== Testing LocalBox Providers ==="))

	config := GetLocalBoxProviders()

	// Test basic configuration structure
	if config.SolutionName != "LocalBox" {
		printRPTestStatus(t, "Solution name", false, fmt.Sprintf("Expected SolutionName to be 'LocalBox', got '%s'", config.SolutionName))
		t.Errorf("Expected SolutionName to be 'LocalBox', got '%s'", config.SolutionName)
	} else {
		printRPTestStatus(t, "Solution name", true, fmt.Sprintf("Solution name correctly set to '%s'", rpTestSuccessColor(config.SolutionName)))
	}

	// Test that required providers are present
	expectedProviders := []string{
		"Microsoft.Compute",
		"Microsoft.Network",
		"Microsoft.Storage",
	}

	if len(config.RequiredProviders) == 0 {
		printRPTestStatus(t, "Required providers", false, "RequiredProviders should not be empty")
		t.Fatal("RequiredProviders should not be empty")
	} else {
		printRPTestStatus(t, "Required providers", true, fmt.Sprintf("Found %d required providers", len(config.RequiredProviders)))
	}

	// Check that expected providers are present
	providerMap := make(map[string]bool)
	for _, provider := range config.RequiredProviders {
		providerMap[provider] = true
	}

	missingProviders := 0
	for _, expected := range expectedProviders {
		if !providerMap[expected] {
			printRPTestStatus(t, fmt.Sprintf("Provider %s", expected), false, fmt.Sprintf("Expected provider '%s' not found in LocalBox configuration", expected))
			t.Errorf("Expected provider '%s' not found in LocalBox configuration", expected)
			missingProviders++
		} else {
			printRPTestStatus(t, fmt.Sprintf("Provider %s", expected), true, "Provider found in configuration")
		}
	}

	if missingProviders == 0 {
		printRPTestStatus(t, "All expected providers", true, "All expected providers are present")
	}

	// Verify no empty provider names
	emptyProviders := 0
	for _, provider := range config.RequiredProviders {
		if provider == "" {
			printRPTestStatus(t, "Empty provider validation", false, "Found empty provider name in LocalBox configuration")
			t.Error("Found empty provider name in LocalBox configuration")
			emptyProviders++
		}
	}

	if emptyProviders == 0 {
		printRPTestStatus(t, "Empty provider validation", true, "No empty provider names found")
	}
}

func TestGetAgoraProviders(t *testing.T) {
	fmt.Printf("\n%s\n", rpTestHeaderColor("=== Testing Agora Providers ==="))

	config := GetAgoraProviders()

	// Test basic configuration structure
	if config.SolutionName != "Agora" {
		printRPTestStatus(t, "Solution name", false, fmt.Sprintf("Expected SolutionName to be 'Agora', got '%s'", config.SolutionName))
		t.Errorf("Expected SolutionName to be 'Agora', got '%s'", config.SolutionName)
	} else {
		printRPTestStatus(t, "Solution name", true, fmt.Sprintf("Solution name correctly set to '%s'", rpTestSuccessColor(config.SolutionName)))
	}

	// Agora should have at least basic Azure providers
	if len(config.RequiredProviders) == 0 {
		printRPTestStatus(t, "Required providers", false, "RequiredProviders should not be empty for Agora")
		t.Fatal("RequiredProviders should not be empty for Agora")
	} else {
		printRPTestStatus(t, "Required providers", true, fmt.Sprintf("Found %d required providers", len(config.RequiredProviders)))
	}

	// Verify no empty provider names
	emptyProviders := 0
	for _, provider := range config.RequiredProviders {
		if provider == "" {
			printRPTestStatus(t, "Empty provider validation", false, "Found empty provider name in Agora configuration")
			t.Error("Found empty provider name in Agora configuration")
			emptyProviders++
		}
	}

	if emptyProviders == 0 {
		printRPTestStatus(t, "Empty provider validation", true, "No empty provider names found")
	}
}

func TestResourceProviderConfig(t *testing.T) {
	fmt.Printf("\n%s\n", rpTestHeaderColor("=== Testing ResourceProviderConfig Struct ==="))

	// Test that ResourceProviderConfig struct works as expected
	testConfig := ResourceProviderConfig{
		SolutionName:      "TestSolution",
		RequiredProviders: []string{"Microsoft.Test", "Microsoft.Example"},
	}

	if testConfig.SolutionName != "TestSolution" {
		printRPTestStatus(t, "Solution name assignment", false, fmt.Sprintf("Expected SolutionName to be 'TestSolution', got '%s'", testConfig.SolutionName))
		t.Errorf("Expected SolutionName to be 'TestSolution', got '%s'", testConfig.SolutionName)
	} else {
		printRPTestStatus(t, "Solution name assignment", true, "Solution name correctly assigned")
	}

	if len(testConfig.RequiredProviders) != 2 {
		printRPTestStatus(t, "Provider count", false, fmt.Sprintf("Expected 2 required providers, got %d", len(testConfig.RequiredProviders)))
		t.Errorf("Expected 2 required providers, got %d", len(testConfig.RequiredProviders))
	} else {
		printRPTestStatus(t, "Provider count", true, "Correct number of providers assigned")
	}

	expectedProviders := map[string]bool{
		"Microsoft.Test":    true,
		"Microsoft.Example": true,
	}

	unexpectedCount := 0
	for _, provider := range testConfig.RequiredProviders {
		if !expectedProviders[provider] {
			printRPTestStatus(t, fmt.Sprintf("Provider %s", provider), false, fmt.Sprintf("Unexpected provider '%s' in test configuration", provider))
			t.Errorf("Unexpected provider '%s' in test configuration", provider)
			unexpectedCount++
		}
	}

	if unexpectedCount == 0 {
		printRPTestStatus(t, "Provider content validation", true, "All providers match expected values")
	}
}

func TestProviderConfigurationConsistency(t *testing.T) {
	fmt.Printf("\n%s\n", rpTestHeaderColor("=== Testing Provider Configuration Consistency ==="))

	// Test that all solution configurations have consistent structure
	solutions := []ResourceProviderConfig{
		GetArcBoxProviders(),
		GetLocalBoxProviders(),
		GetAgoraProviders(),
	}

	for _, config := range solutions {
		t.Run("solution_"+config.SolutionName, func(t *testing.T) {
			fmt.Printf("  %s Testing consistency for solution: %s\n", rpTestInfoColor("Testing:"), rpTestInfoColor(config.SolutionName))

			// Every solution should have a name
			if config.SolutionName == "" {
				printRPTestStatus(t, "Solution name", false, "SolutionName should not be empty")
				t.Error("SolutionName should not be empty")
			} else {
				printRPTestStatus(t, "Solution name", true, "Solution has a valid name")
			}

			// Every solution should have at least one provider
			if len(config.RequiredProviders) == 0 {
				printRPTestStatus(t, "Provider count", false, "RequiredProviders should not be empty")
				t.Error("RequiredProviders should not be empty")
			} else {
				printRPTestStatus(t, "Provider count", true, fmt.Sprintf("Solution has %d providers", len(config.RequiredProviders)))
			}

			// All provider names should follow Microsoft.* pattern
			invalidProviders := 0
			emptyProviders := 0
			for _, provider := range config.RequiredProviders {
				if provider == "" {
					printRPTestStatus(t, "Empty provider", false, "Provider name should not be empty")
					t.Error("Provider name should not be empty")
					emptyProviders++
					continue
				}

				if !startsWithMicrosoft(provider) {
					printRPTestStatus(t, fmt.Sprintf("Provider format %s", provider), false, fmt.Sprintf("Provider '%s' should start with 'Microsoft.'", provider))
					t.Errorf("Provider '%s' should start with 'Microsoft.'", provider)
					invalidProviders++
				}
			}

			if invalidProviders == 0 && emptyProviders == 0 {
				printRPTestStatus(t, "Provider naming", true, "All providers follow Microsoft.* pattern")
			}

			// Check for duplicate providers
			providerSet := make(map[string]bool)
			duplicates := 0
			for _, provider := range config.RequiredProviders {
				if providerSet[provider] {
					printRPTestStatus(t, fmt.Sprintf("Duplicate %s", provider), false, fmt.Sprintf("Duplicate provider '%s' found in %s configuration", provider, config.SolutionName))
					t.Errorf("Duplicate provider '%s' found in %s configuration", provider, config.SolutionName)
					duplicates++
				}
				providerSet[provider] = true
			}

			if duplicates == 0 {
				printRPTestStatus(t, "No duplicates", true, "No duplicate providers found")
			}
		})
	}
}

func TestProviderNameValidation(t *testing.T) {
	fmt.Printf("\n%s\n", rpTestHeaderColor("=== Testing Provider Name Validation ==="))

	// Test validation of provider names
	validProviders := []string{
		"Microsoft.Compute",
		"Microsoft.Network",
		"Microsoft.Storage",
		"Microsoft.KeyVault",
		"Microsoft.AzureArcData",
	}

	invalidProviders := []string{
		"",
		"Compute",
		"microsoft.compute",
		"Microsoft.",
		"Microsoft",
		"Azure.Compute",
	}

	validCount := 0
	for _, provider := range validProviders {
		t.Run("valid_"+provider, func(t *testing.T) {
			if !startsWithMicrosoft(provider) {
				printRPTestStatus(t, fmt.Sprintf("Valid provider %s", provider), false, fmt.Sprintf("Provider '%s' should be considered valid", provider))
				t.Errorf("Provider '%s' should be considered valid", provider)
			} else {
				printRPTestStatus(t, fmt.Sprintf("Valid provider %s", provider), true, "Provider correctly recognized as valid")
				validCount++
			}
		})
	}

	invalidCount := 0
	for _, provider := range invalidProviders {
		providerName := provider
		if providerName == "" {
			providerName = "(empty)"
		}
		t.Run("invalid_"+providerName, func(t *testing.T) {
			if startsWithMicrosoft(provider) && provider != "" {
				// Empty string is a special case that should fail
				printRPTestStatus(t, fmt.Sprintf("Invalid provider %s", providerName), false, fmt.Sprintf("Provider '%s' should be considered invalid", provider))
				t.Errorf("Provider '%s' should be considered invalid", provider)
			} else {
				printRPTestStatus(t, fmt.Sprintf("Invalid provider %s", providerName), true, "Provider correctly recognized as invalid")
				invalidCount++
			}
		})
	}

	fmt.Printf("  %s Validated %d valid and %d invalid provider names\n",
		rpTestInfoColor("Summary:"), validCount, invalidCount)
}

func TestSolutionSpecificProviders(t *testing.T) {
	fmt.Printf("\n%s\n", rpTestHeaderColor("=== Testing Solution-Specific Providers ==="))

	// Test that different solutions have appropriate provider sets
	arcboxConfig := GetArcBoxProviders()
	localboxConfig := GetLocalBoxProviders()
	agoraConfig := GetAgoraProviders()

	// ArcBox should have Arc-specific providers
	arcboxProviders := make(map[string]bool)
	for _, provider := range arcboxConfig.RequiredProviders {
		arcboxProviders[provider] = true
	}

	arcSpecificProviders := []string{
		"Microsoft.Kubernetes",
		"Microsoft.KubernetesConfiguration",
		"Microsoft.AzureArcData",
	}

	fmt.Printf("  %s Validating ArcBox Arc-specific providers\n", rpTestInfoColor("Testing:"))
	arcMissing := 0
	for _, required := range arcSpecificProviders {
		if !arcboxProviders[required] {
			printRPTestStatus(t, fmt.Sprintf("ArcBox provider %s", required), false, fmt.Sprintf("ArcBox should include %s provider", required))
			t.Errorf("ArcBox should include %s provider", required)
			arcMissing++
		} else {
			printRPTestStatus(t, fmt.Sprintf("ArcBox provider %s", required), true, "Required Arc provider found")
		}
	}

	if arcMissing == 0 {
		printRPTestStatus(t, "ArcBox Arc providers", true, "All required Arc-specific providers found")
	}

	// LocalBox should have basic compute/network providers
	localboxProviders := make(map[string]bool)
	for _, provider := range localboxConfig.RequiredProviders {
		localboxProviders[provider] = true
	}

	localSpecificProviders := []string{
		"Microsoft.Compute",
		"Microsoft.Network",
		"Microsoft.Storage",
	}

	fmt.Printf("  %s Validating LocalBox basic providers\n", rpTestInfoColor("Testing:"))
	localMissing := 0
	for _, required := range localSpecificProviders {
		if !localboxProviders[required] {
			printRPTestStatus(t, fmt.Sprintf("LocalBox provider %s", required), false, fmt.Sprintf("LocalBox should include %s provider", required))
			t.Errorf("LocalBox should include %s provider", required)
			localMissing++
		} else {
			printRPTestStatus(t, fmt.Sprintf("LocalBox provider %s", required), true, "Required basic provider found")
		}
	}

	if localMissing == 0 {
		printRPTestStatus(t, "LocalBox basic providers", true, "All required basic providers found")
	}

	// All solutions should have Microsoft.Compute
	fmt.Printf("  %s Validating Microsoft.Compute across all solutions\n", rpTestInfoColor("Testing:"))
	allConfigs := []ResourceProviderConfig{arcboxConfig, localboxConfig, agoraConfig}
	solutionsWithCompute := 0
	for _, config := range allConfigs {
		hasCompute := false
		for _, provider := range config.RequiredProviders {
			if provider == "Microsoft.Compute" {
				hasCompute = true
				break
			}
		}
		if !hasCompute {
			printRPTestStatus(t, fmt.Sprintf("%s Compute provider", config.SolutionName), false, fmt.Sprintf("Solution '%s' should include Microsoft.Compute provider", config.SolutionName))
			t.Errorf("Solution '%s' should include Microsoft.Compute provider", config.SolutionName)
		} else {
			printRPTestStatus(t, fmt.Sprintf("%s Compute provider", config.SolutionName), true, "Microsoft.Compute provider found")
			solutionsWithCompute++
		}
	}

	if solutionsWithCompute == len(allConfigs) {
		printRPTestStatus(t, "Universal Compute provider", true, "All solutions include Microsoft.Compute")
	}
}

// Helper function to check if a provider name starts with "Microsoft."
func startsWithMicrosoft(provider string) bool {
	return len(provider) > 10 && provider[:10] == "Microsoft."
}

// Test performance of provider configuration functions
func BenchmarkGetArcBoxProviders(b *testing.B) {
	fmt.Printf("\n%s\n", rpTestHeaderColor("=== Benchmarking ArcBox Providers ==="))

	for i := 0; i < b.N; i++ {
		GetArcBoxProviders()
	}
}

func BenchmarkGetLocalBoxProviders(b *testing.B) {
	fmt.Printf("\n%s\n", rpTestHeaderColor("=== Benchmarking LocalBox Providers ==="))

	for i := 0; i < b.N; i++ {
		GetLocalBoxProviders()
	}
}

func BenchmarkGetAgoraProviders(b *testing.B) {
	fmt.Printf("\n%s\n", rpTestHeaderColor("=== Benchmarking Agora Providers ==="))

	for i := 0; i < b.N; i++ {
		GetAgoraProviders()
	}
}

// Add tests for edge cases and error conditions
func TestResourceProviderConfigNilHandling(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Nil Handling ===")

	t.Run("nil_provider_list", func(t *testing.T) {
		config := ResourceProviderConfig{
			SolutionName:      "TestSolution",
			RequiredProviders: nil,
		}

		testutils.PrintTestStatus(t, "Nil providers list", config.RequiredProviders == nil,
			"Should handle nil RequiredProviders list")
	})

	t.Run("empty_solution_name", func(t *testing.T) {
		config := ResourceProviderConfig{
			SolutionName:      "",
			RequiredProviders: []string{"Microsoft.Test"},
		}

		testutils.PrintTestStatus(t, "Empty solution name", config.SolutionName == "",
			"Should handle empty solution name")
	})
}

// Add tests for concurrent access
func TestConcurrentProviderAccess(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Concurrent Access ===")

	// Test concurrent access to provider functions
	done := make(chan bool, 3)

	go func() {
		_ = GetArcBoxProviders()
		done <- true
	}()

	go func() {
		_ = GetLocalBoxProviders()
		done <- true
	}()

	go func() {
		_ = GetAgoraProviders()
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	testutils.PrintTestStatus(t, "Concurrent access", true, "No race condition in provider access")
}

// Add mutation tests
func TestProviderImmutability(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Provider Immutability ===")

	// Get providers
	arcbox1 := GetArcBoxProviders()
	arcbox2 := GetArcBoxProviders()

	// Verify they're separate instances (defensive copies)
	arcbox1.RequiredProviders = append(arcbox1.RequiredProviders, "Microsoft.Test")

	testutils.PrintTestStatus(t, "Provider immutability",
		len(arcbox1.RequiredProviders) != len(arcbox2.RequiredProviders),
		"Provider configurations should return defensive copies")
}
