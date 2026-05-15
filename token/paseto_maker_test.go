package token

import (
	"testing"
	"time"

	"github.com/Jingqi0327/eleclog/util"
	"github.com/stretchr/testify/require"
)

// TestPasetoMaker 测试 PasetoMaker 的 CreateToken 和 VerifyToken 方法，确保它们能够正确地生成和验证 PASETO 令牌。
func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomName(6)
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)


	token, payload, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.RegisteredClaims.IssuedAt.Time, time.Second)
	require.WithinDuration(t, expiredAt, payload.RegisteredClaims.ExpiresAt.Time, time.Second)
}

// 测试过期的 PASETO 令牌是否被正确识别为无效
func TestExpiredPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	token, payload, err := maker.CreateToken(util.RandomName(6), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

// 测试无效的 PASETO 令牌是否被正确识别为无效
func TestInvalidPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	payload, err := maker.VerifyToken("invalid_token")
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}
