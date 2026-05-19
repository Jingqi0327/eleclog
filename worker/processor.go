package worker

import (
	"context"
	"errors"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/logger"
	"github.com/Jingqi0327/eleclog/mail"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessTaskSendNotificationEmail(ctx context.Context, task *asynq.Task) error
	ProcessTaskDetectLowBalance(ctx context.Context, task *asynq.Task) error
}

// RedisTaskProcessor 负责从 Redis 队列中取出任务并执行
type RedisTaskProcessor struct {
	server      *asynq.Server // Asynq 服务器，用于连接 Redis 并处理任务
	store       db.Store      // 数据库存储接口，提供访问数据库的方法
	emailSender mail.EmailSender
	distributor *RedisTaskDistributor
}

func NewRedisTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	store db.Store,
	emailSender mail.EmailSender,
	taskDistributor *RedisTaskDistributor,
) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				// 对于明确放弃重试的任务，打 Warn 就行
				if errors.Is(err, asynq.SkipRetry) {
					logger.Log.Warn("[Asynq] Task skipped retry",
						zap.Error(err),
						zap.String("type", task.Type()),
					)
					return
				}

				// 对于普通会重试的错误，打 Error
				logger.Log.Error("[Asynq] Failed to process task",
					zap.Error(err),
					zap.String("type", task.Type()),
					zap.ByteString("payload", task.Payload()))
			}),
			Logger: NewAsynqLogger(),
		},
	)

	return &RedisTaskProcessor{
		server:      server,
		store:       store,
		emailSender: emailSender,
		distributor: taskDistributor,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()                                                            // 创建一个新的 ServeMux，用于注册任务处理函数
	mux.HandleFunc(TaskDetectLowBalance, processor.ProcessTaskDetectLowBalance)           // 注册处理 TaskDetectLowBalance 任务的函数
	mux.HandleFunc(TaskSendNotificationEmail, processor.ProcessTaskSendNotificationEmail) // 注册处理 SendVerifyEmail 任务的函数

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Shutdown() {
	processor.server.Shutdown()
}
