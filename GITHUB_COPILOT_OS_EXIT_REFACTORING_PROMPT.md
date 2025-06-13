<!-- ========== GITHUB COPILOT PROMPT START ========== -->

# GitHub Copilot: Go CLI `os.Exit()` Refactoring Guide

## 🎯 **Objective**
Refactor this Go CLI application to follow the principle: **"Only use `os.Exit()` in `main()` functions"**

## 🧭 **Core Principle**
✅ **Only use `os.Exit()` in `main()`**
- Keeps core logic testable
- Prevents surprises like skipped defer statements or abrupt process termination during testing
- Aligns with Go's philosophy: "errors are values" → prefer returning errors over exiting

## 📊 **Current State Analysis**
This codebase has **42 instances** of `os.Exit()` outside of `main()`:
- 27 instances in Cobra command `Run` functions (CLI entry points) - **KEEP AS-IS**
- 2 instances in utility functions - **REFACTOR PRIORITY HIGH**
- 1 instance in business logic function - **REFACTOR PRIORITY HIGH**
- 12 instances in CLI-specific validation - **REFACTOR PRIORITY MEDIUM**

## 🔧 **Refactoring Strategy**

### **HIGH PRIORITY: Business Logic Functions**
**Target Files:**
- `internal/preflight/arcbox/rp.go` - `RegisterResourceProvider` function
- `internal/utils/utils.go` - `Fatal` function

**Pattern to Apply:**
```go
// ❌ BEFORE: Direct os.Exit()
func RegisterResourceProvider(cli azurecli.AzureCLI, provider string) {
    if err := resourceproviders.RegisterProvider(cli, provider); err != nil {
        fmt.Printf(utils.ErrorColor("❌ [ERROR] Failed to register resource provider '%s': %v\n"), provider, err)
        os.Exit(1)
    }
}

// ✅ AFTER: Return errors
func RegisterResourceProvider(cli azurecli.AzureCLI, provider string) error {
    if err := resourceproviders.RegisterProvider(cli, provider); err != nil {
        return fmt.Errorf("failed to register resource provider '%s': %w", provider, err)
    }
    fmt.Printf(utils.SuccessColor("✅ [SUCCESS] Successfully registered resource provider '%s'\n"), provider)
    return nil
}
```

### **HIGH PRIORITY: Utility Functions**
**Target:** `internal/utils/utils.go`

**Pattern to Apply:**
```go
// ✅ Keep existing Fatal for backward compatibility
func Fatal(msg string, args ...interface{}) {
    fmt.Fprintln(os.Stderr, FatalColor("[FATAL]"), fmt.Sprintf(msg, args...))
    os.Exit(1)
}

// ✅ ADD: New error-returning variant for testable code
func FatalError(msg string, args ...interface{}) error {
    return fmt.Errorf("[FATAL] "+msg, args...)
}

// ✅ REFACTOR: PrintMissingRequiredFlagsError to return boolean
func PrintMissingRequiredFlagsError(cmd *cobra.Command, requiredFlags []string) bool {
    missing := []string{}
    // ...existing logic...
    if len(missing) > 0 {
        fmt.Fprintf(os.Stderr, "%s\n\n", ErrorColor("the following arguments are required: "+strings.Join(missing, ", ")))
        ShowHelpWithoutTypes(cmd)
        return false // Indicate validation failed
    }
    return true // Indicate validation passed
}
```

### **KEEP AS-IS: CLI Command Handlers**
**Target Files:** All `cmd/*/` files with Cobra command `Run` functions

**Reasoning:**
```go
// ✅ ACCEPTABLE: These ARE the application entry points
var arcboxDeployCmd = &cobra.Command{
    Run: func(cmd *cobra.Command, args []string) {
        // CLI command handlers can use os.Exit() - they are entry points
        if err := utils.ValidateAllFlags(cmd); err != nil {
            os.Exit(1) // This is OK - CLI entry point behavior
        }
        
        // But call refactored functions that return errors
        if err := RegisterResourceProvider(cli, provider); err != nil {
            fmt.Printf(utils.ErrorColor("❌ [ERROR] %v\n"), err)
            os.Exit(1)
        }
    },
}
```

