package internals

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
)

type DeploymentDetails struct {
	ImageURL  string
	Subdomain string
}

func getDeploymentDetails(deployID int64) (*DeploymentDetails, error) {
	// imported from db package
	var d DeploymentDetails
	err := getDeploymentFromDB(deployID, &d)
	return &d, err
}

// runContainer starts the Docker container on a random available port
// returns (hostPort, containerID, error)
func runContainer(imageName string, subdomain string) (int, string, error) {
	// We still pick a random port for the database record,
	// but Nginx will talk to the container internally on 3000.
	port := rand.Intn(6000) + 3000

	fmt.Printf("🐳 Deploying %s.vercel-clone.local using image: %s\n", subdomain, imageName)

	cmd := exec.Command(
		"docker", "run",
		"-d",
		// 1. Connect to the same network as the nginx-proxy
		"--network", "vercel_network",
		// 2. Tell Nginx Proxy what domain to listen for
		"-e", fmt.Sprintf("VIRTUAL_HOST=%s.vercel-clone.local", subdomain),
		// 3. Tell Nginx Proxy which port the app uses INSIDE the container
		"-e", "VIRTUAL_PORT=3000",
		// 4. Map a random host port (good for debugging, though Nginx won't use it)
		"-p", fmt.Sprintf("%d:3000", port),
		"--restart", "unless-stopped",
		imageName,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, "", fmt.Errorf("docker run failed: %s - %w", string(output), err)
	}

	containerID := strings.TrimSpace(string(output))
	return port, containerID, nil
}

func getDeploymentFromDB(deployID int64, d *DeploymentDetails) error {
	return db.Pool.QueryRow(
		context.Background(),
		`SELECT d.image_url, p.subdomain
		 FROM deployments d
		 JOIN projects p ON p.id = d.project_id
		 WHERE d.id = $1`,
		deployID,
	).Scan(&d.ImageURL, &d.Subdomain)
}
