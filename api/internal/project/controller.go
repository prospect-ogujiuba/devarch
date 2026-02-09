package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/priz/devarch-api/internal/compose"
	"github.com/priz/devarch-api/internal/container"
)

type Controller struct {
	db              *sql.DB
	containerClient *container.Client
}

func NewController(db *sql.DB, containerClient *container.Client) *Controller {
	return &Controller{db: db, containerClient: containerClient}
}

type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type,omitempty"`
}

type stackInfo struct {
	stackID     int
	stackName   string
	networkName string
	enabled     bool
}

func (c *Controller) getStackInfo(projectName string) (*stackInfo, error) {
	var info stackInfo
	var networkNameNull sql.NullString
	err := c.db.QueryRow(`
		SELECT s.id, s.name, s.network_name, s.enabled
		FROM projects p
		JOIN stacks s ON s.id = p.stack_id
		WHERE p.name = $1 AND s.deleted_at IS NULL
	`, projectName).Scan(&info.stackID, &info.stackName, &networkNameNull, &info.enabled)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	info.networkName = fmt.Sprintf("devarch-%s-net", info.stackName)
	if networkNameNull.Valid && networkNameNull.String != "" {
		info.networkName = networkNameNull.String
	}
	return &info, nil
}

func (c *Controller) Start(ctx context.Context, name string) (string, error) {
	si, err := c.ensureStack(name)
	if err != nil {
		return "", err
	}
	return c.startStack(si)
}

func (c *Controller) Stop(ctx context.Context, name string) (string, error) {
	si, err := c.getStackInfo(name)
	if err != nil {
		return "", err
	}
	if si != nil {
		return c.stopStack(si)
	}
	composePath, err := c.getComposePath(name)
	if err != nil {
		return "", err
	}
	return c.runCompose(composePath, "down")
}

func (c *Controller) Restart(ctx context.Context, name string) (string, error) {
	si, err := c.ensureStack(name)
	if err != nil {
		return "", err
	}
	c.stopStack(si)
	return c.startStack(si)
}

func (c *Controller) Status(ctx context.Context, name string) ([]ServiceStatus, error) {
	si, err := c.getStackInfo(name)
	if err != nil {
		return nil, err
	}
	if si != nil {
		return c.stackStatus(si)
	}
	return c.legacyStatus(name)
}

func (c *Controller) startStack(si *stackInfo) (string, error) {
	if !si.enabled {
		return "", fmt.Errorf("stack %q is disabled", si.stackName)
	}

	labels := map[string]string{
		"devarch.managed_by": "devarch",
		"devarch.stack":      si.stackName,
	}
	if err := c.containerClient.CreateNetwork(si.networkName, labels); err != nil {
		return "", fmt.Errorf("create network: %w", err)
	}

	projName, yamlBytes, err := c.stackCompose(si)
	if err != nil {
		return "", err
	}
	if err := c.containerClient.StartService(projName, yamlBytes); err != nil {
		return "", fmt.Errorf("start stack: %w", err)
	}
	return "started via stack " + si.stackName, nil
}

func (c *Controller) stopStack(si *stackInfo) (string, error) {
	projName, yamlBytes, err := c.stackCompose(si)
	if err != nil {
		return "", err
	}
	if err := c.containerClient.StopStack(projName, yamlBytes); err != nil {
		return "", fmt.Errorf("stop stack: %w", err)
	}
	return "stopped via stack " + si.stackName, nil
}

