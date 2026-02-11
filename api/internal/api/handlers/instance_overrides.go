package handlers

import (
	"github.com/priz/devarch-api/internal/api/respond"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/pkg/models"
)

func (h *InstanceHandler) getInstanceByName(stackName, instanceName string) (int, int, error) {
	var instanceID, stackID int
	err := h.db.QueryRow(`
		SELECT si.id, si.stack_id
		FROM service_instances si
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.instance_id = $2 AND si.deleted_at IS NULL AND st.deleted_at IS NULL
	`, stackName, instanceName).Scan(&instanceID, &stackID)
	return instanceID, stackID, err
}

// UpdatePorts godoc
// @Summary      Update instance port overrides
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body object true "Ports array"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/ports [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdatePorts(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		Ports []models.ServicePort `json:"ports"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to begin transaction: %w", err))
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_ports WHERE instance_id = $1", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete existing ports: %w", err))
		return
	}

	for _, p := range req.Ports {
		_, err = tx.Exec(
			`INSERT INTO instance_ports (instance_id, host_ip, host_port, container_port, protocol) VALUES ($1, $2, $3, $4, $5)`,
			instanceID, p.HostIP, p.HostPort, p.ContainerPort, p.Protocol,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to insert port: %w", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// UpdateVolumes godoc
// @Summary      Update instance volume overrides
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body object true "Volumes array"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/volumes [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdateVolumes(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		Volumes []models.ServiceVolume `json:"volumes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to begin transaction: %w", err))
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_volumes WHERE instance_id = $1", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete existing volumes: %w", err))
		return
	}

	for _, v := range req.Volumes {
		_, err = tx.Exec(
			`INSERT INTO instance_volumes (instance_id, volume_type, source, target, read_only, is_external) VALUES ($1, $2, $3, $4, $5, $6)`,
			instanceID, v.VolumeType, v.Source, v.Target, v.ReadOnly, v.IsExternal,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to insert volume: %w", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// UpdateEnvVars godoc
// @Summary      Update instance environment variable overrides
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body object true "Environment variables array"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/env-vars [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdateEnvVars(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		EnvVars []models.ServiceEnvVar `json:"env_vars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to begin transaction: %w", err))
		return
	}
	defer tx.Rollback()

	type existingSecret struct {
		encryptedValue    sql.NullString
		encryptionVersion int
	}
	existingSecrets := map[string]existingSecret{}
	rows, err := tx.Query("SELECT key, encrypted_value, encryption_version FROM instance_env_vars WHERE instance_id = $1 AND is_secret = true", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to query existing secrets: %w", err))
		return
	}
	for rows.Next() {
		var key string
		var secret existingSecret
		if err := rows.Scan(&key, &secret.encryptedValue, &secret.encryptionVersion); err != nil {
			rows.Close()
			respond.InternalError(w, r, fmt.Errorf("failed to scan existing secret: %w", err))
			return
		}
		existingSecrets[key] = secret
	}
	rows.Close()

	_, err = tx.Exec("DELETE FROM instance_env_vars WHERE instance_id = $1", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete existing env vars: %w", err))
		return
	}

	for _, e := range req.EnvVars {
		value := e.Value
		var encryptedValue sql.NullString
		encryptionVersion := 0

		if e.IsSecret {
			if e.Value == "***" || e.Value == "" {
				if current, ok := existingSecrets[e.Key]; ok {
					encryptedValue = current.encryptedValue
					encryptionVersion = current.encryptionVersion
					value = ""
				}
			} else {
				encrypted, err := h.cipher.Encrypt(e.Value)
				if err != nil {
					respond.InternalError(w, r, fmt.Errorf("failed to encrypt secret: %w", err))
					return
				}
				encryptedValue = sql.NullString{String: encrypted, Valid: true}
				encryptionVersion = 1
				value = ""
			}
		}

		_, err = tx.Exec(
			`INSERT INTO instance_env_vars (instance_id, key, value, is_secret, encrypted_value, encryption_version) VALUES ($1, $2, $3, $4, $5, $6)`,
			instanceID, e.Key, value, e.IsSecret, encryptedValue, encryptionVersion,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to insert env var: %w", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// UpdateLabels godoc
// @Summary      Update instance label overrides
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body object true "Labels array"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/labels [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdateLabels(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		Labels []models.ServiceLabel `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	for _, l := range req.Labels {
		if strings.HasPrefix(l.Key, "devarch.") {
			respond.BadRequest(w, r, fmt.Sprintf("label key %q cannot start with 'devarch.' - this prefix is reserved for system labels", l.Key))
			return
		}
	}

	tx, err := h.db.Begin()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to begin transaction: %w", err))
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_labels WHERE instance_id = $1", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete existing labels: %w", err))
		return
	}

	for _, l := range req.Labels {
		_, err = tx.Exec(
			`INSERT INTO instance_labels (instance_id, key, value) VALUES ($1, $2, $3)`,
			instanceID, l.Key, l.Value,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to insert label: %w", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// UpdateDomains godoc
// @Summary      Update instance domain overrides
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body object true "Domains array"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/domains [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdateDomains(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		Domains []models.ServiceDomain `json:"domains"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to begin transaction: %w", err))
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_domains WHERE instance_id = $1", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete existing domains: %w", err))
		return
	}

	for _, d := range req.Domains {
		_, err = tx.Exec(
			`INSERT INTO instance_domains (instance_id, domain, proxy_port) VALUES ($1, $2, $3)`,
			instanceID, d.Domain, d.ProxyPort,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to insert domain: %w", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// UpdateHealthcheck godoc
// @Summary      Update instance healthcheck override
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body models.ServiceHealthcheck true "Healthcheck configuration"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/healthcheck [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdateHealthcheck(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req *models.ServiceHealthcheck
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	_, err = h.db.Exec("DELETE FROM instance_healthchecks WHERE instance_id = $1", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete existing healthcheck: %w", err))
		return
	}

	if req != nil && req.Test != "" {
		_, err := h.db.Exec(
			`INSERT INTO instance_healthchecks (instance_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds) VALUES ($1, $2, $3, $4, $5, $6)`,
			instanceID, req.Test, req.IntervalSeconds, req.TimeoutSeconds, req.Retries, req.StartPeriodSeconds,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to insert healthcheck: %w", err))
			return
		}
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// UpdateDependencies godoc
// @Summary      Update instance dependency overrides
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body object true "Dependencies array"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/dependencies [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdateDependencies(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		Dependencies []struct {
			DependsOn string `json:"depends_on"`
			Condition string `json:"condition"`
		} `json:"dependencies"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to begin transaction: %w", err))
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_dependencies WHERE instance_id = $1", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete existing dependencies: %w", err))
		return
	}

	for _, d := range req.Dependencies {
		condition := d.Condition
		if condition == "" {
			condition = "service_started"
		}
		_, err = tx.Exec(
			`INSERT INTO instance_dependencies (instance_id, depends_on, condition) VALUES ($1, $2, $3)`,
			instanceID, d.DependsOn, condition,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to insert dependency: %w", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// ListConfigFiles godoc
// @Summary      List instance config files
// @Tags         instances
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/files [get]
// @Security     ApiKeyAuth
func (h *InstanceHandler) ListConfigFiles(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	rows, err := h.db.Query(`
		SELECT id, instance_id, file_path, file_mode, is_template, created_at, updated_at
		FROM instance_config_files WHERE instance_id = $1 ORDER BY file_path
	`, instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to query config files: %w", err))
		return
	}
	defer rows.Close()

	type configFileResponse struct {
		ID         int    `json:"id"`
		InstanceID int    `json:"instance_id"`
		FilePath   string `json:"file_path"`
		FileMode   string `json:"file_mode"`
		IsTemplate bool   `json:"is_template"`
		CreatedAt  string `json:"created_at"`
		UpdatedAt  string `json:"updated_at"`
	}

	files := []configFileResponse{}
	for rows.Next() {
		var f configFileResponse
		if err := rows.Scan(&f.ID, &f.InstanceID, &f.FilePath, &f.FileMode, &f.IsTemplate, &f.CreatedAt, &f.UpdatedAt); err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to scan config file: %w", err))
			return
		}
		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("error iterating config files: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, files)
}

// GetConfigFile godoc
// @Summary      Get instance config file
// @Tags         instances
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        filepath path string true "File path"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/files/{filepath} [get]
// @Security     ApiKeyAuth
func (h *InstanceHandler) GetConfigFile(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")
	filePath := chi.URLParam(r, "*")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	type configFileResponse struct {
		ID         int    `json:"id"`
		InstanceID int    `json:"instance_id"`
		FilePath   string `json:"file_path"`
		Content    string `json:"content"`
		FileMode   string `json:"file_mode"`
		IsTemplate bool   `json:"is_template"`
		CreatedAt  string `json:"created_at"`
		UpdatedAt  string `json:"updated_at"`
	}

	var f configFileResponse
	err = h.db.QueryRow(`
		SELECT id, instance_id, file_path, content, file_mode, is_template, created_at, updated_at
		FROM instance_config_files WHERE instance_id = $1 AND file_path = $2
	`, instanceID, filePath).Scan(&f.ID, &f.InstanceID, &f.FilePath, &f.Content, &f.FileMode, &f.IsTemplate, &f.CreatedAt, &f.UpdatedAt)

	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "config file", filePath)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get config file: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, f)
}

// PutConfigFile godoc
// @Summary      Create or update instance config file
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        filepath path string true "File path"
// @Param        body body object true "Config file content"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/files/{filepath} [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) PutConfigFile(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")
	filePath := chi.URLParam(r, "*")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		Content    string `json:"content"`
		FileMode   string `json:"file_mode"`
		IsTemplate bool   `json:"is_template"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	type configFileResponse struct {
		ID         int    `json:"id"`
		InstanceID int    `json:"instance_id"`
		FilePath   string `json:"file_path"`
		Content    string `json:"content"`
		FileMode   string `json:"file_mode"`
		IsTemplate bool   `json:"is_template"`
		CreatedAt  string `json:"created_at"`
		UpdatedAt  string `json:"updated_at"`
	}

	var f configFileResponse
	err = h.db.QueryRow(`
		INSERT INTO instance_config_files (instance_id, file_path, content, file_mode, is_template)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (instance_id, file_path) DO UPDATE
		SET content = EXCLUDED.content, file_mode = EXCLUDED.file_mode, is_template = EXCLUDED.is_template, updated_at = NOW()
		RETURNING id, instance_id, file_path, content, file_mode, is_template, created_at, updated_at
	`, instanceID, filePath, req.Content, req.FileMode, req.IsTemplate).Scan(
		&f.ID, &f.InstanceID, &f.FilePath, &f.Content, &f.FileMode, &f.IsTemplate, &f.CreatedAt, &f.UpdatedAt,
	)

	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to upsert config file: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, f)
}

// DeleteConfigFile godoc
// @Summary      Delete instance config file
// @Tags         instances
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        filepath path string true "File path"
// @Success      204
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/files/{filepath} [delete]
// @Security     ApiKeyAuth
func (h *InstanceHandler) DeleteConfigFile(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")
	filePath := chi.URLParam(r, "*")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	result, err := h.db.Exec("DELETE FROM instance_config_files WHERE instance_id = $1 AND file_path = $2", instanceID, filePath)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete config file: %w", err))
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get rows affected: %w", err))
		return
	}

	if rowsAffected == 0 {
		respond.NotFound(w, r, "config file", filePath)
		return
	}

	respond.NoContent(w, r)
}

// GetResourceLimits godoc
// @Summary      Get instance resource limits
// @Tags         instances
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/resources [get]
// @Security     ApiKeyAuth
func (h *InstanceHandler) GetResourceLimits(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	type resourceLimitsResponse struct {
		CPULimit          string `json:"cpu_limit,omitempty"`
		CPUReservation    string `json:"cpu_reservation,omitempty"`
		MemoryLimit       string `json:"memory_limit,omitempty"`
		MemoryReservation string `json:"memory_reservation,omitempty"`
	}

	var resp resourceLimitsResponse
	err = h.db.QueryRow(`
		SELECT cpu_limit, cpu_reservation, memory_limit, memory_reservation
		FROM instance_resource_limits WHERE instance_id = $1
	`, instanceID).Scan(
		&resp.CPULimit,
		&resp.CPUReservation,
		&resp.MemoryLimit,
		&resp.MemoryReservation,
	)

	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "resource limits", "")
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get resource limits: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, resp)
}

// UpdateResourceLimits godoc
// @Summary      Update instance resource limits
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body object true "Resource limits"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/resources [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdateResourceLimits(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		CPULimit          string `json:"cpu_limit"`
		CPUReservation    string `json:"cpu_reservation"`
		MemoryLimit       string `json:"memory_limit"`
		MemoryReservation string `json:"memory_reservation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	warnings := []string{}

	if req.MemoryLimit != "" {
		bytes, err := parseMemoryString(req.MemoryLimit)
		if err != nil {
			respond.BadRequest(w, r, fmt.Sprintf("invalid memory_limit: %v", err))
			return
		}
		if bytes < 4*1024*1024 {
			warnings = append(warnings, "memory limit very low (< 4MB), container may fail to start")
		}
	}

	if req.CPULimit != "" {
		cpuLimit, err := strconv.ParseFloat(req.CPULimit, 64)
		if err != nil {
			respond.BadRequest(w, r, fmt.Sprintf("invalid cpu_limit: must be numeric, got %q", req.CPULimit))
			return
		}
		if cpuLimit < 0.01 {
			warnings = append(warnings, "CPU limit extremely low (< 0.01)")
		}
	}

	allEmpty := req.CPULimit == "" && req.CPUReservation == "" && req.MemoryLimit == "" && req.MemoryReservation == ""
	if allEmpty {
		_, err := h.db.Exec("DELETE FROM instance_resource_limits WHERE instance_id = $1", instanceID)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to delete resource limits: %w", err))
			return
		}
		respond.JSON(w, r, http.StatusOK, map[string]interface{}{
			"status": "deleted",
		})
		return
	}

	_, err = h.db.Exec(`
		INSERT INTO instance_resource_limits (instance_id, cpu_limit, cpu_reservation, memory_limit, memory_reservation)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''))
		ON CONFLICT (instance_id) DO UPDATE SET
			cpu_limit = NULLIF(EXCLUDED.cpu_limit, ''),
			cpu_reservation = NULLIF(EXCLUDED.cpu_reservation, ''),
			memory_limit = NULLIF(EXCLUDED.memory_limit, ''),
			memory_reservation = NULLIF(EXCLUDED.memory_reservation, '')
	`, instanceID, req.CPULimit, req.CPUReservation, req.MemoryLimit, req.MemoryReservation)

	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to save resource limits: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]interface{}{
		"status":             "updated",
		"cpu_limit":          req.CPULimit,
		"cpu_reservation":    req.CPUReservation,
		"memory_limit":       req.MemoryLimit,
		"memory_reservation": req.MemoryReservation,
		"warnings":           warnings,
	})
}

