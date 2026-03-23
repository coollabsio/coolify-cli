package storage

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewUpdateCommand returns the service storage update command
func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <service_uuid>",
		Short: "Update a storage for a service",
		Long: `Update a persistent volume or file storage for a service.

The --uuid and --type flags are required. Use 'coolify svc storage list' to find storage UUIDs.

Examples:
  coolify svc storage update <service_uuid> --uuid <storage_uuid> --type persistent --name my-volume
  coolify svc storage update <service_uuid> --uuid <storage_uuid> --type persistent --is-preview-suffix-enabled`,
		Args: cli.ExactArgs(1, "<service_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			storageUUID, _ := cmd.Flags().GetString("uuid")
			storageID, _ := cmd.Flags().GetInt("id")
			storageType, _ := cmd.Flags().GetString("type")

			if storageUUID == "" && storageID == 0 {
				return fmt.Errorf("--uuid is required (or --id as deprecated fallback)")
			}
			if storageType == "" {
				return fmt.Errorf("--type is required (persistent or file)")
			}
			if storageType != "persistent" && storageType != "file" {
				return fmt.Errorf("--type must be 'persistent' or 'file'")
			}

			req := &models.StorageUpdateRequest{
				Type: storageType,
			}

			if storageUUID != "" {
				req.UUID = &storageUUID
			} else {
				req.ID = &storageID
			}

			hasUpdates := false

			if cmd.Flags().Changed("is-preview-suffix-enabled") {
				val, _ := cmd.Flags().GetBool("is-preview-suffix-enabled")
				req.IsPreviewSuffixEnabled = &val
				hasUpdates = true
			}
			if cmd.Flags().Changed("name") {
				val, _ := cmd.Flags().GetString("name")
				req.Name = &val
				hasUpdates = true
			}
			if cmd.Flags().Changed("mount-path") {
				val, _ := cmd.Flags().GetString("mount-path")
				req.MountPath = &val
				hasUpdates = true
			}
			if cmd.Flags().Changed("host-path") {
				val, _ := cmd.Flags().GetString("host-path")
				req.HostPath = &val
				hasUpdates = true
			}
			if cmd.Flags().Changed("content") {
				val, _ := cmd.Flags().GetString("content")
				req.Content = &val
				hasUpdates = true
			}

			if !hasUpdates {
				return fmt.Errorf("no fields to update. Use --help to see available flags")
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := cli.CheckMinimumVersion(ctx, client, "4.0.0-beta.470"); err != nil {
				return err
			}

			svcSvc := service.NewService(client)
			if err := svcSvc.UpdateStorage(ctx, args[0], req); err != nil {
				return fmt.Errorf("failed to update storage: %w", err)
			}

			fmt.Println("Storage updated successfully.")
			return nil
		},
	}

	cmd.Flags().String("uuid", "", "Storage UUID (required, use 'storage list' to find)")
	cmd.Flags().Int("id", 0, "Storage ID (deprecated, use --uuid instead)")
	cmd.Flags().String("type", "", "Storage type: 'persistent' or 'file' (required)")
	cmd.Flags().Bool("is-preview-suffix-enabled", false, "Enable preview suffix for this storage")
	cmd.Flags().String("name", "", "Storage name (persistent only)")
	cmd.Flags().String("mount-path", "", "Mount path inside the container")
	cmd.Flags().String("host-path", "", "Host path (persistent only)")
	cmd.Flags().String("content", "", "File content (file only)")

	return cmd
}
