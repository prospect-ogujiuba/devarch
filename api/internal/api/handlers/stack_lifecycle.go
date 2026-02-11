package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/compose"
)

// Stop godoc
// @Summary      Stop a stack
// @Tags         stacks
// @Produce      json
// @Param        name path string true "Stack name"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/stop [post]
// @Security     ApiKeyAuth
func (h *StackHandler) Stop(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	yamlBytes, err := h.stackCompose(name)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if err := h.containerClient.StopStack("devarch-"+name, yamlBytes); err != nil {
		respond.InternalError(w, r, err)
		return
	}

	respond.Action(w, r, http.StatusOK, "stopped")
}

// Start godoc
// @Summary      Start a stack
// @Description  Start all instances in a stack. Stack must be enabled.
// @Tags         stacks
// @Produce      json
// @Param        name path string true "Stack name"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/start [post]
// @Security     ApiKeyAuth
func (h *StackHandler) Start(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var enabled bool
	err := h.db.QueryRow(`
		SELECT enabled FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(&enabled)
	if err != nil {
		respond.NotFound(w, r, "stack", name)
		return
	}
	if !enabled {
		respond.Conflict(w, r, "stack is disabled — enable it first")
		return
	}

	var networkName *string
	h.db.QueryRow(`SELECT network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL`, name).Scan(&networkName)

	netName := fmt.Sprintf("devarch-%s-net", name)
	if networkName != nil && *networkName != "" {
		netName = *networkName
	}

	labels := map[string]string{
		"devarch.managed_by": "devarch",
		"devarch.stack":      name,
	}
	if err := h.containerClient.CreateNetwork(netName, labels); err != nil {
		respond.InternalError(w, r, err)
		return
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
		if err := gen.MaterializeStackConfigs(name, projectRoot); err != nil {
			respond.InternalError(w, r, err)
			return
		}
	}

	yamlBytes, _, err := gen.GenerateStack(name)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if err := h.containerClient.StartService("devarch-"+name, yamlBytes); err != nil {
		respond.InternalError(w, r, err)
		return
	}

	respond.Action(w, r, http.StatusOK, "started")
}

// Restart godoc
// @Summary      Restart a stack
// @Description  Restart all instances in a stack. Stack must be enabled.
// @Tags         stacks
// @Produce      json
// @Param        name path string true "Stack name"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/restart [post]
// @Security     ApiKeyAuth
func (h *StackHandler) Restart(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var enabled bool
	err := h.db.QueryRow(`
		SELECT enabled FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(&enabled)
	if err != nil {
		respond.NotFound(w, r, "stack", name)
		return
	}
	if !enabled {
		respond.Conflict(w, r, "stack is disabled — enable it first")
		return
	}

	stopYaml, err := h.stackCompose(name)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	h.containerClient.StopStack("devarch-"+name, stopYaml)

	var networkName *string
	h.db.QueryRow(`SELECT network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL`, name).Scan(&networkName)

	netName := fmt.Sprintf("devarch-%s-net", name)
	if networkName != nil && *networkName != "" {
		netName = *networkName
	}

	labels := map[string]string{
		"devarch.managed_by": "devarch",
		"devarch.stack":      name,
	}
	if err := h.containerClient.CreateNetwork(netName, labels); err != nil {
		respond.InternalError(w, r, err)
		return
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
		if err := gen.MaterializeStackConfigs(name, projectRoot); err != nil {
			respond.InternalError(w, r, err)
			return
		}
	}

	yamlBytes, _, err := gen.GenerateStack(name)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if err := h.containerClient.StartService("devarch-"+name, yamlBytes); err != nil {
		respond.InternalError(w, r, err)
		return
	}

	respond.Action(w, r, http.StatusOK, "restarted")
}
