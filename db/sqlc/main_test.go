package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/Jingqi0327/eleclog/util"
	_ "github.com/lib/pq"
)

var testStore Store

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.TestDBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testStore = NewStore(conn)
	os.Exit(m.Run())
}
