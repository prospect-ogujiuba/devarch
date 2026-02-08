package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/plan"
)

func (h *StackHandler) Plan(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var stackID int
	var networkName sql.NullString
	var stackUpdatedAt time.Time
	var enabled bool
	err := h.db.QueryRow(`
		SELECT id, network_name, updated_at, enabled
		FROM stacks
		WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID, &networkName, &stackUpdatedAt, &enabled)

	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
		return
	}

	if !enabled {
		http.Error(w, "stack is disabled — enable it first", http.StatusConflict)
		return
	}

	rows, err := h.db.Query(`
		SELECT si.instance_id, s.name as template_name, si.container_name, si.enabled, si.updated_at
		FROM service_instances si
		JOIN services s ON s.id = si.template_service_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
		ORDER BY si.instance_id
	`, stackID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query instances: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var desired []plan.DesiredInstance
	var timestamps []plan.InstanceTimestamp
	for rows.Next() {
		var instanceID, templateName string
		var containerName sql.NullString
		var enabled bool
		var updatedAt time.Time

		if err := rows.Scan(&instanceID, &templateName, &containerName, &enabled, &updatedAt); err != nil {
			http.Error(w, fmt.Sprintf("failed to scan instance: %v", err), http.StatusInternalServerError)
			return
		}

		name := ""
		if containerName.Valid && containerName.String != "" {
			name = containerName.String
		} else {
			name = container.ContainerName(stackName, instanceID)
		}

		desired = append(desired, plan.DesiredInstance{
			InstanceID:    instanceID,
			TemplateName:  templateName,
			ContainerName: name,
			Enabled:       enabled,
		})

		timestamps = append(timestamps, plan.InstanceTimestamp{
			InstanceID: instanceID,
			UpdatedAt:  updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("failed to iterate instances: %v", err), http.StatusInternalServerError)
		return
	}

	running, err := h.containerClient.ListContainersWithLabels(map[string]string{
		"devarch.stack_id": stackName,
	})
	if err != nil {
		// Log error but continue with empty slice - runtime may be down, plan should still work
		running = []string{}
	}

	changes := plan.ComputeDiff(desired, running)
	if changes == nil {
		changes = []plan.Change{}
	}
	token := plan.GenerateToken(stackUpdatedAt, timestamps)

	planResp := plan.Plan{
		StackName:   stackName,
		StackID:     stackID,
		Changes:     changes,
		Token:       token,
		GeneratedAt: time.Now(),
		Warnings:    []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(planResp)
}
