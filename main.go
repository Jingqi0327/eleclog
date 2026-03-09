package main

import (
	"database/sql"
	"log"

	"github.com/Jingqi0327/eleclog/api"
	"github.com/Jingqi0327/eleclog/collector"
	db "github.com/Jingqi0327/eleclog/db/sqlc"
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

	go runCollector(config,store)

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

	collector.RunNow()

}
