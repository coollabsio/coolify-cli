package env

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewGetEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <app_uuid> <env_uuid_or_key>",
		Short: "Get environment variable details",
		Long:  `Get detailed information about a specific environment variable by UUID or key name.`,
		Args:  cli.ExactArgs(2, "<app_uuid> <env_uuid_or_key>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			appUUID := args[0]
			envUUIDOrKey := args[1]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			appSvc := service.NewApplicationService(client)

			// First try to get by the identifier directly
			env, err := appSvc.GetEnv(ctx, appUUID, envUUIDOrKey)
			if err != nil {
				return fmt.Errorf("failed to get environment variable: %w", err)
			}

			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

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

	return cmd
}
