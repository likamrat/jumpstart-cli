package services

import (
	"fmt"
	"strings"
	"time"

	"jumpstartcli/cmd/arcbox/display"
	"jumpstartcli/cmd/arcbox/models"
	arcboxUtils "jumpstartcli/cmd/arcbox/utils"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"
)

// ListingService handles ArcBox deployment discovery and listing operations
type ListingService struct {
	cli       azurecli.AzureCLI
	formatter *display.ListFormatter
}

// NewListingService creates a new listing service with dependency injection
func NewListingService(cli azurecli.AzureCLI) *ListingService {
	return &ListingService{
		cli:       cli,
		formatter: display.NewListFormatter(),
	}
}

// ListDeployments discovers and lists ArcBox deployments based on specified criteria
func (s *ListingService) ListDeployments(allSubscriptions, currentSubscription bool, subscriptionID, outputFormat string) error {
	var subscriptions []models.AzureSubscription
	var err error

	if allSubscriptions {
		fmt.Println(utils.InfoColor("[INFO] Searching for ArcBox deployments across all subscriptions..."))
		subscriptions, err = s.getAllSubscriptions()
	} else if subscriptionID != "" {
		fmt.Printf(utils.InfoColor("[INFO] Searching for ArcBox deployments in subscription %s...\n"), subscriptionID)
		sub, err := s.GetSubscription(subscriptionID)
		if err != nil {
			return err
		}
		subscriptions = []models.AzureSubscription{sub}
	} else if currentSubscription {
		// Explicit --current-subscription flag
		fmt.Println(utils.InfoColor("[INFO] Searching for ArcBox deployments in current subscription..."))
		sub, err := s.getCurrentSubscription()
		if err != nil {
			return err
		}
		subscriptions = []models.AzureSubscription{sub}
	}

	if err != nil {
		return err
	}

	var allDeployments []models.ArcBoxDeployment
	for _, sub := range subscriptions {
		// Scan each subscription with spinner animation
		deployments := s.formatter.ScanSubscriptionWithSpinner(s.cli, sub, s.discoverArcBoxDeployments)
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
		return s.formatter.OutputArcBoxDeploymentsJSON(allDeployments)
	case "table":
		fallthrough
	default:
		return s.formatter.OutputArcBoxDeploymentsTable(allDeployments)
	}
}

// getAllSubscriptions returns all accessible Azure subscriptions using Azure CLI wrapper
func (s *ListingService) getAllSubscriptions() ([]models.AzureSubscription, error) {
	subs, err := s.cli.ListSubscriptions()
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %v", err)
	}

	// Convert Azure CLI subscriptions to our model
	var subscriptions []models.AzureSubscription
	for _, sub := range subs {
		subscriptions = append(subscriptions, models.AzureSubscription{
			ID:   sub.ID,
			Name: sub.Name,
		})
	}
	return subscriptions, nil
}

// getCurrentSubscription returns the current Azure subscription using Azure CLI wrapper
func (s *ListingService) getCurrentSubscription() (models.AzureSubscription, error) {
	sub, err := s.cli.GetCurrentSubscription()
	if err != nil {
		return models.AzureSubscription{}, fmt.Errorf("failed to get current subscription: %v", err)
	}

	return models.AzureSubscription{ID: sub.ID, Name: sub.Name}, nil
}

// GetSubscription returns the specified Azure subscription using Azure CLI wrapper
func (s *ListingService) GetSubscription(subscriptionID string) (models.AzureSubscription, error) {
	sub, err := s.cli.GetSubscription(subscriptionID)
	if err != nil {
		return models.AzureSubscription{}, fmt.Errorf("failed to get subscription %s: %v", subscriptionID, err)
	}

	return models.AzureSubscription{ID: sub.ID, Name: sub.Name}, nil
}

// discoverArcBoxDeployments discovers ArcBox deployments in a specific subscription
func (s *ListingService) discoverArcBoxDeployments(azCLI azurecli.AzureCLI, subscriptionID, subscriptionName string) ([]models.ArcBoxDeployment, error) {
	// Set subscription context
	if err := arcboxUtils.SetAzureSubscription(azCLI, subscriptionID); err != nil {
		return nil, fmt.Errorf("failed to set subscription context: %v", err)
	}

	// Get all resource groups using Azure CLI wrapper
	resourceGroups, err := azCLI.ListResourceGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to list resource groups: %v", err)
	}

	// Pre-filter resource groups to exclude obvious non-ArcBox ones
	var candidateRGs []azurecli.ResourceGroupInfo
	for _, rg := range resourceGroups {
		rgNameLower := strings.ToLower(rg.Name)

		// Skip obviously non-ArcBox resource groups
		if strings.Contains(rgNameLower, "localbox") ||
			strings.HasPrefix(rg.Name, "MC_") ||
			strings.HasPrefix(rg.Name, "DefaultResourceGroup-") ||
			strings.HasPrefix(rg.Name, "NetworkWatcherRG") ||
			strings.HasPrefix(rg.Name, "AzSecPackAutoConfigRG") {
			continue
		}

		candidateRGs = append(candidateRGs, rg)
	}

	var deployments []models.ArcBoxDeployment
	for _, rg := range candidateRGs {
		rgName := rg.Name
		location := rg.Location

		// Check if this resource group contains an ArcBox deployment
		if s.isArcBoxResourceGroup(rgName) {
			deployment := models.ArcBoxDeployment{
				ResourceGroupName: rgName,
				SubscriptionID:    subscriptionID,
				SubscriptionName:  subscriptionName,
				Location:          location,
			}

			// Enrich deployment information
			s.enrichArcBoxDeployment(&deployment)
			deployments = append(deployments, deployment)
		}
	}

	return deployments, nil
}

