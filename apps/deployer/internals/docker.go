package internals

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
)

type DeploymentDetails struct {
	ImageURL  string
	Subdomain string
}

func getDeploymentDetails(deployID int64) (*DeploymentDetails, error) {
	logf(deployID, "DB", "querying deployment details from database…")

	// imported from db package
	var d DeploymentDetails
	err := getDeploymentFromDB(deployID, &d)

	if err != nil {
		logf(deployID, "DB", " ERROR: query failed: %v", err)
		return nil, err
	}

	logf(deployID, "DB", "✓ Query successful → ImageURL=%q  Subdomain=%q", d.ImageURL, d.Subdomain)
	return &d, nil
}

// runContainer starts the Docker container on a random available port
// returns (hostPort, containerID, error)
func runContainer(imageName string, subdomain string) (int, string, error) {
	// We still pick a random port for the database record,
	// but Nginx will talk to the container internally on 3000.
	port := rand.Intn(6000) + 3000
	deployID := int64(0) // Will be set later if needed for logging context

	log(deployID, "DOCKER", "─────────────────────────────────────────")
	logf(deployID, "DOCKER", "🐳 Starting container deployment")
	logf(deployID, "DOCKER", "  Subdomain: %s.vercel-clone.local", subdomain)
	logf(deployID, "DOCKER", "  Image: %s", imageName)
	logf(deployID, "DOCKER", "  Assigned port: %d (host) → 3000 (container)", port)
	networkCandidates := []string{}
	if envNet := strings.TrimSpace(os.Getenv("DOCKER_NETWORK")); envNet != "" {
		networkCandidates = append(networkCandidates, envNet)
	}
	networkCandidates = append(networkCandidates, "vercel-clone_vercel_network", "vercel_network")

	var output []byte
	var err error
	startTime := time.Now()
	log(deployID, "DOCKER", "⏳ Waiting for container to start…")
	for _, networkName := range networkCandidates {
		logf(deployID, "DOCKER", "  Trying network: %s", networkName)
		cmd := exec.Command(
			"docker", "run",
			"-d",
			"--network", networkName,
			"-e", fmt.Sprintf("VIRTUAL_HOST=%s.vercel-clone.local", subdomain),
			"-e", "VIRTUAL_PORT=3000",
			"-p", fmt.Sprintf("%d:3000", port),
			"--restart", "unless-stopped",
			imageName,
		)

		logf(deployID, "DOCKER", "↳ Executing: docker run -d --network %s -e VIRTUAL_HOST=%s.vercel-clone.local -e VIRTUAL_PORT=3000 -p %d:3000 --restart unless-stopped %s",
			networkName, subdomain, port, imageName)

		output, err = cmd.CombinedOutput()
		if err == nil {
			break
		}

		errText := strings.ToLower(string(output) + " " + err.Error())
		if strings.Contains(errText, "network") && strings.Contains(errText, "not found") {
			logf(deployID, "DOCKER", "  Network %s not found, trying next candidate", networkName)
			continue
		}

		break
	}

	duration := time.Since(startTime)
	if err != nil {
		logf(deployID, "DOCKER", "❌ ERROR: exec error: %v", err)
		logf(deployID, "DOCKER", "  Duration: %v", duration)
		logf(deployID, "DOCKER", "  Output: %s", string(output))
		return 0, "", fmt.Errorf("docker run failed: %s - %w", string(output), err)
	}

	containerID := strings.TrimSpace(string(output))
	logf(deployID, "DOCKER", "✅ Container started successfully")
	logf(deployID, "DOCKER", "✓ Container ID: %s", containerID)
	logf(deployID, "DOCKER", "✓ Duration: %v", duration)
	log(deployID, "DOCKER", "─────────────────────────────────────────")

	return port, containerID, nil
}

func getDeploymentFromDB(deployID int64, d *DeploymentDetails) error {
	logf(deployID, "DB", "SELECT url(as image), subdomain FROM deployments JOIN projects…")

	err := db.Pool.QueryRow(
		context.Background(),
		`SELECT COALESCE(d.url, ''), COALESCE(p.subdomain, '')
		 FROM deployments d
		 JOIN projects p ON p.id = d.project_id
		 WHERE d.id = $1`,
		deployID,
	).Scan(&d.ImageURL, &d.Subdomain)

	if err != nil {
		logf(deployID, "DB", " ERROR: query failed: %v", err)
		return err
	}

	return nil
}
