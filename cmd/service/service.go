package service

import (
	"github.com/spf13/cobra"

	serviceapplication "github.com/coollabsio/coolify-cli/cmd/service/application"
	servicedatabase "github.com/coollabsio/coolify-cli/cmd/service/database"
	"github.com/coollabsio/coolify-cli/cmd/service/env"
	"github.com/coollabsio/coolify-cli/cmd/service/storage"
	"github.com/coollabsio/coolify-cli/cmd/service/tag"
)

// NewServiceCommand creates the service parent command with all subcommands
func NewServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"services", "svc"},
		Short:   "Service related commands",
		Long:    `Manage Coolify one-click services (databases, Redis, PostgreSQL, etc.).`,
	}

	// Add main service commands
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewStartCommand())
	cmd.AddCommand(NewStopCommand())
	cmd.AddCommand(NewRestartCommand())
	cmd.AddCommand(NewDeleteCommand())
	cmd.AddCommand(NewLogsCommand())
	cmd.AddCommand(NewMoveCommand())
	cmd.AddCommand(NewCloneCommand())
	cmd.AddCommand(NewExecuteTaskCommand())
	cmd.AddCommand(NewRunStorageBackupCommand())
	cmd.AddCommand(tag.NewCommand())
	cmd.AddCommand(serviceapplication.NewCommand())
	cmd.AddCommand(servicedatabase.NewCommand())

	// Add env subcommand
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage service environment variables",
	}
	envCmd.AddCommand(env.NewListCommand())
	envCmd.AddCommand(env.NewGetCommand())
	envCmd.AddCommand(env.NewCreateCommand())
	envCmd.AddCommand(env.NewUpdateCommand())
	envCmd.AddCommand(env.NewDeleteCommand())
	envCmd.AddCommand(env.NewSyncCommand())
	cmd.AddCommand(envCmd)

	// Add storage subcommand
	storageCmd := &cobra.Command{
		Use:     "storage",
		Aliases: []string{"storages"},
		Short:   "Manage service storages",
		Long:    `List and manage persistent volumes and file storages for services.`,
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
