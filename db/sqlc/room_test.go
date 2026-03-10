package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Jingqi0327/eleclog/util"
	"github.com/stretchr/testify/require"
)

func createRandomRoom(t *testing.T) Room {
	arg := CreateRoomParams{
		Name:         util.RandomName(6),
		AreaID:       util.RandomCode(6),
		BuildingCode: util.RandomCode(6),
		FloorCode:    util.RandomCode(6),
		RoomCode:     util.RandomCode(6),
	}

	room, err := testStore.CreateRoom(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, room)

	require.Equal(t, arg.Name, room.Name)
	require.Equal(t, arg.AreaID, room.AreaID)
	require.Equal(t, arg.BuildingCode, room.BuildingCode)
	require.Equal(t, arg.FloorCode, room.FloorCode)
	require.Equal(t, arg.RoomCode, room.RoomCode)

	require.NotZero(t, room.ID)
	require.NotZero(t, room.CreatedAt)

	return room
}

func TestCreateRoom(t *testing.T) {
	createRandomRoom(t)
}

func TestGetRoom(t *testing.T) {
	room1 := createRandomRoom(t)

	room2, err := testStore.GetRoom(context.Background(), room1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, room2)

	require.Equal(t, room1.ID, room2.ID)
	require.Equal(t, room1.Name, room2.Name)
	require.Equal(t, room1.AreaID, room2.AreaID)
	require.Equal(t, room1.BuildingCode, room2.BuildingCode)
	require.Equal(t, room1.FloorCode, room2.FloorCode)
	require.Equal(t, room1.RoomCode, room2.RoomCode)
	require.WithinDuration(t, room1.CreatedAt, room2.CreatedAt, time.Minute)
}

func TestListRooms(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomRoom(t)
	}

	arg := ListRoomsParams{
		Limit:  5,
		Offset: 5,
	}

	rooms, err := testStore.ListRooms(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, rooms, 5)

	for _, room := range rooms {
		require.NotEmpty(t, room)
	}
}

func TestListRoomsAll(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomRoom(t)
	}

	rooms, err := testStore.ListRoomsAll(context.Background())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rooms), 10)

	for _, room := range rooms {
		require.NotEmpty(t, room)
	}
}

func TestDeleteRoom(t *testing.T) {
	room1 := createRandomRoom(t)

	err := testStore.DeleteRoom(context.Background(), room1.ID)
	require.NoError(t, err)

	room2, err := testStore.GetRoom(context.Background(), room1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, room2)
}

func TestUpdateRoom(t *testing.T) {
	room1 := createRandomRoom(t)

	newName:=util.RandomName(6)
	newAreaID:=util.RandomCode(6)
	newBuildingCode:=util.RandomCode(6)
	newFloorCode:=util.RandomCode(6)
	newRoomCode:=util.RandomCode(6)

	arg := UpdateRoomParams{
		ID:           room1.ID,
		Name:         sql.NullString{String: newName, Valid: true},
		AreaID:       sql.NullString{String: newAreaID, Valid: true},
		BuildingCode: sql.NullString{String: newBuildingCode, Valid: true},
		FloorCode:    sql.NullString{String: newFloorCode, Valid: true},
		RoomCode:     sql.NullString{String: newRoomCode, Valid: true},
	}

	room2, err := testStore.UpdateRoom(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, room2)

	require.Equal(t, room1.ID, room2.ID)
	require.Equal(t, newName, room2.Name)
	require.Equal(t, newAreaID, room2.AreaID)
	require.Equal(t, newBuildingCode, room2.BuildingCode)
	require.Equal(t, newFloorCode, room2.FloorCode)
	require.Equal(t, newRoomCode, room2.RoomCode)
	require.WithinDuration(t, room1.CreatedAt, room2.CreatedAt, time.Minute)	

}