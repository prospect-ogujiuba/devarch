package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/lock"
	"github.com/priz/devarch-api/internal/orchestration"
)

type applyRequest struct {
	Token string         `json:"token"`
	Lock  *lock.LockFile `json:"lock,omitempty"`
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

	result, err := h.orchestrationService.ApplyPlan(r.Context(), stackName, req.Token, req.Lock)
	if err != nil {
		switch {
		case errors.Is(err, orchestration.ErrStackNotFound):
			respond.NotFound(w, r, "stack", stackName)
		case errors.Is(err, orchestration.ErrStackDisabled):
			respond.Conflict(w, r, "stack is disabled — enable it first")
		case errors.Is(err, orchestration.ErrLockConflict):
			respond.Conflict(w, r, "Stack is being applied by another session")
		case errors.Is(err, orchestration.ErrStalePlan):
			respond.Conflict(w, r, "Plan is stale — stack was modified since plan was generated. Regenerate plan.")
		default:
			respond.InternalError(w, r, err)
		}
		return
	}

	opts := []func(*respond.ActionResponse){
		respond.WithOutput(result.Output),
	}
	if len(result.LockWarnings) > 0 {
		opts = append(opts, respond.WithMetadata("lock_warnings", result.LockWarnings))
	}
	respond.Action(w, r, http.StatusOK, "applied", opts...)
}
