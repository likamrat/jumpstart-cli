# Go CLI Test Coverage Optimization Prompt

## 🚀 Quick Copy Prompt

**Copy this section and paste into GitHub Copilot:**

---
<!-- START COPY - GitHub Copilot Prompt -->

**GO CLI TEST COVERAGE IMPROVEMENT**

**TARGET**: [Replace with specific function/file: `cmd/arcbox/arcbox.go`, `internal/utils/utils.go`, single test file]

**CONTEXT**: Improve Go CLI test coverage using proven methodology. **WORK IN SMALL INCREMENTS** - focus on 1-2 functions or single test file at a time for best results.

**CRITICAL REQUIREMENT**: Create **SCALABLE & ROBUST** test architecture that supports future commands (e.g., `localbox`, new CLI features). Use consistent patterns, reusable components, and extensible structures.

**98% COVERAGE REQUIREMENTS**:
- **All code paths**: Every branch, condition, and error scenario
- **Edge cases**: Empty inputs, nil values, boundary conditions, malformed data
- **Error handling**: All error types, nested errors, timeout scenarios
- **Concurrency**: Race conditions, goroutine safety (if applicable)
- **Integration points**: Azure CLI failures, network issues, file system errors
- **Performance edge cases**: Large inputs, memory constraints, slow operations

**INCREMENTAL 4-PHASE APPROACH** (do ONE phase per session):

1. **ANALYZE** (5-10 mins): `go test -coverprofile=coverage.out ./[SPECIFIC_PACKAGE]` → identify 2-3 lowest coverage functions
2. **ARCHITECTURE** (15-20 mins): Build reusable test infrastructure for JUST the target functions
3. **TEST** (20-30 mins): Add comprehensive cases for 1-2 functions using scalable patterns
4. **VALIDATE** (5 mins): Check coverage improvement for just the targeted functions

**FOCUS THIS SESSION ON**: [Specify: "CreateCommand function", "flag validation", "error handling", or "single test file"]

**SMALL-SCOPE SCALABLE CLI PATTERN** (Copy & Adapt for 1-2 functions):

