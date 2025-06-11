# GitHub Copilot Prompt: Comprehensive Test Coverage Improvement for Go CLI Commands

## Quick Copy Prompt

**Copy this entire section and paste into GitHub Copilot:**

---

**COMPREHENSIVE TEST COVERAGE IMPROVEMENT PROMPT**

**TARGET PACKAGE OR COMMAND**: [REPLACE WITH PACKAGE PATH OR COMMAND - e.g., `cmd/arcbox`, `internal/resourceproviders`, `cmd/agora`]

**STEP 1 - ASSESSMENT**: First, please analyze the current test coverage:
1. Run `go test -coverprofile=coverage.out ./...` and `go tool cover -func=coverage.out`
2. Identify all functions with <90% coverage
3. Review existing test structure and mocking patterns
4. Check for untested error conditions and edge cases

**STEP 2 - SYSTEMATIC IMPROVEMENT**: Then improve test coverage for this Go CLI package using the proven 4-phase methodology that achieved 94.3% coverage in the quota package:

**PHASE 1 - ANALYSIS**: Analyze current test coverage using `go test -coverprofile=coverage.out ./...` and `go tool cover -func=coverage.out`. Identify all functions with <95% coverage.

**PHASE 2 - INFRASTRUCTURE**: Enhance mocks to support all error conditions and edge cases. Create comprehensive table-driven tests with success/failure scenarios.

**PHASE 3 - COMPREHENSIVE TESTING**: For each uncovered function, add tests for:
- **Success scenarios**: Valid inputs with expected outputs
- **Error conditions**: Invalid inputs, external service failures, edge cases
- **Edge cases**: Empty inputs, nil values, boundary conditions
- **Integration scenarios**: Function interactions and dependency chains

**PHASE 4 - VALIDATION**: Run coverage analysis, validate test quality, and ensure real-world error conditions are properly tested.

**Test Pattern Template**:
```go
tests := []struct {
    name        string
    mockSetup   func(*MockInterface)
    expectError bool
    expected    interface{}
    validate    func(t *testing.T, result interface{})
}{
    {
        name: "success case",
        mockSetup: func(mock *MockInterface) {
            mock.SetReturnValue(expectedValue)
        },
        expectError: false,
        expected: expectedResult,
    },
    {
        name: "error case",
        mockSetup: func(mock *MockInterface) {
            mock.SetError(errors.New("test error"))
        },
        expectError: true,
    },
}
```

**Target**: Achieve 90%+ test coverage with robust error handling and comprehensive scenario testing.

Apply this methodology systematically to identify and test all uncovered code paths, especially error conditions and edge cases.

**USAGE EXAMPLES:**
- For arcbox command: Replace `[REPLACE WITH PACKAGE PATH]` with `cmd/arcbox`
- For resource providers: Replace `[REPLACE WITH PACKAGE PATH]` with `internal/resourceproviders`
- For agora command: Replace `[REPLACE WITH PACKAGE PATH]` with `cmd/agora`

**REFERENCE**: See the detailed documentation, templates, and examples in `GITHUB_COPILOT_TEST_COVERAGE_PROMPT.md` below this quick copy section for comprehensive guidance on the 4-phase methodology, test patterns, and quality gates.

---

## Context
You are helping improve test coverage for a Go CLI application following the systematic methodology successfully used to increase quota test coverage from 51.0% to 94.3%. Apply this proven approach to achieve excellent test coverage (90%+) for other command packages.

## Methodology Overview
This prompt follows a 4-phase approach that achieved 94.3% coverage and fixed critical production bugs:

### Phase 1: Coverage Analysis & Function Mapping
1. **Measure baseline coverage**: Run `go test -coverprofile=coverage.out ./...` 
2. **Analyze function-level coverage**: Use `go tool cover -func=coverage.out`
3. **Identify target functions**: List all functions with <90% coverage
4. **Map dependencies**: Identify external dependencies (Azure CLI, HTTP clients, file systems)

