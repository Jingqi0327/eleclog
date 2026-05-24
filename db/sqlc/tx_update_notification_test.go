package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpdateRoomLastNotifiedAtTx(t *testing.T) {
	testCases := []struct {
		name        string
		AfterUpdate func(t *testing.T, n UserRoom, original UserRoom, arg UpdateUserRoomLastNotifiedAtParams) error
		check       func(t *testing.T, err error, original UserRoom, arg UpdateUserRoomLastNotifiedAtParams)
	}{
		{
			name: "Happy Case",
			AfterUpdate: func(t *testing.T, n UserRoom, original UserRoom, arg UpdateUserRoomLastNotifiedAtParams) error {
				require.Equal(t, arg.Username, n.Username)
				require.Equal(t, arg.RoomID, n.RoomID)
				require.WithinDuration(t, time.Now(), n.LastNotifiedAt, time.Second)
				return nil
			},
			check: func(t *testing.T, err error, original UserRoom, arg UpdateUserRoomLastNotifiedAtParams) {
				require.NoError(t, err)
				n, err := testStore.GetUserRoom(context.Background(), GetUserRoomParams{Username: arg.Username, RoomID: arg.RoomID})
				require.NoError(t, err)
				require.NotEqual(t, original.LastNotifiedAt, n.LastNotifiedAt)
				require.Equal(t, arg.Username, n.Username)
				require.Equal(t, arg.RoomID, n.RoomID)
				require.WithinDuration(t, time.Now(), n.LastNotifiedAt, time.Second)
			},
		},
		{
			name: "Rollback Case",
			AfterUpdate: func(t *testing.T, n UserRoom, original UserRoom, arg UpdateUserRoomLastNotifiedAtParams) error {
				return fmt.Errorf("Mock after update error")
			},
			check: func(t *testing.T, err error, original UserRoom, arg UpdateUserRoomLastNotifiedAtParams) {
				require.Error(t, err)
				require.EqualError(t, err, "Mock after update error")

				n, err := testStore.GetUserRoom(context.Background(), GetUserRoomParams{Username: arg.Username, RoomID: arg.RoomID})
				require.NoError(t, err)
				require.Equal(t, original.LastNotifiedAt, n.LastNotifiedAt) // 应该等于最初的那个时间，因为事务回滚了
				require.Equal(t, arg.Username, n.Username)
				require.Equal(t, arg.RoomID, n.RoomID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 在闭包内部创建测试记录，保证每个测试用例隔离，互不影响
			original := createRandomUserRoom(t)
			arg := UpdateUserRoomLastNotifiedAtParams{
				Username: original.Username,
				RoomID:   original.RoomID,
			}

			req := UpdateUserRoomLastNotifiedAtTxParams{
				UpdateUserRoomLastNotifiedAtParams: arg,
				AfterUpdate: func(n UserRoom) error {
					return tc.AfterUpdate(t, n, original, arg)
				},
			}
			_, err := testStore.UpdateRoomLastNotifiedAtTx(context.Background(), req)
			tc.check(t, err, original, arg)
		})
	}
}