```go
// FOCUS: Test ONE function at a time with COMPREHENSIVE 98% coverage pattern
func TestSingleTargetFunction_Comprehensive(t *testing.T) {
    // Setup shared infrastructure (build incrementally)
    suite := NewCLITestSuite() // Create if doesn't exist, reuse if exists
    
    tests := []struct {
        name        string
        mockSetup   func(*MockAzureCLI)
        input       string
        expectError bool
        expectPanic bool
        validate    func(t *testing.T, result interface{})
    }{
        // ===== HAPPY PATH (20% of cases) =====
        {
            name: "success_basic",
            mockSetup: func(mock *MockAzureCLI) {
                mock.SetResponse("success")
            },
            input: "valid-input",
            expectError: false,
        },
        {
            name: "success_with_flags",
            mockSetup: func(mock *MockAzureCLI) {
                mock.SetResponse("success")
            },
            input: "valid-input --flag=value",
            expectError: false,
        },
        
        // ===== ERROR SCENARIOS (60% of cases) =====
        {
            name: "azure_cli_error",
            mockSetup: func(mock *MockAzureCLI) {
                mock.SetError(errors.New("Azure CLI failed"))
            },
            input: "valid-input",
            expectError: true,
        },
        {
            name: "azure_cli_timeout",
            mockSetup: func(mock *MockAzureCLI) {
                mock.SetTimeout()
            },
            input: "valid-input",
            expectError: true,
        },
        {
            name: "azure_cli_non_zero_exit",
            mockSetup: func(mock *MockAzureCLI) {
                mock.SetExitCode(1)
            },
            input: "valid-input",
            expectError: true,
        },
        
        // ===== EDGE CASES (20% of cases) =====
        {
            name: "empty_input",
            input: "",
            expectError: true,
        },
        {
            name: "nil_input_handling",
            input: "null",
            expectError: true,
        },
        {
            name: "malformed_json_response",
            mockSetup: func(mock *MockAzureCLI) {
                mock.SetResponse("invalid-json{")
            },
            input: "valid-input",
            expectError: true,
        },
        {
            name: "extremely_long_input",
            input: strings.Repeat("a", 10000),
            expectError: true,
        },
        {
            name: "special_characters",
            input: "input with spaces & symbols: @#$%^&*()",
            expectError: false, // Should handle gracefully
        },
        {
            name: "unicode_characters",
            input: "unicode-テスト-😀",
            expectError: false,
        },
        
        // ===== BOUNDARY CONDITIONS =====
        {
            name: "max_length_boundary",
            input: strings.Repeat("x", 255), // Assuming 255 is max
            expectError: false,
        },
        {
            name: "over_max_length_boundary",
            input: strings.Repeat("x", 256),
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.expectPanic {
                defer func() {
                    if r := recover(); r == nil {
                        t.Errorf("Expected panic but didn't get one")
                    }
                }()
            }
            
            if tt.mockSetup != nil {
                tt.mockSetup(suite.mockAzureCLI)
            }
            
            result, err := SingleTargetFunction(suite.mockAzureCLI, tt.input)
            
            if (err != nil) != tt.expectError {
                t.Errorf("expectError %v, got error: %v", tt.expectError, err)
            }
            
            if tt.validate != nil {
                tt.validate(t, result)
            }
            
            // Reset mock for next test
            suite.mockAzureCLI.Reset()
        })
    }
}

// 98% COVERAGE HELPER: Test all conditional branches
func TestSingleTargetFunction_AllBranches(t *testing.T) {
    testCases := []struct {
        name      string
        condition string
        setup     func(*MockAzureCLI)
        input     string
        expected  bool
    }{
        // Test every if/else branch in the function
        {"branch_condition_true", "flag_present", setupFlagPresent, "input", true},
        {"branch_condition_false", "flag_absent", setupFlagAbsent, "input", false},
        {"nested_branch_A", "nested_true", setupNestedA, "input", true},
        {"nested_branch_B", "nested_false", setupNestedB, "input", false},
        // Add a case for every conditional path in your function
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            mock := NewMockAzureCLI()
            tc.setup(mock)
            
            result := CheckCondition(mock, tc.input, tc.condition)
            
            if result != tc.expected {
                t.Errorf("expected %v, got %v", tc.expected, result)
            }
        })
    }
}
                mock.SetError(errors.New("simple error"))
            },
            input: "valid-input",
            expectError: true,
        },
        // Add 2-3 more focused test cases
    }
    
    // Test just the target function
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            suite.mockAzureCLI.Reset()
            if tt.mockSetup != nil {
                tt.mockSetup(suite.mockAzureCLI)
            }
            
            result := TargetFunctionOnly(suite.mockAzureCLI, tt.input)
            // Simple, focused validation
        })
    }
}
```

**SMALL INCREMENT SUCCESS TARGETS**:
- **Single session goal**: Improve 1-2 functions by 15-25% coverage (targeting 98% final)
- **Coverage verification**: `go test -coverprofile=coverage.out ./[PACKAGE] && go tool cover -func=coverage.out | grep [FUNCTION]`
- **Quality gate**: Each function must reach 98%+ before moving to next
- **Branch coverage**: Use `go tool cover -html=coverage.out` to verify all branches tested
- **Edge case validation**: Test boundary conditions, error paths, and panic scenarios
- **Build incrementally**: Add to existing test infrastructure, don't rebuild
- **Keep scope narrow**: Test one command flag, one error path, one validation
- **Validate quickly**: Check coverage after each small addition

**PROGRESS TRACKING**: Update this section after each session to maintain momentum and visibility:

```markdown
## Test Optimization Progress Log

### Package: [cmd/arcbox | internal/utils | etc.]
**Overall Goal**: 90%+ coverage | **Current**: __% | **Target Date**: ____

#### Session Log
| Session | Date | Duration | Target Function | Coverage Before | Coverage After | Files Modified | Next Session Focus |
|---------|------|----------|-----------------|-----------------|----------------|----------------|-------------------|
| 1 | 2025-06-12 | 25 min | validateResourceGroup | 45% | 72% | arcbox_test.go | Add error cases |
| 2 | | | | | | | |
| 3 | | | | | | | |

#### Infrastructure Built
- [ ] **CLITestHelper**: Basic shared infrastructure created
- [ ] **MockAzureCLI**: Standard scenarios implemented  
- [ ] **TestDataProvider**: Common test data available
- [ ] **ValidationHelpers**: Output format validation ready
- [ ] **ErrorScenarios**: Standard error patterns established

#### Functions Coverage Progress
| Function Name | Baseline | Current | Target | Status | Notes |
|---------------|----------|---------|--------|--------|-------|
| validateResourceGroup | 45% | 72% | 98% | 🟡 In Progress | Success cases done, need edge cases |
| createCommand | 23% | 23% | 98% | 🔴 Not Started | Next target |
| executeOperation | 67% | 67% | 98% | 🔴 Pending | After createCommand |

#### Session Notes & Learnings
- **Session 1**: Created basic infrastructure, focused on happy path cases
- **Session 2**: [Next session notes]
- **Session 3**: [Next session notes]

#### Patterns Established (for future commands)
- [ ] Command factory pattern implemented
- [ ] Standard test case structure defined
- [ ] Mock scenario library created
- [ ] Output validation helpers ready
- [ ] Error handling patterns documented

---
```

