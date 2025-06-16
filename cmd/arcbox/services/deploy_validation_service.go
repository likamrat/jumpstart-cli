package services

import (
	"fmt"

	"jumpstartcli/internal/preflight/arcbox"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// DeployValidationService handles all validation logic for deploy commands
type DeployValidationService struct{}

// NewDeployValidationService creates a new deployment validation service
func NewDeployValidationService() *DeployValidationService {
	return &DeployValidationService{}
}

// ValidateDeployFlags validates all deploy command flags
func (dvs *DeployValidationService) ValidateDeployFlags(cmd *cobra.Command) ValidationResult {
	if err := utils.ValidateAllFlags(cmd); err != nil {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("flag validation failed: %w", err),
		}
	}
	return ValidationResult{IsValid: true}
}

// ValidateRequiredArguments validates that all required arguments are provided
func (dvs *DeployValidationService) ValidateRequiredArguments(cmd *cobra.Command, requiredArguments []string) ValidationResult {
	if !utils.PrintMissingRequiredFlagsError(cmd, requiredArguments) {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("deploy command missing required arguments: --location, --resource-group, --windows-user, --flavor are required"),
		}
	}
	return ValidationResult{IsValid: true}
}

// ValidateConditionalRequirements validates flavor-specific requirements
func (dvs *DeployValidationService) ValidateConditionalRequirements(cmd *cobra.Command) ValidationResult {
	if !arcbox.ValidateConditionalRequirements(cmd) {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("deployment flavor requirements validation failed: missing required parameters for selected flavor"),
		}
	}
	return ValidationResult{IsValid: true}
}

// RunPreflightChecks runs comprehensive preflight checks
func (dvs *DeployValidationService) RunPreflightChecks(cmd *cobra.Command, skipPreflight bool) ValidationResult {
	if skipPreflight {
		fmt.Println(utils.WarnColor("⚠️  [WARNING] Preflight checks have been skipped. Deployment may fail if prerequisites are not met."))
		return ValidationResult{IsValid: true}
	}

	// Run comprehensive preflight checks including parameter validation
	if !arcbox.RunArcBoxPreflightChecks(cmd) {
		return ValidationResult{
			IsValid: false,
			Error:   fmt.Errorf("deployment preflight checks failed: environment validation did not pass"),
		}
	}

	// Success message is already printed by PrintResults() in the validation engine
	return ValidationResult{IsValid: true}
}

// ValidateAllDeployRequirements runs all validation steps for deploy command
func (dvs *DeployValidationService) ValidateAllDeployRequirements(cmd *cobra.Command) ValidationResult {
	// Required arguments for deploy command
	requiredArguments := []string{"location", "resource-group", "windows-user", "flavor"}

	// Step 1: Validate ALL flags first (before any other operations)
	if result := dvs.ValidateDeployFlags(cmd); !result.IsValid {
		return result
	}

	// Step 2: Validate required arguments
	if result := dvs.ValidateRequiredArguments(cmd, requiredArguments); !result.IsValid {
		return result
	}

	// Step 3: Validate conditional requirements (before preflight checks)
	if result := dvs.ValidateConditionalRequirements(cmd); !result.IsValid {
		return result
	}

	// Step 4: Check if preflight checks should be skipped and run them if needed
	skipPreflight := utils.GetBooleanFlagValue(cmd, "skip-preflight")
	if result := dvs.RunPreflightChecks(cmd, skipPreflight); !result.IsValid {
		return result
	}

	return ValidationResult{IsValid: true}
}
