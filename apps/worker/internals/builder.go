package internals

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/hibiken/asynq"
)

// ─────────────────────────────────────────────
//  Types
// ─────────────────────────────────────────────

type BuildPayload struct {
	DeploymentID int64 `json:"deployment_id"`
	ProjectID    int64 `json:"project_id"`
}

type ProjectDetails struct {
	RepoURL         string
	BuildCommand    sql.NullString
	OutputDirectory sql.NullString
	Subdomain       sql.NullString
}

// ─────────────────────────────────────────────
//  Logger helper  (prefixed, timestamped)
// ─────────────────────────────────────────────

func log(deployID int64, step, msg string) {
	ts := time.Now().Format("2006/01/02 15:04:05.000")
	fmt.Printf("[%s] [deploy=%d] [%s] %s\n", ts, deployID, step, msg)
}

func logf(deployID int64, step, format string, args ...interface{}) {
	log(deployID, step, fmt.Sprintf(format, args...))
}

// ─────────────────────────────────────────────
//  HandleBuild  — asynq task handler
// ─────────────────────────────────────────────

func HandleBuild(ctx context.Context, t *asynq.Task) error {
	fmt.Println("🔨 Build started")
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ PANIC in HandleBuild: %v\n", r)
		}
	}()

	// ── 0. Decode payload ────────────────────────────────────────────────────
	var payload BuildPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		fmt.Printf("[FATAL] [decode] failed to decode task payload: %v | raw: %s\n",
			err, string(t.Payload()))
		return fmt.Errorf("failed to decode payload: %w", err)
	}

	dID := payload.DeploymentID
	log(dID, "START", "========== HandleBuild entered ==========")
	logf(dID, "START", "DeploymentID=%d  ProjectID=%d", payload.DeploymentID, payload.ProjectID)

	// ── 1. Mark as "building" ────────────────────────────────────────────────
	log(dID, "DB", "updating status → building")
	if err := updateDeploymentStatus(dID, "building", "", ""); err != nil {
		logf(dID, "DB", "ERROR: failed to set status=building: %v", err)
		return fmt.Errorf("failed to update status to building: %w", err)
	}
	log(dID, "DB", "status=building  ✓")

	// ── 2. Fetch project details ─────────────────────────────────────────────
	logf(dID, "DB", "fetching project details for projectID=%d", payload.ProjectID)
	project, err := getProjectDetails(payload.ProjectID)
	if err != nil {
		logf(dID, "DB", "ERROR: getProjectDetails: %v", err)
		_ = updateDeploymentStatus(dID, "failed", err.Error(), "")
		return fmt.Errorf("failed to get project details: %w", err)
	}
	logf(dID, "DB", "project fetched → RepoURL=%q  BuildCmd=%q  OutDir=%q  Subdomain=%q",
		project.RepoURL,
		project.BuildCommand.String,
		project.OutputDirectory.String,
		project.Subdomain.String)

	// ── 3. Build Docker image ────────────────────────────────────────────────
	log(dID, "DOCKER", "starting docker build …")
	imageName, buildLogs, err := buildDockerImage(ctx, project, dID)
	if err != nil {
		logf(dID, "DOCKER", "ERROR: docker build failed: %v", err)
		logf(dID, "DOCKER", "--- docker output start ---\n%s\n--- docker output end ---", buildLogs)
		_ = updateDeploymentStatus(dID, "failed", buildLogs, "")
		return fmt.Errorf("docker build failed: %w", err)
	}
	logf(dID, "DOCKER", "image built successfully → imageName=%q", imageName)
	logf(dID, "DOCKER", "--- docker output start ---\n%s\n--- docker output end ---", buildLogs)

	// ── 4. Mark as "built" ───────────────────────────────────────────────────
	log(dID, "DB", "updating status → built")
	if err := updateDeploymentStatus(dID, "built", buildLogs, imageName); err != nil {
		logf(dID, "DB", "ERROR: failed to set status=built: %v", err)
		return fmt.Errorf("failed to update status to built: %w", err)
	}
	log(dID, "DB", "status=built  ✓")

	// ── 5. Enqueue deploy task ───────────────────────────────────────────────
	log(dID, "ENQUEUE", "enqueueing deploy:run task …")
	if err := enqueueDeploy(dID); err != nil {
		// Non-fatal: build succeeded; log and continue
		logf(dID, "ENQUEUE", "WARNING: failed to enqueue deploy task: %v", err)
	} else {
		log(dID, "ENQUEUE", "deploy:run task enqueued  ✓")
	}

	log(dID, "DONE", "========== HandleBuild completed successfully ==========")
	return nil
}

