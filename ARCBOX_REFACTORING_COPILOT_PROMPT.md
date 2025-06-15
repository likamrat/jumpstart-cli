# ArcBox Command Refactoring - GitHub Copilot Prompt

## Context and Background

The `cmd/arcbox/arcbox.go` file is a monolithic 2164-line file that handles all ArcBox functionality including command definitions, deployment logic, listing operations, quota checking, and utilities. This refactoring aims to break it into smaller, testable modules without changing any functionality.

### Current File Structure

- **Lines 1-540**: Command definitions and structure
- **Lines 540-830**: Deployment logic
- **Lines 830-950**: Status monitoring and display
- **Lines 950-1350**: List/discovery logic
- **Lines 1350-1600**: Utility functions
- **Lines 1600-2164**: Quota checking logic

### Refactoring Goals

1. Improve testability by isolating concerns
2. Enable dependency injection for better mocking
3. Create focused, single-responsibility modules
4. Maintain all existing functionality (zero behavior changes)
5. Add comprehensive unit tests for each module

### Refactoring Phases Overview

- **Phase 0**: Comprehensive codebase analysis and planning (CRITICAL - DO NOT SKIP)
- **Phase 1**: Extract models and types (foundation)
- **Phase 2**: Extract utility functions (pure functions, easy to test)
- **Phase 3**: Extract display logic (isolate UI concerns)
- **Phase 4**: Extract services (core business logic)
- **Phase 5**: Refactor commands (clean separation)
- **Phase 6**: Add comprehensive tests

---

## PHASE 0: Codebase Analysis & Planning

### Phase 0.1: Analyze Current ArcBox Structure

#### PROMPT START

Before starting the refactoring, perform a comprehensive analysis of the current ArcBox codebase:

1. **Main File Analysis** - Examine `cmd/arcbox/arcbox.go`:
   - **Line-by-line inventory**: Create a complete inventory of all functions with line numbers
   - **Struct catalog**: Identify all struct definitions, their fields, and purposes
   - **Function categorization**: Map all functions and categorize them by responsibility:
     - Command creation functions (e.g., NewArcboxCmd, subcommand creators)
     - Business logic functions (deployment, listing, quota checking)
     - Utility functions (normalization, parsing, validation, helpers)
     - Display/formatting functions (table output, status display)
     - Validation functions (used by preflight)
   - **Import analysis**: Document all dependencies and their usage patterns
   - **State analysis**: Note any global variables, package-level state, or shared data structures
   - **Error handling patterns**: Document how errors are handled and propagated

2. **External Dependencies Analysis**:
   - **Main.go integration**: Check `main.go` - identify exactly how arcbox is imported and used
   - **Preflight integration**: Examine `internal/preflight/arcbox/arcbox.go` - document:
     - Which specific arcbox functions are called (with exact signatures)
     - How these functions are invoked (parameters, return values)
     - Any error handling or result processing
   - **Examples integration**: Review `internal/examples/examples.go` - document:
     - Which example keys reference arcbox commands
     - How command names map to example content
     - Any dependencies on command structure or naming
   - **Cross-package dependencies**: Search the entire codebase for imports of the arcbox package
   - **Test file dependencies**: Check all test files that import or reference arcbox functionality

3. **Public Interface Deep Analysis**:
   - **Critical exported functions**: Document ALL exported functions with:
     - Exact function signatures (parameters and return types)
     - Current usage locations (which files call them)
     - Required behavior (what they must continue to do)
     - Any side effects or state changes they perform
   - **Command structure mapping**: Document the complete command hierarchy:
     - Main command definition and its properties
     - All subcommands and their relationships
     - Flag definitions, types, and default values
     - Command aliases and shortcuts
   - **Shared data structures**: Identify structs or types used across package boundaries
   - **Interface contracts**: Document any implicit contracts or expected behaviors

4. **Current Testing Analysis**:
   - Review `cmd/arcbox/arcbox_test.go` structure
   - Identify testing patterns and mock usage
   - Note dependency injection mechanisms (`SetAzureCLI`)
   - Document test coverage areas

Create a detailed analysis report covering:

- Function inventory with categorization
- Dependency map (internal and external)
- Public interface documentation
- Current architectural patterns
- Testing approach summary

**Analysis Documentation Template**:

