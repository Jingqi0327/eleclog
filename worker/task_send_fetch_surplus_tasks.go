package worker

import (
	"context"
	"fmt"

	"github.com/Jingqi0327/eleclog/logger"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const TaskSendFetchSurplusTasks = "task:send_fetch_surplus_tasks"

func (distributor *RedisTaskDistributor) DistributeTaskSendFetchSurplusTasks(ctx context.Context, opts ...asynq.Option) error {
	// 创建一个新的 Asynq 任务
	task := asynq.NewTask(TaskSendFetchSurplusTasks, nil, opts...)
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

func (scheduler *RedisTaskScheduler) ScheduleTaskSendFetchSurplusTasks(cron string) error {
	task := asynq.NewTask(TaskSendFetchSurplusTasks, nil)
	// 注册定时任务
	_, err := scheduler.scheduler.Register(cron, task)
	if err != nil {
		return err
	}
	logger.Log.Info("[Scheduler] Registered task: 发送获取剩余电量的任务",
		zap.String("cron", cron),
	)
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendFetchSurplusTasks(ctx context.Context, task *asynq.Task) error {
	logger.Log.Info("[Processor] 开始发送获取剩余电量的任务")

	rooms, err := processor.store.ListRoomsAll(ctx)
	if err != nil {
		return fmt.Errorf("fail to list rooms: %w", err)
	}

	for _, room := range rooms {
		payload := &PayloadFetchSurplusAndStore{
			RoomID:       room.ID,
			RoomName:     room.Name,
			AreaID:       room.AreaID,
			BuildingCode: room.BuildingCode,
			FloorCode:    room.FloorCode,
			RoomCode:     room.RoomCode,
		}
		err := processor.distributor.DistributeTaskFetchSurplusAndStore(ctx, payload)
		if err != nil {
			logger.Log.Error("fail to distribute task",
				zap.String("room_name", room.Name),
				zap.Error(err),
			)
		}
	}

	return nil
}
