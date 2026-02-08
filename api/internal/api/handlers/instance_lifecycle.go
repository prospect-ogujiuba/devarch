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

func (h *InstanceHandler) instanceCompose(stackName string) (projectName string, yamlBytes []byte, err error) {
	var networkName sql.NullString
	err = h.db.QueryRow(`
		SELECT network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&networkName)
	if err != nil {
		return "", nil, fmt.Errorf("lookup stack: %w", err)
	}

	netName := fmt.Sprintf("devarch-%s-net", stackName)
	if networkName.Valid && networkName.String != "" {
		netName = networkName.String
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
			return "", nil, fmt.Errorf("materialize configs: %w", err)
		}
	}

	yaml, _, err := gen.GenerateStack(stackName)
	if err != nil {
		return "", nil, fmt.Errorf("generate compose: %w", err)
	}

	return "devarch-" + stackName, yaml, nil
}

func (h *InstanceHandler) Stop(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	projectName, yamlBytes, err := h.instanceCompose(stackName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate compose: %v", err), http.StatusInternalServerError)
		return
	}

	if err := h.containerClient.StopComposeService(projectName, yamlBytes, instanceName); err != nil {
		http.Error(w, fmt.Sprintf("failed to stop instance: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (h *InstanceHandler) Start(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	projectName, yamlBytes, err := h.instanceCompose(stackName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate compose: %v", err), http.StatusInternalServerError)
		return
	}

	if err := h.containerClient.StartComposeService(projectName, yamlBytes, instanceName); err != nil {
		http.Error(w, fmt.Sprintf("failed to start instance: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (h *InstanceHandler) Restart(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	instanceName := chi.URLParam(r, "instance")

	projectName, yamlBytes, err := h.instanceCompose(stackName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate compose: %v", err), http.StatusInternalServerError)
		return
	}

	if err := h.containerClient.RestartComposeService(projectName, yamlBytes, instanceName); err != nil {
		http.Error(w, fmt.Sprintf("failed to restart instance: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "restarted"})
}
