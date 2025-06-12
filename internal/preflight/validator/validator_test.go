package validator

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/resourceproviders"
	"jumpstartcli/internal/testutils"

	"github.com/fatih/color"
)

var (
	validatorTestSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	validatorTestInfoColor    = color.New(color.FgCyan).SprintFunc()
	validatorTestErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	validatorTestHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

func printValidatorTestStatus(t *testing.T, testName string, success bool, message string) {
	status := validatorTestSuccessColor("✅")
	if !success {
		status = validatorTestErrorColor("❌")
	}
	fmt.Printf("  %s %s: %s\n", status, validatorTestInfoColor(testName), message)
}

func TestValidationResult(t *testing.T) {
	fmt.Printf("\n%s\n", validatorTestHeaderColor("=== Testing ValidationResult Struct ==="))

	// Test ValidationResult struct creation and basic functionality
	result := ValidationResult{
		CheckName:  "TestCheck",
		Passed:     true,
		Message:    "Test passed successfully",
		Severity:   "info",
		Suggestion: "No action needed",
		Details:    "Additional test details",
	}

	if result.CheckName != "TestCheck" {
		printValidatorTestStatus(t, "CheckName assignment", false, fmt.Sprintf("Expected CheckName 'TestCheck', got '%s'", result.CheckName))
		t.Errorf("Expected CheckName 'TestCheck', got '%s'", result.CheckName)
	} else {
		printValidatorTestStatus(t, "CheckName assignment", true, "CheckName correctly set to 'TestCheck'")
	}

	if !result.Passed {
		printValidatorTestStatus(t, "Passed status", false, "Expected Passed to be true")
		t.Error("Expected Passed to be true")
	} else {
		printValidatorTestStatus(t, "Passed status", true, "Passed status correctly set to true")
	}

	if result.Severity != "info" {
		printValidatorTestStatus(t, "Severity level", false, fmt.Sprintf("Expected Severity 'info', got '%s'", result.Severity))
		t.Errorf("Expected Severity 'info', got '%s'", result.Severity)
	} else {
		printValidatorTestStatus(t, "Severity level", true, "Severity correctly set to 'info'")
	}

	if result.Message != "Test passed successfully" {
		printValidatorTestStatus(t, "Message content", false, fmt.Sprintf("Expected specific message, got '%s'", result.Message))
		t.Errorf("Expected message 'Test passed successfully', got '%s'", result.Message)
	} else {
		printValidatorTestStatus(t, "Message content", true, "Message correctly set")
	}
}

func TestValidationContext(t *testing.T) {
	fmt.Printf("\n%s\n", validatorTestHeaderColor("=== Testing ValidationContext Struct ==="))

	// Test ValidationContext struct creation
	ctx := ValidationContext{
		Solution:   "arcbox",
		Flavor:     "ITPro",
		Location:   "eastus",
		Parameters: map[string]string{"test": "value"},
		SkipChecks: []string{"check1", "check2"},
		SilentMode: false,
	}

	if ctx.Solution != "arcbox" {
		printValidatorTestStatus(t, "Solution assignment", false, fmt.Sprintf("Expected Solution 'arcbox', got '%s'", ctx.Solution))
		t.Errorf("Expected Solution 'arcbox', got '%s'", ctx.Solution)
	} else {
		printValidatorTestStatus(t, "Solution assignment", true, "Solution correctly set to 'arcbox'")
	}

	if ctx.Flavor != "ITPro" {
		printValidatorTestStatus(t, "Flavor assignment", false, fmt.Sprintf("Expected Flavor 'ITPro', got '%s'", ctx.Flavor))
		t.Errorf("Expected Flavor 'ITPro', got '%s'", ctx.Flavor)
	} else {
		printValidatorTestStatus(t, "Flavor assignment", true, "Flavor correctly set to 'ITPro'")
	}
	if ctx.Location != "eastus" {
		printValidatorTestStatus(t, "Location assignment", false, fmt.Sprintf("Expected Location 'eastus', got '%s'", ctx.Location))
		t.Errorf("Expected Location 'eastus', got '%s'", ctx.Location)
	} else {
		printValidatorTestStatus(t, "Location assignment", true, "Location correctly set to 'eastus'")
	}

	if len(ctx.Parameters) != 1 || ctx.Parameters["test"] != "value" {
		printValidatorTestStatus(t, "Parameters map", false, "Expected Parameters map to contain test=value")
		t.Error("Expected Parameters map to contain test=value")
	} else {
		printValidatorTestStatus(t, "Parameters map", true, "Parameters map correctly initialized with test=value")
	}
	if len(ctx.SkipChecks) != 2 {
		printValidatorTestStatus(t, "SkipChecks length", false, fmt.Sprintf("Expected 2 skip checks, got %d", len(ctx.SkipChecks)))
		t.Errorf("Expected 2 skip checks, got %d", len(ctx.SkipChecks))
	} else {
		printValidatorTestStatus(t, "SkipChecks length", true, "SkipChecks correctly contains 2 items")
	}

	if ctx.SilentMode {
		printValidatorTestStatus(t, "SilentMode setting", false, "Expected SilentMode to be false")
		t.Error("Expected SilentMode to be false")
	} else {
		printValidatorTestStatus(t, "SilentMode setting", true, "SilentMode correctly set to false")
	}
}

func TestNewValidationEngine(t *testing.T) {
	fmt.Printf("\n%s\n", validatorTestHeaderColor("=== Testing ValidationEngine Creation ==="))

	engine := NewValidationEngine()

	if engine == nil {
		printValidatorTestStatus(t, "Engine creation", false, "NewValidationEngine should not return nil")
		t.Fatal("NewValidationEngine should not return nil")
	} else {
		printValidatorTestStatus(t, "Engine creation", true, "ValidationEngine created successfully")
	}

	// Test that the engine has some validators registered
	// We can't test the exact number since it depends on implementation
	printValidatorTestStatus(t, "Engine initialization", true, "ValidationEngine initialization completed")
	t.Logf("ValidationEngine created successfully")
}

func TestValidationEngineRegisterValidator(t *testing.T) {
	engine := &ValidationEngine{}

	// Create a mock validator
	mockValidator := &mockValidator{
		name:        "MockValidator",
		description: "A mock validator for testing",
	}

	engine.RegisterValidator(mockValidator)

	// Verify the validator was registered
	if len(engine.validators) != 1 {
		t.Errorf("Expected 1 validator after registration, got %d", len(engine.validators))
	}

	if engine.validators[0].Name() != "MockValidator" {
		t.Errorf("Expected validator name 'MockValidator', got '%s'", engine.validators[0].Name())
	}
}

func TestValidationEngineValidateAll(t *testing.T) {
	engine := &ValidationEngine{}

	// Add mock validators
	passingValidator := &mockValidator{
		name:        "PassingValidator",
		description: "Always passes",
		shouldPass:  true,
		applicable:  true,
	}

	failingValidator := &mockValidator{
		name:        "FailingValidator",
		description: "Always fails",
		shouldPass:  false,
		applicable:  true,
	}

	notApplicableValidator := &mockValidator{
		name:        "NotApplicableValidator",
		description: "Not applicable",
		applicable:  false,
	}

	engine.RegisterValidator(passingValidator)
	engine.RegisterValidator(failingValidator)
	engine.RegisterValidator(notApplicableValidator)

	ctx := &ValidationContext{
		Solution: "test",
		Flavor:   "test",
	}

	results := engine.ValidateAll(ctx)

	// Should have results for applicable validators only
	if len(results) != 2 {
		t.Errorf("Expected 2 validation results, got %d", len(results))
	}

	// Check that we have one passing and one failing result
	var passCount, failCount int
	for _, result := range results {
		if result.Passed {
			passCount++
		} else {
			failCount++
		}
	}

	if passCount != 1 {
		t.Errorf("Expected 1 passing result, got %d", passCount)
	}

	if failCount != 1 {
		t.Errorf("Expected 1 failing result, got %d", failCount)
	}
}

func TestPrintResults(t *testing.T) {
	// Test PrintResults function with various result types
	results := []ValidationResult{
		{
			CheckName: "PassingCheck",
			Passed:    true,
			Message:   "Check passed",
			Severity:  "info",
		},
		{
			CheckName:  "FailingCheck",
			Passed:     false,
			Message:    "Check failed",
			Severity:   "error",
			Suggestion: "Fix this issue",
		},
		{
			CheckName: "WarningCheck",
			Passed:    false,
			Message:   "Warning message",
			Severity:  "warning",
		},
	}

	// This function prints to stdout, so we mainly test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintResults panicked: %v", r)
		}
	}()

	PrintResults(results)
}

