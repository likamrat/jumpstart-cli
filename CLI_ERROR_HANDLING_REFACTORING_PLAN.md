# CLI Error Handling Refactoring Plan
## Azure CLI Consistency & Global Error Handling Implementation

**Objective**: Refactor the Jumpstart CLI to achieve Azure CLI-like consistency in error handling while maintaining functionality, testability, and avoiding hard-coded command-specific logic.

**Key Principles**:
- ✅ Preserve ALL existing CLI functionality
- ✅ Maintain testing compatibility and mock-friendly interfaces
- ✅ Create global, reusable error handling patterns
- ✅ Eliminate hard-coded command-specific logic
- ✅ Achieve Azure CLI-like consistency
- ✅ Build and test CLI after each phase

---

## **Current State Assessment**

### **Existing Error Handling Infrastructure**

#### **Current Utils Functions (internal/utils/utils.go)**
```go
// EXISTING - Good foundation but needs enhancement
✅ PrintMissingRequiredFlagsError(cmd, requiredFlags []string) bool
✅ PrintMissingRequiredArgumentsTip(cmd *cobra.Command)
✅ ShowHelpWithoutTypes(cmd *cobra.Command)
✅ PrintDidYouMean(invalid, suggestion string)
✅ SuggestSimilarCommand(input string, commands []string, threshold int) string

// EXISTING - Color functions (keep as-is)
✅ ErrorColor, InfoColor, SuccessColor, etc.

// EXISTING - But needs cleanup/consolidation
⚠️ PrintMissingRequiredArgumentsError(cmd, requiredArgs []string) - calls os.Exit
⚠️ Fatal(msg string, args ...interface{}) - calls os.Exit
✅ FatalError(msg string, args ...interface{}) error - testing-friendly version
```

#### **Current Validation Service Pattern**
```go
// EXISTING - Good pattern used in ArcBox commands
✅ ValidationResult struct with IsValid bool and Error error
✅ Service-based validation (deploy_validation_service.go, delete_validation_service.go, etc.)
✅ Separation of business logic from error display
```

#### **Current Command Error Patterns**
```go
// PROBLEMATIC - Multiple patterns exist:
❌ Hard-coded error messages in commands
❌ Inconsistent use of os.Exit vs return error
❌ Mixed error prefixes (❌ [ERROR] vs clean messages)
❌ Duplicate error messages in some commands
❌ Missing tip messages in some commands
```

### **Files Requiring Assessment/Cleanup**

#### **Files to Create**
- `internal/utils/error_handling.go` - New centralized error handling utilities

#### **Files to Update (Error Handling)**
- `internal/utils/utils.go` - Update existing functions to use centralized approach
- `cmd/subscription/subscription.go` - Standardize error handling
- `cmd/arcbox/deploy_cmd.go` - Fix duplicate errors, use centralized handling
- `cmd/arcbox/delete_cmd.go` - Standardize error handling
- `cmd/arcbox/list_cmd.go` - Standardize subscription validation errors
- `cmd/arcbox/preflight_cmd.go` - Standardize preflight command errors
- `internal/preflight/arcbox/rp.go` - Standardize resource provider command errors
- `cmd/repo/*.go` - Add proper error handling and tips
- `cmd/upgrade/upgrade.go` - Add proper error handling and tips
- `cmd/completion/completion.go` - Standardize error handling
- `main.go` - Ensure root command uses centralized patterns

#### **Files Requiring Cleanup**
```go
// CONSOLIDATE - Remove redundant functions after refactoring
⚠️ PrintMissingRequiredArgumentsError() - replace with centralized version
⚠️ Hard-coded error messages in validation services
⚠️ Inconsistent error handling in command files
```

#### **Files to Keep As-Is**
```go
// PRESERVE - These work well and support the new pattern
✅ All validation service interfaces and business logic
✅ Mock interfaces for testing (azurecli.MockAzureCLI, etc.)
✅ Color utility functions
✅ Help formatting functions
✅ Testing infrastructure and patterns
```

### **Key Cleanup Opportunities**

