package application

import "github.com/spf13/cobra"

// NewCommand creates the service application command tree.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Aliases: []string{"applications", "app", "apps"},
		Short:   "Manage applications within a service",
	}
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newUpdateCommand())
	cmd.AddCommand(newLogsCommand())
	cmd.AddCommand(newStartCommand())
	cmd.AddCommand(newRestartCommand())
	cmd.AddCommand(newStopCommand())
	return cmd
}
