package project

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/gin-gonic/gin"
)

type Project struct {
	UserID    int64   `json:"user_id"`
	Name      string  `json:"name"`
	RepoURL   string  `json:"repo_url"`
	Framework *string `json:"framework"`
	Subdomain *string `json:"subdomain"`
}

func CreateProject(c *gin.Context) {

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
		body.Name, body.RepoURL, body.UserID, providedSubdomain,
	).Scan(&projectid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
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
