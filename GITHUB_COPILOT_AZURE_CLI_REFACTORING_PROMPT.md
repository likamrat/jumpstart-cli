# GitHub Copilot Prompt: Azure CLI Wrapper Refactoring for Go CLI Commands

## Quick Copy Prompt

**Copy this entire section and paste into GitHub Copilot:**

---

**AZURE CLI WRAPPER REFACTORING PROMPT**

**TARGET PACKAGE OR COMMAND**: js arcbox preflight rp

**STEP 1 - ASSESSMENT**: First, please analyze the current codebase:
1. Search for all `exec.Command("az"` calls in this package
2. Identify functions that need Azure CLI access
3. Check current test structure and mocking patterns
4. Review existing imports and dependencies

**STEP 2 - REFACTORING**: Then refactor this Go CLI command package to use the standardized Azure CLI wrapper interface (`internal/azurecli`) following these proven patterns:

**Current Status**: The `cmd/subscription` package (95.9% coverage) and `internal/preflight/arcbox/quota` (94.3% coverage) are fully refactored. The `cmd/arcbox` main functionality and `internal/resourceproviders` packages still have 25+ and 2+ direct Azure CLI calls respectively that need conversion.

**Refactoring Pattern**:
1. **Add dependency injection structure**: Create `defaultAzureCLI` variable and `SetAzureCLI()` function for testing
2. **Split constructors**: Create both `NewCommandCmd()` and `NewCommandCmdWithCLI(azCLI azurecli.AzureCLI)` versions
3. **Replace direct calls**: Convert all `exec.Command("az", ...)` calls to use `azCLI.GetCurrentSubscription()`, `azCLI.IsLoggedIn()`, `azCLI.ListVMUsage()`, etc.
4. **Add comprehensive tests**: Use `azurecli.NewMockAzureCLI()` for systematic testing of success/failure scenarios

**Available Interface Methods**:
- `GetCurrentSubscription()`, `GetSubscription(id)`, `ListSubscriptions()`, `SetSubscription(id)`
- `IsLoggedIn()`
- `ListVMUsage(region)`, `ListVMSKUs(region)`, `CheckSKUAvailability(sku, region)`

**Goal**: Achieve 90%+ test coverage with robust error handling, dependency injection, and comprehensive mocking while maintaining backward compatibility.

Apply this refactoring systematically to eliminate all direct Azure CLI calls and enable comprehensive testing.

**USAGE EXAMPLES:**
- For arcbox command: Replace `[REPLACE WITH PACKAGE PATH]` with `cmd/arcbox`
- For resource providers: Replace `[REPLACE WITH PACKAGE PATH]` with `internal/resourceproviders`
- For agora command: Replace `[REPLACE WITH PACKAGE PATH]` with `cmd/agora`

**REFERENCE**: See the detailed documentation, templates, and examples in `GITHUB_COPILOT_AZURE_CLI_REFACTORING_PROMPT.md` below this quick copy section for comprehensive guidance on patterns, common refactoring scenarios, and quality checklists.

---

## Context
You are refactoring existing Go CLI commands to use the new standardized Azure CLI wrapper interface (`internal/azurecli`) that provides dependency injection, comprehensive mocking, and improved testability. This refactoring follows the successful patterns established in the `cmd/subscription` package and the quota functionality within `internal/preflight/arcbox`.

## Current Refactoring Status

### Fully Refactored Packages
- **`cmd/subscription`**: Complete integration with Azure CLI wrapper (95.9% test coverage)
- **`internal/preflight/arcbox/quota`**: Complete integration for quota functionality (94.3% test coverage)

### Partially Refactored Packages
- **`cmd/arcbox`**: Only quota functionality refactored, 25+ direct Azure CLI calls remain in main command
- **`internal/resourceproviders`**: 2+ direct Azure CLI calls still need refactoring

### Packages Needing Refactoring
- **`cmd/agora`**: In development, no Azure CLI calls yet
- **`cmd/localbox`**: In development, no Azure CLI calls yet