**COPY THIS TEMPLATE** to track your progress session by session.

**SCALABILITY REQUIREMENTS**:
- **Consistent Test Structure**: Same pattern for arcbox, agora, localbox
- **Reusable Mock Infrastructure**: Shared Azure CLI mocks across commands
- **Extensible Validation**: Common validation patterns for all CLI commands
- **Maintainable Patterns**: Easy to add new commands without duplicating code
- **Future-Proof Design**: Support new features, flags, output formats

**FOCUS AREAS**:
- Command validation & flag parsing
- Azure CLI integration errors  
- Output format testing (json/yaml/table)
- Edge cases: empty inputs, invalid flags
- Error message validation

**SUCCESS TARGET**: 98% coverage, all error paths tested, comprehensive edge cases, no panics on invalid input.

<!-- END COPY - GitHub Copilot Prompt -->
---

## Methodology Reference

### Incremental 4-Phase Approach

**Work in SMALL sessions (20-30 minutes each) for best GitHub Copilot results**

1. **Analysis** (1 session): Target 2-3 specific functions with lowest coverage
2. **Architecture** (1-2 sessions): Build minimal infrastructure for just those functions  
3. **Testing** (2-3 sessions): Add tests for 1 function per session
4. **Validation** (quick): Check coverage after each function

### Small-Scope Test Architecture

#### Start Small, Build Incrementally

```go
// SESSION 1: Create minimal shared infrastructure
type SimpleCLITestHelper struct {
    mockAzureCLI *MockAzureCLI
}

func NewSimpleCLITestHelper() *SimpleCLITestHelper {
    return &SimpleCLITestHelper{
        mockAzureCLI: NewMockAzureCLI(),
    }
}

// SESSION 2: Test ONE function with basic pattern
func TestSingleFunction(t *testing.T) {
    helper := NewSimpleCLITestHelper()
    
    // Focus on 3-4 core test cases only
    tests := []struct {
        name      string
        mockSetup func(*MockAzureCLI)
        input     string
        wantError bool
    }{
        {"success", func(m *MockAzureCLI) { m.SetResponse("ok") }, "valid", false},
        {"error", func(m *MockAzureCLI) { m.SetError(errors.New("fail")) }, "valid", true},
        {"empty_input", nil, "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.mockSetup != nil {
                tt.mockSetup(helper.mockAzureCLI)
            }
            
            _, err := SingleTargetFunction(helper.mockAzureCLI, tt.input)
            if (err != nil) != tt.wantError {
                t.Errorf("wantError %v, got %v", tt.wantError, err)
            }
        })
    }
}
```

#### Incremental Enhancement Pattern

```go
// SESSION 3: Extend helper for additional function
func (h *SimpleCLITestHelper) TestAnotherFunction(t *testing.T, input string) {
    // Reuse existing mock, add minimal new functionality
    result, err := AnotherTargetFunction(h.mockAzureCLI, input)
    return result, err
}

// SESSION 4: Add specific validation for command testing
func (h *SimpleCLITestHelper) ValidateCommandOutput(t *testing.T, output string, format string) {
    switch format {
    case "json":
        assert.True(t, json.Valid([]byte(output)))
    case "table":
        assert.Contains(t, output, "|")
    // Add other formats as needed
    }
}
```

### Scalable Test Architecture Patterns

#### Reusable Test Infrastructure

