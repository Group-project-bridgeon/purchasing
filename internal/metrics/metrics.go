package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTPRequestsTotal counts all incoming HTTP requests labeled by method, path template, and status.
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "purchasing",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed by method, path, and HTTP status code.",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration tracks the latency distribution of HTTP requests.
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "purchasing",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Histogram of latencies for HTTP requests.",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestsInFlight gauges the number of concurrent active HTTP requests.
	HTTPRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "purchasing",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Current number of in-flight HTTP requests.",
		},
	)

	// OutboxEventsTotal counts dispatched vs failed domain events.
	OutboxEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "purchasing",
			Subsystem: "outbox",
			Name:      "events_total",
			Help:      "Total number of transactional outbox events processed, partitioned by status.",
		},
		[]string{"status"},
	)

	// OutboxBatchDuration records the duration of background outbox batch processing cycles.
	OutboxBatchDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "purchasing",
			Subsystem: "outbox",
			Name:      "batch_duration_seconds",
			Help:      "Duration of outbox batch processing iterations.",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		},
	)
)

func init() {
	// Register metrics with the default Prometheus registerer safely
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPRequestsInFlight,
		OutboxEventsTotal,
		OutboxBatchDuration,
	)
}

// RecordHTTPRequest records latency, status code, and increments request counters.
func RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	statusStr := strconv.Itoa(statusCode)
	HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
	HTTPRequestDuration.WithLabelValues(method, path, statusStr).Observe(duration.Seconds())
}

// RecordOutboxEvent increments the outbox events counter by status (e.g., "dispatched", "failed").
func RecordOutboxEvent(status string) {
	OutboxEventsTotal.WithLabelValues(status).Inc()
}

// RecordOutboxBatchDuration observes the execution duration of an outbox processing batch.
func RecordOutboxBatchDuration(duration time.Duration) {
	OutboxBatchDuration.Observe(duration.Seconds())
}

// DBPoolCollector implements prometheus.Collector to export live pgxpool connection pool statistics.
type DBPoolCollector struct {
	pool           *pgxpool.Pool
	totalConnsDesc *prometheus.Desc
	idleConnsDesc  *prometheus.Desc
	acqConnsDesc   *prometheus.Desc
	maxConnsDesc   *prometheus.Desc
}

var (
	dbCollectorLock           sync.Mutex
	registeredDBPoolCollector *DBPoolCollector
)

// RegisterDBPoolCollector creates and registers a dynamic PostgreSQL pool metrics collector safely.
func RegisterDBPoolCollector(pool *pgxpool.Pool) *DBPoolCollector {
	dbCollectorLock.Lock()
	defer dbCollectorLock.Unlock()

	if registeredDBPoolCollector != nil {
		registeredDBPoolCollector.pool = pool
		return registeredDBPoolCollector
	}

	c := &DBPoolCollector{
		pool: pool,
		totalConnsDesc: prometheus.NewDesc(
			"purchasing_db_pool_total_connections",
			"Current total number of open database connections in the pool.",
			nil, nil,
		),
		idleConnsDesc: prometheus.NewDesc(
			"purchasing_db_pool_idle_connections",
			"Current number of idle database connections in the pool.",
			nil, nil,
		),
		acqConnsDesc: prometheus.NewDesc(
			"purchasing_db_pool_acquired_connections",
			"Current number of acquired (in-use) database connections.",
			nil, nil,
		),
		maxConnsDesc: prometheus.NewDesc(
			"purchasing_db_pool_max_connections",
			"Maximum configured database connections for the pool.",
			nil, nil,
		),
	}

	_ = prometheus.Register(c)
	registeredDBPoolCollector = c
	return c
}

// Describe sends the metric descriptions to Prometheus.
func (c *DBPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalConnsDesc
	ch <- c.idleConnsDesc
	ch <- c.acqConnsDesc
	ch <- c.maxConnsDesc
}

// Collect reads dynamic stats from pgxpool and publishes gauges to Prometheus.
func (c *DBPoolCollector) Collect(ch chan<- prometheus.Metric) {
	if c.pool == nil {
		return
	}
	stat := c.pool.Stat()

	ch <- prometheus.MustNewConstMetric(
		c.totalConnsDesc,
		prometheus.GaugeValue,
		float64(stat.TotalConns()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.idleConnsDesc,
		prometheus.GaugeValue,
		float64(stat.IdleConns()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.acqConnsDesc,
		prometheus.GaugeValue,
		float64(stat.AcquiredConns()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.maxConnsDesc,
		prometheus.GaugeValue,
		float64(stat.MaxConns()),
	)
}
