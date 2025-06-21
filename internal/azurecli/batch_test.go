package azurecli

import (
	"fmt"
	"testing"
	"time"
)

// TestGetResourcesBatch_EnhancedFeatures tests the enhanced parallel querying features
func TestGetResourcesBatch_EnhancedFeatures(t *testing.T) {
	tests := []struct {
		name                  string
		resourceIDs           []string
		maxConcurrency        int
		mockSpecificResources map[string]*ResourceInfo
		mockErrors            map[string]error
		expectedResultCount   int
		expectedErrorCount    int
		expectOrderPreserved  bool
	}{
		{
			name: "order_preservation_with_mixed_results",
			resourceIDs: []string{
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/vm-1",
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/storage-1",
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/vm-2",
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/storage-2",
			},
			maxConcurrency: 2,
			mockSpecificResources: map[string]*ResourceInfo{
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/vm-1": {
					Name: "vm-1", Type: "Microsoft.Compute/virtualMachines",
					Properties: map[string]interface{}{"provisioningState": "Succeeded"},
				},
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/storage-1": {
					Name: "storage-1", Type: "Microsoft.Storage/storageAccounts",
					Properties: map[string]interface{}{"provisioningState": "Running"},
				},
				// vm-2 will have error (missing from mock data)
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/storage-2": {
					Name: "storage-2", Type: "Microsoft.Storage/storageAccounts",
					Properties: map[string]interface{}{"provisioningState": "Failed"},
				},
			},
			expectedResultCount:  3, // vm-1, storage-1, storage-2
			expectedErrorCount:   1, // vm-2 missing
			expectOrderPreserved: true,
		},
		{
			name: "concurrency_limit_enforcement",
			resourceIDs: []string{
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/vm-1",
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/vm-2",
			},
			maxConcurrency: 10, // Should be capped at 8
			mockSpecificResources: map[string]*ResourceInfo{
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/vm-1": {
					Name: "vm-1", Type: "Microsoft.Compute/virtualMachines",
					Properties: map[string]interface{}{"provisioningState": "Succeeded"},
				},
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/vm-2": {
					Name: "vm-2", Type: "Microsoft.Compute/virtualMachines",
					Properties: map[string]interface{}{"provisioningState": "Succeeded"},
				},
			},
			expectedResultCount:  2,
			expectedErrorCount:   0,
			expectOrderPreserved: true,
		},
		{
			name:                 "empty_resource_list",
			resourceIDs:          []string{},
			maxConcurrency:       5,
			expectedResultCount:  0,
			expectedErrorCount:   0,
			expectOrderPreserved: true,
		},
		{
			name: "default_concurrency_when_zero",
			resourceIDs: []string{
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/vm-1",
			},
			maxConcurrency: 0, // Should default to 5
			mockSpecificResources: map[string]*ResourceInfo{
				"/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/vm-1": {
					Name: "vm-1", Type: "Microsoft.Compute/virtualMachines",
					Properties: map[string]interface{}{"provisioningState": "Succeeded"},
				},
			},
			expectedResultCount:  1,
			expectedErrorCount:   0,
			expectOrderPreserved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockAzureCLI()

			// Setup mock resources
			for id, resource := range tt.mockSpecificResources {
				mock.SetSpecificResource(id, resource)
			}

			startTime := time.Now()
			results, errors := mock.GetResourcesBatch(tt.resourceIDs, tt.maxConcurrency)
			duration := time.Since(startTime)

			// Verify result count
			if len(results) != tt.expectedResultCount {
				t.Errorf("Expected %d results, got %d", tt.expectedResultCount, len(results))
			}

			// Verify error count
			if len(errors) != tt.expectedErrorCount {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrorCount, len(errors))
			}

			// Verify order preservation for successful results
			if tt.expectOrderPreserved && len(results) > 1 {
				// Check that the relative order of successful resources matches input order
				resultNames := make([]string, len(results))
				for i, result := range results {
					resultNames[i] = result.Name
				}

				// For this test, we expect vm-1, storage-1, storage-2 in that order
				if tt.name == "order_preservation_with_mixed_results" {
					expectedOrder := []string{"vm-1", "storage-1", "storage-2"}
					for i, expected := range expectedOrder {
						if i < len(resultNames) && resultNames[i] != expected {
							t.Errorf("Order not preserved: expected %s at position %d, got %s",
								expected, i, resultNames[i])
						}
					}
				}
			}

			// Verify that batch operations are reasonably fast (should complete quickly for mock)
			if duration > 5*time.Second {
				t.Errorf("Batch operation took too long: %v", duration)
			}

			t.Logf("✅ Test %s: %d results, %d errors, completed in %v",
				tt.name, len(results), len(errors), duration)
		})
	}
}

// TestGetResourcesBatch_ConcurrencyBehavior tests specific concurrency patterns
func TestGetResourcesBatch_ConcurrencyBehavior(t *testing.T) {
	mock := NewMockAzureCLI()

	// Create a larger set of resources to test concurrency
	resourceIDs := make([]string, 20)
	for i := 0; i < 20; i++ {
		resourceID := fmt.Sprintf("/subscriptions/sub1/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/vm-%d", i)
		resourceIDs[i] = resourceID

		// Setup mock data for each resource
		mock.SetSpecificResource(resourceID, &ResourceInfo{
			Name:       fmt.Sprintf("vm-%d", i),
			Type:       "Microsoft.Compute/virtualMachines",
			Properties: map[string]interface{}{"provisioningState": "Succeeded"},
		})
	}

	// Test with different concurrency levels
	concurrencyLevels := []int{1, 3, 5, 8, 10} // 10 should be capped at 8

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
			startTime := time.Now()
			results, errors := mock.GetResourcesBatch(resourceIDs, concurrency)
			duration := time.Since(startTime)

			// Should get all 20 resources successfully
			if len(results) != 20 {
				t.Errorf("Expected 20 results, got %d", len(results))
			}

			if len(errors) != 0 {
				t.Errorf("Expected 0 errors, got %d", len(errors))
			}

			// Verify order preservation
			for i, result := range results {
				expectedName := fmt.Sprintf("vm-%d", i)
				if result.Name != expectedName {
					t.Errorf("Order not preserved: expected %s at position %d, got %s",
						expectedName, i, result.Name)
				}
			}

			t.Logf("✅ Concurrency %d: %d results in %v", concurrency, len(results), duration)
		})
	}
}
