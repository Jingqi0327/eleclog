package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 定义密钥的最小长度，确保安全性
const minSecretKeySize = 32

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string // secretKey 是用来加密和解密 JWT 的密钥
	// 这边在本地所以用对称加密
}

// NewJWTMaker 创建一个新的 JWTMaker 实例，使用提供的 secretKey 进行加密和解密操作
func NewJWTMaker(secretKey string) (Maker, error) {
	// 密钥不满足最小长度要求时，返回错误
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}

// CreateToken 生成一个新的令牌，包含username和duration，令牌有效时长
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	//先创建一个payload
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}

	// 创建一个未签名的 JWT 令牌，使用 HS256 签名方法，并将 payload 作为其声明
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// 使用 secretKey 对 JWT 令牌进行签名，并返回完整的 JWT 字符串
	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	return token, payload, err
}

// VerifyToken 验证令牌是否有效，返回载荷信息
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// 定义一个函数，用于提供密钥和验证签名方法
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法是否正确
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	// 解析 JWT 令牌，并验证其有效性
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		// 如果解析过程中发生错误，检查是否是因为令牌过期导致的
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		// 其他解析错误被视为无效令牌
		return nil, ErrInvalidToken
	}

	// 从解析后的 JWT 令牌中提取 payload，并进行类型断言
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
