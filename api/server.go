package api

import (
	"fmt"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	token "github.com/Jingqi0327/eleclog/token"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store      db.Store
	router     *gin.Engine
	config     util.Config
	tokenMaker token.Maker
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
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
	router := gin.Default()

	// 启用CORS支持
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	authRoutes := router.Group("/").Use(authMiddleware(server))
	authRoutes.POST("/rooms", server.createRoom)
	authRoutes.DELETE("/rooms/:id", server.deleteRoom)
	authRoutes.POST("/users", server.createUser)
	authRoutes.PUT("/users", server.UpdateUser)
	authRoutes.POST("/notifications", server.createUserRoomNotification)
	authRoutes.GET("/notifications", server.listUserRoomNotifications)
	authRoutes.GET("/notifications/:room_id", server.getUserRoomNotification)
	authRoutes.PUT("/notifications/:room_id", server.updateUserRoomNotification)
	authRoutes.DELETE("/notifications/:room_id", server.deleteUserRoomNotification)
	// 代理路由：转发到 xiaofubao 外部 API
	authRoutes.GET("/proxy/areas", server.proxyQueryArea)
	authRoutes.GET("/proxy/buildings", server.proxyQueryBuilding)
	authRoutes.GET("/proxy/floors", server.proxyQueryFloor)
	authRoutes.GET("/proxy/rooms", server.proxyQueryRoom)
	authRoutes.GET("/proxy/room-surplus", server.proxyQueryRoomSurplus)

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
	return server.router.Run(address)
}

// 统一的错误响应格式
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
