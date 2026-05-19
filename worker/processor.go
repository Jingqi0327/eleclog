package worker

import (
	"context"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/mail"
	"github.com/hibiken/asynq"
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
