package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/priz/devarch-api/internal/api/respond"
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

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
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

// List godoc
// @Summary      List all projects
// @Description  Returns all projects with optional type and language filters
// @Tags         projects
// @Produce      json
// @Param        type query string false "Filter by project type"
// @Param        language query string false "Filter by programming language"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects [get]
// @Security     ApiKeyAuth
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
		respond.InternalError(w, r, err)
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			respond.InternalError(w, r, err)
			return
		}
		h.enrichRunningCount(&p)
		projects = append(projects, p)
	}

	if projects == nil {
		projects = []models.Project{}
	}

	respond.JSON(w, r, http.StatusOK, projects)
}

// Get godoc
// @Summary      Get project by name
// @Description  Returns a single project by its name
// @Tags         projects
// @Produce      json
// @Param        name path string true "Project name"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name} [get]
// @Security     ApiKeyAuth
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	row := h.db.QueryRow(`SELECT `+projectColumns+projectFrom+` WHERE p.name = $1 AND s.deleted_at IS NULL`, name)
	p, err := scanProject(row)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "project", name)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	h.enrichRunningCount(&p)

	respond.JSON(w, r, http.StatusOK, p)
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

// Create godoc
// @Summary      Create a new project
// @Description  Creates a new project with associated stack
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project body createProjectRequest true "Project details"
// @Success      201 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects [post]
// @Security     ApiKeyAuth
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		respond.BadRequest(w, r, "invalid request body")
		return
	}

	if req.Name == "" {
		respond.BadRequest(w, r, "name is required")
		return
	}
	if err := container.ValidateName(req.Name); err != nil {
		respond.BadRequest(w, r, err.Error())
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
		respond.InternalError(w, r, err)
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
				respond.Conflict(w, r, fmt.Sprintf("name %q already exists", req.Name))
				return
			}
			respond.InternalError(w, r, err)
			return
		}
	} else if err != nil {
		respond.InternalError(w, r, err)
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
			respond.Conflict(w, r, fmt.Sprintf("project %q already exists", req.Name))
			return
		}
		respond.InternalError(w, r, err)
		return
	}

	_, err = tx.Exec(`UPDATE stacks SET project_id = $1 WHERE id = $2`, projectID, stackID)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, err)
		return
	}

	row := h.db.QueryRow(`SELECT `+projectColumns+projectFrom+` WHERE p.id = $1`, projectID)
	p, err := scanProject(row)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	respond.JSON(w, r, http.StatusCreated, p)
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

// Update godoc
// @Summary      Update project
// @Description  Updates an existing project's properties
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        name path string true "Project name"
// @Param        project body updateProjectRequest true "Updated project details"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name} [put]
// @Security     ApiKeyAuth
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var req updateProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		respond.BadRequest(w, r, "invalid request body")
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
		respond.InternalError(w, r, err)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		respond.NotFound(w, r, "project", name)
		return
	}

	row := h.db.QueryRow(`SELECT `+projectColumns+projectFrom+` WHERE p.name = $1 AND s.deleted_at IS NULL`, name)
	p, err := scanProject(row)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	h.enrichRunningCount(&p)

	respond.JSON(w, r, http.StatusOK, p)
}

// Delete godoc
// @Summary      Delete project
// @Description  Deletes a project and moves its stack to trash
// @Tags         projects
// @Produce      json
// @Param        name path string true "Project name"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name} [delete]
// @Security     ApiKeyAuth
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var stackID int
	err := h.db.QueryRow(`SELECT stack_id FROM projects WHERE name = $1`, name).Scan(&stackID)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "project", name)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	_, err = h.db.Exec(`UPDATE stacks SET deleted_at = NOW(), updated_at = NOW(), project_id = NULL WHERE id = $1`, stackID)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	_, err = h.db.Exec(`DELETE FROM projects WHERE name = $1`, name)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("project %q deleted, stack moved to trash", name),
	})
}

// Scan godoc
// @Summary      Scan for projects
// @Description  Scans filesystem for projects and persists them to database
// @Tags         projects
// @Produce      json
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/scan [post]
// @Security     ApiKeyAuth
func (h *ProjectHandler) Scan(w http.ResponseWriter, r *http.Request) {
	projects, err := h.scanner.ScanAndPersist()
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]interface{}{
		"scanned":  len(projects),
		"projects": projects,
	})
}

