package handlers

import (
	"github.com/priz/devarch-api/internal/api/respond"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/compose"
	"github.com/priz/devarch-api/internal/identity"
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

	netName := identity.NetworkName(stackName)
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

// Stop godoc
// @Summary      Stop instance
// @Tags         instances
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/stop [post]
// @Security     ApiKeyAuth
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

// Start godoc
// @Summary      Start instance
// @Tags         instances
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/start [post]
// @Security     ApiKeyAuth
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

// Restart godoc
// @Summary      Restart instance
// @Tags         instances
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/restart [post]
// @Security     ApiKeyAuth
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

// Logs godoc
// @Summary      Get instance logs
// @Tags         instances
// @Produce      plain
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Param        tail query int false "Number of log lines to return" default(100)
// @Success      200 {string} string
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/logs [get]
// @Security     ApiKeyAuth
func (h *InstanceHandler) Logs(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	tailStr := r.URL.Query().Get("tail")
	tail := 100
	if tailStr != "" {
		if parsed, err := strconv.Atoi(tailStr); err == nil && parsed > 0 {
			tail = parsed
		}
	}

	_, containerName, err := h.getInstanceRuntimeInfo(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to query instance: %w", err))
		return
	}

	logs, err := h.containerClient.GetLogs(containerName, strconv.Itoa(tail))
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(logs))
}

// Compose godoc
// @Summary      Get instance compose YAML
// @Tags         instances
// @Produce      plain
// @Param        name path string true "Stack name"
// @Param        instance path string true "Instance ID"
// @Success      200 {string} string
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/instances/{instance}/compose [get]
// @Security     ApiKeyAuth
func (h *InstanceHandler) Compose(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	_, _, err := h.getInstanceRuntimeInfo(stackName, instanceName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "instance", instanceName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to query instance: %w", err))
		return
	}

	_, yamlBytes, err := h.instanceCompose(stackName)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("failed to generate compose: %w", err))
		return
	}

	w.Header().Set("Content-Type", "text/yaml")
	w.Write(yamlBytes)
}
