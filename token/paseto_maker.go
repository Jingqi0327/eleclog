package token

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// NewPasetoMaker 创建一个新的 PasetoMaker 实例，使用提供的 symmetricKey 进行加密和解密操作
func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) < chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}

// CreateToken 生成一个新的令牌，包含username和duration，令牌有效时长
func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}

	// 使用 symmetricKey 对 payload 进行加密，并返回生成的 PASETO 令牌字符串
	// 参数说明：
	// maker.symmetricKey: 用于加密的对称密钥，必须是 32 字节长。
	// payload: 要加密的数据，这里是一个包含用户名和过期时间的 Payload 结构体。
	// nil: 可选参数，可以用来添加额外的 footer 信息，这里我们不使用，所以传入 nil。
	token, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	return token, payload, err
}

// VerifyToken 验证令牌是否有效，返回载荷信息
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	// 使用 symmetricKey 对 token 进行解密，并将解密后的数据存储在 payload 中
	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
