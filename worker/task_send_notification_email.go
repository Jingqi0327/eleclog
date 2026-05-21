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

const TaskSendNotificationEmail = "task:send_notification_email"

// 任务负载（Payload）结构体
type PayloadSendNotificationEmail struct {
	Username  string `json:"username"`
	RoomID    int64  `json:"room_id"`
	Surplus   int64  `json:"surplus"`
	Threshold int64  `json:"threshold"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendNotificationEmail(ctx context.Context, payload *PayloadSendNotificationEmail, opts ...asynq.Option) error {
	// 将任务负载序列化为 JSON 格式
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fail to marshal task payload: %w", err)
	}
	// 创建一个新的 Asynq 任务
	task := asynq.NewTask(TaskSendNotificationEmail, jsonPayload, opts...)
	// 将任务发送到 Redis 队列
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("fail to enqueue task: %w", err)
	}
	logger.Log.Info("[Processor] Enqueued task: 发送电量不足通知",
		zap.String("type", task.Type()),
		zap.ByteString("payload", task.Payload()),
		zap.String("queue", info.Queue),
		zap.Int("max_retry", info.MaxRetry),
	)
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendNotificationEmail(ctx context.Context, task *asynq.Task) error {
	// 将任务负载从 JSON 格式反序列化为结构体
	var payload PayloadSendNotificationEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("fail to unmarshal payload: %w", err)
	}
	// 从数据库中获取用户信息
	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == db.ErrRecordNotFound {
			return fmt.Errorf("user doesn't exist: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("fail to get user: %w", err)
	}

	subject := "寝室电量不足通知"
	content := "您好，您的寝室电量剩余 " + util.FormatCentsToYuan(payload.Surplus) + " 元，已低于您设置的阈值 " + util.FormatCentsToYuan(payload.Threshold*100) + " 元，请及时充值。"
	to := []string{user.Email}
	err = processor.emailSender.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("fail to send email: %w", err)
	}
	logger.Log.Info("[Processor] Processed task",
		zap.String("type", task.Type()),
		zap.ByteString("payload", task.Payload()),
		zap.String("email", user.Email),
	)
	
	return nil
}
