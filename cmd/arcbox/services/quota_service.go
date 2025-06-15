package services

import (
	"encoding/json"
	"fmt"
	"os"
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

// RunQuotaCheckCommand runs the quota check command logic
func (q *QuotaService) RunQuotaCheckCommand(cmd *cobra.Command, args []string) {
	// Get flag values
	locationFlag, _ := cmd.Flags().GetString("location")
	allLocations, _ := cmd.Flags().GetBool("all-locations")
	selectedFlavor, _ := cmd.Flags().GetString("flavor")
	selectedFlavor = strings.TrimSpace(selectedFlavor)

	// Validate required arguments - either location or all-locations must be specified
	if selectedFlavor == "" {
		fmt.Print(utils.ErrorColor("❌ [ERROR] Missing required argument: --flavor/-f\n\n"))
		utils.ShowHelpWithoutTypes(cmd)
		os.Exit(1)
	}

	if locationFlag == "" && !allLocations {
		fmt.Print(utils.ErrorColor("❌ [ERROR] Must specify either --location/-l or --all-locations\n\n"))
		utils.ShowHelpWithoutTypes(cmd)
		os.Exit(1)
	}

	if locationFlag != "" && allLocations {
		fmt.Print(utils.ErrorColor("❌ [ERROR] Cannot specify both --location and --all-locations\n\n"))
		utils.ShowHelpWithoutTypes(cmd)
		os.Exit(1)
	}

	// Get and validate locations early
	var locations []string
	if allLocations {
		// Load all supported ArcBox regions
		var supportedRegions []string
		if err := json.Unmarshal(regions.ArcboxSupportedRegionsData, &supportedRegions); err != nil {
			fmt.Printf(utils.ErrorColor("❌ [ERROR] Failed to load supported regions: %v\n"), err)
			os.Exit(1)
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
		fmt.Printf(utils.ErrorColor("❌ [ERROR] %v\n"), err)
		utils.ShowHelpWithoutTypes(cmd)
		os.Exit(1)
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

		if len(locations) > 1 && utils.OutputFormat == "table" {
			fmt.Printf(utils.InfoColor("\n📍 Checking location %d/%d: %s (%s)\n"), i+1, len(locations), location, utils.GetRegionDisplayName(location))
		}

		// Use the new quota checking with Azure CLI wrapper
		locationPassed, results := quotaDisplay.RunQuotaChecksWithOutput(q.cli, cmd, location, selectedFlavor, arcboxUtils.GetSubscriptionID)
		allResults = append(allResults, results...)

		if !locationPassed {
			allPassed = false
			if len(locations) > 1 && utils.OutputFormat == "table" {
				fmt.Printf(utils.ErrorColor("❌ Location %s failed quota validation\n"), location)
			}
		} else if len(locations) > 1 && utils.OutputFormat == "table" {
			fmt.Printf(utils.SuccessColor("✅ Location %s passed quota validation\n"), location)
		}
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
			fmt.Printf(utils.ErrorColor("❌ [ERROR] Failed to format output: %v\n"), err)
			os.Exit(1)
		}
	}

	if !allPassed && utils.OutputFormat == "table" {
		fmt.Println(utils.ErrorColor("\n❌ [ERROR] Quota validation failed for one or more locations. Please resolve the issues above."))
		os.Exit(1)
	}

	if utils.OutputFormat == "table" {
		fmt.Println(utils.SuccessColor("✅ [SUCCESS] All quota checks passed!"))
	}

	// For non-table formats, exit with error code if any checks failed
	if !allPassed {
		os.Exit(1)
	}
}
