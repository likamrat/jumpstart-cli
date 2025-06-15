// quota_formatter.go - Display formatting for quota and resource checking
package display

import (
	"fmt"

	arcboxUtils "jumpstartcli/cmd/arcbox/utils"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/preflight/arcbox"
	"jumpstartcli/internal/table"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// This file contains functions for formatting and displaying quota information,
// resource availability checks, and preflight validation results.

// QuotaDisplay handles formatting and display of quota checking results
type QuotaDisplay struct{}

// NewQuotaDisplay creates a new QuotaDisplay instance
func NewQuotaDisplay() *QuotaDisplay {
	return &QuotaDisplay{}
}

// RunQuotaChecksWithOutput performs detailed quota checking and returns results for flexible output formatting
// Returns (allPassed bool, results []map[string]interface{})
func (qd *QuotaDisplay) RunQuotaChecksWithOutput(cli azurecli.AzureCLI, cmd *cobra.Command, location, flavor string, subscriptionGetter func(*cobra.Command, azurecli.AzureCLI) string) (bool, []map[string]interface{}) {
	// Normalize flavor and validate
	flavor = arcboxUtils.NormalizeFlavorCase(flavor)

	subscription := subscriptionGetter(cmd, cli)
	if subscription == "" {
		if utils.OutputFormat == "table" {
			fmt.Println(utils.ErrorColor("❌ [ERROR] Unable to get subscription ID. Please ensure Azure CLI is authenticated."))
		}
		return false, nil
	}

	if utils.OutputFormat == "table" {
		fmt.Printf(utils.InfoColor("🔍 Checking vCPU quota and SKU availability for %s flavor...\n"), flavor)
	}

	// Use the new preflight quota checking functionality
	quotaResults, err := arcbox.RunQuotaChecks(cli, location, flavor, subscription)
	if err != nil {
		if utils.OutputFormat == "table" {
			fmt.Printf(utils.ErrorColor("❌ [ERROR] Failed to check quota: %v\n"), err)
		}
		return false, nil
	}

	// Convert to map format for flexible output
	var results []map[string]interface{}
	allPassed := true

	for _, result := range quotaResults {
		// Set deployment status
		if !result.CanDeploy {
			allPassed = false
		}

		resultMap := map[string]interface{}{
			"Flavor":    flavor,
			"Location":  utils.GetRegionDisplayName(location),
			"SKU":       result.SKU,
			"Available": result.Available,
			"Limit":     result.Limit,
			"Required":  result.Required,
			"CanDeploy": result.CanDeploy,
			"Details":   result.Details,
		}
		results = append(results, resultMap)
	}

	// For table format, print the table immediately
	if utils.OutputFormat == "table" {
		qd.printQuotaTable(quotaResults, flavor, location, allPassed)
	}

	return allPassed, results
}

// RunQuotaChecksWithTable performs detailed quota checking with table output using Azure CLI wrapper
// Returns true if all quota checks pass, false otherwise
func (qd *QuotaDisplay) RunQuotaChecksWithTable(cli azurecli.AzureCLI, cmd *cobra.Command, location, flavor string, subscriptionGetter func(*cobra.Command, azurecli.AzureCLI) string) bool {
	allPassed, _ := qd.RunQuotaChecksWithOutput(cli, cmd, location, flavor, subscriptionGetter)
	return allPassed
}

// printQuotaTable prints the quota results in table format
func (qd *QuotaDisplay) printQuotaTable(quotaResults []arcbox.QuotaCheckResult, flavor, location string, allPassed bool) {
	headers := []string{"ArcBox Flavor", "Location", "SKU", "vCPU Quota (Available/Limit)", "Required vCPU", "Can Deploy ArcBox?", "Details"}
	var rows [][]string

	for _, result := range quotaResults {
		// Format status for display
		var canDeploy string
		if result.CanDeploy {
			canDeploy = utils.SuccessColor("Yes")
		} else {
			canDeploy = utils.ErrorColor("No")
		}

		// Format the quota column as "Available/Limit"
		quotaDisplay := fmt.Sprintf("%d/%d", result.Available, result.Limit)

		// Add row to table
		row := []string{
			flavor,
			utils.GetRegionDisplayName(location),
			result.SKU,
			quotaDisplay,
			fmt.Sprintf("%d", result.Required),
			canDeploy,
			result.Details,
		}
		rows = append(rows, row)
	}

	// Print the table
	fmt.Printf("\n")
	table.PrintASCIITable(headers, rows)
	fmt.Printf("\n")

	// Print summary
	if allPassed {
		fmt.Printf(utils.SuccessColor("✅ All quota checks passed for %s flavor in %s\n"),
			flavor, utils.GetRegionDisplayName(location))
	} else {
		fmt.Printf(utils.ErrorColor("❌ Some quota checks failed for %s flavor in %s\n"),
			flavor, utils.GetRegionDisplayName(location))
		fmt.Println(utils.InfoColor("💡 Consider requesting quota increases or choosing a different region"))
	}
}