#### **1. Consolidate Error Functions**
```go
// BEFORE (multiple functions)
PrintMissingRequiredFlagsError() 
PrintMissingRequiredArgumentsError()
PrintMissingRequiredArgumentsTip()

// AFTER (centralized)
HandleMissingRequiredArguments() // combines all three
```

#### **2. Eliminate Hard-coded Messages**
```go
// BEFORE - Command-specific hard-coded messages
fmt.Fprintf(cmd.ErrOrStderr(), utils.ErrorColor("Cannot specify multiple..."))

// AFTER - Centralized, reusable functions
utils.HandleValidationError(err, cmd, true)
```

#### **3. Standardize os.Exit Usage**
```go
// BEFORE - Mixed patterns
os.Exit(1) // in some commands
return error // in others

// AFTER - Consistent testing-friendly pattern
return error // in business logic
os.Exit(1) // only in main command handlers
```

#### **4. Remove Duplicate Logic**
- Multiple commands have similar validation patterns
- Error display logic repeated across files
- Tip message formatting inconsistent

---

## **Phase 1: Foundation - Create Centralized Error Handling**

### **Prompt 1.1: Create Global Error Handling Utilities**

<!-- START PROMPT 1.1 -->
I need to create a new file `internal/utils/error_handling.go` that provides centralized, reusable error handling functions for the CLI. This must be testing-friendly and avoid hard-coded command-specific logic.

**ASSESSMENT**: Current codebase has `PrintMissingRequiredFlagsError`, `PrintMissingRequiredArgumentsTip`, and related functions in `utils.go`, but they have inconsistencies and some call `os.Exit`. We need centralized functions that work with the existing validation service pattern.

Requirements:
1. Create `PrintRequiredArgumentsError(missingArgs []string)` - formats "the following arguments are required: [args]" with NO error prefix
2. Create `PrintStandardHelpTip(cmd *cobra.Command)` - shows "💡 [TIP] Use '[command] --help' to see all required arguments" using InfoColor
3. Create `HandleMissingRequiredArguments(cmd *cobra.Command, missingArgs []string)` - combines error + help + tip
4. Create `HandleValidationError(err error, cmd *cobra.Command, showHelp bool)` - generic validation error handler
5. All functions must be testing-friendly (no os.Exit calls)
6. Use existing color functions from utils package
7. Follow existing code patterns in the codebase
8. Integrate with existing ValidationResult pattern used in validation services

After creating the file, verify it compiles by running `make build`.
<!-- END PROMPT 1.1 -->

### **Prompt 1.2: Update Utils Package Integration and Cleanup**

<!-- START PROMPT 1.2 -->
I need to update the existing `internal/utils/utils.go` file to integrate with the new error handling utilities and clean up redundant functions.

**CLEANUP NEEDED**: Current `utils.go` has `PrintMissingRequiredArgumentsError()` which calls `os.Exit` - this should be updated to use the new centralized functions while maintaining backward compatibility for existing callers.

Requirements:
1. Review existing error handling functions: `PrintMissingRequiredFlagsError`, `PrintMissingRequiredArgumentsTip`, `PrintMissingRequiredArgumentsError`
2. Update `PrintMissingRequiredArgumentsError` to use new centralized functions instead of calling `os.Exit` directly
3. Update `PrintMissingRequiredFlagsError` if needed to work with centralized approach
4. Ensure backward compatibility - existing calls should continue to work
5. Remove any duplicate logic that's now handled by the new utilities
6. Maintain all existing function signatures to avoid breaking tests
7. Keep the existing `FatalError` function (testing-friendly) and `Fatal` function (calls os.Exit)

Build and verify: `make build`
Test that existing functionality still works by running a simple command like `js --help`
<!-- END PROMPT 1.2 -->

---

## **Phase 2: Subscription Commands Refactoring**

### **Prompt 2.1: Refactor Subscription Set Command**

<!-- START PROMPT 2.1 -->
I need to refactor the subscription set command in `cmd/subscription/subscription.go` to use the new centralized error handling approach.

