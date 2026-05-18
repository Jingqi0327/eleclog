package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	token "github.com/Jingqi0327/eleclog/token"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store      db.Store
	router     *gin.Engine
	config     util.Config
	tokenMaker token.Maker
	srv        *http.Server
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		store:      store,
		config:     config,
		tokenMaker: tokenMaker,
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	// router := gin.Default()
	router := gin.New()

	// 启用CORS支持
	router.Use(cors.New(cors.Config{
		// 允许访问的域名列表（替换为你前端真实的生产域名）
		AllowOrigins: []string{"http://localhost:3001"},
		// 允许的请求方法
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		// 允许携带的自定义请求头
		AllowHeaders: []string{"Origin", "Authorization", "Content-Type", "X-Requested-With"},
		// 是否允许客户端浏览器携带 Cookie（如果用了 JWT 存放在 Header，一般不需要开启）
		AllowCredentials: false,
		// 预检请求（OPTIONS）的缓存时间。在 12 小时内，浏览器不需要再重复发送 OPTIONS 请求，大幅提升性能
		MaxAge: 12 * time.Hour,
	}))

	router.Use(GinLogger(), GinRecovery(true))

	authRoutes := router.Group("/").Use(authMiddleware(server))
	authRoutes.POST("/rooms", server.createRoom)
	authRoutes.DELETE("/rooms/:id", server.deleteRoom)
	authRoutes.POST("/users", server.createUser)
	authRoutes.PATCH("/users", server.UpdateUser)
	authRoutes.POST("/notifications", server.createUserRoomNotification)
	authRoutes.GET("/notifications", server.listUserRoomNotifications)
	authRoutes.GET("/notifications/:room_id", server.getUserRoomNotification)
	authRoutes.PATCH("/notifications/:room_id", server.updateUserRoomNotification)
	authRoutes.DELETE("/notifications/:room_id", server.deleteUserRoomNotification)
	// 代理路由：转发到 xiaofubao 外部 API
	authRoutes.GET("/proxy/areas", server.proxyQueryArea)
	authRoutes.GET("/proxy/buildings", server.proxyQueryBuilding)
	authRoutes.GET("/proxy/floors", server.proxyQueryFloor)
	authRoutes.GET("/proxy/rooms", server.proxyQueryRoom)
	authRoutes.GET("/proxy/room-surplus", server.proxyQueryRoomSurplus)
	authRoutes.POST("/electricity-balances/import/:room_id", server.importElectricityRecords)

	router.GET("/rooms/:id", server.getRoom)
	router.GET("/rooms", server.listRooms)
	router.PUT("/rooms/:id", server.updateRoom)

	router.GET("/electricity-balances/latest/:room_id", server.getLatestElectricityBalance)
	router.GET("/electricity-balances/hour-range/:room_id", server.getElectricityRecordByHourRange)

	router.POST("/users/login", server.loginUser)

	server.router = router
}

// 启动服务器
func (server *Server) Start(address string) error {
	server.srv = &http.Server{
		Addr:    address,
		Handler: server.router,
	}
	return server.srv.ListenAndServe()
}

func (server *Server) Shutdown(ctx context.Context) error {
	return server.srv.Shutdown(ctx)
}

// 统一的错误响应格式
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
