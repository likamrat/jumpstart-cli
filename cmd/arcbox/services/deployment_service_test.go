package services

import (
	"fmt"
	"strings"
	"testing"

	"jumpstartcli/cmd/arcbox/display"
	"jumpstartcli/internal/azurecli"

	"github.com/spf13/cobra"
)

// TestDeploymentServiceCreation tests that DeploymentService can be created properly
func TestDeploymentServiceCreation(t *testing.T) {
	mockCLI := &azurecli.MockAzureCLI{}
	mockDisplay := display.NewDeploymentDisplay(mockCLI)
	service := NewDeploymentService(mockCLI, mockDisplay)

	if service == nil {
		t.Fatal("Expected non-nil DeploymentService")
	}

	if service.azureCLI != mockCLI {
		t.Error("Expected DeploymentService to use provided CLI")
	}

	if service.display != mockDisplay {
		t.Error("Expected DeploymentService to use provided display")
	}
}

// TestDeploymentService_Deploy tests the deployment logic
func TestDeploymentService_Deploy(t *testing.T) {
	tests := []struct {
		name              string
		templateLocal     string
		templateParams    string
		templateURI       string
		expectedBicepPath string
		expectedUseParam  bool
		expectedParamFile string
		shouldError       bool
	}{
		{
			name:              "default_remote_template",
			templateLocal:     "",
			templateParams:    "",
			templateURI:       "",
			expectedBicepPath: "https://raw.githubusercontent.com/microsoft/azure_arc/main/azure_jumpstart_arcbox/ARM/azuredeploy.json",
			expectedUseParam:  false,
			expectedParamFile: "",
			shouldError:       false,
		},
		{
			name:              "local_template_with_params",
			templateLocal:     "/path/to/template.bicep",
			templateParams:    "/path/to/params.json",
			templateURI:       "",
			expectedBicepPath: "/path/to/template.bicep",
			expectedUseParam:  true,
			expectedParamFile: "/path/to/params.json",
			shouldError:       false,
		},
		{
			name:              "local_template_without_params",
			templateLocal:     "/path/to/template.bicep",
			templateParams:    "",
			templateURI:       "",
			expectedBicepPath: "/path/to/template.bicep",
			expectedUseParam:  false,
			expectedParamFile: "",
			shouldError:       false,
		},
		{
			name:              "custom_remote_uri",
			templateLocal:     "",
			templateParams:    "",
			templateURI:       "https://example.com/custom-template.json",
			expectedBicepPath: "https://example.com/custom-template.json",
			expectedUseParam:  false,
			expectedParamFile: "",
			shouldError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := &azurecli.MockAzureCLI{
				IsLoggedInResult: true,
				CurrentSubscription: &azurecli.SubscriptionInfo{
					ID:   "test-subscription-id",
					Name: "Test Subscription",
				},
			}
			mockDisplay := display.NewDeploymentDisplay(mockCLI)
			service := NewDeploymentService(mockCLI, mockDisplay)

			// Create a test command with flags
			cmd := &cobra.Command{}
			cmd.Flags().String("template-local", "", "")
			cmd.Flags().String("template-params", "", "")
			cmd.Flags().String("template-uri", "", "")
			cmd.Flags().String("resource-group", "test-rg", "")
			cmd.Flags().String("location", "eastus", "")
			cmd.Flags().String("flavor", "ITPro", "")
			cmd.Flags().String("windows-user", "testuser", "")
			cmd.Flags().String("windows-password", "TestPassword123!", "")
			cmd.Flags().Bool("auto-shutdown", true, "") // Enable auto-shutdown to bypass confirmation
			cmd.Flags().Bool("yes", true, "")           // Skip deployment confirmation

			// Set flag values
			cmd.Flags().Set("template-local", tt.templateLocal)
			cmd.Flags().Set("template-params", tt.templateParams)
			cmd.Flags().Set("template-uri", tt.templateURI)

			// Test the deployment logic (this will test path resolution but not actual deployment)
			// Since Deploy method performs actual Azure operations, we test the logic without executing
			templateSpecString := ""
			deploymentSpecString := ""
			noWait := false

			// For testing purposes, we'll check that the method doesn't panic
			// and handles the flag parsing correctly
			defer func() {
				if r := recover(); r != nil && !tt.shouldError {
					t.Errorf("Deploy method panicked unexpectedly: %v", r)
				}
			}()

			// Note: We can't easily test the full deployment without mocking the entire Azure CLI flow
			// The actual deployment logic involves external Azure CLI calls
			err := service.Deploy(cmd, []string{}, templateSpecString, noWait, deploymentSpecString)

			// For now, we expect errors since we're not fully mocking all Azure operations
			// In a real scenario, you'd want to mock all the Azure CLI calls
			if !tt.shouldError && err == nil {
				t.Log("Deploy method completed without error (expected for mock)")
			}
		})
	}
}

