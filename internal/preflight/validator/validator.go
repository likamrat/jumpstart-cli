// validator.go - Extensible preflight validation system for Jumpstart CLI
package validator

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"

	"jumpstartcli/internal/auth"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/resourceproviders"
	"jumpstartcli/internal/utils"
)

// ValidationResult represents the result of a single validation check
type ValidationResult struct {
	CheckName  string
	Passed     bool
	Message    string
	Severity   string // "error", "warning", "info"
	Suggestion string
	Details    string
}

// ValidationContext holds all the information needed for validation
type ValidationContext struct {
	Solution   string // "arcbox", "localbox", "agora"
	Flavor     string // "ITPro", "DevOps", "DataOps", etc.
	Location   string
	Parameters map[string]string
	SkipChecks []string
	SilentMode bool              // When true, suppress progress messages
	AzureCLI   azurecli.AzureCLI // Azure CLI instance for operations
}

// Validator interface defines a validation check
type Validator interface {
	Name() string
	Description() string
	Validate(ctx *ValidationContext) ValidationResult
	IsApplicable(ctx *ValidationContext) bool
}

// ValidationEngine manages and executes validation checks
type ValidationEngine struct {
	validators []Validator
}

// NewValidationEngine creates a new validation engine with default validators
func NewValidationEngine() *ValidationEngine {
	engine := &ValidationEngine{}

	// 1. Azure CLI is healthy and authenticated (critical first check)
	engine.RegisterValidator(&AzureCLIHealthValidator{})

	// 2. Azure subscription context set successfully
	engine.RegisterValidator(&SubscriptionAccessValidator{})

	// 3. All required resource providers are registered
	engine.RegisterValidator(&ResourceProviderValidator{})

	// 4. SSH key and GitHub user requirement validators (before format validation)
	engine.RegisterValidator(&SSHKeyRequirementValidator{})
	engine.RegisterValidator(&GitHubUserRequirementValidator{})

	// 5. Validates flavor-specific parameter requirements completed
	engine.RegisterValidator(&FlavorSpecificValidator{})

	// 6. Windows password meets complexity requirements
	engine.RegisterValidator(&WindowsPasswordValidator{})

	// Additional fast local validators (format validation)
	engine.RegisterValidator(&SSHKeyValidator{})
	engine.RegisterValidator(&ResourceTagsValidator{})
	engine.RegisterValidator(&GitHubUsernameValidator{})

	// 7. Region validation
	engine.RegisterValidator(&RegionValidator{})

	// 8. VM SKUs availability
	engine.RegisterValidator(&SKUAvailabilityValidator{})

	// 9. vCPU quota validation (last due to slowness - can take 60+ seconds)
	engine.RegisterValidator(&QuotaValidator{})

	return engine
}

// RegisterValidator adds a new validator to the engine
func (e *ValidationEngine) RegisterValidator(validator Validator) {
	e.validators = append(e.validators, validator)
}

// ValidateAll runs all applicable validators and returns results
func (e *ValidationEngine) ValidateAll(ctx *ValidationContext) []ValidationResult {
	var results []ValidationResult

	for _, validator := range e.validators {
		if validator.IsApplicable(ctx) {
			// Skip if this check is in the skip list
			skip := false
			for _, skipCheck := range ctx.SkipChecks {
				if skipCheck == validator.Name() {
					skip = true
					break
				}
			}

			if !skip {
				// Show progress indicator with spinner animation for each validator (unless in silent mode)
				if !ctx.SilentMode {
					result := e.runValidatorWithSpinner(validator, ctx)
					results = append(results, result)
				} else {
					// Silent mode: just run validator without any progress display
					result := validator.Validate(ctx)
					results = append(results, result)
				}
			}
		}
	}

	return results
}

// runValidatorWithSpinner runs a validator with animated spinner and progress display
func (e *ValidationEngine) runValidatorWithSpinner(validator Validator, ctx *ValidationContext) ValidationResult {
	description := validator.Description()

	// Animation frames: Unicode spinner for smooth animation
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIdx := 0
	stopSpinner := make(chan struct{})
	spinnerDone := make(chan struct{})

	// Show initial message
	fmt.Printf("🔍 Running %s...", description)

	// Hide cursor before starting animation
	fmt.Print("\033[?25l")

	// Start spinner animation in goroutine
	go func() {
		for {
			select {
			case <-stopSpinner:
				// Clear the spinner line completely
				fmt.Printf("\r\033[2K")
				// Restore cursor when animation stops
				fmt.Print("\033[?25h")
				close(spinnerDone)
				return
			default:
				// Show spinner with current frame
				fmt.Printf("\r🔍 Running %s... %s", description, frames[frameIdx%len(frames)])
				frameIdx++
				time.Sleep(80 * time.Millisecond) // Smooth 80ms animation
			}
		}
	}()

	// Run the actual validation
	result := validator.Validate(ctx)

	// Stop the spinner
	close(stopSpinner)
	<-spinnerDone

	// Show final result for this validator
	// Skip display if message is empty
	if result.Message != "" {
		switch {
		case result.Passed:
			fmt.Printf("✅ %s\n", result.Message)
		case result.Severity == "error":
			fmt.Printf("❌ %s\n", result.Message)
			if result.Suggestion != "" {
				fmt.Printf("   💡 %s\n", result.Suggestion)
			}
		case result.Severity == "warning":
			fmt.Printf("⚠️ %s\n", result.Message)
			if result.Suggestion != "" {
				fmt.Printf("   💡 %s\n", result.Suggestion)
			}
		default:
			fmt.Printf("ℹ️ %s\n", result.Message)
		}
	} else {
		// If no message, just show completion
		fmt.Printf("✅ %s completed\n", description)
	}

	return result
}

