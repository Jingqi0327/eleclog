package db

import (
	"context"
	"testing"
	"time"

	"github.com/Jingqi0327/eleclog/util"
	"github.com/stretchr/testify/require"
)

func createRandomElectricityRecord(t *testing.T, room Room) ElectricityRecord {

	arg := CreateElectricityRecordParams{
		RoomID:  room.ID,
		Balance: util.RandomBalance(0, 1000),
	}
	electricityRecord, err := testStore.CreateElectricityRecord(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, electricityRecord)

	return electricityRecord
}

func TestCreateElectricityRecord(t *testing.T) {
	room := createRandomRoom(t)
	createRandomElectricityRecord(t, room)
}

func createRandomElectricityRecordwithTime(t *testing.T, room Room, time time.Time) ElectricityRecord {

	arg := CreateElectricityRecordwithTimeParams{
		RoomID:     room.ID,
		Balance:    util.RandomBalance(0, 1000),
		RecordedAt: time,
	}
	electricityRecord, err := testStore.CreateElectricityRecordwithTime(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, electricityRecord)

	return electricityRecord
}

func TestCreateElectricityRecordwithTime(t *testing.T) {
	room := createRandomRoom(t)
	time := util.RandomTime()
	createRandomElectricityRecordwithTime(t, room, time)
}

func TestGetLatestBalance(t *testing.T) {
	room := createRandomRoom(t)
	record1 := createRandomElectricityRecord(t, room)
	record2 := createRandomElectricityRecordwithTime(t, room, record1.RecordedAt.Add(1*time.Hour))

	latestRecord, err := testStore.GetLatestBalance(context.Background(), room.ID)
	require.NoError(t, err)
	require.NotEmpty(t, latestRecord)

	require.Equal(t, record2.ID, latestRecord.ID)
	require.Equal(t, record2.RoomID, latestRecord.RoomID)
	require.Equal(t, record2.Balance, latestRecord.Balance)
	require.WithinDuration(t, record2.RecordedAt, latestRecord.RecordedAt, time.Minute)
}

func TestGetRecordsByHourRange(t *testing.T) {
	room := createRandomRoom(t)
	baseTime := util.RandomTime()
	for i := 0; i < 10; i++ {
		createRandomElectricityRecordwithTime(t, room, baseTime.Add(time.Duration(i) * time.Hour))
	}

	arg := GetRecordsByHourRangeParams{
		RoomID:    room.ID,
		StartTime: baseTime.Add(2 * time.Hour),
		EndTime:   baseTime.Add(5 * time.Hour),
	}
	
	records, err := testStore.GetRecordsByHourRange(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, records, 4)
	for i, record := range records {
		require.Equal(t, room.ID, record.RoomID)
		require.WithinDuration(t, baseTime.Add(time.Duration(i+2)*time.Hour), record.RecordedAt, time.Minute)
	}
}
