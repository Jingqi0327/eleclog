package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jingqi0327/eleclog/api"
	"github.com/Jingqi0327/eleclog/collector"
	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/logger"
	"github.com/Jingqi0327/eleclog/mail"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/Jingqi0327/eleclog/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// 定义停机信号列表，包含常见的中断信号
var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	bootstrapLog, _ := zap.NewProduction()
	defer bootstrapLog.Sync()

	config, err := util.LoadConfig(".")
	if err != nil {
		bootstrapLog.Fatal("cannot load config", zap.Error(err))
	}

	if config.Environment == "development" {
		logger.InitLogger(true)
	} else {
		logger.InitLogger(false)
	}
	defer logger.Log.Sync()

	logger.Log.Info("[System] Starting server...")

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		logger.Log.Fatal("[System] Cannot connect to db:", zap.Error(err))
	}

	store := db.NewStore(connPool)
	runMigrate(config.MigrationURL, config.DBSource)
	addDefaultUser(config, store)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()
	waitGroup, ctx := errgroup.WithContext(ctx)

	switch config.RunMode {
	case "backend":
		logger.Log.Info("[System] Running in backend mode, skipping collector and mail alerter...")
		runGinServer(waitGroup, ctx, config, store)
	case "worker":
		logger.Log.Info("[System] Running in worker mode, skipping API server...")
		go runCollector(waitGroup, ctx, config, store)
		runTaskScheduler(waitGroup, ctx, redisOpt)
		runTaskProcessor(waitGroup, ctx, config, redisOpt, store, taskDistributor)
	default:
		logger.Log.Info("[System] Running in full mode, starting API server, collector and mail alerter...")
		go runCollector(waitGroup, ctx, config, store)
		runTaskScheduler(waitGroup, ctx, redisOpt)
		runTaskProcessor(waitGroup, ctx, config, redisOpt, store, taskDistributor)
		runGinServer(waitGroup, ctx, config, store)
	}

	err = waitGroup.Wait()
	if err != nil {
		logger.Log.Fatal("[System] Error from wait group: ", zap.Error(err))
	}

}

func runGinServer(waitGroup *errgroup.Group, ctx context.Context, config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		logger.Log.Fatal("[Server] Cannot create server:", zap.Error(err))
	}

	waitGroup.Go(func() error {
		logger.Log.Info("[Server] API server started successfully...")
		msg := fmt.Sprintf("[Server] API Server is running on %s ...", config.HTTPServerAddress)
		logger.Log.Info(msg)

		err := server.Start(config.HTTPServerAddress)
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			logger.Log.Error("[Server] Cannot start server:", zap.Error(err))
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()

		logger.Log.Info("[Server] Graceful shutdown API server...")
		err := server.Shutdown(ctx)
		if err != nil {
			logger.Log.Error("[Server] Cannot shutdown API server:", zap.Error(err))
			return err
		}
		logger.Log.Info("[Server] API server stopped successfully")
		return nil
	})

}

func runCollector(waitGroup *errgroup.Group, ctx context.Context, config util.Config, store db.Store) {
	collector := collector.NewCollector(config, store)

	err := collector.Start()
	if err != nil {
		logger.Log.Fatal("[Collector] Cannot start collector", zap.Error(err))
	}
	logger.Log.Info("[Collector] Collector started successfully...")

	waitGroup.Go(func() error {
		<-ctx.Done()
		logger.Log.Info("[Collector] Graceful shutdown collector...")
		c_ctx := collector.Stop()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		select {
		case <-timeoutCtx.Done():
			logger.Log.Warn("[Collector] Collector cannot Gracefully shutdown, forced to abort:", zap.Error(timeoutCtx.Err()))
			return timeoutCtx.Err()
		case <-c_ctx.Done():
			logger.Log.Info("[Collector] Collector stopped successfully")
		}
		return nil
	})
}

func runTaskScheduler(
	waitGroup *errgroup.Group,
	ctx context.Context,
	redisOpt asynq.RedisClientOpt,
) {
	scheduler := worker.NewRedisTaskScheduler(redisOpt)
	err := scheduler.ScheduleDetectLowBalance()
	if err != nil {
		logger.Log.Fatal("[Scheduler] Failed to register scheduler", zap.Error(err))
	}

	logger.Log.Info("[Scheduler] Starting task scheduler...")
	err = scheduler.Start()
	if err != nil {
		logger.Log.Fatal("[Scheduler] Failed to start scheduler", zap.Error(err))
	}
	logger.Log.Info("[Scheduler] Task scheduler started successfully")

	waitGroup.Go(func() error {
		<-ctx.Done()
		logger.Log.Info("[Scheduler] Graceful shutdown scheduler...")
		scheduler.Shutdown()
		logger.Log.Info("[Scheduler] Scheduler stopped")
		return nil
	})
}

func runTaskProcessor(
	waitGroup *errgroup.Group,
	ctx context.Context,
	config util.Config,
	redisOpt asynq.RedisClientOpt,
	store db.Store,
	taskDistributor worker.TaskDistributor,
) {
	mailer := mail.NewQQmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	redisTaskDistributor, ok := taskDistributor.(*worker.RedisTaskDistributor)
	if !ok {
		logger.Log.Fatal("[Processor] TaskDistributor is not a RedisTaskDistributor")
		return
	}
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer, redisTaskDistributor)

	logger.Log.Info("[Processor] Starting task processor...")
	err := taskProcessor.Start()
	if err != nil {
		logger.Log.Fatal("[Processor] Cannot start task processor", zap.Error(err))
		return
	}
	logger.Log.Info("[Processor] Task processor started successfully")

	waitGroup.Go(func() error {
		<-ctx.Done()
		logger.Log.Info("[Processor] Graceful shutdown task processor...")
		taskProcessor.Shutdown()
		logger.Log.Info("[Processor] Task processor stopped successfully")
		return nil
	})
}

func addDefaultUser(config util.Config, store db.Store) {
	// 假如数据库中没有用户，我们就添加一个默认用户
	count, err := store.CountUsers(context.Background())
	if err != nil {
		logger.Log.Fatal("[System] Cannot count users", zap.Error(err))
	}

	hashPassword, err := util.HashPassword(config.Password)
	if err != nil {
		logger.Log.Fatal("[System] Cannot hash password", zap.Error(err))
	}

	if count == 0 {
		logger.Log.Info("[System] Trying to create a default user...")
		arg := db.CreateUserParams{
			Username:       config.Username,
			HashedPassword: hashPassword,
			FullName:       config.FullName,
			Email:          config.Email,
		}

		_, err := store.CreateUser(context.Background(), arg)
		if err != nil {
			logger.Log.Fatal("[System] Cannot create default user", zap.Error(err))
		} else {
			logger.Log.Info("[System] Default user created successfully",
				zap.String("username", config.Username),
				zap.String("password", config.Password),
				zap.String("email", config.Email),
				zap.String("full_name", config.FullName),
			)
		}
	}
}

// 这里是运行数据库迁移的代码
func runMigrate(migrationURL string, dbSource string) {
	// 1. 创建一个新的迁移实例
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		logger.Log.Fatal("[System] Cannot create new migration instance:", zap.Error(err))
	}

	// 2. 执行迁移
	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Log.Fatal("[System] Cannot run migration:", zap.Error(err))
	}

	logger.Log.Info("[System] DB migrated successfully")
}
