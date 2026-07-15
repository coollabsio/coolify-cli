package tag

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "tag", Aliases: []string{"tags"}, Short: "Manage service tags"}
	cmd.AddCommand(newListCommand(), newAddCommand(), newRemoveCommand())
	return cmd
}

func newListCommand() *cobra.Command {
	return &cobra.Command{Use: "list <service-uuid>", Short: "List service tags", Args: cli.ExactArgs(1, "<service-uuid>"), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		tags, err := service.NewTagService(client).ListForResource(cmd.Context(), service.TagResourceServices, args[0])
		if err != nil {
			return fmt.Errorf("failed to list service tags: %w", err)
		}
		return format(cmd, tags)
	}}
}

func newAddCommand() *cobra.Command {
	return &cobra.Command{Use: "add <service-uuid> <name>", Short: "Add a service tag", Args: cli.ExactArgs(2, "<service-uuid> <name>"), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		tags, err := service.NewTagService(client).CreateForResource(cmd.Context(), service.TagResourceServices, args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to add service tag: %w", err)
		}
		return format(cmd, tags)
	}}
}

func newRemoveCommand() *cobra.Command {
	return &cobra.Command{Use: "remove <service-uuid> <tag-uuid>", Short: "Remove a service tag", Args: cli.ExactArgs(2, "<service-uuid> <tag-uuid>"), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		if err := service.NewTagService(client).DeleteForResource(cmd.Context(), service.TagResourceServices, args[0], args[1]); err != nil {
			return fmt.Errorf("failed to remove service tag: %w", err)
		}
		fmt.Println("Service tag removed.")
		return nil
	}}
}

func format(cmd *cobra.Command, value any) error {
	formatName, _ := cmd.Flags().GetString("format")
	formatter, err := output.NewFormatter(formatName, output.Options{})
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}
	return formatter.Format(value)
}
