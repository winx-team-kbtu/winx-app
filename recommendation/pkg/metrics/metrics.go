package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"service", "method", "path", "status"})

	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method", "path", "status"})
)

func Middleware(service string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		ctx.Next()

		status := strconv.Itoa(ctx.Writer.Status())
		duration := time.Since(start).Seconds()
		path := ctx.FullPath()
		if path == "" {
			path = "unknown"
		}

		requestsTotal.WithLabelValues(service, ctx.Request.Method, path, status).Inc()
		requestDuration.WithLabelValues(service, ctx.Request.Method, path, status).Observe(duration)
	}
}
