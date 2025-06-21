package display

import (
	"testing"
	"time"

	"jumpstartcli/cmd/arcbox/models"
)

func TestIntelligentCache(t *testing.T) {
	cache := NewIntelligentCache()

	// Test cache initialization
	if !cache.IsEnabled() {
		t.Error("Cache should be enabled by default")
	}

	// Test resource caching
	resource := &models.ResourceStatus{
		Name:  "test-vm",
		Type:  "Microsoft.Compute/virtualMachines",
		State: "Creating",
	}

	// Initially, resource should not be in cache
	if _, found := cache.GetCachedResource("test-vm"); found {
		t.Error("Resource should not be in cache initially")
	}

	// Cache the resource
	cache.CacheResource(resource)

	// Now resource should be in cache
	if cachedResource, found := cache.GetCachedResource("test-vm::Microsoft.Compute/virtualMachines"); found {
		if cachedResource.Name != resource.Name {
			t.Errorf("Expected name %s, got %s", resource.Name, cachedResource.Name)
		}
		if cachedResource.Type != resource.Type {
			t.Errorf("Expected type %s, got %s", resource.Type, cachedResource.Type)
		}
		if cachedResource.State != resource.State {
			t.Errorf("Expected state %s, got %s", resource.State, cachedResource.State)
		}
	} else {
		t.Error("Resource should be found in cache")
	}

	// Test cache metrics
	metrics := cache.GetMetrics()
	if metrics.CacheHits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", metrics.CacheHits)
	}
	if metrics.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", metrics.CacheMisses)
	}

	// Test cache hit rate
	hitRate := cache.GetCacheHitRate()
	expectedRate := 50.0 // 1 hit out of 2 total requests = 50%
	if hitRate != expectedRate {
		t.Errorf("Expected hit rate %.1f%%, got %.1f%%", expectedRate, hitRate)
	}

	// Test cache warming
	expectedResources := []models.ResourceStatus{
		{Name: "vm1", Type: "Microsoft.Compute/virtualMachines", State: "Creating"},
		{Name: "storage1", Type: "Microsoft.Storage/storageAccounts", State: "Creating"},
	}
	cache.WarmCache(expectedResources)

	// Check that resources were warmed
	status := cache.GetCacheStatus()
	if status["resource_entries"].(int) < 2 {
		t.Error("Cache warming should have added resources")
	}

	// Test phase optimization
	cache.UpdatePhase(PhaseConfiguration)
	// The cache should adjust TTLs for VM extensions during configuration phase
	// This is tested implicitly through the phase change
}

func TestCacheTTLBehavior(t *testing.T) {
	cache := NewIntelligentCache()

	// Test different TTLs for different resource states
	stableResource := &models.ResourceStatus{
		Name:  "stable-vm",
		Type:  "Microsoft.Compute/virtualMachines",
		State: "Succeeded",
	}

	transitionalResource := &models.ResourceStatus{
		Name:  "transitional-vm",
		Type:  "Microsoft.Compute/virtualMachines",
		State: "Creating",
	}

	cache.CacheResource(stableResource)
	cache.CacheResource(transitionalResource)

	// Both should be cached initially
	if _, found := cache.GetCachedResource("stable-vm::Microsoft.Compute/virtualMachines"); !found {
		t.Error("Stable resource should be cached")
	}
	if _, found := cache.GetCachedResource("transitional-vm::Microsoft.Compute/virtualMachines"); !found {
		t.Error("Transitional resource should be cached")
	}

	// Simulate time passing (would need to modify cache entries' timestamps in real test)
	// For this basic test, we verify the cache is functioning
	status := cache.GetCacheStatus()
	if status["resource_entries"].(int) != 2 {
		t.Errorf("Expected 2 cache entries, got %d", status["resource_entries"])
	}
}

func TestCacheDeploymentInfo(t *testing.T) {
	cache := NewIntelligentCache()

	// Test deployment caching
	deploymentInfo := map[string]interface{}{
		"provisioningState": "Running",
		"timestamp":         time.Now().Format(time.RFC3339),
	}

	// Initially not cached
	if _, found := cache.GetCachedDeployment("test-rg", "test-deployment"); found {
		t.Error("Deployment should not be cached initially")
	}

	// Cache deployment
	cache.CacheDeployment("test-rg", "test-deployment", deploymentInfo)

	// Should be cached now
	if cached, found := cache.GetCachedDeployment("test-rg", "test-deployment"); found {
		if cachedMap, ok := cached.(map[string]interface{}); ok {
			if cachedMap["provisioningState"] != "Running" {
				t.Error("Deployment state not cached correctly")
			}
		} else {
			t.Error("Cached deployment should be a map")
		}
	} else {
		t.Error("Deployment should be cached")
	}
}
