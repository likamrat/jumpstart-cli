package azurecli

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
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

// UnmarshalJSON custom unmarshaler to handle string to int conversion for Azure CLI output
func (v *VMUsageInfo) UnmarshalJSON(data []byte) error {
	// Define a temporary struct with string fields
	var temp struct {
		Name         map[string]string `json:"name"`
		CurrentValue string            `json:"currentValue"`
		Limit        string            `json:"limit"`
		Unit         string            `json:"unit"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Convert strings to integers
	currentValue, err := strconv.Atoi(temp.CurrentValue)
	if err != nil {
		return fmt.Errorf("failed to convert currentValue '%s' to int: %v", temp.CurrentValue, err)
	}

	limit, err := strconv.Atoi(temp.Limit)
	if err != nil {
		return fmt.Errorf("failed to convert limit '%s' to int: %v", temp.Limit, err)
	}

	// Set the converted values
	v.Name = temp.Name
	v.CurrentValue = currentValue
	v.Limit = limit
	v.Unit = temp.Unit

	return nil
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

// ResourceGroupInfo represents Azure resource group information
type ResourceGroupInfo struct {
	Name       string                 `json:"name"`
	Location   string                 `json:"location"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// ResourceInfo represents Azure resource information
type ResourceInfo struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Location   string                 `json:"location,omitempty"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// DeploymentInfo represents Azure deployment information
type DeploymentInfo struct {
	Name              string                 `json:"name"`
	Properties        map[string]interface{} `json:"properties,omitempty"`
	ProvisioningState string                 `json:"provisioningState,omitempty"`
}

// VMInfo represents Azure VM information
type VMInfo struct {
	Name          string `json:"name"`
	ResourceGroup string `json:"resourceGroup,omitempty"`
	Location      string `json:"location,omitempty"`
	PowerState    string `json:"powerState,omitempty"`
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
	ListLocations() ([]string, error)

	// VM Quota operations
	ListVMUsage(region string) ([]VMUsageInfo, error)
	ListVMSKUs(region string) ([]SKUInfo, error)
	CheckSKUAvailability(sku, region string) (bool, error)

	// Resource provider operations
	CheckProviderRegistration(provider string) (bool, error)
	RegisterProvider(provider string) error

	// Resource group operations
	CheckResourceGroupExists(name string) (bool, error)
	ListResourceGroups() ([]ResourceGroupInfo, error)
	CreateResourceGroup(name, location string) error
	DeleteResourceGroup(name string, noWait bool) error

	// Resource operations
	ListResources(resourceGroup string) ([]ResourceInfo, error)
	GetResource(resourceID string) (*ResourceInfo, error)

	// Deployment operations
	ListDeployments(resourceGroup string) ([]DeploymentInfo, error)
	GetDeployment(resourceGroup, deploymentName string) (*DeploymentInfo, error)
	CreateDeployment(resourceGroup, deploymentName, templateURI string, parameters []string, noWait bool) error

	// VM operations
	ListVMs(resourceGroup string) ([]VMInfo, error)
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

// ListLocations retrieves available Azure locations
func (r *RealAzureCLI) ListLocations() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "account", "list-locations", "--query", "[].name", "-o", "tsv")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout while listing Azure locations")
		}
		return nil, fmt.Errorf("failed to list Azure locations: %v", err)
	}

	locations := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, loc := range locations {
		if trimmed := strings.TrimSpace(loc); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, nil
}

// ListVMUsage retrieves VM quota/usage information for a specific region
// CRITICAL: This function makes a fresh Azure CLI call every time - no caching
// Each call provides real-time quota information directly from Azure APIs
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get detailed SKU information including restrictions
	query := fmt.Sprintf("[?name=='%s']", sku)
	cmd := exec.CommandContext(ctx, "az", "vm", "list-skus", "--resource-type", "virtualMachines", "--location", region, "--query", query, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return false, fmt.Errorf("timeout while checking SKU %s availability in region %s", sku, region)
		}
		return false, fmt.Errorf("failed to check SKU %s availability in region %s: %v", sku, region, err)
	}

	var skus []map[string]interface{}
	if err := json.Unmarshal(output, &skus); err != nil {
		return false, fmt.Errorf("failed to parse SKU availability data: %v", err)
	}

	if len(skus) == 0 {
		return false, nil // SKU not found
	}

	// Check if the SKU has restrictions that make it unavailable
	skuInfo := skus[0]
	if restrictions, ok := skuInfo["restrictions"].([]interface{}); ok && len(restrictions) > 0 {
		// If there are restrictions, check if any make it unavailable for subscription
		for _, restriction := range restrictions {
			if restrictionMap, ok := restriction.(map[string]interface{}); ok {
				if reasonCode, ok := restrictionMap["reasonCode"].(string); ok {
					if restrictionType, ok := restrictionMap["type"].(string); ok {
						// Only block if it's a location-level restriction
						// Zone-level restrictions don't prevent deployment, just limit zone choices
						if reasonCode == "NotAvailableForSubscription" && restrictionType == "Location" {
							return false, nil // SKU truly unavailable for this subscription at location level
						}
						// Zone restrictions are acceptable - SKU is still deployable
					}
				}
			}
		}
	}

	return true, nil // SKU exists and is available
}

// CheckProviderRegistration checks if a resource provider is registered
func (r *RealAzureCLI) CheckProviderRegistration(provider string) (bool, error) {
	if provider == "" {
		return false, fmt.Errorf("provider cannot be empty")
	}

	// Add timeout context to prevent hanging on Azure CLI calls
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "provider", "show", "--namespace", provider, "-o", "tsv", "--query", "registrationState")
	out, err := cmd.Output()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return false, fmt.Errorf("timeout checking provider %s (Azure CLI took too long)", provider)
		}
		return false, fmt.Errorf("failed to check provider %s: %v", provider, err)
	}

	state := strings.TrimSpace(string(out))
	return state == "Registered", nil
}

// RegisterProvider registers a resource provider
func (r *RealAzureCLI) RegisterProvider(provider string) error {
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}

	// Add timeout context to prevent hanging on Azure CLI calls
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "provider", "register", "--namespace", provider)
	err := cmd.Run()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout registering provider %s (Azure CLI took too long)", provider)
		}
		return fmt.Errorf("failed to register provider %s: %v", provider, err)
	}

	return nil
}

// CheckResourceGroupExists checks if a resource group exists
func (r *RealAzureCLI) CheckResourceGroupExists(name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("resource group name cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "group", "exists", "--name", name, "-o", "tsv")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return false, fmt.Errorf("timeout while checking resource group %s existence", name)
		}
		return false, fmt.Errorf("failed to check resource group %s existence: %v", name, err)
	}

	exists := strings.TrimSpace(string(output))
	return exists == "true", nil
}

// ListResourceGroups lists all resource groups in the current subscription
func (r *RealAzureCLI) ListResourceGroups() ([]ResourceGroupInfo, error) {
	cmd := exec.Command("az", "group", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list resource groups: %v", err)
	}

	var groups []ResourceGroupInfo
	if err := json.Unmarshal(output, &groups); err != nil {
		return nil, fmt.Errorf("failed to parse resource group data: %v", err)
	}

	return groups, nil
}

// CreateResourceGroup creates a new resource group
func (r *RealAzureCLI) CreateResourceGroup(name, location string) error {
	if name == "" {
		return fmt.Errorf("resource group name cannot be empty")
	}
	if location == "" {
		return fmt.Errorf("location cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "group", "create", "--name", name, "--location", location, "-o", "none")
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout while creating resource group %s", name)
		}
		return fmt.Errorf("failed to create resource group %s: %v", name, err)
	}

	return nil
}

// DeleteResourceGroup deletes a resource group
func (r *RealAzureCLI) DeleteResourceGroup(name string, noWait bool) error {
	if name == "" {
		return fmt.Errorf("resource group name cannot be empty")
	}

	args := []string{"group", "delete", "--name", name, "--yes"}
	if noWait {
		args = append(args, "--no-wait")
	}

	cmd := exec.Command("az", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete resource group %s: %v", name, err)
	}

	return nil
}

// ListResources lists all resources in a resource group
func (r *RealAzureCLI) ListResources(resourceGroup string) ([]ResourceInfo, error) {
	if resourceGroup == "" {
		return nil, fmt.Errorf("resource group cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "resource", "list", "--resource-group", resourceGroup, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout while listing resources in resource group %s", resourceGroup)
		}
		return nil, fmt.Errorf("failed to list resources in resource group %s: %v", resourceGroup, err)
	}

	var resources []ResourceInfo
	if err := json.Unmarshal(output, &resources); err != nil {
		return nil, fmt.Errorf("failed to parse resource data: %v", err)
	}

	return resources, nil
}

// GetResource retrieves a specific resource by its ID
func (r *RealAzureCLI) GetResource(resourceID string) (*ResourceInfo, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("resource ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "resource", "show", "--ids", resourceID, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout while retrieving resource %s", resourceID)
		}
		return nil, fmt.Errorf("failed to retrieve resource %s: %v", resourceID, err)
	}

	var resource ResourceInfo
	if err := json.Unmarshal(output, &resource); err != nil {
		return nil, fmt.Errorf("failed to parse resource data: %v", err)
	}

	return &resource, nil
}

// ListDeployments lists all deployments in a resource group
func (r *RealAzureCLI) ListDeployments(resourceGroup string) ([]DeploymentInfo, error) {
	if resourceGroup == "" {
		return nil, fmt.Errorf("resource group cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "deployment", "group", "list", "--resource-group", resourceGroup, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout while listing deployments in resource group %s", resourceGroup)
		}
		return nil, fmt.Errorf("failed to list deployments in resource group %s: %v", resourceGroup, err)
	}

	var deployments []DeploymentInfo
	if err := json.Unmarshal(output, &deployments); err != nil {
		return nil, fmt.Errorf("failed to parse deployment data: %v", err)
	}

	return deployments, nil
}

// GetDeployment retrieves a specific deployment by its name
func (r *RealAzureCLI) GetDeployment(resourceGroup, deploymentName string) (*DeploymentInfo, error) {
	if resourceGroup == "" || deploymentName == "" {
		return nil, fmt.Errorf("resource group and deployment name cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "deployment", "group", "show", "--resource-group", resourceGroup, "--name", deploymentName, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout while retrieving deployment %s in resource group %s", deploymentName, resourceGroup)
		}
		return nil, fmt.Errorf("failed to retrieve deployment %s in resource group %s: %v", deploymentName, resourceGroup, err)
	}

	var deployment DeploymentInfo
	if err := json.Unmarshal(output, &deployment); err != nil {
		return nil, fmt.Errorf("failed to parse deployment data: %v", err)
	}

	return &deployment, nil
}

// CreateDeployment creates a new deployment in a resource group
func (r *RealAzureCLI) CreateDeployment(resourceGroup, deploymentName, templateURI string, parameters []string, noWait bool) error {
	if resourceGroup == "" || deploymentName == "" || templateURI == "" {
		return fmt.Errorf("resource group, deployment name, and template URI cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 minutes for deployment start
	defer cancel()

	args := []string{"deployment", "group", "create", "--resource-group", resourceGroup, "--name", deploymentName}

	// Add template URI or template file
	if strings.HasPrefix(templateURI, "http://") || strings.HasPrefix(templateURI, "https://") {
		args = append(args, "--template-uri", templateURI)
	} else {
		args = append(args, "--template-file", templateURI)
	}

	// Add parameters if provided
	if len(parameters) > 0 {
		args = append(args, "--parameters")
		args = append(args, parameters...)
	}

	// Add no-wait flag if specified
	if noWait {
		args = append(args, "--no-wait")
	}

	cmd := exec.CommandContext(ctx, "az", args...)
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout while creating deployment %s in resource group %s", deploymentName, resourceGroup)
		}
		return fmt.Errorf("failed to create deployment %s in resource group %s: %v", deploymentName, resourceGroup, err)
	}

	return nil
}

// ListVMs lists all VMs in a resource group
func (r *RealAzureCLI) ListVMs(resourceGroup string) ([]VMInfo, error) {
	if resourceGroup == "" {
		return nil, fmt.Errorf("resource group cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "vm", "list", "--resource-group", resourceGroup, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout while listing VMs in resource group %s", resourceGroup)
		}
		return nil, fmt.Errorf("failed to list VMs in resource group %s: %v", resourceGroup, err)
	}

	var vms []VMInfo
	if err := json.Unmarshal(output, &vms); err != nil {
		return nil, fmt.Errorf("failed to parse VM data: %v", err)
	}

	return vms, nil
}