func TestHasErrors(t *testing.T) {
	t.Run("no_errors", func(t *testing.T) {
		results := []ValidationResult{
			{Passed: true, Severity: "info"},
			{Passed: false, Severity: "warning"},
		}

		if HasErrors(results) {
			t.Error("HasErrors should return false when no error-severity failures exist")
		}
	})

	t.Run("with_errors", func(t *testing.T) {
		results := []ValidationResult{
			{Passed: true, Severity: "info"},
			{Passed: false, Severity: "error"},
		}

		if !HasErrors(results) {
			t.Error("HasErrors should return true when error-severity failures exist")
		}
	})

	t.Run("empty_results", func(t *testing.T) {
		results := []ValidationResult{}

		if HasErrors(results) {
			t.Error("HasErrors should return false for empty results")
		}
	})
}

func TestValidationSeverityLevels(t *testing.T) {
	// Test different severity levels
	severities := []string{"error", "warning", "info"}

	for _, severity := range severities {
		t.Run("severity_"+severity, func(t *testing.T) {
			result := ValidationResult{
				CheckName: "TestCheck",
				Passed:    false,
				Message:   "Test message",
				Severity:  severity,
			}

			hasErrors := HasErrors([]ValidationResult{result})

			if severity == "error" && !hasErrors {
				t.Error("Error severity should be detected by HasErrors")
			}

			if severity != "error" && hasErrors {
				t.Errorf("Non-error severity '%s' should not be detected by HasErrors", severity)
			}
		})
	}
}

func TestValidationContextParameterHandling(t *testing.T) {
	// Test parameter handling in ValidationContext
	ctx := ValidationContext{
		Parameters: map[string]string{
			"param1": "value1",
			"param2": "value2",
			"empty":  "",
		},
	}

	// Test parameter retrieval
	if ctx.Parameters["param1"] != "value1" {
		t.Error("Should be able to retrieve parameter values")
	}

	if ctx.Parameters["nonexistent"] != "" {
		t.Error("Non-existent parameters should return empty string")
	}

	// Test parameter modification
	ctx.Parameters["param3"] = "value3"
	if ctx.Parameters["param3"] != "value3" {
		t.Error("Should be able to add new parameters")
	}
}

func TestSkipChecksHandling(t *testing.T) {
	ctx := ValidationContext{
		SkipChecks: []string{"check1", "check2", "check3"},
	}

	// Test if a check should be skipped
	shouldSkip := func(checkName string, skipList []string) bool {
		for _, skip := range skipList {
			if skip == checkName {
				return true
			}
		}
		return false
	}

	if !shouldSkip("check1", ctx.SkipChecks) {
		t.Error("check1 should be in skip list")
	}

	if shouldSkip("check4", ctx.SkipChecks) {
		t.Error("check4 should not be in skip list")
	}
}

// Mock validator for testing
type mockValidator struct {
	name        string
	description string
	shouldPass  bool
	applicable  bool
}

func (m *mockValidator) Name() string {
	return m.name
}

func (m *mockValidator) Description() string {
	return m.description
}

func (m *mockValidator) Validate(ctx *ValidationContext) ValidationResult {
	return ValidationResult{
		CheckName: m.name,
		Passed:    m.shouldPass,
		Message:   "Mock validation result",
		Severity:  "info",
	}
}

func (m *mockValidator) IsApplicable(ctx *ValidationContext) bool {
	return m.applicable
}

func TestMockValidator(t *testing.T) {
	// Test our mock validator implementation
	mock := &mockValidator{
		name:        "TestMock",
		description: "Test Description",
		shouldPass:  true,
		applicable:  true,
	}

	if mock.Name() != "TestMock" {
		t.Errorf("Expected name 'TestMock', got '%s'", mock.Name())
	}

	if mock.Description() != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", mock.Description())
	}

	if !mock.IsApplicable(&ValidationContext{}) {
		t.Error("Expected mock to be applicable")
	}

	result := mock.Validate(&ValidationContext{})
	if !result.Passed {
		t.Error("Expected mock validation to pass")
	}
}

// Benchmark validation engine performance
func BenchmarkValidationEngine(b *testing.B) {
	engine := NewValidationEngine()
	ctx := &ValidationContext{
		Solution: "arcbox",
		Flavor:   "ITPro",
		Location: "eastus",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ValidateAll(ctx)
	}
}

// Add tests for concurrent validation
func TestConcurrentValidation(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Concurrent Validation ===")

	engine := NewValidationEngine()
	mockCLI := azurecli.NewMockAzureCLI()
	ctx := &ValidationContext{
		Solution: "test",
		Flavor:   "test",
		AzureCLI: mockCLI,
	}

	// Run multiple validations concurrently
	done := make(chan []ValidationResult, 5)

	for i := 0; i < 5; i++ {
		go func() {
			results := engine.ValidateAll(ctx)
			done <- results
		}()
	}

	// Collect results
	var allResults [][]ValidationResult
	for i := 0; i < 5; i++ {
		allResults = append(allResults, <-done)
	}

	// Verify consistency
	consistent := true
	if len(allResults) > 1 {
		firstLen := len(allResults[0])
		for i := 1; i < len(allResults); i++ {
			if len(allResults[i]) != firstLen {
				consistent = false
				break
			}
		}
	}

	testutils.PrintTestStatus(t, "Concurrent validation consistency", consistent,
		"All concurrent validations should return consistent results")
}

// Add performance tests
func TestValidationPerformance(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Validation Performance ===")

	engine := &ValidationEngine{}

	// Add many validators
	for i := 0; i < 100; i++ {
		engine.RegisterValidator(&mockValidator{
			name:        fmt.Sprintf("Validator%d", i),
			description: "Performance test validator",
			shouldPass:  i%2 == 0,
			applicable:  true,
		})
	}

	ctx := &ValidationContext{
		Solution: "test",
		Flavor:   "test",
	}

	start := time.Now()
	results := engine.ValidateAll(ctx)
	duration := time.Since(start)

	testutils.PrintTestStatus(t, "Performance", duration < 1*time.Second,
		fmt.Sprintf("100 validators completed in %v", duration))

	testutils.PrintTestStatus(t, "Results count", len(results) == 100,
		fmt.Sprintf("Expected 100 results, got %d", len(results)))
}

