# GitHub Copilot Prompt: Azure CLI Wrapper Refactoring for Go CLI Commands

## Quick Copy Prompt

**Copy this entire section and paste into GitHub Copilot:**

---

**AZURE CLI WRAPPER REFACTORING PROMPT**

**TARGET PACKAGE OR COMMAND**: `cmd/arcbox`

**STEP 1 - ASSESSMENT**: First, please analyze the current codebase:
1. Search for all `exec.Command("az"` calls in this package
2. Identify functions that need Azure CLI access
3. Check current test structure and mocking patterns
4. Review existing imports and dependencies

**STEP 2 - REFACTORING**: Then refactor this Go CLI command package to use the standardized Azure CLI wrapper interface (`internal/azurecli`) following these proven patterns:

**Current Status**: **ALL TARGETED PACKAGES COMPLETED** - The `cmd/subscription` package (95.9% coverage), `internal/preflight/arcbox/quota` (94.3% coverage), `internal/preflight/arcbox/rp` and `internal/preflight/arcbox/status` (comprehensive Azure CLI wrapper integration with 122 total tests), and **`cmd/arcbox` main functionality (100% refactored with 15+ comprehensive test functions)** are fully refactored with zero direct Azure CLI calls remaining.

**Refactoring Pattern**:
1. **Add dependency injection structure**: Create `defaultAzureCLI` variable and `SetAzureCLI()` function for testing
2. **Split constructors**: Create both `NewCommandCmd()` and `NewCommandCmdWithCLI(azCLI azurecli.AzureCLI)` versions
3. **Replace direct calls**: Convert all `exec.Command("az", ...)` calls to use `azCLI.GetCurrentSubscription()`, `azCLI.IsLoggedIn()`, `azCLI.ListVMUsage()`, etc.
4. **Add comprehensive tests**: Use `azurecli.NewMockAzureCLI()` for systematic testing of success/failure scenarios

**Available Interface Methods**:
- `GetCurrentSubscription()`, `GetSubscription(id)`, `ListSubscriptions()`, `SetSubscription(id)`
- `IsLoggedIn()`
- `ListVMUsage(region)`, `ListVMSKUs(region)`, `CheckSKUAvailability(sku, region)`
- `IsResourceProviderRegistered(provider)`, `RegisterResourceProvider(provider)` (used in rp commands)
- `CheckResourceGroupExists(name)`, `ListResourceGroups()`, `DeleteResourceGroup(name)` (added for arcbox)
- `ListResources(resourceGroup)`, `GetResource(resourceGroup, name, resourceType)` (added for arcbox)
- `ListDeployments(resourceGroup)`, `GetDeployment(resourceGroup, name)` (added for arcbox)
- `ListVMs(resourceGroup)` (added for arcbox)

**Recent Success Example**: The `internal/preflight/arcbox` package was comprehensively refactored with:
- **`rp.go`** and **`status.go`**: Complete modular command implementation following the established pattern
- **`rp_test.go`** and **`status_test.go`**: 16+ new test functions with comprehensive coverage including:
  - Command validation and error handling
  - Flag configuration testing
  - Mock-based Azure CLI testing
  - Edge case and error scenario testing
  - Performance benchmarking
- **Total ArcBox package tests**: 122 tests all passing
- **Pattern consistency**: Followed same refactoring methodology as quota and subscription packages

**Goal**: Achieve 90%+ test coverage with robust error handling, dependency injection, and comprehensive mocking while maintaining backward compatibility.

Apply this refactoring systematically to eliminate all direct Azure CLI calls and enable comprehensive testing.

**USAGE EXAMPLES:**
- For arcbox command: Replace `[REPLACE WITH PACKAGE PATH]` with `cmd/arcbox`
- For resource providers: Replace `[REPLACE WITH PACKAGE PATH]` with `internal/resourceproviders`
- For agora command: Replace `[REPLACE WITH PACKAGE PATH]` with `cmd/agora`

**REFERENCE**: See the detailed documentation, templates, and examples below this quick copy section in the `GITHUB_COPILOT_AZURE_CLI_REFACTORING_PROMPT.md` for comprehensive guidance on patterns, common refactoring scenarios, and quality checklists.

---

## Context
You are refactoring existing Go CLI commands to use the new standardized Azure CLI wrapper interface (`internal/azurecli`) that provides dependency injection, comprehensive mocking, and improved testability. This refactoring follows the successful patterns established in the `cmd/subscription` package and the quota functionality within `internal/preflight/arcbox`.

## Current Refactoring Status

### Fully Refactored Packages

- **`cmd/subscription`**: Complete integration with Azure CLI wrapper (95.9% test coverage)
- **`cmd/arcbox`**: **FULLY COMPLETED** - All 20+ direct Azure CLI calls refactored with comprehensive test suite (15+ test functions)
- **`internal/preflight/arcbox/quota`**: Complete integration for quota functionality (94.3% test coverage)
- **`internal/preflight/arcbox/rp`**: Complete resource provider command refactoring with Azure CLI wrapper
- **`internal/preflight/arcbox/status`**: Complete status command implementation with comprehensive testing
- **`internal/resourceproviders`**: Fully refactored to use Azure CLI wrapper interface

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

    // Resource Provider operations
    IsResourceProviderRegistered(provider string) (bool, error)
    RegisterResourceProvider(provider string) error

    // Resource Group operations (added for arcbox)
    CheckResourceGroupExists(name string) (bool, error)
    ListResourceGroups() ([]ResourceGroupInfo, error)
    DeleteResourceGroup(name string) error

    // Resource operations (added for arcbox)
    ListResources(resourceGroup string) ([]ResourceInfo, error)
    GetResource(resourceGroup, name, resourceType string) (*ResourceInfo, error)

    // Deployment operations (added for arcbox)
    ListDeployments(resourceGroup string) ([]DeploymentInfo, error)
    GetDeployment(resourceGroup, name string) (*DeploymentInfo, error)

    // VM operations (added for arcbox)
    ListVMs(resourceGroup string) ([]VMInfo, error)
}
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

