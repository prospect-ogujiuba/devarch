package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/export"
	"github.com/priz/devarch-api/internal/lock"
)

type generateLockRequest struct {
	YmlContent []byte `json:"yml_content,omitempty"`
}

func (h *StackHandler) GenerateLock(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var req generateLockRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
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
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("export failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	generator := lock.NewGenerator(h.db, h.containerClient)
	lockFile, err := generator.Generate(stackName, ymlContent)
	if err != nil {
		http.Error(w, fmt.Sprintf("generate lock failed: %v", err), http.StatusInternalServerError)
		return
	}

	lockJSON, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal lock failed: %v", err), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%s-devarch.lock", stackName)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Write(lockJSON)
}

func (h *StackHandler) ValidateLock(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var lockFile lock.LockFile
	if err := json.NewDecoder(r.Body).Decode(&lockFile); err != nil {
		http.Error(w, fmt.Sprintf("invalid lock file: %v", err), http.StatusBadRequest)
		return
	}

	validator := lock.NewValidator(h.db, h.containerClient)
	result, err := validator.Validate(&lockFile, stackName)
	if err != nil {
		http.Error(w, fmt.Sprintf("validation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *StackHandler) RefreshLock(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	exporter := export.NewExporter(h.db)
	ymlContent, err := exporter.Export(stackName)
	if err != nil {
		if err.Error() == fmt.Sprintf("stack %q not found", stackName) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("export failed: %v", err), http.StatusInternalServerError)
		return
	}

	generator := lock.NewGenerator(h.db, h.containerClient)
	lockFile, err := generator.Generate(stackName, ymlContent)
	if err != nil {
		http.Error(w, fmt.Sprintf("generate lock failed: %v", err), http.StatusInternalServerError)
		return
	}

	lockJSON, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal lock failed: %v", err), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%s-devarch.lock", stackName)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Write(lockJSON)
}
