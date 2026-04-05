package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Rakshit788/VERCEL-CLONE/apps/worker/internals"
	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/Rakshit788/VERCEL-CLONE/packages/redis"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

func main() {
	sep := strings.Repeat("=", 50)
	fmt.Printf("%s\n", sep)
	fmt.Println("🚀 VERCEL CLONE WORKER SERVICE STARTING")
	fmt.Printf("%s\n\n", sep)

	// Initialize database and redis connections
	fmt.Println("[1/3] Initializing database connection...")
	db.InitDB("postgres://postgres:password@localhost:5432/vercel_clone?sslmode=disable")
	fmt.Println("     ✅ Database connected")

	fmt.Println("[2/3] Initializing Redis connection...")
	redis.InitRedis("localhost:6379")
	fmt.Println("     ✅ Redis connected")

	// Start HTTP health check server in a goroutine
	fmt.Println("[3/3] Starting health check server on :8081...")
	go startHealthCheckServer()

	// Small delay to let health server start
	time.Sleep(1 * time.Second)
	fmt.Println("     ✅ Health check server started")

	fmt.Printf("\n%s\n", sep)
	fmt.Println("🎯 STARTING ASYNQ WORKER...")
	fmt.Printf("%s\n\n", sep)

	// Start the main Asynq worker server (blocking call)
	startAsynqWorker()
}

// startHealthCheckServer runs a simple HTTP server for health checks
// doesn't interfere with the worker pool
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

// startAsynqWorker creates and runs the Asynq task queue worker
// This listens to Redis for tasks and processes them concurrently
func startAsynqWorker() {
	fmt.Println("🚀 Starting Asynq Worker Server...")

	// Create Asynq server with Redis connection
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379"},
		asynq.Config{
			// Concurrency: Max number of tasks processed concurrently
			Concurrency: 10,

			// Queue priorities/weights
			Queues: map[string]int{
				"default":  6, // 60% of workers
				"critical": 4, // 40% of workers (higher priority builds)
			},
		},
	)

	// Create a ServeMux to register task handlers
	mux := asynq.NewServeMux()

	// Register the build:project task handler
	mux.HandleFunc("build:project", internals.HandleBuildTask)

	fmt.Println("📋 Registered task handler: build:project")
	fmt.Println("⏳ Waiting for tasks from Redis queue...")
	fmt.Println("   Concurrency: 10 workers (6 default, 4 critical)")

	// Start the server (blocking until shutdown)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("❌ Worker error: %v", err)
	}
}