## Refactoring Goals
1. **Standardize Azure CLI interactions** using the `azurecli.AzureCLI` interface
2. **Enable dependency injection** for better testing and modularity
3. **Improve error handling** with consistent Azure CLI error patterns
4. **Enhance testability** through mockable Azure CLI operations
5. **Maintain backward compatibility** while improving code quality

## Azure CLI Wrapper Interface Overview

The `azurecli.AzureCLI` interface provides these key operations:

```go
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
```

## Refactoring Pattern Templates

### Template 1: Command Structure Refactoring

**Before (Raw Azure CLI usage):**
```go
package command

import (
    "os/exec"
    "encoding/json"
    "github.com/spf13/cobra"
)

func NewCommandCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "command",
        Short: "Command description",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Direct Azure CLI calls
            azCmd := exec.Command("az", "account", "show")
            output, err := azCmd.Output()
            if err != nil {
                return err
            }
            // Process output...
            return nil
        },
    }
}
```

**After (Azure CLI wrapper usage):**
```go
package command

import (
    "jumpstartcli/internal/azurecli"
    "github.com/spf13/cobra"
)

// Default Azure CLI instance - can be overridden for testing
var defaultAzureCLI azurecli.AzureCLI = azurecli.NewAzureCLI()

// SetAzureCLI allows overriding the Azure CLI implementation for testing
func SetAzureCLI(cli azurecli.AzureCLI) {
    defaultAzureCLI = cli
}

// NewCommandCmd creates the command with default Azure CLI
func NewCommandCmd() *cobra.Command {
    return NewCommandCmdWithCLI(defaultAzureCLI)
}

// NewCommandCmdWithCLI creates the command with injectable Azure CLI for testing
func NewCommandCmdWithCLI(azCLI azurecli.AzureCLI) *cobra.Command {
    return &cobra.Command{
        Use:   "command",
        Short: "Command description", 
        RunE: func(cmd *cobra.Command, args []string) error {
            // Use Azure CLI wrapper
            sub, err := azCLI.GetCurrentSubscription()
            if err != nil {
                return fmt.Errorf("failed to get current subscription: %v", err)
            }
            // Process subscription...
            return nil
        },
    }
}
```

### Template 2: Subscription Validation Refactoring

**Before (Manual subscription checking):**
```go
func validateSubscription() error {
    cmd := exec.Command("az", "account", "show")
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("not logged in to Azure")
    }
    
    var sub map[string]interface{}
    if err := json.Unmarshal(output, &sub); err != nil {
        return fmt.Errorf("failed to parse subscription data")
    }
    
    // Manual validation logic...
    return nil
}
```

**After (Azure CLI wrapper):**
```go
func validateSubscription(azCLI azurecli.AzureCLI) error {
    if !azCLI.IsLoggedIn() {
        return fmt.Errorf("not logged in to Azure. Run 'az login' to authenticate")
    }
    
    sub, err := azCLI.GetCurrentSubscription()
    if err != nil {
        return fmt.Errorf("failed to get current subscription: %v", err)
    }
    
    if sub.ID == "" {
        return fmt.Errorf("no active subscription found")
    }
    
    return nil
}
```

### Template 3: Quota Checking Refactoring

**Before (Direct quota commands):**
```go
func checkQuota(region string) error {
    cmd := exec.Command("az", "vm", "list-usage", "--location", region, "--output", "json")
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("failed to get quota data: %v", err)
    }
    
    var usage []map[string]interface{}
    if err := json.Unmarshal(output, &usage); err != nil {
        return fmt.Errorf("failed to parse quota data: %v", err)
    }
    
    // Manual quota parsing...
    return nil
}
```

