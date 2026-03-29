package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitDB() {
	databaseUrl := os.Getenv("DATABASE_URL")

	if databaseUrl == "" {
		log.Fatal("DATABASE_URL not set")
	}

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
