package version

import (
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/version"
	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Current Coolify CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.CliVersion)
		},
	}
}
