package arcbox

import (
	"fmt"
	"os"

	"jumpstartcli/cmd/arcbox/services"
	"jumpstartcli/internal/examples"
	"jumpstartcli/internal/preflight/arcbox"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// createDeployCommand creates the deploy command with the provided deployment service
func createDeployCommand(deployService *services.DeploymentService, cli interface{}) *cobra.Command {
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

			// Use the provided deployment service
			if err := deployService.Deploy(cmd, args, bicepPath, useParamFile, paramFile); err != nil {
				utils.Error("Deployment failed: %v", err)
				os.Exit(1)
			}
		},
	}

	// Add all the deploy command flags
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

	return arcboxDeployCmd
}
