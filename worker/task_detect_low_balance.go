package worker

import (
	"context"
	"fmt"

	"github.com/Jingqi0327/eleclog/logger"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const TaskDetectLowBalance = "task:detect_low_balance"

func (distributor *RedisTaskDistributor) DistributeTaskDetectLowBalance(ctx context.Context, opts ...asynq.Option) error {
	// 创建一个新的 Asynq 任务
	task := asynq.NewTask(TaskDetectLowBalance, nil, opts...)
	// 将任务发送到 Redis 队列
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		logger.Log.Error("[Scheduler] Enqueue task failed",
			zap.String("type", task.Type()),
			zap.ByteString("payload", task.Payload()),
			zap.Error(err),
		)
		return fmt.Errorf("fail to enqueue task: %w", err)
	}
	logger.Log.Info("[Scheduler] Enqueued task",
		zap.String("type", task.Type()),
		zap.ByteString("payload", task.Payload()),
		zap.String("queue", info.Queue),
		zap.Int("max_retry", info.MaxRetry),
	)
	return nil
}

func (scheduler *RedisTaskScheduler) ScheduleDetectLowBalance(cron string) error {
	task := asynq.NewTask(TaskDetectLowBalance, nil)
	// 注册定时任务
	_, err := scheduler.scheduler.Register(cron, task)
	if err != nil {
		return err
	}
	logger.Log.Info("[Scheduler] Registered task: 检测低于余额阈值的房间",
		zap.String("cron", cron),
	)
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskDetectLowBalance(ctx context.Context, task *asynq.Task) error {
	logger.Log.Info("[Processor] Processing task: 开始检测低于余额阈值的房间")

	notifications, err := processor.store.ListDueUserRoomNotifications(ctx)
	if err != nil {
		logger.Log.Error("[Processor] Find due notifications failed",
			zap.Error(err),
		)
		return err
	}

	roomCurrentBalance := make(map[int64]int64) // 将房间当前的余额存入map，避免重复查询
	for _, notification := range notifications {
		if _, exists := roomCurrentBalance[notification.RoomID]; !exists {
			record, err := processor.store.GetLatestBalance(ctx, notification.RoomID)
			if err != nil {
				logger.Log.Error("[Processor] 查询房间当前剩余电量失败",
					zap.Int64("room_id", notification.RoomID),
					zap.Error(err),
				)
				continue
			}
			roomCurrentBalance[notification.RoomID] = record.Balance
		}

		curSurplus := roomCurrentBalance[notification.RoomID]

		if curSurplus < int64(notification.Threshold*100) {
			sendEmailPayload := &PayloadSendNotificationEmail{
				Username:  notification.Username,
				RoomID:    notification.RoomID,
				Surplus:   curSurplus,
				Threshold: int64(notification.Threshold),
			}
			err := processor.distributor.DistributeTaskSendNotificationEmail(ctx, sendEmailPayload)
			if err != nil {
				logger.Log.Error("[Processor] Fail to enqueue send email task",
					zap.Error(err),
				)
				continue
			}
		}
	}

	return nil
}