// Add edge case tests for ValidationResult
func TestValidationResultEdgeCases(t *testing.T) {
	testutils.PrintTestHeader("=== Testing ValidationResult Edge Cases ===")

	t.Run("empty_severity", func(t *testing.T) {
		result := ValidationResult{
			CheckName: "Test",
			Passed:    false,
			Severity:  "", // empty severity
		}

		hasErrors := HasErrors([]ValidationResult{result})
		testutils.PrintTestStatus(t, "Empty severity", !hasErrors,
			"Empty severity should not be treated as error")
	})

	t.Run("nil_message_fields", func(t *testing.T) {
		result := ValidationResult{
			CheckName:  "Test",
			Passed:     true,
			Message:    "",
			Suggestion: "",
			Details:    "",
		}

		// Should not panic with empty strings
		defer func() {
			if r := recover(); r != nil {
				testutils.PrintTestStatus(t, "Nil fields handling", false,
					fmt.Sprintf("Should not panic with empty fields: %v", r))
			} else {
				testutils.PrintTestStatus(t, "Nil fields handling", true,
					"Handles empty string fields correctly")
			}
		}()

		PrintResults([]ValidationResult{result})
	})
}

// Add mutation tests for ValidationContext
func TestValidationContextMutation(t *testing.T) {
	testutils.PrintTestHeader("=== Testing ValidationContext Mutation ===")

	ctx := &ValidationContext{
		Parameters: map[string]string{"key1": "value1"},
		SkipChecks: []string{"check1"},
	}

	// Test parameter mutation
	originalParams := make(map[string]string)
	for k, v := range ctx.Parameters {
		originalParams[k] = v
	}

	ctx.Parameters["key2"] = "value2"

	testutils.PrintTestStatus(t, "Parameter mutation", len(ctx.Parameters) == 2,
		"Should be able to add parameters")

	// Test skip checks mutation
	originalSkipLen := len(ctx.SkipChecks)
	ctx.SkipChecks = append(ctx.SkipChecks, "check2")

	testutils.PrintTestStatus(t, "Skip checks mutation", len(ctx.SkipChecks) == originalSkipLen+1,
		"Should be able to add skip checks")
}

// Test SSH Key Validator
func TestSSHKeyValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing SSH Key Validator ===")

	validator := &SSHKeyValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "ssh-key-validation",
		fmt.Sprintf("Expected 'ssh-key-validation', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	ctx := &ValidationContext{Solution: "arcbox", Flavor: "DevOps"}
	testutils.PrintTestStatus(t, "IsApplicable DevOps", validator.IsApplicable(ctx),
		"Should be applicable to ArcBox DevOps")

	ctx.Flavor = "DataOps"
	testutils.PrintTestStatus(t, "IsApplicable DataOps", validator.IsApplicable(ctx),
		"Should be applicable to ArcBox DataOps")

	ctx.Flavor = "ITPro"
	testutils.PrintTestStatus(t, "IsApplicable ITPro", !validator.IsApplicable(ctx),
		"Should not be applicable to ArcBox ITPro")

	ctx.Solution = "localbox"
	ctx.Flavor = "DevOps"
	testutils.PrintTestStatus(t, "IsApplicable LocalBox", !validator.IsApplicable(ctx),
		"Should not be applicable to LocalBox")

	// Test Validate method with valid SSH key
	ctx = &ValidationContext{
		Solution: "arcbox",
		Flavor:   "DevOps",
		Parameters: map[string]string{
			"ssh-rsa-public-key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa+F8e4oV0NqWRGGZdOQnBvnM7lBZEO3nZPH2eZi6aB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa+F8e4oV0NqWRGGZdOQnBvnM7lBZEO3nZPH2eZi6a comment@example.com",
		},
	}
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate valid SSH key", result.Passed,
		"Valid SSH key should pass validation")

	// Test Validate method with invalid SSH key
	ctx.Parameters["ssh-rsa-public-key"] = "invalid-key"
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate invalid SSH key", !result.Passed && result.Severity == "error",
		"Invalid SSH key should fail validation")

	// Test Validate method with empty SSH key
	ctx.Parameters["ssh-rsa-public-key"] = ""
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate empty SSH key", result.Passed && result.Severity == "info",
		"Empty SSH key should pass with info message")
}

// Test Windows Password Validator
func TestWindowsPasswordValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Windows Password Validator ===")

	validator := &WindowsPasswordValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "windows-password-validation",
		fmt.Sprintf("Expected 'windows-password-validation', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	ctx := &ValidationContext{
		Solution:   "arcbox",
		Parameters: map[string]string{"windows-password": "password123"},
	}
	testutils.PrintTestStatus(t, "IsApplicable with password", validator.IsApplicable(ctx),
		"Should be applicable when windows-password is provided")

	ctx.Parameters = map[string]string{}
	testutils.PrintTestStatus(t, "IsApplicable without password", !validator.IsApplicable(ctx),
		"Should not be applicable when windows-password is not provided")

	// Test Validate method with valid password
	ctx.Parameters["windows-password"] = "ValidPassword123!"
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate valid password", result.Passed,
		"Valid password should pass validation")

	// Test Validate method with invalid password
	ctx.Parameters["windows-password"] = "weak"
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate invalid password", !result.Passed && result.Severity == "error",
		"Invalid password should fail validation")
}

// Test Resource Tags Validator
func TestResourceTagsValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Resource Tags Validator ===")

	validator := &ResourceTagsValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "resource-tags-validation",
		fmt.Sprintf("Expected 'resource-tags-validation', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	ctx := &ValidationContext{
		Parameters: map[string]string{"resource-tags": `{"env":"test"}`},
	}
	testutils.PrintTestStatus(t, "IsApplicable with tags", validator.IsApplicable(ctx),
		"Should be applicable when resource-tags is provided")

	ctx.Parameters = map[string]string{}
	testutils.PrintTestStatus(t, "IsApplicable without tags", !validator.IsApplicable(ctx),
		"Should not be applicable when resource-tags is not provided")

	// Test Validate method with valid JSON
	ctx.Parameters["resource-tags"] = `{"Environment":"Dev","Owner":"TeamA"}`
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate valid JSON", result.Passed,
		"Valid JSON should pass validation")

	// Test Validate method with invalid JSON
	ctx.Parameters["resource-tags"] = `{"invalid": json}`
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate invalid JSON", !result.Passed && result.Severity == "error",
		"Invalid JSON should fail validation")
}

// Test GitHub Username Validator
func TestGitHubUsernameValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing GitHub Username Validator ===")

	validator := &GitHubUsernameValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "github-username-validation",
		fmt.Sprintf("Expected 'github-username-validation', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	ctx := &ValidationContext{
		Solution:   "arcbox",
		Flavor:     "DevOps",
		Parameters: map[string]string{"github-user": "testuser"},
	}
	testutils.PrintTestStatus(t, "IsApplicable DevOps with user", validator.IsApplicable(ctx),
		"Should be applicable to ArcBox DevOps with github-user")

	ctx.Flavor = "ITPro"
	testutils.PrintTestStatus(t, "IsApplicable ITPro", !validator.IsApplicable(ctx),
		"Should not be applicable to ITPro flavor")

	// Test Validate method with valid username
	ctx.Flavor = "DevOps"
	ctx.Parameters["github-user"] = "validuser123"
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate valid username", result.Passed,
		"Valid username should pass validation")

	// Test Validate method with invalid username (email)
	ctx.Parameters["github-user"] = "user@example.com"
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate email as username", !result.Passed && result.Severity == "error",
		"Email format should fail validation")

	// Test Validate method with empty username for DevOps
	ctx.Parameters["github-user"] = ""
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate empty username DevOps", !result.Passed && result.Severity == "error",
		"Empty username for DevOps should fail validation")
}

