package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/token"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type createRoomRequest struct {
	Name         string `json:"name" binding:"required"`
	AreaID       string `json:"area_id" binding:"required"`
	BuildingCode string `json:"building_code" binding:"required"`
	FloorCode    string `json:"floor_code" binding:"required"`
	RoomCode     string `json:"room_code" binding:"required"`
}

type createRoomResponse struct {
	Name         string    `json:"name"`
	AreaID       string    `json:"area_id" `
	BuildingCode string    `json:"building_code"`
	FloorCode    string    `json:"floor_code"`
	RoomCode     string    `json:"room_code"`
	CreatedAt    time.Time `json:"created_at"`
}

func (server *Server) createRoom(ctx *gin.Context) {
	var req createRoomRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateRoomParams{
		Name:         req.Name,
		AreaID:       req.AreaID,
		BuildingCode: req.BuildingCode,
		FloorCode:    req.FloorCode,
		RoomCode:     req.RoomCode,
	}

	room, err := server.store.CreateRoom(ctx, arg)
	if err != nil { 
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("room already exists")))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := createRoomResponse{
		Name:         room.Name,
		AreaID:       room.AreaID,
		BuildingCode: room.BuildingCode,
		FloorCode:    room.FloorCode,
		RoomCode:     room.RoomCode,
		CreatedAt:    room.CreatedAt,
	}

	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) bindRoomToUser(ctx *gin.Context) {
	var req createRoomRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	argFields := db.GetRoomByUniqueFieldsParams{
		Name:         req.Name,
		AreaID:       req.AreaID,
		BuildingCode: req.BuildingCode,
		FloorCode:    req.FloorCode,
		RoomCode:     req.RoomCode,
	}

	var room db.Room
	room, err := server.store.GetRoomByUniqueFields(ctx, argFields)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Room does not exist, create it
			argCreate := db.CreateRoomParams{
				Name:         req.Name,
				AreaID:       req.AreaID,
				BuildingCode: req.BuildingCode,
				FloorCode:    req.FloorCode,
				RoomCode:     req.RoomCode,
			}
			room, err = server.store.CreateRoom(ctx, argCreate)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	// Now we have the room, bind it to the current user
	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	argBind := db.CreateUserRoomParams{
		Username:  payload.Username,
		RoomID:    room.ID,
		Threshold: 10,
		IsEnabled: false,
	}

	_, err = server.store.CreateUserRoom(ctx, argBind)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			// Already bound, we can just return success
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	resp := createRoomResponse{
		Name:         room.Name,
		AreaID:       room.AreaID,
		BuildingCode: room.BuildingCode,
		FloorCode:    room.FloorCode,
		RoomCode:     room.RoomCode,
		CreatedAt:    room.CreatedAt,
	}

	ctx.JSON(http.StatusOK, resp)
}

type getRoomRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type getRoomResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	AreaID       string    `json:"area_id"`
	BuildingCode string    `json:"building_code"`
	FloorCode    string    `json:"floor_code"`
	RoomCode     string    `json:"room_code"`
	CreatedAt    time.Time `json:"created_at"`
}

func newGetRoomResponse(room db.Room) getRoomResponse {
	return getRoomResponse{
		ID:           room.ID,
		Name:         room.Name,
		AreaID:       room.AreaID,
		BuildingCode: room.BuildingCode,
		FloorCode:    room.FloorCode,
		RoomCode:     room.RoomCode,
		CreatedAt:    room.CreatedAt,
	}
}

func (server *Server) getRoom(ctx *gin.Context) {
	var req getRoomRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	room, err := server.store.GetRoom(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newGetRoomResponse(room))
}

type listRoomsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=50"`
}

type listRoomsResponse struct {
	Total int64             `json:"total"`
	Rooms []getRoomResponse `json:"rooms"`
}

func (server *Server) listRooms(ctx *gin.Context) {
	var req listRoomsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	var rooms []db.Room
	var total int64
	var err error

	if payload.Role == util.UserRole {
		argUser := db.ListRoomsByUserParams{
			Username: payload.Username,
			Limit:    req.PageSize,
			Offset:   (req.PageID - 1) * req.PageSize,
		}
		rooms, err = server.store.ListRoomsByUser(ctx, argUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		total, err = server.store.CountRoomsByUser(ctx, payload.Username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	} else {
		argAll := db.ListRoomsParams{
			Limit:  req.PageSize,
			Offset: (req.PageID - 1) * req.PageSize,
		}
		rooms, err = server.store.ListRooms(ctx, argAll)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		total, err = server.store.CountRooms(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	resp := listRoomsResponse{
		Total: total,
		Rooms: make([]getRoomResponse, 0, len(rooms)),
	}
	for _, room := range rooms {
		resp.Rooms = append(resp.Rooms, newGetRoomResponse(room))
	}

	ctx.JSON(http.StatusOK, resp)
}

type updateRoomRequest struct {
	Name         *string `json:"name"`
	AreaID       *string `json:"area_id" binding:"omitempty,alphanum"`
	BuildingCode *string `json:"building_code" binding:"omitempty,alphanum"`
	FloorCode    *string `json:"floor_code" binding:"omitempty,alphanum"`
	RoomCode     *string `json:"room_code" binding:"omitempty,alphanum"`
}

func (server *Server) updateRoom(ctx *gin.Context) {
	var uriReq getRoomRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateRoomRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateRoomParams{
		ID: uriReq.ID,
	}

	if req.Name != nil {
		arg.Name = pgtype.Text{String: *req.Name, Valid: true}
	}

	if req.AreaID != nil {
		arg.AreaID = pgtype.Text{String: *req.AreaID, Valid: true}
	}

	if req.BuildingCode != nil {
		arg.BuildingCode = pgtype.Text{String: *req.BuildingCode, Valid: true}
	}

	if req.FloorCode != nil {
		arg.FloorCode = pgtype.Text{String: *req.FloorCode, Valid: true}
	}

	if req.RoomCode != nil {
		arg.RoomCode = pgtype.Text{String: *req.RoomCode, Valid: true}
	}

	room, err := server.store.UpdateRoom(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newGetRoomResponse(room))
}

func (server *Server) deleteRoom(ctx *gin.Context) {
	var req getRoomRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteRoom(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
