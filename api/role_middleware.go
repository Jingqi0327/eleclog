package api

import (
	"errors"
	"net/http"

	token "github.com/Jingqi0327/eleclog/token"
	"github.com/gin-gonic/gin"
)

func roleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. 安全获取 Token 载荷，防止因路由配置失误触发 panic 导致服务崩溃
		val, exists := ctx.Get(authorizationPayloadKey)
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized: missing token payload")))
			return
		}

		payload, ok := val.(*token.Payload)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(errors.New("internal server error: invalid token payload type")))
			return
		}

		// 2. 校验当前用户的角色是否属于允许列表中
		for _, role := range allowedRoles {
			if payload.Role == role {
				ctx.Next() // 显式放行至后续处理器
				return
			}
		}

		// 3. 角色不匹配，返回 403 Forbidden（无权访问）
		ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(errors.New("forbidden: permission denied")))
	}
}
