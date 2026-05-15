package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/Jingqi0327/eleclog/db/mock"
	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func newElectricityRecord() db.ElectricityRecord {
	return db.ElectricityRecord{
		ID:         util.RandomInt(1, 1000),
		RoomID:     util.RandomInt(1, 1000),
		Balance:    util.RandomInt(0, 10000),
		RecordedAt: time.Now(),
	}
}

func checkElectricityBalanceResponse(t *testing.T, body *bytes.Buffer, expected db.ElectricityRecord) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var got getElectricityBalanceResponse
	err = json.Unmarshal(data, &got)

	require.NoError(t, err)
	require.Equal(t, expected.ID, got.ID)
	require.Equal(t, expected.RoomID, got.RoomID)
	require.Equal(t, expected.Balance, got.Balance)
	require.WithinDuration(t, expected.RecordedAt, got.RecordedAt, time.Second)
}

func TestGetLatestElectricityBalanceAPI(t *testing.T) {
	record := newElectricityRecord()

	testCases := []struct {
		name          string
		roomID        int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			roomID: record.RoomID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLatestBalance(gomock.Any(), gomock.Eq(record.RoomID)).
					Times(1).
					Return(record, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				checkElectricityBalanceResponse(t, recorder.Body, record)
			},
		},
		{
			name:   "NotFound",
			roomID: record.RoomID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLatestBalance(gomock.Any(), gomock.Eq(record.RoomID)).
					Times(1).
					Return(db.ElectricityRecord{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			roomID: record.RoomID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLatestBalance(gomock.Any(), gomock.Eq(record.RoomID)).
					Times(1).
					Return(db.ElectricityRecord{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			roomID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLatestBalance(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/electricity-balances/latest/%d", tc.roomID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetElectricityRecordByHourRangeAPI(t *testing.T) {
	roomID := util.RandomInt(1, 1000)
	now := time.Now()
	startTime := now.Add(-3 * time.Hour)
	endTime := now

	// 模拟3条记录：一条补位（1小时10分前的误差时间），两条在区间内
	records := []db.ElectricityRecord{
		{
			ID:         util.RandomInt(1, 1000),
			RoomID:     roomID,
			Balance:    3000,
			RecordedAt: startTime.Add(-1 * time.Hour), // 基准点
		},
		{
			ID:         util.RandomInt(1, 1000),
			RoomID:     roomID,
			Balance:    2500, // 消耗了500分
			RecordedAt: startTime.Add(1 * time.Hour),
		},
		{
			ID:         util.RandomInt(1, 1000),
			RoomID:     roomID,
			Balance:    2000, // 消耗了500分
			RecordedAt: startTime.Add(2 * time.Hour),
		},
	}

	testCases := []struct {
		name          string
		roomID        int64
		query         gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			roomID: roomID,
			query: gin.H{
				"start_time": startTime.Format(time.RFC3339),
				"end_time":   endTime.Format(time.RFC3339),
			},
			buildStubs: func(store *mockdb.MockStore) {
				// bufferStartTime := req.StartTime.Add(-1 * time.Hour - 10 * time.Minute)
				// 匹配参数
				store.EXPECT().
					GetRecordsByHourRange(gomock.Any(), gomock.Any()). // TODO: match precise params
					Times(1).
					Return(records, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var got []electricityUsageResponse
				err = json.Unmarshal(data, &got)
				require.NoError(t, err)

				require.Len(t, got, 2)
				require.Equal(t, float64(20), got[1].Balance)
			},
		},
		{
			name:   "InternalError",
			roomID: roomID,
			query: gin.H{
				"start_time": startTime.Format(time.RFC3339),
				"end_time":   endTime.Format(time.RFC3339),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetRecordsByHourRange(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ElectricityRecord{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidQuery",
			roomID: roomID,
			query: gin.H{
				"start_time": "invalid-time",
				"end_time":   endTime.Format(time.RFC3339),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetRecordsByHourRange(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/electricity-balances/hour-range/%d", tc.roomID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := request.URL.Query()
			for k, v := range tc.query {
				q.Add(k, v.(string))
			}
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