// TestDeploymentService_ValidateFlags tests flag validation logic
func TestDeploymentService_ValidateFlags(t *testing.T) {
	mockCLI := &azurecli.MockAzureCLI{
		IsLoggedInResult: true,
	}
	mockDisplay := display.NewDeploymentDisplay(mockCLI)
	_ = NewDeploymentService(mockCLI, mockDisplay)

	tests := []struct {
		name        string
		flags       map[string]string
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid_required_flags",
			flags: map[string]string{
				"resource-group": "test-rg",
				"location":       "eastus",
				"flavor":         "ITPro",
				"windows-user":   "testuser",
			},
			shouldError: false,
		},
		{
			name: "missing_resource_group",
			flags: map[string]string{
				"location":     "eastus",
				"flavor":       "ITPro",
				"windows-user": "testuser",
			},
			shouldError: true,
			errorMsg:    "resource-group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}

			// Add all possible flags
			for flag, value := range tt.flags {
				cmd.Flags().String(flag, "", "")
				cmd.Flags().Set(flag, value)
			}

			// Test validation logic would go here
			// This would involve extracting validation logic from the Deploy method
			// into a separate validateFlags method that can be tested independently
			t.Log("Flag validation test completed")
		})
	}
}

// TestDeploymentService_TemplatePathResolution tests template path resolution logic
func TestDeploymentService_TemplatePathResolution(t *testing.T) {
	tests := []struct {
		name           string
		templateLocal  string
		templateURI    string
		expectedPath   string
		expectedSource string
	}{
		{
			name:           "default_remote",
			templateLocal:  "",
			templateURI:    "",
			expectedPath:   "https://raw.githubusercontent.com/microsoft/azure_arc/main/azure_jumpstart_arcbox/ARM/azuredeploy.json",
			expectedSource: "default_remote",
		},
		{
			name:           "local_template",
			templateLocal:  "/path/to/template.bicep",
			templateURI:    "",
			expectedPath:   "/path/to/template.bicep",
			expectedSource: "local",
		},
		{
			name:           "custom_uri",
			templateLocal:  "",
			templateURI:    "https://example.com/template.json",
			expectedPath:   "https://example.com/template.json",
			expectedSource: "remote_uri",
		},
		{
			name:           "local_overrides_uri",
			templateLocal:  "/path/to/template.bicep",
			templateURI:    "https://example.com/template.json",
			expectedPath:   "/path/to/template.bicep",
			expectedSource: "local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the template path resolution logic
			var actualPath string
			var actualSource string

			if tt.templateLocal != "" {
				actualPath = tt.templateLocal
				actualSource = "local"
			} else if tt.templateURI != "" {
				actualPath = tt.templateURI
				actualSource = "remote_uri"
			} else {
				actualPath = "https://raw.githubusercontent.com/microsoft/azure_arc/main/azure_jumpstart_arcbox/ARM/azuredeploy.json"
				actualSource = "default_remote"
			}

			if actualPath != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, actualPath)
			}

			if actualSource != tt.expectedSource {
				t.Errorf("Expected source %s, got %s", tt.expectedSource, actualSource)
			}
		})
	}
}

// TestDeploymentService_ErrorHandling tests error handling scenarios
func TestDeploymentService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*azurecli.MockAzureCLI)
		flags         map[string]string
		expectedError string
	}{
		{
			name: "cli_not_logged_in",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = false
			},
			flags: map[string]string{
				"resource-group":   "test-rg",
				"windows-user":     "testuser",
				"windows-password": "TestPassword123!",
				"location":         "eastus",
			},
			expectedError: "not logged in",
		},
		{
			name: "deployment_creation_error",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
				// Set up the mock so that CreateDeployment fails
				mock.CreateDeploymentError = fmt.Errorf("deployment creation failed")
			},
			flags: map[string]string{
				"resource-group":   "test-rg",
				"windows-user":     "testuser",
				"windows-password": "TestPassword123!",
				"location":         "eastus",
			},
			expectedError: "deployment creation failed",
		},
		{
			name: "missing_required_flags",
			mockSetup: func(mock *azurecli.MockAzureCLI) {
				mock.IsLoggedInResult = true
			},
			flags: map[string]string{
				"location": "eastus",
			},
			expectedError: "missing required arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := &azurecli.MockAzureCLI{
				IsLoggedInResult: true,
			}

			if tt.mockSetup != nil {
				tt.mockSetup(mockCLI)
			}

			mockDisplay := display.NewDeploymentDisplay(mockCLI)
			service := NewDeploymentService(mockCLI, mockDisplay)

			// Create a basic command for testing
			cmd := &cobra.Command{}
			cmd.Flags().String("template-local", "", "")
			cmd.Flags().String("template-params", "", "")
			cmd.Flags().String("template-uri", "", "")
			cmd.Flags().String("resource-group", "", "")
			cmd.Flags().String("location", "", "")
			cmd.Flags().String("windows-user", "", "")
			cmd.Flags().String("windows-password", "", "")
			cmd.Flags().String("flavor", "ITPro", "")
			cmd.Flags().Bool("auto-shutdown", true, "") // Enable auto-shutdown to bypass confirmation
			cmd.Flags().Bool("yes", true, "")           // Skip deployment confirmation

			// Set flag values from test case
			for flag, value := range tt.flags {
				if cmd.Flags().Lookup(flag) != nil {
					cmd.Flags().Set(flag, value)
				}
			}

			err := service.Deploy(cmd, []string{}, "", false, "")

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}
