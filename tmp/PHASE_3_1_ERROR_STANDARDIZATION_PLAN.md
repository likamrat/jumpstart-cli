# Phase 3.1: Error Handling Standardization Plan

## Current Error Handling Analysis

### Current Patterns Found

#### Service Layer Error Patterns (Inconsistent)
1. **Basic error wrapping**: `fmt.Errorf("operation description: %v", err)`
2. **Error wrapping with %w**: `fmt.Errorf("operation description: %w", err)`
3. **Simple error messages**: `fmt.Errorf("validation message")`
4. **Mixed context inclusion**: Some include context, others don't

#### CLI Error Display Patterns (Mixed)
1. **utils.Error()**: `utils.Error("message", args...)`
2. **Direct ErrorColor**: `fmt.Printf(utils.ErrorColor("❌ [ERROR] %v\n"), err)`
3. **Mixed formatting**: Some use ❌, some use [ERROR], some use both

#### Issues to Address
1. **Inconsistent error wrapping**: Mix of %v and %w verbs
2. **Missing context**: Not all errors include relevant context
3. **Inconsistent CLI formatting**: Different error display styles
4. **Service layer output**: Some services print directly instead of returning errors

## Standardization Plan

### 1. Service Layer Error Standard

#### Error Wrapping Pattern
```go
// Standard pattern for service layer errors
return fmt.Errorf("operation description for %s: %w", context, underlyingError)

// Examples:
return fmt.Errorf("failed to validate subscription access for '%s': %w", subscription, err)
return fmt.Errorf("failed to check resource group existence for '%s': %w", resourceGroupName, err)
```

#### Error Message Format Rules
- Start with operation description (lowercase, verb-based)
- Include relevant context in quotes
- Use %w verb for error wrapping (enables errors.Is/As)
- Use user-friendly language
- No service layer should print directly to console

### 2. CLI Error Display Standard

#### Consistent CLI Error Format
```go
// Standard CLI error display
if err != nil {
    utils.Error(err.Error())
    os.Exit(1)
}
```

#### Update utils.Error to be consistent
```go
func Error(msg string, args ...interface{}) {
    fmt.Fprintln(os.Stderr, ErrorColor(fmt.Sprintf("❌ [ERROR] "+msg, args...)))
}
```

### 3. Context Requirements

Each error should include relevant context:
- **Quota operations**: subscription, location, flavor
- **Resource operations**: resource group, subscription  
- **Validation operations**: flag names, required values
- **Azure operations**: subscription, CLI status

## Implementation Steps

### Step 1: Update Service Layer Error Patterns
- [ ] Update delete validation service errors
- [ ] Update list validation service errors  
- [ ] Update deploy validation service errors
- [ ] Update quota service errors
- [ ] Update deployment service errors
- [ ] Update deletion service errors
- [ ] Update listing service errors

### Step 2: Remove Service Layer Console Output
- [ ] Remove utils.Error calls from service layer
- [ ] Remove direct fmt.Printf calls from service layer
- [ ] Ensure all service errors return to caller

### Step 3: Standardize CLI Error Display
- [ ] Update CLI handlers to use consistent error display
- [ ] Ensure all CLI errors use utils.Error()
- [ ] Update utils.Error() to use consistent format

### Step 4: Update Tests
- [ ] Update service layer tests for new error patterns
- [ ] Update error message assertions
- [ ] Add tests for error wrapping

### Step 5: Documentation
- [ ] Document new error handling patterns
- [ ] Update function documentation
- [ ] Add examples of error handling

## Files to Modify

### Service Layer Files
- `cmd/arcbox/services/delete_validation_service.go`
- `cmd/arcbox/services/list_validation_service.go`
- `cmd/arcbox/services/deploy_validation_service.go`
- `cmd/arcbox/services/quota_service.go`
- `cmd/arcbox/services/deployment_service.go`
- `cmd/arcbox/services/deletion_service.go`
- `cmd/arcbox/services/listing_service.go`

### CLI Handler Files
- `cmd/arcbox/delete_cmd.go`
- `cmd/arcbox/deploy_cmd.go`
- `cmd/arcbox/list_cmd.go`
- `cmd/arcbox/preflight_cmd.go`

### Utility Files
- `internal/utils/utils.go`

### Test Files (All service test files)
- Update error message assertions
- Add error wrapping tests

## Success Criteria

1. **Consistency**: All service layer errors follow the same pattern
2. **Context**: All errors include relevant context information
3. **Wrapping**: All errors use %w verb for proper error wrapping
4. **Separation**: Service layer never prints to console directly
5. **CLI Consistency**: All CLI errors use the same display format
6. **Testing**: All error scenarios are properly tested
