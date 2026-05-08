package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Jingqi0327/eleclog/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomName(6),
		HashedPassword: hashedPassword,
		FullName:       util.RandomName(6),
		Email:          util.RandomName(6) + "@example.com",
	}

	user, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testStore.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestUpdateUser(t *testing.T) {
	user1 := createRandomUser(t)

	newFullName := util.RandomName(6)
	newEmail := util.RandomName(6) + "@example.com"
	newPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := UpdateUserParams{
		Username:       user1.Username,
		HashedPassword: sql.NullString{String: newPassword, Valid: true},
		FullName:       sql.NullString{String: newFullName, Valid: true},
		Email:          sql.NullString{String: newEmail, Valid: true},
	}

	user2, err := testStore.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, newPassword, user2.HashedPassword)
	require.Equal(t, newFullName, user2.FullName)
	require.Equal(t, newEmail, user2.Email)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestCountUsers(t *testing.T) {
	beforeCount, err := testStore.CountUsers(context.Background())
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		createRandomUser(t)
	}

	afterCount, err := testStore.CountUsers(context.Background())
	require.NoError(t, err)
	require.Equal(t, beforeCount+5, afterCount)
}
