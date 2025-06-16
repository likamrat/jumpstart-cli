# Test File Reorganization Blueprint

## Executive Summary

Based on the comprehensive coverage analysis showing 48.8% coverage for the main arcbox package, this blueprint provides a detailed plan for reorganizing the monolithic 1,843-line test file into focused, maintainable components while achieving 95%+ coverage across all modules.

## Current State Analysis

### Monolithic Test File Breakdown (`arcbox_test.go` - 1,843 lines)
- **19 test functions** covering multiple concerns
- **Mixed responsibilities**: Core commands, business logic, utility functions
- **Coverage gaps**: Critical functions like `SetAzureCLI` (0.0% coverage)
- **Maintenance challenges**: Large file, difficult to navigate and modify

### Coverage Baseline from Analysis
| Component | Current Coverage | Target Coverage | Priority |
|-----------|------------------|-----------------|----------|
| **arcbox.go** | 77.5% overall | 95%+ | High |
| **deploy_cmd.go** | 60.4% | 95%+ | High |
| **delete_cmd.go** | 23.8% | 95%+ | Critical |
| **list_cmd.go** | 13.2% | 95%+ | Critical |
| **preflight_cmd.go** | 50.0% | 95%+ | High |

## File-by-File Migration Plan

### 1. arcbox_test.go (Target: 300-400 lines, 95%+ coverage)

#### Keep These Tests (Enhanced)
- [x] **TestNewArcboxCmd** (Lines 34-74)
  - **Current**: Basic command structure validation
  - **Enhancement**: Add edge cases, error scenarios, command tree validation
  - **New coverage**: Command factory error handling, invalid configurations

- [x] **TestNewArcboxCmdWithCLI_Comprehensive** (Lines 699-957)
  - **Current**: CLI integration scenarios
  - **Enhancement**: Add more CLI failure scenarios, authentication edge cases
  - **New coverage**: CLI state management, connection failures

- [x] **TestNewArcboxCmdWithCLI_AllSubcommands** (Lines 957-1008)
  - **Current**: Subcommand existence validation
  - **Enhancement**: Deep subcommand hierarchy validation, command relationships
  - **New coverage**: Command tree integrity, parent-child relationships

- [x] **TestNewArcboxCmdWithCLI_CommandProperties** (Lines 1008-1056)
  - **Current**: Command property validation
  - **Enhancement**: Comprehensive property testing, flag inheritance
  - **New coverage**: Help text validation, usage patterns

- [x] **TestNewArcboxCmdWithCLI_ErrorHandling** (Lines 1056-1109)
  - **Current**: Basic error scenarios
  - **Enhancement**: Comprehensive error propagation, suggestion system
  - **New coverage**: Unknown command handling, suggestion accuracy

- [x] **TestNewArcboxCmdWithCLI_NilInputHandling** (Lines 1109-1131)
  - **Current**: Nil input scenarios
  - **Enhancement**: Comprehensive nil safety, defensive programming
  - **New coverage**: Edge cases with partial nil inputs

- [x] **TestNewArcboxCmdWithCLI_RunEFunction_EdgeCases** (Lines 1505-1646)
  - **Current**: Command execution edge cases
  - **Enhancement**: More execution scenarios, argument validation
  - **New coverage**: Command suggestion system, help display logic

#### New Tests to Add (Critical Coverage Gaps)
- [ ] **TestSetAzureCLI** - **CRITICAL (0.0% coverage)**
  ```go
  func TestSetAzureCLI_BasicFunctionality(t *testing.T)
  func TestSetAzureCLI_NilInput(t *testing.T)
  func TestSetAzureCLI_ReplacementChain(t *testing.T)
  ```

- [ ] **TestCommandHierarchy** - Full command tree validation
  ```go
  func TestCommandHierarchy_CompleteTree(t *testing.T)
  func TestCommandHierarchy_SubcommandRelationships(t *testing.T)
  func TestCommandHierarchy_InvalidPaths(t *testing.T)
  ```

