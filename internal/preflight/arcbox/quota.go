// quota.go - ArcBox-specific quota checking functionality
package arcbox

import (
	"fmt"
	"strings"

	"jumpstartcli/internal/azurecli"
)

// QuotaCheckResult represents the result of checking quota for a single SKU
type QuotaCheckResult struct {
	SKU          string
	Required     int
	Current      int
	Limit        int
	Available    int
	QuotaOK      bool
	SKUAvailable bool
	CanDeploy    bool
	Details      string
}

// CheckQuotaForSKU checks quota for a specific SKU using the Azure CLI wrapper
func CheckQuotaForSKU(cli azurecli.AzureCLI, sku string, required int, region, flavor string) QuotaCheckResult {
	result := QuotaCheckResult{
		SKU:      sku,
		Required: required,
	}

	// Get quota data for the region
	usages, err := cli.ListVMUsage(region)
	if err != nil {
		result.Details = fmt.Sprintf("Failed to get quota data: %v", err)
		return result
	}

	// Map SKU to its quota family name
	familyName := mapSKUToFamilyQuotaName(sku)

	// Look for the quota usage entry
	found := false
	for _, usage := range usages {
		val, hasVal := usage.Name["value"]
		localizedValue, hasLocalized := usage.Name["localizedValue"]

		if !hasVal || !hasLocalized {
			continue
		}

		// Normalize for comparison
		familyNorm := strings.ReplaceAll(strings.ToLower(familyName), " ", "")
		valNorm := strings.ReplaceAll(strings.ToLower(val), " ", "")
		localizedNorm := strings.ReplaceAll(strings.ToLower(localizedValue), " ", "")

		if familyName != "" && (valNorm == familyNorm || localizedNorm == familyNorm) {
			result.Current = usage.CurrentValue
			result.Limit = usage.Limit
			result.Available = usage.Limit - usage.CurrentValue
			result.QuotaOK = result.Available >= required
			found = true
			break
		}
	}

	// Fallback: try total regional vCPU quota
	if !found {
		for _, usage := range usages {
			val, hasVal := usage.Name["value"]
			localizedValue, hasLocalized := usage.Name["localizedValue"]

			if !hasVal || !hasLocalized {
				continue
			}

			valNorm := strings.ReplaceAll(strings.ToLower(val), " ", "")
			localizedNorm := strings.ReplaceAll(strings.ToLower(localizedValue), " ", "")

			if strings.Contains(valNorm, "totalregionalvcpu") || strings.Contains(localizedNorm, "totalregionalvcpu") {
				result.Current = usage.CurrentValue
				result.Limit = usage.Limit
				result.Available = usage.Limit - usage.CurrentValue
				result.QuotaOK = result.Available >= required
				found = true
				break
			}
		}
	}

	if !found {
		result.Details = "Quota information not found"
		return result
	}

	// If we have quota information, the SKU family is available in the region
	// Azure wouldn't provide quota limits for unavailable SKU families
	result.SKUAvailable = true

	// Set deployment feasibility based on quota sufficiency only
	result.CanDeploy = result.QuotaOK

	// Set appropriate details message
	if result.QuotaOK {
		result.Details = "Ready to deploy"
	} else {
		result.Details = fmt.Sprintf("Need %d more vCPU", result.Required-result.Available)
	}

	return result
}

// CheckBatchSKUAvailability checks availability for multiple SKUs efficiently
func CheckBatchSKUAvailability(cli azurecli.AzureCLI, skus []string, region string) ([]string, error) {
	if len(skus) == 0 {
		return []string{}, nil
	}

	var unavailableSKUs []string

	// Check each SKU individually using the Azure CLI wrapper
	for _, sku := range skus {
		available, err := cli.CheckSKUAvailability(sku, region)
		if err != nil {
			// If we can't check, assume it's unavailable for safety
			unavailableSKUs = append(unavailableSKUs, sku)
			continue
		}

		if !available {
			unavailableSKUs = append(unavailableSKUs, sku)
		}
	}

	return unavailableSKUs, nil
}