**After (Azure CLI wrapper):**
```go
func checkQuota(azCLI azurecli.AzureCLI, region string) error {
    usage, err := azCLI.ListVMUsage(region)
    if err != nil {
        return fmt.Errorf("failed to get quota data for region %s: %v", region, err)
    }
    
    // Use structured VMUsageInfo data
    for _, u := range usage {
        if u.CurrentValue >= u.Limit {
            return fmt.Errorf("quota exceeded for %s", u.Name["value"])
        }
    }
    
    return nil
}
```

### Template 4: Test Enhancement with Mocks

**Before (Hard to test with real Azure CLI):**
```go
func TestCommand(t *testing.T) {
    // Tests were either integration tests or skipped
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Test with real Azure CLI (fragile and slow)
    cmd := NewCommandCmd()
    err := cmd.Execute()
    if err != nil {
        t.Errorf("Command failed: %v", err)
    }
}
```

**After (Comprehensive mocking):**
```go
func TestCommand(t *testing.T) {
    tests := []struct {
        name        string
        mockSetup   func(*azurecli.MockAzureCLI)
        expectError bool
        errorMsg    string
    }{
        {
            name: "successful execution",
            mockSetup: func(mock *azurecli.MockAzureCLI) {
                mock.SetCurrentSubscription(&azurecli.SubscriptionInfo{
                    ID:   "test-subscription-id",
                    Name: "Test Subscription",
                })
            },
            expectError: false,
        },
        {
            name: "not logged in",
            mockSetup: func(mock *azurecli.MockAzureCLI) {
                mock.SetLoggedIn(false)
            },
            expectError: true,
            errorMsg:    "not logged in",
        },
        {
            name: "subscription error",
            mockSetup: func(mock *azurecli.MockAzureCLI) {
                mock.SetError(fmt.Errorf("subscription not found"))
            },
            expectError: true,
            errorMsg:    "subscription not found",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := azurecli.NewMockAzureCLI()
            tt.mockSetup(mock)
            
            cmd := NewCommandCmdWithCLI(mock)
            err := cmd.Execute()
            
            if tt.expectError && err == nil {
                t.Errorf("Expected error, got nil")
            }
            if !tt.expectError && err != nil {
                t.Errorf("Expected no error, got %v", err)
            }
            if tt.expectError && !strings.Contains(err.Error(), tt.errorMsg) {
                t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
            }
        })
    }
}
```

## Systematic Refactoring Steps

### Step 1: Analysis and Planning
1. **Identify Azure CLI usage**: Search for `exec.Command("az"` calls
2. **Map to wrapper methods**: Determine which `azurecli.AzureCLI` methods to use
3. **Identify dependencies**: Find functions that need Azure CLI access
4. **Plan injection points**: Determine where to inject the Azure CLI dependency

### Step 2: Structure Refactoring
1. **Add wrapper import**: `import "jumpstartcli/internal/azurecli"`
2. **Create default instance**: `var defaultAzureCLI azurecli.AzureCLI = azurecli.NewAzureCLI()`
3. **Add setter function**: For dependency injection in tests
4. **Split command constructors**: Create testable version with CLI parameter
5. **Update function signatures**: Add Azure CLI parameter to functions that need it

### Step 3: Replace Azure CLI Calls
1. **Replace exec.Command calls**: Use wrapper methods instead
2. **Update error handling**: Use wrapper's error patterns
3. **Simplify JSON parsing**: Use wrapper's structured types
4. **Improve error messages**: Leverage wrapper's context-aware errors

### Step 4: Enhance Testing
1. **Create mock-based tests**: Use `azurecli.NewMockAzureCLI()`
2. **Add error condition tests**: Test all Azure CLI failure scenarios
3. **Improve test coverage**: Target 90%+ coverage with systematic testing
4. **Add integration tests**: Keep some real Azure CLI tests for validation

## Common Refactoring Patterns