### Phase 2: Test Infrastructure & Mock Enhancement  
1. **Enhance mock capabilities**: Ensure mocks support all error conditions and edge cases
2. **Create comprehensive test tables**: Use table-driven tests with these fields:
   ```go
   tests := []struct {
       name        string
       // Input parameters
       mockSetup   func(*MockInterface)  // Setup mocks with specific behaviors
       expectError bool
       expected    interface{}  // Expected results
       validate    func(t *testing.T, result interface{})  // Custom validation
   }{
       // Test cases here
   }
   ```
3. **Mock setup patterns**: Create realistic mock data that mirrors actual service responses

### Phase 3: Systematic Test Case Development
Develop test cases in this priority order:

#### A. Happy Path Coverage (Target: Hit main execution flows)
- Valid inputs with successful operations
- Different input combinations and formats
- Multiple configuration scenarios

#### B. Edge Case Coverage (Target: Boundary conditions)
- Empty/nil inputs, zero values, boundary values
- Missing required fields or data
- Malformed inputs and invalid formats
- Unknown/unsupported options

#### C. Error Path Coverage (Target: All error conditions)
- External service failures (network, API errors)
- Invalid credentials or permissions
- Timeout scenarios and service unavailability
- Fallback mechanism testing

#### D. Integration Coverage (Target: Component interactions)
- Multiple function calls in sequence
- State changes between operations
- Cross-package dependencies

### Phase 4: Bug Identification & Quality Assurance
1. **Look for panic conditions**: Test with empty strings, nil pointers, out-of-bounds access
2. **Test defensive programming**: Verify all error handling paths work correctly
3. **Validate assumptions**: Test what happens when external dependencies behave unexpectedly
4. **Run with race detector**: `go test -race` to find concurrency issues

## Required Test Function Templates

### Template 1: Core Function Testing
```go
func TestMainFunction(t *testing.T) {
    tests := []struct {
        name        string
        input1      string
        input2      int
        mockSetup   func(*MockService)
        expectError bool
        expected    ExpectedResult
    }{
        {
            name:   "successful operation",
            input1: "valid-input",
            input2: 42,
            mockSetup: func(mock *MockService) {
                mock.SetResponse("expected-response")
            },
            expectError: false,
            expected:    ExpectedResult{...},
        },
        {
            name:   "service error",
            input1: "valid-input", 
            input2: 42,
            mockSetup: func(mock *MockService) {
                mock.SetError(fmt.Errorf("service unavailable"))
            },
            expectError: true,
        },
        // Add more test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := NewMockService()
            tt.mockSetup(mock)
            
            result, err := MainFunction(mock, tt.input1, tt.input2)
            
            if tt.expectError && err == nil {
                t.Errorf("Expected error, got nil")
            }
            if !tt.expectError && err != nil {
                t.Errorf("Expected no error, got %v", err)
            }
            if !tt.expectError {
                // Validate result matches expected
                validateResult(t, result, tt.expected)
            }
        })
    }
}
```

### Template 2: Edge Cases Testing
```go
func TestEdgeCases(t *testing.T) {
    tests := []struct {
        name      string
        setup     func(*MockService)
        operation func() interface{}
        validate  func(t *testing.T, result interface{})
    }{
        {
            name: "empty input handling",
            setup: func(mock *MockService) {
                // Setup for empty input scenario
            },
            operation: func() interface{} {
                return FunctionUnderTest("")
            },
            validate: func(t *testing.T, result interface{}) {
                // Verify graceful handling of empty input
            },
        },
        {
            name: "nil pointer handling",
            setup: func(mock *MockService) {
                // Setup for nil pointer scenario
            },
            operation: func() interface{} {
                return FunctionUnderTest(nil)
            },
            validate: func(t *testing.T, result interface{}) {
                // Verify no panic, proper error handling
            },
        },
        // Add more edge cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := NewMockService()
            tt.setup(mock)
            
            result := tt.operation()
            tt.validate(t, result)
        })
    }
}
```