// HasErrors returns true if any validation results contain errors
func HasErrors(results []ValidationResult) bool {
	for _, result := range results {
		if !result.Passed && result.Severity == "error" {
			return true
		}
	}
	return false
}

// PrintResults prints validation results summary (individual results are now shown in real-time)
func PrintResults(results []ValidationResult) {
	if len(results) == 0 {
		fmt.Println("📋 No preflight validation checks to run.")
		return
	}

	// Count errors and warnings for summary
	errorCount := 0
	warningCount := 0

	for _, result := range results {
		// Skip results with empty messages (used for validators that want to run but not display)
		if result.Message == "" {
			continue
		}

		switch {
		case result.Severity == "error" && !result.Passed:
			errorCount++
		case result.Severity == "warning":
			warningCount++
		}
	}

	// Print final summary
	if errorCount > 0 || warningCount > 0 {
		fmt.Printf("⚠️ Summary: %d errors, %d warnings\n", errorCount, warningCount)
	} else {
		fmt.Printf("✅ All preflight validation checks passed!\n")
	}
}

// --- Built-in Validators ---

// SSHKeyValidator validates SSH RSA public key format and requirements
type SSHKeyValidator struct{}

func (v *SSHKeyValidator) Name() string {
	return "ssh-key-validation"
}

func (v *SSHKeyValidator) Description() string {
	return "Validates SSH RSA public key format and requirements"
}

func (v *SSHKeyValidator) IsApplicable(ctx *ValidationContext) bool {
	// Only apply to ArcBox DevOps and DataOps flavors (case-insensitive)
	flavor := strings.ToLower(ctx.Flavor)
	return ctx.Solution == "arcbox" && (flavor == "devops" || flavor == "dataops")
}

func (v *SSHKeyValidator) Validate(ctx *ValidationContext) ValidationResult {
	sshKey := ctx.Parameters["ssh-rsa-public-key"]

	// Skip validation if SSH key is not provided - this is handled by FlavorSpecificValidator
	if sshKey == "" {
		return ValidationResult{
			CheckName: v.Name(),
			Passed:    true,
			Message:   "SSH key format validation skipped (no key provided)",
			Severity:  "info",
		}
	}

	// Validate SSH key format
	if !isValidSSHKey(sshKey) {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    "SSH RSA public key format is invalid",
			Severity:   "error",
			Suggestion: "Ensure your SSH key starts with 'ssh-rsa' and contains valid base64 content",
			Details:    "SSH RSA public keys should follow the format: ssh-rsa <base64-content> [comment]",
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   "SSH RSA public key format is valid",
		Severity:  "info",
	}
}

// WindowsPasswordValidator validates Windows password complexity requirements
type WindowsPasswordValidator struct{}

func (v *WindowsPasswordValidator) Name() string {
	return "windows-password-validation"
}

func (v *WindowsPasswordValidator) Description() string {
	return "Validates Windows password complexity requirements"
}

func (v *WindowsPasswordValidator) IsApplicable(ctx *ValidationContext) bool {
	// Apply to ArcBox when windows-password is provided
	return ctx.Solution == "arcbox" && ctx.Parameters["windows-password"] != ""
}

func (v *WindowsPasswordValidator) Validate(ctx *ValidationContext) ValidationResult {
	password := ctx.Parameters["windows-password"]

	if !isValidWindowsPassword(password) {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    "Windows password does not meet complexity requirements",
			Severity:   "error",
			Suggestion: "Password must be 12-123 characters and contain 3 of: lowercase, uppercase, numbers, special characters",
			Details:    "Azure requires strong passwords for Windows VMs to ensure security",
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   "Windows password meets complexity requirements",
		Severity:  "info",
	}
}

// ResourceTagsValidator validates JSON syntax for resource tags
type ResourceTagsValidator struct{}

func (v *ResourceTagsValidator) Name() string {
	return "resource-tags-validation"
}

func (v *ResourceTagsValidator) Description() string {
	return "Validates JSON syntax for resource tags"
}

func (v *ResourceTagsValidator) IsApplicable(ctx *ValidationContext) bool {
	// Apply when resource-tags parameter is provided
	return ctx.Parameters["resource-tags"] != ""
}

