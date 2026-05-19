package worker

import (
	"github.com/hibiken/asynq"
)

type TaskScheduler interface {
	Start() error
	Shutdown()
	ScheduleDetectLowBalance(cron string) error
}

type RedisTaskScheduler struct {
	scheduler *asynq.Scheduler
}

func NewRedisTaskScheduler(redisOpt asynq.RedisClientOpt) TaskScheduler {
	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		Logger: NewAsynqLogger(),
	})
	return &RedisTaskScheduler{
		scheduler: scheduler,
	}
}

func (s *RedisTaskScheduler) Start() error {
	return s.scheduler.Start()
}

func (s *RedisTaskScheduler) Shutdown() {
	s.scheduler.Shutdown()
}
