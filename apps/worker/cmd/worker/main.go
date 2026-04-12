package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Rakshit788/VERCEL-CLONE/apps/worker/internals"
	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/Rakshit788/VERCEL-CLONE/packages/redis"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

func main() {
	fmt.Println("🚀 VERCEL CLONE WORKER SERVICE STARTING")

	fmt.Println("[1/3] Initializing database connection...")
	db.InitDB("postgres://postgres:password@localhost:5432/vercel_clone?sslmode=disable")
	fmt.Println("     ✅ Database connected")

	fmt.Println("[2/3] Initializing Redis connection...")
	redis.InitRedis("localhost:6379")
	fmt.Println("     ✅ Redis connected")

	fmt.Println("[3/3] Starting health check server on :8081...")
	go startHealthCheckServer()

	time.Sleep(1 * time.Second)
	fmt.Println("     ✅ Health check server started")

	fmt.Println("🎯 STARTING ASYNQ WORKER...")
	startAsynqWorker()
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

func startAsynqWorker() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ PANIC in startAsynqWorker: %v\n", r)
		}
	}()

	fmt.Println("🚀 Starting Asynq Worker Server...")

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379"},
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