```markdown
# ArcBox Codebase Analysis Report

## Executive Summary
- Current file size: [X] lines
- Number of functions: [X]
- Number of structs: [X]
- Key architectural patterns: [brief summary]

## Function Inventory

### Command Creation Functions
| Function Name | Line Range | Purpose | External Usage |
|---------------|------------|---------|----------------|
| NewArcboxCmd | [X-Y] | Main command entry | main.go |
| [etc.] | | | |

### Business Logic Functions
| Function Name | Line Range | Purpose | Dependencies | External Usage |
|---------------|------------|---------|--------------|----------------|
| [function] | [X-Y] | [purpose] | [deps] | [usage] |

### Utility Functions
| Function Name | Line Range | Purpose | Pure Function? | External Usage |
|---------------|------------|---------|----------------|----------------|
| [function] | [X-Y] | [purpose] | [yes/no] | [usage] |

### Display Functions
| Function Name | Line Range | Purpose | Dependencies | External Usage |
|---------------|------------|---------|--------------|----------------|
| [function] | [X-Y] | [purpose] | [deps] | [usage] |

## Critical External Dependencies

### Functions Called by Other Packages
| Function | Called From | Required Signature | Notes |
|----------|-------------|-------------------|-------|
| ValidateConditionalRequirements | internal/preflight/arcbox/arcbox.go | func(cmd *cobra.Command) bool | Must remain public |
| RunArcBoxPreflightChecks | internal/preflight/arcbox/arcbox.go | func(cmd *cobra.Command) bool | Must remain public |
| [etc.] | | | |

### Testing Dependencies
| Function | Purpose | Required For |
|----------|---------|--------------|
| NewArcboxCmdWithCLI | Dependency injection | All tests |
| SetAzureCLI | Mock injection | Test setup |
| [etc.] | | |

## Data Flow Analysis

### Shared Data Structures
- [List structures used across functions]
- [Note any mutable state or global variables]

### Function Dependencies
- [Map which functions call which other functions]
- [Identify circular dependencies or complex call chains]

## Risk Assessment

### High-Risk Functions (require careful handling)
- [Functions with complex external dependencies]
- [Functions with side effects]
- [Functions with unclear boundaries]

### Medium-Risk Functions
- [Functions with some dependencies but clearer boundaries]

### Low-Risk Functions (safe to extract first)
- [Pure functions or simple utilities]
- [Functions with minimal dependencies]

## Proposed Migration Strategy

### Phase 1 Targets (Low Risk)
- [List specific functions to extract first]

### Phase 2 Targets (Medium Risk)
- [Functions to extract after dependencies are resolved]

### Phase 3 Targets (High Risk)
- [Functions requiring careful handling last]

## Testing Strategy

### Current Test Coverage
- [What is currently tested]
- [Testing patterns in use]
- [Mock usage patterns]

### Required New Tests
- [What needs new tests after refactoring]
- [Testing strategy for new services]
```

#### PROMPT END

### Phase 0.2: Identify Refactoring Patterns

#### PROMPT START

Based on the analysis from Phase 0.1, identify specific refactoring patterns:

1. **Code Grouping Analysis**:
   - Group related functions that should move together
   - Identify shared dependencies between function groups
   - Map data flow between different functional areas
   - Identify potential service boundaries

2. **Dependency Analysis**:
   - Map which functions depend on `azurecli.AzureCLI`
   - Identify functions that share common data structures
   - Find functions that operate on the same domain objects
   - Identify pure utility functions with no external dependencies

3. **Testing Impact Analysis**:
   - Identify which functions are directly tested
   - Map test dependencies and mock requirements
   - Identify functions that will need new test patterns after refactoring
   - Plan testing strategy for new service architecture

4. **Migration Complexity Assessment**:
   - Rank functions by migration complexity (easy/medium/hard)
   - Identify functions with circular dependencies
   - Note functions that require careful handling due to external usage
   - Plan migration order to minimize breaking changes

Create a refactoring strategy document with:
- Proposed service boundaries and responsibilities
- Migration order and dependencies
- Testing strategy for each new component
- Risk assessment and mitigation plans

#### PROMPT END

### Phase 0.3: Validate Understanding

#### PROMPT START

Before proceeding with refactoring, validate your understanding:

1. **Build and Test Current State**:
   - Ensure `go build` succeeds
   - Run `go test ./cmd/arcbox/...` - all tests must pass
   - Run `go test ./internal/preflight/arcbox/...` - verify external dependencies work
   - Test basic arcbox commands: `go run main.go arcbox --help`

