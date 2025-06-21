package completion

import (
	"os"

	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// NewCompletionCmd creates the completion command
func NewCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(js completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ js completion bash > /etc/bash_completion.d/js
  # macOS:
  $ js completion bash > /usr/local/etc/bash_completion.d/js

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ js completion zsh > "${fpath[1]}/_js"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ js completion fish | source

  # To load completions for each session, execute once:
  $ js completion fish > ~/.config/fish/completions/js.fish

PowerShell:

  PS> js completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> js completion powershell > js.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate arguments using centralized error handling
			if len(args) == 0 {
				utils.HandleMissingRequiredArguments(cmd, []string{"shell"})
				return nil
			}

			if len(args) > 1 {
				utils.PrintRequiredArgumentsError([]string{"shell"})
				return nil
			}

			// Validate shell type
			validShells := []string{"bash", "zsh", "fish", "powershell"}
			shell := args[0]
			isValidShell := false
			for _, validShell := range validShells {
				if shell == validShell {
					isValidShell = true
					break
				}
			}

			if !isValidShell {
				utils.PrintRequiredArgumentsError([]string{"shell (bash|zsh|fish|powershell)"})
				return nil
			}

			// Generate completion for the specified shell
			switch shell {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
}
