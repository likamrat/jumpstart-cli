// arcbox.go - ArcBox-specific preflight validation functions
package arcbox

import (
	"fmt"

	arcboxUtils "jumpstartcli/cmd/arcbox/utils"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/preflight/validator"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// RunArcBoxPreflightChecks performs comprehensive preflight validation for ArcBox deployment
func RunArcBoxPreflightChecks(cmd *cobra.Command) bool {
	// Extract parameters from command flags
	ctx := buildArcBoxValidationContext(cmd)

	// Create validation engine
	engine := validator.NewValidationEngine()

	// Run all applicable validations
	results := engine.ValidateAll(ctx)

	// Print results
	validator.PrintResults(results)

	// Return whether all checks passed (no errors)
	return !validator.HasErrors(results)
}

// RunParameterValidation runs only parameter-specific validation checks
func RunParameterValidation(cmd *cobra.Command) bool {
	ctx := buildArcBoxValidationContext(cmd)

	// Create engine with only parameter validators
	engine := &validator.ValidationEngine{}
	engine.RegisterValidator(&validator.SSHKeyValidator{})
	engine.RegisterValidator(&validator.WindowsPasswordValidator{})
	engine.RegisterValidator(&validator.ResourceTagsValidator{})
	engine.RegisterValidator(&validator.GitHubUsernameValidator{})
	engine.RegisterValidator(&validator.FlavorSpecificValidator{})

	results := engine.ValidateAll(ctx)

	// Print only errors for parameter validation
	hasErrors := false
	for _, result := range results {
		if !result.Passed && result.Severity == "error" {
			fmt.Printf("%s %s\n", utils.ErrorColor("❌"), utils.ErrorColor(result.Message))
			if result.Suggestion != "" {
				fmt.Printf("   💡 %s\n", utils.InfoColor(result.Suggestion))
			}
			if result.Details != "" {
				fmt.Printf("   %s\n", utils.DebugColor(result.Details))
			}
			hasErrors = true
		}
	}

	return !hasErrors
}

// RunArcBoxQuotaChecks performs optimized quota-only validation for standalone quota commands
func RunArcBoxQuotaChecks(cmd *cobra.Command) bool {
	// Extract parameters from command flags
	ctx := buildArcBoxValidationContext(cmd)

	// Create validation engine with only quota-related validators for speed
	engine := &validator.ValidationEngine{}

	// Add critical Azure check first (fail fast if Azure CLI not working)
	engine.RegisterValidator(&validator.AzureCLIHealthValidator{})

	// Add core infrastructure validators needed for quota checks
	engine.RegisterValidator(&validator.SubscriptionAccessValidator{})

	// NOTE: Resource provider validation is handled separately by the 'rp' command
	// This quota command focuses only on quota and SKU validation

	// Add the main quota validators
	engine.RegisterValidator(&validator.SKUAvailabilityValidator{})
	engine.RegisterValidator(&validator.QuotaValidator{})

	// Add region validation to ensure quota checks are meaningful
	engine.RegisterValidator(&validator.RegionValidator{})

	// Run validations
	results := engine.ValidateAll(ctx)

	// Print results with quota-focused messaging
	validator.PrintResults(results)

	// Return whether all checks passed (no errors)
	return !validator.HasErrors(results)
}

// ValidateConditionalRequirements checks flavor-specific requirements and prints errors
func ValidateConditionalRequirements(cmd *cobra.Command) bool {
	flavor, _ := cmd.Flags().GetString("flavor")
	// Normalize flavor for consistent display
	flavor = arcboxUtils.NormalizeFlavorCase(flavor)
	sshKey, _ := cmd.Flags().GetString("ssh-rsa-public-key")
	githubUser, _ := cmd.Flags().GetString("github-user")

	hasErrors := false

	// SSH key requirement for DevOps and DataOps
	if sshKey == "" && (flavor == "DevOps" || flavor == "DataOps") {
		utils.Error("You must provide --ssh-rsa-public-key for %s flavor.", flavor)
		fmt.Printf("   💡 %s\n", utils.InfoColor("Generate an SSH key pair: ssh-keygen -t rsa -b 4096"))
		hasErrors = true
	}

	// GitHub user requirement for DevOps
	if (githubUser == "" || githubUser == "microsoft") && flavor == "DevOps" {
		utils.Error("DevOps flavor requires your personal GitHub username for custom configurations.")
		fmt.Printf("   💡 %s\n", utils.InfoColor("Provide your GitHub username where you forked jumpstart-apps repository"))
		hasErrors = true
	}

	if hasErrors {
		fmt.Println()
		utils.ShowHelpWithoutTypes(cmd)
		return false
	}

	return true
}

// buildArcBoxValidationContext creates a validation context from command flags
func buildArcBoxValidationContext(cmd *cobra.Command) *validator.ValidationContext {
	ctx := &validator.ValidationContext{
		Solution:   "arcbox",
		Parameters: make(map[string]string),
		SkipChecks: []string{},
		SilentMode: false,                  // Show progress indicators to improve user experience during long operations
		AzureCLI:   azurecli.NewAzureCLI(), // Initialize Azure CLI wrapper
	}

	// Extract all flag values
	if flavor, _ := cmd.Flags().GetString("flavor"); flavor != "" {
		ctx.Flavor = arcboxUtils.NormalizeFlavorCase(flavor)
	}

	if location, _ := cmd.Flags().GetString("location"); location != "" {
		ctx.Location = location
		ctx.Parameters["location"] = location
	}

	if sshKey, _ := cmd.Flags().GetString("ssh-rsa-public-key"); sshKey != "" {
		ctx.Parameters["ssh-rsa-public-key"] = sshKey
	}

	if windowsPassword, _ := cmd.Flags().GetString("windows-password"); windowsPassword != "" {
		ctx.Parameters["windows-password"] = windowsPassword
	}

	if resourceTags, _ := cmd.Flags().GetString("resource-tags"); resourceTags != "" {
		// Only include resource tags if the flag was explicitly set by the user
		// (not just the default value)
		if cmd.Flags().Changed("resource-tags") {
			ctx.Parameters["resource-tags"] = resourceTags
		}
	}

	if githubUser, _ := cmd.Flags().GetString("github-user"); githubUser != "" {
		ctx.Parameters["github-user"] = githubUser
	}

	// Add other parameters as needed
	if resourceGroup, _ := cmd.Flags().GetString("resource-group"); resourceGroup != "" {
		ctx.Parameters["resource-group"] = resourceGroup
	}

	if windowsUser, _ := cmd.Flags().GetString("windows-user"); windowsUser != "" {
		ctx.Parameters["windows-user"] = windowsUser
	}

	return ctx
}

// GetFlavorSpecificChecks returns the specific checks that apply to a flavor
func GetFlavorSpecificChecks(flavor string) []string {
	switch flavor {
	case "DevOps":
		return []string{"ssh-rsa-public-key", "github-user"}
	case "DataOps":
		return []string{"ssh-rsa-public-key"}
	case "ITPro":
		return []string{} // No additional requirements
	default:
		return []string{}
	}
}

// PrintFlavorRequirements prints a summary of flavor-specific requirements
func PrintFlavorRequirements(flavor string) {
	// Normalize flavor for consistent display
	flavor = arcboxUtils.NormalizeFlavorCase(flavor)
	fmt.Printf("%s ArcBox %s flavor requirements:\n", utils.InfoColor("📋"), flavor)

	switch flavor {
	case "DevOps":
		fmt.Println("   • --ssh-rsa-public-key (for Linux VM access)")
		fmt.Println("   • --github-user (for custom configurations)")
		fmt.Println("   • Linux and Windows VMs will be deployed")
	case "DataOps":
		fmt.Println("   • --ssh-rsa-public-key (for Linux VM access)")
		fmt.Println("   • Data platform services and Linux VMs will be deployed")
	case "ITPro":
		fmt.Println("   • No additional requirements")
		fmt.Println("   • Windows VM and basic Arc services will be deployed")
	default:
		fmt.Printf("   • Unknown flavor: %s\n", flavor)
	}
	fmt.Println()
}
