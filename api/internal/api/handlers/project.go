package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/project"
	"github.com/priz/devarch-api/internal/scanner"
	"github.com/priz/devarch-api/pkg/models"
)

type ProjectHandler struct {
	db              *sql.DB
	scanner         *scanner.Scanner
	controller      *project.Controller
	containerClient *container.Client
}

func NewProjectHandler(db *sql.DB, s *scanner.Scanner, c *project.Controller, cc *container.Client) *ProjectHandler {
	return &ProjectHandler{db: db, scanner: s, controller: c, containerClient: cc}
}

const projectColumns = `p.id, p.name, p.path, p.project_type, p.framework, p.language, p.package_manager,
	p.description, p.version, p.license, p.entry_point, p.has_frontend, p.frontend_framework,
	p.domain, p.proxy_port, p.compose_path, p.service_count, p.dependencies, p.scripts, p.git_remote, p.git_branch,
	p.stack_id, s.name,
	(SELECT COUNT(*) FROM service_instances si WHERE si.stack_id = p.stack_id AND si.deleted_at IS NULL) AS instance_count,
	p.last_scanned_at, p.created_at, p.updated_at`

const projectFrom = ` FROM projects p JOIN stacks s ON s.id = p.stack_id`

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
		&p.InstanceCount,
		&p.LastScannedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == nil {
		p.ResolveNulls()
	}
	return p, err
}

func (h *ProjectHandler) enrichRunningCount(p *models.Project) {
	count, err := h.containerClient.CountRunningWithLabels(map[string]string{
		container.LabelStackID: p.StackName,
	})
	if err == nil {
		p.RunningCount = count
	}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	typeFilter := r.URL.Query().Get("type")
	langFilter := r.URL.Query().Get("language")

	query := `SELECT ` + projectColumns + projectFrom + ` WHERE s.deleted_at IS NULL`
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
		h.enrichRunningCount(&p)
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

	row := h.db.QueryRow(`SELECT `+projectColumns+projectFrom+` WHERE p.name = $1 AND s.deleted_at IS NULL`, name)
	p, err := scanProject(row)
	if err == sql.ErrNoRows {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.enrichRunningCount(&p)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	ProjectType string `json:"project_type"`
	Framework   string `json:"framework"`
	Language    string `json:"language"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	ProxyPort   *int   `json:"proxy_port"`
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if err := container.ValidateName(req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Path == "" {
		req.Path = "/unknown"
	}
	if req.ProjectType == "" {
		req.ProjectType = "unknown"
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var stackID int
	err = tx.QueryRow(`SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL`, req.Name).Scan(&stackID)
	if err == sql.ErrNoRows {
		networkName := container.NetworkName(req.Name)
		err = tx.QueryRow(`
			INSERT INTO stacks (name, network_name, source)
			VALUES ($1, $2, 'project')
			RETURNING id
		`, req.Name, networkName).Scan(&stackID)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				http.Error(w, fmt.Sprintf("name %q already exists", req.Name), http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var projectID int
	err = tx.QueryRow(`
		INSERT INTO projects (name, path, project_type, framework, language, description, domain, proxy_port, stack_id, dependencies, scripts)
		VALUES ($1, $2, $3, NULLIF($4,''), NULLIF($5,''), NULLIF($6,''), NULLIF($7,''), $8, $9, '{}', '{}')
		RETURNING id
	`, req.Name, req.Path, req.ProjectType, req.Framework, req.Language, req.Description, req.Domain, req.ProxyPort, stackID).Scan(&projectID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			http.Error(w, fmt.Sprintf("project %q already exists", req.Name), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`UPDATE stacks SET project_id = $1 WHERE id = $2`, projectID, stackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	row := h.db.QueryRow(`SELECT `+projectColumns+projectFrom+` WHERE p.id = $1`, projectID)
	p, err := scanProject(row)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

type updateProjectRequest struct {
	Path        *string `json:"path"`
	ProjectType *string `json:"project_type"`
	Framework   *string `json:"framework"`
	Language    *string `json:"language"`
	Description *string `json:"description"`
	Domain      *string `json:"domain"`
	ProxyPort   *int    `json:"proxy_port"`
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(`
		UPDATE projects SET
			path = COALESCE($2, path),
			project_type = COALESCE($3, project_type),
			framework = COALESCE(NULLIF($4,''), framework),
			language = COALESCE(NULLIF($5,''), language),
			description = COALESCE(NULLIF($6,''), description),
			domain = COALESCE(NULLIF($7,''), domain),
			proxy_port = COALESCE($8, proxy_port),
			updated_at = NOW()
		WHERE name = $1
	`, name, req.Path, req.ProjectType, req.Framework, req.Language, req.Description, req.Domain, req.ProxyPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	row := h.db.QueryRow(`SELECT `+projectColumns+projectFrom+` WHERE p.name = $1 AND s.deleted_at IS NULL`, name)
	p, err := scanProject(row)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.enrichRunningCount(&p)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var stackID int
	err := h.db.QueryRow(`SELECT stack_id FROM projects WHERE name = $1`, name).Scan(&stackID)
	if err == sql.ErrNoRows {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec(`UPDATE stacks SET deleted_at = NOW(), updated_at = NOW(), project_id = NULL WHERE id = $1`, stackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec(`DELETE FROM projects WHERE name = $1`, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("project %q deleted, stack moved to trash", name),
	})
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

	var stackID int
	if err := h.db.QueryRow(`SELECT stack_id FROM projects WHERE name = $1`, name).Scan(&stackID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.stackServices(w, stackID)
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
