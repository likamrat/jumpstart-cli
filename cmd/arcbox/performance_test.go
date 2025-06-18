package arcbox

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"jumpstartcli/cmd/arcbox/services"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// TestPerformance_LargeSubscriptionSets tests performance with large numbers of subscriptions
func TestPerformance_LargeSubscriptionSets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Test with increasing subscription counts
	testCases := []struct {
		name                 string
		subscriptionCount    int
		resourceGroupsPerSub int
		maxDuration          time.Duration
	}{
		{"Small Scale", 10, 50, 5 * time.Second},
		{"Medium Scale", 50, 100, 15 * time.Second},
		{"Large Scale", 100, 100, 30 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()

			// Generate mock subscription data
			subscriptions := generateMockSubscriptionData(tc.subscriptionCount)
			mockCLI.Subscriptions = subscriptions

			// Generate mock resource groups for each subscription
			for i := 0; i < tc.subscriptionCount; i++ {
				resourceGroups := generateMockResourceGroupData(tc.resourceGroupsPerSub, fmt.Sprintf("sub-%d", i))
				mockCLI.ResourceGroups = append(mockCLI.ResourceGroups, resourceGroups...)
			}

			// Set up listing service
			listingService := services.NewListingService(mockCLI)

			// Measure performance
			start := time.Now()
			err := listingService.ListDeployments(true, false, "", utils.OutputFormat)
			duration := time.Since(start)

			// Validate results
			if err != nil {
				t.Errorf("Performance test failed: %v", err)
			}

			t.Logf("Processed %d subscriptions with %d resource groups each in %v",
				tc.subscriptionCount, tc.resourceGroupsPerSub, duration)

			if duration > tc.maxDuration {
				t.Errorf("Performance test too slow: %v > %v for %s", duration, tc.maxDuration, tc.name)
			}
		})
	}
}

