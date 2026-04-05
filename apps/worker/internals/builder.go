package internals

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/hibiken/asynq"
)

type BuildPayload struct {
	DeploymentID int64 `json:"deployment_id"`
	ProjectID    int64 `json:"project_id"`
}

type ProjectDetails struct {
	RepoURL         string
	BuildCommand    string
	OutputDirectory string
	Subdomain       string
}

// HandleBuildTask processes the build:project task from the Redis queue
// This function is called concurrently by multiple Asynq workers
func HandleBuildTask(ctx context.Context, t *asynq.Task) error {
	var payload BuildPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	fmt.Printf(" Starting build for deployment %d (project %d)\n", payload.DeploymentID, payload.ProjectID)

	// Update deployment status to "building"
	if err := updateDeploymentStatus(ctx, payload.DeploymentID, "building", ""); err != nil {
		fmt.Printf(" Failed to update status: %v\n", err)
		return err
	}

	// Fetch project details from database
	project, err := getProjectDetails(ctx, payload.ProjectID)
	if err != nil {
		logs := fmt.Sprintf(" Failed to fetch project: %v", err)
		_ = updateDeploymentStatus(ctx, payload.DeploymentID, "failed", logs)
		return err
	}

	fmt.Printf("Project Details:\n  └─ Repository: %s\n  └─ Build Cmd: %s\n  └─ Output Dir: %s\n",
		project.RepoURL, project.BuildCommand, project.OutputDirectory)

	logs, err := buildInDockerContainer(project, payload.DeploymentID)
	if err != nil {
		fmt.Printf(" Build failed: %v\n", err)
		_ = updateDeploymentStatus(ctx, payload.DeploymentID, "failed", logs)
		return err
	}

	// Generate deployment URL (using subdomain if provided)
	deploymentURL := fmt.Sprintf("https://%s.localhost", project.Subdomain)
	if project.Subdomain == "" {
		deploymentURL = fmt.Sprintf("https://deploy-%d.localhost", payload.DeploymentID)
	}

	fmt.Printf(" Build completed successfully!\n")
	fmt.Printf(" Deployment URL: %s\n", deploymentURL)

	// Update deployment status to "deployed" with the generated URL
	if err := updateDeploymentStatusWithURL(ctx, payload.DeploymentID, "deployed", deploymentURL, logs); err != nil {
		fmt.Printf("❌ Failed to update final status: %v\n", err)
		return err
	}

	return nil
}

// getProjectDetails fetches project information from database
func getProjectDetails(ctx context.Context, projectID int64) (*ProjectDetails, error) {
	var project ProjectDetails

	err := db.Pool.QueryRow(ctx,
		`SELECT repo_url, build_command, output_directory, subdomain 
		 FROM projects WHERE id = $1`,
		projectID,
	).Scan(&project.RepoURL, &project.BuildCommand, &project.OutputDirectory, &project.Subdomain)

	if err != nil {
		return nil, err
	}

	// Set defaults if empty
	if project.BuildCommand == "" {
		project.BuildCommand = "npm run build"
	}
	if project.OutputDirectory == "" {
		project.OutputDirectory = "dist"
	}

	return &project, nil
}