func (v *ResourceTagsValidator) Validate(ctx *ValidationContext) ValidationResult {
	tagsJSON := ctx.Parameters["resource-tags"]

	var tags map[string]interface{}
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    "Resource tags JSON syntax is invalid",
			Severity:   "error",
			Suggestion: `Ensure tags are valid JSON, e.g., {"Environment":"Dev","Owner":"TeamA"}`,
			Details:    fmt.Sprintf("JSON parsing error: %v", err),
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   fmt.Sprintf("Resource tags JSON is valid (%d tags)", len(tags)),
		Severity:  "info",
	}
}

// GitHubUsernameValidator validates GitHub username format
type GitHubUsernameValidator struct{}

func (v *GitHubUsernameValidator) Name() string {
	return "github-username-validation"
}

func (v *GitHubUsernameValidator) Description() string {
	return "Validates GitHub username format and requirements"
}

func (v *GitHubUsernameValidator) IsApplicable(ctx *ValidationContext) bool {
	// Apply to ArcBox DevOps flavor when github-user is provided (case-insensitive)
	flavor := strings.ToLower(ctx.Flavor)
	return ctx.Solution == "arcbox" && flavor == "devops" && ctx.Parameters["github-user"] != ""
}

func (v *GitHubUsernameValidator) Validate(ctx *ValidationContext) ValidationResult {
	username := ctx.Parameters["github-user"]
	friendlyFlavorName := "DevOps"

	if username == "" && strings.ToLower(ctx.Flavor) == "devops" {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    fmt.Sprintf("GitHub username is required for %s flavor", friendlyFlavorName),
			Severity:   "error",
			Suggestion: "Provide your GitHub username with --github-user where you have forked the jumpstart-apps repository",
			Details:    fmt.Sprintf("%s flavor requires custom configurations from your forked repository", friendlyFlavorName),
		}
	}

	if !isValidGitHubUsername(username) {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    "GitHub username format is invalid",
			Severity:   "error",
			Suggestion: "Provide only the username (e.g., 'johndoe', not 'github.com/johndoe' or email)",
			Details:    "GitHub usernames can only contain alphanumeric characters and hyphens",
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   fmt.Sprintf("GitHub username '%s' is valid", username),
		Severity:  "info",
	}
}

// FlavorSpecificValidator ensures all flavor-specific requirements are met
type FlavorSpecificValidator struct{}

func (v *FlavorSpecificValidator) Name() string {
	return "flavor-specific-requirements"
}

func (v *FlavorSpecificValidator) Description() string {
	return "Validates flavor-specific parameter requirements"
}

func (v *FlavorSpecificValidator) IsApplicable(ctx *ValidationContext) bool {
	return ctx.Solution == "arcbox"
}

func (v *FlavorSpecificValidator) Validate(ctx *ValidationContext) ValidationResult {
	switch ctx.Flavor {
	default:
		// For ITPro and other flavors with no special requirements, don't show a message
		// This avoids confusing messages like "All ITPro flavor requirements are satisfied"
		// when ITPro has no actual special requirements
		return ValidationResult{
			CheckName: v.Name(),
			Passed:    true,
			Message:   "", // Empty message means this result won't be displayed
			Severity:  "info",
		}
	}
}

// --- Infrastructure Validators ---

// AzureCLIHealthValidator validates Azure CLI is healthy and logged in
type AzureCLIHealthValidator struct{}

func (v *AzureCLIHealthValidator) Name() string {
	return "azure-cli-health"
}

func (v *AzureCLIHealthValidator) Description() string {
	return "Checks Azure CLI health and authentication status"
}

func (v *AzureCLIHealthValidator) IsApplicable(ctx *ValidationContext) bool {
	// Always check Azure CLI health for deployment operations
	return true
}

func (v *AzureCLIHealthValidator) Validate(ctx *ValidationContext) ValidationResult {
	// Check if Azure CLI is responsive and logged in
	if err := checkAzureCLIHealth(ctx.AzureCLI); err != nil {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    "Azure CLI is not responding or not logged in",
			Severity:   "error",
			Suggestion: "Run 'az login' to authenticate with Azure, then try again",
			Details:    fmt.Sprintf("Azure CLI error: %v", err),
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   "Azure CLI is healthy and authenticated",
		Severity:  "info",
	}
}

// SubscriptionAccessValidator validates subscription access and sets context
type SubscriptionAccessValidator struct{}

func (v *SubscriptionAccessValidator) Name() string {
	return "subscription-access"
}

func (v *SubscriptionAccessValidator) Description() string {
	return "Validates Azure subscription access and sets context"
}

func (v *SubscriptionAccessValidator) IsApplicable(ctx *ValidationContext) bool {
	// Always validate subscription access for Azure operations
	return true
}

