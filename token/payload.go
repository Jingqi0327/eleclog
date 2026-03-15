package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = fmt.Errorf("token is invalid")
	ErrExpiredToken = fmt.Errorf("token has expired")
)

type Payload struct {
	Username string    `json:"username"`
	ID       uuid.UUID `json:"id"`
	// 直接嵌入，不再定义冗余的 ID, IssuedAt, ExpiredAt
	jwt.RegisteredClaims
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	payload := &Payload{
		Username: username,
		ID:       tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			// v5 要求使用 NumericDate 类型
			Subject:   username,                              // 对应 JWT 的 sub (可选)
			IssuedAt:  jwt.NewNumericDate(now),               // 对应 JWT 的 iat
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)), // 对应 JWT 的 exp
			NotBefore: jwt.NewNumericDate(now),               // 对应 JWT 的 nbf
		},
	}

	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.RegisteredClaims.ExpiresAt.Time) {
		return ErrExpiredToken
	}
	return nil
}
