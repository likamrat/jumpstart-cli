// arcbox.go - ArcBox command and deployment logic for Jumpstart CLI
package arcbox

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"jumpstartcli/internal/artifacts/regions"
	"jumpstartcli/internal/examples"
	"jumpstartcli/internal/preflight/arcbox"
	"jumpstartcli/internal/preflight/validator"
	"jumpstartcli/internal/resourceproviders"
	"jumpstartcli/internal/table"
	"jumpstartcli/internal/urlutils"
	"jumpstartcli/internal/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Cache to avoid multiple Azure CLI calls for quota data per region
var quotaCache = make(map[string][]map[string]interface{})

// clearQuotaCache clears the quota cache to ensure fresh data
func clearQuotaCache() {
	quotaCache = make(map[string][]map[string]interface{})
}

// Exported for use in main.go
func NewArcboxCmd() *cobra.Command {
	var arcboxCmd = &cobra.Command{
		Use:   "arcbox",
		Short: "Manage Jumpstart ArcBox automation",
		Long: `Manage Jumpstart ArcBox automation resources.

Subcommands:
  • deploy   Deploy a new Jumpstart ArcBox deployment
  • delete   Delete a Jumpstart ArcBox deployment
  • list     List all Jumpstart ArcBox deployments

Use 'js arcbox <subcommand> --help' for more details.`,
		// Disable suggestions to use our custom handling
		DisableSuggestions: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		// Use RunE instead of Args for better control over suggestion handling
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// Check if the first argument matches any subcommand
				validCommands := []string{"deploy", "delete", "list", "preflight"}
				invalidCommand := args[0]

				for _, validCmd := range validCommands {
					if invalidCommand == validCmd {
						return nil // Valid command, continue normal processing
					}
				}

				// Special case for "create" -> "deploy" since they're conceptually the same
				if invalidCommand == "create" {
					utils.PrintDidYouMean(invalidCommand, "deploy")
					return nil
				}

				// General similarity checking with higher threshold
				if suggestion := utils.SuggestSimilarCommand(invalidCommand, validCommands, 3); suggestion != "" {
					utils.PrintDidYouMean(invalidCommand, suggestion)
					return nil
				}

				// No suggestion found, show normal error
				return fmt.Errorf("unknown subcommand '%s' for 'js arcbox'", invalidCommand)
			}
			// If no args, show help
			utils.ShowHelpWithoutTypes(cmd)
			return nil
		},
	}

	// arcbox deploy
	var arcboxDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new Jumpstart ArcBox deployment",
		Long: `Deploy a new Jumpstart ArcBox deployment using remote Bicep or ARM templates from GitHub.

By default, uses the official ArcBox ARM template from GitHub. You can specify:
- Custom remote template URI with --template-uri (ARM templates only - Bicep doesn't support remote templates)
- Local template files with --template-local and --template-params (for local Bicep/ARM templates)

` + examples.GetExamples("arcbox.deploy").FormatExamples(),
		Run: func(cmd *cobra.Command, args []string) {
			// Validate ALL flags first (before any other operations)
			if err := utils.ValidateAllFlags(cmd); err != nil {
				os.Exit(1)
			}

			requiredArguments := []string{"location", "resource-group", "windows-user", "flavor"}
			utils.PrintMissingRequiredArgumentsError(cmd, requiredArguments)

			// Validate conditional requirements (before preflight checks)
			if !arcbox.ValidateConditionalRequirements(cmd) {
				os.Exit(1)
			}

			// Check if preflight checks should be skipped
			skipPreflight := utils.GetBooleanFlagValue(cmd, "skip-pre-flight")
			if skipPreflight {
				fmt.Println(utils.WarnColor("⚠️  [WARNING] Preflight checks have been skipped. Deployment may fail if prerequisites are not met."))
			} else {
				// Run comprehensive preflight checks including parameter validation
				if !arcbox.RunArcBoxPreflightChecks(cmd) {
					fmt.Println(utils.ErrorColor("❌ [ERROR] Preflight checks failed. Please resolve the issues above before proceeding."))
					fmt.Println(utils.InfoColor("💡 [TIP] You can use --skip-preflight to bypass these checks (not recommended)."))
					os.Exit(1)
				}
				// Success message is already printed by PrintResults() in the validation engine
			}

			bicepPath, _ := cmd.Flags().GetString("template-local")
			useParamFile := false // Set based on arguments
			paramFile, _ := cmd.Flags().GetString("template-params")
			deployArcboxWithParamFile(cmd, args, bicepPath, useParamFile, paramFile)
		},
	}

	// arcbox delete
	var arcboxDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a Jumpstart ArcBox deployment",
		Long: `Delete an existing Jumpstart ArcBox deployment by deleting its resource group.

This command will delete the specified resource group and all resources within it.
Use --name to specify the resource group containing your ArcBox deployment.
This operation is irreversible and will permanently remove all ArcBox resources.

` + examples.GetExamples("arcbox.delete").FormatExamples(),
		Run: func(cmd *cobra.Command, args []string) {
			requiredArguments := []string{"name"}
			utils.PrintMissingRequiredArgumentsError(cmd, requiredArguments)

			resourceGroupName, _ := cmd.Flags().GetString("name")
			skipConfirmation, _ := cmd.Flags().GetBool("yes")
			subscription, _ := cmd.Flags().GetString("subscription")

			// Validate Azure CLI is logged in
			if !utils.IsAzureLoggedIn() {
				utils.Error("You are not logged in to Azure. Please run 'az login' and try again.")
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			// Set Azure subscription if provided
			if subscription != "" {
				if err := setAzureSubscription(subscription); err != nil {
					utils.Error("Failed to set subscription: %v", err)
					os.Exit(1)
				}
			}

			// Check if resource group exists
			rgExists, err := checkResourceGroupExists(resourceGroupName, subscription)
			if err != nil {
				utils.Error("Unable to check resource group '%s': %v", resourceGroupName, err)
				utils.Error("Please verify Azure CLI authentication and subscription access.")
				os.Exit(1)
			}

			if !rgExists {
				utils.Error("Resource group '%s' does not exist or you don't have access to it.", resourceGroupName)
				utils.Error("Please check the resource group name and your Azure permissions.")
				os.Exit(1)
			}

			// Confirmation prompt (unless --yes is specified)
			if !skipConfirmation {
				fmt.Printf(utils.WarnColor("⚠️  WARNING: This will permanently delete resource group '%s' and all its resources.\n"), resourceGroupName)
				fmt.Print("Are you sure you want to continue? (y/N): ")
				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Deletion cancelled by user.")
					return
				}
			}

			// Perform deletion
			fmt.Printf(utils.InfoColor("[INFO] Deleting ArcBox resource group '%s'...\n"), resourceGroupName)

			// Use Azure CLI to delete the resource group
			deleteCmd := exec.Command("az", "group", "delete", "--name", resourceGroupName, "--yes", "--no-wait")
			deleteCmd.Stdout = os.Stdout
			deleteCmd.Stderr = os.Stderr

			err = deleteCmd.Run()
			if err != nil {
				utils.Error("Failed to delete resource group '%s': %v", resourceGroupName, err)
				utils.Error("Please check the Azure Portal for more details.")
				os.Exit(1)
			}

			fmt.Printf(utils.SuccessColor("✅ Successfully initiated deletion of resource group '%s'.\n"), resourceGroupName)
			fmt.Println(utils.InfoColor("[INFO] Deletion is running in the background. Check the Azure Portal to monitor progress."))

			// Provide Azure Portal link for monitoring
			if subscription != "" {
				portalUrl := fmt.Sprintf("https://portal.azure.com/#view/HubsExtension/BrowseResource/resourceType/Microsoft.Resources%%2Fresourcegroups")
				fmt.Printf(utils.InfoColor("🔗 [INFO] Monitor deletion progress in the Azure Portal: %s\n"), portalUrl)
			}
		},
	}

	arcboxDeleteCmd.Flags().StringP("name", "n", "", "Resource group name containing the ArcBox deployment to delete")
	arcboxDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt and proceed with deletion")
	arcboxDeleteCmd.Flags().StringP("subscription", "s", "", "Azure subscription ID to use")

	arcboxCmd.AddCommand(arcboxDeleteCmd)

	arcboxCmd.AddCommand(arcboxDeployCmd)

	// arcbox list
	var arcboxListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Jumpstart ArcBox deployments",
		Long: `List all Jumpstart ArcBox deployments across your Azure subscriptions.

Discovers ArcBox deployments by identifying resource groups containing resources with:
- Solution tag "jumpstart_arcbox" (default identification method)
- ArcBox naming prefix (configurable, default: "ArcBox")
- Specific ArcBox resource types (VMs, Key Vaults, etc.)

Requires explicit subscription selection: --current-subscription, --all-subscriptions, or --subscription <id>.

` + examples.GetExamples("arcbox.list").FormatExamples(),
		Run: func(cmd *cobra.Command, args []string) {
			// Validate Azure CLI is logged in
			if !utils.IsAzureLoggedIn() {
				utils.Error("You are not logged in to Azure. Please run 'az login' and try again.")
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			allSubscriptions, _ := cmd.Flags().GetBool("all-subscriptions")
			currentSubscription, _ := cmd.Flags().GetBool("current-subscription")
			subscription, _ := cmd.Flags().GetString("subscription")

			// Check if any subscription selection flag is provided
			flagCount := 0
			if allSubscriptions {
				flagCount++
			}
			if currentSubscription {
				flagCount++
			}
			if subscription != "" {
				flagCount++
			}

			// If no subscription selection flag is provided, show help (like arcbox deploy does for required args)
			if flagCount == 0 {
				fmt.Fprintf(os.Stderr, "%s\n\n", utils.ErrorColor("please specify a subscription selection flag: --current-subscription, --all-subscriptions, or --subscription <id>"))
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			// Validate flag combinations - prevent contradictory flags
			if flagCount > 1 {
				utils.Error("Cannot use multiple subscription selection flags together. Choose one of: --all-subscriptions, --current-subscription, or --subscription")
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			// Validate output format
			if !utils.ValidateOutputFormat(utils.OutputFormat) {
				utils.Error("Invalid output format '%s'. Valid formats are: table, json, yaml, tsv", utils.OutputFormat)
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			// Validate subscription access if specific subscription is provided
			if subscription != "" {
				if _, err := getSubscription(subscription); err != nil {
					utils.Error("Cannot access subscription '%s'. Please verify the subscription ID and your permissions.", subscription)
					os.Exit(1)
				}
			}
			// Run the list operation
			if err := runArcBoxList(allSubscriptions, currentSubscription, subscription, utils.OutputFormat); err != nil {
				utils.Error("Failed to list ArcBox deployments: %v", err)
				os.Exit(1)
			}
		},
	}

	arcboxListCmd.Flags().Bool("all-subscriptions", false, "Search for ArcBox deployments across all accessible subscriptions")
	arcboxListCmd.Flags().Bool("current-subscription", false, "Search for ArcBox deployments in the current subscription")
	arcboxListCmd.Flags().StringP("subscription", "s", "", "Azure subscription ID to search")

	arcboxCmd.AddCommand(arcboxListCmd)

	// arcbox preflight
	var arcboxPreflightCmd = &cobra.Command{
		Use:   "preflight",
		Short: "Run preflight checks for ArcBox deployment",
		Long:  `Run preflight checks to ensure your Azure environment is ready for ArcBox deployment.`,
		// Disable Cobra's built-in suggestions and errors to use our custom ones
		DisableSuggestions: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// List of valid subcommands for arcbox preflight
				validSubcommands := []string{"quota", "status", "rp"}

				// Check if the provided argument is a valid subcommand
				invalidSubcommand := args[0]
				for _, validCmd := range validSubcommands {
					if invalidSubcommand == validCmd {
						return nil // Valid subcommand, continue normal processing
					}
				}

				// If we reach here, it's an invalid subcommand - suggest similar ones
				if suggestion := utils.SuggestSimilarCommand(invalidSubcommand, validSubcommands, 3); suggestion != "" {
					utils.PrintDidYouMean(invalidSubcommand, suggestion)
					return nil
				}

				// No suggestion found, show normal error
				return fmt.Errorf("unknown subcommand '%s' for 'js arcbox preflight'", invalidSubcommand)
			}
			// If no args, show help
			utils.ShowHelpWithoutTypes(cmd)
			return nil
		},
	}

	// arcbox preflight quota
	var arcboxPreflightQuotaCmd = &cobra.Command{
		Use:   "quota",
		Short: "Check vCPU quota for ArcBox flavors",
		Long: `Check if your Azure subscription and region have sufficient vCPU quota for ArcBox ITPro, DevOps, and DataOps flavors.

` + examples.GetExamples("arcbox.preflight.quota").FormatExamples(),
		Run: func(cmd *cobra.Command, args []string) {
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
			if err := validateLocations(locations); err != nil {
				fmt.Printf(utils.ErrorColor("❌ [ERROR] %v\n"), err)
				utils.ShowHelpWithoutTypes(cmd)
				os.Exit(1)
			}

			// Run quota checks for each location with detailed table output
			allPassed := true
			for i, location := range locations {
				// Clear quota cache between locations to ensure fresh data
				if i > 0 {
					validator.ClearQuotaCache()
				}

				if len(locations) > 1 {
					fmt.Printf(utils.InfoColor("\n📍 Checking location %d/%d: %s (%s)\n"), i+1, len(locations), location, utils.GetRegionDisplayName(location))
				}

				// Use detailed table-based quota checking
				locationPassed := runQuotaChecksWithTable(cmd, location, selectedFlavor)
				if !locationPassed {
					allPassed = false
					if len(locations) > 1 {
						fmt.Printf(utils.ErrorColor("❌ Location %s failed quota validation\n"), location)
					}
				} else if len(locations) > 1 {
					fmt.Printf(utils.SuccessColor("✅ Location %s passed quota validation\n"), location)
				}
			}

			if !allPassed {
				fmt.Println(utils.ErrorColor("\n❌ [ERROR] Quota validation failed for one or more locations. Please resolve the issues above."))
				os.Exit(1)
			}

			fmt.Println(utils.SuccessColor("✅ [SUCCESS] All quota checks passed!"))
		},
	}
	arcboxPreflightQuotaCmd.Flags().StringP("flavor", "f", "", "ArcBox flavor to check (ITPro, DevOps, DataOps, all)")
	arcboxPreflightQuotaCmd.Flags().StringP("location", "l", "", "Azure region(s) to check quota in. Use comma-separated values for multiple regions")
	arcboxPreflightQuotaCmd.Flags().Bool("all-locations", false, "Check quota in all ArcBox-supported regions")
	arcboxPreflightQuotaCmd.Flags().String("sku", "", "Custom VM SKU(s) to check. Comma-separated")
	arcboxPreflightQuotaCmd.Flags().StringP("subscription", "s", "", "Azure subscription ID to use")
	arcboxPreflightCmd.AddCommand(arcboxPreflightQuotaCmd)

	// arcbox preflight rp check (enhanced)
	var arcboxPreflightRPCmd = &cobra.Command{
		Use:   "rp",
		Short: "Check and manage Azure resource provider registration",
		Long:  `Check, list, and register required Azure resource providers for ArcBox deployment`,
		// Disable Cobra's built-in suggestions and errors to use our custom ones
		DisableSuggestions: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// List of valid subcommands for arcbox preflight rp
				validSubcommands := []string{"show", "list", "register"}

				// Check if the provided argument is a valid subcommand
				invalidSubcommand := args[0]
				for _, validCmd := range validSubcommands {
					if invalidSubcommand == validCmd {
						return nil // Valid subcommand, continue normal processing
					}
				}

				// If we reach here, it's an invalid subcommand - suggest similar ones
				if suggestion := utils.SuggestSimilarCommand(invalidSubcommand, validSubcommands, 3); suggestion != "" {
					utils.PrintDidYouMean(invalidSubcommand, suggestion)
					return nil
				}

				// No suggestion found, show normal error
				return fmt.Errorf("unknown subcommand '%s' for 'js arcbox preflight rp'", invalidSubcommand)
			}
			// If no args, show help
			utils.ShowHelpWithoutTypes(cmd)
			return nil
		},
	}

	// rp show: check registration status
	var arcboxPreflightRPShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show registration status of required Azure resource providers",
		Long: `Check and display the registration status of all required Azure resource providers for ArcBox deployment.

` + examples.GetExamples("arcbox.preflight.rp.show").FormatExamples(),
		Run: func(cmd *cobra.Command, args []string) {
			config := resourceproviders.GetArcBoxProviders()
			resourceproviders.CheckAllProviders(config)
		},
	}

	// rp list: list required providers
	var arcboxPreflightRPListCmd = &cobra.Command{
		Use:   "list",
		Short: "List required Azure resource providers for ArcBox",
		Long: `List the Azure resource providers required for ArcBox deployment (names only).

` + examples.GetExamples("arcbox.preflight.rp.list").FormatExamples(),
		Run: func(cmd *cobra.Command, args []string) {
			config := resourceproviders.GetArcBoxProviders()
			resourceproviders.ListProviders(config)
		},
	}

	// rp register: register a provider
	var arcboxPreflightRPRegisterCmd = &cobra.Command{
		Use:   "register",
		Short: "Register a required Azure resource provider",
		Long: `Register a required Azure resource provider for ArcBox deployment.

` + examples.GetExamples("arcbox.preflight.rp.register").FormatExamples(),
		Run: func(cmd *cobra.Command, args []string) {
			requiredArguments := []string{"name"}
			utils.PrintMissingRequiredArgumentsError(cmd, requiredArguments)

			provider, _ := cmd.Flags().GetString("name")
			if err := resourceproviders.RegisterProvider(provider); err != nil {
				os.Exit(1)
			}
		},
	}

	arcboxPreflightRPRegisterCmd.Flags().StringP("name", "n", "", "Azure resource provider name to register")

	arcboxPreflightRPCmd.AddCommand(arcboxPreflightRPListCmd)
	arcboxPreflightRPCmd.AddCommand(arcboxPreflightRPRegisterCmd)
	arcboxPreflightRPCmd.AddCommand(arcboxPreflightRPShowCmd)
	arcboxPreflightCmd.AddCommand(arcboxPreflightRPCmd)

	// arcbox preflight status (stub)
	var arcboxPreflightStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show last preflight check status",
		Long:  `Show the results of the last ArcBox preflight check (not yet implemented).`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(utils.InfoColor("[INFO] Preflight status reporting is currently in development."))
		},
	}
	arcboxPreflightCmd.AddCommand(arcboxPreflightStatusCmd)

	arcboxCmd.AddCommand(arcboxPreflightCmd)

	// Define arguments for arcbox deploy command in alphabetical order
	arcboxDeployCmd.Flags().String("admin-username", "", "Admin username for Linux virtual machines that are part of the ArcBox deployment. Overrides windows-user if set")
	arcboxDeployCmd.Flags().String("auto-shutdown", "yes", "Enable automatic shutdown for the ArcBox deployment Client virtual machine to save costs")
	arcboxDeployCmd.Flags().String("auto-shutdown-time", "1800", "Automatic shutdown time in 24-hour format (HHMM). Only used when auto-shutdown is enabled")
	arcboxDeployCmd.Flags().String("auto-shutdown-timezone", "UTC", "Timezone for automatic shutdown time. Only used when auto-shutdown is enabled")
	arcboxDeployCmd.Flags().String("auto-shutdown-email", "", "Email address to notify when automatic shutdown occurs. Only used when auto-shutdown is enabled")
	arcboxDeployCmd.Flags().String("bastion-sku", "Basic", "Bastion host SKU name. Only used when deploy-bastion is true")
	arcboxDeployCmd.Flags().String("deploy-bastion", "no", "Deploy Azure Bastion to connect to the ArcBox deployment Client virtual machine. Provides secure RDP/SSH access through the Azure portal")
	arcboxDeployCmd.Flags().String("enable-spot-pricing", "no", "Enable spot pricing for the ArcBox Client VM to reduce costs")
	arcboxDeployCmd.Flags().StringP("flavor", "f", "ITPro", "ArcBox flavor")
	arcboxDeployCmd.Flags().String("vm-autologon", "yes", "Enable automatic logon into the ArcBox deployment Client virtual machine. Simplifies access to the ArcBox deployment Client virtual machine")
	arcboxDeployCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts and proceed automatically. Useful for automation and CI/CD pipelines")
	arcboxDeployCmd.Flags().String("github-user", "microsoft", "GitHub username for the jumpstart-apps repository. Only change when deploying the DevOps flavor")
	arcboxDeployCmd.Flags().StringP("location", "l", "", "Azure region for deployment")
	arcboxDeployCmd.Flags().String("log-analytics-workspace", "", "Log Analytics workspace name for monitoring")
	arcboxDeployCmd.Flags().String("naming-prefix", "ArcBox", "Naming prefix for the Azure resources deployed. Max 7 chars")
	arcboxDeployCmd.Flags().String("rdp-port", "3389", "Override default RDP port. No changes will be made to the ArcBox deployment Client virtual machine")
	arcboxDeployCmd.Flags().StringP("resource-group", "g", "", "Resource group name for deployment")
	arcboxDeployCmd.Flags().String("resource-tags", `{"Solution":"jumpstart_arcbox"}`, "Resource tags to assign to all ArcBox resources")
	arcboxDeployCmd.Flags().String("skip-preflight", "no", "Skip preflight checks. Azure CLI health, resource providers, region validation, resource group existence, quota checks")
	arcboxDeployCmd.Flags().String("sql-server-edition", "Developer", "SQL Server edition")
	arcboxDeployCmd.Flags().String("ssh-rsa-public-key", "", "SSH RSA public key for Linux VMs. Required for DevOps/DataOps flavors")
	arcboxDeployCmd.Flags().StringP("subscription", "s", "", "Azure subscription ID to use")
	arcboxDeployCmd.Flags().String("template-local", "", "Local Bicep or ARM template file path. Overrides default remote template")
	arcboxDeployCmd.Flags().String("template-params", "", "Local parameters file path. Used with local template deployments")
	arcboxDeployCmd.Flags().String("template-uri", "", "Remote ARM template URI. By default, uses the official Jumpstart ArcBox template from https://github.com/microsoft/azure_arc/blob/main/azure_jumpstart_arcbox/ARM/azuredeploy.json. Note: Bicep templates cannot be remote")
	arcboxDeployCmd.Flags().String("windows-password", "", "ArcBox deployment Client virtual machine password. Password must have 3 of the following: 1 lower case character, 1 upper case character, 1 number, and 1 special character. The value must be between 12 and 123 characters long. If not specified, the default value is generated using the Bicep newGuid() function and stored in the Key Vault")
	arcboxDeployCmd.Flags().String("windows-user", "", "ArcBox deployment Client virtual machine Administrator username")

	return arcboxCmd
}

