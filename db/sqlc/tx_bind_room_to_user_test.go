package db

import (
	"context"
	"testing"

	"github.com/Jingqi0327/eleclog/util"
	"github.com/stretchr/testify/require"
)

func TestBindRoomToUserTx(t *testing.T) {
	// Case 1: Room does not exist, user room does not exist
	t.Run("RoomAndUserRoomNotExist", func(t *testing.T) {
		user := createRandomUser(t)
		arg := BindRoomToUserTxParams{
			Username:     user.Username,
			Name:         util.RandomString(6),
			AreaID:       util.RandomString(6),
			BuildingCode: util.RandomString(6),
			FloorCode:    util.RandomString(6),
			RoomCode:     util.RandomString(6),
			Threshold:    10,
		}

		result, err := testStore.BindRoomToUserTx(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, result.Room)
		require.NotEmpty(t, result.UserRoom)

		require.Equal(t, arg.Name, result.Room.Name)
		require.Equal(t, arg.AreaID, result.Room.AreaID)
		require.Equal(t, arg.BuildingCode, result.Room.BuildingCode)
		require.Equal(t, arg.FloorCode, result.Room.FloorCode)
		require.Equal(t, arg.RoomCode, result.Room.RoomCode)

		require.Equal(t, arg.Username, result.UserRoom.Username)
		require.Equal(t, result.Room.ID, result.UserRoom.RoomID)
		require.Equal(t, arg.Threshold, result.UserRoom.Threshold)
		require.True(t, result.UserRoom.IsEnabled)
	})

	// Case 2: Room exists, user room does not exist
	t.Run("RoomExistsUserRoomNotExists", func(t *testing.T) {
		user := createRandomUser(t)
		room := createRandomRoom(t)

		arg := BindRoomToUserTxParams{
			Username:     user.Username,
			Name:         util.RandomString(6), // Different name to simulate name mismatch
			AreaID:       room.AreaID,
			BuildingCode: room.BuildingCode,
			FloorCode:    room.FloorCode,
			RoomCode:     room.RoomCode,
			Threshold:    20,
		}

		result, err := testStore.BindRoomToUserTx(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, result.Room)
		require.NotEmpty(t, result.UserRoom)

		// It should fetch the existing room and update its name
		require.Equal(t, room.ID, result.Room.ID)
		require.Equal(t, arg.Name, result.Room.Name)

		// The user room should be newly created
		require.Equal(t, arg.Username, result.UserRoom.Username)
		require.Equal(t, room.ID, result.UserRoom.RoomID)
		require.Equal(t, arg.Threshold, result.UserRoom.Threshold)
		require.True(t, result.UserRoom.IsEnabled)
	})

	// Case 3: Room exists, user room already exists
	t.Run("RoomExistsUserRoomExists", func(t *testing.T) {
		userRoom := createRandomUserRoom(t)
		room, err := testStore.GetRoom(context.Background(), userRoom.RoomID)
		require.NoError(t, err)

		arg := BindRoomToUserTxParams{
			Username:     userRoom.Username,
			Name:         util.RandomString(6), // Different name
			AreaID:       room.AreaID,
			BuildingCode: room.BuildingCode,
			FloorCode:    room.FloorCode,
			RoomCode:     room.RoomCode,
			Threshold:    30, // Trying to bind with a new threshold
		}

		result, err := testStore.BindRoomToUserTx(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, result.Room)
		require.NotEmpty(t, result.UserRoom)

		// It should fetch the existing room and update its name
		require.Equal(t, room.ID, result.Room.ID)
		require.Equal(t, arg.Name, result.Room.Name)

		// It should fetch the existing user room, keeping its original threshold
		require.Equal(t, userRoom.Username, result.UserRoom.Username)
		require.Equal(t, userRoom.RoomID, result.UserRoom.RoomID)
		require.Equal(t, userRoom.Threshold, result.UserRoom.Threshold)
		require.Equal(t, userRoom.IsEnabled, result.UserRoom.IsEnabled)
	})

	// Case 4: Transaction Rollback
	t.Run("RollbackOnUserRoomFailure", func(t *testing.T) {
		// We use a completely new room code so CreateRoom will be attempted and succeed.
		// BUT we use a non-existent username, so CreateUserRoom will fail with a foreign key violation.
		// This should cause the transaction to rollback, and the room should NOT be in the database.
		roomCode := util.RandomString(6)
		arg := BindRoomToUserTxParams{
			Username:     "non_existent_user_12345",
			Name:         util.RandomString(6),
			AreaID:       util.RandomString(6),
			BuildingCode: util.RandomString(6),
			FloorCode:    util.RandomString(6),
			RoomCode:     roomCode,
			Threshold:    10,
		}

		_, err := testStore.BindRoomToUserTx(context.Background(), arg)
		require.Error(t, err)

		// Verify that the room was NOT created (transaction rolled back)
		_, err = testStore.GetRoomByCodes(context.Background(), GetRoomByCodesParams{
			AreaID:       arg.AreaID,
			BuildingCode: arg.BuildingCode,
			FloorCode:    arg.FloorCode,
			RoomCode:     arg.RoomCode,
		})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrRecordNotFound)
	})
}
