package azurecli

import (
	"fmt"
)

// MockAzureCLI implements the AzureCLI interface for testing
type MockAzureCLI struct {
	// Control mock behavior
	IsLoggedInResult bool
	ShouldFailLogin  bool

	// Mock data
	CurrentSubscription *SubscriptionInfo
	Subscriptions       []SubscriptionInfo
	Locations           []string
	VMUsages            map[string][]VMUsageInfo // region -> usage data
	VMSKUs              map[string][]SKUInfo     // region -> SKU data
	AvailableSKUs       map[string][]string      // region -> available SKU names

	// Resource provider mock data
	RegisteredProviders map[string]bool // provider -> registration status

	// Resource group mock data
	ResourceGroups      []ResourceGroupInfo
	ResourceGroupExists map[string]bool             // resource group name -> exists
	Resources           map[string][]ResourceInfo   // resource group -> resources
	Deployments         map[string][]DeploymentInfo // resource group -> deployments
	VMs                 map[string][]VMInfo         // resource group -> VMs
	SpecificResources   map[string]*ResourceInfo    // resource ID -> resource
	SpecificDeployments map[string]*DeploymentInfo  // deployment key -> deployment

	// Error injection
	GetCurrentSubscriptionError    error
	GetSubscriptionError           error
	ListSubscriptionsError         error
	SetSubscriptionError           error
	ListVMUsageError               error
	ListVMSKUsError                error
	CheckSKUAvailabilityError      error
	CheckProviderRegistrationError error
	RegisterProviderError          error
	ListLocationsError             error

	// Resource group error injection
	CheckResourceGroupExistsError error
	ListResourceGroupsError       error
	CreateResourceGroupError      error
	DeleteResourceGroupError      error
	ListResourcesError            error
	GetResourceError              error
	ListDeploymentsError          error
	GetDeploymentError            error
	CreateDeploymentError         error
	ListVMsError                  error

	// Call tracking
	GetCurrentSubscriptionCalled        bool
	GetSubscriptionCalled               bool
	GetSubscriptionCalledWith           string
	ListSubscriptionsCalled             bool
	SetSubscriptionCalled               bool
	SetSubscriptionCalledWith           string
	IsLoggedInCalled                    bool
	ListLocationsCalled                 bool
	ListVMUsageCalled                   bool
	ListVMUsageCalledWith               string
	ListVMSKUsCalled                    bool
	ListVMSKUsCalledWith                string
	CheckSKUAvailabilityCalled          bool
	CheckSKUAvailabilityCalledWith      map[string]string // "sku" and "region"
	CheckProviderRegistrationCalled     bool
	CheckProviderRegistrationCalledWith string
	RegisterProviderCalled              bool
	RegisterProviderCalledWith          string

	// Resource group call tracking
	CheckResourceGroupExistsCalled     bool
	CheckResourceGroupExistsCalledWith string
	ListResourceGroupsCalled           bool
	CreateResourceGroupCalled          bool
	CreateResourceGroupCalledWith      map[string]string // name -> location
	DeleteResourceGroupCalled          bool
	DeleteResourceGroupCalledWith      string
	ListResourcesCalled                bool
	ListResourcesCalledWith            string
	GetResourceCalled                  bool
	GetResourceCalledWith              string
	ListDeploymentsCalled              bool
	ListDeploymentsCalledWith          string
	GetDeploymentCalled                bool
	GetDeploymentCalledWith            string
	CreateDeploymentCalled             bool
	CreateDeploymentCalledWith         string
	ListVMsCalled                      bool
	ListVMsCalledWith                  string
}

