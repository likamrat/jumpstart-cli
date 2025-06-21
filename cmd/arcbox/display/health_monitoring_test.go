package display

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestHealthCheckConfig tests the default health check configuration
func TestHealthCheckConfig(t *testing.T) {
	config := DefaultHealthCheckConfig()

	assert.Equal(t, 30*time.Second, config.MaxPollingDelay, "Default max polling delay should be 30 seconds")
	assert.Equal(t, 0.2, config.APIFailureRateThreshold, "Default API failure rate threshold should be 20%")
	assert.Equal(t, 10*time.Minute, config.StateStalenessLimit, "Default staleness limit should be 10 minutes")
	assert.Equal(t, 5*time.Minute, config.ResourceCreationTimeout, "Default resource creation timeout should be 5 minutes")
	assert.Equal(t, 1*time.Minute, config.PerformanceCheckInterval, "Default performance check interval should be 1 minute")
	assert.Equal(t, 10*time.Second, config.PollingLagThreshold, "Default polling lag threshold should be 10 seconds")
	assert.Equal(t, 5*time.Second, config.MonitoringOverheadLimit, "Default monitoring overhead limit should be 5 seconds")
}

// TestMonitoringHealthTracker tests the health tracker initialization and basic functionality
func TestMonitoringHealthTracker(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	assert.NotNil(t, tracker, "Health tracker should be initialized")
	assert.NotNil(t, tracker.MonitoringHealth, "Monitoring health should be initialized")
	assert.NotNil(t, tracker.Config, "Config should be initialized")
	assert.NotNil(t, tracker.ResourceStateHistory, "Resource state history should be initialized")
	assert.NotNil(t, tracker.PerformanceMetrics, "Performance metrics should be initialized")
	assert.Empty(t, tracker.HealthCheckResults, "Health check results should start empty")
}

// TestHealthCheckPollingLag tests polling lag detection
func TestHealthCheckPollingLag(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	now := time.Now()

	// Test no lag case (ExpectedPollingTime not set)
	result := tracker.checkPollingLag(now)
	assert.Nil(t, result, "Should return nil when expected polling time is not set")

	// Test no lag case (on time)
	tracker.ExpectedPollingTime = now.Add(-500 * time.Millisecond) // Less than 1 second, should be ignored
	result = tracker.checkPollingLag(now)
	assert.Nil(t, result, "Should return nil when polling lag is under 1 second")

	// Test warning lag
	tracker.ExpectedPollingTime = now.Add(-15 * time.Second) // 15 seconds behind
	result = tracker.checkPollingLag(now)
	assert.NotNil(t, result, "Should detect polling lag")
	assert.Equal(t, "PollingLag", result.CheckType)
	assert.Equal(t, HealthStatusWarning, result.Status)
	assert.Contains(t, result.Message, "Polling is lagging behind")

	// Test critical lag (add more lag history to trigger average calculation)
	// Clear previous history first
	tracker.PollingLagHistory = []time.Duration{}
	for i := 0; i < 10; i++ {
		tracker.PollingLagHistory = append(tracker.PollingLagHistory, 15*time.Second)
	}
	result = tracker.checkPollingLag(now)
	assert.Equal(t, HealthStatusCritical, result.Status)
	assert.Contains(t, result.Message, "Persistent polling lag detected")
}

// TestHealthCheckAPIHealth tests API health monitoring
func TestHealthCheckAPIHealth(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	now := time.Now()

	// Test no API calls yet
	result := tracker.checkAPIHealth(now)
	assert.Nil(t, result, "Should return nil when no API calls made yet")

	// Test healthy API calls
	tracker.TotalAPICallCount = 100
	tracker.FailedAPICallCount = 5 // 5% failure rate
	result = tracker.checkAPIHealth(now)
	assert.NotNil(t, result)
	assert.Equal(t, "APIHealth", result.CheckType)
	assert.Equal(t, HealthStatusHealthy, result.Status)
	assert.Contains(t, result.Message, "API health is good")

	// Test warning failure rate
	tracker.FailedAPICallCount = 25 // 25% failure rate
	result = tracker.checkAPIHealth(now)
	assert.Equal(t, HealthStatusWarning, result.Status)
	assert.Contains(t, result.Message, "High API failure rate")

	// Test critical failure rate
	tracker.FailedAPICallCount = 60 // 60% failure rate
	result = tracker.checkAPIHealth(now)
	assert.Equal(t, HealthStatusCritical, result.Status)
	assert.Contains(t, result.Message, "Critical API failure rate")
}

