package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"jumpstartcli/internal/examples"
	"jumpstartcli/internal/table"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

var (
	InfoColor    = color.New(color.FgCyan).SprintFunc()
	WarnColor    = color.New(color.FgYellow).SprintFunc()
	ErrorColor   = color.New(color.FgRed).SprintFunc()
	SuccessColor = color.New(color.FgGreen).SprintFunc()
	DebugColor   = color.New(color.FgHiBlack).SprintFunc()
	FatalColor   = color.New(color.FgHiRed, color.Bold).SprintFunc()
	PromptColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

var DebugMode bool
var VerboseMode bool
var OutputFormat string

// CliVersion can be set at build time via ldflags
var CliVersion = "0.1.0"

// IsAzureLoggedIn checks if the user is logged in to Azure CLI
func IsAzureLoggedIn() bool {
	cmd := exec.Command("az", "account", "show")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func ResourceGroupExists(name string) bool {
	cmd := exec.Command("az", "group", "exists", "--name", name)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error checking resource group existence: %v\n", err)
		return false
	}
	return string(output) == "true\n"
}

func CreateResourceGroup(name string, location string) error {
	// Only print a single info message before creation, no spinner, no JSON output
	cmd := exec.Command("az", "group", "create", "--name", name, "--location", location)
	cmd.Stdout = nil // Suppress all output
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		fmt.Printf("[ERROR] Failed to create resource group: %v\n", err)
		return err
	}
	return nil
}

func Info(msg string, args ...interface{}) {
	fmt.Println(InfoColor(fmt.Sprintf("[INFO] "+msg, args...)))
}

func Warn(msg string, args ...interface{}) {
	fmt.Println(WarnColor(fmt.Sprintf("[WARN] "+msg, args...)))
}

func Error(msg string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, ErrorColor(fmt.Sprintf("[ERROR] "+msg, args...)))
}

func Success(msg string, args ...interface{}) {
	fmt.Println(SuccessColor("[SUCCESS]"), fmt.Sprintf(msg, args...))
}

func Debug(msg string, args ...interface{}) {
	if DebugMode {
		fmt.Println(DebugColor("[DEBUG]"), fmt.Sprintf(msg, args...))
	}
}

func Fatal(msg string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, FatalColor("[FATAL]"), fmt.Sprintf(msg, args...))
	os.Exit(1)
}

func Prompt(msg string, args ...interface{}) {
	fmt.Println(PromptColor("[PROMPT]"), fmt.Sprintf(msg, args...))
}

// FriendlyResourceName returns a user-friendly resource type name, with special handling for VM Extensions
func FriendlyResourceName(resourceType, resourceName string) string {
	switch resourceType {
	case "Microsoft.OperationalInsights/workspaces":
		return "Log Analytics workspace"
	case "Microsoft.OperationsManagement/solutions":
		return "Operations Management solution"
	case "Microsoft.Network/networkSecurityGroups":
		return "Network Security Group"
	case "Microsoft.KeyVault/vaults":
		return "Azure Key Vault"
	case "Microsoft.Network/virtualNetworks":
		return "Virtual Network"
	case "Microsoft.Compute/disks":
		return "Disk"
	case "Microsoft.Network/publicIPAddresses":
		return "Public IP Address"
	case "Microsoft.Network/networkInterfaces":
		return "Network Interface"
	case "Microsoft.Compute/virtualMachines":
		return "Virtual Machine"
	case "Microsoft.Resources/deployments":
		return "Nested Deployment"
	case "Microsoft.DevTestLab/schedules":
		return "Schedule"
	case "Microsoft.Compute/virtualMachines/extensions":
		// For VM Extensions, extract the extension name from the resourceName (after last '/')
		parts := strings.Split(resourceName, "/")
		extName := parts[len(parts)-1]
		switch extName {
		case "Microsoft.Azure.Geneva.GenevaMonitoring":
			return "Azure Geneva Monitoring"
		case "Bootstrap":
			return "Bootstrap"
		default:
			return extName // fallback to extension name
		}
	default:
		// Fallback: try to prettify the type
		parts := strings.Split(resourceType, "/")
		if len(parts) > 1 {
			// Remove trailing 's' for plural, e.g. disks -> Disk
			name := parts[1]
			if strings.HasSuffix(name, "s") && len(name) > 1 {
				name = name[:len(name)-1]
			}
			// Capitalize first letter (avoid deprecated strings.Title)
			if len(name) > 0 {
				name = strings.ToUpper(name[:1]) + name[1:]
			}
			return strings.ReplaceAll(name, "_", " ")
		}
		return resourceType
	}
}