// NewMockAzureCLI creates a new mock Azure CLI implementation with default test data
func NewMockAzureCLI() *MockAzureCLI {
	// Default VM usage data for testing
	defaultVMUsages := map[string][]VMUsageInfo{
		"eastus": {
			{
				Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
				CurrentValue: 8,
				Limit:        100,
				Unit:         "Count",
			},
			{
				Name:         map[string]string{"value": "standardBSFamily", "localizedValue": "Standard BS Family vCPUs"},
				CurrentValue: 2,
				Limit:        50,
				Unit:         "Count",
			},
		},
		"westus2": {
			{
				Name:         map[string]string{"value": "standardDSv5Family", "localizedValue": "Standard DSv5 Family vCPUs"},
				CurrentValue: 0,
				Limit:        100,
				Unit:         "Count",
			},
		},
	}

	// Default VM SKU data for testing
	defaultVMSKUs := map[string][]SKUInfo{
		"eastus": {
			{Name: "Standard_D8s_v5"},
			{Name: "Standard_B2ms"},
			{Name: "Standard_B4ms"},
		},
		"westus2": {
			{Name: "Standard_D8s_v5"},
			{Name: "Standard_B2ms"},
		},
	}

	// Default available SKUs
	defaultAvailableSKUs := map[string][]string{
		"eastus":  {"Standard_D8s_v5", "Standard_B2ms", "Standard_B4ms"},
		"westus2": {"Standard_D8s_v5", "Standard_B2ms"},
	}

	// Default resource groups for testing
	defaultResourceGroups := []ResourceGroupInfo{
		{Name: "test-rg-1", Location: "eastus"},
		{Name: "arcbox-rg", Location: "westus2"},
		{Name: "another-rg", Location: "centralus"},
	}

	// Default resources for testing
	defaultResources := map[string][]ResourceInfo{
		"arcbox-rg": {
			{ID: "/subscriptions/test-sub/resourceGroups/arcbox-rg/providers/Microsoft.Compute/virtualMachines/ArcBox-Client", Name: "ArcBox-Client", Type: "Microsoft.Compute/virtualMachines"},
			{ID: "/subscriptions/test-sub/resourceGroups/arcbox-rg/providers/Microsoft.KeyVault/vaults/ArcBox-KeyVault", Name: "ArcBox-KeyVault", Type: "Microsoft.KeyVault/vaults"},
		},
	}

	// Default deployments for testing
	defaultDeployments := map[string][]DeploymentInfo{
		"arcbox-rg": {
			{Name: "arcbox-deployment", ProvisioningState: "Succeeded"},
		},
	}

	// Default VMs for testing
	defaultVMs := map[string][]VMInfo{
		"arcbox-rg": {
			{Name: "ArcBox-Client", ResourceGroup: "arcbox-rg", Location: "westus2", PowerState: "VM running"},
		},
	}

	// Default registered resource providers for testing
	defaultRegisteredProviders := map[string]bool{
		"Microsoft.Compute":              true,
		"Microsoft.Network":              true,
		"Microsoft.Storage":              true,
		"Microsoft.Kubernetes":           false,
		"Microsoft.AzureArcData":         false,
		"Microsoft.ExtendedLocation":     false,
		"Microsoft.HybridConnectivity":   false,
		"Microsoft.OperationsManagement": true,
	}

	return &MockAzureCLI{
		IsLoggedInResult:               true,
		VMUsages:                       defaultVMUsages,
		VMSKUs:                         defaultVMSKUs,
		AvailableSKUs:                  defaultAvailableSKUs,
		RegisteredProviders:            defaultRegisteredProviders,
		ResourceGroups:                 defaultResourceGroups,
		ResourceGroupExists:            map[string]bool{"arcbox-rg": true, "test-rg-1": true, "another-rg": true},
		Resources:                      defaultResources,
		Deployments:                    defaultDeployments,
		VMs:                            defaultVMs,
		SpecificResources:              make(map[string]*ResourceInfo),
		SpecificDeployments:            make(map[string]*DeploymentInfo),
		CheckSKUAvailabilityCalledWith: make(map[string]string),
		CurrentSubscription: &SubscriptionInfo{
			ID:        "608937df-4e8f-4dc5-8bc6-16f30646ebd9",
			Name:      "Jumpstart Development EXT",
			TenantID:  "72f988bf-86f1-41af-91ab-2d7cd011db47",
			State:     "Enabled",
			IsDefault: true,
			User: &struct {
				Name string `json:"name"`
				Type string `json:"type"`
			}{
				Name: "test@microsoft.com",
				Type: "user",
			},
		},
		Subscriptions: []SubscriptionInfo{
			{
				ID:        "608937df-4e8f-4dc5-8bc6-16f30646ebd9",
				Name:      "Jumpstart Development EXT",
				IsDefault: true,
			},
			{
				ID:        "204898ee-cd13-4332-b9d4-55ca5c25496d",
				Name:      "ARC-Testing",
				IsDefault: false,
			},
		},
	}
}