// TestHealthCheckStateStaleness tests resource state staleness detection
func TestHealthCheckStateStaleness(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	now := time.Now()

	// Test no resources
	result := tracker.checkStateStaleness(now)
	assert.Nil(t, result, "Should return nil when no resources to check")

	// Add some resources with recent changes
	tracker.ResourceStateHistory["resource1"] = now.Add(-5 * time.Minute)
	tracker.ResourceStateHistory["resource2"] = now.Add(-3 * time.Minute)

	result = tracker.checkStateStaleness(now)
	assert.NotNil(t, result)
	assert.Equal(t, "StateStaleness", result.CheckType)
	assert.Equal(t, HealthStatusHealthy, result.Status)
	assert.Contains(t, result.Message, "All resources showing recent activity")

	// Add stale resource
	tracker.ResourceStateHistory["resource3"] = now.Add(-15 * time.Minute) // Over staleness limit

	result = tracker.checkStateStaleness(now)
	assert.Equal(t, HealthStatusWarning, result.Status)
	assert.Contains(t, result.Message, "Some resources may be stale")

	// Make most resources stale (need >70% to be stale for critical)
	tracker.ResourceStateHistory["resource4"] = now.Add(-20 * time.Minute)
	tracker.ResourceStateHistory["resource5"] = now.Add(-25 * time.Minute) // Now 3/5 = 60%, still not critical

	result = tracker.checkStateStaleness(now)
	assert.Equal(t, HealthStatusWarning, result.Status) // Should still be warning, not critical
	assert.Contains(t, result.Message, "Some resources may be stale")

	// Add another stale resource to reach critical threshold
	tracker.ResourceStateHistory["resource6"] = now.Add(-30 * time.Minute) // Now 4/6 = 67%, still not critical
	tracker.ResourceStateHistory["resource7"] = now.Add(-35 * time.Minute) // Now 5/7 = 71%, should be critical

	result = tracker.checkStateStaleness(now)
	assert.Equal(t, HealthStatusCritical, result.Status)
	assert.Contains(t, result.Message, "Most resources appear stale")
}

// TestHealthCheckResourceDetection tests resource detection monitoring
func TestHealthCheckResourceDetection(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	now := time.Now()

	// Test no expectation set
	result := tracker.checkResourceDetection(now)
	assert.Nil(t, result, "Should return nil when no expectation is set")

	// Test good detection rate
	tracker.ExpectedResourceCount = 10
	tracker.ActualResourceCount = 9 // 90% detection rate

	result = tracker.checkResourceDetection(now)
	assert.NotNil(t, result)
	assert.Equal(t, "ResourceDetection", result.CheckType)
	assert.Equal(t, HealthStatusHealthy, result.Status)
	assert.Contains(t, result.Message, "Resource detection is good")

	// Test warning detection rate
	tracker.ActualResourceCount = 7 // 70% detection rate

	result = tracker.checkResourceDetection(now)
	assert.Equal(t, HealthStatusWarning, result.Status)
	assert.Contains(t, result.Message, "Some expected resources not found")

	// Test critical detection rate
	tracker.ActualResourceCount = 3 // 30% detection rate

	result = tracker.checkResourceDetection(now)
	assert.Equal(t, HealthStatusCritical, result.Status)
	assert.Contains(t, result.Message, "Missing many expected resources")
}

