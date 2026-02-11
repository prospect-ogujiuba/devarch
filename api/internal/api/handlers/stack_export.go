package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/export"
)

// ExportStack godoc
// @Summary      Export stack as compose YAML
// @Tags         stacks
// @Produce      application/x-yaml
// @Param        name path string true "Stack name"
// @Success      200 {string} string "Compose YAML content"
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/export [get]
// @Security     ApiKeyAuth
func (h *StackHandler) ExportStack(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	exporter := export.NewExporter(h.db)
	yamlBytes, err := exporter.Export(stackName)
	if err != nil {
		if err.Error() == fmt.Sprintf("stack %q not found", stackName) {
			respond.NotFound(w, r, "stack", stackName)
			return
		}
		respond.InternalError(w, r, err)
		return
	}

	filename := fmt.Sprintf("%s-devarch.yml", stackName)
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Write(yamlBytes)
}
