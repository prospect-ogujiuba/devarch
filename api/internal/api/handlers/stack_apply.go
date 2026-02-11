package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/compose"
	"github.com/priz/devarch-api/internal/lock"
	"github.com/priz/devarch-api/internal/plan"
)

type applyRequest struct {
	Token string          `json:"token"`
	Lock  *lock.LockFile  `json:"lock,omitempty"`
}

// Apply godoc
// @Summary      Apply stack deployment plan
// @Description  Apply a generated plan to deploy stack changes
// @Tags         stacks
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        apply body applyRequest true "Apply request with plan token"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/apply [post]
// @Security     ApiKeyAuth
func (h *StackHandler) Apply(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var req applyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.Token == "" {
		respond.BadRequest(w, r, "token is required")
		return
	}

	var stackID int
	var networkName sql.NullString
	var enabled bool
	err := h.db.QueryRow(`
		SELECT id, network_name, enabled
		FROM stacks
		WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID, &networkName, &enabled)

	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "stack", stackName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if !enabled {
		respond.Conflict(w, r, "stack is disabled — enable it first")
		return
	}

	var acquired bool
	err = h.db.QueryRowContext(r.Context(), "SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquired)
	if err != nil || !acquired {
		respond.Conflict(w, r, "Stack is being applied by another session")
		return
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.db.ExecContext(unlockCtx, "SELECT pg_advisory_unlock($1)", stackID)
	}()

	if err := plan.ValidateToken(h.db, stackID, req.Token); err != nil {
		if errors.Is(err, plan.ErrStalePlan) {
			respond.Conflict(w, r, "Plan is stale — stack was modified since plan was generated. Regenerate plan.")
			return
		}
		respond.InternalError(w, r, err)
		return
	}

	netName := ""
	if networkName.Valid && networkName.String != "" {
		netName = networkName.String
	} else {
		netName = fmt.Sprintf("devarch-%s-net", stackName)
	}

	labels := map[string]string{
		"devarch.managed_by": "devarch",
		"devarch.stack":      stackName,
	}
	if err := h.containerClient.CreateNetwork(netName, labels); err != nil {
		respond.InternalError(w, r, err)
		return
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		respond.InternalError(w, r, fmt.Errorf("PROJECT_ROOT environment variable not set"))
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

	if err := gen.MaterializeStackConfigs(stackName, projectRoot); err != nil {
		configDir := filepath.Join(projectRoot, ".runtime", "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		respond.InternalError(w, r, err)
		return
	}

	yamlBytes, _, err := gen.GenerateStack(stackName)
	if err != nil {
		configDir := filepath.Join(projectRoot, ".runtime", "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		respond.InternalError(w, r, err)
		return
	}

	tmpFile, err := os.CreateTemp("", "devarch-apply-*.yml")
	if err != nil {
		configDir := filepath.Join(projectRoot, ".runtime", "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		respond.InternalError(w, r, err)
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(yamlBytes); err != nil {
		tmpFile.Close()
		configDir := filepath.Join(projectRoot, ".runtime", "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		respond.InternalError(w, r, err)
		return
	}
	tmpFile.Close()

	output, err := h.containerClient.RunCompose(tmpFile.Name(), "--project-name", "devarch-"+stackName, "up", "-d")
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("compose up failed: %v\n%s", err, output))
		return
	}

	opts := []func(*respond.ActionResponse){
		respond.WithOutput(output),
	}
	if req.Lock != nil {
		validator := lock.NewValidator(h.db, h.containerClient)
		result, err := validator.Validate(req.Lock, stackName)
		if err == nil && len(result.Warnings) > 0 {
			opts = append(opts, respond.WithMetadata("lock_warnings", result.Warnings))
		}
	}
	respond.Action(w, r, http.StatusOK, "applied", opts...)
}
