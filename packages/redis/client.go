package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var Ctx = context.Background()

func InitRedis(addr string) {
	RDB = redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})

	// test connection
	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("❌ Redis not reachable:", err)
	}

	log.Println("✅ Redis connected")
}
