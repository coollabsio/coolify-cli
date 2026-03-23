package storage

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewCreateCommand returns the service storage create command
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <service_uuid>",
		Short: "Create a storage for a service",
		Long: `Create a persistent volume or file storage for a service.

The --resource-uuid flag is required to specify which service sub-resource
(application or database) the storage belongs to.

Examples:
  coolify svc storage create <service_uuid> --resource-uuid <sub_resource_uuid> --type persistent --name my-volume --mount-path /data
  coolify svc storage create <service_uuid> --resource-uuid <sub_resource_uuid> --type file --mount-path /app/config.yml --content "key: value"`,
		Args: cli.ExactArgs(1, "<service_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			storageType, _ := cmd.Flags().GetString("type")
			mountPath, _ := cmd.Flags().GetString("mount-path")
			resourceUUID, _ := cmd.Flags().GetString("resource-uuid")

			if storageType == "" {
				return fmt.Errorf("--type is required (persistent or file)")
			}
			if storageType != "persistent" && storageType != "file" {
				return fmt.Errorf("--type must be 'persistent' or 'file'")
			}
			if mountPath == "" {
				return fmt.Errorf("--mount-path is required")
			}
			if resourceUUID == "" {
				return fmt.Errorf("--resource-uuid is required (UUID of the service sub-resource)")
			}

			req := &models.ServiceStorageCreateRequest{
				Type:         storageType,
				MountPath:    mountPath,
				ResourceUUID: resourceUUID,
			}

			if cmd.Flags().Changed("name") {
				val, _ := cmd.Flags().GetString("name")
				req.Name = &val
			}
			if cmd.Flags().Changed("host-path") {
				val, _ := cmd.Flags().GetString("host-path")
				req.HostPath = &val
			}
			if cmd.Flags().Changed("content") {
				val, _ := cmd.Flags().GetString("content")
				req.Content = &val
			}
			if cmd.Flags().Changed("is-directory") {
				val, _ := cmd.Flags().GetBool("is-directory")
				req.IsDirectory = &val
			}
			if cmd.Flags().Changed("fs-path") {
				val, _ := cmd.Flags().GetString("fs-path")
				req.FsPath = &val
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := cli.CheckMinimumVersion(ctx, client, "4.0.0-beta.470"); err != nil {
				return err
			}

			svcSvc := service.NewService(client)
			if err := svcSvc.CreateStorage(ctx, args[0], req); err != nil {
				return fmt.Errorf("failed to create storage: %w", err)
			}

			fmt.Println("Storage created successfully.")
			return nil
		},
	}

	cmd.Flags().String("type", "", "Storage type: 'persistent' or 'file' (required)")
	cmd.Flags().String("mount-path", "", "Mount path inside the container (required)")
	cmd.Flags().String("resource-uuid", "", "UUID of the service sub-resource (required)")
	cmd.Flags().String("name", "", "Volume name (persistent only)")
	cmd.Flags().String("host-path", "", "Host path (persistent only)")
	cmd.Flags().String("content", "", "File content (file only)")
	cmd.Flags().Bool("is-directory", false, "Whether this is a directory mount (file only)")
	cmd.Flags().String("fs-path", "", "Host directory path (file only, required when --is-directory is set)")

	return cmd
}
