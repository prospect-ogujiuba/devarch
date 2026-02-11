package handlers

import (
	"github.com/priz/devarch-api/internal/api/respond"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/compose"
)

func (h *InstanceHandler) getInstanceRuntimeInfo(stackName string, instanceName string) (enabled bool, containerName string, err error) {
	err = h.db.QueryRow(`
		SELECT si.enabled, COALESCE(si.container_name, '')
		FROM service_instances si
		JOIN stacks st ON st.id = si.stack_id
		WHERE st.name = $1 AND si.instance_id = $2 AND si.deleted_at IS NULL AND st.deleted_at IS NULL
	`, stackName, instanceName).Scan(&enabled, &containerName)
	return
}

func (h *InstanceHandler) instanceCompose(stackName string) (projectName string, yamlBytes []byte, err error) {
	var networkName sql.NullString
	err = h.db.QueryRow(`
		SELECT network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&networkName)
	if err != nil {
		return "", nil, fmt.Errorf("lookup stack: %w", err)
	}

	netName := fmt.Sprintf("devarch-%s-net", stackName)
	if networkName.Valid && networkName.String != "" {
		netName = networkName.String
	}

	gen := compose.NewGenerator(h.db, netName)
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		gen.SetProjectRoot(root)
	}
	if hostRoot := os.Getenv("HOST_PROJECT_ROOT"); hostRoot != "" {
		gen.SetHostProjectRoot(hostRoot)
	}
	if ws := os.Getenv("WORKSPACE_ROOT"); ws != "" {
		gen.SetWorkspaceRoot(ws)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot != "" {
		if err := gen.MaterializeStackConfigs(stackName, projectRoot); err != nil {
			return "", nil, fmt.Errorf("materialize configs: %w", err)
		}
	}

	yaml, _, err := gen.GenerateStack(stackName)
	if err != nil {
		return "", nil, fmt.Errorf("generate compose: %w", err)
	}

	return "devarch-" + stackName, yaml, nil
}

func (h *InstanceHandler) Stop(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	_, containerName, err := h.getInstanceRuntimeInfo(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to query instance: %w", err))
		return
	}

	if err := h.containerClient.StopContainer(containerName); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to stop instance: %w", err))
		return
	}

	respond.Action(w, r, http.StatusOK, "stopped")
}

func (h *InstanceHandler) Start(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	enabled, _, err := h.getInstanceRuntimeInfo(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to query instance: %w", err))
		return
	}
	if !enabled {
		respond.Conflict(w, r, "instance is disabled — enable it first")
		return
	}

	projectName, yamlBytes, err := h.instanceCompose(stackName)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to generate compose: %w", err))
		return
	}

	if err := h.containerClient.StartComposeService(projectName, yamlBytes, instanceName); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to start instance: %w", err))
		return
	}

	respond.Action(w, r, http.StatusOK, "started")
}

func (h *InstanceHandler) Restart(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	enabled, _, err := h.getInstanceRuntimeInfo(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to query instance: %w", err))
		return
	}
	if !enabled {
		respond.Conflict(w, r, "instance is disabled — enable it first")
		return
	}

	projectName, yamlBytes, err := h.instanceCompose(stackName)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to generate compose: %w", err))
		return
	}

	if err := h.containerClient.RestartComposeService(projectName, yamlBytes, instanceName); err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to restart instance: %w", err))
		return
	}

	respond.Action(w, r, http.StatusOK, "restarted")
}
