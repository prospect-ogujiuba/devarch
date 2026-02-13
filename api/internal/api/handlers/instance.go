package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/crypto"
	"github.com/priz/devarch-api/internal/identity"
	"github.com/priz/devarch-api/pkg/models"
)

type InstanceHandler struct {
	db              *sql.DB
	containerClient *container.Client
	cipher          *crypto.Cipher
}

func NewInstanceHandler(db *sql.DB, cc *container.Client, cipher *crypto.Cipher) *InstanceHandler {
	return &InstanceHandler{
		db:              db,
		containerClient: cc,
		cipher:          cipher,
	}
}

type instanceResponse struct {
	ID                int       `json:"id"`
	StackID           int       `json:"stack_id"`
	InstanceID        string    `json:"instance_id"`
	TemplateServiceID int       `json:"template_service_id"`
	TemplateName      string    `json:"template_name"`
	ContainerName     string    `json:"container_name"`
	Description       string    `json:"description"`
	Enabled           bool      `json:"enabled"`
	Running           bool      `json:"running"`
	OverrideCount     int       `json:"override_count"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type dependencyEntry struct {
	ID         int    `json:"id"`
	InstanceID int    `json:"instance_id"`
	DependsOn  string `json:"depends_on"`
	Condition  string `json:"condition"`
}

type instanceDetailResponse struct {
	instanceResponse
	Ports        []models.ServicePort       `json:"ports"`
	Volumes      []models.ServiceVolume     `json:"volumes"`
	EnvVars      []models.ServiceEnvVar     `json:"env_vars"`
	EnvFiles     []string                   `json:"env_files"`
	Networks     []string                   `json:"networks"`
	ConfigMounts []models.ServiceConfigMount `json:"config_mounts"`
	Labels       []models.ServiceLabel      `json:"labels"`
	Domains      []models.ServiceDomain     `json:"domains"`
	Healthcheck  *models.ServiceHealthcheck `json:"healthcheck"`
	Dependencies []dependencyEntry          `json:"dependencies"`
	ConfigFiles  []models.ServiceConfigFile `json:"config_files"`
}

type createInstanceRequest struct {
	InstanceID        string `json:"instance_id"`
	TemplateServiceID int    `json:"template_service_id"`
	Description       string `json:"description"`
}

type updateInstanceRequest struct {
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
}

type duplicateInstanceRequest struct {
	InstanceID string `json:"instance_id"`
}

type renameInstanceRequest struct {
	InstanceID string `json:"instance_id"`
}

type instanceDeletePreviewResponse struct {
	InstanceName  string `json:"instance_name"`
	TemplateName  string `json:"template_name"`
	OverrideCount int    `json:"override_count"`
	ContainerName string `json:"container_name"`
}

func (h *InstanceHandler) getStackByName(stackName string) (int, string, error) {
	var stackID int
	var stackActualName string
	err := h.db.QueryRow(`
		SELECT id, name FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID, &stackActualName)
	return stackID, stackActualName, err
}

