// Package webhook receives GitHub push events and turns them into
// deployments, so every push to a connected repo is deployed automatically.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	tasks "github.com/Rakshit788/VERCEL-CLONE/packages/queue"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type pushPayload struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	} `json:"head_commit"`
}

type candidateProject struct {
	ID     int64
	Secret string
}

// GitHubWebhook handles POST /webhooks/github. It has no auth middleware —
// GitHub can't send a JWT — instead every request must carry a valid
// X-Hub-Signature-256 HMAC computed with a project's stored webhook secret.
func GitHubWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	event := c.GetHeader("X-GitHub-Event")
	if event == "ping" {
		c.JSON(http.StatusOK, gin.H{"status": "pong"})
		return
	}
	if event != "push" {
		c.JSON(http.StatusOK, gin.H{"ignored": event})
		return
	}

	var payload pushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}
	if payload.Repository.FullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing repository.full_name"})
		return
	}

	candidates, err := findCandidateProjects(c, payload.Repository.FullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db query failed"})
		return
	}

	signature := c.GetHeader("X-Hub-Signature-256")
	project, ok := matchSignature(candidates, signature, body)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "signature verification failed"})
		return
	}

	// Only deploy branch pushes (skip tag pushes, branch deletes, etc.)
	if !strings.HasPrefix(payload.Ref, "refs/heads/") || payload.After == "" || strings.Trim(payload.After, "0") == "" {
		c.JSON(http.StatusOK, gin.H{"ignored": "not a branch push"})
		return
	}

	commitSHA := payload.After
	if payload.HeadCommit.ID != "" {
		commitSHA = payload.HeadCommit.ID
	}

	deploymentID, err := createDeployment(c, project.ID, commitSHA)
	if err != nil {
		log.Printf("webhook: failed to create deployment for project %d: %v", project.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create deployment"})
		return
	}

	if err := enqueueBuild(deploymentID, project.ID); err != nil {
		log.Printf("webhook: failed to enqueue build for deployment %d: %v", deploymentID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue build"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deployment_id": deploymentID,
		"status":        "queued",
		"commit_sha":    commitSHA,
	})
}

// findCandidateProjects returns every project whose repo_url references the
// pushed repo. Matching by full_name (owner/repo) rather than an exact URL
// string tolerates https/ssh/.git URL differences.
func findCandidateProjects(c *gin.Context, fullName string) ([]candidateProject, error) {
	rows, err := db.Pool.Query(c.Request.Context(),
		`SELECT id, COALESCE(webhook_secret, '')
		 FROM projects
		 WHERE webhook_secret IS NOT NULL
		   AND repo_url ILIKE '%' || $1 || '%'`,
		fullName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []candidateProject
	for rows.Next() {
		var p candidateProject
		if err := rows.Scan(&p.ID, &p.Secret); err != nil {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

// matchSignature checks the request's X-Hub-Signature-256 against each
// candidate project's secret and returns the one that matches.
func matchSignature(candidates []candidateProject, signature string, body []byte) (candidateProject, bool) {
	const prefix = "sha256="
	if !strings.HasPrefix(signature, prefix) {
		return candidateProject{}, false
	}
	given, err := hex.DecodeString(strings.TrimPrefix(signature, prefix))
	if err != nil {
		return candidateProject{}, false
	}

	for _, p := range candidates {
		if p.Secret == "" {
			continue
		}
		mac := hmac.New(sha256.New, []byte(p.Secret))
		mac.Write(body)
		expected := mac.Sum(nil)
		if hmac.Equal(expected, given) {
			return p, true
		}
	}
	return candidateProject{}, false
}

func createDeployment(c *gin.Context, projectID int64, commitSHA string) (int64, error) {
	var deploymentID int64
	err := db.Pool.QueryRow(c.Request.Context(),
		`INSERT INTO deployments (project_id, status, commit_sha)
		 VALUES ($1, 'pending', $2)
		 RETURNING id`,
		projectID, commitSHA,
	).Scan(&deploymentID)
	return deploymentID, err
}

func enqueueBuild(deploymentID, projectID int64) error {
	payload, err := json.Marshal(map[string]interface{}{
		"deployment_id": deploymentID,
		"project_id":    projectID,
	})
	if err != nil {
		return err
	}

	task := asynq.NewTask("build:project", payload)
	_, err = tasks.Queue.Enqueue(task)
	return err
}