// GetCurrentSubscription mocks getting the current subscription
func (m *MockAzureCLI) GetCurrentSubscription() (*SubscriptionInfo, error) {
	m.GetCurrentSubscriptionCalled = true

	if m.GetCurrentSubscriptionError != nil {
		return nil, m.GetCurrentSubscriptionError
	}

	if m.CurrentSubscription == nil {
		return nil, fmt.Errorf("no current subscription set")
	}

	return m.CurrentSubscription, nil
}

// GetSubscription mocks getting a specific subscription
func (m *MockAzureCLI) GetSubscription(subscriptionID string) (*SubscriptionInfo, error) {
	m.GetSubscriptionCalled = true
	m.GetSubscriptionCalledWith = subscriptionID

	if m.GetSubscriptionError != nil {
		return nil, m.GetSubscriptionError
	}

	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID cannot be empty")
	}

	// Find subscription by ID or name
	for _, sub := range m.Subscriptions {
		if sub.ID == subscriptionID || sub.Name == subscriptionID {
			return &sub, nil
		}
	}

	return nil, fmt.Errorf("subscription '%s' not found or inaccessible", subscriptionID)
}

// ListSubscriptions mocks listing all subscriptions
func (m *MockAzureCLI) ListSubscriptions() ([]SubscriptionInfo, error) {
	m.ListSubscriptionsCalled = true

	if m.ListSubscriptionsError != nil {
		return nil, m.ListSubscriptionsError
	}

	return m.Subscriptions, nil
}

// SetSubscription mocks setting the current subscription
func (m *MockAzureCLI) SetSubscription(subscriptionID string) error {
	m.SetSubscriptionCalled = true
	m.SetSubscriptionCalledWith = subscriptionID

	if m.SetSubscriptionError != nil {
		return m.SetSubscriptionError
	}

	if subscriptionID == "" {
		return fmt.Errorf("subscription ID cannot be empty")
	}

	// Find and set the subscription as current
	for i, sub := range m.Subscriptions {
		if sub.ID == subscriptionID {
			// Update all subscriptions to not be default
			for j := range m.Subscriptions {
				m.Subscriptions[j].IsDefault = false
			}
			// Set the target subscription as default
			m.Subscriptions[i].IsDefault = true
			m.CurrentSubscription = &m.Subscriptions[i]
			return nil
		}
	}

	return fmt.Errorf("subscription '%s' not found", subscriptionID)
}

// IsLoggedIn mocks checking if user is logged in
func (m *MockAzureCLI) IsLoggedIn() bool {
	m.IsLoggedInCalled = true

	if m.ShouldFailLogin {
		return false
	}

	return m.IsLoggedInResult
}

// ListLocations mocks getting available Azure locations
func (m *MockAzureCLI) ListLocations() ([]string, error) {
	m.ListLocationsCalled = true

	if m.ListLocationsError != nil {
		return nil, m.ListLocationsError
	}

	// Return default locations if none set
	if len(m.Locations) == 0 {
		return []string{"eastus", "westus2", "centralus", "westeurope", "southeastasia"}, nil
	}

	return m.Locations, nil
}

// ListVMUsage mocks getting VM quota/usage information for a region
func (m *MockAzureCLI) ListVMUsage(region string) ([]VMUsageInfo, error) {
	m.ListVMUsageCalled = true
	m.ListVMUsageCalledWith = region

	if m.ListVMUsageError != nil {
		return nil, m.ListVMUsageError
	}

	if region == "" {
		return nil, fmt.Errorf("region cannot be empty")
	}

	if usages, exists := m.VMUsages[region]; exists {
		return usages, nil
	}

	// Return empty list for unknown regions
	return []VMUsageInfo{}, nil
}

// ListVMSKUs mocks getting VM SKUs for a region
func (m *MockAzureCLI) ListVMSKUs(region string) ([]SKUInfo, error) {
	m.ListVMSKUsCalled = true
	m.ListVMSKUsCalledWith = region

	if m.ListVMSKUsError != nil {
		return nil, m.ListVMSKUsError
	}

	if region == "" {
		return nil, fmt.Errorf("region cannot be empty")
	}

	if skus, exists := m.VMSKUs[region]; exists {
		return skus, nil
	}

	// Return empty list for unknown regions
	return []SKUInfo{}, nil
}

