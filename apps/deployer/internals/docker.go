package internals

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
)

type DeploymentDetails struct {
	ImageURL  string
	Subdomain string
}

func deploymentHost(subdomain string) string {
	domainSuffix := strings.TrimSpace(os.Getenv("BASE_DOMAIN_SUFFIX"))
	if domainSuffix == "" {
		domainSuffix = "localhost"
	}

	return fmt.Sprintf("%s.%s", subdomain, domainSuffix)
}

func getDeploymentDetails(deployID int64) (*DeploymentDetails, error) {
	logf(deployID, "DB", "querying deployment details from database…")

	var d DeploymentDetails
	err := getDeploymentFromDB(deployID, &d)

	if err != nil {
		logf(deployID, "DB", " ERROR: query failed: %v", err)
		return nil, err
	}

	logf(deployID, "DB", "✓ Query successful → ImageURL=%q  Subdomain=%q", d.ImageURL, d.Subdomain)
	return &d, nil
}

// runContainer starts a Docker container discoverable by nginx-proxy via VIRTUAL_HOST.
func runContainer(imageName string, subdomain string) (string, error) {
	host := deploymentHost(subdomain)
	deployID := int64(0)

	// Get network from env OR fallback
	network := strings.TrimSpace(os.Getenv("DOCKER_NETWORK"))
	if network == "" {
		network = "vercel-clone_vercel_network"
	}

	logf(deployID, "DOCKER", " Starting container deployment")
	logf(deployID, "DOCKER", "Host: %s", host)
	logf(deployID, "DOCKER", "Network: %s", network)
	logf(deployID, "DOCKER", "Mode: nginx-only (no host port mapping)")

	cmd := exec.Command(
		"docker", "run",
		"-d",
		"--network", network,
		"--label", "com.vercel-clone.managed=true",
		"-e", fmt.Sprintf("VIRTUAL_HOST=%s", host),
		"-e", "VIRTUAL_PORT=3000",
		imageName,
	)

	output, err := cmd.CombinedOutput()

	if err != nil {
		logf(deployID, "DOCKER", " ERROR: %v", err)
		logf(deployID, "DOCKER", "Output: %s", string(output))
		return "", fmt.Errorf("docker run failed: %s - %w", string(output), err)
	}

	containerID := strings.TrimSpace(string(output))

	logf(deployID, "DOCKER", " Container started")
	logf(deployID, "DOCKER", "Container ID: %s", containerID)

	return containerID, nil
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