### **MEDIUM PRIORITY: Validation Functions**
**Target:** `internal/utils/utils.go` - validation functions

**Pattern to Apply:**
```go
// ✅ Make validation functions return errors, let CLI handlers decide exit behavior
func ValidateAllFlags(cmd *cobra.Command) error {
    // Return validation errors instead of exiting
    // Let the caller (CLI handler) decide whether to exit
}
```

## 🏗️ **Implementation Guidelines**

### **1. Function Signature Changes**
```go
// ❌ BEFORE: void function with os.Exit()
func DoSomething(params) {
    if err := operation(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}

// ✅ AFTER: return error
func DoSomething(params) error {
    if err := operation(); err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    return nil
}
```

### **2. Error Wrapping**
```go
// ✅ Use fmt.Errorf with %w verb for error wrapping
return fmt.Errorf("failed to register resource provider '%s': %w", provider, err)
```

### **3. Caller Updates**
```go
// ✅ Update callers to handle returned errors
if err := RegisterResourceProvider(cli, provider); err != nil {
    fmt.Printf(utils.ErrorColor("❌ [ERROR] %v\n"), err)
    os.Exit(1) // Only in CLI handlers
}
```

### **4. Backward Compatibility**
```go
// ✅ Keep existing functions, add new variants
func Fatal(msg string, args ...interface{}) {
    // Keep for backward compatibility
    fmt.Fprintln(os.Stderr, FatalColor("[FATAL]"), fmt.Sprintf(msg, args...))
    os.Exit(1)
}

func FatalError(msg string, args ...interface{}) error {
    // New testable variant
    return fmt.Errorf("[FATAL] "+msg, args...)
}
```

## 🧪 **Testing Considerations**

### **Before Refactoring:**
```go
// ❌ Cannot unit test - calls os.Exit()
func TestRegisterResourceProvider(t *testing.T) {
    // This test would terminate the test process!
    // RegisterResourceProvider(mockCLI, "Microsoft.Compute")
    t.Skip("Cannot test - function calls os.Exit()")
}
```

### **After Refactoring:**
```go
// ✅ Can unit test - returns error
func TestRegisterResourceProvider(t *testing.T) {
    mockCLI := &azurecli.MockAzureCLI{}
    mockCLI.On("RegisterProvider", "Microsoft.Compute").Return(errors.New("registration failed"))
    
    err := RegisterResourceProvider(mockCLI, "Microsoft.Compute")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to register resource provider")
}
```

## 📋 **Refactoring Checklist**

### **Phase 1: High Priority (Business Logic)**
- [ ] Refactor `RegisterResourceProvider` to return error
- [ ] Add `FatalError` variant to utils
- [ ] Update callers of refactored functions
- [ ] Add unit tests for refactored functions

### **Phase 2: Medium Priority (Validation)**
- [ ] Refactor `PrintMissingRequiredFlagsError` to return boolean
- [ ] Refactor `ValidateAllFlags` to return error
- [ ] Update CLI command handlers to handle returned errors

### **Phase 3: Testing & Verification**
- [ ] Add comprehensive unit tests for all refactored functions
- [ ] Run integration tests to ensure CLI behavior unchanged
- [ ] Verify all defer statements execute properly
- [ ] Check that error messages remain user-friendly

## ⚠️ **Important Notes**

1. **CLI Behavior Must Remain Unchanged**: Users should not notice any difference in CLI behavior
2. **Error Messages**: Maintain the same user-friendly error formatting
3. **Exit Codes**: Preserve the same exit codes (0 for success, 1 for errors)
4. **Backward Compatibility**: Keep existing function signatures where possible
5. **Testing**: Add unit tests for all refactored functions

## 🎯 **Success Criteria**

✅ **Business logic functions return errors instead of calling `os.Exit()`**
✅ **All business logic functions have comprehensive unit tests**
✅ **CLI behavior remains exactly the same for end users**
✅ **Error handling is more robust and testable**
✅ **Defer statements execute properly in all code paths**

---

**Remember**: The goal is to make the codebase more testable and robust while maintaining the exact same CLI user experience. Focus on separating business logic (which should return errors) from CLI presentation logic (which can still use `os.Exit()`)

<!-- ========== GITHUB COPILOT PROMPT END ========== -->
