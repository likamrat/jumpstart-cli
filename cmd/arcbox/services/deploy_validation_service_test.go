package services

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDeployValidationService_ValidateDeployFlags(t *testing.T) {
	service := NewDeployValidationService()

	tests := []struct {
		name          string
		setupCmd      func() *cobra.Command
		expectedValid bool
		expectedError string
	}{
		{
			name: "Valid flags",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("location", "", "Location")
				cmd.Flags().String("resource-group", "", "Resource group")
				cmd.Flags().String("windows-user", "", "Windows user")
				cmd.Flags().String("flavor", "", "Flavor")

				// Set valid values
				cmd.Flags().Set("location", "eastus")
				cmd.Flags().Set("resource-group", "test-rg")
				cmd.Flags().Set("windows-user", "testuser")
				cmd.Flags().Set("flavor", "ITPro")

				return cmd
			},
			expectedValid: true,
		},
		{
			name: "No flags set",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("location", "", "Location")
				cmd.Flags().String("resource-group", "", "Resource group")
				cmd.Flags().String("windows-user", "", "Windows user")
				cmd.Flags().String("flavor", "", "Flavor")
				return cmd
			},
			expectedValid: true, // ValidateAllFlags only validates changed flags
		},
		{
			name: "Invalid boolean-like string flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("auto-shutdown", "", "Enable automatic shutdown (yes/no)")
				cmd.Flags().Set("auto-shutdown", "invalid-value")
				return cmd
			},
			expectedValid: false,
			expectedError: "flag validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			result := service.ValidateDeployFlags(cmd)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestDeployValidationService_ValidateRequiredArguments(t *testing.T) {
	service := NewDeployValidationService()

	tests := []struct {
		name          string
		setupCmd      func() *cobra.Command
		requiredArgs  []string
		expectedValid bool
		expectedError string
	}{
		{
			name: "All required arguments provided",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("location", "", "Location")
				cmd.Flags().String("resource-group", "", "Resource group")
				cmd.Flags().String("windows-user", "", "Windows user")
				cmd.Flags().String("flavor", "", "Flavor")

				cmd.Flags().Set("location", "eastus")
				cmd.Flags().Set("resource-group", "test-rg")
				cmd.Flags().Set("windows-user", "testuser")
				cmd.Flags().Set("flavor", "ITPro")

				return cmd
			},
			requiredArgs:  []string{"location", "resource-group", "windows-user", "flavor"},
			expectedValid: true,
		},
		{
			name: "Missing required arguments",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("location", "", "Location")
				cmd.Flags().String("resource-group", "", "Resource group")
				cmd.Flags().String("windows-user", "", "Windows user")
				cmd.Flags().String("flavor", "", "Flavor")

				// Only set some flags, leave others missing
				cmd.Flags().Set("location", "eastus")
				cmd.Flags().Set("resource-group", "test-rg")
				// windows-user and flavor missing

				return cmd
			},
			requiredArgs:  []string{"location", "resource-group", "windows-user", "flavor"},
			expectedValid: false,
			expectedError: "deploy command missing required arguments: --location, --resource-group, --windows-user, --flavor are required",
		},
		{
			name: "Empty string arguments",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("location", "", "Location")
				cmd.Flags().String("resource-group", "", "Resource group")

				cmd.Flags().Set("location", "")
				cmd.Flags().Set("resource-group", "test-rg")

				return cmd
			},
			requiredArgs:  []string{"location", "resource-group"},
			expectedValid: false,
			expectedError: "deploy command missing required arguments: --location, --resource-group, --windows-user, --flavor are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stderr to avoid cluttering test output
			originalStderr := os.Stderr
			defer func() { os.Stderr = originalStderr }()

			cmd := tt.setupCmd()
			result := service.ValidateRequiredArguments(cmd, tt.requiredArgs)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestDeployValidationService_ValidateConditionalRequirements(t *testing.T) {
	service := NewDeployValidationService()

	tests := []struct {
		name          string
		setupCmd      func() *cobra.Command
		expectedValid bool
		expectedError string
	}{
		{
			name: "ITPro flavor - no additional requirements",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("flavor", "", "Flavor")
				cmd.Flags().String("ssh-rsa-public-key", "", "SSH key")
				cmd.Flags().String("github-user", "", "GitHub user")

				cmd.Flags().Set("flavor", "ITPro")

				return cmd
			},
			expectedValid: true,
		},
		{
			name: "DevOps flavor with SSH key and GitHub user",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("flavor", "", "Flavor")
				cmd.Flags().String("ssh-rsa-public-key", "", "SSH key")
				cmd.Flags().String("github-user", "", "GitHub user")

				cmd.Flags().Set("flavor", "DevOps")
				cmd.Flags().Set("ssh-rsa-public-key", "ssh-rsa AAAAB3...")
				cmd.Flags().Set("github-user", "myusername")

				return cmd
			},
			expectedValid: true,
		},
		{
			name: "DevOps flavor missing SSH key",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("flavor", "", "Flavor")
				cmd.Flags().String("ssh-rsa-public-key", "", "SSH key")
				cmd.Flags().String("github-user", "", "GitHub user")

				cmd.Flags().Set("flavor", "DevOps")
				cmd.Flags().Set("github-user", "myusername")
				// SSH key missing

				return cmd
			},
			expectedValid: false,
			expectedError: "deployment flavor requirements validation failed: missing required parameters for selected flavor",
		},
		{
			name: "DevOps flavor missing GitHub user",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("flavor", "", "Flavor")
				cmd.Flags().String("ssh-rsa-public-key", "", "SSH key")
				cmd.Flags().String("github-user", "", "GitHub user")

				cmd.Flags().Set("flavor", "DevOps")
				cmd.Flags().Set("ssh-rsa-public-key", "ssh-rsa AAAAB3...")
				// GitHub user missing or default

				return cmd
			},
			expectedValid: false,
			expectedError: "deployment flavor requirements validation failed",
		},
		{
			name: "DevOps flavor with default GitHub user",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("flavor", "", "Flavor")
				cmd.Flags().String("ssh-rsa-public-key", "", "SSH key")
				cmd.Flags().String("github-user", "", "GitHub user")

				cmd.Flags().Set("flavor", "DevOps")
				cmd.Flags().Set("ssh-rsa-public-key", "ssh-rsa AAAAB3...")
				cmd.Flags().Set("github-user", "microsoft") // Default value, should fail

				return cmd
			},
			expectedValid: false,
			expectedError: "deployment flavor requirements validation failed",
		},
		{
			name: "DataOps flavor with SSH key",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("flavor", "", "Flavor")
				cmd.Flags().String("ssh-rsa-public-key", "", "SSH key")
				cmd.Flags().String("github-user", "", "GitHub user")

				cmd.Flags().Set("flavor", "DataOps")
				cmd.Flags().Set("ssh-rsa-public-key", "ssh-rsa AAAAB3...")

				return cmd
			},
			expectedValid: true,
		},
		{
			name: "DataOps flavor missing SSH key",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("flavor", "", "Flavor")
				cmd.Flags().String("ssh-rsa-public-key", "", "SSH key")
				cmd.Flags().String("github-user", "", "GitHub user")

				cmd.Flags().Set("flavor", "DataOps")
				// SSH key missing

				return cmd
			},
			expectedValid: false,
			expectedError: "deployment flavor requirements validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout/stderr to avoid cluttering test output
			originalStdout := os.Stdout
			originalStderr := os.Stderr
			defer func() {
				os.Stdout = originalStdout
				os.Stderr = originalStderr
			}()

			cmd := tt.setupCmd()
			result := service.ValidateConditionalRequirements(cmd)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestDeployValidationService_RunPreflightChecks(t *testing.T) {
	service := NewDeployValidationService()

	tests := []struct {
		name          string
		skipPreflight bool
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Skip preflight checks",
			skipPreflight: true,
			expectedValid: true,
		},
		// Note: Testing actual preflight checks would require extensive mocking
		// of Azure CLI and other dependencies. These are covered by the existing
		// preflight validation tests.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // Create a minimal command for testing
			cmd := &cobra.Command{}
			cmd.Flags().String("skip-preflight", "", "Skip preflight")

			result := service.RunPreflightChecks(cmd, tt.skipPreflight)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}

