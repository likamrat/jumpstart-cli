package models

// This file contains quota-related types and structs for ArcBox deployments

import "time"

// QuotaCheckResult represents the result of checking quota for a single VM SKU
type QuotaCheckResult struct {
	SKU          string `json:"sku"`
	Required     int    `json:"required"`
	Current      int    `json:"current"`
	Limit        int    `json:"limit"`
	Available    int    `json:"available"`
	QuotaOK      bool   `json:"quotaOk"`
	SKUAvailable bool   `json:"skuAvailable"`
	CanDeploy    bool   `json:"canDeploy"`
	Details      string `json:"details"`
}

// FlavorSKUMapping represents the VM SKU requirements for a specific ArcBox flavor
type FlavorSKUMapping struct {
	Flavor string   `json:"flavor"`
	SKUs   []string `json:"skus"`
}

// SKUSpecification represents detailed specifications for a VM SKU
type SKUSpecification struct {
	Name        string            `json:"name"`
	VCPUCount   int               `json:"vcpuCount"`
	MemoryGB    int               `json:"memoryGb,omitempty"`
	QuotaFamily string            `json:"quotaFamily"`
	Tier        string            `json:"tier,omitempty"`
	Size        string            `json:"size,omitempty"`
	Family      string            `json:"family,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// QuotaUsageInfo represents Azure VM usage/quota information for a specific quota family
type QuotaUsageInfo struct {
	Name         map[string]string `json:"name"`
	CurrentValue int               `json:"currentValue"`
	Limit        int               `json:"limit"`
	Available    int               `json:"available"`
	Unit         string            `json:"unit"`
	Region       string            `json:"region,omitempty"`
}

// QuotaValidationSummary represents a summary of quota validation for an ArcBox deployment
type QuotaValidationSummary struct {
	Flavor          string             `json:"flavor"`
	Region          string             `json:"region"`
	SubscriptionID  string             `json:"subscriptionId"`
	TotalRequired   int                `json:"totalRequired"`
	CanDeployFlavor bool               `json:"canDeployFlavor"`
	CheckedAt       time.Time          `json:"checkedAt"`
	Results         []QuotaCheckResult `json:"results"`
	UnavailableSKUs []string           `json:"unavailableSkus,omitempty"`
	Recommendations []string           `json:"recommendations,omitempty"`
}

// QuotaIncreaseRequest represents a request for quota increase
type QuotaIncreaseRequest struct {
	SubscriptionID string   `json:"subscriptionId"`
	Region         string   `json:"region"`
	QuotaFamily    string   `json:"quotaFamily"`
	CurrentLimit   int      `json:"currentLimit"`
	RequestedLimit int      `json:"requestedLimit"`
	Justification  string   `json:"justification"`
	SKUs           []string `json:"skus,omitempty"`
}

// RegionQuotaAvailability represents quota availability across multiple regions
type RegionQuotaAvailability struct {
	Flavor             string                            `json:"flavor"`
	SubscriptionID     string                            `json:"subscriptionId"`
	CheckedAt          time.Time                         `json:"checkedAt"`
	RegionAvailability map[string]QuotaValidationSummary `json:"regionAvailability"`
	RecommendedRegions []string                          `json:"recommendedRegions"`
}
