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
	"github.com/Jingqi0327/eleclog/token"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestCreateRoomAPI(t *testing.T) {
	room := newRoom()

	testCases := []struct {
		name             string
		body             gin.H
		addAuthorization func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs       func(store *mockdb.MockStore)
		checkResponse    func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"name":          room.Name,
				"area_id":       room.AreaID,
				"building_code": room.BuildingCode,
				"floor_code":    room.FloorCode,
				"room_code":     room.RoomCode,
			},
			addAuthorization: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateRoom(gomock.Any(), gomock.Any()).
					Times(1).
					Return(room, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				checkRoomResponse(t, recorder.Body, room)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/rooms"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.addAuthorization(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func newRoom() db.Room {
	return db.Room{
		Name:         util.RandomName(10),
		AreaID:       util.RandomString(5),
		BuildingCode: util.RandomString(5),
		FloorCode:    util.RandomString(5),
		RoomCode:     util.RandomString(5),
		CreatedAt:    time.Now(),
	}
}

func checkRoomResponse(t *testing.T, body *bytes.Buffer, expected db.Room) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var got db.Room
	err = json.Unmarshal(data, &got)

	require.NoError(t, err)
	require.Equal(t, expected.Name, got.Name)
	require.Equal(t, expected.AreaID, got.AreaID)
	require.Equal(t, expected.BuildingCode, got.BuildingCode)
	require.Equal(t, expected.FloorCode, got.FloorCode)
	require.Equal(t, expected.RoomCode, got.RoomCode)
}

func TestGetRoomAPI(t *testing.T) {
	room := newRoom()
	room.ID = util.RandomInt(1, 1000)

	testCases := []struct {
		name          string
		roomID        int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			roomID: room.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetRoom(gomock.Any(), gomock.Eq(room.ID)).
					Times(1).
					Return(room, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				checkRoomResponse(t, recorder.Body, room)
			},
		},
		{
			name:   "NotFound",
			roomID: room.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetRoom(gomock.Any(), gomock.Eq(room.ID)).
					Times(1).
					Return(db.Room{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			roomID: room.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetRoom(gomock.Any(), gomock.Eq(room.ID)).
					Times(1).
					Return(db.Room{}, sql.ErrConnDone)
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
					GetRoom(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/rooms/%d", tc.roomID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListRoomsAPI(t *testing.T) {
	n := 5
	rooms := make([]db.Room, n)
	for i := 0; i < n; i++ {
		rooms[i] = newRoom()
		rooms[i].ID = int64(i + 1)
	}

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: 5,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListRoomsParams{
					Limit:  5,
					Offset: 0,
				}
				store.EXPECT().
					ListRooms(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(rooms, nil)
				store.EXPECT().
					CountRooms(gomock.Any()).
					Times(1).
					Return(int64(n), nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			query: Query{
				pageID:   1,
				pageSize: 5,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListRooms(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Room{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				pageID:   -1,
				pageSize: 5,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListRooms(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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

			url := "/rooms"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateRoomAPI(t *testing.T) {
	room := newRoom()
	room.ID = util.RandomInt(1, 1000)

	newName := util.RandomName(10)

	testCases := []struct {
		name          string
		roomID        int64
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			roomID: room.ID,
			body: gin.H{
				"name": newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateRoomParams{
					ID:   room.ID,
					Name: pgtype.Text{String: newName, Valid: true},
				}
				updatedRoom := room
				updatedRoom.Name = newName
				store.EXPECT().
					UpdateRoom(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updatedRoom, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				checkRoomResponse(t, recorder.Body, db.Room{
					ID:           room.ID,
					Name:         newName,
					AreaID:       room.AreaID,
					BuildingCode: room.BuildingCode,
					FloorCode:    room.FloorCode,
					RoomCode:     room.RoomCode,
				})
			},
		},
		{
			name:   "NotFound",
			roomID: room.ID,
			body: gin.H{
				"name": newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateRoom(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Room{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			roomID: room.ID,
			body: gin.H{
				"name": newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateRoom(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Room{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			roomID: 0,
			body: gin.H{
				"name": newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateRoom(gomock.Any(), gomock.Any()).
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/rooms/%d", tc.roomID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteRoomAPI(t *testing.T) {
	room := newRoom()
	room.ID = util.RandomInt(1, 1000)

	testCases := []struct {
		name          string
		roomID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			roomID: room.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteRoom(gomock.Any(), gomock.Eq(room.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			roomID: room.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteRoom(gomock.Any(), gomock.Eq(room.ID)).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			roomID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteRoom(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "Unauthorized",
			roomID: room.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteRoom(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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

			url := fmt.Sprintf("/rooms/%d", tc.roomID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
