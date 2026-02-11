package export

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/priz/devarch-api/internal/identity"
	"gopkg.in/yaml.v3"
)

type Exporter struct {
	db *sql.DB
}

type instanceRow struct {
	id                int
	instanceID        string
	enabled           bool
	templateServiceID int
}

func NewExporter(db *sql.DB) *Exporter {
	return &Exporter{db: db}
}

func (e *Exporter) Export(stackName string) ([]byte, error) {
	var stackID int
	var description sql.NullString
	var networkName sql.NullString

	err := e.db.QueryRow(`
		SELECT id, description, network_name
		FROM stacks
		WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID, &description, &networkName)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("stack %q not found", stackName)
	}
	if err != nil {
		return nil, fmt.Errorf("query stack: %w", err)
	}

	netName := identity.NetworkName(stackName)
	if networkName.Valid && networkName.String != "" {
		netName = networkName.String
	}

	devarchFile := DevArchFile{
		Version: 1,
		Stack: StackConfig{
			Name:        stackName,
			NetworkName: netName,
		},
		Instances: make(map[string]InstanceDef),
		Wires:     []WireDef{},
	}

	if description.Valid && description.String != "" {
		devarchFile.Stack.Description = description.String
	}

	rows, err := e.db.Query(`
		SELECT si.id, si.instance_id, si.enabled, si.template_service_id
		FROM service_instances si
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
		ORDER BY si.instance_id
	`, stackID)
	if err != nil {
		return nil, fmt.Errorf("query instances: %w", err)
	}
	defer rows.Close()

	var instances []instanceRow
	instanceMap := make(map[int]string)
	for rows.Next() {
		var inst instanceRow
		if err := rows.Scan(&inst.id, &inst.instanceID, &inst.enabled, &inst.templateServiceID); err != nil {
			return nil, fmt.Errorf("scan instance: %w", err)
		}
		instances = append(instances, inst)
		instanceMap[inst.id] = inst.instanceID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate instances: %w", err)
	}

	for _, inst := range instances {
		instanceDef, err := e.loadInstanceDef(stackName, inst)
		if err != nil {
			return nil, fmt.Errorf("load instance %s: %w", inst.instanceID, err)
		}
		devarchFile.Instances[inst.instanceID] = instanceDef
	}

	wires, err := e.loadWires(stackID, instanceMap)
	if err != nil {
		return nil, fmt.Errorf("load wires: %w", err)
	}
	devarchFile.Wires = wires

	yamlBytes, err := yaml.Marshal(devarchFile)
	if err != nil {
		return nil, fmt.Errorf("marshal yaml: %w", err)
	}

	return yamlBytes, nil
}

func (e *Exporter) loadWires(stackID int, instanceMap map[int]string) ([]WireDef, error) {
	rows, err := e.db.Query(`
		SELECT
			siw.consumer_instance_id,
			siw.provider_instance_id,
			sic.name,
			se.name,
			siw.source
		FROM service_instance_wires siw
		JOIN service_import_contracts sic ON sic.id = siw.import_contract_id
		JOIN service_exports se ON se.id = siw.export_contract_id
		WHERE siw.stack_id = $1
		ORDER BY siw.consumer_instance_id, sic.name
	`, stackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wires []WireDef
	for rows.Next() {
		var consumerPK, providerPK int
		var importName, exportName, source string
		if err := rows.Scan(&consumerPK, &providerPK, &importName, &exportName, &source); err != nil {
			return nil, err
		}

		consumerInstanceID, consumerExists := instanceMap[consumerPK]
		providerInstanceID, providerExists := instanceMap[providerPK]

		if consumerExists && providerExists {
			wires = append(wires, WireDef{
				ConsumerInstance: consumerInstanceID,
				ProviderInstance: providerInstanceID,
				ImportContract:   importName,
				ExportContract:   exportName,
				Source:           source,
			})
		}
	}
	return wires, rows.Err()
}

func (e *Exporter) loadInstanceDef(stackName string, inst instanceRow) (InstanceDef, error) {
	def := InstanceDef{
		Enabled: inst.enabled,
	}

	var templateName string
	var imageName, imageTag string
	var command, userSpec sql.NullString

	err := e.db.QueryRow(`
		SELECT name, image_name, image_tag, command, user_spec
		FROM services
		WHERE id = $1
	`, inst.templateServiceID).Scan(&templateName, &imageName, &imageTag, &command, &userSpec)
	if err != nil {
		return def, fmt.Errorf("query template: %w", err)
	}

	def.Template = templateName
	def.Image = fmt.Sprintf("%s:%s", imageName, imageTag)

	if command.Valid && command.String != "" {
		def.Command = command.String
	}

	ports, err := e.loadEffectivePorts(inst.id, inst.templateServiceID)
	if err != nil {
		return def, fmt.Errorf("load ports: %w", err)
	}
	if len(ports) > 0 {
		def.Ports = ports
	}

	volumes, err := e.loadEffectiveVolumes(inst.id, inst.templateServiceID)
	if err != nil {
		return def, fmt.Errorf("load volumes: %w", err)
	}
	if len(volumes) > 0 {
		def.Volumes = volumes
	}

	envVars, err := e.loadEffectiveEnvVars(inst.id, inst.templateServiceID)
	if err != nil {
		return def, fmt.Errorf("load env vars: %w", err)
	}
	if len(envVars) > 0 {
		def.Environment = RedactSecrets(envVars)
	}

	labels, err := e.loadEffectiveLabels(inst.id, inst.templateServiceID, stackName, inst.instanceID)
	if err != nil {
		return def, fmt.Errorf("load labels: %w", err)
	}
	if len(labels) > 0 {
		def.Labels = labels
	}

	domains, err := e.loadEffectiveDomains(inst.id, inst.templateServiceID)
	if err != nil {
		return def, fmt.Errorf("load domains: %w", err)
	}
	if len(domains) > 0 {
		def.Domains = domains
	}

	healthcheck, err := e.loadEffectiveHealthcheck(inst.id, inst.templateServiceID)
	if err != nil {
		return def, fmt.Errorf("load healthcheck: %w", err)
	}
	if healthcheck != nil {
		def.Healthcheck = healthcheck
	}

	deps, err := e.loadEffectiveDependencies(inst.id, inst.templateServiceID)
	if err != nil {
		return def, fmt.Errorf("load dependencies: %w", err)
	}
	if len(deps) > 0 {
		def.Dependencies = deps
	}

	configFiles, err := e.loadEffectiveConfigFiles(inst.id, inst.templateServiceID)
	if err != nil {
		return def, fmt.Errorf("load config files: %w", err)
	}
	if len(configFiles) > 0 {
		def.ConfigFiles = configFiles
	}

	envFiles, err := e.loadEffectiveEnvFiles(inst.id, inst.templateServiceID)
	if err != nil {
		return def, fmt.Errorf("load env files: %w", err)
	}
	if len(envFiles) > 0 {
		def.EnvFiles = envFiles
	}

	networks, err := e.loadEffectiveNetworks(inst.id, inst.templateServiceID)
	if err != nil {
		return def, fmt.Errorf("load networks: %w", err)
	}
	if len(networks) > 0 {
		def.Networks = networks
	}

	configMounts, err := e.loadEffectiveConfigMounts(inst.id, inst.templateServiceID)
	if err != nil {
		return def, fmt.Errorf("load config mounts: %w", err)
	}
	if len(configMounts) > 0 {
		def.ConfigMounts = configMounts
	}

	return def, nil
}

func (e *Exporter) loadEffectivePorts(instancePK, templateServiceID int) ([]PortDef, error) {
	rows, err := e.db.Query(`
		SELECT host_ip, host_port, container_port, protocol
		FROM instance_ports
		WHERE instance_id = $1
		ORDER BY container_port
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []PortDef
	for rows.Next() {
		var p PortDef
		if err := rows.Scan(&p.HostIP, &p.HostPort, &p.ContainerPort, &p.Protocol); err != nil {
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

	rows, err = e.db.Query(`
		SELECT host_ip, host_port, container_port, protocol
		FROM service_ports
		WHERE service_id = $1
		ORDER BY container_port
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p PortDef
		if err := rows.Scan(&p.HostIP, &p.HostPort, &p.ContainerPort, &p.Protocol); err != nil {
			return nil, err
		}
		ports = append(ports, p)
	}
	return ports, rows.Err()
}

func (e *Exporter) loadEffectiveVolumes(instancePK, templateServiceID int) ([]VolumeDef, error) {
	rows, err := e.db.Query(`
		SELECT source, target, read_only
		FROM instance_volumes
		WHERE instance_id = $1
		ORDER BY target
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var volumes []VolumeDef
	for rows.Next() {
		var v VolumeDef
		if err := rows.Scan(&v.Source, &v.Target, &v.ReadOnly); err != nil {
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

	rows, err = e.db.Query(`
		SELECT source, target, read_only
		FROM service_volumes
		WHERE service_id = $1
		ORDER BY target
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v VolumeDef
		if err := rows.Scan(&v.Source, &v.Target, &v.ReadOnly); err != nil {
			return nil, err
		}
		volumes = append(volumes, v)
	}
	return volumes, rows.Err()
}

func (e *Exporter) loadEffectiveEnvVars(instancePK, templateServiceID int) (map[string]string, error) {
	merged := make(map[string]string)

	rows, err := e.db.Query(`
		SELECT key, value, is_secret FROM service_env_vars WHERE service_id = $1
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		var isSecret bool
		if err := rows.Scan(&k, &v, &isSecret); err != nil {
			return nil, err
		}
		if isSecret {
			merged[k] = fmt.Sprintf("${SECRET:%s}", k)
		} else {
			merged[k] = v
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = e.db.Query(`
		SELECT key, value, is_secret FROM instance_env_vars WHERE instance_id = $1
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		var isSecret bool
		if err := rows.Scan(&k, &v, &isSecret); err != nil {
			return nil, err
		}
		if isSecret {
			merged[k] = fmt.Sprintf("${SECRET:%s}", k)
		} else {
			merged[k] = v
		}
	}
	return merged, rows.Err()
}

func (e *Exporter) loadEffectiveLabels(instancePK, templateServiceID int, stackName, instanceID string) (map[string]string, error) {
	merged := make(map[string]string)

	rows, err := e.db.Query(`
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

	rows, err = e.db.Query(`
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

	identityLabels := identity.BuildLabels(stackName, instanceID, strconv.Itoa(templateServiceID))
	for k, v := range identityLabels {
		if _, exists := merged[k]; !exists {
			merged[k] = v
		}
	}

	return merged, nil
}

func (e *Exporter) loadEffectiveDomains(instancePK, templateServiceID int) ([]DomainDef, error) {
	rows, err := e.db.Query(`
		SELECT domain, proxy_port
		FROM instance_domains
		WHERE instance_id = $1
		ORDER BY domain
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []DomainDef
	for rows.Next() {
		var d DomainDef
		if err := rows.Scan(&d.Domain, &d.ProxyPort); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(domains) > 0 {
		return domains, nil
	}

	rows, err = e.db.Query(`
		SELECT domain, proxy_port
		FROM service_domains
		WHERE service_id = $1
		ORDER BY domain
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d DomainDef
		if err := rows.Scan(&d.Domain, &d.ProxyPort); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (e *Exporter) loadEffectiveHealthcheck(instancePK, templateServiceID int) (*HealthcheckDef, error) {
	var test string
	var interval, timeout, retries, startPeriod int

	err := e.db.QueryRow(`
		SELECT test, interval_seconds, timeout_seconds, retries, start_period_seconds
		FROM instance_healthchecks
		WHERE instance_id = $1
	`, instancePK).Scan(&test, &interval, &timeout, &retries, &startPeriod)

	if err == nil {
		hc := &HealthcheckDef{
			Test:     test,
			Interval: fmt.Sprintf("%ds", interval),
			Timeout:  fmt.Sprintf("%ds", timeout),
			Retries:  retries,
		}
		if startPeriod > 0 {
			hc.StartPeriod = fmt.Sprintf("%ds", startPeriod)
		}
		return hc, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	err = e.db.QueryRow(`
		SELECT test, interval_seconds, timeout_seconds, retries, start_period_seconds
		FROM service_healthchecks
		WHERE service_id = $1
	`, templateServiceID).Scan(&test, &interval, &timeout, &retries, &startPeriod)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	hc := &HealthcheckDef{
		Test:     test,
		Interval: fmt.Sprintf("%ds", interval),
		Timeout:  fmt.Sprintf("%ds", timeout),
		Retries:  retries,
	}
	if startPeriod > 0 {
		hc.StartPeriod = fmt.Sprintf("%ds", startPeriod)
	}
	return hc, nil
}

func (e *Exporter) loadEffectiveDependencies(instancePK, templateServiceID int) ([]string, error) {
	rows, err := e.db.Query(`
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

	rows, err = e.db.Query(`
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

func (e *Exporter) loadEffectiveConfigFiles(instancePK, templateServiceID int) (map[string]ConfigFileDef, error) {
	merged := make(map[string]ConfigFileDef)

	rows, err := e.db.Query(`
		SELECT file_path, content, file_mode, is_template
		FROM service_config_files
		WHERE service_id = $1
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filePath, content, fileMode string
		var isTemplate bool
		if err := rows.Scan(&filePath, &content, &fileMode, &isTemplate); err != nil {
			return nil, err
		}
		merged[filePath] = ConfigFileDef{
			Content:    content,
			FileMode:   fileMode,
			IsTemplate: isTemplate,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = e.db.Query(`
		SELECT file_path, content, file_mode, is_template
		FROM instance_config_files
		WHERE instance_id = $1
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filePath, content, fileMode string
		var isTemplate bool
		if err := rows.Scan(&filePath, &content, &fileMode, &isTemplate); err != nil {
			return nil, err
		}
		merged[filePath] = ConfigFileDef{
			Content:    content,
			FileMode:   fileMode,
			IsTemplate: isTemplate,
		}
	}

	return merged, rows.Err()
}

func (e *Exporter) loadEffectiveEnvFiles(instancePK, templateServiceID int) ([]string, error) {
	rows, err := e.db.Query(`
		SELECT path FROM instance_env_files WHERE instance_id = $1 ORDER BY sort_order
	`, instancePK)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(envFiles) > 0 {
		return envFiles, nil
	}

	rows, err = e.db.Query(`
		SELECT path FROM service_env_files WHERE service_id = $1 ORDER BY sort_order
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		envFiles = append(envFiles, path)
	}
	return envFiles, rows.Err()
}

func (e *Exporter) loadEffectiveNetworks(instancePK, templateServiceID int) ([]string, error) {
	rows, err := e.db.Query(`
		SELECT network_name FROM instance_networks WHERE instance_id = $1 ORDER BY network_name
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var networks []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		networks = append(networks, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(networks) > 0 {
		return networks, nil
	}

	rows, err = e.db.Query(`
		SELECT network_name FROM service_networks WHERE service_id = $1 ORDER BY network_name
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		networks = append(networks, name)
	}
	return networks, rows.Err()
}

func (e *Exporter) loadEffectiveConfigMounts(instancePK, templateServiceID int) ([]ConfigMountDef, error) {
	rows, err := e.db.Query(`
		SELECT icm.source_path, icm.target_path, icm.readonly, scf.file_path
		FROM instance_config_mounts icm
		LEFT JOIN service_config_files scf ON scf.id = icm.config_file_id
		WHERE icm.instance_id = $1 ORDER BY icm.target_path
	`, instancePK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mounts []ConfigMountDef
	for rows.Next() {
		var m ConfigMountDef
		var cfPath sql.NullString
		if err := rows.Scan(&m.SourcePath, &m.TargetPath, &m.ReadOnly, &cfPath); err != nil {
			return nil, err
		}
		if cfPath.Valid {
			m.ConfigFilePath = cfPath.String
		}
		mounts = append(mounts, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(mounts) > 0 {
		return mounts, nil
	}

	rows, err = e.db.Query(`
		SELECT scm.source_path, scm.target_path, scm.readonly, scf.file_path
		FROM service_config_mounts scm
		LEFT JOIN service_config_files scf ON scf.id = scm.config_file_id
		WHERE scm.service_id = $1 ORDER BY scm.target_path
	`, templateServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m ConfigMountDef
		var cfPath sql.NullString
		if err := rows.Scan(&m.SourcePath, &m.TargetPath, &m.ReadOnly, &cfPath); err != nil {
			return nil, err
		}
		if cfPath.Valid {
			m.ConfigFilePath = cfPath.String
		}
		mounts = append(mounts, m)
	}
	return mounts, rows.Err()
}