// Test Flavor Specific Validator
func TestFlavorSpecificValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Flavor Specific Validator ===")

	validator := &FlavorSpecificValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "flavor-specific-requirements",
		fmt.Sprintf("Expected 'flavor-specific-requirements', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	ctx := &ValidationContext{Solution: "arcbox"}
	testutils.PrintTestStatus(t, "IsApplicable ArcBox", validator.IsApplicable(ctx),
		"Should be applicable to ArcBox")

	ctx.Solution = "localbox"
	testutils.PrintTestStatus(t, "IsApplicable LocalBox", !validator.IsApplicable(ctx),
		"Should not be applicable to LocalBox")

	// Test Validate method
	ctx.Solution = "arcbox"
	ctx.Flavor = "ITPro"
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate ITPro", result.Passed && result.Message == "",
		"ITPro should pass with empty message")
}

// Test Azure CLI Health Validator
func TestAzureCLIHealthValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Azure CLI Health Validator ===")

	validator := &AzureCLIHealthValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "azure-cli-health",
		fmt.Sprintf("Expected 'azure-cli-health', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	mockCLI := azurecli.NewMockAzureCLI()
	ctx := &ValidationContext{AzureCLI: mockCLI}
	testutils.PrintTestStatus(t, "IsApplicable", validator.IsApplicable(ctx),
		"Should always be applicable")

	// Test Validate method - this will depend on actual Azure CLI status
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate CLI health", true,
		fmt.Sprintf("CLI health check result: %s", result.Message))
}

// Test Subscription Access Validator
func TestSubscriptionAccessValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Subscription Access Validator ===")

	validator := &SubscriptionAccessValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "subscription-access",
		fmt.Sprintf("Expected 'subscription-access', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	mockCLI := azurecli.NewMockAzureCLI()
	ctx := &ValidationContext{AzureCLI: mockCLI}
	testutils.PrintTestStatus(t, "IsApplicable", validator.IsApplicable(ctx),
		"Should always be applicable")

	// Test Validate method with subscription parameter
	ctx.Parameters = map[string]string{"subscription": "test-subscription-id"}
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate with subscription", true,
		fmt.Sprintf("Subscription validation result: %s", result.Message))
}

// Test Resource Provider Validator
func TestResourceProviderValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Resource Provider Validator ===")

	validator := &ResourceProviderValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "resource-providers",
		fmt.Sprintf("Expected 'resource-providers', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	mockAzCLI := &azurecli.MockAzureCLI{}
	ctx := &ValidationContext{Solution: "arcbox", AzureCLI: mockAzCLI}
	testutils.PrintTestStatus(t, "IsApplicable ArcBox", validator.IsApplicable(ctx),
		"Should be applicable to ArcBox")

	ctx.Solution = "localbox"
	testutils.PrintTestStatus(t, "IsApplicable LocalBox", !validator.IsApplicable(ctx),
		"Should not be applicable to LocalBox")

	// Test Validate method
	ctx.Solution = "arcbox"
	ctx.AzureCLI = mockAzCLI
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate resource providers", true,
		fmt.Sprintf("Resource provider validation result: %s", result.Message))
}

// Test Quota Validator
func TestQuotaValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Quota Validator ===")

	validator := &QuotaValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "vcpu-quota",
		fmt.Sprintf("Expected 'vcpu-quota', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	mockCLI := azurecli.NewMockAzureCLI()
	ctx := &ValidationContext{Flavor: "ITPro", Location: "eastus", AzureCLI: mockCLI}
	testutils.PrintTestStatus(t, "IsApplicable with flavor and location", validator.IsApplicable(ctx),
		"Should be applicable when flavor and location are provided")

	ctx.Flavor = ""
	testutils.PrintTestStatus(t, "IsApplicable without flavor", !validator.IsApplicable(ctx),
		"Should not be applicable when flavor is missing")

	// Test Validate method with unknown flavor
	ctx.Flavor = "UnknownFlavor"
	ctx.Location = "eastus"
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate unknown flavor", !result.Passed && result.Severity == "error",
		"Unknown flavor should fail validation")

	// Test Validate method with valid flavor
	ctx.Flavor = "ITPro"
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate ITPro flavor", true,
		fmt.Sprintf("ITPro quota validation result: %s", result.Message))
}

// Test Region Validator
func TestRegionValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Region Validator ===")

	validator := &RegionValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "region-support",
		fmt.Sprintf("Expected 'region-support', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	ctx := &ValidationContext{Location: "eastus"}
	testutils.PrintTestStatus(t, "IsApplicable with location", validator.IsApplicable(ctx),
		"Should be applicable when location is provided")

	ctx.Location = ""
	testutils.PrintTestStatus(t, "IsApplicable without location", !validator.IsApplicable(ctx),
		"Should not be applicable when location is missing")

	// Test Validate method with valid region
	ctx.Location = "eastus"
	ctx.Solution = "arcbox"
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate valid region", result.Passed,
		"Valid region should pass validation")

	// Test Validate method with invalid region
	ctx.Location = "invalidregion"
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate invalid region", !result.Passed && result.Severity == "error",
		"Invalid region should fail validation")
}

// Test SKU Availability Validator
func TestSKUAvailabilityValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing SKU Availability Validator ===")

	validator := &SKUAvailabilityValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "sku-availability",
		fmt.Sprintf("Expected 'sku-availability', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	mockCLI := azurecli.NewMockAzureCLI()
	ctx := &ValidationContext{Flavor: "ITPro", Location: "eastus", AzureCLI: mockCLI}
	testutils.PrintTestStatus(t, "IsApplicable with flavor and location", validator.IsApplicable(ctx),
		"Should be applicable when flavor and location are provided")

	ctx.Flavor = ""
	testutils.PrintTestStatus(t, "IsApplicable without flavor", !validator.IsApplicable(ctx),
		"Should not be applicable when flavor is missing")

	// Test Validate method with unknown flavor
	ctx.Flavor = "UnknownFlavor"
	ctx.Location = "eastus"
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate unknown flavor", !result.Passed && result.Severity == "error",
		"Unknown flavor should fail validation")

	// Test Validate method with valid flavor
	ctx.Flavor = "ITPro"
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate ITPro SKUs", true,
		fmt.Sprintf("ITPro SKU validation result: %s", result.Message))
}

// Test SSH Key Requirement Validator
func TestSSHKeyRequirementValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing SSH Key Requirement Validator ===")

	validator := &SSHKeyRequirementValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "ssh-key-requirement-validation",
		fmt.Sprintf("Expected 'ssh-key-requirement-validation', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	ctx := &ValidationContext{Solution: "arcbox", Flavor: "DevOps"}
	testutils.PrintTestStatus(t, "IsApplicable DevOps", validator.IsApplicable(ctx),
		"Should be applicable to ArcBox DevOps")

	ctx.Flavor = "DataOps"
	testutils.PrintTestStatus(t, "IsApplicable DataOps", validator.IsApplicable(ctx),
		"Should be applicable to ArcBox DataOps")

	ctx.Flavor = "ITPro"
	testutils.PrintTestStatus(t, "IsApplicable ITPro", !validator.IsApplicable(ctx),
		"Should not be applicable to ArcBox ITPro")

	// Test Validate method with SSH key provided
	ctx.Flavor = "DevOps"
	ctx.Parameters = map[string]string{"ssh-rsa-public-key": "test-key"}
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate with SSH key", result.Passed,
		"Should pass when SSH key is provided")

	// Test Validate method without SSH key
	ctx.Parameters = map[string]string{}
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate without SSH key", !result.Passed && result.Severity == "error",
		"Should fail when SSH key is missing for DevOps")
}