// buildInDockerContainer handles the complete build process:
// 1. Clone the repository
// 2. Check if Dockerfile exists
// 3. If Dockerfile exists, build using Docker
// 4. If no Dockerfile, fail gracefully with clear message
func buildInDockerContainer(project *ProjectDetails, deploymentID int64) (string, error) {
	var logOutput strings.Builder

	// Create temporary directory for cloning
	tmpDir := filepath.Join("/tmp", fmt.Sprintf("build_%d_%d", deploymentID, time.Now().Unix()))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Sprintf("Failed to create temp directory: %v", err), err
	}
	defer os.RemoveAll(tmpDir) // Cleanup after build

	repoDir := filepath.Join(tmpDir, "repo")

	fmt.Printf("📥 Cloning repository...\n")
	fmt.Printf("   └─ URL: %s\n", project.RepoURL)

	// Step 1: Clone the repository
	cloneCmd := exec.Command("git", "clone", project.RepoURL, repoDir)
	cloneOutput, err := cloneCmd.CombinedOutput()
	logOutput.WriteString(string(cloneOutput))
	logOutput.WriteString("\n")

	if err != nil {
		errMsg := fmt.Sprintf("Failed to clone repository: %v", err)
		logOutput.WriteString(errMsg)
		return logOutput.String(), err
	}

	fmt.Printf("Repository cloned successfully\n")

	// Step 2: Check if Dockerfile exists
	dockerfilePath := filepath.Join(repoDir, "Dockerfile")
	fileInfo, err := os.Stat(dockerfilePath)

	if err != nil || fileInfo.IsDir() {
		errMsg := fmt.Sprintf("❌ Dockerfile not found in repository root. Build requires a Dockerfile at the project root.")
		logOutput.WriteString("\n" + errMsg)
		fmt.Printf("%s\n", errMsg)
		return logOutput.String(), fmt.Errorf("no dockerfile found")
	}

	fmt.Printf("✅ Dockerfile found! Starting Docker build...\n")

	// Step 3: Build Docker image
	imageName := fmt.Sprintf("deploy-%d:%d", deploymentID, time.Now().Unix())

	fmt.Printf("🐳 Building Docker image: %s\n", imageName)
	fmt.Printf("   └─ Build context: %s\n", repoDir)

	// Execute: docker build -t IMAGE_NAME .
	buildCmd := exec.Command("docker", "build",
		"-t", imageName, // Tag the image
		"-f", dockerfilePath, // Explicit Dockerfile path
		repoDir, // Build context
	)

	buildOutput, err := buildCmd.CombinedOutput()
	logOutput.WriteString(string(buildOutput))
	logOutput.WriteString("\n")

	if err != nil {
		errMsg := fmt.Sprintf("Docker build failed: %v", err)
		logOutput.WriteString(errMsg)
		fmt.Printf("❌ %s\n", errMsg)
		return logOutput.String(), err
	}

	fmt.Printf("✅ Docker image built successfully: %s\n", imageName)

	// Step 4: Optionally run the image to verify it works
	fmt.Printf("🧪 Running container to verify build...\n")

	runCmd := exec.Command("docker", "run",
		"--rm",                 // Remove container after exit
		"-t",                   // Allocate TTY
		"--entrypoint", "echo", // Run echo command
		imageName,
		"Build verification: Container started successfully!",
	)

	runOutput, err := runCmd.CombinedOutput()
	logOutput.WriteString(string(runOutput))
	logOutput.WriteString("\n")

	if err != nil {
		warnMsg := fmt.Sprintf("Warning: Container test run failed (non-critical): %v", err)
		logOutput.WriteString(warnMsg)
		fmt.Printf("  %s\n", warnMsg)
		// Don't fail here - image build succeeded, runtime issues are separate
	} else {
		fmt.Printf("✅ Container verification passed!\n")
	}

	return logOutput.String(), nil
}

// updateDeploymentStatus updates the deployment status in the database
func updateDeploymentStatus(ctx context.Context, deploymentID int64, status, logs string) error {
	query := `UPDATE deployments SET status = $1, logs = $2, started_at = COALESCE(started_at, CURRENT_TIMESTAMP) WHERE id = $3`

	_, err := db.Pool.Exec(ctx, query, status, logs, deploymentID)
	return err
}

// updateDeploymentStatusWithURL updates deployment with status, URL, logs, and finished time
func updateDeploymentStatusWithURL(ctx context.Context, deploymentID int64, status, url, logs string) error {
	query := `UPDATE deployments 
	          SET status = $1, url = $2, logs = $3, started_at = COALESCE(started_at, CURRENT_TIMESTAMP), finished_at = CURRENT_TIMESTAMP 
	          WHERE id = $4`

	_, err := db.Pool.Exec(ctx, query, status, url, logs, deploymentID)
	return err
}