// flagInfo represents information about a command flag
type flagInfo struct {
	name        string
	shorthand   string
	description string
	defaultVal  string
	isRequired  bool
}

// PrintMissingRequiredFlagsError prints a dynamic error message listing all missing required flags for a command, Azure CLI style.
func PrintMissingRequiredFlagsError(cmd *cobra.Command, requiredFlags []string) {
	missing := []string{}
	for _, name := range requiredFlags {
		f := cmd.Flags().Lookup(name)
		if f == nil || isFlagMissing(cmd, f) {
			flagStr := "--" + name
			if f != nil && f.Shorthand != "" {
				flagStr += "/-" + f.Shorthand
			}
			missing = append(missing, flagStr)
		}
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "%s\n\n", ErrorColor("the following arguments are required: "+strings.Join(missing, ", ")))
		ShowHelpWithoutTypes(cmd)
		os.Exit(1)
	}
}

// isFlagMissing returns true if the required flag is missing (unset or empty string)
func isFlagMissing(cmd *cobra.Command, f *pflag.Flag) bool {
	if !f.Changed {
		return true
	}
	val := f.Value.String()
	return val == "" || val == "[]" // handle string and string slice
}

// ShowHelpWithoutTypes displays help for a command with type annotations removed and [Required] text added
func ShowHelpWithoutTypes(cmd *cobra.Command) {
	// Build custom help output
	output := buildCustomHelpOutput(cmd)
	fmt.Print(output)
}

// buildCustomHelpOutput builds help output with proper text wrapping
func buildCustomHelpOutput(cmd *cobra.Command) string {
	var result strings.Builder
	terminalWidth := 160 // Maximized width for optimal description formatting

	// Usage section
	result.WriteString("Usage:\n")
	if cmd.Runnable() {
		result.WriteString("  " + cmd.UseLine() + "\n")
	}
	if cmd.HasAvailableSubCommands() {
		result.WriteString("  " + cmd.CommandPath() + " [command]\n")
	}

	// Available Commands section
	if cmd.HasAvailableSubCommands() {
		result.WriteString("\nAvailable Commands:\n")

		// Calculate maximum command name width for consistent alignment
		maxCmdWidth := 0
		availableCommands := []*cobra.Command{}
		for _, subCmd := range cmd.Commands() {
			if subCmd.IsAvailableCommand() {
				availableCommands = append(availableCommands, subCmd)
				if len(subCmd.Name()) > maxCmdWidth {
					maxCmdWidth = len(subCmd.Name())
				}
			}
		}

		// Use actual maximum width with some padding for visual spacing
		// Add 2 extra spaces for better readability
		maxCmdWidth += 2

		// Format each command with consistent spacing (single colon)
		for _, subCmd := range availableCommands {
			result.WriteString(fmt.Sprintf("  %-*s: %s\n", maxCmdWidth, subCmd.Name(), subCmd.Short))
		}
	}

	// Local Flags section
	if cmd.HasAvailableLocalFlags() {
		result.WriteString("\nArguments:\n")
		result.WriteString(buildFlagsOutput(cmd, cmd.LocalFlags(), terminalWidth))
	}

	// Global Flags section
	if cmd.HasAvailableInheritedFlags() {
		result.WriteString("\nGlobal Arguments:\n")
		result.WriteString(buildFlagsOutput(cmd, cmd.InheritedFlags(), terminalWidth))
	}

	// Examples section
	commandKey := strings.ReplaceAll(cmd.CommandPath(), " ", ".")
	if exampleSet := examples.GetExamples(commandKey); len(exampleSet.Examples) > 0 {
		result.WriteString("\n")
		result.WriteString(exampleSet.FormatExamples())
	}

	// Additional help section
	if cmd.HasAvailableSubCommands() {
		result.WriteString("\nUse \"" + cmd.CommandPath() + " [command] --help\" for more information about a command.\n")
	}

	return result.String()
}

