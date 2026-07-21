package database

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

// NewCloneCommand clones a database to a destination.
func NewCloneCommand() *cobra.Command {
	var destinationUUID, name string
	var cloneVolumes bool
	cmd := &cobra.Command{
		Use:   "clone <uuid>",
		Args:  cli.ExactArgs(1, "<uuid>"),
		Short: "Clone a database to a destination",
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
			resp, err := service.NewDatabaseService(client).Clone(cmd.Context(), args[0], req)
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

// NewRunStorageBackupCommand runs a volume backup schedule immediately.
func NewRunStorageBackupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run-storage-backup <uuid> <storage_uuid>",
		Args:  cli.ExactArgs(2, "<uuid> <storage_uuid>"),
		Short: "Run a volume backup schedule now",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := service.NewDatabaseService(client).RunStorageBackup(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			return printAny(cmd, resp)
		},
	}
}
