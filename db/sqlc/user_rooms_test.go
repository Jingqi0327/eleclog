package db

import (
	"context"

	"testing"
	"time"

	"github.com/Jingqi0327/eleclog/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomUserRoom(t *testing.T) UserRoom {
	user := createRandomUser(t)
	room := createRandomRoom(t)

	arg := CreateUserRoomParams{
		Username:  user.Username,
		RoomID:    room.ID,
		Threshold: int32(util.RandomInt(10, 100)),
	}

	notification, err := testStore.CreateUserRoom(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notification)

	require.Equal(t, arg.Username, notification.Username)
	require.Equal(t, arg.RoomID, notification.RoomID)
	require.Equal(t, arg.Threshold, notification.Threshold)
	require.False(t, notification.IsEnabled)

	return notification
}

func TestCreateUserRoom(t *testing.T) {
	createRandomUserRoom(t)
}

func TestGetUserRoom(t *testing.T) {
	notif1 := createRandomUserRoom(t)

	arg := GetUserRoomParams{
		Username: notif1.Username,
		RoomID:   notif1.RoomID,
	}

	notif2, err := testStore.GetUserRoom(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notif2)

	require.Equal(t, notif1.Username, notif2.Username)
	require.Equal(t, notif1.RoomID, notif2.RoomID)
	require.Equal(t, notif1.Threshold, notif2.Threshold)
	require.Equal(t, notif1.IsEnabled, notif2.IsEnabled)
	require.WithinDuration(t, notif1.LastNotifiedAt, notif2.LastNotifiedAt, time.Second)
}

func TestDeleteUserRoom(t *testing.T) {
	notif1 := createRandomUserRoom(t)

	argDelete := DeleteUserRoomParams{
		Username: notif1.Username,
		RoomID:   notif1.RoomID,
	}
	err := testStore.DeleteUserRoom(context.Background(), argDelete)
	require.NoError(t, err)

	argGet := GetUserRoomParams{
		Username: notif1.Username,
		RoomID:   notif1.RoomID,
	}
	notif2, err := testStore.GetUserRoom(context.Background(), argGet)
	require.Error(t, err)
	require.EqualError(t, err, ErrRecordNotFound.Error())
	require.Empty(t, notif2)
}

func TestUpdateUserRoom(t *testing.T) {
	notif1 := createRandomUserRoom(t)

	newThreshold := int32(util.RandomInt(10, 100))
	newIsEnabled := false

	arg := UpdateUserRoomParams{
		Username:  notif1.Username,
		RoomID:    notif1.RoomID,
		Threshold: pgtype.Int4{Int32: newThreshold, Valid: true},
		IsEnabled: pgtype.Bool{Bool: newIsEnabled, Valid: true},
	}

	notif2, err := testStore.UpdateUserRoom(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notif2)

	require.Equal(t, notif1.Username, notif2.Username)
	require.Equal(t, notif1.RoomID, notif2.RoomID)
	require.Equal(t, newThreshold, notif2.Threshold)
	require.Equal(t, newIsEnabled, notif2.IsEnabled)
	require.WithinDuration(t, notif1.LastNotifiedAt, notif2.LastNotifiedAt, time.Second)
}

func TestUpdateUserRoomLastNotifiedAt(t *testing.T) {
	notif1 := createRandomUserRoom(t)

	newTime := time.Now().Add(-24 * time.Hour)

	arg := UpdateUserRoomLastNotifiedAtParams{
		Username:       notif1.Username,
		RoomID:         notif1.RoomID,
		LastNotifiedAt: newTime,
	}

	notif2, err := testStore.UpdateUserRoomLastNotifiedAt(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notif2)

	require.Equal(t, notif1.Username, notif2.Username)
	require.Equal(t, notif1.RoomID, notif2.RoomID)
	require.Equal(t, notif1.Threshold, notif2.Threshold)
	require.Equal(t, notif1.IsEnabled, notif2.IsEnabled)
	require.WithinDuration(t, newTime, notif2.LastNotifiedAt, time.Second)
}