2. **Trace External Dependencies**:
   - Verify that `main.go` successfully calls `arcbox.NewArcboxCmd()`
   - Confirm that preflight package calls work correctly
   - Check that examples are properly referenced
   - Validate that all imports resolve correctly

3. **Document Current Behavior**:
   - Test each arcbox subcommand help output
   - Document current flag behavior and defaults
   - Note any specific command interactions or dependencies
   - Record current test suite behavior

4. **Create Baseline**:
   - Document current file sizes and structure
   - Record current test coverage if available
   - Note current build time and test execution time
   - Create a comprehensive functionality checklist including:
     - All command help text outputs
     - All flag behaviors and defaults
     - All error messages and their triggers
     - All success scenarios and their outputs
     - All external function signatures and behaviors

5. **Verify Understanding**:
   - **Cross-reference analysis**: Compare your function inventory with actual external usage
   - **Test signature analysis**: Verify that all functions used in tests are properly documented
   - **Import verification**: Ensure all external package imports of arcbox are accounted for
   - **Command behavior verification**: Test and document current command behavior patterns

**Critical Verification Checklist**:
- [ ] All functions called by `internal/preflight/arcbox/arcbox.go` are identified and documented
- [ ] All test dependencies (`NewArcboxCmdWithCLI`, `SetAzureCLI`) are documented
- [ ] All example keys in `internal/examples/examples.go` are mapped to actual commands
- [ ] All exported functions have documented signatures and usage locations
- [ ] All command structure and flag definitions are documented
- [ ] Current build succeeds: `go build`
- [ ] All tests pass: `go test ./cmd/arcbox/... -v`
- [ ] All related tests pass: `go test ./internal/preflight/arcbox/... -v`
- [ ] Basic command works: `go run main.go arcbox --help`

Output a comprehensive baseline report that will be used to verify the refactoring maintains all existing functionality.

#### PROMPT END

**DECLARE PHASE 0 COMPLETE WHEN**: You have a thorough understanding of the current codebase structure, all dependencies are mapped, and you have a clear refactoring strategy with minimal risk of breaking changes.

### ⚠️ CRITICAL: Phase 0 Verification Requirements

Before proceeding to Phase 1, you MUST have completed:

1. **Comprehensive Function Documentation**: Every function in `arcbox.go` must be cataloged with its purpose, dependencies, and external usage
2. **External Dependency Mapping**: All functions called by external packages must be identified and their required signatures documented
3. **Test Dependency Analysis**: All test-specific functions and their required behaviors must be documented
4. **Command Structure Documentation**: The complete command hierarchy, flags, and behaviors must be mapped
5. **Risk Assessment**: Each function must be categorized by migration risk level
6. **Verification Baseline**: All tests must currently pass and build must succeed

**DO NOT proceed to Phase 1 without completing this analysis. The success of the entire refactoring depends on the thoroughness of Phase 0.**

---

## PHASE 1: Extract Models & Types

### Phase 1.1: Create Models Package Structure

#### PROMPT START

I need to create a new models package for the ArcBox command. Create the following directory structure and files:

```text
cmd/arcbox/models/
├── deployment.go
├── quota.go  
├── subscription.go
└── resource_status.go
```

In each file, add the package declaration and prepare for model extraction:

```go
package models

// This file will contain [SPECIFIC_DOMAIN] related types and structs
```

Replace [SPECIFIC_DOMAIN] with the appropriate domain for each file (deployment, quota, subscription, resource_status).

#### PROMPT END

### Phase 1.2: Extract Deployment Models

#### PROMPT START

From the file `cmd/arcbox/arcbox.go`, extract the following types and move them to `cmd/arcbox/models/deployment.go`:

1. `ArcBoxDeployment` struct (around line 1010)
2. `resourceStatus` struct (around line 835)
3. Any other deployment-related structs or types

Rules:

- Keep the exact same struct definitions
- Add proper package declaration `package models`
- Add necessary imports
- Do not modify any field names or types
- Add descriptive comments for each struct

When done, update the imports in `arcbox.go` to reference `"jumpstartcli/cmd/arcbox/models"` and prefix the moved types with `models.`.

#### PROMPT END

### Phase 1.3: Extract Subscription Models

#### PROMPT START

From the file `cmd/arcbox/arcbox.go`, extract the following type and move it to `cmd/arcbox/models/subscription.go`:

