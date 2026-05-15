package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Jingqi0327/eleclog/util"
	"github.com/stretchr/testify/require"
)

func createRandomUserRoomNotification(t *testing.T) UserRoomNotification {
	user := createRandomUser(t)
	room := createRandomRoom(t)

	arg := CreateUserRoomNotificationParams{
		Username:  user.Username,
		RoomID:    room.ID,
		Threshold: int32(util.RandomInt(10, 100)),
	}

	notification, err := testStore.CreateUserRoomNotification(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notification)

	require.Equal(t, arg.Username, notification.Username)
	require.Equal(t, arg.RoomID, notification.RoomID)
	require.Equal(t, arg.Threshold, notification.Threshold)
	require.True(t, notification.IsEnabled)

	return notification
}

func TestCreateUserRoomNotification(t *testing.T) {
	createRandomUserRoomNotification(t)
}

func TestGetUserRoomNotification(t *testing.T) {
	notif1 := createRandomUserRoomNotification(t)

	arg := GetUserRoomNotificationParams{
		Username: notif1.Username,
		RoomID:   notif1.RoomID,
	}

	notif2, err := testStore.GetUserRoomNotification(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notif2)

	require.Equal(t, notif1.Username, notif2.Username)
	require.Equal(t, notif1.RoomID, notif2.RoomID)
	require.Equal(t, notif1.Threshold, notif2.Threshold)
	require.Equal(t, notif1.IsEnabled, notif2.IsEnabled)
	require.WithinDuration(t, notif1.LastNotifiedAt, notif2.LastNotifiedAt, time.Second)
}

func TestDeleteUserRoomNotification(t *testing.T) {
	notif1 := createRandomUserRoomNotification(t)

	argDelete := DeleteUserRoomNotificationParams{
		Username: notif1.Username,
		RoomID:   notif1.RoomID,
	}
	err := testStore.DeleteUserRoomNotification(context.Background(), argDelete)
	require.NoError(t, err)

	argGet := GetUserRoomNotificationParams{
		Username: notif1.Username,
		RoomID:   notif1.RoomID,
	}
	notif2, err := testStore.GetUserRoomNotification(context.Background(), argGet)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, notif2)
}

func TestUpdateUserRoomNotification(t *testing.T) {
	notif1 := createRandomUserRoomNotification(t)

	newThreshold := int32(util.RandomInt(10, 100))
	newIsEnabled := false

	arg := UpdateUserRoomNotificationParams{
		Username:  notif1.Username,
		RoomID:    notif1.RoomID,
		Threshold: sql.NullInt32{Int32: newThreshold, Valid: true},
		IsEnabled: sql.NullBool{Bool: newIsEnabled, Valid: true},
	}

	notif2, err := testStore.UpdateUserRoomNotification(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notif2)

	require.Equal(t, notif1.Username, notif2.Username)
	require.Equal(t, notif1.RoomID, notif2.RoomID)
	require.Equal(t, newThreshold, notif2.Threshold)
	require.Equal(t, newIsEnabled, notif2.IsEnabled)
	require.WithinDuration(t, notif1.LastNotifiedAt, notif2.LastNotifiedAt, time.Second)
}

func TestUpdateUserRoomNotificationLastNotifiedAt(t *testing.T) {
	notif1 := createRandomUserRoomNotification(t)

	newTime := time.Now().Add(-24 * time.Hour)

	arg := UpdateUserRoomNotificationLastNotifiedAtParams{
		Username:       notif1.Username,
		RoomID:         notif1.RoomID,
		LastNotifiedAt: newTime,
	}

	notif2, err := testStore.UpdateUserRoomNotificationLastNotifiedAt(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notif2)

	require.Equal(t, notif1.Username, notif2.Username)
	require.Equal(t, notif1.RoomID, notif2.RoomID)
	require.Equal(t, notif1.Threshold, notif2.Threshold)
	require.Equal(t, notif1.IsEnabled, notif2.IsEnabled)
	require.WithinDuration(t, newTime, notif2.LastNotifiedAt, time.Second)
}

func TestListUserRoomNotifications(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomUserRoomNotification(t)
	}

	arg := ListUserRoomNotificationsParams{
		Limit:  5,
		Offset: 0,
	}

	notifs, err := testStore.ListUserRoomNotifications(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notifs)
	require.LessOrEqual(t, len(notifs), 5)
}

func TestListUserRoomNotificationsByUser(t *testing.T) {
	user := createRandomUser(t)

	for i := 0; i < 10; i++ {
		room := createRandomRoom(t)
		arg := CreateUserRoomNotificationParams{
			Username:  user.Username,
			RoomID:    room.ID,
			Threshold: int32(util.RandomInt(10, 100)),
		}
		_, err := testStore.CreateUserRoomNotification(context.Background(), arg)
		require.NoError(t, err)
	}

	arg:=ListUserRoomNotificationsByUserParams{
		Username: user.Username,
		Limit: 5,
		Offset: 0,
	}

	notifs, err := testStore.ListUserRoomNotificationsByUser(context.Background(),arg)
	require.NoError(t, err)
	require.Len(t, notifs, 5)

	for _, notif := range notifs {
		require.NotEmpty(t, notif)
		require.Equal(t, user.Username, notif.Username)
	}
}

func TestListUserRoomNotificationsByRoom(t *testing.T) {
	room := createRandomRoom(t)

	for i := 0; i < 5; i++ {
		user := createRandomUser(t)
		arg := CreateUserRoomNotificationParams{
			Username:  user.Username,
			RoomID:    room.ID,
			Threshold: int32(util.RandomInt(10, 100)),
		}
		_, err := testStore.CreateUserRoomNotification(context.Background(), arg)
		require.NoError(t, err)
	}

	notifs, err := testStore.ListUserRoomNotificationsByRoom(context.Background(), room.ID)
	require.NoError(t, err)
	require.Len(t, notifs, 5)

	for _, notif := range notifs {
		require.NotEmpty(t, notif)
		require.Equal(t, room.ID, notif.RoomID)
	}
}

func TestListDueUserRoomNotifications(t *testing.T) {
	notif := createRandomUserRoomNotification(t)

	newTime := time.Now().Add(-25 * time.Hour)
	argUpdate := UpdateUserRoomNotificationLastNotifiedAtParams{
		Username:       notif.Username,
		RoomID:         notif.RoomID,
		LastNotifiedAt: newTime,
	}
	_, err := testStore.UpdateUserRoomNotificationLastNotifiedAt(context.Background(), argUpdate)
	require.NoError(t, err)

	notifs, err := testStore.ListDueUserRoomNotifications(context.Background())
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

func TestCountUserRoomNotifications(t *testing.T) {
	beforeCount, err := testStore.CountUserRoomNotifications(context.Background())
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		createRandomUserRoomNotification(t)
	}

	afterCount, err := testStore.CountUserRoomNotifications(context.Background())
	require.NoError(t, err)
	require.Equal(t, beforeCount+5, afterCount)
}
