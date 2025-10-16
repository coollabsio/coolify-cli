package application

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

func NewStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start <uuid>",
		Aliases: []string{"deploy"},
		Short:   "Start an application",
		Long:    `Start an application (initiates a deployment).`,
		Args:    cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			force, _ := cmd.Flags().GetBool("force")
			instantDeploy, _ := cmd.Flags().GetBool("instant-deploy")

			appSvc := service.NewApplicationService(client)
			resp, err := appSvc.Start(ctx, uuid, force, instantDeploy)
			if err != nil {
				return fmt.Errorf("failed to start application: %w", err)
			}

			fmt.Println(resp.Message)
			if resp.DeploymentUUID != nil && *resp.DeploymentUUID != "" {
				fmt.Printf("Deployment UUID: %s\n", *resp.DeploymentUUID)
			}
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Force rebuild")
	cmd.Flags().Bool("instant-deploy", false, "Instant deploy (skip queuing)")
	return cmd
}
