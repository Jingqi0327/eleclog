package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload" // 为提取出的令牌载荷信息定义一个常量键，以便在 Gin 的上下文中存储和访问
)

// 这个函数是一个高阶函数，返回一个 Gin 的中间件函数
func authMiddleware(server *Server) gin.HandlerFunc {
	// 这个返回的匿名函数才是真正的中间件函数
	return func(ctx *gin.Context) {
		// 从请求头中获取授权信息
		authorizationHeaderKey := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeaderKey) == 0 { // 如果授权信息为空，说明客户端没有提供授权信息
			err := fmt.Errorf("authorization header is not provided")
			// AbortWithStatusJSON 会停止当前请求的处理，并返回一个 JSON 格式的错误响应
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// 授权信息中包含了授权类型和令牌，我们需要将它们分开
		// 例如，授权信息可能是 "Bearer <token>"，我们需要将 "Bearer" 和 "<token>" 分开
		// strings.Fields 函数会将字符串按照空格分割成一个字符串切片
		fields := strings.Fields(authorizationHeaderKey)
		if len(fields) < 2 { // 如果分割后的字符串切片长度小于 2，说明授权信息的格式不正确
			err := fmt.Errorf("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// 提取授权类型，并将其转换为小写字母，以便进行比较
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// 提取令牌，并验证其有效性
		accessToken := fields[1]
		payload, err := server.tokenMaker.VerifyToken(accessToken)
		if err != nil {
			err := fmt.Errorf("invalid token: %w", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// 将验证通过的令牌的载荷信息存储在 Gin 的上下文中，以便后续处理函数使用
		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next() // 调用下一个处理函数，继续处理请求
	}
}
