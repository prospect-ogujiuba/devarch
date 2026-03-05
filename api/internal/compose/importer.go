package compose

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var defaultCategoryOrder = []string{
	"database", "storage", "dbms", "erp", "security", "registry",
	"gateway", "proxy", "management", "backend", "ci", "project",
	"mail", "exporters", "analytics", "messaging", "search",
	"workflow", "docs", "testing", "collaboration", "ai", "support",
}

type Importer struct {
	db          *sql.DB
	composeDir  string
	projectRoot string
}

func NewImporter(db *sql.DB, composeDir string) *Importer {
	return &Importer{
		db:         db,
		composeDir: composeDir,
	}
}

func NewImporterWithRoot(db *sql.DB, composeDir, projectRoot string) *Importer {
	return &Importer{
		db:          db,
		composeDir:  composeDir,
		projectRoot: projectRoot,
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
		if !entry.IsDir() {
			continue
		}

		svcDir := filepath.Join(categoryDir, entry.Name())
		ymlPath := filepath.Join(svcDir, "compose.yml")
		if _, err := os.Stat(ymlPath); err != nil {
			continue
		}

		serviceIDs, err := i.importServices(ymlPath, categoryID, category)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to import %s: %v\n", ymlPath, err)
			continue
		}

		configDir := filepath.Join(svcDir, "config")
		if info, statErr := os.Stat(configDir); statErr == nil && info.IsDir() {
			for _, serviceID := range serviceIDs {
				if _, err := i.importServiceConfigFiles(serviceID, configDir); err != nil {
					fmt.Fprintf(os.Stderr, "warning: config import for %s: %v\n", entry.Name(), err)
				}
				if err := i.resolveServiceConfigMountFKs(serviceID); err != nil {
					fmt.Fprintf(os.Stderr, "warning: config mount FK resolution for %s: %v\n", entry.Name(), err)
				}
			}
		}
	}

	return nil
}

