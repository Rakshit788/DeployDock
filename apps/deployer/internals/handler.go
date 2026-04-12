package internals

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/hibiken/asynq"
)

// ─────────────────────────────────────────
//  Logger helper  (prefixed, timestamped)
// ─────────────────────────────────────────

func log(deployID int64, step, msg string) {
	ts := time.Now().Format("2006/01/02 15:04:05.000")
	fmt.Printf("[%s] [deploy=%d] [%s] %s\n", ts, deployID, step, msg)
}

func logf(deployID int64, step, format string, args ...interface{}) {
	log(deployID, step, fmt.Sprintf(format, args...))
}

type DeployPayload struct {
	DeploymentID int64 `json:"deployment_id"`
}

func HandleDeploy(ctx context.Context, t *asynq.Task) error {
	log(0, "DEPLOY", "─────────────────────────────────────────")
	log(0, "DEPLOY", "HandleDeploy task received")
	log(0, "DEPLOY", "─────────────────────────────────────────")

	var payload DeployPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log(0, "DECODE", "❌ ERROR: failed to decode payload")
		logf(0, "DECODE", "  Error: %v", err)
		logf(0, "DECODE", "  Raw payload: %s", string(t.Payload()))
		return fmt.Errorf("failed to decode payload: %w", err)
	}

	dID := payload.DeploymentID
	logf(dID, "START", "========== HandleDeploy entered ==========")
	logf(dID, "START", "DeploymentID=%d", dID)

	// ── 1. Mark as "deploying" ────────────────────────────────────────
	log(dID, "DB", "updating status → deploying")
	if err := updateDeploymentStatus(dID, "deploying", ""); err != nil {
		logf(dID, "DB", "❌ ERROR: failed to set status=deploying: %v", err)
		return fmt.Errorf("failed to update status to deploying: %w", err)
	}
	log(dID, "DB", "status=deploying  ✓")

	// ── 2. Fetch deployment details ────────────────────────────────────
	logf(dID, "DB", "fetching deployment details for deploymentID=%d", dID)
	details, err := getDeploymentDetails(dID)
	if err != nil {
		logf(dID, "DB", "❌ ERROR: getDeploymentDetails: %v", err)
		_ = updateDeploymentStatus(dID, "failed", fmt.Sprintf("failed to get deployment details: %v", err))
		return fmt.Errorf("failed to get deployment details: %w", err)
	}
	if details.ImageURL == "" {
		err := fmt.Errorf("deployment image is empty in deployments.url")
		logf(dID, "DB", "❌ ERROR: %v", err)
		_ = updateDeploymentStatus(dID, "failed", err.Error())
		return err
	}
	if details.Subdomain == "" {
		err := fmt.Errorf("project subdomain is empty")
		logf(dID, "DB", "❌ ERROR: %v", err)
		_ = updateDeploymentStatus(dID, "failed", err.Error())
		return err
	}
	logf(dID, "DB", "✓ Deployment details fetched → ImageURL=%q  Subdomain=%q", details.ImageURL, details.Subdomain)

	// ── 3. Run Docker container ────────────────────────────────────────
	log(dID, "DOCKER", "starting Docker container…")
	startTime := time.Now()
	port, containerID, err := runContainer(details.ImageURL, details.Subdomain)
	duration := time.Since(startTime)

	if err != nil {
		logf(dID, "DOCKER", "❌ ERROR: runContainer failed: %v", err)
		logf(dID, "DOCKER", "  Duration: %v", duration)
		_ = updateDeploymentStatus(dID, "failed", fmt.Sprintf("failed to run container: %v", err))
		return fmt.Errorf("failed to run container: %w", err)
	}
	logf(dID, "DOCKER", "✓ Container started successfully")
	logf(dID, "DOCKER", "✓ Container ID: %s", containerID)
	logf(dID, "DOCKER", "✓ Port: %d", port)
	logf(dID, "DOCKER", "✓ Duration: %v", duration)

	// ── 4. Save container info ─────────────────────────────────────────
	log(dID, "DB", "saving container info…")
	if err := saveContainerInfo(dID, containerID, port); err != nil {
		logf(dID, "DB", "❌ ERROR: failed to save container info: %v", err)
		return fmt.Errorf("failed to save container info: %w", err)
	}
	log(dID, "DB", "✓ Container info saved")

	liveURL := fmt.Sprintf("http://%s.vercel-clone.local", details.Subdomain)
	logf(dID, "DB", "updating live URL → %s", liveURL)
	if err := updateDeploymentURL(dID, liveURL); err != nil {
		logf(dID, "DB", "⚠️  WARNING: failed to update live URL: %v", err)
	}

	// ── 5. Mark as "deployed" ─────────────────────────────────────────
	log(dID, "DB", "updating status → deployed")
	if err := updateDeploymentStatus(dID, "deployed", ""); err != nil {
		logf(dID, "DB", "⚠️  WARNING: failed to update status to deployed: %v", err)
	} else {
		log(dID, "DB", "status=deployed  ✓")
	}

	logf(dID, "SUCCESS", "✅ Deployment %d is live at http://%s.vercel-clone.local:%d", dID, details.Subdomain, port)
	logf(dID, "DONE", "========== HandleDeploy completed successfully ==========")
	return nil

}

func updateDeploymentStatus(deployID int64, status string, logs string) error {
	logf(deployID, "DB", "updateDeploymentStatus → status=%q  logsLen=%d", status, len(logs))

	if db.Pool == nil {
		logf(deployID, "DB", "❌ ERROR: db.Pool is nil!")
		return fmt.Errorf("database pool not initialized")
	}

	_, err := db.Pool.Exec(
		context.Background(),
		`UPDATE deployments SET status=$1, logs=$2 WHERE id=$3`,
		status, logs, deployID,
	)

	if err != nil {
		logf(deployID, "DB", "❌ ERROR: UPDATE deployments failed: %v", err)
		return err
	}

	logf(deployID, "DB", "✓ Status updated to %q", status)
	return nil
}

func saveContainerInfo(deployID int64, containerID string, port int) error {
	logf(deployID, "DB", "saveContainerInfo → containerID=%s  port=%d", containerID, port)

	if db.Pool == nil {
		logf(deployID, "DB", "❌ ERROR: db.Pool is nil!")
		return fmt.Errorf("database pool not initialized")
	}

	_, err := db.Pool.Exec(
		context.Background(),
		`UPDATE deployments
		 SET logs = COALESCE(logs, '') || $1
		 WHERE id = $2`,
		fmt.Sprintf("\ncontainer_id=%s port=%d", containerID, port), deployID,
	)

	if err != nil {
		logf(deployID, "DB", "❌ ERROR: UPDATE deployments (container_info) failed: %v", err)
		return err
	}

	logf(deployID, "DB", "✓ Container info saved")
	return nil
}

func updateDeploymentURL(deployID int64, liveURL string) error {
	if db.Pool == nil {
		return fmt.Errorf("database pool not initialized")
	}

	_, err := db.Pool.Exec(
		context.Background(),
		`UPDATE deployments SET url=$1 WHERE id=$2`,
		liveURL, deployID,
	)
	return err
}
