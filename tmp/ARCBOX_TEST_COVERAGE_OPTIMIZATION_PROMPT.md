# ArcBox Test Coverage Optimization - GitHub Copilot Prompt

## Executive Summary

This prompt guides an **incremental, modular approach** to optimizing test coverage for the ArcBox CLI by splitting the monolithic 1,842-line test file into focused, maintainable components. Each command test file can be implemented **independently**, allowing for parallel development and easier review.

### Quick Start Options

**Option A: Full Implementation** - Complete all phases for comprehensive coverage optimization  
**Option B: Single Command Focus** - Pick one command (deploy, delete, list, or preflight) and optimize only that component  
**Option C: Incremental Approach** - Implement one command at a time over multiple iterations

### Key Benefits

- 🎯 **Focused Development**: Each command test file is self-contained (200-400 lines vs 1,842 lines)
- 🔄 **Incremental Progress**: Coverage improves with each completed command  
- 👥 **Team Collaboration**: Multiple developers can work on different commands simultaneously
- 🚀 **Faster Feedback**: Smaller chunks enable quicker code reviews and testing
- 📊 **Measurable Progress**: Clear coverage metrics for each component

### Command-Specific Jump Points

Each command test optimization is **independent** and can be tackled separately:

| Command | Complexity | Time Estimate | Coverage Gain | Jump To |
|---------|------------|---------------|---------------|---------|
| **Delete** | Low | 1-2 days | 15-20% | [Phase 3.1.2](#phase-312-delete-command-tests-delete_cmd_testgo) |
| **List** | Medium | 2-3 days | 20-25% | [Phase 3.1.3](#phase-313-list-command-tests-list_cmd_testgo) |
| **Deploy** | High | 2-3 days | 25-30% | [Phase 3.1.1](#phase-311-deploy-command-tests-deploy_cmd_testgo) |
| **Preflight** | High | 3-4 days | 30-35% | [Phase 3.1.4](#phase-314-preflight-command-tests-preflight_cmd_testgo) |

**Pick any command above and jump directly to its section for immediate implementation.**

---

## Context and Background

The ArcBox CLI has undergone a successful modular refactoring from a single 2164-line file into focused, testable components organized across services, display, and utility modules. The current test coverage shows excellent progress:

- **Main package**: 48.8% coverage (1842-line `arcbox_test.go`)
- **Services**: 59.0% coverage (individual test files)
- **Display**: 43.3% coverage (individual test files)  
- **Utils**: 96.8% coverage (individual test files)

### Current Test Structure Status

**Existing Modular Tests** (Already Implemented):
- ✅ `cmd/arcbox/services/`: 6 service files with dedicated test files
- ✅ `cmd/arcbox/display/`: 4 display files with dedicated test files
- ✅ `cmd/arcbox/utils/`: 4 utility files with dedicated test files

**Coverage Optimization Target** (This Effort):
- 🎯 Split the monolithic `cmd/arcbox/arcbox_test.go` (1842 lines, 19 test functions)
- 🎯 Achieve 95-100% test coverage across all ArcBox modules
- 🎯 Create focused, maintainable test files with clear responsibilities

### Monolithic Test File Analysis

The current `cmd/arcbox/arcbox_test.go` contains 19 comprehensive test functions covering:

1. **Command Structure Tests** (Lines 34-140):
   - `TestNewArcboxCmd` - Command creation and subcommand validation
   - Tests basic command properties, subcommand existence, and hierarchy

2. **Individual Subcommand Tests** (Lines 74-568):
   - `TestArcboxDeployCommand` - Deploy command structure and flags
   - `TestArcboxDeleteCommand` - Delete command structure and flags
   - `TestArcboxListCommand` - List command structure and flags
   - `TestArcboxPreflightCommand` - Preflight command structure and flags
   - `TestArcboxPreflightQuotaCommand` - Quota subcommand validation
   - `TestArcboxPreflightRpCommand` - Resource provider subcommand validation
   - `TestArcboxPreflightRpRegisterCommand` - RP register subcommand validation

3. **Integration and Factory Tests** (Lines 699-1131):
   - `TestNewArcboxCmdWithCLI_Comprehensive` - Factory function with CLI injection
   - `TestNewArcboxCmdWithCLI_AllSubcommands` - CLI integration across all commands
   - `TestNewArcboxCmdWithCLI_CommandProperties` - Command property validation
   - `TestNewArcboxCmdWithCLI_ErrorHandling` - Error scenario testing
   - `TestNewArcboxCmdWithCLI_NilInputHandling` - Nil input handling

4. **Business Logic Tests** (Lines 1131-1505):
   - `TestDetectArcBoxFlavor` - Flavor detection algorithm (extensive test cases)
   - `TestDetectArcBoxFlavorFallback` - Fallback scenarios for flavor detection

5. **Command Execution Tests** (Lines 1505-1646):
   - `TestNewArcboxCmdWithCLI_RunEFunction_EdgeCases` - Command execution edge cases

6. **Utility Function Tests** (Lines 1646-1842):
   - `TestNormalizeFlavorCase` - Flavor normalization logic
   - `TestNormalizeSqlServerEditionCase` - SQL Server edition normalization
   - `TestNormalizeBastionSkuCase` - Bastion SKU normalization

### Test Coverage Optimization Goals

1. **Achieve 95-100% coverage** by creating comprehensive, focused test suites
2. **Split monolithic test file** into logical, maintainable test modules
3. **Improve test organization** for better discoverability and maintenance
4. **Maintain existing test quality** while expanding coverage
5. **Ensure zero functionality changes** - all existing behavior must be preserved

### Incremental Implementation Strategy

This optimization can be implemented **incrementally and independently** by command:

#### Implementation Order (Recommended)
1. **Phase 1**: Coverage analysis and planning (1-2 days)
2. **Phase 2**: Core command tests enhancement (2-3 days)
3. **Phase 3**: Individual command test files (can be done in parallel):
   - **3.1**: Deploy command tests (2-3 days)
   - **3.2**: Delete command tests (1-2 days) 
   - **3.3**: List command tests (2-3 days)
   - **3.4**: Preflight command tests (3-4 days)
4. **Phase 4**: Final coverage optimization (2-3 days)
5. **Phase 5**: Documentation and validation (1-2 days)

#### Benefits of Incremental Approach
- ✅ **Smaller, manageable chunks** - Each command can be tackled separately
- ✅ **Parallel development** - Multiple developers can work on different commands
- ✅ **Faster feedback** - Each increment can be reviewed and merged independently
- ✅ **Risk mitigation** - Issues can be caught and fixed in smaller scopes
- ✅ **Progressive improvement** - Coverage improves incrementally with each step

### Test File Organization Strategy

Based on the current modular architecture, the test files should be organized as:

```
cmd/arcbox/
├── arcbox.go                           # Main command factory (keep core tests here)
├── arcbox_test.go                      # Core command tests only (SPLIT THIS)
├── delete_cmd.go                       # Delete command
├── delete_cmd_test.go                  # NEW: Delete command tests
├── deploy_cmd.go                       # Deploy command  
├── deploy_cmd_test.go                  # NEW: Deploy command tests
├── list_cmd.go                         # List command
├── list_cmd_test.go                    # NEW: List command tests
├── preflight_cmd.go                    # Preflight command
├── preflight_cmd_test.go               # NEW: Preflight command tests
├── services/                           # ✅ Already done
├── display/                            # ✅ Already done
└── utils/                              # ✅ Already done
```

---

## PHASE 1: Test Coverage Analysis & Planning

### Quick Start Guide

Choose your approach based on available time and team size:

#### 🚀 Quick Win (1-2 days): Single Command Focus
Pick one command to optimize:
- **Easiest**: Delete command (simple logic, fewer edge cases)
- **Most Impact**: Deploy command (complex logic, high usage)
- **Best ROI**: List command (moderate complexity, good coverage gains)

Jump to the specific command section in Phase 3 (3.1.1 through 3.1.4).

#### 🎯 Systematic Approach (1-2 weeks): Full Implementation
Follow all phases in order for comprehensive optimization.

#### 🔄 Incremental Approach (3-4 sprints): Team Collaboration
Assign each command to different team members:
- **Sprint 1**: Coverage analysis + Deploy command
- **Sprint 2**: Delete + List commands  
- **Sprint 3**: Preflight command + Core tests
- **Sprint 4**: Final optimization + Documentation

---

### Phase 1.1: Comprehensive Test Coverage Assessment

#### PROMPT START

Before reorganizing tests, perform a detailed analysis of current test coverage and gaps:

1. **Current Coverage Analysis**:
   - Run detailed coverage analysis: `go test -v -coverprofile=coverage.out ./cmd/arcbox/...`
   - Generate coverage report: `go tool cover -html=coverage.out -o coverage.html`
   - Identify **uncovered code paths** in each module:
     - Main command factory functions
     - Error handling branches
     - Edge cases in business logic
     - CLI integration points
     - Validation logic gaps

2. **Test Function Inventory**:
   - Create complete inventory of all 19 test functions in `arcbox_test.go`
   - **Map each test to target modules**:
     - Which tests belong with `arcbox.go` (core command structure)
     - Which tests should move to `*_cmd_test.go` files
     - Which tests cover cross-cutting concerns
   - **Identify test dependencies and shared setup**:
     - Mock CLI usage patterns
     - Shared test data structures
     - Common assertion patterns
     - Test utility functions

3. **Coverage Gap Identification**:
   - **Missing test scenarios** for each module:
     - Untested error conditions
     - Boundary value testing gaps
     - Integration scenario gaps
     - Mock interaction coverage
   - **Insufficient test depth**:
     - Functions with partial coverage
     - Complex business logic branches
     - Error propagation paths
     - State management scenarios

4. **Test Quality Assessment**:
   - **Test maintainability** evaluation:
     - Test complexity and readability
     - Mock usage effectiveness
     - Assertion clarity and coverage
     - Test data organization
   - **Test reliability** analysis:
     - Flaky test identification
     - Environment dependencies
     - Race condition potential
     - Resource cleanup patterns

Create a comprehensive coverage analysis report:

```markdown
## Current Test Coverage Analysis

### Coverage by Module
- **arcbox.go**: X% coverage
  - Uncovered lines: [specific line numbers]
  - Missing scenarios: [list gaps]
  
- **deploy_cmd.go**: X% coverage
  - Uncovered lines: [specific line numbers]
  - Missing scenarios: [list gaps]

[Continue for all modules...]

### Test Function Mapping
| Current Test Function | Target Module | Coverage Focus | Action Needed |
|----------------------|---------------|----------------|---------------|
| TestNewArcboxCmd | arcbox_test.go | Core command | Keep, expand |
| TestArcboxDeployCommand | deploy_cmd_test.go | Deploy command | Move, enhance |
[Continue for all 19 functions...]

### Critical Coverage Gaps
1. **High Priority** (Security/Stability):
   - [Specific uncovered critical paths]
   
2. **Medium Priority** (Business Logic):
   - [Specific uncovered business logic]
   
3. **Low Priority** (Edge Cases):
   - [Specific uncovered edge cases]
```

**Analysis Requirements**:
- Use actual coverage data from `go test -cover`
- Document specific line numbers for uncovered code
- Prioritize coverage gaps by business impact
- Create actionable test enhancement plan

#### PROMPT END

---

## PHASE 2: Test File Reorganization Planning

### Phase 2.1: Create Test File Organization Blueprint

#### PROMPT START

Based on the coverage analysis, create a detailed plan for reorganizing the monolithic test file:

1. **Test File Structure Planning**:
   - **Core Command Tests** (`arcbox_test.go`):
     - `TestNewArcboxCmd` - Basic command structure
     - `TestNewArcboxCmdWithCLI_*` - CLI integration tests  
     - Cross-cutting integration scenarios
     - Command hierarchy validation
     - Keep ~300-400 lines focused on core functionality

   - **Deploy Command Tests** (`deploy_cmd_test.go`):
     - Move `TestArcboxDeployCommand` from main test file
     - Add comprehensive flag validation tests
     - Add deployment flow testing with mocks
     - Add error scenario coverage
     - Target: 95%+ coverage of deploy_cmd.go

   - **Delete Command Tests** (`delete_cmd_test.go`):
     - Move `TestArcboxDeleteCommand` from main test file
     - Add deletion flow testing with mocks
     - Add confirmation prompt testing
     - Add error scenario coverage
     - Target: 95%+ coverage of delete_cmd.go

   - **List Command Tests** (`list_cmd_test.go`):
     - Move `TestArcboxListCommand` from main test file
     - Add output format testing (table, JSON, etc.)
     - Add subscription filtering tests
     - Add flavor detection testing
     - Target: 95%+ coverage of list_cmd.go

   - **Preflight Command Tests** (`preflight_cmd_test.go`):
     - Move `TestArcboxPreflightCommand` and related tests
     - Add comprehensive subcommand testing
     - Add quota validation testing  
     - Add resource provider testing
     - Target: 95%+ coverage of preflight_cmd.go

2. **Test Infrastructure Planning**:
   - **Shared Test Utilities**:
     - Create `test_helpers.go` for common test utilities
     - Mock CLI creation patterns
     - Test data builders
     - Assertion helpers
     - Setup/teardown patterns

   - **Test Data Organization**:
     - Create consistent test data structures
     - Shared mock responses
     - Parameterized test cases
     - Test scenario definitions

3. **Coverage Enhancement Planning**:
   - **New Test Scenarios** needed for 95%+ coverage:
     - Error condition testing (CLI failures, network issues)
     - Edge case validation (empty inputs, invalid data)
     - Integration scenarios (command chaining, state management)
     - Mock interaction verification
     - Concurrency safety testing

   - **Test Pattern Standardization**:
     - Consistent test naming conventions
     - Standardized mock setup patterns
     - Common assertion approaches
     - Error testing strategies

Create detailed reorganization blueprint:

```markdown
## Test File Reorganization Blueprint

### File-by-File Migration Plan

#### 1. arcbox_test.go (Target: 300-400 lines, 95%+ coverage)
**Keep These Tests**:
- [ ] TestNewArcboxCmd (expand with more edge cases)
- [ ] TestNewArcboxCmdWithCLI_Comprehensive (enhance CLI scenarios)
- [ ] TestNewArcboxCmdWithCLI_AllSubcommands (expand integration)
- [ ] TestNewArcboxCmdWithCLI_CommandProperties (add property validation)
- [ ] TestNewArcboxCmdWithCLI_ErrorHandling (add more error scenarios)
- [ ] TestNewArcboxCmdWithCLI_NilInputHandling (expand nil scenarios)

**New Tests to Add**:
- [ ] TestSetAzureCLI (CLI injection mechanism)
- [ ] TestCommandHierarchy (full command tree validation)
- [ ] TestCommandIntegration (cross-command interactions)
- [ ] TestFactoryErrorHandling (factory function error scenarios)

#### 2. deploy_cmd_test.go (NEW FILE, Target: 95%+ coverage)
**Migrate These Tests**:
- [ ] TestArcboxDeployCommand (from arcbox_test.go)

**New Tests to Add**:
- [ ] TestDeployCommandFlags (comprehensive flag validation)
- [ ] TestDeployCommandExecution (deployment flow with mocks)
- [ ] TestDeployCommandValidation (parameter validation)
- [ ] TestDeployCommandErrorHandling (error scenarios)
- [ ] TestDeployCommandTemplateHandling (template path resolution)
- [ ] TestDeployCommandOutputFormats (output format testing)

[Continue for all command test files...]

### Coverage Enhancement Strategy

#### High-Priority Coverage Targets (Week 1)
1. **Error Handling Paths** (Currently ~60% covered):
   - CLI command failures
   - Network connectivity issues  
   - Invalid user inputs
   - Permission errors
   - Resource conflicts

2. **Edge Case Scenarios** (Currently ~40% covered):
   - Empty command arguments
   - Malformed configuration
   - Concurrent execution
   - Resource cleanup failures
   - Timeout conditions

#### Medium-Priority Coverage Targets (Week 2)
1. **Integration Scenarios** (Currently ~70% covered):
   - Command chaining workflows
   - State persistence across commands
   - Configuration inheritance
   - Environment variable handling
   - Flag precedence rules

2. **Mock Interaction Verification** (Currently ~50% covered):
   - CLI command execution verification
   - Service interaction patterns
   - Display output verification
   - Error propagation validation
   - Resource cleanup verification

### Test Infrastructure Enhancements

#### Shared Test Utilities (test_helpers.go)
```go
// Mock CLI creation patterns
func createMockCLI(responses map[string]string) *azurecli.MockAzureCLI
func setupTestEnvironment() (*TestEnv, func())
func assertCommandStructure(t *testing.T, cmd *cobra.Command, expected CommandSpec)

// Test data builders  
func buildValidDeploymentFlags() DeploymentFlags
func buildInvalidDeploymentFlags() []InvalidFlagCase
func buildMockAzureResponses() MockResponseSet

// Assertion helpers
func assertErrorContains(t *testing.T, err error, substring string)
func assertCommandExists(t *testing.T, parent *cobra.Command, cmdName string)
func assertFlagExists(t *testing.T, cmd *cobra.Command, flagName string, flagType string)
```

#### Test Pattern Standards
- **Naming**: `Test[Component][Function][Scenario]`
- **Structure**: Arrange, Act, Assert with clear separation
- **Mocking**: Consistent mock setup and verification
- **Data**: Parameterized tests for multiple scenarios
- **Cleanup**: Proper resource cleanup in all tests
```

**Blueprint Requirements**:
- Specific migration plan for each test function
- Coverage targets for each new test file
- Infrastructure improvements needed
- Timeline for implementation

#### PROMPT END

---

## PHASE 3: Test File Migration Implementation

### Phase 3.1: Create Command-Specific Test Files (Incremental Approach)

The following phases can be implemented **independently** and **incrementally**. Each command test file can be created and optimized separately, allowing for focused development and easier review.

---

#### Phase 3.1.1: Deploy Command Tests (`deploy_cmd_test.go`)

##### PROMPT START - Deploy Command Testing

Create comprehensive tests for the deploy command with incremental implementation:

**Step 1: Basic Migration and Structure**
- **Migrate existing test** from `arcbox_test.go`:
  - `TestArcboxDeployCommand` → `TestDeployCommand_BasicStructure`
- **Create test file foundation** with imports and helper setup
- **Target**: Basic test migration working (70%+ coverage)

**Step 2: Flag Validation Enhancement**
- **Add comprehensive flag testing**:
  - `TestDeployCommand_RequiredFlags` - Resource group, subscription validation
  - `TestDeployCommand_OptionalFlags` - Template paths, parameter files
  - `TestDeployCommand_FlagDefaults` - Default values and behavior
- **Target**: All flag scenarios covered (80%+ coverage)

**Step 3: Service Integration Testing**
- **Add service interaction tests**:
  - `TestDeployCommand_ServiceIntegration` - Deployment service calls
  - `TestDeployCommand_ParameterValidation` - Parameter processing
  - `TestDeployCommand_TemplateHandling` - Local vs remote templates
- **Target**: Service integration covered (85%+ coverage)

**Step 4: Error Scenarios and Edge Cases**
- **Add comprehensive error testing**:
  - `TestDeployCommand_ErrorScenarios` - CLI failures, validation errors
  - `TestDeployCommand_EdgeCases` - Empty inputs, invalid data
- **Target**: 95%+ coverage of `deploy_cmd.go`

##### PROMPT END - Deploy Command Testing

---

#### Phase 3.1.2: Delete Command Tests (`delete_cmd_test.go`)

##### PROMPT START - Delete Command Testing

Create comprehensive tests for the delete command with incremental implementation:

**Step 1: Basic Migration and Structure**
- **Migrate existing test** from `arcbox_test.go`:
  - `TestArcboxDeleteCommand` → `TestDeleteCommand_BasicStructure`
- **Create test file foundation** with confirmation testing setup
- **Target**: Basic test migration working (70%+ coverage)

**Step 2: Confirmation Flow Testing**
- **Add user interaction tests**:
  - `TestDeleteCommand_ConfirmationPrompt` - User confirmation scenarios
  - `TestDeleteCommand_SkipConfirmation` - Automated deletion flow
  - `TestDeleteCommand_ConfirmationCancellation` - User cancels operation
- **Target**: User interaction covered (80%+ coverage)

**Step 3: Service Integration and Validation**
- **Add service interaction tests**:
  - `TestDeleteCommand_ServiceIntegration` - Deletion service calls
  - `TestDeleteCommand_ResourceValidation` - Resource existence checking
  - `TestDeleteCommand_DeletionFlow` - End-to-end deletion process
- **Target**: Service integration covered (85%+ coverage)

**Step 4: Error Scenarios and Edge Cases**
- **Add comprehensive error testing**:
  - `TestDeleteCommand_ErrorScenarios` - Resource not found, permission errors
  - `TestDeleteCommand_EdgeCases` - Invalid resource groups, CLI failures
- **Target**: 95%+ coverage of `delete_cmd.go`

##### PROMPT END - Delete Command Testing

---

#### Phase 3.1.3: List Command Tests (`list_cmd_test.go`)

##### PROMPT START - List Command Testing

Create comprehensive tests for the list command with incremental implementation:

**Step 1: Basic Migration and Structure**
- **Migrate existing test** from `arcbox_test.go`:
  - `TestArcboxListCommand` → `TestListCommand_BasicStructure`
- **Create test file foundation** with output format testing setup
- **Target**: Basic test migration working (70%+ coverage)

**Step 2: Output Format Testing**
- **Add output format tests**:
  - `TestListCommand_TableOutput` - Default table format
  - `TestListCommand_JSONOutput` - JSON format validation
  - `TestListCommand_YAMLOutput` - YAML format validation
- **Target**: Output formats covered (80%+ coverage)

**Step 3: Subscription and Filtering**
- **Add subscription handling tests**:
  - `TestListCommand_SubscriptionHandling` - Current, specific, all subscriptions
  - `TestListCommand_FlavorFiltering` - Flavor detection and filtering
  - `TestListCommand_ServiceIntegration` - Listing service interaction
- **Target**: Filtering and subscriptions covered (85%+ coverage)

**Step 4: Edge Cases and Error Scenarios**
- **Add comprehensive edge case testing**:
  - `TestListCommand_EmptyResults` - No deployments found scenarios
  - `TestListCommand_ErrorScenarios` - CLI failures, invalid subscriptions
- **Target**: 95%+ coverage of `list_cmd.go`

##### PROMPT END - List Command Testing

---

#### Phase 3.1.4: Preflight Command Tests (`preflight_cmd_test.go`)

##### PROMPT START - Preflight Command Testing

Create comprehensive tests for the preflight command with incremental implementation:

**Step 1: Basic Migration and Structure**
- **Migrate existing tests** from `arcbox_test.go`:
  - `TestArcboxPreflightCommand` → `TestPreflightCommand_BasicStructure`
  - `TestArcboxPreflightQuotaCommand` → `TestPreflightCommand_QuotaSubcommand`
- **Create test file foundation** with subcommand testing setup
- **Target**: Basic test migration working (70%+ coverage)

**Step 2: Subcommand Structure Testing**
- **Add subcommand validation tests**:
  - `TestPreflightCommand_SubcommandHierarchy` - All subcommands present
  - `TestPreflightCommand_QuotaSubcommand` - Quota check functionality
  - `TestPreflightCommand_RPSubcommand` - Resource provider checks
- **Target**: Subcommand structure covered (80%+ coverage)

**Step 3: Resource Provider Testing**
- **Add RP-specific tests**:
  - `TestPreflightCommand_RPValidation` - RP status checking
  - `TestPreflightCommand_RPRegistration` - RP registration process
  - **Migrate remaining tests**: `TestArcboxPreflightRpCommand`, `TestArcboxPreflightRpRegisterCommand`
- **Target**: RP functionality covered (85%+ coverage)

**Step 4: Integration and Error Scenarios**
- **Add comprehensive testing**:
  - `TestPreflightCommand_ServiceIntegration` - Validation service integration
  - `TestPreflightCommand_ErrorScenarios` - CLI failures, quota insufficient
  - `TestPreflightCommand_OutputFormats` - Table, JSON output validation
- **Target**: 95%+ coverage of `preflight_cmd.go`

##### PROMPT END - Preflight Command Testing

**Implementation Requirements**:

```go
// Example test structure for deploy_cmd_test.go
package arcbox

import (
	"strings"
	"testing"
	"jumpstartcli/internal/azurecli"
	"github.com/spf13/cobra"
)

func TestDeployCommand_FlagValidation(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		setup    func(*azurecli.MockAzureCLI)
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "missing_resource_group",
			args:    []string{"deploy", "--subscription", "test-sub"},
			wantErr: true,
			errMsg:  "resource-group is required",
		},
		{
			name:    "valid_flags",
			args:    []string{"deploy", "--resource-group", "test-rg", "--subscription", "test-sub"},
			setup:   func(m *azurecli.MockAzureCLI) { /* setup mock responses */ },
			wantErr: false,
		},
		// Add comprehensive flag testing...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := azurecli.NewMockAzureCLI()
			if tt.setup != nil {
				tt.setup(mockCLI)
			}
			
			cmd := NewArcboxCmdWithCLI(mockCLI)
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDeployCommand_ServiceIntegration(t *testing.T) {
	// Test integration between command and deployment service
	// Verify service method calls, parameter passing, error handling
}

func TestDeployCommand_ErrorScenarios(t *testing.T) {
	// Comprehensive error scenario testing
	// CLI failures, validation failures, service errors
}

// Continue with all required test functions...
```

**For each new test file, ensure**:
- Comprehensive flag validation (all flags, types, defaults, validation rules)
- Service integration testing (verify service calls, parameters, responses)  
- Error scenario coverage (all error paths, proper error messages)
- Mock usage patterns (consistent with existing tests)
- Business logic validation (parameter validation, conditional logic)
- Output format testing (where applicable)
- Edge case handling (empty inputs, invalid data, boundary conditions)

**Coverage verification for each file**:
- Run `go test -cover ./cmd/arcbox/[command]_cmd_test.go` 
- Verify 95%+ coverage for corresponding command file
- Add additional tests for any uncovered lines
- Document coverage results

#### PROMPT END

### Phase 3.2: Enhance Core Command Tests

#### PROMPT START

Streamline and enhance the core command tests in `arcbox_test.go` to focus on integration and factory functions:

1. **Refactor `arcbox_test.go`** (Target: 300-400 lines, 95%+ coverage):
   - **Keep and enhance core tests**:
     - `TestNewArcboxCmd` → Add comprehensive command structure validation
     - `TestNewArcboxCmdWithCLI_*` → Enhance CLI integration scenarios
     - Add `TestSetAzureCLI` → Test CLI injection mechanism
     - Add `TestCommandHierarchy` → Validate complete command tree
     - Add `TestFactoryErrorHandling` → Factory function error scenarios

   - **Remove migrated tests**: 
     - Remove `TestArcboxDeployCommand` (migrated to `deploy_cmd_test.go`)
     - Remove `TestArcboxDeleteCommand` (migrated to `delete_cmd_test.go`)
     - Remove `TestArcboxListCommand` (migrated to `list_cmd_test.go`)
     - Remove `TestArcboxPreflight*Command` tests (migrated to `preflight_cmd_test.go`)

   - **Keep business logic tests**:
     - `TestDetectArcBoxFlavor` → Move to more appropriate location or keep if used by core
     - `TestDetectArcBoxFlavorFallback` → Same as above
     - `TestNormalize*Case` tests → Move to `utils/` if not already there

2. **Add comprehensive integration tests**:
   - **Command Tree Validation**:
     ```go
     func TestCommandHierarchy(t *testing.T) {
         tests := []struct {
             name        string
             commandPath []string
             shouldExist bool
         }{
             {"root", []string{"arcbox"}, true},
             {"deploy", []string{"arcbox", "deploy"}, true},
             {"delete", []string{"arcbox", "delete"}, true},
             {"list", []string{"arcbox", "list"}, true},
             {"preflight", []string{"arcbox", "preflight"}, true},
             {"preflight_quota", []string{"arcbox", "preflight", "quota"}, true},
             {"preflight_rp", []string{"arcbox", "preflight", "rp"}, true},
             {"preflight_rp_register", []string{"arcbox", "preflight", "rp", "register"}, true},
             {"invalid", []string{"arcbox", "invalid"}, false},
         }
         // Implement comprehensive command tree validation
     }
     ```

   - **CLI Integration Testing**:
     ```go
     func TestSetAzureCLI(t *testing.T) {
         // Test CLI injection mechanism
         // Verify CLI is properly set across all commands
         // Test error scenarios (nil CLI, invalid CLI)
     }

     func TestCLIIntegrationAcrossCommands(t *testing.T) {
         // Test that CLI injection works for all subcommands
         // Verify mock CLI responses are properly used
         // Test CLI state sharing between commands
     }
     ```

   - **Factory Function Testing**:
     ```go
     func TestFactoryErrorHandling(t *testing.T) {
         // Test NewArcboxCmd error scenarios
         // Test NewArcboxCmdWithCLI error scenarios
         // Test invalid configurations
     }
     ```

3. **Move business logic tests to appropriate locations**:
   - **Flavor detection tests**: If `detectArcBoxFlavor` is used only by listing service, move to `services/listing_service_test.go`
   - **Normalization tests**: If not already in `utils/normalizers_test.go`, move them there
   - **Core command logic**: Keep only tests that are specific to the main command factory

4. **Add missing core functionality tests**:
   - **Command configuration validation**
   - **Flag inheritance testing** (global flags vs command-specific flags)
   - **Help text validation** (ensure help text is properly set)
   - **Command aliases testing** (if any aliases exist)
   - **Error message formatting** (ensure consistent error messaging)

**Updated test file structure**:
```go
// arcbox_test.go - Core command and integration tests only

// Core Command Structure Tests
func TestNewArcboxCmd(t *testing.T) { /* Enhanced with more validation */ }

// CLI Integration Tests  
func TestNewArcboxCmdWithCLI_Comprehensive(t *testing.T) { /* Enhanced scenarios */ }
func TestNewArcboxCmdWithCLI_AllSubcommands(t *testing.T) { /* Enhanced coverage */ }
func TestNewArcboxCmdWithCLI_CommandProperties(t *testing.T) { /* Enhanced properties */ }
func TestNewArcboxCmdWithCLI_ErrorHandling(t *testing.T) { /* Enhanced error scenarios */ }
func TestNewArcboxCmdWithCLI_NilInputHandling(t *testing.T) { /* Enhanced nil scenarios */ }
func TestSetAzureCLI(t *testing.T) { /* NEW: CLI injection testing */ }

// Integration and Factory Tests
func TestCommandHierarchy(t *testing.T) { /* NEW: Complete command tree validation */ }
func TestFactoryErrorHandling(t *testing.T) { /* NEW: Factory error scenarios */ }
func TestCLIIntegrationAcrossCommands(t *testing.T) { /* NEW: Cross-command CLI testing */ }

// Keep only if used by core command logic, otherwise move to appropriate locations
func TestDetectArcBoxFlavor(t *testing.T) { /* Keep or move to services */ }
func TestDetectArcBoxFlavorFallback(t *testing.T) { /* Keep or move to services */ }
```

**Coverage enhancement requirements**:
- Verify 95%+ coverage of `arcbox.go` main file
- Ensure all public functions are tested
- Add edge case testing for command creation
- Verify error propagation works correctly
- Test CLI injection mechanism thoroughly

#### PROMPT END

---

## PHASE 4: Coverage Gap Analysis & Enhancement

### Phase 4.1: Comprehensive Coverage Analysis

#### PROMPT START

Perform detailed coverage analysis and create targeted test enhancements to reach 95-100% coverage:

1. **Generate detailed coverage reports**:
   ```bash
   # Generate coverage for each module
   go test -v -coverprofile=arcbox_coverage.out ./cmd/arcbox/
   go test -v -coverprofile=services_coverage.out ./cmd/arcbox/services/
   go test -v -coverprofile=display_coverage.out ./cmd/arcbox/display/
   go test -v -coverprofile=utils_coverage.out ./cmd/arcbox/utils/
   
   # Generate HTML reports for visual analysis
   go tool cover -html=arcbox_coverage.out -o arcbox_coverage.html
   go tool cover -html=services_coverage.out -o services_coverage.html
   go tool cover -html=display_coverage.out -o display_coverage.html
   go tool cover -html=utils_coverage.out -o utils_coverage.html
   
   # Get line-by-line coverage analysis
   go tool cover -func=arcbox_coverage.out | grep -v "100.0%"
   ```

2. **Identify specific uncovered code paths**:
   - **Error handling branches**: Document each uncovered error path
   - **Edge cases**: Identify boundary conditions not tested
   - **Integration points**: Note untested service interactions
   - **Mock scenarios**: Find missing mock interaction tests
   - **Conditional logic**: Locate untested if/else branches

3. **Create targeted test enhancement plan**:

   **For Main Package (Target: 95%+ from 48.8%)**:
   ```markdown
   ### arcbox.go Coverage Gaps
   
   **Uncovered Lines**: [Specific line numbers from coverage report]
   - Line X: Error handling in NewArcboxCmd
   - Line Y: Edge case in command setup
   - Line Z: CLI injection error path
   
   **Required New Tests**:
   - [ ] TestNewArcboxCmd_ErrorScenarios
   - [ ] TestCommandSetup_EdgeCases  
   - [ ] TestCLIInjection_Failures
   
   **Implementation Priority**: High (core functionality)
   ```

   **For Services Package (Target: 95%+ from 59.0%)**:
   ```markdown
   ### Services Coverage Gaps
   
   **deployment_service.go** (Current: X%):
   - Uncovered: [specific lines]
   - Missing: [specific scenarios]
   - New tests needed: [specific test functions]
   
   **deletion_service.go** (Current: X%):
   - Uncovered: [specific lines] 
   - Missing: [specific scenarios]
   - New tests needed: [specific test functions]
   
   [Continue for all service files...]
   ```

   **For Display Package (Target: 95%+ from 43.3%)**:
   ```markdown
   ### Display Coverage Gaps
   
   **deployment_display.go** (Current: X%):
   - Uncovered: [specific lines]
   - Missing: Complex display scenarios, error formatting
   - New tests needed: [specific test functions]
   
   [Continue for all display files...]
   ```

4. **Create specific test implementations for coverage gaps**:

   **Example: Error Handling Coverage**:
   ```go
   func TestNewArcboxCmd_ErrorScenarios(t *testing.T) {
       tests := []struct {
           name     string
           setup    func() // Setup conditions that trigger errors
           wantErr  bool
           errMsg   string
       }{
           {
               name:    "command_creation_failure",
               setup:   func() { /* Create conditions for command creation failure */ },
               wantErr: true,
               errMsg:  "expected error message",
           },
           // Add all uncovered error scenarios
       }
       // Implement test logic targeting uncovered error paths
   }
   ```

   **Example: Service Integration Coverage**:
   ```go
   func TestDeploymentService_ErrorPathCoverage(t *testing.T) {
       // Target specific uncovered lines in deployment service
       // Test error propagation, retry logic, cleanup scenarios
       // Mock failures at different integration points
   }
   ```

   **Example: Display Logic Coverage**:
   ```go
   func TestDisplayFormatting_EdgeCases(t *testing.T) {
       // Test display logic with edge case data
       // Empty data sets, extremely long strings, special characters
       // Unicode handling, formatting edge cases
   }
   ```

**Coverage Analysis Template**:
```markdown
## Detailed Coverage Analysis Results

### Overall Coverage Summary
- **arcbox.go**: X% → Target: 95%
- **deploy_cmd.go**: X% → Target: 95%  
- **delete_cmd.go**: X% → Target: 95%
- **list_cmd.go**: X% → Target: 95%
- **preflight_cmd.go**: X% → Target: 95%
- **Services average**: X% → Target: 95%
- **Display average**: X% → Target: 95%
- **Utils average**: 96.8% → Maintain 95%+

### Priority Coverage Gaps

#### High Priority (Critical paths)
1. **Error handling in core command creation**
   - File: arcbox.go, Lines: [X-Y]
   - Impact: Command factory reliability
   - Tests needed: [Specific test functions]

2. **Service error propagation**
   - File: [service_file.go], Lines: [X-Y]  
   - Impact: Error user experience
   - Tests needed: [Specific test functions]

#### Medium Priority (Business logic)
[Continue with medium priority gaps...]

#### Low Priority (Edge cases)
[Continue with low priority gaps...]

### Implementation Plan
- **Week 1**: High priority gaps (target 85%+ coverage)
- **Week 2**: Medium priority gaps (target 92%+ coverage)  
- **Week 3**: Low priority gaps (target 95%+ coverage)
- **Week 4**: Final optimization (target 98%+ coverage)
```

#### PROMPT END

### Phase 4.2: Implement Coverage Enhancement Tests

#### PROMPT START

Implement the specific test enhancements needed to achieve 95-100% coverage across all modules:

1. **Implement high-priority coverage tests**:
   
   **Error Handling Coverage**:
   ```go
   // Add to appropriate test files
   func TestErrorHandling_ComprehensiveCoverage(t *testing.T) {
       // Test all uncovered error branches
       // CLI failures, validation errors, service errors
       // Network failures, permission errors, resource conflicts
   }

   func TestEdgeCases_BoundaryConditions(t *testing.T) {
       // Test boundary conditions and edge cases
       // Empty inputs, maximum length inputs, special characters
       // Null values, invalid formats, unexpected data types
   }
   ```

   **Integration Path Coverage**:
   ```go
   func TestServiceIntegration_ErrorPropagation(t *testing.T) {
       // Test error propagation between layers
       // Service to command error handling
       // Command to display error handling
       // Cross-service error scenarios
   }

   func TestMockInteraction_ComprehensiveScenarios(t *testing.T) {
       // Test all mock interaction patterns
       // Verify mock method calls
       // Test mock response handling
       // Mock failure scenarios
   }
   ```

2. **Add missing business logic tests**:
   
   **Complex Business Logic**:
   ```go
   func TestBusinessLogic_ComplexScenarios(t *testing.T) {
       // Test complex business logic branches
       // Multi-step workflows
       // Conditional logic trees
       // State management scenarios
   }

   func TestParameterValidation_ComprehensiveCases(t *testing.T) {
       // Test all parameter validation rules
       // Required vs optional parameters
       // Parameter format validation
       // Cross-parameter dependencies
   }
   ```

3. **Enhance service-specific coverage**:

   **Deployment Service**:
   ```go
   func TestDeploymentService_UncoveredPaths(t *testing.T) {
       // Target specific uncovered lines in deployment_service.go
       // Template resolution edge cases
       // Parameter processing error paths
       // Resource creation failure scenarios
   }
   ```

   **Deletion Service**:
   ```go
   func TestDeletionService_UncoveredPaths(t *testing.T) {
       // Target specific uncovered lines in deletion_service.go
       // Resource existence check failures
       // Confirmation prompt edge cases
       // Cleanup failure scenarios
   }
   ```

   **Listing Service**:
   ```go
   func TestListingService_UncoveredPaths(t *testing.T) {
       // Target specific uncovered lines in listing_service.go
       // Subscription enumeration failures
       // Flavor detection edge cases
       // Output formatting edge cases
   }
   ```

4. **Add display logic coverage**:

   **Display Components**:
   ```go
   func TestDisplayComponents_UncoveredPaths(t *testing.T) {
       // Target uncovered lines in display components
       // Complex formatting scenarios
       // Error display handling
       // Progress indicator edge cases
   }

   func TestOutputFormatting_EdgeCases(t *testing.T) {
       // Test output formatting edge cases
       // Large data sets, empty data sets
       // Special characters, Unicode handling
       // Formatting error scenarios
   }
   ```

5. **Create test data and scenarios for comprehensive coverage**:

   **Test Data Builders**:
   ```go
   // In test_helpers.go or appropriate test files
   func buildComplexTestScenarios() []TestScenario {
       return []TestScenario{
           // Scenarios targeting uncovered code paths
           // Edge cases, error conditions, boundary values
           // Integration scenarios, mock interaction cases
       }
   }

   func createErrorConditionMocks() *azurecli.MockAzureCLI {
       // Mock configurations that trigger error paths
       // CLI failures, permission errors, network issues
       // Resource conflicts, quota exceeded scenarios
   }
   ```

**Implementation Strategy**:
- Focus on one module at a time for systematic coverage improvement
- Run coverage analysis after each enhancement to verify improvement
- Prioritize critical error paths and integration points
- Use parameterized tests for comprehensive scenario coverage
- Verify that new tests actually increase coverage (avoid redundant tests)

**Verification Process**:
```bash
# After implementing each enhancement
go test -v -cover ./cmd/arcbox/[module]
go tool cover -func=[module]_coverage.out | grep -v "100.0%"

# Verify overall improvement
go test -v -cover ./cmd/arcbox/...
```

**Success Criteria**:
- Each module achieves 95%+ coverage
- All critical error paths are tested
- Integration scenarios are comprehensively covered
- Mock interactions are properly verified
- Edge cases and boundary conditions are tested

#### PROMPT END

---

## PHASE 5: Test Quality & Maintainability Enhancement

### Phase 5.1: Test Infrastructure Optimization

#### PROMPT START

Optimize test infrastructure for maintainability, reliability, and developer experience:

1. **Create comprehensive test helpers** (`test_helpers.go`):
   
   ```go
   package arcbox

   import (
       "testing"
       "jumpstartcli/internal/azurecli"
       "github.com/spf13/cobra"
   )

   // TestEnvironment provides consistent test setup
   type TestEnvironment struct {
       MockCLI    *azurecli.MockAzureCLI
       TempDir    string
       Cleanup    func()
   }

   // SetupTestEnvironment creates a clean test environment
   func SetupTestEnvironment(t *testing.T) *TestEnvironment {
       // Create mock CLI with default responses
       // Setup temporary directory for test files  
       // Configure test logging
       // Return cleanup function
   }

   // Mock CLI creation patterns
   func CreateMockCLIWithResponses(responses map[string]string) *azurecli.MockAzureCLI {
       // Standardized mock CLI creation
       // Common response patterns
       // Error scenario configurations
   }

   // Test data builders
   func BuildValidDeploymentConfig() DeploymentConfig {
       // Standard valid deployment configuration
   }

   func BuildInvalidDeploymentConfigs() []InvalidConfigTestCase {
       // Standard invalid configuration test cases
   }

   // Assertion helpers
   func AssertCommandExists(t *testing.T, parent *cobra.Command, cmdName string) {
       // Standard command existence validation
   }

   func AssertFlagConfiguration(t *testing.T, cmd *cobra.Command, flagSpecs []FlagSpec) {
       // Standard flag validation
   }

   func AssertErrorMessage(t *testing.T, err error, expectedSubstring string) {
       // Standard error message validation
   }

   // Mock verification helpers
   func VerifyMockCLICalls(t *testing.T, mockCLI *azurecli.MockAzureCLI, expectedCalls []string) {
       // Verify expected CLI calls were made
   }
   ```

2. **Standardize test patterns across all test files**:

   **Consistent Test Structure**:
   ```go
   func TestComponent_Functionality_Scenario(t *testing.T) {
       // Arrange
       env := SetupTestEnvironment(t)
       defer env.Cleanup()
       
       // Configure test-specific setup
       
       // Act
       result, err := performOperation()
       
       // Assert
       if err != nil {
           t.Errorf("Unexpected error: %v", err)
       }
       
       // Verify results
       // Verify mock interactions
   }
   ```

   **Parameterized Test Patterns**:
   ```go
   func TestComponent_ComprehensiveScenarios(t *testing.T) {
       tests := []struct {
           name     string
           setup    func(*TestEnvironment)
           input    interface{}
           want     interface{}
           wantErr  bool
           errMsg   string
       }{
           // Comprehensive test cases
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               env := SetupTestEnvironment(t)
               defer env.Cleanup()
               
               if tt.setup != nil {
                   tt.setup(env)
               }
               
               // Test implementation
           })
       }
   }
   ```

3. **Implement test reliability improvements**:

   **Deterministic Test Execution**:
   ```go
   // Ensure tests are independent and deterministic
   func TestReliability_IndependentExecution(t *testing.T) {
       // Each test should be able to run independently
       // No shared state between tests
       // Proper cleanup after each test
   }

   // Concurrent execution safety
   func TestReliability_ConcurrentSafety(t *testing.T) {
       // Tests should be safe for concurrent execution
       // No race conditions
       // No shared mutable state
   }
   ```

   **Resource Management**:
   ```go
   // Proper resource cleanup
   func (env *TestEnvironment) Cleanup() {
       // Clean up temporary files
       // Reset mock state
       // Clean up any created resources
   }

   // Memory leak prevention
   func TestReliability_MemoryManagement(t *testing.T) {
       // Ensure tests don't leak memory
       // Proper cleanup of large objects
       // Mock object lifecycle management
   }
   ```

4. **Add performance and benchmarking tests**:

   **Performance Benchmarks**:
   ```go
   func BenchmarkCommandCreation(b *testing.B) {
       for i := 0; i < b.N; i++ {
           cmd := NewArcboxCmd()
           _ = cmd
       }
   }

   func BenchmarkCLIIntegration(b *testing.B) {
       mockCLI := CreateMockCLIWithResponses(standardResponses)
       
       b.ResetTimer()
       for i := 0; i < b.N; i++ {
           cmd := NewArcboxCmdWithCLI(mockCLI)
           _ = cmd
       }
   }
   ```

5. **Create test documentation and guidelines**:

   **Test Guidelines Document**:
   ```markdown
   ## ArcBox Test Guidelines

   ### Test Organization
   - Each command has its own test file: `[command]_cmd_test.go`
   - Core integration tests in `arcbox_test.go`
   - Shared test utilities in `test_helpers.go`

   ### Test Naming Conventions
   - `Test[Component]_[Functionality]_[Scenario]`
   - Use descriptive scenario names
   - Group related tests with common prefixes

   ### Test Structure
   - Follow Arrange-Act-Assert pattern
   - Use `SetupTestEnvironment()` for consistent setup
   - Always call cleanup functions
   - Use parameterized tests for multiple scenarios

   ### Mock Usage
   - Use standardized mock creation patterns
   - Verify mock interactions where appropriate
   - Test both success and failure scenarios
   - Use realistic mock responses

   ### Coverage Requirements
   - Maintain 95%+ coverage for all modules
   - Test all error paths
   - Include edge case testing
   - Verify integration points
   ```

**Implementation Priorities**:
1. **Week 1**: Create `test_helpers.go` with core utilities
2. **Week 2**: Standardize existing test patterns
3. **Week 3**: Add reliability and performance tests
4. **Week 4**: Create documentation and guidelines

#### PROMPT END

### Phase 5.2: Final Validation & Documentation

#### PROMPT START

Perform final validation of the test coverage optimization and create comprehensive documentation:

1. **Comprehensive coverage validation**:
   
   ```bash
   # Generate final coverage reports
   go test -v -coverprofile=final_coverage.out ./cmd/arcbox/...
   go tool cover -html=final_coverage.out -o final_coverage.html
   go tool cover -func=final_coverage.out

   # Verify each module meets 95%+ coverage
   go test -cover ./cmd/arcbox/ | grep "coverage:"
   go test -cover ./cmd/arcbox/services/... | grep "coverage:"
   go test -cover ./cmd/arcbox/display/... | grep "coverage:"
   go test -cover ./cmd/arcbox/utils/... | grep "coverage:"
   ```

2. **Functional regression testing**:
   
   ```go
   // Comprehensive integration tests to ensure no functionality changes
   func TestFunctionalRegression_AllCommands(t *testing.T) {
       tests := []struct {
           name    string
           args    []string
           setup   func(*azurecli.MockAzureCLI)
           verify  func(*testing.T, error, string) // Verify behavior matches original
       }{
           {
               name: "deploy_command_behavior",
               args: []string{"deploy", "--resource-group", "test-rg"},
               setup: func(m *azurecli.MockAzureCLI) {
                   // Setup responses identical to original behavior
               },
               verify: func(t *testing.T, err error, output string) {
                   // Verify behavior matches original implementation
               },
           },
           // Test all command variations and scenarios
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Execute test and verify original behavior is preserved
           })
       }
   }
   ```

3. **Performance regression testing**:
   
   ```go
   func TestPerformanceRegression_CommandCreation(t *testing.T) {
       // Ensure command creation performance hasn't degraded
       // Compare against baseline metrics
   }

   func TestPerformanceRegression_TestExecution(t *testing.T) {
       // Ensure test execution time is reasonable
       // No significant slowdown in test suite
   }
   ```

4. **Create final test documentation**:

   **Test Coverage Report**:
   ```markdown
   # ArcBox Test Coverage Optimization - Final Report

   ## Coverage Achievement Summary
   
   ### Before Optimization
   - **Main package**: 48.8% coverage (1842-line monolithic test file)
   - **Services**: 59.0% coverage
   - **Display**: 43.3% coverage
   - **Utils**: 96.8% coverage
   
   ### After Optimization
   - **arcbox.go**: XX% coverage (focused core tests)
   - **deploy_cmd.go**: XX% coverage (dedicated test file)
   - **delete_cmd.go**: XX% coverage (dedicated test file)
   - **list_cmd.go**: XX% coverage (dedicated test file)
   - **preflight_cmd.go**: XX% coverage (dedicated test file)
   - **Services average**: XX% coverage
   - **Display average**: XX% coverage
   - **Utils average**: XX% coverage
   - **Overall average**: XX% coverage

   ## Test File Organization
   
   ### New Test File Structure
   ```
   cmd/arcbox/
   ├── arcbox_test.go              # Core command tests (XXX lines)
   ├── deploy_cmd_test.go          # Deploy command tests (XXX lines) 
   ├── delete_cmd_test.go          # Delete command tests (XXX lines)
   ├── list_cmd_test.go            # List command tests (XXX lines)
   ├── preflight_cmd_test.go       # Preflight command tests (XXX lines)
   ├── test_helpers.go             # Shared test utilities (XXX lines)
   ├── services/
   │   ├── [service]_test.go       # Enhanced service tests
   ├── display/
   │   ├── [display]_test.go       # Enhanced display tests
   └── utils/
       └── [util]_test.go          # Maintained high coverage
   ```

   ### Test Coverage Improvements
   
   #### Major Coverage Gains
   1. **Error handling paths**: From XX% to XX%
   2. **Edge case scenarios**: From XX% to XX%
   3. **Integration points**: From XX% to XX%
   4. **Mock interactions**: From XX% to XX%

   #### New Test Categories Added
   - [ ] Comprehensive flag validation
   - [ ] Service integration testing
   - [ ] Error scenario coverage
   - [ ] Output format validation
   - [ ] CLI injection testing
   - [ ] Command hierarchy validation
   - [ ] Performance benchmarking

   ## Quality Improvements
   
   ### Test Infrastructure
   - ✅ Standardized test patterns
   - ✅ Shared test utilities
   - ✅ Consistent mock usage
   - ✅ Proper resource cleanup
   - ✅ Deterministic test execution

   ### Maintainability
   - ✅ Focused, single-responsibility test files
   - ✅ Clear test naming conventions
   - ✅ Comprehensive test documentation
   - ✅ Parameterized test patterns
   - ✅ Reusable test components

   ## Functionality Validation
   
   ### Regression Testing Results
   - ✅ All original functionality preserved
   - ✅ No behavior changes detected
   - ✅ All public interfaces unchanged
   - ✅ Performance within acceptable ranges
   - ✅ All existing tests still pass

   ## Developer Experience Improvements
   
   ### Test Execution
   - **Faster test feedback**: Individual command tests can be run separately
   - **Better error isolation**: Focused test files make debugging easier
   - **Clear test organization**: Easy to find relevant tests
   - **Comprehensive coverage**: Confidence in code changes

   ### Test Development
   - **Reusable components**: Test helpers reduce duplication
   - **Standard patterns**: Consistent test structure across files
   - **Clear guidelines**: Test development documentation
   - **Mock infrastructure**: Standardized mock usage patterns
   ```

5. **Create future maintenance guidelines**:

   **Test Maintenance Guide**:
   ```markdown
   ## ArcBox Test Maintenance Guide

   ### Adding New Tests
   1. Determine appropriate test file based on component
   2. Use standard test patterns from `test_helpers.go`
   3. Follow naming conventions: `Test[Component]_[Function]_[Scenario]`
   4. Include both success and error scenarios
   5. Verify coverage improvement with `go test -cover`

   ### Updating Existing Tests
   1. Run affected tests before making changes
   2. Update test data and expectations as needed
   3. Verify all tests still pass
   4. Check that coverage is maintained or improved

   ### Coverage Monitoring
   1. Run coverage analysis regularly: `go test -cover ./cmd/arcbox/...`
   2. Investigate any coverage decreases immediately
   3. Add tests for any new code paths
   4. Maintain 95%+ coverage threshold

   ### Performance Monitoring
   1. Run benchmarks periodically: `go test -bench . ./cmd/arcbox/...`
   2. Watch for performance regressions
   3. Optimize test execution time as needed
   4. Monitor test suite execution time

   ### Quality Assurance
   1. Ensure tests are deterministic and can run independently
   2. Verify proper resource cleanup
   3. Check for test flakiness
   4. Maintain test documentation
   ```

**Final Validation Checklist**:
- [ ] All modules achieve 95%+ coverage
- [ ] All original functionality preserved  
- [ ] No performance regressions
- [ ] Test files are well-organized and maintainable
- [ ] Documentation is comprehensive and up-to-date
- [ ] Future maintenance guidelines are clear

#### PROMPT END

---

## Summary

This GitHub Copilot prompt provides a comprehensive, phase-by-phase approach to optimizing test coverage for the ArcBox CLI by:

1. **Analyzing current test structure** and identifying coverage gaps
2. **Planning test file reorganization** to split the monolithic test file
3. **Creating focused test files** for each command with comprehensive coverage
4. **Enhancing core integration tests** while maintaining functionality
5. **Implementing targeted coverage improvements** to reach 95-100%
6. **Optimizing test infrastructure** for maintainability and reliability
7. **Validating results** and creating comprehensive documentation

The approach emphasizes:
- **Zero functionality changes** - all existing behavior must be preserved
- **Systematic coverage improvement** - targeted testing for uncovered code paths  
- **Test organization** - logical, maintainable test file structure
- **Developer experience** - faster feedback, better debugging, clear guidelines
- **Quality assurance** - reliable, deterministic tests with proper cleanup

Each phase includes specific prompts, code examples, and validation steps to guide the implementation process effectively while maintaining the same high-quality, detailed approach as the original refactoring prompt.