- [ ] **TestCommandIntegration** - Cross-command interactions
  ```go
  func TestCommandIntegration_SharedServices(t *testing.T)
  func TestCommandIntegration_CLIStatePersistence(t *testing.T)
  func TestCommandIntegration_ErrorPropagation(t *testing.T)
  ```

- [ ] **TestFactoryErrorHandling** - Factory function edge cases
  ```go
  func TestFactoryErrorHandling_ServiceCreationFailures(t *testing.T)
  func TestFactoryErrorHandling_CLIInjectionFailures(t *testing.T)
  func TestFactoryErrorHandling_CommandCreationEdgeCases(t *testing.T)
  ```

#### Tests to Remove (Migrate to Other Files)
- [x] **TestArcboxDeployCommand** → `deploy_cmd_test.go`
- [x] **TestArcboxDeleteCommand** → `delete_cmd_test.go`
- [x] **TestArcboxListCommand** → `list_cmd_test.go`
- [x] **TestArcboxPreflightCommand** → `preflight_cmd_test.go`
- [x] **TestArcboxPreflightQuotaCommand** → `preflight_cmd_test.go`
- [x] **TestArcboxPreflightRpCommand** → `preflight_cmd_test.go`
- [x] **TestArcboxPreflightRpRegisterCommand** → `preflight_cmd_test.go`
- [x] **TestDetectArcBoxFlavor** → `services/listing_service_test.go` (if not already there)
- [x] **TestDetectArcBoxFlavorFallback** → `services/listing_service_test.go` (if not already there)

### 2. deploy_cmd_test.go (NEW FILE, Target: 95%+ coverage)

#### Migrate These Tests
- [x] **TestArcboxDeployCommand** (Lines 74-204)
  - **Rename to**: `TestDeployCommand_BasicStructure`
  - **Current coverage**: Command structure, basic flag validation
  - **Enhancement needed**: Comprehensive flag testing, error scenarios

#### New Tests to Add (60.4% → 95%+ coverage)
- [ ] **TestDeployCommand_FlagValidation**
  ```go
  func TestDeployCommand_RequiredFlags(t *testing.T)      // --resource-group, --subscription
  func TestDeployCommand_OptionalFlags(t *testing.T)     // --template-local, --template-params
  func TestDeployCommand_FlagDefaults(t *testing.T)      // Default values and behavior
  func TestDeployCommand_FlagConflicts(t *testing.T)     // Mutually exclusive flags
  func TestDeployCommand_FlagValidation(t *testing.T)    // Invalid flag combinations
  ```

- [ ] **TestDeployCommand_ServiceIntegration**
  ```go
  func TestDeployCommand_DeploymentServiceCalls(t *testing.T)  // Service method verification
  func TestDeployCommand_ParameterPassing(t *testing.T)       // Parameter validation
  func TestDeployCommand_TemplateResolution(t *testing.T)     // Local vs remote templates
  func TestDeployCommand_ValidationIntegration(t *testing.T)  // Preflight check integration
  ```

- [ ] **TestDeployCommand_ErrorHandling**
  ```go
  func TestDeployCommand_CLIErrors(t *testing.T)           // CLI command failures
  func TestDeployCommand_ValidationErrors(t *testing.T)    // Parameter validation failures
  func TestDeployCommand_ServiceErrors(t *testing.T)       // Deployment service errors
  func TestDeployCommand_TemplateErrors(t *testing.T)      // Template resolution failures
  ```

- [ ] **TestDeployCommand_OutputFormats**
  ```go
  func TestDeployCommand_ProgressDisplay(t *testing.T)     // Progress indicator testing
  func TestDeployCommand_SuccessOutput(t *testing.T)       // Success message formatting
  func TestDeployCommand_ErrorOutput(t *testing.T)         // Error message formatting
  ```

### 3. delete_cmd_test.go (NEW FILE, Target: 95%+ coverage) - **CRITICAL PRIORITY**

#### Migrate These Tests
- [x] **TestArcboxDeleteCommand** (Lines 204-277)
  - **Rename to**: `TestDeleteCommand_BasicStructure`
  - **Current coverage**: 23.8% - **Critical coverage gap**
  - **Enhancement needed**: Complete functionality testing

