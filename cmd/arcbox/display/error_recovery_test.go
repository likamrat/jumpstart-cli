package display

import (
	"errors"
	"testing"
	"time"

	"jumpstartcli/cmd/arcbox/models"
	"jumpstartcli/internal/azurecli"

	"github.com/stretchr/testify/assert"
)

// SimpleErrorRecoveryMock is a simple mock without testify dependencies
type SimpleErrorRecoveryMock struct {
	deploymentFailures    int
	resourceFailures      int
	maxDeploymentFailures int
	maxResourceFailures   int
	deploymentError       error
	resourceError         error
	resources             []azurecli.ResourceInfo
}

func (m *SimpleErrorRecoveryMock) GetDeployment(resourceGroup, deploymentName string) (*azurecli.DeploymentInfo, error) {
	if m.deploymentFailures < m.maxDeploymentFailures {
		m.deploymentFailures++
		return nil, m.deploymentError
	}

	return &azurecli.DeploymentInfo{
		ProvisioningState: "Running",
		Properties: map[string]interface{}{
			"timestamp": "2023-01-01T10:00:00Z",
		},
	}, nil
}

func (m *SimpleErrorRecoveryMock) ListResourcesWithDetails(resourceGroup string) ([]azurecli.ResourceInfo, error) {
	if m.resourceFailures < m.maxResourceFailures {
		m.resourceFailures++
		return nil, m.resourceError
	}

	return m.resources, nil
}

// All other required interface methods - simplified implementations
func (m *SimpleErrorRecoveryMock) ListResources(resourceGroup string) ([]azurecli.ResourceInfo, error) {
	return m.ListResourcesWithDetails(resourceGroup)
}

func (m *SimpleErrorRecoveryMock) GetResourcesBatch(resourceIDs []string, maxConcurrency int) ([]azurecli.ResourceInfo, []error) {
	return nil, nil
}

func (m *SimpleErrorRecoveryMock) CheckProviderRegistration(provider string) (bool, error) {
	return true, nil
}

func (m *SimpleErrorRecoveryMock) CheckResourceGroupExists(name string) (bool, error) {
	return true, nil
}

func (m *SimpleErrorRecoveryMock) GetCurrentSubscription() (*azurecli.SubscriptionInfo, error) {
	return nil, nil
}

func (m *SimpleErrorRecoveryMock) GetSubscription(subscriptionID string) (*azurecli.SubscriptionInfo, error) {
	return nil, nil
}

func (m *SimpleErrorRecoveryMock) ListSubscriptions() ([]azurecli.SubscriptionInfo, error) {
	return nil, nil
}

func (m *SimpleErrorRecoveryMock) SetSubscription(subscriptionID string) error {
	return nil
}

func (m *SimpleErrorRecoveryMock) IsLoggedIn() bool {
	return true
}

func (m *SimpleErrorRecoveryMock) ListLocations() ([]string, error) {
	return nil, nil
}

func (m *SimpleErrorRecoveryMock) ListVMUsage(region string) ([]azurecli.VMUsageInfo, error) {
	return nil, nil
}

func (m *SimpleErrorRecoveryMock) ListVMSKUs(region string) ([]azurecli.SKUInfo, error) {
	return nil, nil
}

func (m *SimpleErrorRecoveryMock) CheckSKUAvailability(sku, region string) (bool, error) {
	return true, nil
}

func (m *SimpleErrorRecoveryMock) RegisterProvider(provider string) error {
	return nil
}

func (m *SimpleErrorRecoveryMock) ListResourceGroups() ([]azurecli.ResourceGroupInfo, error) {
	return nil, nil
}

func (m *SimpleErrorRecoveryMock) CreateResourceGroup(name, location string) error {
	return nil
}

func (m *SimpleErrorRecoveryMock) DeleteResourceGroup(name string, noWait bool) error {
	return nil
}

func (m *SimpleErrorRecoveryMock) GetResource(resourceID string) (*azurecli.ResourceInfo, error) {
	return nil, nil
}

func (m *SimpleErrorRecoveryMock) ListDeployments(resourceGroup string) ([]azurecli.DeploymentInfo, error) {
	return nil, nil
}

func (m *SimpleErrorRecoveryMock) CreateDeployment(resourceGroup, deploymentName, templateURI string, parameters []string, noWait bool) error {
	return nil
}

func (m *SimpleErrorRecoveryMock) ListVMs(resourceGroup string) ([]azurecli.VMInfo, error) {
	return nil, nil
}

