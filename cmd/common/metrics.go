package common

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// BindMetricsWindowFlags registers --minutes, --from, and --to.
func BindMetricsWindowFlags(cmd *cobra.Command) {
	cmd.Flags().Int("minutes", 0, "Minutes of history before now (API default: 10). Cannot be used with --from or --to")
	cmd.Flags().String("from", "", "Start of the history, ISO 8601 (e.g. 2026-11-20T08:00:00Z)")
	cmd.Flags().String("to", "", "End of the history, ISO 8601 (default: now). Needs --from")
}

// MetricsWindowFromFlags reads the window flags and rejects combinations the API refuses.
func MetricsWindowFromFlags(cmd *cobra.Command) (models.MetricsWindow, error) {
	minutes, _ := cmd.Flags().GetInt("minutes")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	window := models.MetricsWindow{Minutes: minutes, From: from, To: to}
	if cmd.Flags().Changed("minutes") && (from != "" || to != "") {
		return window, fmt.Errorf("--minutes cannot be used with --from or --to")
	}
	if cmd.Flags().Changed("minutes") && minutes < 1 {
		return window, fmt.Errorf("--minutes must be at least 1")
	}
	if to != "" && from == "" {
		return window, fmt.Errorf("--to needs --from")
	}
	return window, nil
}

// NewMetricsService builds a MetricsService from the command's context.
func NewMetricsService(cmd *cobra.Command) (*service.MetricsService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewMetricsService(client), nil
}

// writeRaw prints the API response unchanged for json and pretty and reports whether it did.
func writeRaw(cmd *cobra.Command, raw json.RawMessage) (bool, error) {
	format, _ := cmd.Flags().GetString("format")
	if format == output.FormatTable {
		return false, nil
	}
	formatter, err := output.NewFormatter(format, output.Options{Writer: cmd.OutOrStdout()})
	if err != nil {
		return true, err
	}
	return true, formatter.Format(raw)
}

// WriteMetrics prints the raw response for json and pretty, or the rows that tableRows builds from it.
func WriteMetrics[T any](cmd *cobra.Command, raw json.RawMessage, tableRows func(T) []models.MetricRow) error {
	if done, err := writeRaw(cmd, raw); done {
		return err
	}
	var decoded T
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	return output.NewTableFormatter(output.Options{Writer: cmd.OutOrStdout()}).Format(tableRows(decoded))
}

// NewContainerMetricsCommand builds `metrics <uuid>` for applications and databases.
func NewContainerMetricsCommand(resource string, fetch func(*service.MetricsService, context.Context, string, models.MetricsWindow) (json.RawMessage, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics <uuid>",
		Short: fmt.Sprintf("Get %s CPU and memory metrics", resource),
		Long: fmt.Sprintf(`Get CPU (percent) and memory (MB) metrics of the %s container from Sentinel.

The table shows one row per series with the number of points and the latest,
average, and maximum value. Use --format json for the raw [unix_ms, value] points.`, resource),
		Args: cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			window, err := MetricsWindowFromFlags(cmd)
			if err != nil {
				return err
			}
			svc, err := NewMetricsService(cmd)
			if err != nil {
				return err
			}
			raw, err := fetch(svc, cmd.Context(), args[0], window)
			if err != nil {
				return fmt.Errorf("failed to get %s metrics: %w", resource, err)
			}
			return WriteMetrics(cmd, raw, service.ContainerMetricRows)
		},
	}
	BindMetricsWindowFlags(cmd)
	return cmd
}

// NewTrafficCommand builds `traffic <uuid>` for applications and services.
func NewTrafficCommand(resource string, fetch func(*service.MetricsService, context.Context, string, string) (json.RawMessage, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traffic <uuid>",
		Short: fmt.Sprintf("Get %s traffic analytics", resource),
		Long: fmt.Sprintf(`Get traffic analytics of the %s that Sentinel records from the proxy access log.

The table shows the overview, the top paths, and the top values per dimension
(country, browser, status, and more). Use --format json for the full response,
including the time series.`, resource),
		Args: cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			timeRange, _ := cmd.Flags().GetString("range")
			if !slices.Contains(models.TrafficRanges, timeRange) {
				return fmt.Errorf("--range must be one of: %s", strings.Join(models.TrafficRanges, ", "))
			}
			svc, err := NewMetricsService(cmd)
			if err != nil {
				return err
			}
			raw, err := fetch(svc, cmd.Context(), args[0], timeRange)
			if err != nil {
				return fmt.Errorf("failed to get %s traffic: %w", resource, err)
			}
			if done, err := writeRaw(cmd, raw); done {
				return err
			}
			var traffic models.ResourceTraffic
			if err := json.Unmarshal(raw, &traffic); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			return writeTrafficTables(cmd, traffic)
		},
	}
	cmd.Flags().String("range", "24h", "Time range: "+strings.Join(models.TrafficRanges, ", "))
	return cmd
}

func writeTrafficTables(cmd *cobra.Command, traffic models.ResourceTraffic) error {
	out := cmd.OutOrStdout()
	if !traffic.Enabled {
		_, err := fmt.Fprintln(out, "Traffic analytics is disabled on this server.")
		return err
	}
	table := output.NewTableFormatter(output.Options{Writer: out})
	fmt.Fprintf(out, "Traffic (%s)\n", traffic.Range)
	if err := table.Format(service.TrafficOverviewRows(traffic)); err != nil {
		return err
	}
	fmt.Fprintln(out, "\nTop paths")
	if err := table.Format(service.TrafficPathRows(traffic)); err != nil {
		return err
	}
	for _, section := range service.TrafficBreakdownSections(traffic) {
		fmt.Fprintf(out, "\nTop %s\n", section.Dimension)
		if err := table.Format(section.Rows); err != nil {
			return err
		}
	}
	if traffic.Attribution != nil && *traffic.Attribution != "" {
		fmt.Fprintf(out, "\n%s\n", *traffic.Attribution)
	}
	return nil
}
