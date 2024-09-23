package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"learn.bleckshiba/banking/api"
	db "learn.bleckshiba/banking/db/sqlc"
	"learn.bleckshiba/banking/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	conn, err := sql.Open(config.Database.Driver, config.Database.Uri)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	log.Fatal(server.Start(fmt.Sprintf(":%s", config.App.Port)))
}
