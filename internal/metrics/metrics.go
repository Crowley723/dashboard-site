package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: Namespace + "_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: Namespace + "_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_name"},
	)

	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: Namespace + "_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_name"},
	)

	CacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    Namespace + "_cache_operation_duration_seconds",
			Help:    "Time to complete cache operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"cache_name", "operation"},
	)

	CacheItems = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: Namespace + "_cache_items_total",
			Help: "Current number of items in cache",
		},
		[]string{"cache_name"},
	)

	DataFetchDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "data_fetch_duration_seconds",
			Help:    "Time to fetch data from source",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"query_name", "source"},
	)

	DataFetchErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: Namespace + "_data_fetch_errors_total",
			Help: "Total number of data fetch errors",
		},
		[]string{"query_name", "source"},
	)

	IsLeader = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: Namespace + "_leader_is_leader",
			Help: "1 if this instance is the leader, 0 otherwise",
		},
	)
	LeadershipChanges = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: Namespace + "_leader_changes_total",
			Help: "Total number of leadership changes",
		})
)
