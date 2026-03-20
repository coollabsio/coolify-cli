package database

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/database/backup"
	"github.com/coollabsio/coolify-cli/cmd/database/env"
)

// NewDatabaseCommand creates the database parent command with all subcommands
func NewDatabaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "database",
		Aliases: []string{"databases", "db", "dbs"},
		Short:   "Manage Coolify databases",
		Long:    `Manage Coolify databases (PostgreSQL, MySQL, MongoDB, Redis, MariaDB, KeyDB, Clickhouse, Dragonfly).`,
	}

	// Add main database commands
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewStartCommand())
	cmd.AddCommand(NewStopCommand())
	cmd.AddCommand(NewRestartCommand())
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewUpdateCommand())
	cmd.AddCommand(NewDeleteCommand())

	// Add env subcommand
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage database environment variables",
	}
	envCmd.AddCommand(env.NewListCommand())
	envCmd.AddCommand(env.NewGetCommand())
	envCmd.AddCommand(env.NewCreateCommand())
	envCmd.AddCommand(env.NewUpdateCommand())
	envCmd.AddCommand(env.NewDeleteCommand())
	envCmd.AddCommand(env.NewSyncCommand())
	cmd.AddCommand(envCmd)

	// Add backup subcommand
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage database backups",
	}
	backupCmd.AddCommand(backup.NewCreateCommand())
	backupCmd.AddCommand(backup.NewListCommand())
	backupCmd.AddCommand(backup.NewDeleteCommand())
	backupCmd.AddCommand(backup.NewUpdateCommand())
	backupCmd.AddCommand(backup.NewTriggerCommand())
	backupCmd.AddCommand(backup.NewExecutionCommand())
	backupCmd.AddCommand(backup.NewDeleteExecutionCommand())
	cmd.AddCommand(backupCmd)

	return cmd
}
