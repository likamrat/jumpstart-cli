package arcbox

import (
	"fmt"
	"testing"

	"jumpstartcli/internal/azurecli"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
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
	printTestStatus(t, testName, success, message) // Test that command has the expected flags without setting values that would trigger deployment
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
	} // Test shorthand flags that actually exist
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
	printTestStatus(t, testName, success, message) // Test flags (actual implementation has limited shorthand flags)
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
	printTestStatus(t, testName, success, message) // Test flags
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

// ===== Azure CLI Wrapper Tests =====

func TestDiscoverArcBoxDeployments(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Deployment Discovery ==="))

	// Create mock Azure CLI
	mockCLI := azurecli.NewMockAzureCLI()

	// Set up mock subscriptions first
	mockCLI.Subscriptions = []azurecli.SubscriptionInfo{
		{ID: "test-sub-id", Name: "Test Subscription"},
	}
	mockCLI.CurrentSubscription = &azurecli.SubscriptionInfo{
		ID:   "test-sub-id",
		Name: "Test Subscription",
	}

	// Set up mock data for resource groups
	mockCLI.ResourceGroups = []azurecli.ResourceGroupInfo{
		{Name: "ArcBox-Test-RG", Location: "eastus"},
		{Name: "LocalBox-Test-RG", Location: "westus"}, // Should be filtered out
		{Name: "Regular-RG", Location: "centralus"},
		{Name: "MC_AKS_RG", Location: "westus2"}, // Should be filtered out
	}

	// Set up mock resources with ArcBox solution tag
	mockCLI.Resources = map[string][]azurecli.ResourceInfo{
		"ArcBox-Test-RG": {
			{
				Name: "ArcBox-VM",
				Type: "Microsoft.Compute/virtualMachines",
				Tags: map[string]string{"Solution": "jumpstart_arcbox"},
			},
		},
	}

	// Set up mock deployments
	mockCLI.Deployments = map[string][]azurecli.DeploymentInfo{
		"ArcBox-Test-RG": {
			{Name: "arcbox-main", ProvisioningState: "Succeeded"},
		},
	}

	t.Run("successful_discovery", func(t *testing.T) {
		deployments, err := discoverArcBoxDeployments(mockCLI, "test-sub-id", "Test Subscription")

		testName := "Discovery Success"
		success := err == nil
		message := fmt.Sprintf("Expected no error, got: %v", err)
		printTestStatus(t, testName, success, message)

		testName = "Deployment Count"
		success = len(deployments) == 1
		message = fmt.Sprintf("Expected 1 deployment, got %d", len(deployments))
		printTestStatus(t, testName, success, message)

		if len(deployments) > 0 {
			deployment := deployments[0]

			testName = "Resource Group Name"
			success = deployment.ResourceGroupName == "ArcBox-Test-RG"
			message = fmt.Sprintf("Expected 'ArcBox-Test-RG', got '%s'", deployment.ResourceGroupName)
			printTestStatus(t, testName, success, message)

			testName = "Subscription ID"
			success = deployment.SubscriptionID == "test-sub-id"
			message = fmt.Sprintf("Expected 'test-sub-id', got '%s'", deployment.SubscriptionID)
			printTestStatus(t, testName, success, message)

			testName = "Location"
			success = deployment.Location == "eastus"
			message = fmt.Sprintf("Expected 'eastus', got '%s'", deployment.Location)
			printTestStatus(t, testName, success, message)
		}
	})

	t.Run("error_handling", func(t *testing.T) {
		// Test with error from Azure CLI
		errorMockCLI := azurecli.NewMockAzureCLI()
		errorMockCLI.SetErrorForListResourceGroups(fmt.Errorf("Azure CLI error"))

		deployments, err := discoverArcBoxDeployments(errorMockCLI, "test-sub-id", "Test Subscription")

		testName := "Error Propagation"
		success := err != nil
		message := "Expected error when Azure CLI fails"
		printTestStatus(t, testName, success, message)

		testName = "Empty Results on Error"
		success = len(deployments) == 0
		message = fmt.Sprintf("Expected 0 deployments on error, got %d", len(deployments))
		printTestStatus(t, testName, success, message)
	})
}

