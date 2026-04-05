package main

import (
	"log"

	"github.com/Rakshit788/VERCEL-CLONE/apps/api/internal/auth"
	"github.com/Rakshit788/VERCEL-CLONE/apps/api/internal/deployment"
	"github.com/Rakshit788/VERCEL-CLONE/apps/api/internal/project"
	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	tasks "github.com/Rakshit788/VERCEL-CLONE/packages/queue"
	"github.com/Rakshit788/VERCEL-CLONE/packages/redis"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env found")
	}

	db.InitDB("postgres://postgres:password@postgres:5432/vercel_clone?sslmode=disable")
	redis.InitRedis("redis:6379")

	tasks.InitQueue()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 🔥 auth routes
	r.GET("/auth/github/login", auth.GitHubLogin)
	r.GET("/auth/github/callback", auth.GitHubCallback)

	r.POST("/create-project", project.CreateProject)

	//  deployment routes
	r.POST("/create-deployment", deployment.CreateDeployment)
	r.GET("/deployments/:id/status", deployment.GetDeploymentStatus)

	r.Run(":8080")
}