// buildFlagsOutput builds properly wrapped flag output with Azure CLI-style formatting
func buildFlagsOutput(cmd *cobra.Command, flagSet *pflag.FlagSet, terminalWidth int) string {
	var result strings.Builder

	// Collect flags and their descriptions
	var flags []flagInfo
	flagSet.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		info := flagInfo{
			name:        flag.Name,
			shorthand:   flag.Shorthand,
			description: flag.Usage,
			defaultVal:  flag.DefValue,
			isRequired:  isRequiredFlag(cmd, flag.Name),
		}
		flags = append(flags, info)
	})

	// Sort flags alphabetically by name for consistent display
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].name < flags[j].name
	})

	// Calculate maximum flag width (first column)
	maxFlagWidth := 0
	for _, flag := range flags {
		flagStr := buildFlagString(flag)
		width := len(stripAnsiCodes(flagStr))
		if width > maxFlagWidth {
			maxFlagWidth = width
		}
	}

	// Ensure minimum width for proper alignment
	if maxFlagWidth < 40 {
		maxFlagWidth = 40
	}

	// Format each flag line using Azure CLI style with proper text wrapping
	for _, flag := range flags {
		flagStr := buildFlagString(flag)

		// Build description with default values
		description := buildDescriptionWithDefaults(flag)

		// Calculate available width for description (accounting for columns and margins)
		flagColumn := fmt.Sprintf("    %-*s", maxFlagWidth, flagStr)
		requiredColumn := ""
		if flag.isRequired {
			requiredTag := color.New(color.FgYellow, color.Bold).Sprint("[Required]")
			requiredColumn = " " + requiredTag
		} else {
			requiredColumn = "           " // 11 spaces to match "[Required]" length
		}

		prefixLength := len(stripAnsiCodes(flagColumn + requiredColumn + " : "))
		availableWidth := terminalWidth - prefixLength - 1 // Minimal margin for maximum description width
		if availableWidth < 120 {
			availableWidth = 120 // Maximized minimum width for optimal single-line descriptions
		}

		// Wrap description text if needed
		wrappedLines := wrapTextAtWords(description, availableWidth)

		// Output first line
		result.WriteString(flagColumn + requiredColumn + " : " + wrappedLines[0] + "\n")

		// Output continuation lines with proper indentation
		if len(wrappedLines) > 1 {
			indentStr := strings.Repeat(" ", prefixLength)
			for i := 1; i < len(wrappedLines); i++ {
				result.WriteString(indentStr + wrappedLines[i] + "\n")
			}
		}
	}

	return result.String()
}

// buildFlagString creates the flag name portion (e.g., "-f, --flavor")
func buildFlagString(flag flagInfo) string {
	if flag.shorthand != "" {
		return "-" + flag.shorthand + ", --" + flag.name
	}
	return "    --" + flag.name
}

// buildDescriptionWithDefaults creates description with default values displayed
func buildDescriptionWithDefaults(flag flagInfo) string {
	description := flag.description

	// Convert argument references to italic format (e.g., auto-shutdown-enabled -> auto-shutdown-enabled)
	description = formatArgumentReferences(description)

	// Use colored and italic default value - Blue color with italic styling
	defaultColor := color.New(color.FgBlue, color.Italic).SprintFunc()
	allowedValuesColor := color.New(color.FgCyan).SprintFunc()

	// Add "Allowed values" callout for string flags that accept yes/no values (like Azure CLI)
	if isBooleanStringFlag(flag.name, flag.defaultVal) {
		description += " " + allowedValuesColor("Allowed values: yes, no")
	} else if isEnum, enumValues := isEnumStringFlag(flag.name, flag.defaultVal); isEnum {
		description += " " + allowedValuesColor("Allowed values: "+enumValues)
	} else if isPathFlag(flag.name) {
		description += " " + allowedValuesColor("Expects: local file path")
	} else if isURIFlag(flag.name) {
		description += " " + allowedValuesColor("Expects: remote URI")
	}

	// Handle boolean flags specially to show yes/no instead of true/false
	if flag.defaultVal == "true" {
		description += " (default: " + defaultColor("yes") + ")"
	} else if flag.defaultVal == "false" {
		// For boolean flags, show "no" instead of hiding the default
		// Check if this is likely a boolean flag by looking for common boolean flag patterns
		if isBooleanFlag(flag.name) {
			description += " (default: " + defaultColor("no") + ")"
		}
		// For non-boolean flags with false values, don't show default (maintain existing behavior)
	} else if flag.defaultVal != "" && flag.defaultVal != "[]" {
		// Handle non-boolean flags with non-empty defaults
		description += " (default: " + defaultColor(flag.defaultVal) + ")"
	}

	return description
}

// isBooleanFlag checks if a flag name suggests it's a boolean flag
func isBooleanFlag(flagName string) bool {
	// Common boolean flag patterns
	booleanPatterns := []string{
		"enabled", "disabled", "enable", "disable",
		"deploy", "skip", "auto", "vm-",
		"bastion", "preflight",
	}

	flagLower := strings.ToLower(flagName)
	for _, pattern := range booleanPatterns {
		if strings.Contains(flagLower, pattern) {
			return true
		}
	}
	return false
}