Current behavior to maintain:
- Command requires one of: --subscription/-s, --name/-n, or positional argument
- Should show clean error message: "the following arguments are required (choose one): --subscription/-s, --name/-n, or positional argument"
- Should show help and tip message
- Must handle mutual exclusion validation

Requirements:
1. Use the new `HandleMissingRequiredArguments` function from error_handling.go
2. Remove hard-coded error messages and use centralized approach
3. Maintain existing validation logic
4. Preserve testing compatibility (avoid os.Exit in core logic)
5. Keep all existing functionality intact

After changes:
1. Build: `make build`
2. Test the command: `js subscription set` (should show clean error)
3. Test with valid args: `js subscription set --help` (should work)
4. Verify error message format matches Azure CLI style
<!-- END PROMPT 2.1 -->

### **Prompt 2.2: Refactor All Subscription Subcommands**

<!-- START PROMPT 2.2 -->
I need to refactor all remaining subscription subcommands (`show`, `list`) in `cmd/subscription/subscription.go` to use consistent error handling.

Requirements:
1. Apply the same centralized error handling pattern to all subscription subcommands
2. Ensure consistent tip messages across all subcommands
3. Maintain all existing validation and business logic
4. Remove any hard-coded error formatting
5. Preserve testing interfaces and mock compatibility

After changes:
1. Build: `make build`
2. Test each subcommand:
   - `js subscription show --help`
   - `js subscription list --help`
   - `js subscription` (should show help)
3. Verify all commands work with valid arguments
4. Check that error messages are consistent and follow Azure CLI format
<!-- END PROMPT 2.2 -->

---

## **Phase 3: ArcBox Commands Refactoring**

### **Prompt 3.1: Refactor ArcBox Deploy Command**

<!-- START PROMPT 3.1 -->
I need to refactor the ArcBox deploy command in `cmd/arcbox/deploy_cmd.go` to eliminate duplicate error messages and use centralized error handling.

**CLEANUP NEEDED**: Current deploy command uses `deploy_validation_service.go` which calls `utils.PrintMissingRequiredFlagsError()` and has duplicate error messages. The validation service pattern is good but needs to use the new centralized error handling.

Current issues to fix:
- Duplicate error messages appearing (error shows twice)
- Inconsistent error formatting with ❌ [ERROR] prefix
- Need to standardize required arguments error handling

Requirements:
1. Use new centralized error handling functions from `error_handling.go`
2. Eliminate duplicate error messages completely
3. Maintain all existing validation logic through the validation service (preserve `deploy_validation_service.go`)
4. Preserve testing compatibility (separate business logic from error display)
5. Handle the validation service pattern properly - update `deploy_validation_service.go` if needed
6. Keep all existing functionality and arguments
7. Preserve existing test interfaces and mock compatibility

After changes:
1. Build: `make build`
2. Test deploy command: `js arcbox deploy` (should show clean error, help, and tip - NO duplicates)
3. Test with some args: `js arcbox deploy --location eastus` (should show remaining required args)
4. Verify the command works with all required arguments
5. Check that examples in help output are preserved
<!-- END PROMPT 3.1 -->

### **Prompt 3.2: Refactor ArcBox Delete Command**

<!-- START PROMPT 3.2 -->
I need to refactor the ArcBox delete command in `cmd/arcbox/delete_cmd.go` to use consistent error handling and eliminate any duplicate messages.

Requirements:
1. Apply centralized error handling pattern
2. Ensure consistent language: "the following arguments are required: --name/-n"
3. Remove any ❌ [ERROR] prefixes for required argument errors
4. Add proper tip messages
5. Maintain existing validation and business logic
6. Keep testing compatibility

After changes:
1. Build: `make build`
2. Test delete command: `js arcbox delete` (should show clean error format)
3. Test with valid args: `js arcbox delete --name test-deployment`
4. Verify help and examples are preserved
<!-- END PROMPT 3.2 -->

### **Prompt 3.3: Refactor ArcBox List Command**

