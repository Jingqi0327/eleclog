package main

import (
	"context"

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
)

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

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		logger.Log.Fatal("cannot connect to db:", zap.Error(err))
	}

	store := db.NewStore(connPool)
	runMigrate(config.MigrationURL, config.DBSource)
	addDefaultUser(config, store)

	switch config.RunMode {
	case "backend":
		logger.Log.Info(">> Running in backend mode, skipping collector and mail alerter...")
		runGinServer(config, store)
		return
	case "worker":
		logger.Log.Info(">> Running in worker mode, skipping API server...")
		go runCollector(config, store)
		runMailAlerter(config, store)
		select {} // 阻塞主线程，保持Worker运行
	default:
		logger.Log.Info(">> Running in full mode, starting API server, collector and mail alerter...")
		go runCollector(config, store)
		go runMailAlerter(config, store)
		runGinServer(config, store)
	}

}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		logger.Log.Fatal("cannot create server:", zap.Error(err))
	}
	
	logger.Log.Info(">> API server started successfully...")
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		logger.Log.Fatal("cannot start server:", zap.Error(err))
	}
	logger.Log.Info(">> API server started successfully...")
}

func runCollector(config util.Config, store db.Store) {
	collector := collector.NewCollector(config, store)

	err := collector.Start()
	if err != nil {
		logger.Log.Fatal("cannot start collector", zap.Error(err))
	}

	logger.Log.Info(">> Collector started successfully...")

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

func runMailAlerter(config util.Config, store db.Store) {
	mailer := mail.NewQQmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	alerter := worker.NewMailAlerter(store, mailer)
	err := alerter.Start()
	if err != nil {
		logger.Log.Fatal("cannot start mail alerter", zap.Error(err))
	}

	logger.Log.Info(">> Mail alerter started successfully...")
}