#### New Tests to Add (23.8% → 95%+ coverage) - **HIGH PRIORITY**
- [ ] **TestDeleteCommand_ConfirmationFlow** - **Critical for data safety**
  ```go
  func TestDeleteCommand_UserConfirmation_Yes(t *testing.T)    // User confirms deletion
  func TestDeleteCommand_UserConfirmation_No(t *testing.T)     // User cancels deletion
  func TestDeleteCommand_SkipConfirmation(t *testing.T)        // --yes flag behavior
  func TestDeleteCommand_ConfirmationPrompt(t *testing.T)      // Prompt display testing
  ```

- [ ] **TestDeleteCommand_DeletionFlow**
  ```go
  func TestDeleteCommand_ResourceGroupValidation(t *testing.T) // RG existence checking
  func TestDeleteCommand_DeletionExecution(t *testing.T)       // Actual deletion process
  func TestDeleteCommand_ProgressMonitoring(t *testing.T)      // Background deletion tracking
  func TestDeleteCommand_CleanupVerification(t *testing.T)     // Post-deletion validation
  ```

- [ ] **TestDeleteCommand_ErrorScenarios** - **Critical for safety**
  ```go
  func TestDeleteCommand_ResourceNotFound(t *testing.T)        // Non-existent RG
  func TestDeleteCommand_PermissionErrors(t *testing.T)        // Access denied scenarios
  func TestDeleteCommand_CLIFailures(t *testing.T)             // Azure CLI failures
  func TestDeleteCommand_PartialDeletionFailures(t *testing.T) // Incomplete deletion handling
  ```

### 4. list_cmd_test.go (NEW FILE, Target: 95%+ coverage) - **CRITICAL PRIORITY**

#### Migrate These Tests
- [x] **TestArcboxListCommand** (Lines 277-350)
  - **Rename to**: `TestListCommand_BasicStructure`
  - **Current coverage**: 13.2% - **Critical coverage gap**
  - **Enhancement needed**: Complete functionality testing

#### New Tests to Add (13.2% → 95%+ coverage) - **HIGH PRIORITY**
- [ ] **TestListCommand_OutputFormats**
  ```go
  func TestListCommand_TableOutput(t *testing.T)      // Default table format
  func TestListCommand_JSONOutput(t *testing.T)       // JSON format validation
  func TestListCommand_YAMLOutput(t *testing.T)       // YAML format validation
  func TestListCommand_TSVOutput(t *testing.T)        // TSV format validation
  ```

- [ ] **TestListCommand_SubscriptionHandling**
  ```go
  func TestListCommand_CurrentSubscription(t *testing.T)   // Default subscription
  func TestListCommand_SpecificSubscription(t *testing.T)  // --subscription flag
  func TestListCommand_AllSubscriptions(t *testing.T)      // --all-subscriptions flag
  func TestListCommand_SubscriptionFiltering(t *testing.T) // Invalid subscriptions
  ```

- [ ] **TestListCommand_ServiceIntegration**
  ```go
  func TestListCommand_ListingServiceCalls(t *testing.T)   // Service method verification
  func TestListCommand_FlavorDetection(t *testing.T)       // Flavor detection testing
  func TestListCommand_ResourceGroupScanning(t *testing.T) // RG discovery testing
  func TestListCommand_DeploymentEnrichment(t *testing.T)  // Deployment data enrichment
  ```

- [ ] **TestListCommand_EdgeCases**
  ```go
  func TestListCommand_EmptyResults(t *testing.T)          // No deployments found
  func TestListCommand_LargeResultSets(t *testing.T)       // Performance with many results
  func TestListCommand_MalformedData(t *testing.T)         // Invalid deployment data
  func TestListCommand_PartialErrors(t *testing.T)         // Some subscriptions fail
  ```

### 5. preflight_cmd_test.go (NEW FILE, Target: 95%+ coverage)

#### Migrate These Tests
- [x] **TestArcboxPreflightCommand** (Lines 350-405)
- [x] **TestArcboxPreflightQuotaCommand** (Lines 405-497)
- [x] **TestArcboxPreflightRpCommand** (Lines 497-568)
- [x] **TestArcboxPreflightRpRegisterCommand** (Lines 568-699)