func TestIsArcBoxResourceGroup(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Resource Group Detection ==="))

	mockCLI := azurecli.NewMockAzureCLI()

	t.Run("name_contains_arcbox", func(t *testing.T) {
		result := isArcBoxResourceGroup(mockCLI, "MyArcBoxRG")

		testName := "Name Contains ArcBox"
		success := result == true
		message := "Resource group with 'arcbox' in name should be detected"
		printTestStatus(t, testName, success, message)
	})

	t.Run("localbox_excluded", func(t *testing.T) {
		result := isArcBoxResourceGroup(mockCLI, "MyLocalBoxRG")

		testName := "LocalBox Exclusion"
		success := result == false
		message := "Resource group with 'localbox' in name should be excluded"
		printTestStatus(t, testName, success, message)
	})

	t.Run("aks_managed_excluded", func(t *testing.T) {
		result := isArcBoxResourceGroup(mockCLI, "MC_MyCluster_myRG_eastus")

		testName := "AKS Managed Exclusion"
		success := result == false
		message := "AKS managed resource groups should be excluded"
		printTestStatus(t, testName, success, message)
	})

	t.Run("solution_tag_detection", func(t *testing.T) {
		// Set up mock resources with ArcBox solution tag
		mockCLI.Resources = map[string][]azurecli.ResourceInfo{
			"TestRG": {
				{
					Name: "test-vm",
					Type: "Microsoft.Compute/virtualMachines",
					Tags: map[string]string{"Solution": "jumpstart_arcbox"},
				},
			},
		}

		result := isArcBoxResourceGroup(mockCLI, "TestRG")

		testName := "Solution Tag Detection"
		success := result == true
		message := "Resource group with ArcBox solution tag should be detected"
		printTestStatus(t, testName, success, message)
	})

	t.Run("deployment_name_detection", func(t *testing.T) {
		// Set up mock deployments with ArcBox name
		mockCLI.Deployments = map[string][]azurecli.DeploymentInfo{
			"TestRG2": {
				{Name: "arcbox-main", ProvisioningState: "Succeeded"},
			},
		}

		result := isArcBoxResourceGroup(mockCLI, "TestRG2")

		testName := "Deployment Name Detection"
		success := result == true
		message := "Resource group with ArcBox deployment should be detected"
		printTestStatus(t, testName, success, message)
	})
}

func TestDetectArcBoxFlavor(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Flavor Detection ==="))

	mockCLI := azurecli.NewMockAzureCLI()

	t.Run("flavor_from_deployment_parameters", func(t *testing.T) {
		// Set up mock deployments with ArcBox name
		mockCLI.Deployments = map[string][]azurecli.DeploymentInfo{
			"TestRG": {
				{Name: "arcbox-main", ProvisioningState: "Succeeded"},
			},
		}

		// Set up specific deployment with parameters
		mockCLI.SpecificDeployments = map[string]*azurecli.DeploymentInfo{
			"TestRG/arcbox-main": {
				Name:              "arcbox-main",
				ProvisioningState: "Succeeded",
				Properties: map[string]interface{}{
					"parameters": map[string]interface{}{
						"flavor": map[string]interface{}{
							"value": "DevOps",
						},
					},
				},
			},
		}

		flavor, prefix := detectArcBoxFlavor(mockCLI, "TestRG")

		testName := "Flavor Detection"
		success := flavor == "DevOps"
		message := fmt.Sprintf("Expected 'DevOps', got '%s'", flavor)
		printTestStatus(t, testName, success, message)

		testName = "Prefix Detection"
		success = prefix == "ArcBox"
		message = fmt.Sprintf("Expected 'ArcBox', got '%s'", prefix)
		printTestStatus(t, testName, success, message)
	})

	t.Run("fallback_to_resource_inspection", func(t *testing.T) {
		// Set up mock resources with SQL Server (indicates DataOps)
		mockCLI.Resources = map[string][]azurecli.ResourceInfo{
			"TestRG2": {
				{
					Name: "test-sql-server",
					Type: "Microsoft.Sql/servers",
				},
			},
		}

		flavor, prefix := detectArcBoxFlavor(mockCLI, "TestRG2")

		testName := "Fallback Flavor Detection"
		success := flavor == "DataOps"
		message := fmt.Sprintf("Expected 'DataOps', got '%s'", flavor)
		printTestStatus(t, testName, success, message)

		testName = "Fallback Prefix"
		success = prefix == "ArcBox"
		message = fmt.Sprintf("Expected 'ArcBox', got '%s'", prefix)
		printTestStatus(t, testName, success, message)
	})
}