1. `AzureSubscription` struct (around line 1055)

Rules:

- Keep the exact same struct definition
- Add proper package declaration `package models`
- Add JSON tags if they exist
- Add descriptive comments

When done, update all references in `arcbox.go` to use `models.AzureSubscription`.

**PHASE 1 VERIFICATION**: After each extraction, verify:
- [ ] `go build` succeeds
- [ ] `go test ./cmd/arcbox/... -v` passes
- [ ] `go test ./internal/preflight/arcbox/... -v` passes
- [ ] No external imports of arcbox are broken

#### PROMPT END

### Phase 1.4: Create Quota Models

#### PROMPT START

Create quota-related models in `cmd/arcbox/models/quota.go` based on the quota checking logic in `arcbox.go` (lines 1600-2164).

Analyze the quota functions and create appropriate structs for:

1. Quota check results
2. SKU information
3. vCPU requirements

Example structure:

```go
package models

type QuotaCheckResult struct {
    SKU       string
    Available int
    Limit     int
    Required  int
    CanDeploy bool
    Details   string
}

// Add other quota-related types based on the existing code
```

Do not modify the existing functions yet, just create the models that represent the data they work with.

#### PROMPT END

**DECLARE PHASE 1 COMPLETE WHEN**: All model files are created and arcbox.go compiles successfully with updated imports.

---

## PHASE 2: Extract Utility Functions

### Phase 2.1: Create Utils Package Structure

#### PROMPT START

Create a utils package for ArcBox-specific utilities:

```text
cmd/arcbox/utils/
├── normalizers.go
├── parsers.go
├── validators.go
└── azure_helpers.go
```

Each file should have the package declaration `package utils` and be prepared for function extraction.

#### PROMPT END

### Phase 2.2: Extract Normalization Functions

#### PROMPT START

From `cmd/arcbox/arcbox.go`, extract these normalization functions to `cmd/arcbox/utils/normalizers.go`:

1. `normalizeFlavorCase` (around line 1475)
2. `normalizeSqlServerEditionCase` (around line 1485)
3. `normalizeBastionSkuCase` (around line 1495)

Rules:

- Move the complete function implementations
- Keep all functionality identical
- Add proper package declaration
- Update imports in arcbox.go to reference the new location
- Test that the application still compiles and works

#### PROMPT END

### Phase 2.3: Extract Parser Functions

#### PROMPT START

From `cmd/arcbox/arcbox.go`, extract these parsing functions to `cmd/arcbox/utils/parsers.go`:

1. `parseISO8601Duration` (around line 1435)
2. `parseInt64` (around line 2100)
3. Any other parsing-related functions

Add necessary imports and ensure all function signatures remain identical.

#### PROMPT END

### Phase 2.4: Extract Azure Helper Functions

#### PROMPT START

From `cmd/arcbox/arcbox.go`, extract these Azure helper functions to `cmd/arcbox/utils/azure_helpers.go`:

1. `getSubscriptionID` (around line 1505)
2. `setAzureSubscription` (around line 1595)
3. `checkResourceGroupExists` (around line 1605)
4. `mapSKUToFamilyQuotaName` (around line 2055)
5. `getRequiredVCPUForSKU` (around line 2040)

Ensure all dependencies are properly imported and function signatures remain unchanged.

#### PROMPT END

**DECLARE PHASE 2 COMPLETE WHEN**: All utility functions are extracted and arcbox.go compiles successfully with updated imports.

---

## PHASE 3: Extract Display Logic

### Phase 3.1: Create Display Package Structure

#### PROMPT START

Create a display package for formatting and output logic:

```text
cmd/arcbox/display/
├── deployment_display.go
├── list_formatter.go
├── quota_formatter.go
└── status_display.go
```

Each file should have the package declaration `package display`.

#### PROMPT END

### Phase 3.2: Extract Deployment Display Functions

#### PROMPT START

From `cmd/arcbox/arcbox.go`, extract these deployment display functions to `cmd/arcbox/display/deployment_display.go`:

1. `printDeploymentResourceList` (around line 870)
2. `waitForDeploymentAndShowStatus` (around line 890)
3. `printDeploymentErrorDetails` (around line 1385)

Create a struct to encapsulate dependencies:

```go
package display

type DeploymentDisplay struct {
    azureCLI azurecli.AzureCLI
}

func NewDeploymentDisplay(cli azurecli.AzureCLI) *DeploymentDisplay {
    return &DeploymentDisplay{azureCLI: cli}
}
```

