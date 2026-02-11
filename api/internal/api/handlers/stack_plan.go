package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/orchestration"
)

// Plan godoc
// @Summary      Generate stack deployment plan
// @Tags         stacks
// @Produce      json
// @Param        name path string true "Stack name"
// @Success      200 {object} respond.SuccessEnvelope{data=object}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/plan [get]
// @Security     ApiKeyAuth
func (h *StackHandler) Plan(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	planResp, err := h.orchestrationService.GeneratePlan(stackName)
	if err != nil {
		switch {
		case errors.Is(err, orchestration.ErrStackNotFound):
			respond.NotFound(w, r, "stack", stackName)
		case errors.Is(err, orchestration.ErrStackDisabled):
			respond.Conflict(w, r, "stack is disabled — enable it first")
		default:
			respond.InternalError(w, r, err)
		}
		return
	}

	respond.JSON(w, r, http.StatusOK, planResp)
}