// TestPerformance_ConcurrentExecution tests concurrent command execution safety
func TestPerformance_ConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	const numWorkers = 10
	const iterationsPerWorker = 20

	mockCLI := azurecli.NewMockAzureCLI()
	setupMockDataForConcurrency(mockCLI)

	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	// Test concurrent list operations
	t.Run("Concurrent_List_Operations", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < iterationsPerWorker; j++ {
					listingService := services.NewListingService(mockCLI)
					err := listingService.ListDeployments(false, true, "", "table")

					if err != nil {
						mu.Lock()
						errors = append(errors, fmt.Errorf("worker %d iteration %d: %v", workerID, j, err))
						mu.Unlock()
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		// Validate concurrent execution
		if len(errors) > 0 {
			t.Errorf("Concurrent execution had %d errors. First error: %v", len(errors), errors[0])
		}

		totalOperations := numWorkers * iterationsPerWorker
		avgDuration := duration / time.Duration(totalOperations)
		t.Logf("Completed %d concurrent operations in %v (avg: %v per operation)",
			totalOperations, duration, avgDuration)

		// Performance threshold: should complete within reasonable time
		if duration > 60*time.Second {
			t.Errorf("Concurrent execution too slow: %v > 60s", duration)
		}
	})

	// Test concurrent deploy command creation (safe command building)
	t.Run("Concurrent_Deploy_Command_Creation", func(t *testing.T) {
		start := time.Now()
		wg = sync.WaitGroup{}
		errors = make([]error, 0)

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < iterationsPerWorker; j++ {
					// Test concurrent command creation - this should be thread-safe
					cmd := NewArcboxCmdWithCLI(mockCLI)
					deployCmd := getDeployCommandByUse(cmd, "deploy")

					if deployCmd == nil {
						mu.Lock()
						errors = append(errors, fmt.Errorf("worker %d iteration %d: deploy command not found", workerID, j))
						mu.Unlock()
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		if len(errors) > 0 {
			t.Errorf("Concurrent command creation had %d errors. First error: %v", len(errors), errors[0])
		}

		t.Logf("Created %d commands concurrently in %v", numWorkers*iterationsPerWorker, duration)
	})
}

// TestPerformance_MemoryUsage tests memory usage and leak detection
func TestPerformance_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Force garbage collection to get baseline
	runtime.GC()
	runtime.GC() // Call twice to ensure cleanup

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Perform memory-intensive operations
	mockCLI := azurecli.NewMockAzureCLI()

	// Generate large datasets
	subscriptions := generateMockSubscriptionData(100)
	mockCLI.Subscriptions = subscriptions

	for i := 0; i < 100; i++ {
		resourceGroups := generateMockResourceGroupData(500, fmt.Sprintf("sub-%d", i)) // Large resource group count
		mockCLI.ResourceGroups = append(mockCLI.ResourceGroups, resourceGroups...)
	}

	// Perform multiple operations to test memory usage
	for iteration := 0; iteration < 10; iteration++ {
		listingService := services.NewListingService(mockCLI)
		err := listingService.ListDeployments(true, false, "", "table")
		if err != nil {
			t.Errorf("Memory test iteration %d failed: %v", iteration, err)
		}
	}

	// Force garbage collection and measure memory
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate memory usage
	allocatedBytes := m2.TotalAlloc - m1.TotalAlloc
	heapGrowth := int64(m2.HeapInuse) - int64(m1.HeapInuse)

	t.Logf("Memory usage: Allocated %d bytes, Heap growth: %d bytes", allocatedBytes, heapGrowth)
	t.Logf("Number of GC runs: %d", m2.NumGC-m1.NumGC)

	// Memory thresholds (these are reasonable for processing large datasets)
	maxHeapGrowth := int64(50 * 1024 * 1024) // 50MB max heap growth
	if heapGrowth > maxHeapGrowth {
		t.Errorf("Excessive heap growth: %d bytes > %d bytes", heapGrowth, maxHeapGrowth)
	}

	// Test for potential memory leaks by running operations multiple times
	t.Run("Memory_Leak_Detection", func(t *testing.T) {
		var heapSizes []uint64

		for i := 0; i < 5; i++ {
			// Perform operations
			for j := 0; j < 20; j++ {
				listingService := services.NewListingService(mockCLI)
				listingService.ListDeployments(false, true, "", "json")
			}

			// Force GC and measure
			runtime.GC()
			runtime.GC()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			heapSizes = append(heapSizes, m.HeapInuse)

			t.Logf("Iteration %d heap size: %d bytes", i, m.HeapInuse)
		}

		// Check for significant heap growth across iterations (potential leak)
		if len(heapSizes) >= 2 {
			firstHeap := heapSizes[0]
			lastHeap := heapSizes[len(heapSizes)-1]
			growth := float64(lastHeap) / float64(firstHeap)

			// If heap grows by more than 200% across iterations, we might have a leak
			if growth > 2.0 {
				t.Errorf("Potential memory leak detected: heap grew from %d to %d bytes (%.2fx growth)",
					firstHeap, lastHeap, growth)
			}
		}
	})
}

// TestPerformance_ResourceGroupDiscovery tests performance of resource discovery algorithms
func TestPerformance_ResourceGroupDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Test with varying numbers of resource groups
	testCases := []struct {
		name          string
		resourceCount int
		arcboxCount   int // How many should match ArcBox criteria
		maxDuration   time.Duration
	}{
		{"Small Dataset", 100, 5, 2 * time.Second},
		{"Medium Dataset", 500, 25, 5 * time.Second},
		{"Large Dataset", 1000, 50, 10 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()

			// Set up proper mock data for resource discovery
			setupMockDataForResourceDiscovery(mockCLI, tc.resourceCount, tc.arcboxCount)

			listingService := services.NewListingService(mockCLI)

			start := time.Now()
			err := listingService.ListDeployments(false, false, "test-sub", "table")
			duration := time.Since(start)

			if err != nil {
				t.Errorf("Resource discovery test failed: %v", err)
			}

			t.Logf("Discovered ArcBox deployments from %d resource groups in %v", tc.resourceCount, duration)

			if duration > tc.maxDuration {
				t.Errorf("Resource discovery too slow: %v > %v for %s", duration, tc.maxDuration, tc.name)
			}
		})
	}
}

// Helper functions for generating mock data

func generateMockSubscriptionData(count int) []azurecli.SubscriptionInfo {
	subscriptions := make([]azurecli.SubscriptionInfo, count)
	for i := 0; i < count; i++ {
		subscriptions[i] = azurecli.SubscriptionInfo{
			ID:        fmt.Sprintf("sub-%d", i),
			Name:      fmt.Sprintf("Test Subscription %d", i),
			TenantID:  "tenant-1",
			State:     "Enabled",
			IsDefault: i == 0,
		}
	}
	return subscriptions
}

