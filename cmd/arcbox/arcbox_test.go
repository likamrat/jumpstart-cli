package arcbox

import (
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Test color functions for better visual feedback
var (
	testSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	testInfoColor    = color.New(color.FgCyan).SprintFunc()
	testWarnColor    = color.New(color.FgYellow).SprintFunc()
	testErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	testHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

// Helper function to print colored test output
func printTestStatus(t *testing.T, testName string, success bool, message string) {
	if success {
		fmt.Printf("%s ✅ %s: %s\n", testSuccessColor("PASS"), testHeaderColor(testName), testInfoColor(message))
	} else {
		fmt.Printf("%s ❌ %s: %s\n", testErrorColor("FAIL"), testHeaderColor(testName), testErrorColor(message))
		t.Error(message)
	}
}

func TestNewArcboxCmd(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Command Creation ==="))
	
	cmd := NewArcboxCmd()

	// Test command basic structure
	testName := "Command Use"
	success := cmd.Use == "arcbox"
	message := fmt.Sprintf("Expected 'arcbox', got '%s'", cmd.Use)
	printTestStatus(t, testName, success, message)

	testName = "Command Short Description"
	expected := "Manage Jumpstart ArcBox automation"
	success = cmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, cmd.Short)
	printTestStatus(t, testName, success, message)

	// Test that command has expected subcommands
	expectedSubcommands := []string{"deploy", "delete", "list", "preflight"}
	subcommands := cmd.Commands()

	testName = "Subcommand Count"
	success = len(subcommands) == len(expectedSubcommands)
	message = fmt.Sprintf("Expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	printTestStatus(t, testName, success, message)

	for _, expectedSubcmd := range expectedSubcommands {
		testName = fmt.Sprintf("Subcommand '%s' exists", expectedSubcmd)
		found := false
		for _, subcmd := range subcommands {
			if subcmd.Use == expectedSubcmd {
				found = true
				break
			}
		}
		message = fmt.Sprintf("Subcommand '%s' validation", expectedSubcmd)
		printTestStatus(t, testName, found, message)
	}
}

func TestArcboxDeployCommand(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Deploy Command ==="))
	
	cmd := NewArcboxCmd()

	// Find the deploy subcommand
	var deployCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "deploy" {
			deployCmd = subCmd
			break
		}
	}

	testName := "Deploy Subcommand Exists"
	success := deployCmd != nil
	message := "Deploy subcommand should be available"
	if !success {
		message = "Deploy subcommand not found"
	}
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	// Test basic command structure
	testName = "Deploy Short Description"
	expected := "Deploy a new Jumpstart ArcBox deployment"
	success = deployCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, deployCmd.Short)
	printTestStatus(t, testName, success, message)

	// Test that the command has a Run function (but don't execute it)
	testName = "Deploy Run Function"
	success = deployCmd.Run != nil
	message = "Deploy command should have a Run function"
	printTestStatus(t, testName, success, message)

	// Test that the command has proper structure without executing deployment
	testName = "Deploy Command Use"
	success = deployCmd.Use == "deploy"
	message = fmt.Sprintf("Expected 'deploy', got '%s'", deployCmd.Use)
	printTestStatus(t, testName, success, message)	// Test that command has the expected flags without setting values that would trigger deployment
	expectedFlags := []string{"location", "resource-group", "flavor", "windows-user", "windows-password"}
	for _, flagName := range expectedFlags {
		testName = fmt.Sprintf("Required Flag '%s'", flagName)
		flag := deployCmd.Flags().Lookup(flagName)
		success = flag != nil
		message = fmt.Sprintf("Flag '%s' validation", flagName)
		printTestStatus(t, testName, success, message)
	}

	// Test optional flags with their default values
	optionalFlags := []struct {
		name         string
		defaultValue string
	}{
		{"auto-shutdown", "yes"},
		{"auto-shutdown-time", "1800"},
		{"auto-shutdown-timezone", "UTC"},
		{"auto-shutdown-email", ""},
		{"bastion-sku", "Basic"},
		{"deploy-bastion", "no"},
		{"enable-spot-pricing", "no"},
		{"vm-autologon", "yes"},
		{"github-user", "microsoft"},
		{"log-analytics-workspace", ""},
		{"naming-prefix", "ArcBox"},
		{"rdp-port", "3389"},
		{"resource-tags", `{"Solution":"jumpstart_arcbox"}`},
		{"skip-preflight", "no"},
		{"sql-server-edition", "Developer"},
		{"ssh-rsa-public-key", ""},
		{"subscription", ""},
		{"template-local", ""},
		{"template-params", ""},
		{"template-uri", ""},
	}
	for _, of := range optionalFlags {
		testName = fmt.Sprintf("Optional Flag '%s'", of.name)
		flag := deployCmd.Flags().Lookup(of.name)
		success = flag != nil
		message = fmt.Sprintf("Flag '%s' exists", of.name)
		printTestStatus(t, testName, success, message)
		
		if flag != nil {
			testName = fmt.Sprintf("Default Value for '%s'", of.name)
			success = flag.DefValue == of.defaultValue
			message = fmt.Sprintf("Expected '%s', got '%s'", of.defaultValue, flag.DefValue)
			printTestStatus(t, testName, success, message)
		}
	}
	// Test boolean flags
	booleanFlags := []string{"yes"}
	for _, bf := range booleanFlags {
		testName = fmt.Sprintf("Boolean Flag '%s'", bf)
		flag := deployCmd.Flags().Lookup(bf)
		success = flag != nil
		message = fmt.Sprintf("Boolean flag '%s' exists", bf)
		printTestStatus(t, testName, success, message)
		
		if flag != nil {
			testName = fmt.Sprintf("Flag Type for '%s'", bf)
			success = flag.Value.Type() == "bool"
			message = fmt.Sprintf("Expected bool, got '%s'", flag.Value.Type())
			printTestStatus(t, testName, success, message)

			testName = fmt.Sprintf("Default Bool Value for '%s'", bf)
			success = flag.DefValue == "false"
			message = fmt.Sprintf("Expected 'false', got '%s'", flag.DefValue)
			printTestStatus(t, testName, success, message)
		}
	}	// Test shorthand flags that actually exist
	shorthandFlags := []struct {
		shorthand string
		fullName  string
	}{
		{"y", "yes"},
		{"s", "subscription"},
	}

	for _, sf := range shorthandFlags {
		testName = fmt.Sprintf("Shorthand Flag '%s' for '%s'", sf.shorthand, sf.fullName)
		shortFlag := deployCmd.Flags().ShorthandLookup(sf.shorthand)
		success = shortFlag != nil
		message = fmt.Sprintf("Shorthand '%s' for '%s'", sf.shorthand, sf.fullName)
		printTestStatus(t, testName, success, message)
	}
}

