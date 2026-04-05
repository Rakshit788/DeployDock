package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitDB(databaseUrl string) {
	var err error

	Pool, err = pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		log.Fatal("Failed to create pool:", err)
	}

	err = Pool.Ping(context.Background())
	if err != nil {
		log.Fatal("DB not reachable:", err)
	}

	log.Println("✅ DB connected")
}
