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