Convert the extracted functions to methods on this struct.

#### PROMPT END

### Phase 3.3: Extract List Formatting Functions

#### PROMPT START

From `cmd/arcbox/arcbox.go`, extract these list formatting functions to `cmd/arcbox/display/list_formatter.go`:

1. `outputArcBoxDeploymentsTable` (around line 1360)
2. `outputArcBoxDeploymentsJSON` (around line 1380)
3. `getStatusIcon` (around line 1395)
4. `scanSubscriptionWithSpinner` (around line 1530)

Create appropriate structs and methods for these functions.

#### PROMPT END

### Phase 3.4: Extract Quota Display Functions

#### PROMPT START

From `cmd/arcbox/arcbox.go`, extract quota display functions to `cmd/arcbox/display/quota_formatter.go`:

1. `runQuotaChecksWithOutput` (around line 1760)
2. `runQuotaChecksWithTable` (around line 2025)

Convert these to methods on a QuotaDisplay struct with proper dependency injection.

#### PROMPT END

**DECLARE PHASE 3 COMPLETE WHEN**: All display functions are extracted and arcbox.go compiles successfully.

**PHASE 3 VERIFICATION**: After all display extractions, verify:
- [ ] `go build` succeeds
- [ ] `go test ./cmd/arcbox/... -v` passes  
- [ ] `go test ./internal/preflight/arcbox/... -v` passes
- [ ] All external dependencies still function correctly
- [ ] Command output format and behavior unchanged

---

## PHASE 4: Extract Services

### Phase 4.1: Create Services Package Structure

#### PROMPT START

Create a services package for business logic:

```text
cmd/arcbox/services/
├── deployment_service.go
├── listing_service.go
├── monitoring_service.go
├── quota_service.go
└── validation_service.go
```

Each file should have the package declaration `package services`.

#### PROMPT END

### Phase 4.2: Extract Deployment Service

#### PROMPT START

From `cmd/arcbox/arcbox.go`, extract the core deployment logic to `cmd/arcbox/services/deployment_service.go`:

1. `deployArcboxWithParamFile` function (around line 550)
2. Related deployment helper functions

Create a service struct:

```go
package services

type DeploymentService struct {
    azureCLI azurecli.AzureCLI
    display  *display.DeploymentDisplay
}

func NewDeploymentService(cli azurecli.AzureCLI, disp *display.DeploymentDisplay) *DeploymentService {
    return &DeploymentService{
        azureCLI: cli,
        display:  disp,
    }
}

func (s *DeploymentService) Deploy(cmd *cobra.Command, args []string) error {
    // Move the deployment logic here
    return nil
}
```

Ensure all dependencies are properly injected and imports are updated.

#### PROMPT END

### Phase 4.3: Extract Listing Service

#### PROMPT START

From `cmd/arcbox/arcbox.go`, extract the listing/discovery logic to `cmd/arcbox/services/listing_service.go`:

1. `runArcBoxList` (around line 1020)
2. `discoverArcBoxDeployments` (around line 1100)
3. `isArcBoxResourceGroup` (around line 1160)
4. All the `hasArcBox*` functions (around lines 1200-1280)
5. `enrichArcBoxDeployment` (around line 1300)

Create a service struct with proper dependency injection for the Azure CLI.

#### PROMPT END

### Phase 4.4: Extract Quota Service

#### PROMPT START

From `cmd/arcbox/arcbox.go`, extract quota checking logic to `cmd/arcbox/services/quota_service.go`:

1. `validateLocations` (around line 1650)
2. `getFlavorSKUs` (around line 2030)
3. All quota-related helper functions

Create a QuotaService struct with proper dependency injection.

#### PROMPT END

**DECLARE PHASE 4 COMPLETE WHEN**: All services are extracted and core business logic is properly encapsulated.

---

## PHASE 5: Refactor Command Structure

### Phase 5.1: Extract Deploy Command

#### PROMPT START

Create a new file `cmd/arcbox/deploy_cmd.go` and extract the deploy command definition from the main `NewArcboxCmdWithCLI` function.

Create a function:

```go
package arcbox

func createDeployCommand(deployService *services.DeploymentService) *cobra.Command {
    // Move the deploy command definition here
    // Update the Run function to use deployService.Deploy()
}
```

