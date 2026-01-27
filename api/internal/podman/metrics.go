package podman

import (
	"context"
	"time"

	"github.com/priz/devarch-api/pkg/models"
)

func (c *Client) GetServiceMetrics(ctx context.Context, name string) (*models.ContainerMetrics, error) {
	stats, err := c.ContainerStats(ctx, name)
	if err != nil {
		return nil, err
	}

	return ConvertToMetrics(stats), nil
}

func ConvertToMetrics(stats *ContainerStats) *models.ContainerMetrics {
	metrics := &models.ContainerMetrics{
		RecordedAt: time.Now(),
	}

	metrics.CPUPercentage = calculateCPUPercent(stats)

	if stats.MemoryStats.Limit > 0 {
		metrics.MemoryUsedMB = float64(stats.MemoryStats.Usage) / (1024 * 1024)
		metrics.MemoryLimitMB = float64(stats.MemoryStats.Limit) / (1024 * 1024)
		metrics.MemoryPercentage = (float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit)) * 100
	}

	var rxBytes, txBytes uint64
	for _, netStats := range stats.NetworkStats {
		rxBytes += netStats.RxBytes
		txBytes += netStats.TxBytes
	}
	metrics.NetworkRxBytes = int64(rxBytes)
	metrics.NetworkTxBytes = int64(txBytes)

	return metrics
}

func calculateCPUPercent(stats *ContainerStats) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)

	if systemDelta > 0 && cpuDelta > 0 {
		cpuPercent := (cpuDelta / systemDelta) * float64(stats.CPUStats.OnlineCPUs) * 100.0
		return cpuPercent
	}

	return 0.0
}

func (c *Client) GetServiceState(ctx context.Context, name string) (*models.ContainerState, error) {
	health, err := c.GetContainerHealth(ctx, name)
	if err != nil {
		return &models.ContainerState{
			Status: "stopped",
		}, nil
	}

	state := &models.ContainerState{
		Status:       health.Status,
		RestartCount: health.RestartCount,
		ErrorStr:     health.Error,
	}

	if health.Health != "" {
		state.HealthStr = health.Health
	}

	if health.StartedAt != nil {
		state.StartedAtStr = health.StartedAt
	}

	if health.FinishedAt != nil {
		state.FinishedStr = health.FinishedAt
	}

	if health.ExitCode != nil {
		state.ExitCodeInt = health.ExitCode
	}

	return state, nil
}

type BatchMetrics struct {
	States  map[string]*models.ContainerState
	Metrics map[string]*models.ContainerMetrics
}

func (c *Client) GetBatchServiceData(ctx context.Context, names []string) (*BatchMetrics, error) {
	result := &BatchMetrics{
		States:  make(map[string]*models.ContainerState, len(names)),
		Metrics: make(map[string]*models.ContainerMetrics, len(names)),
	}

	for _, name := range names {
		state, err := c.GetServiceState(ctx, name)
		if err == nil {
			result.States[name] = state
		}

		if state != nil && state.Status == "running" {
			metrics, err := c.GetServiceMetrics(ctx, name)
			if err == nil {
				result.Metrics[name] = metrics
			}
		}
	}

	return result, nil
}

func (c *Client) GetRunningContainerNames(ctx context.Context) ([]string, error) {
	containers, err := c.ListContainers(ctx, false)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(containers))
	for _, container := range containers {
		if len(container.Names) > 0 {
			names = append(names, container.Names[0])
		}
	}

	return names, nil
}

func (c *Client) GetContainerCounts(ctx context.Context) (running int, stopped int, err error) {
	containers, err := c.ListContainers(ctx, true)
	if err != nil {
		return 0, 0, err
	}

	for _, container := range containers {
		if container.State == "running" {
			running++
		} else {
			stopped++
		}
	}

	return running, stopped, nil
}
