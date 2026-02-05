package sync

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/priz/devarch-api/internal/podman"
)

const (
	metricsCleanupTimeout = 60 * time.Second
	dailyCleanupTimeout   = 5 * time.Minute
	dailyOpTimeout        = 90 * time.Second
	maxBatchesPerOp       = 10
)

type Manager struct {
	db           *sql.DB
	podman       *podman.Client
	jobs         map[string]*Job
	jobsMu       sync.RWMutex
	statusCache  *StatusCache
	eventCancel  context.CancelFunc
	registrySync *RegistrySync
	trivyScanner *TrivyScanner

	metricsInterval  time.Duration
	metricsRetention time.Duration
	cleanupInterval  time.Duration
	cleanupBatch     int

	configVersionsMax      int
	registryImageRetention time.Duration
	vulnOrphanRetention    time.Duration
	softDeleteRetention    time.Duration

	lastDailyCleanup time.Time
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
	metricsInterval := mustParseDurationEnv("DEVARCH_METRICS_INTERVAL", "30s")
	if metricsInterval < 5*time.Second {
		metricsInterval = 5 * time.Second
	}

	metricsRetention := mustParseDurationEnv("DEVARCH_METRICS_RETENTION", "3d")
	if metricsRetention < time.Hour {
		metricsRetention = time.Hour
	}

	cleanupInterval := mustParseDurationEnv("DEVARCH_CLEANUP_INTERVAL", "1h")
	if cleanupInterval < time.Minute {
		cleanupInterval = time.Minute
	}

	cleanupBatch := mustParseIntEnv("DEVARCH_CLEANUP_BATCH", 50000)
	if cleanupBatch < 1000 {
		cleanupBatch = 1000
	}

	configVersionsMax := mustParseIntEnv("DEVARCH_CONFIG_VERSIONS_MAX", 25)
	if configVersionsMax < 0 {
		configVersionsMax = 0
	}

	// Optional retention: 0 disables, negative rejected (falls back to default via parseDurationWithDays)
	registryImageRetention := mustParseDurationEnv("DEVARCH_REGISTRY_IMAGE_RETENTION", "30d")
	vulnOrphanRetention := mustParseDurationEnv("DEVARCH_VULN_ORPHAN_RETENTION", "30d")
	softDeleteRetention := mustParseDurationEnv("DEVARCH_SOFT_DELETE_RETENTION", "30d")

	log.Printf("sync: config — metrics every %s retain %s, cleanup every %s batch %d, config versions max %d",
		metricsInterval, metricsRetention, cleanupInterval, cleanupBatch, configVersionsMax)

	return &Manager{
		db:           db,
		podman:       pc,
		jobs:         make(map[string]*Job),
		registrySync: NewRegistrySync(db),
		trivyScanner: NewTrivyScanner(db),
		statusCache: &StatusCache{
			data:       make(map[string]interface{}),
			containers: make(map[string]string),
		},

		metricsInterval:  metricsInterval,
		metricsRetention: metricsRetention,
		cleanupInterval:  cleanupInterval,
		cleanupBatch:     cleanupBatch,

		configVersionsMax:      configVersionsMax,
		registryImageRetention: registryImageRetention,
		vulnOrphanRetention:    vulnOrphanRetention,
		softDeleteRetention:    softDeleteRetention,
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
	ticker := time.NewTicker(m.metricsInterval)
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
	// Jitter initial delay to avoid aligned DB pressure across instances
	if maxJitter := m.cleanupInterval / 10; maxJitter > 0 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(rng.Int63n(int64(maxJitter)))):
		}
	}

	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mCtx, mCancel := context.WithTimeout(ctx, metricsCleanupTimeout)
			m.cleanupOldMetrics(mCtx)
			mCancel()

			dCtx, dCancel := context.WithTimeout(ctx, dailyCleanupTimeout)
			m.runDailyCleanupIfDue(dCtx)
			dCancel()
		}
	}
}

func (m *Manager) cleanupOldMetrics(ctx context.Context) {
	cutoff := time.Now().Add(-m.metricsRetention)
	m.deleteInBatches(
		ctx,
		`WITH doomed AS (
			SELECT id FROM container_metrics
			WHERE recorded_at < $1
			LIMIT $2
		)
		DELETE FROM container_metrics cm
		USING doomed d
		WHERE cm.id = d.id`,
		cutoff,
		m.cleanupBatch,
		"cleanup old metrics",
	)
}

func (m *Manager) runDailyCleanupIfDue(ctx context.Context) {
	if !m.lastDailyCleanup.IsZero() && time.Since(m.lastDailyCleanup) < 24*time.Hour {
		return
	}
	for _, op := range []struct {
		name string
		fn   func(context.Context)
	}{
		{"config versions", m.cleanupConfigVersions},
		{"registry images", m.cleanupRegistryImages},
		{"orphan vulns", m.cleanupOrphanVulnerabilities},
		{"soft-deleted", m.cleanupSoftDeleted},
	} {
		if ctx.Err() != nil {
			return
		}
		opCtx, cancel := context.WithTimeout(ctx, dailyOpTimeout)
		op.fn(opCtx)
		cancel()
	}
	m.lastDailyCleanup = time.Now()
}

