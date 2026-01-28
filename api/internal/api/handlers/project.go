package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/project"
	"github.com/priz/devarch-api/internal/scanner"
	"github.com/priz/devarch-api/pkg/models"
)

type ProjectHandler struct {
	db         *sql.DB
	scanner    *scanner.Scanner
	controller *project.Controller
}

func NewProjectHandler(db *sql.DB, s *scanner.Scanner, c *project.Controller) *ProjectHandler {
	return &ProjectHandler{db: db, scanner: s, controller: c}
}

const projectColumns = `id, name, path, project_type, framework, language, package_manager,
	description, version, license, entry_point, has_frontend, frontend_framework,
	domain, proxy_port, compose_path, service_count, dependencies, scripts, git_remote, git_branch,
	last_scanned_at, created_at, updated_at`

func scanProject(rows interface{ Scan(dest ...interface{}) error }) (models.Project, error) {
	var p models.Project
	err := rows.Scan(
		&p.ID, &p.Name, &p.Path, &p.ProjectType,
		&p.Framework, &p.Language, &p.PackageManager,
		&p.Description, &p.Version, &p.License, &p.EntryPoint,
		&p.HasFrontend, &p.FrontendFramework,
		&p.Domain, &p.ProxyPort,
		&p.ComposePath, &p.ServiceCount,
		&p.Dependencies, &p.Scripts,
		&p.GitRemote, &p.GitBranch,
		&p.LastScannedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == nil {
		p.ResolveNulls()
	}
	return p, err
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	typeFilter := r.URL.Query().Get("type")
	langFilter := r.URL.Query().Get("language")

	query := `SELECT ` + projectColumns + ` FROM projects WHERE 1=1`
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
		p, err := scanProject(rows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

	row := h.db.QueryRow(`SELECT `+projectColumns+` FROM projects WHERE name = $1`, name)
	p, err := scanProject(row)
	if err == sql.ErrNoRows {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
		"scanned":  len(projects),
		"projects": projects,
	})
}

func (h *ProjectHandler) Services(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	rows, err := h.db.Query(`
		SELECT ps.id, ps.project_id, ps.service_name, ps.container_name, ps.image,
			ps.service_type, ps.ports, ps.depends_on
		FROM project_services ps
		JOIN projects p ON p.id = ps.project_id
		WHERE p.name = $1
		ORDER BY ps.service_name`, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []models.ProjectService
	for rows.Next() {
		var s models.ProjectService
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.ServiceName, &s.ContainerName,
			&s.Image, &s.ServiceType, &s.Ports, &s.DependsOn); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.ResolveNulls()
		services = append(services, s)
	}

	if services == nil {
		services = []models.ProjectService{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (h *ProjectHandler) Start(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	output, err := h.controller.Start(r.Context(), name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "output": output})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started", "output": output})
}

func (h *ProjectHandler) Stop(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	output, err := h.controller.Stop(r.Context(), name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "output": output})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped", "output": output})
}

func (h *ProjectHandler) Restart(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	output, err := h.controller.Restart(r.Context(), name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "output": output})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "restarted", "output": output})
}

func (h *ProjectHandler) Status(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	statuses, err := h.controller.Status(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}
