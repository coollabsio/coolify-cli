package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestMetricsService_RequestPathsAndQueries(t *testing.T) {
	var gotURI string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		gotURI = r.URL.RequestURI()
		_, _ = w.Write([]byte(`{"cpu":null}`))
	}))
	defer server.Close()
	svc := NewMetricsService(api.NewClient(server.URL, "token", api.WithRetries(0)))
	ctx := context.Background()

	tests := []struct {
		name string
		call func() (json.RawMessage, error)
		want string
	}{
		{"server defaults", func() (json.RawMessage, error) {
			return svc.ServerMetrics(ctx, "srv-1", models.MetricsWindow{}, nil)
		}, "/api/v1/servers/srv-1/metrics"},
		{"server minutes and types", func() (json.RawMessage, error) {
			return svc.ServerMetrics(ctx, "srv-1", models.MetricsWindow{Minutes: 30}, []string{"disk", "load"})
		}, "/api/v1/servers/srv-1/metrics?minutes=30&types=disk%2Cload"},
		{"application from and to", func() (json.RawMessage, error) {
			return svc.ApplicationMetrics(ctx, "app-1", models.MetricsWindow{From: "2026-11-20T08:00:00Z", To: "2026-11-20T09:00:00Z"})
		}, "/api/v1/applications/app-1/metrics?from=2026-11-20T08%3A00%3A00Z&to=2026-11-20T09%3A00%3A00Z"},
		{"database minutes", func() (json.RawMessage, error) {
			return svc.DatabaseMetrics(ctx, "db-1", models.MetricsWindow{Minutes: 5})
		}, "/api/v1/databases/db-1/metrics?minutes=5"},
		{"application traffic", func() (json.RawMessage, error) {
			return svc.ApplicationTraffic(ctx, "app-1", "7d")
		}, "/api/v1/applications/app-1/traffic?range=7d"},
		{"service traffic", func() (json.RawMessage, error) {
			return svc.ServiceTraffic(ctx, "svc-1", "30d")
		}, "/api/v1/services/svc-1/traffic?range=30d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := tt.call()
			require.NoError(t, err)
			assert.Equal(t, tt.want, gotURI)
			assert.JSONEq(t, `{"cpu":null}`, string(raw))
		})
	}
}

func TestSummarizeSeries(t *testing.T) {
	// Points out of time order: latest is the point with the highest timestamp, not the last one.
	series := models.MetricSeries{{1000, 10}, {3000, 20}, {2000, 45}}
	assert.Equal(t,
		models.MetricRow{Series: "cpu", Points: "3", Latest: "20.00", Avg: "25.00", Max: "45.00", Unit: "%"},
		SummarizeSeries("cpu", "%", series))
	assert.Equal(t,
		models.MetricRow{Series: "cpu", Points: "-", Latest: "metrics disabled", Avg: "-", Max: "-", Unit: "%"},
		SummarizeSeries("cpu", "%", nil))
	assert.Equal(t,
		models.MetricRow{Series: "cpu", Points: "0", Latest: "-", Avg: "-", Max: "-", Unit: "%"},
		SummarizeSeries("cpu", "%", models.MetricSeries{}))
}

func TestServerMetricRows_DisabledTypesAndSubSeries(t *testing.T) {
	var metrics models.ServerMetrics
	require.NoError(t, json.Unmarshal([]byte(`{
		"cpu": [[1000, 2], [2000, 4]],
		"disk": {"/data": [], "/": [[1000, 40], [2000, 41]]},
		"network": {"rx": [[1000, 100], [2000, 300]], "tx": null},
		"load": null
	}`), &metrics))

	rows := ServerMetricRows(metrics, []string{"cpu", "memory", "disk", "network", "load"})

	assert.Equal(t, []models.MetricRow{
		{Series: "cpu", Points: "2", Latest: "4.00", Avg: "3.00", Max: "4.00", Unit: "%"},
		{Series: "memory", Points: "-", Latest: "metrics disabled", Avg: "-", Max: "-", Unit: "%"},
		{Series: "disk /", Points: "2", Latest: "41.00", Avg: "40.50", Max: "41.00", Unit: "%"},
		{Series: "disk /data", Points: "0", Latest: "-", Avg: "-", Max: "-", Unit: "%"},
		{Series: "network rx", Points: "2", Latest: "300.00", Avg: "200.00", Max: "300.00", Unit: "B/s"},
		{Series: "network tx", Points: "-", Latest: "metrics disabled", Avg: "-", Max: "-", Unit: "B/s"},
		{Series: "load", Points: "-", Latest: "metrics disabled", Avg: "-", Max: "-", Unit: ""},
	}, rows)
}

func TestFormatBytes(t *testing.T) {
	assert.Equal(t, "0 B", FormatBytes(0))
	assert.Equal(t, "1023 B", FormatBytes(1023))
	assert.Equal(t, "14.0 KiB", FormatBytes(14365))
	assert.Equal(t, "1.5 MiB", FormatBytes(1572864))
}
