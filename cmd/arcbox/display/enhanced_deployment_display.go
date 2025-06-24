package display

import (
	"fmt"
	"jumpstartcli/cmd/arcbox/models"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"
	"time"
)

// EnhancedDeploymentDisplay provides enhanced deployment monitoring with accurate state detection
type EnhancedDeploymentDisplay struct {
	*DeploymentDisplay // Embed existing display
	stateDetector      *utils.DeploymentStateDetector
	deploymentMonitor  *utils.DeploymentMonitor
	exitFunc           utils.ExitFunc
	debug              bool
	verbose            bool
}

// NewEnhancedDeploymentDisplay creates a new enhanced deployment display with state validation
func NewEnhancedDeploymentDisplay(cli azurecli.AzureCLI, exitFunc utils.ExitFunc, debug, verbose bool) *EnhancedDeploymentDisplay {
	baseDisplay := NewDeploymentDisplay(cli)
	deploymentMonitor := utils.NewDeploymentMonitor(cli, exitFunc, debug, verbose)

	return &EnhancedDeploymentDisplay{
		DeploymentDisplay: baseDisplay,
		deploymentMonitor: deploymentMonitor,
		exitFunc:          exitFunc,
		debug:             debug,
		verbose:           verbose,
	}
}

// WaitForDeploymentAndShowStatusWithValidation provides enhanced deployment monitoring with state validation
func (edd *EnhancedDeploymentDisplay) WaitForDeploymentAndShowStatusWithValidation(resourceGroup, deploymentName string) {
	start := time.Now()

	// Initialize monitoring configuration
	config := utils.DefaultMonitoringConfig
	config.StateValidationEnabled = true

	if edd.verbose {
		fmt.Printf("🚀 Starting enhanced deployment monitoring for %s with state validation...\n", deploymentName)
	} else {
		fmt.Printf("🚀 Starting deployment monitoring for %s...\n", deploymentName)
	}

	var (
		lastStatusTime       = time.Now()
		timeoutWarningShown  = false
		pollCount            = 0
		consecutiveNoChanges = 0
		lastResourceCount    = 0
	)

	for {
		pollCount++
		pollStart := time.Now()

		// Get comprehensive deployment status with validation
		status, err := edd.deploymentMonitor.GetDeploymentStatus(resourceGroup, deploymentName, config)
		if err != nil {
			if edd.debug {
				fmt.Printf("[DEBUG] Error getting deployment status: %v\n", err)
			}
			// Continue monitoring - don't fail on transient errors
			time.Sleep(config.PollInterval)
			continue
		}

		// Convert to models.ResourceStatus for compatibility with existing display
		legacyResources := edd.convertToLegacyResourceStatus(status.Resources)

		// Detect if we have changes worth displaying
		hasChanges := len(status.Resources) != lastResourceCount ||
			time.Since(lastStatusTime) > 30*time.Second

		if hasChanges {
			// Log state validation information in verbose mode
			if edd.verbose && !status.IsAccurate {
				fmt.Printf("⚠️  State validation corrected inaccuracies in deployment status\n")
			}

			// Check for potential false failures
			if edd.deploymentMonitor.DetectPotentialFalseFailure(status) {
				fmt.Printf("🔍 Potential false failure detected - using validated states for accuracy\n")
			}

			// Display progress update using validated states
			edd.displayProgressWithValidation(status, legacyResources)

			lastStatusTime = time.Now()
			consecutiveNoChanges = 0
		} else {
			consecutiveNoChanges++

			// Show minimal activity after several polls with no changes
			if consecutiveNoChanges > 5 {
				fmt.Print(".")
			}
		}

		// Check for completion using enhanced logic
		isComplete, isSuccessful := edd.deploymentMonitor.IsDeploymentComplete(status)
		if isComplete {
			if isSuccessful {
				duration := time.Since(start)
				fmt.Printf("\n✅ Deployment completed successfully in %v\n", edd.formatDuration(duration))

				// Show final status summary in verbose mode
				if edd.verbose {
					edd.deploymentMonitor.LogStatusSummary(status)
				}
			} else {
				fmt.Printf("\n❌ Deployment failed\n")

				// In debug mode, show detailed failure analysis
				if edd.debug {
					edd.analyzeFailureReasons(status)
				}
			}
			break
		}

		// Enhanced timeout warning (only show once at 30 minutes)
		deploymentDuration := time.Since(start)
		if deploymentDuration > 30*time.Minute && !timeoutWarningShown {
			edd.showEnhancedTimeoutWarning(deploymentDuration, status)
			timeoutWarningShown = true
		}

		// Update tracking variables
		lastResourceCount = len(status.Resources)

		// Calculate next poll time with adaptive interval
		pollDuration := time.Since(pollStart)
		adaptiveInterval := edd.calculateAdaptiveInterval(config.PollInterval, pollDuration, hasChanges)

		time.Sleep(adaptiveInterval)
	}

	// Print final metrics if in verbose mode
	if edd.verbose {
		edd.printEnhancedMetrics(pollCount, time.Since(start))
	}
}