// isBooleanStringFlag checks if a flag is a string-based boolean flag that accepts yes/no values
func isBooleanStringFlag(flagName, defaultVal string) bool {
	// Check if this is a known boolean string flag
	knownBooleanStringFlags := []string{
		"auto-shutdown",
		"deploy-bastion",
		"skip-preflight",
		"vm-autologon",
	}

	for _, knownFlag := range knownBooleanStringFlags {
		if flagName == knownFlag {
			return true
		}
	}

	// Also check if default value suggests boolean (yes/no)
	if defaultVal == "yes" || defaultVal == "no" {
		return true
	}

	return false
}

// isEnumStringFlag checks if a flag has specific allowed values that should be displayed
func isEnumStringFlag(flagName, defaultVal string) (bool, string) {
	// Map of flag names to their allowed values
	enumFlags := map[string]string{
		"flavor":             "ITPro, DevOps, DataOps",
		"sql-server-edition": "Developer, Standard, Enterprise",
		"bastion-sku":        "Basic, Standard, Developer",
	}

	if allowedValues, exists := enumFlags[flagName]; exists {
		return true, allowedValues
	}

	return false, ""
}

// isPathFlag checks if a flag expects a local file path
func isPathFlag(flagName string) bool {
	// Flags that expect local file paths
	pathFlags := []string{
		"template-local",
		"template-params",
	}

	for _, pathFlag := range pathFlags {
		if flagName == pathFlag {
			return true
		}
	}

	return false
}

// isURIFlag checks if a flag expects a remote URI
func isURIFlag(flagName string) bool {
	// Flags that expect remote URIs
	uriFlags := []string{
		"template-uri",
	}

	for _, uriFlag := range uriFlags {
		if flagName == uriFlag {
			return true
		}
	}

	return false
}

// formatArgumentReferences converts argument references in descriptions to italic format
func formatArgumentReferences(description string) string {
	// Pattern to match argument references (hyphenated words that look like CLI flags)
	// This matches words like "auto-shutdown-enabled", "resource-group", etc.
	argPattern := regexp.MustCompile(`\b([a-z]+-[a-z-]+)\b`)

	// Common argument names that should be italicized when mentioned in descriptions
	commonArgs := map[string]bool{
		"auto-shutdown-enabled":   true,
		"auto-shutdown-time":      true,
		"bastion-sku":             true,
		"deploy-bastion":          true,
		"resource-group":          true,
		"windows-user":            true,
		"windows-password":        true,
		"naming-prefix":           true,
		"rdp-port":                true,
		"log-analytics-workspace": true,
		"resource-tags":           true,
		"ssh-rsa-public-key":      true,
		"sql-server-edition":      true,
		"template-local":          true,
		"template-params":         true,
		"template-uri":            true,
		"github-user":             true,
		"skip-preflight":          true,
		"vm-autologon":            true,
	}

	italicColor := color.New(color.Italic).SprintFunc()

	return argPattern.ReplaceAllStringFunc(description, func(match string) string {
		if commonArgs[match] {
			return italicColor(match)
		}
		return match
	})
}

