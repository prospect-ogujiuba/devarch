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

const projectColumns = `p.id, p.name, p.path, p.project_type, p.framework, p.language, p.package_manager,
	p.description, p.version, p.license, p.entry_point, p.has_frontend, p.frontend_framework,
	p.domain, p.proxy_port, p.compose_path, p.service_count, p.dependencies, p.scripts, p.git_remote, p.git_branch,
	p.stack_id, s.name,
	p.last_scanned_at, p.created_at, p.updated_at`

const projectFrom = ` FROM projects p LEFT JOIN stacks s ON s.id = p.stack_id`

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
		&p.StackID, &p.StackName,
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

	query := `SELECT ` + projectColumns + projectFrom + ` WHERE 1=1`
	var args []interface{}
	argN := 1

	if typeFilter != "" {
		query += fmt.Sprintf(" AND p.project_type = $%d", argN)
		args = append(args, typeFilter)
		argN++
	}
	if langFilter != "" {
		query += fmt.Sprintf(" AND p.language = $%d", argN)
		args = append(args, langFilter)
		argN++
	}

	query += " ORDER BY p.name"

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

	row := h.db.QueryRow(`SELECT `+projectColumns+projectFrom+` WHERE p.name = $1`, name)
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

	var stackID sql.NullInt32
	if err := h.db.QueryRow(`SELECT stack_id FROM projects WHERE name = $1`, name).Scan(&stackID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if stackID.Valid {
		h.stackServices(w, int(stackID.Int32))
		return
	}

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

func (h *ProjectHandler) stackServices(w http.ResponseWriter, stackID int) {
	rows, err := h.db.Query(`
		SELECT si.id, si.instance_id, si.container_name, si.enabled,
			svc.name AS template_name, svc.image_name, svc.image_tag
		FROM service_instances si
		JOIN services svc ON svc.id = si.template_service_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
		ORDER BY si.instance_id`, stackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type stackSvc struct {
		ID            int    `json:"id"`
		InstanceID    string `json:"instance_id"`
		ContainerName string `json:"container_name"`
		Enabled       bool   `json:"enabled"`
		TemplateName  string `json:"template_name"`
		Image         string `json:"image"`
	}

	var services []stackSvc
	for rows.Next() {
		var s stackSvc
		var imageName, imageTag string
		if err := rows.Scan(&s.ID, &s.InstanceID, &s.ContainerName, &s.Enabled,
			&s.TemplateName, &imageName, &imageTag); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.Image = imageName + ":" + imageTag
		services = append(services, s)
	}

	if services == nil {
		services = []stackSvc{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (h *ProjectHandler) LinkStack(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var body struct {
		StackID *int `json:"stack_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.StackID != nil {
		var exists bool
		err := h.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM stacks WHERE id = $1 AND deleted_at IS NULL)`, *body.StackID).Scan(&exists)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "stack not found", http.StatusNotFound)
			return
		}
	}

	result, err := h.db.Exec(`UPDATE projects SET stack_id = $1, updated_at = NOW() WHERE name = $2`, body.StackID, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	row := h.db.QueryRow(`SELECT `+projectColumns+projectFrom+` WHERE p.name = $1`, name)
	p, err := scanProject(row)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
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

func (h *ProjectHandler) StartService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	service := chi.URLParam(r, "service")
	if err := h.controller.StartService(r.Context(), name, service); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (h *ProjectHandler) StopService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	service := chi.URLParam(r, "service")
	if err := h.controller.StopService(r.Context(), name, service); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (h *ProjectHandler) RestartService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	service := chi.URLParam(r, "service")
	if err := h.controller.RestartService(r.Context(), name, service); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "restarted"})
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