// isArcBoxResourceGroup determines if a resource group contains an ArcBox deployment
func (s *ListingService) isArcBoxResourceGroup(resourceGroupName string) bool {
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
	if s.hasArcBoxSolutionTag(resourceGroupName) {
		return true
	}

	// Method 2: Check for ArcBox deployments by name pattern (faster)
	if s.hasArcBoxDeployments(resourceGroupName) {
		return true
	}

	// Method 3: Check for ArcBox naming patterns in resources (slower)
	if s.hasArcBoxNamingPattern(resourceGroupName) {
		return true
	}

	// Do NOT use hasArcBoxResources as it's too broad and causes false positives
	// with LocalBox and other deployments that have VMs, VNets, and Key Vaults

	return false
}

// hasArcBoxSolutionTag checks if any resource has the ArcBox solution tag
func (s *ListingService) hasArcBoxSolutionTag(resourceGroupName string) bool {
	resources, err := s.cli.ListResources(resourceGroupName)
	if err != nil {
		return false
	}

	// Check if any resource has the ArcBox solution tag
	for _, resource := range resources {
		if tags, ok := resource.Tags["Solution"]; ok && tags == "jumpstart_arcbox" {
			return true
		}
	}
	return false
}

// hasArcBoxDeployments checks for deployments with the ArcBox naming pattern using Azure CLI wrapper
func (s *ListingService) hasArcBoxDeployments(resourceGroupName string) bool {
	deployments, err := s.cli.ListDeployments(resourceGroupName)
	if err != nil {
		return false
	}

	// Check if any deployment has ArcBox in its name (case-insensitive)
	for _, deployment := range deployments {
		if strings.Contains(strings.ToLower(deployment.Name), "arcbox") {
			return true
		}
	}
	return false
}

// hasArcBoxNamingPattern checks for ArcBox naming patterns in resources using Azure CLI wrapper
func (s *ListingService) hasArcBoxNamingPattern(resourceGroupName string) bool {
	resources, err := s.cli.ListResources(resourceGroupName)
	if err != nil {
		return false
	}

	// Check if any resource has ArcBox naming pattern
	for _, resource := range resources {
		resourceName := strings.ToLower(resource.Name)
		if strings.HasPrefix(resourceName, "arcbox") {
			return true
		}
	}
	return false
}

