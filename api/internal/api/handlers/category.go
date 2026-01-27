package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/compose"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/pkg/models"
)

type CategoryHandler struct {
	db              *sql.DB
	containerClient *container.Client
	generator       *compose.Generator
}

func NewCategoryHandler(db *sql.DB, cc *container.Client) *CategoryHandler {
	return &CategoryHandler{
		db:              db,
		containerClient: cc,
		generator:       compose.NewGenerator(db, "microservices-net"),
	}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT c.id, c.name, c.display_name, c.color, c.startup_order,
			c.created_at, c.updated_at,
			COUNT(s.id) as service_count
		FROM categories c
		LEFT JOIN services s ON s.category_id = c.id
		GROUP BY c.id
		ORDER BY c.startup_order
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type categoryWithCount struct {
		models.Category
		ServiceCount int `json:"service_count"`
	}

	var categories []categoryWithCount
	for rows.Next() {
		var c categoryWithCount
		var displayName, color sql.NullString
		if err := rows.Scan(
			&c.ID, &c.Name, &displayName, &color, &c.StartupOrder,
			&c.CreatedAt, &c.UpdatedAt, &c.ServiceCount,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if displayName.Valid {
			c.DisplayName = displayName.String
		}
		if color.Valid {
			c.Color = color.String
		}
		categories = append(categories, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var c models.Category
	var displayName, color sql.NullString
	err := h.db.QueryRow(`
		SELECT id, name, display_name, color, startup_order, created_at, updated_at
		FROM categories WHERE name = $1
	`, name).Scan(&c.ID, &c.Name, &displayName, &color, &c.StartupOrder, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if displayName.Valid {
		c.DisplayName = displayName.String
	}
	if color.Valid {
		c.Color = color.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *CategoryHandler) Services(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var categoryID int
	err := h.db.QueryRow("SELECT id FROM categories WHERE name = $1", name).Scan(&categoryID)
	if err == sql.ErrNoRows {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, name, image_name, image_tag, restart_policy, enabled, created_at, updated_at
		FROM services WHERE category_id = $1 ORDER BY name
	`, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		services = append(services, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (h *CategoryHandler) Start(w http.ResponseWriter, r *http.Request) {
	if h.containerClient == nil {
		http.Error(w, "container client not initialized", http.StatusInternalServerError)
		return
	}

	name := chi.URLParam(r, "name")

	var categoryID int
	err := h.db.QueryRow("SELECT id FROM categories WHERE name = $1", name).Scan(&categoryID)
	if err == sql.ErrNoRows {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, name, image_name, image_tag, restart_policy, command, user_spec
		FROM services WHERE category_id = $1 AND enabled = true ORDER BY name
	`, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *CategoryHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if h.containerClient == nil {
		http.Error(w, "container client not initialized", http.StatusInternalServerError)
		return
	}

	name := chi.URLParam(r, "name")

	var categoryID int
	err := h.db.QueryRow("SELECT id FROM categories WHERE name = $1", name).Scan(&categoryID)
	if err == sql.ErrNoRows {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, name, image_name, image_tag
		FROM services WHERE category_id = $1 ORDER BY name
	`, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
