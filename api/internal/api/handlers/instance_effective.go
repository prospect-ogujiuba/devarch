package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/pkg/models"
)

type effectiveConfigResponse struct {
	InstanceID       string                   `json:"instance_id"`
	StackName        string                   `json:"stack_name"`
	TemplateName     string                   `json:"template_name"`
	ImageName        string                   `json:"image_name"`
	ImageTag         string                   `json:"image_tag"`
	RestartPolicy    string                   `json:"restart_policy"`
	Command          string                   `json:"command,omitempty"`
	UserSpec         string                   `json:"user_spec,omitempty"`
	Ports            []models.ServicePort     `json:"ports"`
	Volumes          []models.ServiceVolume   `json:"volumes"`
	EnvVars          []models.ServiceEnvVar   `json:"env_vars"`
	Labels           []models.ServiceLabel    `json:"labels"`
	Domains          []models.ServiceDomain   `json:"domains"`
	Healthcheck      *models.ServiceHealthcheck `json:"healthcheck,omitempty"`
	Dependencies     []string                 `json:"dependencies"`
	ConfigFiles      []models.ServiceConfigFile `json:"config_files"`
	OverridesApplied overrideMetadata         `json:"overrides_applied"`
}

type overrideMetadata struct {
	Ports        bool `json:"ports"`
	Volumes      bool `json:"volumes"`
	EnvVars      bool `json:"env_vars"`
	Labels       bool `json:"labels"`
	Domains      bool `json:"domains"`
	Healthcheck  bool `json:"healthcheck"`
	Dependencies bool `json:"dependencies"`
	ConfigFiles  bool `json:"config_files"`
}

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
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("failed to get template service: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("failed to load template ports: %v", err), http.StatusInternalServerError)
		return
	}

	instancePorts, err := h.loadInstancePorts(instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load instance ports: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("failed to load template volumes: %v", err), http.StatusInternalServerError)
		return
	}

	instanceVolumes, err := h.loadInstanceVolumes(instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load instance volumes: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("failed to load template env vars: %v", err), http.StatusInternalServerError)
		return
	}

	instanceEnvVars, err := h.loadInstanceEnvVars(instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load instance env vars: %v", err), http.StatusInternalServerError)
		return
	}

	resp.EnvVars = mergeEnvVars(templateEnvVars, instanceEnvVars)
	resp.OverridesApplied.EnvVars = len(instanceEnvVars) > 0

	templateLabels, err := h.loadServiceLabels(templateServiceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load template labels: %v", err), http.StatusInternalServerError)
		return
	}

	instanceLabels, err := h.loadInstanceLabels(instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load instance labels: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("failed to load template domains: %v", err), http.StatusInternalServerError)
		return
	}

	instanceDomains, err := h.loadInstanceDomains(instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load instance domains: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("failed to load template healthcheck: %v", err), http.StatusInternalServerError)
		return
	}

	instanceHealthcheck, err := h.loadInstanceHealthcheck(instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load instance healthcheck: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("failed to load template dependencies: %v", err), http.StatusInternalServerError)
		return
	}

	instanceDeps, err := h.loadInstanceDependencies(instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load instance dependencies: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("failed to load template config files: %v", err), http.StatusInternalServerError)
		return
	}

	instanceConfigFiles, err := h.loadInstanceConfigFiles(instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load instance config files: %v", err), http.StatusInternalServerError)
		return
	}

	resp.ConfigFiles = mergeConfigFiles(templateConfigFiles, instanceConfigFiles)
	resp.OverridesApplied.ConfigFiles = len(instanceConfigFiles) > 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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
		SELECT id, service_id, key, value, is_secret
		FROM service_env_vars WHERE service_id = $1 ORDER BY key
	`, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []models.ServiceEnvVar
	for rows.Next() {
		var e models.ServiceEnvVar
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.Key, &e.Value, &e.IsSecret); err != nil {
			return nil, err
		}
		envVars = append(envVars, e)
	}
	return envVars, rows.Err()
}

func (h *InstanceHandler) loadInstanceEnvVars(instanceID int) ([]models.ServiceEnvVar, error) {
	rows, err := h.db.Query(`
		SELECT id, instance_id, key, value, is_secret
		FROM instance_env_vars WHERE instance_id = $1 ORDER BY key
	`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []models.ServiceEnvVar
	for rows.Next() {
		var e models.ServiceEnvVar
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.Key, &e.Value, &e.IsSecret); err != nil {
			return nil, err
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