func TestArcboxDeleteCommand(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Delete Command ==="))
	
	cmd := NewArcboxCmd()

	// Find the delete subcommand
	var deleteCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "delete" {
			deleteCmd = subCmd
			break
		}
	}

	testName := "Delete Subcommand Exists"
	success := deleteCmd != nil
	message := "Delete subcommand should be available"
	if !success {
		message = "Delete subcommand not found"
	}
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	// Test basic command structure
	testName = "Delete Short Description"
	expected := "Delete a Jumpstart ArcBox deployment"
	success = deleteCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, deleteCmd.Short)
	printTestStatus(t, testName, success, message)	// Test flags (actual implementation has limited shorthand flags)
	expectedFlags := []struct {
		name         string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{"name", "n", "", "string"},
		{"yes", "", "false", "bool"},
		{"subscription", "s", "", "string"},
	}

	for _, ef := range expectedFlags {
		testName = fmt.Sprintf("Flag '%s'", ef.name)
		flag := deleteCmd.Flags().Lookup(ef.name)
		success = flag != nil
		message = fmt.Sprintf("Flag '%s' exists", ef.name)
		printTestStatus(t, testName, success, message)
		
		if flag == nil {
			continue
		}

		if ef.shorthand != "" {
			testName = fmt.Sprintf("Shorthand '%s' for '%s'", ef.shorthand, ef.name)
			shortFlag := deleteCmd.Flags().ShorthandLookup(ef.shorthand)
			success = shortFlag != nil
			message = fmt.Sprintf("Shorthand '%s' for '%s'", ef.shorthand, ef.name)
			printTestStatus(t, testName, success, message)
		}

		testName = fmt.Sprintf("Default Value for '%s'", ef.name)
		success = flag.DefValue == ef.defaultValue
		message = fmt.Sprintf("Expected '%s', got '%s'", ef.defaultValue, flag.DefValue)
		printTestStatus(t, testName, success, message)

		testName = fmt.Sprintf("Flag Type for '%s'", ef.name)
		success = flag.Value.Type() == ef.flagType
		message = fmt.Sprintf("Expected %s, got '%s'", ef.flagType, flag.Value.Type())
		printTestStatus(t, testName, success, message)
	}
}