Keep all flag definitions and validation logic, but delegate actual deployment to the service.

#### PROMPT END

### Phase 5.2: Extract Delete Command

#### PROMPT START

Create `cmd/arcbox/delete_cmd.go` and extract the delete command definition.

Create appropriate service dependencies and update the command to use service methods.

#### PROMPT END

### Phase 5.3: Extract List Command

#### PROMPT START

Create `cmd/arcbox/list_cmd.go` and extract the list command definition.

Update to use the ListingService for actual operations.

#### PROMPT END

### Phase 5.4: Extract Preflight Command

#### PROMPT START

Create `cmd/arcbox/preflight_cmd.go` and extract the preflight command definitions.

This includes the main preflight command and its subcommands (quota, rp, status).

#### PROMPT END

### Phase 5.5: Update Main ArcBox Command

#### PROMPT START

Refactor the main `cmd/arcbox/arcbox.go` file to:

1. Keep only the main command definition and dependency injection setup
2. Use the extracted command creation functions
3. Set up proper service dependencies
4. Ensure the file is now under 300 lines

Example structure:

```go
package arcbox

func NewArcboxCmd() *cobra.Command {
    return NewArcboxCmdWithCLI(defaultAzureCLI)
}

func NewArcboxCmdWithCLI(cli azurecli.AzureCLI) *cobra.Command {
    // Create services
    deployDisplay := display.NewDeploymentDisplay(cli)
    deployService := services.NewDeploymentService(cli, deployDisplay)
    // ... other services
    
    // Create main command
    arcboxCmd := &cobra.Command{
        Use:   "arcbox",
        Short: "Manage Jumpstart ArcBox automation",
        // ... rest of main command definition
    }
    
    // Add subcommands using extracted functions
    arcboxCmd.AddCommand(createDeployCommand(deployService))
    arcboxCmd.AddCommand(createDeleteCommand(...))
    arcboxCmd.AddCommand(createListCommand(...))
    arcboxCmd.AddCommand(createPreflightCommand(...))
    
    return arcboxCmd
}
```

#### PROMPT END

**DECLARE PHASE 5 COMPLETE WHEN**: The main arcbox.go file is under 300 lines and all commands are properly extracted.

**PHASE 5 CRITICAL VERIFICATION**: This is the most important verification phase:
- [ ] `go build` succeeds
- [ ] `go test ./cmd/arcbox/... -v` passes (ALL tests must pass)
- [ ] `go test ./internal/preflight/arcbox/... -v` passes
- [ ] `go test ./internal/examples/... -v` passes  
- [ ] External function signatures are preserved:
  - [ ] `ValidateConditionalRequirements(cmd *cobra.Command) bool` works
  - [ ] `RunArcBoxPreflightChecks(cmd *cobra.Command) bool` works
  - [ ] `NewArcboxCmd() *cobra.Command` works
  - [ ] `NewArcboxCmdWithCLI(cli azurecli.AzureCLI) *cobra.Command` works
  - [ ] `SetAzureCLI(cli azurecli.AzureCLI)` works
- [ ] Command behavior unchanged: `go run main.go arcbox --help`
- [ ] All subcommands work: `go run main.go arcbox deploy --help`, etc.
- [ ] Example references still resolve correctly

---

## PHASE 6: Add Comprehensive Tests

### Phase 6.1: Create Service Tests

#### PROMPT START

Create unit tests for each service in the services package:

1. `cmd/arcbox/services/deployment_service_test.go`
2. `cmd/arcbox/services/listing_service_test.go`
3. `cmd/arcbox/services/quota_service_test.go`

Use the existing `azurecli.MockAzureCLI` for mocking Azure CLI operations.

Example test structure:

```go
package services

func TestDeploymentService_Deploy(t *testing.T) {
    mockCLI := &azurecli.MockAzureCLI{}
    mockDisplay := &display.DeploymentDisplay{}
    service := NewDeploymentService(mockCLI, mockDisplay)
    
    // Test various deployment scenarios
}
```

Focus on testing the business logic without external dependencies.

#### PROMPT END

### Phase 6.2: Create Display Tests

#### PROMPT START

Create unit tests for display components:

1. `cmd/arcbox/display/deployment_display_test.go`
2. `cmd/arcbox/display/list_formatter_test.go`
3. `cmd/arcbox/display/quota_formatter_test.go`

Test output formatting without side effects by capturing output and verifying content.

