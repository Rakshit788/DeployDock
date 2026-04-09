package internals

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/hibiken/asynq"
)

type DeployPayload struct {
	DeploymentID int64 `json:"deployment_id"`
}

func HandleDeploy(ctx context.Context, t *asynq.Task) error {
	var payload DeployPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}

	if err := updateDeploymentStatus(payload.DeploymentID, "deploying", ""); err != nil {
		return fmt.Errorf("failed to update status to deploying: %w", err)
	}

	fmt.Printf(" Deploy started for deployment %d\n", payload.DeploymentID)

	details, err := getDeploymentDetails(payload.DeploymentID)
	if err != nil {
		updateDeploymentStatus(payload.DeploymentID, "failed", fmt.Sprintf("failed to get deployment details: %v", err))
		return fmt.Errorf("failed to get deployment details: %w", err)
	}

	port, containerID, err := runContainer(details.ImageURL, details.Subdomain)
	if err != nil {
		updateDeploymentStatus(payload.DeploymentID, "failed", fmt.Sprintf("failed to run container: %v", err))
		return fmt.Errorf("failed to run container: %w", err)
	}

	if err := saveContainerInfo(payload.DeploymentID, containerID, port); err != nil {
		return fmt.Errorf("failed to save container info: %w", err)
	}

	fmt.Printf(" Deployment %d is live at http://%s.vercel-clone.local:%d\n", payload.DeploymentID, details.Subdomain, port)
	return nil

}

func updateDeploymentStatus(deployID int64, status string, logs string) error {
	_, err := db.Pool.Exec(
		context.Background(),
		`UPDATE deployments SET status=$1, logs=$2 WHERE id=$3`,
		status, logs, deployID,
	)
	return err
}

func saveContainerInfo(deployID int64, containerID string, port int) error {
	_, err := db.Pool.Exec(
		context.Background(),
		`UPDATE deployments SET container_id=$1, container_port=$2 WHERE id=$3`,
		containerID, port, deployID,
	)
	return err
}
