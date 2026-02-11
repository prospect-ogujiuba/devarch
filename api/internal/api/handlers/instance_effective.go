package handlers

import (
	"github.com/priz/devarch-api/internal/api/respond"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/wiring"
	"github.com/priz/devarch-api/pkg/models"
)

type effectiveConfigResponse struct {
	InstanceID       string                     `json:"instance_id"`
	StackName        string                     `json:"stack_name"`
	TemplateName     string                     `json:"template_name"`
	ImageName        string                     `json:"image_name"`
	ImageTag         string                     `json:"image_tag"`
	RestartPolicy    string                     `json:"restart_policy"`
	Command          string                     `json:"command,omitempty"`
	UserSpec         string                     `json:"user_spec,omitempty"`
	Ports            []models.ServicePort       `json:"ports"`
	Volumes          []models.ServiceVolume     `json:"volumes"`
	EnvVars          []models.ServiceEnvVar     `json:"env_vars"`
	EnvFiles         []string                   `json:"env_files"`
	Networks         []string                   `json:"networks"`
	ConfigMounts     []models.ServiceConfigMount `json:"config_mounts"`
	WiredEnvVars     map[string]string          `json:"wired_env_vars,omitempty"`
	Labels           []models.ServiceLabel      `json:"labels"`
	Domains          []models.ServiceDomain     `json:"domains"`
	Healthcheck      *models.ServiceHealthcheck `json:"healthcheck,omitempty"`
	Dependencies     []string                   `json:"dependencies"`
	ConfigFiles      []models.ServiceConfigFile `json:"config_files"`
	OverridesApplied overrideMetadata           `json:"overrides_applied"`
}

type overrideMetadata struct {
	Ports        bool `json:"ports"`
	Volumes      bool `json:"volumes"`
	EnvVars      bool `json:"env_vars"`
	EnvFiles     bool `json:"env_files"`
	Networks     bool `json:"networks"`
	ConfigMounts bool `json:"config_mounts"`
	Labels       bool `json:"labels"`
	Domains      bool `json:"domains"`
	Healthcheck  bool `json:"healthcheck"`
	Dependencies bool `json:"dependencies"`
	ConfigFiles  bool `json:"config_files"`
}

