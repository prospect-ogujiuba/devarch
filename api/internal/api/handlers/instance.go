package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/priz/devarch-api/internal/container"
)

type InstanceHandler struct {
	db              *sql.DB
	containerClient *container.Client
}

func NewInstanceHandler(db *sql.DB, cc *container.Client) *InstanceHandler {
	return &InstanceHandler{
		db:              db,
		containerClient: cc,
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
	OverrideCount     int       `json:"override_count"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
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
	InstanceName   string `json:"instance_name"`
	TemplateName   string `json:"template_name"`
	OverrideCount  int    `json:"override_count"`
	ContainerName  string `json:"container_name"`
}

func (h *InstanceHandler) getStackByName(stackName string) (int, string, error) {
	var stackID int
	var stackActualName string
	err := h.db.QueryRow(`
		SELECT id, name FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID, &stackActualName)
	return stackID, stackActualName, err
}

func (h *InstanceHandler) Create(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var req createInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if err := container.ValidateName(req.InstanceID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stackID, stackActualName, err := h.getStackByName(stackName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
		return
	}

	var templateExists bool
	err = h.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM services WHERE id = $1)`, req.TemplateServiceID).Scan(&templateExists)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to check template: %v", err), http.StatusInternalServerError)
		return
	}
	if !templateExists {
		http.Error(w, fmt.Sprintf("template service ID %d not found", req.TemplateServiceID), http.StatusBadRequest)
		return
	}

	// Validate combined container name length
	if err := container.ValidateContainerName(stackActualName, req.InstanceID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	containerName := container.ContainerName(stackActualName, req.InstanceID)

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
			http.Error(w, fmt.Sprintf("instance %q already exists in stack %q", req.InstanceID, stackName), http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("failed to create instance: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.db.QueryRow(`SELECT name FROM services WHERE id = $1`, instance.TemplateServiceID).Scan(&instance.TemplateName)
	if err != nil {
		instance.TemplateName = ""
	}

	instance.OverrideCount = 0

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(instance)
}

func (h *InstanceHandler) List(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	stackID, _, err := h.getStackByName(stackName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
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
			) as override_count
		FROM service_instances si
		JOIN services s ON s.id = si.template_service_id
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
		http.Error(w, fmt.Sprintf("failed to query instances: %v", err), http.StatusInternalServerError)
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
			http.Error(w, fmt.Sprintf("failed to scan instance: %v", err), http.StatusInternalServerError)
			return
		}
		instances = append(instances, inst)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("error iterating instances: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instances)
}

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
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instance)
}

func (h *InstanceHandler) Update(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	var req updateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "no fields to update", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("failed to update instance: %v", err), http.StatusInternalServerError)
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
			(SELECT COUNT(*) FROM instance_labels WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_domains WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_healthchecks WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_dependencies WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_config_files WHERE instance_id = $1)
	`, instance.ID).Scan(&instance.OverrideCount)
	if err != nil {
		instance.OverrideCount = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instance)
}

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
		http.Error(w, fmt.Sprintf("failed to delete instance: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get rows affected: %v", err), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *InstanceHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	sourceInstanceName := chi.URLParam(r, "instance")

	var req duplicateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	newInstanceID := req.InstanceID
	if newInstanceID == "" {
		newInstanceID = sourceInstanceName + "-copy"
	}

	if err := container.ValidateName(newInstanceID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, stackActualName, err := h.getStackByName(stackName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
		return
	}

	// Validate combined container name length
	if err := container.ValidateContainerName(stackActualName, newInstanceID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	containerName := container.ContainerName(stackActualName, newInstanceID)

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
			http.Error(w, fmt.Sprintf("instance %q already exists in stack %q", newInstanceID, stackName), http.StatusConflict)
			return
		}
		if err == sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("source instance %q not found in stack %q", sourceInstanceName, stackName), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to duplicate instance: %v", err), http.StatusInternalServerError)
		return
	}

	var sourceInstanceID int
	err = tx.QueryRow(`
		SELECT si.id FROM service_instances si
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.instance_id = $2 AND si.deleted_at IS NULL
	`, stackName, sourceInstanceName).Scan(&sourceInstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get source instance: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_ports (instance_id, host_ip, host_port, container_port, protocol)
		SELECT $1, host_ip, host_port, container_port, protocol FROM instance_ports WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to copy ports: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_volumes (instance_id, volume_type, source, target, read_only, is_external)
		SELECT $1, volume_type, source, target, read_only, is_external FROM instance_volumes WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to copy volumes: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_env_vars (instance_id, key, value, is_secret)
		SELECT $1, key, value, is_secret FROM instance_env_vars WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to copy env vars: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_labels (instance_id, key, value)
		SELECT $1, key, value FROM instance_labels WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to copy labels: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_domains (instance_id, domain, proxy_port)
		SELECT $1, domain, proxy_port FROM instance_domains WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to copy domains: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_healthchecks (instance_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds)
		SELECT $1, test, interval_seconds, timeout_seconds, retries, start_period_seconds FROM instance_healthchecks WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to copy healthcheck: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_dependencies (instance_id, depends_on, condition)
		SELECT $1, depends_on, condition FROM instance_dependencies WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to copy dependencies: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO instance_config_files (instance_id, file_path, content, file_mode, is_template)
		SELECT $1, file_path, content, file_mode, is_template FROM instance_config_files WHERE instance_id = $2`, newInstance.ID, sourceInstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to copy config files: %v", err), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("failed to commit transaction: %v", err), http.StatusInternalServerError)
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
			(SELECT COUNT(*) FROM instance_labels WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_domains WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_healthchecks WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_dependencies WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_config_files WHERE instance_id = $1)
	`, newInstance.ID).Scan(&newInstance.OverrideCount)
	if err != nil {
		newInstance.OverrideCount = 0
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newInstance)
}

func (h *InstanceHandler) Rename(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	oldInstanceName := chi.URLParam(r, "instance")

	var req renameInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if err := container.ValidateName(req.InstanceID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stackID, stackActualName, err := h.getStackByName(stackName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
		return
	}

	var exists bool
	err = h.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM service_instances WHERE stack_id = $1 AND instance_id = $2 AND deleted_at IS NULL)`,
		stackID, req.InstanceID).Scan(&exists)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to check instance name: %v", err), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, fmt.Sprintf("instance %q already exists in stack %q", req.InstanceID, stackName), http.StatusConflict)
		return
	}

	// Validate combined container name length
	if err := container.ValidateContainerName(stackActualName, req.InstanceID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newContainerName := container.ContainerName(stackActualName, req.InstanceID)

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
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", oldInstanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to rename instance: %v", err), http.StatusInternalServerError)
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
			(SELECT COUNT(*) FROM instance_labels WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_domains WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_healthchecks WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_dependencies WHERE instance_id = $1) +
			(SELECT COUNT(*) FROM instance_config_files WHERE instance_id = $1)
	`, instance.ID).Scan(&instance.OverrideCount)
	if err != nil {
		instance.OverrideCount = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instance)
}

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
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get delete preview: %v", err), http.StatusInternalServerError)
		return
	}

	preview.OverrideCount = overrideCount

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
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
