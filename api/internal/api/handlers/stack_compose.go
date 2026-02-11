package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
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
		respond.NotFound(w, r, "stack", stackName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
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
	if ws := os.Getenv("WORKSPACE_ROOT"); ws != "" {
		gen.SetWorkspaceRoot(ws)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot != "" {
		if err := gen.MaterializeStackConfigs(stackName, projectRoot); err != nil {
			respond.InternalError(w, r, err)
			return
		}
	}

	yamlBytes, warnings, err := gen.GenerateStackWithRedaction(stackName, true)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	var instanceCount int
	h.db.QueryRow(`
		SELECT COUNT(*) FROM service_instances WHERE stack_id = $1 AND enabled = true AND deleted_at IS NULL
	`, stackID).Scan(&instanceCount)

	respond.JSON(w, r, http.StatusOK, map[string]interface{}{
		"yaml":           string(yamlBytes),
		"warnings":       warnings,
		"instance_count": instanceCount,
	})
}
