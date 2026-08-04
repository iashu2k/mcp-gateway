package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_gateway_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_gateway_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Invocation metrics
	InvocationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_gateway_invocations_total",
			Help: "Total number of tool invocations",
		},
		[]string{"server", "tool", "status"},
	)

	InvocationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_gateway_invocation_duration_seconds",
			Help:    "Tool invocation duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"server", "tool"},
	)

	// Database metrics
	DatabaseConnectionsOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_gateway_database_connections_open",
			Help: "Number of open database connections",
		},
	)

	// Upstream API metrics
	UpstreamRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_gateway_upstream_requests_total",
			Help: "Total number of upstream API requests (GitHub, etc.)",
		},
		[]string{"service", "status"},
	)
)
