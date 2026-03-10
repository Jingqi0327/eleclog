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

type electricityUsageResponse struct {
    StartTime  time.Time `json:"start_time"`
    EndTime    time.Time `json:"end_time"`
    Usage      float64   `json:"usage"`   // 本周期用电量（度）
    Balance    float64   `json:"balance"` // 结束时的余额（元）
}

func (server *Server) getElectricityRecordByHourRange(ctx *gin.Context) {
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

    // 💡 核心改动 1：为了计算第一个点的用量，我们需要往前多抓一个点
    // 假设是一小时采集一次，我们往前推 1 小时 10 分钟（预留一点误差）
    bufferStartTime := req.StartTime.Add(-1*time.Hour - 10*time.Minute)

    arg := db.GetRecordsByHourRangeParams{
        RoomID:    uriReq.RoomID,
        StartTime: bufferStartTime,
        EndTime:   req.EndTime,
    }

    records, err := server.store.GetRecordsByHourRange(ctx, arg)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }

    // 💡 核心改动 2：通过循环计算差值
    resp := make([]electricityUsageResponse, 0)

    // 我们需要至少 2 条数据才能算出 1 个区间的用量
    for i := 1; i < len(records); i++ {
        prev := records[i-1]
        curr := records[i]

        // 如果当前点的时间早于用户要求的开始时间，说明这是用来补位的基准点，不放入结果集
        // 但我们要用它来计算 i=1 时的 usage
        if curr.RecordedAt.Before(req.StartTime) {
            continue
        }

        // 计算用量 (前一次余额 - 本次余额) / 电价
        // 记得将 int64 的分转为 float64 的元
        usage := float64(prev.Balance-curr.Balance) / 100.0 / server.config.PricePerKWh
        
        // 容错处理：如果充值了，差值为负，此时用量计为 0
        if usage < 0 {
            usage = 0
        }

        resp = append(resp, electricityUsageResponse{
            StartTime: prev.RecordedAt,
            EndTime:   curr.RecordedAt,
            Usage:     usage,
            Balance:   float64(curr.Balance) / 100.0,
        })
    }

    ctx.JSON(http.StatusOK, resp)
}