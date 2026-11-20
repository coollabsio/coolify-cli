package common_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/cmd/application"
	"github.com/coollabsio/coolify-cli/cmd/database"
	"github.com/coollabsio/coolify-cli/cmd/server"
	"github.com/coollabsio/coolify-cli/cmd/service"
)

// fakeAPI answers every request with body and records the request URIs.
func fakeAPI(t *testing.T, body string) *[]string {
	t.Helper()
	var uris []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uris = append(uris, r.URL.RequestURI())
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "coolify")
	require.NoError(t, os.MkdirAll(configDir, 0o750))
	cfg, err := json.Marshal(map[string]any{"instances": []map[string]any{{
		"name": "test", "fqdn": srv.URL, "token": "test-token", "default": true,
	}}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), cfg, 0o600))
	return &uris
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := &cobra.Command{Use: "coolify", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("token", "", "")
	root.PersistentFlags().String("context", "", "")
	root.PersistentFlags().String("format", "table", "")
	root.PersistentFlags().BoolP("show-sensitive", "s", false, "")
	root.PersistentFlags().Bool("debug", false, "")
	root.AddCommand(server.NewServerCommand(), application.NewAppCommand(), database.NewDatabaseCommand(), service.NewServiceCommand())
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), err
}

// tableRow returns the cells of the table row whose first data cell is first, without the # column.
func tableRow(t *testing.T, out, first string) []string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		cells := strings.Split(line, "│")
		if len(cells) < 4 {
			continue
		}
		var trimmed []string
		for _, c := range cells[2 : len(cells)-1] {
			trimmed = append(trimmed, strings.TrimSpace(c))
		}
		if trimmed[0] == first {
			return trimmed
		}
	}
	t.Fatalf("no table row %q in:\n%s", first, out)
	return nil
}

func TestMetricsCommands_RequestPathAndQuery(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"server", "metrics", "srv-1"}, "/api/v1/servers/srv-1/metrics"},
		{[]string{"server", "metrics", "srv-1", "--minutes", "60", "--types", "cpu, disk,cpu,load"}, "/api/v1/servers/srv-1/metrics?minutes=60&types=cpu%2Cdisk%2Cload"},
		{[]string{"server", "metrics", "srv-1", "--from", "2026-11-20T08:00:00Z", "--to", "2026-11-20T09:00:00Z"}, "/api/v1/servers/srv-1/metrics?from=2026-11-20T08%3A00%3A00Z&to=2026-11-20T09%3A00%3A00Z"},
		{[]string{"app", "metrics", "app-1", "--minutes", "15"}, "/api/v1/applications/app-1/metrics?minutes=15"},
		{[]string{"database", "metrics", "db-1", "--from", "2026-11-20T08:00:00Z"}, "/api/v1/databases/db-1/metrics?from=2026-11-20T08%3A00%3A00Z"},
		{[]string{"app", "traffic", "app-1"}, "/api/v1/applications/app-1/traffic?range=24h"},
		{[]string{"service", "traffic", "svc-1", "--range", "30d"}, "/api/v1/services/svc-1/traffic?range=30d"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			uris := fakeAPI(t, `{}`)
			_, err := run(t, append(tt.args, "--format", "json")...)
			require.NoError(t, err)
			assert.Equal(t, []string{tt.want}, *uris)
		})
	}
}

func TestMetricsCommands_RejectInvalidFlagsBeforeCallingAPI(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"server", "metrics", "srv-1", "--minutes", "5", "--from", "2026-11-20T08:00:00Z"}, "--minutes cannot be used with --from or --to"},
		{[]string{"app", "metrics", "app-1", "--minutes", "5", "--to", "2026-11-20T09:00:00Z"}, "--minutes cannot be used with --from or --to"},
		{[]string{"database", "metrics", "db-1", "--to", "2026-11-20T09:00:00Z"}, "--to needs --from"},
		{[]string{"server", "metrics", "srv-1", "--types", "cpu,gpu"}, `unknown metric type "gpu", use: cpu, memory, disk, network, load`},
		{[]string{"app", "traffic", "app-1", "--range", "1y"}, "--range must be one of: 24h, 7d, 30d"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			uris := fakeAPI(t, `{}`)
			_, err := run(t, tt.args...)
			require.EqualError(t, err, tt.want)
			assert.Empty(t, *uris)
		})
	}
}