func generateMockResourceGroupData(count int, subscriptionID string) []azurecli.ResourceGroupInfo {
	resourceGroups := make([]azurecli.ResourceGroupInfo, count)
	for i := 0; i < count; i++ {
		resourceGroups[i] = azurecli.ResourceGroupInfo{
			Name:     fmt.Sprintf("rg-%s-%d", subscriptionID, i),
			Location: "eastus",
			Properties: map[string]interface{}{
				"provisioningState": "Succeeded",
			},
		}
	}
	return resourceGroups
}

func generateMixedResourceGroupData(totalCount, arcboxCount int, subscriptionID string) []azurecli.ResourceGroupInfo {
	resourceGroups := make([]azurecli.ResourceGroupInfo, totalCount)

	// Generate ArcBox resource groups
	for i := 0; i < arcboxCount && i < totalCount; i++ {
		resourceGroups[i] = azurecli.ResourceGroupInfo{
			Name:     fmt.Sprintf("arcbox-rg-%d", i),
			Location: "eastus",
			Properties: map[string]interface{}{
				"provisioningState": "Succeeded",
				"tags": map[string]string{
					"Solution":    "jumpstart_arcbox",
					"Environment": "Test",
				},
			},
		}
	}

	// Generate non-ArcBox resource groups
	for i := arcboxCount; i < totalCount; i++ {
		resourceGroups[i] = azurecli.ResourceGroupInfo{
			Name:     fmt.Sprintf("regular-rg-%d", i),
			Location: "westus",
			Properties: map[string]interface{}{
				"provisioningState": "Succeeded",
				"tags": map[string]string{
					"Environment": "Production",
				},
			},
		}
	}

	return resourceGroups
}

func setupMockDataForConcurrency(mockCLI *azurecli.MockAzureCLI) {
	// Set up mock responses for concurrent testing
	subscriptions := generateMockSubscriptionData(5)
	mockCLI.Subscriptions = subscriptions

	resourceGroups := generateMockResourceGroupData(50, "test-sub")
	mockCLI.ResourceGroups = resourceGroups

	// Set current subscription
	mockCLI.CurrentSubscription = &subscriptions[0]
}

func setupMockDataForResourceDiscovery(mockCLI *azurecli.MockAzureCLI, resourceCount, arcboxCount int) {
	// Set up test subscription
	testSub := azurecli.SubscriptionInfo{
		ID:        "test-sub",
		Name:      "Test Subscription",
		TenantID:  "tenant-1",
		State:     "Enabled",
		IsDefault: true,
	}
	mockCLI.CurrentSubscription = &testSub
	mockCLI.Subscriptions = []azurecli.SubscriptionInfo{testSub}

	// Generate mixed resource groups (some ArcBox, some not)
	resourceGroups := generateMixedResourceGroupData(resourceCount, arcboxCount, "test-sub")
	mockCLI.ResourceGroups = resourceGroups

	// Initialize Resources map
	if mockCLI.Resources == nil {
		mockCLI.Resources = make(map[string][]azurecli.ResourceInfo)
	}

	// Set up resources for ArcBox resource groups to simulate real deployments
	for i := 0; i < arcboxCount; i++ {
		rgName := fmt.Sprintf("arcbox-rg-%d", i)
		mockCLI.Resources[rgName] = []azurecli.ResourceInfo{
			{
				ID:   fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/arcbox-vm", rgName),
				Name: "arcbox-vm",
				Type: "Microsoft.Compute/virtualMachines",
				Tags: map[string]string{"Solution": "jumpstart_arcbox"},
			},
			{
				ID:   fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/arcbox-vnet", rgName),
				Name: "arcbox-vnet",
				Type: "Microsoft.Network/virtualNetworks",
				Tags: map[string]string{"Solution": "jumpstart_arcbox"},
			},
		}
	}
}

// Helper function to get deploy command by use string
func getDeployCommandByUse(cmd *cobra.Command, use string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == use {
			return subCmd
		}
	}
	return nil
}