// CheckSKUAvailability mocks checking if a SKU is available in a region
func (m *MockAzureCLI) CheckSKUAvailability(sku, region string) (bool, error) {
	m.CheckSKUAvailabilityCalled = true
	m.CheckSKUAvailabilityCalledWith["sku"] = sku
	m.CheckSKUAvailabilityCalledWith["region"] = region

	if m.CheckSKUAvailabilityError != nil {
		return false, m.CheckSKUAvailabilityError
	}

	if sku == "" {
		return false, fmt.Errorf("SKU cannot be empty")
	}
	if region == "" {
		return false, fmt.Errorf("region cannot be empty")
	}

	if availableSKUs, exists := m.AvailableSKUs[region]; exists {
		for _, availableSKU := range availableSKUs {
			if availableSKU == sku {
				return true, nil
			}
		}
	}

	return false, nil
}

// SetVMUsage mocks setting VM usage data for a region
func (m *MockAzureCLI) SetVMUsage(region string, usage []VMUsageInfo) {
	m.VMUsages[region] = usage
}

// SetVMUsageForRegion is an alias for SetVMUsage for test compatibility
func (m *MockAzureCLI) SetVMUsageForRegion(region string, usage []VMUsageInfo) {
	m.SetVMUsage(region, usage)
}

// SetVMSKUs mocks setting VM SKUs data for a region
func (m *MockAzureCLI) SetVMSKUs(region string, skus []SKUInfo) {
	m.VMSKUs[region] = skus
}

// SetVMSKUsForRegion is an alias for SetVMSKUs for test compatibility
func (m *MockAzureCLI) SetVMSKUsForRegion(region string, skus []SKUInfo) {
	m.SetVMSKUs(region, skus)
}

// SetSKUAvailability mocks setting SKU availability data for a region
func (m *MockAzureCLI) SetSKUAvailability(region string, skus []string) {
	m.AvailableSKUs[region] = skus
}

// SetAvailableSKUsForRegion is an alias for SetSKUAvailability for test compatibility
func (m *MockAzureCLI) SetAvailableSKUsForRegion(region string, skus []string) {
	m.SetSKUAvailability(region, skus)
}

// ClearVMData clears all VM-related data
func (m *MockAzureCLI) ClearVMData() {
	m.VMUsages = make(map[string][]VMUsageInfo)
	m.VMSKUs = make(map[string][]SKUInfo)
	m.AvailableSKUs = make(map[string][]string)
}

// Reset clears all call tracking flags
func (m *MockAzureCLI) Reset() {
	m.GetCurrentSubscriptionCalled = false
	m.GetSubscriptionCalled = false
	m.GetSubscriptionCalledWith = ""
	m.ListSubscriptionsCalled = false
	m.SetSubscriptionCalled = false
	m.SetSubscriptionCalledWith = ""
	m.IsLoggedInCalled = false
	m.ListLocationsCalled = false
	m.ListVMUsageCalled = false
	m.ListVMUsageCalledWith = ""
	m.ListVMSKUsCalled = false
	m.ListVMSKUsCalledWith = ""
	m.CheckSKUAvailabilityCalled = false
	m.CheckSKUAvailabilityCalledWith = make(map[string]string)
}

// Error injection helper methods for VM operations
func (m *MockAzureCLI) SetErrorForListVMUsage(err error) {
	m.ListVMUsageError = err
}

func (m *MockAzureCLI) SetErrorForListVMSKUs(err error) {
	m.ListVMSKUsError = err
}

func (m *MockAzureCLI) SetErrorForCheckSKUAvailability(err error) {
	m.CheckSKUAvailabilityError = err
}

// Missing helper methods for subscription operations
func (m *MockAzureCLI) SetErrorForGetCurrentSubscription(err error) {
	m.GetCurrentSubscriptionError = err
}

func (m *MockAzureCLI) SetErrorForGetSubscription(err error) {
	m.GetSubscriptionError = err
}

func (m *MockAzureCLI) SetErrorForListSubscriptions(err error) {
	m.ListSubscriptionsError = err
}

func (m *MockAzureCLI) SetErrorForSetSubscription(err error) {
	m.SetSubscriptionError = err
}

func (m *MockAzureCLI) SetErrorForListLocations(err error) {
	m.ListLocationsError = err
}

func (m *MockAzureCLI) ClearSubscriptions() {
	m.Subscriptions = []SubscriptionInfo{}
	m.CurrentSubscription = nil
}

func (m *MockAzureCLI) AddSubscription(sub SubscriptionInfo) {
	m.Subscriptions = append(m.Subscriptions, sub)
}