```go
// SHARED TEST SUITE for all CLI commands
type CLITestSuite struct {
    MockAzureCLI    *MockAzureCLI
    TestData        *TestDataProvider
    Validator       *CommandValidator
    OutputCapture   *OutputCapture
}

func NewCLITestSuite() *CLITestSuite {
    return &CLITestSuite{
        MockAzureCLI:  NewMockAzureCLI(),
        TestData:      NewTestDataProvider(),
        Validator:     NewCommandValidator(),
        OutputCapture: NewOutputCapture(),
    }
}

// EXTENSIBLE COMMAND FACTORY for future commands
type CommandFactory interface {
    CreateCommand() *cobra.Command
    GetCommandName() string
    GetRequiredFlags() []string
    GetSupportedOutputFormats() []string
}

// CONSISTENT TEST RUNNER for all commands
func RunStandardCommandTests(t *testing.T, factory CommandFactory, tests []CommandTestCase) {
    suite := NewCLITestSuite()
    
    for _, tt := range tests {
        t.Run(tt.Name, func(t *testing.T) {
            // Setup
            cmd := factory.CreateCommand()
            if tt.MockSetup != nil {
                tt.MockSetup(suite.MockAzureCLI)
            }
            
            // Execute with consistent pattern
            output, err := suite.ExecuteCommand(cmd, tt.Args)
            
            // Validate with extensible validation
            suite.ValidateResult(t, tt, output, err)
        })
    }
}
```

#### Extensible Mock Infrastructure

```go
// SHARED MOCK INTERFACE for all commands
type MockAzureCLI struct {
    responses    map[string]string
    errors       map[string]error
    callHistory  []MockCall
    scenarios    map[string]MockScenario
}

// FUTURE-PROOF MOCK SCENARIOS
func (m *MockAzureCLI) SetScenario(scenario string) {
    switch scenario {
    case "arcbox_success":
        m.SetArcBoxSuccessResponses()
    case "agora_success":
        m.SetAgoraSuccessResponses()
    case "localbox_success":  // Ready for future localbox command
        m.SetLocalBoxSuccessResponses()
    case "network_failure":
        m.SetNetworkFailureResponses()
    case "auth_failure":
        m.SetAuthFailureResponses()
    }
}

// CONSISTENT ERROR PATTERNS for all commands
func (m *MockAzureCLI) SetStandardErrorScenarios() {
    m.errors["subscription_not_found"] = errors.New("subscription not found")
    m.errors["resource_group_not_found"] = errors.New("resource group not found") 
    m.errors["permission_denied"] = errors.New("insufficient permissions")
    m.errors["network_timeout"] = errors.New("network timeout")
    m.errors["quota_exceeded"] = errors.New("quota exceeded")
}
```

#### Standard Test Data Patterns

```go
// REUSABLE TEST DATA for consistency across commands
type TestDataProvider struct {
    ResourceGroups  []string
    Locations       []string
    Subscriptions   []string
    InvalidInputs   map[string]interface{}
    ValidInputs     map[string]interface{}
}

func NewTestDataProvider() *TestDataProvider {
    return &TestDataProvider{
        ResourceGroups: []string{"test-rg", "prod-rg", "dev-rg"},
        Locations:      []string{"eastus", "westus2", "westeurope"},
        Subscriptions:  []string{"test-sub-id", "prod-sub-id"},
        InvalidInputs: map[string]interface{}{
            "empty_string":    "",
            "nil_value":       nil,
            "invalid_region":  "invalid-region",
            "malformed_json":  "{invalid json}",
        },
        ValidInputs: map[string]interface{}{
            "standard_rg":     "test-rg",
            "standard_location": "eastus",
            "standard_sub":    "test-sub-id",
        },
    }
}

// EXTENSIBLE VALIDATION PATTERNS
func (tdp *TestDataProvider) GetTestCasesFor(commandType string) []CommandTestCase {
    base := tdp.getBaseTestCases()
    
    switch commandType {
    case "arcbox":
        return append(base, tdp.getArcBoxSpecificCases()...)
    case "agora":
        return append(base, tdp.getAgoraSpecificCases()...)
    case "localbox":  // Ready for future expansion
        return append(base, tdp.getLocalBoxSpecificCases()...)
    default:
        return base
    }
}
```

### Essential Test Templates

