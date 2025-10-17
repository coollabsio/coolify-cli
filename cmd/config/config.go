package config

import (
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/config"
	"github.com/spf13/cobra"
)

// NewConfigCommand creates the config command
func NewConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show configuration file location",
		Long:  "Display the path to the Coolify CLI configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.Path())
		},
	}
}