#### PROMPT END

### Phase 6.3: Create Utils Tests

#### PROMPT START

Create unit tests for utility functions:

1. `cmd/arcbox/utils/normalizers_test.go`
2. `cmd/arcbox/utils/parsers_test.go`
3. `cmd/arcbox/utils/validators_test.go`

These should be pure function tests with various input/output scenarios.

#### PROMPT END

### Phase 6.4: Update Command Tests

#### PROMPT START

Update the existing `cmd/arcbox/arcbox_test.go` to work with the new structure:

1. Test command creation and structure
2. Test flag definitions and defaults
3. Test command relationships
4. Mock service dependencies for integration tests

Ensure all existing test functionality is preserved and enhanced.

#### PROMPT END

**DECLARE PHASE 6 COMPLETE WHEN**: All tests pass and coverage is maintained or improved.

---

## Final Verification

#### PROMPT START

Perform final verification of the refactoring:

1. **Core Functionality Tests**:
   - Run all arcbox tests: `go test ./cmd/arcbox/...`
   - Build the application: `go build`
   - Test basic arcbox commands to ensure functionality is preserved

2. **External Dependency Verification**:
   - Verify main.go integration: `go run main.go arcbox --help`
   - Test preflight integration: `go test ./internal/preflight/arcbox/...`
   - Test examples integration: `go test ./internal/examples/...`
   - Verify all imports resolve correctly

3. **Architecture Verification**:
   - Verify that the original 2164-line file is now broken into focused modules
   - Check that each file has a single responsibility
   - Ensure all dependencies are properly injected
   - Confirm public interfaces (NewArcboxCmd, NewArcboxCmdWithCLI, SetAzureCLI) remain intact

4. **Integration Tests**:
   - Test that `arcbox.NewArcboxCmd()` works in main.go context
   - Test that `arcbox.SetAzureCLI()` works for test dependency injection
   - Verify command structure, flags, and help text are unchanged

Create a summary report of:

- Original file size vs new structure
- Number of new files created
- Test coverage improvements
- External dependency compatibility status
- Any issues found during verification

#### PROMPT END

**DECLARE COMPLETE WHEN**: All functionality works exactly as before, tests pass, and the code is properly modularized.

---

## Success Criteria

✅ **Zero Behavior Changes**: All arcbox commands work exactly as before  
✅ **Improved Testability**: Each component can be unit tested independently  
✅ **Proper Separation of Concerns**: Models, services, display, and commands are separate  
✅ **Dependency Injection**: Services can be easily mocked for testing  
✅ **Maintainable Code**: Each file has a single, clear responsibility  
✅ **Comprehensive Tests**: All new modules have appropriate test coverage

---

## CRITICAL DEPENDENCIES & EXTERNAL INTERFACES

⚠️ **IMPORTANT**: The following external dependencies MUST be preserved during refactoring:

### 1. Main CLI Integration
- **File**: `main.go`
- **Dependency**: `import "jumpstartcli/cmd/arcbox"`
- **Usage**: `rootCmd.AddCommand(arcbox.NewArcboxCmd())`
- **Requirement**: The `NewArcboxCmd()` function MUST remain accessible and return the same *cobra.Command

### 2. Testing Interface
- **Function**: `NewArcboxCmdWithCLI(cli azurecli.AzureCLI) *cobra.Command`
- **Function**: `SetAzureCLI(cli azurecli.AzureCLI)`
- **Usage**: Used in all test files for dependency injection
- **Requirement**: These functions MUST remain public and maintain the same signatures

### 3. Internal Preflight Dependencies
- **File**: `internal/preflight/arcbox/arcbox.go`
- **Usage**: Calls functions from the arcbox package for validation
- **Functions Used**:
  - `arcbox.ValidateConditionalRequirements(cmd)`
  - `arcbox.RunArcBoxPreflightChecks(cmd)`
- **Requirement**: These functions must remain accessible and maintain signatures

### 4. Examples Package Dependencies
- **File**: `internal/examples/examples.go`
- **Usage**: Provides command examples for help text
- **Examples Keys**: `"arcbox.deploy"`, `"arcbox.delete"`, `"arcbox.list"`, `"arcbox.preflight.quota"`
- **Requirement**: Example keys must remain consistent with command structure

### 5. Package Structure Dependencies
- **Current Package**: `package arcbox`
- **Import Path**: `jumpstartcli/cmd/arcbox`
- **Requirement**: Main package name and import path MUST NOT change

