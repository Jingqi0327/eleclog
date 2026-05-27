package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/Jingqi0327/eleclog/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	// 提取元数据
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	// 获取授权头
	value := md.Get(authorizationHeader)
	if len(value) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	// 解析授权头
	authHeader := value[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	// 验证授权类型
	authType := fields[0]
	authType = strings.ToLower(authType)
	if authType != authorizationBearer {
		return nil, fmt.Errorf("unsupported authorization type %s", authType)
	}

	// 验证访问令牌
	accessToken := fields[1]
	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization token")
	}
	return payload, nil
}
