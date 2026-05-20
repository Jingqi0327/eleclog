package worker

import (
	"context"
	"encoding/json"
	"fmt"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/logger"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const TaskFetchSurplusAndStore = "task:fetch_surplus_and_store"

type PayloadFetchSurplusAndStore struct {
	RoomID       int64  `json:"room_id"`
	RoomName     string `json:"room_name"`
	AreaID       string `json:"area_id"`
	BuildingCode string `json:"building_code"`
	FloorCode    string `json:"floor_code"`
	RoomCode     string `json:"room_code"`
}

func (distributor *RedisTaskDistributor) DistributeTaskFetchSurplusAndStore(ctx context.Context, payload *PayloadFetchSurplusAndStore, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fail to marshal task payload: %w", err)
	}
	// 创建一个新的 Asynq 任务
	task := asynq.NewTask(TaskFetchSurplusAndStore, jsonPayload, opts...)
	// 将任务发送到 Redis 队列
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("Fail to enqueue task: %w", err)
	}
	logger.Log.Info("[Scheduler] Enqueued task",
		zap.String("type", task.Type()),
		zap.ByteString("payload", task.Payload()),
		zap.String("queue", info.Queue),
		zap.Int("max_retry", info.MaxRetry),
	)
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskFetchSurplusAndStore(ctx context.Context, task *asynq.Task) error {
	var payload PayloadFetchSurplusAndStore
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("fail to unmarshal payload: %w", err)
	}

	balance, err := processor.fetchSurplus(payload.AreaID, payload.BuildingCode, payload.FloorCode, payload.RoomCode)
	if err != nil {
		return fmt.Errorf("fail to fetch surplus: %w", err)
	}

	_, err = processor.store.CreateElectricityRecord(ctx, db.CreateElectricityRecordParams{
		RoomID:  payload.RoomID,
		Balance: util.ToCents(balance), // 转换为分存储
	})

	if err != nil {
		return fmt.Errorf("fail to record electricity record: %w", err)
	}

	logger.Log.Info("成功记录寝室电量",
		zap.String("room_name", payload.RoomName),
		zap.String("balance", util.FormatCentsToYuan(util.ToCents(balance))),
	)

	return nil
}

type SurplusResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Amount          float64 `json:"amount"`          // 剩余金额 (元)
		DisplayRoomName string  `json:"displayRoomName"` // 房间全称
	} `json:"data"`
}

func (processor *RedisTaskProcessor) fetchSurplus(areaID, buildingCode, floorCode, roomCode string) (float64, error) {
	apiURL := "https://application.xiaofubao.com/app/electric/queryRoomSurplus"

	var result SurplusResponse
	resp, err := processor.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36").
		SetHeader("Cookie", fmt.Sprintf("shiroJID=%s", processor.config.ShiroJID)).
		SetQueryParams(map[string]string{
			"platform":     "YUNMA_APP",
			"areaId":       areaID,
			"buildingCode": buildingCode,
			"floorCode":    floorCode,
			"roomCode":     roomCode,
		}).
		SetResult(&result). // 自动将 JSON 解析到 struct
		Post(apiURL)

	if err != nil {
		return 0, fmt.Errorf("请求失败: %v", err)
	}

	if !resp.IsSuccess() {
		return 0, fmt.Errorf("接口返回错误代码: %d", resp.StatusCode())
	}

	if !result.Success {
		return 0, fmt.Errorf("接口返回失败: %v", result)
	}

	return result.Data.Amount, nil
}