// TestHealthCheckPerformanceMetrics tests performance monitoring
func TestHealthCheckPerformanceMetrics(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	now := time.Now()

	// Test no polls completed yet
	result := tracker.checkPerformanceMetrics(now)
	assert.Nil(t, result, "Should return nil when no polls completed")

	// Test good performance
	tracker.PerformanceMetrics.TotalPolls = 10
	tracker.PerformanceMetrics.AveragePollingDuration = 2 * time.Second

	result = tracker.checkPerformanceMetrics(now)
	assert.NotNil(t, result)
	assert.Equal(t, "Performance", result.CheckType)
	assert.Equal(t, HealthStatusHealthy, result.Status)
	assert.Contains(t, result.Message, "Monitoring performance is good")

	// Test warning performance
	tracker.PerformanceMetrics.AveragePollingDuration = 7 * time.Second // Over limit

	result = tracker.checkPerformanceMetrics(now)
	assert.Equal(t, HealthStatusWarning, result.Status)
	assert.Contains(t, result.Message, "High monitoring overhead")

	// Test critical performance
	tracker.PerformanceMetrics.AveragePollingDuration = 12 * time.Second // Way over limit

	result = tracker.checkPerformanceMetrics(now)
	assert.Equal(t, HealthStatusCritical, result.Status)
	assert.Contains(t, result.Message, "Excessive monitoring overhead")
}

// TestRunHealthChecks tests the comprehensive health check execution
func TestRunHealthChecks(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	// Set up conditions for multiple health checks to trigger
	now := time.Now()
	tracker.ExpectedPollingTime = now.Add(-15 * time.Second) // Polling lag
	tracker.TotalAPICallCount = 100
	tracker.FailedAPICallCount = 30                                        // High failure rate
	tracker.ResourceStateHistory["resource1"] = now.Add(-15 * time.Minute) // Stale
	tracker.ExpectedResourceCount = 10
	tracker.ActualResourceCount = 6 // Low detection rate
	tracker.PerformanceMetrics.TotalPolls = 10
	tracker.PerformanceMetrics.AveragePollingDuration = 8 * time.Second // High overhead

	results := tracker.RunHealthChecks()

	assert.Len(t, results, 5, "Should return 5 health check results")
	assert.NotEmpty(t, tracker.HealthCheckResults, "Results should be stored in tracker")

	// Verify all check types are present
	checkTypes := make(map[string]bool)
	for _, result := range results {
		checkTypes[result.CheckType] = true
	}

	assert.True(t, checkTypes["PollingLag"], "Should include polling lag check")
	assert.True(t, checkTypes["APIHealth"], "Should include API health check")
	assert.True(t, checkTypes["StateStaleness"], "Should include staleness check")
	assert.True(t, checkTypes["ResourceDetection"], "Should include resource detection check")
	assert.True(t, checkTypes["Performance"], "Should include performance check")
}

// TestUpdatePollingMetrics tests polling metrics updates
func TestUpdatePollingMetrics(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	// Test first poll
	tracker.UpdatePollingMetrics(2*time.Second, true, 5)

	metrics := tracker.PerformanceMetrics
	assert.Equal(t, 1, metrics.TotalPolls)
	assert.Equal(t, 1, metrics.SuccessfulPolls)
	assert.Equal(t, 2*time.Second, metrics.AveragePollingDuration)
	assert.Equal(t, 2*time.Second, metrics.MinPollingDuration)
	assert.Equal(t, 2*time.Second, metrics.MaxPollingDuration)
	assert.Equal(t, 5, tracker.ActualResourceCount)

	// Test second poll with different duration
	tracker.UpdatePollingMetrics(4*time.Second, false, 7)

	assert.Equal(t, 2, metrics.TotalPolls)
	assert.Equal(t, 1, metrics.SuccessfulPolls)                    // Still 1 successful
	assert.Equal(t, 3*time.Second, metrics.AveragePollingDuration) // (2+4)/2
	assert.Equal(t, 2*time.Second, metrics.MinPollingDuration)
	assert.Equal(t, 4*time.Second, metrics.MaxPollingDuration)
	assert.Equal(t, 7, tracker.ActualResourceCount)
	assert.Equal(t, 0.5, metrics.MonitoringEfficiency) // 1/2 success rate
}

// TestUpdateResourceStateHistory tests resource state history updates
func TestUpdateResourceStateHistory(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	startTime := time.Now()

	// Test resource change
	tracker.UpdateResourceStateHistory("resource1", true)
	assert.Contains(t, tracker.ResourceStateHistory, "resource1")
	assert.True(t, tracker.ResourceStateHistory["resource1"].After(startTime))

	// Test resource no change
	oldTime := tracker.ResourceStateHistory["resource1"]
	time.Sleep(10 * time.Millisecond) // Small delay to ensure time difference
	tracker.UpdateResourceStateHistory("resource1", false)
	assert.Equal(t, oldTime, tracker.ResourceStateHistory["resource1"], "Time should not change for unchanged resource")
}