### Pattern 1: Subscription Access Validation
```go
// Before: Direct Azure CLI call
func ensureLoggedIn() error {
    cmd := exec.Command("az", "account", "show")
    _, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("not logged in to Azure")
    }
    return nil
}

// After: Wrapper usage
func ensureLoggedIn(azCLI azurecli.AzureCLI) error {
    if !azCLI.IsLoggedIn() {
        return fmt.Errorf("not logged in to Azure. Run 'az login' to authenticate")
    }
    
    sub, err := azCLI.GetCurrentSubscription()
    if err != nil {
        return fmt.Errorf("failed to access current subscription: %v", err)
    }
    
    if sub.ID == "" {
        return fmt.Errorf("no active subscription found. Run 'az account set --subscription <subscription-id>'")
    }
    
    return nil
}
```

### Pattern 2: Resource Validation
```go
// Before: Manual command construction
func checkResourceAvailability(region, sku string) (bool, error) {
    cmd := exec.Command("az", "vm", "list-skus", "--location", region, "--query", fmt.Sprintf("[?name=='%s']", sku))
    output, err := cmd.Output()
    if err != nil {
        return false, err
    }
    
    var skus []map[string]interface{}
    if err := json.Unmarshal(output, &skus); err != nil {
        return false, err
    }
    
    return len(skus) > 0, nil
}

// After: Wrapper method
func checkResourceAvailability(azCLI azurecli.AzureCLI, region, sku string) (bool, error) {
    available, err := azCLI.CheckSKUAvailability(sku, region)
    if err != nil {
        return false, fmt.Errorf("failed to check SKU availability for %s in %s: %v", sku, region, err)
    }
    
    return available, nil
}
```

### Pattern 3: Error Handling Standardization
```go
// Before: Inconsistent error handling
func processAzureOperation() error {
    cmd := exec.Command("az", "...")
    output, err := cmd.Output()
    if err != nil {
        // Various error handling approaches across commands
        return err
    }
    return nil
}

// After: Standardized error handling
func processAzureOperation(azCLI azurecli.AzureCLI) error {
    result, err := azCLI.SomeOperation()
    if err != nil {
        // Consistent error context and formatting
        return fmt.Errorf("failed to perform Azure operation: %v", err)
    }
    
    // Use structured data instead of parsing JSON
    if result.ID == "" {
        return fmt.Errorf("operation completed but returned invalid data")
    }
    
    return nil
}
```

## Quality Checklist

### After Refactoring, Ensure:
1. **No direct Azure CLI calls**: All `exec.Command("az"` calls replaced
2. **Dependency injection**: Commands accept Azure CLI interface parameter
3. **Backward compatibility**: Original command constructors still work
4. **Enhanced testing**: Mock-based tests with comprehensive scenarios
5. **Improved error handling**: Consistent error messages and context
6. **Documentation updates**: Update function comments and usage examples

### Validation Steps:
1. **Run existing tests**: Ensure no regressions
2. **Add new tests**: Achieve target coverage with mocks
3. **Integration testing**: Verify real Azure CLI still works
4. **Error scenario testing**: Test all failure modes
5. **Performance check**: Ensure no performance degradation

## Example Application

When I say "Refactor the `cmd/agora` package to use the Azure CLI wrapper", you should:

1. **Analyze current Azure CLI usage** in agora.go
2. **Add Azure CLI wrapper import** and dependency injection structure
3. **Replace direct Azure CLI calls** with wrapper methods
4. **Update function signatures** to accept Azure CLI parameter
5. **Create comprehensive tests** using mocks
6. **Validate functionality** with both mocks and real Azure CLI

This systematic approach ensures consistent Azure CLI integration across all command packages while dramatically improving testability and maintainability.

---

**Proven Benefits**: The subscription package using this pattern achieved 95.9% test coverage, and the quota functionality within arcbox achieved 94.3% test coverage, both with robust error handling and comprehensive test suites. This refactoring approach provides a systematic way to modernize Azure CLI integration across the remaining packages that still have direct Azure CLI calls (primarily `cmd/arcbox` main functionality and `internal/resourceproviders`).