<!-- START PROMPT 3.3 -->
I need to refactor the ArcBox list command in `cmd/arcbox/list_cmd.go` to use consistent error handling for subscription selection validation.

Current issues:
- Mixed error handling approaches in `executeListCommand` and `handleListCommandError`
- Inconsistent error formatting for subscription selection errors

Requirements:
1. Use centralized error handling for subscription selection validation
2. Maintain the separation between `executeListCommand` and `handleListCommandError`
3. Preserve all existing subscription selection logic
4. Keep testing-friendly pattern (return errors, don't exit in core logic)
5. Standardize error messages for subscription requirements

After changes:
1. Build: `make build`
2. Test list command: `js arcbox list` (should show clean subscription selection error)
3. Test with valid subscription args: `js arcbox list --current-subscription`
4. Verify all subscription selection scenarios work correctly
<!-- END PROMPT 3.3 -->

### **Prompt 3.4: Refactor ArcBox Preflight Commands**

<!-- START PROMPT 3.4 -->
I need to refactor the ArcBox preflight commands in `cmd/arcbox/preflight_cmd.go` and related files to use consistent error handling.

Focus areas:
- `preflight quota` command error handling
- `preflight rp` subcommands (show, list, register)
- Resource provider command validation

Requirements:
1. Apply centralized error handling to all preflight subcommands
2. Handle the quota service error patterns properly
3. Maintain existing timeout and validation logic
4. Preserve testing compatibility (avoid os.Exit in testable functions)
5. Keep all business logic intact

Special considerations:
- Some commands like `js arcbox preflight quota` may take time - ensure proper timeout handling
- Resource provider commands have complex validation - preserve all existing logic

After changes:
1. Build: `make build`
2. Test preflight commands:
   - `js arcbox preflight` (should show help)
   - `js arcbox preflight quota` (should show clean error for missing args)
   - `js arcbox preflight rp` (should show help)
   - `js arcbox preflight rp show` (test with timeout if needed)
3. Verify all validation scenarios work correctly
4. Check that examples and help text are preserved
<!-- END PROMPT 3.4 -->

---

## **Phase 4: Remaining Commands Refactoring**

### **Prompt 4.1: Refactor Repo Commands**

<!-- START PROMPT 4.1 -->
I need to refactor all repo commands in `cmd/repo/` to use consistent error handling and add proper tip messages.

Current state: Basic error handling, missing helpful tips

Requirements:
1. Apply centralized error handling to all repo subcommands (init, update, delete)
2. Add proper tip messages where missing
3. Standardize error message formats
4. Maintain all existing business logic and validation
5. Preserve testing compatibility

After changes:
1. Build: `make build`
2. Test all repo commands:
   - `js repo` (should show help)
   - `js repo init` (test error handling)
   - `js repo update` (test error handling)
   - `js repo delete` (test error handling)
3. Verify all commands work with valid arguments
4. Check that help and examples are preserved
<!-- END PROMPT 4.1 -->

### **Prompt 4.2: Refactor Upgrade Commands**

<!-- START PROMPT 4.2 -->
I need to refactor upgrade commands in `cmd/upgrade/upgrade.go` to use consistent error handling.

Current state: Minimal error guidance, needs standardization

Requirements:
1. Apply centralized error handling to all upgrade subcommands (check, install, rollback, list)
2. Add proper tip messages
3. Maintain existing --yes flag validation logic
4. Preserve all upgrade functionality and safety checks
5. Keep testing-friendly patterns

After changes:
1. Build: `make build`
2. Test all upgrade commands:
   - `js upgrade` (should show help)
   - `js upgrade check` (test without --yes flag)
   - `js upgrade install` (test without --yes flag)
   - Test with --yes flag to ensure functionality preserved
3. Verify all safety checks and validations work
4. Check that examples and help are preserved
<!-- END PROMPT 4.2 -->

### **Prompt 4.3: Refactor Completion Command**

<!-- START PROMPT 4.3 -->
I need to refactor the completion command in `cmd/completion/completion.go` to use consistent error handling.

Requirements:
1. Apply centralized error handling patterns
2. Add proper tip messages for missing arguments
3. Maintain all existing shell completion functionality
4. Preserve validation for supported shells
5. Keep testing compatibility

After changes:
1. Build: `make build`
2. Test completion command:
   - `js completion` (should show clean error for missing shell)
   - `js completion bash` (should work)
   - `js completion zsh` (should work)
3. Verify all supported shells work correctly
4. Check help and examples are preserved
<!-- END PROMPT 4.3 -->

---

## **Phase 5: Root Command and Global Consistency**

### **Prompt 5.1: Review and Standardize Root Command Behavior**

<!-- START PROMPT 5.1 -->
I need to review the root command in `main.go` and ensure consistent error handling for unknown commands and global error patterns.

Requirements:
1. Ensure root command error handling uses centralized functions
2. Verify "did you mean" suggestions work consistently
3. Check that unknown command handling is standardized
4. Maintain all existing suggestion logic
5. Preserve version and help command functionality

After changes:
1. Build: `make build`
2. Test root command scenarios:
   - `js` (should show help)
   - `js unknowncommand` (should show suggestion if similar)
   - `js --version` (should work)
   - `js --help` (should work)
3. Verify all global command patterns work correctly
<!-- END PROMPT 5.1 -->

### **Prompt 5.2: Comprehensive CLI Testing and Validation**

<!-- START PROMPT 5.2 -->
I need to perform comprehensive testing of all CLI commands to ensure the refactoring didn't break any functionality and that error handling is consistent across all commands.

Testing checklist:
1. **Build verification**: `make build` should complete successfully
2. **Command structure verification**: All commands and subcommands should be available
3. **Error message consistency**: All required argument errors should follow Azure CLI format
4. **Tip message presence**: All commands should show helpful tip messages
5. **Functionality preservation**: All existing features should work as before
6. **Examples preservation**: Help text and examples should be intact

Systematic testing approach:
1. Test each command family (subscription, arcbox, repo, upgrade, completion)
2. Test both error scenarios and successful execution paths
3. Verify help text and examples are preserved
4. Check that all error messages are consistent
5. Ensure no duplicate error messages exist
6. Confirm all tip messages use proper coloring (InfoColor)

Create a testing summary report documenting:
- All commands tested
- Error message formats verified
- Any issues found and resolved
- Confirmation that functionality is preserved
<!-- END PROMPT 5.2 -->

---

## **Phase 6: Test Suite Updates (Final Phase)**

### **Prompt 6.1: Update Test Suite for New Error Handling**

<!-- START PROMPT 6.1 -->
Now that all core functionality is working, I need to update the test suite to work with the new centralized error handling while maintaining all existing test coverage.

Requirements:
1. Review all test files for commands that were refactored
2. Update test expectations to match new error message formats
3. Ensure tests still validate core functionality
4. Maintain test isolation and mock compatibility
5. Add tests for new centralized error handling functions
6. Preserve all existing test coverage

Focus areas:
- `cmd/subscription/subscription_test.go`
- `cmd/arcbox/*_test.go` files
- `cmd/repo/*_test.go` files
- `cmd/upgrade/upgrade_test.go`
- `internal/utils/utils_test.go`

After changes:
1. Build: `make build`
2. Run all tests: `make test` or `go test ./...`
3. Ensure all tests pass
4. Verify test coverage is maintained
5. Check that mock interfaces still work correctly
<!-- END PROMPT 6.1 -->

### **Prompt 6.2: Final Integration Testing and Documentation**

<!-- START PROMPT 6.2 -->
I need to perform final integration testing and update documentation to reflect the new consistent error handling approach.

Requirements:
1. **Final comprehensive testing**: Test all command paths and scenarios
2. **Performance verification**: Ensure no performance regressions
3. **Error handling documentation**: Update any relevant documentation
4. **Integration test verification**: Run integration tests if they exist
5. **Backward compatibility confirmation**: Ensure all existing usage patterns work

Final validation checklist:
- [ ] All commands build successfully
- [ ] All error messages follow Azure CLI format
- [ ] All commands show helpful tips
- [ ] No duplicate error messages exist
- [ ] All existing functionality preserved
- [ ] All tests pass
- [ ] Examples and help text intact
- [ ] Performance is acceptable
- [ ] Mock interfaces work correctly

Create a final report documenting:
- Summary of changes made
- Commands refactored
- Error handling improvements
- Testing results
- Any remaining considerations
<!-- END PROMPT 6.2 -->

---

## **Key Success Criteria**

### **Functionality Preservation**
- ✅ All existing CLI commands work exactly as before
- ✅ All validation logic maintained
- ✅ All business logic preserved
- ✅ All examples and help text intact

### **Consistency Achieved**
- ✅ All required argument errors follow format: "the following arguments are required: [args]"
- ✅ No ❌ [ERROR] prefixes for required argument errors
- ✅ All commands show helpful tip messages
- ✅ Consistent coloring and formatting

### **Technical Quality**
- ✅ No hard-coded command-specific error handling
- ✅ Centralized, reusable error handling functions
- ✅ Testing-friendly design (no os.Exit in core logic)
- ✅ Mock compatibility preserved
- ✅ All tests pass

### **Build and Performance**
- ✅ CLI builds successfully after each phase
- ✅ No performance regressions
- ✅ Commands execute within acceptable timeouts
- ✅ No broken functionality

---

## **Post-Refactoring Cleanup Actions**

### **Functions to Deprecate/Remove (After Phase 6)**
```go
// These functions will be consolidated into centralized error handling
❌ PrintMissingRequiredArgumentsError() // Replace with HandleMissingRequiredArguments()
⚠️ Hard-coded error messages in validation services // Replace with centralized functions
⚠️ Command-specific error handling patterns // Replace with generic patterns
```

### **Files Modified Summary**
```
CREATED:
✅ internal/utils/error_handling.go

UPDATED:
✅ internal/utils/utils.go - Updated existing functions
✅ cmd/subscription/subscription.go - Standardized error handling  
✅ cmd/arcbox/deploy_cmd.go - Fixed duplicates, centralized handling
✅ cmd/arcbox/delete_cmd.go - Standardized error handling
✅ cmd/arcbox/list_cmd.go - Standardized subscription validation
✅ cmd/arcbox/preflight_cmd.go - Standardized all preflight commands
✅ internal/preflight/arcbox/rp.go - Standardized resource provider commands
✅ cmd/repo/*.go - Added proper error handling and tips
✅ cmd/upgrade/upgrade.go - Added proper error handling and tips
✅ cmd/completion/completion.go - Standardized error handling
✅ main.go - Root command uses centralized patterns
✅ All test files updated for new error message formats

PRESERVED:
✅ All validation service business logic
✅ All mock interfaces and testing infrastructure
✅ All existing functionality and command behavior
✅ All examples and help text
```

### **Validation Checklist (Post-Implementation)**
- [ ] `make build` completes successfully
- [ ] All commands show Azure CLI-style error messages
- [ ] No duplicate error messages exist anywhere
- [ ] All commands show helpful tip messages  
- [ ] All existing functionality preserved
- [ ] All tests pass
- [ ] Mock interfaces work correctly
- [ ] Performance is acceptable
- [ ] Examples and help text intact

---

## **Implementation Notes**

### **Special Considerations**
- Commands like `js arcbox preflight quota` may take time - handle with appropriate timeouts
- Mock interfaces must remain compatible for testing
- Validation services should be preserved, not replaced
- Examples and help text are critical user-facing features
- Error handling should be generic, not command-specific

### **Risk Mitigation**
- Build and test after each prompt/phase
- Maintain backward compatibility throughout
- Preserve existing test interfaces
- Keep core business logic separate from error display
- Test both success and failure scenarios

### **Quality Assurance**
- Run actual commands, don't just report success
- Verify error message formats manually
- Check that all functionality works with real arguments
- Ensure help text and examples are preserved
- Confirm no duplicate error messages exist
