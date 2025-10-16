package env

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <service_uuid>",
		Short: "List all environment variables for a service",
		Long:  `List all environment variables for a specific service.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			serviceSvc := service.NewServiceService(client)
			envs, err := serviceSvc.ListEnvs(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to list environment variables: %w", err)
			}

			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			// Mask sensitive values unless --show-sensitive is used
			if !showSensitive {
				for i := range envs {
					envs[i].Value = "********"
					if envs[i].RealValue != nil {
						masked := "********"
						envs[i].RealValue = &masked
					}
				}
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(envs)
		},
	}
}