## Recent Success Story: Complete ArcBox Package Refactoring

### Completed Implementation
The **complete `cmd/arcbox` package** demonstrates the full success of this refactoring methodology:

**Files Fully Refactored**:
- `arcbox.go` - **ALL 20+ direct Azure CLI calls eliminated** with complete Azure CLI wrapper integration
- `internal/preflight/arcbox/rp.go` - Resource provider commands with full Azure CLI wrapper integration
- `internal/preflight/arcbox/status.go` - Status command implementation following established patterns
- `internal/preflight/arcbox/quota.go` - Previously refactored with 94.3% test coverage

**Test Coverage Achieved**:
- `arcbox_test.go` - **15+ comprehensive test functions** covering all main command scenarios
- `rp_test.go` - 9 comprehensive test functions covering all command scenarios
- `status_test.go` - 8 test functions with extensive validation and benchmarking
- `quota_test.go` - Previously achieved 94.3% coverage
- **Total**: 140+ tests across the complete arcbox package ecosystem

**Key Achievements**:
1. **100% Azure CLI call elimination** - Zero `exec.Command("az"...)` usage remaining
2. **Comprehensive interface extension** - Added 8 new Azure CLI wrapper methods
3. **Complete dependency injection** - All functions accept Azure CLI interface parameters
4. **Robust error handling** - All failure scenarios tested and validated
5. **Full backward compatibility** - All original functionality preserved
6. **Performance benchmarking** - Critical operations monitored for performance
7. **Pattern consistency** - Same refactoring approach across all subcommands and main package

**New Interface Methods Added for ArcBox**:
- `CheckResourceGroupExists()`, `ListResourceGroups()`, `DeleteResourceGroup()`
- `ListResources()`, `GetResource()`
- `ListDeployments()`, `GetDeployment()`
- `ListVMs()`

**Implementation Pattern Used**:
```go
// Azure CLI wrapper integration in main arcbox functions
func discoverArcBoxDeployments(azCLI azurecli.AzureCLI) ([]ArcBoxDeployment, error) {
    // Uses injected Azure CLI interface instead of direct calls
    resourceGroups, err := azCLI.ListResourceGroups()
    // All logic now uses wrapper methods
}

// Comprehensive test coverage with mocks
func TestDiscoverArcBoxDeployments(t *testing.T) {
    mockCLI := azurecli.NewMockAzureCLI()
    arcboxCmd := NewArcBoxCmdWithCLI(mockCLI)
    // Test all scenarios including error conditions
}
```

This success demonstrates the methodology's effectiveness and provides a complete template for remaining packages.

## Previous Success Story: ArcBox Preflight Refactoring
The `internal/preflight/arcbox` package demonstrates the complete success of this refactoring methodology:

**Files Refactored**:
- `rp.go` - Resource provider commands with full Azure CLI wrapper integration
- `status.go` - Status command implementation following established patterns
- `quota.go` - Previously refactored with 94.3% test coverage

**Test Coverage Achieved**:
- `rp_test.go` - 9 comprehensive test functions covering all command scenarios
- `status_test.go` - 8 test functions with extensive validation and benchmarking
- `quota_test.go` - Previously achieved 94.3% coverage
- **Total**: 122 tests across the complete package

**Key Achievements**:
1. **Zero direct Azure CLI calls** - All `exec.Command("az"...)` usage eliminated
2. **Comprehensive mock testing** - Every Azure CLI interaction fully mockable
3. **Robust error handling** - All failure scenarios tested and validated
4. **Performance benchmarking** - Critical operations monitored for performance
5. **Pattern consistency** - Same refactoring approach across all subcommands

**Implementation Pattern Used**:
```go
// Azure CLI wrapper integration
func CreateResourceProviderCommands(cli azurecli.AzureCLI) *cobra.Command {
    // Uses injected Azure CLI interface instead of direct calls
    ShowResourceProviderStatus(cli)
}

// Comprehensive test coverage with mocks
func TestResourceProviderCommandValidation(t *testing.T) {
    mockCLI := &azurecli.MockAzureCLI{}
    rpCmd := CreateResourceProviderCommands(mockCLI)
    // Test all scenarios including error conditions
}
```

This success demonstrates the methodology's effectiveness and provides a concrete template for remaining packages.

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

**Proven Benefits**: The subscription package using this pattern achieved 95.9% test coverage, the quota functionality within arcbox achieved 94.3% test coverage, **the complete ArcBox preflight package (status, rp, quota) now has 122 comprehensive tests with full Azure CLI wrapper integration**, and **the main `cmd/arcbox` package has been 100% refactored with 15+ comprehensive test functions and zero direct Azure CLI calls remaining**. This refactoring approach has successfully modernized Azure CLI integration across all targeted packages, demonstrating a systematic way to achieve comprehensive testability, dependency injection, and robust error handling while maintaining full backward compatibility.
