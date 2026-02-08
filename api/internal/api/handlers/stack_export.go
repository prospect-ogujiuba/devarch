package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/export"
)

func (h *StackHandler) ExportStack(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	exporter := export.NewExporter(h.db)
	yamlBytes, err := exporter.Export(stackName)
	if err != nil {
		if err.Error() == fmt.Sprintf("stack %q not found", stackName) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("export failed: %v", err), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%s-devarch.yml", stackName)
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Write(yamlBytes)
}
