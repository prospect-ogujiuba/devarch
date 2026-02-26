package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/compose"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/crypto"
	"github.com/priz/devarch-api/internal/podman"
	"github.com/priz/devarch-api/pkg/models"
)

type CategoryHandler struct {
	db              *sql.DB
	containerClient *container.Client
	podmanClient    *podman.Client
	generator       *compose.Generator
	cipher          *crypto.Cipher
}

func NewCategoryHandler(db *sql.DB, cc *container.Client, pc *podman.Client, cipher *crypto.Cipher) *CategoryHandler {
	return &CategoryHandler{
		db:              db,
		containerClient: cc,
		podmanClient:    pc,
		generator:       compose.NewGenerator(db, "microservices-net", cipher),
		cipher:          cipher,
	}
}

// List godoc
// @Summary      List all categories
// @Description  Returns all categories with service counts and running container counts
// @Tags         categories
// @Produce      json
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /categories [get]
// @Security     ApiKeyAuth
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	runningContainers, _ := h.podmanClient.ListContainers(ctx, false)
	runningByName := make(map[string]bool)
	for _, c := range runningContainers {
		if len(c.Names) > 0 {
			runningByName[c.Names[0]] = true
		}
	}

	rows, err := h.db.Query(`
		SELECT c.id, c.name, c.display_name, c.color, c.startup_order,
			c.created_at, c.updated_at,
			COUNT(s.id) as service_count,
			ARRAY_AGG(s.name) FILTER (WHERE s.name IS NOT NULL)
		FROM categories c
		LEFT JOIN services s ON s.category_id = c.id
		GROUP BY c.id
		ORDER BY c.startup_order
	`)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	defer rows.Close()

	type categoryWithCount struct {
		models.Category
		ServiceCount int `json:"service_count"`
		RunningCount int `json:"runningCount"`
	}

	var categories []categoryWithCount
	for rows.Next() {
		var c categoryWithCount
		var displayName, color sql.NullString
		var serviceNames []string
		if err := rows.Scan(
			&c.ID, &c.Name, &displayName, &color, &c.StartupOrder,
			&c.CreatedAt, &c.UpdatedAt, &c.ServiceCount, pq.Array(&serviceNames),
		); err != nil {
			respond.InternalError(w, r, err)
			return
		}
		if displayName.Valid {
			c.DisplayName = displayName.String
		}
		if color.Valid {
			c.Color = color.String
		}
		for _, name := range serviceNames {
			if runningByName[name] {
				c.RunningCount++
			}
		}
		categories = append(categories, c)
	}

	respond.JSON(w, r, http.StatusOK, categories)
}

// Get godoc
// @Summary      Get category by name
// @Description  Returns a single category by its name
// @Tags         categories
// @Produce      json
// @Param        name path string true "Category name"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /categories/{name} [get]
// @Security     ApiKeyAuth
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var c models.Category
	var displayName, color sql.NullString
	err := h.db.QueryRow(`
		SELECT id, name, display_name, color, startup_order, created_at, updated_at
		FROM categories WHERE name = $1
	`, name).Scan(&c.ID, &c.Name, &displayName, &color, &c.StartupOrder, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "category", name)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if displayName.Valid {
		c.DisplayName = displayName.String
	}
	if color.Valid {
		c.Color = color.String
	}

	respond.JSON(w, r, http.StatusOK, c)
}

// Services godoc
// @Summary      List services in category
// @Description  Returns all services that belong to the specified category
// @Tags         categories
// @Produce      json
// @Param        name path string true "Category name"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /categories/{name}/services [get]
// @Security     ApiKeyAuth
func (h *CategoryHandler) Services(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var categoryID int
	err := h.db.QueryRow("SELECT id FROM categories WHERE name = $1", name).Scan(&categoryID)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "category", name)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, name, image_name, image_tag, restart_policy, enabled, created_at, updated_at
		FROM services WHERE category_id = $1 ORDER BY name
	`, categoryID)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(
			&s.ID, &s.Name, &s.ImageName, &s.ImageTag,
			&s.RestartPolicy, &s.Enabled, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			respond.InternalError(w, r, err)
			return
		}
		services = append(services, s)
	}

	respond.JSON(w, r, http.StatusOK,services)
}

// Start godoc
// @Summary      Start all services in category
// @Description  Starts all enabled services that belong to the specified category
// @Tags         categories
// @Produce      json
// @Param        name path string true "Category name"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /categories/{name}/start [post]
// @Security     ApiKeyAuth
func (h *CategoryHandler) Start(w http.ResponseWriter, r *http.Request) {
	if h.containerClient == nil {
		respond.InternalError(w, r, fmt.Errorf("container client not initialized"))
		return
	}

	name := chi.URLParam(r, "name")

	var categoryID int
	err := h.db.QueryRow("SELECT id FROM categories WHERE name = $1", name).Scan(&categoryID)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "category", name)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, name, image_name, image_tag, restart_policy, command, user_spec
		FROM services WHERE category_id = $1 AND enabled = true ORDER BY name
	`, categoryID)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	defer rows.Close()

	results := make(map[string]string)
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.ImageName, &s.ImageTag, &s.RestartPolicy, &s.Command, &s.UserSpec); err != nil {
			continue
		}

		composeYAML, err := h.generator.Generate(&s)
		if err != nil {
			results[s.Name] = "compose error: " + err.Error()
			continue
		}

		if err := h.containerClient.StartService(s.Name, composeYAML); err != nil {
			results[s.Name] = "error: " + err.Error()
		} else {
			results[s.Name] = "started"
		}
	}

	respond.Action(w, r, http.StatusOK, "completed",
		respond.WithMetadata("services", results),
	)
}

