package env

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <service_uuid> <env_uuid_or_key>",
		Short: "Get environment variable details",
		Long:  `Get detailed information about a specific environment variable. First UUID is the service, second is the environment variable UUID or key name.`,
		Args:  cli.ExactArgs(2, "<uuid1> <uuid2>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			serviceUUID := args[0]
			envUUID := args[1]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			serviceSvc := service.NewService(client)
			env, err := serviceSvc.GetEnv(ctx, serviceUUID, envUUID)
			if err != nil {
				return fmt.Errorf("failed to get environment variable: %w", err)
			}

			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			// Mask sensitive value unless --show-sensitive is used
			if !showSensitive {
				env.Value = "********"
				if env.RealValue != nil {
					masked := "********"
					env.RealValue = &masked
				}
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(env)
		},
	}
}