func TestEnrichArcBoxDeployment(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ArcBox Deployment Enrichment ==="))

	mockCLI := azurecli.NewMockAzureCLI()

	// Set up mock data
	mockCLI.Resources = map[string][]azurecli.ResourceInfo{
		"TestRG": {
			{Name: "vm1", Type: "Microsoft.Compute/virtualMachines"},
			{Name: "vnet1", Type: "Microsoft.Network/virtualNetworks"},
			{Name: "kv1", Type: "Microsoft.KeyVault/vaults"},
		},
	}

	mockCLI.Deployments = map[string][]azurecli.DeploymentInfo{
		"TestRG": {
			{
				Name:              "arcbox-main",
				ProvisioningState: "Succeeded",
				Properties: map[string]interface{}{
					"timestamp": "2023-01-15T10:00:00Z",
				},
			},
		},
	}

	t.Run("enrichment_success", func(t *testing.T) {
		deployment := &ArcBoxDeployment{
			ResourceGroupName: "TestRG",
			SubscriptionID:    "test-sub-id",
			SubscriptionName:  "Test Subscription",
			Location:          "eastus",
		}

		enrichArcBoxDeployment(mockCLI, deployment)

		testName := "Resource Count"
		success := deployment.ResourceCount == 3
		message := fmt.Sprintf("Expected 3 resources, got %d", deployment.ResourceCount)
		printTestStatus(t, testName, success, message)

		testName = "Status Detection"
		success = deployment.Status == "Succeeded"
		message = fmt.Sprintf("Expected 'Succeeded', got '%s'", deployment.Status)
		printTestStatus(t, testName, success, message)

		testName = "Creation Date Format"
		success = len(deployment.CreatedDate) == 10 // YYYY-MM-DD format
		message = fmt.Sprintf("Expected date format YYYY-MM-DD, got '%s'", deployment.CreatedDate)
		printTestStatus(t, testName, success, message)

		testName = "Flavor Assignment"
		success = deployment.Flavor == "ITPro" // Default fallback
		message = fmt.Sprintf("Expected 'ITPro', got '%s'", deployment.Flavor)
		printTestStatus(t, testName, success, message)
	})
}

func TestGetResourceCount(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Resource Count Function ==="))

	mockCLI := azurecli.NewMockAzureCLI()

	t.Run("count_resources", func(t *testing.T) {
		// Set up mock resources
		mockCLI.Resources = map[string][]azurecli.ResourceInfo{
			"TestRG": {
				{Name: "resource1", Type: "Microsoft.Compute/virtualMachines"},
				{Name: "resource2", Type: "Microsoft.Network/virtualNetworks"},
				{Name: "resource3", Type: "Microsoft.KeyVault/vaults"},
			},
		}

		count := getResourceCount(mockCLI, "TestRG")

		testName := "Resource Count"
		success := count == 3
		message := fmt.Sprintf("Expected 3 resources, got %d", count)
		printTestStatus(t, testName, success, message)
	})

	t.Run("error_handling", func(t *testing.T) {
		errorMockCLI := azurecli.NewMockAzureCLI()
		errorMockCLI.SetErrorForListResources(fmt.Errorf("Azure CLI error"))

		count := getResourceCount(errorMockCLI, "TestRG")

		testName := "Error Handling"
		success := count == 0
		message := fmt.Sprintf("Expected 0 on error, got %d", count)
		printTestStatus(t, testName, success, message)
	})
}

func TestGetDeploymentStatus(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Deployment Status Function ==="))

	mockCLI := azurecli.NewMockAzureCLI()

	t.Run("arcbox_deployment_priority", func(t *testing.T) {
		// Set up mock deployments with ArcBox deployment and NO failed deployments
		mockCLI.Deployments = map[string][]azurecli.DeploymentInfo{
			"TestRG": {
				{Name: "other-deployment", ProvisioningState: "Succeeded"},
				{Name: "arcbox-main", ProvisioningState: "Succeeded"},
			},
		}

		status := getDeploymentStatus(mockCLI, "TestRG")

		testName := "ArcBox Priority"
		success := status == "Succeeded"
		message := fmt.Sprintf("Expected 'Succeeded' (ArcBox priority), got '%s'", status)
		printTestStatus(t, testName, success, message)
	})

	t.Run("failed_deployment_handling", func(t *testing.T) {
		// Set up mock deployments with failed ArcBox deployment
		mockCLI.Deployments = map[string][]azurecli.DeploymentInfo{
			"TestRG2": {
				{Name: "arcbox-main", ProvisioningState: "Failed"},
				{Name: "other-deployment", ProvisioningState: "Succeeded"},
			},
		}

		status := getDeploymentStatus(mockCLI, "TestRG2")

		testName := "Failed ArcBox Deployment"
		success := status == "Failed"
		message := fmt.Sprintf("Expected 'Failed', got '%s'", status)
		printTestStatus(t, testName, success, message)
	})

	t.Run("no_deployments", func(t *testing.T) {
		emptyMockCLI := azurecli.NewMockAzureCLI()

		status := getDeploymentStatus(emptyMockCLI, "EmptyRG")

		testName := "No Deployments"
		success := status == "Unknown"
		message := fmt.Sprintf("Expected 'Unknown', got '%s'", status)
		printTestStatus(t, testName, success, message)
	})
}