func TestListUserRooms(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomUserRoom(t)
	}

	arg := ListUserRoomsParams{
		Limit:  5,
		Offset: 0,
	}

	notifs, err := testStore.ListUserRooms(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notifs)
	require.LessOrEqual(t, len(notifs), 5)
}

func TestListUserRoomsByUser(t *testing.T) {
	user := createRandomUser(t)

	for i := 0; i < 10; i++ {
		room := createRandomRoom(t)
		arg := CreateUserRoomParams{
			Username:  user.Username,
			RoomID:    room.ID,
			Threshold: int32(util.RandomInt(10, 100)),
		}
		_, err := testStore.CreateUserRoom(context.Background(), arg)
		require.NoError(t, err)
	}

	arg := ListUserRoomsByUserParams{
		Username: user.Username,
		Limit:    5,
		Offset:   0,
	}

	notifs, err := testStore.ListUserRoomsByUser(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, notifs, 5)

	for _, notif := range notifs {
		require.NotEmpty(t, notif)
		require.Equal(t, user.Username, notif.Username)
	}
}

func TestListUserRoomsByRoom(t *testing.T) {
	room := createRandomRoom(t)

	for i := 0; i < 5; i++ {
		user := createRandomUser(t)
		arg := CreateUserRoomParams{
			Username:  user.Username,
			RoomID:    room.ID,
			Threshold: int32(util.RandomInt(10, 100)),
		}
		_, err := testStore.CreateUserRoom(context.Background(), arg)
		require.NoError(t, err)
	}

	notifs, err := testStore.ListUserRoomsByRoom(context.Background(), room.ID)
	require.NoError(t, err)
	require.Len(t, notifs, 5)

	for _, notif := range notifs {
		require.NotEmpty(t, notif)
		require.Equal(t, room.ID, notif.RoomID)
	}
}

func TestListDueUserRooms(t *testing.T) {
	notif := createRandomUserRoom(t)

	newTime := time.Now().Add(-25 * time.Hour)
	argUpdate := UpdateUserRoomLastNotifiedAtParams{
		Username:       notif.Username,
		RoomID:         notif.RoomID,
		LastNotifiedAt: newTime,
	}
	_, err := testStore.UpdateUserRoomLastNotifiedAt(context.Background(), argUpdate)
	require.NoError(t, err)

	argEnable := UpdateUserRoomParams{
		Username:  notif.Username,
		RoomID:    notif.RoomID,
		IsEnabled: pgtype.Bool{Bool: true, Valid: true},
	}
	_, err = testStore.UpdateUserRoom(context.Background(), argEnable)
	require.NoError(t, err)

	notifs, err := testStore.ListDueUserRooms(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, notifs)

	found := false
	for _, n := range notifs {
		if n.Username == notif.Username && n.RoomID == notif.RoomID {
			found = true
			require.Equal(t, notif.Threshold, n.Threshold)
			break
		}
	}
	require.True(t, found)
}

func TestCountUserRooms(t *testing.T) {
	beforeCount, err := testStore.CountUserRooms(context.Background())
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		createRandomUserRoom(t)
	}

	afterCount, err := testStore.CountUserRooms(context.Background())
	require.NoError(t, err)
	require.Equal(t, beforeCount+5, afterCount)
}


func TestCountRoomsByUser(t *testing.T) {
	user:=createRandomUser(t)
	for i:=0;i<5;i++{
		room:=createRandomRoom(t)
		arg:=CreateUserRoomParams{
			Username: user.Username,
			RoomID: room.ID,
			Threshold: int32(util.RandomInt(10,100)),
		}
		_,err:=testStore.CreateUserRoom(context.Background(),arg)
		require.NoError(t,err)
	}

	count,err:=testStore.CountRoomsByUser(context.Background(),user.Username)
	require.NoError(t,err)
	require.Equal(t,int64(5),count)
}