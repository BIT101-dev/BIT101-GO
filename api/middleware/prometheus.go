/*
 * @Author: flwfdd
 * @Date: 2025-03-18 00:54:50
 * @LastEditTime: 2025-03-18 14:35:56
 * @Description: _(:з」∠)_
 */
package middleware

import (
	"BIT101-GO/pkg/cache"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

var (
	// 请求总数计数器
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bit101_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
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
			Name:    "bit101_http_request_duration_seconds",
			Help:    "Request processing time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// 请求大小直方图
	requestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bit101_http_request_size_bytes",
			Help:    "Size of HTTP requests in bytes",
			Buckets: prometheus.ExponentialBuckets(1, 10, 8),
		},
		[]string{"method", "path"},
	)

	// 响应大小直方图
	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bit101_http_response_size_bytes",
			Help:    "Size of HTTP responses in bytes",
			Buckets: prometheus.ExponentialBuckets(1, 10, 8),
		},
		[]string{"method", "path"},
	)

	// 活跃用户数直方图
	activeUsers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bit101_active_users",
			Help: "Current number of active users",
		},
		[]string{"window"},
	)
	timeWindows = map[string]time.Duration{
		"10m": 10 * time.Minute,
		"1h":  time.Hour,
		"12h": 12 * time.Hour,
		"1d":  24 * time.Hour,
		"7d":  7 * 24 * time.Hour,
		"30d": 30 * 24 * time.Hour,
		"90d": 90 * 24 * time.Hour,
		"1y":  365 * 24 * time.Hour,
	}
)

// updateActiveUsers 更新活跃用户数
func updateActiveUsers(c *gin.Context) {
	// 获取用户ID
	userCtx, err := GetUserContext(c)
	if err != nil {
		return
	}
	uid := userCtx.UIDUint
	rdb := cache.RDB()
	go func() {
		// 更新活跃用户数到Redis
		for window := range timeWindows {
			key := fmt.Sprintf("active_users:%s", window)
			rdb.ZAdd(cache.Context, key, redis.Z{Score: float64(time.Now().Unix()), Member: uid})
		}
	}()
}

// init 初始化
func init() {
	go func() {
		// 定时更新活跃用户数到Prometheus
		for {
			rdb := cache.RDB()
			for window, duration := range timeWindows {
				key := fmt.Sprintf("active_users:%s", window)
				rdb.ZRemRangeByScore(cache.Context, key, "-inf", fmt.Sprintf("%d", time.Now().Add(-duration).Unix()))
				count, _ := rdb.ZCard(cache.Context, key).Result()
				activeUsers.WithLabelValues(window).Set(float64(count))
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

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
			requestSize.WithLabelValues(method, path).Observe(float64(size))
		}

		// 处理请求
		c.Next()

		// 减少活跃请求计数
		httpActiveRequests.Dec()

		// 记录请求状态
		status := fmt.Sprintf("%d", c.Writer.Status())
		requestsTotal.WithLabelValues(method, path, status).Inc()

		// 记录响应时间
		duration := time.Since(start).Seconds()
		requestDuration.WithLabelValues(method, path).Observe(duration)

		// 记录响应大小
		if size := c.Writer.Size(); size > 0 {
			responseSize.WithLabelValues(method, path).Observe(float64(size))
		}

		// 更新活跃用户数
		updateActiveUsers(c)
	}
}

// PrometheusHandler 返回Prometheus指标处理函数
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
