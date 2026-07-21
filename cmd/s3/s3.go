package s3

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewS3Command() *cobra.Command {
	cmd := &cobra.Command{Use: "s3", Aliases: []string{"s3-storage", "s3-storages"}, Short: "Manage S3-compatible storage destinations"}
	cmd.AddCommand(newList(), newGet(), newCreate(), newUpdate(), newDelete(), newValidate())
	return cmd
}

func svc(cmd *cobra.Command) (*service.S3StorageService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewS3StorageService(client), nil
}

func format(cmd *cobra.Command, value any) error {
	formatName, _ := cmd.Flags().GetString("format")
	showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
	if !showSensitive {
		switch typed := value.(type) {
		case *models.S3Storage:
			redacted := *typed
			if typed.Key != nil {
				o := output.SensitiveOverlay
				redacted.Key = &o
			}
			if typed.Secret != nil {
				o := output.SensitiveOverlay
				redacted.Secret = &o
			}
			value = &redacted
		case []models.S3Storage:
			redacted := append([]models.S3Storage(nil), typed...)
			for i := range redacted {
				if redacted[i].Key != nil {
					o := output.SensitiveOverlay
					redacted[i].Key = &o
				}
				if redacted[i].Secret != nil {
					o := output.SensitiveOverlay
					redacted[i].Secret = &o
				}
			}
			value = redacted
		}
	}
	formatter, err := output.NewFormatter(formatName, output.Options{ShowSensitive: showSensitive})
	if err != nil {
		return err
	}
	return formatter.Format(value)
}

func newList() *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List S3 storages", RunE: func(cmd *cobra.Command, _ []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		items, err := s.List(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list S3 storages: %w", err)
		}
		return format(cmd, items)
	}}
}

func newGet() *cobra.Command {
	return &cobra.Command{Use: "get <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Get an S3 storage", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		item, err := s.Get(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get S3 storage: %w", err)
		}
		return format(cmd, item)
	}}
}

func newCreate() *cobra.Command {
	var name, description, endpoint, bucket, region, key, secret string
	var isUsable bool
	cmd := &cobra.Command{Use: "create", Short: "Create an S3 storage", RunE: func(cmd *cobra.Command, _ []string) error {
		if name == "" || endpoint == "" || bucket == "" || region == "" || key == "" || secret == "" {
			return fmt.Errorf("--name, --endpoint, --bucket, --region, --key, and --secret are required")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		req := models.S3StorageCreateRequest{
			Name: name, Endpoint: endpoint, Bucket: bucket, Region: region, Key: key, Secret: secret,
		}
		if description != "" {
			req.Description = &description
		}
		if cmd.Flags().Changed("usable") {
			req.IsUsable = &isUsable
		}
		resp, err := s.Create(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create S3 storage: %w", err)
		}
		return format(cmd, resp)
	}}
	cmd.Flags().StringVar(&name, "name", "", "Storage name")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "S3 endpoint URL")
	cmd.Flags().StringVar(&bucket, "bucket", "", "Bucket name")
	cmd.Flags().StringVar(&region, "region", "", "Region")
	cmd.Flags().StringVar(&key, "key", "", "Access key")
	cmd.Flags().StringVar(&secret, "secret", "", "Secret key")
	cmd.Flags().BoolVar(&isUsable, "usable", true, "Mark storage as usable")
	return cmd
}

func newUpdate() *cobra.Command {
	var name, description, endpoint, bucket, region, key, secret string
	var isUsable bool
	cmd := &cobra.Command{Use: "update <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Update an S3 storage", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		req := models.S3StorageUpdateRequest{}
		if cmd.Flags().Changed("name") {
			req.Name = &name
		}
		if cmd.Flags().Changed("description") {
			req.Description = &description
		}
		if cmd.Flags().Changed("endpoint") {
			req.Endpoint = &endpoint
		}
		if cmd.Flags().Changed("bucket") {
			req.Bucket = &bucket
		}
		if cmd.Flags().Changed("region") {
			req.Region = &region
		}
		if cmd.Flags().Changed("key") {
			req.Key = &key
		}
		if cmd.Flags().Changed("secret") {
			req.Secret = &secret
		}
		if cmd.Flags().Changed("usable") {
			req.IsUsable = &isUsable
		}
		resp, err := s.Update(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update S3 storage: %w", err)
		}
		return format(cmd, resp)
	}}
	cmd.Flags().StringVar(&name, "name", "", "Storage name")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "S3 endpoint URL")
	cmd.Flags().StringVar(&bucket, "bucket", "", "Bucket name")
	cmd.Flags().StringVar(&region, "region", "", "Region")
	cmd.Flags().StringVar(&key, "key", "", "Access key")
	cmd.Flags().StringVar(&secret, "secret", "", "Secret key")
	cmd.Flags().BoolVar(&isUsable, "usable", true, "Mark storage as usable")
	return cmd
}

func newDelete() *cobra.Command {
	return &cobra.Command{Use: "delete <uuid>", Aliases: []string{"remove"}, Args: cli.ExactArgs(1, "<uuid>"), Short: "Delete an S3 storage", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		if err := s.Delete(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete S3 storage: %w", err)
		}
		fmt.Println("S3 storage deleted successfully.")
		return nil
	}}
}

func newValidate() *cobra.Command {
	return &cobra.Command{Use: "validate <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Validate S3 connection", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		resp, err := s.Validate(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to validate S3 storage: %w", err)
		}
		return format(cmd, resp)
	}}
}