func deployArcboxWithParamFile(cmd *cobra.Command, args []string, _ string, _ bool, _ string) {
	defaultRemote := "https://raw.githubusercontent.com/microsoft/azure_arc/main/azure_jumpstart_arcbox/ARM/azuredeploy.json"
	templateLocalPath, _ := cmd.Flags().GetString("template-local")
	templateLocalParamPath, _ := cmd.Flags().GetString("template-params")
	templateURI, _ := cmd.Flags().GetString("template-uri")
	bicepPath := ""
	useParamFile := false
	paramFile := ""

	if templateLocalPath != "" {
		bicepPath = templateLocalPath
		if templateLocalParamPath != "" {
			useParamFile = true
			paramFile = templateLocalParamPath
		}
	} else if templateURI != "" {
		bicepPath = templateURI
	} else {
		bicepPath = defaultRemote
	}

	resourceGroup, _ := cmd.Flags().GetString("resource-group")
	location, _ := cmd.Flags().GetString("location")
	windowsAdminUsername, _ := cmd.Flags().GetString("windows-user")
	windowsAdminPassword, _ := cmd.Flags().GetString("windows-password")
	flavorRaw, _ := cmd.Flags().GetString("flavor")
	sqlServerEditionRaw, _ := cmd.Flags().GetString("sql-server-edition")

	// Normalize flavor to proper case for ARM template (case-insensitive input support)
	flavor := normalizeFlavorCase(flavorRaw)

	// Normalize SQL Server edition to proper case (case-insensitive input support)
	sqlServerEdition := normalizeSqlServerEditionCase(sqlServerEditionRaw)

	sshRSAPublicKey, _ := cmd.Flags().GetString("ssh-rsa-public-key")
	logAnalyticsWorkspaceName, _ := cmd.Flags().GetString("log-analytics-workspace")
	namingPrefix, _ := cmd.Flags().GetString("naming-prefix")
	adminUsername, _ := cmd.Flags().GetString("admin-username")
	if adminUsername == "" {
		adminUsername = windowsAdminUsername
	}

	// Get deployment options
	resourceTags, _ := cmd.Flags().GetString("resource-tags")

	// Use global yes/no boolean flag parsing pattern
	autoShutdownEnabled := utils.GetBooleanFlagValue(cmd, "auto-shutdown")
	vmAutologon := utils.GetBooleanFlagValue(cmd, "vm-autologon")
	deployBastion := utils.GetBooleanFlagValue(cmd, "deploy-bastion")
	enableSpotPricing := utils.GetBooleanFlagValue(cmd, "enable-spot-pricing")

	autoShutdownTime, _ := cmd.Flags().GetString("auto-shutdown-time")
	autoShutdownTimezone, _ := cmd.Flags().GetString("auto-shutdown-timezone")
	autoShutdownEmail, _ := cmd.Flags().GetString("auto-shutdown-email")
	githubUser, _ := cmd.Flags().GetString("github-user")
	rdpPort, _ := cmd.Flags().GetString("rdp-port")
	bastionSkuRaw, _ := cmd.Flags().GetString("bastion-sku")

	// Normalize Bastion SKU to proper case (case-insensitive input support)
	bastionSku := normalizeBastionSkuCase(bastionSkuRaw)

	yesFlag, _ := cmd.Flags().GetBool("yes")

	if utils.DebugMode {
		utils.Info("Debug mode enabled")
	}

	fmt.Println(utils.InfoColor("[INFO] Starting Jumpstart ArcBox deployment..."))
	utils.Info("Jumpstart CLI version %s", utils.CliVersion)
	utils.Info("ArcBox Selected flavor: %s", flavor)
	utils.Info("Resource Group: %s", resourceGroup)
	utils.Info("Location: %s", location)
	utils.Info("Bicep Template: %s", bicepPath)

	if utils.DebugMode {
		debugParams := fmt.Sprintf("windowsAdminUsername=%s, adminUsername=%s, windowsAdminPassword=****, flavor=%s, sqlServerEdition=%s, logAnalyticsWorkspaceName=%s, namingPrefix=%s", windowsAdminUsername, adminUsername, flavor, sqlServerEdition, logAnalyticsWorkspaceName, namingPrefix)
		if (flavor == "DevOps" || flavor == "DataOps") && sshRSAPublicKey != "" {
			debugParams += fmt.Sprintf(", sshRSAPublicKey=%s", sshRSAPublicKey)
		}
		utils.Info("Parameters: %s", debugParams)
	}

	if len(bicepPath) > 4 && (bicepPath[:4] == "http") {
		_, _ = exec.LookPath("curl") // ignore error, just check presence
	}

	if resourceGroup == "" || windowsAdminUsername == "" || windowsAdminPassword == "" {
		utils.Error("Usage: js arcbox deploy --resource-group <name> --windows-user <username> --windows-password <password> [other arguments]")
		utils.ShowHelpWithoutTypes(cmd)
		os.Exit(1)
	}

	if !utils.IsAzureLoggedIn() {
		utils.Error("You are not logged in to Azure. Please run 'az login' and try again.")
		utils.ShowHelpWithoutTypes(cmd)
		os.Exit(1)
	}

	// Security warnings for VM-related settings (unless --yes is specified)
	if !yesFlag {
		// Warning for VM autologon enabled
		if vmAutologon {
			fmt.Print(utils.WarnColor("⚠️  WARNING: VM autologon is enabled. This reduces security by automatically logging in users.\n"))
			fmt.Print("Are you sure you want to continue with autologon enabled? (y/N): ")
			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Deployment cancelled by user.")
				return
			}
		}

		// Warning for auto-shutdown disabled
		if !autoShutdownEnabled {
			fmt.Print(utils.WarnColor("⚠️  WARNING: Auto-shutdown is disabled. This may result in unexpected costs from running VMs.\n"))
			fmt.Print("Are you sure you want to continue without auto-shutdown? (y/N): ")
			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Deployment cancelled by user.")
				return
			}
		}
	}

	if !utils.ResourceGroupExists(resourceGroup) {
		utils.Info("Resource group '%s' does not exist. Creating it...", resourceGroup)
		err := utils.CreateResourceGroup(resourceGroup, location)
		if err != nil {
			utils.Error("Failed to create resource group: %v", err)
			utils.ShowHelpWithoutTypes(cmd)
			os.Exit(1)
		}
	}

	var azArgs []string
	if useParamFile && paramFile != "" {
		azArgs = []string{"deployment", "group", "create", "--resource-group", resourceGroup, "--template-file", bicepPath, "--parameters", "@" + paramFile}
	} else if len(bicepPath) > 4 && (bicepPath[:4] == "http") {
		params := []string{}
		if windowsAdminUsername != "" {
			params = append(params, fmt.Sprintf("windowsAdminUsername=%s", windowsAdminUsername))
		}
		if windowsAdminPassword != "" {
			params = append(params, fmt.Sprintf("windowsAdminPassword=%s", windowsAdminPassword))
		}
		if flavor != "" {
			params = append(params, fmt.Sprintf("flavor=%s", flavor))
		}
		if sqlServerEdition != "" {
			params = append(params, fmt.Sprintf("sqlServerEdition=%s", sqlServerEdition))
		}
		if logAnalyticsWorkspaceName != "" {
			params = append(params, fmt.Sprintf("logAnalyticsWorkspaceName=%s", logAnalyticsWorkspaceName))
		}
		if namingPrefix != "" {
			params = append(params, fmt.Sprintf("namingPrefix=%s", namingPrefix))
		}
		if (flavor == "DevOps" || flavor == "DataOps") && sshRSAPublicKey != "" {
			params = append(params, fmt.Sprintf("sshRSAPublicKey=%s", sshRSAPublicKey))
		}

		// Add new parameters
		params = append(params, fmt.Sprintf("deployBastion=%t", deployBastion))
		params = append(params, fmt.Sprintf("resourceTags=%s", resourceTags))
		params = append(params, fmt.Sprintf("vmAutologon=%t", vmAutologon))
		params = append(params, fmt.Sprintf("autoShutdownEnabled=%t", autoShutdownEnabled))
		if autoShutdownEnabled {
			params = append(params, fmt.Sprintf("autoShutdownTime=%s", autoShutdownTime))
			params = append(params, fmt.Sprintf("autoShutdownTimezone=%s", autoShutdownTimezone))
			if autoShutdownEmail != "" {
				params = append(params, fmt.Sprintf("autoShutdownEmailRecipient=%s", autoShutdownEmail))
			}
		}
		params = append(params, fmt.Sprintf("enableAzureSpotPricing=%t", enableSpotPricing))
		params = append(params, fmt.Sprintf("githubUser=%s", githubUser))
		params = append(params, fmt.Sprintf("rdpPort=%s", rdpPort))
		if deployBastion {
			params = append(params, fmt.Sprintf("bastionSku=%s", bastionSku))
		}
		azArgs = []string{"deployment", "group", "create", "--resource-group", resourceGroup, "--template-uri", bicepPath}
		if len(params) > 0 {
			azArgs = append(azArgs, "--parameters")
			azArgs = append(azArgs, params...)
		}
	} else {
		params := []string{}
		if windowsAdminUsername != "" {
			params = append(params, fmt.Sprintf("windowsAdminUsername=%s", windowsAdminUsername))
		}
		if windowsAdminPassword != "" {
			params = append(params, fmt.Sprintf("windowsAdminPassword=%s", windowsAdminPassword))
		}
		if flavor != "" {
			params = append(params, fmt.Sprintf("flavor=%s", flavor))
		}
		if sqlServerEdition != "" {
			params = append(params, fmt.Sprintf("sqlServerEdition=%s", sqlServerEdition))
		}
		if logAnalyticsWorkspaceName != "" {
			params = append(params, fmt.Sprintf("logAnalyticsWorkspaceName=%s", logAnalyticsWorkspaceName))
		}
		if namingPrefix != "" {
			params = append(params, fmt.Sprintf("namingPrefix=%s", namingPrefix))
		}
		if (flavor == "DevOps" || flavor == "DataOps") && sshRSAPublicKey != "" {
			params = append(params, fmt.Sprintf("sshRSAPublicKey=%s", sshRSAPublicKey))
		}

		// Add new parameters
		params = append(params, fmt.Sprintf("deployBastion=%t", deployBastion))
		params = append(params, fmt.Sprintf("resourceTags=%s", resourceTags))
		params = append(params, fmt.Sprintf("vmAutologon=%t", vmAutologon))
		params = append(params, fmt.Sprintf("autoShutdownEnabled=%t", autoShutdownEnabled))
		if autoShutdownEnabled {
			params = append(params, fmt.Sprintf("autoShutdownTime=%s", autoShutdownTime))
			params = append(params, fmt.Sprintf("autoShutdownTimezone=%s", autoShutdownTimezone))
			if autoShutdownEmail != "" {
				params = append(params, fmt.Sprintf("autoShutdownEmailRecipient=%s", autoShutdownEmail))
			}
		}
		params = append(params, fmt.Sprintf("enableAzureSpotPricing=%t", enableSpotPricing))
		params = append(params, fmt.Sprintf("githubUser=%s", githubUser))
		params = append(params, fmt.Sprintf("rdpPort=%s", rdpPort))
		if deployBastion {
			params = append(params, fmt.Sprintf("bastionSku=%s", bastionSku))
		}
		azArgs = []string{"deployment", "group", "create", "--resource-group", resourceGroup, "--template-file", bicepPath}
		if len(params) > 0 {
			azArgs = append(azArgs, "--parameters")
			azArgs = append(azArgs, params...)
		}
	}

	// Generate a unique deployment name
	deploymentName := fmt.Sprintf("arcbox-%d", time.Now().Unix())
	azArgs = append(azArgs, "--name", deploymentName)
	azArgs = append(azArgs, "--no-wait")

	azCmd := exec.Command("az", azArgs...)
	azCmd.Stdout = os.Stdout
	azCmd.Stderr = os.Stderr
	err := azCmd.Run()
	if err != nil {
		utils.Error("Error starting az deployment: %v", err)
		fmt.Println("Please check your parameters, resource group, and Azure login status.")
		portalUrl := fmt.Sprintf("https://portal.azure.com/#view/HubsExtension/BrowseResource/resourceType/Microsoft.Resources%%2Fdeployments/resourceGroup/%s", resourceGroup)
		fmt.Printf("View failed deployment details in the Azure Portal: %s\n", portalUrl)
		os.Exit(1)
	}

	// Get subscription ID for portal link
	subscription := getSubscriptionID(cmd)

	// Provide Azure Portal link for tracking deployment
	if subscription != "" {
		portalUrl := fmt.Sprintf("https://portal.azure.com/#view/HubsExtension/DeploymentDetailsBlade/~/overview/id/%%2Fsubscriptions%%2F%s%%2FresourceGroups%%2F%s%%2Fproviders%%2FMicrosoft.Resources%%2Fdeployments%%2F%s", subscription, resourceGroup, deploymentName)

		// Try to shorten the URL (with a brief indication)
		fmt.Print(utils.InfoColor("🔗 [INFO] Generating portal link... "))
		shortURL := urlutils.ShortenURL(portalUrl)
		fmt.Print("\r\033[2K") // Clear the "generating" message

		fmt.Println(utils.InfoColor("🔗 [INFO] Track your deployment in the Azure Portal:"))
		fmt.Printf("   %s\n", shortURL)
		fmt.Println()
	}

	waitForDeploymentAndShowStatus(resourceGroup, deploymentName)
}

