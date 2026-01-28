package compose

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

var defaultCategoryOrder = []string{
	"database", "storage", "dbms", "erp", "security", "registry",
	"gateway", "proxy", "management", "backend", "ci", "project",
	"mail", "exporters", "analytics", "messaging", "search",
	"workflow", "docs", "testing", "collaboration", "ai", "support",
}

type Importer struct {
	db         *sql.DB
	composeDir string
}

func NewImporter(db *sql.DB, composeDir string) *Importer {
	return &Importer{
		db:         db,
		composeDir: composeDir,
	}
}

func (i *Importer) ImportAll() error {
	categories, err := i.discoverCategories()
	if err != nil {
		return fmt.Errorf("discover categories: %w", err)
	}

	if err := i.importCategories(categories); err != nil {
		return fmt.Errorf("import categories: %w", err)
	}

	for _, category := range categories {
		if err := i.importCategoryServices(category); err != nil {
			return fmt.Errorf("import %s services: %w", category, err)
		}
	}

	if err := i.resolveDependencies(); err != nil {
		return fmt.Errorf("resolve dependencies: %w", err)
	}

	return nil
}

func (i *Importer) discoverCategories() ([]string, error) {
	entries, err := os.ReadDir(i.composeDir)
	if err != nil {
		return nil, err
	}

	found := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() {
			found[entry.Name()] = true
		}
	}

	var ordered []string
	for _, cat := range defaultCategoryOrder {
		if found[cat] {
			ordered = append(ordered, cat)
			delete(found, cat)
		}
	}

	var remaining []string
	for cat := range found {
		remaining = append(remaining, cat)
	}
	sort.Strings(remaining)
	ordered = append(ordered, remaining...)

	return ordered, nil
}

func (i *Importer) importCategories(categories []string) error {
	for idx, name := range categories {
		_, err := i.db.Exec(`
			INSERT INTO categories (name, startup_order)
			VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET startup_order = $2, updated_at = NOW()
		`, name, idx+1)
		if err != nil {
			return fmt.Errorf("insert category %s: %w", name, err)
		}
	}
	return nil
}

func (i *Importer) importCategoryServices(category string) error {
	categoryDir := filepath.Join(i.composeDir, category)
	entries, err := os.ReadDir(categoryDir)
	if err != nil {
		return err
	}

	var categoryID int
	err = i.db.QueryRow("SELECT id FROM categories WHERE name = $1", category).Scan(&categoryID)
	if err != nil {
		return fmt.Errorf("get category id: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		path := filepath.Join(categoryDir, entry.Name())
		if err := i.importServices(path, categoryID); err != nil {
			fmt.Printf("warning: failed to import %s: %v\n", path, err)
		}
	}

	return nil
}

func (i *Importer) importServices(path string, categoryID int) error {
	services, err := ParseFileAll(path)
	if err != nil {
		return err
	}

	for _, parsed := range services {
		if err := i.importService(parsed, categoryID); err != nil {
			fmt.Printf("warning: failed to import service %s: %v\n", parsed.Name, err)
		}
	}
	return nil
}

func (i *Importer) importService(parsed *ParsedService, categoryID int) error {

	tx, err := i.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	overridesJSON := OverridesToJSON(parsed.Overrides)

	var serviceID int
	err = tx.QueryRow(`
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
	`, parsed.Name, categoryID, parsed.ImageName, parsed.ImageTag,
		parsed.RestartPolicy, parsed.Command, parsed.UserSpec, overridesJSON).Scan(&serviceID)
	if err != nil {
		return fmt.Errorf("insert service: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM service_ports WHERE service_id = $1", serviceID); err != nil {
		return err
	}
	for _, port := range parsed.Ports {
		_, err := tx.Exec(`
			INSERT INTO service_ports (service_id, host_ip, host_port, container_port, protocol)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (host_ip, host_port) DO NOTHING
		`, serviceID, port.HostIP, port.HostPort, port.ContainerPort, port.Protocol)
		if err != nil {
			return fmt.Errorf("insert port: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM service_volumes WHERE service_id = $1", serviceID); err != nil {
		return err
	}
	for _, vol := range parsed.Volumes {
		_, err := tx.Exec(`
			INSERT INTO service_volumes (service_id, volume_type, source, target, read_only)
			VALUES ($1, $2, $3, $4, $5)
		`, serviceID, vol.VolumeType, vol.Source, vol.Target, vol.ReadOnly)
		if err != nil {
			return fmt.Errorf("insert volume: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM service_env_vars WHERE service_id = $1", serviceID); err != nil {
		return err
	}
	for _, env := range parsed.EnvVars {
		_, err := tx.Exec(`
			INSERT INTO service_env_vars (service_id, key, value, is_secret)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (service_id, key) DO UPDATE SET value = $3, is_secret = $4
		`, serviceID, env.Key, env.Value, env.IsSecret)
		if err != nil {
			return fmt.Errorf("insert env var: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM service_labels WHERE service_id = $1", serviceID); err != nil {
		return err
	}
	for _, label := range parsed.Labels {
		_, err := tx.Exec(`
			INSERT INTO service_labels (service_id, key, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (service_id, key) DO UPDATE SET value = $3
		`, serviceID, label.Key, label.Value)
		if err != nil {
			return fmt.Errorf("insert label: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM service_healthchecks WHERE service_id = $1", serviceID); err != nil {
		return err
	}
	if parsed.Healthcheck != nil {
		_, err := tx.Exec(`
			INSERT INTO service_healthchecks (service_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, serviceID, parsed.Healthcheck.Test, parsed.Healthcheck.IntervalSeconds,
			parsed.Healthcheck.TimeoutSeconds, parsed.Healthcheck.Retries, parsed.Healthcheck.StartPeriodSeconds)
		if err != nil {
			return fmt.Errorf("insert healthcheck: %w", err)
		}
	}

	return tx.Commit()
}

func (i *Importer) resolveDependencies() error {
	rows, err := i.db.Query(`
		SELECT s.id, s.name FROM services s
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	serviceIDs := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		serviceIDs[name] = id
	}

	categories, err := i.discoverCategories()
	if err != nil {
		return err
	}

	for _, category := range categories {
		categoryDir := filepath.Join(i.composeDir, category)
		entries, err := os.ReadDir(categoryDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" {
				continue
			}

			path := filepath.Join(categoryDir, entry.Name())
			services, err := ParseFileAll(path)
			if err != nil {
				continue
			}

			for _, parsed := range services {
				serviceID, ok := serviceIDs[parsed.Name]
				if !ok {
					continue
				}

				i.db.Exec("DELETE FROM service_dependencies WHERE service_id = $1", serviceID)

				for _, dep := range parsed.Dependencies {
					depID, ok := serviceIDs[dep]
					if !ok {
						continue
					}
					i.db.Exec(`
						INSERT INTO service_dependencies (service_id, depends_on_service_id, condition)
						VALUES ($1, $2, 'service_started')
						ON CONFLICT (service_id, depends_on_service_id) DO NOTHING
					`, serviceID, depID)
				}
			}
		}
	}

	return nil
}
