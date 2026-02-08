package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/export"
	"gopkg.in/yaml.v3"
)

func (h *StackHandler) ImportStack(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse multipart form: %v", err), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get file from form: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	var devarchFile export.DevArchFile
	if err := yaml.NewDecoder(file).Decode(&devarchFile); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode YAML: %v", err), http.StatusBadRequest)
		return
	}

	if devarchFile.Version != 1 {
		http.Error(w, fmt.Sprintf("unsupported version %d — only version 1 is supported", devarchFile.Version), http.StatusBadRequest)
		return
	}

	if devarchFile.Stack.Name == "" {
		http.Error(w, "stack name is required", http.StatusBadRequest)
		return
	}

	if err := container.ValidateName(devarchFile.Stack.Name); err != nil {
		http.Error(w, fmt.Sprintf("invalid stack name: %v", err), http.StatusBadRequest)
		return
	}

	importer := export.NewImporter(h.db)
	result, err := importer.Import(&devarchFile)
	if err != nil {
		if strings.Contains(err.Error(), "not found in catalog") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":  err.Error(),
				"result": result,
			})
			return
		}
		if strings.Contains(err.Error(), "locked") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("import failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
