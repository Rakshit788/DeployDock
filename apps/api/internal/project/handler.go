package project

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Rakshit788/VERCEL-CLONE/apps/api/internal/auth"
	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/gin-gonic/gin"
)

type Project struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	RepoURL   string    `json:"repo_url"`
	Framework *string   `json:"framework"`
	Subdomain *string   `json:"subdomain"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateProject(c *gin.Context) {

	// Extract user_id from JWT
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var body Project

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var projectid int64

	providedSubdomain := ""
	if body.Subdomain != nil {
		providedSubdomain = sanitizeSubdomain(*body.Subdomain)
	}

	err := db.Pool.QueryRow(
		c.Request.Context(),
		`INSERT INTO projects (name, repo_url, user_id, subdomain)
		 VALUES ($1, $2, $3, NULLIF($4, ''))
		 RETURNING id`,
		body.Name, body.RepoURL, userID, providedSubdomain,
	).Scan(&projectid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed", "details": err.Error()})
		return
	}

	if providedSubdomain == "" {
		generated := fmt.Sprintf("%s-%d", sanitizeSubdomain(body.Name), projectid)

		_, err = db.Pool.Exec(
			c.Request.Context(),
			`UPDATE projects SET subdomain=$1 WHERE id=$2`,
			generated, projectid,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set project subdomain"})
			return
		}
		providedSubdomain = generated
	}

	c.JSON(http.StatusOK, gin.H{
		"project_id": projectid,
		"subdomain":  providedSubdomain,
	})

}

func sanitizeSubdomain(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	if s == "" {
		return "project"
	}

	re := regexp.MustCompile(`[^a-z0-9-]`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	if s == "" {
		return "project"
	}
	return s
}

func ListProjects(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	rows, err := db.Pool.Query(
		c.Request.Context(),
		`SELECT id, user_id, name, repo_url, framework, subdomain, created_at
		 FROM projects
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db query failed"})
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.RepoURL, &p.Framework, &p.Subdomain, &p.CreatedAt)
		if err != nil {
			continue
		}
		projects = append(projects, p)
	}

	if projects == nil {
		projects = []Project{}
	}

	c.JSON(http.StatusOK, projects)
}
