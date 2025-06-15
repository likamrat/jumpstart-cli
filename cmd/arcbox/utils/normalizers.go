package utils

import "strings"

// This file contains normalization functions for ArcBox deployment parameters
// and configuration values to ensure consistent formatting and casing

// NormalizeFlavorCase normalizes ArcBox flavor names to proper case
func NormalizeFlavorCase(flavor string) string {
	switch strings.ToLower(flavor) {
	case "itpro", "it-pro":
		return "ITPro"
	case "devops", "dev-ops":
		return "DevOps"
	case "dataops", "data-ops":
		return "DataOps"
	case "all":
		return "all"
	default:
		return flavor
	}
}

// NormalizeSqlServerEditionCase normalizes SQL Server edition names to proper case
func NormalizeSqlServerEditionCase(edition string) string {
	switch strings.ToLower(edition) {
	case "developer":
		return "Developer"
	case "standard":
		return "Standard"
	case "enterprise":
		return "Enterprise"
	default:
		return edition
	}
}

// NormalizeBastionSkuCase normalizes Bastion SKU names to proper case
func NormalizeBastionSkuCase(sku string) string {
	switch strings.ToLower(sku) {
	case "basic":
		return "Basic"
	case "standard":
		return "Standard"
	case "developer":
		return "Developer"
	default:
		return sku
	}
}
