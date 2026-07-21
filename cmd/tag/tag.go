package tag

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewTagCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "tag", Aliases: []string{"tags"}, Short: "Manage resource tags"}
	cmd.AddCommand(&cobra.Command{
		Use: "list", Short: "List all team tags",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			tags, err := service.NewTagService(client).List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list tags: %w", err)
			}
			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(tags)
		},
	})
	create := &cobra.Command{Use: "create", Short: "Create a team tag", RunE: func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		tag, err := service.NewTagService(client).Create(cmd.Context(), name)
		if err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}
		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return err
		}
		return formatter.Format(tag)
	}}
	create.Flags().String("name", "", "Tag name")
	cmd.AddCommand(create)

	update := &cobra.Command{Use: "update <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Rename a team tag", RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		tag, err := service.NewTagService(client).Update(cmd.Context(), args[0], name)
		if err != nil {
			return fmt.Errorf("failed to update tag: %w", err)
		}
		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return err
		}
		return formatter.Format(tag)
	}}
	update.Flags().String("name", "", "New tag name")
	cmd.AddCommand(update)

	cmd.AddCommand(&cobra.Command{Use: "delete <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Delete a team tag", RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		if err := service.NewTagService(client).Delete(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete tag: %w", err)
		}
		fmt.Println("Tag deleted.")
		return nil
	}})
	return cmd
}
