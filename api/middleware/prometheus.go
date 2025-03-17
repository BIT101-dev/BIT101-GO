/*
 * @Author: flwfdd
 * @Date: 2025-03-18 00:54:50
 * @LastEditTime: 2025-03-18 01:32:36
 * @Description: _(:з」∠)_
 */
package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// 请求总数计数器
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bit101_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status", "ip", "user"},
	)

	// 活跃连接数
	httpActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "bit101_http_active_requests",
			Help: "Current number of active HTTP requests",
		},
	)

	// 请求处理时间直方图
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bit101_request_duration_seconds",
			Help:    "Request processing time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// 请求大小 - 修正语法并使用 promauto
	requestSize = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bit101_request_size_bytes",
			Help: "Size of HTTP requests in bytes",
		},
		[]string{"method", "path"},
	)

	// 响应大小 - 修正语法并使用 promauto
	responseSize = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bit101_response_size_bytes",
			Help: "Size of HTTP responses in bytes",
		},
		[]string{"method", "path"},
	)
)

// PrometheusMiddleware 返回一个Gin中间件，用于收集Prometheus指标
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method

		// 开始计时
		start := time.Now()

		// 增加活跃请求计数
		httpActiveRequests.Inc()

		// 记录请求大小
		if size := c.Request.ContentLength; size > 0 {
			requestSize.WithLabelValues(method, path).Add(float64(size))
		}

		// 处理请求
		c.Next()

		// 减少活跃请求计数
		httpActiveRequests.Dec()

		// 记录请求状态
		status := fmt.Sprintf("%d", c.Writer.Status())
		clientIP := c.ClientIP()
		user := "unknown"
		userCtx, err := GetUserContext(c)
		if err != nil {
			user = userCtx.UIDStr
		}
		requestsTotal.WithLabelValues(method, path, status, clientIP, user).Inc()

		// 记录响应时间
		duration := time.Since(start).Seconds()
		requestDuration.WithLabelValues(method, path).Observe(duration)

		// 记录响应大小
		if size := c.Writer.Size(); size > 0 {
			responseSize.WithLabelValues(method, path).Add(float64(size))
		}
	}
}

// PrometheusHandler 返回Prometheus指标处理函数
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
