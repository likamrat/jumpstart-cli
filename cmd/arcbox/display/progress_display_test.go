package display

import (
	"testing"
	"time"

	"jumpstartcli/cmd/arcbox/models"

	"github.com/stretchr/testify/assert"
)

func TestProgressDisplay_NewProgressDisplay(t *testing.T) {
	pd := NewProgressDisplay()

	assert.NotNil(t, pd)
	assert.Equal(t, PhaseInitialization, pd.CurrentPhase)
	assert.Equal(t, 5, pd.MaxRecentChanges)
	assert.NotNil(t, pd.RecentChanges)
	assert.Equal(t, 0, len(pd.RecentChanges))
}

func TestProgressDisplay_UpdateProgress(t *testing.T) {
	pd := NewProgressDisplay()

	resources := []models.ResourceStatus{
		{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", State: "Succeeded"},
		{Name: "storage1", Type: "Microsoft.Storage/storageAccounts", State: "Creating"},
		{Name: "network1", Type: "Microsoft.Network/virtualNetworks", State: "Failed"},
	}

	pd.UpdateProgress(resources)

	assert.Equal(t, 3, pd.TotalResources)
	assert.Equal(t, 1, pd.CompletedResources)
	assert.Equal(t, 1, pd.InProgressResources)
	assert.Equal(t, 1, pd.FailedResources)
	assert.Equal(t, float64(1)/float64(3)*100, pd.ProgressPercentage)
}

func TestProgressDisplay_PhaseDetection(t *testing.T) {
	tests := []struct {
		name      string
		resources []models.ResourceStatus
		expected  DeploymentPhase
	}{
		{
			name:      "empty_resources_initialization",
			resources: []models.ResourceStatus{},
			expected:  PhaseInitialization,
		},
		{
			name: "infrastructure_phase",
			resources: []models.ResourceStatus{
				{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", State: "Creating"},
				{Name: "storage1", Type: "Microsoft.Storage/storageAccounts", State: "Creating"},
			},
			expected: PhaseInfrastructure,
		},
		{
			name: "configuration_phase",
			resources: []models.ResourceStatus{
				{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", State: "Succeeded"},
				{Name: "vm1-ext", Type: "Microsoft.Compute/virtualMachines/extensions", State: "Creating"},
			},
			expected: PhaseConfiguration,
		},
		{
			name: "completion_phase",
			resources: []models.ResourceStatus{
				{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", State: "Succeeded"},
				{Name: "storage1", Type: "Microsoft.Storage/storageAccounts", State: "Succeeded"},
				{Name: "network1", Type: "Microsoft.Network/virtualNetworks", State: "Succeeded"},
				{Name: "remaining", Type: "Microsoft.Other/resource", State: "Creating"},
			},
			expected: PhaseCompletion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pd := NewProgressDisplay()
			pd.UpdateProgress(tt.resources)
			assert.Equal(t, tt.expected, pd.CurrentPhase)
		})
	}
}

func TestProgressDisplay_GenerateProgressBar(t *testing.T) {
	pd := NewProgressDisplay()
	pd.ProgressPercentage = 33.3

	progressBar := pd.GenerateProgressBar(30)
	assert.Contains(t, progressBar, "33.3%")
	assert.Contains(t, progressBar, "[")
	assert.Contains(t, progressBar, "]")
}

func TestProgressDisplay_AddRecentChange(t *testing.T) {
	pd := NewProgressDisplay()
	pd.MaxRecentChanges = 2

	change1 := ResourceChange{
		Name:      "vm1",
		FromState: "Creating",
		ToState:   "Succeeded",
		Icon:      "✅",
	}

	change2 := ResourceChange{
		Name:      "storage1",
		FromState: "Creating",
		ToState:   "Running",
		Icon:      "⌛",
	}

	change3 := ResourceChange{
		Name:      "network1",
		FromState: "Creating",
		ToState:   "Failed",
		Icon:      "❌",
	}

	pd.AddRecentChange(change1)
	assert.Equal(t, 1, len(pd.RecentChanges))

	pd.AddRecentChange(change2)
	assert.Equal(t, 2, len(pd.RecentChanges))

	pd.AddRecentChange(change3)
	assert.Equal(t, 2, len(pd.RecentChanges))           // Should still be 2 due to limit
	assert.Contains(t, pd.RecentChanges[1], "network1") // Latest change should be included
}

func TestProgressDisplay_GetPhaseInfo(t *testing.T) {
	pd := NewProgressDisplay()

	phases := []DeploymentPhase{
		PhaseInitialization,
		PhaseInfrastructure,
		PhaseConfiguration,
		PhaseCompletion,
		PhaseFinalized,
	}

	for _, phase := range phases {
		info := pd.GetPhaseInfo(phase)
		assert.NotEmpty(t, info.Name)
		assert.NotEmpty(t, info.Description)
		assert.NotEmpty(t, info.Icon)
		assert.NotNil(t, info.Color)
	}
}

func TestDeploymentDisplay_DetectResourceChanges(t *testing.T) {
	dd := NewDeploymentDisplay(nil)

	// Previous state
	previousStates := map[string]string{
		"vm1":      "Creating",
		"storage1": "Succeeded",
	}

	// Current resources
	currentResources := []models.ResourceStatus{
		{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", State: "Succeeded"},      // Changed
		{Name: "storage1", Type: "Microsoft.Storage/storageAccounts", State: "Succeeded"}, // No change
		{Name: "network1", Type: "Microsoft.Network/virtualNetworks", State: "Creating"},  // New
	}

	hasChanges, changes := dd.detectResourceChanges(currentResources, previousStates)

	assert.True(t, hasChanges)
	assert.Equal(t, 2, len(changes)) // vm1 changed, network1 is new

	// Check vm1 change
	vm1Change := changes[0]
	assert.Equal(t, "vm1", vm1Change.Name)
	assert.Equal(t, "Creating", vm1Change.FromState)
	assert.Equal(t, "Succeeded", vm1Change.ToState)

	// Check network1 new resource
	network1Change := changes[1]
	assert.Equal(t, "network1", network1Change.Name)
	assert.Equal(t, "New", network1Change.FromState)
	assert.Equal(t, "Creating", network1Change.ToState)
}

func TestDeploymentDisplay_GetResourceStateIcon(t *testing.T) {
	dd := NewDeploymentDisplay(nil)

	tests := []struct {
		state    string
		expected string
	}{
		{"Succeeded", "✅"},
		{"Failed", "❌"},
		{"Canceled", "❌"},
		{"Creating", "⌛"},
		{"Running", "⌛"},
		{"InProgress", "⌛"},
		{"Updating", "🔄"},
		{"Deleting", "🗑️"},
		{"Unknown", "❓"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			icon := dd.getResourceStateIcon(tt.state)
			assert.Equal(t, tt.expected, icon)
		})
	}
}

func TestProgressDisplay_EstimateTimeRemaining(t *testing.T) {
	pd := NewProgressDisplay()
	pd.StartTime = time.Now().Add(-10 * time.Minute) // Started 10 minutes ago
	pd.ProgressPercentage = 25.0                     // 25% complete

	pd.estimateTimeRemaining()

	// With 25% done in 10 minutes, estimated total is 40 minutes, so 30 minutes remaining
	expectedRemaining := 30 * time.Minute
	tolerance := 1 * time.Minute

	assert.True(t, pd.EstimatedTimeLeft >= expectedRemaining-tolerance &&
		pd.EstimatedTimeLeft <= expectedRemaining+tolerance,
		"Expected around %v, got %v", expectedRemaining, pd.EstimatedTimeLeft)
}

func TestProgressDisplay_EstimateTimeRemaining_Complete(t *testing.T) {
	pd := NewProgressDisplay()
	pd.ProgressPercentage = 100.0

	pd.estimateTimeRemaining()

	assert.Equal(t, time.Duration(0), pd.EstimatedTimeLeft)
}
