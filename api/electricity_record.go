package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/gin-gonic/gin"
)

type getElectricityBalanceRequest struct {
	RoomID int64 `uri:"room_id" binding:"required,min=1"`
}

type getElectricityBalanceResponse struct {
	ID         int64     `json:"id"`
	RoomID     int64     `json:"room_id"`
	Balance    int64     `json:"balance"`
	RecordedAt time.Time `json:"recorded_at"`
}

func newGetElectricityBalanceResponse(record db.ElectricityRecord) getElectricityBalanceResponse {
	return getElectricityBalanceResponse{
		ID:         record.ID,
		RoomID:     record.RoomID,
		Balance:    record.Balance,
		RecordedAt: record.RecordedAt,
	}
}

// 获取最新的电费余额
func (server *Server) getLatestElectricityBalance(ctx *gin.Context) {
	var req getElectricityBalanceRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	record, err := server.store.GetLatestBalance(ctx, req.RoomID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newGetElectricityBalanceResponse(record))
}

type getElectricityBalanceRangeRequest struct {
	StartTime time.Time `form:"start_time" binding:"required"`
	EndTime   time.Time `form:"end_time" binding:"required"`
}

// 获取指定时间范围内的电费余额记录
func (server *Server) getElectricityBalanceByHourRange(ctx *gin.Context) {
	var uriReq getElectricityBalanceRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req getElectricityBalanceRangeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetRecordsByHourRangeParams{
		RoomID:    uriReq.RoomID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	records, err := server.store.GetRecordsByHourRange(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := make([]getElectricityBalanceResponse, 0, len(records))
	for _, record := range records {
		resp = append(resp, newGetElectricityBalanceResponse(record))
	}

	ctx.JSON(http.StatusOK, resp)
}
