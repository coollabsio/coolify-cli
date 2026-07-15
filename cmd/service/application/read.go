package application

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use: "list <service-uuid>", Short: "List applications in a service", Args: cli.ExactArgs(1, "<service-uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			applications, err := service.NewService(client).ListApplications(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return format(cmd, applications)
		},
	}
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use: "get <service-uuid> <application-uuid>", Short: "Get an application in a service", Args: cli.ExactArgs(2, "<service-uuid> <application-uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			application, err := service.NewService(client).GetApplication(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			return format(cmd, application)
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
