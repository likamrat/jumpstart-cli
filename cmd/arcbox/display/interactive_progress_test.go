package display

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"jumpstartcli/cmd/arcbox/models"

	"github.com/stretchr/testify/assert"
)

func TestInteractiveProgressIndicator_NewInteractiveProgressIndicator(t *testing.T) {
	ipi := NewInteractiveProgressIndicator(true, 1)

	assert.NotNil(t, ipi)
	assert.True(t, ipi.IsInteractive)
	assert.Equal(t, 1, ipi.VerbosityLevel)
	assert.NotNil(t, ipi.SmartSpinner)
	assert.NotNil(t, ipi.ResourceCounters)
	assert.NotNil(t, ipi.TimeIndicators)
	assert.NotNil(t, ipi.ActivityFeed)
	assert.Equal(t, 5, len(ipi.PhaseProgressBars))
}

func TestSmartSpinner_ContextualFrames(t *testing.T) {
	spinner := NewSmartSpinner()
	spinner.Start()

	tests := []struct {
		context        SpinnerContext
		expectedFrames int
		minUpdateTime  time.Duration
	}{
		{SpinnerInitializing, 4, 200 * time.Millisecond},
		{SpinnerProvisioning, 4, 300 * time.Millisecond},
		{SpinnerConfiguring, 4, 250 * time.Millisecond},
		{SpinnerValidating, 4, 150 * time.Millisecond},
		{SpinnerCompleting, 4, 400 * time.Millisecond},
		{SpinnerWaiting, 4, 500 * time.Millisecond},
		{SpinnerIdle, 10, 80 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.context.String(), func(t *testing.T) {
			frame := spinner.UpdateSpinner(tt.context)
			assert.NotEmpty(t, frame)
			assert.Equal(t, tt.expectedFrames, len(spinner.CurrentFrames))
			assert.Equal(t, tt.minUpdateTime, spinner.UpdateInterval)
		})
	}
}

func TestPhaseProgressBar_Render(t *testing.T) {
	bar := NewPhaseProgressBar(PhaseInfrastructure)
	bar.UpdateProgress(75.0, "Deploying VMs")

	rendered := bar.Render()
	assert.Contains(t, rendered, "75.0%")
	assert.Contains(t, rendered, "Deploying VMs")
	assert.Contains(t, rendered, "[")
	assert.Contains(t, rendered, "]")
}