// hasArcBoxResources checks for characteristic ArcBox resource types using Azure CLI wrapper
func (s *ListingService) hasArcBoxResources(resourceGroupName string) bool {
	// Look for typical ArcBox resources: Key Vault + VM + specific extensions
	arcboxResourceTypes := []string{
		"Microsoft.KeyVault/vaults",
		"Microsoft.Compute/virtualMachines",
		"Microsoft.Network/virtualNetworks",
	}

	// Get all resources in the resource group
	resources, err := s.cli.ListResources(resourceGroupName)
	if err != nil {
		return false
	}

	// Check if each required resource type exists
	for _, requiredType := range arcboxResourceTypes {
		found := false
		for _, resource := range resources {
			if resource.Type == requiredType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// enrichArcBoxDeployment enriches deployment information with additional details
func (s *ListingService) enrichArcBoxDeployment(deployment *models.ArcBoxDeployment) {
	// Get resource count
	deployment.ResourceCount = s.getResourceCount(deployment.ResourceGroupName)

	// Get creation date from resource group
	deployment.CreatedDate = s.getResourceGroupCreationDate(deployment.ResourceGroupName)

	// Determine ArcBox flavor and naming prefix
	deployment.Flavor, deployment.NamingPrefix = s.DetectArcBoxFlavor(deployment.ResourceGroupName)

	// Get deployment status
	deployment.Status = s.getDeploymentStatus(deployment.ResourceGroupName)
}

// getResourceCount returns the number of resources in a resource group using Azure CLI wrapper
func (s *ListingService) getResourceCount(resourceGroupName string) int {
	resources, err := s.cli.ListResources(resourceGroupName)
	if err != nil {
		return 0
	}
	return len(resources)
}

// getResourceGroupCreationDate returns the creation date of a resource group using Azure CLI wrapper
func (s *ListingService) getResourceGroupCreationDate(resourceGroupName string) string {
	// Try to get creation date from deployment history - use the EARLIEST deployment
	deployments, err := s.cli.ListDeployments(resourceGroupName)
	if err == nil && len(deployments) > 0 {
		// Find the earliest deployment by timestamp
		var earliestTime time.Time
		for i, deployment := range deployments {
			// Extract timestamp from properties
			if properties, ok := deployment.Properties["timestamp"]; ok {
				if timestampStr, ok := properties.(string); ok {
					if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
						if i == 0 || timestamp.Before(earliestTime) {
							earliestTime = timestamp
						}
					}
				}
			}
		}
		if !earliestTime.IsZero() {
			// Convert UTC time to local time before formatting the date
			localTime := earliestTime.Local()
			return localTime.Format("2006-01-02")
		}
	}

	// Fallback: use today's date if we can't determine the actual creation date
	return time.Now().Format("2006-01-02")
}

// DetectArcBoxFlavor detects the ArcBox flavor and naming prefix from resources using Azure CLI wrapper
func (s *ListingService) DetectArcBoxFlavor(resourceGroupName string) (string, string) {
	// First, find the ArcBox deployment by looking for deployments containing "arcbox" (case-insensitive)
	deployments, err := s.cli.ListDeployments(resourceGroupName)
	if err != nil {
		// Fallback to old method if deployment not found
		return s.DetectArcBoxFlavorFallback(resourceGroupName)
	}

	for _, deployment := range deployments {
		deploymentName := strings.ToLower(deployment.Name)
		if strings.Contains(deploymentName, "arcbox") {
			// Extract parameters from the deployment outputs to determine flavor
			if outputs, ok := deployment.Properties["outputs"]; ok {
				if outputsMap, ok := outputs.(map[string]interface{}); ok {
					// Look for flavor in deployment outputs
					if flavorOutput, ok := outputsMap["flavor"]; ok {
						if flavorMap, ok := flavorOutput.(map[string]interface{}); ok {
							if flavorValue, ok := flavorMap["value"].(string); ok {
								// Also try to extract naming prefix from outputs
								namingPrefix := "ArcBox" // default
								if namingOutput, ok := outputsMap["namingPrefix"]; ok {
									if namingMap, ok := namingOutput.(map[string]interface{}); ok {
										if namingValue, ok := namingMap["value"].(string); ok {
											namingPrefix = namingValue
										}
									}
								}
								return flavorValue, namingPrefix
							}
						}
					}
				}
			}
		}
	}

	// Fallback to the old detection method
	return s.DetectArcBoxFlavorFallback(resourceGroupName)
}

// DetectArcBoxFlavorFallback provides fallback flavor detection when deployment outputs are not available
func (s *ListingService) DetectArcBoxFlavorFallback(resourceGroupName string) (string, string) {
	resources, err := s.cli.ListResources(resourceGroupName)
	if err != nil {
		return "Unknown", "ArcBox"
	}

	// Check for flavor-specific resources
	hasDataControllersExtension := false
	hasSQLMIResources := false
	hasArcDataServices := false
	hasLinuxVM := false

	for _, resource := range resources {
		resourceType := resource.Type
		resourceName := strings.ToLower(resource.Name)

		// Check for DataOps-specific resources
		if strings.Contains(resourceName, "datacontroller") ||
			strings.Contains(resourceName, "sqlmi") ||
			resourceType == "Microsoft.AzureArcData/dataControllers" ||
			resourceType == "Microsoft.AzureArcData/sqlManagedInstances" {
			hasDataControllersExtension = true
			hasSQLMIResources = true
			hasArcDataServices = true
		}

		// Check for Linux VMs (DevOps and DataOps use Linux VMs)
		if resourceType == "Microsoft.Compute/virtualMachines" &&
			(strings.Contains(resourceName, "ubuntu") || strings.Contains(resourceName, "linux")) {
			hasLinuxVM = true
		}
	}

	// Determine flavor based on resources found
	namingPrefix := "ArcBox" // Default naming prefix

	if hasArcDataServices || hasDataControllersExtension || hasSQLMIResources {
		return "DataOps", namingPrefix
	}

	if hasLinuxVM {
		return "DevOps", namingPrefix
	}

	// Default to ITPro if no specific indicators are found
	return "ITPro", namingPrefix
}

// getDeploymentStatus returns the status of the most recent ArcBox deployment
func (s *ListingService) getDeploymentStatus(resourceGroupName string) string {
	deployments, err := s.cli.ListDeployments(resourceGroupName)
	if err != nil {
		return "Unknown"
	}

	// Find the most recent ArcBox deployment
	var mostRecentDeployment *azurecli.DeploymentInfo
	var mostRecentTime time.Time

	for _, deployment := range deployments {
		if strings.Contains(strings.ToLower(deployment.Name), "arcbox") {
			// Extract timestamp from properties
			if properties, ok := deployment.Properties["timestamp"]; ok {
				if timestampStr, ok := properties.(string); ok {
					if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
						if mostRecentDeployment == nil || timestamp.After(mostRecentTime) {
							deploymentCopy := deployment
							mostRecentDeployment = &deploymentCopy
							mostRecentTime = timestamp
						}
					}
				}
			}
		}
	}

	if mostRecentDeployment != nil {
		// Get the provisioning state from properties
		if properties, ok := mostRecentDeployment.Properties["provisioningState"]; ok {
			if status, ok := properties.(string); ok {
				return status
			}
		}
	}

	return "Unknown"
}
