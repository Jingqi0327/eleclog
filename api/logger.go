package api

import (
	"fmt"
	"net"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Jingqi0327/eleclog/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinLogger 动态适应环境的请求日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 执行后续业务逻辑
		c.Next()

		// 1. 实例化 Gin 的日志参数结构体，收集当前请求的所有上下文
		param := gin.LogFormatterParams{
			Request:      c.Request,
			ClientIP:     c.ClientIP(),
			Method:       c.Request.Method,
			StatusCode:   c.Writer.Status(),
			Latency:      time.Since(start),
			Path:         path,
			ErrorMessage: c.Errors.ByType(gin.ErrorTypePrivate).String(),
		}

		if query != "" {
			param.Path = path + "?" + query
		}

		// 2. 环境分流处理
		if logger.IsDev {
			// ==========================================
			// 开发环境：白嫖 Gin 的原生颜色方法进行排版
			// ==========================================
			// 使用 param.StatusCodeColor() 和 param.MethodColor() 注入 ANSI 颜色码
			msg := fmt.Sprintf("|%s %3d %s| %13v | %15s |%s %-7s %s %#v",
				param.StatusCodeColor(), param.StatusCode, param.ResetColor(),
				param.Latency,
				param.ClientIP,
				param.MethodColor(), param.Method, param.ResetColor(),
				param.Path,
			)

			if len(param.ErrorMessage) > 0 {
				msg += " | errors: " + param.ErrorMessage
			}

			// 交给 Zap 输出。终端上你会看到和原生 Gin 完全一样的彩色高亮！
			if param.StatusCode >= 500 {
				logger.Log.Error(msg)
			} else if param.StatusCode >= 400 {
				logger.Log.Warn(msg)
			} else {
				logger.Log.Info(msg)
			}

		} else {
			// ==========================================
			// 生产环境：纯净结构化 JSON，为后续持久化到文件做准备
			// ==========================================
			logger.Log.Info("HTTP Request",
				zap.Int("status", param.StatusCode),
				zap.String("method", param.Method),
				zap.String("path", param.Path),
				zap.String("ip", param.ClientIP),
				zap.String("errors", param.ErrorMessage),
				zap.Duration("cost", param.Latency),
			)
		}
	}
}

// GinRecovery 替换 Gin 默认的 Recovery，使用 zap 记录 Panic 堆栈
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查是否是客户端断开连接导致的 Panic (例如 Broken pipe)
				// 这种通常不需要记录完整的堆栈
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Log.Error("Client disconnected",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// 由于连接已断开，不需要返回状态码了
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				// 如果是服务器内部代码错误导致的 Panic，记录详细堆栈
				if stack {
					logger.Log.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Log.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}

				// 向客户端返回 500 状态码
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
