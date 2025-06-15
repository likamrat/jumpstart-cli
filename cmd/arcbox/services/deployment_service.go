package services

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"jumpstartcli/cmd/arcbox/display"
	arcboxUtils "jumpstartcli/cmd/arcbox/utils"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/urlutils"
	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// DeploymentService handles ArcBox deployment operations
type DeploymentService struct {
	azureCLI azurecli.AzureCLI
	display  *display.DeploymentDisplay
}

// NewDeploymentService creates a new deployment service with dependency injection
func NewDeploymentService(cli azurecli.AzureCLI, disp *display.DeploymentDisplay) *DeploymentService {
	return &DeploymentService{
		azureCLI: cli,
		display:  disp,
	}
}

// Deploy handles the core deployment logic for ArcBox
func (s *DeploymentService) Deploy(cmd *cobra.Command, args []string, templateSpecString string, noWait bool, deploymentSpecString string) error {
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
	flavor := arcboxUtils.NormalizeFlavorCase(flavorRaw)

	// Normalize SQL Server edition to proper case (case-insensitive input support)
	sqlServerEdition := arcboxUtils.NormalizeSqlServerEditionCase(sqlServerEditionRaw)

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
	bastionSku := arcboxUtils.NormalizeBastionSkuCase(bastionSkuRaw)

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
		utils.Error("Usage: js arcbox deploy --resource-group <n> --windows-user <username> --windows-password <password> [other arguments]")
		utils.ShowHelpWithoutTypes(cmd)
		return fmt.Errorf("missing required arguments: resource-group, windows-user, or windows-password")
	}

	if !utils.IsAzureLoggedInWithCLI(s.azureCLI) {
		utils.Error("You are not logged in to Azure. Please run 'az login' and try again.")
		utils.ShowHelpWithoutTypes(cmd)
		return fmt.Errorf("not logged in to Azure")
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
				return fmt.Errorf("deployment cancelled by user")
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
				return fmt.Errorf("deployment cancelled by user")
			}
		}
	}

	if !utils.ResourceGroupExistsWithCLI(s.azureCLI, resourceGroup) {
		utils.Info("Resource group '%s' does not exist. Creating it...", resourceGroup)
		err := utils.CreateResourceGroupWithCLI(s.azureCLI, resourceGroup, location)
		if err != nil {
			utils.Error("Failed to create resource group: %v", err)
			utils.ShowHelpWithoutTypes(cmd)
			return fmt.Errorf("failed to create resource group: %v", err)
		}
	}

	var azArgs []string
	var params []string // Declare params outside the conditionals

	if useParamFile && paramFile != "" {
		azArgs = []string{"deployment", "group", "create", "--resource-group", resourceGroup, "--template-file", bicepPath, "--parameters", "@" + paramFile}
	} else if len(bicepPath) > 4 && (bicepPath[:4] == "http") {
		params = []string{} // Initialize for HTTP case
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
		params = []string{} // Initialize for local file case
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

	// Handle deployment creation based on the deployment method
	if useParamFile && paramFile != "" {
		// For parameter file case, still use the CreateDeployment method
		// but params will be empty since they're in the file
		err := s.azureCLI.CreateDeployment(resourceGroup, deploymentName, bicepPath, []string{"@" + paramFile}, true)
		if err != nil {
			utils.Error("Error starting az deployment: %v", err)
			fmt.Println("Please check your parameters, resource group, and Azure login status.")
			portalUrl := fmt.Sprintf("https://portal.azure.com/#view/HubsExtension/BrowseResource/resourceType/Microsoft.Resources%%2Fdeployments/resourceGroup/%s", resourceGroup)
			fmt.Printf("View failed deployment details in the Azure Portal: %s\n", portalUrl)
			return fmt.Errorf("error starting deployment: %v", err)
		}
	} else {
		// For both HTTP and local file cases, use the constructed params
		err := s.azureCLI.CreateDeployment(resourceGroup, deploymentName, bicepPath, params, true) // noWait = true
		if err != nil {
			utils.Error("Error starting az deployment: %v", err)
			fmt.Println("Please check your parameters, resource group, and Azure login status.")
			portalUrl := fmt.Sprintf("https://portal.azure.com/#view/HubsExtension/BrowseResource/resourceType/Microsoft.Resources%%2Fdeployments/resourceGroup/%s", resourceGroup)
			fmt.Printf("View failed deployment details in the Azure Portal: %s\n", portalUrl)
			return fmt.Errorf("error starting deployment: %v", err)
		}
	}

	// Get subscription ID for portal link
	subscription := arcboxUtils.GetSubscriptionID(cmd, s.azureCLI)

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

	// Create Azure CLI instance for deployment monitoring
	// Skip monitoring when using mock CLI for testing
	if _, isMock := s.azureCLI.(*azurecli.MockAzureCLI); !isMock {
		s.display.WaitForDeploymentAndShowStatus(resourceGroup, deploymentName)
	}
	return nil
}
