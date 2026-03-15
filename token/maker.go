package token

import "time"

type Maker interface {
	// CreateToken 生成一个新的令牌，包含username和duration，令牌有效时长
	CreateToken(username string, duration time.Duration) (string, *Payload, error)

	// VerifyToken 验证令牌是否有效，返回载荷信息
	VerifyToken(token string) (*Payload, error)
}