// Services godoc
// @Summary      List project services
// @Description  Returns all service instances that belong to the project's stack
// @Tags         projects
// @Produce      json
// @Param        name path string true "Project name"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name}/services [get]
// @Security     ApiKeyAuth
func (h *ProjectHandler) Services(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var stackID int
	if err := h.db.QueryRow(`SELECT stack_id FROM projects WHERE name = $1`, name).Scan(&stackID); err != nil {
		if err == sql.ErrNoRows {
			respond.NotFound(w, r, "project", name)
			return
		}
		respond.InternalError(w, r, err)
		return
	}

	h.stackServices(w, r, stackID)
}

func (h *ProjectHandler) stackServices(w http.ResponseWriter, r *http.Request, stackID int) {
	rows, err := h.db.Query(`
		SELECT si.id, si.instance_id, si.container_name, si.enabled,
			svc.name AS template_name, svc.image_name, svc.image_tag
		FROM service_instances si
		JOIN services svc ON svc.id = si.template_service_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
		ORDER BY si.instance_id`, stackID)
	if err != nil {
		respond.InternalError(w, r, err)
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
			respond.InternalError(w, r, err)
			return
		}
		s.Image = imageName + ":" + imageTag
		services = append(services, s)
	}

	if services == nil {
		services = []stackSvc{}
	}

	respond.JSON(w, r, http.StatusOK, services)
}

// Start godoc
// @Summary      Start project
// @Description  Starts all services in the project's stack
// @Tags         projects
// @Produce      json
// @Param        name path string true "Project name"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name}/start [post]
// @Security     ApiKeyAuth
func (h *ProjectHandler) Start(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	output, err := h.controller.Start(r.Context(), name)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	respond.Action(w, r, http.StatusOK, "started", respond.WithOutput(output))
}

// Stop godoc
// @Summary      Stop project
// @Description  Stops all services in the project's stack
// @Tags         projects
// @Produce      json
// @Param        name path string true "Project name"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name}/stop [post]
// @Security     ApiKeyAuth
func (h *ProjectHandler) Stop(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	output, err := h.controller.Stop(r.Context(), name)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	respond.Action(w, r, http.StatusOK, "stopped", respond.WithOutput(output))
}

// Restart godoc
// @Summary      Restart project
// @Description  Restarts all services in the project's stack
// @Tags         projects
// @Produce      json
// @Param        name path string true "Project name"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name}/restart [post]
// @Security     ApiKeyAuth
func (h *ProjectHandler) Restart(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	output, err := h.controller.Restart(r.Context(), name)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	respond.Action(w, r, http.StatusOK, "restarted", respond.WithOutput(output))
}

// StartService godoc
// @Summary      Start project service
// @Description  Starts a specific service within the project's stack
// @Tags         projects
// @Produce      json
// @Param        name path string true "Project name"
// @Param        service path string true "Service instance ID"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name}/services/{service}/start [post]
// @Security     ApiKeyAuth
func (h *ProjectHandler) StartService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	service := chi.URLParam(r, "service")
	if err := h.controller.StartService(r.Context(), name, service); err != nil {
		respond.InternalError(w, r, err)
		return
	}
	respond.Action(w, r, http.StatusOK, "started")
}

// StopService godoc
// @Summary      Stop project service
// @Description  Stops a specific service within the project's stack
// @Tags         projects
// @Produce      json
// @Param        name path string true "Project name"
// @Param        service path string true "Service instance ID"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name}/services/{service}/stop [post]
// @Security     ApiKeyAuth
func (h *ProjectHandler) StopService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	service := chi.URLParam(r, "service")
	if err := h.controller.StopService(r.Context(), name, service); err != nil {
		respond.InternalError(w, r, err)
		return
	}
	respond.Action(w, r, http.StatusOK, "stopped")
}

// RestartService godoc
// @Summary      Restart project service
// @Description  Restarts a specific service within the project's stack
// @Tags         projects
// @Produce      json
// @Param        name path string true "Project name"
// @Param        service path string true "Service instance ID"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name}/services/{service}/restart [post]
// @Security     ApiKeyAuth
func (h *ProjectHandler) RestartService(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	service := chi.URLParam(r, "service")
	if err := h.controller.RestartService(r.Context(), name, service); err != nil {
		respond.InternalError(w, r, err)
		return
	}
	respond.Action(w, r, http.StatusOK, "restarted")
}

// Status godoc
// @Summary      Get project status
// @Description  Returns runtime status of all services in the project's stack
// @Tags         projects
// @Produce      json
// @Param        name path string true "Project name"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /projects/{name}/status [get]
// @Security     ApiKeyAuth
func (h *ProjectHandler) Status(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	statuses, err := h.controller.Status(r.Context(), name)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	respond.JSON(w, r, http.StatusOK, statuses)
}