### Template 3: Helper Functions Comprehensive Testing
```go
func TestHelperFunctionsComprehensive(t *testing.T) {
    // Test helper functions with all possible inputs including edge cases
    tests := []struct {
        name     string
        input    interface{}
        expected interface{}
    }{
        {"valid input", "valid", "expected_output"},
        {"empty input", "", "default_or_error"},
        {"unknown input", "unknown", "fallback_value"},
        {"nil input", nil, "nil_handling"},
        // Add all possible input variations
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := HelperFunction(tt.input)
            if result != tt.expected {
                t.Errorf("Expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

### Template 4: Error Condition Testing
```go
func TestErrorConditions(t *testing.T) {
    tests := []struct {
        name          string
        mockSetup     func(*MockService)
        expectedError string
    }{
        {
            name: "network timeout",
            mockSetup: func(mock *MockService) {
                mock.SetTimeout()
            },
            expectedError: "timeout",
        },
        {
            name: "invalid credentials",
            mockSetup: func(mock *MockService) {
                mock.SetAuthError()
            },
            expectedError: "authentication failed",
        },
        // Add all error scenarios
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := NewMockService()
            tt.mockSetup(mock)
            
            _, err := FunctionUnderTest(mock)
            
            if err == nil {
                t.Errorf("Expected error, got nil")
            }
            if !strings.Contains(err.Error(), tt.expectedError) {
                t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
            }
        })
    }
}
```

## Critical Success Patterns

### 1. Mock Setup Best Practices
- **Realistic data**: Use actual service response formats
- **Error injection**: Test all external dependency failures
- **State simulation**: Mock services should maintain state between calls
- **Edge case data**: Include malformed, missing, and boundary data

### 2. Test Validation Strategies
- **Comprehensive assertions**: Validate all important fields, not just success/failure
- **Error message validation**: Ensure error messages are meaningful and consistent
- **State verification**: Check that operations leave system in expected state
- **Performance checks**: Verify operations complete within reasonable time

### 3. Coverage Improvement Tactics
- **Uncovered line analysis**: Use `go tool cover -html=coverage.out` to see exact uncovered lines
- **Defensive code testing**: Focus on error handling and boundary checks
- **Fallback mechanism testing**: Ensure backup logic works when primary fails
- **Configuration variations**: Test different settings and environment conditions

## Quality Gates

### Before Completion, Ensure:
1. **Coverage target met**: Achieve 90%+ overall package coverage
2. **All functions tested**: Every public function has dedicated test cases
3. **No panics**: All edge cases handle gracefully without runtime panics
4. **Error paths covered**: All error conditions properly tested
5. **Mock completeness**: All external dependencies fully mocked with error scenarios
6. **Documentation updated**: Add package to TESTING_ASSESSMENT.md with achievement details

### Success Metrics:
- **Function-level coverage**: Individual functions at 90%+ coverage
- **Edge case resilience**: No crashes on invalid/malformed input
- **Error handling quality**: Meaningful error messages for all failure modes
- **Test maintainability**: Clear test names and comprehensive setup/validation

## Example Application

When I say "Apply this methodology to improve test coverage for the `cmd/arcbox` package", you should:

1. **Analyze current coverage** of cmd/arcbox package
2. **Identify functions** with low coverage, especially main command functions
3. **Enhance mocks** to support arcbox-specific external dependencies
4. **Create comprehensive test tables** using the templates above
5. **Implement systematic test cases** following the 4-phase approach
6. **Validate improvements** and update documentation

Apply this methodology systematically and you'll achieve the same 90%+ coverage improvement seen in the quota testing success.

**USAGE EXAMPLES:**
- For arcbox command: Replace `[REPLACE WITH PACKAGE PATH]` with `cmd/arcbox`
- For resource providers: Replace `[REPLACE WITH PACKAGE PATH]` with `internal/resourceproviders`
- For agora command: Replace `[REPLACE WITH PACKAGE PATH]` with `cmd/agora`

---

**Proven Results**: This methodology increased quota package coverage from 51.0% to 94.3% (+43.3% improvement) while identifying and fixing critical production bugs. It provides a systematic, repeatable approach for achieving excellent test coverage across all CLI command packages.
