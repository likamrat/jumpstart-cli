package azurecli

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// SubscriptionInfo represents Azure subscription information
type SubscriptionInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TenantID  string `json:"tenantId,omitempty"`
	State     string `json:"state,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
	User      *struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"user,omitempty"`
}

// VMUsageInfo represents Azure VM usage/quota information
type VMUsageInfo struct {
	Name         map[string]string `json:"name"`
	CurrentValue int               `json:"currentValue"`
	Limit        int               `json:"limit"`
	Unit         string            `json:"unit"`
}

// SKUInfo represents Azure VM SKU information
type SKUInfo struct {
	Name     string            `json:"name"`
	Location string            `json:"location,omitempty"`
	Tier     string            `json:"tier,omitempty"`
	Size     string            `json:"size,omitempty"`
	Family   string            `json:"family,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// AzureCLI defines the interface for Azure CLI operations
type AzureCLI interface {
	// Subscription operations
	GetCurrentSubscription() (*SubscriptionInfo, error)
	GetSubscription(subscriptionID string) (*SubscriptionInfo, error)
	ListSubscriptions() ([]SubscriptionInfo, error)
	SetSubscription(subscriptionID string) error

	// Account operations
	IsLoggedIn() bool

	// VM Quota operations
	ListVMUsage(region string) ([]VMUsageInfo, error)
	ListVMSKUs(region string) ([]SKUInfo, error)
	CheckSKUAvailability(sku, region string) (bool, error)
}

// RealAzureCLI implements the AzureCLI interface using actual Azure CLI commands
type RealAzureCLI struct{}

// NewAzureCLI creates a new instance of the real Azure CLI implementation
func NewAzureCLI() AzureCLI {
	return &RealAzureCLI{}
}

// GetCurrentSubscription retrieves the current Azure subscription
func (r *RealAzureCLI) GetCurrentSubscription() (*SubscriptionInfo, error) {
	cmd := exec.Command("az", "account", "show", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current subscription: %v", err)
	}

	var sub SubscriptionInfo
	if err := json.Unmarshal(output, &sub); err != nil {
		return nil, fmt.Errorf("failed to parse current subscription: %v", err)
	}

	return &sub, nil
}

// GetSubscription retrieves information about a specific subscription
func (r *RealAzureCLI) GetSubscription(subscriptionID string) (*SubscriptionInfo, error) {
	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID cannot be empty")
	}

	cmd := exec.Command("az", "account", "show", "--subscription", subscriptionID, "--query", "{id:id,name:name}", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("subscription '%s' not found or inaccessible: %v", subscriptionID, err)
	}

	var sub SubscriptionInfo
	if err := json.Unmarshal(output, &sub); err != nil {
		return nil, fmt.Errorf("failed to parse subscription information: %v", err)
	}

	return &sub, nil
}

// ListSubscriptions retrieves all accessible Azure subscriptions
func (r *RealAzureCLI) ListSubscriptions() ([]SubscriptionInfo, error) {
	cmd := exec.Command("az", "account", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %v", err)
	}

	var subs []SubscriptionInfo
	if err := json.Unmarshal(output, &subs); err != nil {
		return nil, fmt.Errorf("failed to parse subscriptions: %v", err)
	}

	return subs, nil
}

// SetSubscription sets the current Azure subscription
func (r *RealAzureCLI) SetSubscription(subscriptionID string) error {
	if subscriptionID == "" {
		return fmt.Errorf("subscription ID cannot be empty")
	}

	cmd := exec.Command("az", "account", "set", "--subscription", subscriptionID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set subscription: %v", err)
	}

	return nil
}

// IsLoggedIn checks if the user is logged in to Azure CLI
func (r *RealAzureCLI) IsLoggedIn() bool {
	cmd := exec.Command("az", "account", "show")
	err := cmd.Run()
	return err == nil
}

// ListVMUsage retrieves VM quota/usage information for a specific region
func (r *RealAzureCLI) ListVMUsage(region string) ([]VMUsageInfo, error) {
	if region == "" {
		return nil, fmt.Errorf("region cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "vm", "list-usage", "--location", region, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout while retrieving VM usage for region %s", region)
		}
		return nil, fmt.Errorf("failed to get VM usage for region %s: %v", region, err)
	}

	var usages []VMUsageInfo
	if err := json.Unmarshal(output, &usages); err != nil {
		return nil, fmt.Errorf("failed to parse VM usage data: %v", err)
	}

	return usages, nil
}

// ListVMSKUs retrieves available VM SKUs for a specific region
func (r *RealAzureCLI) ListVMSKUs(region string) ([]SKUInfo, error) {
	if region == "" {
		return nil, fmt.Errorf("region cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "vm", "list-skus", "--resource-type", "virtualMachines", "--location", region, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout while retrieving VM SKUs for region %s", region)
		}
		return nil, fmt.Errorf("failed to get VM SKUs for region %s: %v", region, err)
	}

	var skus []SKUInfo
	if err := json.Unmarshal(output, &skus); err != nil {
		return nil, fmt.Errorf("failed to parse VM SKU data: %v", err)
	}

	return skus, nil
}

// CheckSKUAvailability checks if a specific VM SKU is available in a region
func (r *RealAzureCLI) CheckSKUAvailability(sku, region string) (bool, error) {
	if sku == "" {
		return false, fmt.Errorf("SKU cannot be empty")
	}
	if region == "" {
		return false, fmt.Errorf("region cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use a targeted query to check for the specific SKU
	query := fmt.Sprintf("[?name=='%s'].name", sku)
	cmd := exec.CommandContext(ctx, "az", "vm", "list-skus", "--resource-type", "virtualMachines", "--location", region, "--query", query, "-o", "tsv")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return false, fmt.Errorf("timeout while checking SKU %s availability in region %s", sku, region)
		}
		return false, fmt.Errorf("failed to check SKU %s availability in region %s: %v", sku, region, err)
	}

	result := strings.TrimSpace(string(output))
	return result == sku, nil
}
