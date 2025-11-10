package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/version"
)

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Current Coolify CLI version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.GetVersion())
		},
	}
}