#### Core Function Testing
```go
func TestMainFunction(t *testing.T) {
    tests := []struct {
        name        string
        mockSetup   func(*MockAzureCLI)
        input       string
        expectError bool
        expected    interface{}
    }{
        {
            name: "success_case",
            mockSetup: func(mock *MockAzureCLI) {
                mock.SetResponse("success")
            },
            input: "valid-input",
            expectError: false,
            expected: ExpectedResult{},
        },
        {
            name: "azure_cli_error",
            mockSetup: func(mock *MockAzureCLI) {
                mock.SetError(errors.New("az cli failed"))
            },
            input: "valid-input",
            expectError: true,
        },
        {
            name: "empty_input",
            input: "",
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := NewMockAzureCLI()
            if tt.mockSetup != nil {
                tt.mockSetup(mock)
            }
            
            result, err := FunctionUnderTest(mock, tt.input)
            
            if tt.expectError != (err != nil) {
                t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
            }
            if !tt.expectError && result != tt.expected {
                t.Errorf("Expected: %v, got: %v", tt.expected, result)
            }
        })
    }
}
```

#### CLI Command Testing
```go
func TestCommandExecution(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        mockSetup   func(*MockAzureCLI)
        expectError bool
        validate    func(t *testing.T, output string)
    }{
        {
            name: "valid_command",
            args: []string{"--resource-group", "test-rg", "--location", "eastus"},
            mockSetup: func(mock *MockAzureCLI) {
                mock.SetResponse("operation successful")
            },
            expectError: false,
            validate: func(t *testing.T, output string) {
                assert.Contains(t, output, "successful")
            },
        },
        {
            name: "missing_required_flag",
            args: []string{"--location", "eastus"},
            expectError: true,
        },
        {
            name: "azure_service_error",
            args: []string{"--resource-group", "test-rg", "--location", "eastus"},
            mockSetup: func(mock *MockAzureCLI) {
                mock.SetError(errors.New("subscription not found"))
            },
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := CreateTestCommand()
            if tt.mockSetup != nil {
                mock := NewMockAzureCLI()
                tt.mockSetup(mock)
                cmd.SetAzureCLI(mock)
            }
            
            output, err := ExecuteCommand(cmd, tt.args)
            
            if tt.expectError != (err != nil) {
                t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
            }
            if tt.validate != nil {
                tt.validate(t, output)
            }
        })
    }
}
```

### CLI-Specific Test Patterns

#### Flag Validation Testing
```go
func TestCommandFlags(t *testing.T) {
    cmd := CreateCommand()
    
    // Test required flags
    err := cmd.Execute()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "required flag")
    
    // Test flag combinations
    cmd.SetArgs([]string{"--resource-group", "test", "--location", "eastus"})
    assert.NoError(t, cmd.Execute())
    
    // Test invalid flag values
    cmd.SetArgs([]string{"--location", "invalid-region"})
    assert.Error(t, cmd.Execute())
}
```

#### Output Format Testing
```go
func TestOutputFormats(t *testing.T) {
    formats := []string{"json", "yaml", "table", "tsv"}
    
    for _, format := range formats {
        t.Run(format, func(t *testing.T) {
            cmd := CreateCommand()
            cmd.SetArgs([]string{"--output", format, "--resource-group", "test"})
            
            output, err := ExecuteCommand(cmd)
            assert.NoError(t, err)
            
            switch format {
            case "json":
                assert.True(t, json.Valid([]byte(output)))
            case "yaml":
                var data interface{}
                assert.NoError(t, yaml.Unmarshal([]byte(output), &data))
            case "table":
                assert.Contains(t, output, "|")
            case "tsv":
                assert.Contains(t, output, "\t")
            }
        })
    }
}
```

### Incremental Quality Gates

**Per Session Goals (20-30 minutes):**

- [ ] **Target 1-2 functions**: Focus on specific functions, not entire packages
- [ ] **Coverage increment**: Improve targeted functions by 20-30% coverage
- [ ] **Small scope**: Add 3-5 test cases maximum per session
- [ ] **Build incrementally**: Reuse existing test infrastructure when possible
- [ ] **Quick validation**: Run coverage on just the changed functions
- [ ] **One concept**: Focus on either success cases OR error cases OR edge cases per session
- [ ] **Clean workspace**: No temporary files left behind (coverage.out, *.tmp, *.log)

**Session Completion Checklist (98% Coverage Standards):**

