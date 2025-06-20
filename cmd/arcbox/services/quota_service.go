package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"jumpstartcli/cmd/arcbox/display"
	arcboxUtils "jumpstartcli/cmd/arcbox/utils"
	"jumpstartcli/internal/artifacts/regions"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// QuotaService handles quota checking functionality for ArcBox
type QuotaService struct {
	cli azurecli.AzureCLI
}

// NewQuotaService creates a new QuotaService with the provided Azure CLI
func NewQuotaService(cli azurecli.AzureCLI) *QuotaService {
	return &QuotaService{
		cli: cli,
	}
}

// GetFlavorSKUs returns the VM SKUs required for a specific ArcBox flavor
func (q *QuotaService) GetFlavorSKUs(flavor string) []string {
	switch strings.ToLower(flavor) {
	case "itpro":
		return []string{"Standard_D8s_v5"}
	case "devops":
		return []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}
	case "dataops":
		return []string{"Standard_D8s_v5", "Standard_B2ms", "Standard_B8ms", "Standard_B4ms", "Standard_D8s_v4"}
	default:
		return []string{}
	}
}

// CheckQuota performs quota validation and returns structured results and errors
func (q *QuotaService) CheckQuota(locationFlag string, allLocations bool, selectedFlavor string, subscriptionID string) ([]map[string]interface{}, error) {
	selectedFlavor = strings.TrimSpace(selectedFlavor)
	// Normalize flavor for consistent display
	selectedFlavor = arcboxUtils.NormalizeFlavorCase(selectedFlavor)

	// Validate required arguments - either location or all-locations must be specified
	if selectedFlavor == "" {
		return nil, fmt.Errorf("required argument missing: --flavor flag must specify an ArcBox flavor")
	}

	if locationFlag == "" && !allLocations {
		return nil, fmt.Errorf("location specification required: specify either --location or --all-locations")
	}

	if locationFlag != "" && allLocations {
		return nil, fmt.Errorf("conflicting location flags: cannot specify both --location and --all-locations")
	}

	// Get and validate locations early
	var locations []string
	if allLocations {
		// Load all supported ArcBox regions
		var supportedRegions []string
		if err := json.Unmarshal(regions.ArcboxSupportedRegionsData, &supportedRegions); err != nil {
			return nil, fmt.Errorf("supported regions loading failed for quota check: unable to access ArcBox region configuration: %w", err)
		}

		// Convert display names to normalized names
		for _, region := range supportedRegions {
			locations = append(locations, utils.NormalizeRegion(region))
		}
	} else {
		// Parse comma-separated locations
		locationParts := strings.Split(locationFlag, ",")
		for _, loc := range locationParts {
			trimmed := strings.TrimSpace(loc)
			if trimmed != "" {
				locations = append(locations, utils.NormalizeRegion(trimmed))
			}
		}
	}

	// Validate locations immediately
	if err := arcboxUtils.ValidateLocations(locations); err != nil {
		return nil, fmt.Errorf("failed to validate locations for flavor '%s': %w", selectedFlavor, err)
	}

	// Run quota checks for each location with spinner animation
	allPassed := true
	var allResults []map[string]interface{}

	// Show time warning before starting checks (only for table output)
	if utils.OutputFormat == "table" {
		duration := q.GetExpectedDuration(selectedFlavor)
		fmt.Printf("⏱️  %s Note: Quota checks for %s flavor may take %s\n\n",
			utils.InfoColor(""), selectedFlavor, duration)
	}

	for _, location := range locations {
		// Create quota display
		quotaDisplay := display.NewQuotaDisplay()

		// Create location description for spinner
		locationDesc := utils.GetRegionDisplayName(location)
		if selectedFlavor != "" {
			locationDesc = fmt.Sprintf("%s flavor in %s", selectedFlavor, locationDesc)
		}

		// Use spinner for quota checking (only for table output to avoid interfering with other formats)
		var locationPassed bool
		var results []map[string]interface{}

		if utils.OutputFormat == "table" {
			// Run with spinner for better UX
			checkFunc := func() ([]map[string]interface{}, error) {
				_, res := quotaDisplay.RunQuotaChecksWithSubscriptionSilent(q.cli, location, selectedFlavor, subscriptionID)
				return res, nil
			}

			var err error
			results, err = quotaDisplay.RunQuotaCheckWithSpinner(checkFunc, locationDesc)
			if err != nil {
				return allResults, fmt.Errorf("quota check failed for location %s: %w", location, err)
			}

			// Check if all quotas passed for this location
			locationPassed = true
			for _, result := range results {
				if !result["CanDeploy"].(bool) {
					locationPassed = false
					break
				}
			}
		} else {
			// For non-table output, run without spinner to avoid interfering with output formatting
			locationPassed, results = quotaDisplay.RunQuotaChecksWithSubscription(q.cli, location, selectedFlavor, subscriptionID)
		}

		allResults = append(allResults, results...)

		if !locationPassed {
			allPassed = false
		}
	}

	if !allPassed {
		// Analyze the failure reasons to provide more accurate error message
		hasQuotaIssues := false
		hasSKUIssues := false

		for _, result := range allResults {
			if !result["CanDeploy"].(bool) {
				details := result["Details"].(string)
				if strings.Contains(details, "Need") && strings.Contains(details, "more vCPU") {
					hasQuotaIssues = true
				} else if strings.Contains(details, "SKU not available") {
					hasSKUIssues = true
				}
			}
		}

		var errorMsg string
		if hasQuotaIssues && hasSKUIssues {
			errorMsg = fmt.Sprintf("deployment validation failed for flavor '%s': insufficient vCPU quota and SKU availability issues in one or more locations", selectedFlavor)
		} else if hasQuotaIssues {
			errorMsg = fmt.Sprintf("quota validation failed for flavor '%s': insufficient vCPU quota in one or more locations", selectedFlavor)
		} else if hasSKUIssues {
			errorMsg = fmt.Sprintf("deployment validation failed for flavor '%s': SKU availability issues in one or more locations", selectedFlavor)
		} else {
			errorMsg = fmt.Sprintf("deployment validation failed for flavor '%s': deployment requirements not met in one or more locations", selectedFlavor)
		}

		return allResults, fmt.Errorf("%s", errorMsg)
	}

	return allResults, nil
}

