package sync

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/priz/devarch-api/internal/podman"
)

type Manager struct {
	db          *sql.DB
	podman      *podman.Client
	jobs        map[string]*Job
	jobsMu      sync.RWMutex
	statusCache *StatusCache
	eventCancel context.CancelFunc
}

type Job struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Error     string     `json:"error,omitempty"`
}

type StatusCache struct {
	mu         sync.RWMutex
	data       map[string]interface{}
	updated    time.Time
	containers map[string]string
}

type StatusUpdate struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

func NewManager(db *sql.DB, pc *podman.Client) *Manager {
	return &Manager{
		db:     db,
		podman: pc,
		jobs:   make(map[string]*Job),
		statusCache: &StatusCache{
			data:       make(map[string]interface{}),
			containers: make(map[string]string),
		},
	}
}

func (m *Manager) Start(ctx context.Context) {
	go m.containerStatusLoop(ctx)
	go m.metricsLoop(ctx)
	go m.cleanupLoop(ctx)
	// Event loop disabled - causes connection spam with rootless podman
	// go m.eventLoop(ctx)
}

func (m *Manager) containerStatusLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	m.syncContainerStatus(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.syncContainerStatus(ctx)
		}
	}
}

func (m *Manager) syncContainerStatus(ctx context.Context) {
	containers, err := m.podman.ListContainers(ctx, true)
	if err != nil {
		log.Printf("sync: failed to list containers: %v", err)
		return
	}

	runningSet := make(map[string]string)
	for _, c := range containers {
		if len(c.Names) > 0 {
			runningSet[c.Names[0]] = c.State
		}
	}

	rows, err := m.db.Query("SELECT id, name FROM services")
	if err != nil {
		log.Printf("sync: failed to query services: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}

		status := "stopped"
		if state, ok := runningSet[name]; ok {
			status = state
		}

		_, err := m.db.Exec(`
			INSERT INTO container_states (service_id, status, updated_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (service_id) DO UPDATE SET status = $2, updated_at = NOW()
		`, id, status)
		if err != nil {
			log.Printf("sync: failed to update status for %s: %v", name, err)
		}
	}

	m.statusCache.mu.Lock()
	m.statusCache.containers = runningSet
	m.statusCache.data["containers"] = runningSet
	m.statusCache.updated = time.Now()
	m.statusCache.mu.Unlock()
}

func (m *Manager) metricsLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics(ctx)
		}
	}
}

func (m *Manager) collectMetrics(ctx context.Context) {
	rows, err := m.db.Query(`
		SELECT s.id, s.name FROM services s
		JOIN container_states cs ON cs.service_id = s.id
		WHERE cs.status = 'running'
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}

		metrics, err := m.podman.GetServiceMetrics(ctx, name)
		if err != nil {
			continue
		}

		_, err = m.db.Exec(`
			INSERT INTO container_metrics (service_id, cpu_percentage, memory_used_mb, memory_limit_mb, memory_percentage, network_rx_bytes, network_tx_bytes)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, id, metrics.CPUPercentage, metrics.MemoryUsedMB, metrics.MemoryLimitMB, metrics.MemoryPercentage, metrics.NetworkRxBytes, metrics.NetworkTxBytes)
		if err != nil {
			log.Printf("sync: failed to insert metrics for %s: %v", name, err)
		}
	}
}

func (m *Manager) eventLoop(ctx context.Context) {
	eventCtx, cancel := context.WithCancel(ctx)
	m.eventCancel = cancel

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := m.podman.StreamEvents(eventCtx, time.Now(), &podman.EventFilter{
			Type: []string{"container"},
		}, func(event *podman.Event, err error) bool {
			if err != nil {
				log.Printf("sync: event error: %v", err)
				return true
			}

			m.handleContainerEvent(ctx, event)
			return true
		})

		if err != nil && ctx.Err() == nil {
			log.Printf("sync: event stream disconnected: %v, reconnecting...", err)
			time.Sleep(2 * time.Second)
		}
	}
}

func (m *Manager) handleContainerEvent(ctx context.Context, event *podman.Event) {
	name := podman.GetContainerName(event)

	switch event.Action {
	case "start", "restart":
		m.updateContainerStatus(ctx, name, "running")
	case "stop", "die", "kill":
		m.updateContainerStatus(ctx, name, "stopped")
	case "pause":
		m.updateContainerStatus(ctx, name, "paused")
	case "unpause":
		m.updateContainerStatus(ctx, name, "running")
	case "health_status":
		m.updateContainerHealth(ctx, name, event)
	}
}

func (m *Manager) updateContainerStatus(ctx context.Context, name string, status string) {
	_, err := m.db.Exec(`
		UPDATE container_states cs
		SET status = $1, updated_at = NOW()
		FROM services s
		WHERE s.id = cs.service_id AND s.name = $2
	`, status, name)
	if err != nil {
		log.Printf("sync: failed to update status for %s: %v", name, err)
	}

	m.statusCache.mu.Lock()
	m.statusCache.containers[name] = status
	m.statusCache.updated = time.Now()
	m.statusCache.mu.Unlock()
}

func (m *Manager) updateContainerHealth(ctx context.Context, name string, event *podman.Event) {
	healthStatus := ""
	if attrs := event.Actor.Attributes; attrs != nil {
		if hs, ok := attrs["health_status"]; ok {
			healthStatus = hs
		}
	}

	if healthStatus == "" {
		return
	}

	_, err := m.db.Exec(`
		UPDATE container_states cs
		SET health_status = $1, updated_at = NOW()
		FROM services s
		WHERE s.id = cs.service_id AND s.name = $2
	`, healthStatus, name)
	if err != nil {
		log.Printf("sync: failed to update health for %s: %v", name, err)
	}
}

func (m *Manager) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.cleanupOldMetrics()
		}
	}
}

func (m *Manager) cleanupOldMetrics() {
	_, err := m.db.Exec("DELETE FROM container_metrics WHERE recorded_at < NOW() - INTERVAL '7 days'")
	if err != nil {
		log.Printf("sync: failed to cleanup old metrics: %v", err)
	}
}

func (m *Manager) TriggerSync(syncType string) string {
	jobID := time.Now().Format("20060102150405")

	job := &Job{
		ID:        jobID,
		Type:      syncType,
		Status:    "running",
		StartedAt: time.Now(),
	}

	m.jobsMu.Lock()
	m.jobs[jobID] = job
	m.jobsMu.Unlock()

	go func() {
		ctx := context.Background()
		var err error

		switch syncType {
		case "containers":
			m.syncContainerStatus(ctx)
		case "metrics":
			m.collectMetrics(ctx)
		case "all":
			m.syncContainerStatus(ctx)
			m.collectMetrics(ctx)
		}

		m.jobsMu.Lock()
		now := time.Now()
		job.EndedAt = &now
		if err != nil {
			job.Status = "failed"
			job.Error = err.Error()
		} else {
			job.Status = "completed"
		}
		m.jobsMu.Unlock()
	}()

	return jobID
}

func (m *Manager) GetJobs() []*Job {
	m.jobsMu.RLock()
	defer m.jobsMu.RUnlock()

	jobs := make([]*Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (m *Manager) GetStatusUpdate() *StatusUpdate {
	m.statusCache.mu.RLock()
	defer m.statusCache.mu.RUnlock()

	if m.statusCache.data == nil || len(m.statusCache.data) == 0 {
		return nil
	}

	return &StatusUpdate{
		Type:      "status",
		Timestamp: m.statusCache.updated,
		Data:      m.statusCache.data,
	}
}

func (m *Manager) PodmanClient() *podman.Client {
	return m.podman
}
