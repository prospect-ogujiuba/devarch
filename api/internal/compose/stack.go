package compose

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/priz/devarch-api/internal/container"
	"gopkg.in/yaml.v3"
)

type stackCompose struct {
	Networks map[string]networkConfig     `yaml:"networks"`
	Volumes  map[string]interface{}       `yaml:"volumes,omitempty"`
	Services map[string]stackServiceEntry `yaml:"services"`
}

type stackServiceEntry struct {
	Image         string            `yaml:"image,omitempty"`
	ContainerName string            `yaml:"container_name"`
	Restart       string            `yaml:"restart,omitempty"`
	Command       interface{}       `yaml:"command,omitempty"`
	User          string            `yaml:"user,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	DependsOn     interface{}       `yaml:"depends_on,omitempty"`
	Labels        []string          `yaml:"labels,omitempty"`
	Healthcheck   *healthcheckConfig `yaml:"healthcheck,omitempty"`
	Networks      []string          `yaml:"networks"`
}

type stackInstance struct {
	id                int
	instanceID        string
	enabled           bool
	containerName     string
	templateServiceID int
}

type stackInstanceConfig struct {
	imageName     string
	imageTag      string
	restartPolicy string
	command       sql.NullString
	userSpec      sql.NullString
	ports         []portEntry
	volumes       []volumeEntry
	envVars       map[string]string
	labels        map[string]string
	healthcheck   *healthcheckConfig
	dependencies  []string
	configFiles   []configFileEntry
}

type portEntry struct {
	hostIP        string
	hostPort      int
	containerPort int
	protocol      string
}

type volumeEntry struct {
	volumeType string
	source     string
	target     string
	readOnly   bool
	isExternal bool
}

type configFileEntry struct {
	filePath string
	content  string
	fileMode string
}

func (g *Generator) GenerateStack(stackName string) ([]byte, []string, error) {
	var warnings []string

	var networkName sql.NullString
	err := g.db.QueryRow(`
		SELECT network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&networkName)
	if err != nil {
		return nil, nil, fmt.Errorf("query stack: %w", err)
	}

	netName := g.networkName
	if netName == "" {
		if networkName.Valid && networkName.String != "" {
			netName = networkName.String
		} else {
			netName = fmt.Sprintf("devarch-%s-net", stackName)
		}
	}

	instances, err := g.loadStackInstances(stackName)
	if err != nil {
		return nil, nil, fmt.Errorf("load instances: %w", err)
	}

	enabledMap := make(map[string]bool)
	allInstanceIDs := make(map[string]bool)
	for _, inst := range instances {
		allInstanceIDs[inst.instanceID] = true
		if inst.enabled {
			enabledMap[inst.instanceID] = true
		} else {
			warnings = append(warnings, fmt.Sprintf("Skipped disabled instance: %s", inst.instanceID))
		}
	}

	configs := make(map[string]*stackInstanceConfig)
	for _, inst := range instances {
		if !inst.enabled {
			continue
		}
		cfg, err := g.loadInstanceEffectiveConfig(stackName, inst)
		if err != nil {
			return nil, nil, fmt.Errorf("load config for %s: %w", inst.instanceID, err)
		}
		configs[inst.instanceID] = cfg
	}

	compose := stackCompose{
		Networks: map[string]networkConfig{
			netName: {External: true},
		},
		Services: make(map[string]stackServiceEntry),
	}

	namedVolumes := make(map[string]interface{})

	portUsage := make(map[string][]string)

	for _, inst := range instances {
		if !inst.enabled {
			continue
		}
		cfg := configs[inst.instanceID]

		svc := stackServiceEntry{
			ContainerName: inst.containerName,
			Restart:       cfg.restartPolicy,
			Networks:      []string{netName},
		}

		if cfg.imageName != "" {
			svc.Image = fmt.Sprintf("%s:%s", cfg.imageName, cfg.imageTag)
		}
		if cfg.command.Valid && cfg.command.String != "" {
			svc.Command = cfg.command.String
		}
		if cfg.userSpec.Valid && cfg.userSpec.String != "" {
			svc.User = cfg.userSpec.String
		}

		for _, p := range cfg.ports {
			portStr := fmt.Sprintf("%s:%d:%d", p.hostIP, p.hostPort, p.containerPort)
			if p.protocol != "tcp" {
				portStr += "/" + p.protocol
			}
			svc.Ports = append(svc.Ports, portStr)

			key := fmt.Sprintf("%s:%d", p.hostIP, p.hostPort)
			portUsage[key] = append(portUsage[key], inst.instanceID)
		}

		for _, v := range cfg.volumes {
			source := g.resolveStackVolumePath(v.source, stackName, inst.instanceID)
			volStr := fmt.Sprintf("%s:%s", source, v.target)
			if v.readOnly {
				volStr += ":ro"
			}
			svc.Volumes = append(svc.Volumes, volStr)

			isNamed := v.volumeType == "named" || extractNamedVolume(volStr) != ""
			if isNamed {
				if v.isExternal {
					namedVolumes[v.source] = map[string]interface{}{"external": true}
				} else {
					namedVolumes[v.source] = nil
				}
			}
		}

		if len(cfg.envVars) > 0 {
			svc.Environment = cfg.envVars
		}

		labelKeys := make([]string, 0, len(cfg.labels))
		for k := range cfg.labels {
			labelKeys = append(labelKeys, k)
		}
		sort.Strings(labelKeys)
		for _, k := range labelKeys {
			svc.Labels = append(svc.Labels, fmt.Sprintf("%s=%s", k, cfg.labels[k]))
		}

		svc.Healthcheck = cfg.healthcheck

		svc.DependsOn = g.buildDependsOn(inst.instanceID, cfg.dependencies, enabledMap, allInstanceIDs, configs, &warnings)

		compose.Services[inst.instanceID] = svc
	}

	for key, users := range portUsage {
		if len(users) > 1 {
			warnings = append(warnings, fmt.Sprintf("Port conflict: host port %s used by instances %s", key, strings.Join(users, ", ")))
		}
	}

	if len(namedVolumes) > 0 {
		compose.Volumes = namedVolumes
	}

	yamlBytes, err := yaml.Marshal(compose)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal yaml: %w", err)
	}

	return yamlBytes, warnings, nil
}

