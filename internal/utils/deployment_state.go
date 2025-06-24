package utils

import (
	"fmt"
	"jumpstartcli/internal/azurecli"
	"strings"
	"time"
)

// ExitFunc type for customizable exit behavior (matches main.go pattern)
type ExitFunc func(int)

// DeploymentStateDetector handles accurate state detection for deployment resources
type DeploymentStateDetector struct {
	azureCLI azurecli.AzureCLI
	exitFunc ExitFunc
	debug    bool
}

// StateValidationResult represents the result of resource state validation
type StateValidationResult struct {
	ResourceName    string
	ReportedState   string
	ValidatedState  string
	IsStateAccurate bool
	ValidationError error
	RetryAttempts   int
	ValidationTime  time.Time
}

// RetryConfig defines retry behavior for transient failures
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	BackoffFactor   float64
}

// DefaultRetryConfig provides sensible defaults for deployment state detection
var DefaultRetryConfig = RetryConfig{
	MaxRetries:      3,
	InitialInterval: 5 * time.Second,
	MaxInterval:     30 * time.Second,
	BackoffFactor:   2.0,
}

// NewDeploymentStateDetector creates a new state detector with the Azure CLI wrapper
func NewDeploymentStateDetector(azureCLI azurecli.AzureCLI, exitFunc ExitFunc, debug bool) *DeploymentStateDetector {
	return &DeploymentStateDetector{
		azureCLI: azureCLI,
		exitFunc: exitFunc,
		debug:    debug,
	}
}

// ValidateResourceState performs enhanced validation of resource state to prevent false failures
func (dsd *DeploymentStateDetector) ValidateResourceState(resourceName, resourceID, reportedState string) StateValidationResult {
	result := StateValidationResult{
		ResourceName:    resourceName,
		ReportedState:   reportedState,
		ValidatedState:  reportedState, // Default to reported state
		IsStateAccurate: true,          // Assume accurate unless proven otherwise
		ValidationTime:  time.Now(),
	}

	// Only validate potentially problematic state transitions
	if !dsd.shouldValidateState(reportedState) {
		return result
	}

	// Perform validation with retry logic
	validatedState, err := dsd.validateWithRetry(resourceID, DefaultRetryConfig)
	if err != nil {
		result.ValidationError = err
		result.IsStateAccurate = false
		if dsd.debug {
			fmt.Printf("[DEBUG] State validation failed for %s: %v\n", resourceName, err)
		}
		return result
	}

	result.ValidatedState = validatedState
	result.IsStateAccurate = (reportedState == validatedState)

	// Log discrepancies for debugging
	if !result.IsStateAccurate && dsd.debug {
		fmt.Printf("[DEBUG] State discrepancy detected for %s: reported=%s, validated=%s\n",
			resourceName, reportedState, validatedState)
	}

	return result
}

// shouldValidateState determines if a state requires additional validation
func (dsd *DeploymentStateDetector) shouldValidateState(state string) bool {
	// Validate states that are commonly misreported or transient
	problematicStates := []string{
		"Failed",
		"Canceled",
		"Unknown",
		"NotFound",
	}

	for _, problematicState := range problematicStates {
		if state == problematicState {
			return true
		}
	}
	return false
}

// validateWithRetry performs state validation with exponential backoff retry logic
func (dsd *DeploymentStateDetector) validateWithRetry(resourceID string, config RetryConfig) (string, error) {
	var lastErr error
	interval := config.InitialInterval

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			if dsd.debug {
				fmt.Printf("[DEBUG] Retrying state validation for %s (attempt %d/%d) after %v\n",
					resourceID, attempt, config.MaxRetries, interval)
			}
			time.Sleep(interval)

			// Calculate next interval with exponential backoff
			interval = time.Duration(float64(interval) * config.BackoffFactor)
			if interval > config.MaxInterval {
				interval = config.MaxInterval
			}
		}

		// Use Azure CLI wrapper to get resource details
		resource, err := dsd.azureCLI.GetResource(resourceID)
		if err != nil {
			lastErr = fmt.Errorf("failed to get resource details: %w", err)
			continue
		}

		// Extract provisioning state from resource properties
		state := dsd.extractProvisioningState(resource)
		if state != "" {
			if dsd.debug && attempt > 0 {
				fmt.Printf("[DEBUG] State validation succeeded for %s after %d retries: %s\n",
					resourceID, attempt, state)
			}
			return state, nil
		}

		lastErr = fmt.Errorf("unable to determine provisioning state from resource properties")
	}

	return "", fmt.Errorf("state validation failed after %d attempts: %w", config.MaxRetries, lastErr)
}

// extractProvisioningState safely extracts provisioning state from resource properties
func (dsd *DeploymentStateDetector) extractProvisioningState(resource *azurecli.ResourceInfo) string {
	if resource == nil || resource.Properties == nil {
		return ""
	}

	// Try different property paths where provisioning state might be stored
	statePaths := []string{
		"provisioningState",
		"properties.provisioningState",
		"status.provisioningState",
	}

	for _, path := range statePaths {
		if state := dsd.getNestedProperty(resource.Properties, path); state != "" {
			return state
		}
	}

	return ""
}

// getNestedProperty safely extracts nested properties from a map
func (dsd *DeploymentStateDetector) getNestedProperty(props map[string]interface{}, path string) string {
	parts := strings.Split(path, ".")
	current := props

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - extract the string value
			if val, ok := current[part]; ok {
				if strVal, ok := val.(string); ok {
					return strVal
				}
			}
		} else {
			// Intermediate part - navigate deeper
			if val, ok := current[part]; ok {
				if mapVal, ok := val.(map[string]interface{}); ok {
					current = mapVal
				} else {
					break
				}
			} else {
				break
			}
		}
	}

	return ""
}

// IsTransientFailure determines if a failure is likely transient and worth retrying
func (dsd *DeploymentStateDetector) IsTransientFailure(err error) bool {
	if err == nil {
		return false
	}

	errorStr := strings.ToLower(err.Error())
	transientErrors := []string{
		"timeout",
		"network",
		"connection",
		"temporary",
		"throttl",
		"rate limit",
		"service unavailable",
		"internal server error",
	}

	for _, transientError := range transientErrors {
		if strings.Contains(errorStr, transientError) {
			return true
		}
	}

	return false
}

// HandleDetectionError handles errors during state detection using the CLI's error handling patterns
func (dsd *DeploymentStateDetector) HandleDetectionError(err error, context string) {
	if err == nil {
		return
	}

	// Log error details in debug mode
	if dsd.debug {
		fmt.Printf("[DEBUG] State detection error in %s: %v\n", context, err)
	}

	// Check if this is a critical error that should cause the CLI to exit
	if dsd.isCriticalError(err) {
		fmt.Printf("%s Critical deployment state detection error: %v\n", ErrorColor("❌"), err)
		if dsd.exitFunc != nil {
			dsd.exitFunc(1)
		}
	} else {
		// Non-critical error - log and continue
		fmt.Printf("%s Warning during state detection: %v\n", WarnColor("⚠️"), err)
	}
}

// isCriticalError determines if an error should cause the CLI to exit
func (dsd *DeploymentStateDetector) isCriticalError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := strings.ToLower(err.Error())
	criticalErrors := []string{
		"authentication failed",
		"access denied",
		"subscription not found",
		"resource group not found",
		"deployment not found",
	}

	for _, criticalError := range criticalErrors {
		if strings.Contains(errorStr, criticalError) {
			return true
		}
	}

	return false
}