### 6. Test Dependencies
- **Files**: All `*_test.go` files in the arcbox package
- **Dependencies**: Tests depend on command structure, flag names, and public functions
- **Requirement**: All existing test functionality must continue to work

## ⚠️ CRITICAL PRESERVATION REQUIREMENTS

The following specific functions are called by external packages and MUST remain accessible with identical signatures:

### From `internal/preflight/arcbox/arcbox.go`:
- `arcbox.ValidateConditionalRequirements(cmd *cobra.Command) bool`
- `arcbox.RunArcBoxPreflightChecks(cmd *cobra.Command) bool`

These functions may need to be:
1. Kept in the main `arcbox.go` file, OR
2. Re-exported from the main package if moved to services, OR  
3. Moved to a dedicated public interface package

### Testing Functions (MUST remain public):
- `NewArcboxCmd() *cobra.Command`
- `NewArcboxCmdWithCLI(cli azurecli.AzureCLI) *cobra.Command`
- `SetAzureCLI(cli azurecli.AzureCLI)`

---

## How to Use This Prompt

### Step 1: Complete Phase 0 (Analysis)
**CRITICAL**: Do not skip Phase 0. Complete all analysis sub-phases and create the documentation using the provided template. This analysis will inform all subsequent phases and help avoid breaking changes.

### Step 2: Execute Phases 1-6 Sequentially  
Use the findings from Phase 0 to:
- Identify the exact functions to extract in each phase
- Understand dependencies between functions
- Plan the migration order to minimize risks
- Validate that external interfaces remain intact

### Step 3: Use Phase 0 Documentation Throughout
Refer back to your Phase 0 analysis documentation throughout the refactoring to:
- Ensure all external dependencies are preserved
- Verify that public interfaces remain unchanged
- Check that no functionality is lost during migration
- Validate that testing approaches are compatible

### Step 4: Verify After Every Phase
After EACH phase (1-6), run the verification steps provided. Do not proceed to the next phase until all verifications pass. This prevents accumulating problems that become harder to fix later.

### Step 5: Final Comprehensive Verification
After completing all phases, run the complete verification checklist to ensure the refactoring was successful and maintains 100% backward compatibility.

---

## 🎯 REFACTORING SUCCESS CHECKLIST

Upon completion of all phases, verify the following comprehensive checklist:

### ✅ **Structural Success**
- [ ] Original 2164-line `arcbox.go` is now under 300 lines
- [ ] At least 6 new focused modules created (models, services, display, commands)
- [ ] Each new file has a single, clear responsibility
- [ ] Proper package structure and organization established

### ✅ **Functional Preservation**
- [ ] `go build` succeeds without errors
- [ ] All existing tests pass: `go test ./cmd/arcbox/... -v`
- [ ] Preflight integration works: `go test ./internal/preflight/arcbox/... -v`
- [ ] Examples integration works: `go test ./internal/examples/... -v`
- [ ] Full test suite passes: `go test ./... -v`

### ✅ **External Interface Compatibility**
- [ ] `main.go` integration unchanged: `go run main.go arcbox --help`
- [ ] All arcbox subcommands work: deploy, delete, list, preflight
- [ ] Function signatures preserved:
  - [ ] `NewArcboxCmd() *cobra.Command`
  - [ ] `NewArcboxCmdWithCLI(cli azurecli.AzureCLI) *cobra.Command`
  - [ ] `SetAzureCLI(cli azurecli.AzureCLI)`
  - [ ] `ValidateConditionalRequirements(cmd *cobra.Command) bool`
  - [ ] `RunArcBoxPreflightChecks(cmd *cobra.Command) bool`

### ✅ **Quality Improvements**
- [ ] Test coverage significantly improved
- [ ] All new modules have comprehensive unit tests
- [ ] Dependency injection properly implemented
- [ ] Services can be easily mocked for testing
- [ ] Clear separation of concerns achieved

### ✅ **Documentation & Maintainability**
- [ ] All new packages have proper documentation
- [ ] Function and struct comments are clear and accurate
- [ ] Import statements are organized and minimal
- [ ] Code follows Go best practices and conventions

**🏆 REFACTORING COMPLETE**: When all items above are checked, the massive arcbox.go refactor is successfully complete with zero behavior changes and dramatically improved maintainability and testability!
