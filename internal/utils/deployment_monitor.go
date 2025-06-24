package utils

import (
	"fmt"
	"jumpstartcli/internal/azurecli"
	"time"
)

// DeploymentMonitor provides enhanced deployment monitoring capabilities for all deployment types
type DeploymentMonitor struct {
	azureCLI      azurecli.AzureCLI
	stateDetector *DeploymentStateDetector
	exitFunc      ExitFunc
	debug         bool
	verbose       bool
}

// DeploymentStatus represents the comprehensive status of a deployment
type DeploymentStatus struct {
	DeploymentName      string
	ResourceGroup       string
	State               string
	ValidatedState      string
	Resources           []ResourceStatus
	StartTime           time.Time
	LastUpdateTime      time.Time
	TotalResources      int
	CompletedResources  int
	FailedResources     int
	InProgressResources int
	ValidationResults   []StateValidationResult
	IsAccurate          bool
}

// ResourceStatus represents the status of an individual resource with validation
type ResourceStatus struct {
	Name            string
	Type            string
	ID              string
	ReportedState   string
	ValidatedState  string
	IsStateAccurate bool
	LastValidated   time.Time
}

// MonitoringConfig defines configuration for deployment monitoring
type MonitoringConfig struct {
	PollInterval           time.Duration
	StateValidationEnabled bool
	RetryConfig            RetryConfig
	TimeoutWarningMinutes  int
	MaxPollingDuration     time.Duration
}

// DefaultMonitoringConfig provides sensible defaults for deployment monitoring
var DefaultMonitoringConfig = MonitoringConfig{
	PollInterval:           5 * time.Second,
	StateValidationEnabled: true,
	RetryConfig:            DefaultRetryConfig,
	TimeoutWarningMinutes:  30,
	MaxPollingDuration:     120 * time.Minute, // 2 hours max
}

// NewDeploymentMonitor creates a new deployment monitor with enhanced state detection
func NewDeploymentMonitor(azureCLI azurecli.AzureCLI, exitFunc ExitFunc, debug, verbose bool) *DeploymentMonitor {
	stateDetector := NewDeploymentStateDetector(azureCLI, exitFunc, debug)

	return &DeploymentMonitor{
		azureCLI:      azureCLI,
		stateDetector: stateDetector,
		exitFunc:      exitFunc,
		debug:         debug,
		verbose:       verbose,
	}
}

