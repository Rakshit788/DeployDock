package main

import (
	"log"
	"os"

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

	db.InitDB(getenv("DATABASE_URL", "postgres://postgres:password@postgres:5432/vercel_clone?sslmode=disable"))

	redisAddr := getenv("REDIS_ADDR", "redis:6379")
	redis.InitRedis(redisAddr)

	tasks.InitQueue(redisAddr)

	r := gin.Default()

	// Allow frontend app to call API from browser during local development.
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// auth routes
	r.GET("/auth/github/login", auth.GitHubLogin)
	r.GET("/auth/github/callback", auth.GitHubCallback)
	r.POST("/api/auth/github", auth.GitHubAuthAPI)
	r.GET("/api/auth/user", auth.AuthMiddleware(), auth.GetUser)
	r.GET("/api/test/token", auth.GenerateTestToken) // Development only

	// Protected routes (require authentication)
	r.POST("/create-project", auth.AuthMiddleware(), project.CreateProject)
	r.GET("/api/project/list", auth.AuthMiddleware(), project.ListProjects)

	//  deployment routes - temporarily without auth for testing
	r.POST("/create-deployment", auth.AuthMiddleware(), deployment.CreateDeployment)
	r.GET("/deployments/:id/status", auth.AuthMiddleware(), deployment.GetDeploymentStatus)
	r.GET("/api/deployment/list", auth.AuthMiddleware(), deployment.ListDeployments)

	r.Run(":8080")
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
