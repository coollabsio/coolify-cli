package env

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

func NewListEnvCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <app_uuid>",
		Short: "List all environment variables for an application",
		Long:  `List all environment variables for a specific application.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			appSvc := service.NewApplicationService(client)
			envs, err := appSvc.ListEnvs(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to list environment variables: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			if !showSensitive {
				for i := range envs {
					envs[i].Value = "********"
					if envs[i].RealValue != nil {
						masked := "********"
						envs[i].RealValue = &masked
					}
				}
			}

			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			return formatter.Format(envs)
		},
	}
}
