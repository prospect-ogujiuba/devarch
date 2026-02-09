package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
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

func (h *NetworkHandler) List(w http.ResponseWriter, r *http.Request) {
	names, err := h.containerClient.ListAllNetworks()
	if err != nil {
		http.Error(w, fmt.Sprintf("list networks: %v", err), http.StatusInternalServerError)
		return
	}

	stackSet := map[string]bool{}
	rows, err := h.db.Query(`SELECT name FROM stacks WHERE deleted_at IS NULL`)
	if err != nil {
		http.Error(w, fmt.Sprintf("query stacks: %v", err), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

type createNetworkRequest struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
}

var networkNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

func (h *NetworkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if !networkNameRe.MatchString(req.Name) {
		http.Error(w, "invalid network name", http.StatusBadRequest)
		return
	}

	labels := map[string]string{
		container.LabelManagedBy: container.ManagedByValue,
	}

	if err := h.containerClient.CreateNetwork(req.Name, labels); err != nil {
		http.Error(w, fmt.Sprintf("create network: %v", err), http.StatusInternalServerError)
		return
	}

	info, err := h.containerClient.InspectNetwork(req.Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "created",
			"network": req.Name,
		})
		return
	}

	containers := info.Containers
	if containers == nil {
		containers = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(networkListItem{
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

func (h *NetworkHandler) Remove(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	info, err := h.containerClient.InspectNetwork(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("network not found: %v", err), http.StatusNotFound)
		return
	}

	if len(info.Containers) > 0 {
		http.Error(w, "network has connected containers", http.StatusConflict)
		return
	}

	if err := h.containerClient.RemoveNetwork(name); err != nil {
		http.Error(w, fmt.Sprintf("remove network: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
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

func (h *NetworkHandler) BulkRemove(w http.ResponseWriter, r *http.Request) {
	var req bulkRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