// UpdateEnvFiles godoc
// @Summary      Update instance env_files overrides
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body object true "Env files array"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/env-files [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdateEnvFiles(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		EnvFiles []string `json:"env_files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to begin transaction: %w", err))
		return
	}
	defer tx.Rollback()

	for _, path := range req.EnvFiles {
		if strings.TrimSpace(path) == "" {
			respond.BadRequest(w, r, "env_file path cannot be empty")
			return
		}
		if strings.Contains(path, "..") || strings.HasPrefix(path, "/") {
			respond.BadRequest(w, r, "env_file path must be relative and cannot contain '..'")
			return
		}
	}

	_, err = tx.Exec("DELETE FROM instance_env_files WHERE instance_id = $1", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete existing env_files: %w", err))
		return
	}

	for idx, path := range req.EnvFiles {
		_, err = tx.Exec(
			`INSERT INTO instance_env_files (instance_id, path, sort_order) VALUES ($1, $2, $3)`,
			instanceID, path, idx,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to insert env_file: %w", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// UpdateNetworks godoc
// @Summary      Update instance network overrides
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body object true "Networks array"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/networks [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdateNetworks(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		Networks []string `json:"networks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	for _, name := range req.Networks {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			respond.BadRequest(w, r, "network name cannot be empty")
			return
		}
		if strings.ContainsAny(trimmed, " \t/\\:") {
			respond.BadRequest(w, r, fmt.Sprintf("invalid network name %q: contains disallowed characters", trimmed))
			return
		}
	}

	tx, err := h.db.Begin()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to begin transaction: %w", err))
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_networks WHERE instance_id = $1", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete existing networks: %w", err))
		return
	}

	for _, name := range req.Networks {
		_, err = tx.Exec(
			`INSERT INTO instance_networks (instance_id, network_name) VALUES ($1, $2)`,
			instanceID, name,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to insert network: %w", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

// UpdateConfigMounts godoc
// @Summary      Update instance config mount overrides
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        body body object true "Config mounts array"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/config-mounts [put]
// @Security     ApiKeyAuth
func (h *InstanceHandler) UpdateConfigMounts(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to get instance: %w", err))
		return
	}

	var req struct {
		ConfigMounts []models.ServiceConfigMount `json:"config_mounts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	for _, m := range req.ConfigMounts {
		if strings.TrimSpace(m.TargetPath) == "" {
			respond.BadRequest(w, r, "config mount target_path cannot be empty")
			return
		}
		if strings.Contains(m.SourcePath, "..") {
			respond.BadRequest(w, r, "config mount source_path cannot contain '..'")
			return
		}
	}

	tx, err := h.db.Begin()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to begin transaction: %w", err))
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_config_mounts WHERE instance_id = $1", instanceID)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to delete existing config_mounts: %w", err))
		return
	}

	for _, m := range req.ConfigMounts {
		var cfgFileID sql.NullInt32
		if m.ConfigFileID != nil {
			cfgFileID = sql.NullInt32{Int32: int32(*m.ConfigFileID), Valid: true}
		}
		_, err = tx.Exec(
			`INSERT INTO instance_config_mounts (instance_id, config_file_id, source_path, target_path, readonly) VALUES ($1, $2, $3, $4, $5)`,
			instanceID, cfgFileID, m.SourcePath, m.TargetPath, m.ReadOnly,
		)
		if err != nil {
			respond.InternalError(w, r, fmt.Errorf("failed to insert config_mount: %w", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

func parseMemoryString(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	multiplier := int64(1)
	lastChar := strings.ToLower(string(s[len(s)-1]))

	if lastChar == "k" {
		multiplier = 1024
		s = s[:len(s)-1]
	} else if lastChar == "m" {
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	} else if lastChar == "g" {
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}

	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %w", err)
	}

	return num * multiplier, nil
}