#### New Tests to Add (50.0% → 95%+ coverage)
- [ ] **TestPreflightCommand_SubcommandStructure**
  ```go
  func TestPreflightCommand_SubcommandHierarchy(t *testing.T)  // All subcommands present
  func TestPreflightCommand_QuotaSubcommand(t *testing.T)      // Quota check functionality
  func TestPreflightCommand_RPSubcommand(t *testing.T)         // Resource provider checks
  func TestPreflightCommand_RPRegisterSubcommand(t *testing.T) // RP registration
  ```

- [ ] **TestPreflightCommand_ValidationIntegration**
  ```go
  func TestPreflightCommand_ValidationServiceCalls(t *testing.T) // Service integration
  func TestPreflightCommand_QuotaServiceCalls(t *testing.T)      // Quota service integration
  func TestPreflightCommand_FullPreflightFlow(t *testing.T)      // End-to-end validation
  func TestPreflightCommand_SkipPreflightChecks(t *testing.T)    // Skip flag behavior
  ```

- [ ] **TestPreflightCommand_OutputFormats**
  ```go
  func TestPreflightCommand_TableOutput(t *testing.T)      // Table format validation
  func TestPreflightCommand_JSONOutput(t *testing.T)       // JSON format validation
  func TestPreflightCommand_VerboseOutput(t *testing.T)    // Verbose mode testing
  ```

### 6. test_helpers.go (NEW FILE - Shared Test Infrastructure)

#### Mock CLI Creation Patterns
```go
// Standard mock CLI creation
func CreateMockCLI(responses map[string]string) *azurecli.MockAzureCLI
func CreateMockCLIWithDefaults() *azurecli.MockAzureCLI
func CreateMockCLIWithErrors(errorCommands map[string]error) *azurecli.MockAzureCLI

// Test environment setup
func SetupTestEnvironment(t *testing.T) (*TestEnvironment, func())
func CreateTempDir(t *testing.T) (string, func())
func SetupTestFiles(t *testing.T, files map[string]string) (string, func())
```

#### Test Data Builders
```go
// Command test data
func BuildValidDeploymentFlags() map[string]string
func BuildInvalidDeploymentFlags() []InvalidFlagTestCase
func BuildValidDeleteFlags() map[string]string
func BuildValidListFlags() map[string]string

// Mock response builders
func BuildMockAzureResponses() MockResponseSet
func BuildMockSubscriptionResponses() map[string]string
func BuildMockResourceGroupResponses() map[string]string
func BuildMockDeploymentResponses() map[string]string
```

#### Assertion Helpers
```go
// Command structure assertions
func AssertCommandExists(t *testing.T, parent *cobra.Command, cmdName string)
func AssertFlagExists(t *testing.T, cmd *cobra.Command, flagName, flagType string)
func AssertCommandStructure(t *testing.T, cmd *cobra.Command, expected CommandSpec)

// Error and output assertions
func AssertErrorContains(t *testing.T, err error, substring string)
func AssertOutputContains(t *testing.T, output, substring string)
func AssertMockCLICalls(t *testing.T, mockCLI *azurecli.MockAzureCLI, expectedCalls []string)
```

## Coverage Enhancement Strategy

### High-Priority Coverage Targets (Week 1) - **Critical Gaps**

#### 1. SetAzureCLI Function (0.0% → 100% coverage) - **BLOCKING**
- **Impact**: Core dependency injection mechanism
- **Risk**: All CLI mocking depends on this function
- **Tests needed**: 
  - Basic functionality testing
  - Nil input handling
  - CLI replacement verification

#### 2. Delete Command (23.8% → 95% coverage) - **DATA SAFETY**
- **Impact**: Resource deletion safety, user confirmation
- **Risk**: Accidental data loss without proper validation
- **Critical paths**:
  - User confirmation flows
  - Resource existence validation
  - Error handling and rollback

