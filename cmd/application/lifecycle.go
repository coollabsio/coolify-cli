package application

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func printAny(cmd *cobra.Command, value any) error {
	formatName, _ := cmd.Flags().GetString("format")
	if formatName == "table" {
		formatName = "pretty"
	}
	formatter, err := output.NewFormatter(formatName, output.Options{})
	if err != nil {
		return err
	}
	return formatter.Format(value)
}

func NewCloneCommand() *cobra.Command {
	var destinationUUID, name string
	var cloneVolumes bool
	cmd := &cobra.Command{
		Use:   "clone <uuid>",
		Args:  cli.ExactArgs(1, "<uuid>"),
		Short: "Clone an application to a destination",
		RunE: func(cmd *cobra.Command, args []string) error {
			if destinationUUID == "" {
				return fmt.Errorf("--destination is required")
			}
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			req := models.ResourceCloneRequest{DestinationUUID: destinationUUID}
			if name != "" {
				req.Name = &name
			}
			if cmd.Flags().Changed("clone-volumes") {
				req.CloneVolumes = &cloneVolumes
			}
			resp, err := service.NewApplicationService(client).Clone(cmd.Context(), args[0], req)
			if err != nil {
				return err
			}
			return printAny(cmd, resp)
		},
	}
	cmd.Flags().StringVar(&destinationUUID, "destination", "", "Destination UUID")
	cmd.Flags().StringVar(&name, "name", "", "Optional new name")
	cmd.Flags().BoolVar(&cloneVolumes, "clone-volumes", false, "Clone volume data")
	return cmd
}

func NewRollbackCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "rollback", Short: "Application rollback helpers"}
	cmd.AddCommand(&cobra.Command{
		Use: "images <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "List rollback images",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := service.NewApplicationService(client).ListRollbackImages(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printAny(cmd, resp)
		},
	})
	run := &cobra.Command{
		Use: "run <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Rollback application to a commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			commit, _ := cmd.Flags().GetString("commit")
			if commit == "" {
				return fmt.Errorf("--commit is required")
			}
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := service.NewApplicationService(client).Rollback(cmd.Context(), args[0], commit)
			if err != nil {
				return err
			}
			return printAny(cmd, resp)
		},
	}
	run.Flags().String("commit", "", "Git commit SHA")
	cmd.AddCommand(run)
	return cmd
}

func NewDestinationsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "destinations", Aliases: []string{"destination"}, Short: "Manage application additional destinations"}
	cmd.AddCommand(&cobra.Command{Use: "list <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "List application destinations", RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return err
		}
		resp, err := service.NewApplicationService(client).ListDestinations(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printAny(cmd, resp)
	}})
	add := &cobra.Command{Use: "add <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Attach a destination", RunE: func(cmd *cobra.Command, args []string) error {
		dest, _ := cmd.Flags().GetString("destination")
		if dest == "" {
			return fmt.Errorf("--destination is required")
		}
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return err
		}
		resp, err := service.NewApplicationService(client).AddDestination(cmd.Context(), args[0], dest)
		if err != nil {
			return err
		}
		return printAny(cmd, resp)
	}}
	add.Flags().String("destination", "", "Destination UUID")
	cmd.AddCommand(add)
	cmd.AddCommand(&cobra.Command{Use: "remove <uuid> <destination_uuid>", Args: cli.ExactArgs(2, "<uuid> <destination_uuid>"), Short: "Detach a destination", RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return err
		}
		if err := service.NewApplicationService(client).RemoveDestination(cmd.Context(), args[0], args[1]); err != nil {
			return err
		}
		fmt.Println("Destination removed.")
		return nil
	}})
	return cmd
}

func NewExecuteTaskCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "execute-task <uuid> <task_uuid>",
		Args:  cli.ExactArgs(2, "<uuid> <task_uuid>"),
		Short: "Execute a scheduled task now",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := service.NewApplicationService(client).ExecuteScheduledTask(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			return printAny(cmd, resp)
		},
	}
}

func NewRunStorageBackupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run-storage-backup <uuid> <storage_uuid>",
		Args:  cli.ExactArgs(2, "<uuid> <storage_uuid>"),
		Short: "Run volume backup schedule now",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := service.NewApplicationService(client).RunStorageBackup(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			return printAny(cmd, resp)
		},
	}
}