func TestCheckResourceGroupExists(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Resource Group Existence Check ==="))

	mockCLI := azurecli.NewMockAzureCLI()

	// Set up mock subscriptions
	mockCLI.Subscriptions = []azurecli.SubscriptionInfo{
		{ID: "test-sub-id", Name: "Test Subscription"},
	}
	mockCLI.CurrentSubscription = &azurecli.SubscriptionInfo{
		ID:   "test-sub-id",
		Name: "Test Subscription",
	}

	t.Run("existing_resource_group", func(t *testing.T) {
		// Use a resource group that exists in the mock data
		exists, err := checkResourceGroupExists(mockCLI, "arcbox-rg", "test-sub-id")

		testName := "Check Success"
		success := err == nil
		message := fmt.Sprintf("Expected no error, got: %v", err)
		printTestStatus(t, testName, success, message)

		testName = "Resource Group Exists"
		success = exists == true
		message = fmt.Sprintf("Expected true, got %t", exists)
		printTestStatus(t, testName, success, message)
	})

	t.Run("error_handling", func(t *testing.T) {
		errorMockCLI := azurecli.NewMockAzureCLI()
		errorMockCLI.SetErrorForCheckResourceGroupExists(fmt.Errorf("Azure CLI error"))

		exists, err := checkResourceGroupExists(errorMockCLI, "ErrorRG", "test-sub-id")

		testName := "Error Propagation"
		success := err != nil
		message := "Expected error when Azure CLI fails"
		printTestStatus(t, testName, success, message)

		testName = "False on Error"
		success = exists == false
		message = fmt.Sprintf("Expected false on error, got %t", exists)
		printTestStatus(t, testName, success, message)
	})
}

func TestSetAzureCLI(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Azure CLI Dependency Injection ==="))

	t.Run("dependency_injection", func(t *testing.T) {
		// Save original CLI
		originalCLI := defaultAzureCLI

		// Create and set mock CLI
		mockCLI := azurecli.NewMockAzureCLI()
		SetAzureCLI(mockCLI)

		testName := "CLI Injection"
		success := defaultAzureCLI == mockCLI
		message := "Azure CLI should be injected successfully"
		printTestStatus(t, testName, success, message)

		// Restore original CLI
		SetAzureCLI(originalCLI)

		testName = "CLI Restoration"
		success = defaultAzureCLI == originalCLI
		message = "Original Azure CLI should be restored"
		printTestStatus(t, testName, success, message)
	})
}

func TestAzureCLIWrapperIntegration(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Azure CLI Wrapper Integration ==="))

	t.Run("mock_cli_integration", func(t *testing.T) {
		// Create comprehensive test scenario
		mockCLI := azurecli.NewMockAzureCLI()

		// Set up complete mock scenario
		mockCLI.Subscriptions = []azurecli.SubscriptionInfo{
			{ID: "test-sub-id", Name: "Test Subscription"},
		}

		mockCLI.ResourceGroups = []azurecli.ResourceGroupInfo{
			{Name: "ArcBox-Integration-RG", Location: "eastus"},
		}

		mockCLI.Resources = map[string][]azurecli.ResourceInfo{
			"ArcBox-Integration-RG": {
				{
					Name: "ArcBox-VM",
					Type: "Microsoft.Compute/virtualMachines",
					Tags: map[string]string{"Solution": "jumpstart_arcbox"},
				},
			},
		}

		mockCLI.Deployments = map[string][]azurecli.DeploymentInfo{
			"ArcBox-Integration-RG": {
				{
					Name:              "arcbox-main",
					ProvisioningState: "Succeeded",
					Properties: map[string]interface{}{
						"timestamp": "2023-01-15T10:00:00Z",
					},
				},
			},
		}

		// Test full integration workflow
		deployments, err := discoverArcBoxDeployments(mockCLI, "test-sub-id", "Test Subscription")

		testName := "Integration Success"
		success := err == nil && len(deployments) == 1
		message := fmt.Sprintf("Expected 1 deployment with no errors, got %d deployments with error: %v", len(deployments), err)
		printTestStatus(t, testName, success, message)

		if len(deployments) > 0 {
			deployment := deployments[0]

			testName := "Integration Data Completeness"
			success = deployment.ResourceGroupName != "" &&
				deployment.SubscriptionID != "" &&
				deployment.Location != "" &&
				deployment.Status != "" &&
				deployment.ResourceCount > 0
			message = "All deployment fields should be populated"
			printTestStatus(t, testName, success, message)
		}
	})
}
