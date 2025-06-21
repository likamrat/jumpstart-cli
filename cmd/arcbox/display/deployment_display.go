// deployment_display.go - Display formatting for ArcBox deployment information
package display

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"jumpstartcli/cmd/arcbox/models"
	"jumpstartcli/internal/azurecli"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

// DeploymentPhase represents different phases of ArcBox deployment
type DeploymentPhase int

const (
	PhaseInitialization DeploymentPhase = iota
	PhaseInfrastructure
	PhaseConfiguration
	PhaseCompletion
	PhaseFinalized
)

// ProgressDisplay manages enhanced progress display functionality
type ProgressDisplay struct {
	StartTime           time.Time
	LastUpdateTime      time.Time
	TotalResources      int
	CompletedResources  int
	FailedResources     int
	InProgressResources int
	CurrentPhase        DeploymentPhase
	PhaseStartTime      time.Time
	EstimatedTimeLeft   time.Duration
	ProgressPercentage  float64
	DisplayBuffer       []string
	LastDisplayHeight   int
	RecentChanges       []string
	MaxRecentChanges    int
	IsInteractiveMode   bool
	VerbosityLevel      int
}

// PhaseInfo contains information about deployment phases
type PhaseInfo struct {
	Name         string
	Description  string
	Icon         string
	Color        func(a ...interface{}) string
	ExpectedTime time.Duration
}

// ResourceChange represents a change in resource state
type ResourceChange struct {
	Name      string
	Type      string
	FromState string
	ToState   string
	Timestamp time.Time
	Icon      string
}

// DeploymentDisplay encapsulates dependencies for deployment display operations
type DeploymentDisplay struct {
	azureCLI         azurecli.AzureCLI
	progressDisplay  *ProgressDisplay
	cache            *IntelligentCache
	metricsCollector *PerformanceMetricsCollector
}

// NewDeploymentDisplay creates a new DeploymentDisplay instance
func NewDeploymentDisplay(cli azurecli.AzureCLI) *DeploymentDisplay {
	return &DeploymentDisplay{
		azureCLI:         cli,
		progressDisplay:  NewProgressDisplay(),
		cache:            NewIntelligentCache(),
		metricsCollector: NewPerformanceMetricsCollector(MetricsDetailed, false),
	}
}

// NewDeploymentDisplayWithMetrics creates a new DeploymentDisplay instance with configurable metrics
func NewDeploymentDisplayWithMetrics(cli azurecli.AzureCLI, metricsLevel MetricsCollectionLevel, verbose bool) *DeploymentDisplay {
	return &DeploymentDisplay{
		azureCLI:         cli,
		progressDisplay:  NewProgressDisplay(),
		cache:            NewIntelligentCache(),
		metricsCollector: NewPerformanceMetricsCollector(metricsLevel, verbose),
	}
}

// NewProgressDisplay creates a new progress display manager
func NewProgressDisplay() *ProgressDisplay {
	return &ProgressDisplay{
		StartTime:         time.Now(),
		LastUpdateTime:    time.Now(),
		CurrentPhase:      PhaseInitialization,
		PhaseStartTime:    time.Now(),
		MaxRecentChanges:  5,
		IsInteractiveMode: false,
		VerbosityLevel:    1,
		DisplayBuffer:     make([]string, 0),
		RecentChanges:     make([]string, 0),
	}
}

// SetInteractiveMode sets whether progress display should use interactive mode
func (pd *ProgressDisplay) SetInteractiveMode(interactive bool) {
	pd.IsInteractiveMode = interactive
}

// UpdatePhase updates the current deployment phase
func (pd *ProgressDisplay) UpdatePhase(phase DeploymentPhase) {
	pd.CurrentPhase = phase
	pd.PhaseStartTime = time.Now()
}

// AddAccessibilityAnnouncement adds an accessibility announcement (placeholder)
func (pd *ProgressDisplay) AddAccessibilityAnnouncement(message string) {
	// Implementation would be added for accessibility support
}

// GetAccessibleStatusSummary returns a summary for accessibility (placeholder)
func (pd *ProgressDisplay) GetAccessibleStatusSummary() string {
	return fmt.Sprintf("Deployment phase: %d, Total resources: %d, Completed: %d, Failed: %d",
		pd.CurrentPhase, pd.TotalResources, pd.CompletedResources, pd.FailedResources)
}

// UpdateProgress updates progress display with current resource state
func (pd *ProgressDisplay) UpdateProgress(resources []models.ResourceStatus) {
	pd.TotalResources = len(resources)
	pd.CompletedResources = 0
	pd.FailedResources = 0
	pd.InProgressResources = 0

	for _, res := range resources {
		switch res.State {
		case "Succeeded":
			pd.CompletedResources++
		case "Failed", "Canceled":
			pd.FailedResources++
		case "Creating", "Running", "Updating", "InProgress":
			pd.InProgressResources++
		}
	}

	if pd.TotalResources > 0 {
		pd.ProgressPercentage = float64(pd.CompletedResources) / float64(pd.TotalResources) * 100
	}

	pd.LastUpdateTime = time.Now()
}

// AddRecentChange adds a resource change to recent changes list
func (pd *ProgressDisplay) AddRecentChange(change ResourceChange) {
	pd.RecentChanges = append(pd.RecentChanges, change.Name+" "+change.FromState+" → "+change.ToState)

	// Keep only the most recent changes
	if len(pd.RecentChanges) > pd.MaxRecentChanges {
		pd.RecentChanges = pd.RecentChanges[1:]
	}
}

// DisplayProgressUpdate displays current progress (simplified implementation)
func (pd *ProgressDisplay) DisplayProgressUpdate() {
	fmt.Printf("\r🔄 Progress: %d/%d resources completed (%.1f%%), %d in progress, %d failed",
		pd.CompletedResources, pd.TotalResources, pd.ProgressPercentage,
		pd.InProgressResources, pd.FailedResources)

	if len(pd.RecentChanges) > 0 {
		fmt.Printf("\n   Recent: %s", pd.RecentChanges[len(pd.RecentChanges)-1])
	}

	fmt.Println()
}

// PrintResourceList displays the deployment resource list with status icons
func (dd *DeploymentDisplay) PrintResourceList(resources []models.ResourceStatus) {
	for _, res := range resources {
		// Hide DevTestLab schedules (e.g. auto-shutdown)
		if res.Type == "Microsoft.DevTestLab/schedules" {
			continue
		}
		icon := "❓"
		// Set emoji for each state
		switch res.State {
		case "Succeeded":
			icon = "✅"
		case "Failed":
			icon = "❌"
		case "Running", "Creating", "Accepted", "InProgress":
			icon = "⌛"
		case "Updating":
			icon = "🔄"
		case "Deleting":
			icon = "🗑️"
		}
		if res.Type == "Microsoft.Compute/virtualMachines/extensions" {
			// For VM Extensions, show only the extension name as resource name, and 'VM Extension' as friendly type
			parts := strings.Split(res.Name, "/")
			extName := parts[len(parts)-1]
			fval := "VM Extension"
			// Map extension name to friendly name if needed
			if extName == "Microsoft.Azure.Geneva.GenevaMonitoring" {
				extName = "Azure Geneva Monitoring"
			}
			msg := fmt.Sprintf("%s \"%s\" %s: %s", icon, extName, fval, res.State)
			fmt.Println(msg)
			continue
		}
		// Use a simplified friendly name without the utils package
		fval := res.Name
		msg := fmt.Sprintf("%s \"%s\" %s: %s", icon, res.Name, fval, res.State)
		fmt.Println(msg)
	}
}

