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
	VMUsages            map[string][]VMUsageInfo // region -> usage data
	VMSKUs              map[string][]SKUInfo     // region -> SKU data
	AvailableSKUs       map[string][]string      // region -> available SKU names

	// Resource provider mock data
	RegisteredProviders map[string]bool // provider -> registration status

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

	// Call tracking
	GetCurrentSubscriptionCalled        bool
	GetSubscriptionCalled               bool
	GetSubscriptionCalledWith           string
	ListSubscriptionsCalled             bool
	SetSubscriptionCalled               bool
	SetSubscriptionCalledWith           string
	IsLoggedInCalled                    bool
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
