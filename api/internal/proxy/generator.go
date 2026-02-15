package proxy

import (
	"bytes"
	"database/sql"
	"fmt"
	"text/template"
)

// Generator fetches proxy targets from the DB and renders config for any
// supported reverse proxy.
type Generator struct {
	db *sql.DB
}

func NewGenerator(db *sql.DB) *Generator {
	return &Generator{db: db}
}

// GenerateForService builds proxy config for a standalone service's domains.
func (g *Generator) GenerateForService(proxyType ProxyType, serviceName string) (*ProxyConfigResult, error) {
	targets, err := g.serviceTargets(serviceName)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("service %q has no domains configured", serviceName)
	}
	config, err := render(proxyType, targets)
	if err != nil {
		return nil, err
	}
	return &ProxyConfigResult{
		ProxyType: proxyType,
		Scope:     "service",
		Name:      serviceName,
		Config:    config,
		Targets:   targets,
	}, nil
}

// GenerateForStack builds proxy config for every enabled instance in a stack.
func (g *Generator) GenerateForStack(proxyType ProxyType, stackName string) (*ProxyConfigResult, error) {
	targets, err := g.stackTargets(stackName)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("stack %q has no instances with domains configured", stackName)
	}
	config, err := render(proxyType, targets)
	if err != nil {
		return nil, err
	}
	return &ProxyConfigResult{
		ProxyType: proxyType,
		Scope:     "stack",
		Name:      stackName,
		Config:    config,
		Targets:   targets,
	}, nil
}

// GenerateForProject builds proxy config for a project's domain and services.
func (g *Generator) GenerateForProject(proxyType ProxyType, projectName string) (*ProxyConfigResult, error) {
	targets, err := g.projectTargets(projectName)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("project %q has no domain configured", projectName)
	}
	config, err := render(proxyType, targets)
	if err != nil {
		return nil, err
	}
	return &ProxyConfigResult{
		ProxyType: proxyType,
		Scope:     "project",
		Name:      projectName,
		Config:    config,
		Targets:   targets,
	}, nil
}

// GenerateForInstance builds proxy config for a single instance.
func (g *Generator) GenerateForInstance(proxyType ProxyType, stackName string, instanceID string) (*ProxyConfigResult, error) {
	targets, err := g.instanceTargets(stackName, instanceID)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("instance %q in stack %q has no domains configured", instanceID, stackName)
	}
	config, err := render(proxyType, targets)
	if err != nil {
		return nil, err
	}
	return &ProxyConfigResult{
		ProxyType: proxyType,
		Scope:     "instance",
		Name:      instanceID,
		Config:    config,
		Targets:   targets,
	}, nil
}

// --- DB queries ---

func (g *Generator) serviceTargets(name string) ([]ProxyTarget, error) {
	rows, err := g.db.Query(`
		SELECT sd.domain, sd.proxy_port, s.name
		FROM service_domains sd
		JOIN services s ON s.id = sd.service_id
		WHERE s.name = $1 AND s.enabled = true
	`, name)
	if err != nil {
		return nil, fmt.Errorf("query service domains: %w", err)
	}
	defer rows.Close()

	var targets []ProxyTarget
	for rows.Next() {
		var domain, svcName string
		var port int
		if err := rows.Scan(&domain, &port, &svcName); err != nil {
			return nil, fmt.Errorf("scan service domain: %w", err)
		}
		targets = append(targets, ProxyTarget{
			Name:       svcName,
			Domain:     domain,
			TargetHost: svcName,
			TargetPort: port,
			HTTPS:      true,
		})
	}
	return targets, rows.Err()
}