func (v *SubscriptionAccessValidator) Validate(ctx *ValidationContext) ValidationResult {
	subscription := getSubscriptionFromContext(ctx)

	if subscription == "" {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    "No Azure subscription specified or available",
			Severity:   "error",
			Suggestion: "Specify subscription with --subscription flag or set AZURE_SUBSCRIPTION_ID environment variable",
		}
	}

	// Set subscription context
	if err := setAzureSubscription(ctx.AzureCLI, subscription); err != nil {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    "Failed to set Azure subscription context",
			Severity:   "error",
			Suggestion: "Verify subscription ID/name is correct and you have access",
			Details:    fmt.Sprintf("Subscription error: %v", err),
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   fmt.Sprintf("Azure subscription context set successfully (%s)", subscription),
		Severity:  "info",
	}
}

// ResourceProviderValidator validates required resource providers are registered
type ResourceProviderValidator struct{}

func (v *ResourceProviderValidator) Name() string {
	return "resource-providers"
}

func (v *ResourceProviderValidator) Description() string {
	return "Validates required Azure resource providers are registered"
}

func (v *ResourceProviderValidator) IsApplicable(ctx *ValidationContext) bool {
	// Check resource providers for all Azure deployments
	return ctx.Solution == "arcbox"
}

func (v *ResourceProviderValidator) Validate(ctx *ValidationContext) ValidationResult {
	// Get required providers for the solution
	config := getResourceProviderConfig(ctx.Solution)
	if config == nil {
		return ValidationResult{
			CheckName: v.Name(),
			Passed:    true,
			Message:   "No resource provider requirements defined",
			Severity:  "info",
		}
	}

	// Check all required providers
	allRegistered, missingProviders := checkAllResourceProviders(ctx.AzureCLI, *config)

	if !allRegistered {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    fmt.Sprintf("Required resource providers not registered (%d missing)", len(missingProviders)),
			Severity:   "error",
			Suggestion: "Run 'js arcbox preflight rp register' to register missing providers",
			Details:    fmt.Sprintf("Missing providers: %s", strings.Join(missingProviders, ", ")),
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   fmt.Sprintf("All required resource providers are registered (%d checked)", len(config.RequiredProviders)),
		Severity:  "info",
	}
}

// QuotaValidator validates vCPU quota availability for selected flavor
type QuotaValidator struct{}

func (v *QuotaValidator) Name() string {
	return "vcpu-quota"
}

func (v *QuotaValidator) Description() string {
	return "Validates sufficient vCPU quota for deployment"
}

func (v *QuotaValidator) IsApplicable(ctx *ValidationContext) bool {
	// Check quota for all flavors except when custom SKUs are specified
	return ctx.Flavor != "" && ctx.Location != ""
}

func (v *QuotaValidator) Validate(ctx *ValidationContext) ValidationResult {
	flavor := ctx.Flavor
	location := ctx.Location

	// Get SKUs for the flavor
	skus := getFlavorSKUsForValidation(flavor)
	if len(skus) == 0 {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    fmt.Sprintf("Unknown flavor for quota validation: %s", flavor),
			Severity:   "error",
			Suggestion: "Use a valid flavor: ITPro, DevOps, DataOps, or all",
		}
	}

	// Check quota for each SKU using Azure CLI wrapper
	var failedChecks []string
	var warnings []string
	totalRequired := 0

	for _, sku := range skus {
		required := getRequiredVCPUForSKU(sku)
		totalRequired += required

		// Use Azure CLI wrapper for quota checking
		quotaOK := false
		if ctx.AzureCLI != nil {
			// Get quota data for the region using Azure CLI wrapper
			usages, err := ctx.AzureCLI.ListVMUsage(location)
			if err == nil {
				// Map SKU to its quota family name
				familyName := mapSKUToFamilyQuotaName(sku)

				// Look for the quota usage entry
				for _, usage := range usages {
					val, hasVal := usage.Name["value"]
					localizedValue, hasLocalized := usage.Name["localizedValue"]

					if !hasVal || !hasLocalized {
						continue
					}

					// Normalize for comparison
					familyNorm := strings.ReplaceAll(strings.ToLower(familyName), " ", "")
					valNorm := strings.ReplaceAll(strings.ToLower(val), " ", "")
					localizedNorm := strings.ReplaceAll(strings.ToLower(localizedValue), " ", "")

					if familyName != "" && (valNorm == familyNorm || localizedNorm == familyNorm) {
						available := usage.Limit - usage.CurrentValue
						quotaOK = available >= required
						break
					}
				}

				// Fallback: try total regional vCPU quota
				if !quotaOK {
					for _, usage := range usages {
						val, hasVal := usage.Name["value"]
						localizedValue, hasLocalized := usage.Name["localizedValue"]

						if !hasVal || !hasLocalized {
							continue
						}

						valNorm := strings.ReplaceAll(strings.ToLower(val), " ", "")
						localizedNorm := strings.ReplaceAll(strings.ToLower(localizedValue), " ", "")

						if strings.Contains(valNorm, "totalregionalvcpu") || strings.Contains(localizedNorm, "totalregionalvcpu") {
							available := usage.Limit - usage.CurrentValue
							quotaOK = available >= required
							break
						}
					}
				}
			}
		} else {
			// Azure CLI wrapper is required for quota validation
			failedChecks = append(failedChecks, fmt.Sprintf("%s (Azure CLI wrapper required)", sku))
			continue
		}

		if !quotaOK {
			failedChecks = append(failedChecks, fmt.Sprintf("%s (%d vCPU)", sku, required))
		}
	}

	if len(failedChecks) > 0 {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    fmt.Sprintf("Insufficient vCPU quota in \"%s\" (%d vCPU required)", utils.GetRegionDisplayName(location), totalRequired),
			Severity:   "error",
			Suggestion: "Request quota increase in Azure portal or choose a different region",
			Details:    fmt.Sprintf("Failed checks: %s", strings.Join(failedChecks, ", ")),
		}
	}

	message := fmt.Sprintf("Sufficient vCPU quota available in \"%s\" (%d vCPU required)", utils.GetRegionDisplayName(location), totalRequired)
	if len(warnings) > 0 {
		message += fmt.Sprintf(" - %d warnings", len(warnings))
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   message,
		Severity:  "info",
	}
}

