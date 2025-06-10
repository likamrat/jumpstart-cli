package azurecli

import (
	"encoding/json"
	"fmt"
	"os/exec"
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

// AzureCLI defines the interface for Azure CLI operations
type AzureCLI interface {
	// Subscription operations
	GetCurrentSubscription() (*SubscriptionInfo, error)
	GetSubscription(subscriptionID string) (*SubscriptionInfo, error)
	ListSubscriptions() ([]SubscriptionInfo, error)
	SetSubscription(subscriptionID string) error

	// Account operations
	IsLoggedIn() bool
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
