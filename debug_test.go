package main

import (
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/resourceproviders"
	"testing"
)

func TestDebugMockCLI(t *testing.T) {
	// Test 1: Initialize mock with explicit map
	mockCLI := &azurecli.MockAzureCLI{
		RegisteredProviders: make(map[string]bool),
	}

	// Test that we can assign to the map
	mockCLI.RegisteredProviders["Microsoft.Test"] = true
	t.Logf("Test 1 passed: assigned to RegisteredProviders map")

	// Test 2: Get ArcBox providers
	config := resourceproviders.GetArcBoxProviders()
	t.Logf("Test 2 passed: got %d providers", len(config.RequiredProviders))

	// Test 3: Try to assign all providers
	for _, provider := range config.RequiredProviders {
		mockCLI.RegisteredProviders[provider] = true
		t.Logf("Assigned provider: %s", provider)
	}

	t.Logf("Test 3 passed: assigned all providers")
}
