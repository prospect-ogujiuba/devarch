package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/priz/devarch-api/internal/compose"
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

func (h *StackHandler) runningCount(stackName string) int {
	count, err := h.containerClient.CountRunningWithLabels(map[string]string{
		container.LabelStackID: stackName,
	})
	if err != nil {
		return 0
	}
	return count
}

func (h *StackHandler) stackCompose(stackName string) ([]byte, error) {
	var networkName sql.NullString
	err := h.db.QueryRow(`
		SELECT network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&networkName)
	if err != nil {
		return nil, fmt.Errorf("lookup stack: %w", err)
	}

	netName := fmt.Sprintf("devarch-%s-net", stackName)
	if networkName.Valid && networkName.String != "" {
		netName = networkName.String
	}

	gen := compose.NewGenerator(h.db, netName)
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		gen.SetProjectRoot(root)
	}
	if hostRoot := os.Getenv("HOST_PROJECT_ROOT"); hostRoot != "" {
		gen.SetHostProjectRoot(hostRoot)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot != "" {
		if err := gen.MaterializeStackConfigs(stackName, projectRoot); err != nil {
			return nil, fmt.Errorf("materialize configs: %w", err)
		}
	}

	yamlBytes, _, err := gen.GenerateStack(stackName)
	if err != nil {
		return nil, fmt.Errorf("generate compose: %w", err)
	}

	return yamlBytes, nil
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
	Name        string  `json:"name"`
	Description string  `json:"description"`
	NetworkName *string `json:"network_name,omitempty"`
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

	// Compute or validate network_name
	networkName := container.NetworkName(req.Name)
	if req.NetworkName != nil && *req.NetworkName != "" {
		// User provided custom network name, validate it
		if err := container.ValidateName(*req.NetworkName); err != nil {
			http.Error(w, fmt.Sprintf("invalid network_name: %v", err), http.StatusBadRequest)
			return
		}
		networkName = *req.NetworkName
	}

	// Insert stack with network_name
	var stack stackResponse
	err := h.db.QueryRow(`
		INSERT INTO stacks (name, description, network_name)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, network_name, enabled, created_at, updated_at
	`, req.Name, req.Description, networkName).Scan(
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
		LEFT JOIN service_instances si ON si.stack_id = s.id AND si.deleted_at IS NULL
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

		stack.RunningCount = h.runningCount(stack.Name)

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
		LEFT JOIN service_instances si ON si.stack_id = s.id AND si.deleted_at IS NULL
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

	stack.RunningCount = h.runningCount(stack.Name)

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

	stack.RunningCount = h.runningCount(stack.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stack)
}

func (h *StackHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var stackID int
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

	yamlBytes, err := h.stackCompose(name)
	if err != nil {
		log.Printf("warning: could not generate compose for stack %q (may not have been deployed): %v", name, err)
	} else {
		if err := h.containerClient.StopStack("devarch-"+name, yamlBytes); err != nil {
			log.Printf("warning: failed to stop containers for stack %q: %v", name, err)
		}
	}

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

// Enable handles POST /stacks/{name}/enable
func (h *StackHandler) Enable(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var stack stackResponse
	err := h.db.QueryRow(`
		UPDATE stacks
		SET enabled = true, updated_at = NOW()
		WHERE name = $1 AND deleted_at IS NULL
		RETURNING id, name, description, network_name, enabled, created_at, updated_at
	`, name).Scan(
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
		http.Error(w, fmt.Sprintf("failed to enable stack: %v", err), http.StatusInternalServerError)
		return
	}

	// Query instance count
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM service_instances WHERE stack_id = $1
	`, stack.ID).Scan(&stack.InstanceCount)
	if err != nil {
		stack.InstanceCount = 0
	}

	stack.RunningCount = h.runningCount(stack.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stack)
}

