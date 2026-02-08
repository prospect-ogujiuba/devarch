package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/compose"
)

func (h *StackHandler) Stop(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	yamlBytes, err := h.stackCompose(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate compose: %v", err), http.StatusInternalServerError)
		return
	}

	if err := h.containerClient.StopStack("devarch-"+name, yamlBytes); err != nil {
		http.Error(w, fmt.Sprintf("failed to stop stack: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (h *StackHandler) Start(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var enabled bool
	err := h.db.QueryRow(`
		SELECT enabled FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(&enabled)
	if err != nil {
		http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)
		return
	}
	if !enabled {
		http.Error(w, "stack is disabled — enable it first", http.StatusConflict)
		return
	}

	var networkName *string
	h.db.QueryRow(`SELECT network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL`, name).Scan(&networkName)

	netName := fmt.Sprintf("devarch-%s-net", name)
	if networkName != nil && *networkName != "" {
		netName = *networkName
	}

	labels := map[string]string{
		"devarch.managed_by": "devarch",
		"devarch.stack":      name,
	}
	if err := h.containerClient.CreateNetwork(netName, labels); err != nil {
		http.Error(w, fmt.Sprintf("failed to create network: %v", err), http.StatusInternalServerError)
		return
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
		if err := gen.MaterializeStackConfigs(name, projectRoot); err != nil {
			http.Error(w, fmt.Sprintf("failed to materialize configs: %v", err), http.StatusInternalServerError)
			return
		}
	}

	yamlBytes, _, err := gen.GenerateStack(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate compose: %v", err), http.StatusInternalServerError)
		return
	}

	if err := h.containerClient.StartService("devarch-"+name, yamlBytes); err != nil {
		http.Error(w, fmt.Sprintf("failed to start stack: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (h *StackHandler) Restart(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var enabled bool
	err := h.db.QueryRow(`
		SELECT enabled FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(&enabled)
	if err != nil {
		http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)
		return
	}
	if !enabled {
		http.Error(w, "stack is disabled — enable it first", http.StatusConflict)
		return
	}

	stopYaml, err := h.stackCompose(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate compose: %v", err), http.StatusInternalServerError)
		return
	}

	h.containerClient.StopStack("devarch-"+name, stopYaml)

	var networkName *string
	h.db.QueryRow(`SELECT network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL`, name).Scan(&networkName)

	netName := fmt.Sprintf("devarch-%s-net", name)
	if networkName != nil && *networkName != "" {
		netName = *networkName
	}

	labels := map[string]string{
		"devarch.managed_by": "devarch",
		"devarch.stack":      name,
	}
	if err := h.containerClient.CreateNetwork(netName, labels); err != nil {
		http.Error(w, fmt.Sprintf("failed to create network: %v", err), http.StatusInternalServerError)
		return
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
		if err := gen.MaterializeStackConfigs(name, projectRoot); err != nil {
			http.Error(w, fmt.Sprintf("failed to materialize configs: %v", err), http.StatusInternalServerError)
			return
		}
	}

	yamlBytes, _, err := gen.GenerateStack(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate compose: %v", err), http.StatusInternalServerError)
		return
	}

	if err := h.containerClient.StartService("devarch-"+name, yamlBytes); err != nil {
		http.Error(w, fmt.Sprintf("failed to start stack: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "restarted"})
}
