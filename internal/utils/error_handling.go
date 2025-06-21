package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// PrintRequiredArgumentsError formats and prints the Azure CLI-style required arguments error message
// This formats exactly "the following arguments are required: [args]" with RED color and NO prefix/suffix
// CRITICAL: Matches Azure CLI format exactly - clean, minimal, professional
func PrintRequiredArgumentsError(missingArgs []string) {
	if len(missingArgs) > 0 {
		message := "the following arguments are required: " + strings.Join(missingArgs, ", ")
		// Azure CLI style: RED color, single newline, NO extra text
		fmt.Fprintf(os.Stderr, "%s\n", ErrorColor(message))
	}
}

// HandleMissingRequiredArguments provides centralized required arguments error handling
// This exactly matches Azure CLI behavior: ONLY the error message, no help, no usage, no descriptions
// CRITICAL: Minimal output like Azure CLI - just the red error message, nothing else
func HandleMissingRequiredArguments(cmd *cobra.Command, missingArgs []string) {
	// Azure CLI style: Print ONLY the clean error message
	// NO help footers, NO tip messages, NO usage, NO descriptions
	PrintRequiredArgumentsError(missingArgs)
}

// HandleValidationError provides generic validation error handling
// This works with the existing ValidationResult pattern used in validation services
// TESTING-FRIENDLY: No os.Exit calls, returns errors for testing
func HandleValidationError(err error, cmd *cobra.Command, showHelp bool) {
	if err == nil {
		return
	}

	// Print the error message (no special formatting, let the error speak for itself)
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())

	// Show help if explicitly requested (for backward compatibility only)
	// FUTURE: This should be removed in later phases to match Azure CLI minimal output
	if showHelp {
		fmt.Fprintf(os.Stderr, "\n")
		ShowHelpWithoutTypes(cmd)
	}
}

// HandleRequiredFlagsValidation handles the specific case of required flag validation
// This integrates with the existing PrintMissingRequiredFlagsError pattern but uses Azure CLI style
// TESTING-FRIENDLY: Returns validation result instead of calling os.Exit
func HandleRequiredFlagsValidation(cmd *cobra.Command, requiredFlags []string) bool {
	// Use the existing PrintMissingRequiredFlagsError function for validation
	// This already implements the Azure CLI style error message format
	return PrintMissingRequiredFlagsError(cmd, requiredFlags)
}

// HandleSubscriptionSelectionError handles subscription selection validation errors
// This provides a centralized way to handle subscription-related validation with Azure CLI style
// MINIMAL OUTPUT: Just the error message in RED color, no extra help text (matching Azure CLI behavior)
func HandleSubscriptionSelectionError(cmd *cobra.Command, errorMessage string) {
	// Azure CLI style: Error message in RED color with single newline, no prefixes
	fmt.Fprintf(os.Stderr, "%s\n", ErrorColor(errorMessage))
}

// PrintAuthenticationError prints a minimal authentication error message
// This matches Azure CLI's authentication error style
func PrintAuthenticationError() {
	fmt.Fprintf(os.Stderr, "Please run 'az login' to setup account.\n")
}
