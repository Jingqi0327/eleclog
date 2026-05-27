package gapi

import (
	"fmt"

	"github.com/Jingqi0327/eleclog/pb"
	"github.com/Jingqi0327/eleclog/token"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/go-resty/resty/v2"
)

// Server serves gRPC requests for proxy service.
type Server struct {
	pb.UnimplementedProxyServiceServer
	config     util.Config
	tokenMaker token.Maker
}

// NewServer creates a new gRPC server.
func NewServer(config util.Config) (*Server, error){
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, err
	}

	server := &Server{
		config:     config,
		tokenMaker: tokenMaker,
	}

	return server, nil
}

// newRestyClient 辅助方法，生成配置了 header 的 resty 客户端
func (server *Server) newRestyClient() *resty.Client {
	return resty.New().
		SetHeader("Cookie", fmt.Sprintf("shiroJID=%s", server.config.ShiroJID)).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36")
}
