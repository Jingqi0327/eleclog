package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	token "github.com/Jingqi0327/eleclog/token"
	"github.com/gin-gonic/gin"
)

type createUserRoomNotificationRequest struct {
	RoomID    int64 `json:"room_id" binding:"required,min=1"`
	Threshold int32 `json:"threshold" binding:"required,min=0"`
}

type userRoomNotificationResponse struct {
	Username       string    `json:"username"`
	RoomID         int64     `json:"room_id"`
	Threshold      int32     `json:"threshold"`
	IsEnabled      bool      `json:"is_enabled"`
	LastNotifiedAt time.Time `json:"last_notified_at"`
}

func newUserRoomNotificationResponse(n db.UserRoomNotification) userRoomNotificationResponse {
	return userRoomNotificationResponse{
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
func (server *Server) createUserRoomNotification(ctx *gin.Context) {
	var req createUserRoomNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateUserRoomNotificationParams{
		Username:  getAuthorizedUsername(ctx),
		RoomID:    req.RoomID,
		Threshold: req.Threshold,
	}

	notification, err := server.store.CreateUserRoomNotification(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserRoomNotificationResponse(notification))
}

type userRoomNotificationURIRequest struct {
	RoomID int64 `uri:"room_id" binding:"required,min=1"`
}

func (server *Server) getUserRoomNotification(ctx *gin.Context) {
	var uriReq userRoomNotificationURIRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetUserRoomNotificationParams{
		Username: getAuthorizedUsername(ctx),
		RoomID:   uriReq.RoomID,
	}

	notification, err := server.store.GetUserRoomNotification(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserRoomNotificationResponse(notification))
}

type listUserRoomNotificationsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=50"`
}

type listUserRoomNotificationsResponse struct {
	Total         int64                          `json:"total"`
	Notifications []userRoomNotificationResponse `json:"notifications"`
}

func (server *Server) listUserRoomNotifications(ctx *gin.Context) {
	var req listUserRoomNotificationsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	username := getAuthorizedUsername(ctx)
	arg:=db.ListUserRoomNotificationsByUserParams{
		Username: username,
		Limit: req.PageSize,
		Offset: (req.PageID-1)*req.PageSize,
	}
	notifications, err := server.store.ListUserRoomNotificationsByUser(ctx, arg)
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

	resp := listUserRoomNotificationsResponse{
		Total:         int64(len(notifications)),
		Notifications: make([]userRoomNotificationResponse, 0, end-start),
	}

	for _, n := range notifications[start:end] {
		resp.Notifications = append(resp.Notifications, newUserRoomNotificationResponse(n))
	}

	ctx.JSON(http.StatusOK, resp)
}

type updateUserRoomNotificationRequest struct {
	Threshold *int32 `json:"threshold" binding:"omitempty,min=0"`
	IsEnabled *bool  `json:"is_enabled"`
}

func (server *Server) updateUserRoomNotification(ctx *gin.Context) {
	var uriReq userRoomNotificationURIRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateUserRoomNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateUserRoomNotificationParams{
		Username: getAuthorizedUsername(ctx),
		RoomID:   uriReq.RoomID,
	}

	if req.Threshold != nil {
		arg.Threshold = sql.NullInt32{Int32: *req.Threshold, Valid: true}
	}

	if req.IsEnabled != nil {
		arg.IsEnabled = sql.NullBool{Bool: *req.IsEnabled, Valid: true}
	}

	notification, err := server.store.UpdateUserRoomNotification(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserRoomNotificationResponse(notification))
}

func (server *Server) deleteUserRoomNotification(ctx *gin.Context) {
	var uriReq userRoomNotificationURIRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.DeleteUserRoomNotificationParams{
		Username: getAuthorizedUsername(ctx),
		RoomID:   uriReq.RoomID,
	}

	err := server.store.DeleteUserRoomNotification(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
