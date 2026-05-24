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

func newUserRoom() db.UserRoom {
	return db.UserRoom{
		Username:       util.RandomName(6),
		RoomID:         util.RandomInt(1, 1000),
		Threshold:      int32(util.RandomInt(10, 100)),
		IsEnabled:      true,
		LastNotifiedAt: time.Now(),
	}
}

func checkUserRoomResponse(t *testing.T, body *bytes.Buffer, expected db.UserRoom) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var got userRoomResponse
	err = json.Unmarshal(data, &got)

	require.NoError(t, err)
	require.Equal(t, expected.Username, got.Username)
	require.Equal(t, expected.RoomID, got.RoomID)
	require.Equal(t, expected.Threshold, got.Threshold)
	require.Equal(t, expected.IsEnabled, got.IsEnabled)
	require.WithinDuration(t, expected.LastNotifiedAt, got.LastNotifiedAt, time.Second)
}

func TestCreateUserRoomAPI(t *testing.T) {
	notification := newUserRoom()

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"room_id":   notification.RoomID,
				"threshold": notification.Threshold,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserRoomParams{
					Username:  notification.Username,
					RoomID:    notification.RoomID,
					Threshold: notification.Threshold,
					IsEnabled: true,
				}
				store.EXPECT().
					CreateUserRoom(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(notification, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				checkUserRoomResponse(t, recorder.Body, notification)
			},
		},
		{
			name: "Unauthorized",
			body: gin.H{
				"room_id":   notification.RoomID,
				"threshold": notification.Threshold,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserRoom(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"room_id":   notification.RoomID,
				"threshold": notification.Threshold,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserRoom(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserRoom{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidBody",
			body: gin.H{
				"room_id":   0,
				"threshold": -1,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserRoom(gomock.Any(), gomock.Any()).
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

			url := "/user-rooms"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetUserRoomAPI(t *testing.T) {
	notification := newUserRoom()

	testCases := []struct {
		name          string
		roomID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			roomID: notification.RoomID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetUserRoomParams{
					Username: notification.Username,
					RoomID:   notification.RoomID,
				}
				store.EXPECT().
					GetUserRoom(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(notification, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				checkUserRoomResponse(t, recorder.Body, notification)
			},
		},
		{
			name:   "NotFound",
			roomID: notification.RoomID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserRoom(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserRoom{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:   "Unauthorized",
			roomID: notification.RoomID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserRoom(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			roomID: notification.RoomID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserRoom(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserRoom{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			roomID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserRoom(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/user-rooms/%d", tc.roomID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListUserRoomsAPI(t *testing.T) {
	n := 5
	notifications := make([]db.UserRoom, n)
	username := util.RandomName(6)
	for i := 0; i < n; i++ {
		notifications[i] = newUserRoom()
		notifications[i].Username = username
	}

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: 5,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListUserRoomsByUserParams{
					Username: username,
					Limit:    5,
					Offset:   0,
				}
				store.EXPECT().
					ListUserRoomsByUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(notifications, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			query: Query{
				pageID:   1,
				pageSize: 5,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUserRoomsByUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			query: Query{
				pageID:   1,
				pageSize: 5,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUserRoomsByUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.UserRoom{}, sql.ErrConnDone)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUserRoomsByUser(gomock.Any(), gomock.Any()).
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

			url := "/user-rooms"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateUserRoomAPI(t *testing.T) {
	notification := newUserRoom()
	newThreshold := int32(util.RandomInt(10, 100))
	newIsEnabled := false

	testCases := []struct {
		name          string
		roomID        int64
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			roomID: notification.RoomID,
			body: gin.H{
				"threshold":  newThreshold,
				"is_enabled": newIsEnabled,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserRoomParams{
					Username:  notification.Username,
					RoomID:    notification.RoomID,
					Threshold: pgtype.Int4{Int32: newThreshold, Valid: true},
					IsEnabled: pgtype.Bool{Bool: newIsEnabled, Valid: true},
				}
				updatedNotification := notification
				updatedNotification.Threshold = newThreshold
				updatedNotification.IsEnabled = newIsEnabled
				store.EXPECT().
					UpdateUserRoom(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updatedNotification, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				checkUserRoomResponse(t, recorder.Body, db.UserRoom{
					Username:       notification.Username,
					RoomID:         notification.RoomID,
					Threshold:      newThreshold,
					IsEnabled:      newIsEnabled,
					LastNotifiedAt: notification.LastNotifiedAt,
				})
			},
		},
		{
			name:   "NotFound",
			roomID: notification.RoomID,
			body: gin.H{
				"threshold":  newThreshold,
				"is_enabled": newIsEnabled,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserRoom(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserRoom{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:   "Unauthorized",
			roomID: notification.RoomID,
			body: gin.H{
				"threshold":  newThreshold,
				"is_enabled": newIsEnabled,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserRoom(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			roomID: notification.RoomID,
			body: gin.H{
				"threshold":  newThreshold,
				"is_enabled": newIsEnabled,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserRoom(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserRoom{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			roomID: 0,
			body: gin.H{
				"threshold":  newThreshold,
				"is_enabled": newIsEnabled,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserRoom(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/user-rooms/%d", tc.roomID)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteUserRoomAPI(t *testing.T) {
	notification := newUserRoom()

	testCases := []struct {
		name          string
		roomID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			roomID: notification.RoomID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.DeleteUserRoomParams{
					Username: notification.Username,
					RoomID:   notification.RoomID,
				}
				store.EXPECT().
					DeleteUserRoom(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name:   "Unauthorized",
			roomID: notification.RoomID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUserRoom(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			roomID: notification.RoomID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUserRoom(gomock.Any(), gomock.Any()).
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
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, notification.Username, util.UserRole, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUserRoom(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/user-rooms/%d", tc.roomID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
