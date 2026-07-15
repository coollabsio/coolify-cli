package application

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "logs <service-uuid> <application-uuid>", Short: "Get logs for an application in a service", Args: cli.ExactArgs(2, "<service-uuid> <application-uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			lines, _ := cmd.Flags().GetInt("lines")
			response, err := service.NewService(client).ApplicationLogs(cmd.Context(), args[0], args[1], lines)
			if err != nil {
				return err
			}
			fmt.Print(response.Logs)
			return nil
		},
	}
	cmd.Flags().IntP("lines", "n", 100, "Number of log lines to retrieve")
	return cmd
}

func newStartCommand() *cobra.Command {
	cmd := lifecycleCommand("start", "Start or redeploy an application in a service", func(cmd *cobra.Command, svc *service.Service, serviceUUID, applicationUUID string) (*models.ServiceLifecycleResponse, error) {
		force, _ := cmd.Flags().GetBool("force")
		latest, _ := cmd.Flags().GetBool("latest")
		return svc.StartApplication(cmd.Context(), serviceUUID, applicationUUID, force, latest)
	})
	cmd.Flags().Bool("force", false, "Rebuild the application")
	cmd.Flags().Bool("latest", false, "Pull the latest image")
	return cmd
}

func newRestartCommand() *cobra.Command {
	return lifecycleCommand("restart", "Restart an application in a service", func(cmd *cobra.Command, svc *service.Service, serviceUUID, applicationUUID string) (*models.ServiceLifecycleResponse, error) {
		return svc.RestartApplication(cmd.Context(), serviceUUID, applicationUUID)
	})
}

func newStopCommand() *cobra.Command {
	return lifecycleCommand("stop", "Stop an application in a service", func(cmd *cobra.Command, svc *service.Service, serviceUUID, applicationUUID string) (*models.ServiceLifecycleResponse, error) {
		return svc.StopApplication(cmd.Context(), serviceUUID, applicationUUID)
	})
}

func lifecycleCommand(name, short string, run func(*cobra.Command, *service.Service, string, string) (*models.ServiceLifecycleResponse, error)) *cobra.Command {
	return &cobra.Command{
		Use: name + " <service-uuid> <application-uuid>", Short: short, Args: cli.ExactArgs(2, "<service-uuid> <application-uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			response, err := run(cmd, service.NewService(client), args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Println(response.Message)
			return nil
		},
	}
}