func TestArcboxListCommand(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox List Command ==="))
	
	cmd := NewArcboxCmd()

	// Find the list subcommand
	var listCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "list" {
			listCmd = subCmd
			break
		}
	}

	testName := "List Subcommand Exists"
	success := listCmd != nil
	message := "List subcommand should be available"
	if !success {
		message = "List subcommand not found"
	}
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	// Test basic command structure
	testName = "List Short Description"
	expected := "List Jumpstart ArcBox deployments"
	success = listCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, listCmd.Short)
	printTestStatus(t, testName, success, message)	// Test flags
	expectedFlags := []struct {
		name         string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{"all-subscriptions", "", "false", "bool"},
		{"current-subscription", "", "false", "bool"},
		{"subscription", "s", "", "string"},
	}

	for _, ef := range expectedFlags {
		testName = fmt.Sprintf("Flag '%s'", ef.name)
		flag := listCmd.Flags().Lookup(ef.name)
		success = flag != nil
		message = fmt.Sprintf("Flag '%s' exists", ef.name)
		printTestStatus(t, testName, success, message)
		
		if flag == nil {
			continue
		}

		if ef.shorthand != "" {
			testName = fmt.Sprintf("Shorthand '%s' for '%s'", ef.shorthand, ef.name)
			shortFlag := listCmd.Flags().ShorthandLookup(ef.shorthand)
			success = shortFlag != nil
			message = fmt.Sprintf("Shorthand '%s' for '%s'", ef.shorthand, ef.name)
			printTestStatus(t, testName, success, message)
		}

		testName = fmt.Sprintf("Default Value for '%s'", ef.name)
		success = flag.DefValue == ef.defaultValue
		message = fmt.Sprintf("Expected '%s', got '%s'", ef.defaultValue, flag.DefValue)
		printTestStatus(t, testName, success, message)

		testName = fmt.Sprintf("Flag Type for '%s'", ef.name)
		success = flag.Value.Type() == ef.flagType
		message = fmt.Sprintf("Expected %s, got '%s'", ef.flagType, flag.Value.Type())
		printTestStatus(t, testName, success, message)
	}
}

func TestArcboxPreflightCommand(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Preflight Command ==="))
	
	cmd := NewArcboxCmd()

	// Find the preflight subcommand
	var preflightCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "preflight" {
			preflightCmd = subCmd
			break
		}
	}

	testName := "Preflight Subcommand Exists"
	success := preflightCmd != nil
	message := "Preflight subcommand should be available"
	if !success {
		message = "Preflight subcommand not found"
	}
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	// Test basic command structure
	testName = "Preflight Short Description"
	expected := "Run preflight checks for ArcBox deployment"
	success = preflightCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, preflightCmd.Short)
	printTestStatus(t, testName, success, message)

	// Test that preflight has expected subcommands
	expectedSubcommands := []string{"quota", "status", "rp"}
	subcommands := preflightCmd.Commands()

	testName = "Preflight Subcommand Count"
	success = len(subcommands) == len(expectedSubcommands)
	message = fmt.Sprintf("Expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	printTestStatus(t, testName, success, message)

	for _, expectedSubcmd := range expectedSubcommands {
		testName = fmt.Sprintf("Preflight Subcommand '%s'", expectedSubcmd)
		found := false
		for _, subcmd := range subcommands {
			if subcmd.Use == expectedSubcmd {
				found = true
				break
			}
		}
		message = fmt.Sprintf("Subcommand '%s' exists", expectedSubcmd)
		printTestStatus(t, testName, found, message)
	}
}

