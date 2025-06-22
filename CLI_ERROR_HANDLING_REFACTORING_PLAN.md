# CLI Error Handling Refactoring Plan
## Azure CLI Consistency & Global Error Handling Implementation

**Objective**: Refactor the Jumpstart CLI to achieve Azure CLI-like consistency in error handling while maintaining functionality, testability, and avoiding hard-coded command-specific logic.

## **🎯 CRITICAL GLOBAL REQUIREMENTS**

**EXACT ERROR MESSAGE FORMAT** (applies to ALL commands and subcommands):
- **Format**: `"the following arguments are required: [args]"`
- **Color**: RED (using ErrorColor function)
- **No prefixes**: Absolutely no "❌ [ERROR]" or any prefix
- **No suffixes**: Absolutely no "(choose one)" or any extra text
- **Global consistency**: Every command must produce IDENTICAL error message format

### **🎯 DESIRED STATE - EXACT AZURE CLI BEHAVIOR**

Our CLI must behave EXACTLY like Azure CLI. Here are the exact patterns to replicate:

**1. Missing Subcommand Example:**
```bash
~/repos/jumpstart-cli initial_commit ❯ az vm
the following arguments are required: _subcommand
```

**2. Missing Required Flag Example:**
```bash
~/repos/jumpstart-cli initial_commit ❯ az vm list-usage
the following arguments are required: --location/-l
```

