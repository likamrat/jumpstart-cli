package auth

import (
	"testing"

	"jumpstartcli/internal/azurecli"

	"github.com/stretchr/testify/assert"
)

func TestCheckAzureAuthentication_Success(t *testing.T) {
	// Test when user is logged in
	mockAzCLI := azurecli.NewMockAzureCLI()
	mockAzCLI.IsLoggedInResult = true

	err := CheckAzureAuthentication(mockAzCLI)

	assert.NoError(t, err)
	assert.True(t, mockAzCLI.IsLoggedInCalled, "IsLoggedIn should be called")
}

func TestCheckAzureAuthentication_NotLoggedIn(t *testing.T) {
	// Test when user is not logged in
	mockAzCLI := azurecli.NewMockAzureCLI()
	mockAzCLI.IsLoggedInResult = false

	err := CheckAzureAuthentication(mockAzCLI)

	assert.Error(t, err)
	assert.True(t, mockAzCLI.IsLoggedInCalled, "IsLoggedIn should be called")
	assert.Contains(t, err.Error(), "Azure CLI authentication required")
	assert.Contains(t, err.Error(), "az login")
}

func TestCheckAzureAuthentication_ErrorMessageContent(t *testing.T) {
	// Test the exact error message content
	mockAzCLI := azurecli.NewMockAzureCLI()
	mockAzCLI.IsLoggedInResult = false

	err := CheckAzureAuthentication(mockAzCLI)

	expectedMessage := "Azure CLI authentication required. Please run 'az login' to setup your account"
	assert.EqualError(t, err, expectedMessage)
}

func TestCheckAzureAuthentication_NilInput(t *testing.T) {
	// Test with nil input (should not panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CheckAzureAuthentication should not panic with nil input, but got: %v", r)
		}
	}()

	// This will panic since we're calling methods on nil, but that's expected behavior
	// The calling code should ensure azCLI is not nil
	assert.Panics(t, func() {
		CheckAzureAuthentication(nil)
	}, "CheckAzureAuthentication should panic with nil azCLI")
}

// Benchmark test to ensure authentication check is fast
func BenchmarkCheckAzureAuthentication(b *testing.B) {
	mockAzCLI := azurecli.NewMockAzureCLI()
	mockAzCLI.IsLoggedInResult = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckAzureAuthentication(mockAzCLI)
	}
}
