package server

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/common"
	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewMetricsCommand gets server metrics from Sentinel.
func NewMetricsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics <server_uuid>",
		Short: "Get server metrics",
		Long: `Get server metrics from Sentinel: CPU and memory in percent, disk usage in percent
per mount, network rx/tx in bytes per second, and load averages.

The table shows one row per series with the number of points and the latest,
average, and maximum value. Use --format json for the raw [unix_ms, value] points.`,
		Example: `  coolify server metrics <server_uuid>
  coolify server metrics <server_uuid> --minutes 60 --types cpu,memory,disk,network,load
  coolify server metrics <server_uuid> --from 2026-11-20T08:00:00Z --to 2026-11-20T09:00:00Z`,
		Args: cli.ExactArgs(1, "<server_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			window, err := common.MetricsWindowFromFlags(cmd)
			if err != nil {
				return err
			}
			typesFlag, _ := cmd.Flags().GetString("types")
			var types []string
			for _, t := range strings.Split(typesFlag, ",") {
				t = strings.TrimSpace(t)
				if t == "" || slices.Contains(types, t) {
					continue
				}
				if !slices.Contains(models.ServerMetricTypes, t) {
					return fmt.Errorf("unknown metric type %q, use: %s", t, strings.Join(models.ServerMetricTypes, ", "))
				}
				types = append(types, t)
			}
			svc, err := common.NewMetricsService(cmd)
			if err != nil {
				return err
			}
			raw, err := svc.ServerMetrics(cmd.Context(), args[0], window, types)
			if err != nil {
				return fmt.Errorf("failed to get server metrics: %w", err)
			}
			if len(types) == 0 {
				types = models.DefaultServerMetricTypes
			}
			return common.WriteMetrics(cmd, raw, func(m models.ServerMetrics) []models.MetricRow {
				return service.ServerMetricRows(m, types)
			})
		},
	}
	common.BindMetricsWindowFlags(cmd)
	cmd.Flags().String("types", "", "Comma-separated metric types: cpu, memory, disk, network, load (API default: cpu,memory)")
	return cmd
}
