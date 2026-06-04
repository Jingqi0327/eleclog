package api

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// 定义请求总数的计数器 (CounterVec)
	// 包含三个标签：method (GET/POST等), path (请求路径), status (HTTP状态码)
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eleclog_http_requests_total",      // 指标在 Prometheus 中显示的名字
			Help: "Total number of HTTP requests.",   // 指标的帮助说明
		},
		[]string{"method", "path", "status"},         // 声明标签名称
	)

	// 定义请求耗时的直方图 (HistogramVec)
	// 同样包含 method 和 path 两个标签 (通常不把 status 放进来以免数据维度过大)
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "eleclog_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}, // 定义耗时的分布桶 (单位:秒)
		},
		[]string{"method", "path"},
	)
)