func (m *Manager) cleanupConfigVersions(ctx context.Context) {
	if m.configVersionsMax <= 0 {
		return
	}

	m.deleteInBatches(
		ctx,
		`WITH doomed AS (
			SELECT v.id
			FROM (SELECT DISTINCT service_id FROM service_config_versions) s,
			LATERAL (
				SELECT id FROM service_config_versions
				WHERE service_id = s.service_id
				ORDER BY version DESC
				OFFSET $1
			) v
			LIMIT $2
		)
		DELETE FROM service_config_versions
		USING doomed WHERE service_config_versions.id = doomed.id`,
		m.configVersionsMax,
		m.cleanupBatch,
		"cleanup old config versions",
	)
}

func (m *Manager) cleanupRegistryImages(ctx context.Context) {
	if m.registryImageRetention <= 0 {
		return
	}
	cutoff := time.Now().Add(-m.registryImageRetention)
	m.deleteInBatches(ctx,
		`WITH doomed AS (
			SELECT i.id FROM images i
			WHERE i.last_synced_at < $1
			AND NOT EXISTS (
				SELECT 1 FROM services s
				WHERE s.enabled = true
				AND s.image_name = i.repository
			)
			LIMIT $2
		)
		DELETE FROM images USING doomed WHERE images.id = doomed.id`,
		cutoff, m.cleanupBatch, "cleanup registry images",
	)
}

func (m *Manager) cleanupOrphanVulnerabilities(ctx context.Context) {
	if m.vulnOrphanRetention <= 0 {
		return
	}
	cutoff := time.Now().Add(-m.vulnOrphanRetention)
	m.deleteInBatches(ctx,
		`WITH doomed AS (
			SELECT v.id FROM vulnerabilities v
			WHERE v.created_at < $1
			AND NOT EXISTS (
				SELECT 1 FROM image_tag_vulnerabilities tv
				WHERE tv.vulnerability_id = v.id
			)
			LIMIT $2
		)
		DELETE FROM vulnerabilities USING doomed WHERE vulnerabilities.id = doomed.id`,
		cutoff, m.cleanupBatch, "cleanup orphan vulnerabilities",
	)
}

func (m *Manager) cleanupSoftDeleted(ctx context.Context) {
	if m.softDeleteRetention <= 0 {
		return
	}
	cutoff := time.Now().Add(-m.softDeleteRetention)
	m.deleteInBatches(ctx,
		`WITH doomed AS (
			SELECT id FROM service_instances
			WHERE deleted_at IS NOT NULL AND deleted_at < $1
			LIMIT $2
		)
		DELETE FROM service_instances USING doomed WHERE service_instances.id = doomed.id`,
		cutoff, m.cleanupBatch, "cleanup soft-deleted instances",
	)
	m.deleteInBatches(ctx,
		`WITH doomed AS (
			SELECT id FROM stacks
			WHERE deleted_at IS NOT NULL AND deleted_at < $1
			LIMIT $2
		)
		DELETE FROM stacks USING doomed WHERE stacks.id = doomed.id`,
		cutoff, m.cleanupBatch, "cleanup soft-deleted stacks",
	)
}

func (m *Manager) deleteInBatches(ctx context.Context, query string, arg1 any, batchSize int, label string) {
	var total int64
	for i := 0; i < maxBatchesPerOp; i++ {
		if ctx.Err() != nil {
			if total > 0 {
				log.Printf("sync: %s: deleted %d rows (interrupted)", label, total)
			}
			return
		}
		res, err := m.db.ExecContext(ctx, query, arg1, batchSize)
		if err != nil {
			if total > 0 {
				log.Printf("sync: %s: deleted %d rows before error: %v", label, total, err)
			} else {
				log.Printf("sync: %s failed: %v", label, err)
			}
			return
		}
		n, _ := res.RowsAffected()
		total += n
		if n < int64(batchSize) {
			break
		}
	}
	if total >= int64(batchSize) {
		log.Printf("sync: %s: deleted %d rows", label, total)
	}
}

func mustParseIntEnv(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("sync: invalid %s=%q, using %d", key, v, def)
		return def
	}
	return i
}

func mustParseDurationEnv(key string, def string) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		v = def
	}
	d, err := parseDurationWithDays(v)
	if err != nil {
		dd, derr := parseDurationWithDays(def)
		if derr != nil {
			log.Printf("sync: invalid %s=%q and default %q, using 0", key, v, def)
			return 0
		}
		log.Printf("sync: invalid %s=%q, using default %q", key, v, def)
		return dd
	}
	return d
}

func parseDurationWithDays(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if strings.HasSuffix(s, "d") {
		base := strings.TrimSpace(strings.TrimSuffix(s, "d"))
		if base == "" {
			return 0, fmt.Errorf("invalid day duration")
		}
		days, err := strconv.ParseFloat(base, 64)
		if err != nil {
			return 0, err
		}
		if days < 0 || math.IsNaN(days) || math.IsInf(days, 0) {
			return 0, fmt.Errorf("invalid day value: %v", days)
		}
		ns := days * 24 * float64(time.Hour)
		if ns > float64(math.MaxInt64) {
			return 0, fmt.Errorf("duration overflow: %vd", days)
		}
		return time.Duration(ns), nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, fmt.Errorf("negative duration: %s", s)
	}
	return d, nil
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
		case "registry":
			err = m.registrySync.SyncAll(ctx)
		case "trivy":
			err = m.trivyScanner.ScanAll(ctx)
		case "all":
			m.syncContainerStatus(ctx)
			m.collectMetrics(ctx)
			if syncErr := m.registrySync.SyncAll(ctx); syncErr != nil {
				log.Printf("sync: registry sync failed: %v", syncErr)
			}
			err = m.trivyScanner.ScanAll(ctx)
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
