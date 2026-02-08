package export

import (
	"database/sql"
	"fmt"
	"strings"
)

type ImportResult struct {
	StackName    string   `json:"stack_name"`
	StackCreated bool     `json:"stack_created"`
	Created      []string `json:"created"`
	Updated      []string `json:"updated"`
	Errors       []string `json:"errors,omitempty"`
}

type Importer struct {
	db *sql.DB
}

func NewImporter(db *sql.DB) *Importer {
	return &Importer{db: db}
}

func (imp *Importer) Import(file *DevArchFile) (*ImportResult, error) {
	result := &ImportResult{
		StackName: file.Stack.Name,
		Created:   []string{},
		Updated:   []string{},
		Errors:    []string{},
	}

	tx, err := imp.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	templateNames := make(map[string]bool)
	for _, inst := range file.Instances {
		templateNames[inst.Template] = true
	}

	var missingTemplates []string
	for name := range templateNames {
		var exists bool
		err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM services WHERE name = $1 AND deleted_at IS NULL)`, name).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("check template %q: %w", name, err)
		}
		if !exists {
			missingTemplates = append(missingTemplates, name)
		}
	}

	if len(missingTemplates) > 0 {
		return nil, fmt.Errorf("Template %s not found in catalog. Import the template first.", strings.Join(missingTemplates, ", "))
	}

	var stackID int
	var stackExists bool
	err = tx.QueryRow(`SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL`, file.Stack.Name).Scan(&stackID)
	if err == sql.ErrNoRows {
		stackExists = false
	} else if err != nil {
		return nil, fmt.Errorf("check stack: %w", err)
	} else {
		stackExists = true
	}

	if stackExists {
		var acquired bool
		err = tx.QueryRow(`SELECT pg_try_advisory_xact_lock($1)`, stackID).Scan(&acquired)
		if err != nil || !acquired {
			return nil, fmt.Errorf("stack is locked by another operation")
		}

		networkName := file.Stack.NetworkName
		if networkName == "" {
			networkName = fmt.Sprintf("devarch-%s-net", file.Stack.Name)
		}

		_, err = tx.Exec(`UPDATE stacks SET description = $1, network_name = $2, updated_at = NOW() WHERE id = $3`,
			file.Stack.Description, networkName, stackID)
		if err != nil {
			return nil, fmt.Errorf("update stack: %w", err)
		}
		result.StackCreated = false
	} else {
		networkName := file.Stack.NetworkName
		if networkName == "" {
			networkName = fmt.Sprintf("devarch-%s-net", file.Stack.Name)
		}

		err = tx.QueryRow(`
			INSERT INTO stacks (name, description, network_name, enabled)
			VALUES ($1, $2, $3, true)
			RETURNING id
		`, file.Stack.Name, file.Stack.Description, networkName).Scan(&stackID)
		if err != nil {
			return nil, fmt.Errorf("create stack: %w", err)
		}
		result.StackCreated = true
	}

	for instanceName, inst := range file.Instances {
		var templateServiceID int
		err := tx.QueryRow(`SELECT id FROM services WHERE name = $1 AND deleted_at IS NULL`, inst.Template).Scan(&templateServiceID)
		if err != nil {
			return nil, fmt.Errorf("get template service ID for %q: %w", inst.Template, err)
		}

		var instanceID int
		var instanceExists bool
		err = tx.QueryRow(`
			SELECT id FROM service_instances
			WHERE instance_id = $1 AND stack_id = $2 AND deleted_at IS NULL
		`, instanceName, stackID).Scan(&instanceID)
		if err == sql.ErrNoRows {
			instanceExists = false
		} else if err != nil {
			return nil, fmt.Errorf("check instance %q: %w", instanceName, err)
		} else {
			instanceExists = true
		}

		containerName := fmt.Sprintf("devarch-%s-%s", file.Stack.Name, instanceName)

		if instanceExists {
			_, err = tx.Exec(`
				UPDATE service_instances
				SET template_service_id = $1, enabled = $2, updated_at = NOW()
				WHERE id = $3
			`, templateServiceID, inst.Enabled, instanceID)
			if err != nil {
				return nil, fmt.Errorf("update instance %q: %w", instanceName, err)
			}

			if err := imp.deleteOverrides(tx, instanceID); err != nil {
				return nil, fmt.Errorf("delete overrides for %q: %w", instanceName, err)
			}

			if err := imp.insertOverrides(tx, instanceID, &inst); err != nil {
				return nil, fmt.Errorf("insert overrides for %q: %w", instanceName, err)
			}

			result.Updated = append(result.Updated, instanceName)
		} else {
			err = tx.QueryRow(`
				INSERT INTO service_instances (stack_id, instance_id, template_service_id, container_name, enabled)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id
			`, stackID, instanceName, templateServiceID, containerName, inst.Enabled).Scan(&instanceID)
			if err != nil {
				return nil, fmt.Errorf("create instance %q: %w", instanceName, err)
			}

			if err := imp.insertOverrides(tx, instanceID, &inst); err != nil {
				return nil, fmt.Errorf("insert overrides for %q: %w", instanceName, err)
			}

			result.Created = append(result.Created, instanceName)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return result, nil
}

func (imp *Importer) deleteOverrides(tx *sql.Tx, instanceID int) error {
	tables := []string{
		"instance_ports",
		"instance_volumes",
		"instance_env_vars",
		"instance_labels",
		"instance_domains",
		"instance_healthchecks",
		"instance_dependencies",
		"instance_config_files",
	}

	for _, table := range tables {
		_, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE instance_id = $1", table), instanceID)
		if err != nil {
			return fmt.Errorf("delete from %s: %w", table, err)
		}
	}

	return nil
}

func (imp *Importer) insertOverrides(tx *sql.Tx, instanceID int, inst *InstanceDef) error {
	for _, p := range inst.Ports {
		_, err := tx.Exec(`
			INSERT INTO instance_ports (instance_id, host_ip, host_port, container_port, protocol)
			VALUES ($1, $2, $3, $4, $5)
		`, instanceID, p.HostIP, p.HostPort, p.ContainerPort, p.Protocol)
		if err != nil {
			return fmt.Errorf("insert port: %w", err)
		}
	}

	for _, v := range inst.Volumes {
		volumeType := "bind"
		if v.Source != "" && !strings.HasPrefix(v.Source, "/") && !strings.HasPrefix(v.Source, ".") {
			volumeType = "volume"
		}
		_, err := tx.Exec(`
			INSERT INTO instance_volumes (instance_id, volume_type, source, target, read_only, is_external)
			VALUES ($1, $2, $3, $4, $5, false)
		`, instanceID, volumeType, v.Source, v.Target, v.ReadOnly)
		if err != nil {
			return fmt.Errorf("insert volume: %w", err)
		}
	}

	for key, value := range inst.Environment {
		isSecret := strings.Contains(value, "${SECRET:")
		_, err := tx.Exec(`
			INSERT INTO instance_env_vars (instance_id, key, value, is_secret)
			VALUES ($1, $2, $3, $4)
		`, instanceID, key, value, isSecret)
		if err != nil {
			return fmt.Errorf("insert env var: %w", err)
		}
	}

	for key, value := range inst.Labels {
		if strings.HasPrefix(key, "devarch.") {
			continue
		}
		_, err := tx.Exec(`
			INSERT INTO instance_labels (instance_id, key, value)
			VALUES ($1, $2, $3)
		`, instanceID, key, value)
		if err != nil {
			return fmt.Errorf("insert label: %w", err)
		}
	}

	for _, d := range inst.Domains {
		_, err := tx.Exec(`
			INSERT INTO instance_domains (instance_id, domain, proxy_port)
			VALUES ($1, $2, $3)
		`, instanceID, d.Domain, d.ProxyPort)
		if err != nil {
			return fmt.Errorf("insert domain: %w", err)
		}
	}

	if inst.Healthcheck != nil && inst.Healthcheck.Test != "" {
		intervalSec := parseDuration(inst.Healthcheck.Interval, 30)
		timeoutSec := parseDuration(inst.Healthcheck.Timeout, 10)
		startPeriodSec := parseDuration(inst.Healthcheck.StartPeriod, 0)

		_, err := tx.Exec(`
			INSERT INTO instance_healthchecks (instance_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, instanceID, inst.Healthcheck.Test, intervalSec, timeoutSec, inst.Healthcheck.Retries, startPeriodSec)
		if err != nil {
			return fmt.Errorf("insert healthcheck: %w", err)
		}
	}

	for _, dep := range inst.Dependencies {
		_, err := tx.Exec(`
			INSERT INTO instance_dependencies (instance_id, depends_on, condition)
			VALUES ($1, $2, 'service_started')
		`, instanceID, dep)
		if err != nil {
			return fmt.Errorf("insert dependency: %w", err)
		}
	}

	for path, cfg := range inst.ConfigFiles {
		_, err := tx.Exec(`
			INSERT INTO instance_config_files (instance_id, file_path, content, file_mode, is_template)
			VALUES ($1, $2, $3, $4, $5)
		`, instanceID, path, cfg.Content, cfg.FileMode, cfg.IsTemplate)
		if err != nil {
			return fmt.Errorf("insert config file: %w", err)
		}
	}

	return nil
}

func parseDuration(duration string, defaultSec int) int {
	if duration == "" {
		return defaultSec
	}
	duration = strings.TrimSpace(duration)
	if strings.HasSuffix(duration, "s") {
		var sec int
		fmt.Sscanf(duration, "%ds", &sec)
		return sec
	}
	if strings.HasSuffix(duration, "m") {
		var min int
		fmt.Sscanf(duration, "%dm", &min)
		return min * 60
	}
	var sec int
	fmt.Sscanf(duration, "%d", &sec)
	return sec
}
