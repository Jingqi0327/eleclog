package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware 拦截请求并记录指标
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 记录请求开始的时间
		start := time.Now()

		// 2. 交由后续的业务逻辑处理
		c.Next()

		// ==== 下面的代码会在业务逻辑处理完毕后执行 ====

		// 3. 计算耗时（秒）
		duration := time.Since(start).Seconds()

		// 获取我们需要记录的 Label 标签信息
		method := c.Request.Method
		// 重点提醒：这里一定要用 c.FullPath() 而不是 c.Request.URL.Path
		// 防止/users/:id接口， /users/1 和 /users/2 被当成两个不同的接口记录
		path := c.FullPath() 
		
		// 如果路径没有匹配到任何路由（比如 404 请求），FullPath 会为空
		if path == "" {
			path = "unknown"
		}

		status := strconv.Itoa(c.Writer.Status())

		// 4. 更新我们之前定义的指标
		// Counter: 填入标签并调用 Inc() 加 1
		HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
		
		// Histogram: 填入标签并调用 Observe() 记录耗时
		HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