func (g *Generator) stackTargets(stackName string) ([]ProxyTarget, error) {
	// First try instance-level domains (overrides)
	rows, err := g.db.Query(`
		SELECT id.domain, id.proxy_port, si.container_name, si.instance_id
		FROM instance_domains id
		JOIN service_instances si ON si.id = id.instance_id
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND st.deleted_at IS NULL AND si.deleted_at IS NULL AND si.enabled = true
	`, stackName)
	if err != nil {
		return nil, fmt.Errorf("query instance domains: %w", err)
	}
	defer rows.Close()

	var targets []ProxyTarget
	for rows.Next() {
		var domain, containerName, instanceID string
		var port int
		if err := rows.Scan(&domain, &port, &containerName, &instanceID); err != nil {
			return nil, fmt.Errorf("scan instance domain: %w", err)
		}
		targets = append(targets, ProxyTarget{
			Name:       instanceID,
			Domain:     domain,
			TargetHost: containerName,
			TargetPort: port,
			HTTPS:      true,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Also include template-level domains for instances without overrides
	templateRows, err := g.db.Query(`
		SELECT sd.domain, sd.proxy_port, si.container_name, si.instance_id
		FROM service_domains sd
		JOIN services s ON s.id = sd.service_id
		JOIN service_instances si ON si.template_service_id = s.id
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND st.deleted_at IS NULL AND si.deleted_at IS NULL AND si.enabled = true
		AND NOT EXISTS (
			SELECT 1 FROM instance_domains id2 WHERE id2.instance_id = si.id
		)
	`, stackName)
	if err != nil {
		return nil, fmt.Errorf("query template domains for stack: %w", err)
	}
	defer templateRows.Close()

	for templateRows.Next() {
		var domain, containerName, instanceID string
		var port int
		if err := templateRows.Scan(&domain, &port, &containerName, &instanceID); err != nil {
			return nil, fmt.Errorf("scan template domain: %w", err)
		}
		targets = append(targets, ProxyTarget{
			Name:       instanceID,
			Domain:     domain,
			TargetHost: containerName,
			TargetPort: port,
			HTTPS:      true,
		})
	}
	return targets, templateRows.Err()
}

func (g *Generator) instanceTargets(stackName string, instanceID string) ([]ProxyTarget, error) {
	rows, err := g.db.Query(`
		SELECT id.domain, id.proxy_port, si.container_name, si.instance_id
		FROM instance_domains id
		JOIN service_instances si ON si.id = id.instance_id
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.instance_id = $2 AND st.deleted_at IS NULL AND si.deleted_at IS NULL AND si.enabled = true
	`, stackName, instanceID)
	if err != nil {
		return nil, fmt.Errorf("query instance domains: %w", err)
	}
	defer rows.Close()

	var targets []ProxyTarget
	for rows.Next() {
		var domain, containerName, instID string
		var port int
		if err := rows.Scan(&domain, &port, &containerName, &instID); err != nil {
			return nil, fmt.Errorf("scan instance domain: %w", err)
		}
		targets = append(targets, ProxyTarget{
			Name:       instID,
			Domain:     domain,
			TargetHost: containerName,
			TargetPort: port,
			HTTPS:      true,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(targets) > 0 {
		return targets, nil
	}

	templateRows, err := g.db.Query(`
		SELECT sd.domain, sd.proxy_port, si.container_name, si.instance_id
		FROM service_domains sd
		JOIN services s ON s.id = sd.service_id
		JOIN service_instances si ON si.template_service_id = s.id
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.instance_id = $2 AND st.deleted_at IS NULL AND si.deleted_at IS NULL AND si.enabled = true
	`, stackName, instanceID)
	if err != nil {
		return nil, fmt.Errorf("query template domains for instance: %w", err)
	}
	defer templateRows.Close()

	for templateRows.Next() {
		var domain, containerName, instID string
		var port int
		if err := templateRows.Scan(&domain, &port, &containerName, &instID); err != nil {
			return nil, fmt.Errorf("scan template domain: %w", err)
		}
		targets = append(targets, ProxyTarget{
			Name:       instID,
			Domain:     domain,
			TargetHost: containerName,
			TargetPort: port,
			HTTPS:      true,
		})
	}
	return targets, templateRows.Err()
}

func (g *Generator) projectTargets(projectName string) ([]ProxyTarget, error) {
	var domain sql.NullString
	var proxyPort sql.NullInt32
	var projectType string

	err := g.db.QueryRow(`
		SELECT domain, proxy_port, project_type FROM projects WHERE name = $1
	`, projectName).Scan(&domain, &proxyPort, &projectType)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project %q not found", projectName)
	}
	if err != nil {
		return nil, fmt.Errorf("query project: %w", err)
	}

	if !domain.Valid || domain.String == "" {
		return nil, nil
	}

	port := 80
	if proxyPort.Valid {
		port = int(proxyPort.Int32)
	}

	containerName := projectName + "-app"
	maxBody := ""
	if projectType == "laravel" || projectType == "wordpress" || projectType == "php" {
		maxBody = "100M"
	}

	targets := []ProxyTarget{{
		Name:          projectName,
		Domain:        domain.String,
		TargetHost:    containerName,
		TargetPort:    port,
		HTTPS:         true,
		ClientMaxBody: maxBody,
	}}

	return targets, nil
}

// --- Rendering ---

func render(proxyType ProxyType, targets []ProxyTarget) (string, error) {
	tmplStr, ok := templates[proxyType]
	if !ok {
		return "", fmt.Errorf("unsupported proxy type: %s", proxyType)
	}
	tmpl, err := template.New(string(proxyType)).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, targets); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