// EffectiveConfig godoc
// @Summary      Get effective instance configuration (template merged with overrides)
// @Tags         instances
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Success      200 {object} respond.SuccessEnvelope{data=effectiveConfigResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/effective-config [get]
// @Security     ApiKeyAuth
func (h *InstanceHandler) EffectiveConfig(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	var instanceID, stackID, templateServiceID int
	var command, userSpec sql.NullString
	err := h.db.QueryRow(`
		SELECT si.id, si.stack_id, si.template_service_id
		FROM service_instances si
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.instance_id = $2 AND si.deleted_at IS NULL AND st.deleted_at IS NULL
	`, stackName, instanceName).Scan(&instanceID, &stackID, &templateServiceID)

	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var resp effectiveConfigResponse
	resp.InstanceID = instanceName
	resp.StackName = stackName

	err = h.db.QueryRow(`
		SELECT s.name, s.image_name, s.image_tag, s.restart_policy, s.command, s.user_spec
		FROM services s WHERE s.id = $1
	`, templateServiceID).Scan(&resp.TemplateName, &resp.ImageName, &resp.ImageTag, &resp.RestartPolicy, &command, &userSpec)

	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get template service: %w", err))
		return
	}

	if command.Valid {
		resp.Command = command.String
	}
	if userSpec.Valid {
		resp.UserSpec = userSpec.String
	}

	templatePorts, err := h.loadServicePorts(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template ports: %w", err))
		return
	}

	instancePorts, err := h.loadInstancePorts(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance ports: %w", err))
		return
	}

	if len(instancePorts) > 0 {
		resp.Ports = instancePorts
		resp.OverridesApplied.Ports = true
	} else {
		resp.Ports = templatePorts
	}

	templateVolumes, err := h.loadServiceVolumes(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template volumes: %w", err))
		return
	}

	instanceVolumes, err := h.loadInstanceVolumes(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance volumes: %w", err))
		return
	}

	if len(instanceVolumes) > 0 {
		resp.Volumes = instanceVolumes
		resp.OverridesApplied.Volumes = true
	} else {
		resp.Volumes = templateVolumes
	}

	templateEnvVars, err := h.loadServiceEnvVars(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template env vars: %w", err))
		return
	}

	wiredEnvVars, err := h.loadWiredEnvVarsForEffective(instanceID, stackName)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load wired env vars: %w", err))
		return
	}

	instanceEnvVars, err := h.loadInstanceEnvVars(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance env vars: %w", err))
		return
	}

	resp.EnvVars = mergeEnvVarsThreeLayer(templateEnvVars, wiredEnvVars, instanceEnvVars)
	resp.WiredEnvVars = wiredEnvVars
	resp.OverridesApplied.EnvVars = len(instanceEnvVars) > 0

	templateLabels, err := h.loadServiceLabels(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template labels: %w", err))
		return
	}

	instanceLabels, err := h.loadInstanceLabels(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance labels: %w", err))
		return
	}

	resp.Labels = mergeLabels(templateLabels, instanceLabels)
	resp.OverridesApplied.Labels = len(instanceLabels) > 0

	// Inject identity labels from container package (NETW-04 requirement)
	// User overrides take precedence — only add if not already present
	identityLabels := container.BuildLabels(stackName, instanceName, strconv.Itoa(templateServiceID))
	existingKeys := make(map[string]bool)
	for _, l := range resp.Labels {
		existingKeys[l.Key] = true
	}
	for key, value := range identityLabels {
		if !existingKeys[key] {
			resp.Labels = append(resp.Labels, models.ServiceLabel{Key: key, Value: value})
		}
	}

	templateDomains, err := h.loadServiceDomains(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template domains: %w", err))
		return
	}

	instanceDomains, err := h.loadInstanceDomains(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance domains: %w", err))
		return
	}

	if len(instanceDomains) > 0 {
		resp.Domains = instanceDomains
		resp.OverridesApplied.Domains = true
	} else {
		resp.Domains = templateDomains
	}

	templateHealthcheck, err := h.loadServiceHealthcheck(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template healthcheck: %w", err))
		return
	}

	instanceHealthcheck, err := h.loadInstanceHealthcheck(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance healthcheck: %w", err))
		return
	}

	if instanceHealthcheck != nil {
		resp.Healthcheck = instanceHealthcheck
		resp.OverridesApplied.Healthcheck = true
	} else {
		resp.Healthcheck = templateHealthcheck
	}

	templateDeps, err := h.loadServiceDependencies(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template dependencies: %w", err))
		return
	}

	instanceDeps, err := h.loadInstanceDependencies(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance dependencies: %w", err))
		return
	}

	if len(instanceDeps) > 0 {
		resp.Dependencies = instanceDeps
		resp.OverridesApplied.Dependencies = true
	} else {
		resp.Dependencies = templateDeps
	}

	templateConfigFiles, err := h.loadServiceConfigFiles(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template config files: %w", err))
		return
	}

	instanceConfigFiles, err := h.loadInstanceConfigFiles(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance config files: %w", err))
		return
	}

	resp.ConfigFiles = mergeConfigFiles(templateConfigFiles, instanceConfigFiles)
	resp.OverridesApplied.ConfigFiles = len(instanceConfigFiles) > 0

	templateEnvFiles, err := h.loadServiceEnvFiles(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template env_files: %w", err))
		return
	}

	instanceEnvFiles, err := h.loadInstanceEnvFiles(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance env_files: %w", err))
		return
	}

	if len(instanceEnvFiles) > 0 {
		resp.EnvFiles = instanceEnvFiles
		resp.OverridesApplied.EnvFiles = true
	} else {
		resp.EnvFiles = templateEnvFiles
	}

	templateNetworks, err := h.loadServiceNetworks(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template networks: %w", err))
		return
	}

	instanceNetworks, err := h.loadInstanceNetworks(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance networks: %w", err))
		return
	}

	if len(instanceNetworks) > 0 {
		resp.Networks = instanceNetworks
		resp.OverridesApplied.Networks = true
	} else {
		resp.Networks = templateNetworks
	}

	templateConfigMounts, err := h.loadServiceConfigMounts(templateServiceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load template config_mounts: %w", err))
		return
	}

	instanceConfigMounts, err := h.loadInstanceConfigMounts(instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to load instance config_mounts: %w", err))
		return
	}

	if len(instanceConfigMounts) > 0 {
		resp.ConfigMounts = instanceConfigMounts
		resp.OverridesApplied.ConfigMounts = true
	} else {
		resp.ConfigMounts = templateConfigMounts
	}

	respond.JSON(w, r, http.StatusOK, resp)
}

