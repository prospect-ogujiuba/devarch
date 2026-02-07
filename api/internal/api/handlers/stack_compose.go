package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/compose"
)

func (h *StackHandler) Compose(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var stackID int
	var networkName sql.NullString
	err := h.db.QueryRow(`
		SELECT id, network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID, &networkName)

	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
		return
	}

	netName := ""
	if networkName.Valid && networkName.String != "" {
		netName = networkName.String
	} else {
		netName = fmt.Sprintf("devarch-%s-net", stackName)
	}

	gen := compose.NewGenerator(h.db, netName)
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		gen.SetProjectRoot(root)
	}
	if hostRoot := os.Getenv("HOST_PROJECT_ROOT"); hostRoot != "" {
		gen.SetHostProjectRoot(hostRoot)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot != "" {
		if err := gen.MaterializeStackConfigs(stackName, projectRoot); err != nil {
			http.Error(w, fmt.Sprintf("failed to materialize config files: %v", err), http.StatusInternalServerError)
			return
		}
	}

	yamlBytes, warnings, err := gen.GenerateStack(stackName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate compose: %v", err), http.StatusInternalServerError)
		return
	}

	var instanceCount int
	h.db.QueryRow(`
		SELECT COUNT(*) FROM service_instances WHERE stack_id = $1 AND enabled = true AND deleted_at IS NULL
	`, stackID).Scan(&instanceCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"yaml":           string(yamlBytes),
		"warnings":       warnings,
		"instance_count": instanceCount,
	})
}
