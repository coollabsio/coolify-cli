package service

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	internalservice "github.com/coollabsio/coolify-cli/internal/service"
)

func NewLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <uuid>",
		Short: "Get logs for a service sub-resource",
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			subServiceName, _ := cmd.Flags().GetString("sub-service-name")
			if subServiceName == "" {
				return fmt.Errorf("--sub-service-name is required")
			}
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			lines, _ := cmd.Flags().GetInt("lines")
			showTimestamps, _ := cmd.Flags().GetBool("show-timestamps")
			response, err := internalservice.NewService(client).Logs(cmd.Context(), args[0], subServiceName, lines, showTimestamps)
			if err != nil {
				return fmt.Errorf("failed to get service logs: %w", err)
			}
			fmt.Print(response.Logs)
			return nil
		},
	}
	cmd.Flags().String("sub-service-name", "", "Sub-service name from the service applications or databases list (required)")
	cmd.Flags().IntP("lines", "n", 100, "Number of log lines to retrieve")
	cmd.Flags().Bool("show-timestamps", false, "Include timestamps in logs")
	return cmd
}