func TestServerMetrics_TableSummarizesEachSeries(t *testing.T) {
	fakeAPI(t, `{
		"cpu": [[1000, 10], [3000, 20], [2000, 45]],
		"memory": null,
		"disk": {"\/": [[1000, 40], [2000, 41]]},
		"network": {"rx": [[1000, 100], [2000, 300]], "tx": [[1000, 1], [2000, 2]]},
		"load": null
	}`)

	out, err := run(t, "server", "metrics", "srv-1", "--types", "cpu,memory,disk,network,load")
	require.NoError(t, err)

	assert.Equal(t, []string{"series", "points", "latest", "avg", "max", "unit"}, tableRow(t, out, "series"))
	assert.Equal(t, []string{"cpu", "3", "20.00", "25.00", "45.00", "%"}, tableRow(t, out, "cpu"))
	assert.Equal(t, []string{"memory", "-", "metrics disabled", "-", "-", "%"}, tableRow(t, out, "memory"))
	assert.Equal(t, []string{"disk /", "2", "41.00", "40.50", "41.00", "%"}, tableRow(t, out, "disk /"))
	assert.Equal(t, []string{"network rx", "2", "300.00", "200.00", "300.00", "B/s"}, tableRow(t, out, "network rx"))
	assert.Equal(t, []string{"network tx", "2", "2.00", "1.50", "2.00", "B/s"}, tableRow(t, out, "network tx"))
	assert.Equal(t, []string{"load", "-", "metrics disabled", "-", "-", ""}, tableRow(t, out, "load"))
}

func TestContainerMetrics_TableUsesMegabytesForMemory(t *testing.T) {
	fakeAPI(t, `{"cpu": [[1000, 0.5], [2000, 1.5]], "memory": [[1000, 100], [2000, 120], [3000, 110]]}`)

	out, err := run(t, "database", "metrics", "db-1")
	require.NoError(t, err)

	assert.Equal(t, []string{"cpu", "2", "1.50", "1.00", "1.50", "%"}, tableRow(t, out, "cpu"))
	assert.Equal(t, []string{"memory", "3", "110.00", "110.00", "120.00", "MB"}, tableRow(t, out, "memory"))
}

func TestMetrics_JSONPrintsAPIResponseUnchanged(t *testing.T) {
	body := `{"cpu":[[1000,1.5]],"disk":{"\/":[[1000,40]]},"memory":null}`
	fakeAPI(t, body)

	out, err := run(t, "server", "metrics", "srv-1", "--types", "cpu,disk,memory", "--format", "json")
	require.NoError(t, err)
	assert.Equal(t, body+"\n", out)
}

func TestTraffic_TableShowsOverviewPathsAndBreakdowns(t *testing.T) {
	fakeAPI(t, `{
		"enabled": true, "range": "7d",
		"overview": {"requests": 35, "unique_visitors": 3, "bytes_in": 512, "bytes_out": 14365, "error_rate": 14.29,
			"latency": {"p50": 0.5, "p95": 12.25, "p99": 40}, "status": {"s2xx": 30, "s3xx": 0, "s4xx": 4, "s5xx": 1}},
		"latency_approximate": true, "uniques_approximate": false,
		"series": [],
		"top_paths": [{"path": "/missing", "domain": "a.test", "requests": 5, "bytes_out": 55, "s4xx": 4, "s5xx": 1, "p95": 0.3}],
		"breakdowns": {"country": [], "status": [{"value": "404", "requests": 5, "bytes_out": 2048}]},
		"attribution": "GeoLite2 by MaxMind"
	}`)

	out, err := run(t, "service", "traffic", "svc-1", "--range", "7d")
	require.NoError(t, err)

	assert.Contains(t, out, "Traffic (7d)")
	assert.Equal(t, []string{"requests", "35"}, tableRow(t, out, "requests")[:2])
	assert.Equal(t, []string{"unique visitors", "3"}, tableRow(t, out, "unique visitors"))
	assert.Equal(t, []string{"error rate", "14.29%"}, tableRow(t, out, "error rate"))
	assert.Equal(t, []string{"latency p95", "12.25 ms (approximate)"}, tableRow(t, out, "latency p95"))
	assert.Equal(t, []string{"bytes in", "512 B"}, tableRow(t, out, "bytes in"))
	assert.Equal(t, []string{"bytes out", "14.0 KiB"}, tableRow(t, out, "bytes out"))
	assert.Equal(t, []string{"status 5xx", "1"}, tableRow(t, out, "status 5xx"))
	assert.Equal(t, []string{"path", "requests", "p95", "4xx", "5xx"}, tableRow(t, out, "path"))
	assert.Equal(t, []string{"/missing", "5", "0.30 ms", "4", "1"}, tableRow(t, out, "/missing"))
	assert.Contains(t, out, "Top status")
	assert.Equal(t, []string{"404", "5", "2.0 KiB"}, tableRow(t, out, "404"))
	assert.NotContains(t, out, "Top country", "empty dimensions are left out")
	assert.Contains(t, out, "GeoLite2 by MaxMind")
}

func TestTraffic_Disabled(t *testing.T) {
	fakeAPI(t, `{"enabled": false, "range": "24h", "overview": null, "latency_approximate": null, "uniques_approximate": null,
		"series": null, "top_paths": null, "breakdowns": null, "attribution": null}`)

	out, err := run(t, "app", "traffic", "app-1")
	require.NoError(t, err)
	assert.Equal(t, "Traffic analytics is disabled on this server.\n", out)
}