// WaitForDeploymentAndShowStatus provides the standard interface method for deployment monitoring
// This delegates to the enhanced validation method
func (edd *EnhancedDeploymentDisplay) WaitForDeploymentAndShowStatus(resourceGroup, deploymentName string) {
	edd.WaitForDeploymentAndShowStatusWithValidation(resourceGroup, deploymentName)
}

// displayProgressWithValidation shows progress using validated resource states
func (edd *EnhancedDeploymentDisplay) displayProgressWithValidation(status *utils.DeploymentStatus, legacyResources []models.ResourceStatus) {
	// Show recent changes with state validation indicators
	for _, resource := range status.Resources {
		if !resource.IsStateAccurate && edd.verbose {
			fmt.Printf("🔄 %s: %s → %s (validated)\n",
				resource.Name, resource.ReportedState, resource.ValidatedState)
		} else {
			// Show normal state transitions
			fmt.Printf("🔄 %s: %s\n", resource.Name, resource.ValidatedState)
		}
	}

	// Enhanced progress summary
	progressPercentage := float64(status.CompletedResources) / float64(status.TotalResources) * 100
	fmt.Printf("🔄 Progress: %d/%d resources completed (%.1f%%), %d in progress, %d failed\n",
		status.CompletedResources, status.TotalResources, progressPercentage,
		status.InProgressResources, status.FailedResources)

	// Show most recent significant change
	if len(status.Resources) > 0 {
		lastResource := status.Resources[len(status.Resources)-1]
		stateToShow := lastResource.ValidatedState
		if !lastResource.IsStateAccurate {
			stateToShow = fmt.Sprintf("%s (validated)", lastResource.ValidatedState)
		}
		fmt.Printf("   Recent: %s %s\n", lastResource.Name, stateToShow)
	}
}

// showEnhancedTimeoutWarning provides constructive timeout guidance
func (edd *EnhancedDeploymentDisplay) showEnhancedTimeoutWarning(duration time.Duration, status *utils.DeploymentStatus) {
	fmt.Printf("\n⚠️  Deployment has been running for %v (longer than expected)\n", edd.formatDuration(duration))

	// Provide constructive guidance based on current state
	if status.InProgressResources > 0 {
		fmt.Printf("   • %d resources are still being configured\n", status.InProgressResources)
		fmt.Printf("   • This is normal for complex deployments with VM extensions\n")
	}

	if status.CompletedResources > 0 {
		progressPercentage := float64(status.CompletedResources) / float64(status.TotalResources) * 100
		fmt.Printf("   • Progress: %.1f%% complete (%d/%d resources)\n",
			progressPercentage, status.CompletedResources, status.TotalResources)

		// Estimate completion time based on current progress
		if progressPercentage > 10 { // Only estimate if we have meaningful progress
			estimatedTotal := duration * 100 / time.Duration(progressPercentage)
			estimatedRemaining := estimatedTotal - duration
			fmt.Printf("   • Estimated completion in ~%v\n", edd.formatDuration(estimatedRemaining))
		}
	}

	fmt.Printf("   • Monitor progress in Azure Portal: https://portal.azure.com\n")
	fmt.Printf("   • Use Ctrl+C to stop monitoring (deployment will continue)\n")
}

