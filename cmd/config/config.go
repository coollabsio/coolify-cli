package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/config"
)

// NewConfigCommand creates the config command
func NewConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show configuration file location",
		Long:  "Display the path to the Coolify CLI configuration file",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(config.Path())
		},
	}
}
