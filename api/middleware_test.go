package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jingqi0327/eleclog/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// 这个函数用于在测试中为HTTP请求添加认证信息
func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{ // 正常情况，提供了正确的认证信息，应该返回200 OK
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{ // 没有提供认证信息，应该返回401 Unauthorized
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// 不添加认证信息
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{ // 提供了一个不支持的认证类型，应该返回401 Unauthorized
			name: "UnsupportedAuthorizationType",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// 添加一个不支持的认证类型，例如 "unsupported"
				addAuthorization(t, request, tokenMaker, "unsupported", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{ // 提供了一个格式不正确的认证信息，例如缺少认证类型，应该返回401 Unauthorized
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//不添加认证类型，只给令牌，导致认证信息格式不正确
				addAuthorization(t, request, tokenMaker, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{ // 提供了一个过期的令牌，应该返回401 Unauthorized
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// 创建一个过期的令牌，过期时间设置为负数
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// 创建一个新的测试服务器实例,我们不需要store，所以传入nil
			server := newTestServer(t, nil)

			// 在测试服务器的路由中注册一个需要认证的测试路由
			authPath := "/auth"
			server.router.GET(authPath, authMiddleware(server), func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{})
			})

			// 创建一个新的HTTP响应记录器，用于捕获测试路由的响应结果
			recorder := httptest.NewRecorder()
			// 创建一个新的HTTP请求，方法为GET，路径为/auth，且没有请求体
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			// 为请求添加认证信息
			tc.setupAuth(t, request, server.tokenMaker)
			// 让测试服务器处理这个请求，并将响应结果记录在recorder中
			server.router.ServeHTTP(recorder, request)
			// 检查响应结果是否符合预期
			tc.checkResponse(t, recorder)
		})
	}
}