// analyzeFailureReasons provides detailed failure analysis in debug mode
func (edd *EnhancedDeploymentDisplay) analyzeFailureReasons(status *utils.DeploymentStatus) {
	fmt.Printf("\n🔍 Failure Analysis:\n")

	// Check for false failure scenarios
	if edd.deploymentMonitor.DetectPotentialFalseFailure(status) {
		fmt.Printf("   • ⚠️  Potential false failure detected\n")
		fmt.Printf("   • State validation found discrepancies between reported and actual states\n")
		fmt.Printf("   • Recommendation: Verify resources in Azure Portal before concluding failure\n")
	}

	// Analyze failed resources
	if status.FailedResources > 0 {
		fmt.Printf("   • %d resources reported as failed:\n", status.FailedResources)
		for _, resource := range status.Resources {
			if resource.ValidatedState == "Failed" {
				fmt.Printf("     - %s (%s)\n", resource.Name, resource.Type)
			}
		}
	}

	// Show state validation summary
	inaccurateStates := 0
	for _, resource := range status.Resources {
		if !resource.IsStateAccurate {
			inaccurateStates++
		}
	}

	if inaccurateStates > 0 {
		fmt.Printf("   • %d resources had inaccurate initial state reporting\n", inaccurateStates)
		fmt.Printf("   • Validated states were used for final determination\n")
	}
}

// calculateAdaptiveInterval adjusts polling interval based on activity and performance
func (edd *EnhancedDeploymentDisplay) calculateAdaptiveInterval(baseInterval, lastPollDuration time.Duration, hasChanges bool) time.Duration {
	// Start with base interval
	interval := baseInterval

	// Reduce interval if there are active changes
	if hasChanges {
		interval = interval / 2
		if interval < 2*time.Second {
			interval = 2 * time.Second
		}
	} else {
		// Increase interval gradually if no changes
		interval = interval * 3 / 2
		if interval > 15*time.Second {
			interval = 15 * time.Second
		}
	}

	// Adjust for API performance - if last poll was slow, wait a bit longer
	if lastPollDuration > 5*time.Second {
		interval += lastPollDuration / 2
	}

	return interval
}

// convertToLegacyResourceStatus converts new ResourceStatus to legacy models.ResourceStatus
func (edd *EnhancedDeploymentDisplay) convertToLegacyResourceStatus(resources []utils.ResourceStatus) []models.ResourceStatus {
	legacy := make([]models.ResourceStatus, len(resources))
	for i, resource := range resources {
		legacy[i] = models.ResourceStatus{
			Name:  resource.Name,
			Type:  resource.Type,
			State: resource.ValidatedState, // Use validated state for accuracy
		}
	}
	return legacy
}

// formatDuration formats duration in a human-readable way
func (edd *EnhancedDeploymentDisplay) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), d.Seconds()-60*d.Minutes())
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) - 60*hours
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// printEnhancedMetrics shows enhanced performance metrics
func (edd *EnhancedDeploymentDisplay) printEnhancedMetrics(pollCount int, totalDuration time.Duration) {
	fmt.Printf("\n📊 Enhanced Monitoring Metrics:\n")
	fmt.Printf("   • Total polling cycles: %d\n", pollCount)
	fmt.Printf("   • Average poll interval: %v\n", totalDuration/time.Duration(pollCount))
	fmt.Printf("   • State validation: Enabled\n")
	fmt.Printf("   • Monitoring duration: %v\n", edd.formatDuration(totalDuration))
}
