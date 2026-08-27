package project

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Rakshit788/VERCEL-CLONE/apps/api/internal/auth"
	"github.com/Rakshit788/VERCEL-CLONE/apps/api/internal/githubapi"
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

	webhookRegistered, webhookErr := registerGitHubWebhook(int64(userID), projectid, body.RepoURL)
	if webhookErr != nil {
		log.Printf("CreateProject: webhook registration skipped for project %d: %v", projectid, webhookErr)
	}

	resp := gin.H{
		"project_id":         projectid,
		"subdomain":          providedSubdomain,
		"webhook_registered": webhookRegistered,
	}
	if !webhookRegistered && webhookErr != nil {
		resp["webhook_error"] = webhookErr.Error()
	}

	c.JSON(http.StatusOK, resp)

}

// registerGitHubWebhook creates a push webhook on the project's GitHub repo
// so that every push triggers a new deployment automatically. It uses the
// access token captured during GitHub OAuth login. Failure here is
// non-fatal to project creation — the user can still trigger deploys
// manually via create-deployment.
func registerGitHubWebhook(userID, projectID int64, repoURL string) (bool, error) {
	var accessToken string
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COALESCE(github_access_token, '') FROM users WHERE id = $1`, userID,
	).Scan(&accessToken)
	if err != nil {
		return false, fmt.Errorf("looking up github access token: %w", err)
	}
	if accessToken == "" {
		return false, fmt.Errorf("user has no stored github access token (log in via GitHub OAuth first)")
	}

	owner, repo, err := githubapi.ParseOwnerRepo(repoURL)
	if err != nil {
		return false, err
	}

	secret, err := generateWebhookSecret()
	if err != nil {
		return false, fmt.Errorf("generating webhook secret: %w", err)
	}

	baseURL := strings.TrimRight(os.Getenv("WEBHOOK_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	callbackURL := baseURL + "/webhooks/github"

	hookID, err := githubapi.CreateWebhook(accessToken, owner, repo, callbackURL, secret)
	if err != nil {
		return false, fmt.Errorf("creating github webhook: %w", err)
	}

	_, err = db.Pool.Exec(context.Background(),
		`UPDATE projects SET webhook_secret = $1, github_hook_id = $2 WHERE id = $3`,
		secret, hookID, projectID,
	)
	if err != nil {
		return false, fmt.Errorf("saving webhook secret: %w", err)
	}

	return true, nil
}

func generateWebhookSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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
