package deployment

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewNameCommand deploys a resource by name
func NewNameCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "name <resource_name>",
		Short: "Deploy by resource name",
		Args:  cli.ExactArgs(1, "<resource_name>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			name := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Find resource by name
			resourceSvc := service.NewResourceService(client)
			resources, err := resourceSvc.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}

			var matchedUUID string
			for _, r := range resources {
				if r.Name == name {
					matchedUUID = r.UUID
					break
				}
			}

			if matchedUUID == "" {
				return fmt.Errorf("resource with name '%s' not found", name)
			}

			// Deploy using the found UUID
			force, _ := cmd.Flags().GetBool("force")
			deploySvc := service.NewDeploymentService(client)
			result, err := deploySvc.Deploy(ctx, matchedUUID, force)
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