// WaitForDeploymentAndShowStatus polls deployment status with enhanced interactive progress visualization
func (dd *DeploymentDisplay) WaitForDeploymentAndShowStatus(resourceGroup, deploymentName string) {
	start := time.Now()
	cWarn := color.New(color.FgHiYellow, color.Bold).SprintFunc()

	// Initialize cache phase and metrics collection
	dd.cache.UpdatePhase(PhaseInitialization)

	// Detect if we're in an interactive environment
	isInteractive := dd.isInteractiveEnvironment()
	dd.progressDisplay.SetInteractiveMode(isInteractive)

	// Start interactive mode if appropriate
	if isInteractive {
		dd.StartInteractiveMode()
		defer dd.StopInteractiveMode()
	}

	// Animation frames for fallback spinner during waiting periods
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIdx := 0

	// Hide cursor during deployment monitoring
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h") // Ensure cursor is restored on exit

	pollInterval := 5 * time.Second
	nextPoll := time.Now()

	// Track previous resource states for change detection
	previousResourceStates := make(map[string]string)
	previousResourceCount := 0
	lastDisplayUpdate := time.Now()
	noChangeCount := 0
	pollingCycle := 0

	// Cache warming - try to pre-populate with expected resources
	expectedResources := []models.ResourceStatus{
		{Name: "ArcBox-VM", Type: "Microsoft.Compute/virtualMachines", State: "Creating"},
		{Name: "ArcBox-Storage", Type: "Microsoft.Storage/storageAccounts", State: "Creating"},
		{Name: "ArcBox-VNet", Type: "Microsoft.Network/virtualNetworks", State: "Creating"},
	}
	dd.cache.WarmCache(expectedResources)

	// Variables for deployment state tracking
	var (
		state        string
		resources    []models.ResourceStatus
		allSucceeded bool
		anyFailed    bool
		lastPhase    DeploymentPhase = PhaseInitialization
	)

	// Initial message with accessibility support
	if isInteractive {
		fmt.Println("🚀 Starting ArcBox deployment monitoring with enhanced interactive progress display...")
		dd.progressDisplay.AddAccessibilityAnnouncement("Starting enhanced deployment monitoring")
	} else {
		fmt.Println("🚀 Starting ArcBox deployment monitoring...")
	}
	fmt.Println()

	for {
		currentTime := time.Now()

		if currentTime.After(nextPoll) {
			pollingCycle++
			pollStartTime := time.Now()

			// Calculate actual polling interval
			actualInterval := currentTime.Sub(nextPoll.Add(-pollInterval))

			// Get current deployment and resource status
			state = dd.getDeploymentProvisioningState(resourceGroup, deploymentName)
			resources = dd.getDeploymentResourceStatus(resourceGroup)

			pollEndTime := time.Now()

			// Detect deployment phase based on resource states and progress
			currentPhase := dd.detectDeploymentPhase(resources, state)
			if currentPhase != lastPhase {
				dd.cache.UpdatePhase(currentPhase)
				dd.progressDisplay.UpdatePhase(currentPhase)
				lastPhase = currentPhase
			}

			// Analyze changes since last poll and record state change metrics
			hasChanges, resourceChanges := dd.detectResourceChanges(resources, previousResourceStates)
			resourceCountChanged := len(resources) != previousResourceCount

			// Record state changes with timing metrics
			changeDetectionTime := time.Now()
			for _, change := range resourceChanges {
				dd.metricsCollector.RecordStateChange(change.Name, change.Type, change.FromState,
					change.ToState, changeDetectionTime, changeDetectionTime) // Display time same as detection for now
			}

			// Update progress display with current resource state
			dd.progressDisplay.UpdateProgress(resources)

			// Add detected changes to recent changes
			for _, change := range resourceChanges {
				dd.progressDisplay.AddRecentChange(change)

				// Log the change for monitoring
				fmt.Printf("🔄 %s: %s → %s\n", change.Name, change.FromState, change.ToState)
			}

			// Analyze resource states for completion/failure detection
			allSucceeded = true
			anyFailed = false
			for _, r := range resources {
				if r.State == "Failed" || r.State == "Canceled" {
					anyFailed = true
				}
				if r.State != "Succeeded" {
					allSucceeded = false
				}
			}

			// Update resource counters for progress tracking
			// (Interactive display components would be updated here)

			// Determine if we should update the display
			shouldUpdateDisplay := false
			timeSinceLastDisplay := currentTime.Sub(lastDisplayUpdate)

			switch {
			case len(resources) == 0:
				// Special case for waiting for resources to appear
				if isInteractive {
					dd.displayInteractiveWaiting()
				} else {
					dd.displayWaitingForResources(frames, &frameIdx)
				}
			case hasChanges || resourceCountChanged:
				// Resource changes detected - definitely update
				shouldUpdateDisplay = true
				noChangeCount = 0
			case timeSinceLastDisplay > 30*time.Second:
				// Periodic update even without changes
				shouldUpdateDisplay = true
			case dd.progressDisplay.CurrentPhase == PhaseInitialization && timeSinceLastDisplay > 10*time.Second:
				// More frequent updates during initialization
				shouldUpdateDisplay = true
			}

			// Update display if needed
			if shouldUpdateDisplay && len(resources) > 0 {
				dd.progressDisplay.DisplayProgressUpdate()

				// Accessibility announcement for significant changes
				if isInteractive && (hasChanges || resourceCountChanged) {
					accessibleSummary := dd.progressDisplay.GetAccessibleStatusSummary()
					dd.progressDisplay.AddAccessibilityAnnouncement(accessibleSummary)
				}

				lastDisplayUpdate = currentTime
				noChangeCount = 0
			} else if len(resources) > 0 {
				noChangeCount++
				// Show minimal activity indicator if no changes for a while
				if noChangeCount > 3 {
					if isInteractive {
						dd.displayMinimalInteractiveActivity()
					} else {
						dd.displayMinimalActivity(frames, &frameIdx)
					}
				}
			}

			// Check for deployment completion or failure
			if anyFailed || state == "Failed" || state == "Canceled" {
				dd.displayDeploymentFailure(resourceGroup, deploymentName, cWarn)
				break
			}

			if state == "Succeeded" && allSucceeded {
				dd.displayDeploymentSuccess(resources, start, resourceGroup, deploymentName)
				break
			}

			// Update tracking variables
			previousResourceStates = dd.createResourceStateMap(resources)
			previousResourceCount = len(resources)
			nextPoll = currentTime.Add(pollInterval)

			// Record polling metrics
			apiCallsMade := 1 + len(resources) // Deployment call + resource calls (approximation)
			changesDetected := len(resourceChanges)

			dd.metricsCollector.RecordPollingCycle(pollingCycle, pollInterval, actualInterval,
				pollStartTime, pollEndTime, len(resources), changesDetected, apiCallsMade)
		}

		// Timeout check (30 minutes)
		if time.Since(start) > 30*time.Minute {
			timeoutMsg := "⚠️  [WARN] Deployment is taking longer than 30 minutes. Please check the Azure Portal for more details."
			fmt.Println(cWarn(timeoutMsg))

			if isInteractive {
				dd.progressDisplay.AddAccessibilityAnnouncement("Deployment timeout warning: Please check Azure Portal")
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Print cache performance metrics at the end
	dd.printCacheMetrics()

	// Print comprehensive performance metrics
	dd.printPerformanceMetrics()

	// Stop metrics collection
	dd.metricsCollector.Stop()
}

// detectDeploymentPhase analyzes resources to determine the current deployment phase
func (dd *DeploymentDisplay) detectDeploymentPhase(resources []models.ResourceStatus, deploymentState string) DeploymentPhase {
	if len(resources) == 0 {
		return PhaseInitialization
	}

	// Count resources by type and state
	vmCount := 0
	storageCount := 0
	networkCount := 0
	extensionCount := 0
	succeededCount := 0
	totalCount := len(resources)

	for _, res := range resources {
		if res.State == "Succeeded" {
			succeededCount++
		}

		switch {
		case strings.Contains(res.Type, "Microsoft.Compute/virtualMachines"):
			if !strings.Contains(res.Type, "extensions") {
				vmCount++
			} else {
				extensionCount++
			}
		case strings.Contains(res.Type, "Microsoft.Storage"):
			storageCount++
		case strings.Contains(res.Type, "Microsoft.Network"):
			networkCount++
		}
	}

	completionPercentage := float64(succeededCount) / float64(totalCount)

	// Phase detection logic
	switch {
	case completionPercentage >= 0.95:
		return PhaseFinalized
	case completionPercentage >= 0.8:
		return PhaseCompletion
	case extensionCount > 0 && (completionPercentage >= 0.6 || vmCount > 0):
		return PhaseConfiguration
	case vmCount > 0 || storageCount > 0 || networkCount > 0:
		return PhaseInfrastructure
	default:
		return PhaseInitialization
	}
}

// printCacheMetrics prints cache performance metrics at the end of deployment
func (dd *DeploymentDisplay) printCacheMetrics() {
	metrics := dd.cache.GetMetrics()
	hitRate := dd.cache.GetCacheHitRate()

	fmt.Println("\n📊 Cache Performance Metrics:")
	fmt.Printf("   • Cache Hit Rate: %.1f%% (%d hits / %d requests)\n", hitRate, metrics.CacheHits, metrics.TotalRequests)
	fmt.Printf("   • API Calls Saved: %d\n", metrics.APICallsSaved)
	fmt.Printf("   • Cache Invalidations: %d\n", metrics.Invalidations)

	status := dd.cache.GetCacheStatus()
	fmt.Printf("   • Cache Entries: %d resources, %d deployments\n",
		status["resource_entries"], status["deployment_entries"])

	if hitRate >= 50.0 {
		fmt.Printf("   ✅ Cache effectiveness target achieved (%.1f%% >= 50%%)\n", hitRate)
	} else {
		fmt.Printf("   ⚠️  Cache effectiveness below target (%.1f%% < 50%%)\n", hitRate)
	}
}

// printPerformanceMetrics prints comprehensive performance metrics at the end of deployment monitoring
func (dd *DeploymentDisplay) printPerformanceMetrics() {
	if dd.metricsCollector == nil {
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 DEPLOYMENT MONITORING PERFORMANCE METRICS")
	fmt.Println(strings.Repeat("=", 60))

	// Print the main performance summary
	dd.metricsCollector.PrintPerformanceSummary()

	// For detailed level and above, provide additional insights
	if dd.metricsCollector.CollectionLevel >= MetricsDetailed {
		dd.printDetailedPerformanceAnalysis()
	}

	// For debug level, provide export options
	if dd.metricsCollector.CollectionLevel >= MetricsDebug {
		dd.printMetricsExportInfo()
	}
}

// printDetailedPerformanceAnalysis prints detailed performance analysis
func (dd *DeploymentDisplay) printDetailedPerformanceAnalysis() {
	report := dd.metricsCollector.GetPerformanceReport()

	fmt.Println("\n🔍 Detailed Performance Analysis")
	fmt.Println("================================")

	// API Call Analysis
	if apiCalls, ok := report["api_call_details"].([]APICallMetric); ok && len(apiCalls) > 0 {
		fmt.Printf("\n🔌 API Call Breakdown:\n")

		// Group by operation type
		operationStats := make(map[string]struct {
			count         int
			totalTime     time.Duration
			successCount  int
			cacheHitCount int
		})

		for _, call := range apiCalls {
			stats := operationStats[call.Operation]
			stats.count++
			stats.totalTime += call.Duration
			if call.Success {
				stats.successCount++
			}
			if call.CacheHit {
				stats.cacheHitCount++
			}
			operationStats[call.Operation] = stats
		}

		for operation, stats := range operationStats {
			avgDuration := stats.totalTime / time.Duration(stats.count)
			successRate := float64(stats.successCount) / float64(stats.count) * 100
			cacheHitRate := float64(stats.cacheHitCount) / float64(stats.count) * 100

			fmt.Printf("   • %s: %d calls, avg: %v, success: %.1f%%, cache: %.1f%%\n",
				operation, stats.count, avgDuration, successRate, cacheHitRate)
		}
	}

	// Polling Analysis
	if pollingCycles, ok := report["polling_details"].([]PollingMetric); ok && len(pollingCycles) > 0 {
		fmt.Printf("\n⏱️  Polling Performance:\n")

		totalDeviation := time.Duration(0)
		maxDeviation := time.Duration(0)
		totalChanges := 0
		totalAPICalls := 0

		for _, cycle := range pollingCycles {
			deviation := cycle.ActualInterval - cycle.IntendedInterval
			if deviation < 0 {
				deviation = -deviation
			}
			totalDeviation += deviation
			if deviation > maxDeviation {
				maxDeviation = deviation
			}
			totalChanges += cycle.ChangesDetected
			totalAPICalls += cycle.APICallsMade
		}

		avgDeviation := totalDeviation / time.Duration(len(pollingCycles))
		avgChangesPerCycle := float64(totalChanges) / float64(len(pollingCycles))
		avgAPICallsPerCycle := float64(totalAPICalls) / float64(len(pollingCycles))

		fmt.Printf("   • Polling Cycles: %d\n", len(pollingCycles))
		fmt.Printf("   • Average Timing Deviation: %v\n", avgDeviation)
		fmt.Printf("   • Maximum Timing Deviation: %v\n", maxDeviation)
		fmt.Printf("   • Average Changes per Cycle: %.1f\n", avgChangesPerCycle)
		fmt.Printf("   • Average API Calls per Cycle: %.1f\n", avgAPICallsPerCycle)
	}

	// Resource Query Analysis
	if resourceQueries, ok := report["resource_query_details"].([]ResourceQueryMetric); ok && len(resourceQueries) > 0 {
		fmt.Printf("\n💾 Resource Query Performance:\n")

		// Group by resource type
		resourceTypeStats := make(map[string]struct {
			count          int
			totalDuration  time.Duration
			totalResources int
			totalCacheHits int
			totalCacheMiss int
		})

		for _, query := range resourceQueries {
			stats := resourceTypeStats[query.ResourceType]
			stats.count++
			stats.totalDuration += query.QueryDuration
			stats.totalResources += query.ResourceCount
			stats.totalCacheHits += query.CacheHitCount
			stats.totalCacheMiss += query.CacheMissCount
			resourceTypeStats[query.ResourceType] = stats
		}

		for resourceType, stats := range resourceTypeStats {
			avgDuration := stats.totalDuration / time.Duration(stats.count)
			avgResourcesPerQuery := float64(stats.totalResources) / float64(stats.count)
			totalCacheRequests := stats.totalCacheHits + stats.totalCacheMiss
			cacheHitRate := float64(0)
			if totalCacheRequests > 0 {
				cacheHitRate = float64(stats.totalCacheHits) / float64(totalCacheRequests) * 100
			}

			fmt.Printf("   • %s: %d queries, avg: %v, resources/query: %.1f, cache: %.1f%%\n",
				resourceType, stats.count, avgDuration, avgResourcesPerQuery, cacheHitRate)
		}
	}

	// State Change Analysis
	if stateChanges, ok := report["state_change_details"].([]StateChangeMetric); ok && len(stateChanges) > 0 {
		fmt.Printf("\n🔍 State Change Performance:\n")

		totalDetectionLatency := time.Duration(0)
		totalDisplayLatency := time.Duration(0)
		stateTransitions := make(map[string]int)

		for _, change := range stateChanges {
			totalDetectionLatency += change.DetectionLatency
			totalDisplayLatency += change.DisplayLatency
			transition := change.FromState + "→" + change.ToState
			stateTransitions[transition]++
		}

		avgDetectionLatency := totalDetectionLatency / time.Duration(len(stateChanges))
		avgDisplayLatency := totalDisplayLatency / time.Duration(len(stateChanges))

		fmt.Printf("   • State Changes Detected: %d\n", len(stateChanges))
		fmt.Printf("   • Average Detection Latency: %v\n", avgDetectionLatency)
		fmt.Printf("   • Average Display Latency: %v\n", avgDisplayLatency)

		if len(stateTransitions) > 0 {
			fmt.Printf("   • Common Transitions:\n")
			for transition, count := range stateTransitions {
				if count > 1 { // Only show transitions that happened multiple times
					fmt.Printf("     - %s: %d times\n", transition, count)
				}
			}
		}
	}
}

// printMetricsExportInfo prints information about exporting metrics for debug level
func (dd *DeploymentDisplay) printMetricsExportInfo() {
	fmt.Println("\n📄 Metrics Export Options")
	fmt.Println("=========================")
	fmt.Println("For detailed analysis, metrics can be exported in JSON or YAML format.")
	fmt.Println("To export metrics programmatically, use:")
	fmt.Println("  • JSON: metricsCollector.ExportMetrics(\"json\")")
	fmt.Println("  • YAML: metricsCollector.ExportMetrics(\"yaml\")")

	// Provide a sample of how to access the data
	report := dd.metricsCollector.GetPerformanceReport()

	if dd.metricsCollector.VerboseOutput {
		fmt.Printf("\n🔍 Quick Export Sample (JSON snippet):\n")

		// Create a smaller sample for display
		sample := map[string]interface{}{
			"monitoring_duration":  report["monitoring_duration"],
			"total_api_calls":      report["total_api_calls"],
			"api_success_rate":     report["api_success_rate"],
			"total_polling_cycles": report["total_polling_cycles"],
			"efficiency_scores":    report["efficiency_scores"],
		}

		if jsonData, err := json.MarshalIndent(sample, "  ", "  "); err == nil {
			// Limit output to first few lines for readability
			lines := strings.Split(string(jsonData), "\n")
			maxLines := 10
			if len(lines) > maxLines {
				lines = lines[:maxLines]
				lines = append(lines, "  ... (truncated)")
			}
			fmt.Println("  " + strings.Join(lines, "\n  "))
		}
	}
}

// MetricsCollectionLevel defines the level of metrics collection
type MetricsCollectionLevel int

const (
	MetricsBasic MetricsCollectionLevel = iota
	MetricsDetailed
	MetricsDebug
)

// APICallMetric represents metrics for a single API call
type APICallMetric struct {
	Operation    string
	ResourceType string
	StartTime    time.Time
	Duration     time.Duration
	Success      bool
	ErrorMsg     string
	CacheHit     bool
}

// PollingMetric represents metrics for polling performance
type PollingMetric struct {
	Cycle            int
	IntendedInterval time.Duration
	ActualInterval   time.Duration
	PollStartTime    time.Time
	PollEndTime      time.Time
	ResourcesPolled  int
	ChangesDetected  int
	APICallsMade     int
}

// ResourceQueryMetric represents metrics for resource queries
type ResourceQueryMetric struct {
	ResourceType   string
	QueryStartTime time.Time
	QueryDuration  time.Duration
	ResourceCount  int
	CacheHitCount  int
	CacheMissCount int
	Success        bool
	ErrorMsg       string
}

// StateChangeMetric represents metrics for state change detection
type StateChangeMetric struct {
	ResourceName     string
	ResourceType     string
	FromState        string
	ToState          string
	DetectionTime    time.Time
	DisplayTime      time.Time
	DetectionLatency time.Duration
	DisplayLatency   time.Duration
}

// PerformanceMetricsCollector manages comprehensive performance metrics
type PerformanceMetricsCollector struct {
	CollectionLevel        MetricsCollectionLevel
	StartTime              time.Time
	APICallMetrics         []APICallMetric
	PollingMetrics         []PollingMetric
	ResourceQueryMetrics   []ResourceQueryMetric
	StateChangeMetrics     []StateChangeMetric
	TotalAPICallCount      int64
	TotalAPICallDuration   time.Duration
	TotalPollingCycles     int64
	TotalResourceQueries   int64
	TotalStateChanges      int64
	APISuccessRate         float64
	AveragePollingInterval time.Duration
	AverageAPICallDuration time.Duration
	mutex                  sync.RWMutex
	VerboseOutput          bool
	metricsChannel         chan interface{}
	stopChannel            chan bool
}

// EfficiencyScore represents calculated efficiency metrics
type EfficiencyScore struct {
	OverallScore           float64
	APIEfficiencyScore     float64
	PollingEfficiencyScore float64
	CacheEfficiencyScore   float64
	StateDetectionScore    float64
	BottleneckIndicators   []string
	ImprovementSuggestions []string
	LastCalculated         time.Time
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Data        interface{}
	Timestamp   time.Time
	TTL         time.Duration
	AccessCount int
	LastAccess  time.Time
}

// CacheMetrics tracks cache performance for optimization
type CacheMetrics struct {
	TotalRequests int64
	CacheHits     int64
	CacheMisses   int64
	Invalidations int64
	APICallsSaved int64
	LastResetTime time.Time
	mutex         sync.RWMutex
}

// ResourceTypeConfig defines caching behavior for different resource types
type ResourceTypeConfig struct {
	StableTTL       time.Duration // TTL for stable states (Succeeded/Failed)
	TransitionalTTL time.Duration // TTL for transitional states (Creating/Running)
	PollPriority    int           // Higher = more likely to be refreshed
}

// IntelligentCache manages deployment resource and state caching
type IntelligentCache struct {
	resourceCache   map[string]*CacheEntry
	deploymentCache map[string]*CacheEntry
	resourceTypes   map[string]*ResourceTypeConfig
	metrics         *CacheMetrics
	mutex           sync.RWMutex
	enabled         bool
	maxEntries      int
	currentPhase    DeploymentPhase
	lastPhaseChange time.Time
}

// CacheInvalidationReason represents why cache was invalidated
type CacheInvalidationReason int

const (
	InvalidationStateChange CacheInvalidationReason = iota
	InvalidationPhaseChange
	InvalidationTTLExpired
	InvalidationManual
	InvalidationCapacityLimit
)

// NewPerformanceMetricsCollector creates a new performance metrics collector
func NewPerformanceMetricsCollector(level MetricsCollectionLevel, verbose bool) *PerformanceMetricsCollector {
	pmc := &PerformanceMetricsCollector{
		CollectionLevel:      level,
		StartTime:            time.Now(),
		APICallMetrics:       make([]APICallMetric, 0),
		PollingMetrics:       make([]PollingMetric, 0),
		ResourceQueryMetrics: make([]ResourceQueryMetric, 0),
		StateChangeMetrics:   make([]StateChangeMetric, 0),
		VerboseOutput:        verbose,
		metricsChannel:       make(chan interface{}, 1000),
		stopChannel:          make(chan bool, 1),
	}

	// Start background metrics processing for detailed/debug levels
	if level >= MetricsDetailed {
		go pmc.backgroundMetricsProcessor()
	}

	return pmc
}

// backgroundMetricsProcessor processes metrics in the background
func (pmc *PerformanceMetricsCollector) backgroundMetricsProcessor() {
	for {
		select {
		case metric := <-pmc.metricsChannel:
			pmc.processMetricAsync(metric)
		case <-pmc.stopChannel:
			return
		}
	}
}

// processMetricAsync processes a metric asynchronously
func (pmc *PerformanceMetricsCollector) processMetricAsync(metric interface{}) {
	switch m := metric.(type) {
	case APICallMetric:
		pmc.aggregateAPICallMetrics(m)
	case PollingMetric:
		pmc.aggregatePollingMetrics(m)
	case ResourceQueryMetric:
		pmc.aggregateResourceQueryMetrics(m)
	case StateChangeMetric:
		pmc.aggregateStateChangeMetrics(m)
	}
}

// RecordAPICall records metrics for an API call
func (pmc *PerformanceMetricsCollector) RecordAPICall(operation, resourceType string, duration time.Duration, success bool, errorMsg string, cacheHit bool) {
	if pmc.CollectionLevel == MetricsBasic && !success {
		// Basic level only tracks failures
		return
	}

	metric := APICallMetric{
		Operation:    operation,
		ResourceType: resourceType,
		StartTime:    time.Now().Add(-duration),
		Duration:     duration,
		Success:      success,
		ErrorMsg:     errorMsg,
		CacheHit:     cacheHit,
	}

	if pmc.CollectionLevel >= MetricsDetailed {
		// Send to background processor for detailed processing
		select {
		case pmc.metricsChannel <- metric:
		default:
			// Channel full, process synchronously
			pmc.processAPICallMetric(metric)
		}
	} else {
		pmc.processAPICallMetric(metric)
	}

	if pmc.VerboseOutput && pmc.CollectionLevel >= MetricsDebug {
		fmt.Printf("[METRICS] API Call: %s (%s) - Duration: %v, Success: %t, Cache: %t\n",
			operation, resourceType, duration, success, cacheHit)
	}
}

// processAPICallMetric processes an API call metric
func (pmc *PerformanceMetricsCollector) processAPICallMetric(metric APICallMetric) {
	pmc.mutex.Lock()
	defer pmc.mutex.Unlock()

	if pmc.CollectionLevel >= MetricsDetailed {
		pmc.APICallMetrics = append(pmc.APICallMetrics, metric)
	}

	pmc.TotalAPICallCount++
	pmc.TotalAPICallDuration += metric.Duration

	// Calculate running success rate
	successCount := int64(0)
	if pmc.CollectionLevel >= MetricsDetailed {
		for _, m := range pmc.APICallMetrics {
			if m.Success {
				successCount++
			}
		}
		pmc.APISuccessRate = float64(successCount) / float64(len(pmc.APICallMetrics)) * 100
	} else {
		// For basic level, maintain a simple running average
		if metric.Success {
			successCount = pmc.TotalAPICallCount // Simplified for basic level
		}
		pmc.APISuccessRate = float64(successCount) / float64(pmc.TotalAPICallCount) * 100
	}

	pmc.AverageAPICallDuration = pmc.TotalAPICallDuration / time.Duration(pmc.TotalAPICallCount)
}

// RecordPollingCycle records metrics for a polling cycle
func (pmc *PerformanceMetricsCollector) RecordPollingCycle(cycle int, intendedInterval, actualInterval time.Duration,
	pollStart, pollEnd time.Time, resourcesPolled, changesDetected, apiCallsMade int) {

	metric := PollingMetric{
		Cycle:            cycle,
		IntendedInterval: intendedInterval,
		ActualInterval:   actualInterval,
		PollStartTime:    pollStart,
		PollEndTime:      pollEnd,
		ResourcesPolled:  resourcesPolled,
		ChangesDetected:  changesDetected,
		APICallsMade:     apiCallsMade,
	}

	if pmc.CollectionLevel >= MetricsDetailed {
		select {
		case pmc.metricsChannel <- metric:
		default:
			pmc.processPollingMetric(metric)
		}
	} else {
		pmc.processPollingMetric(metric)
	}

	if pmc.VerboseOutput && pmc.CollectionLevel >= MetricsDebug {
		fmt.Printf("[METRICS] Polling Cycle %d: Intended: %v, Actual: %v, Resources: %d, Changes: %d, API Calls: %d\n",
			cycle, intendedInterval, actualInterval, resourcesPolled, changesDetected, apiCallsMade)
	}
}

// processPollingMetric processes a polling metric
func (pmc *PerformanceMetricsCollector) processPollingMetric(metric PollingMetric) {
	pmc.mutex.Lock()
	defer pmc.mutex.Unlock()

	if pmc.CollectionLevel >= MetricsDetailed {
		pmc.PollingMetrics = append(pmc.PollingMetrics, metric)
	}

	pmc.TotalPollingCycles++

	// Calculate running average polling interval
	totalInterval := time.Duration(0)
	if pmc.CollectionLevel >= MetricsDetailed {
		for _, m := range pmc.PollingMetrics {
			totalInterval += m.ActualInterval
		}
		pmc.AveragePollingInterval = totalInterval / time.Duration(len(pmc.PollingMetrics))
	} else {
		// Simplified for basic level
		pmc.AveragePollingInterval = metric.ActualInterval
	}
}

// RecordResourceQuery records metrics for resource queries
func (pmc *PerformanceMetricsCollector) RecordResourceQuery(resourceType string, queryStart time.Time,
	queryDuration time.Duration, resourceCount, cacheHitCount, cacheMissCount int, success bool, errorMsg string) {

	metric := ResourceQueryMetric{
		ResourceType:   resourceType,
		QueryStartTime: queryStart,
		QueryDuration:  queryDuration,
		ResourceCount:  resourceCount,
		CacheHitCount:  cacheHitCount,
		CacheMissCount: cacheMissCount,
		Success:        success,
		ErrorMsg:       errorMsg,
	}

	if pmc.CollectionLevel >= MetricsDetailed {
		select {
		case pmc.metricsChannel <- metric:
		default:
			pmc.processResourceQueryMetric(metric)
		}
	} else {
		pmc.processResourceQueryMetric(metric)
	}

	if pmc.VerboseOutput && pmc.CollectionLevel >= MetricsDebug {
		fmt.Printf("[METRICS] Resource Query: %s - Duration: %v, Count: %d, Cache Hits: %d, Misses: %d, Success: %t\n",
			resourceType, queryDuration, resourceCount, cacheHitCount, cacheMissCount, success)
	}
}

// processResourceQueryMetric processes a resource query metric
func (pmc *PerformanceMetricsCollector) processResourceQueryMetric(metric ResourceQueryMetric) {
	pmc.mutex.Lock()
	defer pmc.mutex.Unlock()

	if pmc.CollectionLevel >= MetricsDetailed {
		pmc.ResourceQueryMetrics = append(pmc.ResourceQueryMetrics, metric)
	}

	pmc.TotalResourceQueries++
}

// RecordStateChange records metrics for state change detection
func (pmc *PerformanceMetricsCollector) RecordStateChange(resourceName, resourceType, fromState, toState string,
	detectionTime, displayTime time.Time) {

	detectionLatency := detectionTime.Sub(time.Now()) // This would need the actual state change time
	displayLatency := displayTime.Sub(detectionTime)

	metric := StateChangeMetric{
		ResourceName:     resourceName,
		ResourceType:     resourceType,
		FromState:        fromState,
		ToState:          toState,
		DetectionTime:    detectionTime,
		DisplayTime:      displayTime,
		DetectionLatency: detectionLatency,
		DisplayLatency:   displayLatency,
	}

	if pmc.CollectionLevel >= MetricsDetailed {
		select {
		case pmc.metricsChannel <- metric:
		default:
			pmc.processStateChangeMetric(metric)
		}
	} else {
		pmc.processStateChangeMetric(metric)
	}

	if pmc.VerboseOutput && pmc.CollectionLevel >= MetricsDebug {
		fmt.Printf("[METRICS] State Change: %s (%s) %s→%s - Detection: %v, Display: %v\n",
			resourceName, resourceType, fromState, toState, detectionLatency, displayLatency)
	}
}

// processStateChangeMetric processes a state change metric
func (pmc *PerformanceMetricsCollector) processStateChangeMetric(metric StateChangeMetric) {
	pmc.mutex.Lock()
	defer pmc.mutex.Unlock()

	if pmc.CollectionLevel >= MetricsDetailed {
		pmc.StateChangeMetrics = append(pmc.StateChangeMetrics, metric)
	}

	pmc.TotalStateChanges++
}

// aggregateAPICallMetrics aggregates API call metrics for background processing
func (pmc *PerformanceMetricsCollector) aggregateAPICallMetrics(metric APICallMetric) {
	// Additional aggregation logic for detailed/debug levels
	pmc.processAPICallMetric(metric)
}

// aggregatePollingMetrics aggregates polling metrics for background processing
func (pmc *PerformanceMetricsCollector) aggregatePollingMetrics(metric PollingMetric) {
	// Additional aggregation logic for detailed/debug levels
	pmc.processPollingMetric(metric)
}

// aggregateResourceQueryMetrics aggregates resource query metrics for background processing
func (pmc *PerformanceMetricsCollector) aggregateResourceQueryMetrics(metric ResourceQueryMetric) {
	// Additional aggregation logic for detailed/debug levels
	pmc.processResourceQueryMetric(metric)
}

// aggregateStateChangeMetrics aggregates state change metrics for background processing
func (pmc *PerformanceMetricsCollector) aggregateStateChangeMetrics(metric StateChangeMetric) {
	// Additional aggregation logic for detailed/debug levels
	pmc.processStateChangeMetric(metric)
}

// CalculateEfficiencyScore calculates comprehensive efficiency scores
func (pmc *PerformanceMetricsCollector) CalculateEfficiencyScore() EfficiencyScore {
	pmc.mutex.RLock()
	defer pmc.mutex.RUnlock()

	score := EfficiencyScore{
		LastCalculated:         time.Now(),
		BottleneckIndicators:   make([]string, 0),
		ImprovementSuggestions: make([]string, 0),
	}

	// Calculate API efficiency score (0-100)
	if pmc.TotalAPICallCount > 0 {
		score.APIEfficiencyScore = pmc.APISuccessRate

		// Check for API bottlenecks
		if pmc.AverageAPICallDuration > 2*time.Second {
			score.BottleneckIndicators = append(score.BottleneckIndicators, "Slow API calls")
			score.ImprovementSuggestions = append(score.ImprovementSuggestions, "Consider optimizing Azure CLI call patterns")
		}

		if pmc.APISuccessRate < 95.0 {
			score.BottleneckIndicators = append(score.BottleneckIndicators, "High API failure rate")
			score.ImprovementSuggestions = append(score.ImprovementSuggestions, "Implement better error handling and retry logic")
		}
	} else {
		score.APIEfficiencyScore = 100.0 // No API calls means perfect efficiency
	}

	// Calculate polling efficiency score
	if pmc.TotalPollingCycles > 0 {
		// Efficiency based on how close actual intervals are to intended intervals
		intervalAccuracy := 100.0
		if pmc.CollectionLevel >= MetricsDetailed && len(pmc.PollingMetrics) > 0 {
			totalDeviation := float64(0)
			for _, metric := range pmc.PollingMetrics {
				deviation := float64(metric.ActualInterval-metric.IntendedInterval) / float64(metric.IntendedInterval)
				totalDeviation += deviation * deviation // Squared deviation
			}
			avgDeviation := totalDeviation / float64(len(pmc.PollingMetrics))
			intervalAccuracy = math.Max(0, 100.0-avgDeviation*100)
		}
		score.PollingEfficiencyScore = intervalAccuracy

		// Check for polling bottlenecks
		if pmc.AveragePollingInterval > 10*time.Second {
			score.BottleneckIndicators = append(score.BottleneckIndicators, "Slow polling intervals")
			score.ImprovementSuggestions = append(score.ImprovementSuggestions, "Implement adaptive polling strategies")
		}
	} else {
		score.PollingEfficiencyScore = 100.0
	}

	// Calculate cache efficiency score (this would need integration with cache metrics)
	// For now, use a placeholder that can be updated when cache metrics are available
	score.CacheEfficiencyScore = 75.0 // Placeholder

	// Calculate state detection score
	if pmc.TotalStateChanges > 0 {
		avgDetectionLatency := time.Duration(0)
		avgDisplayLatency := time.Duration(0)

		if pmc.CollectionLevel >= MetricsDetailed && len(pmc.StateChangeMetrics) > 0 {
			totalDetectionLatency := time.Duration(0)
			totalDisplayLatency := time.Duration(0)

			for _, metric := range pmc.StateChangeMetrics {
				totalDetectionLatency += metric.DetectionLatency
				totalDisplayLatency += metric.DisplayLatency
			}

			avgDetectionLatency = totalDetectionLatency / time.Duration(len(pmc.StateChangeMetrics))
			avgDisplayLatency = totalDisplayLatency / time.Duration(len(pmc.StateChangeMetrics))
		}

		// Score based on latency (lower is better)
		detectionScore := math.Max(0, 100.0-float64(avgDetectionLatency/time.Second)*10)
		displayScore := math.Max(0, 100.0-float64(avgDisplayLatency/time.Second)*20)
		score.StateDetectionScore = (detectionScore + displayScore) / 2

		// Check for state detection bottlenecks
		if avgDetectionLatency > 5*time.Second {
			score.BottleneckIndicators = append(score.BottleneckIndicators, "Slow state change detection")
			score.ImprovementSuggestions = append(score.ImprovementSuggestions, "Optimize state change detection algorithms")
		}

		if avgDisplayLatency > 2*time.Second {
			score.BottleneckIndicators = append(score.BottleneckIndicators, "Slow state change display")
			score.ImprovementSuggestions = append(score.ImprovementSuggestions, "Optimize console output performance")
		}
	} else {
		score.StateDetectionScore = 100.0
	}

	// Calculate overall score (weighted average)
	score.OverallScore = (score.APIEfficiencyScore*0.3 +
		score.PollingEfficiencyScore*0.2 +
		score.CacheEfficiencyScore*0.3 +
		score.StateDetectionScore*0.2)

	return score
}

// GetPerformanceReport generates a comprehensive performance report
func (pmc *PerformanceMetricsCollector) GetPerformanceReport() map[string]interface{} {
	pmc.mutex.RLock()
	defer pmc.mutex.RUnlock()

	report := make(map[string]interface{})

	// Basic metrics
	report["collection_level"] = pmc.CollectionLevel
	report["monitoring_duration"] = time.Since(pmc.StartTime)
	report["total_api_calls"] = pmc.TotalAPICallCount
	report["total_api_duration"] = pmc.TotalAPICallDuration
	report["average_api_duration"] = pmc.AverageAPICallDuration
	report["api_success_rate"] = pmc.APISuccessRate
	report["total_polling_cycles"] = pmc.TotalPollingCycles
	report["average_polling_interval"] = pmc.AveragePollingInterval
	report["total_resource_queries"] = pmc.TotalResourceQueries
	report["total_state_changes"] = pmc.TotalStateChanges

	// Detailed metrics (if collection level allows)
	if pmc.CollectionLevel >= MetricsDetailed {
		report["api_call_details"] = pmc.APICallMetrics
		report["polling_details"] = pmc.PollingMetrics
		report["resource_query_details"] = pmc.ResourceQueryMetrics
		report["state_change_details"] = pmc.StateChangeMetrics
	}

	// Efficiency scores
	efficiency := pmc.CalculateEfficiencyScore()
	report["efficiency_scores"] = efficiency

	return report
}

// ExportMetrics exports metrics in structured format
func (pmc *PerformanceMetricsCollector) ExportMetrics(format string) ([]byte, error) {
	report := pmc.GetPerformanceReport()

	switch format {
	case "json":
		return json.Marshal(report)
	case "yaml":
		return yaml.Marshal(report)
	default:
		return json.Marshal(report) // Default to JSON
	}
}

// Stop stops the background metrics processor
func (pmc *PerformanceMetricsCollector) Stop() {
	if pmc.CollectionLevel >= MetricsDetailed {
		select {
		case pmc.stopChannel <- true:
		default:
		}
	}
}

// PrintPerformanceSummary prints a formatted performance summary
func (pmc *PerformanceMetricsCollector) PrintPerformanceSummary() {
	efficiency := pmc.CalculateEfficiencyScore()

	fmt.Println("\n📊 Performance Metrics Summary")
	fmt.Println("==============================")

	fmt.Printf("📈 Overall Efficiency Score: %.1f/100\n", efficiency.OverallScore)
	fmt.Printf("🔌 API Efficiency: %.1f%% (Success Rate: %.1f%%)\n", efficiency.APIEfficiencyScore, pmc.APISuccessRate)
	fmt.Printf("⏱️  Polling Efficiency: %.1f%%\n", efficiency.PollingEfficiencyScore)
	fmt.Printf("💾 Cache Efficiency: %.1f%%\n", efficiency.CacheEfficiencyScore)
	fmt.Printf("🔍 State Detection Efficiency: %.1f%%\n", efficiency.StateDetectionScore)

	fmt.Printf("\n📊 Key Metrics:\n")
	fmt.Printf("   • Total API Calls: %d\n", pmc.TotalAPICallCount)
	fmt.Printf("   • Average API Duration: %v\n", pmc.AverageAPICallDuration)
	fmt.Printf("   • Polling Cycles: %d\n", pmc.TotalPollingCycles)
	fmt.Printf("   • Average Polling Interval: %v\n", pmc.AveragePollingInterval)
	fmt.Printf("   • Resource Queries: %d\n", pmc.TotalResourceQueries)
	fmt.Printf("   • State Changes: %d\n", pmc.TotalStateChanges)
	fmt.Printf("   • Monitoring Duration: %v\n", time.Since(pmc.StartTime))

	if len(efficiency.BottleneckIndicators) > 0 {
		fmt.Printf("\n⚠️  Bottlenecks Detected:\n")
		for _, bottleneck := range efficiency.BottleneckIndicators {
			fmt.Printf("   • %s\n", bottleneck)
		}
	}

	if len(efficiency.ImprovementSuggestions) > 0 {
		fmt.Printf("\n💡 Improvement Suggestions:\n")
		for _, suggestion := range efficiency.ImprovementSuggestions {
			fmt.Printf("   • %s\n", suggestion)
		}
	}

	// Performance grade
	grade := "F"
	switch {
	case efficiency.OverallScore >= 90:
		grade = "A"
	case efficiency.OverallScore >= 80:
		grade = "B"
	case efficiency.OverallScore >= 70:
		grade = "C"
	case efficiency.OverallScore >= 60:
		grade = "D"
	}

	fmt.Printf("\n🎯 Performance Grade: %s\n", grade)

	if pmc.VerboseOutput && pmc.CollectionLevel >= MetricsDebug {
		fmt.Printf("\n🔍 Debug Information:\n")
		fmt.Printf("   • Collection Level: %d\n", pmc.CollectionLevel)
		fmt.Printf("   • Metrics Channel Capacity: %d\n", cap(pmc.metricsChannel))
		fmt.Printf("   • Metrics Channel Length: %d\n", len(pmc.metricsChannel))
	}
}

// NewIntelligentCache creates a new intelligent cache instance
func NewIntelligentCache() *IntelligentCache {
	return &IntelligentCache{
		resourceCache:   make(map[string]*CacheEntry),
		deploymentCache: make(map[string]*CacheEntry),
		resourceTypes:   initializeResourceTypeConfigs(),
		metrics:         &CacheMetrics{LastResetTime: time.Now()},
		enabled:         true,
		maxEntries:      1000, // Configurable limit
		currentPhase:    PhaseInitialization,
		lastPhaseChange: time.Now(),
	}
}

// initializeResourceTypeConfigs sets up caching behavior for different Azure resource types
func initializeResourceTypeConfigs() map[string]*ResourceTypeConfig {
	configs := make(map[string]*ResourceTypeConfig)

	// Virtual Machines - slower to change, higher cache times
	configs["Microsoft.Compute/virtualMachines"] = &ResourceTypeConfig{
		StableTTL:       2 * time.Minute,
		TransitionalTTL: 30 * time.Second,
		PollPriority:    2,
	}

	// VM Extensions - can change frequently during configuration
	configs["Microsoft.Compute/virtualMachines/extensions"] = &ResourceTypeConfig{
		StableTTL:       45 * time.Second,
		TransitionalTTL: 15 * time.Second,
		PollPriority:    3,
	}

	// Storage Accounts - relatively stable once created
	configs["Microsoft.Storage/storageAccounts"] = &ResourceTypeConfig{
		StableTTL:       3 * time.Minute,
		TransitionalTTL: 45 * time.Second,
		PollPriority:    1,
	}

	// Network resources - generally stable
	configs["Microsoft.Network/virtualNetworks"] = &ResourceTypeConfig{
		StableTTL:       2 * time.Minute,
		TransitionalTTL: 30 * time.Second,
		PollPriority:    1,
	}
	configs["Microsoft.Network/networkSecurityGroups"] = &ResourceTypeConfig{
		StableTTL:       2 * time.Minute,
		TransitionalTTL: 30 * time.Second,
		PollPriority:    1,
	}
	configs["Microsoft.Network/networkInterfaces"] = &ResourceTypeConfig{
		StableTTL:       90 * time.Second,
		TransitionalTTL: 25 * time.Second,
		PollPriority:    2,
	}

	// Default configuration for unknown resource types
	configs["default"] = &ResourceTypeConfig{
		StableTTL:       60 * time.Second,
		TransitionalTTL: 20 * time.Second,
		PollPriority:    2,
	}

	return configs
}

// UpdatePhase updates the current deployment phase for cache optimization
func (ic *IntelligentCache) UpdatePhase(phase DeploymentPhase) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	if ic.currentPhase != phase {
		ic.currentPhase = phase
		ic.lastPhaseChange = time.Now()

		// Adjust cache behavior based on phase
		ic.optimizeForPhase(phase)
	}
}

// optimizeForPhase adjusts cache settings based on deployment phase
func (ic *IntelligentCache) optimizeForPhase(phase DeploymentPhase) {
	switch phase {
	case PhaseInitialization:
		// During initialization, cache aggressively as resources are just being created
		ic.adjustAllTTLs(0.5) // Reduce TTLs by 50%
	case PhaseInfrastructure:
		// During infrastructure phase, moderate caching
		ic.adjustAllTTLs(0.8) // Reduce TTLs by 20%
	case PhaseConfiguration:
		// During configuration, VM extensions change frequently
		ic.adjustSpecificTTL("Microsoft.Compute/virtualMachines/extensions", 0.3)
	case PhaseCompletion:
		// During completion, most resources are stable
		ic.adjustAllTTLs(1.2) // Increase TTLs by 20%
	case PhaseFinalized:
		// After completion, resources are very stable
		ic.adjustAllTTLs(2.0) // Double TTLs
	}
}

// adjustAllTTLs adjusts all resource type TTLs by a multiplier
func (ic *IntelligentCache) adjustAllTTLs(multiplier float64) {
	for _, config := range ic.resourceTypes {
		config.StableTTL = time.Duration(float64(config.StableTTL) * multiplier)
		config.TransitionalTTL = time.Duration(float64(config.TransitionalTTL) * multiplier)
	}
}

// adjustSpecificTTL adjusts TTL for a specific resource type
func (ic *IntelligentCache) adjustSpecificTTL(resourceType string, multiplier float64) {
	if config, exists := ic.resourceTypes[resourceType]; exists {
		config.StableTTL = time.Duration(float64(config.StableTTL) * multiplier)
		config.TransitionalTTL = time.Duration(float64(config.TransitionalTTL) * multiplier)
	}
}

// WarmCache pre-populates cache with expected resources
func (ic *IntelligentCache) WarmCache(expectedResources []models.ResourceStatus) {
	if !ic.enabled {
		return
	}

	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	for _, resource := range expectedResources {
		// Pre-cache with shorter TTL since these might not be accurate
		ttl := 30 * time.Second

		entry := &CacheEntry{
			Data:        &resource,
			Timestamp:   time.Now(),
			TTL:         ttl,
			AccessCount: 0,
			LastAccess:  time.Now(),
		}

		key := ic.generateResourceKey(resource.Name, resource.Type)
		ic.resourceCache[key] = entry
	}
}

// generateResourceKey creates a unique key for resource caching
func (ic *IntelligentCache) generateResourceKey(name, resourceType string) string {
	return fmt.Sprintf("%s::%s", resourceType, name)
}

// GetMetrics returns current cache metrics (copy without mutex)
func (ic *IntelligentCache) GetMetrics() CacheMetrics {
	ic.metrics.mutex.RLock()
	defer ic.metrics.mutex.RUnlock()

	// Return a copy without the mutex
	return CacheMetrics{
		TotalRequests: ic.metrics.TotalRequests,
		CacheHits:     ic.metrics.CacheHits,
		CacheMisses:   ic.metrics.CacheMisses,
		Invalidations: ic.metrics.Invalidations,
		APICallsSaved: ic.metrics.APICallsSaved,
		LastResetTime: ic.metrics.LastResetTime,
		// Note: mutex field is omitted from the copy
	}
}

// GetCacheHitRate returns the current cache hit rate as a percentage
func (ic *IntelligentCache) GetCacheHitRate() float64 {
	ic.metrics.mutex.RLock()
	defer ic.metrics.mutex.RUnlock()

	if ic.metrics.TotalRequests == 0 {
		return 0.0
	}

	return float64(ic.metrics.CacheHits) / float64(ic.metrics.TotalRequests) * 100.0
}

// GetCacheStatus returns a summary of current cache status
func (ic *IntelligentCache) GetCacheStatus() map[string]interface{} {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	metrics := ic.GetMetrics()

	return map[string]interface{}{
		"enabled":            ic.enabled,
		"resource_entries":   len(ic.resourceCache),
		"deployment_entries": len(ic.deploymentCache),
		"total_entries":      len(ic.resourceCache) + len(ic.deploymentCache),
		"max_entries":        ic.maxEntries,
		"current_phase":      ic.currentPhase,
		"hit_rate":           ic.GetCacheHitRate(),
		"total_requests":     metrics.TotalRequests,
		"cache_hits":         metrics.CacheHits,
		"cache_misses":       metrics.CacheMisses,
		"invalidations":      metrics.Invalidations,
		"api_calls_saved":    metrics.APICallsSaved,
	}
}

// GetCachedResource retrieves a cached resource if valid
func (ic *IntelligentCache) GetCachedResource(resourceID string) (*models.ResourceStatus, bool) {
	if !ic.enabled {
		ic.recordCacheMiss()
		return nil, false
	}

	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	entry, exists := ic.resourceCache[resourceID]
	if !exists {
		ic.recordCacheMiss()
		return nil, false
	}

	// Check if entry is still valid
	if time.Since(entry.Timestamp) > entry.TTL {
		ic.recordCacheMiss()
		return nil, false
	}

	// Update access information
	entry.AccessCount++
	entry.LastAccess = time.Now()

	if resource, ok := entry.Data.(*models.ResourceStatus); ok {
		ic.recordCacheHit()
		return resource, true
	}

	ic.recordCacheMiss()
	return nil, false
}

// CacheResource stores a resource in the cache with appropriate TTL
func (ic *IntelligentCache) CacheResource(resource *models.ResourceStatus) {
	if !ic.enabled {
		return
	}

	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	// Determine TTL based on resource type and state
	ttl := ic.calculateTTL(resource.Type, resource.State)

	// Create cache entry
	entry := &CacheEntry{
		Data:        resource,
		Timestamp:   time.Now(),
		TTL:         ttl,
		AccessCount: 0,
		LastAccess:  time.Now(),
	}

	// Store in cache
	resourceKey := ic.generateResourceKey(resource.Name, resource.Type)
	ic.resourceCache[resourceKey] = entry

	// Enforce cache size limits
	ic.enforceCapacityLimits()
}

// calculateTTL determines the appropriate TTL for a resource based on type and state
func (ic *IntelligentCache) calculateTTL(resourceType, state string) time.Duration {
	config, exists := ic.resourceTypes[resourceType]
	if !exists {
		config = ic.resourceTypes["default"]
	}

	// Use different TTLs based on resource state
	switch state {
	case "Succeeded", "Failed", "Canceled":
		return config.StableTTL
	case "Creating", "Running", "Updating", "InProgress", "Deleting":
		return config.TransitionalTTL
	default:
		// For unknown states, use transitional TTL to be safe
		return config.TransitionalTTL
	}
}

// GetCachedDeployment retrieves cached deployment information
func (ic *IntelligentCache) GetCachedDeployment(resourceGroup, deploymentName string) (interface{}, bool) {
	if !ic.enabled {
		ic.recordCacheMiss()
		return nil, false
	}

	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	key := fmt.Sprintf("deployment::%s::%s", resourceGroup, deploymentName)
	entry, exists := ic.deploymentCache[key]
	if !exists {
		ic.recordCacheMiss()
		return nil, false
	}

	// Check if entry is still valid
	if time.Since(entry.Timestamp) > entry.TTL {
		ic.recordCacheMiss()
		return nil, false
	}

	entry.AccessCount++
	entry.LastAccess = time.Now()
	ic.recordCacheHit()

	return entry.Data, true
}

// CacheDeployment stores deployment information in cache
func (ic *IntelligentCache) CacheDeployment(resourceGroup, deploymentName string, deployment interface{}) {
	if !ic.enabled {
		return
	}

	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	// Deployment info changes less frequently, use longer TTL
	ttl := 2 * time.Minute

	// Adjust TTL based on current phase
	switch ic.currentPhase {
	case PhaseInitialization:
		ttl = 30 * time.Second // Deployment state changes rapidly during init
	case PhaseCompletion, PhaseFinalized:
		ttl = 5 * time.Minute // Very stable during completion
	}

	entry := &CacheEntry{
		Data:        deployment,
		Timestamp:   time.Now(),
		TTL:         ttl,
		AccessCount: 0,
		LastAccess:  time.Now(),
	}

	key := fmt.Sprintf("deployment::%s::%s", resourceGroup, deploymentName)
	ic.deploymentCache[key] = entry

	ic.enforceCapacityLimits()
}

// enforceCapacityLimits removes old entries when cache is full
func (ic *IntelligentCache) enforceCapacityLimits() {
	totalEntries := len(ic.resourceCache) + len(ic.deploymentCache)
	if totalEntries <= ic.maxEntries {
		return
	}

	// Remove oldest, least accessed entries
	type entryInfo struct {
		key        string
		entry      *CacheEntry
		isResource bool
	}

	var allEntries []entryInfo

	for key, entry := range ic.resourceCache {
		allEntries = append(allEntries, entryInfo{key, entry, true})
	}

	for key, entry := range ic.deploymentCache {
		allEntries = append(allEntries, entryInfo{key, entry, false})
	}

	// Sort by last access time (oldest first)
	for i := 0; i < len(allEntries)-1; i++ {
		for j := i + 1; j < len(allEntries); j++ {
			if allEntries[i].entry.LastAccess.After(allEntries[j].entry.LastAccess) {
				allEntries[i], allEntries[j] = allEntries[j], allEntries[i]
			}
		}
	}

	// Remove oldest entries until we're under the limit
	entriesToRemove := totalEntries - ic.maxEntries + 10 // Remove extra for buffer
	for i := 0; i < entriesToRemove && i < len(allEntries); i++ {
		entry := allEntries[i]
		if entry.isResource {
			delete(ic.resourceCache, entry.key)
		} else {
			delete(ic.deploymentCache, entry.key)
		}
	}

	ic.recordInvalidation(InvalidationCapacityLimit)
}

// recordCacheHit increments cache hit metrics
func (ic *IntelligentCache) recordCacheHit() {
	ic.metrics.mutex.Lock()
	defer ic.metrics.mutex.Unlock()

	ic.metrics.TotalRequests++
	ic.metrics.CacheHits++
}

// recordCacheMiss increments cache miss metrics
func (ic *IntelligentCache) recordCacheMiss() {
	ic.metrics.mutex.Lock()
	defer ic.metrics.mutex.Unlock()

	ic.metrics.TotalRequests++
	ic.metrics.CacheMisses++
}

// recordInvalidation increments invalidation metrics
func (ic *IntelligentCache) recordInvalidation(reason CacheInvalidationReason) {
	ic.metrics.mutex.Lock()
	defer ic.metrics.mutex.Unlock()

	ic.metrics.Invalidations++
}

// RecordAPICallSaved increments the API calls saved metric
func (ic *IntelligentCache) RecordAPICallSaved() {
	ic.metrics.mutex.Lock()
	defer ic.metrics.mutex.Unlock()

	ic.metrics.APICallsSaved++
}

// Simple implementations for missing methods to avoid compilation errors
func (dd *DeploymentDisplay) StartInteractiveMode() {
	// Implementation would be added in future interactive display work
}

func (dd *DeploymentDisplay) StopInteractiveMode() {
	// Implementation would be added in future interactive display work
}

// Simple placeholder methods for missing functionality
func (dd *DeploymentDisplay) getDeploymentProvisioningState(resourceGroup, deploymentName string) string {
	queryStart := time.Now()

	// Try to get deployment info from cache first
	if cachedDeployment, found := dd.cache.GetCachedDeployment(resourceGroup, deploymentName); found {
		if deploymentMap, ok := cachedDeployment.(map[string]interface{}); ok {
			if state, ok := deploymentMap["provisioningState"].(string); ok {
				dd.cache.RecordAPICallSaved()
				dd.metricsCollector.RecordAPICall("GetDeployment", "Deployment", time.Since(queryStart), true, "", true)
				return state
			}
		}
	}

	// Cache miss - get deployment info from Azure CLI
	deployment, err := dd.azureCLI.GetDeployment(resourceGroup, deploymentName)
	apiCallDuration := time.Since(queryStart)

	if err != nil {
		dd.metricsCollector.RecordAPICall("GetDeployment", "Deployment", apiCallDuration, false, err.Error(), false)
		return "Unknown"
	}

	dd.metricsCollector.RecordAPICall("GetDeployment", "Deployment", apiCallDuration, true, "", false)

	// Cache the deployment info
	dd.cache.CacheDeployment(resourceGroup, deploymentName, deployment)

	return deployment.ProvisioningState
}

func (dd *DeploymentDisplay) getDeploymentResourceStatus(resourceGroup string) []models.ResourceStatus {
	queryStart := time.Now()
	resourceType := "ResourceGroup" // General resource query

	resources, err := dd.azureCLI.ListResources(resourceGroup)
	if err != nil {
		dd.metricsCollector.RecordAPICall("ListResources", resourceType, time.Since(queryStart), false, err.Error(), false)
		return nil
	}

	dd.metricsCollector.RecordAPICall("ListResources", resourceType, time.Since(queryStart), true, "", false)

	var result []models.ResourceStatus
	cacheHitCount := 0
	cacheMissCount := 0

	for _, res := range resources {
		resourceQueryStart := time.Now()

		// Try to get resource from cache first
		if cachedResource, found := dd.cache.GetCachedResource(res.ID); found {
			result = append(result, *cachedResource)
			dd.cache.RecordAPICallSaved() // Record that we saved an API call
			cacheHitCount++
			dd.metricsCollector.RecordAPICall("GetResource", res.Type, time.Since(resourceQueryStart), true, "", true)
			continue
		}

		// Cache miss - get detailed state from Azure CLI
		cacheMissCount++
		detailedResource, err := dd.azureCLI.GetResource(res.ID)
		apiCallDuration := time.Since(resourceQueryStart)

		provState := "Unknown"
		success := true
		errorMsg := ""

		if err == nil && detailedResource != nil {
			if props, ok := detailedResource.Properties["provisioningState"]; ok {
				if ps, ok := props.(string); ok {
					provState = ps
				}
			}
		} else {
			success = false
			if err != nil {
				errorMsg = err.Error()
			}
		}

		dd.metricsCollector.RecordAPICall("GetResource", res.Type, apiCallDuration, success, errorMsg, false)

		resourceStatus := models.ResourceStatus{
			Name:  res.Name,
			Type:  res.Type,
			State: provState,
		}

		// Cache the resource for future requests
		dd.cache.CacheResource(&resourceStatus)

		result = append(result, resourceStatus)
	}

	// Record overall resource query metrics
	dd.metricsCollector.RecordResourceQuery(resourceType, queryStart, time.Since(queryStart),
		len(result), cacheHitCount, cacheMissCount, true, "")

	return result
}

func (dd *DeploymentDisplay) detectResourceChanges(current []models.ResourceStatus, previous map[string]string) (bool, []ResourceChange) {
	var changes []ResourceChange
	hasChanges := false

	// Check for state changes in existing resources
	for _, resource := range current {
		if previousState, exists := previous[resource.Name]; exists {
			if previousState != resource.State {
				changes = append(changes, ResourceChange{
					Name:      resource.Name,
					Type:      resource.Type,
					FromState: previousState,
					ToState:   resource.State,
					Timestamp: time.Now(),
					Icon:      dd.getResourceStateIcon(resource.State),
				})
				hasChanges = true
			}
		} else {
			// New resource appeared
			changes = append(changes, ResourceChange{
				Name:      resource.Name,
				Type:      resource.Type,
				FromState: "NotFound",
				ToState:   resource.State,
				Timestamp: time.Now(),
				Icon:      dd.getResourceStateIcon(resource.State),
			})
			hasChanges = true
		}
	}

	return hasChanges, changes
}

func (dd *DeploymentDisplay) getResourceStateIcon(state string) string {
	switch state {
	case "Succeeded":
		return "✅"
	case "Failed":
		return "❌"
	case "Canceled":
		return "⚠️"
	case "Creating", "Running", "Updating", "InProgress":
		return "🔄"
	case "Deleting":
		return "🗑️"
	default:
		return "❓"
	}
}

func (dd *DeploymentDisplay) displayInteractiveWaiting() {
	fmt.Print(".")
}

func (dd *DeploymentDisplay) displayWaitingForResources(frames []string, frameIdx *int) {
	fmt.Printf("\r%s Waiting for resources to appear...", frames[*frameIdx])
	*frameIdx = (*frameIdx + 1) % len(frames)
}

func (dd *DeploymentDisplay) displayMinimalInteractiveActivity() {
	fmt.Print(".")
}

func (dd *DeploymentDisplay) displayMinimalActivity(frames []string, frameIdx *int) {
	fmt.Printf("\r%s", frames[*frameIdx])
	*frameIdx = (*frameIdx + 1) % len(frames)
}

func (dd *DeploymentDisplay) displayDeploymentFailure(resourceGroup, deploymentName string, cWarn func(a ...interface{}) string) {
	fmt.Println(cWarn("❌ Deployment failed"))
}

func (dd *DeploymentDisplay) displayDeploymentSuccess(resources []models.ResourceStatus, start time.Time, resourceGroup, deploymentName string) {
	fmt.Printf("✅ Deployment completed successfully in %v\n", time.Since(start))
}

func (dd *DeploymentDisplay) createResourceStateMap(resources []models.ResourceStatus) map[string]string {
	stateMap := make(map[string]string)
	for _, res := range resources {
		stateMap[res.Name] = res.State
	}
	return stateMap
}

// isInteractiveEnvironment checks if we're running in an interactive environment
func (dd *DeploymentDisplay) isInteractiveEnvironment() bool {
	// For now, just return true as a placeholder
	// This could be enhanced to check for CI/CD environments
	return true
}