func (h *CategoryHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if h.containerClient == nil {
		respond.InternalError(w, r, fmt.Errorf("container client not initialized"))
		return
	}

	name := chi.URLParam(r, "name")

	var categoryID int
	err := h.db.QueryRow("SELECT id FROM categories WHERE name = $1", name).Scan(&categoryID)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "category", name)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, name, image_name, image_tag
		FROM services WHERE category_id = $1 ORDER BY name
	`, categoryID)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	defer rows.Close()

	results := make(map[string]string)
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.ImageName, &s.ImageTag); err != nil {
			continue
		}

		composeYAML, err := h.generator.Generate(&s)
		if err != nil {
			results[s.Name] = "compose error: " + err.Error()
			continue
		}

		if err := h.containerClient.StopService(s.Name, composeYAML); err != nil {
			results[s.Name] = "error: " + err.Error()
		} else {
			results[s.Name] = "stopped"
		}
	}

	respond.Action(w, r, http.StatusOK, "completed",
		respond.WithMetadata("services", results),
	)
}

var validCategoryName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

type createCategoryRequest struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Color        string `json:"color"`
	StartupOrder *int   `json:"startup_order"`
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid JSON body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.BadRequest(w, r, "name is required")
		return
	}
	if len(req.Name) > 64 || !validCategoryName.MatchString(req.Name) {
		respond.BadRequest(w, r, "name must be lowercase alphanumeric with hyphens, max 64 characters")
		return
	}

	startupOrder := 0
	if req.StartupOrder != nil {
		startupOrder = *req.StartupOrder
	}

	var c models.Category
	var displayName, color sql.NullString
	err := h.db.QueryRow(`
		INSERT INTO categories (name, display_name, color, startup_order)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), $4)
		RETURNING id, name, display_name, color, startup_order, created_at, updated_at
	`, req.Name, req.DisplayName, req.Color, startupOrder).Scan(
		&c.ID, &c.Name, &displayName, &color, &c.StartupOrder, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			respond.Conflict(w, r, fmt.Sprintf("category '%s' already exists", req.Name))
			return
		}
		respond.InternalError(w, r, err)
		return
	}

	if displayName.Valid {
		c.DisplayName = displayName.String
	}
	if color.Valid {
		c.Color = color.String
	}

	respond.JSON(w, r, http.StatusCreated, c)
}

type updateCategoryRequest struct {
	Name         *string `json:"name"`
	DisplayName  *string `json:"display_name"`
	Color        *string `json:"color"`
	StartupOrder *int    `json:"startup_order"`
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var req updateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid JSON body")
		return
	}

	if req.Name != nil {
		*req.Name = strings.TrimSpace(*req.Name)
		if len(*req.Name) > 64 || !validCategoryName.MatchString(*req.Name) {
			respond.BadRequest(w, r, "name must be lowercase alphanumeric with hyphens, max 64 characters")
			return
		}
	}

	var c models.Category
	var displayName, color sql.NullString
	err := h.db.QueryRow(`
		UPDATE categories SET
			name = CASE WHEN $2 THEN $3 ELSE name END,
			display_name = CASE WHEN $4 THEN NULLIF($5, '') ELSE display_name END,
			color = CASE WHEN $6 THEN NULLIF($7, '') ELSE color END,
			startup_order = CASE WHEN $8 THEN $9 ELSE startup_order END,
			updated_at = NOW()
		WHERE name = $1
		RETURNING id, name, display_name, color, startup_order, created_at, updated_at
	`,
		name,
		req.Name != nil, sqlNullString(req.Name),
		req.DisplayName != nil, sqlNullString(req.DisplayName),
		req.Color != nil, sqlNullString(req.Color),
		req.StartupOrder != nil, sqlNullInt(req.StartupOrder),
	).Scan(&c.ID, &c.Name, &displayName, &color, &c.StartupOrder, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "category", name)
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			respond.Conflict(w, r, fmt.Sprintf("category '%s' already exists", *req.Name))
			return
		}
		respond.InternalError(w, r, err)
		return
	}

	if displayName.Valid {
		c.DisplayName = displayName.String
	}
	if color.Valid {
		c.Color = color.String
	}

	respond.JSON(w, r, http.StatusOK, c)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var serviceCount int
	err := h.db.QueryRow(`
		SELECT COUNT(*) FROM services s
		JOIN categories c ON s.category_id = c.id
		WHERE c.name = $1
	`, name).Scan(&serviceCount)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	if serviceCount > 0 {
		respond.Conflict(w, r, fmt.Sprintf("cannot delete category '%s': has %d service(s)", name, serviceCount))
		return
	}

	result, err := h.db.Exec("DELETE FROM categories WHERE name = $1", name)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	if affected == 0 {
		respond.NotFound(w, r, "category", name)
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"status": "deleted"})
}

func sqlNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func sqlNullInt(i *int) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(*i), Valid: true}
}
