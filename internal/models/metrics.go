package models

// MetricSeries holds Sentinel data points as [unix_timestamp_ms, value].
// It is nil when the API returns null, which means metrics are disabled for the server.
type MetricSeries [][2]float64

// MetricsWindow selects the metric history: the last Minutes, or From to To (ISO 8601).
// Zero values are not sent, so the API applies its defaults.
type MetricsWindow struct {
	Minutes int
	From    string
	To      string
}

// Server metric types accepted by GET /servers/{uuid}/metrics.
const (
	MetricTypeCPU     = "cpu"
	MetricTypeMemory  = "memory"
	MetricTypeDisk    = "disk"
	MetricTypeNetwork = "network"
	MetricTypeLoad    = "load"
)

// ServerMetricTypes lists every server metric type.
var ServerMetricTypes = []string{MetricTypeCPU, MetricTypeMemory, MetricTypeDisk, MetricTypeNetwork, MetricTypeLoad}

// DefaultServerMetricTypes are the types the API returns when types is not sent.
var DefaultServerMetricTypes = []string{MetricTypeCPU, MetricTypeMemory}

// ServerMetrics is the response of GET /servers/{uuid}/metrics.
// Only requested types are present; a requested type is null when metrics are disabled.
type ServerMetrics struct {
	CPU     MetricSeries            `json:"cpu"`
	Memory  MetricSeries            `json:"memory"`
	Disk    map[string]MetricSeries `json:"disk"`
	Network *struct {
		RX MetricSeries `json:"rx"`
		TX MetricSeries `json:"tx"`
	} `json:"network"`
	Load *struct {
		Load1  MetricSeries `json:"load1"`
		Load5  MetricSeries `json:"load5"`
		Load15 MetricSeries `json:"load15"`
	} `json:"load"`
}

// ContainerMetrics is the response of GET /applications/{uuid}/metrics and GET /databases/{uuid}/metrics.
// CPU is in percent and Memory in MB.
type ContainerMetrics struct {
	CPU    MetricSeries `json:"cpu"`
	Memory MetricSeries `json:"memory"`
}

// MetricRow is the table view of one metric series.
type MetricRow struct {
	Series string `json:"series"`
	Points string `json:"points"`
	Latest string `json:"latest"`
	Avg    string `json:"avg"`
	Max    string `json:"max"`
	Unit   string `json:"unit"`
}

// TrafficRanges are the time ranges accepted by the traffic endpoints.
var TrafficRanges = []string{"24h", "7d", "30d"}

// ResourceTraffic is the response of GET /applications/{uuid}/traffic and GET /services/{uuid}/traffic.
// When Enabled is false, every field except Range is null.
type ResourceTraffic struct {
	Enabled  bool   `json:"enabled"`
	Range    string `json:"range"`
	Overview *struct {
		Requests       int64   `json:"requests"`
		UniqueVisitors int64   `json:"unique_visitors"`
		BytesIn        int64   `json:"bytes_in"`
		BytesOut       int64   `json:"bytes_out"`
		ErrorRate      float64 `json:"error_rate"`
		Latency        struct {
			P50 float64 `json:"p50"`
			P95 float64 `json:"p95"`
			P99 float64 `json:"p99"`
		} `json:"latency"`
		Status struct {
			S2xx int64 `json:"s2xx"`
			S3xx int64 `json:"s3xx"`
			S4xx int64 `json:"s4xx"`
			S5xx int64 `json:"s5xx"`
		} `json:"status"`
	} `json:"overview"`
	LatencyApproximate *bool `json:"latency_approximate"`
	UniquesApproximate *bool `json:"uniques_approximate"`
	TopPaths           []struct {
		Path     string  `json:"path"`
		Domain   *string `json:"domain"`
		Requests int64   `json:"requests"`
		BytesOut int64   `json:"bytes_out"`
		S4xx     int64   `json:"s4xx"`
		S5xx     int64   `json:"s5xx"`
		P95      float64 `json:"p95"`
	} `json:"top_paths"`
	Breakdowns map[string][]struct {
		Value    string `json:"value"`
		Requests int64  `json:"requests"`
		BytesOut int64  `json:"bytes_out"`
	} `json:"breakdowns"`
	Attribution *string `json:"attribution"`
}

// TrafficOverviewRow is one line of the traffic overview table.
type TrafficOverviewRow struct {
	Metric string `json:"metric"`
	Value  string `json:"value"`
}

// TrafficPathRow is the table view of one top path.
type TrafficPathRow struct {
	Path     string `json:"path"`
	Requests int64  `json:"requests"`
	P95      string `json:"p95"`
	S4xx     int64  `json:"4xx"`
	S5xx     int64  `json:"5xx"`
}

// TrafficBreakdownRow is the table view of one breakdown value.
type TrafficBreakdownRow struct {
	Value    string `json:"value"`
	Requests int64  `json:"requests"`
	BytesOut string `json:"bytes_out"`
}
