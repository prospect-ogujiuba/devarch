package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/export"
	"github.com/priz/devarch-api/internal/lock"
)

type generateLockRequest struct {
	YmlContent []byte `json:"yml_content,omitempty"`
}

// GenerateLock godoc
// @Summary      Generate stack lock file
// @Tags         stacks
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        request body generateLockRequest false "Optional YAML content"
// @Success      200 {object} respond.SuccessEnvelope{data=object}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/lock [post]
// @Security     ApiKeyAuth
func (h *StackHandler) GenerateLock(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var req generateLockRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respond.BadRequest(w, r, fmt.Sprintf("invalid request body: %v", err))
			return
		}
	}

	var ymlContent []byte
	if len(req.YmlContent) > 0 {
		ymlContent = req.YmlContent
	} else {
		exporter := export.NewExporter(h.db)
		var err error
		ymlContent, err = exporter.Export(stackName)
		if err != nil {
			if err.Error() == fmt.Sprintf("stack %q not found", stackName) {
				respond.NotFound(w, r, "stack", stackName)
				return
			}
			respond.InternalError(w, r, err)
			return
		}
	}

	generator := lock.NewGenerator(h.db, h.containerClient)
	lockFile, err := generator.Generate(stackName, ymlContent)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	lockJSON, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	filename := fmt.Sprintf("%s-devarch.lock", stackName)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Write(lockJSON)
}

// ValidateLock godoc
// @Summary      Validate stack lock file
// @Tags         stacks
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        lock body lock.LockFile true "Lock file to validate"
// @Success      200 {object} respond.SuccessEnvelope{data=object}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/lock/validate [post]
// @Security     ApiKeyAuth
func (h *StackHandler) ValidateLock(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var lockFile lock.LockFile
	if err := json.NewDecoder(r.Body).Decode(&lockFile); err != nil {
		respond.BadRequest(w, r, fmt.Sprintf("invalid lock file: %v", err))
		return
	}

	validator := lock.NewValidator(h.db, h.containerClient)
	result, err := validator.Validate(&lockFile, stackName)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	respond.JSON(w, r, http.StatusOK, result)
}

// RefreshLock godoc
// @Summary      Refresh stack lock file
// @Description  Regenerate lock file from current stack state
// @Tags         stacks
// @Produce      json
// @Param        name path string true "Stack name"
// @Success      200 {object} respond.SuccessEnvelope{data=object}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/lock/refresh [post]
// @Security     ApiKeyAuth
func (h *StackHandler) RefreshLock(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	exporter := export.NewExporter(h.db)
	ymlContent, err := exporter.Export(stackName)
	if err != nil {
		if err.Error() == fmt.Sprintf("stack %q not found", stackName) {
			respond.NotFound(w, r, "stack", stackName)
			return
		}
		respond.InternalError(w, r, err)
		return
	}

	generator := lock.NewGenerator(h.db, h.containerClient)
	lockFile, err := generator.Generate(stackName, ymlContent)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	lockJSON, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	filename := fmt.Sprintf("%s-devarch.lock", stackName)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Write(lockJSON)
}