// ─────────────────────────────────────────────
//  buildDockerImage
// ─────────────────────────────────────────────

func buildDockerImage(ctx context.Context, project *ProjectDetails, deployID int64) (string, string, error) {
	log(deployID, "DOCKER", "─────────────────────────────────────────")
	log(deployID, "DOCKER", "Starting Docker image build process")
	log(deployID, "DOCKER", "─────────────────────────────────────────")

	imageTag := fmt.Sprintf("%d", time.Now().Unix())
	imageName := fmt.Sprintf("deploy-%d:%s", deployID, imageTag)

	logf(deployID, "DOCKER", "✓ Generated image tag: %q", imageName)
	logf(deployID, "DOCKER", "✓ Repository URL: %q", project.RepoURL)

	// Guard: make sure repoURL is not empty
	if project.RepoURL == "" {
		log(deployID, "DOCKER", "❌ ERROR: project.RepoURL is empty")
		return "", "", fmt.Errorf("project.RepoURL is empty — cannot build Docker image")
	}
	log(deployID, "DOCKER", "✓ RepoURL validation passed")

	// Wrap in a hard 10-minute timeout so it never hangs forever
	buildCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	log(deployID, "DOCKER", "✓ Context timeout set to 10 minutes")
	logf(deployID, "DOCKER", "↳ Executing: docker build -t %s %s", imageName, project.RepoURL)
	log(deployID, "DOCKER", "↳ Flags: --no-cache, --progress=plain")

	startTime := time.Now()

	cmd := exec.CommandContext(buildCtx,
		"docker", "build",
		"--no-cache",
		"--progress=plain", // full layer-by-layer output
		"-t", imageName,
		project.RepoURL,
	)

	// Capture combined stdout+stderr
	log(deployID, "DOCKER", "⏳ Waiting for docker build to complete...")
	output, err := cmd.CombinedOutput()
	logs := string(output)
	buildDuration := time.Since(startTime)

	logf(deployID, "DOCKER", "↳ Build duration: %v", buildDuration)
	logf(deployID, "DOCKER", "↳ Output length: %d bytes", len(logs))

	if buildCtx.Err() == context.DeadlineExceeded {
		log(deployID, "DOCKER", "❌ ERROR: Build timed out after 10 minutes")
		logf(deployID, "DOCKER", "❌ Context error: %v", buildCtx.Err())
		return "", logs, fmt.Errorf("docker build timed out after 10 minutes")
	}

	if err != nil {
		logf(deployID, "DOCKER", "❌ ERROR: Command execution failed")
		logf(deployID, "DOCKER", "  Error: %v", err)
		logf(deployID, "DOCKER", "  Process state: %v", cmd.ProcessState)
		logf(deployID, "DOCKER", "  Exit code: %d", cmd.ProcessState.ExitCode())
		log(deployID, "DOCKER", "--- Build output start ---")
		logf(deployID, "DOCKER", "%s", logs)
		log(deployID, "DOCKER", "--- Build output end ---")
		return "", logs, err
	}

	log(deployID, "DOCKER", "✅ Docker build completed successfully!")
	logf(deployID, "DOCKER", "✅ Image name: %s", imageName)
	logf(deployID, "DOCKER", "✅ Total build time: %v", buildDuration)
	log(deployID, "DOCKER", "--- Build output start ---")
	logf(deployID, "DOCKER", "%s", logs)
	log(deployID, "DOCKER", "--- Build output end ---")
	log(deployID, "DOCKER", "─────────────────────────────────────────")

	return imageName, logs, nil
}

