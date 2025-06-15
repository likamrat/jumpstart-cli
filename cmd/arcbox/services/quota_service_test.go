package services

import (
	"testing"

	"jumpstartcli/internal/azurecli"
)

// TestQuotaServiceCreation tests that QuotaService can be created properly
func TestQuotaServiceCreation(t *testing.T) {
	mockCLI := &azurecli.MockAzureCLI{}
	service := NewQuotaService(mockCLI)

	if service == nil {
		t.Fatal("Expected non-nil QuotaService")
	}

	if service.cli != mockCLI {
		t.Error("Expected QuotaService to use provided CLI")
	}

	if service.quotaCache == nil {
		t.Error("Expected quotaCache to be initialized")
	}
}

// TestGetFlavorSKUs tests the GetFlavorSKUs method
func TestGetFlavorSKUs(t *testing.T) {
	mockCLI := &azurecli.MockAzureCLI{}
	service := NewQuotaService(mockCLI)

	tests := []struct {
		flavor   string
		expected []string
	}{
		{"itpro", []string{"Standard_D8s_v5"}},
		{"ITPro", []string{"Standard_D8s_v5"}},
		{"ITPRO", []string{"Standard_D8s_v5"}},
		{"devops", []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}},
		{"DevOps", []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}},
		{"dataops", []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}},
		{"DataOps", []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}},
		{"unknown", []string{}},
		{"", []string{}},
	}

	for _, tt := range tests {
		t.Run("flavor_"+tt.flavor, func(t *testing.T) {
			result := service.GetFlavorSKUs(tt.flavor)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d SKUs for flavor %s, got %d", len(tt.expected), tt.flavor, len(result))
				return
			}
			for i, sku := range result {
				if sku != tt.expected[i] {
					t.Errorf("Expected SKU %s at index %d for flavor %s, got %s", tt.expected[i], i, tt.flavor, sku)
				}
			}
		})
	}
}

// TestClearQuotaCache tests the ClearQuotaCache method
func TestClearQuotaCache(t *testing.T) {
	mockCLI := &azurecli.MockAzureCLI{}
	service := NewQuotaService(mockCLI)

	// Add some data to the cache
	service.quotaCache["test"] = []azurecli.VMUsageInfo{{
		Name:         map[string]string{"value": "test"},
		CurrentValue: 0,
		Limit:        10,
		Unit:         "cores",
	}}

	if len(service.quotaCache) == 0 {
		t.Error("Expected cache to contain test data")
	}

	// Clear the cache
	service.ClearQuotaCache()

	if len(service.quotaCache) != 0 {
		t.Error("Expected cache to be empty after clearing")
	}
}
