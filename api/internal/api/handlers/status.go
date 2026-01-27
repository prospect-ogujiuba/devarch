package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/lib/pq"
	"github.com/priz/devarch-api/internal/podman"
	"github.com/priz/devarch-api/internal/sync"
)

type StatusHandler struct {
	db          *sql.DB
	podman      *podman.Client
	syncManager *sync.Manager
}

func NewStatusHandler(db *sql.DB, pc *podman.Client, sm *sync.Manager) *StatusHandler {
	return &StatusHandler{
		db:          db,
		podman:      pc,
		syncManager: sm,
	}
}

type overviewResponse struct {
	TotalServices    int                `json:"total_services"`
	EnabledServices  int                `json:"enabled_services"`
	RunningServices  int                `json:"running_services"`
	StoppedServices  int                `json:"stopped_services"`
	Categories       []categoryOverview `json:"categories"`
	ContainerRuntime string             `json:"container_runtime"`
	SocketPath       string             `json:"socket_path"`
}

type categoryOverview struct {
	Name            string `json:"name"`
	TotalServices   int    `json:"total_services"`
	RunningServices int    `json:"running_services"`
}

func (h *StatusHandler) Overview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var total, enabled int
	h.db.QueryRow("SELECT COUNT(*) FROM services").Scan(&total)
	h.db.QueryRow("SELECT COUNT(*) FROM services WHERE enabled = true").Scan(&enabled)

	running, stopped, _ := h.podman.GetContainerCounts(ctx)

	runningContainers, _ := h.podman.ListContainers(ctx, false)
	runningByName := make(map[string]bool)
	for _, c := range runningContainers {
		if len(c.Names) > 0 {
			runningByName[c.Names[0]] = true
		}
	}

	rows, err := h.db.Query(`
		SELECT c.name, COUNT(s.id) as total, ARRAY_AGG(s.name) FILTER (WHERE s.name IS NOT NULL)
		FROM categories c
		LEFT JOIN services s ON s.category_id = c.id AND s.enabled = true
		GROUP BY c.id, c.name
		ORDER BY c.startup_order
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []categoryOverview
	for rows.Next() {
		var co categoryOverview
		var serviceNames []string
		if err := rows.Scan(&co.Name, &co.TotalServices, pq.Array(&serviceNames)); err != nil {
			continue
		}
		for _, name := range serviceNames {
			if runningByName[name] {
				co.RunningServices++
			}
		}
		categories = append(categories, co)
	}

	resp := overviewResponse{
		TotalServices:    total,
		EnabledServices:  enabled,
		RunningServices:  running,
		StoppedServices:  stopped,
		Categories:       categories,
		ContainerRuntime: "podman",
		SocketPath:       h.podman.SocketPath(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *StatusHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	syncType := r.URL.Query().Get("type")
	if syncType == "" {
		syncType = "all"
	}

	jobID := h.syncManager.TriggerSync(syncType)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
		"status": "started",
	})
}

func (h *StatusHandler) SyncJobs(w http.ResponseWriter, r *http.Request) {
	jobs := h.syncManager.GetJobs()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}
