package compose

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/wiring"
	"gopkg.in/yaml.v3"
)

type stackCompose struct {
	Networks map[string]networkConfig     `yaml:"networks"`
	Volumes  map[string]interface{}       `yaml:"volumes,omitempty"`
	Services map[string]stackServiceEntry `yaml:"services"`
}

type stackServiceEntry struct {
	Image         string             `yaml:"image,omitempty"`
	ContainerName string             `yaml:"container_name"`
	Restart       string             `yaml:"restart,omitempty"`
	Command       interface{}        `yaml:"command,omitempty"`
	User          string             `yaml:"user,omitempty"`
	Ports         []string           `yaml:"ports,omitempty"`
	Volumes       []string           `yaml:"volumes,omitempty"`
	Environment   map[string]string  `yaml:"environment,omitempty"`
	EnvFile       []string           `yaml:"env_file,omitempty"`
	DependsOn     interface{}        `yaml:"depends_on,omitempty"`
	Labels        []string           `yaml:"labels,omitempty"`
	Healthcheck   *healthcheckConfig `yaml:"healthcheck,omitempty"`
	Networks      []string           `yaml:"networks"`
	Deploy        *deployConfig      `yaml:"deploy,omitempty"`
}

type deployConfig struct {
	Resources *resourcesConfig `yaml:"resources,omitempty"`
}

type resourcesConfig struct {
	Limits       *resourceLimits `yaml:"limits,omitempty"`
	Reservations *resourceLimits `yaml:"reservations,omitempty"`
}

