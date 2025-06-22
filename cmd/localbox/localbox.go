package localbox

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

func buildNormalizedLocalboxRegionMap() map[string]string {
	m := make(map[string]string)
	file := "internal/artifacts/localbox_supported_regions.json"
	data, err := os.ReadFile(file)
	if err == nil {
		var regions []string
		if err := json.Unmarshal(data, &regions); err == nil {
			for _, r := range regions {
				normR := utils.NormalizeRegion(r)
				m[normR] = r
			}
		}
	}
	return m
}

func NewLocalboxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "localbox",
		Short: "Manage Jumpstart LocalBox automation",
		Long:  `(Implementation in progress)`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate ALL flags first (before any other operations)
			if err := utils.ValidateAllFlags(cmd); err != nil {
				os.Exit(1)
			}

			requiredArguments := []string{} // Add required argument names here if needed in the future
			utils.PrintMissingRequiredArgumentsError(cmd, requiredArguments)
			region := ""
			if len(args) > 0 {
				region = args[0]
			}
			if region != "" {
				normMap := buildNormalizedLocalboxRegionMap()
				normRegion := utils.NormalizeRegion(region)
				if _, ok := normMap[normRegion]; !ok {
					utils.Warn("%s is NOT an officially supported region for LocalBox.", region)
					utils.Info("Supported regions are:")
					for _, display := range normMap {
						fmt.Println("  -", display)
					}
					utils.Prompt("Do you want to continue anyway? (y/N): ")
					var response string
					fmt.Scanln(&response)
					response = strings.ToLower(strings.TrimSpace(response))
					if response != "y" && response != "yes" {
						utils.Warn("Aborting command.")
						utils.ShowHelpWithoutTypes(cmd)
						os.Exit(0)
					}
				}
			}

			// Show implementation status message
			fmt.Println(utils.InfoColor("[INFO] LocalBox automation functionality is currently in development."))
		},
	}
}
