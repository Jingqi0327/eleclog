package api

import (
	"fmt"
	"net/http"

	"github.com/Jingqi0327/eleclog/cache"
	"github.com/Jingqi0327/eleclog/logger"
	"github.com/Jingqi0327/eleclog/token"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// rateLimitMiddleware 创建一个基于用户的限流中间件
func rateLimitMiddleware(limiter cache.RateLimiter, capacity int, rate float64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if limiter == nil {
			ctx.Next()
			return
		}

		// 从上下文中获取授权负载信息 (必须放在 authMiddleware 之后使用)
		payloadVal, exists := ctx.Get(authorizationPayloadKey)
		if !exists {
			// 如果没有登录信息，可以选择按 IP 限流或者直接放行
			// 这里我们降级为按 IP 限流
			clientIP := ctx.ClientIP()
			key := fmt.Sprintf("rate_limit:ip:%s", clientIP)
			checkRateLimit(ctx, limiter, key, capacity, rate)
			return
		}

		payload, ok := payloadVal.(*token.Payload)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("invalid payload type")))
			return
		}

		// 按登录用户名(UserID)进行限流
		key := fmt.Sprintf("rate_limit:user:%s", payload.Username)
		checkRateLimit(ctx, limiter, key, capacity, rate)
	}
}

func checkRateLimit(ctx *gin.Context, limiter cache.RateLimiter, key string, capacity int, rate float64) {
	allowed, err := limiter.Allow(ctx, key, capacity, rate)
	if err != nil {
		// 记录错误但放行，避免 Redis 故障导致全站不可用。
		logger.Log.Error("[Middleware] Rate limiter error", zap.Error(err), zap.String("key", key))
		ctx.Next()
		return
	}

	if !allowed {
		ctx.AbortWithStatusJSON(http.StatusTooManyRequests, errorResponse(fmt.Errorf("too many requests")))
		return
	}

	ctx.Next()
}