// Create godoc
// @Summary      Create instance in stack
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        body body createInstanceRequest true "Instance creation request"
// @Success      201 {object} respond.SuccessEnvelope{data=instanceResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances [post]
// @Security     ApiKeyAuth
func (h *InstanceHandler) Create(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var req createInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if err := identity.ValidateName(req.InstanceID); err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	stackID, stackActualName, err := h.getStackByName(stackName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "stack", stackName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get stack: %w", err))
		return
	}

	var templateExists bool
	err = h.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM services WHERE id = $1)`, req.TemplateServiceID).Scan(&templateExists)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to check template: %w", err))
		return
	}
	if !templateExists {
		respond.BadRequest(w, r, fmt.Sprintf("template service ID %d not found", req.TemplateServiceID))
		return
	}

	// Validate combined container name length
	if err := identity.ValidateContainerName(stackActualName, req.InstanceID); err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	containerName := identity.ContainerName(stackActualName, req.InstanceID)

	var instance instanceResponse
	err = h.db.QueryRow(`
		INSERT INTO service_instances (stack_id, instance_id, template_service_id, container_name, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, stack_id, instance_id, template_service_id, container_name, description, enabled, created_at, updated_at
	`, stackID, req.InstanceID, req.TemplateServiceID, containerName, req.Description).Scan(
		&instance.ID,
		&instance.StackID,
		&instance.InstanceID,
		&instance.TemplateServiceID,
		&instance.ContainerName,
		&instance.Description,
		&instance.Enabled,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			respond.Conflict(w, r, fmt.Sprintf("instance %q already exists in stack %q", req.InstanceID, stackName))
			return
		}
		respond.InternalError(w, r, fmt.Errorf("failed to create instance: %w", err))
		return
	}

	err = h.db.QueryRow(`SELECT name FROM services WHERE id = $1`, instance.TemplateServiceID).Scan(&instance.TemplateName)
	if err != nil {
		instance.TemplateName = ""
	}

	instance.OverrideCount = 0

	respond.JSON(w, r, http.StatusCreated, instance)
}

// List godoc
// @Summary      List instances in stack
// @Tags         instances
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        search query string false "Search by instance ID"
// @Param        enabled query boolean false "Filter by enabled status"
// @Success      200 {object} respond.SuccessEnvelope{data=[]instanceResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances [get]
// @Security     ApiKeyAuth
func (h *InstanceHandler) List(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	stackID, _, err := h.getStackByName(stackName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "stack", stackName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get stack: %w", err))
		return
	}

	search := r.URL.Query().Get("search")
	enabledParam := r.URL.Query().Get("enabled")

	query := `
		SELECT
			si.id,
			si.stack_id,
			si.instance_id,
			si.template_service_id,
			s.name as template_name,
			si.container_name,
			si.description,
			si.enabled,
			si.created_at,
			si.updated_at,
			COALESCE(oc.total, 0) as override_count
		FROM service_instances si
		JOIN services s ON s.id = si.template_service_id
		LEFT JOIN (
			SELECT instance_id, COUNT(*) as total
			FROM (
				SELECT instance_id FROM instance_ports
				UNION ALL
				SELECT instance_id FROM instance_volumes
				UNION ALL
				SELECT instance_id FROM instance_env_vars
				UNION ALL
				SELECT instance_id FROM instance_labels
				UNION ALL
				SELECT instance_id FROM instance_domains
				UNION ALL
				SELECT instance_id FROM instance_healthchecks
				UNION ALL
				SELECT instance_id FROM instance_dependencies
				UNION ALL
				SELECT instance_id FROM instance_config_files
				UNION ALL
				SELECT instance_id FROM instance_env_files
				UNION ALL
				SELECT instance_id FROM instance_networks
				UNION ALL
				SELECT instance_id FROM instance_config_mounts
			) all_overrides
			GROUP BY instance_id
		) oc ON oc.instance_id = si.id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
	`

	args := []interface{}{stackID}
	argPos := 2

	if search != "" {
		query += fmt.Sprintf(" AND si.instance_id ILIKE $%d", argPos)
		args = append(args, "%"+search+"%")
		argPos++
	}

	if enabledParam == "true" {
		query += " AND si.enabled = true"
	} else if enabledParam == "false" {
		query += " AND si.enabled = false"
	}

	query += " ORDER BY si.created_at ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to query instances: %w", err))
		return
	}
	defer rows.Close()

	instances := []instanceResponse{}
	for rows.Next() {
		var inst instanceResponse
		err := rows.Scan(
			&inst.ID,
			&inst.StackID,
			&inst.InstanceID,
			&inst.TemplateServiceID,
			&inst.TemplateName,
			&inst.ContainerName,
			&inst.Description,
			&inst.Enabled,
			&inst.CreatedAt,
			&inst.UpdatedAt,
			&inst.OverrideCount,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to scan instance: %w", err))
			return
		}
		instances = append(instances, inst)
	}

	if err := rows.Err(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("error iterating instances: %w", err))
		return
	}

	runningNames, _ := h.containerClient.ListRunningContainersWithLabels(map[string]string{
		identity.LabelStackID: stackName,
	})
	runningSet := make(map[string]bool, len(runningNames))
	for _, n := range runningNames {
		runningSet[n] = true
	}
	for i := range instances {
		if instances[i].ContainerName != "" {
			instances[i].Running = runningSet[instances[i].ContainerName]
		}
	}

	respond.JSON(w, r, http.StatusOK, instances)
}

// Get godoc
// @Summary      Get instance details
// @Tags         instances
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Success      200 {object} respond.SuccessEnvelope{data=instanceDetailResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance} [get]
// @Security     ApiKeyAuth
func (h *InstanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	var instance instanceResponse
	err := h.db.QueryRow(`
		SELECT
			si.id,
			si.stack_id,
			si.instance_id,
			si.template_service_id,
			s.name as template_name,
			si.container_name,
			si.description,
			si.enabled,
			si.created_at,
			si.updated_at,
			(
				SELECT COUNT(*) FROM instance_ports WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_volumes WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_env_vars WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_labels WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_domains WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_healthchecks WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_dependencies WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_config_files WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_env_files WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_networks WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_config_mounts WHERE instance_id = si.id
			) as override_count
		FROM service_instances si
		JOIN services s ON s.id = si.template_service_id
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.instance_id = $2 AND si.deleted_at IS NULL AND st.deleted_at IS NULL
	`, stackName, instanceName).Scan(
		&instance.ID,
		&instance.StackID,
		&instance.InstanceID,
		&instance.TemplateServiceID,
		&instance.TemplateName,
		&instance.ContainerName,
		&instance.Description,
		&instance.Enabled,
		&instance.CreatedAt,
		&instance.UpdatedAt,
		&instance.OverrideCount,
	)

	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	runningNames, _ := h.containerClient.ListRunningContainersWithLabels(map[string]string{
		identity.LabelStackID: stackName,
	})
	for _, n := range runningNames {
		if n == instance.ContainerName {
			instance.Running = true
			break
		}
	}

	// Load instance override data
	detail := instanceDetailResponse{
		instanceResponse: instance,
		Ports:        []models.ServicePort{},
		Volumes:      []models.ServiceVolume{},
		EnvVars:      []models.ServiceEnvVar{},
		EnvFiles:     []string{},
		Networks:     []string{},
		ConfigMounts: []models.ServiceConfigMount{},
		Labels:       []models.ServiceLabel{},
		Domains:      []models.ServiceDomain{},
		Dependencies: []string{},
		ConfigFiles:  []models.ServiceConfigFile{},
	}

	if ports, err := h.loadInstancePorts(instance.ID); err == nil && len(ports) > 0 {
		detail.Ports = ports
	}
	if volumes, err := h.loadInstanceVolumes(instance.ID); err == nil && len(volumes) > 0 {
		detail.Volumes = volumes
	}
	if envVars, err := h.loadInstanceEnvVars(instance.ID); err == nil && len(envVars) > 0 {
		detail.EnvVars = envVars
	}
	if envFiles, err := h.loadInstanceEnvFiles(instance.ID); err == nil && len(envFiles) > 0 {
		detail.EnvFiles = envFiles
	}
	if networks, err := h.loadInstanceNetworks(instance.ID); err == nil && len(networks) > 0 {
		detail.Networks = networks
	}
	if configMounts, err := h.loadInstanceConfigMounts(instance.ID); err == nil && len(configMounts) > 0 {
		detail.ConfigMounts = configMounts
	}
	if labels, err := h.loadInstanceLabels(instance.ID); err == nil && len(labels) > 0 {
		detail.Labels = labels
	}
	if domains, err := h.loadInstanceDomains(instance.ID); err == nil && len(domains) > 0 {
		detail.Domains = domains
	}
	if hc, err := h.loadInstanceHealthcheck(instance.ID); err == nil && hc != nil {
		detail.Healthcheck = hc
	}
	if deps, err := h.loadInstanceDependencies(instance.ID); err == nil && len(deps) > 0 {
		detail.Dependencies = deps
	}
	if configFiles, err := h.loadInstanceConfigFiles(instance.ID); err == nil && len(configFiles) > 0 {
		detail.ConfigFiles = configFiles
	}

	respond.JSON(w, r, http.StatusOK, detail)
}

// Update godoc
// @Summary      Update instance
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body updateInstanceRequest true "Instance update request"
// @Success      200 {object} respond.SuccessEnvelope{data=instanceResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance} [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) Update(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	var req updateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	var instanceID int
	err := h.db.QueryRow(`
		SELECT si.id
		FROM service_instances si
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.instance_id = $2 AND si.deleted_at IS NULL AND st.deleted_at IS NULL
	`, stackName, instanceName).Scan(&instanceID)

	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argPos))
		args = append(args, *req.Description)
		argPos++
	}

	if req.Enabled != nil {
		updates = append(updates, fmt.Sprintf("enabled = $%d", argPos))
		args = append(args, *req.Enabled)
		argPos++
	}

	if len(updates) == 0 {
		respond.BadRequest(w, r, "no fields to update")
		return
	}

	updates = append(updates, "updated_at = NOW()")
	args = append(args, instanceID)

	query := fmt.Sprintf(`
		UPDATE service_instances
		SET %s
		WHERE id = $%d
		RETURNING id, stack_id, instance_id, template_service_id, container_name, description, enabled, created_at, updated_at
	`, join(updates, ", "), argPos)

	var instance instanceResponse
	err = h.db.QueryRow(query, args...).Scan(
		&instance.ID,
		&instance.StackID,
		&instance.InstanceID,
		&instance.TemplateServiceID,
		&instance.ContainerName,
		&instance.Description,
		&instance.Enabled,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)

	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to update instance: %w", err))
		return
	}

	if req.Enabled != nil && !*req.Enabled {
		if err := h.containerClient.StopContainer(instance.ContainerName); err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to stop instance: %w", err))
			return
		}
	}

	err = h.db.QueryRow(`SELECT name FROM services WHERE id = $1`, instance.TemplateServiceID).Scan(&instance.TemplateName)
	if err != nil {
		instance.TemplateName = ""
	}

	err = h.db.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM instance_ports WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_volumes WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_env_vars WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_env_files WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_networks WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_labels WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_domains WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_healthchecks WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_dependencies WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_config_files WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_config_mounts WHERE instance_id = $1)
	`, instance.ID).Scan(&instance.OverrideCount)
	if err != nil {
		instance.OverrideCount = 0
	}

	respond.JSON(w, r, http.StatusOK, instance)
}

