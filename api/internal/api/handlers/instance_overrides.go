package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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

func (h *InstanceHandler) UpdatePorts(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
		return
	}

	var req struct {
		Ports []models.ServicePort `json:"ports"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_ports WHERE instance_id = $1", instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete existing ports: %v", err), http.StatusInternalServerError)
		return
	}

	for _, p := range req.Ports {
		_, err = tx.Exec(
			`INSERT INTO instance_ports (instance_id, host_ip, host_port, container_port, protocol) VALUES ($1, $2, $3, $4, $5)`,
			instanceID, p.HostIP, p.HostPort, p.ContainerPort, p.Protocol,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to insert port: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("failed to commit transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *InstanceHandler) UpdateVolumes(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
		return
	}

	var req struct {
		Volumes []models.ServiceVolume `json:"volumes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_volumes WHERE instance_id = $1", instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete existing volumes: %v", err), http.StatusInternalServerError)
		return
	}

	for _, v := range req.Volumes {
		_, err = tx.Exec(
			`INSERT INTO instance_volumes (instance_id, volume_type, source, target, read_only, is_external) VALUES ($1, $2, $3, $4, $5, $6)`,
			instanceID, v.VolumeType, v.Source, v.Target, v.ReadOnly, v.IsExternal,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to insert volume: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("failed to commit transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *InstanceHandler) UpdateEnvVars(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
		return
	}

	var req struct {
		EnvVars []models.ServiceEnvVar `json:"env_vars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_env_vars WHERE instance_id = $1", instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete existing env vars: %v", err), http.StatusInternalServerError)
		return
	}

	for _, e := range req.EnvVars {
		_, err = tx.Exec(
			`INSERT INTO instance_env_vars (instance_id, key, value, is_secret) VALUES ($1, $2, $3, $4)`,
			instanceID, e.Key, e.Value, e.IsSecret,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to insert env var: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("failed to commit transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *InstanceHandler) UpdateLabels(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
		return
	}

	var req struct {
		Labels []models.ServiceLabel `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	for _, l := range req.Labels {
		if strings.HasPrefix(l.Key, "devarch.") {
			http.Error(w, fmt.Sprintf("label key %q cannot start with 'devarch.' - this prefix is reserved for system labels", l.Key), http.StatusBadRequest)
			return
		}
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_labels WHERE instance_id = $1", instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete existing labels: %v", err), http.StatusInternalServerError)
		return
	}

	for _, l := range req.Labels {
		_, err = tx.Exec(
			`INSERT INTO instance_labels (instance_id, key, value) VALUES ($1, $2, $3)`,
			instanceID, l.Key, l.Value,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to insert label: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("failed to commit transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *InstanceHandler) UpdateDomains(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
		return
	}

	var req struct {
		Domains []models.ServiceDomain `json:"domains"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM instance_domains WHERE instance_id = $1", instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete existing domains: %v", err), http.StatusInternalServerError)
		return
	}

	for _, d := range req.Domains {
		_, err = tx.Exec(
			`INSERT INTO instance_domains (instance_id, domain, proxy_port) VALUES ($1, $2, $3)`,
			instanceID, d.Domain, d.ProxyPort,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to insert domain: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("failed to commit transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *InstanceHandler) UpdateHealthcheck(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
		return
	}

	var req *models.ServiceHealthcheck
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec("DELETE FROM instance_healthchecks WHERE instance_id = $1", instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete existing healthcheck: %v", err), http.StatusInternalServerError)
		return
	}

	if req != nil && req.Test != "" {
		_, err := h.db.Exec(
			`INSERT INTO instance_healthchecks (instance_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds) VALUES ($1, $2, $3, $4, $5, $6)`,
			instanceID, req.Test, req.IntervalSeconds, req.TimeoutSeconds, req.Retries, req.StartPeriodSeconds,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to insert healthcheck: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *InstanceHandler) ListConfigFiles(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, instance_id, file_path, file_mode, is_template, created_at, updated_at
		FROM instance_config_files WHERE instance_id = $1 ORDER BY file_path
	`, instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query config files: %v", err), http.StatusInternalServerError)
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
			http.Error(w, fmt.Sprintf("failed to scan config file: %v", err), http.StatusInternalServerError)
			return
		}
		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("error iterating config files: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (h *InstanceHandler) GetConfigFile(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")
	filePath := chi.URLParam(r, "*")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("config file %q not found", filePath), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get config file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

func (h *InstanceHandler) PutConfigFile(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")
	filePath := chi.URLParam(r, "*")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
		return
	}

	var req struct {
		Content    string `json:"content"`
		FileMode   string `json:"file_mode"`
		IsTemplate bool   `json:"is_template"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("failed to upsert config file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

func (h *InstanceHandler) DeleteConfigFile(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")
	filePath := chi.URLParam(r, "*")

	instanceID, _, err := h.getInstanceByName(stackName, instanceName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("instance %q not found in stack %q", instanceName, stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get instance: %v", err), http.StatusInternalServerError)
		return
	}

	result, err := h.db.Exec("DELETE FROM instance_config_files WHERE instance_id = $1 AND file_path = $2", instanceID, filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete config file: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get rows affected: %v", err), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, fmt.Sprintf("config file %q not found", filePath), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