type resourceLimits struct {
	CPUs   string `yaml:"cpus,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

type stackInstance struct {
	id                int
	instanceID        string
	enabled           bool
	containerName     string
	templateServiceID int
}

type stackInstanceConfig struct {
	imageName        string
	imageTag         string
	restartPolicy    string
	command          sql.NullString
	userSpec         sql.NullString
	ports            []portEntry
	volumes          []volumeEntry
	envVars          map[string]string
	envFiles         []string
	networks         []string
	labels           map[string]string
	healthcheck      *healthcheckConfig
	dependencies     []string
	configFiles      []configFileEntry
	configMounts     []configMountEntry
	composeOverrides map[string]interface{}
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

type configMountEntry struct {
	sourcePath     string
	targetPath     string
	readonly       bool
	configFilePath sql.NullString
}

func (g *Generator) GenerateStack(stackName string) ([]byte, []string, error) {
	return g.GenerateStackWithRedaction(stackName, false)
}

func (g *Generator) GenerateStackWithRedaction(stackName string, redactSecrets bool) ([]byte, []string, error) {
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
	secretKeys := make(map[string]map[string]bool)
	for _, inst := range instances {
		if !inst.enabled {
			continue
		}
		cfg, secrets, err := g.loadInstanceEffectiveConfigWithSecrets(stackName, inst)
		if err != nil {
			return nil, nil, fmt.Errorf("load config for %s: %w", inst.instanceID, err)
		}
		configs[inst.instanceID] = cfg
		secretKeys[inst.instanceID] = secrets
	}

	// Build networks map from all instances' networks
	networksMap := make(map[string]networkConfig)
	for _, inst := range instances {
		if !inst.enabled {
			continue
		}
		cfg := configs[inst.instanceID]
		for _, netName := range cfg.networks {
			networksMap[netName] = networkConfig{External: true}
		}
	}

	compose := stackCompose{
		Networks: networksMap,
		Services: make(map[string]stackServiceEntry),
	}

	namedVolumes := make(map[string]interface{})

	portUsage := make(map[string][]string)

	for _, inst := range instances {
		if !inst.enabled {
			continue
		}
		cfg := configs[inst.instanceID]

		networks := cfg.networks

		svc := stackServiceEntry{
			ContainerName: inst.containerName,
			Restart:       cfg.restartPolicy,
			Networks:      networks,
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
			if redactSecrets && secretKeys[inst.instanceID] != nil {
				envCopy := make(map[string]string, len(cfg.envVars))
				for k, v := range cfg.envVars {
					if secretKeys[inst.instanceID][k] {
						envCopy[k] = "***"
					} else {
						envCopy[k] = v
					}
				}
				svc.Environment = envCopy
			} else {
				svc.Environment = cfg.envVars
			}
		}

		svc.EnvFile = cfg.envFiles

		// Merge config mounts into volumes
		for _, mount := range cfg.configMounts {
			var resolvedSource string
			if mount.configFilePath.Valid {
				// Resolve to materialized path: compose/stacks/{stackName}/{instanceID}/{file_path}
				resolvedSource = filepath.Join("compose", "stacks", stackName, inst.instanceID, mount.configFilePath.String)
				resolvedSource = g.resolveStackVolumePath(resolvedSource, stackName, inst.instanceID)
			} else {
				// Use raw source_path for NULL config_file_id
				resolvedSource = g.resolveStackVolumePath(mount.sourcePath, stackName, inst.instanceID)
			}

			mountStr := fmt.Sprintf("%s:%s", resolvedSource, mount.targetPath)
			if mount.readonly {
				mountStr += ":ro"
			}
			svc.Volumes = append(svc.Volumes, mountStr)
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

		deployConfig, err := g.loadResourceLimits(inst.id)
		if err != nil {
			return nil, nil, fmt.Errorf("load resource limits for %s: %w", inst.instanceID, err)
		}
		if deployConfig != nil {
			svc.Deploy = deployConfig
		}

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

	hasOverrides := false
	for _, inst := range instances {
		if inst.enabled && len(configs[inst.instanceID].composeOverrides) > 0 {
			hasOverrides = true
			break
		}
	}

	if !hasOverrides {
		yamlBytes, err := yaml.Marshal(compose)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal yaml: %w", err)
		}
		return yamlBytes, warnings, nil
	}

	rawServices := make(map[string]interface{})
	for id, svc := range compose.Services {
		cfg := configs[id]
		if len(cfg.composeOverrides) == 0 {
			rawServices[id] = svc
			continue
		}

		svcBytes, _ := yaml.Marshal(svc)
		var svcMap map[string]interface{}
		yaml.Unmarshal(svcBytes, &svcMap)

		overrides := cfg.composeOverrides
		if volumes, ok := overrides["volumes"]; ok {
			if volList, ok := volumes.([]interface{}); ok {
				for i, v := range volList {
					if volStr, ok := v.(string); ok {
						volName := extractNamedVolume(volStr)
						if volName != "" && namedVolumes[volName] == nil {
							namedVolumes[volName] = nil
						}
						volList[i] = g.resolveStackVolumePath(volStr, stackName, id)
					}
				}
			}
		}
		if build, ok := overrides["build"]; ok {
			if buildMap, ok := build.(map[string]interface{}); ok {
				if ctx, ok := buildMap["context"].(string); ok {
					buildMap["context"] = g.resolveRelativePath(ctx, "")
				}
			} else if ctx, ok := build.(string); ok {
				overrides["build"] = g.resolveRelativePath(ctx, "")
			}
		}

		for k, v := range overrides {
			svcMap[k] = v
		}
		rawServices[id] = svcMap
	}

	rawCompose := map[string]interface{}{
		"networks": compose.Networks,
		"services": rawServices,
	}
	if len(namedVolumes) > 0 {
		rawCompose["volumes"] = namedVolumes
	}

	yamlBytes, err := yaml.Marshal(rawCompose)
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
	cfg, _, err := g.loadInstanceEffectiveConfigWithSecrets(stackName, inst)
	return cfg, err
}

func (g *Generator) loadInstanceEffectiveConfigWithSecrets(stackName string, inst stackInstance) (*stackInstanceConfig, map[string]bool, error) {
	cfg := &stackInstanceConfig{
		envVars: make(map[string]string),
		labels:  make(map[string]string),
	}

	var overridesJSON []byte
	err := g.db.QueryRow(`
		SELECT name, image_name, image_tag, restart_policy, command, user_spec, compose_overrides
		FROM services WHERE id = $1
	`, inst.templateServiceID).Scan(
		new(string),
		&cfg.imageName,
		&cfg.imageTag,
		&cfg.restartPolicy,
		&cfg.command,
		&cfg.userSpec,
		&overridesJSON,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("template service: %w", err)
	}
	if len(overridesJSON) > 2 {
		json.Unmarshal(overridesJSON, &cfg.composeOverrides)
	}

	cfg.ports, err = g.loadEffectivePorts(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, nil, fmt.Errorf("ports: %w", err)
	}

	cfg.volumes, err = g.loadEffectiveVolumes(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, nil, fmt.Errorf("volumes: %w", err)
	}

	var secretKeys map[string]bool
	cfg.envVars, secretKeys, err = g.loadEffectiveEnvVarsWithSecrets(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, nil, fmt.Errorf("env vars: %w", err)
	}

	cfg.labels, err = g.loadEffectiveLabels(inst.id, inst.templateServiceID, stackName, inst.instanceID)
	if err != nil {
		return nil, nil, fmt.Errorf("labels: %w", err)
	}

	cfg.healthcheck, err = g.loadEffectiveHealthcheck(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, nil, fmt.Errorf("healthcheck: %w", err)
	}

	cfg.dependencies, err = g.loadEffectiveDependencies(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, nil, fmt.Errorf("dependencies: %w", err)
	}

	wireDeps, err := g.loadWireDependencies(inst.id)
	if err != nil {
		return nil, nil, fmt.Errorf("wire dependencies: %w", err)
	}
	cfg.dependencies = append(cfg.dependencies, wireDeps...)

	cfg.configFiles, err = g.loadEffectiveConfigFiles(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, nil, fmt.Errorf("config files: %w", err)
	}

	cfg.envFiles, err = g.loadEffectiveEnvFiles(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, nil, fmt.Errorf("env files: %w", err)
	}

	cfg.networks, err = g.loadEffectiveNetworks(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, nil, fmt.Errorf("networks: %w", err)
	}

	cfg.configMounts, err = g.loadEffectiveConfigMounts(inst.id, inst.templateServiceID)
	if err != nil {
		return nil, nil, fmt.Errorf("config mounts: %w", err)
	}

	return cfg, secretKeys, nil
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
	merged, _, err := g.loadEffectiveEnvVarsWithSecrets(instancePK, templateServiceID)
	return merged, err
}

func (g *Generator) loadEffectiveEnvVarsWithSecrets(instancePK, templateServiceID int) (map[string]string, map[string]bool, error) {
	merged := make(map[string]string)
	secretKeys := make(map[string]bool)

	rows, err := g.db.Query(`
		SELECT key, value, is_secret FROM service_env_vars WHERE service_id = $1
	`, templateServiceID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		var isSecret bool
		if err := rows.Scan(&k, &v, &isSecret); err != nil {
			return nil, nil, err
		}
		merged[k] = v
		if isSecret {
			secretKeys[k] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	wiredEnvVars, err := g.loadWiredEnvVarsForInstance(instancePK)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range wiredEnvVars {
		merged[k] = v
	}

	rows, err = g.db.Query(`
		SELECT key, value, is_secret FROM instance_env_vars WHERE instance_id = $1
	`, instancePK)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		var isSecret bool
		if err := rows.Scan(&k, &v, &isSecret); err != nil {
			return nil, nil, err
		}
		merged[k] = v
		if isSecret {
			secretKeys[k] = true
		}
	}
	return merged, secretKeys, rows.Err()
}

func (g *Generator) loadWiredEnvVarsForInstance(instancePK int) (map[string]string, error) {
	var stackName string
	err := g.db.QueryRow(`
		SELECT st.name
		FROM service_instances si
		JOIN stacks st ON st.id = si.stack_id
		WHERE si.id = $1
	`, instancePK).Scan(&stackName)
	if err != nil {
		return nil, err
	}

	rows, err := g.db.Query(`
		SELECT
			si_provider.instance_id,
			se.name,
			se.port,
			se.protocol,
			COALESCE(sic.env_vars, '{}')
		FROM service_instance_wires siw
		JOIN service_instances si_provider ON si_provider.id = siw.provider_instance_id
		JOIN service_exports se ON se.id = siw.export_contract_id
		JOIN service_import_contracts sic ON sic.id = siw.import_contract_id
		WHERE siw.consumer_instance_id = $1
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merged := make(map[string]string)
	for rows.Next() {
		var providerInstanceName string
		var contractName string
		var port int
		var protocol string
		var envVarsJSON []byte

		if err := rows.Scan(&providerInstanceName, &contractName, &port, &protocol, &envVarsJSON); err != nil {
			return nil, err
		}

		envVars := make(map[string]string)
		if len(envVarsJSON) > 0 {
			if err := json.Unmarshal(envVarsJSON, &envVars); err != nil {
				return nil, fmt.Errorf("unmarshal env_vars: %w", err)
			}
		}

		provider := wiring.Provider{
			InstanceName: providerInstanceName,
			ContractName: contractName,
			Port:         port,
			Protocol:     protocol,
		}
		consumer := wiring.Consumer{
			EnvVars: envVars,
		}

		injections := wiring.InjectEnvVars(stackName, provider, consumer)
		for k, v := range injections {
			merged[k] = v
		}
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
	} else if len(testParts) > 0 && testParts[0] == "CMD-SHELL" {
		hc.Test = append([]string{"CMD-SHELL"}, strings.Join(testParts[1:], " "))
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

func (g *Generator) loadWireDependencies(instancePK int) ([]string, error) {
	rows, err := g.db.Query(`
		SELECT si_provider.instance_id
		FROM service_instance_wires siw
		JOIN service_instances si_provider ON si_provider.id = siw.provider_instance_id
		WHERE siw.consumer_instance_id = $1
		ORDER BY si_provider.instance_id
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

func (g *Generator) loadEffectiveEnvFiles(instancePK, templateServiceID int) ([]string, error) {
	rows, err := g.db.Query(`
		SELECT path FROM service_env_files WHERE service_id = $1 ORDER BY sort_order
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envFiles []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		envFiles = append(envFiles, path)
	}
	return envFiles, rows.Err()
}

func (g *Generator) loadEffectiveNetworks(instancePK, templateServiceID int) ([]string, error) {
	rows, err := g.db.Query(`
		SELECT network_name FROM service_networks WHERE service_id = $1 ORDER BY id
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var networks []string
	for rows.Next() {
		var network string
		if err := rows.Scan(&network); err != nil {
			return nil, err
		}
		networks = append(networks, network)
	}
	return networks, rows.Err()
}

func (g *Generator) loadEffectiveConfigMounts(instancePK, templateServiceID int) ([]configMountEntry, error) {
	rows, err := g.db.Query(`
		SELECT scm.source_path, scm.target_path, scm.readonly, scf.file_path
		FROM service_config_mounts scm
		LEFT JOIN service_config_files scf ON scf.id = scm.config_file_id
		WHERE scm.service_id = $1
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mounts []configMountEntry
	for rows.Next() {
		var m configMountEntry
		if err := rows.Scan(&m.sourcePath, &m.targetPath, &m.readonly, &m.configFilePath); err != nil {
			return nil, err
		}
		mounts = append(mounts, m)
	}
	return mounts, rows.Err()
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

func (g *Generator) loadResourceLimits(instancePK int) (*deployConfig, error) {
	var cpuLimit, cpuReservation, memoryLimit, memoryReservation sql.NullString

	err := g.db.QueryRow(`
		SELECT cpu_limit, cpu_reservation, memory_limit, memory_reservation
		FROM instance_resource_limits WHERE instance_id = $1
	`, instancePK).Scan(&cpuLimit, &cpuReservation, &memoryLimit, &memoryReservation)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	hasLimits := cpuLimit.Valid || memoryLimit.Valid
	hasReservations := cpuReservation.Valid || memoryReservation.Valid

	if !hasLimits && !hasReservations {
		return nil, nil
	}

	deploy := &deployConfig{
		Resources: &resourcesConfig{},
	}

	if hasLimits {
		deploy.Resources.Limits = &resourceLimits{}
		if cpuLimit.Valid && cpuLimit.String != "" {
			deploy.Resources.Limits.CPUs = cpuLimit.String
		}
		if memoryLimit.Valid && memoryLimit.String != "" {
			deploy.Resources.Limits.Memory = memoryLimit.String
		}
	}

	if hasReservations {
		deploy.Resources.Reservations = &resourceLimits{}
		if cpuReservation.Valid && cpuReservation.String != "" {
			deploy.Resources.Reservations.CPUs = cpuReservation.String
		}
		if memoryReservation.Valid && memoryReservation.String != "" {
			deploy.Resources.Reservations.Memory = memoryReservation.String
		}
	}

	return deploy, nil
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
