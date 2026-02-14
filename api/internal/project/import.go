package project

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/priz/devarch-api/internal/compose"
)

func (c *Controller) EnsureStack(projectName string) error {
	_, err := c.ensureStack(projectName)
	return err
}

func (c *Controller) ensureStack(projectName string) (*stackInfo, error) {
	si, err := c.getStackInfo(projectName)
	if err != nil {
		return nil, err
	}

	if si != nil {
		var count int
		c.db.QueryRow(`SELECT COUNT(*) FROM service_instances WHERE stack_id = $1 AND deleted_at IS NULL`, si.stackID).Scan(&count)
		if count > 0 {
			return si, nil
		}
	}

	var composePath sql.NullString
	err = c.db.QueryRow(`SELECT compose_path FROM projects WHERE name = $1`, projectName).Scan(&composePath)
	if err != nil {
		return nil, fmt.Errorf("project not found: %s", projectName)
	}
	if !composePath.Valid || composePath.String == "" {
		if si != nil {
			return si, nil
		}
		return nil, fmt.Errorf("no compose file for project: %s", projectName)
	}

	services, err := compose.ParseFileAll(composePath.String)
	if err != nil {
		return nil, fmt.Errorf("parse compose: %w", err)
	}
	if len(services) == 0 {
		if si != nil {
			return si, nil
		}
		return nil, fmt.Errorf("no services in compose file for %s", projectName)
	}

	composeDir := filepath.Dir(composePath.String)
	for _, svc := range services {
		resolveServicePaths(svc, composeDir)
	}

	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	categoryID, err := ensureCategory(tx, "projects")
	if err != nil {
		return nil, fmt.Errorf("ensure category: %w", err)
	}

	templateIDs := make(map[string]int, len(services))
	for _, svc := range services {
		templateName := projectName + "-" + svc.Name
		id, err := importTemplate(tx, templateName, categoryID, svc)
		if err != nil {
			return nil, fmt.Errorf("import template %s: %w", templateName, err)
		}
		templateIDs[svc.Name] = id
	}

	for _, svc := range services {
		svcID := templateIDs[svc.Name]
		for _, dep := range svc.Dependencies {
			depID, ok := templateIDs[dep]
			if !ok {
				continue
			}
			_, err := tx.Exec(`
				INSERT INTO service_dependencies (service_id, depends_on_service_id, condition)
				VALUES ($1, $2, 'service_started')
				ON CONFLICT (service_id, depends_on_service_id) DO NOTHING
			`, svcID, depID)
			if err != nil {
				return nil, fmt.Errorf("insert dependency: %w", err)
			}
		}
	}

	if err := importBuildContextFiles(tx, composePath.String, templateIDs); err != nil {
		return nil, fmt.Errorf("import build context files: %w", err)
	}

	if err := importEnvFiles(tx, composePath.String, templateIDs); err != nil {
		return nil, fmt.Errorf("import env files: %w", err)
	}

	if err := importProjectConfigFiles(tx, composePath.String, templateIDs); err != nil {
		return nil, fmt.Errorf("import config files: %w", err)
	}

	stackID := 0
	networkName := "devarch-" + projectName + "-net"

	if si != nil {
		stackID = si.stackID
		networkName = si.networkName
	} else {
		err = tx.QueryRow(`
			INSERT INTO stacks (name, network_name, source)
			VALUES ($1, $2, 'project-import')
			RETURNING id
		`, projectName, networkName).Scan(&stackID)
		if err != nil {
			return nil, fmt.Errorf("create stack: %w", err)
		}
		_, err = tx.Exec(`UPDATE projects SET stack_id = $1 WHERE name = $2`, stackID, projectName)
		if err != nil {
			return nil, fmt.Errorf("link project to stack: %w", err)
		}
	}

	for _, svc := range services {
		containerName := "devarch-" + projectName + "-" + svc.Name
		_, err := tx.Exec(`
			INSERT INTO service_instances (stack_id, instance_id, template_service_id, container_name)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (stack_id, instance_id) WHERE deleted_at IS NULL DO NOTHING
		`, stackID, svc.Name, templateIDs[svc.Name], containerName)
		if err != nil {
			return nil, fmt.Errorf("create instance %s: %w", svc.Name, err)
		}
	}

	for _, svc := range services {
		for _, dep := range svc.Dependencies {
			_, err := tx.Exec(`
				INSERT INTO instance_dependencies (instance_id, depends_on, condition)
				VALUES (
					(SELECT id FROM service_instances WHERE stack_id = $1 AND instance_id = $2),
					$3,
					'service_started'
				)
				ON CONFLICT (instance_id, depends_on) DO NOTHING
			`, stackID, svc.Name, dep)
			if err != nil {
				return nil, fmt.Errorf("insert instance dep: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &stackInfo{
		stackID:     stackID,
		stackName:   projectName,
		networkName: networkName,
		enabled:     true,
	}, nil
}

func ensureCategory(tx *sql.Tx, name string) (int, error) {
	var id int
	err := tx.QueryRow(`
		INSERT INTO categories (name, startup_order)
		VALUES ($1, 999)
		ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, name).Scan(&id)
	return id, err
}

func importTemplate(tx *sql.Tx, name string, categoryID int, parsed *compose.ParsedService) (int, error) {
	overridesJSON := compose.OverridesToJSON(parsed.Overrides)

	var serviceID int
	err := tx.QueryRow(`
		INSERT INTO services (name, category_id, image_name, image_tag, restart_policy, command, user_spec, compose_overrides)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8)
		ON CONFLICT (name) DO UPDATE SET
			category_id = $2,
			image_name = $3,
			image_tag = $4,
			restart_policy = $5,
			command = NULLIF($6, ''),
			user_spec = NULLIF($7, ''),
			compose_overrides = $8,
			updated_at = NOW()
		RETURNING id
	`, name, categoryID, parsed.ImageName, parsed.ImageTag,
		parsed.RestartPolicy, parsed.Command, parsed.UserSpec, overridesJSON).Scan(&serviceID)
	if err != nil {
		return 0, fmt.Errorf("insert service: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM service_ports WHERE service_id = $1", serviceID); err != nil {
		return 0, err
	}
	for _, port := range parsed.Ports {
		_, err := tx.Exec(`
			INSERT INTO service_ports (service_id, host_ip, host_port, container_port, protocol)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (service_id, host_ip, host_port) DO NOTHING
		`, serviceID, port.HostIP, port.HostPort, port.ContainerPort, port.Protocol)
		if err != nil {
			return 0, fmt.Errorf("insert port: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM service_volumes WHERE service_id = $1", serviceID); err != nil {
		return 0, err
	}
	for _, vol := range parsed.Volumes {
		_, err := tx.Exec(`
			INSERT INTO service_volumes (service_id, volume_type, source, target, read_only, is_external)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, serviceID, vol.VolumeType, vol.Source, vol.Target, vol.ReadOnly, vol.IsExternal)
		if err != nil {
			return 0, fmt.Errorf("insert volume: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM service_env_vars WHERE service_id = $1", serviceID); err != nil {
		return 0, err
	}
	for _, env := range parsed.EnvVars {
		_, err := tx.Exec(`
			INSERT INTO service_env_vars (service_id, key, value, is_secret)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (service_id, key) DO UPDATE SET value = $3, is_secret = $4
		`, serviceID, env.Key, env.Value, env.IsSecret)
		if err != nil {
			return 0, fmt.Errorf("insert env var: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM service_labels WHERE service_id = $1", serviceID); err != nil {
		return 0, err
	}
	for _, label := range parsed.Labels {
		_, err := tx.Exec(`
			INSERT INTO service_labels (service_id, key, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (service_id, key) DO UPDATE SET value = $3
		`, serviceID, label.Key, label.Value)
		if err != nil {
			return 0, fmt.Errorf("insert label: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM service_healthchecks WHERE service_id = $1", serviceID); err != nil {
		return 0, err
	}
	if parsed.Healthcheck != nil {
		_, err := tx.Exec(`
			INSERT INTO service_healthchecks (service_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, serviceID, parsed.Healthcheck.Test, parsed.Healthcheck.IntervalSeconds,
			parsed.Healthcheck.TimeoutSeconds, parsed.Healthcheck.Retries, parsed.Healthcheck.StartPeriodSeconds)
		if err != nil {
			return 0, fmt.Errorf("insert healthcheck: %w", err)
		}
	}

	return serviceID, nil
}

func resolveServicePaths(svc *compose.ParsedService, composeDir string) {
	for i := range svc.Volumes {
		v := &svc.Volumes[i]
		if v.VolumeType == "bind" && !filepath.IsAbs(v.Source) {
			v.Source = filepath.Clean(filepath.Join(composeDir, v.Source))
		}
	}

	if svc.Overrides == nil {
		return
	}

	if build, ok := svc.Overrides["build"]; ok {
		switch b := build.(type) {
		case string:
			if !filepath.IsAbs(b) {
				svc.Overrides["build"] = filepath.Clean(filepath.Join(composeDir, b))
			}
		case map[string]interface{}:
			if ctx, ok := b["context"].(string); ok && !filepath.IsAbs(ctx) {
				b["context"] = filepath.Clean(filepath.Join(composeDir, ctx))
			}
		}
	}

	if volumes, ok := svc.Overrides["volumes"]; ok {
		if volList, ok := volumes.([]interface{}); ok {
			for i, v := range volList {
				if volStr, ok := v.(string); ok {
					parts := strings.SplitN(volStr, ":", 3)
					if len(parts) >= 2 && !filepath.IsAbs(parts[0]) && (strings.HasPrefix(parts[0], ".") || strings.Contains(parts[0], "/")) {
						parts[0] = filepath.Clean(filepath.Join(composeDir, parts[0]))
						volList[i] = strings.Join(parts, ":")
					}
				}
			}
		}
	}
}

func isBinaryContent(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	contentType := http.DetectContentType(data)
	if strings.HasPrefix(contentType, "text/") || strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "application/xml") || strings.HasPrefix(contentType, "application/javascript") {
		return false
	}
	if contentType == "application/octet-stream" {
		check := data
		if len(check) > 512 {
			check = check[:512]
		}
		for _, b := range check {
			if b == 0 {
				return true
			}
		}
		return false
	}
	return true
}

func importBuildContextFiles(tx *sql.Tx, composePath string, templateIDs map[string]int) error {
	services, err := compose.ParseFileAll(composePath)
	if err != nil {
		return fmt.Errorf("parse compose: %w", err)
	}

	composeDir := filepath.Dir(composePath)

	for _, svc := range services {
		serviceID, ok := templateIDs[svc.Name]
		if !ok {
			continue
		}

		if svc.Overrides == nil {
			continue
		}

		build, ok := svc.Overrides["build"]
		if !ok {
			continue
		}

		var buildContext string
		switch b := build.(type) {
		case string:
			buildContext = b
		case map[string]interface{}:
			if ctx, ok := b["context"].(string); ok {
				buildContext = ctx
			}
		}

		if buildContext == "" {
			continue
		}

		if !filepath.IsAbs(buildContext) {
			buildContext = filepath.Clean(filepath.Join(composeDir, buildContext))
		}

		dockerfilePath := filepath.Join(buildContext, "Dockerfile")
		if _, err := os.Stat(dockerfilePath); err == nil {
			content, err := os.ReadFile(dockerfilePath)
			if err == nil && !isBinaryContent(content) {
				_, err = tx.Exec(`
					INSERT INTO service_config_files (service_id, file_path, content, file_mode)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (service_id, file_path) DO UPDATE SET
						content = $3, file_mode = $4, updated_at = NOW()
				`, serviceID, "Dockerfile", string(content), "0644")
				if err != nil {
					return fmt.Errorf("insert Dockerfile for %s: %w", svc.Name, err)
				}
			}
		}

		dockerignorePath := filepath.Join(buildContext, ".dockerignore")
		if _, err := os.Stat(dockerignorePath); err == nil {
			content, err := os.ReadFile(dockerignorePath)
			if err == nil && !isBinaryContent(content) {
				_, err = tx.Exec(`
					INSERT INTO service_config_files (service_id, file_path, content, file_mode)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (service_id, file_path) DO UPDATE SET
						content = $3, file_mode = $4, updated_at = NOW()
				`, serviceID, ".dockerignore", string(content), "0644")
				if err != nil {
					return fmt.Errorf("insert .dockerignore for %s: %w", svc.Name, err)
				}
			}
		}
	}

	return nil
}

func importEnvFiles(tx *sql.Tx, composePath string, templateIDs map[string]int) error {
	services, err := compose.ParseFileAll(composePath)
	if err != nil {
		return fmt.Errorf("parse compose: %w", err)
	}

	composeDir := filepath.Dir(composePath)

	for _, svc := range services {
		serviceID, ok := templateIDs[svc.Name]
		if !ok {
			continue
		}

		for _, envFile := range svc.EnvFiles {
			envPath := envFile
			if !filepath.IsAbs(envPath) {
				envPath = filepath.Clean(filepath.Join(composeDir, envPath))
			}

			if _, err := os.Stat(envPath); err != nil {
				continue
			}

			content, err := os.ReadFile(envPath)
			if err != nil {
				continue
			}

			if isBinaryContent(content) {
				continue
			}

			relPath := filepath.Base(envFile)

			_, err = tx.Exec(`
				INSERT INTO service_config_files (service_id, file_path, content, file_mode)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (service_id, file_path) DO UPDATE SET
					content = $3, file_mode = $4, updated_at = NOW()
			`, serviceID, relPath, string(content), "0644")
			if err != nil {
				return fmt.Errorf("insert env file %s for %s: %w", relPath, svc.Name, err)
			}
		}
	}

	return nil
}

func importProjectConfigFiles(tx *sql.Tx, composePath string, templateIDs map[string]int) error {
	services, err := compose.ParseFileAll(composePath)
	if err != nil {
		return fmt.Errorf("parse compose: %w", err)
	}

	composeDir := filepath.Dir(composePath)
	configDir := filepath.Join(composeDir, "config")

	if _, err := os.Stat(configDir); err != nil {
		return nil
	}

	for _, svc := range services {
		serviceID, ok := templateIDs[svc.Name]
		if !ok {
			continue
		}

		serviceConfigDir := filepath.Join(configDir, svc.Name)
		if _, err := os.Stat(serviceConfigDir); err != nil {
			continue
		}

		err := filepath.Walk(serviceConfigDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(serviceConfigDir, path)
			if err != nil {
				return err
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			if isBinaryContent(content) {
				return nil
			}

			fileMode := "0644"
			if info.Mode()&0111 != 0 {
				fileMode = "0755"
			}

			_, err = tx.Exec(`
				INSERT INTO service_config_files (service_id, file_path, content, file_mode)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (service_id, file_path) DO UPDATE SET
					content = $3, file_mode = $4, updated_at = NOW()
			`, serviceID, relPath, string(content), fileMode)
			if err != nil {
				return fmt.Errorf("insert config file %s for %s: %w", relPath, svc.Name, err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("walk config dir for %s: %w", svc.Name, err)
		}
	}

	return nil
}
