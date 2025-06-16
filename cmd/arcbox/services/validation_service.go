package services

import (
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/preflight/arcbox"
	"jumpstartcli/internal/preflight/validator"

	"github.com/spf13/cobra"
)

// ValidationService handles validation operations for ArcBox deployments
// It encapsulates preflight checks, parameter validation, and environment verification.
//
// This service acts as a bridge between the command layer and the internal validation
// system, providing a clean interface for:
// - Comprehensive preflight checks (Azure CLI, subscriptions, quotas, etc.)
// - Parameter validation (SSH keys, passwords, tags, etc.)
// - Conditional requirement checks (flavor-specific validations)
// - Deployment readiness assessment
//
// The service leverages the existing validation framework in internal/preflight
// while providing a service-oriented interface that fits with the refactored
// command architecture.
type ValidationService struct {
	cli azurecli.AzureCLI
}

// NewValidationService creates a new ValidationService instance
func NewValidationService(cli azurecli.AzureCLI) *ValidationService {
	return &ValidationService{
		cli: cli,
	}
}

// RunFullPreflightChecks performs comprehensive preflight validation for ArcBox deployment
// This includes Azure CLI health, subscription access, resource providers, quotas, and parameters
func (s *ValidationService) RunFullPreflightChecks(cmd *cobra.Command) bool {
	return arcbox.RunArcBoxPreflightChecks(cmd)
}

// RunParameterValidation runs only parameter-specific validation checks
// This is useful for validating user input without checking infrastructure requirements
func (s *ValidationService) RunParameterValidation(cmd *cobra.Command) bool {
	return arcbox.RunParameterValidation(cmd)
}

// RunQuotaChecks performs optimized quota-only validation for standalone quota commands
// This includes SKU availability and vCPU quota checks but skips other infrastructure checks
func (s *ValidationService) RunQuotaChecks(cmd *cobra.Command) bool {
	return arcbox.RunArcBoxQuotaChecks(cmd)
}

// ValidateConditionalRequirements checks flavor-specific requirements
// This validates requirements that depend on the selected ArcBox flavor
func (s *ValidationService) ValidateConditionalRequirements(cmd *cobra.Command) bool {
	return arcbox.ValidateConditionalRequirements(cmd)
}

// CreateValidationContext creates a validation context from command flags
// This is used to prepare validation parameters for the validation engine
func (s *ValidationService) CreateValidationContext(cmd *cobra.Command) (*validator.ValidationContext, error) {
	ctx := &validator.ValidationContext{
		Solution:   "arcbox",
		Parameters: make(map[string]string),
		AzureCLI:   s.cli,
	}

	// Extract common parameters from command flags
	if cmd.Flags().Changed("flavor") {
		if flavor, err := cmd.Flags().GetString("flavor"); err == nil {
			ctx.Flavor = flavor
		}
	}

	if cmd.Flags().Changed("location") {
		if location, err := cmd.Flags().GetString("location"); err == nil {
			ctx.Location = location
		}
	}

	if cmd.Flags().Changed("subscription") {
		if subscription, err := cmd.Flags().GetString("subscription"); err == nil {
			ctx.Parameters["subscription"] = subscription
		}
	}

	// Extract skip-preflight flag if present
	if cmd.Flags().Changed("skip-preflight") {
		if skipPreflight, err := cmd.Flags().GetString("skip-preflight"); err == nil {
			if skipPreflight == "yes" || skipPreflight == "true" {
				ctx.SkipChecks = []string{"all"}
			}
		}
	}

	// Extract deployment-specific parameters
	deploymentFlags := []string{
		"windows-user", "windows-password", "admin-username",
		"ssh-rsa-public-key", "resource-tags", "github-user",
		"sql-server-edition", "bastion-sku", "naming-prefix",
	}

	for _, flag := range deploymentFlags {
		if cmd.Flags().Changed(flag) {
			if value, err := cmd.Flags().GetString(flag); err == nil && value != "" {
				ctx.Parameters[flag] = value
			}
		}
	}

	return ctx, nil
}

// ValidateDeploymentParameters validates parameters specific to deployment operations
// This is a convenience method for deployment command validation
func (s *ValidationService) ValidateDeploymentParameters(cmd *cobra.Command) bool {
	// Create validation context
	ctx, err := s.CreateValidationContext(cmd)
	if err != nil {
		return false
	}

	// Create validation engine with deployment-specific validators
	engine := &validator.ValidationEngine{}

	// Add parameter validators
	engine.RegisterValidator(&validator.SSHKeyValidator{})
	engine.RegisterValidator(&validator.WindowsPasswordValidator{})
	engine.RegisterValidator(&validator.ResourceTagsValidator{})
	engine.RegisterValidator(&validator.GitHubUsernameValidator{})
	engine.RegisterValidator(&validator.FlavorSpecificValidator{})

	// Run validations
	results := engine.ValidateAll(ctx)

	// Check if all validations passed
	return !validator.HasErrors(results)
}

// ShouldSkipPreflightChecks determines if preflight checks should be skipped
// based on command flags and configuration
func (s *ValidationService) ShouldSkipPreflightChecks(cmd *cobra.Command) bool {
	if cmd.Flags().Changed("skip-preflight") {
		if skipPreflight, err := cmd.Flags().GetString("skip-preflight"); err == nil {
			return skipPreflight == "yes" || skipPreflight == "true"
		}
	}
	return false
}
