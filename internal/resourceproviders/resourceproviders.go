// resourceproviders.go - Shared Azure resource provider management utilities
// This package provides reusable resource provider management functionality
// for all Jumpstart solutions (ArcBox, LocalBox, Agora, etc.)
package resourceproviders

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	InfoColor  = color.New(color.FgCyan).SprintFunc()
	ErrorColor = color.New(color.FgRed).SprintFunc()
)

// ResourceProviderConfig defines the configuration for a solution's resource providers
type ResourceProviderConfig struct {
	SolutionName      string   // Name of the solution (e.g., "ArcBox", "LocalBox", "Agora")
	RequiredProviders []string // List of required resource provider namespaces
}

// GetArcBoxProviders returns the resource provider configuration for ArcBox
func GetArcBoxProviders() ResourceProviderConfig {
	return ResourceProviderConfig{
		SolutionName: "ArcBox",
		RequiredProviders: []string{
			"Microsoft.Compute",
			"Microsoft.Kubernetes",
			"Microsoft.KubernetesConfiguration",
			"Microsoft.ExtendedLocation",
			"Microsoft.AzureArcData",
			"Microsoft.OperationsManagement",
			"Microsoft.HybridConnectivity",
		},
	}
}

// GetLocalBoxProviders returns the resource provider configuration for LocalBox
// TODO: Update with actual LocalBox requirements when implemented
func GetLocalBoxProviders() ResourceProviderConfig {
	return ResourceProviderConfig{
		SolutionName: "LocalBox",
		RequiredProviders: []string{
			"Microsoft.Compute",
			"Microsoft.Network",
			"Microsoft.Storage",
			// Add LocalBox-specific providers as needed
		},
	}
}

// GetAgoraProviders returns the resource provider configuration for Agora
// TODO: Update with actual Agora requirements when implemented
func GetAgoraProviders() ResourceProviderConfig {
	return ResourceProviderConfig{
		SolutionName: "Agora",
		RequiredProviders: []string{
			"Microsoft.Compute",
			"Microsoft.Network",
			"Microsoft.Storage",
			// Add Agora-specific providers as needed
		},
	}
}

// CheckProviderRegistration checks if a resource provider is registered
func CheckProviderRegistration(provider string) (bool, error) {
	// Add timeout context to prevent hanging on Azure CLI calls
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "provider", "show", "--namespace", provider, "-o", "tsv", "--query", "registrationState")
	out, err := cmd.Output()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return false, fmt.Errorf("timeout checking provider %s (Azure CLI took too long)", provider)
		}
		return false, fmt.Errorf("failed to check provider %s: %v", provider, err)
	}

	state := strings.TrimSpace(string(out))
	return state == "Registered", nil
}

// RegisterProvider registers a resource provider
func RegisterProvider(provider string) error {
	fmt.Printf(InfoColor("[INFO] Registering provider: %s\n"), provider)

	// Add timeout context to prevent hanging on Azure CLI calls
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmdOut := exec.CommandContext(ctx, "az", "provider", "register", "--namespace", provider)
	cmdOut.Stdout = os.Stdout
	cmdOut.Stderr = os.Stderr
	err := cmdOut.Run()
	if err != nil {
		fmt.Printf(ErrorColor("[ERROR] Failed to register provider %s: %v\n"), provider, err)
		return err
	}
	fmt.Println(InfoColor(fmt.Sprintf("[INFO] Provider %s registration initiated.", provider)))
	return nil
}

// CheckAllProviders checks the registration status of all providers for a solution
func CheckAllProviders(config ResourceProviderConfig) (bool, []string) {
	fmt.Printf(InfoColor("[INFO] Checking required Azure resource providers for %s...\n"), config.SolutionName)
	missing := false
	missingProviders := []string{}

	for _, rp := range config.RequiredProviders {
		isRegistered, err := CheckProviderRegistration(rp)
		if err != nil || !isRegistered {
			fmt.Printf("❌ %s: Not registered\n", rp)
			missing = true
			missingProviders = append(missingProviders, rp)
		} else {
			fmt.Printf("✅ %s: Registered\n", rp)
		}
	}

	if missing {
		fmt.Println(ErrorColor("[ERROR] One or more required resource providers are not registered."))
		fmt.Println("Use the register command to register missing providers.")
	} else {
		fmt.Println(InfoColor("[INFO] All required resource providers are registered."))
	}

	return !missing, missingProviders
}

// ListProviders lists all required providers for a solution
func ListProviders(config ResourceProviderConfig) {
	fmt.Printf(InfoColor("[INFO] Required Azure resource providers for %s:\n"), config.SolutionName)
	for _, rp := range config.RequiredProviders {
		fmt.Println(" -", rp)
	}
}

// GetRequiredProvidersForSolution returns a list of required providers for a given solution
// This is a convenience function for accessing provider lists without needing the full config
func GetRequiredProvidersForSolution(solutionName string) []string {
	switch strings.ToLower(solutionName) {
	case "arcbox":
		return GetArcBoxProviders().RequiredProviders
	case "localbox":
		return GetLocalBoxProviders().RequiredProviders
	case "agora":
		return GetAgoraProviders().RequiredProviders
	default:
		return nil
	}
}

// RegisterAllProviders registers all required providers for a solution
// Returns the number of providers that failed to register
func RegisterAllProviders(config ResourceProviderConfig) int {
	fmt.Printf(InfoColor("[INFO] Registering all required Azure resource providers for %s...\n"), config.SolutionName)
	failures := 0

	for _, provider := range config.RequiredProviders {
		if err := RegisterProvider(provider); err != nil {
			failures++
		}
	}

	if failures == 0 {
		fmt.Printf(InfoColor("[INFO] Successfully initiated registration for all %d providers.\n"), len(config.RequiredProviders))
	} else {
		fmt.Printf(ErrorColor("[ERROR] Failed to register %d out of %d providers.\n"), failures, len(config.RequiredProviders))
	}

	return failures
}
