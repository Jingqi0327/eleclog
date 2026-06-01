package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Jingqi0327/eleclog/cache"
	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/pb"
	token "github.com/Jingqi0327/eleclog/token"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	store       db.Store
	cache       cache.Cache
	router      *gin.Engine
	config      util.Config
	tokenMaker  token.Maker
	srv         *http.Server
	proxyClient pb.ProxyServiceClient
	rateLimiter cache.RateLimiter
}

func NewServer(config util.Config, store db.Store, c cache.Cache, proxyClient pb.ProxyServiceClient, rateLimiter cache.RateLimiter) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		store:       store,
		cache:       c,
		config:      config,
		tokenMaker:  tokenMaker,
		proxyClient: proxyClient,
		rateLimiter: rateLimiter,
	}

	// 注册验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("role", validRole)
	}

	// 设置路由
	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	// router := gin.Default()
	router := gin.New()

	origins := []string{"http://localhost:3001"}
	if server.config.FrontendOrigin != "" {
		origins = append(origins, server.config.FrontendOrigin)
	}

	// 启用CORS支持
	router.Use(cors.New(cors.Config{
		// 允许访问的域名列表（替换为你前端真实的生产域名）
		AllowOrigins: origins,
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

	authRoutes := router.Group("/")
	authRoutes.Use(
		authMiddleware(server),
		rateLimitMiddleware(server.rateLimiter, server.config.RedisLimiterCapacity, server.config.RedisLimiterRate),
	)

	userRoutes := authRoutes.Group("/")
	userRoutes.Use(roleMiddleware(util.UserRole, util.ManagerRole, util.AdminRole))

	managerRoutes := authRoutes.Group("/")
	managerRoutes.Use(roleMiddleware(util.ManagerRole, util.AdminRole))

	adminRoutes := authRoutes.Group("/")
	adminRoutes.Use(roleMiddleware(util.AdminRole))

	managerRoutes.POST("/rooms", server.createRoom)
	adminRoutes.DELETE("/rooms/:id", server.deleteRoom)
	managerRoutes.PUT("/rooms/:id", server.updateRoom)
	userRoutes.GET("/rooms/:id", server.getRoom)
	userRoutes.GET("/rooms", server.listRooms)

	managerRoutes.POST("/users", server.createUser)
	userRoutes.PATCH("/users", server.UpdateUser)
	userRoutes.POST("/users/rooms/bind", server.bindRoomToUser)
	managerRoutes.GET("/users", server.ListUsers)
	managerRoutes.DELETE("/users/:username", server.deleteUser)
	// 针对登录接口单独限流，更严格的限制 (基于 IP，容量 5，每秒 1 次) 防止暴力破解
	router.POST("/users/login", rateLimitMiddleware(server.rateLimiter, 5, 1), server.loginUser)

	userRoutes.POST("/user-rooms", server.createUserRoom)
	userRoutes.GET("/user-rooms", server.listUserRooms)
	userRoutes.GET("/user-rooms/:room_id", server.getUserRoom)
	userRoutes.PATCH("/user-rooms/:room_id", server.updateUserRoom)
	userRoutes.DELETE("/user-rooms/:room_id", server.deleteUserRoom)
	// 代理路由：转发到 xiaofubao 外部 API
	userRoutes.GET("/proxy/areas", server.proxyQueryArea)
	userRoutes.GET("/proxy/buildings", server.proxyQueryBuilding)
	userRoutes.GET("/proxy/floors", server.proxyQueryFloor)
	userRoutes.GET("/proxy/rooms", server.proxyQueryRoom)
	userRoutes.GET("/proxy/room-surplus", server.proxyQueryRoomSurplus)

	managerRoutes.POST("/electricity-balances/import/:room_id", server.importElectricityRecords)

	userRoutes.GET("/electricity-balances/latest/:room_id", server.getLatestElectricityBalance)
	userRoutes.GET("/electricity-balances/hour-range/:room_id", server.getElectricityRecordByHourRange)

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