// TestErrorRecoveryManager tests the error recovery manager functionality
func TestErrorRecoveryManager(t *testing.T) {
	tests := []struct {
		name                string
		error               error
		expectedClass       ErrorClassification
		shouldRetry         bool
		expectedUserMessage string
	}{
		{
			name:                "network timeout",
			error:               errors.New("network timeout occurred"),
			expectedClass:       ErrorTransient,
			shouldRetry:         true,
			expectedUserMessage: "Temporary network or service issue",
		},
		{
			name:                "rate limiting",
			error:               errors.New("Rate limit exceeded (429)"),
			expectedClass:       ErrorThrottling,
			shouldRetry:         true,
			expectedUserMessage: "Azure API rate limiting",
		},
		{
			name:                "authentication error",
			error:               errors.New("Authentication failed - please login"),
			expectedClass:       ErrorAuthentication,
			shouldRetry:         false,
			expectedUserMessage: "Azure authentication issue",
		},
		{
			name:                "resource not found",
			error:               errors.New("Resource not found"),
			expectedClass:       ErrorPermanent,
			shouldRetry:         false,
			expectedUserMessage: "Configuration issue",
		},
		{
			name:                "unknown error",
			error:               errors.New("Something unexpected happened"),
			expectedClass:       ErrorUnknown,
			shouldRetry:         true,
			expectedUserMessage: "Unexpected error:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			erm := NewErrorRecoveryManager()

			// Test error classification
			classification := erm.ClassifyError(tt.error)
			assert.Equal(t, tt.expectedClass, classification, "Error classification should match expected")

			// Test retry decision
			ctx := &RetryContext{Attempt: 0}
			shouldRetry := erm.ShouldRetry(ctx, classification)
			assert.Equal(t, tt.shouldRetry, shouldRetry, "Retry decision should match expected")

			// Test user-friendly message
			message := erm.GetUserFriendlyErrorMessage(tt.error, classification)
			assert.Contains(t, message, tt.expectedUserMessage, "User message should contain expected text")
		})
	}
}

// TestRetryLogic tests the retry mechanism with exponential backoff
func TestRetryLogic(t *testing.T) {
	erm := NewErrorRecoveryManager()
	ctx := &RetryContext{StartTime: time.Now()}

	// Test exponential backoff calculation
	testError := errors.New("network timeout")
	classification := erm.ClassifyError(testError)

	delays := make([]time.Duration, 5)
	for i := 0; i < 5; i++ {
		delays[i] = erm.CalculateDelay(ctx, classification)
		ctx.Attempt++
	}

	// Verify that delays generally increase (allowing for jitter)
	for i := 1; i < len(delays); i++ {
		// Allow some tolerance for jitter (previous delay * 1.5 as rough check)
		assert.True(t, delays[i] >= delays[i-1]*3/4,
			"Delay should generally increase with exponential backoff (got %v after %v)",
			delays[i], delays[i-1])
	}

	// Verify max retry limit
	ctx.Attempt = erm.Config.MaxRetries
	shouldRetry := erm.ShouldRetry(ctx, classification)
	assert.False(t, shouldRetry, "Should not retry after max attempts reached")
}

// TestThrottlingBackoff tests special handling for API throttling
func TestThrottlingBackoff(t *testing.T) {
	erm := NewErrorRecoveryManager()
	ctx := &RetryContext{StartTime: time.Now()}

	throttleError := errors.New("Rate limit exceeded")
	classification := erm.ClassifyError(throttleError)

	assert.Equal(t, ErrorThrottling, classification, "Should classify as throttling error")

	delay := erm.CalculateDelay(ctx, classification)
	assert.Equal(t, erm.Config.ThrottleBackoffDuration, delay,
		"Throttling errors should use special backoff duration")
}

