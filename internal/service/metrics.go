package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// MetricsService reads Sentinel metrics and traffic analytics.
// It returns the raw response so JSON output shows the API data unchanged.
type MetricsService struct {
	client *api.Client
}

func NewMetricsService(client *api.Client) *MetricsService {
	return &MetricsService{client: client}
}

// ServerMetrics fetches GET /servers/{uuid}/metrics. Empty types lets the API use its default.
func (s *MetricsService) ServerMetrics(ctx context.Context, uuid string, window models.MetricsWindow, types []string) (json.RawMessage, error) {
	query := windowQuery(window)
	if len(types) > 0 {
		query.Set("types", strings.Join(types, ","))
	}
	return s.get(ctx, "servers/"+url.PathEscape(uuid)+"/metrics", query)
}

// ApplicationMetrics fetches GET /applications/{uuid}/metrics.
func (s *MetricsService) ApplicationMetrics(ctx context.Context, uuid string, window models.MetricsWindow) (json.RawMessage, error) {
	return s.get(ctx, "applications/"+url.PathEscape(uuid)+"/metrics", windowQuery(window))
}

// DatabaseMetrics fetches GET /databases/{uuid}/metrics.
func (s *MetricsService) DatabaseMetrics(ctx context.Context, uuid string, window models.MetricsWindow) (json.RawMessage, error) {
	return s.get(ctx, "databases/"+url.PathEscape(uuid)+"/metrics", windowQuery(window))
}

// ApplicationTraffic fetches GET /applications/{uuid}/traffic.
func (s *MetricsService) ApplicationTraffic(ctx context.Context, uuid, timeRange string) (json.RawMessage, error) {
	return s.get(ctx, "applications/"+url.PathEscape(uuid)+"/traffic", url.Values{"range": {timeRange}})
}

// ServiceTraffic fetches GET /services/{uuid}/traffic.
func (s *MetricsService) ServiceTraffic(ctx context.Context, uuid, timeRange string) (json.RawMessage, error) {
	return s.get(ctx, "services/"+url.PathEscape(uuid)+"/traffic", url.Values{"range": {timeRange}})
}