// RegionValidator validates Azure region support and availability
type RegionValidator struct{}

func (v *RegionValidator) Name() string {
	return "region-support"
}

func (v *RegionValidator) Description() string {
	return "Validates Azure region support and resource availability"
}

func (v *RegionValidator) IsApplicable(ctx *ValidationContext) bool {
	// Validate region for all deployments
	return ctx.Location != ""
}

func (v *RegionValidator) Validate(ctx *ValidationContext) ValidationResult {
	location := ctx.Location

	// Validate region format and availability
	if !isValidAzureRegion(location) {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    fmt.Sprintf("Invalid or unsupported Azure region: \"%s\"", utils.GetRegionDisplayName(location)),
			Severity:   "error",
			Suggestion: "Use a valid Azure region name like 'eastus', 'westus2', 'centralus'",
		}
	}

	// Check region capabilities for the solution
	if ctx.Solution == "arcbox" && !isRegionSupportedForArcBox(location) {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    fmt.Sprintf("Region \"%s\" has limited ArcBox service support", utils.GetRegionDisplayName(location)),
			Severity:   "error",
			Suggestion: "For optimal ArcBox experience, use a fully supported region like eastus, westus2, or westeurope",
			Details:    "Some ArcBox features may not be available in all regions due to service dependencies",
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   fmt.Sprintf("Region \"%s\" is valid and supports ArcBox deployment", utils.GetRegionDisplayName(location)),
		Severity:  "info",
	}
}

// SKUAvailabilityValidator validates that required VM SKUs are available in the target region
type SKUAvailabilityValidator struct{}

func (v *SKUAvailabilityValidator) Name() string {
	return "sku-availability"
}

func (v *SKUAvailabilityValidator) Description() string {
	return "Validates VM SKUs are available in the target region"
}

func (v *SKUAvailabilityValidator) IsApplicable(ctx *ValidationContext) bool {
	// Check SKU availability for all flavors
	return ctx.Flavor != "" && ctx.Location != ""
}

func (v *SKUAvailabilityValidator) Validate(ctx *ValidationContext) ValidationResult {
	flavor := ctx.Flavor
	location := ctx.Location

	// Get SKUs for the flavor
	skus := getFlavorSKUsForValidation(flavor)
	if len(skus) == 0 {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    fmt.Sprintf("Unknown flavor for SKU validation: %s", flavor),
			Severity:   "error",
			Suggestion: "Use a valid flavor: ITPro, DevOps, DataOps, or all",
		}
	}

	// Check SKU availability using Azure CLI wrapper
	var unavailableSkus []string
	if ctx.AzureCLI != nil {
		// Use Azure CLI wrapper to check SKU availability
		for _, sku := range skus {
			available, err := ctx.AzureCLI.CheckSKUAvailability(sku, location)
			if err != nil || !available {
				unavailableSkus = append(unavailableSkus, sku)
			}
		}
	} else {
		// Azure CLI wrapper is required for SKU availability validation
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    "Azure CLI wrapper required for SKU availability validation",
			Severity:   "error",
			Suggestion: "Ensure Azure CLI wrapper is properly initialized",
		}
	}

	if len(unavailableSkus) > 0 {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    fmt.Sprintf("VM SKUs not available in %s", location),
			Severity:   "error",
			Suggestion: "Choose a different region or contact Azure support for SKU availability",
			Details:    fmt.Sprintf("Unavailable SKUs: %s", strings.Join(unavailableSkus, ", ")),
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   fmt.Sprintf("All required VM SKUs are available in \"%s\"", utils.GetRegionDisplayName(location)),
		Severity:  "info",
	}
}

// --- Helper Functions ---

