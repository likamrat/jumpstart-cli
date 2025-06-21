package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// PrintRequiredArgumentsError formats and prints the Azure CLI-style required arguments error message
// This formats "the following arguments are required: [args]" with NO error prefix in RED color
func PrintRequiredArgumentsError(missingArgs []string) {
	if len(missingArgs) > 0 {
		message := "the following arguments are required: " + strings.Join(missingArgs, ", ")
		fmt.Fprintf(os.Stderr, "%s\n\n", ErrorColor(message))
	}
}

// HandleMissingRequiredArguments provides the centralized required arguments error handling
// This matches Azure CLI behavior: clean error message only, no tips or help footers
func HandleMissingRequiredArguments(cmd *cobra.Command, missingArgs []string) {
	// Print the clean error message (no error prefix, no tips, no help footer)
	PrintRequiredArgumentsError(missingArgs)
}

// HandleValidationError provides generic validation error handling
// This works with the existing ValidationResult pattern used in validation services
func HandleValidationError(err error, cmd *cobra.Command, showHelp bool) {
	if err == nil {
		return
	}

	// Check if this is a "validation_failed_with_output_already_shown" special case
	// This is used by validation services to indicate they've already displayed the error
	if err.Error() == "validation_failed_with_output_already_shown" {
		PrintStandardHelpTip(cmd)
		return
	}

	// Print the error message
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())

	// Show help if requested
	if showHelp {
		fmt.Fprintf(os.Stderr, "\n")
		ShowHelpWithoutTypes(cmd)
		PrintStandardHelpTip(cmd)
	}
}

// HandleRequiredFlagsValidation handles the specific case of required flag validation
// This integrates with the existing PrintMissingRequiredFlagsError pattern
func HandleRequiredFlagsValidation(cmd *cobra.Command, requiredFlags []string) bool {
	// Use the existing PrintMissingRequiredFlagsError function for validation
	isValid := PrintMissingRequiredFlagsError(cmd, requiredFlags)

	if !isValid {
		// Add the standard tip message
		PrintStandardHelpTip(cmd)
	}

	return isValid
}

// HandleSubscriptionSelectionError handles subscription selection validation errors
// This provides a centralized way to handle subscription-related validation
func HandleSubscriptionSelectionError(cmd *cobra.Command, errorMessage string) {
	fmt.Fprintf(os.Stderr, "%s\n\n", errorMessage)
	ShowHelpWithoutTypes(cmd)
	PrintStandardHelpTip(cmd)
}