#### 3. List Command (13.2% → 95% coverage) - **DISCOVERY**
- **Impact**: Resource discovery and status reporting
- **Risk**: Incorrect deployment information, poor UX
- **Critical paths**:
  - Output format handling
  - Subscription filtering
  - Flavor detection accuracy

### Medium-Priority Coverage Targets (Week 2)

#### 1. Integration Scenarios (Currently ~70% covered)
- Command chaining workflows
- State persistence across commands
- Configuration inheritance
- Environment variable handling
- Flag precedence rules

#### 2. Mock Interaction Verification (Currently ~50% covered)
- CLI command execution verification
- Service interaction patterns
- Display output verification
- Error propagation validation
- Resource cleanup verification

### Low-Priority Coverage Targets (Week 3-4)

#### 1. Edge Case Scenarios (Currently ~40% covered)
- Empty command arguments
- Malformed configuration
- Concurrent execution
- Resource cleanup failures
- Timeout conditions

#### 2. Performance and Reliability
- Benchmark testing for command creation
- Memory usage validation
- Concurrent execution safety
- Resource leak detection

## Test Pattern Standardization

### Naming Conventions
- **Pattern**: `Test[Component]_[Function]_[Scenario]`
- **Examples**: 
  - `TestDeployCommand_FlagValidation_RequiredFlags`
  - `TestListCommand_OutputFormat_JSONFormat`
  - `TestDeleteCommand_ConfirmationFlow_UserCancel`

### Test Structure Standards
```go
func TestComponent_Function_Scenario(t *testing.T) {
    // Arrange
    env, cleanup := SetupTestEnvironment(t)
    defer cleanup()
    
    mockCLI := CreateMockCLIWithDefaults()
    // Additional setup...
    
    // Act
    result, err := performOperation()
    
    // Assert
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    AssertExpectedBehavior(t, result)
    AssertMockCLICalls(t, mockCLI, expectedCalls)
}
```

### Parameterized Test Patterns
```go
func TestComponent_ComprehensiveScenarios(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(*TestEnvironment)
        input    TestInput
        want     TestOutput
        wantErr  bool
        errMsg   string
    }{
        // Comprehensive test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Implementation...
        })
    }
}
```

## Implementation Timeline

| Week | Focus | Files | Coverage Target | Critical Deliverables |
|------|-------|-------|-----------------|----------------------|
| **Week 1** | Critical gaps | `arcbox_test.go`, `delete_cmd_test.go` | 85%+ | SetAzureCLI coverage, Delete command safety |
| **Week 2** | Command files | `deploy_cmd_test.go`, `list_cmd_test.go` | 92%+ | Command-specific test files |
| **Week 3** | Integration | `preflight_cmd_test.go`, `test_helpers.go` | 95%+ | Shared infrastructure, preflight testing |
| **Week 4** | Optimization | All files | 98%+ | Edge cases, performance, documentation |

## Success Criteria

### Coverage Targets
- [ ] **arcbox.go**: 77.5% → 95%+ (SetAzureCLI: 0% → 100%)
- [ ] **deploy_cmd.go**: 60.4% → 95%+
- [ ] **delete_cmd.go**: 23.8% → 95%+ (Critical priority)
- [ ] **list_cmd.go**: 13.2% → 95%+ (Critical priority)
- [ ] **preflight_cmd.go**: 50.0% → 95%+
- [ ] **Overall package**: 48.8% → 95%+

### Quality Metrics
- [ ] **Zero functionality changes**: All existing behavior preserved
- [ ] **Test isolation**: Each test can run independently
- [ ] **Deterministic execution**: No flaky tests
- [ ] **Comprehensive error coverage**: All error paths tested
- [ ] **Mock verification**: All service interactions verified

### File Organization
- [ ] **Monolithic file split**: 1,843 lines → ~300-400 core lines + focused command files
- [ ] **Shared infrastructure**: Reusable test utilities and patterns
- [ ] **Clear separation**: Command-specific tests in dedicated files
- [ ] **Maintainable structure**: Easy to find and modify relevant tests

This blueprint provides the detailed roadmap for transforming the monolithic test file into a maintainable, comprehensive test suite with 95%+ coverage while preserving all existing functionality.
