package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/priz/devarch-api/internal/container"
)

type StackHandler struct {
	db              *sql.DB
	containerClient *container.Client
}

func NewStackHandler(db *sql.DB, cc *container.Client) *StackHandler {
	return &StackHandler{
		db:              db,
		containerClient: cc,
	}
}

type stackResponse struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	NetworkName   *string   `json:"network_name"`
	Enabled       bool      `json:"enabled"`
	InstanceCount int       `json:"instance_count"`
	RunningCount  int       `json:"running_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type createStackRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateStackRequest struct {
	Description string `json:"description"`
}

// Create handles POST /stacks
func (h *StackHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createStackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate name using existing container validation
	if err := container.ValidateName(req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert stack
	var stack stackResponse
	err := h.db.QueryRow(`
		INSERT INTO stacks (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, network_name, enabled, created_at, updated_at
	`, req.Name, req.Description).Scan(
		&stack.ID,
		&stack.Name,
		&stack.Description,
		&stack.NetworkName,
		&stack.Enabled,
		&stack.CreatedAt,
		&stack.UpdatedAt,
	)

	if err != nil {
		// Handle unique constraint violation with prescriptive error
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			http.Error(w, fmt.Sprintf("stack name %q already exists", req.Name), http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("failed to create stack: %v", err), http.StatusInternalServerError)
		return
	}

	// New stack has no instances yet
	stack.InstanceCount = 0
	stack.RunningCount = 0

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(stack)
}

// List handles GET /stacks
func (h *StackHandler) List(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			s.id,
			s.name,
			s.description,
			s.network_name,
			s.enabled,
			s.created_at,
			s.updated_at,
			COUNT(si.id) AS instance_count
		FROM stacks s
		LEFT JOIN service_instances si ON si.stack_id = s.id
		WHERE s.deleted_at IS NULL
		GROUP BY s.id
		ORDER BY s.name ASC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query stacks: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	stacks := []stackResponse{}
	for rows.Next() {
		var stack stackResponse
		err := rows.Scan(
			&stack.ID,
			&stack.Name,
			&stack.Description,
			&stack.NetworkName,
			&stack.Enabled,
			&stack.CreatedAt,
			&stack.UpdatedAt,
			&stack.InstanceCount,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to scan stack: %v", err), http.StatusInternalServerError)
			return
		}

		// Running count placeholder (Phase 3+ will wire container client queries)
		stack.RunningCount = 0

		stacks = append(stacks, stack)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("error iterating stacks: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stacks)
}

// Get handles GET /stacks/{name}
func (h *StackHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var stack stackResponse
	err := h.db.QueryRow(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.network_name,
			s.enabled,
			s.created_at,
			s.updated_at,
			COUNT(si.id) AS instance_count
		FROM stacks s
		LEFT JOIN service_instances si ON si.stack_id = s.id
		WHERE s.name = $1 AND s.deleted_at IS NULL
		GROUP BY s.id
	`, name).Scan(
		&stack.ID,
		&stack.Name,
		&stack.Description,
		&stack.NetworkName,
		&stack.Enabled,
		&stack.CreatedAt,
		&stack.UpdatedAt,
		&stack.InstanceCount,
	)

	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
		return
	}

	// Running count placeholder (Phase 3+ will wire container client queries)
	stack.RunningCount = 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stack)
}

// Update handles PUT /stacks/{name}
func (h *StackHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var req updateStackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Update description only (name is immutable)
	var stack stackResponse
	err := h.db.QueryRow(`
		UPDATE stacks
		SET description = $1, updated_at = NOW()
		WHERE name = $2 AND deleted_at IS NULL
		RETURNING id, name, description, network_name, enabled, created_at, updated_at
	`, req.Description, name).Scan(
		&stack.ID,
		&stack.Name,
		&stack.Description,
		&stack.NetworkName,
		&stack.Enabled,
		&stack.CreatedAt,
		&stack.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update stack: %v", err), http.StatusInternalServerError)
		return
	}

	// Query instance count
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM service_instances WHERE stack_id = $1
	`, stack.ID).Scan(&stack.InstanceCount)
	if err != nil {
		stack.InstanceCount = 0
	}

	stack.RunningCount = 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stack)
}

// Delete handles DELETE /stacks/{name} (soft delete)
func (h *StackHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	// Get stack ID and instances before soft-deleting
	var stackID int
	var instanceIDs []int
	err := h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(&stackID)

	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
		return
	}

	// Get all instances for this stack to stop their containers
	rows, err := h.db.Query(`
		SELECT id FROM service_instances WHERE stack_id = $1
	`, stackID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query instances: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		instanceIDs = append(instanceIDs, id)
	}

	// Stop containers for all instances (Phase 3+ will implement actual container stopping)
	// For now, this is a placeholder since instance management is in Phase 3
	for range instanceIDs {
		// TODO: Stop container via containerClient when instance management is implemented
	}

	// Soft delete the stack
	_, err = h.db.Exec(`
		UPDATE stacks
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE name = $1 AND deleted_at IS NULL
	`, name)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete stack: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("stack %q deleted successfully", name),
	})
}