func TestLiveResourceCounters_UpdateAndRender(t *testing.T) {
	counters := NewLiveResourceCounters()

	resources := []models.ResourceStatus{
		{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", State: "Succeeded"},
		{Name: "vm2", Type: "Microsoft.Compute/virtualMachines", State: "Creating"},
		{Name: "storage1", Type: "Microsoft.Storage/storageAccounts", State: "Failed"},
		{Name: "network1", Type: "Microsoft.Network/virtualNetworks", State: "InProgress"},
	}

	counters.UpdateCounters(resources)

	assert.Equal(t, 4, counters.Total)
	assert.Equal(t, 1, counters.Succeeded)
	assert.Equal(t, 1, counters.Failed)
	assert.Equal(t, 1, counters.Creating)
	assert.Equal(t, 1, counters.InProgress)

	rendered := counters.RenderCounters(true)
	assert.Contains(t, rendered, "Total: 4")
	assert.Contains(t, rendered, "✅ Succeeded: 1")
	assert.Contains(t, rendered, "❌ Failed: 1")
	assert.Contains(t, rendered, "⌛ In Progress: 2")
}

func TestLiveResourceCounters_Trending(t *testing.T) {
	counters := NewLiveResourceCounters()

	// First update
	resources1 := []models.ResourceStatus{
		{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", State: "Creating"},
	}
	counters.UpdateCounters(resources1)

	// Second update with more succeeded resources
	resources2 := []models.ResourceStatus{
		{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", State: "Succeeded"},
		{Name: "vm2", Type: "Microsoft.Compute/virtualMachines", State: "Succeeded"},
	}
	counters.UpdateCounters(resources2)

	// Check trends
	assert.Equal(t, TrendIncreasing, counters.Trending["succeeded"])
	assert.Equal(t, TrendDecreasing, counters.Trending["creating"])
}

func TestDetailedTimeIndicators_UpdateAndRender(t *testing.T) {
	startTime := time.Now().Add(-10 * time.Minute)
	indicators := NewDetailedTimeIndicators()
	indicators.DeploymentStart = startTime
	indicators.CurrentPhaseStart = startTime.Add(5 * time.Minute)

	// Force the last update to be old enough to allow updates
	indicators.LastUpdate = startTime

	indicators.UpdateTimeIndicators(PhaseInfrastructure, 50.0)

	lines := indicators.RenderTimeIndicators(1)
	assert.GreaterOrEqual(t, len(lines), 2)
	assert.Contains(t, lines[0], "Elapsed:")

	// Check that estimated remaining time was calculated
	assert.Greater(t, indicators.EstimatedRemaining, time.Duration(0), "Estimated remaining time should be > 0 with 50% progress")

	// Only test for additional lines if they exist
	if len(lines) > 1 {
		// Look for either "Estimated remaining:" or "Current phase:" in the second line
		assert.True(t, strings.Contains(lines[1], "Estimated remaining:") || strings.Contains(lines[1], "Current phase:"),
			"Second line should contain either estimated remaining or current phase, got: %s", lines[1])
	}
}

func TestScrollingActivityFeed_AddAndRender(t *testing.T) {
	feed := NewScrollingActivityFeed(3)

	activities := []ActivityEntry{
		{
			Timestamp: time.Now(),
			Message:   "VM created",
			Type:      ActivityResourceChange,
			Icon:      "✅",
		},
		{
			Timestamp: time.Now(),
			Message:   "Storage provisioning",
			Type:      ActivityResourceChange,
			Icon:      "⌛",
		},
		{
			Timestamp: time.Now(),
			Message:   "Network configured",
			Type:      ActivityResourceChange,
			Icon:      "✅",
		},
		{
			Timestamp: time.Now(),
			Message:   "Phase completed",
			Type:      ActivityPhaseChange,
			Icon:      "🎉",
		},
	}

	for _, activity := range activities {
		feed.AddActivity(activity)
	}

	// Should only keep the last 3 activities due to MaxEntries limit
	assert.Equal(t, 3, len(feed.Activities))

	lines := feed.RenderActivityFeed(5)
	assert.Equal(t, 3, len(lines))
	assert.Contains(t, lines[2], "Phase completed")
}

func TestProgressDisplay_InteractiveMode(t *testing.T) {
	pd := NewProgressDisplay()

	// Test setting interactive mode
	pd.SetInteractiveMode(true)
	assert.True(t, pd.IsInteractiveMode)
	assert.True(t, pd.InteractiveIndicators.IsInteractive)

	// Test setting verbosity level
	pd.SetVerbosityLevel(2)
	assert.Equal(t, 2, pd.VerbosityLevel)
	assert.Equal(t, 2, pd.InteractiveIndicators.VerbosityLevel)
}

func TestProgressDisplay_AccessibleStatusSummary(t *testing.T) {
	pd := NewProgressDisplay()
	pd.TotalResources = 10
	pd.CompletedResources = 6
	pd.InProgressResources = 3
	pd.FailedResources = 1
	pd.ProgressPercentage = 60.0
	pd.EstimatedTimeLeft = 15 * time.Minute

	summary := pd.GetAccessibleStatusSummary()

	assert.Contains(t, summary, "Progress: 60 percent complete")
	assert.Contains(t, summary, "Resources: 10 total, 6 completed, 3 in progress, 1 failed")
	assert.Contains(t, summary, "Estimated time remaining: 15m0s")
}

func TestProgressDisplay_ContextualTips(t *testing.T) {
	pd := NewProgressDisplay()
	pd.StartTime = time.Now().Add(-25 * time.Minute) // Started 25 minutes ago

	tests := []struct {
		phase               DeploymentPhase
		inProgressResources int
		progressPercentage  float64
		expectedTipContains string
	}{
		{PhaseInitialization, 0, 10, "Initial resource creation"},
		{PhaseInfrastructure, 5, 30, "VM and storage provisioning"},
		{PhaseConfiguration, 2, 70, "VM extensions and software"},
		{PhaseInfrastructure, 0, 50, "resources seem stuck"},
	}

	for _, tt := range tests {
		pd.CurrentPhase = tt.phase
		pd.InProgressResources = tt.inProgressResources
		pd.ProgressPercentage = tt.progressPercentage

		tip := pd.getContextualTip()
		if tt.expectedTipContains != "" {
			assert.Contains(t, tip, tt.expectedTipContains)
		}
	}
}

func TestInteractiveProgressIndicator_TerminalResize(t *testing.T) {
	ipi := NewInteractiveProgressIndicator(true, 1)

	// Simulate terminal resize
	ipi.TerminalSize.Width = 120
	ipi.TerminalSize.Height = 30
	ipi.HandleTerminalResize()

	// Check that progress bars were resized appropriately
	for _, phaseBar := range ipi.PhaseProgressBars {
		assert.LessOrEqual(t, phaseBar.Width, 80)    // Should be capped at 80
		assert.GreaterOrEqual(t, phaseBar.Width, 20) // Should be at least 20
	}

	// Check activity feed display lines were adjusted
	assert.Equal(t, 6, ipi.ActivityFeed.DisplayLines)
}

func TestDeploymentDisplay_IsInteractiveEnvironment(t *testing.T) {
	dd := NewDeploymentDisplay(nil)

	// Test default case (should be interactive)
	assert.True(t, dd.isInteractiveEnvironment())

	// Test with CI environment variable set
	t.Setenv("CI", "true")
	assert.False(t, dd.isInteractiveEnvironment())

	// Test with explicit non-interactive setting
	t.Setenv("CI", "")
	t.Setenv("ARCBOX_NON_INTERACTIVE", "true")
	assert.False(t, dd.isInteractiveEnvironment())
}

func TestProgressDisplay_EnhancedProgressBar(t *testing.T) {
	pd := NewProgressDisplay()

	tests := []struct {
		percentage  float64
		width       int
		expectColor bool
	}{
		{25.0, 30, true},
		{50.0, 40, true},
		{75.0, 50, true},
		{95.0, 60, true},
	}

	for _, tt := range tests {
		pd.ProgressPercentage = tt.percentage
		bar := pd.generateEnhancedProgressBar(tt.width)

		assert.Contains(t, bar, fmt.Sprintf("%.1f%%", tt.percentage))
		assert.Contains(t, bar, "[")
		assert.Contains(t, bar, "]")
		// Visual progress should be proportional
		filledCount := int(tt.percentage / 100 * float64(tt.width))
		if filledCount > 0 {
			assert.Contains(t, bar, "█") // Should contain filled characters
		}
	}
}
