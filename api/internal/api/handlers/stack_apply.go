package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/compose"
	"github.com/priz/devarch-api/internal/plan"
)

type applyRequest struct {
	Token string `json:"token"`
}

func (h *StackHandler) Apply(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var req applyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	var stackID int
	var networkName sql.NullString
	err := h.db.QueryRow(`
		SELECT id, network_name
		FROM stacks
		WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID, &networkName)

	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", stackName), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
		return
	}

	var acquired bool
	err = h.db.QueryRowContext(r.Context(), "SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquired)
	if err != nil || !acquired {
		http.Error(w, "Stack is being applied by another session", http.StatusConflict)
		return
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.db.ExecContext(unlockCtx, "SELECT pg_advisory_unlock($1)", stackID)
	}()

	if err := plan.ValidateToken(h.db, stackID, req.Token); err != nil {
		if errors.Is(err, plan.ErrStalePlan) {
			http.Error(w, "Plan is stale — stack was modified since plan was generated. Regenerate plan.", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("failed to validate token: %v", err), http.StatusInternalServerError)
		return
	}

	netName := ""
	if networkName.Valid && networkName.String != "" {
		netName = networkName.String
	} else {
		netName = fmt.Sprintf("devarch-%s-net", stackName)
	}

	labels := map[string]string{
		"devarch.managed_by": "devarch",
		"devarch.stack":      stackName,
	}
	if err := h.containerClient.CreateNetwork(netName, labels); err != nil {
		http.Error(w, fmt.Sprintf("failed to create network: %v", err), http.StatusInternalServerError)
		return
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		http.Error(w, "PROJECT_ROOT environment variable not set", http.StatusInternalServerError)
		return
	}

	gen := compose.NewGenerator(h.db, netName)
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		gen.SetProjectRoot(root)
	}
	if hostRoot := os.Getenv("HOST_PROJECT_ROOT"); hostRoot != "" {
		gen.SetHostProjectRoot(hostRoot)
	}

	if err := gen.MaterializeStackConfigs(stackName, projectRoot); err != nil {
		configDir := filepath.Join(projectRoot, "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		http.Error(w, fmt.Sprintf("failed to materialize config files: %v", err), http.StatusInternalServerError)
		return
	}

	yamlBytes, _, err := gen.GenerateStack(stackName)
	if err != nil {
		configDir := filepath.Join(projectRoot, "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		http.Error(w, fmt.Sprintf("failed to generate compose: %v", err), http.StatusInternalServerError)
		return
	}

	tmpFile, err := os.CreateTemp("", "devarch-apply-*.yml")
	if err != nil {
		configDir := filepath.Join(projectRoot, "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		http.Error(w, fmt.Sprintf("failed to create temp file: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(yamlBytes); err != nil {
		tmpFile.Close()
		configDir := filepath.Join(projectRoot, "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		http.Error(w, fmt.Sprintf("failed to write compose file: %v", err), http.StatusInternalServerError)
		return
	}
	tmpFile.Close()

	output, err := h.containerClient.RunCompose(tmpFile.Name(), "--project-name", "devarch-"+stackName, "up", "-d")
	if err != nil {
		http.Error(w, fmt.Sprintf("compose up failed: %v\n%s", err, output), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "applied",
		"output": output,
	})
}