func TestArcboxPreflightQuotaCommand(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Preflight Quota Command ==="))
	
	cmd := NewArcboxCmd()

	// Navigate to preflight quota subcommand
	var preflightCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "preflight" {
			preflightCmd = subCmd
			break
		}
	}

	testName := "Preflight Command Available"
	success := preflightCmd != nil
	message := "Preflight subcommand should be available"
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	var quotaCmd *cobra.Command
	for _, subCmd := range preflightCmd.Commands() {
		if subCmd.Use == "quota" {
			quotaCmd = subCmd
			break
		}
	}

	testName = "Quota Subcommand Exists"
	success = quotaCmd != nil
	message = "Quota subcommand should be available"
	if !success {
		message = "Quota subcommand not found"
	}
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	// Test basic command structure
	testName = "Quota Short Description"
	expected := "Check vCPU quota for ArcBox flavors"
	success = quotaCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, quotaCmd.Short)
	printTestStatus(t, testName, success, message)
	// Test flags
	expectedFlags := []struct {
		name         string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{"flavor", "f", "", "string"},
		{"location", "l", "", "string"},
		{"all-locations", "", "false", "bool"},
		{"sku", "", "", "string"},
		{"subscription", "s", "", "string"},
	}

	for _, ef := range expectedFlags {
		testName = fmt.Sprintf("Flag '%s'", ef.name)
		flag := quotaCmd.Flags().Lookup(ef.name)
		success = flag != nil
		message = fmt.Sprintf("Flag '%s' exists", ef.name)
		printTestStatus(t, testName, success, message)
		
		if flag == nil {
			continue
		}
		
		if ef.shorthand != "" {
			testName = fmt.Sprintf("Shorthand '%s' for '%s'", ef.shorthand, ef.name)
			shortFlag := quotaCmd.Flags().ShorthandLookup(ef.shorthand)
			success = shortFlag != nil
			message = fmt.Sprintf("Shorthand '%s' for '%s'", ef.shorthand, ef.name)
			printTestStatus(t, testName, success, message)
		}

		testName = fmt.Sprintf("Default Value for '%s'", ef.name)
		success = flag.DefValue == ef.defaultValue
		message = fmt.Sprintf("Expected '%s', got '%s'", ef.defaultValue, flag.DefValue)
		printTestStatus(t, testName, success, message)

		testName = fmt.Sprintf("Flag Type for '%s'", ef.name)
		success = flag.Value.Type() == ef.flagType
		message = fmt.Sprintf("Expected %s, got '%s'", ef.flagType, flag.Value.Type())
		printTestStatus(t, testName, success, message)
	}
}

func TestArcboxPreflightRpCommand(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Preflight RP Command ==="))
	
	cmd := NewArcboxCmd()

	// Navigate to preflight rp subcommand
	var preflightCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "preflight" {
			preflightCmd = subCmd
			break
		}
	}

	testName := "Preflight Command Available"
	success := preflightCmd != nil
	message := "Preflight subcommand should be available"
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	var rpCmd *cobra.Command
	for _, subCmd := range preflightCmd.Commands() {
		if subCmd.Use == "rp" {
			rpCmd = subCmd
			break
		}
	}

	testName = "RP Subcommand Exists"
	success = rpCmd != nil
	message = "RP subcommand should be available"
	if !success {
		message = "RP subcommand not found"
	}
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	// Test basic command structure
	testName = "RP Short Description"
	expected := "Check and manage Azure resource provider registration"
	success = rpCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, rpCmd.Short)
	printTestStatus(t, testName, success, message)

	// Test that rp has expected subcommands
	expectedSubcommands := []string{"show", "list", "register"}
	subcommands := rpCmd.Commands()

	testName = "RP Subcommand Count"
	success = len(subcommands) == len(expectedSubcommands)
	message = fmt.Sprintf("Expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	printTestStatus(t, testName, success, message)

	for _, expectedSubcmd := range expectedSubcommands {
		testName = fmt.Sprintf("RP Subcommand '%s'", expectedSubcmd)
		found := false
		for _, subcmd := range subcommands {
			if subcmd.Use == expectedSubcmd {
				found = true
				break
			}
		}
		message = fmt.Sprintf("Subcommand '%s' exists", expectedSubcmd)
		printTestStatus(t, testName, found, message)
	}
}