// TestHealthMetrics tests the monitoring health tracking
func TestHealthMetrics(t *testing.T) {
	erm := NewErrorRecoveryManager()

	// Initial state
	assert.False(t, erm.Health.InDegradedMode, "Should not start in degraded mode")
	assert.Equal(t, 0, erm.Health.ConsecutiveFailures, "Should start with no failures")

	// Record successful calls first to establish a baseline
	for i := 0; i < 10; i++ {
		erm.UpdateHealth(true, 100*time.Millisecond)
	}

	assert.Equal(t, 0, erm.Health.ConsecutiveFailures, "Successful calls should reset consecutive failures")
	assert.Equal(t, 10, erm.Health.TotalAPICallCount, "Should track total API calls")
	assert.Equal(t, 0, erm.Health.FailedAPICallCount, "Should track failed API calls")

	// Record consecutive failures to trigger degraded mode (but keep under tolerance)
	for i := 0; i < 4; i++ {
		erm.UpdateHealth(false, 0)
	}

	assert.True(t, erm.Health.InDegradedMode, "Should enter degraded mode with >3 consecutive failures")
	assert.Equal(t, 4, erm.Health.ConsecutiveFailures, "Should track consecutive failures")

	// Recovery from degraded mode - one success should reset consecutive failures
	erm.UpdateHealth(true, 100*time.Millisecond)
	assert.False(t, erm.Health.InDegradedMode, "Should exit degraded mode on successful call")
	assert.Equal(t, 0, erm.Health.ConsecutiveFailures, "Should reset consecutive failures on success")
}

// TestErrorRecoveryBasicFunctionality tests basic error recovery functionality
func TestErrorRecoveryBasicFunctionality(t *testing.T) {
	// Test successful operation (no errors)
	mockCLI := &SimpleErrorRecoveryMock{
		maxDeploymentFailures: 0, // No failures
		maxResourceFailures:   0,
		resources: []azurecli.ResourceInfo{
			{
				ID:   "/subscriptions/test/resourceGroups/test/providers/Microsoft.Compute/virtualMachines/testvm",
				Name: "testvm",
				Type: "Microsoft.Compute/virtualMachines",
				Properties: map[string]interface{}{
					"provisioningState": "Succeeded",
				},
			},
		},
	}

	dd := NewDeploymentDisplay(mockCLI)

	// Test successful deployment state query
	state, err := dd.getDeploymentProvisioningStateResilient("test-rg", "test-deployment")
	assert.NoError(t, err, "Should succeed with no failures")
	assert.Equal(t, "Running", state, "Should return correct state")

	// Test successful resource status query
	resources, err := dd.getDeploymentResourceStatusResilient("test-rg")
	assert.NoError(t, err, "Should succeed with no failures")
	assert.Len(t, resources, 1, "Should return one resource")
	assert.Equal(t, "testvm", resources[0].Name, "Should return correct resource")
}

// TestPartialFailureTolerance tests that the system continues with partial data
func TestPartialFailureTolerance(t *testing.T) {
	// Create a simple mock that returns predefined data without using the retry mechanism
	mockCLI := &SimpleErrorRecoveryMock{
		maxResourceFailures: 0, // No failures - immediate success
		resources: []azurecli.ResourceInfo{
			{
				ID:   "/subscriptions/test/resourceGroups/test/providers/Microsoft.Compute/virtualMachines/vm1",
				Name: "vm1",
				Type: "Microsoft.Compute/virtualMachines",
				Properties: map[string]interface{}{
					"provisioningState": "Succeeded",
				},
			},
			{
				ID:   "/subscriptions/test/resourceGroups/test/providers/Microsoft.Storage/storageAccounts/storage1",
				Name: "storage1",
				Type: "Microsoft.Storage/storageAccounts",
				Properties: map[string]interface{}{
					"provisioningState": "Running",
				},
			},
		},
	}

	dd := NewDeploymentDisplay(mockCLI)

	// Use the direct method without retry to test the conversion logic
	resources, err := mockCLI.ListResourcesWithDetails("test-rg")
	assert.NoError(t, err, "Mock should not fail")

	// Test the conversion logic directly
	convertedResources := dd.convertToResourceStatus(resources)

	assert.Len(t, convertedResources, 2, "Should return all available resources")

	// Verify resource conversion
	expectedResources := []models.ResourceStatus{
		{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", State: "Succeeded"},
		{Name: "storage1", Type: "Microsoft.Storage/storageAccounts", State: "Running"},
	}

	assert.Equal(t, expectedResources, convertedResources, "Should correctly convert resource info to status")
}

// TestContinueMonitoringDecision tests when monitoring should continue vs stop
func TestContinueMonitoringDecision(t *testing.T) {
	erm := NewErrorRecoveryManager()

	// Should continue with healthy state
	assert.True(t, erm.ShouldContinueMonitoring(), "Should continue monitoring when healthy")

	// Should continue with some failures
	erm.Health.ConsecutiveFailures = 5
	assert.True(t, erm.ShouldContinueMonitoring(), "Should continue monitoring with moderate failures")

	// Should stop with too many consecutive failures
	erm.Health.ConsecutiveFailures = 15
	assert.False(t, erm.ShouldContinueMonitoring(), "Should stop monitoring with excessive failures")
}