// isValidSSHKey validates SSH RSA public key format
func isValidSSHKey(key string) bool {
	// Basic SSH RSA public key validation
	parts := strings.Fields(key)
	if len(parts) < 2 {
		return false
	}

	// Must start with ssh-rsa
	if parts[0] != "ssh-rsa" {
		return false
	}

	// Base64 content validation (basic check)
	base64Pattern := regexp.MustCompile(`^[A-Za-z0-9+/]+=*$`)
	return base64Pattern.MatchString(parts[1]) && len(parts[1]) > 100
}

// isValidWindowsPassword validates Windows password complexity
func isValidWindowsPassword(password string) bool {
	if len(password) < 12 || len(password) > 123 {
		return false
	}

	complexity := 0
	patterns := []string{
		`[a-z]`,        // lowercase
		`[A-Z]`,        // uppercase
		`[0-9]`,        // numbers
		`[^a-zA-Z0-9]`, // special characters
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, password); matched {
			complexity++
		}
	}

	return complexity >= 3
}

// isValidGitHubUsername validates GitHub username format
func isValidGitHubUsername(username string) bool {
	// Check for common invalid formats
	if strings.Contains(username, "@") ||
		strings.Contains(username, ".com") ||
		strings.Contains(username, "github.com/") ||
		strings.Contains(username, "GitHub/") {
		return false
	}

	// GitHub username pattern: alphanumeric and hyphens, 1-39 characters
	// Cannot start or end with hyphen, cannot have consecutive hyphens
	pattern := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)
	if !pattern.MatchString(username) {
		return false
	}

	// Check for consecutive hyphens
	if strings.Contains(username, "--") {
		return false
	}

	// Check length
	return len(username) >= 1 && len(username) <= 39
}

// ValidateEmail validates email format (helper for future use)
func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// --- Helper functions for infrastructure validation ---

func checkAzureCLIHealth(azCLI azurecli.AzureCLI) error {
	return auth.CheckAzureAuthentication(azCLI)
}

func getSubscriptionFromContext(ctx *ValidationContext) string {
	if sub, ok := ctx.Parameters["subscription"]; ok && sub != "" {
		return sub
	}
	if env := os.Getenv("AZURE_SUBSCRIPTION_ID"); env != "" {
		return env
	}
	// Fallback: use Azure CLI wrapper
	if ctx.AzureCLI != nil {
		sub, err := ctx.AzureCLI.GetCurrentSubscription()
		if err == nil && sub != nil {
			return sub.ID
		}
	}
	return ""
}

func setAzureSubscription(azCLI azurecli.AzureCLI, subID string) error {
	if subID == "" {
		return fmt.Errorf("subscription ID is empty")
	}
	return azCLI.SetSubscription(subID)
}

func getResourceProviderConfig(solution string) *resourceproviders.ResourceProviderConfig {
	switch strings.ToLower(solution) {
	case "arcbox":
		config := resourceproviders.GetArcBoxProviders()
		return &config
	default:
		return nil
	}
}

func checkAllResourceProviders(azCLI azurecli.AzureCLI, config resourceproviders.ResourceProviderConfig) (bool, []string) {
	missing := false
	missingProviders := []string{}

	for _, rp := range config.RequiredProviders {
		isRegistered, err := resourceproviders.CheckProviderRegistration(azCLI, rp)
		if err != nil || !isRegistered {
			missing = true
			missingProviders = append(missingProviders, rp)
		}
	}

	return !missing, missingProviders
}

func getFlavorSKUsForValidation(flavor string) []string {
	// This should use the same logic as the existing getFlavorSKUs function
	switch strings.ToLower(flavor) {
	case "itpro":
		return []string{"Standard_D8s_v5"}
	case "devops":
		return []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}
	case "dataops":
		return []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}
	case "all":
		// Return all unique SKUs across all flavors
		return []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}
	default:
		return []string{}
	}
}

func getRequiredVCPUForSKU(sku string) int {
	// Map SKUs to their vCPU requirements
	vcpuMap := map[string]int{
		"Standard_D8s_v5": 8,
		"Standard_D8s_v4": 8,
		"Standard_B2ms":   2,
		"Standard_B4ms":   4,
		"Standard_B8ms":   8,
	}
	if vcpu, ok := vcpuMap[sku]; ok {
		return vcpu
	}
	return 1 // Default fallback
}

// CheckQuotaForSKU is an exported wrapper for quota checking
// DEPRECATED: Use CheckQuotaForSKUWithCLI directly
func CheckQuotaForSKU(sku string, required int, region, subscription, flavor string) (bool, int, int, int) {
	// Use default Azure CLI instance
	defaultAzCLI := azurecli.NewAzureCLI()
	return CheckQuotaForSKUWithCLI(defaultAzCLI, sku, required, region, flavor)
}