- [ ] **Coverage verification**: Run `go test -coverprofile=coverage.out ./[PACKAGE] && go tool cover -func=coverage.out`
- [ ] **Target achieved**: Each worked function shows 98%+ coverage
- [ ] **Visual inspection**: `go tool cover -html=coverage.out` shows no red/uncovered lines
- [ ] **All branches tested**: Every if/else, switch case, and error path covered
- [ ] **Edge cases included**: Boundary conditions, nil inputs, malformed data
- [ ] **Error scenarios comprehensive**: All error types, timeouts, failures
- [ ] **Panic recovery tested**: Functions that might panic have recovery tests
- [ ] **Performance edge cases**: Large inputs, memory constraints covered
- [ ] **Integration failures**: Azure CLI errors, network issues, file system problems
- [ ] **Tests pass**: `go test ./[target_package]` succeeds with no failures
- [ ] **No race conditions**: `go test -race ./[PACKAGE]` passes (if applicable)
- [ ] **Clean workspace**: `git status` shows only intended changes
- [ ] **Progress logged**: Session results recorded in tracking table
- [ ] **Next session planned**: Clear focus for next increment

**98% Coverage Quality Gates:**

- **Function coverage**: ≥98% for each individual function
- **Branch coverage**: Every conditional path executed  
- **Error coverage**: All error types and scenarios tested
- **Edge case coverage**: Boundary conditions and malformed inputs
- **Integration coverage**: External dependency failures covered

**Session-by-Session Building:**

- **Session 1**: Basic test infrastructure + 1 function success case
- **Session 2**: Add error cases for same function
- **Session 3**: Move to next function, reuse infrastructure
- **Session 4**: Add edge cases or output format testing
- **Session 5**: Refactor common patterns into shared helpers

**Small Success Metrics (per session - 98% Target):**

- **Single function coverage**: Target function goes from <70% to 98%+
- **Comprehensive validation**: Test all branches, errors, and edge cases
- **Quality-focused building**: Each session achieves near-perfect coverage for targeted functions
- **Visual verification**: HTML coverage report shows no red uncovered lines
- **Incremental excellence**: Build on previous work to maintain 98% standard across all functions

### Small-Scope Examples

**Instead of**: "Test the entire arcbox command"
**Do**: "Test the validateResourceGroup function"

**Instead of**: "Add comprehensive error handling"  
**Do**: "Add Azure CLI error handling for createCommand function"

**Instead of**: "Implement full output format testing"
**Do**: "Add JSON output validation for status command"

**Session Planning Template:**

```text
Session N Goal: Test [specific function] [specific aspect]
Duration: 20-30 minutes  
Files to modify: [single test file]
Coverage target: [current]% → [target]% for [function name]
Focus: [success cases|error cases|edge cases|output format]
Dependencies: [existing helpers to reuse]
```

**After Each Session - Quick Update Checklist:**

