package main

import (
	"fmt"

	"github.com/Rakshit788/VERCEL-CLONE/apps/deployer/internals"
	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/Rakshit788/VERCEL-CLONE/packages/redis"
	"github.com/hibiken/asynq"
)

func main() {
	fmt.Println("VERCEL CLONE DEPLOYER SERVICE STARTING")

	fmt.Println("Initializing database connection...")
	db.InitDB("postgres://postgres:password@postgres:5432/vercel_clone?sslmode=disable")
	fmt.Println("Database connected")

	redis.InitRedis("redis:6379")
	fmt.Println("Redis connected")

	fmt.Println("Deployer service is running...")
	runDeployer()
}

func runDeployer() {
	fmt.Println("DEPLOYER SETUP")

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "redis:6379"},
		asynq.Config{
			Concurrency: 4,
			Queues: map[string]int{
				"default": 4,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("deploy:run", internals.HandleDeploy)
	fmt.Println("Task handler registered: deploy:run -> HandleDeploy")
	fmt.Println("Starting deployer server...")

	if err := srv.Run(mux); err != nil {
		fmt.Printf("Deployer server error: %v\n", err)
	}
}