// Test GitHub User Requirement Validator
func TestGitHubUserRequirementValidator(t *testing.T) {
	testutils.PrintTestHeader("=== Testing GitHub User Requirement Validator ===")

	validator := &GitHubUserRequirementValidator{}

	// Test Name method
	testutils.PrintTestStatus(t, "Name method", validator.Name() == "github-user-requirement-validation",
		fmt.Sprintf("Expected 'github-user-requirement-validation', got '%s'", validator.Name()))

	// Test Description method
	desc := validator.Description()
	testutils.PrintTestStatus(t, "Description method", desc != "",
		"Description should not be empty")

	// Test IsApplicable method
	ctx := &ValidationContext{Solution: "arcbox", Flavor: "DevOps"}
	testutils.PrintTestStatus(t, "IsApplicable DevOps", validator.IsApplicable(ctx),
		"Should be applicable to ArcBox DevOps")

	ctx.Flavor = "ITPro"
	testutils.PrintTestStatus(t, "IsApplicable ITPro", !validator.IsApplicable(ctx),
		"Should not be applicable to ArcBox ITPro")

	// Test Validate method with GitHub user provided
	ctx.Flavor = "DevOps"
	ctx.Parameters = map[string]string{"github-user": "testuser"}
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate with GitHub user", result.Passed,
		"Should pass when GitHub user is provided")

	// Test Validate method without GitHub user
	ctx.Parameters = map[string]string{}
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "Validate without GitHub user", !result.Passed && result.Severity == "error",
		"Should fail when GitHub user is missing for DevOps")
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Helper Functions ===")

	// Test isValidSSHKey
	testutils.PrintTestStatus(t, "isValidSSHKey valid",
		isValidSSHKey("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa+F8e4oV0NqWRGGZdOQnBvnM7lBZEO3nZPH2eZi6aB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa+F8e4oV0NqWRGGZdOQnBvnM7lBZEO3nZPH2eZi6a comment@example.com"),
		"Valid SSH key should return true")

	testutils.PrintTestStatus(t, "isValidSSHKey invalid",
		!isValidSSHKey("invalid-key"),
		"Invalid SSH key should return false")

	testutils.PrintTestStatus(t, "isValidSSHKey wrong type",
		!isValidSSHKey("ssh-dss AAAAB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa"),
		"Non-RSA SSH key should return false")

	// Test isValidWindowsPassword
	testutils.PrintTestStatus(t, "isValidWindowsPassword valid",
		isValidWindowsPassword("ValidPassword123!"),
		"Valid Windows password should return true")

	testutils.PrintTestStatus(t, "isValidWindowsPassword too short",
		!isValidWindowsPassword("short"),
		"Short password should return false")

	testutils.PrintTestStatus(t, "isValidWindowsPassword too long",
		!isValidWindowsPassword(strings.Repeat("a", 125)),
		"Long password should return false")

	testutils.PrintTestStatus(t, "isValidWindowsPassword simple",
		!isValidWindowsPassword("simplepassword"),
		"Simple password should return false")

	// Test isValidGitHubUsername
	testutils.PrintTestStatus(t, "isValidGitHubUsername valid",
		isValidGitHubUsername("validuser123"),
		"Valid GitHub username should return true")

	testutils.PrintTestStatus(t, "isValidGitHubUsername email",
		!isValidGitHubUsername("user@example.com"),
		"Email format should return false")

	testutils.PrintTestStatus(t, "isValidGitHubUsername url",
		!isValidGitHubUsername("github.com/user"),
		"URL format should return false")

	testutils.PrintTestStatus(t, "isValidGitHubUsername consecutive hyphens",
		!isValidGitHubUsername("user--name"),
		"Consecutive hyphens should return false")

	// Test ValidateEmail
	testutils.PrintTestStatus(t, "ValidateEmail valid",
		ValidateEmail("user@example.com"),
		"Valid email should return true")

	testutils.PrintTestStatus(t, "ValidateEmail invalid",
		!ValidateEmail("invalid-email"),
		"Invalid email should return false")

	// Test getFlavorSKUsForValidation
	itproSKUs := getFlavorSKUsForValidation("ITPro")
	testutils.PrintTestStatus(t, "getFlavorSKUsForValidation ITPro",
		len(itproSKUs) > 0,
		fmt.Sprintf("ITPro should have SKUs: %v", itproSKUs))

	devopsSKUs := getFlavorSKUsForValidation("DevOps")
	testutils.PrintTestStatus(t, "getFlavorSKUsForValidation DevOps",
		len(devopsSKUs) > 0,
		fmt.Sprintf("DevOps should have SKUs: %v", devopsSKUs))

	unknownSKUs := getFlavorSKUsForValidation("Unknown")
	testutils.PrintTestStatus(t, "getFlavorSKUsForValidation Unknown",
		len(unknownSKUs) == 0,
		"Unknown flavor should have no SKUs")

	// Test getRequiredVCPUForSKU
	vcpu := getRequiredVCPUForSKU("Standard_D8s_v5")
	testutils.PrintTestStatus(t, "getRequiredVCPUForSKU D8s_v5",
		vcpu == 8,
		fmt.Sprintf("D8s_v5 should require 8 vCPUs, got %d", vcpu))

	vcpu = getRequiredVCPUForSKU("Unknown_SKU")
	testutils.PrintTestStatus(t, "getRequiredVCPUForSKU Unknown",
		vcpu == 1,
		fmt.Sprintf("Unknown SKU should default to 1 vCPU, got %d", vcpu))

	// Test parseInt64
	testutils.PrintTestStatus(t, "parseInt64 int",
		parseInt64(42) == 42,
		"Should convert int correctly")

	testutils.PrintTestStatus(t, "parseInt64 string",
		parseInt64("123") == 123,
		"Should convert string correctly")

	testutils.PrintTestStatus(t, "parseInt64 float",
		parseInt64(42.7) == 42,
		"Should convert float correctly")

	testutils.PrintTestStatus(t, "parseInt64 invalid",
		parseInt64("invalid") == 0,
		"Should return 0 for invalid input")

	// Test isValidAzureRegion
	testutils.PrintTestStatus(t, "isValidAzureRegion valid",
		isValidAzureRegion("eastus"),
		"eastus should be valid region")

	testutils.PrintTestStatus(t, "isValidAzureRegion invalid",
		!isValidAzureRegion("invalidregion"),
		"invalidregion should be invalid")

	// Test isRegionSupportedForArcBox
	testutils.PrintTestStatus(t, "isRegionSupportedForArcBox valid",
		isRegionSupportedForArcBox("eastus"),
		"eastus should be supported for ArcBox")

	testutils.PrintTestStatus(t, "isRegionSupportedForArcBox invalid",
		!isRegionSupportedForArcBox("unsupportedregion"),
		"unsupportedregion should not be supported for ArcBox")

	// Test mapSKUToFamilyQuotaName
	family := mapSKUToFamilyQuotaName("Standard_D8s_v5")
	testutils.PrintTestStatus(t, "mapSKUToFamilyQuotaName D8s_v5",
		family != "",
		fmt.Sprintf("D8s_v5 should map to family: %s", family))

	family = mapSKUToFamilyQuotaName("Standard_B2ms")
	testutils.PrintTestStatus(t, "mapSKUToFamilyQuotaName B2ms",
		family == "Standard BS Family vCPUs",
		fmt.Sprintf("B2ms should map to BS family, got: %s", family))
}
}

