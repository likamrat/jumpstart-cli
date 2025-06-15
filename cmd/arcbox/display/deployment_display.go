// deployment_display.go - Display formatting for ArcBox deployment information
package display

import (
	"fmt"
	"strings"
	"time"

	"jumpstartcli/cmd/arcbox/models"
	arcboxUtils "jumpstartcli/cmd/arcbox/utils"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/utils"

	"github.com/fatih/color"
)

// DeploymentDisplay encapsulates dependencies for deployment display operations
type DeploymentDisplay struct {
	azureCLI azurecli.AzureCLI
}

// NewDeploymentDisplay creates a new DeploymentDisplay instance
func NewDeploymentDisplay(cli azurecli.AzureCLI) *DeploymentDisplay {
	return &DeploymentDisplay{azureCLI: cli}
}

// PrintResourceList displays the deployment resource list with status icons
func (dd *DeploymentDisplay) PrintResourceList(resources []models.ResourceStatus) {
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

// WaitForDeploymentAndShowStatus polls deployment status and prints resource-level operations
func (dd *DeploymentDisplay) WaitForDeploymentAndShowStatus(resourceGroup, deploymentName string) {
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
		resources             []models.ResourceStatus
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
			state = dd.getDeploymentProvisioningState(resourceGroup, deploymentName)
			resources = dd.getDeploymentResourceStatus(resourceGroup)
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
				dd.PrintResourceList(resources)
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
				dd.PrintErrorDetails(resourceGroup, deploymentName)
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
				dd.PrintResourceList(resources)

				// Get deployment duration from Azure as source of truth
				azureDuration, err := dd.getAzureDeploymentDuration(resourceGroup, deploymentName)
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
}

// PrintErrorDetails displays deployment error information
func (dd *DeploymentDisplay) PrintErrorDetails(resourceGroup, deploymentName string) {
	deployment, err := dd.azureCLI.GetDeployment(resourceGroup, deploymentName)
	if err != nil {
		fmt.Println(utils.ErrorColor("[ERROR] Unable to retrieve deployment error details."))
		return
	}

	if errorInterface, ok := deployment.Properties["error"]; ok {
		if errObj, ok := errorInterface.(map[string]interface{}); ok {
			if msg, ok := errObj["message"].(string); ok {
				fmt.Println(utils.ErrorColor("[ERROR]"), msg)
			}
		}
	}
}

// getDeploymentProvisioningState gets the deployment provisioning state
func (dd *DeploymentDisplay) getDeploymentProvisioningState(resourceGroup, deploymentName string) string {
	deployment, err := dd.azureCLI.GetDeployment(resourceGroup, deploymentName)
	if err != nil {
		return "Unknown"
	}
	return deployment.ProvisioningState
}

// getDeploymentResourceStatus gets the status of deployment resources
func (dd *DeploymentDisplay) getDeploymentResourceStatus(resourceGroup string) []models.ResourceStatus {
	resources, err := dd.azureCLI.ListResources(resourceGroup)
	if err != nil {
		return nil
	}

	var result []models.ResourceStatus
	for _, res := range resources {
		// Get detailed state for each resource
		detailedResource, err := dd.azureCLI.GetResource(res.ID)
		provState := "Unknown"
		if err == nil && detailedResource != nil {
			if props, ok := detailedResource.Properties["provisioningState"]; ok {
				if ps, ok := props.(string); ok {
					provState = ps
				}
			}
		}

		result = append(result, models.ResourceStatus{
			Name:  res.Name,
			Type:  res.Type,
			State: provState,
		})
	}
	return result
}

// getAzureDeploymentDuration gets the actual deployment duration from Azure timestamps
func (dd *DeploymentDisplay) getAzureDeploymentDuration(resourceGroup, deploymentName string) (time.Duration, error) {
	deployment, err := dd.azureCLI.GetDeployment(resourceGroup, deploymentName)
	if err != nil {
		return 0, err
	}

	if props, ok := deployment.Properties["timestamp"].(string); ok {
		// Get start and end timestamps
		var startTimeStr, endTimeStr string
		startTimeStr = props
		if duration, ok := deployment.Properties["duration"].(string); ok {
			// Azure returns duration in ISO 8601 format like "PT1H30M45S"
			if duration != "" {
				parsed, err := arcboxUtils.ParseISO8601Duration(duration)
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