// wrapTextAtWords wraps text at word boundaries
func wrapTextAtWords(text string, maxWidth int) []string {
	// Remove color codes for width calculation
	cleanText := stripAnsiCodes(text)
	if len(cleanText) <= maxWidth {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	var lines []string
	var currentLine strings.Builder
	var currentLength int

	for _, word := range words {
		cleanWord := stripAnsiCodes(word)
		wordLen := len(cleanWord)

		// Check if adding this word would exceed maxWidth
		if currentLength > 0 && currentLength+1+wordLen > maxWidth {
			// Start new line
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
			currentLength = wordLen
		} else {
			// Add to current line
			if currentLength > 0 {
				currentLine.WriteString(" ")
				currentLength++
			}
			currentLine.WriteString(word)
			currentLength += wordLen
		}
	}

	// Add final line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// stripAnsiCodes removes ANSI color codes for accurate length calculation
func stripAnsiCodes(text string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(text, "")
}

// isRequiredFlag checks if a flag is required for the given command
func isRequiredFlag(cmd *cobra.Command, flagName string) bool {
	// Check if flag exists
	flag := cmd.Flags().Lookup(flagName)
	if flag == nil {
		return false
	}

	// Check if the flag has required annotations
	if flag.Annotations != nil {
		if _, exists := flag.Annotations[cobra.BashCompOneRequiredFlag]; exists {
			return true
		}
	}

	// Fallback: check against hardcoded required flags for specific commands
	return isHardcodedRequiredFlag(cmd, flagName)
}

// isHardcodedRequiredFlag checks against hardcoded lists of required flags for specific commands
func isHardcodedRequiredFlag(cmd *cobra.Command, flagName string) bool {
	cmdPath := cmd.CommandPath()

	// Map command paths to their required flags
	// Note: Conditionally required flags (like github-user, ssh-rsa-public-key) are handled in runtime validation
	requiredFlagsByCommand := map[string][]string{
		"js arcbox deploy":                {"location", "resource-group", "windows-user", "flavor"},
		"js arcbox delete":                {"name"},
		"js arcbox preflight quota":       {"flavor"}, // location OR all-locations are required, but handled in runtime validation
		"js arcbox preflight rp register": {"name"},
	}

	if requiredFlags, exists := requiredFlagsByCommand[cmdPath]; exists {
		for _, required := range requiredFlags {
			if required == flagName {
				return true
			}
		}
	}

	return false
}

// NormalizeRegion normalizes Azure region names to lowercase without spaces
func NormalizeRegion(region string) string {
	return strings.ToLower(strings.ReplaceAll(region, " ", ""))
}

// PrintMissingRequiredArgumentsError prints an error message for missing required arguments
// This is an alias for PrintMissingRequiredFlagsError to maintain compatibility
func PrintMissingRequiredArgumentsError(cmd *cobra.Command, requiredArgs []string) {
	PrintMissingRequiredFlagsError(cmd, requiredArgs)
}

// SuggestSimilarCommand suggests similar commands based on edit distance
func SuggestSimilarCommand(input string, commands []string, threshold int) string {
	bestMatch := ""
	minDistance := threshold + 1

	for _, cmd := range commands {
		distance := levenshteinDistance(input, cmd)
		if distance < minDistance {
			minDistance = distance
			bestMatch = cmd
		}
	}

	if minDistance <= threshold {
		return bestMatch
	}
	return ""
}

// PrintDidYouMean prints a "did you mean" suggestion
func PrintDidYouMean(invalid string, suggestion string) {
	fmt.Printf("%s unknown command '%s'. Did you mean '%s'?\n",
		ErrorColor("[ERROR]"), invalid, suggestion)
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

// GetRegionDisplayName returns the display name for an Azure region
func GetRegionDisplayName(region string) string {
	// Common Azure region mappings
	regionMap := map[string]string{
		"eastus":             "East US",
		"eastus2":            "East US 2",
		"westus":             "West US",
		"westus2":            "West US 2",
		"westus3":            "West US 3",
		"centralus":          "Central US",
		"northcentralus":     "North Central US",
		"southcentralus":     "South Central US",
		"westcentralus":      "West Central US",
		"canadacentral":      "Canada Central",
		"canadaeast":         "Canada East",
		"brazilsouth":        "Brazil South",
		"northeurope":        "North Europe",
		"westeurope":         "West Europe",
		"uksouth":            "UK South",
		"ukwest":             "UK West",
		"francecentral":      "France Central",
		"francesouth":        "France South",
		"germanywestcentral": "Germany West Central",
		"norwayeast":         "Norway East",
		"switzerlandnorth":   "Switzerland North",
		"eastasia":           "East Asia",
		"southeastasia":      "Southeast Asia",
		"australiaeast":      "Australia East",
		"australiasoutheast": "Australia Southeast",
		"japaneast":          "Japan East",
		"japanwest":          "Japan West",
		"koreacentral":       "Korea Central",
		"koreasouth":         "Korea South",
		"centralindia":       "Central India",
		"southindia":         "South India",
		"westindia":          "West India",
		"southafricanorth":   "South Africa North",
		"uaenorth":           "UAE North",
	}

	normalized := NormalizeRegion(region)
	if displayName, exists := regionMap[normalized]; exists {
		return displayName
	}

	// If not found in map, return a formatted version
	return strings.ToTitle(strings.ReplaceAll(region, " ", " "))
}

// RegionExistsInAzure checks if a region exists in Azure using Azure CLI
func RegionExistsInAzure(region string) bool {
	cmd := exec.Command("az", "account", "list-locations", "--query", "[?name=='"+NormalizeRegion(region)+"'].name", "--output", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	result := strings.TrimSpace(string(output))
	return result != ""
}

// GetBooleanFlagValue handles the global yes/no pattern for boolean flags
// Supports both positive flags (--flag-name) and negative flags (--no-flag-name)
// Also handles string-based yes/no values for flags without negative counterparts
// Returns the final boolean value after applying the yes/no logic
func GetBooleanFlagValue(cmd *cobra.Command, flagName string) bool {
	flag := cmd.Flags().Lookup(flagName)
	if flag == nil {
		return false
	}

	// Check if there's a corresponding --no-{flag-name} flag
	noFlagName := "no-" + flagName
	noFlag := cmd.Flags().Lookup(noFlagName)

	if noFlag != nil {
		// Pattern 1: Positive + Negative flags (e.g., --auto-shutdown-enabled + --no-auto-shutdown)
		positiveValue, _ := cmd.Flags().GetBool(flagName)
		noValue, _ := cmd.Flags().GetBool(noFlagName)
		return positiveValue && !noValue
	} else {
		// Pattern 2: Single flag that can accept yes/no string values
		if flag.Value.Type() == "string" {
			// String flag - parse yes/no values
			stringValue, _ := cmd.Flags().GetString(flagName)
			if stringValue != "" {
				if parsed, err := ParseYesNoToBool(stringValue); err == nil {
					return parsed
				} else {
					// Invalid boolean value - this should have been caught by early validation
					// Return false as fallback (early validation should prevent reaching here)
					return false
				}
			}
			// Default value for string flags
			return false
		} else {
			// Boolean flag - handle normally
			value, _ := cmd.Flags().GetBool(flagName)
			return value
		}
	}
}

// ParseYesNoToBool converts yes/no strings to boolean values
// Supports: yes, y, true, 1 → true | no, n, false, 0 → false
// Case-insensitive parsing
func ParseYesNoToBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "yes", "y", "true", "1", "on", "enable", "enabled":
		return true, nil
	case "no", "n", "false", "0", "off", "disable", "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s (expected: yes/no, y/n, true/false, 1/0)", value)
	}
}

// RegisterBooleanFlagPair registers both positive and negative versions of a boolean flag
// Example: RegisterBooleanFlagPair(cmd, "auto-shutdown-enabled", "n", true, "Enable automatic VM shutdown")
func RegisterBooleanFlagPair(cmd *cobra.Command, flagName, shortFlag string, defaultValue bool, description string) {
	// Register the positive flag
	if shortFlag != "" {
		cmd.Flags().BoolP(flagName, shortFlag, defaultValue, description)
	} else {
		cmd.Flags().Bool(flagName, defaultValue, description)
	}

	// Register the negative flag (--no-{flag-name})
	noFlagName := "no-" + flagName
	noDescription := fmt.Sprintf("Disable %s", strings.ToLower(description))
	cmd.Flags().Bool(noFlagName, false, noDescription)
}

// ValidateAllFlags performs comprehensive validation of all command flags
// This should be called early in command execution, before any expensive operations
func ValidateAllFlags(cmd *cobra.Command) error {
	var validationErrors []string

	// Get all flags defined for this command
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if !flag.Changed {
			// Skip validation for flags that weren't provided by user
			return
		}

		flagName := flag.Name
		flagValue := flag.Value.String()
		flagType := flag.Value.Type()

		// Validate based on flag type
		switch flagType {
		case "string":
			if err := validateStringFlag(cmd, flagName, flagValue); err != nil {
				validationErrors = append(validationErrors, err.Error())
			}
		case "bool":
			if err := validateBooleanFlag(cmd, flagName, flagValue); err != nil {
				validationErrors = append(validationErrors, err.Error())
			}
		case "int", "int32", "int64":
			if err := validateIntFlag(cmd, flagName, flagValue); err != nil {
				validationErrors = append(validationErrors, err.Error())
			}
		case "float32", "float64":
			if err := validateFloatFlag(cmd, flagName, flagValue); err != nil {
				validationErrors = append(validationErrors, err.Error())
			}
		case "stringSlice":
			if err := validateStringSliceFlag(cmd, flagName, flagValue); err != nil {
				validationErrors = append(validationErrors, err.Error())
			}
		}
	})

	// If we have validation errors, report them and exit
	if len(validationErrors) > 0 {
		fmt.Printf("Error: Invalid flag values detected:\n")
		for _, err := range validationErrors {
			fmt.Printf("  %s\n", err)
		}
		return fmt.Errorf("flag validation failed")
	}

	return nil
}

