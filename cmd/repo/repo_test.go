package repo

import (
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Test color functions for better visual feedback
var (
	testSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	testInfoColor    = color.New(color.FgCyan).SprintFunc()
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

func TestNewRepoCmd(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Repo Command Creation ==="))

	cmd := NewRepoCmd()

	// Test command basic structure
	testName := "Command Use"
	success := cmd.Use == "repo"
	message := fmt.Sprintf("Expected 'repo', got '%s'", cmd.Use)
	printTestStatus(t, testName, success, message)

	testName = "Command Short Description"
	expected := "Manage Jumpstart user local source code repository"
	success = cmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, cmd.Short)
	printTestStatus(t, testName, success, message)

	// Test that command has expected subcommands
	expectedSubcommands := []string{"init", "update", "delete"}
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

func TestRepoInitCommand(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Repo Init Command ==="))

	cmd := NewRepoCmd()

	// Find the init subcommand
	var initCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "init" {
			initCmd = subCmd
			break
		}
	}

	testName := "Init Subcommand Exists"
	success := initCmd != nil
	message := "Init subcommand should be available"
	if !success {
		message = "Init subcommand not found"
	}
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	// Test basic command structure
	testName = "Init Short Description"
	expected := "Initialize a new local Jumpstart repository"
	success = initCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, initCmd.Short)
	printTestStatus(t, testName, success, message)

	// Test flags
	expectedFlags := []struct {
		name         string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{"path", "p", "", "string"},
		{"branch", "b", "main", "string"},
		{"force", "f", "false", "bool"},
	}

	for _, ef := range expectedFlags {
		testName = fmt.Sprintf("Flag '%s'", ef.name)
		flag := initCmd.Flags().Lookup(ef.name)
		success = flag != nil
		message = fmt.Sprintf("Flag '%s' exists", ef.name)
		printTestStatus(t, testName, success, message)

		if flag == nil {
			continue
		}

		if ef.shorthand != "" {
			testName = fmt.Sprintf("Shorthand '%s' for '%s'", ef.shorthand, ef.name)
			shortFlag := initCmd.Flags().ShorthandLookup(ef.shorthand)
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

func TestRepoUpdateCommand(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Repo Update Command ==="))

	cmd := NewRepoCmd()

	// Find the update subcommand
	var updateCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "update" {
			updateCmd = subCmd
			break
		}
	}

	testName := "Update Subcommand Exists"
	success := updateCmd != nil
	message := "Update subcommand should be available"
	if !success {
		message = "Update subcommand not found"
	}
	printTestStatus(t, testName, success, message)
	if !success {
		return
	}

	// Test basic command structure
	testName = "Update Short Description"
	expected := "Update existing repository with latest templates"
	success = updateCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, updateCmd.Short)
	printTestStatus(t, testName, success, message)

	// Test flags
	expectedFlags := []struct {
		name         string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{"path", "p", "", "string"},
		{"branch", "b", "main", "string"},
	}

	for _, ef := range expectedFlags {
		testName = fmt.Sprintf("Flag '%s'", ef.name)
		flag := updateCmd.Flags().Lookup(ef.name)
		success = flag != nil
		message = fmt.Sprintf("Flag '%s' exists", ef.name)
		printTestStatus(t, testName, success, message)

		if flag == nil {
			continue
		}

		if ef.shorthand != "" {
			testName = fmt.Sprintf("Shorthand '%s' for '%s'", ef.shorthand, ef.name)
			shortFlag := updateCmd.Flags().ShorthandLookup(ef.shorthand)
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

func TestRepoDeleteCommand(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Repo Delete Command ==="))

	cmd := NewRepoCmd()

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
	expected := "Delete local repository"
	success = deleteCmd.Short == expected
	message = fmt.Sprintf("Expected '%s', got '%s'", expected, deleteCmd.Short)
	printTestStatus(t, testName, success, message)

	// Test flags
	expectedFlags := []struct {
		name         string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{"path", "p", "", "string"},
		{"force", "f", "false", "bool"},
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
