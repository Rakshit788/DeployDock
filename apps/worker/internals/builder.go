package internals

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
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

// HandleBuild is the asynq task handler for "build:project"
func HandleBuild(ctx context.Context, t *asynq.Task) error {
	// Decode the raw task payload
	var payload BuildPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}

	fmt.Printf(" Build started for deployment %d\n", payload.DeploymentID)

	// 1. Update status → building
	if err := updateDeploymentStatus(payload.DeploymentID, "building", "", ""); err != nil {
		return fmt.Errorf("failed to update status to building: %w", err)
	}

	// 2. Fetch project details
	project, err := getProjectDetails(payload.ProjectID)
	if err != nil {
		updateDeploymentStatus(payload.DeploymentID, "failed", err.Error(), "")
		return fmt.Errorf("failed to get project details: %w", err)
	}

	// 3. Build Docker image
	imageName, logs, err := buildDockerImage(project, payload.DeploymentID)
	if err != nil {
		updateDeploymentStatus(payload.DeploymentID, "failed", logs, "")
		return fmt.Errorf("docker build failed: %w", err)
	}

	// 4. Update status → built with image name and logs
	if err := updateDeploymentStatus(payload.DeploymentID, "built", logs, imageName); err != nil {
		return fmt.Errorf("failed to update status to built: %w", err)
	}

	// 5. Enqueue the deploy task
	if err := enqueueDeploy(payload.DeploymentID); err != nil {
		// Not a fatal error — build succeeded, log and continue
		fmt.Printf(" Failed to enqueue deploy for deployment %d: %v\n", payload.DeploymentID, err)
	}

	fmt.Printf(" Build completed successfully for deployment %d\n", payload.DeploymentID)
	return nil
}

func buildDockerImage(project *ProjectDetails, deployID int64) (string, string, error) {
	fmt.Println(" Building Docker image from remote repository...")

	imageTag := fmt.Sprintf("%d", time.Now().Unix())
	imageName := fmt.Sprintf("deploy-%d:%s", deployID, imageTag)

	cmd := exec.Command(
		"docker", "build",
		"-t", imageName,
		project.RepoURL,
	)

	output, err := cmd.CombinedOutput()
	logs := string(output)
	if err != nil {
		fmt.Printf(" Error building Docker image: %v\n", err)
		return "", logs, err
	}

	return imageName, logs, nil
}

func updateDeploymentStatus(deployID int64, status string, logs string, image string) error {
	_, err := db.Pool.Exec(
		context.Background(),
		`UPDATE deployments 
		 SET status=$1, logs=$2, image_url=$3 
		 WHERE id=$4`,
		status, logs, image, deployID,
	)
	return err
}

func getProjectDetails(projectID int64) (*ProjectDetails, error) {
	var project ProjectDetails

	err := db.Pool.QueryRow(
		context.Background(),
		`SELECT repo_url, build_command, output_directory, subdomain 
		 FROM projects WHERE id = $1`,
		projectID,
	).Scan(
		&project.RepoURL,
		&project.BuildCommand,
		&project.OutputDirectory,
		&project.Subdomain,
	)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func enqueueDeploy(deploymentID int64) error {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	defer client.Close()

	payload, err := json.Marshal(map[string]interface{}{
		"deployment_id": deploymentID,
	})
	if err != nil {
		return err
	}

	task := asynq.NewTask("deploy:run", payload)
	_, err = client.Enqueue(task)
	return err
}
