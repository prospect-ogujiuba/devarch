package container

import "time"

type RuntimeType string

const (
	RuntimePodman RuntimeType = "podman"
	RuntimeDocker RuntimeType = "docker"
)

// Runtime is the unified container runtime interface.
type Runtime interface {
	// Identity
	RuntimeName() RuntimeType

	// Container lifecycle (compose-based)
	StartService(name string, composeYAML []byte) error
	StopService(name string, composeYAML []byte) error
	RestartService(name string, composeYAML []byte) error
	RebuildService(name string, composeYAML []byte, noCache bool) error

	// Container inspection
	GetContainerStatus(name string) (*ContainerStatus, error)
	GetLogs(name string, tail string) (string, error)
	GetContainerMetrics(name string) (*ContainerMetrics, error)
	ListContainers() ([]string, error)
	ListContainersWithLabels(labels map[string]string) ([]string, error)
	GetRunningCount() (running int, stopped int)

	// Container exec
	Exec(containerName string, command []string) (string, error)

	// Network operations (stubs in Phase 1, implemented Phase 4)
	CreateNetwork(name string, labels map[string]string) error
	RemoveNetwork(name string) error
	ListNetworks() ([]string, error)
}

// ContainerStatus holds container state info
type ContainerStatus struct {
	Status       string
	Health       string
	StartedAt    *time.Time
	FinishedAt   *time.Time
	ExitCode     *int
	Error        string
	RestartCount int
}

// ContainerMetrics holds resource usage
type ContainerMetrics struct {
	CPUPercentage    float64
	MemoryUsedMB     float64
	MemoryLimitMB    float64
	MemoryPercentage float64
	NetworkRxBytes   int64
	NetworkTxBytes   int64
	RecordedAt       time.Time
}
