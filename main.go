package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"learn.bleckshiba/banking/api"
	db "learn.bleckshiba/banking/db/sqlc"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:shiba@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:5000"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	log.Fatal(server.Start(serverAddress))
}
