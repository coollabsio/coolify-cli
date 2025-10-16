package deployment

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
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
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			force, _ := cmd.Flags().GetBool("force")
			deploySvc := service.NewDeploymentService(client)
			result, err := deploySvc.Deploy(ctx, uuid, force)
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

	cmd.Flags().Bool("force", false, "Force deployment")
	return cmd
}
