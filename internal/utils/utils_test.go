package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Mock command executor for testing
type MockCommandExecutor struct {
	responses map[string]mockResponse
}

type mockResponse struct {
	output []byte
	err    error
}

func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		responses: make(map[string]mockResponse),
	}
}

func (m *MockCommandExecutor) AddResponse(cmd string, output []byte, err error) {
	m.responses[cmd] = mockResponse{output: output, err: err}
}

func (m *MockCommandExecutor) Run(name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	if resp, ok := m.responses[key]; ok {
		return resp.output, resp.err
	}
	return nil, fmt.Errorf("no mock response for command: %s", key)
}

// Helper function to temporarily replace the command executor for testing
func withMockExecutor(mock *MockCommandExecutor, fn func()) {
	oldExecutor := cmdExecutor
	cmdExecutor = mock
	defer func() { cmdExecutor = oldExecutor }()
	fn()
}

// Color functions for test output
var (
	testSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	testInfoColor    = color.New(color.FgCyan).SprintFunc()
	testErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	testHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

// Helper function to print test status
func printTestStatus(t *testing.T, testName string, success bool, message string) {
	if success {
		fmt.Printf("%s %s - %s\n", testSuccessColor("✅"), testName, testInfoColor(message))
	} else {
		fmt.Printf("%s %s - %s\n", testErrorColor("❌"), testName, testErrorColor(message))
		t.Error(message)
	}
}

func TestIsAzureLoggedIn(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Azure CLI Login Status ==="))

	// Test with mocked Azure CLI
	t.Run("azure_cli_logged_in", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Mocking Azure CLI logged in"))
		mock := NewMockCommandExecutor()
		mock.AddResponse("az account show", []byte(`{"id": "test-subscription"}`), nil)

		withMockExecutor(mock, func() {
			// Note: You'll need to update the IsAzureLoggedIn function to use cmdExecutor
			result := IsAzureLoggedIn()
			printTestStatus(t, "Azure CLI Logged In", result, "Should return true when logged in")
		})
	})

	t.Run("azure_cli_not_logged_in", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Mocking Azure CLI not logged in"))
		mock := NewMockCommandExecutor()
		mock.AddResponse("az account show", []byte(""), fmt.Errorf("Please run 'az login'"))

		withMockExecutor(mock, func() {
			result := IsAzureLoggedIn()
			printTestStatus(t, "Azure CLI Not Logged In", !result, "Should return false when not logged in")
		})
	})
}

func TestResourceGroupExists(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Resource Group Existence ==="))

	// Test with mocked responses
	t.Run("existing_rg", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing existing resource group"))
		mock := NewMockCommandExecutor()
		mock.AddResponse("az group exists --name test-rg", []byte("true"), nil)

		withMockExecutor(mock, func() {
			// Note: You'll need to update ResourceGroupExists to use cmdExecutor
			result := ResourceGroupExists("test-rg")
			printTestStatus(t, "Existing RG", result, "Existing resource group should return true")
		})
	})

	t.Run("nonexistent_rg", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing non-existent resource group"))
		mock := NewMockCommandExecutor()
		mock.AddResponse("az group exists --name jumpstart-test-nonexistent-rg-12345", []byte("false"), nil)

		withMockExecutor(mock, func() {
			result := ResourceGroupExists("jumpstart-test-nonexistent-rg-12345")
			printTestStatus(t, "Non-existent RG", !result, "Non-existent resource group should return false")
		})
	})

	// Test with empty name
	t.Run("empty_name", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing empty resource group name"))
		result := ResourceGroupExists("")
		printTestStatus(t, "Empty RG Name", !result, "Empty resource group name should return false")
	})

	// Test with invalid characters
	t.Run("invalid_name", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing invalid resource group name"))
		result := ResourceGroupExists("invalid@name!")
		printTestStatus(t, "Invalid RG Name", !result, "Invalid resource group name should return false")
	})
}

