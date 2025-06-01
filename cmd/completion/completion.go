package completion

import (
	"fmt"
	"os"

	"jumpstartcli/internal/examples"

	"github.com/spf13/cobra"
)

// NewCompletionCmd creates and returns the completion command
func NewCompletionCmd() *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for Jumpstart CLI commands and arguments.

To load completions:

Bash:
  $ source <(js completion bash)
  # To load completions for each session, add to your ~/.bashrc:
  #   source <(js completion bash)

Zsh:
  $ source <(js completion zsh)
  # To load completions for each session, add to your ~/.zshrc:
  #   source <(js completion zsh)

Fish:
  $ js completion fish | source
  # To load completions for each session, add to your ~/.config/fish/config.fish:
  #   js completion fish | source

PowerShell:
  PS> js completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, add to your profile:
  #   js completion powershell | Out-String | Invoke-Expression

` + examples.GetExamples("js.completion").FormatExamples(),
		Args: cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			// If no arguments provided, show help
			if len(args) == 0 {
				cmd.Help()
				return
			}

			// If too many arguments provided
			if len(args) > 1 {
				fmt.Printf("[ERROR] accepts at most 1 arg(s), received %d\n", len(args))
				return
			}

			// Validate the single argument
			validArgs := []string{"bash", "zsh", "fish", "powershell"}
			validArg := false
			for _, validArgument := range validArgs {
				if args[0] == validArgument {
					validArg = true
					break
				}
			}

			if !validArg {
				fmt.Printf("[ERROR] invalid argument %q for %q\n", args[0], cmd.CommandPath())
				return
			}

			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}

	return completionCmd
}