// CheckProviderRegistration checks if a resource provider is registered (mock implementation)
func (m *MockAzureCLI) CheckProviderRegistration(provider string) (bool, error) {
	m.CheckProviderRegistrationCalled = true
	m.CheckProviderRegistrationCalledWith = provider

	if m.CheckProviderRegistrationError != nil {
		return false, m.CheckProviderRegistrationError
	}

	if provider == "" {
		return false, fmt.Errorf("provider cannot be empty")
	}

	// Return mock registration status
	if m.RegisteredProviders == nil {
		return false, nil
	}

	isRegistered, exists := m.RegisteredProviders[provider]
	if !exists {
		return false, nil // Provider not found, treat as not registered
	}

	return isRegistered, nil
}

// RegisterProvider registers a resource provider (mock implementation)
func (m *MockAzureCLI) RegisterProvider(provider string) error {
	m.RegisterProviderCalled = true
	m.RegisterProviderCalledWith = provider

	if m.RegisterProviderError != nil {
		return m.RegisterProviderError
	}

	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}

	// Mark provider as registered in mock data
	if m.RegisteredProviders == nil {
		m.RegisteredProviders = make(map[string]bool)
	}
	m.RegisteredProviders[provider] = true

	return nil
}

// SetProviderRegistrationStatus sets the registration status for a provider (test helper)
func (m *MockAzureCLI) SetProviderRegistrationStatus(provider string, isRegistered bool) {
	if m.RegisteredProviders == nil {
		m.RegisteredProviders = make(map[string]bool)
	}
	m.RegisteredProviders[provider] = isRegistered
}

// SetErrorForCheckProviderRegistration sets an error for CheckProviderRegistration calls (test helper)
func (m *MockAzureCLI) SetErrorForCheckProviderRegistration(err error) {
	m.CheckProviderRegistrationError = err
}

// SetErrorForRegisterProvider sets an error for RegisterProvider calls (test helper)
func (m *MockAzureCLI) SetErrorForRegisterProvider(err error) {
	m.RegisterProviderError = err
}

// Resource group operations mock implementations

