package application

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewDeploymentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployments",
		Short: "Deployment related commands for an application",
		Long:  `Manage deployments for a specific application. List deployments or view deployment logs.`,
	}

	cmd.AddCommand(NewListDeploymentsCommand())
	cmd.AddCommand(NewLogsDeploymentsCommand())
	return cmd
}

func NewListDeploymentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <app-uuid>",
		Short: "List all deployments for an application",
		Long:  `Retrieve a list of all deployments for a specific application.`,
		Args:  cli.ExactArgs(1, "<app-uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			appUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			deploySvc := service.NewDeploymentService(client)
			deployments, err := deploySvc.ListByApplication(ctx, appUUID)
			if err != nil {
				return fmt.Errorf("failed to list deployments: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(deployments)
		},
	}

	return cmd
}

func NewLogsDeploymentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <app-uuid> [deployment-uuid]",
		Short: "Get deployment logs for an application",
		Long: `Retrieve deployment logs for a specific application or deployment.

If only app-uuid is provided, retrieves logs from the latest deployment.
If deployment-uuid is also provided, retrieves logs for that specific deployment.

Use --follow to continuously stream new logs.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			appUUID := args[0]
			var deploymentUUID string
			if len(args) == 2 {
				deploymentUUID = args[1]
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			lines, _ := cmd.Flags().GetInt("lines")
			follow, _ := cmd.Flags().GetBool("follow")
			debugLogs, _ := cmd.Flags().GetBool("debuglogs")
			format, _ := cmd.Flags().GetString("format")
			deploySvc := service.NewDeploymentService(client)

			// Function to get logs based on whether we have a deployment UUID
			// Returns raw or formatted based on format flag
			getLogs := func() (string, error) {
				if deploymentUUID != "" {
					return deploySvc.GetLogsByDeploymentWithFormat(ctx, deploymentUUID, debugLogs, format)
				}
				// Get logs from the latest deployment
				// Use take=1 internally to efficiently fetch only the most recent deployment
				return deploySvc.GetLogsByApplicationWithFormat(ctx, appUUID, 1, debugLogs, format)
			}

			if !follow {
				logs, err := getLogs()
				if err != nil {
					return fmt.Errorf("failed to get deployment logs: %w", err)
				}

				// Apply line limit if specified (only for text output)
				if lines > 0 && format == "table" {
					logs = limitLogLines(logs, lines)
				}

				fmt.Print(logs)
				return nil
			}

			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			lastLogs := ""

			logs, err := getLogs()
			if err != nil {
				return fmt.Errorf("failed to get deployment logs: %w", err)
			}
			fmt.Print(logs)
			lastLogs = logs

			for {
				select {
				case <-sigChan:
					fmt.Println("\nStopping log follow...")
					return nil
				case <-ticker.C:
					logs, err := getLogs()
					if err != nil {
						continue
					}
					if logs != lastLogs {
						if len(logs) > len(lastLogs) && strings.HasPrefix(logs, lastLogs) {
							fmt.Print(logs[len(lastLogs):])
						} else {
							fmt.Print(logs)
						}
						lastLogs = logs
					}
				}
			}
		},
	}

	cmd.Flags().IntP("lines", "n", 0, "Number of log lines to display (0 = all)")
	cmd.Flags().BoolP("follow", "f", false, "Follow log output (like tail -f)")
	cmd.Flags().Bool("debuglogs", false, "Show debug logs (includes hidden commands and internal operations)")
	return cmd
}

// limitLogLines limits the output to the last N lines
func limitLogLines(logs string, n int) string {
	if n <= 0 {
		return logs
	}

	// Trim trailing newline to avoid empty element at the end
	logs = strings.TrimRight(logs, "\n")
	lines := strings.Split(logs, "\n")

	// If we have fewer lines than requested, return all
	if len(lines) <= n {
		return logs + "\n"
	}

	// Get the last N lines
	lastLines := lines[len(lines)-n:]
	return strings.Join(lastLines, "\n") + "\n"
}