// validateStringFlag validates string-type flags with special handling for boolean-like strings
func validateStringFlag(cmd *cobra.Command, flagName, flagValue string) error {
	// Check if this is a boolean-disguised-as-string flag by looking at the usage
	flag := cmd.Flags().Lookup(flagName)
	if flag == nil {
		return nil
	}

	// Look for boolean indicators in the usage text
	usage := strings.ToLower(flag.Usage)
	if strings.Contains(usage, "(yes/no)") || strings.Contains(usage, "yes/no") {
		// This is a boolean flag disguised as string - validate yes/no format
		if _, err := ParseYesNoToBool(flagValue); err != nil {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Expected: yes/no (also accepts y/n, true/false, 1/0)", flagValue, flagName)
		}
	}

	// Add other string validation rules here
	// For example, validate specific enum values, length constraints, etc.
	if err := validateSpecificStringFlag(flagName, flagValue); err != nil {
		return err
	}

	return nil
}

// validateBooleanFlag validates traditional boolean flags
func validateBooleanFlag(cmd *cobra.Command, flagName, flagValue string) error {
	// Traditional bool flags should only accept true/false
	if flagValue != "true" && flagValue != "false" {
		return fmt.Errorf("Invalid value '%s' for boolean flag --%s. Expected: true or false", flagValue, flagName)
	}
	return nil
}