1. Update the Session Log table with results
2. Update function coverage progress
3. Add session notes (what worked, what didn't)
4. Set next session focus
5. Run quick coverage check: `go test -coverprofile=coverage.out ./[package] && go tool cover -func=coverage.out | grep [function_name]`
6. **Clean up temporary files**: `rm -f coverage.out coverage.html *.tmp *.log`

**Session Cleanup Commands:**
```bash
# Clean test artifacts
make clean  # or: rm -f coverage.out coverage.html
go clean -testcache

# Remove any temporary files created during testing
find . -name "*.tmp" -delete
find . -name "*.log" -delete
find . -name ".DS_Store" -delete

# Verify no junk files remain
git status --porcelain | grep -E "\.(tmp|log|out|html)$" || echo "✅ No junk files found"
```

**Files to Keep vs Clean:**

**✅ Keep (commit these):**
- `*_test.go` - Your new test files
- Updated existing test files
- Progress log updates in documentation
- Any new test helpers or infrastructure

**🗑️ Clean (don't commit):**
- `coverage.out` - Coverage profile data
- `coverage.html` - HTML coverage reports  
- `*.tmp` - Temporary files
- `*.log` - Log files from test runs
- `.vscode/settings.json` changes (unless intentional)
- Any backup files created by editors

**Git Status Check:**
Before ending each session, run:
```bash
git status
# Should only show intended test file changes, no temporary artifacts
```

**Weekly Review Questions:**
- Which patterns are working best?
- What infrastructure can be reused for other commands?
- Are we building toward scalable architecture?
- What should be documented for future developers?

This incremental approach ensures GitHub Copilot can provide focused, high-quality suggestions without being overwhelmed by scope, while still building toward the same scalable, robust test architecture.

### Future Command Integration Guide

**When adding new commands (e.g., localbox):**

1. **Import shared test infrastructure**: Use `CLITestSuite` and `CommandFactory`
2. **Extend mock scenarios**: Add command-specific responses to shared mocks
3. **Follow naming conventions**: Use consistent test function and file naming
4. **Reuse validation patterns**: Leverage existing flag and output format testing
5. **Add to test data provider**: Extend `TestDataProvider` with command-specific cases
6. **Maintain coverage standards**: Achieve same 90%+ coverage with shared patterns

**Example for future localbox command:**

```go
// cmd/localbox/localbox_test.go
func TestLocalBoxCommand(t *testing.T) {
    factory := &LocalBoxCommandFactory{}  // Implements CommandFactory
    testCases := NewTestDataProvider().GetTestCasesFor("localbox")
    
    // Uses same infrastructure as arcbox and agora
    RunStandardCommandTests(t, factory, testCases)
}

// Minimal command-specific code needed
type LocalBoxCommandFactory struct{}
func (f *LocalBoxCommandFactory) CreateCommand() *cobra.Command { /* */ }
func (f *LocalBoxCommandFactory) GetCommandName() string { return "localbox" }
// ... implement interface
```

This architecture ensures that adding localbox or any future CLI command requires minimal test code duplication while maintaining comprehensive coverage and consistent quality.

### Proven Success Examples

**ArcBox Package (122 tests, comprehensive coverage):**

- `status_test.go`: Command validation, flags, error scenarios, output formats
- `quota_test.go`: 94.3% coverage with systematic error condition testing  
- `rp_test.go`: Resource provider integration with full mock scenarios

**Key Success Patterns:**

- Modular test functions per component
- Comprehensive Azure CLI mock integration
- Table-driven tests with realistic scenarios
- Performance benchmarks for critical operations
- Clear visual test output with pass/fail indicators

Apply this methodology systematically to achieve 98%+ coverage with robust CLI testing.

---

## 🎯 Strategic Starting Point Reference

### **RECOMMENDED FIRST TARGET: `cmd/arcbox/arcbox.go`**

**Current State Analysis (as of baseline assessment):**
- **Coverage**: 17.7% (massive improvement opportunity)
- **Priority**: Highest - core user-facing CLI functionality
- **Test Infrastructure**: Existing (`arcbox_test.go` with 122 tests)
- **Functions with 0% Coverage**: 15+ functions ready for immediate wins

### **Specific Function Priority Queue:**

#### **Session 1-2: `runArcBoxList` Function**
```go
// File: cmd/arcbox/arcbox.go:1058 (0% coverage → target 98%+)
func runArcBoxList(allSubscriptions, currentSubscription bool, subscriptionID, outputFormat string) error
```

**Why Start Here:**
- ✅ Zero coverage baseline - complete opportunity
- ✅ Clear business logic with multiple code paths
- ✅ Well-defined error scenarios (auth, subscription validation)
- ✅ Output format branching (table/JSON) - good for branch coverage
- ✅ Limited external dependencies - easier to mock

**98% Coverage Test Plan:**
```go
func TestRunArcBoxList_Comprehensive(t *testing.T) {
    tests := []struct {
        name            string
        allSubs         bool
        currentSub      bool  
        subscriptionID  string
        outputFormat    string
        mockSetup       func(*MockAzureCLI)
        expectError     bool
        expectOutput    string
    }{
        // HAPPY PATH (20% of tests)
        {"all_subscriptions_table", true, false, "", "table", setupValidSubs, false, ""},
        {"current_subscription_json", false, true, "", "json", setupCurrentSub, false, ""},
        {"specific_subscription_table", false, false, "valid-sub-id", "table", setupSpecificSub, false, ""},
        
        // ERROR SCENARIOS (60% of tests) 
        {"azure_cli_not_authenticated", true, false, "", "table", setupNoAuth, true, "not logged in"},
        {"invalid_subscription_id", false, false, "invalid-sub", "table", setupInvalidSub, true, "Cannot access subscription"},
        {"azure_cli_timeout", true, false, "", "table", setupTimeout, true, "timeout"},
        {"malformed_subscription_response", true, false, "", "json", setupMalformedJSON, true, ""},
        {"network_error", false, true, "", "table", setupNetworkError, true, "network"},
        {"empty_subscription_list", true, false, "", "table", setupEmptySubs, false, "No ArcBox deployments"},
        {"subscription_access_denied", false, false, "denied-sub", "table", setupAccessDenied, true, "access denied"},
        
        // EDGE CASES (20% of tests)
        {"contradictory_flags_all_and_current", true, true, "", "table", nil, true, "Cannot use multiple"},
        {"contradictory_flags_all_and_specific", true, false, "sub-id", "table", nil, true, "Cannot use multiple"},
        {"empty_deployments_found", false, true, "", "json", setupNoDeployments, false, "No ArcBox deployments"},
        {"very_large_deployment_list", true, false, "", "table", setupLargeDeploymentList, false, ""},
        {"special_chars_subscription_name", false, false, "sub-with-chars-@#$", "json", setupSpecialCharsSub, false, ""},
        {"invalid_output_format", false, true, "", "xml", setupCurrentSub, true, "Invalid output format"},
        {"nil_subscription_response", true, false, "", "table", setupNilResponse, true, ""},
        {"unicode_subscription_name", false, false, "测试-subscription", "table", setupUnicodeSub, false, ""},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockCLI := NewMockAzureCLI()
            if tt.mockSetup != nil {
                tt.mockSetup(mockCLI)
            }
            
            err := runArcBoxList(tt.allSubs, tt.currentSub, tt.subscriptionID, tt.outputFormat)
            
            if (err != nil) != tt.expectError {
                t.Errorf("expectError %v, got error: %v", tt.expectError, err)
            }
            
            if tt.expectOutput != "" && err != nil {
                if !strings.Contains(err.Error(), tt.expectOutput) {
                    t.Errorf("expected error to contain %q, got %q", tt.expectOutput, err.Error())
                }
            }
            
            mockCLI.Reset()
        })
    }
}
```

#### **Sessions 3-4: Zero Coverage Utility Functions**
```go
// All at 0% coverage - quick wins for 98% target
func getAllSubscriptions(azCLI azurecli.AzureCLI) ([]AzureSubscription, error)  // Line 1122
func getCurrentSubscription(azCLI azurecli.AzureCLI) (AzureSubscription, error) // Line 1140  
func getSubscription(azCLI azurecli.AzureCLI, subscriptionID string) (AzureSubscription, error) // Line 1153
func normalizeFlavorCase(flavor string) string // Line 1738
func normalizeSqlServerEditionCase(edition string) string // Line 1754
func normalizeBastionSkuCase(sku string) string // Line 1768
```

#### **Sessions 5-8: Complex Functions (Require More Comprehensive Testing)**
```go
func runQuotaChecksWithOutput(cli azurecli.AzureCLI, cmd *cobra.Command, location, flavor string) (bool, []map[string]interface{}) // Needs branch coverage
func deployArcboxWithParamFile(cmd *cobra.Command, args []string, ...) // Complex deployment logic
func outputArcBoxDeploymentsTable(deployments []ArcBoxDeployment) error // Output formatting
func outputArcBoxDeploymentsJSON(deployments []ArcBoxDeployment) error // JSON serialization
```

### **Coverage Impact Projection:**

| Session | Target Function(s) | Estimated Coverage Gain | Cumulative Coverage |
|---------|-------------------|------------------------|-------------------|
| 1-2 | `runArcBoxList` + utilities | +25% | ~43% |
| 3-4 | Normalization functions | +15% | ~58% |
| 5-6 | `runQuotaChecksWithOutput` | +20% | ~78% |
| 7-8 | Output functions | +15% | ~93% |
| 9-10 | `deployArcboxWithParamFile` | +10% | **98%+** |

### **Why NOT Start Elsewhere:**

❌ **`main.go`**: Already 94.8% coverage  
❌ **`cmd/subscription`**: Already 95.9% coverage  
❌ **`cmd/agora`, `cmd/localbox`**: Lower impact, smaller codebase  
❌ **Internal packages**: Focus on user-facing commands first for maximum ROI

### **Success Validation Commands:**

```bash
# Check function-specific coverage
go test -coverprofile=coverage.out ./cmd/arcbox && go tool cover -func=coverage.out | grep "runArcBoxList"

# Visual verification (should show no red lines)
go tool cover -html=coverage.out -o coverage.html && open coverage.html

# Overall package coverage target
go tool cover -func=coverage.out | grep "cmd/arcbox" | tail -1
# Target: "total: (statements) 98.0%"
```
