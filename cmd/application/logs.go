package application

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <uuid>",
		Short: "Get application logs",
		Long: `Retrieve logs for an application. Use --follow to continuously stream new logs.

For Docker Compose applications with multiple services, pass --service <name>
(the compose service key, e.g. web or db) to select which container's logs to return.
Without --service, the API returns logs from the first running container.`,
		Args: cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			lines, _ := cmd.Flags().GetInt("lines")
			follow, _ := cmd.Flags().GetBool("follow")
			showTimestamps, _ := cmd.Flags().GetBool("show-timestamps")
			serviceName, _ := cmd.Flags().GetString("service")
			appSvc := service.NewApplicationService(client)

			if !follow {
				resp, err := appSvc.Logs(ctx, uuid, lines, showTimestamps, serviceName)
				if err != nil {
					return fmt.Errorf("failed to get logs: %w", err)
				}
				fmt.Print(resp.Logs)
				return nil
			}

			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			lastLogs := ""

			resp, err := appSvc.Logs(ctx, uuid, lines, showTimestamps, serviceName)
			if err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
			}
			fmt.Print(resp.Logs)
			lastLogs = resp.Logs

			for {
				select {
				case <-sigChan:
					fmt.Println("\nStopping log follow...")
					return nil
				case <-ticker.C:
					resp, err := appSvc.Logs(ctx, uuid, lines, showTimestamps, serviceName)
					if err != nil {
						continue
					}
					if resp.Logs != lastLogs {
						if len(resp.Logs) > len(lastLogs) && strings.HasPrefix(resp.Logs, lastLogs) {
							fmt.Print(resp.Logs[len(lastLogs):])
						} else {
							fmt.Print(resp.Logs)
						}
						lastLogs = resp.Logs
					}
				}
			}
		},
	}

	cmd.Flags().IntP("lines", "n", 100, "Number of log lines to retrieve")
	cmd.Flags().BoolP("follow", "f", false, "Follow log output (like tail -f)")
	cmd.Flags().Bool("show-timestamps", false, "Show timestamps in log output")
	cmd.Flags().String("service", "", "Docker Compose service name (selects one container in multi-service apps)")
	return cmd
}