func (s *MetricsService) get(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var raw json.RawMessage
	if err := s.client.Get(ctx, path, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func windowQuery(window models.MetricsWindow) url.Values {
	query := url.Values{}
	if window.Minutes > 0 {
		query.Set("minutes", strconv.Itoa(window.Minutes))
	}
	if window.From != "" {
		query.Set("from", window.From)
	}
	if window.To != "" {
		query.Set("to", window.To)
	}
	return query
}

// SummarizeSeries builds the table row of one series: point count, latest, average, and maximum value.
func SummarizeSeries(name, unit string, series models.MetricSeries) models.MetricRow {
	row := models.MetricRow{Series: name, Points: strconv.Itoa(len(series)), Latest: "-", Avg: "-", Max: "-", Unit: unit}
	if series == nil {
		row.Points, row.Latest = "-", "metrics disabled"
		return row
	}
	if len(series) == 0 {
		return row
	}
	latest, peak, sum := series[0], series[0][1], 0.0
	for _, point := range series {
		if point[0] >= latest[0] {
			latest = point
		}
		peak = max(peak, point[1])
		sum += point[1]
	}
	row.Latest = formatMetric(latest[1])
	row.Avg = formatMetric(sum / float64(len(series)))
	row.Max = formatMetric(peak)
	return row
}

// ServerMetricRows builds one row per series of the requested types, in the requested order.
func ServerMetricRows(metrics models.ServerMetrics, types []string) []models.MetricRow {
	var rows []models.MetricRow
	for _, metricType := range types {
		switch metricType {
		case models.MetricTypeCPU:
			rows = append(rows, SummarizeSeries("cpu", "%", metrics.CPU))
		case models.MetricTypeMemory:
			rows = append(rows, SummarizeSeries("memory", "%", metrics.Memory))
		case models.MetricTypeDisk:
			if metrics.Disk == nil {
				rows = append(rows, SummarizeSeries("disk", "%", nil))
				continue
			}
			mounts := make([]string, 0, len(metrics.Disk))
			for mount := range metrics.Disk {
				mounts = append(mounts, mount)
			}
			sort.Strings(mounts)
			for _, mount := range mounts {
				rows = append(rows, SummarizeSeries("disk "+mount, "%", metrics.Disk[mount]))
			}
		case models.MetricTypeNetwork:
			if metrics.Network == nil {
				rows = append(rows, SummarizeSeries("network", "B/s", nil))
				continue
			}
			rows = append(rows,
				SummarizeSeries("network rx", "B/s", metrics.Network.RX),
				SummarizeSeries("network tx", "B/s", metrics.Network.TX))
		case models.MetricTypeLoad:
			if metrics.Load == nil {
				rows = append(rows, SummarizeSeries("load", "", nil))
				continue
			}
			rows = append(rows,
				SummarizeSeries("load load1", "", metrics.Load.Load1),
				SummarizeSeries("load load5", "", metrics.Load.Load5),
				SummarizeSeries("load load15", "", metrics.Load.Load15))
		}
	}
	return rows
}

// ContainerMetricRows builds the rows of an application or database: CPU in percent, memory in MB.
func ContainerMetricRows(metrics models.ContainerMetrics) []models.MetricRow {
	return []models.MetricRow{
		SummarizeSeries("cpu", "%", metrics.CPU),
		SummarizeSeries("memory", "MB", metrics.Memory),
	}
}

// TrafficOverviewRows builds the overview table. It is empty when traffic analytics is disabled.
func TrafficOverviewRows(traffic models.ResourceTraffic) []models.TrafficOverviewRow {
	o := traffic.Overview
	if o == nil {
		return nil
	}
	approx := func(flag *bool) string {
		if flag != nil && *flag {
			return " (approximate)"
		}
		return ""
	}
	latency := approx(traffic.LatencyApproximate)
	return []models.TrafficOverviewRow{
		{Metric: "requests", Value: strconv.FormatInt(o.Requests, 10)},
		{Metric: "unique visitors", Value: strconv.FormatInt(o.UniqueVisitors, 10) + approx(traffic.UniquesApproximate)},
		{Metric: "error rate", Value: formatMetric(o.ErrorRate) + "%"},
		{Metric: "latency p50", Value: formatMetric(o.Latency.P50) + " ms" + latency},
		{Metric: "latency p95", Value: formatMetric(o.Latency.P95) + " ms" + latency},
		{Metric: "latency p99", Value: formatMetric(o.Latency.P99) + " ms" + latency},
		{Metric: "bytes in", Value: FormatBytes(o.BytesIn)},
		{Metric: "bytes out", Value: FormatBytes(o.BytesOut)},
		{Metric: "status 2xx", Value: strconv.FormatInt(o.Status.S2xx, 10)},
		{Metric: "status 3xx", Value: strconv.FormatInt(o.Status.S3xx, 10)},
		{Metric: "status 4xx", Value: strconv.FormatInt(o.Status.S4xx, 10)},
		{Metric: "status 5xx", Value: strconv.FormatInt(o.Status.S5xx, 10)},
	}
}

// TrafficPathRows builds the top paths table.
func TrafficPathRows(traffic models.ResourceTraffic) []models.TrafficPathRow {
	rows := make([]models.TrafficPathRow, 0, len(traffic.TopPaths))
	for _, p := range traffic.TopPaths {
		rows = append(rows, models.TrafficPathRow{Path: p.Path, Requests: p.Requests, P95: formatMetric(p.P95) + " ms", S4xx: p.S4xx, S5xx: p.S5xx})
	}
	return rows
}

// TrafficBreakdownSection is one breakdown dimension with its rows.
type TrafficBreakdownSection struct {
	Dimension string
	Rows      []models.TrafficBreakdownRow
}

// TrafficBreakdownSections returns the non-empty breakdown dimensions, sorted by name.
func TrafficBreakdownSections(traffic models.ResourceTraffic) []TrafficBreakdownSection {
	var sections []TrafficBreakdownSection
	for dimension, values := range traffic.Breakdowns {
		if len(values) == 0 {
			continue
		}
		section := TrafficBreakdownSection{Dimension: dimension}
		for _, v := range values {
			section.Rows = append(section.Rows, models.TrafficBreakdownRow{Value: v.Value, Requests: v.Requests, BytesOut: FormatBytes(v.BytesOut)})
		}
		sections = append(sections, section)
	}
	sort.Slice(sections, func(i, j int) bool { return sections[i].Dimension < sections[j].Dimension })
	return sections
}

// FormatBytes formats a byte count with binary units, e.g. 14365 -> "14.0 KiB".
func FormatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func formatMetric(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}
