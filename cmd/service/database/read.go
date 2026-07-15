package database

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use: "list <service-uuid>", Short: "List databases in a service", Args: cli.ExactArgs(1, "<service-uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			databases, err := service.NewService(client).ListDatabases(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return format(cmd, databases)
		},
	}
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use: "get <service-uuid> <database-uuid>", Short: "Get a database in a service", Args: cli.ExactArgs(2, "<service-uuid> <database-uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			database, err := service.NewService(client).GetDatabase(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			return format(cmd, database)
		},
	}
}

func format(cmd *cobra.Command, value any) error {
	formatName, _ := cmd.Flags().GetString("format")
	formatter, err := output.NewFormatter(formatName, output.Options{})
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}
	return formatter.Format(value)
}