// ─────────────────────────────────────────────
//  updateDeploymentStatus
// ─────────────────────────────────────────────

func updateDeploymentStatus(deployID int64, status, logs, image string) error {
	if db.Pool == nil {
		logf(deployID, "DB", "ERROR: db.Pool is nil!")
		return fmt.Errorf("database pool not initialized")
	}

	logf(deployID, "DB", "updateDeploymentStatus → status=%q  image=%q  logsLen=%d",
		status, image, len(logs))

	_, err := db.Pool.Exec(
		context.Background(),
		`UPDATE deployments
		 SET status = $1,
		     logs   = $2,
		     url    = CASE
		                 WHEN $4 <> '' THEN $4
		                 ELSE url
		              END
		 WHERE id = $3`,
		status, logs, deployID, image,
	)
	if err != nil {
		logf(deployID, "DB", "ERROR: UPDATE deployments failed: %v", err)
	}
	return err
}

// ─────────────────────────────────────────────
//  getProjectDetails
// ─────────────────────────────────────────────

func getProjectDetails(projectID int64) (*ProjectDetails, error) {
	if db.Pool == nil {
		logf(0, "DB", "ERROR: db.Pool is nil in getProjectDetails!")
		return nil, fmt.Errorf("database pool not initialized")
	}

	logf(0, "DB", "SELECT projects WHERE id=%d", projectID)

	var project ProjectDetails
	err := db.Pool.QueryRow(
		context.Background(),
		`SELECT repo_url, build_command, output_directory, subdomain
		 FROM projects
		 WHERE id = $1`,
		projectID,
	).Scan(
		&project.RepoURL,
		&project.BuildCommand,
		&project.OutputDirectory,
		&project.Subdomain,
	)
	if err != nil {
		logf(0, "DB", "ERROR: QueryRow projects id=%d: %v", projectID, err)
		return nil, err
	}

	logf(0, "DB", "getProjectDetails OK → %+v", project)
	return &project, nil
}

// ─────────────────────────────────────────────
//  enqueueDeploy
// ─────────────────────────────────────────────

func enqueueDeploy(deploymentID int64) error {
	redisAddr := os.Getenv("REDIS_ADDR")
	candidates := make([]string, 0, 3)
	if redisAddr != "" {
		for _, addr := range strings.Split(redisAddr, ",") {
			trimmed := strings.TrimSpace(addr)
			if trimmed != "" {
				candidates = append(candidates, trimmed)
			}
		}
	}
	if len(candidates) == 0 {
		candidates = append(candidates, "redis:6379", "localhost:6379")
	}
	logf(deploymentID, "ENQUEUE", "redis candidates: %v", candidates)

	payload, err := json.Marshal(map[string]interface{}{
		"deployment_id": deploymentID,
	})
	if err != nil {
		logf(deploymentID, "ENQUEUE", "ERROR: failed to marshal payload: %v", err)
		return err
	}

	task := asynq.NewTask(
		"deploy:run",
		payload,
		asynq.MaxRetry(3),
		asynq.Timeout(15*time.Minute),
		asynq.Unique(10*time.Minute), // prevents duplicate enqueue
	)

	var lastErr error
	for _, addr := range candidates {
		logf(deploymentID, "ENQUEUE", "connecting to Redis at %s", addr)
		client := asynq.NewClient(asynq.RedisClientOpt{Addr: addr, DialTimeout: 2 * time.Second})

		for attempt := 1; attempt <= 3; attempt++ {
			info, err := client.Enqueue(task)
			if err == nil {
				_ = client.Close()
				logf(deploymentID, "ENQUEUE", "task enqueued → id=%s  queue=%s  state=%s",
					info.ID, info.Queue, info.State)
				return nil
			}

			lastErr = err
			logf(deploymentID, "ENQUEUE", "ERROR: enqueue attempt %d via %s failed: %v", attempt, addr, err)
			time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
		}

		_ = client.Close()
	}

	return fmt.Errorf("failed to enqueue deploy task after trying %d redis address(es): %w", len(candidates), lastErr)
}