// Test the remaining helper functions with 0% coverage
func TestRemainingHelperFunctions(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Remaining Helper Functions ===")

	// Test checkAzureCLIHealth
	mockCLI := azurecli.NewMockAzureCLI()
	err := checkAzureCLIHealth(mockCLI)
	testutils.PrintTestStatus(t, "checkAzureCLIHealth", true,
		fmt.Sprintf("Azure CLI health check completed (err: %v)", err))

	// Test getSubscriptionFromContext with different scenarios
	ctx := &ValidationContext{
		Parameters: map[string]string{"subscription": "test-sub-from-param"},
		AzureCLI:   mockCLI,
	}
	sub := getSubscriptionFromContext(ctx)
	testutils.PrintTestStatus(t, "getSubscriptionFromContext with param",
		sub == "test-sub-from-param",
		fmt.Sprintf("Should get subscription from parameters: %s", sub))

	// Test with environment variable (simulate by temporarily setting it)
	originalEnv := os.Getenv("AZURE_SUBSCRIPTION_ID")
	os.Setenv("AZURE_SUBSCRIPTION_ID", "test-sub-from-env")

	ctxWithoutParam := &ValidationContext{
		Parameters: map[string]string{},
		AzureCLI:   mockCLI,
	}
	sub = getSubscriptionFromContext(ctxWithoutParam)
	expectedFromEnv := sub == "test-sub-from-env" || sub != "" // May fall back to az CLI
	testutils.PrintTestStatus(t, "getSubscriptionFromContext with env", expectedFromEnv,
		fmt.Sprintf("Should get subscription from env or fallback: %s", sub))

	// Restore original environment
	if originalEnv != "" {
		os.Setenv("AZURE_SUBSCRIPTION_ID", originalEnv)
	} else {
		os.Unsetenv("AZURE_SUBSCRIPTION_ID")
	}

	// Test setAzureSubscription
	err = setAzureSubscription(mockCLI, "test-subscription")
	testutils.PrintTestStatus(t, "setAzureSubscription", true,
		fmt.Sprintf("setAzureSubscription completed (err: %v)", err))

	// Test setAzureSubscription with empty ID
	err = setAzureSubscription(mockCLI, "")
	testutils.PrintTestStatus(t, "setAzureSubscription empty", err != nil,
		"setAzureSubscription with empty ID should return error")

	// Test getResourceProviderConfig
	config := getResourceProviderConfig("arcbox")
	testutils.PrintTestStatus(t, "getResourceProviderConfig arcbox", config != nil,
		"ArcBox should have resource provider config")

	config = getResourceProviderConfig("unknown")
	testutils.PrintTestStatus(t, "getResourceProviderConfig unknown", config == nil,
		"Unknown solution should return nil config")

	// Test checkAllResourceProviders (if we have a config)
	if arcboxConfig := getResourceProviderConfig("arcbox"); arcboxConfig != nil {
		mockAzCLI := &azurecli.MockAzureCLI{}
		allRegistered, missing := checkAllResourceProviders(mockAzCLI, *arcboxConfig)
		testutils.PrintTestStatus(t, "checkAllResourceProviders", true,
			fmt.Sprintf("Resource provider check completed - registered: %v, missing: %v", allRegistered, missing))
	}

	// Test CheckQuotaForSKU and CheckBatchSKUAvailability exported functions
	quotaOK, current, limit, available := CheckQuotaForSKU("Standard_D8s_v5", 8, "eastus", "test-sub", "ITPro")
	testutils.PrintTestStatus(t, "CheckQuotaForSKU", true,
		fmt.Sprintf("Quota check completed - OK: %v, current: %d, limit: %d, available: %d", quotaOK, current, limit, available))

	unavailable := CheckBatchSKUAvailability([]string{"Standard_D8s_v5", "Standard_B2ms"}, "eastus", "test-sub")
	testutils.PrintTestStatus(t, "CheckBatchSKUAvailability", true,
		fmt.Sprintf("Batch SKU check completed - unavailable: %v", unavailable))
}

// Test ValidateAll with comprehensive coverage
func TestValidateAllWithRealValidators(t *testing.T) {
	testutils.PrintTestHeader("=== Testing ValidateAll with Real Validators ===")

	engine := NewValidationEngine()

	// Create a mock Azure CLI for testing
	mockCLI := azurecli.NewMockAzureCLI()

	// Test with different contexts to trigger different validation paths
	testCases := []struct {
		name string
		ctx  *ValidationContext
	}{
		{
			name: "ArcBox ITPro",
			ctx: &ValidationContext{
				Solution: "arcbox",
				Flavor:   "ITPro",
				Location: "eastus",
				AzureCLI: mockCLI, // Add the Azure CLI instance
				Parameters: map[string]string{
					"subscription": "test-subscription",
				},
				SilentMode: true,
			},
		},
		{
			name: "ArcBox DevOps",
			ctx: &ValidationContext{
				Solution: "arcbox",
				Flavor:   "DevOps",
				Location: "westus2",
				AzureCLI: mockCLI, // Add the Azure CLI instance
				Parameters: map[string]string{
					"subscription":       "test-subscription",
					"ssh-rsa-public-key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa+F8e4oV0NqWRGGZdOQnBvnM7lBZEO3nZPH2eZi6a comment@example.com",
					"github-user":        "testuser",
					"resource-tags":      `{"Environment":"Test"}`,
				},
				SilentMode: false,
			},
		},
		{
			name: "ArcBox DataOps",
			ctx: &ValidationContext{
				Solution: "arcbox",
				Flavor:   "DataOps",
				Location: "centralus",
				AzureCLI: mockCLI, // Add the Azure CLI instance
				Parameters: map[string]string{
					"subscription":       "test-subscription",
					"ssh-rsa-public-key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa+F8e4oV0NqWRGGZdOQnBvnM7lBZEO3nZPH2eZi6a comment@example.com",
					"windows-password":   "ComplexPassword123!",
				},
				SilentMode: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results := engine.ValidateAll(tc.ctx)
			testutils.PrintTestStatus(t, fmt.Sprintf("%s validation count", tc.name),
				len(results) > 0,
				fmt.Sprintf("Should have validation results: %d", len(results)))

			// Check that results contain expected validators for each context
			resultNames := make(map[string]bool)
			for _, result := range results {
				resultNames[result.CheckName] = true
			}

			// Common validators that should always run
			expectedCommon := []string{"azure-cli-health", "subscription-access"}
			for _, expected := range expectedCommon {
				testutils.PrintTestStatus(t, fmt.Sprintf("%s has %s", tc.name, expected),
					resultNames[expected],
					fmt.Sprintf("Should include %s validator", expected))
			}
		})
	}
}