func (c *Controller) stackStatus(si *stackInfo) ([]ServiceStatus, error) {
	rows, err := c.db.Query(`
		SELECT si.container_name, svc.name
		FROM service_instances si
		JOIN services svc ON svc.id = si.template_service_id
		WHERE si.stack_id = $1 AND si.enabled = true AND si.deleted_at IS NULL
		ORDER BY si.instance_id`, si.stackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []ServiceStatus
	for rows.Next() {
		var containerName, templateName string
		if err := rows.Scan(&containerName, &templateName); err != nil {
			continue
		}
		status := "not-created"
		state, err := c.containerClient.GetStatus(containerName)
		if err == nil && state != nil {
			status = state.Status
		}
		statuses = append(statuses, ServiceStatus{
			Name:   containerName,
			Status: status,
			Type:   templateName,
		})
	}
	if statuses == nil {
		statuses = []ServiceStatus{}
	}
	return statuses, nil
}

func (c *Controller) legacyStatus(name string) ([]ServiceStatus, error) {
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
		state, err := c.containerClient.GetStatus(cName)
		if err == nil && state != nil {
			status = state.Status
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

func (c *Controller) stackCompose(si *stackInfo) (string, []byte, error) {
	gen := compose.NewGenerator(c.db, si.networkName)
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		gen.SetProjectRoot(root)
	}
	if hostRoot := os.Getenv("HOST_PROJECT_ROOT"); hostRoot != "" {
		gen.SetHostProjectRoot(hostRoot)
	}
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot != "" {
		if err := gen.MaterializeStackConfigs(si.stackName, projectRoot); err != nil {
			return "", nil, fmt.Errorf("materialize configs: %w", err)
		}
	}
	yamlBytes, _, err := gen.GenerateStack(si.stackName)
	if err != nil {
		return "", nil, fmt.Errorf("generate compose: %w", err)
	}
	return "devarch-" + si.stackName, yamlBytes, nil
}

func (c *Controller) getInstanceInfo(projectName, service string) (*stackInfo, string, string, bool, error) {
	si, err := c.getStackInfo(projectName)
	if err != nil {
		return nil, "", "", false, err
	}
	if si == nil {
		return nil, "", "", false, nil
	}
	var instanceID, containerName string
	var enabled bool
	err = c.db.QueryRow(`
		SELECT si.instance_id, si.container_name, si.enabled
		FROM service_instances si
		WHERE si.stack_id = $1 AND si.instance_id = $2 AND si.deleted_at IS NULL
	`, si.stackID, service).Scan(&instanceID, &containerName, &enabled)
	if err == sql.ErrNoRows {
		return si, "", "", false, fmt.Errorf("service %q not found in stack", service)
	}
	if err != nil {
		return si, "", "", false, err
	}
	return si, instanceID, containerName, enabled, nil
}

func (c *Controller) StartService(ctx context.Context, projectName, service string) error {
	si, err := c.ensureStack(projectName)
	if err != nil {
		return err
	}
	_, instanceID, _, enabled, err := c.getInstanceInfo(projectName, service)
	if err != nil {
		return err
	}
	if !enabled {
		return fmt.Errorf("instance %q is disabled", service)
	}
	projName, yamlBytes, err := c.stackCompose(si)
	if err != nil {
		return err
	}
	return c.containerClient.StartComposeService(projName, yamlBytes, instanceID)
}

func (c *Controller) StopService(ctx context.Context, projectName, service string) error {
	si, _, containerName, _, err := c.getInstanceInfo(projectName, service)
	if err != nil {
		return err
	}
	if si != nil {
		return c.containerClient.StopContainer(containerName)
	}
	composePath, err := c.getComposePath(projectName)
	if err != nil {
		return err
	}
	_, err = c.containerClient.RunCompose(composePath, "stop", service)
	return err
}

func (c *Controller) RestartService(ctx context.Context, projectName, service string) error {
	si, err := c.ensureStack(projectName)
	if err != nil {
		return err
	}
	_, instanceID, _, enabled, err := c.getInstanceInfo(projectName, service)
	if err != nil {
		return err
	}
	if !enabled {
		return fmt.Errorf("instance %q is disabled", service)
	}
	projName, yamlBytes, err := c.stackCompose(si)
	if err != nil {
		return err
	}
	return c.containerClient.RestartComposeService(projName, yamlBytes, instanceID)
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
	return c.containerClient.RunCompose(composePath, args...)
}

func (c *Controller) StatusJSON(ctx context.Context, name string) (json.RawMessage, error) {
	statuses, err := c.Status(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(statuses)
}
