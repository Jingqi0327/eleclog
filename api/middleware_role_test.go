package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jingqi0327/eleclog/token"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRoleMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		allowedRoles  []string
		setupCtx      func(ctx *gin.Context)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			allowedRoles: []string{
				util.AdminRole,
			},
			setupCtx: func(ctx *gin.Context) {
				ctx.Set(authorizationPayloadKey, &token.Payload{
					Role: util.AdminRole,
				})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "MultipleRolesOK",
			allowedRoles: []string{
				util.AdminRole,
				util.ManagerRole,
			},
			setupCtx: func(ctx *gin.Context) {
				ctx.Set(authorizationPayloadKey, &token.Payload{
					Role: util.ManagerRole,
				})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:         "EmptyRoles",
			allowedRoles: []string{},
			setupCtx: func(ctx *gin.Context) {
				ctx.Set(authorizationPayloadKey, &token.Payload{
					Role: util.UserRole,
				})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "Forbidden",
			allowedRoles: []string{
				util.AdminRole,
			},
			setupCtx: func(ctx *gin.Context) {
				ctx.Set(authorizationPayloadKey, &token.Payload{
					Role: util.UserRole,
				})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:         "MissingTokenPayload",
			allowedRoles: []string{util.AdminRole},
			setupCtx: func(ctx *gin.Context) {
				// 不进行任何设置
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:         "InvalidTokenPayloadType",
			allowedRoles: []string{util.AdminRole},
			setupCtx: func(ctx *gin.Context) {
				// 设置错误的类型以触发断言失败
				ctx.Set(authorizationPayloadKey, "invalid_type")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			// 在集成/单元测试中，我们直接构造一个新的 router 来测试特定的中间件
			gin.SetMode(gin.TestMode)
			router := gin.New()

			rolePath := "/role"
			router.GET(
				rolePath,
				func(ctx *gin.Context) {
					tc.setupCtx(ctx)
				},
				roleMiddleware(tc.allowedRoles...),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, rolePath, nil)
			require.NoError(t, err)

			router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
