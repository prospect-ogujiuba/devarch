package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/priz/devarch-api/internal/podman"
)

type Controller struct {
	db           *sql.DB
	podmanClient *podman.Client
}

func NewController(db *sql.DB, podmanClient *podman.Client) *Controller {
	return &Controller{db: db, podmanClient: podmanClient}
}

type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type,omitempty"`
}

func (c *Controller) Start(ctx context.Context, name string) (string, error) {
	composePath, err := c.getComposePath(name)
	if err != nil {
		return "", err
	}
	return c.runCompose(composePath, "up", "-d")
}

func (c *Controller) Stop(ctx context.Context, name string) (string, error) {
	composePath, err := c.getComposePath(name)
	if err != nil {
		return "", err
	}
	return c.runCompose(composePath, "down")
}

func (c *Controller) Restart(ctx context.Context, name string) (string, error) {
	composePath, err := c.getComposePath(name)
	if err != nil {
		return "", err
	}
	if _, err := c.runCompose(composePath, "down"); err != nil {
		return "", err
	}
	return c.runCompose(composePath, "up", "-d")
}

func (c *Controller) Status(ctx context.Context, name string) ([]ServiceStatus, error) {
	rows, err := c.db.Query(`
		SELECT ps.service_name, ps.container_name, ps.service_type
		FROM project_services ps
		JOIN projects p ON p.id = ps.project_id
		WHERE p.name = $1`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []ServiceStatus
	for rows.Next() {
		var svcName string
		var containerName sql.NullString
		var svcType sql.NullString

		if err := rows.Scan(&svcName, &containerName, &svcType); err != nil {
			continue
		}

		cName := containerName.String
		if cName == "" {
			cName = svcName
		}

		status := "unknown"
		inspectOut, err := exec.Command("podman", "inspect", "--format", "{{.State.Status}}", cName).Output()
		if err == nil {
			status = strings.TrimSpace(string(inspectOut))
		} else {
			status = "not-created"
		}

		st := ServiceStatus{
			Name:   svcName,
			Status: status,
		}
		if svcType.Valid {
			st.Type = svcType.String
		}
		statuses = append(statuses, st)
	}

	if statuses == nil {
		statuses = []ServiceStatus{}
	}
	return statuses, nil
}

func (c *Controller) getComposePath(name string) (string, error) {
	var composePath sql.NullString
	err := c.db.QueryRow(`SELECT compose_path FROM projects WHERE name = $1`, name).Scan(&composePath)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("project not found: %s", name)
	}
	if err != nil {
		return "", err
	}
	if !composePath.Valid || composePath.String == "" {
		return "", fmt.Errorf("no compose file for project: %s", name)
	}
	return composePath.String, nil
}

func (c *Controller) runCompose(composePath string, args ...string) (string, error) {
	cmdArgs := []string{"compose", "-f", composePath}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command("podman", cmdArgs...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("compose error: %s: %w", string(out), err)
	}
	return string(out), nil
}

func (c *Controller) StatusJSON(ctx context.Context, name string) (json.RawMessage, error) {
	statuses, err := c.Status(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(statuses)
}
