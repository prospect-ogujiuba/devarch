package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/compose"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/crypto"
	"github.com/priz/devarch-api/internal/podman"
	"github.com/priz/devarch-api/pkg/models"
)

type ServiceHandler struct {
	db              *sql.DB
	containerClient *container.Client
	podmanClient    *podman.Client
	generator       *compose.Generator
	validator       *compose.Validator
	projectRoot     string
	cipher          *crypto.Cipher
}

func NewServiceHandler(db *sql.DB, cc *container.Client, pc *podman.Client, cipher *crypto.Cipher) *ServiceHandler {
	projectRoot := os.Getenv("PROJECT_ROOT")
	gen := compose.NewGenerator(db, "microservices-net")
	if projectRoot != "" {
		gen.SetProjectRoot(projectRoot)
	}
	if hostRoot := os.Getenv("HOST_PROJECT_ROOT"); hostRoot != "" {
		gen.SetHostProjectRoot(hostRoot)
	}
	if ws := os.Getenv("WORKSPACE_ROOT"); ws != "" {
		gen.SetWorkspaceRoot(ws)
	}
	return &ServiceHandler{
		db:              db,
		containerClient: cc,
		podmanClient:    pc,
		generator:       gen,
		validator:       compose.NewValidator(db),
		projectRoot:     projectRoot,
		cipher:          cipher,
	}
}