// CheckQuotaForSKUWithCLI is the preferred method that uses Azure CLI wrapper
func CheckQuotaForSKUWithCLI(azCLI azurecli.AzureCLI, sku string, required int, region, flavor string) (bool, int, int, int) {
	// Get quota data for the region using Azure CLI wrapper
	usages, err := azCLI.ListVMUsage(region)
	if err != nil {
		return false, 0, 0, 0
	}

	// Map SKU to its quota family name
	familyName := mapSKUToFamilyQuotaName(sku)

	// Look for the quota usage entry
	for _, usage := range usages {
		val, hasVal := usage.Name["value"]
		localizedValue, hasLocalized := usage.Name["localizedValue"]

		if !hasVal || !hasLocalized {
			continue
		}

		// Normalize for comparison
		familyNorm := strings.ReplaceAll(strings.ToLower(familyName), " ", "")
		valNorm := strings.ReplaceAll(strings.ToLower(val), " ", "")
		localizedNorm := strings.ReplaceAll(strings.ToLower(localizedValue), " ", "")

		if familyName != "" && (valNorm == familyNorm || localizedNorm == familyNorm) {
			current := usage.CurrentValue
			limit := usage.Limit
			available := limit - current
			return available >= required, current, limit, available
		}
	}

	// Fallback: try total regional vCPU quota
	for _, usage := range usages {
		val, hasVal := usage.Name["value"]
		localizedValue, hasLocalized := usage.Name["localizedValue"]

		if !hasVal || !hasLocalized {
			continue
		}

		valNorm := strings.ReplaceAll(strings.ToLower(val), " ", "")
		localizedNorm := strings.ReplaceAll(strings.ToLower(localizedValue), " ", "")

		if strings.Contains(valNorm, "totalregionalvcpu") || strings.Contains(localizedNorm, "totalregionalvcpu") {
			current := usage.CurrentValue
			limit := usage.Limit
			available := limit - current
			return available >= required, current, limit, available
		}
	}

	return false, 0, 0, 0
}

// CheckBatchSKUAvailability is an exported wrapper for SKU availability checking
// DEPRECATED: Use CheckBatchSKUAvailabilityWithCLI directly
func CheckBatchSKUAvailability(skus []string, region, subscription string) []string {
	// Use default Azure CLI instance
	defaultAzCLI := azurecli.NewAzureCLI()
	return CheckBatchSKUAvailabilityWithCLI(defaultAzCLI, skus, region)
}

// CheckBatchSKUAvailabilityWithCLI is the preferred method that uses Azure CLI wrapper
func CheckBatchSKUAvailabilityWithCLI(azCLI azurecli.AzureCLI, skus []string, region string) []string {
	var unavailable []string
	for _, sku := range skus {
		available, err := azCLI.CheckSKUAvailability(sku, region)
		if err != nil || !available {
			unavailable = append(unavailable, sku)
		}
	}
	return unavailable
}

// mapSKUToFamilyQuotaName maps a VM SKU to its Azure vCPU family quota name
func mapSKUToFamilyQuotaName(sku string) string {
	sku = strings.TrimPrefix(sku, "Standard_")
	parts := strings.Split(sku, "_")

	var main, ver string
	if len(parts) >= 2 {
		// Standard pattern: D8s_v5 -> main=D8s, ver=v5
		main = parts[0]
		ver = parts[1]
	} else if len(parts) == 1 {
		// B-series pattern: B2ms -> main=B2ms, ver=""
		main = parts[0]
		ver = ""

		// Special handling for B-series SKUs - they all map to "Standard BS Family vCPUs"
		if strings.HasPrefix(strings.ToUpper(main), "B") {
			return "Standard BS Family vCPUs"
		}
	} else {
		return ""
	}

	// Remove digits from main part, keep only letters
	letters := ""
	for _, r := range main {
		if r >= '0' && r <= '9' {
			continue
		}
		letters += string(r)
	}

	// If no letters found, return empty string
	if letters == "" {
		return ""
	}

	// Validate that this looks like a real Azure SKU pattern
	// Azure SKUs typically have 1-3 letters, some digits, and optional 's'
	if len(letters) > 4 || !isValidAzureSKUPattern(main) {
		return ""
	}

	// If ends with 's', keep it (e.g. D8s → Ds)
	if strings.HasSuffix(main, "s") && !strings.HasSuffix(letters, "s") {
		letters += "s"
	}

	family := "Standard " + strings.ToUpper(letters[:1]) + letters[1:] + ver + " Family vCPUs"
	return family
}

// isValidAzureSKUPattern checks if the SKU follows Azure naming patterns
func isValidAzureSKUPattern(sku string) bool {
	// Azure SKUs typically start with letters, followed by digits, and optionally end with 's'
	// Examples: D8s, F4s, B2ms, E16s, etc.
	if len(sku) == 0 {
		return false
	}

	// Must start with a letter
	if sku[0] < 'A' || (sku[0] > 'Z' && sku[0] < 'a') || sku[0] > 'z' {
		return false
	}

	// Check pattern: letters followed by digits, optionally followed by 's'
	hasDigits := false
	lettersDone := false

	for i, r := range sku {
		if i == 0 {
			continue // Already checked first character
		}

		if r >= '0' && r <= '9' {
			hasDigits = true
			lettersDone = true
		} else if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' {
			if lettersDone && r != 's' {
				return false // Letters after digits (except 's' at end)
			}
			if lettersDone && r == 's' && i != len(sku)-1 {
				return false // 's' not at the end
			}
		} else {
			return false // Invalid character
		}
	}

	return hasDigits // Must have at least one digit
}