**CRITICAL IMPLEMENTATION REQUIREMENTS:**
- ✅ **Exact text**: "the following arguments are required: [args]"
- ✅ **RED color**: The entire error message must be displayed in red
- ✅ **No prefixes**: No "Error:", no "❌", no brackets, no decorations
- ✅ **No suffixes**: No "(choose one)", no extra explanatory text
- ✅ **Clean format**: Just the plain error message in red, exactly like Azure CLI
- ✅ **Global consistency**: ALL commands must use this identical format
- ✅ **Remove help footer**: Remove "Use [command] --help" messages from help output (Azure CLI doesn't show these)
- ✅ **Remove tip messages**: Remove "💡 [TIP] Use '<command>' to see all required arguments" messages (Azure CLI doesn't show these)
- ✅ **Clean minimal output**: Remove excessive help text after errors - show only the required arguments error message, not the full help text with descriptions and examples (Azure CLI shows minimal output)

**Our CLI Target Examples:**
```bash
# When missing subcommand (CLEAN OUTPUT - just the error message)
~/repos/jumpstart-cli ❯ js arcbox
the following arguments are required: _subcommand

# When missing required flags (CLEAN OUTPUT - just the error message)  
~/repos/jumpstart-cli ❯ js arcbox deploy
the following arguments are required: --location/-l, --name/-n

# When missing subscription argument (CLEAN OUTPUT - just the error message)
~/repos/jumpstart-cli ❯ js subscription set
the following arguments are required: --subscription/-s, --name/-n, or positional argument
```

**CURRENT PROBLEM - TOO MUCH TEXT:**  
```bash
# BEFORE (TOO VERBOSE - like current CLI):
~/repos/jumpstart-cli ❯ js arcbox deploy
the following arguments are required: --location/-l, --resource-group/-g, --windows-user, --flavor/-f

Usage:
  js arcbox deploy [flags]

Deploy a new Jumpstart ArcBox deployment

Deploy a new Jumpstart ArcBox deployment using remote Bicep or ARM templates from GitHub.

By default, uses the official ArcBox ARM template from GitHub. You can specify:
- Custom remote template URI with --template-uri (ARM templates only - Bicep doesn't support remote templates)
- Local template files with --template-local and --template-params (for local Bicep/ARM templates)

# AFTER (CLEAN - like Azure CLI):
~/repos/jumpstart-cli ❯ js arcbox deploy
the following arguments are required: --location/-l, --resource-group/-g, --windows-user, --flavor/-f
```

**AZURE CLI CLEAN OUTPUT REFERENCE:**
```bash
# Azure CLI is CLEAN and MINIMAL on errors:
~/repos/jumpstart-cli ❯ az vm list-usage
the following arguments are required: --location/-l

# No extra help text, no usage, no descriptions - JUST the error!
```

**COMMANDS IN SCOPE** (ALL must use identical error handling):
- **subscription**: set, show, list
- **arcbox**: deploy, delete, list, preflight quota, preflight check, show  
- **repo**: list, set, show, init, update, delete
- **upgrade**: check, apply, version, install, rollback, list
- **completion**: bash, zsh, fish, powershell
- **agora**, **localbox**, **version** commands
- Any other commands and subcommands

**Key Principles**:
- ✅ Preserve ALL existing CLI functionality
- ✅ Maintain testing compatibility and mock-friendly interfaces
- ✅ Create global, reusable error handling patterns
- ✅ Eliminate hard-coded command-specific logic
- ✅ Achieve Azure CLI-like consistency
- ✅ Build and test CLI after each phase
- ✅ **PRESERVE FLAGS & EXAMPLES**: Do NOT change any existing flags, arguments, examples, or help text - only standardize error message format
- ✅ **AZURE CLI COMPARISON**: Test against actual Azure CLI commands during refactoring to ensure identical error message formatting
- ✅ **PRESERVE ROOT COMMAND HELP**: Keep our excellent UX where root commands (like `js arcbox`) automatically show help - this is BETTER than Azure CLI and should be preserved
- ✅ **FIX DUPLICATE HELP SECTIONS**: Remove duplicate command listings - show only "Subcommands:" section, not both "Subcommands:" and "Available Commands:"
- ✅ **FIX DUPLICATE USAGE LINES**: Remove duplicate usage lines like "js arcbox [flags]" and "js arcbox [command]" - show single clean usage line like Azure CLI

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
❌ "Use [command] --help" footer messages in help output (Azure CLI doesn't show these)
❌ "💡 [TIP] Use '<command>' to see all required arguments" messages (Azure CLI doesn't show these)
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

#### **5. Remove Help Footer Messages**
- Remove "Use [command] --help" messages from help output
- Azure CLI doesn't show these footer messages
- Clean, minimal help output like Azure CLI

#### **6. Remove Tip Messages**
- Remove "💡 [TIP] Use '<command>' to see all required arguments" messages
- Azure CLI doesn't show these tip messages after errors
- Clean, minimal error output like Azure CLI

#### **7. Fix Duplicate Help Sections**
- Remove duplicate command listings in help output
- Keep only "Subcommands:" section, remove redundant "Available Commands:" section
- Preserve our excellent UX where root commands automatically show help (better than Azure CLI)
- Clean up help formatting while maintaining functionality

#### **8. Clean Up Usage Lines**
- Remove duplicate usage lines like "js arcbox [flags]" and "js arcbox [command]"
- Show single, clean usage line like Azure CLI: just the command description
- Match Azure CLI's minimal, professional help format
- Eliminate redundant usage patterns that clutter the output

---

## **✅ PHASE 1 COMPLETION STATUS**

**COMPLETED SUCCESSFULLY**:
- ✅ Created `internal/utils/error_handling.go` with centralized error handling functions
- ✅ Updated `internal/utils/utils.go` to integrate with new centralized functions
- ✅ Refactored subscription set command to use the new error handling pattern
- ✅ Verified CLI builds and basic functionality works after each change
- ✅ Removed "(choose one)" text from subscription set command error messages
- ✅ All Phase 1 functions are ready and working

**READY FOR PHASE 2+**: The foundation is complete. ALL remaining commands must now be refactored to use the centralized error handling functions to achieve global consistency.

---

## **Phase 1: Foundation - Create Centralized Error Handling**

### **Prompt 1.1: Create Global Error Handling Utilities**

<!-- START PROMPT 1.1 -->
I need to create a new file `internal/utils/error_handling.go` that provides centralized, reusable error handling functions for the CLI. This must be testing-friendly and avoid hard-coded command-specific logic.

**ASSESSMENT**: Current codebase has `PrintMissingRequiredFlagsError`, `PrintMissingRequiredArgumentsTip`, and related functions in `utils.go`, but they have inconsistencies and some call `os.Exit`. We need centralized functions that work with the existing validation service pattern.

**CRITICAL GLOBAL ERROR MESSAGE REQUIREMENTS**:
This is a GLOBAL refactoring - ALL commands and subcommands (subscription, arcbox, repo, upgrade, completion, etc.) must use the exact same error message format:

- **Exact Format**: "the following arguments are required: [args]"
- **Color**: RED (using ErrorColor function) 
- **No prefixes**: Absolutely no "❌ [ERROR]" or any prefix
- **No suffixes**: Absolutely no "(choose one)" or any extra text
- **Azure CLI style**: Clean, direct, professional - exactly like Azure CLI
- **Global consistency**: Every command must produce identical error message format

**REFERENCE - EXACT AZURE CLI BEHAVIOR TO REPLICATE**:
```bash
# Azure CLI missing subcommand example:
~/repos/jumpstart-cli initial_commit ❯ az vm
the following arguments are required: _subcommand

# Azure CLI missing flag example:
~/repos/jumpstart-cli initial_commit ❯ az vm list-usage
the following arguments are required: --location/-l
```

**OUR CLI MUST PRODUCE IDENTICAL OUTPUT**:
```bash
# Our CLI examples (RED text, no prefixes/suffixes):
~/repos/jumpstart-cli ❯ js arcbox
the following arguments are required: _subcommand

~/repos/jumpstart-cli ❯ js arcbox deploy
the following arguments are required: --location/-l, --name/-n

~/repos/jumpstart-cli ❯ js subscription set
the following arguments are required: --subscription/-s, --name/-n, or positional argument
```

Requirements:
1. Create `PrintRequiredArgumentsError(missingArgs []string)` - formats exactly "the following arguments are required: [args]" with **RED color** (using ErrorColor) and **NO error prefix or suffix**
2. Create `HandleMissingRequiredArguments(cmd *cobra.Command, missingArgs []string)` - shows ONLY the error message (NO help, NO usage, NO descriptions - Azure CLI style)
3. Create `HandleValidationError(err error, cmd *cobra.Command, showHelp bool)` - generic validation error handler
4. All functions must be testing-friendly (no os.Exit calls)
5. Use existing color functions from utils package - **ErrorColor for the required arguments message**
6. Follow existing code patterns in the codebase
7. Integrate with existing ValidationResult pattern used in validation services
8. **CRITICAL**: The required arguments error message must be RED and match Azure CLI format exactly: "the following arguments are required: --flag1, --flag2" (absolutely no extra text, no prefixes, no suffixes)
9. **GLOBAL SCOPE**: This will be used by ALL commands - subscription, arcbox, repo, upgrade, completion, agora, localbox, version, etc.
10. **REMOVE HELP FOOTERS**: Ensure commands don't show "Use [command] --help" footer messages (Azure CLI doesn't show these)
11. **REMOVE TIP MESSAGES**: Don't create or show "💡 [TIP] Use '<command>' to see all required arguments" messages (Azure CLI doesn't show these)
12. **MINIMAL OUTPUT**: When showing required argument errors, show ONLY the error message - no usage, no help text, no descriptions, no examples (exactly like Azure CLI)

After creating the file, verify it compiles by running `make build`.

**TESTING VERIFICATION REQUIREMENT**: During ALL phases, compare our CLI error output against actual Azure CLI commands to ensure exact consistency:
```bash
# Test Azure CLI behavior for reference:
az vm                     # Missing subcommand
az vm list-usage          # Missing required flag

# Then test our CLI and compare:
js arcbox                 # Should match Azure CLI format exactly
js arcbox deploy          # Should match Azure CLI format exactly
```
**CRITICAL**: Preserve ALL existing flags, examples, and functionality - only change error message format to match Azure CLI style.
<!-- END PROMPT 1.1 -->

### **Prompt 1.2: Update Utils Package Integration and Cleanup**

<!-- START PROMPT 1.2 -->
I need to update the existing `internal/utils/utils.go` file to integrate with the new error handling utilities and clean up redundant functions.

**CLEANUP NEEDED**: Current `utils.go` has `PrintMissingRequiredArgumentsError()` which calls `os.Exit` - this should be updated to use the new centralized functions while maintaining backward compatibility for existing callers.

**GLOBAL SCOPE REMINDER**: This refactoring affects ALL commands and subcommands across the entire CLI:
- subscription (set, show, list)
- arcbox (deploy, delete, list, preflight quota, preflight check, show)
- repo (list, set, show)
- upgrade (check, apply, version)
- completion (bash, zsh, fish, powershell)
- agora, localbox, version commands
- All must use IDENTICAL error message format: "the following arguments are required: [args]" in RED

Requirements:

1. Review existing error handling functions: `PrintMissingRequiredFlagsError`, `PrintMissingRequiredArgumentsTip`, `PrintMissingRequiredArgumentsError`
2. Update `PrintMissingRequiredArgumentsError` to use new centralized functions instead of calling `os.Exit` directly
3. Update `PrintMissingRequiredFlagsError` if needed to work with centralized approach
4. Ensure backward compatibility - existing calls should continue to work
5. Remove any duplicate logic that's now handled by the new utilities
6. Maintain all existing function signatures to avoid breaking tests
7. Keep the existing `FatalError` function (testing-friendly) and `Fatal` function (calls os.Exit)
8. **CRITICAL**: Ensure all functions produce the exact Azure CLI error format globally

Build and verify: `make build`
Test that existing functionality still works by running a simple command like `js --help`
<!-- END PROMPT 1.2 -->

---

## **Phase 2: Subscription Commands Refactoring**

### **Prompt 2.1: Refactor Subscription Set Command**

<!-- START PROMPT 2.1 -->
I need to refactor the subscription set command in `cmd/subscription/subscription.go` to use the new centralized error handling approach.

**GLOBAL CONSISTENCY REQUIREMENT**: This command must produce the EXACT same error message format as ALL other commands in the CLI - subscription, arcbox, repo, upgrade, completion, agora, localbox, etc. The error message format must be identical across the entire codebase.

Current behavior to maintain:
- Command requires one of: --subscription/-s, --name/-n, or positional argument
- Should show clean error message: "the following arguments are required: --subscription/-s, --name/-n, or positional argument" (**RED color, NO "(choose one)" text**)
- Should show help (but NO tip messages like Azure CLI)
- Must handle mutual exclusion validation

Requirements:

1. Use the new `HandleMissingRequiredArguments` function from error_handling.go
2. Remove hard-coded error messages and use centralized approach
3. Maintain existing validation logic but **fix the error message format** - remove any "(choose one)" text
4. Preserve testing compatibility (avoid os.Exit in core logic)
5. Keep all existing functionality intact
6. **CRITICAL**: Error message must be exactly "the following arguments are required: --subscription/-s, --name/-n, or positional argument" in RED color
7. **GLOBAL CONSISTENCY**: This error format must match what ALL other commands will use

After changes:

1. Build: `make build`
2. Test the command: `js subscription set` (should show clean error)
3. Test with valid args: `js subscription set --help` (should work)
4. Verify error message format matches Azure CLI style and is consistent with global pattern
5. **AZURE CLI COMPARISON**: Compare against Azure CLI commands with missing arguments to ensure identical error format:
   ```bash
   # Test Azure CLI for reference:
   az account set                    # Missing required args
   
   # Test our CLI (should match format exactly):
   js subscription set               # Should show identical error format
   ```
6. **PRESERVE FUNCTIONALITY**: Ensure all existing flags, examples, and command behavior remain unchanged
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
2. Test deploy command: `js arcbox deploy` (should show ONLY clean error message - NO usage, NO help, NO descriptions)
3. Test with some args: `js arcbox deploy --location eastus` (should show remaining required args only)
4. Verify the command works with all required arguments
5. **CRITICAL VERIFICATION**: Error output should be MINIMAL like Azure CLI - just the red error message, nothing else
<!-- END PROMPT 3.1 -->

### **Prompt 3.2: Refactor ArcBox Delete Command**

<!-- START PROMPT 3.2 -->
I need to refactor the ArcBox delete command in `cmd/arcbox/delete_cmd.go` to use consistent error handling and eliminate any duplicate messages.

Requirements:
1. Apply centralized error handling pattern
2. Ensure consistent language: "the following arguments are required: --name/-n"
3. Remove any ❌ [ERROR] prefixes for required argument errors
4. Maintain existing validation and business logic
5. Keep testing compatibility
6. **NO tip messages**: Don't show "💡 [TIP]" messages (Azure CLI doesn't show these)

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

**✅ COMPLETED SUCCESSFULLY**:
- ✅ Updated `cmd/arcbox/preflight_cmd.go` to use centralized error handling
- ✅ Refactored `cmd/arcbox/services/quota_service.go` to use `PrintRequiredArgumentsError` and clean error handling
- ✅ Updated `internal/preflight/arcbox/rp.go` to remove ❌ [ERROR] prefixes and use centralized error handling
- ✅ Fixed duplicate error message issue in preflight quota command
- ✅ Removed tip messages from resource provider commands (no more "💡 [TIP]" messages)
- ✅ Preserved all existing business logic and timeout handling
- ✅ All preflight subcommands now use consistent Azure CLI-style error format
- ✅ Verified that help text and examples are preserved for all commands

**TESTING RESULTS**:
- ✅ `js arcbox preflight` - Shows help correctly
- ✅ `js arcbox preflight quota` - Shows clean error: "the following arguments are required: --flavor/-f, --location/-l or --all-locations"
- ✅ `js arcbox preflight quota --flavor itpro` - Shows remaining required args: "the following arguments are required: --location/-l or --all-locations"
- ✅ `js arcbox preflight rp` - Shows help correctly
- ✅ `js arcbox preflight rp register` - Shows clean error: "the following arguments are required: --name/-n"
- ✅ `js arcbox preflight rp show` - Works correctly and shows resource provider status
- ✅ `js arcbox preflight rp list` - Works correctly and shows required providers
- ✅ `js arcbox preflight status` - Works correctly and shows preflight status (removed all tip messages)
- ✅ All help text and examples are preserved (verified with --help flag)
- ✅ All error messages match Azure CLI format exactly (red color, no prefixes, clean output)
- ✅ All tip messages removed from status command to match Azure CLI minimal output style

**READY FOR NEXT PHASE**: All ArcBox preflight commands now use centralized error handling and match Azure CLI consistency requirements.
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

**✅ COMPLETED SUCCESSFULLY**:
- ✅ Updated `cmd/repo/repo.go` to use centralized error handling
- ✅ Refactored `repo delete` command to use `PrintRequiredArgumentsError` instead of hard-coded error formatting
- ✅ Removed tip messages to match Azure CLI minimal output style (no `💡 [TIP]` messages)
- ✅ Preserved all existing business logic and functionality
- ✅ All repo subcommands now use consistent Azure CLI-style error format

**TESTING RESULTS**:
- ✅ `js repo` - Shows help correctly (with both Subcommands and Available Commands sections - will be fixed in Phase 5.1)
- ✅ `js repo delete` - Shows clean error: "the following arguments are required: --force"
- ✅ `js repo delete --force` - Works correctly and shows placeholder implementation
- ✅ `js repo init` - Works correctly (no required arguments)
- ✅ `js repo update` - Works correctly (no required arguments)
- ✅ All help text and examples are preserved (verified with --help flag)
- ✅ All error messages match Azure CLI format exactly (red color, no prefixes, clean output)

**NOTE**: The original requirement mentioned "Add proper tip messages where missing" but this conflicts with our global Azure CLI-style requirements. Following our critical global requirements, tip messages were NOT added since Azure CLI doesn't show them. The goal is clean, minimal output matching Azure CLI behavior exactly.

**READY FOR NEXT PHASE**: All repo commands now use centralized error handling and match Azure CLI consistency requirements.
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

**PRESERVE EXCELLENT UX**: Our CLI has better UX than Azure CLI - when users run root commands like `js arcbox`, we automatically show help. This is EXCELLENT and should be preserved. Azure CLI just shows error messages, but our approach is more user-friendly.

**FIX DUPLICATE HELP SECTIONS**: Currently, help output shows both "Subcommands:" and "Available Commands:" sections, which is redundant and confusing. We should only show "Subcommands:" section.

**FIX DUPLICATE USAGE LINES**: Currently, help output shows multiple usage lines:
```
Usage:
  js arcbox [flags]
  js arcbox [command]
```
This should be consolidated to a single, clean usage line like Azure CLI:
```
Usage:
  js arcbox [command]
```

**CURRENT PROBLEM** (example from `js arcbox`):
```
Subcommands:
  • deploy     Deploy a new Jumpstart ArcBox deployment
  • delete     Delete a Jumpstart ArcBox deployment
  • list       List all Jumpstart ArcBox deployments
  • preflight  Run preflight checks for ArcBox deployment

Available Commands:
  delete     : Delete a Jumpstart ArcBox deployment
  deploy     : Deploy a new Jumpstart ArcBox deployment
  list       : List Jumpstart ArcBox deployments
  preflight  : Run preflight checks for ArcBox deployment
```

**TARGET CLEAN OUTPUT** (remove duplicate):
```
Subcommands:
  • deploy     Deploy a new Jumpstart ArcBox deployment
  • delete     Delete a Jumpstart ArcBox deployment
  • list       List all Jumpstart ArcBox deployments
  • preflight  Run preflight checks for ArcBox deployment
```

Requirements:
1. Ensure root command error handling uses centralized functions
2. Verify "did you mean" suggestions work consistently
3. Check that unknown command handling is standardized
4. Maintain all existing suggestion logic
5. Preserve version and help command functionality
6. **PRESERVE ROOT HELP UX**: Keep our excellent behavior where `js arcbox` shows help automatically
7. **FIX DUPLICATE SECTIONS**: Remove redundant "Available Commands:" section, keep only "Subcommands:"
8. **NO AZURE CLI COMPARISON**: Our root command help behavior is BETTER than Azure CLI - preserve it!

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
4. **Azure CLI comparison testing**: Systematically compare our CLI against Azure CLI for identical error output
5. **Functionality preservation**: All existing features should work as before
6. **Examples preservation**: Help text and examples should be intact
7. **Flag compatibility**: All existing flags and arguments must remain unchanged

**AZURE CLI COMPARISON TESTING PROTOCOL**:
For each command family, test matching Azure CLI commands and compare output:

```bash
# ARCBOX COMMANDS - Compare against Azure resource group commands:
az group create                    # Missing required args
js arcbox deploy                   # Should match error format

# SUBSCRIPTION COMMANDS - Compare against Azure account commands:  
az account set                     # Missing required args
js subscription set                # Should match error format

# COMPLETION COMMANDS - Compare against Azure completion:
az completion                      # Missing shell argument
js completion                      # Should match error format
```

**PRESERVATION VERIFICATION**:
- ✅ All flags remain exactly the same (no changes to --location, --name, etc.)
- ✅ All examples in help text remain intact
- ✅ All command functionality preserved
- ✅ All tests pass

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

### **Prompt 5.3: Enhanced Global Fix - Standardize Help Output Format**

<!-- START PROMPT 5.3 -->
I need to implement a comprehensive global fix to standardize help output format across all commands and subcommands in the entire CLI.

**PROBLEMS IDENTIFIED**:
1. **Inconsistent capitalization**: Help output shows "FLAGS" (all caps) instead of "Flags" (proper case)
2. **Duplicate Short/Long descriptions**: Many commands show redundant description lines
3. **Missing "Use --help" line removal**: Some commands may still show redundant help instructions
4. **Inconsistent section spacing**: Missing blank lines before major sections
5. **Missing blank lines around descriptions**: Command descriptions need blank lines before and after for proper formatting

**SOLUTION**: Implement a comprehensive global solution to ensure:
1. **Proper case for sections**: Change "FLAGS" to "Flags", "Global FLAGS" to "Global Flags"
2. **Single description line**: Each command shows only one concise description (use only Long descriptions)
3. **Clean section spacing**: Ensure blank lines before major sections
4. **Proper description formatting**: Ensure blank lines before AND after command descriptions
5. **Remove redundant help instructions**: No "Use <command> --help" lines

**TARGET HELP OUTPUT FORMAT**:
```
Usage:
  js arcbox deploy [flags]

Deploy a new Jumpstart ArcBox deployment using remote Bicep or ARM templates from GitHub.

Flags:
  --admin-username    : Admin username for Linux virtual machines
  --location/-l       : Azure region for deployment
```

**CRITICAL FORMATTING REQUIREMENTS**:
- **Blank line before description**: After usage line, add blank line, then description
- **Blank line after description**: After description, add blank line before next section (Flags/Subcommands)
- **Proper case sections**: "Flags" not "FLAGS", "Global Flags" not "Global FLAGS"

**IMPLEMENTATION APPROACH**:
1. Fix help output generation in `internal/utils/utils.go` to use proper case
2. Ensure all commands have Long descriptions (add where missing)
3. Remove any remaining "Use --help" instructional lines
4. Test systematically across ALL commands and subcommands

**COMPREHENSIVE VALIDATION REQUIREMENT**:
Test EVERY command and subcommand individually:

**Main Commands:**
- `js --help`
- `js version --help`
- `js completion --help`
- `js agora --help`
- `js localbox --help`

**ArcBox Commands (test ALL subcommands):**
- `js arcbox --help`
- `js arcbox deploy --help`
- `js arcbox delete --help`
- `js arcbox list --help`
- `js arcbox preflight --help`
- `js arcbox preflight quota --help`
- `js arcbox preflight rp --help`
- `js arcbox preflight status --help`

**Subscription Commands (test ALL subcommands):**
- `js subscription --help`
- `js subscription set --help`
- `js subscription show --help`
- `js subscription list --help`

**Repo Commands (test ALL subcommands):**
- `js repo --help`
- `js repo init --help`
- `js repo update --help`
- `js repo delete --help`

**Upgrade Commands (test ALL subcommands):**
- `js upgrade --help`
- `js upgrade check --help`
- `js upgrade install --help`
- `js upgrade rollback --help`
- `js upgrade list --help`

**Completion Commands (test ALL shell options):**
- `js completion bash --help`
- `js completion zsh --help`
- `js completion fish --help`
- `js completion powershell --help`

**VALIDATION CHECKLIST** (verify each item):
- [ ] All commands show "Flags" not "FLAGS"
- [ ] All commands show "Global Flags" not "Global FLAGS"
- [ ] All commands show only Long descriptions (no Short fallback)
- [ ] No commands show "Use <command> --help" lines
- [ ] Blank lines appear before major sections
- [ ] All subcommands are tested individually
- [ ] Build succeeds: `make build`
- [ ] Error handling still works: test missing args scenarios

**TARGET**: Professional, consistently formatted help output with proper case and clean format across ALL commands and subcommands.
<!-- END PROMPT 5.3 -->

### **Prompt 5.4: Global Fix - Remove "Use <command> --help" Lines**

<!-- START PROMPT 5.4 -->
I need to implement a global fix to remove all "Use <command> --help" or similar instructional lines from help output across the entire CLI.

**PROBLEM IDENTIFIED**: Help output includes lines like:
- "Use 'js arcbox <subcommand> --help' for more details."
- "Use 'js completion <shell> --help' for more information."

These lines are unnecessary and redundant since the user is already viewing the help menu. Azure CLI doesn't show these lines in help output.

**SOLUTION**: Remove all "Use <command> --help" or similar instructional lines from help output globally.

**IMPLEMENTATION APPROACH**:
1. Search for all occurrences of "Use" + "help" patterns in help output generation
2. Identify where these lines are generated (likely in help templates or command setup)
3. Remove or comment out these lines globally
4. Ensure help output is cleaner without these redundant instructions

**SEARCH PATTERNS TO FIND**:
- "Use" + command name + "help"
- "--help for more"
- "for more details"
- "for more information"

**COMMANDS TO REVIEW**:
- All help output generation logic
- Command setup in all `cmd/` directories
- Help template configurations
- Usage text generation

**VALIDATION**:
After changes:
1. Build: `make build`
2. Test help output for all commands to ensure no "Use --help" lines appear
3. Verify help output is cleaner and more focused
4. Check that essential help information is still present

**TARGET**: Clean help output without redundant "Use --help" instructions, matching Azure CLI's help format.
<!-- END PROMPT 5.4 -->

### **Prompt 5.5: Global Fix - Standardize "Global Flags" and Section Spacing**

<!-- START PROMPT 5.5 -->
I need to implement a global fix to standardize section headers and spacing in help output across all commands.

**PROBLEMS IDENTIFIED**:
1. **Inconsistent capitalization**: Help output shows "Global FLAGS" (all caps) instead of "Global Flags" (proper case)
2. **Missing blank lines**: Sometimes there's no blank line before "Subcommands:" and "Global Flags:" sections
3. **Inconsistent section formatting**: Section headers may not follow consistent formatting patterns

**SOLUTION**: Standardize help output formatting globally to ensure:
1. **Proper case**: Change "Global FLAGS" to "Global Flags"
2. **Consistent spacing**: Ensure blank lines before major sections
3. **Section header consistency**: Standardize all section header formatting

**IMPLEMENTATION APPROACH**:
1. Find where "Global FLAGS" is generated in help output
2. Change it to "Global Flags" (proper case)
3. Ensure blank lines are added before major sections:
   - Blank line before "Subcommands:"
   - Blank line before "Global Flags:"
   - Blank line before "Examples:" (if present)
4. Verify consistent formatting across all commands

**HELP OUTPUT SECTIONS TO STANDARDIZE**:
- "Subcommands:" (with blank line before)
- "Global Flags:" (with blank line before, proper case)
- "Examples:" (with blank line before, if present)
- Usage lines (ensure consistent formatting)

**VALIDATION**:
After changes:
1. Build: `make build`
2. Test help output for all commands to verify:
   - "Global Flags" appears in proper case (not "Global FLAGS")
   - Blank lines appear before major sections
   - Section headers are consistently formatted
3. Compare with Azure CLI help format for consistency reference

**TARGET**: Professional, consistently formatted help output with proper case and spacing, similar to Azure CLI's clean help format.
<!-- END PROMPT 5.5 -->

---

## **Phase 6: Test Suite Updates and Final Integration (Final Phase)**

### **Prompt 6.1: Update Test Suite for New Error Handling**

<!-- START PROMPT 6.1 -->
Now that all core functionality is working and global help output fixes have been implemented (Phases 5.3, 5.4, 5.5), I need to update the test suite to work with the new centralized error handling while maintaining all existing test coverage.

**PREREQUISITE**: Ensure Phases 5.3, 5.4, and 5.5 are completed first:
- ✅ Phase 5.3: Global fix to remove duplicate Short/Long descriptions
- ✅ Phase 5.4: Global fix to remove "Use <command> --help" lines
- ✅ Phase 5.5: Global fix to standardize "Global Flags" and section spacing

Requirements:
1. Review all test files for commands that were refactored
2. Update test expectations to match new error message formats
3. Update test expectations to match new help output formats (no duplicate descriptions, no "Use --help" lines, proper "Global Flags" case)
4. Ensure tests still validate core functionality
5. Maintain test isolation and mock compatibility
6. Add tests for new centralized error handling functions
7. Add tests for new help output formatting
8. Preserve all existing test coverage

Focus areas:
- `cmd/subscription/subscription_test.go`
- `cmd/arcbox/*_test.go` files
- `cmd/repo/*_test.go` files
- `cmd/upgrade/upgrade_test.go`
- `cmd/completion/completion_test.go`
- `internal/utils/utils_test.go`
- `internal/utils/error_handling_test.go` (if needs creation)

After changes:
1. Build: `make build`
2. Run all tests: `make test` or `go test ./...`
3. Ensure all tests pass
4. Verify test coverage is maintained
5. Check that mock interfaces still work correctly
6. Verify help output tests match new clean format
<!-- END PROMPT 6.1 -->

### **Prompt 6.2: Final Integration Testing and Documentation**

<!-- START PROMPT 6.2 -->
I need to perform final integration testing and update documentation to reflect the new consistent error handling and clean help output approach.

**PREREQUISITES**: Ensure all previous phases are completed:
- ✅ Phase 5.1: Global centralized error handling refactoring
- ✅ Phase 5.2: Global help output duplicate fixes
- ✅ Phase 5.3: Global fix to remove duplicate Short/Long descriptions
- ✅ Phase 5.4: Global fix to remove "Use <command> --help" lines
- ✅ Phase 5.5: Global fix to standardize "Global Flags" and section spacing
- ✅ Phase 6.1: Test suite updates

Requirements:
1. **Final comprehensive testing**: Test all command paths and scenarios
2. **Performance verification**: Ensure no performance regressions
3. **Error handling documentation**: Update any relevant documentation
4. **Help output documentation**: Document new clean help format
5. **Integration test verification**: Run integration tests if they exist
6. **Backward compatibility confirmation**: Ensure all existing usage patterns work

Final validation checklist:
- [ ] All commands build successfully
- [ ] All error messages follow Azure CLI format (no prefixes, red color, minimal output)
- [ ] All help output is clean (no duplicates, no "Use --help" lines, proper "Global Flags" case)
- [ ] No duplicate error messages exist
- [ ] All existing functionality preserved
- [ ] All tests pass
- [ ] Mock interfaces work correctly
- [ ] Performance is acceptable
- [ ] Examples and help text intact
- [ ] **No help footer messages**: "Use [command] --help" messages removed from all help output
- [ ] **No tip messages**: "💡 [TIP]" messages removed from all error output
- [ ] **Minimal error output**: Commands show ONLY the required arguments error message - no usage, help text, descriptions, or examples (like Azure CLI)
- [ ] **Root command help preserved**: Commands like `js arcbox` still automatically show help (our excellent UX)
- [ ] **Clean help sections**: Only "Subcommands:" section shown, "Available Commands:" section removed
- [ ] **No duplicate descriptions**: Short and Long descriptions are not duplicated in help output
- [ ] **No redundant help instructions**: All "Use <command> --help" lines removed from help output
- [ ] **Proper section formatting**: "Global Flags" appears in proper case (not "Global FLAGS")
- [ ] **Consistent section spacing**: Blank lines appear before major sections

Create a final report documenting:
- Summary of changes made across all phases
- Commands refactored with error handling improvements
- Global help output improvements implemented
- Testing results and coverage verification
- Performance impact analysis
- Any remaining considerations or future improvements
<!-- END PROMPT 6.2 -->

---

## **Key Success Criteria**

### **Functionality Preservation**
- ✅ All existing CLI commands work exactly as before
- ✅ All validation logic maintained
- ✅ All business logic preserved
- ✅ All examples and help text intact

### **Consistency Achieved**
- ✅ All required argument errors follow exact Azure CLI format: "the following arguments are required: [args]"
- ✅ All error messages displayed in RED color (no other colors for required argument errors)
- ✅ No ❌ [ERROR] prefixes for required argument errors
- ✅ No "(choose one)" or any suffix text for required argument errors
- ✅ Consistent coloring and formatting across all commands
- ✅ **EXACT REPLICATION**: Our CLI behaves identically to Azure CLI error patterns
- ✅ **CLEAN HELP OUTPUT**: No "Use [command] --help" footer messages (like Azure CLI)
- ✅ **CLEAN ERROR OUTPUT**: No "💡 [TIP]" messages after errors (like Azure CLI)
- ✅ **MINIMAL ERROR OUTPUT**: Show ONLY the required arguments error message, not full help text with usage, descriptions, and examples (like Azure CLI)

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
✅ cmd/arcbox/delete_cmd.go - Standardize error handling
✅ cmd/arcbox/list_cmd.go - Standardize subscription validation errors
✅ cmd/arcbox/preflight_cmd.go - Standardize preflight command errors
✅ internal/preflight/arcbox/rp.go - Standardize resource provider command errors
✅ cmd/repo/*.go - Add proper error handling and tips
✅ cmd/upgrade/upgrade.go - Add proper error handling and tips
✅ cmd/completion/completion.go - Standardize error handling
✅ main.go - Ensure root command uses centralized patterns
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
- [ ] All existing functionality preserved
- [ ] All tests pass
- [ ] Mock interfaces work correctly
- [ ] Performance is acceptable
- [ ] Examples and help text intact
- [ ] **No help footer messages**: "Use [command] --help" messages removed from all help output
- [ ] **No tip messages**: "💡 [TIP]" messages removed from all error output
- [ ] **Minimal error output**: Commands show ONLY the required arguments error message - no usage, help text, descriptions, or examples (like Azure CLI)
- [ ] **Root command help preserved**: Commands like `js arcbox` still automatically show help (our excellent UX)
- [ ] **Clean help sections**: Only "Subcommands:" section shown, "Available Commands:" section removed
- [ ] **No duplicate descriptions**: Short and Long descriptions are not duplicated in help output
- [ ] **No redundant help instructions**: All "Use <command> --help" lines removed from help output
- [ ] **Proper section formatting**: "Global Flags" appears in proper case (not "Global FLAGS")
- [ ] **Consistent section spacing**: Blank lines appear before major sections

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
- **Azure CLI comparison testing**: Systematically compare our CLI error output against matching Azure CLI commands to ensure identical formatting
- **Preserve all functionality**: Verify that NO flags, examples, or command behavior changes - only error message format should match Azure CLI style
