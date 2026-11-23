package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/config"
)

// NewConfigCommand creates the config command. Without a subcommand it shows
// the CLI configuration file location. Its subcommands manage coolify.json.
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show configuration file location and manage coolify.json",
		Long: `Without a subcommand, display the path to the Coolify CLI configuration file.

The subcommands manage coolify.json (config as code): a versioned file that
defines the resources of a project environment.

Examples:
  coolify config
  coolify config export <environment_uuid> -o coolify.json
  coolify config plan <environment_uuid> -f coolify.json
  coolify config link create --project my-app --environment production --mode api
  coolify config apply <link_uuid> -f coolify.json`,
		Args: cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(config.Path())
		},
	}

	cmd.AddCommand(newExportCommand())
	cmd.AddCommand(newPlanCommand())
	cmd.AddCommand(newLinkCommand())
	cmd.AddCommand(newRevisionsCommand())
	cmd.AddCommand(newApplyCommand())
	cmd.AddCommand(newSuspendCommand())
	cmd.AddCommand(newResumeCommand())
	cmd.AddCommand(newApproveCommand())
	cmd.AddCommand(newReleaseCommand())
	cmd.AddCommand(newRetryCommand())
	return cmd
}
