package project

import (
	"net/http"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/gin-gonic/gin"
)

// ✅ Proper struct
type Project struct {
	UserID    int64   `json:"user_id"`
	Name      string  `json:"name"`
	RepoURL   string  `json:"repo_url"`
	Framework *string `json:"framework"`
	Subdomain *string `json:"subdomain"`
}

func CreateProject(c *gin.Context) {

	var body Project

	// ✅ Bind JSON
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var projectid int

	err := db.Pool.QueryRow(
		c.Request.Context(),
		`INSERT INTO projects (name, repo_url, user_id)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		body.Name, body.RepoURL, body.UserID,
	).Scan(&projectid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project_id": projectid,
	})
}
