package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendNotificationEmail(ctx context.Context, payload *PayloadSendNotificationEmail, opts ...asynq.Option) error
	DistributeTaskDetectLowBalance(ctx context.Context, opts ...asynq.Option) error
}

// RedisTaskDistributor 负责将任务分发到 Redis 队列
type RedisTaskDistributor struct {
	client *asynq.Client // Asynq 客户端，用于连接 Redis 并发送任务
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
