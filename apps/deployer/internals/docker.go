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

func deploymentHost(subdomain string, deployID int64) string {
	domainSuffix := strings.TrimSpace(os.Getenv("BASE_DOMAIN_SUFFIX"))
	if domainSuffix == "" {
		domainSuffix = "localhost"
	}

	return fmt.Sprintf("%s-%d.%s", subdomain, deployID, domainSuffix)
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

func detectExposedPort(imageName string) string {
	out, err := exec.Command(
		"docker", "image", "inspect",
		"--format", "{{range $p,$_ := .Config.ExposedPorts}}{{$p}}{{\"\\n\"}}{{end}}",
		imageName,
	).Output()
	if err != nil {
		logf(0, "DOCKER", "WARNING: could not inspect image ports: %v, defaulting to 3000", err)
		return "3000"
	}
	ports := strings.Fields(string(out))
	for _, p := range ports {
		parts := strings.SplitN(p, "/", 2)
		if len(parts) > 0 && parts[0] != "" {
			logf(0, "DOCKER", "detected exposed port: %s", parts[0])
			return parts[0]
		}
	}
	logf(0, "DOCKER", "no exposed ports found in image, defaulting to 3000")
	return "3000"
}

// runContainer starts a Docker container discoverable by nginx-proxy via VIRTUAL_HOST.
func runContainer(imageName string, subdomain string, deployID int64) (string, error) {
	host := deploymentHost(subdomain, deployID)

	// Auto-detect exposed port from the Docker image
	port := detectExposedPort(imageName)

	// Get network from env OR fallback
	network := strings.TrimSpace(os.Getenv("DOCKER_NETWORK"))
	if network == "" {
		network = "vercel-clone_vercel_network"
	}

	logf(deployID, "DOCKER", " Starting container deployment")
	logf(deployID, "DOCKER", "Host: %s", host)
	logf(deployID, "DOCKER", "Network: %s", network)
	logf(deployID, "DOCKER", "Port: %s (auto-detected from image)", port)
	logf(deployID, "DOCKER", "Mode: nginx-only (no host port mapping)")

	cmd := exec.Command(
		"docker", "run",
		"-d",
		"--network", network,
		"--label", "com.vercel-clone.managed=true",
		"--label", fmt.Sprintf("com.vercel-clone.deployment-id=%d", deployID),
		"-e", fmt.Sprintf("VIRTUAL_HOST=%s", host),
		"-e", fmt.Sprintf("VIRTUAL_PORT=%s", port),
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
