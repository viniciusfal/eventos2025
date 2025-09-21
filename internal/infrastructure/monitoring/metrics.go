package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Métricas do sistema
var (
	// HTTP Requests
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_active",
			Help: "Number of active HTTP requests",
		},
	)

	// Database Operations
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	DBQueryCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	// Cache Operations
	CacheHitCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMissCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	// Business Logic Metrics
	CheckinCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "checkins_total",
			Help: "Total number of check-ins",
		},
		[]string{"tenant", "event_type", "method"},
	)

	CheckoutCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "checkouts_total",
			Help: "Total number of check-outs",
		},
		[]string{"tenant", "event_type"},
	)

	UserLoginCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "Total number of user logins",
		},
		[]string{"tenant", "status"},
	)

	// System Metrics
	GoroutinesCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "goroutines_count",
			Help: "Number of goroutines",
		},
	)

	MemoryUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
	)
)

// HTTP Metrics
func RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	HTTPRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
	HTTPRequestCount.WithLabelValues(method, endpoint, status).Inc()
}

func IncActiveRequests() {
	HTTPActiveRequests.Inc()
}

func DecActiveRequests() {
	HTTPActiveRequests.Dec()
}

// Database Metrics
func RecordDBQuery(operation, table string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
	DBQueryCount.WithLabelValues(operation, table, status).Inc()
}

// Cache Metrics
func RecordCacheHit(cacheType string) {
	CacheHitCount.WithLabelValues(cacheType).Inc()
}

func RecordCacheMiss(cacheType string) {
	CacheMissCount.WithLabelValues(cacheType).Inc()
}

// Business Metrics
func RecordCheckin(tenant, eventType, method string) {
	CheckinCount.WithLabelValues(tenant, eventType, method).Inc()
}

func RecordCheckout(tenant, eventType string) {
	CheckoutCount.WithLabelValues(tenant, eventType).Inc()
}

func RecordUserLogin(tenant, status string) {
	UserLoginCount.WithLabelValues(tenant, status).Inc()
}

// System Metrics
func UpdateSystemMetrics() {
	// Goroutines count
	GoroutinesCount.Set(float64(GetGoroutinesCount()))

	// Memory usage (simplified)
	MemoryUsage.Set(float64(GetMemoryUsage()))
}

// Helper functions (implementação simplificada)
func GetGoroutinesCount() int {
	return 100 // Placeholder - implementar com runtime.NumGoroutine()
}

func GetMemoryUsage() int64 {
	return 1024 * 1024 * 100 // Placeholder - implementar com runtime.MemStats
}

// GetMetricsRegistry retorna o registry padrão do Prometheus
func GetMetricsRegistry() prometheus.Gatherer {
	return prometheus.DefaultGatherer
}
