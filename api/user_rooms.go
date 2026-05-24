package api

import (
	"errors"
	"net/http"
	"time"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	token "github.com/Jingqi0327/eleclog/token"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type createUserRoomRequest struct {
	RoomID    int64 `json:"room_id" binding:"required,min=1"`
	Threshold int32 `json:"threshold" binding:"required,min=0"`
}

type userRoomResponse struct {
	Username       string    `json:"username"`
	RoomID         int64     `json:"room_id"`
	Threshold      int32     `json:"threshold"`
	IsEnabled      bool      `json:"is_enabled"`
	LastNotifiedAt time.Time `json:"last_notified_at"`
}

func newUserRoomResponse(n db.UserRoom) userRoomResponse {
	return userRoomResponse{
		Username:       n.Username,
		RoomID:         n.RoomID,
		Threshold:      n.Threshold,
		IsEnabled:      n.IsEnabled,
		LastNotifiedAt: n.LastNotifiedAt,
	}
}

func getAuthorizedUsername(ctx *gin.Context) string {
	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	return payload.Username
}

// 创建用户-寝室通知订阅
func (server *Server) createUserRoom(ctx *gin.Context) {
	var req createUserRoomRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateUserRoomParams{
		Username:  getAuthorizedUsername(ctx),
		RoomID:    req.RoomID,
		Threshold: req.Threshold,
		IsEnabled: true,
	}

	notification, err := server.store.CreateUserRoom(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserRoomResponse(notification))
}

type userRoomURIRequest struct {
	RoomID int64 `uri:"room_id" binding:"required,min=1"`
}

func (server *Server) getUserRoom(ctx *gin.Context) {
	var uriReq userRoomURIRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetUserRoomParams{
		Username: getAuthorizedUsername(ctx),
		RoomID:   uriReq.RoomID,
	}

	notification, err := server.store.GetUserRoom(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserRoomResponse(notification))
}

type listUserRoomsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=50"`
}

type listUserRoomsResponse struct {
	Total         int64              `json:"total"`
	Notifications []userRoomResponse `json:"notifications"`
}

func (server *Server) listUserRooms(ctx *gin.Context) {
	var req listUserRoomsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	username := getAuthorizedUsername(ctx)
	arg := db.ListUserRoomsByUserParams{
		Username: username,
		Limit:    req.PageSize,
		Offset:   (req.PageID - 1) * req.PageSize,
	}
	notifications, err := server.store.ListUserRoomsByUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	start := int((req.PageID - 1) * req.PageSize)
	if start > len(notifications) {
		start = len(notifications)
	}
	end := start + int(req.PageSize)
	if end > len(notifications) {
		end = len(notifications)
	}

	resp := listUserRoomsResponse{
		Total:         int64(len(notifications)),
		Notifications: make([]userRoomResponse, 0, end-start),
	}

	for _, n := range notifications[start:end] {
		resp.Notifications = append(resp.Notifications, newUserRoomResponse(n))
	}

	ctx.JSON(http.StatusOK, resp)
}

type updateUserRoomRequest struct {
	Threshold *int32 `json:"threshold" binding:"omitempty,min=0"`
	IsEnabled *bool  `json:"is_enabled"`
}

func (server *Server) updateUserRoom(ctx *gin.Context) {
	var uriReq userRoomURIRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateUserRoomRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateUserRoomParams{
		Username: getAuthorizedUsername(ctx),
		RoomID:   uriReq.RoomID,
	}

	if req.Threshold != nil {
		arg.Threshold = pgtype.Int4{Int32: *req.Threshold, Valid: true}
	}

	if req.IsEnabled != nil {
		arg.IsEnabled = pgtype.Bool{Bool: *req.IsEnabled, Valid: true}
	}

	notification, err := server.store.UpdateUserRoom(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserRoomResponse(notification))
}

func (server *Server) deleteUserRoom(ctx *gin.Context) {
	var uriReq userRoomURIRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.DeleteUserRoomParams{
		Username: getAuthorizedUsername(ctx),
		RoomID:   uriReq.RoomID,
	}

	err := server.store.DeleteUserRoom(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
