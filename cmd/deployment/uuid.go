package deployment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// ResultDisplay represents a deploy result for table display
type ResultDisplay struct {
	Message        string `json:"message"`
	DeploymentUUID string `json:"deployment_uuid"`
}

// NewUUIDCommand deploys a resource by UUID
func NewUUIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uuid <uuid>",
		Short: "Deploy by uuid",
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			if err := validateDeployFlags(ctx, cmd, client); err != nil {
				return err
			}

			deploySvc := service.NewDeploymentService(client)
			result, err := deploySvc.Deploy(ctx, getDeployRequest(cmd, uuid))
			if err != nil {
				return fmt.Errorf("failed to deploy resource: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			// For table format, convert deployment info array to display format
			if format == output.FormatTable {
				displays := make([]ResultDisplay, len(result.Deployments))
				for i, dep := range result.Deployments {
					displays[i] = ResultDisplay{
						Message:        dep.Message,
						DeploymentUUID: dep.DeploymentUUID,
					}
				}
				return formatter.Format(displays)
			}

			return formatter.Format(result)
		},
	}

	addDeployFlags(cmd)
	return cmd
}