func TestArcboxPreflightRpRegisterCommand(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Preflight RP Register Command ==="))
	
	cmd := NewArcboxCmd()

	// Navigate to preflight rp register subcommand
	var preflightCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "preflight" {
			preflightCmd = subCmd
			break
		}
	}

	testName := "Preflight Command Available"
	success := preflightCmd != nil
	message := "Preflight subcommand should be available"
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	var rpCmd *cobra.Command
	for _, subCmd := range preflightCmd.Commands() {
		if subCmd.Use == "rp" {
			rpCmd = subCmd
			break
		}
	}

	testName = "RP Command Available"
	success = rpCmd != nil
	message = "RP subcommand should be available"
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	var registerCmd *cobra.Command
	for _, subCmd := range rpCmd.Commands() {
		if subCmd.Use == "register" {
			registerCmd = subCmd
			break
		}
	}

	testName = "Register Subcommand Exists"
	success = registerCmd != nil
	message = "Register subcommand should be available"
	if !success {
		message = "Register subcommand not found"
	}
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	// Test basic command structure
	testName = "Register Short Description"
	expected := "Register a required Azure resource provider"
	success = registerCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, registerCmd.Short)
	printTestStatus(t, testName, success, message)
	// Test flags
	expectedFlags := []struct {
		name         string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{"name", "n", "", "string"},
	}

	for _, ef := range expectedFlags {
		testName = fmt.Sprintf("Flag '%s'", ef.name)
		flag := registerCmd.Flags().Lookup(ef.name)
		success = flag != nil
		message = fmt.Sprintf("Flag '%s' exists", ef.name)
		printTestStatus(t, testName, success, message)
		
		if flag == nil {
			continue
		}
		
		if ef.shorthand != "" {
			testName = fmt.Sprintf("Shorthand '%s' for '%s'", ef.shorthand, ef.name)
			shortFlag := registerCmd.Flags().ShorthandLookup(ef.shorthand)
			success = shortFlag != nil
			message = fmt.Sprintf("Shorthand '%s' for '%s'", ef.shorthand, ef.name)
			printTestStatus(t, testName, success, message)
		}

		testName = fmt.Sprintf("Default Value for '%s'", ef.name)
		success = flag.DefValue == ef.defaultValue
		message = fmt.Sprintf("Expected '%s', got '%s'", ef.defaultValue, flag.DefValue)
		printTestStatus(t, testName, success, message)

		testName = fmt.Sprintf("Flag Type for '%s'", ef.name)
		success = flag.Value.Type() == ef.flagType
		message = fmt.Sprintf("Expected %s, got '%s'", ef.flagType, flag.Value.Type())
		printTestStatus(t, testName, success, message)
	}
}

func TestQuotaCacheHelpers(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Quota Cache Helpers ==="))
	
	// Test clearQuotaCache function
	// Set some dummy data in cache
	quotaCache["test-region"] = []map[string]interface{}{
		{"name": "test", "value": 100},
	}

	testName := "Cache Has Data Before Clear"
	success := len(quotaCache) > 0
	message := "Cache should have data before clearing"
	printTestStatus(t, testName, success, message)
	clearQuotaCache()

	testName = "Cache Is Empty After Clear"
	success = len(quotaCache) == 0
	message = "Cache should be empty after clearing"
	printTestStatus(t, testName, success, message)
}

// Helper function to get a subcommand by name
func getSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == name {
			return subCmd
		}
	}
	return nil
}

func TestArcboxCommandStructure(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Command Structure ==="))
	
	cmd := NewArcboxCmd()

	// Test that command has proper error handling setup
	testName := "DisableSuggestions Setting"
	success := cmd.DisableSuggestions
	message := "DisableSuggestions should be true"
	printTestStatus(t, testName, success, message)

	testName = "SilenceErrors Setting"
	success = cmd.SilenceErrors
	message = "SilenceErrors should be true"
	printTestStatus(t, testName, success, message)

	testName = "SilenceUsage Setting"
	success = cmd.SilenceUsage
	message = "SilenceUsage should be true"
	printTestStatus(t, testName, success, message)

	// Test that RunE is set (for custom suggestion handling)
	testName = "RunE Function Set"
	success = cmd.RunE != nil
	message = "RunE should be set for custom suggestion handling"
	printTestStatus(t, testName, success, message)
}