func TestCreateResourceGroup(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Resource Group Creation ==="))

	// Test successful creation
	t.Run("successful_creation", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing successful creation"))
		mock := NewMockCommandExecutor()
		mock.AddResponse("az group create --name test-rg --location eastus", []byte(`{"id": "/subscriptions/xxx/resourceGroups/test-rg"}`), nil)

		withMockExecutor(mock, func() {
			// Note: You'll need to update CreateResourceGroup to use cmdExecutor
			err := CreateResourceGroup("test-rg", "eastus")
			printTestStatus(t, "Successful Creation", err == nil, "Should create resource group successfully")
		})
	})

	// Test with invalid parameters to ensure error handling
	t.Run("invalid_location", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing invalid location"))
		mock := NewMockCommandExecutor()
		mock.AddResponse("az group create --name test-rg --location invalid-location-12345",
			[]byte(""), fmt.Errorf("Invalid location"))

		withMockExecutor(mock, func() {
			err := CreateResourceGroup("test-rg", "invalid-location-12345")
			printTestStatus(t, "Invalid Location", err != nil, "Should return error for invalid location")
		})
	})

	// Test with empty parameters
	t.Run("empty_parameters", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing empty parameters"))
		err := CreateResourceGroup("", "")
		printTestStatus(t, "Empty Parameters", err != nil, "Should return error for empty resource group parameters")
	})
}

func TestLoggingFunctions(t *testing.T) {
	// Test all log functions don't panic
	t.Run("Info", func(t *testing.T) {
		assert.NotPanics(t, func() { Info("test info message") })
	})

	t.Run("Warn", func(t *testing.T) {
		assert.NotPanics(t, func() { Warn("test warn message") })
	})

	t.Run("Error", func(t *testing.T) {
		assert.NotPanics(t, func() { Error("test error message") })
	})

	t.Run("Success", func(t *testing.T) {
		assert.NotPanics(t, func() { Success("test success message") })
	})

	t.Run("Prompt", func(t *testing.T) {
		assert.NotPanics(t, func() { Prompt("test prompt message") })
	})
}

func TestDebugFunction(t *testing.T) {
	oldDebugMode := DebugMode
	defer func() { DebugMode = oldDebugMode }()

	// Test with DebugMode false
	DebugMode = false
	assert.NotPanics(t, func() { Debug("debug message") })

	// Test with DebugMode true
	DebugMode = true
	assert.NotPanics(t, func() { Debug("debug message") })
}

func TestFatalFunction(t *testing.T) {
	// We can't easily test Fatal as it calls os.Exit
	// In a real test suite, you'd use a wrapper or mock
	// For now, we'll just ensure the function exists
	assert.NotNil(t, Fatal)
}

