package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"learn.bleckshiba/banking/util"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Failed to load config", err)
	}

	testDB, err = sql.Open(config.Database.Driver, config.Database.Uri)
	if err != nil {
		log.Fatal("Failed to connect to db", err.Error())
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
