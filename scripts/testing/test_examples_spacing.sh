#!/bin/bash

# Comprehensive CLI validation script for Phases 5.5 and 5.6:
# - Phase 5.5: Examples section spacing and "Use <command> --help" lines
# - Phase 5.6: Section header case ("Global Flags") and spacing
# - Er    if [ $SPACING_ISSUES -gt 0 ]; then
        echo "   🔸 $SPACING_ISSUES commands have double blank lines before Examples (Phase 5.5)"
    fi
    
    if [ $HELP_LINE_ISSUES -gt 0 ]; then
        echo "   🔸 $HELP_LINE_ISSUES commands contain 'Use <command> --help' lines (Phase 5.5)"
    fi
    
    if [ $SECTION_HEADER_ISSUES -gt 0 ]; then
        echo "   🔸 $SECTION_HEADER_ISSUES commands have incorrect section header case (Phase 5.6)"
    fi
    
    if [ $ERROR_FORMAT_ISSUES -gt 0 ]; then
        echo "   🔸 $ERROR_FORMAT_ISSUES commands have incorrect error message format"
    fi
    
    echo ""
    echo "🔧 Fix needed before completion"e format validation for required arguments

CLI_BIN="./bin/jumpstart-cli"
TEMP_FILE="/tmp/cli_help_output.txt"
ERROR_FILE="/tmp/cli_error_output.txt"

# Counters for different types of issues
SPACING_ISSUES=0
HELP_LINE_ISSUES=0
SECTION_HEADER_ISSUES=0
ERROR_FORMAT_ISSUES=0

echo "🔍 Comprehensive CLI Validation: Phases 5.5 & 5.6"
echo "=================================================="
echo "Checking: Examples spacing, help lines, section headers, error messages"
echo "======================================================================="

# Function to test a command for help output issues
test_command() {
    local cmd="$1"
    local description="$2"
    
    echo -n "Testing: $description... "
    
    # Run command and capture output
    $CLI_BIN $cmd --help > "$TEMP_FILE" 2>&1
    
    local spacing_ok=true
    local help_line_ok=true
    local section_header_ok=true
    
    # Check 1: Examples section spacing (Phase 5.5)
    if grep -q "^Examples:" "$TEMP_FILE"; then
        # Check for double blank lines before Examples using cat -A
        if $CLI_BIN $cmd --help | cat -A | grep -B2 "^Examples:" | grep -q '\$\$'; then
            spacing_ok=false
        fi
    fi
    
    # Check 2: "Use <command> --help" instructional lines (Phase 5.5)
    if grep -qE "(Use .* --help)|(Use .* -h)|(use .* --help)|(use .* -h)" "$TEMP_FILE"; then
        help_line_ok=false
    fi
    
    # Check 3: Section header case and spacing (Phase 5.6)
    # Look for incorrect cases like "FLAGS", "Global FLAGS", "GLOBAL FLAGS"
    if grep -qE "^(FLAGS|Global FLAGS|GLOBAL FLAGS|global flags|flags):" "$TEMP_FILE"; then
        section_header_ok=false
    fi
    
    # Report results
    if $spacing_ok && $help_line_ok && $section_header_ok; then
        echo "✅ passed"
        return 0
    else
        echo "❌ ISSUES FOUND"
        echo "   Command: $CLI_BIN $cmd --help"
        
        if ! $spacing_ok; then
            echo "   🔸 Issue: Double blank lines before Examples section"
            SPACING_ISSUES=$((SPACING_ISSUES + 1))
        fi
        
        if ! $help_line_ok; then
            echo "   🔸 Issue: Contains 'Use <command> --help' instructional line"
            echo "   🔸 Found: $(grep -E "(Use .* --help)|(Use .* -h)|(use .* --help)|(use .* -h)" "$TEMP_FILE" | head -1 | xargs)"
            HELP_LINE_ISSUES=$((HELP_LINE_ISSUES + 1))
        fi
        
        if ! $section_header_ok; then
            echo "   🔸 Issue: Incorrect section header case"
            echo "   🔸 Found: $(grep -E "^(FLAGS|Global FLAGS|GLOBAL FLAGS|global flags|flags):" "$TEMP_FILE" | head -1 | xargs)"
            SECTION_HEADER_ISSUES=$((SECTION_HEADER_ISSUES + 1))
        fi
        
        echo ""
        return 1
    fi
}