func (g *Generator) buildDependsOn(instanceID string, deps []string, enabledMap, allInstanceIDs map[string]bool, configs map[string]*stackInstanceConfig, warnings *[]string) interface{} {
	if len(deps) == 0 {
		return nil
	}

	var validDeps []string
	hasHealthcheck := make(map[string]bool)

	for _, target := range deps {
		if !allInstanceIDs[target] {
			*warnings = append(*warnings, fmt.Sprintf("Instance %s: dependency %s not found in stack", instanceID, target))
			continue
		}
		if !enabledMap[target] {
			*warnings = append(*warnings, fmt.Sprintf("Instance %s: stripped dependency on disabled instance %s", instanceID, target))
			continue
		}
		validDeps = append(validDeps, target)
		if cfg, ok := configs[target]; ok && cfg.healthcheck != nil {
			hasHealthcheck[target] = true
		}
	}

	if len(validDeps) == 0 {
		return nil
	}

	anyHealthcheck := false
	for _, target := range validDeps {
		if hasHealthcheck[target] {
			anyHealthcheck = true
			break
		}
	}

	if !anyHealthcheck {
		return validDeps
	}

	depMap := make(map[string]interface{})
	for _, target := range validDeps {
		if hasHealthcheck[target] {
			depMap[target] = map[string]string{"condition": "service_healthy"}
		} else {
			depMap[target] = map[string]string{"condition": "service_started"}
		}
	}
	return depMap
}