func TestArcboxSubcommandFlags(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Subcommand Flags ==="))
	
	cmd := NewArcboxCmd()

	// Test deploy command has all expected flags
	deployCmd := getSubcommand(cmd, "deploy")
	testName := "Deploy Command Exists"
	success := deployCmd != nil
	message := "Deploy subcommand should be available"
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}
	
	// Count total flags for deploy command
	flagCount := 0
	deployCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flagCount++
	})

	// Deploy command should have many flags (25+)
	testName = "Deploy Flag Count"
	success = flagCount >= 25
	message = fmt.Sprintf("Expected at least 25 flags, got %d", flagCount)
	printTestStatus(t, testName, success, message)

	// Test delete command has minimal flags
	deleteCmd := getSubcommand(cmd, "delete")
	testName = "Delete Command Exists"
	success = deleteCmd != nil
	message = "Delete subcommand should be available"
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}
	
	deleteFlagCount := 0
	deleteCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		deleteFlagCount++
	})

	testName = "Delete Flag Count"
	success = deleteFlagCount == 3
	message = fmt.Sprintf("Expected exactly 3 flags, got %d", deleteFlagCount)
	printTestStatus(t, testName, success, message)

	// Test list command has subscription selection flags
	listCmd := getSubcommand(cmd, "list")
	testName = "List Command Exists"
	success = listCmd != nil
	message = "List subcommand should be available"
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}
	
	listFlagCount := 0
	listCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		listFlagCount++
	})

	testName = "List Flag Count"
	success = listFlagCount == 3
	message = fmt.Sprintf("Expected exactly 3 flags, got %d", listFlagCount)
	printTestStatus(t, testName, success, message)
}

func TestArcboxComplexStructure(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Complex Structure ==="))
	
	cmd := NewArcboxCmd()

	// Test nested command structure: arcbox preflight rp show
	showCmd := getNestedSubcommand(cmd, "preflight", "rp", "show")
	testName := "Nested Command 'preflight rp show'"
	success := showCmd != nil
	message := "Command 'arcbox preflight rp show' should exist"
	if !success {
		message = "Command 'arcbox preflight rp show' not found"
	}
	printTestStatus(t, testName, success, message)

	if showCmd != nil {
		testName = "Show Command Description"
		expected := "Show registration status of required Azure resource providers"
		success = showCmd.Short == expected
		message = fmt.Sprintf("Expected '%s', got '%s'", expected, showCmd.Short)
		printTestStatus(t, testName, success, message)
	}

	// Test nested command structure: arcbox preflight rp list
	listRpCmd := getNestedSubcommand(cmd, "preflight", "rp", "list")
	testName = "Nested Command 'preflight rp list'"
	success = listRpCmd != nil
	message = "Command 'arcbox preflight rp list' should exist"
	if !success {
		message = "Command 'arcbox preflight rp list' not found"
	}
	printTestStatus(t, testName, success, message)

	if listRpCmd != nil {
		testName = "List RP Command Description"
		expected := "List required Azure resource providers for ArcBox"
		success = listRpCmd.Short == expected
		message = fmt.Sprintf("Expected '%s', got '%s'", expected, listRpCmd.Short)
		printTestStatus(t, testName, success, message)
	}
}

// Helper function to get a nested subcommand
func getNestedSubcommand(cmd *cobra.Command, path ...string) *cobra.Command {
	current := cmd
	for _, name := range path {
		found := false
		for _, subCmd := range current.Commands() {
			if subCmd.Use == name {
				current = subCmd
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return current
}

func TestArcboxCommandPersistence(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Command Persistence ==="))
	
	// Test that creating multiple instances returns consistent structure
	cmd1 := NewArcboxCmd()
	cmd2 := NewArcboxCmd()

	testName := "Consistent Command Use"
	success := cmd1.Use == cmd2.Use
	message := "Commands should have consistent structure across instances"
	printTestStatus(t, testName, success, message)

	testName = "Consistent Subcommand Count"
	success = len(cmd1.Commands()) == len(cmd2.Commands())
	message = fmt.Sprintf("Expected same subcommand count, got %d vs %d", len(cmd1.Commands()), len(cmd2.Commands()))
	printTestStatus(t, testName, success, message)

	// Test that deploy command flags are consistent
	deploy1 := getSubcommand(cmd1, "deploy")
	deploy2 := getSubcommand(cmd2, "deploy")

	testName = "Deploy Commands Exist"
	success = deploy1 != nil && deploy2 != nil
	message = "Deploy commands should exist in both instances"
	printTestStatus(t, testName, success, message)
	
	if deploy1 != nil && deploy2 != nil {
		flag1Count := 0
		deploy1.Flags().VisitAll(func(flag *pflag.Flag) {
			flag1Count++
		})

		flag2Count := 0
		deploy2.Flags().VisitAll(func(flag *pflag.Flag) {
			flag2Count++
		})

		testName = "Consistent Deploy Flag Count"
		success = flag1Count == flag2Count
		message = fmt.Sprintf("Expected consistent flag count, got %d vs %d", flag1Count, flag2Count)
		printTestStatus(t, testName, success, message)
	}
}
