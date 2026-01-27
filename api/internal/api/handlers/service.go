package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/compose"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/podman"
	"github.com/priz/devarch-api/pkg/models"
)

type ServiceHandler struct {
	db              *sql.DB
	containerClient *container.Client
	podmanClient    *podman.Client
	generator       *compose.Generator
}

func NewServiceHandler(db *sql.DB, cc *container.Client, pc *podman.Client) *ServiceHandler {
	return &ServiceHandler{
		db:              db,
		containerClient: cc,
		podmanClient:    pc,
		generator:       compose.NewGenerator(db, "microservices-net"),
	}
}

func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT s.id, s.name, s.category_id, s.image_name, s.image_tag,
			s.restart_policy, s.command, s.user_spec, s.enabled,
			s.created_at, s.updated_at, c.name as category_name
		FROM services s
		JOIN categories c ON s.category_id = c.id
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if cat := r.URL.Query().Get("category"); cat != "" {
		query += " AND c.name = $" + strconv.Itoa(argIdx)
		args = append(args, cat)
		argIdx++
	}

	if search := r.URL.Query().Get("search"); search != "" {
		query += " AND s.name ILIKE $" + strconv.Itoa(argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	if enabled := r.URL.Query().Get("enabled"); enabled != "" {
		query += " AND s.enabled = $" + strconv.Itoa(argIdx)
		args = append(args, enabled == "true")
		argIdx++
	}

	sortCol := "s.name"
	if sort := r.URL.Query().Get("sort"); sort != "" {
		switch sort {
		case "name", "category", "image_name", "created_at", "updated_at":
			sortCol = "s." + sort
		}
	}
	order := "ASC"
	if r.URL.Query().Get("order") == "desc" {
		order = "DESC"
	}
	query += " ORDER BY " + sortCol + " " + order

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}
	query += " LIMIT $" + strconv.Itoa(argIdx)
	args = append(args, limit)
	argIdx++

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 1 {
			offset := (p - 1) * limit
			query += " OFFSET $" + strconv.Itoa(argIdx)
			args = append(args, offset)
		}
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		var catName string
		if err := rows.Scan(
			&s.ID, &s.Name, &s.CategoryID, &s.ImageName, &s.ImageTag,
			&s.RestartPolicy, &s.Command, &s.UserSpec, &s.Enabled,
			&s.CreatedAt, &s.UpdatedAt, &catName,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.Category = &models.Category{Name: catName}
		if s.Command.Valid {
			s.CommandStr = s.Command.String
		}
		if s.UserSpec.Valid {
			s.UserSpecStr = s.UserSpec.String
		}
		services = append(services, s)
	}

	if r.URL.Query().Get("include") != "" {
		ctx := r.Context()
		for i := range services {
			h.loadServiceIncludes(ctx, &services[i], r.URL.Query().Get("include"))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (h *ServiceHandler) loadServiceIncludes(ctx context.Context, s *models.Service, includes string) {
	for _, inc := range []string{"status", "metrics"} {
		if !containsInclude(includes, inc) {
			continue
		}
		switch inc {
		case "status":
			state, err := h.podmanClient.GetServiceState(ctx, s.Name)
			if err == nil {
				s.Status = state
			}
		case "metrics":
			metrics, err := h.podmanClient.GetServiceMetrics(ctx, s.Name)
			if err == nil {
				s.Metrics = metrics
			}
		}
	}
}

func containsInclude(includes, target string) bool {
	for _, inc := range []string{target} {
		if includes == inc || includes == "all" {
			return true
		}
	}
	return false
}

func (h *ServiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var s models.Service
	var catName string
	err := h.db.QueryRow(`
		SELECT s.id, s.name, s.category_id, s.image_name, s.image_tag,
			s.restart_policy, s.command, s.user_spec, s.enabled,
			s.created_at, s.updated_at, c.name as category_name
		FROM services s
		JOIN categories c ON s.category_id = c.id
		WHERE s.name = $1
	`, name).Scan(
		&s.ID, &s.Name, &s.CategoryID, &s.ImageName, &s.ImageTag,
		&s.RestartPolicy, &s.Command, &s.UserSpec, &s.Enabled,
		&s.CreatedAt, &s.UpdatedAt, &catName,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.Category = &models.Category{Name: catName}
	if s.Command.Valid {
		s.CommandStr = s.Command.String
	}
	if s.UserSpec.Valid {
		s.UserSpecStr = s.UserSpec.String
	}

	h.loadServiceRelations(&s)
	h.loadServiceIncludes(r.Context(), &s, r.URL.Query().Get("include"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func (h *ServiceHandler) loadServiceRelations(s *models.Service) {
	rows, _ := h.db.Query(`SELECT host_ip, host_port, container_port, protocol FROM service_ports WHERE service_id = $1`, s.ID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var p models.ServicePort
			rows.Scan(&p.HostIP, &p.HostPort, &p.ContainerPort, &p.Protocol)
			s.Ports = append(s.Ports, p)
		}
	}

	rows, _ = h.db.Query(`SELECT volume_type, source, target, read_only FROM service_volumes WHERE service_id = $1`, s.ID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var v models.ServiceVolume
			rows.Scan(&v.VolumeType, &v.Source, &v.Target, &v.ReadOnly)
			s.Volumes = append(s.Volumes, v)
		}
	}

	rows, _ = h.db.Query(`SELECT key, value, is_secret FROM service_env_vars WHERE service_id = $1`, s.ID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var e models.ServiceEnvVar
			rows.Scan(&e.Key, &e.Value, &e.IsSecret)
			if e.IsSecret {
				e.Value = "***"
			}
			s.EnvVars = append(s.EnvVars, e)
		}
	}

	rows, _ = h.db.Query(`SELECT key, value FROM service_labels WHERE service_id = $1`, s.ID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var l models.ServiceLabel
			rows.Scan(&l.Key, &l.Value)
			s.Labels = append(s.Labels, l)
		}
	}

	rows, _ = h.db.Query(`SELECT srv.name FROM service_dependencies d JOIN services srv ON d.depends_on_service_id = srv.id WHERE d.service_id = $1`, s.ID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			rows.Scan(&name)
			s.Dependencies = append(s.Dependencies, name)
		}
	}

	var hc models.ServiceHealthcheck
	err := h.db.QueryRow(`SELECT test, interval_seconds, timeout_seconds, retries, start_period_seconds FROM service_healthchecks WHERE service_id = $1`, s.ID).
		Scan(&hc.Test, &hc.IntervalSeconds, &hc.TimeoutSeconds, &hc.Retries, &hc.StartPeriodSeconds)
	if err == nil {
		s.Healthcheck = &hc
	}
}

type createServiceRequest struct {
	Name          string                 `json:"name"`
	CategoryID    int                    `json:"category_id"`
	ImageName     string                 `json:"image_name"`
	ImageTag      string                 `json:"image_tag"`
	RestartPolicy string                 `json:"restart_policy"`
	Command       string                 `json:"command"`
	UserSpec      string                 `json:"user_spec"`
	Ports         []models.ServicePort   `json:"ports"`
	Volumes       []models.ServiceVolume `json:"volumes"`
	EnvVars       []models.ServiceEnvVar `json:"env_vars"`
	Dependencies  []string               `json:"dependencies"`
}

func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ImageTag == "" {
		req.ImageTag = "latest"
	}
	if req.RestartPolicy == "" {
		req.RestartPolicy = "unless-stopped"
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var serviceID int
	err = tx.QueryRow(`
		INSERT INTO services (name, category_id, image_name, image_tag, restart_policy, command, user_spec)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''))
		RETURNING id
	`, req.Name, req.CategoryID, req.ImageName, req.ImageTag, req.RestartPolicy, req.Command, req.UserSpec).Scan(&serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, port := range req.Ports {
		_, err = tx.Exec(`
			INSERT INTO service_ports (service_id, host_ip, host_port, container_port, protocol)
			VALUES ($1, $2, $3, $4, $5)
		`, serviceID, port.HostIP, port.HostPort, port.ContainerPort, port.Protocol)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, vol := range req.Volumes {
		_, err = tx.Exec(`
			INSERT INTO service_volumes (service_id, volume_type, source, target, read_only)
			VALUES ($1, $2, $3, $4, $5)
		`, serviceID, vol.VolumeType, vol.Source, vol.Target, vol.ReadOnly)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, env := range req.EnvVars {
		_, err = tx.Exec(`
			INSERT INTO service_env_vars (service_id, key, value, is_secret)
			VALUES ($1, $2, $3, $4)
		`, serviceID, env.Key, env.Value, env.IsSecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": serviceID})
}

func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var serviceID int
	err := h.db.QueryRow("SELECT id FROM services WHERE name = $1", name).Scan(&serviceID)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req createServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec(`
		UPDATE services SET
			image_name = COALESCE(NULLIF($1, ''), image_name),
			image_tag = COALESCE(NULLIF($2, ''), image_tag),
			restart_policy = COALESCE(NULLIF($3, ''), restart_policy),
			command = NULLIF($4, ''),
			user_spec = NULLIF($5, ''),
			updated_at = NOW()
		WHERE id = $6
	`, req.ImageName, req.ImageTag, req.RestartPolicy, req.Command, req.UserSpec, serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	result, err := h.db.Exec("DELETE FROM services WHERE name = $1", name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *ServiceHandler) Start(w http.ResponseWriter, r *http.Request) {
	if h.containerClient == nil {
		http.Error(w, "compose operations unavailable", http.StatusServiceUnavailable)
		return
	}

	name := chi.URLParam(r, "name")

	var s models.Service
	err := h.db.QueryRow(`SELECT id, name, image_name, image_tag, restart_policy, command, user_spec FROM services WHERE name = $1`, name).
		Scan(&s.ID, &s.Name, &s.ImageName, &s.ImageTag, &s.RestartPolicy, &s.Command, &s.UserSpec)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	composeYAML, err := h.generator.Generate(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.containerClient.StartService(name, composeYAML); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (h *ServiceHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if h.containerClient == nil {
		http.Error(w, "compose operations unavailable", http.StatusServiceUnavailable)
		return
	}

	name := chi.URLParam(r, "name")

	var s models.Service
	err := h.db.QueryRow(`SELECT id, name, image_name, image_tag FROM services WHERE name = $1`, name).
		Scan(&s.ID, &s.Name, &s.ImageName, &s.ImageTag)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	composeYAML, err := h.generator.Generate(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.containerClient.StopService(name, composeYAML); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (h *ServiceHandler) Restart(w http.ResponseWriter, r *http.Request) {
	if h.containerClient == nil {
		http.Error(w, "compose operations unavailable", http.StatusServiceUnavailable)
		return
	}

	name := chi.URLParam(r, "name")

	var s models.Service
	err := h.db.QueryRow(`SELECT id, name, image_name, image_tag, restart_policy, command, user_spec FROM services WHERE name = $1`, name).
		Scan(&s.ID, &s.Name, &s.ImageName, &s.ImageTag, &s.RestartPolicy, &s.Command, &s.UserSpec)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	composeYAML, err := h.generator.Generate(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.containerClient.RestartService(name, composeYAML); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "restarted"})
}

func (h *ServiceHandler) Rebuild(w http.ResponseWriter, r *http.Request) {
	if h.containerClient == nil {
		http.Error(w, "compose operations unavailable", http.StatusServiceUnavailable)
		return
	}

	name := chi.URLParam(r, "name")

	var s models.Service
	err := h.db.QueryRow(`SELECT id, name, image_name, image_tag, restart_policy, command, user_spec FROM services WHERE name = $1`, name).
		Scan(&s.ID, &s.Name, &s.ImageName, &s.ImageTag, &s.RestartPolicy, &s.Command, &s.UserSpec)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	composeYAML, err := h.generator.Generate(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	noCache := r.URL.Query().Get("no_cache") == "true"
	if err := h.containerClient.RebuildService(name, composeYAML, noCache); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "rebuilt"})
}

func (h *ServiceHandler) Status(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	status, err := h.podmanClient.GetServiceState(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *ServiceHandler) Logs(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	tailStr := r.URL.Query().Get("tail")
	tail := 100
	if tailStr != "" {
		if parsed, err := strconv.Atoi(tailStr); err == nil && parsed > 0 {
			tail = parsed
		}
	}

	logs, err := h.podmanClient.ContainerLogsString(r.Context(), name, tail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(logs))
}

func (h *ServiceHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	metrics, err := h.podmanClient.GetServiceMetrics(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *ServiceHandler) Compose(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var s models.Service
	err := h.db.QueryRow(`SELECT id, name, image_name, image_tag, restart_policy, command, user_spec FROM services WHERE name = $1`, name).
		Scan(&s.ID, &s.Name, &s.ImageName, &s.ImageTag, &s.RestartPolicy, &s.Command, &s.UserSpec)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	composeYAML, err := h.generator.Generate(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/yaml")
	w.Write(composeYAML)
}

type bulkRequest struct {
	Action   string   `json:"action"`
	Services []string `json:"services"`
}

func (h *ServiceHandler) Bulk(w http.ResponseWriter, r *http.Request) {
	if h.containerClient == nil {
		http.Error(w, "compose operations unavailable", http.StatusServiceUnavailable)
		return
	}

	var req bulkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results := make(map[string]string)
	for _, name := range req.Services {
		var s models.Service
		err := h.db.QueryRow(`SELECT id, name, image_name, image_tag, restart_policy, command, user_spec FROM services WHERE name = $1`, name).
			Scan(&s.ID, &s.Name, &s.ImageName, &s.ImageTag, &s.RestartPolicy, &s.Command, &s.UserSpec)
		if err != nil {
			results[name] = "not found"
			continue
		}

		composeYAML, err := h.generator.Generate(&s)
		if err != nil {
			results[name] = "compose error: " + err.Error()
			continue
		}

		var opErr error
		switch req.Action {
		case "start":
			opErr = h.containerClient.StartService(name, composeYAML)
		case "stop":
			opErr = h.containerClient.StopService(name, composeYAML)
		case "restart":
			opErr = h.containerClient.RestartService(name, composeYAML)
		default:
			results[name] = "unknown action"
			continue
		}

		if opErr != nil {
			results[name] = "error: " + opErr.Error()
		} else {
			results[name] = "ok"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
