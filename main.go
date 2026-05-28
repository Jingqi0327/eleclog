package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Jingqi0327/eleclog/api"
	"github.com/Jingqi0327/eleclog/cache"
	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/gapi"
	"github.com/Jingqi0327/eleclog/logger"
	"github.com/Jingqi0327/eleclog/mail"
	"github.com/Jingqi0327/eleclog/pb"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/Jingqi0327/eleclog/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
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

	logger.Log.Info("[System] Starting processes...")

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		logger.Log.Fatal("[System] Cannot connect to db:", zap.Error(err))
	}

	store := db.NewStore(connPool)
	runMigrate(config.MigrationURL, config.DBSource)
	initDefaultUser(config, store)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddress,
	})
	redisCache := cache.NewRedisCache(redisClient)

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()
	waitGroup, ctx := errgroup.WithContext(ctx)

	proxyClient, cleanupProxy, err := createProxyClient(config)
	if err != nil {
		logger.Log.Fatal("[System] Cannot create proxy client:", zap.Error(err))
	}
	defer cleanupProxy()

	switch config.RunMode {
	case "backend":
		logger.Log.Info("[System] Running in backend mode, skipping collector and mail alerter...")
		runGinServer(waitGroup, ctx, config, store, redisCache, proxyClient)
	case "worker":
		logger.Log.Info("[System] Running in worker mode, skipping API server...")
		runTaskScheduler(waitGroup, ctx, config, redisOpt)
		runTaskProcessor(waitGroup, ctx, config, redisOpt, store, taskDistributor, proxyClient)
	case "proxy":
		logger.Log.Info("[System] Running in proxy mode...")
		runGrpcServer(waitGroup, ctx, config)
	case "main":
		logger.Log.Info("[System] Running in main mode...")
		runTaskScheduler(waitGroup, ctx, config, redisOpt)
		runTaskProcessor(waitGroup, ctx, config, redisOpt, store, taskDistributor, proxyClient)
		runGinServer(waitGroup, ctx, config, store, redisCache, proxyClient)
	default:
		logger.Log.Info("[System] Running in full mode, starting API server, mail alerter...")
		runTaskScheduler(waitGroup, ctx, config, redisOpt)
		runTaskProcessor(waitGroup, ctx, config, redisOpt, store, taskDistributor, proxyClient)
		runGinServer(waitGroup, ctx, config, store, redisCache, proxyClient)
		runGrpcServer(waitGroup, ctx, config)
	}

	err = waitGroup.Wait()
	if err != nil {
		logger.Log.Fatal("[System] Error from wait group: ", zap.Error(err))
	}
	logger.Log.Info("[System] All processes have stopped")
}

func runGinServer(waitGroup *errgroup.Group, ctx context.Context, config util.Config, store db.Store, redisCache cache.Cache, proxyClient pb.ProxyServiceClient) {
	server, err := api.NewServer(config, store, redisCache, proxyClient)
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

func runTaskScheduler(
	waitGroup *errgroup.Group,
	ctx context.Context,
	config util.Config,
	redisOpt asynq.RedisClientOpt,
) {
	scheduler := worker.NewRedisTaskScheduler(redisOpt)
	err := scheduler.ScheduleDetectLowBalance(config.DetectLowBalanceCron)
	if err != nil {
		logger.Log.Fatal("[Scheduler] Failed to register scheduler", zap.Error(err))
	}

	err = scheduler.ScheduleTaskSendFetchSurplusTasks(config.FetchSurplusCron)
	if err != nil {
		logger.Log.Fatal("[Scheduler] Failed to register fetch surplus scheduler", zap.Error(err))
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
	proxyClient pb.ProxyServiceClient,
) {
	mailer := mail.NewQQmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	redisTaskDistributor, ok := taskDistributor.(*worker.RedisTaskDistributor)
	if !ok {
		logger.Log.Fatal("[Processor] TaskDistributor is not a RedisTaskDistributor")
		return
	}
	taskProcessor, err := worker.NewRedisTaskProcessor(redisOpt, store, mailer, redisTaskDistributor, config, proxyClient)
	if err != nil {
		logger.Log.Fatal("[Processor] Cannot create task processor", zap.Error(err))
		return
	}

	logger.Log.Info("[Processor] Starting task processor...")
	err = taskProcessor.Start()
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

func runGrpcServer(waitGroup *errgroup.Group, ctx context.Context, config util.Config) {
	server, err := gapi.NewServer(config)
	if err != nil {
		logger.Log.Fatal("[GRPC] Cannot create server:", zap.Error(err))
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)

	pb.RegisterProxyServiceServer(grpcServer, server)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GrpcServerAddress)
	if err != nil {
		logger.Log.Fatal("[GRPC] Cannot create listener", zap.Error(err))
	}

	// 6. 启动服务器（阻塞调用）
	waitGroup.Go(func() error {
		logger.Log.Info("[GRPC] Starting gRPC server...")

		err = grpcServer.Serve(listener)
		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}
			logger.Log.Fatal("[GRPC] Cannot serve gRPC server", zap.Error(err))
			return err
		}

		return nil
	})

	// 7. 等待停机信号，并优雅地关闭服务器
	waitGroup.Go(func() error {
		<-ctx.Done()
		logger.Log.Info("[GRPC] Graceful shutdown gRPC server...")
		grpcServer.GracefulStop()
		logger.Log.Info("[GRPC] gRPC server stopped successfully")
		return nil
	})

}

// createProxyClient 辅助函数，返回 client 和用于释放连接的 cleanup 函数
func createProxyClient(config util.Config) (pb.ProxyServiceClient, func(), error) {
	if config.RunMode == "proxy" {
		return nil, func() {}, nil
	}
	conn, err := grpc.NewClient(
		config.ProxygRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}

	client := pb.NewProxyServiceClient(conn)

	cleanup := func() {
		logger.Log.Info("[GRPC] Closing proxy client connection...")
		conn.Close()
	}

	return client, cleanup, nil
}

func initDefaultUser(config util.Config, store db.Store) {
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
			Role:           util.AdminRole,
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