// Delete godoc
// @Summary      Delete instance (soft delete)
// @Tags         instances
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Success      204
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance} [delete]
// @Security     ApiKeyAuth
func (h *InstanceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	result, err := h.db.Exec(`
		UPDATE service_instances
		SET deleted_at = NOW(), updated_at = NOW()
		FROM stacks st
		WHERE service_instances.stack_id = st.id
		AND st.name = $1
		AND service_instances.instance_id = $2
		AND service_instances.deleted_at IS NULL
		AND st.deleted_at IS NULL
	`, stackName, instanceName)

	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete instance: %w", err))
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get rows affected: %w", err))
		return
	}

	if rowsAffected == 0 {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}

	respond.NoContent(w, r)
}

// Duplicate godoc
// @Summary      Duplicate instance with all overrides
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Source instance ID"
// @Param        body body duplicateInstanceRequest true "Duplicate request"
// @Success      201 {object} respond.SuccessEnvelope{data=instanceResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/duplicate [post]
// @Security     ApiKeyAuth
func (h *InstanceHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	sourceInstanceName := chi.URLParam(r, "instance")

	var req duplicateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	newInstanceID := req.InstanceID
	if newInstanceID == "" {
		newInstanceID = sourceInstanceName + "-copy"
	}

	if err := identity.ValidateName(newInstanceID); err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	_, stackActualName, err := h.getStackByName(stackName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "stack", stackName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get stack: %w", err))
		return
	}

	// Validate combined container name length
	if err := identity.ValidateContainerName(stackActualName, newInstanceID); err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to begin transaction: %w", err))
		return
	}
	defer tx.Rollback()

	containerName := identity.ContainerName(stackActualName, newInstanceID)

	var newInstance instanceResponse
	err = tx.QueryRow(`
		INSERT INTO service_instances (stack_id, instance_id, template_service_id, container_name, description, enabled)
		SELECT stack_id, $1, template_service_id, $2, description, enabled
		FROM service_instances si
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $3 AND si.instance_id = $4 AND si.deleted_at IS NULL AND st.deleted_at IS NULL
		RETURNING id, stack_id, instance_id, template_service_id, container_name, description, enabled, created_at, updated_at
	`, newInstanceID, containerName, stackName, sourceInstanceName).Scan(
		&newInstance.ID,
		&newInstance.StackID,
		&newInstance.InstanceID,
		&newInstance.TemplateServiceID,
		&newInstance.ContainerName,
		&newInstance.Description,
		&newInstance.Enabled,
		&newInstance.CreatedAt,
		&newInstance.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			respond.Conflict(w, r, fmt.Sprintf("instance %q already exists in stack %q", newInstanceID, stackName))
			return
		}
		if err == sql.ErrNoRows {
			respond.NotFound(w, r, "instance", sourceInstanceName)
			return
		}
		respond.InternalError(w, r, fmt.Errorf("failed to duplicate instance: %w", err))
		return
	}

	var sourceInstanceID int
	err = tx.QueryRow(`
		SELECT si.id FROM service_instances si
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.instance_id = $2 AND si.deleted_at IS NULL
	`, stackName, sourceInstanceName).Scan(&sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get source instance: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_ports (instance_id, host_ip, host_port, container_port, protocol)
		SELECT $1, host_ip, host_port, container_port, protocol FROM instance_ports WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy ports: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_volumes (instance_id, volume_type, source, target, read_only, is_external)
		SELECT $1, volume_type, source, target, read_only, is_external FROM instance_volumes WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy volumes: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_env_vars (instance_id, key, value, is_secret)
		SELECT $1, key, value, is_secret FROM instance_env_vars WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy env vars: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_labels (instance_id, key, value)
		SELECT $1, key, value FROM instance_labels WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy labels: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_domains (instance_id, domain, proxy_port)
		SELECT $1, domain, proxy_port FROM instance_domains WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy domains: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_healthchecks (instance_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds)
		SELECT $1, test, interval_seconds, timeout_seconds, retries, start_period_seconds FROM instance_healthchecks WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy healthcheck: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_dependencies (instance_id, depends_on, condition)
		SELECT $1, depends_on, condition FROM instance_dependencies WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy dependencies: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_config_files (instance_id, file_path, content, file_mode, is_template)
		SELECT $1, file_path, content, file_mode, is_template FROM instance_config_files WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy config files: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_env_files (instance_id, path, sort_order)
		SELECT $1, path, sort_order FROM instance_env_files WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy env files: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_networks (instance_id, network_name)
		SELECT $1, network_name FROM instance_networks WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy networks: %w", err))
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_config_mounts (instance_id, config_file_id, source_path, target_path, readonly)
		SELECT $1, config_file_id, source_path, target_path, readonly FROM instance_config_mounts WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to copy config mounts: %w", err))
		return
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	err = h.db.QueryRow(`SELECT name FROM services WHERE id = $1`, newInstance.TemplateServiceID).Scan(&newInstance.TemplateName)
	if err != nil {
		newInstance.TemplateName = ""
	}

	err = h.db.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM instance_ports WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_volumes WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_env_vars WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_env_files WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_networks WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_labels WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_domains WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_healthchecks WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_dependencies WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_config_files WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_config_mounts WHERE instance_id = $1)
	`, newInstance.ID).Scan(&newInstance.OverrideCount)
	if err != nil {
		newInstance.OverrideCount = 0
	}

		respond.JSON(w, r, http.StatusCreated, newInstance)
}

// Rename godoc
// @Summary      Rename instance
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Current instance ID"
// @Param        body body renameInstanceRequest true "Rename request"
// @Success      200 {object} respond.SuccessEnvelope{data=instanceResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/rename [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) Rename(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	oldInstanceName := chi.URLParam(r, "instance")

	var req renameInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if err := identity.ValidateName(req.InstanceID); err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	stackID, stackActualName, err := h.getStackByName(stackName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "stack", stackName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get stack: %w", err))
		return
	}

	var exists bool
	err = h.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM service_instances WHERE stack_id = $1 AND instance_id = $2 AND deleted_at IS NULL)`,
		stackID, req.InstanceID).Scan(&exists)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to check instance name: %w", err))
		return
	}
	if exists {
		respond.Conflict(w, r, fmt.Sprintf("instance %q already exists in stack %q", req.InstanceID, stackName))
		return
	}

	// Validate combined container name length
	if err := identity.ValidateContainerName(stackActualName, req.InstanceID); err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	newContainerName := identity.ContainerName(stackActualName, req.InstanceID)

	var instance instanceResponse
	err = h.db.QueryRow(`
		UPDATE service_instances
		SET instance_id = $1, container_name = $2, updated_at = NOW()
		FROM stacks st
		WHERE service_instances.stack_id = st.id
		AND st.name = $3
		AND service_instances.instance_id = $4
		AND service_instances.deleted_at IS NULL
		AND st.deleted_at IS NULL
		RETURNING service_instances.id, service_instances.stack_id, service_instances.instance_id,
			service_instances.template_service_id, service_instances.container_name,
			service_instances.description, service_instances.enabled,
			service_instances.created_at, service_instances.updated_at
	`, req.InstanceID, newContainerName, stackName, oldInstanceName).Scan(
		&instance.ID,
		&instance.StackID,
		&instance.InstanceID,
		&instance.TemplateServiceID,
		&instance.ContainerName,
		&instance.Description,
		&instance.Enabled,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", oldInstanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to rename instance: %w", err))
		return
	}

	err = h.db.QueryRow(`SELECT name FROM services WHERE id = $1`, instance.TemplateServiceID).Scan(&instance.TemplateName)
	if err != nil {
		instance.TemplateName = ""
	}

	err = h.db.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM instance_ports WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_volumes WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_env_vars WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_env_files WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_networks WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_labels WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_domains WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_healthchecks WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_dependencies WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_config_files WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_config_mounts WHERE instance_id = $1)
	`, instance.ID).Scan(&instance.OverrideCount)
	if err != nil {
		instance.OverrideCount = 0
	}

	respond.JSON(w, r, http.StatusOK, instance)
}

// DeletePreview godoc
// @Summary      Preview instance deletion impact
// @Tags         instances
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Success      200 {object} respond.SuccessEnvelope{data=instanceDeletePreviewResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/delete-preview [get]
// @Security     ApiKeyAuth
func (h *InstanceHandler) DeletePreview(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	var preview instanceDeletePreviewResponse
	var overrideCount int
	err := h.db.QueryRow(`
		SELECT
			si.instance_id,
			s.name,
			si.container_name,
			(
				SELECT COUNT(*) FROM instance_ports WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_volumes WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_env_vars WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_labels WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_domains WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_healthchecks WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_dependencies WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_config_files WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_env_files WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_networks WHERE instance_id = si.id
			) + (
				SELECT COUNT(*) FROM instance_config_mounts WHERE instance_id = si.id
			) as override_count
		FROM service_instances si
		JOIN services s ON s.id = si.template_service_id
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.instance_id = $2 AND si.deleted_at IS NULL AND st.deleted_at IS NULL
	`, stackName, instanceName).Scan(
		&preview.InstanceName,
		&preview.TemplateName,
		&preview.ContainerName,
		&overrideCount,
	)

	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get delete preview: %w", err))
		return
	}

	preview.OverrideCount = overrideCount

	respond.JSON(w, r, http.StatusOK, preview)
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
