package services

import (
	"testing"

	"jumpstartcli/internal/azurecli"

	"github.com/spf13/cobra"
)

func TestNewValidationService(t *testing.T) {
	// Test service creation
	mockCLI := &azurecli.MockAzureCLI{}
	service := NewValidationService(mockCLI)

	if service == nil {
		t.Fatal("NewValidationService returned nil")
	}

	if service.cli != mockCLI {
		t.Error("ValidationService should store the provided CLI instance")
	}
}

func TestCreateValidationContext(t *testing.T) {
	mockCLI := &azurecli.MockAzureCLI{}
	service := NewValidationService(mockCLI)

	// Create a test command with flags
	cmd := &cobra.Command{}
	cmd.Flags().String("flavor", "", "ArcBox flavor")
	cmd.Flags().String("location", "", "Azure region")
	cmd.Flags().String("subscription", "", "Azure subscription")
	cmd.Flags().String("skip-preflight", "", "Skip preflight checks")

	// Set some flag values
	cmd.Flags().Set("flavor", "ITPro")
	cmd.Flags().Set("location", "eastus")
	cmd.Flags().Set("subscription", "test-subscription-id")

	// Test context creation
	ctx, err := service.CreateValidationContext(cmd)

	if err != nil {
		t.Fatalf("CreateValidationContext returned error: %v", err)
	}

	if ctx == nil {
		t.Fatal("CreateValidationContext returned nil context")
	}

	// Verify context values
	if ctx.Solution != "arcbox" {
		t.Errorf("Expected solution 'arcbox', got '%s'", ctx.Solution)
	}

	if ctx.Flavor != "ITPro" {
		t.Errorf("Expected flavor 'ITPro', got '%s'", ctx.Flavor)
	}

	if ctx.Location != "eastus" {
		t.Errorf("Expected location 'eastus', got '%s'", ctx.Location)
	}

	if ctx.Parameters["subscription"] != "test-subscription-id" {
		t.Errorf("Expected subscription 'test-subscription-id', got '%s'", ctx.Parameters["subscription"])
	}

	if ctx.AzureCLI != mockCLI {
		t.Error("Context should contain the CLI instance")
	}
}

func TestShouldSkipPreflightChecks(t *testing.T) {
	mockCLI := &azurecli.MockAzureCLI{}
	service := NewValidationService(mockCLI)

	tests := []struct {
		name       string
		flagValue  string
		shouldSkip bool
	}{
		{"skip_yes", "yes", true},
		{"skip_true", "true", true},
		{"skip_no", "no", false},
		{"skip_false", "false", false},
		{"skip_empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("skip-preflight", "", "Skip preflight checks")

			if tt.flagValue != "" {
				cmd.Flags().Set("skip-preflight", tt.flagValue)
			}

			result := service.ShouldSkipPreflightChecks(cmd)
			if result != tt.shouldSkip {
				t.Errorf("Expected ShouldSkipPreflightChecks to return %v for value '%s', got %v",
					tt.shouldSkip, tt.flagValue, result)
			}
		})
	}
}

func TestValidationServiceMethods(t *testing.T) {
	mockCLI := &azurecli.MockAzureCLI{}
	service := NewValidationService(mockCLI)

	// Create a minimal command for testing
	cmd := &cobra.Command{}
	cmd.Flags().String("flavor", "ITPro", "ArcBox flavor")
	cmd.Flags().String("location", "eastus", "Azure region")

	// Test that methods don't panic and return boolean values
	// Note: These will likely return false due to missing Azure environment,
	// but we're testing that the service layer works correctly

	t.Run("RunFullPreflightChecks", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RunFullPreflightChecks panicked: %v", r)
			}
		}()
		result := service.RunFullPreflightChecks(cmd)
		_ = result // We expect this might be false in test environment
	})

	t.Run("RunParameterValidation", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RunParameterValidation panicked: %v", r)
			}
		}()
		result := service.RunParameterValidation(cmd)
		_ = result
	})

	t.Run("RunQuotaChecks", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RunQuotaChecks panicked: %v", r)
			}
		}()
		result := service.RunQuotaChecks(cmd)
		_ = result
	})

	t.Run("ValidateConditionalRequirements", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ValidateConditionalRequirements panicked: %v", r)
			}
		}()
		result := service.ValidateConditionalRequirements(cmd)
		_ = result
	})

	t.Run("ValidateDeploymentParameters", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ValidateDeploymentParameters panicked: %v", r)
			}
		}()
		result := service.ValidateDeploymentParameters(cmd)
		_ = result
	})
}