type networkStatusResponse struct {
	NetworkName string            `json:"network_name"`
	Status      string            `json:"status"` // "active" or "not_created"
	Driver      string            `json:"driver,omitempty"`
	Containers  []string          `json:"containers"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// NetworkStatus handles GET /stacks/{name}/network
func (h *StackHandler) NetworkStatus(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	// Get stack with network_name
	var networkName *string
	err := h.db.QueryRow(`
		SELECT network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(&networkName)

	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
		return
	}

	response := networkStatusResponse{
		Containers: []string{},
	}

	if networkName != nil && *networkName != "" {
		response.NetworkName = *networkName

		// Inspect network via container client
		info, err := h.containerClient.InspectNetwork(*networkName)
		if err != nil {
			// Network doesn't exist yet
			response.Status = "not_created"
		} else {
			// Network exists
			response.Status = "active"
			response.Driver = info.Driver
			response.Containers = info.Containers
			response.Labels = info.Labels
		}
	} else {
		// No network_name set (shouldn't happen after migration)
		response.NetworkName = container.NetworkName(name)
		response.Status = "not_created"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *StackHandler) CreateNetwork(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var networkName *string
	err := h.db.QueryRow(`
		SELECT network_name FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(&networkName)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
		return
	}

	netName := container.NetworkName(name)
	if networkName != nil && *networkName != "" {
		netName = *networkName
	}

	labels := map[string]string{
		"devarch.managed_by": "devarch",
		"devarch.stack":      name,
	}
	if err := h.containerClient.CreateNetwork(netName, labels); err != nil {
		http.Error(w, fmt.Sprintf("failed to create network: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "created", "network": netName})
}

type disableResponse struct {
	Stack             stackResponse `json:"stack"`
	StoppedContainers []string      `json:"stopped_containers"`
}

func (h *StackHandler) Disable(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var stackID int
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

	var stoppedContainers []string
	yamlBytes, err := h.stackCompose(name)
	if err != nil {
		log.Printf("warning: could not generate compose for stack %q: %v", name, err)
	} else {
		running, _ := h.containerClient.ListContainersWithLabels(map[string]string{
			container.LabelStackID: name,
		})
		stoppedContainers = running

		if err := h.containerClient.StopStack("devarch-"+name, yamlBytes); err != nil {
			log.Printf("warning: failed to stop containers for stack %q: %v", name, err)
		}
	}

	var stack stackResponse
	err = h.db.QueryRow(`
		UPDATE stacks
		SET enabled = false, updated_at = NOW()
		WHERE name = $1 AND deleted_at IS NULL
		RETURNING id, name, description, network_name, enabled, created_at, updated_at
	`, name).Scan(
		&stack.ID,
		&stack.Name,
		&stack.Description,
		&stack.NetworkName,
		&stack.Enabled,
		&stack.CreatedAt,
		&stack.UpdatedAt,
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to disable stack: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM service_instances WHERE stack_id = $1
	`, stack.ID).Scan(&stack.InstanceCount)
	if err != nil {
		stack.InstanceCount = 0
	}

	stack.RunningCount = h.runningCount(stack.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disableResponse{
		Stack:             stack,
		StoppedContainers: stoppedContainers,
	})
}

type cloneRequest struct {
	Name string `json:"name"`
}

// Clone handles POST /stacks/{name}/clone
func (h *StackHandler) Clone(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var req cloneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate new name
	if err := container.ValidateName(req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Compute network_name for cloned stack (not copied from source)
	newNetworkName := container.NetworkName(req.Name)

	// Begin transaction
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Copy stack record with new name and new network_name
	var newStack stackResponse
	err = tx.QueryRow(`
		INSERT INTO stacks (name, description, network_name, enabled)
		SELECT $1, description, $3, true
		FROM stacks
		WHERE name = $2 AND deleted_at IS NULL
		RETURNING id, name, description, network_name, enabled, created_at, updated_at
	`, req.Name, name, newNetworkName).Scan(
		&newStack.ID,
		&newStack.Name,
		&newStack.Description,
		&newStack.NetworkName,
		&newStack.Enabled,
		&newStack.CreatedAt,
		&newStack.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			http.Error(w, fmt.Sprintf("stack name %q already exists", req.Name), http.StatusConflict)
			return
		}
		if err == sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to clone stack: %v", err), http.StatusInternalServerError)
		return
	}

	// Copy all service_instances from source to new stack
	_, err = tx.Exec(`
		INSERT INTO service_instances (stack_id, instance_id, template_service_id)
		SELECT $1, instance_id, template_service_id
		FROM service_instances si
		JOIN stacks s ON s.id = si.stack_id
		WHERE s.name = $2 AND s.deleted_at IS NULL
	`, newStack.ID, name)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to clone instances: %v", err), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("failed to commit transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Query instance count for new stack
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM service_instances WHERE stack_id = $1
	`, newStack.ID).Scan(&newStack.InstanceCount)
	if err != nil {
		newStack.InstanceCount = 0
	}

	newStack.RunningCount = 0

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newStack)
}

type renameRequest struct {
	Name string `json:"name"`
}

// Rename handles POST /stacks/{name}/rename
func (h *StackHandler) Rename(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var req renameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate new name
	if err := container.ValidateName(req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Compute network_name for renamed stack
	newNetworkName := container.NetworkName(req.Name)

	// Begin transaction - rename is clone + soft-delete in single tx
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Clone stack record with new name and new network_name
	var newStack stackResponse
	err = tx.QueryRow(`
		INSERT INTO stacks (name, description, network_name, enabled)
		SELECT $1, description, $3, enabled
		FROM stacks
		WHERE name = $2 AND deleted_at IS NULL
		RETURNING id, name, description, network_name, enabled, created_at, updated_at
	`, req.Name, name, newNetworkName).Scan(
		&newStack.ID,
		&newStack.Name,
		&newStack.Description,
		&newStack.NetworkName,
		&newStack.Enabled,
		&newStack.CreatedAt,
		&newStack.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			http.Error(w, fmt.Sprintf("stack name %q already exists", req.Name), http.StatusConflict)
			return
		}
		if err == sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to rename stack: %v", err), http.StatusInternalServerError)
		return
	}

	// Copy all service_instances from source to new stack
	_, err = tx.Exec(`
		INSERT INTO service_instances (stack_id, instance_id, template_service_id)
		SELECT $1, instance_id, template_service_id
		FROM service_instances si
		JOIN stacks s ON s.id = si.stack_id
		WHERE s.name = $2 AND s.deleted_at IS NULL
	`, newStack.ID, name)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to copy instances: %v", err), http.StatusInternalServerError)
		return
	}

	// Soft-delete the original stack
	_, err = tx.Exec(`
		UPDATE stacks
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE name = $1 AND deleted_at IS NULL
	`, name)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete original stack: %v", err), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("failed to commit transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Query instance count for new stack
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM service_instances WHERE stack_id = $1
	`, newStack.ID).Scan(&newStack.InstanceCount)
	if err != nil {
		newStack.InstanceCount = 0
	}

	newStack.RunningCount = 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newStack)
}

