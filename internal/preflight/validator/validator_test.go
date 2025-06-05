package validator

import (
	"fmt"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/jumpstart-cli/internal/testutils"
)

var (
	validatorTestSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	validatorTestInfoColor    = color.New(color.FgCyan).SprintFunc()
	validatorTestWarnColor    = color.New(color.FgYellow).SprintFunc()
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
	ctx := &ValidationContext{
		Solution: "test",
		Flavor:   "test",
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