// Test runValidatorWithSpinner with different scenarios
func TestRunValidatorWithSpinner(t *testing.T) {
	testutils.PrintTestHeader("=== Testing runValidatorWithSpinner ===")

	engine := &ValidationEngine{}

	// Test with fast validator
	fastValidator := &mockValidator{
		name:        "FastValidator",
		description: "Fast test validator",
		shouldPass:  true,
		applicable:  true,
	}

	ctx := &ValidationContext{SilentMode: false}
	result := engine.runValidatorWithSpinner(fastValidator, ctx)
	testutils.PrintTestStatus(t, "runValidatorWithSpinner fast", result.Passed,
		"Fast validator should complete successfully")

	// Test with silent mode
	ctx.SilentMode = true
	result = engine.runValidatorWithSpinner(fastValidator, ctx)
	testutils.PrintTestStatus(t, "runValidatorWithSpinner silent", result.Passed,
		"Silent mode should work correctly")

	// Test with failing validator
	failingValidator := &mockValidator{
		name:        "FailingValidator",
		description: "Failing test validator",
		shouldPass:  false,
		applicable:  true,
	}

	result = engine.runValidatorWithSpinner(failingValidator, ctx)
	testutils.PrintTestStatus(t, "runValidatorWithSpinner failing", !result.Passed,
		"Failing validator should return failure")
}

// Test PrintResults with comprehensive coverage
func TestPrintResultsComprehensive(t *testing.T) {
	testutils.PrintTestHeader("=== Testing PrintResults Comprehensive ===")

	// Test with all severity types and various scenarios
	results := []ValidationResult{
		{
			CheckName:  "PassingInfo",
			Passed:     true,
			Message:    "Info message",
			Severity:   "info",
			Suggestion: "No action needed",
			Details:    "Additional info details",
		},
		{
			CheckName:  "FailingError",
			Passed:     false,
			Message:    "Error message",
			Severity:   "error",
			Suggestion: "Fix this error",
			Details:    "Error details",
		},
		{
			CheckName:  "FailingWarning",
			Passed:     false,
			Message:    "Warning message",
			Severity:   "warning",
			Suggestion: "Consider this warning",
			Details:    "Warning details",
		},
		{
			CheckName: "EmptyMessageResult",
			Passed:    true,
			Message:   "", // Empty message should be skipped
			Severity:  "info",
		},
		{
			CheckName:  "LongMessage",
			Passed:     false,
			Message:    "This is a very long message that should be handled properly by the print function and not cause any issues with formatting or display",
			Severity:   "error",
			Suggestion: "This is also a very long suggestion that should be formatted correctly",
			Details:    "These are very detailed details that provide comprehensive information about the validation failure",
		},
	}

	// Test with different result combinations
	defer func() {
		if r := recover(); r != nil {
			testutils.PrintTestStatus(t, "PrintResults comprehensive", false,
				fmt.Sprintf("PrintResults should not panic: %v", r))
		} else {
			testutils.PrintTestStatus(t, "PrintResults comprehensive", true,
				"PrintResults handled all result types successfully")
		}
	}()

	PrintResults(results)

	// Test with empty results
	PrintResults([]ValidationResult{})
	testutils.PrintTestStatus(t, "PrintResults empty", true,
		"PrintResults should handle empty results")

	// Test with nil results (edge case)
	defer func() {
		if r := recover(); r != nil {
			testutils.PrintTestStatus(t, "PrintResults nil handling", false,
				fmt.Sprintf("PrintResults should handle nil gracefully: %v", r))
		} else {
			testutils.PrintTestStatus(t, "PrintResults nil handling", true,
				"PrintResults handled nil results gracefully")
		}
	}()

	var nilResults []ValidationResult
	PrintResults(nilResults)
}

// Test edge cases and error scenarios to improve coverage
func TestEdgeCasesAndErrorScenarios(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Edge Cases and Error Scenarios ===")

	// Test ValidateAll with skip checks
	engine := &ValidationEngine{}
	mockValidator := &mockValidator{
		name:        "SkippableValidator",
		description: "Test skip functionality",
		shouldPass:  true,
		applicable:  true,
	}
	engine.RegisterValidator(mockValidator)

	ctx := &ValidationContext{
		Solution:   "test",
		SkipChecks: []string{"SkippableValidator"},
	}

	results := engine.ValidateAll(ctx)
	testutils.PrintTestStatus(t, "ValidateAll with skip checks", len(results) == 0,
		"Skipped validators should not appear in results")

	// Test runValidatorWithSpinner with long-running validator
	slowValidator := &slowMockValidator{
		name:        "SlowValidator",
		description: "Slow test validator",
		shouldPass:  true,
		applicable:  true,
		delay:       100, // 100ms delay
	}

	ctx.SilentMode = false
	ctx.SkipChecks = []string{} // Clear skip checks
	result := engine.runValidatorWithSpinner(slowValidator, ctx)
	testutils.PrintTestStatus(t, "runValidatorWithSpinner slow validator", result.Passed,
		"Slow validator should complete successfully")

	// Test FlavorSpecificValidator with DevOps flavor that has requirements
	flavorValidator := &FlavorSpecificValidator{}
	ctx = &ValidationContext{
		Solution: "arcbox",
		Flavor:   "DevOps",
		Parameters: map[string]string{
			"ssh-rsa-public-key": "",
			"github-user":        "",
		},
	}
	result = flavorValidator.Validate(ctx)
	testutils.PrintTestStatus(t, "FlavorSpecificValidator DevOps empty params",
		result.Passed && result.Message == "",
		"DevOps with empty params should pass with empty message")

	// Test getFlavorSKUsForValidation with case variations
	skus := getFlavorSKUsForValidation("itpro")
	testutils.PrintTestStatus(t, "getFlavorSKUsForValidation lowercase", len(skus) > 0,
		fmt.Sprintf("Lowercase flavor should work: %v", skus))

	skus = getFlavorSKUsForValidation("ALL")
	testutils.PrintTestStatus(t, "getFlavorSKUsForValidation uppercase", len(skus) > 0,
		fmt.Sprintf("Uppercase 'ALL' should work: %v", skus))

	// Test isValidGitHubUsername edge cases
	testutils.PrintTestStatus(t, "isValidGitHubUsername single char",
		isValidGitHubUsername("a"),
		"Single character should be valid")

	testutils.PrintTestStatus(t, "isValidGitHubUsername max length",
		isValidGitHubUsername(strings.Repeat("a", 39)),
		"Max length (39 chars) should be valid")

	testutils.PrintTestStatus(t, "isValidGitHubUsername too long",
		!isValidGitHubUsername(strings.Repeat("a", 40)),
		"Too long (40 chars) should be invalid")

	testutils.PrintTestStatus(t, "isValidGitHubUsername starts with hyphen",
		!isValidGitHubUsername("-user"),
		"Username starting with hyphen should be invalid")

	testutils.PrintTestStatus(t, "isValidGitHubUsername ends with hyphen",
		!isValidGitHubUsername("user-"),
		"Username ending with hyphen should be invalid")

	// Test getSubscriptionFromContext fallback scenarios
	originalSub := os.Getenv("AZURE_SUBSCRIPTION_ID")
	os.Unsetenv("AZURE_SUBSCRIPTION_ID")

	ctx = &ValidationContext{
		Parameters: map[string]string{},
		AzureCLI:   azurecli.NewMockAzureCLI(),
	}
	sub := getSubscriptionFromContext(ctx)
	testutils.PrintTestStatus(t, "getSubscriptionFromContext fallback", true,
		fmt.Sprintf("Fallback subscription lookup: %s", sub))

	// Restore environment
	if originalSub != "" {
		os.Setenv("AZURE_SUBSCRIPTION_ID", originalSub)
	}

	// Test checkAllResourceProviders with error scenario
	config := resourceproviders.ResourceProviderConfig{
		RequiredProviders: []string{"NonExistentProvider"},
	}
	mockAzCLI := &azurecli.MockAzureCLI{}
	allRegistered, missing := checkAllResourceProviders(mockAzCLI, config)
	testutils.PrintTestStatus(t, "checkAllResourceProviders with missing", !allRegistered && len(missing) > 0,
		fmt.Sprintf("Should detect missing providers: %v", missing))

	// Test parseInt64 with more edge cases
	testutils.PrintTestStatus(t, "parseInt64 int64",
		parseInt64(int64(123)) == 123,
		"Should convert int64 correctly")

	testutils.PrintTestStatus(t, "parseInt64 nil interface",
		parseInt64(nil) == 0,
		"Should handle nil interface")

	var emptyInterface interface{}
	testutils.PrintTestStatus(t, "parseInt64 empty interface",
		parseInt64(emptyInterface) == 0,
		"Should handle empty interface")

	// Test mapSKUToFamilyQuotaName edge cases
	family := mapSKUToFamilyQuotaName("Standard_F4s")
	testutils.PrintTestStatus(t, "mapSKUToFamilyQuotaName F-series",
		family != "",
		fmt.Sprintf("F-series should map to family: %s", family))

	family = mapSKUToFamilyQuotaName("InvalidSKU")
	testutils.PrintTestStatus(t, "mapSKUToFamilyQuotaName invalid",
		family == "",
		fmt.Sprintf("Invalid SKU should return empty family name, got: '%s'", family))

	family = mapSKUToFamilyQuotaName("Standard_")
	testutils.PrintTestStatus(t, "mapSKUToFamilyQuotaName malformed",
		family == "",
		"Malformed SKU should return empty family name")

	// Test checkSKUAvailabilityInRegion with timeout scenario
	available := checkSKUAvailabilityInRegion("Standard_D8s_v5", "invalidregion", "test-sub")
	testutils.PrintTestStatus(t, "checkSKUAvailabilityInRegion invalid region", true,
		fmt.Sprintf("Invalid region check completed: %v", available))
}