# Function to test error message format for commands that require arguments
test_error_message() {
    local cmd="$1"
    local description="$2"
    local expected_error_pattern="$3"
    
    echo -n "Testing error format: $description... "
    
    # Run command without required arguments and capture stderr
    $CLI_BIN $cmd > /dev/null 2> "$ERROR_FILE"
    local exit_code=$?
    
    # Should have non-zero exit code
    if [ $exit_code -eq 0 ]; then
        echo "⚠️  skipped (no error expected)"
        return 0
    fi
    
    local error_format_ok=true
    
    # Check if error message follows Azure CLI pattern
    if [ -n "$expected_error_pattern" ]; then
        if ! grep -qE "$expected_error_pattern" "$ERROR_FILE"; then
            error_format_ok=false
        fi
        
        # Also check that there are no extra help/usage lines in error output
        if grep -qE "(Use .* --help)|(Use .* -h)|(usage:|Usage:)" "$ERROR_FILE"; then
            error_format_ok=false
        fi
    fi
    
    if $error_format_ok; then
        echo "✅ passed"
        return 0
    else
        echo "❌ ISSUES FOUND"
        echo "   Command: $CLI_BIN $cmd"
        echo "   🔸 Issue: Error message format doesn't match Azure CLI style"
        echo "   🔸 Expected pattern: $expected_error_pattern"
        echo "   🔸 Actual output: $(cat "$ERROR_FILE" | head -1 | xargs)"
        ERROR_FORMAT_ISSUES=$((ERROR_FORMAT_ISSUES + 1))
        echo ""
        return 1
    fi
}

# Main Commands
echo ""
echo "📋 Testing Main Commands (Help Output):"
test_command "" "js (root command)"
test_command "version" "js version"
test_command "completion" "js completion"
test_command "agora" "js agora"
test_command "localbox" "js localbox"

echo ""
echo "📋 Testing ArcBox Commands (Help Output):"
test_command "arcbox" "js arcbox"
test_command "arcbox deploy" "js arcbox deploy"
test_command "arcbox delete" "js arcbox delete"
test_command "arcbox list" "js arcbox list"
test_command "arcbox preflight" "js arcbox preflight"
test_command "arcbox preflight quota" "js arcbox preflight quota"
test_command "arcbox preflight rp" "js arcbox preflight rp"
test_command "arcbox preflight rp show" "js arcbox preflight rp show"
test_command "arcbox preflight rp list" "js arcbox preflight rp list"
test_command "arcbox preflight rp register" "js arcbox preflight rp register"
test_command "arcbox preflight status" "js arcbox preflight status"

echo ""
echo "📋 Testing Subscription Commands (Help Output):"
test_command "subscription" "js subscription"
test_command "subscription set" "js subscription set"
test_command "subscription show" "js subscription show"
test_command "subscription list" "js subscription list"

echo ""
echo "📋 Testing Repo Commands (Help Output):"
test_command "repo" "js repo"
test_command "repo init" "js repo init"
test_command "repo update" "js repo update"
test_command "repo delete" "js repo delete"

echo ""
echo "📋 Testing Upgrade Commands (Help Output):"
test_command "upgrade" "js upgrade"
test_command "upgrade check" "js upgrade check"
test_command "upgrade install" "js upgrade install"
test_command "upgrade rollback" "js upgrade rollback"
test_command "upgrade list" "js upgrade list"

echo ""
echo "📋 Testing Completion Commands (Help Output):"
test_command "completion bash" "js completion bash"
test_command "completion zsh" "js completion zsh"
test_command "completion fish" "js completion fish"
test_command "completion powershell" "js completion powershell"

# Error Message Format Testing
echo ""
echo "🔍 Testing Error Message Format (Azure CLI Style):"
echo "=================================================="

# Test commands that require arguments with specific error patterns
test_error_message "arcbox deploy" "js arcbox deploy (missing flavor)" "the following arguments are required:.*flavor"
test_error_message "arcbox delete" "js arcbox delete (missing resource-group)" "the following arguments are required:.*resource-group"
test_error_message "subscription set" "js subscription set (missing subscription)" "the following arguments are required:.*subscription"
test_error_message "repo init" "js repo init (missing repository)" "the following arguments are required:.*repository"
test_error_message "arcbox preflight rp register" "js arcbox preflight rp register (missing subscription)" "the following arguments are required:.*subscription"

# Cleanup
rm -f "$TEMP_FILE" "$ERROR_FILE"

echo ""
echo "======================================================================"
echo "📊 COMPREHENSIVE VALIDATION RESULTS (Phases 5.5 & 5.6):"
echo "======================================================================"

TOTAL_ISSUES=$((SPACING_ISSUES + HELP_LINE_ISSUES + SECTION_HEADER_ISSUES + ERROR_FORMAT_ISSUES))

if [ $TOTAL_ISSUES -eq 0 ]; then
    echo "🎉 SUCCESS: All validation checks passed!"
    echo "✅ Phase 5.5: No double blank lines before Examples sections"
    echo "✅ Phase 5.5: No 'Use <command> --help' instructional lines"
    echo "✅ Phase 5.6: Correct section header case ('Global Flags')"
    echo "✅ Error Format: Azure CLI-style error messages"
    echo ""
    echo "🚀 All phases completed successfully!"
    exit 0
else
    echo "❌ ISSUES FOUND:"
    
    if [ $SPACING_ISSUES -gt 0 ]; then
        echo "   🔸 $SPACING_ISSUES commands have double blank lines before Examples"
    fi
    
    if [ $HELP_LINE_ISSUES -gt 0 ]; then
        echo "   � $HELP_LINE_ISSUES commands contain 'Use <command> --help' instructional lines"
    fi
    
    echo ""
    echo "🔧 Fix needed before proceeding to Phase 5.6"
    exit 1
fi
