package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/scanner"
	"github.com/priz/devarch-api/pkg/models"
)

type ProjectHandler struct {
	db      *sql.DB
	scanner *scanner.Scanner
}

func NewProjectHandler(db *sql.DB, s *scanner.Scanner) *ProjectHandler {
	return &ProjectHandler{db: db, scanner: s}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	typeFilter := r.URL.Query().Get("type")
	langFilter := r.URL.Query().Get("language")

	query := `SELECT id, name, path, project_type, framework, language, package_manager,
		description, version, license, entry_point, has_frontend, frontend_framework,
		domain, proxy_port, dependencies, scripts, git_remote, git_branch,
		last_scanned_at, created_at, updated_at
		FROM projects WHERE 1=1`
	var args []interface{}
	argN := 1

	if typeFilter != "" {
		query += fmt.Sprintf(" AND project_type = $%d", argN)
		args = append(args, typeFilter)
		argN++
	}
	if langFilter != "" {
		query += fmt.Sprintf(" AND language = $%d", argN)
		args = append(args, langFilter)
		argN++
	}

	query += " ORDER BY name"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Path, &p.ProjectType,
			&p.Framework, &p.Language, &p.PackageManager,
			&p.Description, &p.Version, &p.License, &p.EntryPoint,
			&p.HasFrontend, &p.FrontendFramework,
			&p.Domain, &p.ProxyPort,
			&p.Dependencies, &p.Scripts,
			&p.GitRemote, &p.GitBranch,
			&p.LastScannedAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.ResolveNulls()
		projects = append(projects, p)
	}

	if projects == nil {
		projects = []models.Project{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var p models.Project
	err := h.db.QueryRow(`SELECT id, name, path, project_type, framework, language, package_manager,
		description, version, license, entry_point, has_frontend, frontend_framework,
		domain, proxy_port, dependencies, scripts, git_remote, git_branch,
		last_scanned_at, created_at, updated_at
		FROM projects WHERE name = $1`, name).Scan(
		&p.ID, &p.Name, &p.Path, &p.ProjectType,
		&p.Framework, &p.Language, &p.PackageManager,
		&p.Description, &p.Version, &p.License, &p.EntryPoint,
		&p.HasFrontend, &p.FrontendFramework,
		&p.Domain, &p.ProxyPort,
		&p.Dependencies, &p.Scripts,
		&p.GitRemote, &p.GitBranch,
		&p.LastScannedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p.ResolveNulls()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProjectHandler) Scan(w http.ResponseWriter, r *http.Request) {
	projects, err := h.scanner.ScanAndPersist()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scanned": len(projects),
		"projects": projects,
	})
}

