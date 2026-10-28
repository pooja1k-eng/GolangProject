package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_api_requests_total",
			Help: "Total HTTP requests processed",
		},
		[]string{"method", "path", "status"},
	)
	RequestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_api_request_duration_seconds",
			Help:    "Request latency distributions",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	ErrorCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "order_api_errors_total",
			Help: "Total application level errors",
		},
	)
	DatabaseErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "order_api_database_errors_total",
			Help: "Total database level errors",
		},
	)
)

func init() {
	prometheus.MustRegister(RequestCount, RequestLatency, ErrorCount, DatabaseErrors)
}

func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())
		RequestCount.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		RequestLatency.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)

		if c.Writer.Status() >= 400 {
			ErrorCount.Inc()
		}
	}
}