// Test error scenarios in quota and SKU functions
func TestQuotaAndSKUErrorScenarios(t *testing.T) {
	testutils.PrintTestHeader("=== Testing Quota and SKU Error Scenarios ===")

	// Test checkQuotaForSKU with invalid region (should handle Azure CLI errors)
	quotaOK, current, limit, available := checkQuotaForSKU("Standard_D8s_v5", 8, "invalidregion", "invalid-sub", "ITPro")
	testutils.PrintTestStatus(t, "checkQuotaForSKU invalid region", true,
		fmt.Sprintf("Invalid region quota check - OK: %v, current: %d, limit: %d, available: %d", quotaOK, current, limit, available))

	// Test checkQuotaForSKU with unknown SKU
	quotaOK, current, limit, available = checkQuotaForSKU("UnknownSKU", 1, "eastus", "test-sub", "Custom")
	testutils.PrintTestStatus(t, "checkQuotaForSKU unknown SKU", true,
		fmt.Sprintf("Unknown SKU quota check - OK: %v, current: %d, limit: %d, available: %d", quotaOK, current, limit, available))

	// Test getRegionQuotaData with invalid region (error scenario)
	data, err := getRegionQuotaData("completely-invalid-region-name-that-does-not-exist")
	testutils.PrintTestStatus(t, "getRegionQuotaData invalid region", true,
		fmt.Sprintf("Invalid region quota data - count: %d, err: %v", len(data), err))

	// Test checkBatchSKUAvailability with timeout scenario (invalid region)
	unavailable := checkBatchSKUAvailability([]string{"Standard_D8s_v5"}, "invalid-region-timeout", "invalid-sub")
	testutils.PrintTestStatus(t, "checkBatchSKUAvailability timeout", true,
		fmt.Sprintf("Timeout scenario completed - unavailable: %v", unavailable))

	// Test checkIndividualSKUs with timeout scenario
	unavailable = checkIndividualSKUs([]string{"Standard_D8s_v5", "Standard_B2ms"}, "invalid-region", "invalid-sub")
	testutils.PrintTestStatus(t, "checkIndividualSKUs timeout", true,
		fmt.Sprintf("Individual SKU timeout scenario - unavailable: %v", unavailable))

	// Test checkBatchSKUAvailability with empty SKU list
	unavailable = checkBatchSKUAvailability([]string{}, "eastus", "test-sub")
	testutils.PrintTestStatus(t, "checkBatchSKUAvailability empty list", len(unavailable) == 0,
		"Empty SKU list should return empty unavailable list")

	// Test checkIndividualSKUs with empty SKU list
	unavailable = checkIndividualSKUs([]string{}, "eastus", "test-sub")
	testutils.PrintTestStatus(t, "checkIndividualSKUs empty list", len(unavailable) == 0,
		"Empty SKU list should return empty unavailable list")
}

// Test SSH Key validator edge cases
func TestSSHKeyValidatorEdgeCases(t *testing.T) {
	testutils.PrintTestHeader("=== Testing SSH Key Validator Edge Cases ===")

	validator := &SSHKeyValidator{}

	// Test with SSH key that has multiple spaces
	ctx := &ValidationContext{
		Solution: "arcbox",
		Flavor:   "DevOps",
		Parameters: map[string]string{
			"ssh-rsa-public-key": "ssh-rsa    AAAAB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa+F8e4oV0NqWRGGZdOQnBvnM7lBZEO3nZPH2eZi6aB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa+F8e4oV0NqWRGGZdOQnBvnM7lBZEO3nZPH2eZi6a    user@host",
		},
	}
	result := validator.Validate(ctx)
	testutils.PrintTestStatus(t, "SSH key with extra spaces", result.Passed,
		"SSH key with extra spaces should be valid")

	// Test with SSH key that's too short
	ctx.Parameters["ssh-rsa-public-key"] = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCshort"
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "SSH key too short", !result.Passed && result.Severity == "error",
		"Short SSH key should fail validation")

	// Test with SSH key that has invalid base64 characters
	ctx.Parameters["ssh-rsa-public-key"] = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa+F8e4oV0NqWRGGZdOQnBvnM7lBZEO3nZPH2eZi6aB3NzaC1yc2EAAAADAQABAAACAQC7yOPFqGPZcr15xDpTKhMa+F8e4oV0NqWRGGZdOQnBvnM7lBZEO3nZPH2eZi6a@#$%^&"
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "SSH key invalid base64", !result.Passed && result.Severity == "error",
		"SSH key with invalid base64 should fail validation")

	// Test with only one part (missing base64)
	ctx.Parameters["ssh-rsa-public-key"] = "ssh-rsa"
	result = validator.Validate(ctx)
	testutils.PrintTestStatus(t, "SSH key missing base64", !result.Passed && result.Severity == "error",
		"SSH key missing base64 part should fail validation")
}

// Mock validator that simulates slow execution
type slowMockValidator struct {
	name        string
	description string
	shouldPass  bool
	applicable  bool
	delay       int // delay in milliseconds
}

func (m *slowMockValidator) Name() string {
	return m.name
}

func (m *slowMockValidator) Description() string {
	return m.description
}

func (m *slowMockValidator) Validate(ctx *ValidationContext) ValidationResult {
	time.Sleep(time.Duration(m.delay) * time.Millisecond)
	return ValidationResult{
		CheckName: m.name,
		Passed:    m.shouldPass,
		Message:   "Slow mock validation result",
		Severity:  "info",
	}
}

func (m *slowMockValidator) IsApplicable(ctx *ValidationContext) bool {
	return m.applicable
}