// parseInt64 safely converts interface{} to int
func parseInt64(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := fmt.Sscanf(v, "%d", new(int)); err == nil && i == 1 {
			var result int
			fmt.Sscanf(v, "%d", &result)
			return result
		}
	}
	return 0
}

func isValidAzureRegion(region string) bool {
	// List of common Azure regions - in practice this could be more comprehensive
	validRegions := []string{
		"eastus", "eastus2", "westus", "westus2", "westus3", "centralus", "southcentralus",
		"northcentralus", "westcentralus", "canadacentral", "canadaeast", "brazilsouth",
		"northeurope", "westeurope", "francecentral", "germanywestcentral", "norwayeast",
		"switzerlandnorth", "uksouth", "ukwest", "southeastasia", "eastasia", "australiaeast",
		"australiasoutheast", "centralindia", "southindia", "japaneast", "japanwest",
		"koreacentral", "southafricanorth",
	}

	regionLower := strings.ToLower(region)
	for _, validRegion := range validRegions {
		if regionLower == validRegion {
			return true
		}
	}
	return false
}

func isRegionSupportedForArcBox(region string) bool {
	// ArcBox has specific region requirements - this is a simplified check
	// In practice, this would check for specific service availability
	supportedRegions := []string{
		"eastus", "eastus2", "westus2", "centralus", "westeurope", "northeurope",
		"southeastasia", "australiaeast", "japaneast", "uksouth",
	}

	regionLower := strings.ToLower(region)
	for _, supportedRegion := range supportedRegions {
		if regionLower == supportedRegion {
			return true
		}
	}
	return false
}

// SSHKeyRequirementValidator validates SSH key requirement for DevOps/DataOps flavors
type SSHKeyRequirementValidator struct{}

func (v *SSHKeyRequirementValidator) Name() string {
	return "ssh-key-requirement-validation"
}

func (v *SSHKeyRequirementValidator) Description() string {
	return "Validates SSH key requirement for DevOps/DataOps flavors"
}

func (v *SSHKeyRequirementValidator) IsApplicable(ctx *ValidationContext) bool {
	// Only apply to ArcBox DevOps and DataOps flavors (case-insensitive)
	flavor := strings.ToLower(ctx.Flavor)
	return ctx.Solution == "arcbox" && (flavor == "devops" || flavor == "dataops")
}

func (v *SSHKeyRequirementValidator) Validate(ctx *ValidationContext) ValidationResult {
	sshKey := ctx.Parameters["ssh-rsa-public-key"]

	// Map internal flavor names to friendly display names
	friendlyFlavorName := ctx.Flavor
	switch strings.ToLower(ctx.Flavor) {
	case "devops":
		friendlyFlavorName = "DevOps"
	case "dataops":
		friendlyFlavorName = "DataOps"
	}

	if sshKey == "" {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    fmt.Sprintf("%s flavor requires SSH key for Linux VM access", friendlyFlavorName),
			Severity:   "error",
			Suggestion: "Generate an SSH key pair using `ssh-keygen -t rsa -b 4096` and provide the public key with --ssh-rsa-public-key",
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   "SSH key requirement satisfied",
		Severity:  "info",
	}
}

// GitHubUserRequirementValidator validates GitHub user requirement for DevOps flavor
type GitHubUserRequirementValidator struct{}

func (v *GitHubUserRequirementValidator) Name() string {
	return "github-user-requirement-validation"
}

func (v *GitHubUserRequirementValidator) Description() string {
	return "Validates GitHub user requirement for DevOps flavor"
}

func (v *GitHubUserRequirementValidator) IsApplicable(ctx *ValidationContext) bool {
	// Only apply to ArcBox DevOps flavor (case-insensitive)
	flavor := strings.ToLower(ctx.Flavor)
	return ctx.Solution == "arcbox" && flavor == "devops"
}

func (v *GitHubUserRequirementValidator) Validate(ctx *ValidationContext) ValidationResult {
	githubUser := ctx.Parameters["github-user"]

	if githubUser == "" {
		return ValidationResult{
			CheckName:  v.Name(),
			Passed:     false,
			Message:    "DevOps flavor requires GitHub username for custom configurations",
			Severity:   "error",
			Suggestion: "Provide your GitHub username with --github-user where you have forked the jumpstart-apps repository",
		}
	}

	return ValidationResult{
		CheckName: v.Name(),
		Passed:    true,
		Message:   "GitHub user requirement satisfied",
		Severity:  "info",
	}
}