// validateIntFlag validates integer flags
func validateIntFlag(cmd *cobra.Command, flagName, flagValue string) error {
	if flagValue == "" {
		return nil
	}

	// Parse as int to validate
	if _, err := strconv.Atoi(flagValue); err != nil {
		return fmt.Errorf("Invalid value '%s' for integer flag --%s. Expected: valid integer", flagValue, flagName)
	}

	// Add specific integer range validation
	if err := validateSpecificIntFlag(flagName, flagValue); err != nil {
		return err
	}

	return nil
}

// validateFloatFlag validates float flags
func validateFloatFlag(cmd *cobra.Command, flagName, flagValue string) error {
	if flagValue == "" {
		return nil
	}

	// Parse as float to validate
	if _, err := strconv.ParseFloat(flagValue, 64); err != nil {
		return fmt.Errorf("Invalid value '%s' for numeric flag --%s. Expected: valid number", flagValue, flagName)
	}

	return nil
}

// validateStringSliceFlag validates string slice flags
func validateStringSliceFlag(cmd *cobra.Command, flagName, flagValue string) error {
	// String slices are usually comma-separated values
	// Basic validation - could be enhanced based on specific flag requirements
	if flagValue == "" {
		return nil
	}

	// Add specific string slice validation
	if err := validateSpecificStringSliceFlag(flagName, flagValue); err != nil {
		return err
	}

	return nil
}

// validateSpecificStringFlag validates specific string flags with custom rules
func validateSpecificStringFlag(flagName, flagValue string) error {
	switch flagName {
	case "flavor":
		validFlavors := []string{"ITPro", "DevOps", "DataOps"}
		if !containsCaseInsensitive(validFlavors, flagValue) {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Expected one of: %s (case-insensitive)", flagValue, flagName, strings.Join(validFlavors, ", "))
		}
	case "bastion-sku":
		validSkus := []string{"Basic", "Standard", "Developer"}
		if !containsCaseInsensitive(validSkus, flagValue) {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Expected one of: %s (case-insensitive)", flagValue, flagName, strings.Join(validSkus, ", "))
		}
	case "sql-server-edition":
		validEditions := []string{"Developer", "Standard", "Enterprise"}
		if !containsCaseInsensitive(validEditions, flagValue) {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Expected one of: %s (case-insensitive)", flagValue, flagName, strings.Join(validEditions, ", "))
		}
	case "location":
		// Basic location validation - must not be empty and should be lowercase
		if flagValue == "" {
			return fmt.Errorf("Flag --%s cannot be empty", flagName)
		}
		if flagValue != strings.ToLower(flagValue) {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Azure region names should be lowercase (e.g., 'eastus2')", flagValue, flagName)
		}
	case "windows-password":
		if err := validatePasswordComplexity(flagValue); err != nil {
			return fmt.Errorf("Invalid value for flag --%s. %s", flagName, err.Error())
		}
	case "naming-prefix":
		if len(flagValue) > 7 {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Maximum length is 7 characters", flagValue, flagName)
		}
	case "auto-shutdown-time":
		// Validate HHMM format
		if len(flagValue) != 4 {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Expected format: HHMM (e.g., '1800')", flagValue, flagName)
		}
		if _, err := strconv.Atoi(flagValue); err != nil {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Expected format: HHMM (e.g., '1800')", flagValue, flagName)
		}
		// Validate hour and minute ranges
		hour, _ := strconv.Atoi(flagValue[:2])
		minute, _ := strconv.Atoi(flagValue[2:])
		if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Invalid time format (hour: 00-23, minute: 00-59)", flagValue, flagName)
		}
	case "rdp-port":
		// Validate port number (stored as string but should be valid integer port)
		if port, err := strconv.Atoi(flagValue); err != nil {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Expected numeric port value", flagValue, flagName)
		} else if port < 1 || port > 65535 {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Port must be between 1 and 65535", flagValue, flagName)
		}
	}
	return nil
}

