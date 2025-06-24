package utils

import (
	"errors"
	"jumpstartcli/internal/azurecli"
	"testing"
)

func TestNewDeploymentStateDetector(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	mockExitFunc := func(code int) {} // Mock exit function for testing

	detector := NewDeploymentStateDetector(mockCLI, mockExitFunc, true)

	if detector == nil {
		t.Fatal("NewDeploymentStateDetector returned nil")
	}

	if detector.azureCLI != mockCLI {
		t.Error("DeploymentStateDetector should store the provided Azure CLI instance")
	}

	if detector.exitFunc == nil {
		t.Error("DeploymentStateDetector should store the provided exit function")
	}

	if !detector.debug {
		t.Error("DeploymentStateDetector should have debug mode enabled")
	}
}

func TestShouldValidateState(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	detector := NewDeploymentStateDetector(mockCLI, nil, false)

	tests := []struct {
		state    string
		expected bool
	}{
		{"Succeeded", false},
		{"Running", false},
		{"Creating", false},
		{"Failed", true},
		{"Canceled", true},
		{"Unknown", true},
		{"NotFound", true},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			result := detector.shouldValidateState(tt.state)
			if result != tt.expected {
				t.Errorf("shouldValidateState(%s) = %v, expected %v", tt.state, result, tt.expected)
			}
		})
	}
}

func TestIsTransientFailure(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	detector := NewDeploymentStateDetector(mockCLI, nil, false)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"timeout error", errors.New("connection timeout"), true},
		{"network error", errors.New("network unreachable"), true},
		{"throttling error", errors.New("rate limit exceeded"), true},
		{"permanent error", errors.New("resource not found"), false},
		{"authentication error", errors.New("authentication failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.IsTransientFailure(tt.err)
			if result != tt.expected {
				t.Errorf("IsTransientFailure(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsCriticalError(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	detector := NewDeploymentStateDetector(mockCLI, nil, false)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"authentication error", errors.New("authentication failed"), true},
		{"access denied", errors.New("access denied"), true},
		{"subscription not found", errors.New("subscription not found"), true},
		{"timeout error", errors.New("timeout"), false},
		{"general error", errors.New("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isCriticalError(tt.err)
			if result != tt.expected {
				t.Errorf("isCriticalError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestExtractProvisioningState(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	detector := NewDeploymentStateDetector(mockCLI, nil, false)

	tests := []struct {
		name     string
		resource *azurecli.ResourceInfo
		expected string
	}{
		{
			name:     "nil resource",
			resource: nil,
			expected: "",
		},
		{
			name: "resource with direct provisioningState",
			resource: &azurecli.ResourceInfo{
				Properties: map[string]interface{}{
					"provisioningState": "Succeeded",
				},
			},
			expected: "Succeeded",
		},
		{
			name: "resource with nested provisioningState",
			resource: &azurecli.ResourceInfo{
				Properties: map[string]interface{}{
					"properties": map[string]interface{}{
						"provisioningState": "Failed",
					},
				},
			},
			expected: "Failed",
		},
		{
			name: "resource without provisioningState",
			resource: &azurecli.ResourceInfo{
				Properties: map[string]interface{}{
					"location": "eastus",
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractProvisioningState(tt.resource)
			if result != tt.expected {
				t.Errorf("extractProvisioningState() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestValidateResourceState_NoValidationNeeded(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	detector := NewDeploymentStateDetector(mockCLI, nil, false)

	result := detector.ValidateResourceState("test-vm", "/subscriptions/test", "Succeeded")

	if !result.IsStateAccurate {
		t.Error("State should be considered accurate for Succeeded state")
	}

	if result.ResourceName != "test-vm" {
		t.Errorf("ResourceName = %s, expected test-vm", result.ResourceName)
	}

	if result.ReportedState != "Succeeded" {
		t.Errorf("ReportedState = %s, expected Succeeded", result.ReportedState)
	}

	if result.ValidatedState != "Succeeded" {
		t.Errorf("ValidatedState = %s, expected Succeeded", result.ValidatedState)
	}
}

func TestNewDeploymentMonitor(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	mockExitFunc := func(code int) {}

	monitor := NewDeploymentMonitor(mockCLI, mockExitFunc, true, true)

	if monitor == nil {
		t.Fatal("NewDeploymentMonitor returned nil")
	}

	if monitor.azureCLI != mockCLI {
		t.Error("DeploymentMonitor should store the provided Azure CLI instance")
	}

	if !monitor.debug {
		t.Error("DeploymentMonitor should have debug mode enabled")
	}

	if !monitor.verbose {
		t.Error("DeploymentMonitor should have verbose mode enabled")
	}
}

func TestDetectPotentialFalseFailure(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	monitor := NewDeploymentMonitor(mockCLI, nil, false, false)

	tests := []struct {
		name     string
		status   *DeploymentStatus
		expected bool
	}{
		{
			name: "deployment failed but no failed resources",
			status: &DeploymentStatus{
				State:              "Failed",
				FailedResources:    0,
				CompletedResources: 3,
			},
			expected: true,
		},
		{
			name: "deployment failed with failed resources",
			status: &DeploymentStatus{
				State:              "Failed",
				FailedResources:    2,
				CompletedResources: 1,
			},
			expected: false,
		},
		{
			name: "resource-level false failure",
			status: &DeploymentStatus{
				State: "Succeeded",
				Resources: []ResourceStatus{
					{ReportedState: "Failed", ValidatedState: "Succeeded"},
					{ReportedState: "Succeeded", ValidatedState: "Succeeded"},
				},
			},
			expected: true,
		},
		{
			name: "normal successful deployment",
			status: &DeploymentStatus{
				State:              "Succeeded",
				FailedResources:    0,
				CompletedResources: 3,
				Resources: []ResourceStatus{
					{ReportedState: "Succeeded", ValidatedState: "Succeeded"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monitor.DetectPotentialFalseFailure(tt.status)
			if result != tt.expected {
				t.Errorf("DetectPotentialFalseFailure() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsDeploymentComplete(t *testing.T) {
	mockCLI := azurecli.NewMockAzureCLI()
	monitor := NewDeploymentMonitor(mockCLI, nil, false, false)

	tests := []struct {
		name               string
		status             *DeploymentStatus
		expectedComplete   bool
		expectedSuccessful bool
	}{
		{
			name: "successful deployment",
			status: &DeploymentStatus{
				State:               "Succeeded",
				InProgressResources: 0,
				FailedResources:     0,
			},
			expectedComplete:   true,
			expectedSuccessful: true,
		},
		{
			name: "failed deployment",
			status: &DeploymentStatus{
				State:               "Failed",
				InProgressResources: 0,
				FailedResources:     2,
			},
			expectedComplete:   true,
			expectedSuccessful: false,
		},
		{
			name: "in progress deployment",
			status: &DeploymentStatus{
				State:               "Running",
				InProgressResources: 3,
				FailedResources:     0,
			},
			expectedComplete:   false,
			expectedSuccessful: false,
		},
		{
			name: "succeeded but resources still in progress",
			status: &DeploymentStatus{
				State:               "Succeeded",
				InProgressResources: 1,
				FailedResources:     0,
			},
			expectedComplete:   true,
			expectedSuccessful: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complete, successful := monitor.IsDeploymentComplete(tt.status)
			if complete != tt.expectedComplete {
				t.Errorf("IsDeploymentComplete() complete = %v, expected %v", complete, tt.expectedComplete)
			}
			if successful != tt.expectedSuccessful {
				t.Errorf("IsDeploymentComplete() successful = %v, expected %v", successful, tt.expectedSuccessful)
			}
		})
	}
}