func TestDeployValidationService_ValidateAllDeployRequirements(t *testing.T) {
	service := NewDeployValidationService()

	tests := []struct {
		name          string
		setupCmd      func() *cobra.Command
		expectedValid bool
		expectedError string
	}{
		{
			name: "All validations pass",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("location", "", "Location")
				cmd.Flags().String("resource-group", "", "Resource group")
				cmd.Flags().String("windows-user", "", "Windows user")
				cmd.Flags().String("flavor", "", "Flavor")
				cmd.Flags().String("ssh-rsa-public-key", "", "SSH key")
				cmd.Flags().String("github-user", "", "GitHub user")
				cmd.Flags().String("skip-preflight", "", "Skip preflight")

				// Set all required values
				cmd.Flags().Set("location", "eastus")
				cmd.Flags().Set("resource-group", "test-rg")
				cmd.Flags().Set("windows-user", "testuser")
				cmd.Flags().Set("flavor", "ITPro")
				cmd.Flags().Set("skip-preflight", "yes") // Skip preflight to avoid Azure CLI dependency

				return cmd
			},
			expectedValid: true,
		},
		{
			name: "Missing required arguments",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("location", "", "Location")
				cmd.Flags().String("resource-group", "", "Resource group")
				cmd.Flags().String("windows-user", "", "Windows user")
				cmd.Flags().String("flavor", "", "Flavor")
				cmd.Flags().String("skip-preflight", "", "Skip preflight")

				// Only set some required values
				cmd.Flags().Set("location", "eastus")
				cmd.Flags().Set("resource-group", "test-rg")
				// windows-user and flavor missing

				return cmd
			},
			expectedValid: false,
			expectedError: "missing required arguments",
		},
		{
			name: "DevOps flavor missing SSH key",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("location", "", "Location")
				cmd.Flags().String("resource-group", "", "Resource group")
				cmd.Flags().String("windows-user", "", "Windows user")
				cmd.Flags().String("flavor", "", "Flavor")
				cmd.Flags().String("ssh-rsa-public-key", "", "SSH key")
				cmd.Flags().String("github-user", "", "GitHub user")
				cmd.Flags().String("skip-preflight", "", "Skip preflight")

				// Set required values but miss SSH key for DevOps flavor
				cmd.Flags().Set("location", "eastus")
				cmd.Flags().Set("resource-group", "test-rg")
				cmd.Flags().Set("windows-user", "testuser")
				cmd.Flags().Set("flavor", "DevOps")
				cmd.Flags().Set("github-user", "myuser")
				// SSH key missing

				return cmd
			},
			expectedValid: false,
			expectedError: "deployment flavor requirements validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout/stderr to avoid cluttering test output
			originalStdout := os.Stdout
			originalStderr := os.Stderr
			defer func() {
				os.Stdout = originalStdout
				os.Stderr = originalStderr
			}()

			cmd := tt.setupCmd()
			result := service.ValidateAllDeployRequirements(cmd)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			if !tt.expectedValid {
				assert.Contains(t, result.Error.Error(), tt.expectedError)
			}
		})
	}
}