// GetDeploymentStatus retrieves and validates comprehensive deployment status
func (dm *DeploymentMonitor) GetDeploymentStatus(resourceGroup, deploymentName string, config MonitoringConfig) (*DeploymentStatus, error) {
	status := &DeploymentStatus{
		DeploymentName:    deploymentName,
		ResourceGroup:     resourceGroup,
		StartTime:         time.Now(),
		LastUpdateTime:    time.Now(),
		ValidationResults: make([]StateValidationResult, 0),
		IsAccurate:        true,
	}

	// Get deployment information using Azure CLI wrapper
	deployment, err := dm.azureCLI.GetDeployment(resourceGroup, deploymentName)
	if err != nil {
		dm.stateDetector.HandleDetectionError(err, "GetDeployment")
		return nil, fmt.Errorf("failed to get deployment status: %w", err)
	}

	if deployment != nil {
		status.State = deployment.ProvisioningState
		status.ValidatedState = deployment.ProvisioningState // Will be updated if validation is enabled
	}

	// Get resource list using Azure CLI wrapper
	resources, err := dm.azureCLI.ListResources(resourceGroup)
	if err != nil {
		dm.stateDetector.HandleDetectionError(err, "ListResources")
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	// Process and validate resource states
	status.Resources = make([]ResourceStatus, 0, len(resources))
	for _, resource := range resources {
		resourceStatus := dm.processResource(resource, config)
		status.Resources = append(status.Resources, resourceStatus)

		// Update counters
		status.TotalResources++
		switch resourceStatus.ValidatedState {
		case "Succeeded":
			status.CompletedResources++
		case "Failed", "Canceled":
			status.FailedResources++
		case "Creating", "Running", "Updating", "InProgress":
			status.InProgressResources++
		}

		// Track validation accuracy
		if !resourceStatus.IsStateAccurate {
			status.IsAccurate = false
		}
	}

	// Validate deployment state if enabled and there are concerning resource states
	if config.StateValidationEnabled && dm.shouldValidateDeploymentState(status) {
		dm.validateDeploymentState(status, config)
	}

	return status, nil
}

// processResource processes and optionally validates an individual resource
func (dm *DeploymentMonitor) processResource(resource azurecli.ResourceInfo, config MonitoringConfig) ResourceStatus {
	// Extract provisioning state from resource
	provisioningState := dm.extractResourceState(resource)

	resourceStatus := ResourceStatus{
		Name:            resource.Name,
		Type:            resource.Type,
		ID:              resource.ID,
		ReportedState:   provisioningState,
		ValidatedState:  provisioningState,
		IsStateAccurate: true,
		LastValidated:   time.Now(),
	}

	// Perform state validation if enabled and necessary
	if config.StateValidationEnabled {
		validationResult := dm.stateDetector.ValidateResourceState(resource.Name, resource.ID, provisioningState)

		resourceStatus.ValidatedState = validationResult.ValidatedState
		resourceStatus.IsStateAccurate = validationResult.IsStateAccurate

		// Log validation discrepancies in verbose mode
		if dm.verbose && !validationResult.IsStateAccurate {
			fmt.Printf("%s State validation corrected %s: %s → %s\n",
				InfoColor("ℹ️"), resource.Name, provisioningState, validationResult.ValidatedState)
		}
	}

	return resourceStatus
}

// extractResourceState safely extracts the provisioning state from a resource
func (dm *DeploymentMonitor) extractResourceState(resource azurecli.ResourceInfo) string {
	if resource.Properties == nil {
		return "Unknown"
	}

	// Try to extract provisioning state
	if state, ok := resource.Properties["provisioningState"]; ok {
		if stateStr, ok := state.(string); ok {
			return stateStr
		}
	}

	return "Unknown"
}

// shouldValidateDeploymentState determines if deployment-level validation is needed
func (dm *DeploymentMonitor) shouldValidateDeploymentState(status *DeploymentStatus) bool {
	// Validate if deployment is reported as failed but resources suggest otherwise
	if status.State == "Failed" && status.FailedResources == 0 {
		return true
	}

	// Validate if there are discrepancies in resource states
	if !status.IsAccurate {
		return true
	}

	// Validate if deployment state doesn't match resource completion
	if status.State == "Succeeded" && status.CompletedResources < status.TotalResources {
		return true
	}

	return false
}

// validateDeploymentState performs deployment-level state validation
func (dm *DeploymentMonitor) validateDeploymentState(status *DeploymentStatus, config MonitoringConfig) {
	if dm.debug {
		fmt.Printf("[DEBUG] Validating deployment state for %s\n", status.DeploymentName)
	}

	// Re-fetch deployment with retry logic
	validatedState, err := dm.stateDetector.validateWithRetry(
		fmt.Sprintf("/subscriptions/current/resourceGroups/%s/deployments/%s",
			status.ResourceGroup, status.DeploymentName),
		config.RetryConfig)

	if err != nil {
		dm.stateDetector.HandleDetectionError(err, "deployment state validation")
		return
	}

	// Update deployment state if validation found a discrepancy
	if validatedState != status.State {
		if dm.verbose {
			fmt.Printf("%s Deployment state validation corrected %s: %s → %s\n",
				InfoColor("ℹ️"), status.DeploymentName, status.State, validatedState)
		}
		status.ValidatedState = validatedState
		status.IsAccurate = false
	}
}

// IsDeploymentComplete determines if a deployment has completed (successfully or with failure)
func (dm *DeploymentMonitor) IsDeploymentComplete(status *DeploymentStatus) (bool, bool) {
	// Use validated state for accuracy
	deploymentState := status.ValidatedState
	if deploymentState == "" {
		deploymentState = status.State
	}

	// Check deployment-level completion
	switch deploymentState {
	case "Succeeded":
		// Verify all resources are also completed
		allResourcesComplete := (status.InProgressResources == 0)
		return true, allResourcesComplete && (status.FailedResources == 0)
	case "Failed", "Canceled":
		return true, false
	default:
		return false, false
	}
}

// DetectPotentialFalseFailure analyzes a deployment to detect potential false failure scenarios
func (dm *DeploymentMonitor) DetectPotentialFalseFailure(status *DeploymentStatus) bool {
	// Classic false failure: deployment reports failed but all resources succeeded
	if (status.State == "Failed" || status.ValidatedState == "Failed") &&
		status.FailedResources == 0 && status.CompletedResources > 0 {
		return true
	}

	// Resource-level false failure: some resources report failed but validation shows success
	falseFailureCount := 0
	for _, resource := range status.Resources {
		if resource.ReportedState == "Failed" && resource.ValidatedState == "Succeeded" {
			falseFailureCount++
		}
	}

	// If more than one resource had false failures, likely a systematic issue
	return falseFailureCount > 0
}

// LogStatusSummary provides a comprehensive status summary for verbose output
func (dm *DeploymentMonitor) LogStatusSummary(status *DeploymentStatus) {
	if !dm.verbose {
		return
	}

	fmt.Printf("\n%s Deployment Status Summary:\n", InfoColor("📊"))
	fmt.Printf("   • Deployment: %s (State: %s)\n", status.DeploymentName, status.ValidatedState)
	fmt.Printf("   • Resources: %d total, %d completed, %d failed, %d in progress\n",
		status.TotalResources, status.CompletedResources, status.FailedResources, status.InProgressResources)
	fmt.Printf("   • State Accuracy: %t\n", status.IsAccurate)

	if dm.DetectPotentialFalseFailure(status) {
		fmt.Printf("   %s Potential false failure detected - validation recommended\n", WarnColor("⚠️"))
	}
}
