package teams

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewUpdateCommand creates the command that updates the token's team (PATCH /team).
func NewUpdateCommand() *cobra.Command {
	var buildServerFallback bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update settings of the team bound to the API token",
		Long: `Update settings of the team associated with the authentication token (API: PATCH /team).

--build-server-fallback controls whether deployments may fall back to the
deployment server when no usable dedicated build server is available.`,
		Example: `  coolify teams update --build-server-fallback=false
  coolify teams update --build-server-fallback`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !cmd.Flags().Changed("build-server-fallback") {
				return fmt.Errorf("--build-server-fallback is required (use --build-server-fallback=true or --build-server-fallback=false)")
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			team, err := service.NewTeamService(client).UpdateCurrent(cmd.Context(), models.TeamUpdateRequest{
				IsBuildServerFallbackEnabled: buildServerFallback,
			})
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{Writer: cmd.OutOrStdout()})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}
			return formatter.Format(team)
		},
	}
	cmd.Flags().BoolVar(&buildServerFallback, "build-server-fallback", false, "Allow deployments to fall back to the deployment server when no build server is usable (required)")
	return cmd
}
