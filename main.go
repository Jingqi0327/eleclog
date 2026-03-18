package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/Jingqi0327/eleclog/api"
	"github.com/Jingqi0327/eleclog/collector"
	db "github.com/Jingqi0327/eleclog/db/sqlc"
	_ "github.com/Jingqi0327/eleclog/testdata"
	"github.com/Jingqi0327/eleclog/util"
	_ "github.com/lib/pq"
	"github.com/golang-migrate/migrate/v4"
  	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config")
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	runMigrate(config.MigrationURL, config.DBSource)
	addDefaultUser(config, store)

	go runCollector(config, store)

	runGinServer(config, store)
	//testdata.Insert_data(store)
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

func runCollector(config util.Config, store db.Store) {
	collector := collector.NewCollector(config, store)

	//collector.RunNow()

	collector.Start()
}

func addDefaultUser(config util.Config, store db.Store) {
	// 假如数据库中没有用户，我们就添加一个默认用户
	count, err := store.CountUsers(context.Background())
	if err != nil {
		log.Printf("无法查询用户数量: %v", err)
		return
	}

	hashPassword, err := util.HashPassword(config.Password)
	if err != nil {
		log.Printf("无法哈希密码: %v", err)
		return
	}

	if count == 0 {
		arg := db.CreateUserParams{
			Username:       config.Username,
			HashedPassword: hashPassword,
			FullName:       config.FullName,
			Email:          config.Email,
		}

		_, err := store.CreateUser(context.Background(), arg)
		if err != nil {
			log.Printf("无法创建默认用户: %v", err)
		} else {
			log.Printf("默认用户已创建:\n Username: %s\n Password: %s", config.Username, config.Password)
		}
	}
}

// 这里是运行数据库迁移的代码
func runMigrate(migrationURL string, dbSource string) {
	// 1. 创建一个新的迁移实例
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal("cannot create new migration instance:", err)
	}

	// 2. 执行迁移
	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("cannot run migration:", err)
	}

	log.Println("db migrated successfully")
}