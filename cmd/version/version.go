package version

import (
	"fmt"

	"jumpstartcli/internal/utils"

	"github.com/spf13/cobra"
)

// NewVersionCmd creates and returns the version command
func NewVersionCmd() *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display the current version of the CLI",
		Long:  "Display the current version of the Jumpstart CLI",		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "Jumpstart CLI version: %s\n", utils.CliVersion)
		},
	}

	return versionCmd
}