// resourceStatus holds resource info for status output
// Used by getDeploymentResourceStatus and printDeploymentResourceList
type resourceStatus struct {
	Name  string
	Type  string
	State string
}

// getDeploymentResourceStatus returns a slice of resourceStatus for the resource group
func getDeploymentResourceStatus(resourceGroup string) []resourceStatus {
	cmd := exec.Command("az", "resource", "list", "--resource-group", resourceGroup, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	var resources []map[string]interface{}
	if err := json.Unmarshal(output, &resources); err != nil {
		return nil
	}
	var result []resourceStatus
	for _, res := range resources {
		name, _ := res["name"].(string)
		typeStr, _ := res["type"].(string)
		id, _ := res["id"].(string)
		// Get detailed state for each resource
		showCmd := exec.Command("az", "resource", "show", "--ids", id, "-o", "json")
		showOut, showErr := showCmd.Output()
		provState := "Unknown"
		if showErr == nil {
			var showRes map[string]interface{}
			if err := json.Unmarshal(showOut, &showRes); err == nil {
				if props, ok := showRes["properties"].(map[string]interface{}); ok {
					if ps, ok := props["provisioningState"].(string); ok {
						provState = ps
					}
				}
			}
		}
		result = append(result, resourceStatus{
			Name:  name,
			Type:  typeStr,
			State: provState,
		})
	}
	return result
}

// printDeploymentResourceList prints a summary of resources and their status
func printDeploymentResourceList(resources []resourceStatus) {
	for _, res := range resources {
		// Hide DevTestLab schedules (e.g. auto-shutdown)
		if res.Type == "Microsoft.DevTestLab/schedules" {
			continue
		}
		icon := "❓"
		// Set emoji for each state
		switch res.State {
		case "Succeeded":
			icon = "✅"
		case "Failed":
			icon = "❌"
		case "Running", "Creating", "Accepted", "InProgress":
			icon = "⌛"
		case "Updating":
			icon = "🔄"
		case "Deleting":
			icon = "🗑️"
		}
		if res.Type == "Microsoft.Compute/virtualMachines/extensions" {
			// For VM Extensions, show only the extension name as resource name, and 'VM Extension' as friendly type
			parts := strings.Split(res.Name, "/")
			extName := parts[len(parts)-1]
			fval := "VM Extension"
			// Map extension name to friendly name if needed
			if extName == "Microsoft.Azure.Geneva.GenevaMonitoring" {
				extName = "Azure Geneva Monitoring"
			}
			msg := fmt.Sprintf("%s \"%s\" %s: %s", icon, extName, fval, res.State)
			fmt.Println(msg)
			continue
		}
		fval := utils.FriendlyResourceName(res.Type, res.Name)
		msg := fmt.Sprintf("%s \"%s\" %s: %s", icon, res.Name, fval, res.State)
		fmt.Println(msg)
	}
}

// waitForDeploymentAndShowStatus polls deployment status and prints resource-level operations
func waitForDeploymentAndShowStatus(resourceGroup, deploymentName string) {
	start := time.Now()
	cState := color.New(color.FgHiBlue, color.Bold).SprintFunc()
	cWarn := color.New(color.FgHiYellow, color.Bold).SprintFunc()

	// Animation frames: Unicode spinner for smooth animation (consistent with preflight validation)
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIdx := 0
	stopSpinner := make(chan struct{})
	spinnerDone := make(chan struct{})

	// Hide cursor before starting animation
	fmt.Print("\033[?25l")
	go func() {
		for {
			select {
			case <-stopSpinner:
				fmt.Printf("\r\033[2K") // clear spinner line
				// Restore cursor when animation stops
				fmt.Print("\033[?25h")
				close(spinnerDone)
				return
			default:
				// Print: 🚀 Deployment in progress... (with Unicode spinner)
				fmt.Printf("\r\033[2K🚀 Deployment in progress... %s\033[0K", frames[frameIdx%len(frames)])
				frameIdx++
				time.Sleep(80 * time.Millisecond) // Smooth 80ms animation (consistent with preflight validation)
			}
		}
	}()

	pollInterval := 5 * time.Second
	nextPoll := time.Now()
	var (
		state                 string
		resources             []resourceStatus
		resourceNames         map[string]bool
		allSucceeded          bool
		anyFailed             bool
		inProgress            int
		lastResourceCount     int
		lastProvisioningState string
	)
	lastResourceStates := make(map[string]string)
	for {
		if time.Now().After(nextPoll) {
			state = getDeploymentProvisioningState(resourceGroup, deploymentName)
			resources = getDeploymentResourceStatus(resourceGroup)
			resourceNames = make(map[string]bool)
			stateChanged := false
			for _, r := range resources {
				resourceNames[r.Name] = true
				if lastResourceStates[r.Name] != r.State {
					stateChanged = true
				}
			}
			allSucceeded = true
			anyFailed = false
			inProgress = 0
			for _, r := range resources {
				if r.State == "Failed" {
					anyFailed = true
				}
				if r.State != "Succeeded" {
					allSucceeded = false
				}
				if r.State == "Creating" || r.State == "Updating" || r.State == "Running" || r.State == "InProgress" || r.State == "Deleting" || r.State == "Unknown" {
					inProgress++
				}
			}

			// Print the resource list if any state changed or a new resource is detected
			if len(resources) == 0 {
				fmt.Printf("\r\033[2K") // clear spinner line
				fmt.Println(utils.InfoColor("[INFO] Waiting for resources to appear in the resource group..."))
			} else if stateChanged || len(resourceNames) > lastResourceCount {
				fmt.Printf("\r\033[2K") // clear spinner line
				fmt.Println()
				printDeploymentResourceList(resources)
				lastResourceCount = len(resourceNames)
				lastResourceStates = make(map[string]string)
				for _, r := range resources {
					lastResourceStates[r.Name] = r.State
				}
			}

			if anyFailed || state == "Failed" || state == "Canceled" {
				close(stopSpinner)
				<-spinnerDone
				// Show cursor again
				fmt.Print("\033[?25h")
				fmt.Println() // Move to new line before error
				printDeploymentErrorDetails(resourceGroup, deploymentName)
				fmt.Println(utils.ErrorColor("[ERROR] Deployment failed or was canceled, or a resource failed. Please check the Azure Portal for details."))
				break
			}
			if state == "Succeeded" && allSucceeded {
				close(stopSpinner)
				<-spinnerDone
				// Show cursor again
				fmt.Print("\033[?25h")
				// Print final resource list with all resources marked as Succeeded
				for i := range resources {
					resources[i].State = "Succeeded"
				}
				fmt.Println()
				printDeploymentResourceList(resources)

				// Get deployment duration from Azure as source of truth
				azureDuration, err := getAzureDeploymentDuration(resourceGroup, deploymentName)
				if err != nil {
					// Fallback to local timing if Azure timing is unavailable
					elapsed := time.Since(start)
					fmt.Printf("\n🎁 ArcBox successfully deployed! (deployment took %s)\n", elapsed.Truncate(time.Second))
				} else {
					// Use Azure's official deployment timing
					fmt.Printf("\n🎁 ArcBox successfully deployed! (Azure deployment duration: %s)\n", azureDuration.Truncate(time.Second))
				}
				break
			}
			if state != lastProvisioningState {
				if !(state == "Succeeded" && !allSucceeded) {
					fmt.Printf("\r\033[2K") // clear spinner line
					fmt.Printf("%s [%s] Deployment state: %s\n", cState("[STATUS]"), time.Now().Format("15:04:05"), state)
					lastProvisioningState = state
				}
			}
			nextPoll = time.Now().Add(pollInterval)
		}
		if time.Since(start) > 30*time.Minute {
			fmt.Println(cWarn("\n[WARN] Deployment is taking longer than 30 minutes. Please check the Azure Portal for more details."))
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Print("\n")
}

// ArcBoxDeployment represents an ArcBox deployment discovered in a resource group
type ArcBoxDeployment struct {
	ResourceGroupName string
	SubscriptionID    string
	SubscriptionName  string
	Location          string
	CreatedDate       string
	Status            string
	ResourceCount     int
	Flavor            string
	NamingPrefix      string
}

// runArcBoxList discovers and lists ArcBox deployments
func runArcBoxList(allSubscriptions, currentSubscription bool, subscriptionID, outputFormat string) error {
	var subscriptions []AzureSubscription
	var err error

	if allSubscriptions {
		fmt.Println(utils.InfoColor("[INFO] Searching for ArcBox deployments across all subscriptions..."))
		subscriptions, err = getAllSubscriptions()
	} else if subscriptionID != "" {
		fmt.Printf(utils.InfoColor("[INFO] Searching for ArcBox deployments in subscription %s...\n"), subscriptionID)
		sub, err := getSubscription(subscriptionID)
		if err != nil {
			return err
		}
		subscriptions = []AzureSubscription{sub}
	} else if currentSubscription {
		// Explicit --current-subscription flag
		fmt.Println(utils.InfoColor("[INFO] Searching for ArcBox deployments in current subscription..."))
		sub, err := getCurrentSubscription()
		if err != nil {
			return err
		}
		subscriptions = []AzureSubscription{sub}
	}

	if err != nil {
		return err
	}

	var allDeployments []ArcBoxDeployment
	for _, sub := range subscriptions {
		// Scan each subscription with spinner animation
		deployments := scanSubscriptionWithSpinner(sub)
		allDeployments = append(allDeployments, deployments...)
	}

	if len(allDeployments) == 0 {
		fmt.Println(utils.InfoColor("🔍 No ArcBox deployments found."))
		if !allSubscriptions && subscriptionID == "" {
			fmt.Println("💡 Tip: Use --all-subscriptions to search across all accessible subscriptions.")
		}
		return nil
	}

	// Output results
	switch strings.ToLower(outputFormat) {
	case "json":
		return outputArcBoxDeploymentsJSON(allDeployments)
	case "table":
		fallthrough
	default:
		return outputArcBoxDeploymentsTable(allDeployments)
	}
}

// AzureSubscription represents an Azure subscription
type AzureSubscription struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// getAllSubscriptions returns all accessible Azure subscriptions
func getAllSubscriptions() ([]AzureSubscription, error) {
	cmd := exec.Command("az", "account", "list", "--query", "[].{id:id,name:name}", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %v", err)
	}

	var subscriptions []AzureSubscription
	if err := json.Unmarshal(output, &subscriptions); err != nil {
		return nil, fmt.Errorf("failed to parse subscriptions: %v", err)
	}

	return subscriptions, nil
}

// getCurrentSubscription returns the current Azure subscription
func getCurrentSubscription() (AzureSubscription, error) {
	cmd := exec.Command("az", "account", "show", "--query", "{id:id,name:name}", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return AzureSubscription{}, fmt.Errorf("failed to get current subscription: %v", err)
	}

	var subscription AzureSubscription
	if err := json.Unmarshal(output, &subscription); err != nil {
		return AzureSubscription{}, fmt.Errorf("failed to parse current subscription: %v", err)
	}

	return subscription, nil
}

// getSubscription returns information about a specific subscription
func getSubscription(subscriptionID string) (AzureSubscription, error) {
	cmd := exec.Command("az", "account", "show", "--subscription", subscriptionID, "--query", "{id:id,name:name}", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return AzureSubscription{}, fmt.Errorf("failed to get subscription %s: %v", subscriptionID, err)
	}

	var subscription AzureSubscription
	if err := json.Unmarshal(output, &subscription); err != nil {
		return AzureSubscription{}, fmt.Errorf("failed to parse subscription %s: %v", subscriptionID, err)
	}

	return subscription, nil
}

// discoverArcBoxDeployments discovers ArcBox deployments in a subscription
func discoverArcBoxDeployments(subscriptionID, subscriptionName string) ([]ArcBoxDeployment, error) {
	// Set subscription context
	if err := setAzureSubscription(subscriptionID); err != nil {
		return nil, fmt.Errorf("failed to set subscription context: %v", err)
	}

	// Get all resource groups
	cmd := exec.Command("az", "group", "list", "--query", "[].{name:name,location:location,properties:properties}", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list resource groups: %v", err)
	}

	var resourceGroups []map[string]interface{}
	if err := json.Unmarshal(output, &resourceGroups); err != nil {
		return nil, fmt.Errorf("failed to parse resource groups: %v", err)
	}

	// Silently process resource groups (spinner provides visual feedback)
	// fmt.Printf("📋 Found %d resource groups to check\n", len(resourceGroups))

	// Pre-filter resource groups to exclude obvious non-ArcBox ones
	var candidateRGs []map[string]interface{}
	for _, rg := range resourceGroups {
		rgName, _ := rg["name"].(string)
		rgNameLower := strings.ToLower(rgName)

		// Skip obviously non-ArcBox resource groups
		if strings.Contains(rgNameLower, "localbox") ||
			strings.HasPrefix(rgName, "MC_") ||
			strings.HasPrefix(rgName, "DefaultResourceGroup-") ||
			strings.HasPrefix(rgName, "NetworkWatcherRG") ||
			strings.HasPrefix(rgName, "AzSecPackAutoConfigRG") {
			continue
		}

		candidateRGs = append(candidateRGs, rg)
	}

	// Silently pre-filter candidate resource groups (spinner provides visual feedback)
	// fmt.Printf("🔍 Pre-filtered to %d candidate resource groups\n", len(candidateRGs))

	var deployments []ArcBoxDeployment
	for _, rg := range candidateRGs {
		rgName, _ := rg["name"].(string)
		location, _ := rg["location"].(string)

		// Silently check resource groups (spinner provides visual feedback)

		// Check if this resource group contains an ArcBox deployment
		if isArcBoxResourceGroup(rgName) {
			deployment := ArcBoxDeployment{
				ResourceGroupName: rgName,
				SubscriptionID:    subscriptionID,
				SubscriptionName:  subscriptionName,
				Location:          location,
			}

			// Enrich deployment information
			enrichArcBoxDeployment(&deployment)
			deployments = append(deployments, deployment)
			// Silently collect deployments (results shown after spinner completes)
		}
	}

	return deployments, nil
}

// isArcBoxResourceGroup checks if a resource group contains an ArcBox deployment
func isArcBoxResourceGroup(resourceGroupName string) bool {
	// First, exclude resource groups that are clearly LocalBox deployments
	resourceGroupNameLower := strings.ToLower(resourceGroupName)
	if strings.Contains(resourceGroupNameLower, "localbox") {
		return false
	}

	// Exclude AKS-managed resource groups (automatically created by Azure Kubernetes Service)
	if strings.HasPrefix(resourceGroupName, "MC_") {
		return false
	}

	// Quick check: if the resource group name contains "arcbox", it's likely an ArcBox deployment
	if strings.Contains(resourceGroupNameLower, "arcbox") {
		return true
	}

	// Method 1: Check for resources with ArcBox solution tag (most reliable but slower)
	if hasArcBoxSolutionTag(resourceGroupName) {
		return true
	}

	// Method 2: Check for ArcBox deployments by name pattern (faster)
	if hasArcBoxDeployments(resourceGroupName) {
		return true
	}

	// Method 3: Check for ArcBox naming patterns in resources (slower)
	if hasArcBoxNamingPattern(resourceGroupName) {
		return true
	}

	// Do NOT use hasArcBoxResources as it's too broad and causes false positives
	// with LocalBox and other deployments that have VMs, VNets, and Key Vaults

	return false
}

// hasArcBoxSolutionTag checks for resources with the ArcBox solution tag
func hasArcBoxSolutionTag(resourceGroupName string) bool {
	cmd := exec.Command("timeout", "10s", "az", "resource", "list", "--resource-group", resourceGroupName,
		"--query", "[?tags.Solution=='jumpstart_arcbox'].id", "-o", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// hasArcBoxDeployments checks for deployments with the ArcBox naming pattern
func hasArcBoxDeployments(resourceGroupName string) bool {
	cmd := exec.Command("timeout", "10s", "az", "deployment", "group", "list", "--resource-group", resourceGroupName,
		"--query", "[?contains(name, 'arcbox') || contains(name, 'ArcBox') || contains(name, 'Arcbox') || contains(name, 'ARCBOX')].name", "-o", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// hasArcBoxNamingPattern checks for ArcBox naming patterns in resources
func hasArcBoxNamingPattern(resourceGroupName string) bool {
	cmd := exec.Command("timeout", "10s", "az", "resource", "list", "--resource-group", resourceGroupName,
		"--query", "[?starts_with(name,'ArcBox') || starts_with(name,'arcbox')].id", "-o", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// hasArcBoxResources checks for characteristic ArcBox resource types
func hasArcBoxResources(resourceGroupName string) bool {
	// Look for typical ArcBox resources: Key Vault + VM + specific extensions
	arcboxResourceTypes := []string{
		"Microsoft.KeyVault/vaults",
		"Microsoft.Compute/virtualMachines",
		"Microsoft.Network/virtualNetworks",
	}

	for _, resourceType := range arcboxResourceTypes {
		cmd := exec.Command("az", "resource", "list", "--resource-group", resourceGroupName,
			"--resource-type", resourceType, "--query", "[].id", "-o", "tsv")
		output, err := cmd.Output()
		if err != nil || strings.TrimSpace(string(output)) == "" {
			return false
		}
	}
	return true
}

// enrichArcBoxDeployment adds additional information to an ArcBox deployment
func enrichArcBoxDeployment(deployment *ArcBoxDeployment) {
	// Get resource count
	deployment.ResourceCount = getResourceCount(deployment.ResourceGroupName)

	// Get creation date from resource group
	deployment.CreatedDate = getResourceGroupCreationDate(deployment.ResourceGroupName)

	// Determine ArcBox flavor and naming prefix
	deployment.Flavor, deployment.NamingPrefix = detectArcBoxFlavor(deployment.ResourceGroupName)

	// Get deployment status
	deployment.Status = getDeploymentStatus(deployment.ResourceGroupName)
}

// getResourceCount returns the number of resources in a resource group
func getResourceCount(resourceGroupName string) int {
	cmd := exec.Command("az", "resource", "list", "--resource-group", resourceGroupName, "--query", "length(@)", "-o", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0
	}
	return count
}

// getResourceGroupCreationDate returns the creation date of a resource group
func getResourceGroupCreationDate(resourceGroupName string) string {
	// Try to get creation date from deployment history - use the EARLIEST deployment
	deployCmd := exec.Command("az", "deployment", "group", "list", "--resource-group", resourceGroupName,
		"--query", "min_by(@, &properties.timestamp).properties.timestamp", "-o", "tsv")
	deployOutput, err := deployCmd.Output()
	if err == nil && strings.TrimSpace(string(deployOutput)) != "" {
		if timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(string(deployOutput))); err == nil {
			// Convert UTC time to local time before formatting the date
			localTime := timestamp.Local()
			return localTime.Format("2006-01-02")
		}
	}

	// Fallback: try to get resource group creation time from Azure Resource Graph
	// This queries the resource group itself for its creation time
	rgCmd := exec.Command("az", "group", "show", "--name", resourceGroupName,
		"--query", "properties.provisioningState", "-o", "tsv")
	if rgOutput, rgErr := rgCmd.Output(); rgErr == nil && strings.TrimSpace(string(rgOutput)) == "Succeeded" {
		// If resource group exists but no deployments found, try alternative method
		// Get the earliest resource creation time as a proxy
		resourceCmd := exec.Command("az", "resource", "list", "--resource-group", resourceGroupName,
			"--query", "min_by(@, &properties.timeCreated).properties.timeCreated", "-o", "tsv")
		if resourceOutput, resourceErr := resourceCmd.Output(); resourceErr == nil && strings.TrimSpace(string(resourceOutput)) != "" {
			if timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(string(resourceOutput))); err == nil {
				// Convert UTC time to local time before formatting the date
				localTime := timestamp.Local()
				return localTime.Format("2006-01-02")
			}
		}
	}

	return "Unknown"
}

// detectArcBoxFlavor detects the ArcBox flavor and naming prefix from resources
// detectArcBoxFlavor determines the flavor by checking deployment parameters
func detectArcBoxFlavor(resourceGroupName string) (string, string) {
	// First, find the ArcBox deployment by looking for deployments containing "arcbox" (case-insensitive)
	cmd := exec.Command("az", "deployment", "group", "list", "--resource-group", resourceGroupName,
		"--query", "[?contains(name, 'arcbox') || contains(name, 'ArcBox') || contains(name, 'Arcbox') || contains(name, 'ARCBOX')].name", "-o", "tsv")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to old method if deployment not found
		return detectArcBoxFlavorFallback(resourceGroupName)
	}

	deploymentNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(deploymentNames) == 0 || deploymentNames[0] == "" {
		// Fallback to old method if no arcbox deployment found
		return detectArcBoxFlavorFallback(resourceGroupName)
	}

	// Use the first (or most recent) arcbox deployment
	deploymentName := strings.TrimSpace(deploymentNames[0])

	// Get the flavor parameter from the deployment
	flavorCmd := exec.Command("az", "deployment", "group", "show", "--resource-group", resourceGroupName,
		"--name", deploymentName, "--query", "properties.parameters.flavor.value", "-o", "tsv")
	flavorOutput, err := flavorCmd.Output()
	if err != nil {
		// Fallback to old method if parameter not found
		return detectArcBoxFlavorFallback(resourceGroupName)
	}

	flavor := strings.TrimSpace(string(flavorOutput))
	if flavor == "" {
		// Fallback to old method if flavor parameter is empty
		return detectArcBoxFlavorFallback(resourceGroupName)
	}

	// Return the detected flavor and default prefix
	return flavor, "ArcBox"
}

// detectArcBoxFlavorFallback provides fallback flavor detection using resource inspection
func detectArcBoxFlavorFallback(resourceGroupName string) (string, string) {
	// Check for SQL Server (indicates DataOps)
	sqlCmd := exec.Command("az", "resource", "list", "--resource-group", resourceGroupName,
		"--resource-type", "Microsoft.Sql/servers", "--query", "[].id", "-o", "tsv")
	if sqlOutput, err := sqlCmd.Output(); err == nil && strings.TrimSpace(string(sqlOutput)) != "" {
		return "DataOps", "ArcBox"
	}

	// Check for AKS cluster (indicates DevOps)
	aksCmd := exec.Command("az", "resource", "list", "--resource-group", resourceGroupName,
		"--resource-type", "Microsoft.ContainerService/managedClusters", "--query", "[].id", "-o", "tsv")
	if aksOutput, err := aksCmd.Output(); err == nil && strings.TrimSpace(string(aksOutput)) != "" {
		return "DevOps", "ArcBox"
	}

	// Check for naming prefix from VM names
	vmCmd := exec.Command("az", "vm", "list", "--resource-group", resourceGroupName, "--query", "[].name", "-o", "tsv")
	if vmOutput, err := vmCmd.Output(); err == nil {
		vmNames := strings.Split(strings.TrimSpace(string(vmOutput)), "\n")
		for _, vmName := range vmNames {
			vmName = strings.TrimSpace(vmName)
			if vmName != "" {
				// Extract prefix (everything before "Client" or "VM")
				if strings.Contains(vmName, "Client") {
					prefix := strings.Split(vmName, "Client")[0]
					if prefix != "" {
						return "ITPro", prefix
					}
				}
			}
		}
	}

	return "ITPro", "ArcBox"
}

// getDeploymentStatus returns the overall deployment status with improved logic
func getDeploymentStatus(resourceGroupName string) string {
	// Get all deployments in the resource group with their states
	cmd := exec.Command("az", "deployment", "group", "list", "--resource-group", resourceGroupName,
		"--query", "[].{name:name,state:properties.provisioningState,timestamp:properties.timestamp}", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown"
	}

	var deployments []map[string]interface{}
	if err := json.Unmarshal(output, &deployments); err != nil {
		return "Unknown"
	}

	if len(deployments) == 0 {
		return "Unknown"
	}

	// Check for any failed deployments first
	var hasArcBoxDeployment bool
	var arcBoxDeploymentStatus string
	var anyFailed bool

	for _, deployment := range deployments {
		name, _ := deployment["name"].(string)
		state, _ := deployment["state"].(string)

		// Check if this is an ArcBox deployment (contains "arcbox" case-insensitive)
		if strings.Contains(strings.ToLower(name), "arcbox") {
			hasArcBoxDeployment = true
			arcBoxDeploymentStatus = state
		}

		// If any deployment failed, mark overall status as failed
		if state == "Failed" {
			anyFailed = true
		}
	}

	// Prioritize ArcBox deployment status if found
	if hasArcBoxDeployment {
		// If the specific ArcBox deployment failed, return Failed
		if arcBoxDeploymentStatus == "Failed" {
			return "Failed"
		}
		// If any deployment failed but ArcBox succeeded, still return Failed for overall status
		if anyFailed {
			return "Failed"
		}
		// Return the ArcBox deployment status
		return arcBoxDeploymentStatus
	}

	// Fallback: if any deployment failed, return Failed
	if anyFailed {
		return "Failed"
	}

	// Get the most recent deployment status as fallback
	mostRecentCmd := exec.Command("az", "deployment", "group", "list", "--resource-group", resourceGroupName,
		"--query", "max_by(@, &properties.timestamp).properties.provisioningState", "-o", "tsv")
	recentOutput, err := mostRecentCmd.Output()
	if err != nil {
		return "Unknown"
	}

	status := strings.TrimSpace(string(recentOutput))
	if status == "" {
		return "Unknown"
	}

	return status
}

// outputArcBoxDeploymentsTable outputs deployments in table format
func outputArcBoxDeploymentsTable(deployments []ArcBoxDeployment) error {
	fmt.Printf("\n🎯 Found %d ArcBox deployment(s):\n\n", len(deployments))

	headers := []string{"Resource Group", "Subscription", "Location", "Flavor", "Status", "Resources", "Created"}
	rows := [][]string{}

	for _, deployment := range deployments {
		statusIcon := getStatusIcon(deployment.Status)
		statusDisplay := fmt.Sprintf("%s %s", statusIcon, deployment.Status)

		// Truncate long subscription names
		subName := deployment.SubscriptionName
		if len(subName) > 25 {
			subName = subName[:22] + "..."
		}

		rows = append(rows, []string{
			deployment.ResourceGroupName,
			subName,
			deployment.Location,
			deployment.Flavor,
			statusDisplay,
			fmt.Sprintf("%d", deployment.ResourceCount),
			deployment.CreatedDate,
		})
	}

	table.PrintASCIITable(headers, rows)
	fmt.Println()

	return nil
}

// outputArcBoxDeploymentsJSON outputs deployments in JSON format
func outputArcBoxDeploymentsJSON(deployments []ArcBoxDeployment) error {
	jsonData, err := json.MarshalIndent(deployments, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal deployments to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// getStatusIcon returns an appropriate icon for deployment status
func getStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "succeeded":
		return "✅"
	case "failed":
		return "❌"
	case "running", "creating", "accepted", "inprogress":
		return "⌛"
	case "canceled", "cancelled":
		return "🚫"
	default:
		return "❓"
	}
}

// getDeploymentProvisioningState returns the provisioning state of the deployment
func getDeploymentProvisioningState(resourceGroup, deploymentName string) string {
	cmd := exec.Command("az", "deployment", "group", "show", "--resource-group", resourceGroup, "--name", deploymentName, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown"
	}
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return "Unknown"
	}
	if props, ok := result["properties"].(map[string]interface{}); ok {
		if state, ok := props["provisioningState"].(string); ok {
			return state
		}
	}
	return "Unknown"
}

// printDeploymentErrorDetails prints error details for a failed deployment
func printDeploymentErrorDetails(resourceGroup, deploymentName string) {
	cmd := exec.Command("az", "deployment", "group", "show", "--resource-group", resourceGroup, "--name", deploymentName, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(utils.ErrorColor("[ERROR] Unable to retrieve deployment error details."))
		return
	}
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		fmt.Println(utils.ErrorColor("[ERROR] Unable to parse deployment error details."))
		return
	}
	if props, ok := result["properties"].(map[string]interface{}); ok {
		if errObj, ok := props["error"].(map[string]interface{}); ok {
			if msg, ok := errObj["message"].(string); ok {
				fmt.Println(utils.ErrorColor("[ERROR]"), msg)
			}
		}
	}
}

// getAzureDeploymentDuration gets the actual deployment duration from Azure timestamps
func getAzureDeploymentDuration(resourceGroup, deploymentName string) (time.Duration, error) {
	cmd := exec.Command("az", "deployment", "group", "show", "--resource-group", resourceGroup, "--name", deploymentName, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return 0, err
	}

	if props, ok := result["properties"].(map[string]interface{}); ok {
		// Get start and end timestamps
		var startTimeStr, endTimeStr string
		if ts, ok := props["timestamp"].(string); ok {
			startTimeStr = ts
		}
		if duration, ok := props["duration"].(string); ok {
			// Azure returns duration in ISO 8601 format like "PT1H30M45S"
			if duration != "" {
				parsed, err := parseISO8601Duration(duration)
				if err == nil {
					return parsed, nil
				}
			}
		}

		// Fallback: calculate from timestamps if duration field isn't available
		if startTimeStr != "" {
			if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
				// If we have an end time, use it; otherwise use current time
				endTime := time.Now()
				if endTimeStr != "" {
					if et, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
						endTime = et
					}
				}
				return endTime.Sub(startTime), nil
			}
		}
	}

	return 0, fmt.Errorf("unable to determine deployment duration from Azure")
}

// parseISO8601Duration parses Azure's ISO 8601 duration format (e.g., "PT1H30M45S")
func parseISO8601Duration(duration string) (time.Duration, error) {
	// Remove "PT" prefix
	if !strings.HasPrefix(duration, "PT") {
		return 0, fmt.Errorf("invalid ISO 8601 duration format")
	}
	duration = duration[2:]

	var totalDuration time.Duration

	// Parse hours
	if idx := strings.Index(duration, "H"); idx != -1 {
		hours, err := strconv.Atoi(duration[:idx])
		if err != nil {
			return 0, err
		}
		totalDuration += time.Duration(hours) * time.Hour
		duration = duration[idx+1:]
	}

	// Parse minutes
	if idx := strings.Index(duration, "M"); idx != -1 {
		minutes, err := strconv.Atoi(duration[:idx])
		if err != nil {
			return 0, err
		}
		totalDuration += time.Duration(minutes) * time.Minute
		duration = duration[idx+1:]
	}

	// Parse seconds
	if idx := strings.Index(duration, "S"); idx != -1 {
		seconds, err := strconv.ParseFloat(duration[:idx], 64)
		if err != nil {
			return 0, err
		}
		totalDuration += time.Duration(seconds * float64(time.Second))
	}

	return totalDuration, nil
}

// --- Utility functions ---

// normalizeFlavorCase normalizes flavor names to proper case
func normalizeFlavorCase(flavor string) string {
	switch strings.ToLower(flavor) {
	case "itpro", "it-pro":
		return "ITPro"
	case "devops", "dev-ops":
		return "DevOps"
	case "dataops", "data-ops":
		return "DataOps"
	case "all":
		return "all"
	default:
		return flavor
	}
}

// normalizeSqlServerEditionCase normalizes SQL Server edition names to proper case
func normalizeSqlServerEditionCase(edition string) string {
	switch strings.ToLower(edition) {
	case "developer":
		return "Developer"
	case "standard":
		return "Standard"
	case "enterprise":
		return "Enterprise"
	default:
		return edition
	}
}

// normalizeBastionSkuCase normalizes Bastion SKU names to proper case
func normalizeBastionSkuCase(sku string) string {
	switch strings.ToLower(sku) {
	case "basic":
		return "Basic"
	case "standard":
		return "Standard"
	case "developer":
		return "Developer"
	default:
		return sku
	}
}

// getSubscriptionID gets the subscription ID from command flags or default
func getSubscriptionID(cmd *cobra.Command) string {
	// Try to get from command flag first
	subscription, _ := cmd.Flags().GetString("subscription")
	if subscription != "" {
		return subscription
	}

	// Try to get from environment variable
	if env := os.Getenv("AZURE_SUBSCRIPTION_ID"); env != "" {
		return env
	}

	// Fallback: use current Azure CLI subscription
	out, err := exec.Command("az", "account", "show", "--query", "id", "-o", "tsv").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	return ""
}

// scanSubscriptionWithSpinner scans a subscription for ArcBox deployments with spinner
func scanSubscriptionWithSpinner(sub AzureSubscription) []ArcBoxDeployment {
	// Set up spinner for scanning
	stopSpinner := make(chan struct{})
	spinnerDone := make(chan struct{})

	// Animation frames: Unicode spinner for smooth animation
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIdx := 0

	// Show initial message
	fmt.Printf("🔍 Scanning subscription %s...", sub.Name)

	// Hide cursor before starting animation
	fmt.Print("\033[?25l")

	// Start spinner animation in goroutine
	go func() {
		for {
			select {
			case <-stopSpinner:
				// Clear the spinner line completely
				fmt.Printf("\r\033[2K")
				// Restore cursor when animation stops
				fmt.Print("\033[?25h")
				close(spinnerDone)
				return
			default:
				// Update spinner frame
				fmt.Printf("\r🔍 Scanning subscription %s... %s", sub.Name, frames[frameIdx])
				frameIdx = (frameIdx + 1) % len(frames)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Perform the actual scanning
	deployments, err := discoverArcBoxDeployments(sub.ID, sub.Name)

	// Stop spinner and wait for cleanup
	close(stopSpinner)
	<-spinnerDone

	if err != nil {
		fmt.Printf("🔍 Scanning subscription %s... %s\n", sub.Name, utils.ErrorColor("❌"))
		fmt.Printf("Error scanning subscription %s: %v\n", sub.Name, err)
		return []ArcBoxDeployment{}
	}

	// Show final result with success indicator
	fmt.Printf("🔍 Scanning subscription %s... %s\n", sub.Name, utils.SuccessColor("✅"))
	fmt.Printf("Found %d ArcBox deployment(s) in %s\n", len(deployments), sub.Name)
	return deployments
}

// setAzureSubscription sets the Azure CLI subscription context
func setAzureSubscription(subscriptionID string) error {
	if subscriptionID == "" {
		return fmt.Errorf("subscription ID is empty")
	}
	cmd := exec.Command("az", "account", "set", "--subscription", subscriptionID)
	return cmd.Run()
}

// checkResourceGroupExists checks if a resource group exists
func checkResourceGroupExists(resourceGroupName, subscriptionID string) (bool, error) {
	if subscriptionID != "" {
		if err := setAzureSubscription(subscriptionID); err != nil {
			return false, fmt.Errorf("failed to set subscription context: %v", err)
		}
	}

	cmd := exec.Command("az", "group", "exists", "--name", resourceGroupName)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check resource group existence: %v", err)
	}

	exists := strings.TrimSpace(string(output)) == "true"
	return exists, nil
}

// --- Location validation functions ---

// validateLocations checks if all provided locations are valid Azure regions supported by ArcBox
// Returns an error immediately if any location is invalid, providing informative error messages
func validateLocations(locations []string) error {
	if len(locations) == 0 {
		return fmt.Errorf("no locations provided")
	}

	// Load supported ArcBox regions and create lookup maps
	var supportedRegions []string
	if err := json.Unmarshal(regions.ArcboxSupportedRegionsData, &supportedRegions); err != nil {
		return fmt.Errorf("failed to load supported regions: %v", err)
	}

	// Create maps for both normalized names and display names
	normalizedToDisplay := make(map[string]string)
	displayToNormalized := make(map[string]string)

	for _, region := range supportedRegions {
		normalized := utils.NormalizeRegion(region)
		normalizedToDisplay[normalized] = region
		displayToNormalized[region] = normalized
	}

	var invalidLocations []string
	var suggestions []string

	for _, location := range locations {
		location = strings.TrimSpace(location)
		if location == "" {
			continue
		}

		normalized := utils.NormalizeRegion(location)

		// Check if location is valid (either as normalized name or display name)
		isValid := false

		// First check if it's a supported ArcBox region (normalized)
		if _, exists := normalizedToDisplay[normalized]; exists {
			isValid = true
		}

		// Then check if it's a display name that maps to a supported region
		if !isValid {
			if normalizedFromDisplay, exists := displayToNormalized[location]; exists {
				if _, supported := normalizedToDisplay[normalizedFromDisplay]; supported {
					isValid = true
				}
			}
		}

		if !isValid {
			invalidLocations = append(invalidLocations, location)

			// Try to suggest a similar region
			suggestion := utils.SuggestSimilarCommand(normalized, getSupportedRegionsList(normalizedToDisplay), 3)
			if suggestion != "" {
				if displayName, exists := normalizedToDisplay[suggestion]; exists {
					suggestions = append(suggestions, fmt.Sprintf("'%s' → '%s' (%s)", location, suggestion, displayName))
				}
			}
		}
	}

	if len(invalidLocations) > 0 {
		errorMsg := fmt.Sprintf("invalid location(s): %s", strings.Join(invalidLocations, ", "))

		if len(suggestions) > 0 {
			errorMsg += "\n\nDid you mean:\n  " + strings.Join(suggestions, "\n  ")
		}

		errorMsg += fmt.Sprintf("\n\nSupported ArcBox regions:\n  %s", strings.Join(getSupportedRegionsDisplayList(normalizedToDisplay), "\n  "))

		return fmt.Errorf("%s", errorMsg)
	}

	return nil
}

// getSupportedRegionsList returns a list of supported region names (normalized)
func getSupportedRegionsList(normalizedToDisplay map[string]string) []string {
	var regions []string
	for normalized := range normalizedToDisplay {
		regions = append(regions, normalized)
	}
	return regions
}

// getSupportedRegionsDisplayList returns a list of supported region display names
func getSupportedRegionsDisplayList(normalizedToDisplay map[string]string) []string {
	var regions []string
	for _, display := range normalizedToDisplay {
		regions = append(regions, display)
	}
	return regions
}

// runQuotaChecksWithTable performs detailed quota checking with table output
// Returns true if all quota checks pass, false otherwise
func runQuotaChecksWithTable(cmd *cobra.Command, location, flavor string) bool {
	// Normalize flavor and validate
	flavor = normalizeFlavorCase(flavor)

	subscription := getSubscriptionID(cmd)
	if subscription == "" {
		fmt.Println(utils.ErrorColor("❌ [ERROR] Unable to get subscription ID. Please ensure Azure CLI is authenticated."))
		return false
	}

	// Get SKUs for the flavor(s)
	var allSKUs []string
	if flavor == "all" {
		// Get SKUs for all flavors
		for _, f := range []string{"ITPro", "DevOps", "DataOps"} {
			skus := getFlavorSKUs(f)
			allSKUs = append(allSKUs, skus...)
		}
		// Remove duplicates
		uniqueSKUs := make(map[string]bool)
		for _, sku := range allSKUs {
			uniqueSKUs[sku] = true
		}
		allSKUs = []string{}
		for sku := range uniqueSKUs {
			allSKUs = append(allSKUs, sku)
		}
	} else {
		allSKUs = getFlavorSKUs(flavor)
	}

	if len(allSKUs) == 0 {
		fmt.Printf(utils.ErrorColor("❌ [ERROR] No SKUs found for flavor: %s\n"), flavor)
		return false
	}

	fmt.Printf(utils.InfoColor("🔍 Checking vCPU quota and SKU availability for %s flavor...\n"), flavor)

	// Step 1: Check SKU availability for all SKUs in the region (with spinner)
	startSKUCheck := time.Now()

	// Show spinner during SKU availability check
	stopSpinner := make(chan struct{})
	spinnerDone := make(chan struct{})

	// Animation frames: Unicode spinner for smooth animation
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIdx := 0

	// Show initial message
	fmt.Printf("🔍 Checking SKU availability in %s...", location)

	// Hide cursor before starting animation
	fmt.Print("\033[?25l")

	// Start spinner animation in goroutine
	go func() {
		for {
			select {
			case <-stopSpinner:
				// Clear the spinner line completely
				fmt.Printf("\r\033[2K")
				// Restore cursor when animation stops
				fmt.Print("\033[?25h")
				close(spinnerDone)
				return
			default:
				// Update spinner frame
				fmt.Printf("\r🔍 Checking SKU availability in %s... %s", location, frames[frameIdx])
				frameIdx = (frameIdx + 1) % len(frames)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Call the batch SKU availability check function
	unavailableSKUs := validator.CheckBatchSKUAvailability(allSKUs, location, subscription)

	// Stop spinner and wait for cleanup
	close(stopSpinner)
	<-spinnerDone

	// Create a map for quick lookup
	skuAvailabilityMap := make(map[string]bool)
	for _, sku := range allSKUs {
		skuAvailabilityMap[sku] = true // Assume available by default
	}
	for _, sku := range unavailableSKUs {
		skuAvailabilityMap[sku] = false // Mark unavailable SKUs
	}

	elapsedSKU := time.Since(startSKUCheck)
	if len(unavailableSKUs) == 0 {
		fmt.Printf("🔍 Checking SKU availability in %s... %s%s\n", location, utils.SuccessColor("✓"), utils.DebugColor(fmt.Sprintf(" (%.1fs)", elapsedSKU.Seconds())))
	} else {
		fmt.Printf("🔍 Checking SKU availability in %s... %s%s\n", location, utils.ErrorColor("✗"), utils.DebugColor(fmt.Sprintf(" (%.1fs) - %d unavailable SKUs", elapsedSKU.Seconds(), len(unavailableSKUs))))
	}

	// Prepare table data with original 8-column format
	headers := []string{"ArcBox Flavor", "Location", "SKU", "vCPU Quota (Available/Limit)", "Required vCPU", "Can Deploy ArcBox?", "Details"}
	var rows [][]string
	allPassed := true

	// Track whether this is the first quota call for this region (for timing indicators)
	isFirstCallForRegion := true

	// Step 2: Check quota for each SKU with progress messages and timing (with spinners)
	for _, sku := range allSKUs {
		required := getRequiredVCPUForSKU(sku)

		// Track timing for Azure CLI call (first call will be slow, subsequent calls should be fast due to caching)
		startTime := time.Now()

		// Set up spinner for quota check
		stopQuotaSpinner := make(chan struct{})
		quotaSpinnerDone := make(chan struct{})
		quotaFrameIdx := 0

		// Show initial message
		fmt.Printf("🔍 Checking quota for %s / %s in %s...", flavor, sku, location)

		// Hide cursor before starting animation
		fmt.Print("\033[?25l")

		// Start spinner animation in goroutine
		go func() {
			for {
				select {
				case <-stopQuotaSpinner:
					// Clear the spinner line completely
					fmt.Printf("\r\033[2K")
					// Restore cursor when animation stops
					fmt.Print("\033[?25h")
					close(quotaSpinnerDone)
					return
				default:
					// Update spinner frame
					fmt.Printf("\r🔍 Checking quota for %s / %s in %s... %s", flavor, sku, location, frames[quotaFrameIdx])
					quotaFrameIdx = (quotaFrameIdx + 1) % len(frames)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()

		// Use the real checkQuotaForSKU function from preflight/validator.go
		quotaOK, _, limit, available := validator.CheckQuotaForSKU(sku, required, location, subscription, flavor)

		// Stop spinner and wait for cleanup
		close(stopQuotaSpinner)
		<-quotaSpinnerDone

		// Check if SKU is available in this region
		skuAvailable := skuAvailabilityMap[sku]

		// Calculate elapsed time and show appropriate indicator
		elapsed := time.Since(startTime)
		overallOK := quotaOK && skuAvailable
		if overallOK {
			if isFirstCallForRegion && elapsed > 1*time.Second {
				// First call for this region took a while - this is expected for real Azure CLI
				fmt.Printf("🔍 Checking quota for %s / %s in %s... %s%s\n", flavor, sku, location, utils.SuccessColor("✓"), utils.DebugColor(fmt.Sprintf(" (%.1fs)", elapsed.Seconds())))
				isFirstCallForRegion = false // Mark that we've made the first call for this region
			} else if !isFirstCallForRegion && elapsed < 200*time.Millisecond {
				// Subsequent calls in same region should be very fast due to caching
				fmt.Printf("🔍 Checking quota for %s / %s in %s... %s%s\n", flavor, sku, location, utils.SuccessColor("✓"), utils.DebugColor(" (cached)"))
			} else {
				fmt.Printf("🔍 Checking quota for %s / %s in %s... %s\n", flavor, sku, location, utils.SuccessColor("✓"))
				isFirstCallForRegion = false // Mark first call complete even if timing was unexpected
			}
		} else {
			fmt.Printf("🔍 Checking quota for %s / %s in %s... %s\n", flavor, sku, location, utils.ErrorColor("✗"))
			isFirstCallForRegion = false // Mark first call complete even on failure
		}

		// Determine deployment status and details based on both quota and SKU availability
		var canDeploy string
		var details string
		if quotaOK && skuAvailable {
			canDeploy = utils.SuccessColor("Yes")
			details = "Sufficient quota and SKU available"
			allPassed = allPassed && true
		} else if !quotaOK && !skuAvailable {
			canDeploy = utils.ErrorColor("No")
			details = fmt.Sprintf("Need %d more vCPU; SKU not available", required-available)
			allPassed = false
		} else if !quotaOK {
			canDeploy = utils.ErrorColor("No")
			details = fmt.Sprintf("Need %d more vCPU", required-available)
			allPassed = false
		} else { // !skuAvailable
			canDeploy = utils.ErrorColor("No")
			details = "SKU not available in region"
			allPassed = false
		}

		// Format the quota column as "Available/Limit"
		quotaDisplay := fmt.Sprintf("%d/%d", available, limit)

		// Add row to table with original 8-column format
		row := []string{
			flavor,
			utils.GetRegionDisplayName(location),
			sku,
			quotaDisplay,
			fmt.Sprintf("%d", required),
			canDeploy,
			details,
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

	return allPassed
}

// getFlavorSKUs returns the VM SKUs required for a specific ArcBox flavor
func getFlavorSKUs(flavor string) []string {
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

// getRequiredVCPUForSKU returns the vCPU requirement for a VM SKU
func getRequiredVCPUForSKU(sku string) int {
	vcpuMap := map[string]int{
		"Standard_D8s_v5": 8,
		"Standard_D8s_v4": 8,
		"Standard_B2ms":   2,
		"Standard_B4ms":   4,
		"Standard_B8ms":   8,
	}
	if vcpu, ok := vcpuMap[sku]; ok {
		return vcpu
	}
	return 1 // Default fallback
}

// mapSKUToFamilyQuotaName maps a VM SKU to its Azure vCPU family quota name
func mapSKUToFamilyQuotaName(sku string) string {
	sku = strings.TrimPrefix(sku, "Standard_")
	parts := strings.Split(sku, "_")
	if len(parts) < 2 {
		return ""
	}

	main := parts[0]
	ver := parts[1]

	// Remove digits from main part, keep only letters
	letters := ""
	for _, r := range main {
		if r >= '0' && r <= '9' {
			continue
		}
		letters += string(r)
	}

	// If ends with 's', keep it (e.g. D8s → Ds)
	if strings.HasSuffix(main, "s") && !strings.HasSuffix(letters, "s") {
		letters += "s"
	}

	family := "Standard " + strings.ToUpper(letters[:1]) + letters[1:] + ver + " Family vCPUs"
	return family
}

// parseInt64 safely converts interface{} to int64
func parseInt64(val interface{}) int {
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return 0
}
