package previews

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// logsMinimumVersion is the first Coolify version with the preview logs endpoint.
const logsMinimumVersion = "4.3.23"

// followInterval is the polling interval used by --follow.
const followInterval = 2 * time.Second

func NewLogsPreviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <app_uuid> <pr_id>",
		Short: "Get preview deployment logs",
		Long: `Retrieve runtime logs for a preview deployment by pull request ID. Use --follow to continuously stream new logs.

For Docker Compose applications with multiple services, pass --service <name>
(the compose service key, e.g. web or db) to select which container's logs to return.
Without --service, the API returns logs from the first running container.`,
		Args: cli.ExactArgs(2, "<app_uuid> <pr_id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			appUUID := args[0]

			prID, err := parsePullRequestID(args[1])
			if err != nil {
				return err
			}

			lines, _ := cmd.Flags().GetInt("lines")
			if lines < -1 {
				return fmt.Errorf("invalid --lines: must be -1 (all) or a non-negative integer")
			}
			follow, _ := cmd.Flags().GetBool("follow")
			showTimestamps, _ := cmd.Flags().GetBool("show-timestamps")
			serviceName, _ := cmd.Flags().GetString("service")

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := cli.CheckMinimumVersion(ctx, client, logsMinimumVersion); err != nil {
				return err
			}

			appSvc := service.NewApplicationService(client)
			fetch := func(ctx context.Context) (string, error) {
				resp, err := appSvc.PreviewLogs(ctx, appUUID, prID, lines, showTimestamps, serviceName)
				if err != nil {
					return "", err
				}
				return resp.Logs, nil
			}

			out := cmd.OutOrStdout()
			if !follow {
				logs, err := fetch(ctx)
				if err != nil {
					return fmt.Errorf("failed to get preview logs: %w", err)
				}
				fmt.Fprint(out, logs)
				return nil
			}

			sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
			return followLogs(sigCtx, out, followInterval, fetch)
		},
	}

	cmd.Flags().IntP("lines", "n", 100, "Number of log lines to retrieve (-1 for all)")
	cmd.Flags().BoolP("follow", "f", false, "Follow log output (like tail -f)")
	cmd.Flags().Bool("show-timestamps", false, "Show timestamps in log output")
	cmd.Flags().String("service", "", "Docker Compose service name (selects one container in multi-service apps)")
	return cmd
}

// followLogs prints the initial logs, then polls fetch every interval and
// prints only what changed, until ctx is cancelled. Same semantics as
// `coolify app logs --follow`: transient poll errors are ignored, appended
// output is printed as a suffix, and rotated output is reprinted in full.
func followLogs(ctx context.Context, out io.Writer, interval time.Duration, fetch func(context.Context) (string, error)) error {
	lastLogs, err := fetch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get preview logs: %w", err)
	}
	fmt.Fprint(out, lastLogs)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(out, "\nStopping log follow...")
			return nil
		case <-ticker.C:
			logs, err := fetch(ctx)
			if err != nil || logs == lastLogs {
				continue
			}
			if len(logs) > len(lastLogs) && strings.HasPrefix(logs, lastLogs) {
				fmt.Fprint(out, logs[len(lastLogs):])
			} else {
				fmt.Fprint(out, logs)
			}
			lastLogs = logs
		}
	}
}
