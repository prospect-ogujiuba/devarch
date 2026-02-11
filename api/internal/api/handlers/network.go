package handlers

import (
	"encoding/json"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/container"
)

type NetworkHandler struct {
	db              *sql.DB
	containerClient *container.Client
}

func NewNetworkHandler(db *sql.DB, cc *container.Client) *NetworkHandler {
	return &NetworkHandler{
		db:              db,
		containerClient: cc,
	}
}

type networkListItem struct {
	Name           string            `json:"name"`
	ID             string            `json:"id"`
	Driver         string            `json:"driver"`
	Containers     []string          `json:"containers"`
	ContainerCount int               `json:"container_count"`
	Labels         map[string]string `json:"labels"`
	StackName      string            `json:"stack_name,omitempty"`
	Managed        bool              `json:"managed"`
	Orphaned       bool              `json:"orphaned"`
	Created        time.Time         `json:"created"`
}

// List godoc
// @Summary      List container networks
// @Description  Returns all container networks with metadata including managed status, containers, and orphan detection
// @Tags         networks
// @Produce      json
// @Success      200 {object} respond.SuccessEnvelope{data=[]networkListItem}
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /networks [get]
// @Security     ApiKeyAuth
func (h *NetworkHandler) List(w http.ResponseWriter, r *http.Request) {
	names, err := h.containerClient.ListAllNetworks()
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("list networks: %w", err))
		return
	}

	stackSet := map[string]bool{}
	rows, err := h.db.Query(`SELECT name FROM stacks WHERE deleted_at IS NULL`)
	if err != nil {
		respond.InternalError(w, r, fmt.Errorf("query stacks: %w", err))
		return
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			stackSet[name] = true
		}
	}

	items := make([]networkListItem, 0, len(names))
	for _, name := range names {
		info, err := h.containerClient.InspectNetwork(name)
		if err != nil {
			continue
		}

		containers := info.Containers
		if containers == nil {
			containers = []string{}
		}

		managed := container.IsDevArchManaged(info.Labels)

		item := networkListItem{
			Name:           info.Name,
			ID:             info.ID,
			Driver:         info.Driver,
			Containers:     containers,
			ContainerCount: len(containers),
			Labels:         info.Labels,
			Managed:        managed,
			Created:        info.Created,
		}

		if managed {
			if stackName, ok := info.Labels["devarch.stack"]; ok {
				item.StackName = stackName
				if !stackSet[stackName] {
					item.Orphaned = true
				}
			} else {
				item.Orphaned = true
			}
		}

		items = append(items, item)
	}

	respond.JSON(w, r, http.StatusOK,items)
}

type createNetworkRequest struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
}

var networkNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

// Create godoc
// @Summary      Create container network
// @Description  Creates a new container network with DevArch management labels
// @Tags         networks
// @Accept       json
// @Produce      json
// @Param        request body createNetworkRequest true "Network creation request"
// @Success      201 {object} respond.SuccessEnvelope{data=networkListItem}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /networks [post]
// @Security     ApiKeyAuth
func (h *NetworkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid request body")
		return
	}

	if req.Name == "" {
		respond.BadRequest(w, r, "name is required")
		return
	}

	if !networkNameRe.MatchString(req.Name) {
		respond.BadRequest(w, r, "invalid network name")
		return
	}

	labels := map[string]string{
		container.LabelManagedBy: container.ManagedByValue,
	}

	if err := h.containerClient.CreateNetwork(req.Name, labels); err != nil {
		respond.InternalError(w, r, fmt.Errorf("create network: %w", err))
		return
	}

	info, err := h.containerClient.InspectNetwork(req.Name)
	if err != nil {
		respond.JSON(w, r, http.StatusCreated, map[string]string{
			"status":  "created",
			"network": req.Name,
		})
		return
	}

	containers := info.Containers
	if containers == nil {
		containers = []string{}
	}

	respond.JSON(w, r, http.StatusCreated, networkListItem{
		Name:           info.Name,
		ID:             info.ID,
		Driver:         info.Driver,
		Containers:     containers,
		ContainerCount: len(containers),
		Labels:         info.Labels,
		Managed:        true,
		Created:        info.Created,
	})
}

// Remove godoc
// @Summary      Remove container network
// @Description  Removes a container network by name (must have no connected containers)
// @Tags         networks
// @Produce      json
// @Param        name path string true "Network name"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      409 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /networks/{name} [delete]
// @Security     ApiKeyAuth
func (h *NetworkHandler) Remove(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	info, err := h.containerClient.InspectNetwork(name)
	if err != nil {
		respond.NotFound(w, r, "network", name)
		return
	}

	if len(info.Containers) > 0 {
		respond.Conflict(w, r, "network has connected containers")
		return
	}

	if err := h.containerClient.RemoveNetwork(name); err != nil {
		respond.InternalError(w, r, fmt.Errorf("remove network: %w", err))
		return
	}

	respond.JSON(w, r, http.StatusOK,map[string]string{
		"status":  "removed",
		"network": name,
	})
}

type bulkRemoveRequest struct {
	Names []string `json:"names"`
}

type bulkRemoveResponse struct {
	Removed []string    `json:"removed"`
	Errors  []bulkError `json:"errors"`
}

type bulkError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

// BulkRemove godoc
// @Summary      Remove multiple networks
// @Description  Removes multiple container networks by name with partial success handling
// @Tags         networks
// @Accept       json
// @Produce      json
// @Param        request body bulkRemoveRequest true "Bulk remove request"
// @Success      200 {object} respond.SuccessEnvelope{data=bulkRemoveResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Router       /networks/bulk-remove [post]
// @Security     ApiKeyAuth
func (h *NetworkHandler) BulkRemove(w http.ResponseWriter, r *http.Request) {
	var req bulkRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid request body")
		return
	}

	resp := bulkRemoveResponse{
		Removed: []string{},
		Errors:  []bulkError{},
	}

	for _, name := range req.Names {
		info, err := h.containerClient.InspectNetwork(name)
		if err != nil {
			resp.Errors = append(resp.Errors, bulkError{Name: name, Error: "network not found"})
			continue
		}

		if len(info.Containers) > 0 {
			resp.Errors = append(resp.Errors, bulkError{Name: name, Error: "has connected containers"})
			continue
		}

		if err := h.containerClient.RemoveNetwork(name); err != nil {
			resp.Errors = append(resp.Errors, bulkError{Name: name, Error: err.Error()})
			continue
		}

		resp.Removed = append(resp.Removed, name)
	}

	respond.JSON(w, r, http.StatusOK,resp)
}