func (h *InstanceHandler) loadServicePorts(serviceID int) ([]models.ServicePort, error) {
	rows, err := h.db.Query(`
		SELECT id, service_id, host_ip, host_port, container_port, protocol
		FROM service_ports WHERE service_id = $1 ORDER BY container_port
	`, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []models.ServicePort
	for rows.Next() {
		var p models.ServicePort
		if err := rows.Scan(&p.ID, &p.ServiceID, &p.HostIP, &p.HostPort, &p.ContainerPort, &p.Protocol); err != nil {
			return nil, err
		}
		ports = append(ports, p)
	}
	return ports, rows.Err()
}

func (h *InstanceHandler) loadInstancePorts(instanceID int) ([]models.ServicePort, error) {
	rows, err := h.db.Query(`
		SELECT id, instance_id, host_ip, host_port, container_port, protocol
		FROM instance_ports WHERE instance_id = $1 ORDER BY container_port
	`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []models.ServicePort
	for rows.Next() {
		var p models.ServicePort
		if err := rows.Scan(&p.ID, &p.ServiceID, &p.HostIP, &p.HostPort, &p.ContainerPort, &p.Protocol); err != nil {
			return nil, err
		}
		ports = append(ports, p)
	}
	return ports, rows.Err()
}

func (h *InstanceHandler) loadServiceVolumes(serviceID int) ([]models.ServiceVolume, error) {
	rows, err := h.db.Query(`
		SELECT id, service_id, volume_type, source, target, read_only, is_external
		FROM service_volumes WHERE service_id = $1 ORDER BY target
	`, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var volumes []models.ServiceVolume
	for rows.Next() {
		var v models.ServiceVolume
		if err := rows.Scan(&v.ID, &v.ServiceID, &v.VolumeType, &v.Source, &v.Target, &v.ReadOnly, &v.IsExternal); err != nil {
			return nil, err
		}
		volumes = append(volumes, v)
	}
	return volumes, rows.Err()
}

func (h *InstanceHandler) loadInstanceVolumes(instanceID int) ([]models.ServiceVolume, error) {
	rows, err := h.db.Query(`
		SELECT id, instance_id, volume_type, source, target, read_only, is_external
		FROM instance_volumes WHERE instance_id = $1 ORDER BY target
	`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var volumes []models.ServiceVolume
	for rows.Next() {
		var v models.ServiceVolume
		if err := rows.Scan(&v.ID, &v.ServiceID, &v.VolumeType, &v.Source, &v.Target, &v.ReadOnly, &v.IsExternal); err != nil {
			return nil, err
		}
		volumes = append(volumes, v)
	}
	return volumes, rows.Err()
}

func (h *InstanceHandler) loadServiceEnvVars(serviceID int) ([]models.ServiceEnvVar, error) {
	rows, err := h.db.Query(`
		SELECT id, service_id, key, value, is_secret, encrypted_value, encryption_version
		FROM service_env_vars WHERE service_id = $1 ORDER BY key
	`, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []models.ServiceEnvVar
	for rows.Next() {
		var e models.ServiceEnvVar
		var encryptedValue sql.NullString
		var encryptionVersion int
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.Key, &e.Value, &e.IsSecret, &encryptedValue, &encryptionVersion); err != nil {
			return nil, err
		}

		if encryptionVersion > 0 && encryptedValue.Valid {
			e.Value = "***"
		} else if e.IsSecret {
			e.Value = "***"
		}

		envVars = append(envVars, e)
	}
	return envVars, rows.Err()
}

func (h *InstanceHandler) loadInstanceEnvVars(instanceID int) ([]models.ServiceEnvVar, error) {
	rows, err := h.db.Query(`
		SELECT id, instance_id, key, value, is_secret, encrypted_value, encryption_version
		FROM instance_env_vars WHERE instance_id = $1 ORDER BY key
	`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []models.ServiceEnvVar
	for rows.Next() {
		var e models.ServiceEnvVar
		var encryptedValue sql.NullString
		var encryptionVersion int
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.Key, &e.Value, &e.IsSecret, &encryptedValue, &encryptionVersion); err != nil {
			return nil, err
		}

		if encryptionVersion > 0 && encryptedValue.Valid {
			e.Value = "***"
		} else if e.IsSecret {
			e.Value = "***"
		}

		envVars = append(envVars, e)
	}
	return envVars, rows.Err()
}

func (h *InstanceHandler) loadServiceLabels(serviceID int) ([]models.ServiceLabel, error) {
	rows, err := h.db.Query(`
		SELECT id, service_id, key, value
		FROM service_labels WHERE service_id = $1 ORDER BY key
	`, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []models.ServiceLabel
	for rows.Next() {
		var l models.ServiceLabel
		if err := rows.Scan(&l.ID, &l.ServiceID, &l.Key, &l.Value); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

func (h *InstanceHandler) loadInstanceLabels(instanceID int) ([]models.ServiceLabel, error) {
	rows, err := h.db.Query(`
		SELECT id, instance_id, key, value
		FROM instance_labels WHERE instance_id = $1 ORDER BY key
	`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []models.ServiceLabel
	for rows.Next() {
		var l models.ServiceLabel
		if err := rows.Scan(&l.ID, &l.ServiceID, &l.Key, &l.Value); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

func (h *InstanceHandler) loadServiceDomains(serviceID int) ([]models.ServiceDomain, error) {
	rows, err := h.db.Query(`
		SELECT id, service_id, domain, proxy_port
		FROM service_domains WHERE service_id = $1 ORDER BY domain
	`, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []models.ServiceDomain
	for rows.Next() {
		var d models.ServiceDomain
		if err := rows.Scan(&d.ID, &d.ServiceID, &d.Domain, &d.ProxyPort); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (h *InstanceHandler) loadInstanceDomains(instanceID int) ([]models.ServiceDomain, error) {
	rows, err := h.db.Query(`
		SELECT id, instance_id, domain, proxy_port
		FROM instance_domains WHERE instance_id = $1 ORDER BY domain
	`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []models.ServiceDomain
	for rows.Next() {
		var d models.ServiceDomain
		if err := rows.Scan(&d.ID, &d.ServiceID, &d.Domain, &d.ProxyPort); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (h *InstanceHandler) loadServiceHealthcheck(serviceID int) (*models.ServiceHealthcheck, error) {
	var hc models.ServiceHealthcheck
	err := h.db.QueryRow(`
		SELECT id, service_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds
		FROM service_healthchecks WHERE service_id = $1
	`, serviceID).Scan(&hc.ID, &hc.ServiceID, &hc.Test, &hc.IntervalSeconds, &hc.TimeoutSeconds, &hc.Retries, &hc.StartPeriodSeconds)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &hc, nil
}

func (h *InstanceHandler) loadInstanceHealthcheck(instanceID int) (*models.ServiceHealthcheck, error) {
	var hc models.ServiceHealthcheck
	err := h.db.QueryRow(`
		SELECT id, instance_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds
		FROM instance_healthchecks WHERE instance_id = $1
	`, instanceID).Scan(&hc.ID, &hc.ServiceID, &hc.Test, &hc.IntervalSeconds, &hc.TimeoutSeconds, &hc.Retries, &hc.StartPeriodSeconds)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &hc, nil
}

func (h *InstanceHandler) loadServiceDependencies(serviceID int) ([]string, error) {
	rows, err := h.db.Query(`
		SELECT s.name
		FROM service_dependencies sd
		JOIN services s ON s.id = sd.depends_on_service_id
		WHERE sd.service_id = $1
		ORDER BY s.name
	`, serviceID)
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

func (h *InstanceHandler) loadInstanceDependencies(instanceID int) ([]string, error) {
	rows, err := h.db.Query(`
		SELECT depends_on FROM instance_dependencies WHERE instance_id = $1 ORDER BY depends_on
	`, instanceID)
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

func (h *InstanceHandler) loadServiceConfigFiles(serviceID int) ([]models.ServiceConfigFile, error) {
	rows, err := h.db.Query(`
		SELECT id, service_id, file_path, content, file_mode, is_template, created_at, updated_at
		FROM service_config_files WHERE service_id = $1 ORDER BY file_path
	`, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.ServiceConfigFile
	for rows.Next() {
		var f models.ServiceConfigFile
		if err := rows.Scan(&f.ID, &f.ServiceID, &f.FilePath, &f.Content, &f.FileMode, &f.IsTemplate, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func (h *InstanceHandler) loadInstanceConfigFiles(instanceID int) ([]models.ServiceConfigFile, error) {
	rows, err := h.db.Query(`
		SELECT id, instance_id, file_path, content, file_mode, is_template, created_at, updated_at
		FROM instance_config_files WHERE instance_id = $1 ORDER BY file_path
	`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.ServiceConfigFile
	for rows.Next() {
		var f models.ServiceConfigFile
		if err := rows.Scan(&f.ID, &f.ServiceID, &f.FilePath, &f.Content, &f.FileMode, &f.IsTemplate, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func mergeEnvVars(template, instance []models.ServiceEnvVar) []models.ServiceEnvVar {
	merged := make(map[string]models.ServiceEnvVar)
	for _, e := range template {
		merged[e.Key] = e
	}
	for _, e := range instance {
		merged[e.Key] = e
	}

	result := make([]models.ServiceEnvVar, 0, len(merged))
	for _, e := range merged {
		result = append(result, e)
	}
	return result
}

func mergeEnvVarsThreeLayer(template []models.ServiceEnvVar, wired map[string]string, instance []models.ServiceEnvVar) []models.ServiceEnvVar {
	merged := make(map[string]models.ServiceEnvVar)
	for _, e := range template {
		merged[e.Key] = e
	}
	for k, v := range wired {
		merged[k] = models.ServiceEnvVar{Key: k, Value: v}
	}
	for _, e := range instance {
		merged[e.Key] = e
	}

	result := make([]models.ServiceEnvVar, 0, len(merged))
	for _, e := range merged {
		result = append(result, e)
	}
	return result
}

func (h *InstanceHandler) loadWiredEnvVarsForEffective(instanceID int, stackName string) (map[string]string, error) {
	rows, err := h.db.Query(`
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
	`, instanceID)
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

func mergeLabels(template, instance []models.ServiceLabel) []models.ServiceLabel {
	merged := make(map[string]models.ServiceLabel)
	for _, l := range template {
		merged[l.Key] = l
	}
	for _, l := range instance {
		merged[l.Key] = l
	}

	result := make([]models.ServiceLabel, 0, len(merged))
	for _, l := range merged {
		result = append(result, l)
	}
	return result
}

func mergeConfigFiles(template, instance []models.ServiceConfigFile) []models.ServiceConfigFile {
	merged := make(map[string]models.ServiceConfigFile)
	for _, f := range template {
		merged[f.FilePath] = f
	}
	for _, f := range instance {
		merged[f.FilePath] = f
	}

	result := make([]models.ServiceConfigFile, 0, len(merged))
	for _, f := range merged {
		result = append(result, f)
	}
	return result
}

func (h *InstanceHandler) loadServiceEnvFiles(serviceID int) ([]string, error) {
	rows, err := h.db.Query(`
		SELECT path FROM service_env_files WHERE service_id = $1 ORDER BY sort_order
	`, serviceID)
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

func (h *InstanceHandler) loadInstanceEnvFiles(instanceID int) ([]string, error) {
	rows, err := h.db.Query(`
		SELECT path FROM instance_env_files WHERE instance_id = $1 ORDER BY sort_order
	`, instanceID)
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

func (h *InstanceHandler) loadServiceNetworks(serviceID int) ([]string, error) {
	rows, err := h.db.Query(`
		SELECT network_name FROM service_networks WHERE service_id = $1 ORDER BY network_name
	`, serviceID)
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
	return networks, rows.Err()
}

func (h *InstanceHandler) loadInstanceNetworks(instanceID int) ([]string, error) {
	rows, err := h.db.Query(`
		SELECT network_name FROM instance_networks WHERE instance_id = $1 ORDER BY network_name
	`, instanceID)
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
	return networks, rows.Err()
}

func (h *InstanceHandler) loadServiceConfigMounts(serviceID int) ([]models.ServiceConfigMount, error) {
	rows, err := h.db.Query(`
		SELECT id, service_id, config_file_id, source_path, target_path, readonly
		FROM service_config_mounts WHERE service_id = $1 ORDER BY target_path
	`, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mounts []models.ServiceConfigMount
	for rows.Next() {
		var m models.ServiceConfigMount
		var cfgFileID sql.NullInt32
		if err := rows.Scan(&m.ID, &m.ServiceID, &cfgFileID, &m.SourcePath, &m.TargetPath, &m.ReadOnly); err != nil {
			return nil, err
		}
		if cfgFileID.Valid {
			val := int(cfgFileID.Int32)
			m.ConfigFileID = &val
		}
		mounts = append(mounts, m)
	}
	return mounts, rows.Err()
}

func (h *InstanceHandler) loadInstanceConfigMounts(instanceID int) ([]models.ServiceConfigMount, error) {
	rows, err := h.db.Query(`
		SELECT id, config_file_id, source_path, target_path, readonly
		FROM instance_config_mounts WHERE instance_id = $1 ORDER BY target_path
	`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mounts []models.ServiceConfigMount
	for rows.Next() {
		var m models.ServiceConfigMount
		var cfgFileID sql.NullInt32
		if err := rows.Scan(&m.ID, &cfgFileID, &m.SourcePath, &m.TargetPath, &m.ReadOnly); err != nil {
			return nil, err
		}
		if cfgFileID.Valid {
			val := int(cfgFileID.Int32)
			m.ConfigFileID = &val
		}
		mounts = append(mounts, m)
	}
	return mounts, rows.Err()
}
