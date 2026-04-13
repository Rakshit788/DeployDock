package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Rakshit788/VERCEL-CLONE/apps/worker/internals"
	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/Rakshit788/VERCEL-CLONE/packages/redis"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

func main() {
	fmt.Println("🚀 VERCEL CLONE WORKER SERVICE STARTING")

	postgresDSN := getenv("WORKER_DB_DSN", "postgres://postgres:password@postgres:5432/vercel_clone?sslmode=disable")
	redisAddr := getenv("REDIS_ADDR", "redis:6379")

	fmt.Println("[1/3] Initializing database connection...")
	db.InitDB(postgresDSN)
	fmt.Println("     ✅ Database connected")

	fmt.Println("[2/3] Initializing Redis connection...")
	redis.InitRedis(redisAddr)
	fmt.Println("     ✅ Redis connected")

	fmt.Println("[3/3] Starting health check server on :8081...")
	go startHealthCheckServer()

	time.Sleep(1 * time.Second)
	fmt.Println("     ✅ Health check server started")

	fmt.Println("🎯 STARTING ASYNQ WORKER...")
	startAsynqWorker(redisAddr)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func startHealthCheckServer() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "worker",
		})
	})

	if err := r.Run(":8081"); err != nil {
		log.Printf("❌ Health server error: %v", err)
	}
}

func startAsynqWorker(redisAddr string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ PANIC in startAsynqWorker: %v\n", r)
		}
	}()

	fmt.Println("🚀 Starting Asynq Worker Server...")

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 4,
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("build:project", internals.HandleBuild)

	fmt.Println("📋 Registered task handler: build:project")
	fmt.Println("⏳ Waiting for tasks from Redis queue...")
	fmt.Println("   Concurrency: 4 workers")

	if err := srv.Run(mux); err != nil {
		log.Fatalf("❌ Worker error: %v", err)
	}
}