func (i *Importer) importServices(path string, categoryID int, categoryName string) ([]int, error) {
	services, err := ParseFileAll(path)
	if err != nil {
		return nil, err
	}

	var ids []int
	for _, parsed := range services {
		i.resolvePaths(parsed, path)
		id, err := i.importService(parsed, categoryID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to import service %s: %v\n", parsed.Name, err)
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (i *Importer) resolvePaths(parsed *ParsedService, composePath string) {
	if i.projectRoot == "" {
		return
	}
	composeDir := filepath.Dir(composePath)

	// Resolve bind mount sources
	for idx := range parsed.Volumes {
		v := &parsed.Volumes[idx]
		if v.VolumeType == "bind" && !filepath.IsAbs(v.Source) && !strings.HasPrefix(v.Source, "compose/") {
			v.Source = filepath.Clean(filepath.Join(composeDir, v.Source))
		}
	}

	// Resolve build context in overrides
	if parsed.Overrides != nil {
		if build, ok := parsed.Overrides["build"]; ok {
			switch b := build.(type) {
			case string:
				if !filepath.IsAbs(b) && !strings.HasPrefix(b, "compose/") {
					parsed.Overrides["build"] = filepath.Clean(filepath.Join(composeDir, b))
				}
			case map[string]interface{}:
				if ctx, ok := b["context"].(string); ok && !filepath.IsAbs(ctx) && !strings.HasPrefix(ctx, "compose/") {
					b["context"] = filepath.Clean(filepath.Join(composeDir, ctx))
				}
			}
		}
	}
}

// stripProjectRoot converts absolute paths back to relative from project root for portability
func (i *Importer) stripProjectRoot(absPath string) string {
	if i.projectRoot == "" || absPath == "" {
		return absPath
	}
	abs, err := filepath.Abs(i.projectRoot)
	if err != nil {
		return absPath
	}
	if strings.HasPrefix(absPath, abs) {
		rel, err := filepath.Rel(abs, absPath)
		if err == nil {
			return rel
		}
	}
	return absPath
}

func (i *Importer) ImportParsedService(parsed *ParsedService, categoryID int) (int, error) {
	return i.importService(parsed, categoryID)
}

func (i *Importer) importService(parsed *ParsedService, categoryID int) (int, error) {
	tx, err := i.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	overridesJSON := OverridesToJSON(parsed.Overrides)

	var serviceID int
	err = tx.QueryRow(`
		INSERT INTO services (name, category_id, image_name, image_tag, restart_policy, command, user_spec, compose_overrides, container_name_template)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8, NULLIF($9, ''))
		ON CONFLICT (name) DO UPDATE SET
			category_id = $2,
			image_name = $3,
			image_tag = $4,
			restart_policy = $5,
			command = NULLIF($6, ''),
			user_spec = NULLIF($7, ''),
			compose_overrides = $8,
			container_name_template = NULLIF($9, ''),
			updated_at = NOW()
		RETURNING id
	`, parsed.Name, categoryID, parsed.ImageName, parsed.ImageTag,
		parsed.RestartPolicy, parsed.Command, parsed.UserSpec, overridesJSON, parsed.ContainerName).Scan(&serviceID)
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

	if _, err := tx.Exec("DELETE FROM service_env_files WHERE service_id = $1", serviceID); err != nil {
		return 0, err
	}
	for idx, path := range parsed.EnvFiles {
		_, err := tx.Exec(`
			INSERT INTO service_env_files (service_id, path, sort_order)
			VALUES ($1, $2, $3)
		`, serviceID, path, idx)
		if err != nil {
			return 0, fmt.Errorf("insert env_file: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM service_networks WHERE service_id = $1", serviceID); err != nil {
		return 0, err
	}
	for _, network := range parsed.Networks {
		_, err := tx.Exec(`
			INSERT INTO service_networks (service_id, network_name)
			VALUES ($1, $2)
		`, serviceID, network)
		if err != nil {
			return 0, fmt.Errorf("insert network: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM service_config_mounts WHERE service_id = $1", serviceID); err != nil {
		return 0, err
	}
	for _, mount := range parsed.ConfigMounts {
		_, err := tx.Exec(`
			INSERT INTO service_config_mounts (service_id, config_file_id, source_path, target_path, readonly)
			VALUES ($1, NULL, $2, $3, $4)
		`, serviceID, mount.SourcePath, mount.TargetPath, mount.ReadOnly)
		if err != nil {
			return 0, fmt.Errorf("insert config mount: %w", err)
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

	return serviceID, tx.Commit()
}

func (i *Importer) importServiceConfigFiles(serviceID int, configPath string) (int, error) {
	count := 0
	err := filepath.Walk(configPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(configPath, path)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", relPath, err)
		}

		// Skip binary files
		if isBinaryContent(content) {
			return nil
		}

		fileMode := "0644"
		if info.Mode()&0111 != 0 {
			fileMode = "0755"
		}

		_, err = i.db.Exec(`
			INSERT INTO service_config_files (service_id, file_path, content, file_mode)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (service_id, file_path) DO UPDATE SET
				content = $3, file_mode = $4, updated_at = NOW()
		`, serviceID, relPath, string(content), fileMode)
		if err != nil {
			return fmt.Errorf("insert config file %s: %w", relPath, err)
		}

		count++
		return nil
	})
	return count, err
}

func isBinaryContent(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	// Use net/http content type detection
	contentType := http.DetectContentType(data)
	if strings.HasPrefix(contentType, "text/") || strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "application/xml") || strings.HasPrefix(contentType, "application/javascript") {
		return false
	}
	// Also check for common config file patterns that might be detected as octet-stream
	if contentType == "application/octet-stream" {
		// Check first 512 bytes for null bytes
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

func (i *Importer) resolveServiceConfigMountFKs(serviceID int) error {
	type mountRecord struct {
		id         int
		sourcePath string
	}

	rows, err := i.db.Query(`
		SELECT id, source_path FROM service_config_mounts
		WHERE service_id = $1 AND config_file_id IS NULL
	`, serviceID)
	if err != nil {
		return fmt.Errorf("query config mounts: %w", err)
	}
	defer rows.Close()

	var mounts []mountRecord
	for rows.Next() {
		var m mountRecord
		if err := rows.Scan(&m.id, &m.sourcePath); err != nil {
			return err
		}
		mounts = append(mounts, m)
	}

	for _, mount := range mounts {
		cleaned := mount.sourcePath
		for strings.HasPrefix(cleaned, "../") {
			cleaned = strings.TrimPrefix(cleaned, "../")
		}

		if !strings.HasPrefix(cleaned, "config/") {
			continue
		}

		rest := strings.TrimPrefix(cleaned, "config/")
		var configOwner, configRelPath string
		if idx := strings.Index(rest, "/"); idx >= 0 {
			configOwner = rest[:idx]
			configRelPath = rest[idx+1:]
		} else {
			continue
		}

		if configRelPath == "" {
			continue
		}

		var configFileID int
		err := i.db.QueryRow(`
			SELECT scf.id FROM service_config_files scf
			JOIN services s ON s.id = scf.service_id
			WHERE s.name = $1 AND scf.file_path = $2
		`, configOwner, configRelPath).Scan(&configFileID)

		if err == sql.ErrNoRows {
			continue
		} else if err != nil {
			return fmt.Errorf("resolve config_file_id for mount %d: %w", mount.id, err)
		}

		_, err = i.db.Exec(`
			UPDATE service_config_mounts SET config_file_id = $1 WHERE id = $2
		`, configFileID, mount.id)
		if err != nil {
			return fmt.Errorf("update config mount %d: %w", mount.id, err)
		}
	}
	return nil
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
			if !entry.IsDir() {
				continue
			}

			ymlPath := filepath.Join(categoryDir, entry.Name(), "compose.yml")
			services, err := ParseFileAll(ymlPath)
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
