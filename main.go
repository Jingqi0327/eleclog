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

	logger.Log.Info(">> Starting server...")

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		logger.Log.Fatal("cannot connect to db:", zap.Error(err))
	}

	store := db.NewStore(connPool)
	runMigrate(config.MigrationURL, config.DBSource)
	addDefaultUser(config, store)

	waitGroup, ctx := errgroup.WithContext(ctx)

	switch config.RunMode {
	case "backend":
		logger.Log.Info(">> Running in backend mode, skipping collector and mail alerter...")
		runGinServer(waitGroup, ctx, config, store)
	case "worker":
		logger.Log.Info(">> Running in worker mode, skipping API server...")
		go runCollector(waitGroup, ctx, config, store)
		runMailAlerter(waitGroup, ctx, config, store)
		select {} // 阻塞主线程，保持Worker运行
	default:
		logger.Log.Info(">> Running in full mode, starting API server, collector and mail alerter...")
		go runCollector(waitGroup, ctx, config, store)
		go runMailAlerter(waitGroup, ctx, config, store)
		runGinServer(waitGroup, ctx, config, store)
	}

	err = waitGroup.Wait()
	if err != nil {
		logger.Log.Fatal(">> Error from wait group: ", zap.Error(err))
	}

}

func runGinServer(waitGroup *errgroup.Group, ctx context.Context, config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		logger.Log.Fatal("cannot create server:", zap.Error(err))
	}

	waitGroup.Go(func() error {
		logger.Log.Info(">> API server started successfully...")
		msg := fmt.Sprintf(">> API Server is running on %s ...", config.HTTPServerAddress)
		logger.Log.Info(msg)

		err := server.Start(config.HTTPServerAddress)
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			logger.Log.Error("cannot start server:", zap.Error(err))
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()

		logger.Log.Info(">> Graceful shutdown API server...")
		err := server.Shutdown(ctx)
		if err != nil {
			logger.Log.Error(">> Cannot shutdown API server:", zap.Error(err))
			return err
		}
		logger.Log.Info(">> API server stopped successfully")
		return nil
	})

}

func runCollector(waitGroup *errgroup.Group, ctx context.Context, config util.Config, store db.Store) {
	collector := collector.NewCollector(config, store)

	err := collector.Start()
	if err != nil {
		logger.Log.Fatal("cannot start collector", zap.Error(err))
	}
	logger.Log.Info(">> Collector started successfully...")

	waitGroup.Go(func() error {
		<-ctx.Done()
		logger.Log.Info(">> Graceful shutdown collector...")
		c_ctx := collector.Stop()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		select {
		case <-timeoutCtx.Done():
			logger.Log.Warn(">> Collector cannot Gracefully shutdown, forced to abort:", zap.Error(timeoutCtx.Err()))
			return timeoutCtx.Err()
		case <-c_ctx.Done():
			logger.Log.Info(">> Collector stopped successfully")
		}
		return nil
	})
}

func runMailAlerter(waitGroup *errgroup.Group, ctx context.Context, config util.Config, store db.Store) {
	mailer := mail.NewQQmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	alerter := worker.NewMailAlerter(store, mailer)

	err := alerter.Start()
	if err != nil {
		logger.Log.Fatal("cannot start mail alerter", zap.Error(err))
	}
	logger.Log.Info(">> Mail alerter started successfully...")

	waitGroup.Go(func() error {
		<-ctx.Done()
		logger.Log.Info(">> Graceful shutdown mail alerter...")
		m_ctx := alerter.Stop()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		select {
		case <-timeoutCtx.Done():
			logger.Log.Warn(">> Mail alerter cannot Gracefully shutdown, forced to abort:", zap.Error(timeoutCtx.Err()))
			return timeoutCtx.Err()
		case <-m_ctx.Done():
			logger.Log.Info(">> Mail alerter stopped successfully")
		}
		return nil
	})
}

func addDefaultUser(config util.Config, store db.Store) {
	// 假如数据库中没有用户，我们就添加一个默认用户
	count, err := store.CountUsers(context.Background())
	if err != nil {
		logger.Log.Fatal("cannot count users", zap.Error(err))
	}

	hashPassword, err := util.HashPassword(config.Password)
	if err != nil {
		logger.Log.Fatal("cannot hash password", zap.Error(err))
	}

	if count == 0 {
		logger.Log.Info(">> trying to create a default user...")
		arg := db.CreateUserParams{
			Username:       config.Username,
			HashedPassword: hashPassword,
			FullName:       config.FullName,
			Email:          config.Email,
		}

		_, err := store.CreateUser(context.Background(), arg)
		if err != nil {
			logger.Log.Fatal(">> cannot create default user", zap.Error(err))
		} else {
			logger.Log.Info(">> default user created successfully",
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
		logger.Log.Fatal("cannot create new migration instance:", zap.Error(err))
	}

	// 2. 执行迁移
	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Log.Fatal("cannot run migration:", zap.Error(err))
	}

	logger.Log.Info(">> db migrated successfully")
}
