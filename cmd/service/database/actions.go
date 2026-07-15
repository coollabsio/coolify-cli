package database

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "logs <service-uuid> <database-uuid>", Short: "Get logs for a database in a service", Args: cli.ExactArgs(2, "<service-uuid> <database-uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			lines, _ := cmd.Flags().GetInt("lines")
			response, err := service.NewService(client).DatabaseLogs(cmd.Context(), args[0], args[1], lines)
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
	cmd := databaseLifecycleCommand("start", "Start or redeploy a database in a service", func(cmd *cobra.Command, svc *service.Service, serviceUUID, databaseUUID string) (*models.ServiceLifecycleResponse, error) {
		force, _ := cmd.Flags().GetBool("force")
		latest, _ := cmd.Flags().GetBool("latest")
		return svc.StartDatabase(cmd.Context(), serviceUUID, databaseUUID, force, latest)
	})
	cmd.Flags().Bool("force", false, "Rebuild the database container")
	cmd.Flags().Bool("latest", false, "Pull the latest image")
	return cmd
}

func newRestartCommand() *cobra.Command {
	return databaseLifecycleCommand("restart", "Restart a database in a service", func(cmd *cobra.Command, svc *service.Service, serviceUUID, databaseUUID string) (*models.ServiceLifecycleResponse, error) {
		return svc.RestartDatabase(cmd.Context(), serviceUUID, databaseUUID)
	})
}

func newStopCommand() *cobra.Command {
	return databaseLifecycleCommand("stop", "Stop a database in a service", func(cmd *cobra.Command, svc *service.Service, serviceUUID, databaseUUID string) (*models.ServiceLifecycleResponse, error) {
		return svc.StopDatabase(cmd.Context(), serviceUUID, databaseUUID)
	})
}

func databaseLifecycleCommand(name, short string, run func(*cobra.Command, *service.Service, string, string) (*models.ServiceLifecycleResponse, error)) *cobra.Command {
	return &cobra.Command{
		Use: name + " <service-uuid> <database-uuid>", Short: short, Args: cli.ExactArgs(2, "<service-uuid> <database-uuid>"),
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