// RunQuotaChecks performs comprehensive quota checking for ArcBox flavors
func RunQuotaChecks(cli azurecli.AzureCLI, location, flavor string, subscription string) ([]QuotaCheckResult, error) {
	// First, verify Azure CLI authentication by checking current subscription
	if _, err := cli.GetCurrentSubscription(); err != nil {
		return nil, fmt.Errorf("Azure CLI authentication required: %v", err)
	}

	// Get SKUs for the flavor(s)
	var allSKUs []string
	if flavor == "all" {
		// Get SKUs for all flavors
		for _, f := range []string{"ITPro", "DevOps", "DataOps"} {
			skus := getFlavorSKUs(f)
			allSKUs = append(allSKUs, skus...)
		}
		// Remove duplicates
		uniqueSKUs := make(map[string]bool)
		for _, sku := range allSKUs {
			uniqueSKUs[sku] = true
		}
		allSKUs = []string{}
		for sku := range uniqueSKUs {
			allSKUs = append(allSKUs, sku)
		}
	} else {
		allSKUs = getFlavorSKUs(flavor)
	}

	if len(allSKUs) == 0 {
		return nil, fmt.Errorf("no SKUs found for flavor: %s", flavor)
	}

	// Check quota for each SKU
	var results []QuotaCheckResult
	for _, sku := range allSKUs {
		required := getRequiredVCPUForSKU(sku)
		result := CheckQuotaForSKU(cli, sku, required, location, flavor)

		// Check if the quota check failed due to an error (e.g. authentication)
		if strings.Contains(result.Details, "Failed to get quota data") {
			return nil, fmt.Errorf("quota check failed for SKU %s: %s", sku, result.Details)
		}

		// Use the results from CheckQuotaForSKU as-is
		// If quota exists for a SKU family, the SKU is deployable
		results = append(results, result)
	}

	return results, nil
}

// getFlavorSKUs returns the VM SKUs required for a specific ArcBox flavor
func getFlavorSKUs(flavor string) []string {
	switch strings.ToLower(flavor) {
	case "itpro":
		return []string{"Standard_D8s_v5"}
	case "devops":
		return []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}
	case "dataops":
		return []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}
	default:
		return []string{}
	}
}

// getRequiredVCPUForSKU returns the vCPU requirement for a VM SKU
func getRequiredVCPUForSKU(sku string) int {
	vcpuMap := map[string]int{
		"Standard_D8s_v5": 8,
		"Standard_D8s_v4": 8,
		"Standard_B2ms":   2,
		"Standard_B4ms":   4,
		"Standard_B8ms":   8,
	}
	if vcpu, ok := vcpuMap[sku]; ok {
		return vcpu
	}
	return 1 // Default fallback
}

// mapSKUToFamilyQuotaName maps a VM SKU to its Azure vCPU family quota name
func mapSKUToFamilyQuotaName(sku string) string {
	sku = strings.TrimPrefix(sku, "Standard_")
	parts := strings.Split(sku, "_")

	var main, ver string
	if len(parts) >= 2 {
		// Standard pattern: D8s_v5 -> main=D8s, ver=v5
		main = parts[0]
		ver = parts[1]
	} else if len(parts) == 1 {
		// B-series pattern: B2ms -> main=B2ms, ver=""
		main = parts[0]
		ver = ""

		// Special handling for B-series SKUs - they all map to "Standard BS Family vCPUs"
		if strings.HasPrefix(strings.ToUpper(main), "B") {
			return "Standard BS Family vCPUs"
		}
	} else {
		return ""
	}

	// Remove digits from main part, keep only letters
	letters := ""
	for _, r := range main {
		if r >= '0' && r <= '9' {
			continue
		}
		letters += string(r)
	}

	// If ends with 's', keep it (e.g. D8s → Ds)
	if strings.HasSuffix(main, "s") && !strings.HasSuffix(letters, "s") {
		letters += "s"
	}

	// Handle case where letters is empty (e.g., SKU with only digits)
	if letters == "" {
		return ""
	}

	family := "Standard " + strings.ToUpper(letters[:1]) + letters[1:] + ver + " Family vCPUs"
	return family
}
