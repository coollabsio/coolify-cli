package database

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/database/backup"
	"github.com/coollabsio/coolify-cli/cmd/database/env"
	"github.com/coollabsio/coolify-cli/cmd/database/storage"
	"github.com/coollabsio/coolify-cli/cmd/database/tag"
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
	cmd.AddCommand(NewLogsCommand())
	cmd.AddCommand(NewMoveCommand())
	cmd.AddCommand(NewCloneCommand())
	cmd.AddCommand(NewRunStorageBackupCommand())
	cmd.AddCommand(tag.NewCommand())

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

	// Add storage subcommand
	storageCmd := &cobra.Command{
		Use:     "storage",
		Aliases: []string{"storages"},
		Short:   "Manage database storages",
		Long:    `List and manage persistent volumes and file storages for databases.`,
	}
	storageCmd.AddCommand(storage.NewListCommand())
	storageCmd.AddCommand(storage.NewCreateCommand())
	storageCmd.AddCommand(storage.NewUpdateCommand())
	storageCmd.AddCommand(storage.NewDeleteCommand())
	storageCmd.AddCommand(storage.NewSetBackupScheduleCommand())
	storageCmd.AddCommand(storage.NewDeleteBackupScheduleCommand())
	storageCmd.AddCommand(storage.NewRunBackupCommand())
	cmd.AddCommand(storageCmd)

	return cmd
}
