package database

import "github.com/spf13/cobra"

// NewCommand creates the service database command tree.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "database",
		Aliases: []string{"databases", "db", "dbs"},
		Short:   "Manage databases within a service",
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