// CheckResourceGroupExists checks if a resource group exists (mock implementation)
func (m *MockAzureCLI) CheckResourceGroupExists(name string) (bool, error) {
	m.CheckResourceGroupExistsCalled = true
	m.CheckResourceGroupExistsCalledWith = name

	if m.CheckResourceGroupExistsError != nil {
		return false, m.CheckResourceGroupExistsError
	}

	if name == "" {
		return false, fmt.Errorf("resource group name cannot be empty")
	}

	// Check mock data
	if exists, found := m.ResourceGroupExists[name]; found {
		return exists, nil
	}

	// Check if it's in the default resource groups
	for _, rg := range m.ResourceGroups {
		if rg.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// ListResourceGroups lists all resource groups (mock implementation)
func (m *MockAzureCLI) ListResourceGroups() ([]ResourceGroupInfo, error) {
	m.ListResourceGroupsCalled = true

	if m.ListResourceGroupsError != nil {
		return nil, m.ListResourceGroupsError
	}

	return m.ResourceGroups, nil
}

// CreateResourceGroup creates a resource group (mock implementation)
func (m *MockAzureCLI) CreateResourceGroup(name, location string) error {
	m.CreateResourceGroupCalled = true
	if m.CreateResourceGroupCalledWith == nil {
		m.CreateResourceGroupCalledWith = make(map[string]string)
	}
	m.CreateResourceGroupCalledWith[name] = location

	if m.CreateResourceGroupError != nil {
		return m.CreateResourceGroupError
	}

	// Add the resource group to the mock data if it doesn't exist
	for _, rg := range m.ResourceGroups {
		if rg.Name == name {
			return nil // Already exists
		}
	}

	m.ResourceGroups = append(m.ResourceGroups, ResourceGroupInfo{
		Name:     name,
		Location: location,
	})

	return nil
}

// DeleteResourceGroup deletes a resource group (mock implementation)
func (m *MockAzureCLI) DeleteResourceGroup(name string, noWait bool) error {
	m.DeleteResourceGroupCalled = true
	m.DeleteResourceGroupCalledWith = name

	if m.DeleteResourceGroupError != nil {
		return m.DeleteResourceGroupError
	}

	if name == "" {
		return fmt.Errorf("resource group name cannot be empty")
	}

	// Remove from mock data
	delete(m.ResourceGroupExists, name)
	delete(m.Resources, name)
	delete(m.Deployments, name)
	delete(m.VMs, name)

	// Remove from resource groups list
	for i, rg := range m.ResourceGroups {
		if rg.Name == name {
			m.ResourceGroups = append(m.ResourceGroups[:i], m.ResourceGroups[i+1:]...)
			break
		}
	}

	return nil
}

// ListResources lists all resources in a resource group (mock implementation)
func (m *MockAzureCLI) ListResources(resourceGroup string) ([]ResourceInfo, error) {
	m.ListResourcesCalled = true
	m.ListResourcesCalledWith = resourceGroup

	if m.ListResourcesError != nil {
		return nil, m.ListResourcesError
	}

	if resourceGroup == "" {
		return nil, fmt.Errorf("resource group cannot be empty")
	}

	if resources, exists := m.Resources[resourceGroup]; exists {
		return resources, nil
	}

	return []ResourceInfo{}, nil
}

// GetResource retrieves a specific resource by its ID (mock implementation)
func (m *MockAzureCLI) GetResource(resourceID string) (*ResourceInfo, error) {
	m.GetResourceCalled = true
	m.GetResourceCalledWith = resourceID

	if m.GetResourceError != nil {
		return nil, m.GetResourceError
	}

	if resourceID == "" {
		return nil, fmt.Errorf("resource ID cannot be empty")
	}

	if resource, exists := m.SpecificResources[resourceID]; exists {
		return resource, nil
	}

	// Search in all resource groups
	for _, resources := range m.Resources {
		for _, resource := range resources {
			if resource.ID == resourceID {
				return &resource, nil
			}
		}
	}

	return nil, fmt.Errorf("resource not found: %s", resourceID)
}

// ListDeployments lists all deployments in a resource group (mock implementation)
func (m *MockAzureCLI) ListDeployments(resourceGroup string) ([]DeploymentInfo, error) {
	m.ListDeploymentsCalled = true
	m.ListDeploymentsCalledWith = resourceGroup

	if m.ListDeploymentsError != nil {
		return nil, m.ListDeploymentsError
	}

	if resourceGroup == "" {
		return nil, fmt.Errorf("resource group cannot be empty")
	}

	if deployments, exists := m.Deployments[resourceGroup]; exists {
		return deployments, nil
	}

	return []DeploymentInfo{}, nil
}

// GetDeployment retrieves a specific deployment by its name (mock implementation)
func (m *MockAzureCLI) GetDeployment(resourceGroup, deploymentName string) (*DeploymentInfo, error) {
	m.GetDeploymentCalled = true
	m.GetDeploymentCalledWith = fmt.Sprintf("%s/%s", resourceGroup, deploymentName)

	if m.GetDeploymentError != nil {
		return nil, m.GetDeploymentError
	}

	if resourceGroup == "" || deploymentName == "" {
		return nil, fmt.Errorf("resource group and deployment name cannot be empty")
	}

	// Check specific deployments first
	key := fmt.Sprintf("%s/%s", resourceGroup, deploymentName)
	if deployment, exists := m.SpecificDeployments[key]; exists {
		return deployment, nil
	}

	// Search in resource group deployments
	if deployments, exists := m.Deployments[resourceGroup]; exists {
		for _, deployment := range deployments {
			if deployment.Name == deploymentName {
				return &deployment, nil
			}
		}
	}

	return nil, fmt.Errorf("deployment not found: %s in resource group %s", deploymentName, resourceGroup)
}

// ListVMs lists all VMs in a resource group (mock implementation)
func (m *MockAzureCLI) ListVMs(resourceGroup string) ([]VMInfo, error) {
	m.ListVMsCalled = true
	m.ListVMsCalledWith = resourceGroup

	if m.ListVMsError != nil {
		return nil, m.ListVMsError
	}

	if resourceGroup == "" {
		return nil, fmt.Errorf("resource group cannot be empty")
	}

	if vms, exists := m.VMs[resourceGroup]; exists {
		return vms, nil
	}

	return []VMInfo{}, nil
}

// CreateDeployment creates a new deployment in a resource group (mock implementation)
func (m *MockAzureCLI) CreateDeployment(resourceGroup, deploymentName, templateURI string, parameters []string, noWait bool) error {
	m.CreateDeploymentCalled = true
	m.CreateDeploymentCalledWith = fmt.Sprintf("%s/%s/%s", resourceGroup, deploymentName, templateURI)

	if m.CreateDeploymentError != nil {
		return m.CreateDeploymentError
	}

	if resourceGroup == "" || deploymentName == "" || templateURI == "" {
		return fmt.Errorf("resource group, deployment name, and template URI cannot be empty")
	}

	// Mock successful deployment creation
	return nil
}

// Test helper methods for resource group operations

// SetResourceGroupExists sets the existence status for a resource group (test helper)
func (m *MockAzureCLI) SetResourceGroupExists(name string, exists bool) {
	if m.ResourceGroupExists == nil {
		m.ResourceGroupExists = make(map[string]bool)
	}
	m.ResourceGroupExists[name] = exists
}

// SetResourcesForGroup sets the resources for a specific resource group (test helper)
func (m *MockAzureCLI) SetResourcesForGroup(resourceGroup string, resources []ResourceInfo) {
	if m.Resources == nil {
		m.Resources = make(map[string][]ResourceInfo)
	}
	m.Resources[resourceGroup] = resources
}

// SetDeploymentsForGroup sets the deployments for a specific resource group (test helper)
func (m *MockAzureCLI) SetDeploymentsForGroup(resourceGroup string, deployments []DeploymentInfo) {
	if m.Deployments == nil {
		m.Deployments = make(map[string][]DeploymentInfo)
	}
	m.Deployments[resourceGroup] = deployments
}

// SetVMsForGroup sets the VMs for a specific resource group (test helper)
func (m *MockAzureCLI) SetVMsForGroup(resourceGroup string, vms []VMInfo) {
	if m.VMs == nil {
		m.VMs = make(map[string][]VMInfo)
	}
	m.VMs[resourceGroup] = vms
}

// SetSpecificResource sets a specific resource by ID (test helper)
func (m *MockAzureCLI) SetSpecificResource(resourceID string, resource *ResourceInfo) {
	if m.SpecificResources == nil {
		m.SpecificResources = make(map[string]*ResourceInfo)
	}
	m.SpecificResources[resourceID] = resource
}

// SetSpecificDeployment sets a specific deployment (test helper)
func (m *MockAzureCLI) SetSpecificDeployment(resourceGroup, deploymentName string, deployment *DeploymentInfo) {
	if m.SpecificDeployments == nil {
		m.SpecificDeployments = make(map[string]*DeploymentInfo)
	}
	key := fmt.Sprintf("%s/%s", resourceGroup, deploymentName)
	m.SpecificDeployments[key] = deployment
}

// Error injection helper methods for resource group operations

// SetErrorForCheckResourceGroupExists sets an error for CheckResourceGroupExists calls (test helper)
func (m *MockAzureCLI) SetErrorForCheckResourceGroupExists(err error) {
	m.CheckResourceGroupExistsError = err
}

// SetErrorForListResourceGroups sets an error for ListResourceGroups calls (test helper)
func (m *MockAzureCLI) SetErrorForListResourceGroups(err error) {
	m.ListResourceGroupsError = err
}

// SetErrorForCreateResourceGroup sets an error for CreateResourceGroup calls (test helper)
func (m *MockAzureCLI) SetErrorForCreateResourceGroup(err error) {
	m.CreateResourceGroupError = err
}

// SetErrorForDeleteResourceGroup sets an error for DeleteResourceGroup calls (test helper)
func (m *MockAzureCLI) SetErrorForDeleteResourceGroup(err error) {
	m.DeleteResourceGroupError = err
}

// SetErrorForListResources sets an error for ListResources calls (test helper)
func (m *MockAzureCLI) SetErrorForListResources(err error) {
	m.ListResourcesError = err
}

// SetErrorForGetResource sets an error for GetResource calls (test helper)
func (m *MockAzureCLI) SetErrorForGetResource(err error) {
	m.GetResourceError = err
}

// SetErrorForListDeployments sets an error for ListDeployments calls (test helper)
func (m *MockAzureCLI) SetErrorForListDeployments(err error) {
	m.ListDeploymentsError = err
}

// SetErrorForGetDeployment sets an error for GetDeployment calls (test helper)
func (m *MockAzureCLI) SetErrorForGetDeployment(err error) {
	m.GetDeploymentError = err
}

// SetErrorForCreateDeployment sets an error for CreateDeployment calls (test helper)
func (m *MockAzureCLI) SetErrorForCreateDeployment(err error) {
	m.CreateDeploymentError = err
}
