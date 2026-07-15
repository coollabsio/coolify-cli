package database

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <uuid>",
		Short: "Get database logs",
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			lines, _ := cmd.Flags().GetInt("lines")
			showTimestamps, _ := cmd.Flags().GetBool("show-timestamps")
			response, err := service.NewDatabaseService(client).Logs(cmd.Context(), args[0], lines, showTimestamps)
			if err != nil {
				return fmt.Errorf("failed to get database logs: %w", err)
			}
			fmt.Print(response.Logs)
			return nil
		},
	}
	cmd.Flags().IntP("lines", "n", 100, "Number of log lines to retrieve")
	cmd.Flags().Bool("show-timestamps", false, "Include timestamps in logs")
	return cmd
}
