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

	// Error injection
	GetCurrentSubscriptionError error
	GetSubscriptionError        error
	ListSubscriptionsError      error
	SetSubscriptionError        error

	// Call tracking
	GetCurrentSubscriptionCalled bool
	GetSubscriptionCalled        bool
	GetSubscriptionCalledWith    string
	ListSubscriptionsCalled      bool
	SetSubscriptionCalled        bool
	SetSubscriptionCalledWith    string
	IsLoggedInCalled             bool
}

// NewMockAzureCLI creates a new mock Azure CLI implementation with default test data
func NewMockAzureCLI() *MockAzureCLI {
	return &MockAzureCLI{
		IsLoggedInResult: true,
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

// Reset clears all call tracking flags
func (m *MockAzureCLI) Reset() {
	m.GetCurrentSubscriptionCalled = false
	m.GetSubscriptionCalled = false
	m.GetSubscriptionCalledWith = ""
	m.ListSubscriptionsCalled = false
	m.SetSubscriptionCalled = false
	m.SetSubscriptionCalledWith = ""
	m.IsLoggedInCalled = false
}

// SetErrorForGetCurrentSubscription sets up the mock to return an error for GetCurrentSubscription
func (m *MockAzureCLI) SetErrorForGetCurrentSubscription(err error) {
	m.GetCurrentSubscriptionError = err
}

// SetErrorForGetSubscription sets up the mock to return an error for GetSubscription
func (m *MockAzureCLI) SetErrorForGetSubscription(err error) {
	m.GetSubscriptionError = err
}

// SetErrorForListSubscriptions sets up the mock to return an error for ListSubscriptions
func (m *MockAzureCLI) SetErrorForListSubscriptions(err error) {
	m.ListSubscriptionsError = err
}

// SetErrorForSetSubscription sets up the mock to return an error for SetSubscription
func (m *MockAzureCLI) SetErrorForSetSubscription(err error) {
	m.SetSubscriptionError = err
}

// AddSubscription adds a subscription to the mock data
func (m *MockAzureCLI) AddSubscription(sub SubscriptionInfo) {
	m.Subscriptions = append(m.Subscriptions, sub)
}

// ClearSubscriptions removes all subscriptions from the mock data
func (m *MockAzureCLI) ClearSubscriptions() {
	m.Subscriptions = []SubscriptionInfo{}
	m.CurrentSubscription = nil
}
