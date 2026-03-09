package api

import (
	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store  db.Store
	router *gin.Engine
	config util.Config
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	server := &Server{
		store:  store,
		config: config,
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	// 定义API路由
	router.POST("/rooms", server.createRoom)
	router.GET("/rooms/:id", server.getRoom)
	router.GET("/rooms", server.listRooms)
	router.PUT("/rooms/:id", server.updateRoom)
	router.DELETE("/rooms/:id", server.deleteRoom)

	router.GET("/electricity-balances/latest/:room_id", server.getLatestElectricityBalance)
	router.GET("/electricity-balances/hour-range/:room_id", server.getElectricityBalanceByHourRange)

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