func (g *Generator) loadStackInstances(stackName string) ([]stackInstance, error) {
	rows, err := g.db.Query(`
		SELECT si.id, si.instance_id, si.enabled, si.container_name, si.template_service_id
		FROM service_instances si
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.deleted_at IS NULL AND st.deleted_at IS NULL
		ORDER BY si.instance_id
	`, stackName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []stackInstance
	for rows.Next() {
		var inst stackInstance
		var containerName sql.NullString
		if err := rows.Scan(&inst.id, &inst.instanceID, &inst.enabled, &containerName, &inst.templateServiceID); err != nil {
			return nil, err
		}
		if containerName.Valid && containerName.String != "" {
			inst.containerName = containerName.String
		} else {
			inst.containerName = container.ContainerName(stackName, inst.instanceID)
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

func (g *Generator) loadInstanceEffectiveConfig(stackName string, inst stackInstance) (*stackInstanceConfig, error) {
	cfg := &stackInstanceConfig{
		envVars: make(map[string]string),
		labels:  make(map[string]string),
	}

	err := g.db.QueryRow(`
		SELECT name, image_name, image_tag, restart_policy, command, user_spec
		FROM services WHERE id = $1
	`, inst.templateServiceID).Scan(
		new(string),
		&cfg.imageName,
		&cfg.imageTag,
		&cfg.restartPolicy,
		&cfg.command,
		&cfg.userSpec,
	)
	if err != nil {
		return nil, fmt.Errorf("template service: %w", err)
	}

	cfg.ports, err = g.loadEffectivePorts(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, fmt.Errorf("ports: %w", err)
	}

	cfg.volumes, err = g.loadEffectiveVolumes(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, fmt.Errorf("volumes: %w", err)
	}

	cfg.envVars, err = g.loadEffectiveEnvVars(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, fmt.Errorf("env vars: %w", err)
	}

	cfg.labels, err = g.loadEffectiveLabels(inst.id, inst.templateServiceID, stackName, inst.instanceID)
	if err != nil {
		return nil, fmt.Errorf("labels: %w", err)
	}

	cfg.healthcheck, err = g.loadEffectiveHealthcheck(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, fmt.Errorf("healthcheck: %w", err)
	}

	cfg.dependencies, err = g.loadEffectiveDependencies(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, fmt.Errorf("dependencies: %w", err)
	}

	cfg.configFiles, err = g.loadEffectiveConfigFiles(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, fmt.Errorf("config files: %w", err)
	}

	return cfg, nil
}

func (g *Generator) loadEffectivePorts(instancePK, templateServiceID int) ([]portEntry, error) {
	rows, err := g.db.Query(`
		SELECT host_ip, host_port, container_port, protocol
		FROM instance_ports WHERE instance_id = $1
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []portEntry
	for rows.Next() {
		var p portEntry
		if err := rows.Scan(&p.hostIP, &p.hostPort, &p.containerPort, &p.protocol); err != nil {
			return nil, err
		}
		ports = append(ports, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ports) > 0 {
		return ports, nil
	}

	rows, err = g.db.Query(`
		SELECT host_ip, host_port, container_port, protocol
		FROM service_ports WHERE service_id = $1
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p portEntry
		if err := rows.Scan(&p.hostIP, &p.hostPort, &p.containerPort, &p.protocol); err != nil {
			return nil, err
		}
		ports = append(ports, p)
	}
	return ports, rows.Err()
}

func (g *Generator) loadEffectiveVolumes(instancePK, templateServiceID int) ([]volumeEntry, error) {
	rows, err := g.db.Query(`
		SELECT volume_type, source, target, read_only, is_external
		FROM instance_volumes WHERE instance_id = $1
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var volumes []volumeEntry
	for rows.Next() {
		var v volumeEntry
		if err := rows.Scan(&v.volumeType, &v.source, &v.target, &v.readOnly, &v.isExternal); err != nil {
			return nil, err
		}
		volumes = append(volumes, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(volumes) > 0 {
		return volumes, nil
	}

	rows, err = g.db.Query(`
		SELECT volume_type, source, target, read_only, is_external
		FROM service_volumes WHERE service_id = $1
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v volumeEntry
		if err := rows.Scan(&v.volumeType, &v.source, &v.target, &v.readOnly, &v.isExternal); err != nil {
			return nil, err
		}
		volumes = append(volumes, v)
	}
	return volumes, rows.Err()
}

func (g *Generator) loadEffectiveEnvVars(instancePK, templateServiceID int) (map[string]string, error) {
	merged := make(map[string]string)

	rows, err := g.db.Query(`
		SELECT key, value FROM service_env_vars WHERE service_id = $1
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		merged[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = g.db.Query(`
		SELECT key, value FROM instance_env_vars WHERE instance_id = $1
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		merged[k] = v
	}
	return merged, rows.Err()
}

func (g *Generator) loadEffectiveLabels(instancePK, templateServiceID int, stackName, instanceID string) (map[string]string, error) {
	merged := make(map[string]string)

	rows, err := g.db.Query(`
		SELECT key, value FROM service_labels WHERE service_id = $1
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		merged[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = g.db.Query(`
		SELECT key, value FROM instance_labels WHERE instance_id = $1
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		merged[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	identityLabels := container.BuildLabels(stackName, instanceID, strconv.Itoa(templateServiceID))
	for k, v := range identityLabels {
		if _, exists := merged[k]; !exists {
			merged[k] = v
		}
	}

	return merged, nil
}

func (g *Generator) loadEffectiveHealthcheck(instancePK, templateServiceID int) (*healthcheckConfig, error) {
	hc, err := g.scanHealthcheck(`
		SELECT test, interval_seconds, timeout_seconds, retries, start_period_seconds
		FROM instance_healthchecks WHERE instance_id = $1
	`, instancePK)
	if err != nil {
		return nil, err
	}
	if hc != nil {
		return hc, nil
	}

	return g.scanHealthcheck(`
		SELECT test, interval_seconds, timeout_seconds, retries, start_period_seconds
		FROM service_healthchecks WHERE service_id = $1
	`, templateServiceID)
}

func (g *Generator) scanHealthcheck(query string, id int) (*healthcheckConfig, error) {
	var test string
	var interval, timeout, retries, startPeriod int

	err := g.db.QueryRow(query, id).Scan(&test, &interval, &timeout, &retries, &startPeriod)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	hc := &healthcheckConfig{
		Retries: retries,
	}

	testParts := strings.Fields(test)
	if len(testParts) > 0 && testParts[0] == "CMD" {
		hc.Test = testParts
	} else {
		hc.Test = append([]string{"CMD-SHELL"}, test)
	}

	hc.Interval = fmt.Sprintf("%ds", interval)
	hc.Timeout = fmt.Sprintf("%ds", timeout)
	if startPeriod > 0 {
		hc.StartPeriod = fmt.Sprintf("%ds", startPeriod)
	}

	return hc, nil
}

func (g *Generator) loadEffectiveDependencies(instancePK, templateServiceID int) ([]string, error) {
	rows, err := g.db.Query(`
		SELECT depends_on FROM instance_dependencies WHERE instance_id = $1 ORDER BY depends_on
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []string
	for rows.Next() {
		var dep string
		if err := rows.Scan(&dep); err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(deps) > 0 {
		return deps, nil
	}

	rows, err = g.db.Query(`
		SELECT s.name
		FROM service_dependencies sd
		JOIN services s ON s.id = sd.depends_on_service_id
		WHERE sd.service_id = $1
		ORDER BY s.name
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dep string
		if err := rows.Scan(&dep); err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	return deps, rows.Err()
}

func (g *Generator) loadEffectiveConfigFiles(instancePK, templateServiceID int) ([]configFileEntry, error) {
	merged := make(map[string]configFileEntry)

	rows, err := g.db.Query(`
		SELECT file_path, content, file_mode
		FROM service_config_files WHERE service_id = $1
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var f configFileEntry
		if err := rows.Scan(&f.filePath, &f.content, &f.fileMode); err != nil {
			return nil, err
		}
		merged[f.filePath] = f
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = g.db.Query(`
		SELECT file_path, content, file_mode
		FROM instance_config_files WHERE instance_id = $1
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var f configFileEntry
		if err := rows.Scan(&f.filePath, &f.content, &f.fileMode); err != nil {
			return nil, err
		}
		merged[f.filePath] = f
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]configFileEntry, 0, len(merged))
	for _, f := range merged {
		result = append(result, f)
	}
	return result, nil
}

func (g *Generator) resolveStackVolumePath(source, stackName, instanceID string) string {
	if strings.HasPrefix(source, "compose/stacks/") {
		if g.hostProjectRoot != "" {
			return filepath.Join(g.hostProjectRoot, "api", source)
		}
		return source
	}
	return g.resolveRelativePath(source, "")
}

func (g *Generator) MaterializeStackConfigs(stackName, baseDir string) error {
	instances, err := g.loadStackInstances(stackName)
	if err != nil {
		return fmt.Errorf("load instances: %w", err)
	}

	tmpDir := filepath.Join(baseDir, "compose", "stacks", ".tmp-"+stackName)
	finalDir := filepath.Join(baseDir, "compose", "stacks", stackName)

	if err := os.MkdirAll(filepath.Dir(tmpDir), 0755); err != nil {
		return fmt.Errorf("mkdir parent: %w", err)
	}

	os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("mkdir tmp: %w", err)
	}

	for _, inst := range instances {
		if !inst.enabled {
			continue
		}

		configFiles, err := g.loadEffectiveConfigFiles(inst.id, inst.templateServiceID)
		if err != nil {
			os.RemoveAll(tmpDir)
			return fmt.Errorf("load config files for %s: %w", inst.instanceID, err)
		}

		for _, f := range configFiles {
			fullPath := filepath.Join(tmpDir, inst.instanceID, f.filePath)
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				os.RemoveAll(tmpDir)
				return fmt.Errorf("mkdir %s: %w", dir, err)
			}

			mode := parseFileMode(f.fileMode)
			if err := os.WriteFile(fullPath, []byte(f.content), mode); err != nil {
				os.RemoveAll(tmpDir)
				return fmt.Errorf("write %s: %w", fullPath, err)
			}
		}
	}

	os.RemoveAll(finalDir)
	if err := os.Rename(tmpDir, finalDir); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("atomic swap: %w", err)
	}

	return nil
}
