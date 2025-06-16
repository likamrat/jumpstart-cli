package display

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"jumpstartcli/cmd/arcbox/models"
	"jumpstartcli/internal/azurecli"
)

func TestNewListFormatter(t *testing.T) {
	formatter := NewListFormatter()

	if formatter == nil {
		t.Fatal("NewListFormatter returned nil")
	}
}

func TestListFormatter_OutputArcBoxDeploymentsTable(t *testing.T) {
	tests := []struct {
		name        string
		deployments []models.ArcBoxDeployment
		expected    []string
		excluded    []string
	}{
		{
			name: "single_deployment",
			deployments: []models.ArcBoxDeployment{
				{
					ResourceGroupName: "arcbox-rg",
					SubscriptionID:    "12345678-1234-1234-1234-123456789012",
					SubscriptionName:  "Test Subscription",
					Location:          "eastus",
					CreatedDate:       "2025-06-15",
					Status:            "Succeeded",
					ResourceCount:     15,
					Flavor:            "ITPro",
					NamingPrefix:      "ArcBox",
				},
			},
			expected: []string{
				"🎯 Found 1 ArcBox deployment(s)",
				"arcbox-rg",
				"Test Subscription",
				"eastus",
				"✅ Succeeded",
				"15",
				"ITPro",
				"2025-06-15",
			},
		},
		{
			name: "multiple_deployments",
			deployments: []models.ArcBoxDeployment{
				{
					ResourceGroupName: "arcbox-rg-1",
					SubscriptionID:    "sub1",
					SubscriptionName:  "Subscription 1",
					Location:          "eastus",
					CreatedDate:       "2025-06-15",
					Status:            "Succeeded",
					ResourceCount:     10,
					Flavor:            "ITPro",
					NamingPrefix:      "ArcBox",
				},
				{
					ResourceGroupName: "arcbox-rg-2",
					SubscriptionID:    "sub2",
					SubscriptionName:  "Subscription 2",
					Location:          "westus",
					CreatedDate:       "2025-06-14",
					Status:            "Failed",
					ResourceCount:     5,
					Flavor:            "DevOps",
					NamingPrefix:      "ArcBox",
				},
			},
			expected: []string{
				"🎯 Found 2 ArcBox deployment(s)",
				"arcbox-rg-1",
				"arcbox-rg-2",
				"Subscription 1",
				"Subscription 2",
				"✅ Succeeded",
				"❌ Failed",
				"ITPro",
				"DevOps",
				"10",
				"5",
			},
		},
		{
			name: "long_subscription_name_truncated",
			deployments: []models.ArcBoxDeployment{
				{
					ResourceGroupName: "arcbox-rg",
					SubscriptionID:    "sub1",
					SubscriptionName:  "This is a very long subscription name that should be truncated",
					Location:          "eastus",
					CreatedDate:       "2025-06-15",
					Status:            "Succeeded",
					ResourceCount:     10,
					Flavor:            "ITPro",
					NamingPrefix:      "ArcBox",
				},
			},
			expected: []string{
				"This is a very long su...",
			},
			excluded: []string{
				"This is a very long subscription name that should be truncated",
			},
		},
		{
			name: "various_status_icons",
			deployments: []models.ArcBoxDeployment{
				{
					ResourceGroupName: "rg-succeeded",
					SubscriptionName:  "Sub1",
					Status:            "succeeded",
					Flavor:            "ITPro",
				},
				{
					ResourceGroupName: "rg-failed",
					SubscriptionName:  "Sub2",
					Status:            "failed",
					Flavor:            "DevOps",
				},
				{
					ResourceGroupName: "rg-running",
					SubscriptionName:  "Sub3",
					Status:            "running",
					Flavor:            "DataOps",
				},
				{
					ResourceGroupName: "rg-canceled",
					SubscriptionName:  "Sub4",
					Status:            "canceled",
					Flavor:            "ITPro",
				},
				{
					ResourceGroupName: "rg-unknown",
					SubscriptionName:  "Sub5",
					Status:            "unknown",
					Flavor:            "ITPro",
				},
			},
			expected: []string{
				"✅ succeeded",
				"❌ failed",
				"⌛ running",
				"🚫 canceled",
				"❓ unknown",
			},
		},
		{
			name:        "empty_deployments",
			deployments: []models.ArcBoxDeployment{},
			expected: []string{
				"🎯 Found 0 ArcBox deployment(s)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewListFormatter()

			output := captureOutput(func() {
				err := formatter.OutputArcBoxDeploymentsTable(tt.deployments)
				if err != nil {
					t.Errorf("OutputArcBoxDeploymentsTable returned error: %v", err)
				}
			})

			// Check expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, output)
				}
			}

			// Check excluded strings are absent
			for _, excluded := range tt.excluded {
				if strings.Contains(output, excluded) {
					t.Errorf("Expected output to NOT contain '%s', but it did. Output: %s", excluded, output)
				}
			}
		})
	}
}

