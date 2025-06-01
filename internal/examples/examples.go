// examples.go - Centralized examples management for Jumpstart CLI
package examples

import (
	"fmt"
	"strings"
)

// Example represents a single command example
type Example struct {
	Description string
	Command     string
}

// ExampleSet holds multiple examples for a command
type ExampleSet struct {
	Examples []Example
}

// FormatExamples returns a formatted string of examples for help output
func (es *ExampleSet) FormatExamples() string {
	if len(es.Examples) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("Examples\n")

	for i, example := range es.Examples {
		// Format similar to Azure CLI: description followed by indented command
		builder.WriteString(fmt.Sprintf("    %s\n", example.Description))
		builder.WriteString(fmt.Sprintf("        %s\n", example.Command))

		// Add blank line between examples for better readability (except after the last example)
		if i < len(es.Examples)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// GetExamples returns examples for a specific command path
func GetExamples(commandPath string) *ExampleSet {
	examples, exists := examplesRegistry[commandPath]
	if !exists {
		return &ExampleSet{}
	}
	return examples
}

// Registry of all examples organized by command path
var examplesRegistry = map[string]*ExampleSet{
	"js.arcbox.deploy": {
		Examples: []Example{
			{
				Description: "Deploy ITPro ArcBox in East US 2 with default settings",
				Command:     "js arcbox deploy -g MyResourceGroup -l eastus2 --windows-user jumpstart --flavor ITPro",
			},
			{
				Description: "Deploy DevOps ArcBox with custom SSH key and GitHub user",
				Command:     "js arcbox deploy -g DevOpsRG -l westus2 --windows-user jumpstart --flavor DevOps --ssh-rsa-public-key \"ssh-rsa AAAA...\" --github-user myusername",
			},
			{
				Description: "Deploy DataOps ArcBox with custom resource tags and bastion",
				Command:     "js arcbox deploy -g DataOpsRG -l centralus --windows-user jumpstart --flavor DataOps --deploy-bastion=yes --resource-tags '{\"Environment\":\"Production\",\"Team\":\"DataTeam\"}'",
			},
			{
				Description: "Deploy ArcBox using a local custom template file",
				Command:     "js arcbox deploy -g CustomRG -l westeurope --windows-user jumpstart --flavor ITPro --template-local my-custom-template.bicep",
			},
			{
				Description: "Deploy ArcBox using local template with custom parameters file",
				Command:     "js arcbox deploy -g CustomRG -l westeurope --windows-user jumpstart --flavor ITPro --template-local custom-arcbox.bicep --template-params custom-params.json",
			},
			{
				Description: "Deploy ArcBox using a remote ARM template from GitHub",
				Command:     "js arcbox deploy -g RemoteRG -l eastus --windows-user jumpstart --flavor ITPro --template-uri https://raw.githubusercontent.com/myorg/arcbox-templates/main/custom-arcbox.json",
			},
			{
				Description: "Deploy ArcBox with custom naming prefix and automatic shutdown disabled",
				Command:     "js arcbox deploy -g ProdRG -l eastus --windows-user jumpstart --flavor ITPro --naming-prefix MyArc --auto-shutdown=no",
			},
			{
				Description: "Deploy ArcBox with custom RDP port and bastion SKU",
				Command:     "js arcbox deploy -g SecureRG -l westus2 --windows-user jumpstart --flavor ITPro --rdp-port 3390 --deploy-bastion=yes --bastion-sku Standard",
			},
			{
				Description: "Deploy ArcBox with VM autologon enabled, bypassing security prompts",
				Command:     "js arcbox deploy -g AutoRG -l eastus --windows-user jumpstart --flavor ITPro --vm-autologon=yes --yes",
			},
			{
				Description: "Deploy ArcBox with auto-shutdown disabled, bypassing confirmation prompts",
				Command:     "js arcbox deploy -g NoshutdownRG -l westus --windows-user jumpstart --flavor ITPro --auto-shutdown=no -y",
			},
			{
				Description: "Deploy ArcBox with custom auto-shutdown time and timezone",
				Command:     "js arcbox deploy -g TimedRG -l eastus2 --windows-user jumpstart --flavor ITPro --auto-shutdown-time 2200 --auto-shutdown-timezone \"America/New_York\"",
			},
			{
				Description: "Deploy ArcBox with auto-shutdown email notifications",
				Command:     "js arcbox deploy -g NotifyRG -l centralus --windows-user jumpstart --flavor ITPro --auto-shutdown-email admin@company.com --auto-shutdown-time 1900",
			},
			{
				Description: "Deploy ArcBox with spot pricing enabled to reduce costs",
				Command:     "js arcbox deploy -g SpotRG -l westus2 --windows-user jumpstart --flavor ITPro --enable-spot-pricing=yes",
			},
		},
	},
	"js.arcbox.delete": {
		Examples: []Example{
			{
				Description: "Delete an ArcBox deployment by resource group name",
				Command:     "js arcbox delete --name MyArcBoxRG",
			},
			{
				Description: "Delete an ArcBox deployment without confirmation prompt",
				Command:     "js arcbox delete --name MyArcBoxRG --yes",
			},
			{
				Description: "Delete an ArcBox deployment in a specific subscription",
				Command:     "js arcbox delete -n MyArcBoxRG -s 12345678-1234-1234-1234-123456789012",
			},
		},
	},
	"arcbox.delete": {
		Examples: []Example{
			{
				Description: "Delete an ArcBox deployment by resource group name",
				Command:     "js arcbox delete --name MyArcBoxRG",
			},
			{
				Description: "Delete an ArcBox deployment without confirmation prompt",
				Command:     "js arcbox delete --name MyArcBoxRG --yes",
			},
			{
				Description: "Delete an ArcBox deployment in a specific subscription",
				Command:     "js arcbox delete -n MyArcBoxRG -s 12345678-1234-1234-1234-123456789012",
			},
		},
	},
	"js.arcbox.preflight.quota": {
		Examples: []Example{
			{
				Description: "Check quota for ITPro flavor in East US 2",
				Command:     "js arcbox preflight quota -l eastus2 --flavor ITPro",
			},
			{
				Description: "Check quota for all ArcBox flavors in West US 2",
				Command:     "js arcbox preflight quota -l westus2 --flavor all",
			},
			{
				Description: "Check quota for DevOps flavor in multiple locations",
				Command:     "js arcbox preflight quota -l eastus2,westus2,centralus --flavor DevOps",
			},
			{
				Description: "Check quota for ITPro flavor in all ArcBox-supported locations",
				Command:     "js arcbox preflight quota --all-locations --flavor ITPro",
			},
			{
				Description: "Check quota for all flavors across all supported locations",
				Command:     "js arcbox preflight quota --all-locations --flavor all",
			},
			{
				Description: "Check quota for custom VM SKUs in East US 2",
				Command:     "js arcbox preflight quota -l eastus2 --sku Standard_D8s_v5,Standard_B2ms",
			},
			{
				Description: "Check quota with specific subscription ID across multiple locations",
				Command:     "js arcbox preflight quota -l eastus2,westeurope --flavor DevOps -s 12345678-1234-1234-1234-123456789012",
			},
		},
	},
	"arcbox.preflight.rp.register": {
		Examples: []Example{
			{
				Description: "Register the Kubernetes resource provider.",
				Command:     "js arcbox preflight rp register --name Microsoft.Kubernetes",
			},
			{
				Description: "Register the Azure Arc Data resource provider.",
				Command:     "js arcbox preflight rp register -n Microsoft.AzureArcData",
			},
		},
	},
	"arcbox.preflight.rp.show": {
		Examples: []Example{
			{
				Description: "Check registration status of all required resource providers.",
				Command:     "js arcbox preflight rp show",
			},
		},
	},
	"arcbox.preflight.rp.list": {
		Examples: []Example{
			{
				Description: "List all required resource provider names.",
				Command:     "js arcbox preflight rp list",
			},
		},
	},
	"js.arcbox.preflight.rp": {
		Examples: []Example{
			{
				Description: "Check registration status of all required resource providers.",
				Command:     "js arcbox preflight rp show",
			},
			{
				Description: "List all required resource provider names.",
				Command:     "js arcbox preflight rp list",
			},
			{
				Description: "Register a specific resource provider.",
				Command:     "js arcbox preflight rp register --name Microsoft.Kubernetes",
			},
		},
	},
	"js.arcbox.preflight.rp.list": {
		Examples: []Example{
			{
				Description: "List all required resource provider names.",
				Command:     "js arcbox preflight rp list",
			},
		},
	},
	"js.arcbox.preflight.rp.register": {
		Examples: []Example{
			{
				Description: "Register the Kubernetes resource provider.",
				Command:     "js arcbox preflight rp register --name Microsoft.Kubernetes",
			},
			{
				Description: "Register the Azure Arc Data resource provider.",
				Command:     "js arcbox preflight rp register -n Microsoft.AzureArcData",
			},
		},
	},
	"js.arcbox.preflight.rp.show": {
		Examples: []Example{
			{
				Description: "Check registration status of all required resource providers.",
				Command:     "js arcbox preflight rp show",
			},
		},
	},
	"js.arcbox.list": {
		Examples: []Example{
			{
				Description: "List ArcBox deployments in the current subscription (explicit)",
				Command:     "js arcbox list --current-subscription",
			},
			{
				Description: "List ArcBox deployments across all subscriptions",
				Command:     "js arcbox list --all-subscriptions",
			},
			{
				Description: "List ArcBox deployments in a specific subscription",
				Command:     "js arcbox list --subscription 12345678-1234-1234-1234-123456789012",
			},
			{
				Description: "List ArcBox deployments with JSON output format",
				Command:     "js arcbox list --output json",
			},
			{
				Description: "List ArcBox deployments across all subscriptions in JSON format",
				Command:     "js arcbox list --all-subscriptions --output json",
			},
		},
	},

	"js.subscription.set": {
		Examples: []Example{
			{
				Description: "Set current subscription using subscription ID",
				Command:     "js subscription set --subscription 12345678-1234-1234-1234-123456789012",
			},
			{
				Description: "Set current subscription using subscription ID (short form)",
				Command:     "js subscription set -s 12345678-1234-1234-1234-123456789012",
			},
			{
				Description: "Set current subscription using subscription name",
				Command:     "js subscription set --name \"My Production Subscription\"",
			},
			{
				Description: "Set current subscription using subscription name (short form)",
				Command:     "js subscription set -n \"My Development Subscription\"",
			},
			{
				Description: "Set current subscription using positional argument (backward compatibility)",
				Command:     "js subscription set \"My Subscription Name\"",
			},
			{
				Description: "Set current subscription using positional argument with ID",
				Command:     "js subscription set 12345678-1234-1234-1234-123456789012",
			},
		},
	},

	"js.subscription.show": {
		Examples: []Example{
			{
				Description: "Show current subscription in default table format",
				Command:     "js subscription show",
			},
			{
				Description: "Show current subscription with explicit table format",
				Command:     "js subscription show --output table",
			},
			{
				Description: "Show current subscription in JSON format",
				Command:     "js subscription show --output json",
			},
			{
				Description: "Show current subscription in JSON format (short form)",
				Command:     "js subscription show -o json",
			},
			{
				Description: "Show detailed subscription information with verbose output",
				Command:     "js subscription show --verbose",
			},
			{
				Description: "Show only the subscription ID",
				Command:     "js subscription show --id",
			},
			{
				Description: "Show only the subscription name",
				Command:     "js subscription show --name",
			},
			{
				Description: "Show subscription in YAML format",
				Command:     "js subscription show --output yaml",
			},
			{
				Description: "Show subscription in TSV format for scripting",
				Command:     "js subscription show --output tsv",
			},
			{
				Description: "Show detailed subscription in verbose TSV format",
				Command:     "js subscription show --output tsv --verbose",
			},
		},
	},

	"js.completion": {
		Examples: []Example{
			{
				Description: "Generate bash completion script",
				Command:     "js completion bash",
			},
			{
				Description: "Generate zsh completion script",
				Command:     "js completion zsh",
			},
			{
				Description: "Generate fish completion script",
				Command:     "js completion fish",
			},
			{
				Description: "Generate PowerShell completion script",
				Command:     "js completion powershell",
			},
			{
				Description: "Load bash completions for current session",
				Command:     "source <(js completion bash)",
			},
			{
				Description: "Load zsh completions for current session",
				Command:     "source <(js completion zsh)",
			},
			{
				Description: "Load fish completions for current session",
				Command:     "js completion fish | source",
			},
			{
				Description: "Load PowerShell completions for current session",
				Command:     "js completion powershell | Out-String | Invoke-Expression",
			},
			{
				Description: "Save bash completions to file for persistent use",
				Command:     "js completion bash > ~/.bash_completion.d/js",
			},
			{
				Description: "Add zsh completions to shell configuration",
				Command:     "echo 'source <(js completion zsh)' >> ~/.zshrc",
			},
		},
	},
}
