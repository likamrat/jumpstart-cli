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
	cli        azurecli.AzureCLI
	quotaCache map[string][]azurecli.VMUsageInfo
}

// NewQuotaService creates a new QuotaService with the provided Azure CLI
func NewQuotaService(cli azurecli.AzureCLI) *QuotaService {
	return &QuotaService{
		cli:        cli,
		quotaCache: make(map[string][]azurecli.VMUsageInfo),
	}
}

// ClearQuotaCache clears the quota cache to ensure fresh data
func (q *QuotaService) ClearQuotaCache() {
	q.quotaCache = make(map[string][]azurecli.VMUsageInfo)
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

	// Run quota checks for each location with configurable output format
	allPassed := true
	var allResults []map[string]interface{}

	for i, location := range locations {
		// Clear quota cache between locations to ensure fresh data
		if i > 0 {
			q.ClearQuotaCache()
		}

		// Create quota display
		quotaDisplay := display.NewQuotaDisplay()

		// Use the new quota checking with Azure CLI wrapper, passing subscription ID directly
		locationPassed, results := quotaDisplay.RunQuotaChecksWithSubscription(q.cli, location, selectedFlavor, subscriptionID)
		allResults = append(allResults, results...)

		if !locationPassed {
			allPassed = false
		}
	}

	if !allPassed {
		return allResults, fmt.Errorf("quota validation failed for flavor '%s': insufficient vCPU quota in one or more locations", selectedFlavor)
	}

	return allResults, nil
}

// RunQuotaCheckCommand runs the quota check command logic
func (q *QuotaService) RunQuotaCheckCommand(cmd *cobra.Command, args []string) error {
	// Get flag values
	locationFlag, _ := cmd.Flags().GetString("location")
	allLocations, _ := cmd.Flags().GetBool("all-locations")
	selectedFlavor, _ := cmd.Flags().GetString("flavor")
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