func TestListFormatter_OutputArcBoxDeploymentsJSON(t *testing.T) {
	tests := []struct {
		name        string
		deployments []models.ArcBoxDeployment
		expected    []string
	}{
		{
			name: "single_deployment_json",
			deployments: []models.ArcBoxDeployment{
				{
					ResourceGroupName: "arcbox-rg",
					SubscriptionID:    "12345678-1234-1234-1234-123456789012",
					SubscriptionName:  "Test Subscription",
					Location:          "eastus",
					CreatedDate:       "2025-06-15",
					Status:            "Succeeded",
					ResourceCount:     15,
					Flavor:            "ITPro",
					NamingPrefix:      "ArcBox",
				},
			},
			expected: []string{
				`"ResourceGroupName": "arcbox-rg"`,
				`"SubscriptionID": "12345678-1234-1234-1234-123456789012"`,
				`"SubscriptionName": "Test Subscription"`,
				`"Location": "eastus"`,
				`"Status": "Succeeded"`,
				`"Flavor": "ITPro"`,
				`"ResourceCount": 15`,
			},
		},
		{
			name: "multiple_deployments_json",
			deployments: []models.ArcBoxDeployment{
				{
					ResourceGroupName: "rg1",
					SubscriptionName:  "Sub1",
					Flavor:            "ITPro",
				},
				{
					ResourceGroupName: "rg2",
					SubscriptionName:  "Sub2",
					Flavor:            "DevOps",
				},
			},
			expected: []string{
				`"ResourceGroupName": "rg1"`,
				`"ResourceGroupName": "rg2"`,
				`"Flavor": "ITPro"`,
				`"Flavor": "DevOps"`,
			},
		},
		{
			name:        "empty_deployments_json",
			deployments: []models.ArcBoxDeployment{},
			expected: []string{
				"[]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewListFormatter()

			output := captureOutput(func() {
				err := formatter.OutputArcBoxDeploymentsJSON(tt.deployments)
				if err != nil {
					t.Errorf("OutputArcBoxDeploymentsJSON returned error: %v", err)
				}
			})

			// Check that output is valid-looking JSON
			if len(tt.deployments) > 0 {
				if !strings.HasPrefix(strings.TrimSpace(output), "[") || !strings.HasSuffix(strings.TrimSpace(output), "]") {
					t.Errorf("Expected JSON array format, got: %s", output)
				}
			}

			// Check expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, output)
				}
			}
		})
	}
}

func TestListFormatter_getStatusIcon(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"succeeded_lowercase", "succeeded", "✅"},
		{"succeeded_uppercase", "SUCCEEDED", "✅"},
		{"failed_lowercase", "failed", "❌"},
		{"failed_uppercase", "FAILED", "❌"},
		{"running", "running", "⌛"},
		{"creating", "creating", "⌛"},
		{"accepted", "accepted", "⌛"},
		{"inprogress", "inprogress", "⌛"},
		{"canceled", "canceled", "🚫"},
		{"cancelled", "cancelled", "🚫"},
		{"unknown", "unknown", "❓"},
		{"empty", "", "❓"},
		{"random", "SomeRandomStatus", "❓"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewListFormatter()
			result := formatter.getStatusIcon(tt.status)

			if result != tt.expected {
				t.Errorf("getStatusIcon(%q) = %q, expected %q", tt.status, result, tt.expected)
			}
		})
	}
}

func TestListFormatter_ScanSubscriptionWithSpinner(t *testing.T) {
	tests := []struct {
		name            string
		subscription    models.AzureSubscription
		mockDeployments []models.ArcBoxDeployment
		discoverError   error
		expectedOutput  []string
	}{
		{
			name: "successful_scan_with_deployments",
			subscription: models.AzureSubscription{
				ID:   "sub1",
				Name: "Test Subscription",
			},
			mockDeployments: []models.ArcBoxDeployment{
				{ResourceGroupName: "rg1", Flavor: "ITPro"},
				{ResourceGroupName: "rg2", Flavor: "DevOps"},
			},
			discoverError: nil,
			expectedOutput: []string{
				"🔍 Scanning subscription Test Subscription...",
				"✅",
				"Found 2 ArcBox deployment(s) in Test Subscription",
			},
		},
		{
			name: "successful_scan_no_deployments",
			subscription: models.AzureSubscription{
				ID:   "sub2",
				Name: "Empty Subscription",
			},
			mockDeployments: []models.ArcBoxDeployment{},
			discoverError:   nil,
			expectedOutput: []string{
				"🔍 Scanning subscription Empty Subscription...",
				"✅",
				"Found 0 ArcBox deployment(s) in Empty Subscription",
			},
		},
		{
			name: "failed_scan",
			subscription: models.AzureSubscription{
				ID:   "sub3",
				Name: "Failed Subscription",
			},
			mockDeployments: nil,
			discoverError:   fmt.Errorf("subscription access denied"),
			expectedOutput: []string{
				"🔍 Scanning subscription Failed Subscription...",
				"❌",
				"Error scanning subscription Failed Subscription: subscription access denied",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewListFormatter()
			mockCLI := azurecli.NewMockAzureCLI()

			// Create a mock discover function
			discoverFunc := func(azurecli.AzureCLI, string, string) ([]models.ArcBoxDeployment, error) {
				if tt.discoverError != nil {
					return nil, tt.discoverError
				}
				return tt.mockDeployments, nil
			}

			// Run the scan with a short timeout to avoid hanging tests
			done := make(chan []models.ArcBoxDeployment, 1)
			go func() {
				result := formatter.ScanSubscriptionWithSpinner(mockCLI, tt.subscription, discoverFunc)
				done <- result
			}()

			// Wait for result or timeout
			select {
			case result := <-done:
				// Verify the result matches expected deployments
				if tt.discoverError == nil {
					if len(result) != len(tt.mockDeployments) {
						t.Errorf("Expected %d deployments, got %d", len(tt.mockDeployments), len(result))
					}
				} else {
					if len(result) != 0 {
						t.Errorf("Expected empty result on error, got %d deployments", len(result))
					}
				}
			case <-time.After(5 * time.Second):
				t.Error("Test timed out - spinner may have hung")
			}
		})
	}
}