type deletePreviewResponse struct {
	StackName      string   `json:"stack_name"`
	InstanceCount  int      `json:"instance_count"`
	ContainerNames []string `json:"container_names"`
}

// DeletePreview handles GET /stacks/{name}/delete-preview
func (h *StackHandler) DeletePreview(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	// Get stack ID
	var stackID int
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

	// Get all instance IDs and build container names
	rows, err := h.db.Query(`
		SELECT id FROM service_instances WHERE stack_id = $1
	`, stackID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query instances: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var containerNames []string
	instanceCount := 0
	for rows.Next() {
		var instanceID int
		if err := rows.Scan(&instanceID); err != nil {
			continue
		}
		instanceCount++
		containerNames = append(containerNames, container.ContainerName(name, fmt.Sprintf("%d", instanceID)))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deletePreviewResponse{
		StackName:      name,
		InstanceCount:  instanceCount,
		ContainerNames: containerNames,
	})
}

type trashStackResponse struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	NetworkName   *string    `json:"network_name"`
	Enabled       bool       `json:"enabled"`
	InstanceCount int        `json:"instance_count"`
	DeletedAt     *time.Time `json:"deleted_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ListTrash handles GET /stacks/trash
func (h *StackHandler) ListTrash(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			s.id,
			s.name,
			s.description,
			s.network_name,
			s.enabled,
			s.created_at,
			s.updated_at,
			s.deleted_at,
			COUNT(si.id) AS instance_count
		FROM stacks s
		LEFT JOIN service_instances si ON si.stack_id = s.id AND si.deleted_at IS NULL
		WHERE s.deleted_at IS NOT NULL
		GROUP BY s.id
		ORDER BY s.deleted_at DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query trash: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	stacks := []trashStackResponse{}
	for rows.Next() {
		var stack trashStackResponse
		err := rows.Scan(
			&stack.ID,
			&stack.Name,
			&stack.Description,
			&stack.NetworkName,
			&stack.Enabled,
			&stack.CreatedAt,
			&stack.UpdatedAt,
			&stack.DeletedAt,
			&stack.InstanceCount,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to scan stack: %v", err), http.StatusInternalServerError)
			return
		}

		stacks = append(stacks, stack)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("error iterating trash: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stacks)
}

// Restore handles POST /stacks/trash/{name}/restore
func (h *StackHandler) Restore(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	// Check if name conflicts with an active stack
	var existingID int
	err := h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, name).Scan(&existingID)

	if err == nil {
		http.Error(w, fmt.Sprintf("Stack name %q is already in use. Rename before restoring.", name), http.StatusConflict)
		return
	}
	if err != sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("failed to check name conflict: %v", err), http.StatusInternalServerError)
		return
	}

	// Restore stack by clearing deleted_at
	var stack stackResponse
	err = h.db.QueryRow(`
		UPDATE stacks
		SET deleted_at = NULL, updated_at = NOW()
		WHERE name = $1 AND deleted_at IS NOT NULL
		RETURNING id, name, description, network_name, enabled, created_at, updated_at
	`, name).Scan(
		&stack.ID,
		&stack.Name,
		&stack.Description,
		&stack.NetworkName,
		&stack.Enabled,
		&stack.CreatedAt,
		&stack.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("stack %q not found in trash", name), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to restore stack: %v", err), http.StatusInternalServerError)
		return
	}

	// Query instance count
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM service_instances WHERE stack_id = $1
	`, stack.ID).Scan(&stack.InstanceCount)
	if err != nil {
		stack.InstanceCount = 0
	}

	stack.RunningCount = h.runningCount(stack.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stack)
}

// PermanentDelete handles DELETE /stacks/trash/{name}
func (h *StackHandler) PermanentDelete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	result, err := h.db.Exec(`
		DELETE FROM stacks
		WHERE name = $1 AND deleted_at IS NOT NULL
	`, name)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to permanently delete stack: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get rows affected: %v", err), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, fmt.Sprintf("stack %q not found in trash", name), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("stack %q permanently deleted", name),
	})
}
