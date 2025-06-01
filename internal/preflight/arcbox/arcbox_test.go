package arcbox

import (
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	preflightTestSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	preflightTestInfoColor    = color.New(color.FgCyan).SprintFunc()
	preflightTestWarnColor    = color.New(color.FgYellow).SprintFunc()
	preflightTestErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	preflightTestHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

func printPreflightTestStatus(t *testing.T, testName string, success bool, message string) {
	status := preflightTestSuccessColor("✅")
	if !success {
		status = preflightTestErrorColor("❌")
	}
	fmt.Printf("  %s %s: %s\n", status, preflightTestInfoColor(testName), message)
}

func TestBuildArcBoxValidationContext(t *testing.T) {
	fmt.Printf("\n%s\n", preflightTestHeaderColor("=== Testing ArcBox Validation Context ==="))

	// Create a mock command with flags
	cmd := &cobra.Command{}
	cmd.Flags().String("flavor", "ITPro", "ArcBox flavor")
	cmd.Flags().String("location", "eastus", "Azure region")
	cmd.Flags().String("ssh-rsa-public-key", "", "SSH public key")
	cmd.Flags().String("windows-password", "", "Windows password")
	cmd.Flags().String("resource-tags", `{"Solution":"jumpstart_arcbox"}`, "Resource tags")
	cmd.Flags().String("github-user", "microsoft", "GitHub user")
	cmd.Flags().String("resource-group", "test-rg", "Resource group")
	cmd.Flags().String("windows-user", "azureuser", "Windows user")

	// Set some flag values
	cmd.Flags().Set("flavor", "DevOps")
	cmd.Flags().Set("location", "eastus")
	cmd.Flags().Set("ssh-rsa-public-key", "ssh-rsa AAAAB3...")
	cmd.Flags().Set("github-user", "testuser")

	ctx := buildArcBoxValidationContext(cmd)

	// Test basic properties
	if ctx.Solution != "arcbox" {
		printPreflightTestStatus(t, "Solution validation", false, fmt.Sprintf("Expected Solution to be 'arcbox', got '%s'", ctx.Solution))
		t.Errorf("Expected Solution to be 'arcbox', got '%s'", ctx.Solution)
	} else {
		printPreflightTestStatus(t, "Solution validation", true, "Solution correctly set to 'arcbox'")
	}

	if ctx.Flavor != "DevOps" {
		printPreflightTestStatus(t, "Flavor validation", false, fmt.Sprintf("Expected Flavor to be 'DevOps', got '%s'", ctx.Flavor))
		t.Errorf("Expected Flavor to be 'DevOps', got '%s'", ctx.Flavor)
	} else {
		printPreflightTestStatus(t, "Flavor validation", true, "Flavor correctly set to 'DevOps'")
	}

	if ctx.Location != "eastus" {
		printPreflightTestStatus(t, "Location validation", false, fmt.Sprintf("Expected Location to be 'eastus', got '%s'", ctx.Location))
		t.Errorf("Expected Location to be 'eastus', got '%s'", ctx.Location)
	} else {
		printPreflightTestStatus(t, "Location validation", true, "Location correctly set to 'eastus'")
	}

	// Test parameters
	if ctx.Parameters["location"] != "eastus" {
		printPreflightTestStatus(t, "Location parameter", false, fmt.Sprintf("Expected location parameter to be 'eastus', got '%s'", ctx.Parameters["location"]))
		t.Errorf("Expected location parameter to be 'eastus', got '%s'", ctx.Parameters["location"])
	} else {
		printPreflightTestStatus(t, "Location parameter", true, "Location parameter correctly set")
	}

	if ctx.Parameters["ssh-rsa-public-key"] != "ssh-rsa AAAAB3..." {
		printPreflightTestStatus(t, "SSH key parameter", false, "Expected ssh key parameter to be set correctly")
		t.Errorf("Expected ssh key parameter to be set correctly")
	} else {
		printPreflightTestStatus(t, "SSH key parameter", true, "SSH key parameter correctly set")
	}
	if ctx.Parameters["github-user"] != "testuser" {
		printPreflightTestStatus(t, "GitHub user parameter", false, fmt.Sprintf("Expected github-user parameter to be 'testuser', got '%s'", ctx.Parameters["github-user"]))
		t.Errorf("Expected github-user parameter to be 'testuser', got '%s'", ctx.Parameters["github-user"])
	} else {
		printPreflightTestStatus(t, "GitHub user parameter", true, "GitHub user parameter correctly set")
	}

	// Test that SilentMode is false (for progress indicators)
	if ctx.SilentMode {
		printPreflightTestStatus(t, "Silent mode setting", false, "Expected SilentMode to be false for ArcBox validation")
		t.Error("Expected SilentMode to be false for ArcBox validation")
	} else {
		printPreflightTestStatus(t, "Silent mode setting", true, "SilentMode correctly set to false")
	}
}

func TestGetFlavorSpecificChecks(t *testing.T) {
	fmt.Printf("\n%s\n", preflightTestHeaderColor("=== Testing Flavor-Specific Checks ==="))

	tests := []struct {
		flavor   string
		expected []string
	}{
		{"DevOps", []string{"ssh-rsa-public-key", "github-user"}},
		{"DataOps", []string{"ssh-rsa-public-key"}},
		{"ITPro", []string{}},
		{"Unknown", []string{}},
		{"", []string{}},
	}

	for _, test := range tests {
		t.Run("flavor_"+test.flavor, func(t *testing.T) {
			fmt.Printf("  %s Testing flavor: %s\n", preflightTestInfoColor("Testing:"), test.flavor)

			result := GetFlavorSpecificChecks(test.flavor)

			if len(result) != len(test.expected) {
				printPreflightTestStatus(t, fmt.Sprintf("Check count for %s", test.flavor), false, fmt.Sprintf("Expected %d checks, got %d", len(test.expected), len(result)))
				t.Errorf("Expected %d checks for flavor '%s', got %d", len(test.expected), test.flavor, len(result))
				return
			} else {
				printPreflightTestStatus(t, fmt.Sprintf("Check count for %s", test.flavor), true, fmt.Sprintf("Found %d expected checks", len(result)))
			}

			// Check that all expected checks are present
			expectedMap := make(map[string]bool)
			for _, check := range test.expected {
				expectedMap[check] = true
			}

			unexpectedChecks := 0
			for _, check := range result {
				if !expectedMap[check] {
					printPreflightTestStatus(t, fmt.Sprintf("Unexpected check %s", check), false, fmt.Sprintf("Unexpected check '%s' for flavor '%s'", check, test.flavor))
					t.Errorf("Unexpected check '%s' for flavor '%s'", check, test.flavor)
					unexpectedChecks++
				}
			}

			if unexpectedChecks == 0 && len(result) > 0 {
				printPreflightTestStatus(t, fmt.Sprintf("All checks for %s", test.flavor), true, fmt.Sprintf("All %d checks are expected", len(result)))
			} else if len(result) == 0 {
				printPreflightTestStatus(t, fmt.Sprintf("No checks for %s", test.flavor), true, "No checks required (as expected)")
			}
		})
	}
}

func TestValidateConditionalRequirements(t *testing.T) {
	fmt.Printf("\n%s\n", preflightTestHeaderColor("=== Testing Conditional Requirements Validation ==="))

	// Test DevOps flavor requirements
	t.Run("devops_missing_ssh_key", func(t *testing.T) {
		fmt.Printf("  %s Testing DevOps flavor missing SSH key\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("flavor", "DevOps", "ArcBox flavor")
		cmd.Flags().String("ssh-rsa-public-key", "", "SSH public key")
		cmd.Flags().String("github-user", "testuser", "GitHub user")

		cmd.Flags().Set("flavor", "DevOps")
		cmd.Flags().Set("ssh-rsa-public-key", "") // Missing SSH key
		cmd.Flags().Set("github-user", "testuser")

		// This should return false because SSH key is required for DevOps
		// Note: This test captures output but doesn't validate it in detail
		// as the function prints directly to stdout
		result := ValidateConditionalRequirements(cmd)
		if result {
			printPreflightTestStatus(t, "DevOps missing SSH key", false, "Expected validation to fail for DevOps without SSH key")
			t.Error("Expected ValidateConditionalRequirements to return false for DevOps without SSH key")
		} else {
			printPreflightTestStatus(t, "DevOps missing SSH key", true, "Validation correctly failed for missing SSH key")
		}
	})

	t.Run("devops_missing_github_user", func(t *testing.T) {
		fmt.Printf("  %s Testing DevOps flavor missing GitHub user\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("flavor", "DevOps", "ArcBox flavor")
		cmd.Flags().String("ssh-rsa-public-key", "ssh-rsa AAAAB3...", "SSH public key")
		cmd.Flags().String("github-user", "", "GitHub user")

		cmd.Flags().Set("flavor", "DevOps")
		cmd.Flags().Set("ssh-rsa-public-key", "ssh-rsa AAAAB3...")
		cmd.Flags().Set("github-user", "") // Missing GitHub user

		result := ValidateConditionalRequirements(cmd)
		if result {
			printPreflightTestStatus(t, "DevOps missing GitHub user", false, "Expected validation to fail for DevOps without GitHub user")
			t.Error("Expected ValidateConditionalRequirements to return false for DevOps without GitHub user")
		} else {
			printPreflightTestStatus(t, "DevOps missing GitHub user", true, "Validation correctly failed for missing GitHub user")
		}
	})

	t.Run("devops_microsoft_github_user", func(t *testing.T) {
		fmt.Printf("  %s Testing DevOps flavor with 'microsoft' GitHub user\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("flavor", "DevOps", "ArcBox flavor")
		cmd.Flags().String("ssh-rsa-public-key", "ssh-rsa AAAAB3...", "SSH public key")
		cmd.Flags().String("github-user", "microsoft", "GitHub user")

		cmd.Flags().Set("flavor", "DevOps")
		cmd.Flags().Set("ssh-rsa-public-key", "ssh-rsa AAAAB3...")
		cmd.Flags().Set("github-user", "microsoft") // Should not use "microsoft"

		result := ValidateConditionalRequirements(cmd)
		if result {
			printPreflightTestStatus(t, "DevOps microsoft GitHub user", false, "Expected validation to fail for DevOps with 'microsoft' GitHub user")
			t.Error("Expected ValidateConditionalRequirements to return false for DevOps with 'microsoft' GitHub user")
		} else {
			printPreflightTestStatus(t, "DevOps microsoft GitHub user", true, "Validation correctly failed for 'microsoft' GitHub user")
		}
	})
	t.Run("dataops_missing_ssh_key", func(t *testing.T) {
		fmt.Printf("  %s Testing DataOps flavor missing SSH key\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("flavor", "DataOps", "ArcBox flavor")
		cmd.Flags().String("ssh-rsa-public-key", "", "SSH public key")
		cmd.Flags().String("github-user", "", "GitHub user")

		cmd.Flags().Set("flavor", "DataOps")
		cmd.Flags().Set("ssh-rsa-public-key", "") // Missing SSH key

		result := ValidateConditionalRequirements(cmd)
		if result {
			printPreflightTestStatus(t, "DataOps missing SSH key", false, "Expected validation to fail for DataOps without SSH key")
			t.Error("Expected ValidateConditionalRequirements to return false for DataOps without SSH key")
		} else {
			printPreflightTestStatus(t, "DataOps missing SSH key", true, "Validation correctly failed for missing SSH key")
		}
	})

	t.Run("dataops_valid", func(t *testing.T) {
		fmt.Printf("  %s Testing valid DataOps configuration\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("flavor", "DataOps", "ArcBox flavor")
		cmd.Flags().String("ssh-rsa-public-key", "ssh-rsa AAAAB3...", "SSH public key")
		cmd.Flags().String("github-user", "", "GitHub user")

		cmd.Flags().Set("flavor", "DataOps")
		cmd.Flags().Set("ssh-rsa-public-key", "ssh-rsa AAAAB3...")

		result := ValidateConditionalRequirements(cmd)
		if !result {
			printPreflightTestStatus(t, "DataOps valid config", false, "Expected validation to pass for valid DataOps configuration")
			t.Error("Expected ValidateConditionalRequirements to return true for valid DataOps configuration")
		} else {
			printPreflightTestStatus(t, "DataOps valid config", true, "Validation correctly passed for valid DataOps configuration")
		}
	})

	t.Run("itpro_no_requirements", func(t *testing.T) {
		fmt.Printf("  %s Testing ITPro flavor (no additional requirements)\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("flavor", "ITPro", "ArcBox flavor")
		cmd.Flags().String("ssh-rsa-public-key", "", "SSH public key")
		cmd.Flags().String("github-user", "", "GitHub user")

		cmd.Flags().Set("flavor", "ITPro")

		result := ValidateConditionalRequirements(cmd)
		if !result {
			printPreflightTestStatus(t, "ITPro no requirements", false, "Expected validation to pass for ITPro (no additional requirements)")
			t.Error("Expected ValidateConditionalRequirements to return true for ITPro (no additional requirements)")
		} else {
			printPreflightTestStatus(t, "ITPro no requirements", true, "Validation correctly passed for ITPro (no additional requirements)")
		}
	})

	t.Run("devops_valid", func(t *testing.T) {
		fmt.Printf("  %s Testing valid DevOps configuration\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("flavor", "DevOps", "ArcBox flavor")
		cmd.Flags().String("ssh-rsa-public-key", "ssh-rsa AAAAB3...", "SSH public key")
		cmd.Flags().String("github-user", "testuser", "GitHub user")

		cmd.Flags().Set("flavor", "DevOps")
		cmd.Flags().Set("ssh-rsa-public-key", "ssh-rsa AAAAB3...")
		cmd.Flags().Set("github-user", "testuser")

		result := ValidateConditionalRequirements(cmd)
		if !result {
			printPreflightTestStatus(t, "DevOps valid config", false, "Expected validation to pass for valid DevOps configuration")
			t.Error("Expected ValidateConditionalRequirements to return true for valid DevOps configuration")
		} else {
			printPreflightTestStatus(t, "DevOps valid config", true, "Validation correctly passed for valid DevOps configuration")
		}
	})
}

func TestRunArcBoxPreflightChecks(t *testing.T) {
	fmt.Printf("\n%s\n", preflightTestHeaderColor("=== Testing ArcBox Preflight Checks ==="))

	// Test that the function can be called without panicking
	// This is more of an integration test since it depends on external validators
	t.Run("basic_execution", func(t *testing.T) {
		fmt.Printf("  %s Testing basic preflight checks execution\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("flavor", "ITPro", "ArcBox flavor")
		cmd.Flags().String("location", "eastus", "Azure region")
		cmd.Flags().String("ssh-rsa-public-key", "", "SSH public key")
		cmd.Flags().String("windows-password", "", "Windows password")
		cmd.Flags().String("resource-tags", `{"Solution":"jumpstart_arcbox"}`, "Resource tags")
		cmd.Flags().String("github-user", "microsoft", "GitHub user")
		cmd.Flags().String("resource-group", "test-rg", "Resource group")

		cmd.Flags().Set("flavor", "ITPro")
		cmd.Flags().Set("location", "eastus")

		// This will likely fail due to missing Azure CLI or other dependencies,
		// but we're testing that it doesn't panic and returns a boolean
		defer func() {
			if r := recover(); r != nil {
				printPreflightTestStatus(t, "Panic prevention", false, fmt.Sprintf("RunArcBoxPreflightChecks panicked: %v", r))
				t.Errorf("RunArcBoxPreflightChecks panicked: %v", r)
			} else {
				printPreflightTestStatus(t, "Panic prevention", true, "RunArcBoxPreflightChecks executed without panicking")
			}
		}()

		result := RunArcBoxPreflightChecks(cmd)
		// Result can be true or false, we just want to ensure no panic
		printPreflightTestStatus(t, "Execution result", true, fmt.Sprintf("RunArcBoxPreflightChecks returned: %v", result))
		_ = result
	})
}

func TestRunParameterValidation(t *testing.T) {
	fmt.Printf("\n%s\n", preflightTestHeaderColor("=== Testing Parameter Validation ==="))

	// Test parameter-only validation
	t.Run("basic_execution", func(t *testing.T) {
		fmt.Printf("  %s Testing parameter validation execution\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("flavor", "ITPro", "ArcBox flavor")
		cmd.Flags().String("location", "eastus", "Azure region")
		cmd.Flags().String("ssh-rsa-public-key", "", "SSH public key")
		cmd.Flags().String("windows-password", "P@ssw0rd123!", "Windows password")
		cmd.Flags().String("resource-tags", `{"Solution":"jumpstart_arcbox"}`, "Resource tags")
		cmd.Flags().String("github-user", "microsoft", "GitHub user")

		cmd.Flags().Set("flavor", "ITPro")
		cmd.Flags().Set("location", "eastus")
		cmd.Flags().Set("windows-password", "P@ssw0rd123!")

		// This should focus only on parameter validation
		defer func() {
			if r := recover(); r != nil {
				printPreflightTestStatus(t, "Panic prevention", false, fmt.Sprintf("RunParameterValidation panicked: %v", r))
				t.Errorf("RunParameterValidation panicked: %v", r)
			} else {
				printPreflightTestStatus(t, "Panic prevention", true, "RunParameterValidation executed without panicking")
			}
		}()

		result := RunParameterValidation(cmd)
		// Result can be true or false, we just want to ensure no panic
		printPreflightTestStatus(t, "Execution result", true, fmt.Sprintf("RunParameterValidation returned: %v", result))
		_ = result
	})
}

func TestRunArcBoxQuotaChecks(t *testing.T) {
	fmt.Printf("\n%s\n", preflightTestHeaderColor("=== Testing ArcBox Quota Checks ==="))

	// Test quota-specific validation
	t.Run("basic_execution", func(t *testing.T) {
		fmt.Printf("  %s Testing quota checks execution\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("flavor", "ITPro", "ArcBox flavor")
		cmd.Flags().String("location", "eastus", "Azure region")
		cmd.Flags().String("subscription", "", "Azure subscription")

		cmd.Flags().Set("flavor", "ITPro")
		cmd.Flags().Set("location", "eastus")

		// This will likely fail due to missing Azure CLI or other dependencies,
		// but we're testing that it doesn't panic
		defer func() {
			if r := recover(); r != nil {
				printPreflightTestStatus(t, "Panic prevention", false, fmt.Sprintf("RunArcBoxQuotaChecks panicked: %v", r))
				t.Errorf("RunArcBoxQuotaChecks panicked: %v", r)
			} else {
				printPreflightTestStatus(t, "Panic prevention", true, "RunArcBoxQuotaChecks executed without panicking")
			}
		}()

		result := RunArcBoxQuotaChecks(cmd)
		// Result can be true or false, we just want to ensure no panic
		printPreflightTestStatus(t, "Execution result", true, fmt.Sprintf("RunArcBoxQuotaChecks returned: %v", result))
		_ = result
	})
}

func TestArcBoxValidationContextResourceTags(t *testing.T) {
	fmt.Printf("\n%s\n", preflightTestHeaderColor("=== Testing Resource Tags Validation Context ==="))

	// Test that resource-tags parameter is only included when explicitly set
	t.Run("resource_tags_not_changed", func(t *testing.T) {
		fmt.Printf("  %s Testing resource tags not explicitly set\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("resource-tags", `{"Solution":"jumpstart_arcbox"}`, "Resource tags")

		// Don't explicitly set the flag (use default value)
		ctx := buildArcBoxValidationContext(cmd)

		// Should not include resource-tags in parameters when using default value
		if _, exists := ctx.Parameters["resource-tags"]; exists {
			printPreflightTestStatus(t, "Default resource tags exclusion", false, "Expected resource-tags to not be included when using default value")
			t.Error("Expected resource-tags to not be included when using default value")
		} else {
			printPreflightTestStatus(t, "Default resource tags exclusion", true, "Resource-tags correctly excluded when using default value")
		}
	})

	t.Run("resource_tags_explicitly_set", func(t *testing.T) {
		fmt.Printf("  %s Testing resource tags explicitly set\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		cmd.Flags().String("resource-tags", `{"Solution":"jumpstart_arcbox"}`, "Resource tags")

		// Explicitly set the flag
		cmd.Flags().Set("resource-tags", `{"Environment":"test","Solution":"jumpstart_arcbox"}`)

		ctx := buildArcBoxValidationContext(cmd)

		// Should include resource-tags in parameters when explicitly set
		if resourceTags, exists := ctx.Parameters["resource-tags"]; !exists {
			printPreflightTestStatus(t, "Explicit resource tags inclusion", false, "Expected resource-tags to be included when explicitly set")
			t.Error("Expected resource-tags to be included when explicitly set")
		} else if resourceTags != `{"Environment":"test","Solution":"jumpstart_arcbox"}` {
			printPreflightTestStatus(t, "Resource tags value", false, fmt.Sprintf("Expected custom resource-tags value, got '%s'", resourceTags))
			t.Errorf("Expected resource-tags value to be custom value, got '%s'", resourceTags)
		} else {
			printPreflightTestStatus(t, "Explicit resource tags inclusion", true, "Resource-tags correctly included with custom value")
		}
	})
}

func TestPrintFlavorRequirements(t *testing.T) {
	fmt.Printf("\n%s\n", preflightTestHeaderColor("=== Testing Print Flavor Requirements ==="))

	// Test that PrintFlavorRequirements doesn't panic for different flavors
	flavors := []string{"DevOps", "DataOps", "ITPro", "Unknown", ""}

	for _, flavor := range flavors {
		t.Run("flavor_"+flavor, func(t *testing.T) {
			fmt.Printf("  %s Testing print requirements for flavor: %s\n", preflightTestInfoColor("Testing:"), flavor)

			defer func() {
				if r := recover(); r != nil {
					printPreflightTestStatus(t, fmt.Sprintf("Panic prevention for %s", flavor), false, fmt.Sprintf("PrintFlavorRequirements panicked for flavor '%s': %v", flavor, r))
					t.Errorf("PrintFlavorRequirements panicked for flavor '%s': %v", flavor, r)
				} else {
					printPreflightTestStatus(t, fmt.Sprintf("Print requirements for %s", flavor), true, fmt.Sprintf("PrintFlavorRequirements executed without panicking for flavor '%s'", flavor))
				}
			}()

			PrintFlavorRequirements(flavor)
		})
	}
}

func TestArcBoxValidationEdgeCases(t *testing.T) {
	fmt.Printf("\n%s\n", preflightTestHeaderColor("=== Testing ArcBox Validation Edge Cases ==="))

	t.Run("empty_command", func(t *testing.T) {
		fmt.Printf("  %s Testing with empty command\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}

		defer func() {
			if r := recover(); r != nil {
				printPreflightTestStatus(t, "Empty command panic prevention", false, fmt.Sprintf("buildArcBoxValidationContext panicked with empty command: %v", r))
				t.Errorf("buildArcBoxValidationContext panicked with empty command: %v", r)
			} else {
				printPreflightTestStatus(t, "Empty command panic prevention", true, "buildArcBoxValidationContext handled empty command without panicking")
			}
		}()

		ctx := buildArcBoxValidationContext(cmd)

		// Should have default values
		if ctx.Solution != "arcbox" {
			printPreflightTestStatus(t, "Default solution", false, "Expected Solution to be 'arcbox' even with empty command")
			t.Error("Expected Solution to be 'arcbox' even with empty command")
		} else {
			printPreflightTestStatus(t, "Default solution", true, "Solution correctly defaulted to 'arcbox'")
		}

		if ctx.SilentMode {
			printPreflightTestStatus(t, "Default silent mode", false, "Expected SilentMode to be false by default")
			t.Error("Expected SilentMode to be false by default")
		} else {
			printPreflightTestStatus(t, "Default silent mode", true, "SilentMode correctly defaulted to false")
		}
	})

	t.Run("nil_parameters_map", func(t *testing.T) {
		fmt.Printf("  %s Testing parameters map initialization\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		ctx := buildArcBoxValidationContext(cmd)

		// Parameters map should be initialized
		if ctx.Parameters == nil {
			printPreflightTestStatus(t, "Parameters map initialization", false, "Expected Parameters map to be initialized")
			t.Error("Expected Parameters map to be initialized")
		} else {
			printPreflightTestStatus(t, "Parameters map initialization", true, "Parameters map correctly initialized")
		}
	})

	t.Run("nil_skip_checks_slice", func(t *testing.T) {
		fmt.Printf("  %s Testing skip checks slice initialization\n", preflightTestInfoColor("Testing:"))

		cmd := &cobra.Command{}
		ctx := buildArcBoxValidationContext(cmd)

		// SkipChecks slice should be initialized
		if ctx.SkipChecks == nil {
			printPreflightTestStatus(t, "SkipChecks slice initialization", false, "Expected SkipChecks slice to be initialized")
			t.Error("Expected SkipChecks slice to be initialized")
		} else {
			printPreflightTestStatus(t, "SkipChecks slice initialization", true, "SkipChecks slice correctly initialized")
		}
	})
}