// TestSetExpectedResourceCount tests expected resource count setting
func TestSetExpectedResourceCount(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	tracker.SetExpectedResourceCount(15)
	assert.Equal(t, 15, tracker.ExpectedResourceCount)
}

// TestGetHealthSummary tests health summary generation
func TestGetHealthSummary(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	// Test no results yet
	summary := tracker.GetHealthSummary()
	assert.Equal(t, "Health monitoring not yet available", summary)

	// Add some health check results
	now := time.Now()
	tracker.HealthCheckResults = []HealthCheckResult{
		{CheckType: "Test1", Status: HealthStatusHealthy, Timestamp: now},
		{CheckType: "Test2", Status: HealthStatusHealthy, Timestamp: now},
	}

	summary = tracker.GetHealthSummary()
	assert.Equal(t, "HEALTHY: All checks passing", summary)

	// Add warning
	tracker.HealthCheckResults = append(tracker.HealthCheckResults,
		HealthCheckResult{CheckType: "Test3", Status: HealthStatusWarning, Timestamp: now})

	summary = tracker.GetHealthSummary()
	assert.Contains(t, summary, "WARNING: 1 issues detected")

	// Add critical
	tracker.HealthCheckResults = append(tracker.HealthCheckResults,
		HealthCheckResult{CheckType: "Test4", Status: HealthStatusCritical, Timestamp: now})

	summary = tracker.GetHealthSummary()
	assert.Contains(t, summary, "CRITICAL: 1 issues detected")
}

// TestShouldAdjustStrategy tests strategy adjustment recommendations
func TestShouldAdjustStrategy(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	now := time.Now()

	// Test no adjustments needed
	shouldAdjust, reason := tracker.ShouldAdjustStrategy()
	assert.False(t, shouldAdjust)
	assert.Empty(t, reason)

	// Test multiple critical issues
	tracker.HealthCheckResults = []HealthCheckResult{
		{CheckType: "Test1", Status: HealthStatusCritical, Timestamp: now},
		{CheckType: "Test2", Status: HealthStatusCritical, Timestamp: now},
	}

	shouldAdjust, reason = tracker.ShouldAdjustStrategy()
	assert.True(t, shouldAdjust)
	assert.Contains(t, reason, "degraded monitoring mode")

	// Test multiple warnings
	tracker.HealthCheckResults = []HealthCheckResult{
		{CheckType: "Test1", Status: HealthStatusWarning, Timestamp: now},
		{CheckType: "Test2", Status: HealthStatusWarning, Timestamp: now},
		{CheckType: "Test3", Status: HealthStatusWarning, Timestamp: now},
	}

	shouldAdjust, reason = tracker.ShouldAdjustStrategy()
	assert.True(t, shouldAdjust)
	assert.Contains(t, reason, "Reduce polling frequency")

	// Test suggestion for increased monitoring
	tracker.HealthCheckResults = []HealthCheckResult{
		{CheckType: "Test1", Status: HealthStatusHealthy, Timestamp: now},
		{CheckType: "Test2", Status: HealthStatusHealthy, Timestamp: now},
		{CheckType: "Test3", Status: HealthStatusHealthy, Timestamp: now},
		{CheckType: "Test4", Status: HealthStatusHealthy, Timestamp: now},
		{CheckType: "Test5", Status: HealthStatusHealthy, Timestamp: now},
	}

	shouldAdjust, reason = tracker.ShouldAdjustStrategy()
	assert.True(t, shouldAdjust)
	assert.Contains(t, reason, "increasing monitoring frequency")
}

// TestCalculateAveragePollingLag tests polling lag average calculation
func TestCalculateAveragePollingLag(t *testing.T) {
	erm := NewErrorRecoveryManager()
	tracker := erm.HealthTracker

	// Test empty history
	avgLag := tracker.calculateAveragePollingLag()
	assert.Equal(t, time.Duration(0), avgLag)

	// Test with lag history
	tracker.PollingLagHistory = []time.Duration{
		2 * time.Second,
		4 * time.Second,
		6 * time.Second,
	}

	avgLag = tracker.calculateAveragePollingLag()
	assert.Equal(t, 4*time.Second, avgLag) // (2+4+6)/3
}