// validateSpecificIntFlag validates specific integer flags with custom rules
func validateSpecificIntFlag(flagName, flagValue string) error {
	value, _ := strconv.Atoi(flagValue)

	switch flagName {
	case "rdp-port":
		if value < 1 || value > 65535 {
			return fmt.Errorf("Invalid value '%s' for flag --%s. Port must be between 1 and 65535", flagValue, flagName)
		}
	}
	return nil
}

// validateSpecificStringSliceFlag validates specific string slice flags with custom rules
func validateSpecificStringSliceFlag(flagName, flagValue string) error {
	switch flagName {
	case "resource-tags":
		// Basic JSON validation for resource tags
		if flagValue != "" && !strings.HasPrefix(flagValue, "{") {
			return fmt.Errorf("Invalid value for flag --%s. Expected JSON format: '{\"key\":\"value\"}'", flagName)
		}
	}
	return nil
}

// validatePasswordComplexity validates Windows password complexity requirements
func validatePasswordComplexity(password string) error {
	if len(password) < 12 || len(password) > 123 {
		return fmt.Errorf("Password must be between 12 and 123 characters long")
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	complexityCount := 0
	if hasLower {
		complexityCount++
	}
	if hasUpper {
		complexityCount++
	}
	if hasDigit {
		complexityCount++
	}
	if hasSpecial {
		complexityCount++
	}

	if complexityCount < 3 {
		return fmt.Errorf("Password must have 3 of the following: 1 lower case character, 1 upper case character, 1 number, and 1 special character")
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// containsCaseInsensitive checks if a slice contains a string (case-insensitive)
func containsCaseInsensitive(slice []string, item string) bool {
	itemLower := strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == itemLower {
			return true
		}
	}
	return false
}

// Global Output Format Functions

// ValidateOutputFormat checks if the provided output format is supported
func ValidateOutputFormat(format string) bool {
	supportedFormats := []string{"table", "json", "yaml", "tsv"}
	for _, supported := range supportedFormats {
		if strings.ToLower(format) == supported {
			return true
		}
	}
	return false
}

// PrintOutput prints data in the specified format (table, json, yaml, tsv)
func PrintOutput(data interface{}, headers []string, rows [][]string) error {
	format := strings.ToLower(OutputFormat)

	if !ValidateOutputFormat(format) {
		return fmt.Errorf("unsupported output format: %s. Supported formats: table, json, yaml, tsv", OutputFormat)
	}

	switch format {
	case "table":
		table.PrintASCIITable(headers, rows)
	case "json":
		return PrintJSON(data)
	case "yaml":
		return PrintYAML(data)
	case "tsv":
		PrintTSV(headers, rows)
	}
	return nil
}

// PrintJSON prints data in JSON format
func PrintJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %v", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// PrintYAML prints data in YAML format
func PrintYAML(data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to YAML: %v", err)
	}
	fmt.Print(string(yamlData))
	return nil
}

// PrintTSV prints data in TSV (Tab-Separated Values) format
func PrintTSV(headers []string, rows [][]string) {
	// Print headers
	fmt.Println(strings.Join(headers, "\t"))

	// Print rows
	for _, row := range rows {
		fmt.Println(strings.Join(row, "\t"))
	}
}

// PrintStructuredOutput is a convenience function for commands that have structured data
// It automatically converts structs/slices to table format with headers and rows
func PrintStructuredOutput(data interface{}) error {
	format := strings.ToLower(OutputFormat)

	if !ValidateOutputFormat(format) {
		return fmt.Errorf("unsupported output format: %s. Supported formats: table, json, yaml, tsv", OutputFormat)
	}

	switch format {
	case "json":
		return PrintJSON(data)
	case "yaml":
		return PrintYAML(data)
	case "table", "tsv":
		// For table and TSV, caller needs to provide headers and rows
		// This function is mainly for JSON/YAML convenience
		return PrintJSON(data) // fallback to JSON for complex structures
	}
	return nil
}