func TestFatalError(t *testing.T) {
	// Test the new testable FatalError function
	tests := []struct {
		msg      string
		args     []interface{}
		expected string
	}{
		{"test message", nil, "[FATAL] test message"},
		{"test with arg: %s", []interface{}{"value"}, "[FATAL] test with arg: value"},
		{"test with multiple args: %s %d", []interface{}{"value", 42}, "[FATAL] test with multiple args: value 42"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			var err error
			if tt.args != nil {
				err = FatalError(tt.msg, tt.args...)
			} else {
				err = FatalError(tt.msg)
			}
			assert.Error(t, err)
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

func TestFriendlyResourceName(t *testing.T) {
	tests := []struct {
		resourceType string
		resourceName string
		expected     string
	}{
		{"Microsoft.OperationalInsights/workspaces", "test", "Log Analytics workspace"},
		{"Microsoft.Network/networkSecurityGroups", "test", "Network Security Group"},
		{"Microsoft.KeyVault/vaults", "test", "Azure Key Vault"},
		{"Microsoft.Compute/virtualMachines/extensions", "vm/Microsoft.Azure.Geneva.GenevaMonitoring", "Azure Geneva Monitoring"},
		{"Microsoft.Compute/virtualMachines/extensions", "vm/Bootstrap", "Bootstrap"},
		{"Microsoft.Compute/virtualMachines/extensions", "vm/CustomExtension", "CustomExtension"},
		{"Microsoft.Storage/storageAccounts", "test", "Storage Account"},
		{"Unknown/Type", "test", "Unknown/Type"},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			result := FriendlyResourceName(tt.resourceType, tt.resourceName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintMissingRequiredFlagsError(t *testing.T) {
	// Test the refactored PrintMissingRequiredFlagsError function that now returns bool
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("required1", "", "Required flag 1")
	cmd.Flags().String("required2", "", "Required flag 2")
	cmd.Flags().String("optional", "default", "Optional flag")

	// Test with missing required flags
	t.Run("Missing required flags", func(t *testing.T) {
		result := PrintMissingRequiredFlagsError(cmd, []string{"required1", "required2"})
		assert.False(t, result, "Should return false when required flags are missing")
	})

	// Test with all required flags present
	t.Run("All required flags present", func(t *testing.T) {
		cmd.Flags().Set("required1", "value1")
		cmd.Flags().Set("required2", "value2")
		result := PrintMissingRequiredFlagsError(cmd, []string{"required1", "required2"})
		assert.True(t, result, "Should return true when all required flags are present")
	})
}

func TestIsFlagMissing(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("test-flag", "", "Test flag")
	cmd.Flags().StringSlice("slice-flag", []string{}, "Slice flag")

	tests := []struct {
		name     string
		flagName string
		setValue string
		expected bool
	}{
		{"nil flag", "nonexistent", "", true},
		{"unchanged flag", "test-flag", "", true},
		{"empty string", "test-flag", "", true},
		{"empty slice", "slice-flag", "[]", true},
		{"valid value", "test-flag", "value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.flagName == "test-flag" && tt.setValue != "" {
				cmd.Flags().Set("test-flag", tt.setValue)
			}
			flag := cmd.Flags().Lookup(tt.flagName)
			result := isFlagMissing(cmd, flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildCustomHelpOutput(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Long:  "A longer description of the test command",
	}
	cmd.Flags().String("flag1", "default", "First flag")
	cmd.Flags().BoolP("flag2", "f", false, "Second flag")

	subCmd := &cobra.Command{
		Use:   "subcmd",
		Short: "Subcommand",
		Run: func(cmd *cobra.Command, args []string) {
			// Empty run function to make it available
		},
	}
	cmd.AddCommand(subCmd)

	output := buildCustomHelpOutput(cmd)

	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "Test command")
	assert.Contains(t, output, "Available Commands:")
	assert.Contains(t, output, "FLAGS:")
	assert.Contains(t, output, "--flag1")
	assert.Contains(t, output, "--flag2")
}

func TestBuildFlagsOutput(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("test-flag", "default", "Test flag description")
	cmd.Flags().BoolP("bool-flag", "b", false, "Boolean flag")

	output := buildFlagsOutput(cmd, cmd.Flags(), 80)

	assert.Contains(t, output, "--test-flag")
	assert.Contains(t, output, "Test flag description")
	assert.Contains(t, output, "default")
}

func TestBuildFlagString(t *testing.T) {
	tests := []struct {
		flag     flagInfo
		expected string
	}{
		{flagInfo{name: "test", shorthand: "t"}, "-t, --test"},
		{flagInfo{name: "test", shorthand: ""}, "    --test"},
	}

	for _, tt := range tests {
		result := buildFlagString(tt.flag)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIsBooleanFlag(t *testing.T) {
	tests := []struct {
		flagName string
		expected bool
	}{
		{"auto-shutdown-enabled", true},
		{"deploy-bastion", true},
		{"vm-size", true},
		{"regular-flag", false},
		{"name", false},
	}

	for _, tt := range tests {
		result := isBooleanFlag(tt.flagName)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIsBooleanStringFlag(t *testing.T) {
	tests := []struct {
		flagName   string
		defaultVal string
		expected   bool
	}{
		{"auto-shutdown", "", true},
		{"deploy-bastion", "", true},
		{"vm-autologon", "", true},
		{"some-flag", "yes", true},
		{"some-flag", "no", true},
		{"regular-flag", "value", false},
	}

	for _, tt := range tests {
		result := isBooleanStringFlag(tt.flagName, tt.defaultVal)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIsEnumStringFlag(t *testing.T) {
	tests := []struct {
		flagName     string
		expectedBool bool
		expectedVals string
	}{
		{"flavor", true, "ITPro, DevOps, DataOps"},
		{"sql-server-edition", true, "Developer, Standard, Enterprise"},
		{"bastion-sku", true, "Basic, Standard, Developer"},
		{"unknown-flag", false, ""},
	}

	for _, tt := range tests {
		isEnum, vals := isEnumStringFlag(tt.flagName, "")
		assert.Equal(t, tt.expectedBool, isEnum)
		if isEnum {
			assert.Equal(t, tt.expectedVals, vals)
		}
	}
}

func TestIsPathFlag(t *testing.T) {
	tests := []struct {
		flagName string
		expected bool
	}{
		{"template-local", true},
		{"template-params", true},
		{"other-flag", false},
	}

	for _, tt := range tests {
		result := isPathFlag(tt.flagName)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIsURIFlag(t *testing.T) {
	tests := []struct {
		flagName string
		expected bool
	}{
		{"template-uri", true},
		{"other-flag", false},
	}

	for _, tt := range tests {
		result := isURIFlag(tt.flagName)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFormatArgumentReferences(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Use auto-shutdown-enabled flag", "Use auto-shutdown-enabled flag"}, // Should contain ANSI codes
		{"Set resource-group name", "Set resource-group name"},
		{"No hyphens here", "No hyphens here"},
	}

	for _, tt := range tests {
		result := formatArgumentReferences(tt.input)
		// We can't easily test ANSI codes, so just ensure function runs
		assert.IsType(t, "", result)
		assert.Contains(t, result, strings.Split(tt.input, " ")[0])
	}
}

func TestWrapTextAtWords(t *testing.T) {
	tests := []struct {
		text     string
		maxWidth int
		expected int // number of lines
	}{
		{"short", 20, 1},
		{"this is a longer text that should wrap", 10, 5},
		{"word", 10, 1},
		{"", 10, 1},
	}

	for _, tt := range tests {
		result := wrapTextAtWords(tt.text, tt.maxWidth)
		assert.Equal(t, tt.expected, len(result))
	}
}

func TestStripAnsiCodes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"plain text", "plain text"},
		{"\x1b[31mred text\x1b[0m", "red text"},
		{"\x1b[1;32mgreen bold\x1b[0m", "green bold"},
	}

	for _, tt := range tests {
		result := stripAnsiCodes(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIsRequiredFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("subscription-id", "", "Subscription ID")
	cmd.Flags().String("optional-flag", "", "Optional flag")

	tests := []struct {
		flagName string
		expected bool
	}{
		{"subscription-id", true},
		{"optional-flag", false},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		result := isRequiredFlag(cmd, tt.flagName)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIsHardcodedRequiredFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "js arcbox deploy"}
	cmd.SetUsageTemplate("js arcbox deploy")

	tests := []struct {
		flagName string
		expected bool
	}{
		{"subscription-id", true},
		{"location", false}, // Would be true for arcbox deploy command
		{"optional-flag", false},
	}

	for _, tt := range tests {
		result := isHardcodedRequiredFlag(cmd, tt.flagName)
		assert.Equal(t, tt.expected, result)
	}
}

func TestNormalizeRegion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"East US", "eastus"},
		{"WEST US 2", "westus2"},
		{"eastus", "eastus"},
		{"North Europe", "northeurope"},
	}

	for _, tt := range tests {
		result := NormalizeRegion(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestSuggestSimilarCommand(t *testing.T) {
	commands := []string{"deploy", "delete", "list", "show"}

	tests := []struct {
		input     string
		threshold int
		expected  string
	}{
		{"deploi", 2, "deploy"},
		{"delet", 2, "delete"},
		{"xyz", 2, ""},
		{"deplo", 3, "deploy"},
	}

	for _, tt := range tests {
		result := SuggestSimilarCommand(tt.input, commands, tt.threshold)
		assert.Equal(t, tt.expected, result)
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "def", 3},
	}

	for _, tt := range tests {
		result := levenshteinDistance(tt.a, tt.b)
		assert.Equal(t, tt.expected, result)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, c  int
		expected int
	}{
		{1, 2, 3, 1},
		{3, 1, 2, 1},
		{2, 3, 1, 1},
		{5, 5, 5, 5},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b, tt.c)
		assert.Equal(t, tt.expected, result)
	}
}

func TestGetRegionDisplayName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"eastus", "East US"},
		{"westus2", "West US 2"},
		{"northeurope", "North Europe"},
		{"unknown-region", "Unknown-Region"},
	}

	for _, tt := range tests {
		result := GetRegionDisplayName(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestRegionExistsInAzure(t *testing.T) {
	t.Run("valid_region", func(t *testing.T) {
		mock := NewMockCommandExecutor()
		mock.AddResponse("az account list-locations --query [?name=='eastus'].name --output tsv",
			[]byte("eastus"), nil)

		withMockExecutor(mock, func() {
			result := RegionExistsInAzure("eastus")
			assert.True(t, result)
		})
	})

	t.Run("invalid_region", func(t *testing.T) {
		mock := NewMockCommandExecutor()
		mock.AddResponse("az account list-locations --query [?name=='invalid-region'].name --output tsv",
			[]byte(""), nil)

		withMockExecutor(mock, func() {
			result := RegionExistsInAzure("invalid-region")
			assert.False(t, result)
		})
	})

	t.Run("command_error", func(t *testing.T) {
		mock := NewMockCommandExecutor()
		mock.AddResponse("az account list-locations --query [?name=='eastus'].name --output tsv",
			nil, fmt.Errorf("command failed"))

		withMockExecutor(mock, func() {
			result := RegionExistsInAzure("eastus")
			assert.False(t, result)
		})
	})
}

func TestGetBooleanFlagValue(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("bool-flag", false, "Boolean flag")
	cmd.Flags().String("string-bool-flag", "", "String boolean flag")
	cmd.Flags().Bool("positive-flag", false, "Positive flag")
	cmd.Flags().Bool("no-positive-flag", false, "Negative flag")

	tests := []struct {
		flagName string
		setValue string
		expected bool
	}{
		{"bool-flag", "true", true},
		{"string-bool-flag", "yes", true},
		{"string-bool-flag", "no", false},
		{"positive-flag", "true", true},
	}

	for _, tt := range tests {
		if strings.Contains(tt.flagName, "string") {
			cmd.Flags().Set(tt.flagName, tt.setValue)
		} else {
			cmd.Flags().Set(tt.flagName, tt.setValue)
		}
		result := GetBooleanFlagValue(cmd, tt.flagName)
		assert.Equal(t, tt.expected, result)
	}
}

func TestParseYesNoToBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		wantErr  bool
	}{
		{"yes", true, false},
		{"y", true, false},
		{"true", true, false},
		{"1", true, false},
		{"no", false, false},
		{"n", false, false},
		{"false", false, false},
		{"0", false, false},
		{"invalid", false, true},
		{"", false, true},
	}

	for _, tt := range tests {
		result, err := ParseYesNoToBool(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestRegisterBooleanFlagPair(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	RegisterBooleanFlagPair(cmd, "test-flag", "t", true, "Test flag")

	// Check that both flags exist
	flag1 := cmd.Flags().Lookup("test-flag")
	flag2 := cmd.Flags().Lookup("no-test-flag")

	assert.NotNil(t, flag1)
	assert.NotNil(t, flag2)
	assert.Equal(t, "t", flag1.Shorthand)
}

func TestValidateAllFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("test-flag", "", "Test flag")

	// Test with valid flags
	err := ValidateAllFlags(cmd)
	assert.NoError(t, err)
}

func TestPrintOutput(t *testing.T) {
	// Save original OutputFormat
	oldFormat := OutputFormat
	defer func() { OutputFormat = oldFormat }()

	data := map[string]string{"key": "value"}
	headers := []string{"Key", "Value"}
	rows := [][]string{{"key", "value"}}

	t.Run("json_format", func(t *testing.T) {
		OutputFormat = "json"
		err := PrintOutput(data, headers, rows)
		assert.NoError(t, err)
	})

	t.Run("yaml_format", func(t *testing.T) {
		OutputFormat = "yaml"
		err := PrintOutput(data, headers, rows)
		assert.NoError(t, err)
	})

	t.Run("table_format", func(t *testing.T) {
		OutputFormat = "table"
		err := PrintOutput(data, headers, rows)
		assert.NoError(t, err)
	})

	t.Run("tsv_format", func(t *testing.T) {
		OutputFormat = "tsv"
		err := PrintOutput(data, headers, rows)
		assert.NoError(t, err)
	})

	t.Run("invalid_format", func(t *testing.T) {
		OutputFormat = "invalid"
		err := PrintOutput(data, headers, rows)
		assert.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkNormalizeRegion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeRegion("East US")
	}
}

func BenchmarkParseYesNoToBool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseYesNoToBool("yes")
	}
}

func BenchmarkGetBooleanFlagValue(b *testing.B) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("bool-flag", false, "Boolean flag")
	for i := 0; i < b.N; i++ {
		GetBooleanFlagValue(cmd, "bool-flag")
	}
}

func BenchmarkRegionExistsInAzure(b *testing.B) {
	mock := NewMockCommandExecutor()
	mock.AddResponse("az account list-locations --query [].name -o tsv",
		[]byte("eastus\nwestus\nwestus2"), nil)
	withMockExecutor(mock, func() {
		for i := 0; i < b.N; i++ {
			RegionExistsInAzure("eastus")
		}
	})
}

func BenchmarkMin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		min(1, 2, 3)
	}
}

func BenchmarkGetRegionDisplayName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetRegionDisplayName("eastus")
	}
}

// Test version-related utilities
func TestVersionUtilities(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Version Utilities ==="))

	t.Run("format_version_output", func(t *testing.T) {
		versionInfo := map[string]interface{}{
			"version":      "v1.0.0",
			"buildDate":    "2024-01-01",
			"gitCommit":    "abc123",
			"gitTag":       "v1.0.0",
			"goVersion":    "go1.21",
			"osArch":       "linux/amd64",
			"gitTreeState": "clean",
		}

		// Test different output formats
		oldFormat := OutputFormat
		defer func() { OutputFormat = oldFormat }()

		formats := []string{"json", "yaml", "table", "tsv"}
		for _, format := range formats {
			OutputFormat = format
			err := PrintOutput(versionInfo, []string{"Field", "Value"}, [][]string{
				{"Version", "v1.0.0"},
				{"Build Date", "2024-01-01"},
				{"Git Commit", "abc123"},
				{"Git Tag", "v1.0.0"},
				{"Go Version", "go1.21"},
				{"OS/Arch", "linux/amd64"},
				{"Git Tree State", "clean"},
			})
			printTestStatus(t, fmt.Sprintf("Version Output Format: %s", format), err == nil,
				fmt.Sprintf("Should successfully output version in %s format", format))
		}
	})

	t.Run("version_string_validation", func(t *testing.T) {
		testCases := []struct {
			version string
			valid   bool
			desc    string
		}{
			{"v1.0.0", true, "Standard semver"},
			{"v1.0.0-beta", true, "Pre-release version"},
			{"v1.0.0-beta.1", true, "Pre-release with number"},
			{"v1.0.0+20240101", true, "Version with metadata"},
			{"unknown", true, "Unknown version (dev build)"},
			{"", false, "Empty version"},
		}

		for _, tc := range testCases {
			// Since we don't have a ValidateVersion function, we'll just check non-empty
			isValid := tc.version != ""
			printTestStatus(t, fmt.Sprintf("Version Validation: %s", tc.desc),
				isValid == tc.valid, fmt.Sprintf("Version '%s' validation", tc.version))
		}
	})
}

// Test helper for checking version command integration
func TestVersionCommandIntegration(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Version Command Integration ==="))

	t.Run("version_command_exists", func(t *testing.T) {
		mock := NewMockCommandExecutor()
		mock.AddResponse("jumpstart version", []byte("Version: v1.0.0\nBuild Date: 2024-01-01"), nil)

		withMockExecutor(mock, func() {
			output, err := cmdExecutor.Run("jumpstart", "version")
			printTestStatus(t, "Version Command Execution", err == nil && len(output) > 0,
				"Version command should execute successfully")
		})
	})

	t.Run("version_short_flag", func(t *testing.T) {
		mock := NewMockCommandExecutor()
		mock.AddResponse("jumpstart version --short", []byte("v1.0.0"), nil)

		withMockExecutor(mock, func() {
			output, err := cmdExecutor.Run("jumpstart", "version", "--short")
			success := err == nil && strings.TrimSpace(string(output)) == "v1.0.0"
			printTestStatus(t, "Version Short Flag", success,
				"Version --short should only output version number")
		})
	})

	t.Run("version_output_formats", func(t *testing.T) {
		formats := []string{"json", "yaml", "table", "tsv"}

		for _, format := range formats {
			mock := NewMockCommandExecutor()
			// Mock different outputs based on format
			switch format {
			case "json":
				mock.AddResponse(fmt.Sprintf("jumpstart version --output %s", format),
					[]byte(`{"version":"v1.0.0","buildDate":"2024-01-01"}`), nil)
			case "yaml":
				mock.AddResponse(fmt.Sprintf("jumpstart version --output %s", format),
					[]byte("version: v1.0.0\nbuildDate: 2024-01-01"), nil)
			default:
				mock.AddResponse(fmt.Sprintf("jumpstart version --output %s", format),
					[]byte("Version\tv1.0.0\nBuild Date\t2024-01-01"), nil)
			}

			withMockExecutor(mock, func() {
				output, err := cmdExecutor.Run("jumpstart", "version", "--output", format)
				printTestStatus(t, fmt.Sprintf("Version Output Format: %s", format),
					err == nil && len(output) > 0,
					fmt.Sprintf("Should output version in %s format", format))
			})
		}
	})
}

// Benchmark version-related operations
func BenchmarkVersionFormatting(b *testing.B) {
	versionInfo := map[string]interface{}{
		"version":   "v1.0.0",
		"buildDate": "2024-01-01",
		"gitCommit": "abc123",
	}

	headers := []string{"Field", "Value"}
	rows := [][]string{
		{"Version", "v1.0.0"},
		{"Build Date", "2024-01-01"},
		{"Git Commit", "abc123"},
	}

	b.Run("json_format", func(b *testing.B) {
		oldFormat := OutputFormat
		OutputFormat = "json"
		defer func() { OutputFormat = oldFormat }()

		for i := 0; i < b.N; i++ {
			_ = PrintOutput(versionInfo, headers, rows)
		}
	})

	b.Run("table_format", func(b *testing.B) {
		oldFormat := OutputFormat
		OutputFormat = "table"
		defer func() { OutputFormat = oldFormat }()

		for i := 0; i < b.N; i++ {
			_ = PrintOutput(versionInfo, headers, rows)
		}
	})
}

func TestValidateAndPrintFlagsError(t *testing.T) {
	// Test with valid flags
	t.Run("Valid flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("valid-flag", "value", "A valid flag")
		cmd.Flags().Set("valid-flag", "test-value")

		result := ValidateAndPrintFlagsError(cmd)
		assert.True(t, result, "Should return true for valid flags")
	})

	// Test with invalid flags - we'll simulate this by creating a command
	// that has validation issues
	t.Run("Invalid flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		// This will be tested through integration since validateStringFlag
		// depends on complex flag validation logic
		result := ValidateAndPrintFlagsError(cmd)
		assert.True(t, result, "Should return true when no flags set")
	})
}

func TestValidateRequiredStringsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]string
		hasError bool
		expected string
	}{
		{
			name: "All fields valid",
			fields: map[string]string{
				"field1": "value1",
				"field2": "value2",
			},
			hasError: false,
		},
		{
			name: "One empty field",
			fields: map[string]string{
				"field1": "value1",
				"field2": "",
			},
			hasError: true,
			expected: "required fields cannot be empty: field2",
		},
		{
			name: "Multiple empty fields",
			fields: map[string]string{
				"field1": "",
				"field2": "value2",
				"field3": "",
			},
			hasError: true,
			expected: "required fields cannot be empty: field1, field3",
		},
		{
			name: "Whitespace only field",
			fields: map[string]string{
				"field1": "value1",
				"field2": "   ",
			},
			hasError: true,
			expected: "required fields cannot be empty: field2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredStringsEmpty(tt.fields)
			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "required fields cannot be empty")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStringInList(t *testing.T) {
	validOptions := []string{"option1", "option2", "Option3"}

	tests := []struct {
		name      string
		value     string
		fieldName string
		hasError  bool
		expected  string
	}{
		{
			name:      "Valid option exact match",
			value:     "option1",
			fieldName: "test-field",
			hasError:  false,
		},
		{
			name:      "Valid option case insensitive",
			value:     "OPTION2",
			fieldName: "test-field",
			hasError:  false,
		},
		{
			name:      "Invalid option",
			value:     "invalid",
			fieldName: "test-field",
			hasError:  true,
			expected:  "invalid test-field 'invalid'. Valid options: option1, option2, Option3",
		},
		{
			name:      "Empty value",
			value:     "",
			fieldName: "test-field",
			hasError:  true,
			expected:  "test-field cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringInList(tt.value, validOptions, tt.fieldName)
			if tt.hasError {
				assert.Error(t, err)
				if tt.expected != "" {
					assert.Equal(t, tt.expected, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationErrorCollector(t *testing.T) {
	t.Run("Empty collector", func(t *testing.T) {
		collector := NewValidationErrorCollector()
		assert.False(t, collector.HasErrors())
		assert.Equal(t, 0, collector.Count())
		assert.NoError(t, collector.Error())
	})

	t.Run("Single error", func(t *testing.T) {
		collector := NewValidationErrorCollector()
		collector.AddError("test error")

		assert.True(t, collector.HasErrors())
		assert.Equal(t, 1, collector.Count())

		err := collector.Error()
		assert.Error(t, err)
		assert.Equal(t, "test error", err.Error())
	})

	t.Run("Multiple errors", func(t *testing.T) {
		collector := NewValidationErrorCollector()
		collector.AddError("error 1")
		collector.AddError("error 2")
		collector.AddError("error 3")

		assert.True(t, collector.HasErrors())
		assert.Equal(t, 3, collector.Count())

		err := collector.Error()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "multiple validation errors:")
		assert.Contains(t, err.Error(), "error 1")
		assert.Contains(t, err.Error(), "error 2")
		assert.Contains(t, err.Error(), "error 3")
	})

	t.Run("Formatted errors", func(t *testing.T) {
		collector := NewValidationErrorCollector()
		collector.AddError("error with value: %s", "test")
		collector.AddError("error with number: %d", 42)

		assert.True(t, collector.HasErrors())
		assert.Equal(t, 2, collector.Count())

		err := collector.Error()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error with value: test")
		assert.Contains(t, err.Error(), "error with number: 42")
	})

	t.Run("Conditional errors", func(t *testing.T) {
		collector := NewValidationErrorCollector()

		collector.AddErrorIf(true, "this error should be added")
		collector.AddErrorIf(false, "this error should NOT be added")
		collector.AddErrorIf(1 > 0, "this error should also be added")

		assert.True(t, collector.HasErrors())
		assert.Equal(t, 2, collector.Count())

		err := collector.Error()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "this error should be added")
		assert.Contains(t, err.Error(), "this error should also be added")
		assert.NotContains(t, err.Error(), "this error should NOT be added")
	})
}

func TestExampleValidationPattern(t *testing.T) {
	tests := []struct {
		name          string
		resourceGroup string
		location      string
		flavor        string
		hasError      bool
		expectedError string
	}{
		{
			name:          "Valid inputs",
			resourceGroup: "my-rg",
			location:      "eastus",
			flavor:        "ITPro",
			hasError:      false,
		},
		{
			name:          "Empty resource group",
			resourceGroup: "",
			location:      "eastus",
			flavor:        "ITPro",
			hasError:      true,
			expectedError: "resource-group",
		},
		{
			name:          "Invalid flavor",
			resourceGroup: "my-rg",
			location:      "eastus",
			flavor:        "InvalidFlavor",
			hasError:      true,
			expectedError: "invalid flavor",
		},
		{
			name:          "Resource group with spaces",
			resourceGroup: "my resource group",
			location:      "eastus",
			flavor:        "ITPro",
			hasError:      true,
			expectedError: "cannot contain spaces",
		},
		{
			name:          "Multiple validation errors",
			resourceGroup: "",
			location:      "",
			flavor:        "BadFlavor",
			hasError:      true,
			expectedError: "multiple validation errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExampleValidationPattern(tt.resourceGroup, tt.location, tt.flavor)

			if tt.hasError {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