func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT s.id, s.name, s.category_id, s.image_name, s.image_tag,
			s.restart_policy, s.command, s.user_spec, s.enabled,
			s.created_at, s.updated_at, c.name as category_name,
			COALESCE(s.compose_overrides, '{}')
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
		case "name", "image_name", "created_at", "updated_at":
			sortCol = "s." + sort
		case "category":
			sortCol = "c.name"
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
			&s.ComposeOverrides,
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
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("include") != "" {
		ctx := r.Context()
		for i := range services {
			h.loadServiceIncludes(ctx, &services[i], r.URL.Query().Get("include"))
		}
	}

	var total int
	h.db.QueryRow("SELECT COUNT(*) FROM services s JOIN categories c ON s.category_id = c.id WHERE 1=1").Scan(&total)

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	totalPages := (total + limit - 1) / limit

	if services == nil {
		services = []models.Service{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	w.Header().Set("X-Page", strconv.Itoa(page))
	w.Header().Set("X-Per-Page", strconv.Itoa(limit))
	w.Header().Set("X-Total-Pages", strconv.Itoa(totalPages))
	if err := json.NewEncoder(w).Encode(services); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

var serviceNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}[a-z0-9]$`)

func isValidServiceName(name string) bool {
	if len(name) < 2 || len(name) > 64 {
		return false
	}
	return serviceNameRe.MatchString(name)
}

func containsInclude(includes, target string) bool {
	if includes == "all" {
		return true
	}
	for _, inc := range strings.Split(includes, ",") {
		if strings.TrimSpace(inc) == target {
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
			s.created_at, s.updated_at, c.name as category_name,
			s.config_status, s.last_validated_at, s.validation_errors,
			s.compose_overrides
		FROM services s
		JOIN categories c ON s.category_id = c.id
		WHERE s.name = $1
	`, name).Scan(
		&s.ID, &s.Name, &s.CategoryID, &s.ImageName, &s.ImageTag,
		&s.RestartPolicy, &s.Command, &s.UserSpec, &s.Enabled,
		&s.CreatedAt, &s.UpdatedAt, &catName,
		&s.ConfigStatus, &s.LastValidatedAt, &s.ValidationErrors,
		&s.ComposeOverrides,
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
	if s.LastValidatedAt.Valid {
		s.LastValidatedStr = &s.LastValidatedAt.Time
	}

	h.loadServiceRelations(&s)
	h.loadServiceIncludes(r.Context(), &s, r.URL.Query().Get("include"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func (h *ServiceHandler) loadServiceByName(name string) (*models.Service, error) {
	var s models.Service
	err := h.db.QueryRow(
		`SELECT id, name, image_name, image_tag, restart_policy, command, user_spec, compose_overrides, container_name_template FROM services WHERE name = $1`,
		name,
	).Scan(&s.ID, &s.Name, &s.ImageName, &s.ImageTag, &s.RestartPolicy, &s.Command, &s.UserSpec, &s.ComposeOverrides, &s.ContainerNameTemplate)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (h *ServiceHandler) loadServiceRelations(s *models.Service) {
	if rows, err := h.db.Query(`SELECT host_ip, host_port, container_port, protocol FROM service_ports WHERE service_id = $1`, s.ID); err != nil {
		log.Printf("loadServiceRelations: ports query error for service %d: %v", s.ID, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var p models.ServicePort
			if err := rows.Scan(&p.HostIP, &p.HostPort, &p.ContainerPort, &p.Protocol); err != nil {
				log.Printf("loadServiceRelations: ports scan error: %v", err)
				continue
			}
			s.Ports = append(s.Ports, p)
		}
		if err := rows.Err(); err != nil {
			log.Printf("loadServiceRelations: ports iteration error: %v", err)
		}
	}

	if rows, err := h.db.Query(`SELECT volume_type, source, target, read_only, is_external FROM service_volumes WHERE service_id = $1`, s.ID); err != nil {
		log.Printf("loadServiceRelations: volumes query error for service %d: %v", s.ID, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var v models.ServiceVolume
			if err := rows.Scan(&v.VolumeType, &v.Source, &v.Target, &v.ReadOnly, &v.IsExternal); err != nil {
				log.Printf("loadServiceRelations: volumes scan error: %v", err)
				continue
			}
			s.Volumes = append(s.Volumes, v)
		}
		if err := rows.Err(); err != nil {
			log.Printf("loadServiceRelations: volumes iteration error: %v", err)
		}
	}

	if rows, err := h.db.Query(`SELECT key, value, is_secret, encrypted_value, encryption_version FROM service_env_vars WHERE service_id = $1`, s.ID); err != nil {
		log.Printf("loadServiceRelations: env_vars query error for service %d: %v", s.ID, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var e models.ServiceEnvVar
			var encryptedValue sql.NullString
			var encryptionVersion int
			if err := rows.Scan(&e.Key, &e.Value, &e.IsSecret, &encryptedValue, &encryptionVersion); err != nil {
				log.Printf("loadServiceRelations: env_vars scan error: %v", err)
				continue
			}

			if encryptionVersion > 0 && encryptedValue.Valid {
				_, err := h.cipher.Decrypt(encryptedValue.String)
				if err == nil {
					e.Value = "***"
				}
			} else if e.IsSecret && encryptionVersion == 0 && e.Value != "" {
				encrypted, err := h.cipher.Encrypt(e.Value)
				if err == nil {
					h.db.Exec(`UPDATE service_env_vars SET encrypted_value = $1, encryption_version = 1, value = '' WHERE service_id = $2 AND key = $3`, encrypted, s.ID, e.Key)
				}
				e.Value = "***"
			} else if e.IsSecret {
				e.Value = "***"
			}

			s.EnvVars = append(s.EnvVars, e)
		}
		if err := rows.Err(); err != nil {
			log.Printf("loadServiceRelations: env_vars iteration error: %v", err)
		}
	}

	if rows, err := h.db.Query(`SELECT key, value FROM service_labels WHERE service_id = $1`, s.ID); err != nil {
		log.Printf("loadServiceRelations: labels query error for service %d: %v", s.ID, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var l models.ServiceLabel
			if err := rows.Scan(&l.Key, &l.Value); err != nil {
				log.Printf("loadServiceRelations: labels scan error: %v", err)
				continue
			}
			s.Labels = append(s.Labels, l)
		}
		if err := rows.Err(); err != nil {
			log.Printf("loadServiceRelations: labels iteration error: %v", err)
		}
	}

	if rows, err := h.db.Query(`SELECT domain, proxy_port FROM service_domains WHERE service_id = $1`, s.ID); err != nil {
		log.Printf("loadServiceRelations: domains query error for service %d: %v", s.ID, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var d models.ServiceDomain
			if err := rows.Scan(&d.Domain, &d.ProxyPort); err != nil {
				log.Printf("loadServiceRelations: domains scan error: %v", err)
				continue
			}
			s.Domains = append(s.Domains, d)
		}
		if err := rows.Err(); err != nil {
			log.Printf("loadServiceRelations: domains iteration error: %v", err)
		}
	}

	if rows, err := h.db.Query(`SELECT srv.name FROM service_dependencies d JOIN services srv ON d.depends_on_service_id = srv.id WHERE d.service_id = $1`, s.ID); err != nil {
		log.Printf("loadServiceRelations: dependencies query error for service %d: %v", s.ID, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				log.Printf("loadServiceRelations: dependencies scan error: %v", err)
				continue
			}
			s.Dependencies = append(s.Dependencies, name)
		}
		if err := rows.Err(); err != nil {
			log.Printf("loadServiceRelations: dependencies iteration error: %v", err)
		}
	}

	var hc models.ServiceHealthcheck
	err := h.db.QueryRow(`SELECT test, interval_seconds, timeout_seconds, retries, start_period_seconds FROM service_healthchecks WHERE service_id = $1`, s.ID).
		Scan(&hc.Test, &hc.IntervalSeconds, &hc.TimeoutSeconds, &hc.Retries, &hc.StartPeriodSeconds)
	if err == nil {
		s.Healthcheck = &hc
	}

	if rows, err := h.db.Query(`SELECT path FROM service_env_files WHERE service_id = $1 ORDER BY sort_order`, s.ID); err != nil {
		log.Printf("loadServiceRelations: env_files query error for service %d: %v", s.ID, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var path string
			if err := rows.Scan(&path); err != nil {
				log.Printf("loadServiceRelations: env_files scan error: %v", err)
				continue
			}
			s.EnvFiles = append(s.EnvFiles, path)
		}
		if err := rows.Err(); err != nil {
			log.Printf("loadServiceRelations: env_files iteration error: %v", err)
		}
	}

	if rows, err := h.db.Query(`SELECT network_name FROM service_networks WHERE service_id = $1 ORDER BY network_name`, s.ID); err != nil {
		log.Printf("loadServiceRelations: networks query error for service %d: %v", s.ID, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				log.Printf("loadServiceRelations: networks scan error: %v", err)
				continue
			}
			s.Networks = append(s.Networks, name)
		}
		if err := rows.Err(); err != nil {
			log.Printf("loadServiceRelations: networks iteration error: %v", err)
		}
	}

	if rows, err := h.db.Query(`SELECT id, service_id, config_file_id, source_path, target_path, readonly FROM service_config_mounts WHERE service_id = $1`, s.ID); err != nil {
		log.Printf("loadServiceRelations: config_mounts query error for service %d: %v", s.ID, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var m models.ServiceConfigMount
			var cfgFileID sql.NullInt32
			if err := rows.Scan(&m.ID, &m.ServiceID, &cfgFileID, &m.SourcePath, &m.TargetPath, &m.ReadOnly); err != nil {
				log.Printf("loadServiceRelations: config_mounts scan error: %v", err)
				continue
			}
			if cfgFileID.Valid {
				val := int(cfgFileID.Int32)
				m.ConfigFileID = &val
			}
			s.ConfigMounts = append(s.ConfigMounts, m)
		}
		if err := rows.Err(); err != nil {
			log.Printf("loadServiceRelations: config_mounts iteration error: %v", err)
		}
	}
}

type createServiceRequest struct {
	Name          string                     `json:"name"`
	CategoryID    int                        `json:"category_id"`
	ImageName     string                     `json:"image_name"`
	ImageTag      string                     `json:"image_tag"`
	RestartPolicy string                     `json:"restart_policy"`
	Command       string                     `json:"command"`
	UserSpec      string                     `json:"user_spec"`
	Ports         []models.ServicePort       `json:"ports"`
	Volumes       []models.ServiceVolume     `json:"volumes"`
	EnvVars       []models.ServiceEnvVar     `json:"env_vars"`
	Dependencies  []string                   `json:"dependencies"`
	Labels        []models.ServiceLabel      `json:"labels"`
	Domains       []models.ServiceDomain     `json:"domains"`
	Healthcheck   *models.ServiceHealthcheck `json:"healthcheck"`
}

func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.CategoryID == 0 {
		http.Error(w, "name and category_id are required", http.StatusBadRequest)
		return
	}
	if !isValidServiceName(req.Name) {
		http.Error(w, "name must be lowercase alphanumeric with hyphens only", http.StatusBadRequest)
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
			INSERT INTO service_volumes (service_id, volume_type, source, target, read_only, is_external)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, serviceID, vol.VolumeType, vol.Source, vol.Target, vol.ReadOnly, vol.IsExternal)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, env := range req.EnvVars {
		value := env.Value
		var encryptedValue sql.NullString
		encryptionVersion := 0

		if env.IsSecret && value != "" {
			encrypted, err := h.cipher.Encrypt(value)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to encrypt secret: %v", err), http.StatusInternalServerError)
				return
			}
			encryptedValue = sql.NullString{String: encrypted, Valid: true}
			encryptionVersion = 1
			value = ""
		}

		_, err = tx.Exec(`
			INSERT INTO service_env_vars (service_id, key, value, is_secret, encrypted_value, encryption_version)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, serviceID, env.Key, value, env.IsSecret, encryptedValue, encryptionVersion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, depName := range req.Dependencies {
		var depID int
		err := tx.QueryRow("SELECT id FROM services WHERE name = $1", depName).Scan(&depID)
		if err != nil {
			http.Error(w, "dependency not found: "+depName, http.StatusBadRequest)
			return
		}
		_, err = tx.Exec(`
			INSERT INTO service_dependencies (service_id, depends_on_service_id)
			VALUES ($1, $2)
		`, serviceID, depID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, label := range req.Labels {
		_, err = tx.Exec(`
			INSERT INTO service_labels (service_id, key, value)
			VALUES ($1, $2, $3)
		`, serviceID, label.Key, label.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, domain := range req.Domains {
		_, err = tx.Exec(`
			INSERT INTO service_domains (service_id, domain, proxy_port)
			VALUES ($1, $2, $3)
		`, serviceID, domain.Domain, domain.ProxyPort)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if req.Healthcheck != nil && req.Healthcheck.Test != "" {
		_, err = tx.Exec(`
			INSERT INTO service_healthchecks (service_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, serviceID, req.Healthcheck.Test, req.Healthcheck.IntervalSeconds, req.Healthcheck.TimeoutSeconds, req.Healthcheck.Retries, req.Healthcheck.StartPeriodSeconds)
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

	var s models.Service
	var catName string
	err := h.db.QueryRow(`
		SELECT s.id, s.name, s.category_id, s.image_name, s.image_tag,
			s.restart_policy, s.command, s.user_spec, s.enabled,
			s.created_at, s.updated_at, c.name, s.compose_overrides
		FROM services s JOIN categories c ON s.category_id = c.id
		WHERE s.name = $1
	`, name).Scan(
		&s.ID, &s.Name, &s.CategoryID, &s.ImageName, &s.ImageTag,
		&s.RestartPolicy, &s.Command, &s.UserSpec, &s.Enabled,
		&s.CreatedAt, &s.UpdatedAt, &catName, &s.ComposeOverrides,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Snapshot current config before updating
	h.loadServiceRelations(&s)
	h.snapshotConfig(&s, "pre-update")

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
			config_status = 'modified',
			updated_at = NOW()
		WHERE id = $6
	`, req.ImageName, req.ImageTag, req.RestartPolicy, req.Command, req.UserSpec, s.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) snapshotConfig(s *models.Service, summary string) {
	snapshot := map[string]interface{}{
		"image_name":     s.ImageName,
		"image_tag":      s.ImageTag,
		"restart_policy": s.RestartPolicy,
		"ports":          s.Ports,
		"volumes":        s.Volumes,
		"env_vars":       s.EnvVars,
		"dependencies":   s.Dependencies,
		"healthcheck":    s.Healthcheck,
		"labels":         s.Labels,
	}
	if s.Command.Valid {
		snapshot["command"] = s.Command.String
	}
	if s.UserSpec.Valid {
		snapshot["user_spec"] = s.UserSpec.String
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return
	}

	var nextVersion int
	if err := h.db.QueryRow(`SELECT COALESCE(MAX(version), 0) + 1 FROM service_config_versions WHERE service_id = $1`, s.ID).Scan(&nextVersion); err != nil {
		log.Printf("snapshotConfig: failed to get next version for service %d: %v", s.ID, err)
		return
	}
	if _, err := h.db.Exec(`
		INSERT INTO service_config_versions (service_id, version, config_snapshot, change_summary)
		VALUES ($1, $2, $3, $4)
	`, s.ID, nextVersion, snapshotJSON, summary); err != nil {
		log.Printf("snapshotConfig: failed to insert version %d for service %d: %v", nextVersion, s.ID, err)
	}
}

func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	result, err := h.db.Exec("DELETE FROM services WHERE name = $1", name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if affected == 0 {
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

	s, err := h.loadServiceByName(name)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Materialize config files before starting
	if h.projectRoot != "" {
		if err := h.generator.MaterializeConfigFiles(s, h.projectRoot); err != nil {
			http.Error(w, "materialize config: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	composeYAML, err := h.generator.Generate(s)
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

	s, err := h.loadServiceByName(name)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	composeYAML, err := h.generator.Generate(s)
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

	s, err := h.loadServiceByName(name)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	composeYAML, err := h.generator.Generate(s)
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

	s, err := h.loadServiceByName(name)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Materialize config files before rebuilding
	if h.projectRoot != "" {
		if err := h.generator.MaterializeConfigFiles(s, h.projectRoot); err != nil {
			http.Error(w, "materialize config: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	composeYAML, err := h.generator.Generate(s)
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

	s, err := h.loadServiceByName(name)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	composeYAML, err := h.generator.Generate(s)
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
		s, err := h.loadServiceByName(name)
		if err != nil {
			results[name] = "not found"
			continue
		}

		composeYAML, err := h.generator.Generate(s)
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

func (h *ServiceHandler) Versions(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var serviceID int
	if err := h.db.QueryRow("SELECT id FROM services WHERE name = $1", name).Scan(&serviceID); err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, service_id, version, config_snapshot, COALESCE(change_summary, ''), created_at
		FROM service_config_versions
		WHERE service_id = $1
		ORDER BY version DESC
	`, serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var versions []models.ConfigVersion
	for rows.Next() {
		var v models.ConfigVersion
		if err := rows.Scan(&v.ID, &v.ServiceID, &v.Version, &v.ConfigSnapshot, &v.ChangeSummary, &v.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		versions = append(versions, v)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

func (h *ServiceHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	versionStr := chi.URLParam(r, "version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		http.Error(w, "invalid version", http.StatusBadRequest)
		return
	}

	var serviceID int
	if err := h.db.QueryRow("SELECT id FROM services WHERE name = $1", name).Scan(&serviceID); err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var v models.ConfigVersion
	err = h.db.QueryRow(`
		SELECT id, service_id, version, config_snapshot, COALESCE(change_summary, ''), created_at
		FROM service_config_versions
		WHERE service_id = $1 AND version = $2
	`, serviceID, version).Scan(&v.ID, &v.ServiceID, &v.Version, &v.ConfigSnapshot, &v.ChangeSummary, &v.CreatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *ServiceHandler) Validate(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	s, err := h.loadServiceByName(name)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.loadServiceRelations(s)
	result := h.validator.Validate(s)

	// Update validation status in DB
	status := "validated"
	if !result.Valid {
		status = "broken"
	}
	errorsJSON, _ := json.Marshal(result.Errors)
	h.db.Exec(`
		UPDATE services SET config_status = $1, last_validated_at = NOW(), validation_errors = $2
		WHERE id = $3
	`, status, errorsJSON, s.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *ServiceHandler) Export(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	s, err := h.loadServiceByName(name)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var catName string
	h.db.QueryRow(`SELECT c.name FROM categories c JOIN services s ON s.category_id = c.id WHERE s.id = $1`, s.ID).Scan(&catName)

	composeYAML, err := h.generator.Generate(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":  name,
		"category": catName,
		"yaml":     string(composeYAML),
	})
}

// Materialize writes config files to disk without starting the service
func (h *ServiceHandler) Materialize(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if h.projectRoot == "" {
		http.Error(w, "PROJECT_ROOT not set", http.StatusInternalServerError)
		return
	}

	s, err := h.loadServiceByName(name)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.generator.MaterializeConfigFiles(s, h.projectRoot); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "materialized"})
}

// ListConfigFiles returns config files for a service
func (h *ServiceHandler) ListConfigFiles(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var serviceID int
	if err := h.db.QueryRow("SELECT id FROM services WHERE name = $1", name).Scan(&serviceID); err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, service_id, file_path, file_mode, is_template, created_at, updated_at
		FROM service_config_files WHERE service_id = $1 ORDER BY file_path
	`, serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []models.ServiceConfigFile
	for rows.Next() {
		var f models.ServiceConfigFile
		if err := rows.Scan(&f.ID, &f.ServiceID, &f.FilePath, &f.FileMode, &f.IsTemplate, &f.CreatedAt, &f.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		files = append(files, f)
	}

	if files == nil {
		files = []models.ServiceConfigFile{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// GetConfigFile returns a single config file's content
func (h *ServiceHandler) GetConfigFile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	filePath := chi.URLParam(r, "*")

	var serviceID int
	if err := h.db.QueryRow("SELECT id FROM services WHERE name = $1", name).Scan(&serviceID); err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var f models.ServiceConfigFile
	err := h.db.QueryRow(`
		SELECT id, service_id, file_path, content, file_mode, is_template, created_at, updated_at
		FROM service_config_files WHERE service_id = $1 AND file_path = $2
	`, serviceID, filePath).Scan(&f.ID, &f.ServiceID, &f.FilePath, &f.Content, &f.FileMode, &f.IsTemplate, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

// PutConfigFile creates or updates a config file
func (h *ServiceHandler) PutConfigFile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	filePath := chi.URLParam(r, "*")

	var serviceID int
	if err := h.db.QueryRow("SELECT id FROM services WHERE name = $1", name).Scan(&serviceID); err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req struct {
		Content  string `json:"content"`
		FileMode string `json:"file_mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Fall back to reading raw body as content
		r.Body = io.NopCloser(strings.NewReader(""))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.FileMode == "" {
		req.FileMode = "0644"
	}

	_, err := h.db.Exec(`
		INSERT INTO service_config_files (service_id, file_path, content, file_mode)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (service_id, file_path) DO UPDATE SET
			content = $3, file_mode = $4, updated_at = NOW()
	`, serviceID, filePath, req.Content, req.FileMode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (h *ServiceHandler) lookupServiceID(w http.ResponseWriter, r *http.Request) (int, bool) {
	name := chi.URLParam(r, "name")
	var id int
	err := h.db.QueryRow("SELECT id FROM services WHERE name = $1", name).Scan(&id)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return 0, false
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 0, false
	}
	return id, true
}

func (h *ServiceHandler) snapshotAndUpdate(serviceID int) {
	var s models.Service
	s.ID = serviceID
	h.db.QueryRow(
		`SELECT name, image_name, image_tag, restart_policy, command, user_spec, compose_overrides FROM services WHERE id = $1`,
		serviceID,
	).Scan(&s.Name, &s.ImageName, &s.ImageTag, &s.RestartPolicy, &s.Command, &s.UserSpec, &s.ComposeOverrides)
	h.loadServiceRelations(&s)
	h.snapshotConfig(&s, "sub-resource update")
	h.db.Exec(`UPDATE services SET config_status = 'modified', updated_at = NOW() WHERE id = $1`, serviceID)
}

func (h *ServiceHandler) UpdatePorts(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	var req struct {
		Ports []models.ServicePort `json:"ports"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.snapshotAndUpdate(serviceID)

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM service_ports WHERE service_id = $1", serviceID)
	for _, p := range req.Ports {
		_, err = tx.Exec(
			`INSERT INTO service_ports (service_id, host_ip, host_port, container_port, protocol) VALUES ($1, $2, $3, $4, $5)`,
			serviceID, p.HostIP, p.HostPort, p.ContainerPort, p.Protocol,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) UpdateVolumes(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	var req struct {
		Volumes []models.ServiceVolume `json:"volumes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.snapshotAndUpdate(serviceID)

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM service_volumes WHERE service_id = $1", serviceID)
	for _, v := range req.Volumes {
		_, err = tx.Exec(
			`INSERT INTO service_volumes (service_id, volume_type, source, target, read_only, is_external) VALUES ($1, $2, $3, $4, $5, $6)`,
			serviceID, v.VolumeType, v.Source, v.Target, v.ReadOnly, v.IsExternal,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) UpdateEnvVars(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	var req struct {
		EnvVars []models.ServiceEnvVar `json:"env_vars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.snapshotAndUpdate(serviceID)

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	type existingSecret struct {
		encryptedValue    sql.NullString
		encryptionVersion int
	}
	existingSecrets := map[string]existingSecret{}
	rows, err := tx.Query("SELECT key, encrypted_value, encryption_version FROM service_env_vars WHERE service_id = $1 AND is_secret = true", serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var key string
		var secret existingSecret
		if err := rows.Scan(&key, &secret.encryptedValue, &secret.encryptionVersion); err != nil {
			rows.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		existingSecrets[key] = secret
	}
	rows.Close()

	tx.Exec("DELETE FROM service_env_vars WHERE service_id = $1", serviceID)
	for _, e := range req.EnvVars {
		value := e.Value
		var encryptedValue sql.NullString
		encryptionVersion := 0

		if e.IsSecret {
			if e.Value == "***" || e.Value == "" {
				if current, ok := existingSecrets[e.Key]; ok {
					encryptedValue = current.encryptedValue
					encryptionVersion = current.encryptionVersion
					value = ""
				}
			} else {
				encrypted, err := h.cipher.Encrypt(e.Value)
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to encrypt secret: %v", err), http.StatusInternalServerError)
					return
				}
				encryptedValue = sql.NullString{String: encrypted, Valid: true}
				encryptionVersion = 1
				value = ""
			}
		}

		_, err = tx.Exec(
			`INSERT INTO service_env_vars (service_id, key, value, is_secret, encrypted_value, encryption_version) VALUES ($1, $2, $3, $4, $5, $6)`,
			serviceID, e.Key, value, e.IsSecret, encryptedValue, encryptionVersion,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) UpdateDependencies(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	var req struct {
		Dependencies []string `json:"dependencies"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.snapshotAndUpdate(serviceID)

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM service_dependencies WHERE service_id = $1", serviceID)
	for _, depName := range req.Dependencies {
		var depID int
		err := tx.QueryRow("SELECT id FROM services WHERE name = $1", depName).Scan(&depID)
		if err != nil {
			http.Error(w, "dependency not found: "+depName, http.StatusBadRequest)
			return
		}
		_, err = tx.Exec(
			`INSERT INTO service_dependencies (service_id, depends_on_service_id) VALUES ($1, $2)`,
			serviceID, depID,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) UpdateHealthcheck(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	var req *models.ServiceHealthcheck
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.snapshotAndUpdate(serviceID)

	h.db.Exec("DELETE FROM service_healthchecks WHERE service_id = $1", serviceID)

	if req != nil && req.Test != "" {
		_, err := h.db.Exec(
			`INSERT INTO service_healthchecks (service_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds) VALUES ($1, $2, $3, $4, $5, $6)`,
			serviceID, req.Test, req.IntervalSeconds, req.TimeoutSeconds, req.Retries, req.StartPeriodSeconds,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) UpdateLabels(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	var req struct {
		Labels []models.ServiceLabel `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.snapshotAndUpdate(serviceID)

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM service_labels WHERE service_id = $1", serviceID)
	for _, l := range req.Labels {
		_, err = tx.Exec(
			`INSERT INTO service_labels (service_id, key, value) VALUES ($1, $2, $3)`,
			serviceID, l.Key, l.Value,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) UpdateDomains(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	var req struct {
		Domains []models.ServiceDomain `json:"domains"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.snapshotAndUpdate(serviceID)

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM service_domains WHERE service_id = $1", serviceID)
	for _, d := range req.Domains {
		_, err = tx.Exec(
			`INSERT INTO service_domains (service_id, domain, proxy_port) VALUES ($1, $2, $3)`,
			serviceID, d.Domain, d.ProxyPort,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) UpdateEnvFiles(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	var req struct {
		EnvFiles []string `json:"env_files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, path := range req.EnvFiles {
		if strings.TrimSpace(path) == "" {
			http.Error(w, "env_file path cannot be empty", http.StatusBadRequest)
			return
		}
		if strings.Contains(path, "..") || strings.HasPrefix(path, "/") {
			http.Error(w, "env_file path must be relative and cannot contain '..'", http.StatusBadRequest)
			return
		}
	}

	h.snapshotAndUpdate(serviceID)

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM service_env_files WHERE service_id = $1", serviceID)
	for idx, path := range req.EnvFiles {
		_, err = tx.Exec(
			`INSERT INTO service_env_files (service_id, path, sort_order) VALUES ($1, $2, $3)`,
			serviceID, path, idx,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) UpdateNetworks(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	var req struct {
		Networks []string `json:"networks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, name := range req.Networks {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			http.Error(w, "network name cannot be empty", http.StatusBadRequest)
			return
		}
		if strings.ContainsAny(trimmed, " \t/\\:") {
			http.Error(w, fmt.Sprintf("invalid network name %q: contains disallowed characters", trimmed), http.StatusBadRequest)
			return
		}
	}

	h.snapshotAndUpdate(serviceID)

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM service_networks WHERE service_id = $1", serviceID)
	for _, name := range req.Networks {
		_, err = tx.Exec(
			`INSERT INTO service_networks (service_id, network_name) VALUES ($1, $2)`,
			serviceID, name,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ServiceHandler) UpdateConfigMounts(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	var req struct {
		ConfigMounts []models.ServiceConfigMount `json:"config_mounts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, m := range req.ConfigMounts {
		if strings.TrimSpace(m.TargetPath) == "" {
			http.Error(w, "config mount target_path cannot be empty", http.StatusBadRequest)
			return
		}
		if strings.Contains(m.SourcePath, "..") {
			http.Error(w, "config mount source_path cannot contain '..'", http.StatusBadRequest)
			return
		}
	}

	h.snapshotAndUpdate(serviceID)

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM service_config_mounts WHERE service_id = $1", serviceID)
	for _, m := range req.ConfigMounts {
		var cfgFileID sql.NullInt32
		if m.ConfigFileID != nil {
			cfgFileID = sql.NullInt32{Int32: int32(*m.ConfigFileID), Valid: true}
		}
		_, err = tx.Exec(
			`INSERT INTO service_config_mounts (service_id, config_file_id, source_path, target_path, readonly) VALUES ($1, $2, $3, $4, $5)`,
			serviceID, cfgFileID, m.SourcePath, m.TargetPath, m.ReadOnly,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// DeleteConfigFile removes a config file
func (h *ServiceHandler) DeleteConfigFile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	filePath := chi.URLParam(r, "*")

	var serviceID int
	if err := h.db.QueryRow("SELECT id FROM services WHERE name = $1", name).Scan(&serviceID); err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := h.db.Exec("DELETE FROM service_config_files WHERE service_id = $1 AND file_path = $2", serviceID, filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if affected == 0 {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// ListExports returns export contracts for a service template
func (h *ServiceHandler) ListExports(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	rows, err := h.db.Query(`
		SELECT id, name, type, port, protocol
		FROM service_exports
		WHERE service_id = $1
		ORDER BY name
	`, serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Export struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
	}

	var exports []Export
	for rows.Next() {
		var e Export
		if err := rows.Scan(&e.ID, &e.Name, &e.Type, &e.Port, &e.Protocol); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		exports = append(exports, e)
	}

	if exports == nil {
		exports = []Export{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exports)
}

// UpdateExports replaces all export contracts for a service template
func (h *ServiceHandler) UpdateExports(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	type Export struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
	}

	var req struct {
		Exports []Export `json:"exports"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM service_exports WHERE service_id = $1", serviceID)
	for _, e := range req.Exports {
		protocol := e.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		_, err = tx.Exec(
			`INSERT INTO service_exports (service_id, name, type, port, protocol) VALUES ($1, $2, $3, $4, $5)`,
			serviceID, e.Name, e.Type, e.Port, protocol,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// ListImports returns import contracts for a service template
func (h *ServiceHandler) ListImports(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	rows, err := h.db.Query(`
		SELECT id, name, type, required, COALESCE(env_vars, '{}')
		FROM service_import_contracts
		WHERE service_id = $1
		ORDER BY name
	`, serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Import struct {
		ID       int               `json:"id"`
		Name     string            `json:"name"`
		Type     string            `json:"type"`
		Required bool              `json:"required"`
		EnvVars  map[string]string `json:"env_vars"`
	}

	var imports []Import
	for rows.Next() {
		var i Import
		var envVarsJSON []byte
		if err := rows.Scan(&i.ID, &i.Name, &i.Type, &i.Required, &envVarsJSON); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(envVarsJSON, &i.EnvVars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		imports = append(imports, i)
	}

	if imports == nil {
		imports = []Import{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(imports)
}

// UpdateImports replaces all import contracts for a service template
func (h *ServiceHandler) UpdateImports(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := h.lookupServiceID(w, r)
	if !ok {
		return
	}

	type Import struct {
		Name     string            `json:"name"`
		Type     string            `json:"type"`
		Required bool              `json:"required"`
		EnvVars  map[string]string `json:"env_vars"`
	}

	var req struct {
		Imports []Import `json:"imports"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM service_import_contracts WHERE service_id = $1", serviceID)
	for _, i := range req.Imports {
		envVarsJSON, err := json.Marshal(i.EnvVars)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec(
			`INSERT INTO service_import_contracts (service_id, name, type, required, env_vars) VALUES ($1, $2, $3, $4, $5)`,
			serviceID, i.Name, i.Type, i.Required, envVarsJSON,
		)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}
