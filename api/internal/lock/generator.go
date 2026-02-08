package lock

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/priz/devarch-api/internal/container"
)

type Generator struct {
	db              *sql.DB
	containerClient *container.Client
}

func NewGenerator(db *sql.DB, cc *container.Client) *Generator {
	return &Generator{
		db:              db,
		containerClient: cc,
	}
}

func (g *Generator) Generate(stackName string, ymlContent []byte) (*LockFile, error) {
	lock := &LockFile{
		Version:     1,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		ConfigHash:  ComputeHash(ymlContent),
		Instances:   make(map[string]InstLock),
	}

	var stackID int
	var networkName sql.NullString
	err := g.db.QueryRow(`
		SELECT id, network_name
		FROM stacks
		WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID, &networkName)
	if err != nil {
		return nil, fmt.Errorf("query stack: %w", err)
	}

	netName := fmt.Sprintf("devarch-%s-net", stackName)
	if networkName.Valid && networkName.String != "" {
		netName = networkName.String
	}

	lock.Stack = StackLock{
		Name:        stackName,
		NetworkName: netName,
	}

	netInfo, err := g.containerClient.InspectNetwork(netName)
	if err == nil {
		lock.Stack.NetworkID = netInfo.ID
	}

	containers, err := g.containerClient.ListContainersWithLabels(map[string]string{
		"devarch.stack_id": stackName,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	for _, containerName := range containers {
		instanceName := extractInstanceName(stackName, containerName)
		if instanceName == "" {
			continue
		}

		instLock, err := g.buildInstanceLock(stackName, instanceName, containerName)
		if err != nil {
			return nil, fmt.Errorf("build instance lock %s: %w", instanceName, err)
		}

		lock.Instances[instanceName] = instLock
	}

	return lock, nil
}

func extractInstanceName(stackName, containerName string) string {
	prefix := fmt.Sprintf("devarch-%s-", stackName)
	if !strings.HasPrefix(containerName, prefix) {
		return ""
	}
	return strings.TrimPrefix(containerName, prefix)
}

func (g *Generator) buildInstanceLock(stackName, instanceName, containerName string) (InstLock, error) {
	var templateName string
	var templateCreatedAt time.Time

	err := g.db.QueryRow(`
		SELECT s.name, s.created_at
		FROM service_instances si
		JOIN services s ON si.template_service_id = s.id
		JOIN stacks st ON si.stack_id = st.id
		WHERE st.name = $1 AND si.instance_id = $2 AND si.deleted_at IS NULL
	`, stackName, instanceName).Scan(&templateName, &templateCreatedAt)
	if err != nil {
		return InstLock{}, fmt.Errorf("query instance: %w", err)
	}

	templateHash := g.getTemplateHash(templateName, templateCreatedAt)

	imageRef, err := g.getContainerImage(containerName)
	if err != nil {
		return InstLock{}, fmt.Errorf("get container image: %w", err)
	}

	imageDigest, _ := g.getImageDigest(imageRef)

	ports, err := g.getResolvedPorts(containerName)
	if err != nil {
		return InstLock{}, fmt.Errorf("get resolved ports: %w", err)
	}

	return InstLock{
		Template:      templateName,
		TemplateHash:  templateHash,
		ImageTag:      imageRef,
		ImageDigest:   imageDigest,
		ResolvedPorts: ports,
	}, nil
}

func (g *Generator) getTemplateHash(templateName string, createdAt time.Time) string {
	combined := templateName + createdAt.Format(time.RFC3339Nano)
	hash := ComputeHash([]byte(combined))
	if len(hash) > 16 {
		return hash[:16]
	}
	return hash
}

func (g *Generator) getImageDigest(imageRef string) (string, error) {
	runtime := g.containerClient.RuntimeName()

	var cmd *exec.Cmd
	if runtime == container.RuntimePodman {
		cmd = exec.Command("podman", "image", "inspect", imageRef, "--format", "{{.Digest}}")
	} else {
		cmd = exec.Command("docker", "image", "inspect", imageRef, "--format", "{{index .RepoDigests 0}}")
	}

	out, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	digest := strings.TrimSpace(string(out))

	if runtime == container.RuntimeDocker && strings.Contains(digest, "@") {
		parts := strings.Split(digest, "@")
		if len(parts) == 2 {
			digest = parts[1]
		}
	}

	return digest, nil
}

func (g *Generator) getContainerImage(containerName string) (string, error) {
	runtime := g.containerClient.RuntimeName()

	cmd := exec.Command(string(runtime), "inspect", containerName, "--format", "{{.Config.Image}}")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("inspect container: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

func (g *Generator) getResolvedPorts(containerName string) (map[string]int, error) {
	runtime := g.containerClient.RuntimeName()

	cmd := exec.Command(string(runtime), "inspect", containerName, "--format", "{{json .NetworkSettings.Ports}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("inspect ports: %w", err)
	}

	var portsData map[string][]struct {
		HostPort string `json:"HostPort"`
	}
	if err := json.Unmarshal(out, &portsData); err != nil {
		return nil, fmt.Errorf("parse ports: %w", err)
	}

	resolved := make(map[string]int)
	for portProto, bindings := range portsData {
		if len(bindings) > 0 && bindings[0].HostPort != "" {
			var hostPort int
			fmt.Sscanf(bindings[0].HostPort, "%d", &hostPort)
			if hostPort > 0 {
				resolved[portProto] = hostPort
			}
		}
	}

	return resolved, nil
}
