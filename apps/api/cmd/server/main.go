package main

import (
	"log"

	"github.com/Rakshit788/VERCEL-CLONE/apps/api/internal/auth"
	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env found")
	}

	db.InitDB()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 🔥 auth routes
	r.GET("/auth/github/login", auth.GitHubLogin)
	r.GET("/auth/github/callback", auth.GitHubCallback)

	r.Run(":8080")
}