// RunQuotaCheckCommand runs the quota check command logic
func (q *QuotaService) RunQuotaCheckCommand(cmd *cobra.Command, args []string) error {
	// Get flag values
	locationFlag, _ := cmd.Flags().GetString("location")
	allLocations, _ := cmd.Flags().GetBool("all-locations")
	selectedFlavor, _ := cmd.Flags().GetString("flavor")
	// Normalize flavor for consistent display
	selectedFlavor = arcboxUtils.NormalizeFlavorCase(selectedFlavor)
	subscriptionID := arcboxUtils.GetSubscriptionID(cmd, q.cli)

	// Run quota checks using the service method
	allResults, err := q.CheckQuota(locationFlag, allLocations, selectedFlavor, subscriptionID)
	if err != nil {
		return fmt.Errorf("quota check failed: %w", err)
	}

	// Handle non-table output formats
	if utils.OutputFormat != "table" {
		headers := []string{"ArcBox Flavor", "Location", "SKU", "vCPU Quota (Available/Limit)", "Required vCPU", "Can Deploy ArcBox?", "Details"}
		var rows [][]string

		for _, result := range allResults {
			canDeploy := "No"
			if result["CanDeploy"].(bool) {
				canDeploy = "Yes"
			}

			quotaDisplay := fmt.Sprintf("%d/%d", result["Available"].(int), result["Limit"].(int))

			row := []string{
				result["Flavor"].(string),
				result["Location"].(string),
				result["SKU"].(string),
				quotaDisplay,
				fmt.Sprintf("%d", result["Required"].(int)),
				canDeploy,
				result["Details"].(string),
			}
			rows = append(rows, row)
		}

		if err := utils.PrintOutput(allResults, headers, rows); err != nil {
			return fmt.Errorf("output formatting failed: %w", err)
		}
	}

	if utils.OutputFormat == "table" {
		fmt.Println(utils.SuccessColor("✅ [SUCCESS] All quota checks passed!"))
	}

	return nil
}

// GetExpectedDuration returns expected duration message based on flavor complexity
func (q *QuotaService) GetExpectedDuration(flavor string) string {
	// Normalize flavor for consistent checking
	flavor = arcboxUtils.NormalizeFlavorCase(flavor)

	switch flavor {
	case "ITPro":
		return "~1 minute (checking 1 SKU)"
	case "DevOps", "DataOps":
		return "3-5 minutes (checking 5 SKUs)"
	case "all":
		return "3-5 minutes (checking all SKUs for multiple flavors)"
	default:
		return "~1 minute"
	}
}
