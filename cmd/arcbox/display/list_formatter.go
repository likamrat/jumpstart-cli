// list_formatter.go - Formatting utilities for ArcBox deployment lists
package display

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"jumpstartcli/cmd/arcbox/models"
	"jumpstartcli/internal/azurecli"
	"jumpstartcli/internal/table"
	"jumpstartcli/internal/utils"
)

// This file contains functions for formatting and displaying lists of ArcBox
// deployments, including table formatting, JSON output, and filtering options.

// ListFormatter handles formatting and display of ArcBox deployment lists
type ListFormatter struct{}

// NewListFormatter creates a new ListFormatter instance
func NewListFormatter() *ListFormatter {
	return &ListFormatter{}
}

// OutputArcBoxDeploymentsTable outputs deployments in table format
func (lf *ListFormatter) OutputArcBoxDeploymentsTable(deployments []models.ArcBoxDeployment) error {
	fmt.Printf("\n🎯 Found %d ArcBox deployment(s):\n\n", len(deployments))

	headers := []string{"Resource Group", "Subscription", "Location", "Flavor", "Status", "Resources", "Created"}
	rows := [][]string{}

	for _, deployment := range deployments {
		statusIcon := lf.getStatusIcon(deployment.Status)
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

// OutputArcBoxDeploymentsJSON outputs deployments in JSON format
func (lf *ListFormatter) OutputArcBoxDeploymentsJSON(deployments []models.ArcBoxDeployment) error {
	jsonData, err := json.MarshalIndent(deployments, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal deployments to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// getStatusIcon returns an appropriate icon for deployment status
func (lf *ListFormatter) getStatusIcon(status string) string {
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

// ScanSubscriptionWithSpinner scans a subscription for ArcBox deployments with spinner
func (lf *ListFormatter) ScanSubscriptionWithSpinner(azCLI azurecli.AzureCLI, sub models.AzureSubscription, discoverFunc func(azurecli.AzureCLI, string, string) ([]models.ArcBoxDeployment, error)) []models.ArcBoxDeployment {
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
	deployments, err := discoverFunc(azCLI, sub.ID, sub.Name)

	// Stop spinner and wait for cleanup
	close(stopSpinner)
	<-spinnerDone

	if err != nil {
		fmt.Printf("🔍 Scanning subscription %s... %s\n", sub.Name, utils.ErrorColor("❌"))
		fmt.Printf("Error scanning subscription %s: %v\n", sub.Name, err)
		return []models.ArcBoxDeployment{}
	}

	// Show final result with success indicator
	fmt.Printf("🔍 Scanning subscription %s... %s\n", sub.Name, utils.SuccessColor("✅"))
	fmt.Printf("Found %d ArcBox deployment(s) in %s\n", len(deployments), sub.Name)
	return deployments
}
