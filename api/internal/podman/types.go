package podman

import (
	"encoding/json"
	"strconv"
	"time"
)

type FlexibleTime int64

func (f *FlexibleTime) UnmarshalJSON(data []byte) error {
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		*f = FlexibleTime(i)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			*f = FlexibleTime(i)
			return nil
		}
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			*f = FlexibleTime(t.Unix())
			return nil
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			*f = FlexibleTime(t.Unix())
			return nil
		}
	}

	*f = 0
	return nil
}

type SystemInfo struct {
	Host    HostInfo    `json:"host"`
	Store   StoreInfo   `json:"store"`
	Version VersionInfo `json:"version"`
}

type HostInfo struct {
	Arch              string `json:"arch"`
	Hostname          string `json:"hostname"`
	Kernel            string `json:"kernel"`
	OS                string `json:"os"`
	CPUs              int    `json:"cpus"`
	MemTotal          int64  `json:"memTotal"`
	SwapTotal         int64  `json:"swapTotal"`
	RemoteSocket      *RemoteSocket `json:"remoteSocket,omitempty"`
	Security          SecurityInfo  `json:"security"`
	ServiceIsRemote   bool          `json:"serviceIsRemote"`
}

type RemoteSocket struct {
	Exists bool   `json:"exists"`
	Path   string `json:"path"`
}

type SecurityInfo struct {
	AppArmorEnabled     bool   `json:"apparmorEnabled"`
	RootlessMode        bool   `json:"rootless"`
	SELinuxEnabled      bool   `json:"selinuxEnabled"`
	SeccompEnabled      bool   `json:"seccompEnabled"`
	SeccompProfilePath  string `json:"seccompProfilePath"`
}

type StoreInfo struct {
	ConfigFile      string         `json:"configFile"`
	ContainerStore  ContainerStore `json:"containerStore"`
	GraphDriverName string         `json:"graphDriverName"`
	GraphRoot       string         `json:"graphRoot"`
	ImageStore      ImageStore     `json:"imageStore"`
	RunRoot         string         `json:"runRoot"`
}

type ContainerStore struct {
	Number  int `json:"number"`
	Paused  int `json:"paused"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
}

type ImageStore struct {
	Number int `json:"number"`
}

type VersionInfo struct {
	APIVersion string `json:"APIVersion"`
	Version    string `json:"Version"`
	GoVersion  string `json:"GoVersion"`
	GitCommit  string `json:"GitCommit"`
	Built      int64  `json:"Built"`
	OsArch     string `json:"OsArch"`
}

type Container struct {
	ID         string            `json:"Id"`
	Names      []string          `json:"Names"`
	Image      string            `json:"Image"`
	ImageID    string            `json:"ImageID"`
	Command    interface{}       `json:"Command"`
	Created    FlexibleTime      `json:"Created"`
	State      string            `json:"State"`
	Status     string            `json:"Status"`
	Ports      []PortMapping     `json:"Ports"`
	Labels     map[string]string `json:"Labels"`
	SizeRw     int64             `json:"SizeRw,omitempty"`
	SizeRootFs int64             `json:"SizeRootFs,omitempty"`
	Mounts     interface{}       `json:"Mounts"`
	Networks   interface{}       `json:"Networks,omitempty"`
}

type PortMapping struct {
	HostIP        string `json:"host_ip,omitempty"`
	HostPort      uint16 `json:"host_port"`
	ContainerPort uint16 `json:"container_port"`
	Protocol      string `json:"protocol"`
	Range         uint16 `json:"range,omitempty"`
}

type MountPoint struct {
	Type        string `json:"Type"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Mode        string `json:"Mode"`
	RW          bool   `json:"RW"`
	Propagation string `json:"Propagation"`
}

type NetworkSettings struct {
	EndpointID  string `json:"EndpointID"`
	Gateway     string `json:"Gateway"`
	IPAddress   string `json:"IPAddress"`
	MacAddress  string `json:"MacAddress"`
	NetworkID   string `json:"NetworkID"`
}

type ContainerInspect struct {
	ID              string            `json:"Id"`
	Created         string            `json:"Created"`
	Path            string            `json:"Path"`
	Args            []string          `json:"Args"`
	State           ContainerState    `json:"State"`
	Image           string            `json:"Image"`
	ImageName       string            `json:"ImageName"`
	Name            string            `json:"Name"`
	RestartCount    int               `json:"RestartCount"`
	MountLabel      string            `json:"MountLabel"`
	ProcessLabel    string            `json:"ProcessLabel"`
	Config          ContainerConfig   `json:"Config"`
	HostConfig      HostConfig        `json:"HostConfig"`
	NetworkSettings NetworkSettingsInspect `json:"NetworkSettings"`
}

type ContainerState struct {
	Status     string     `json:"Status"`
	Running    bool       `json:"Running"`
	Paused     bool       `json:"Paused"`
	Restarting bool       `json:"Restarting"`
	OOMKilled  bool       `json:"OOMKilled"`
	Dead       bool       `json:"Dead"`
	Pid        int        `json:"Pid"`
	ExitCode   int        `json:"ExitCode"`
	Error      string     `json:"Error"`
	StartedAt  string     `json:"StartedAt"`
	FinishedAt string     `json:"FinishedAt"`
	Health     *HealthState `json:"Health,omitempty"`
}

type HealthState struct {
	Status        string      `json:"Status"`
	FailingStreak int         `json:"FailingStreak"`
	Log           []HealthLog `json:"Log,omitempty"`
}

type HealthLog struct {
	Start    time.Time `json:"Start"`
	End      time.Time `json:"End"`
	ExitCode int       `json:"ExitCode"`
	Output   string    `json:"Output"`
}

type ContainerConfig struct {
	Hostname     string              `json:"Hostname"`
	User         string              `json:"User"`
	Env          []string            `json:"Env"`
	Cmd          []string            `json:"Cmd"`
	Image        string              `json:"Image"`
	WorkingDir   string              `json:"WorkingDir"`
	Entrypoint   []string            `json:"Entrypoint"`
	Labels       map[string]string   `json:"Labels"`
	StopSignal   string              `json:"StopSignal"`
	StopTimeout  *uint               `json:"StopTimeout,omitempty"`
	Healthcheck  *HealthConfig       `json:"Healthcheck,omitempty"`
}

type HealthConfig struct {
	Test        []string      `json:"Test"`
	Interval    time.Duration `json:"Interval"`
	Timeout     time.Duration `json:"Timeout"`
	Retries     int           `json:"Retries"`
	StartPeriod time.Duration `json:"StartPeriod"`
}

type HostConfig struct {
	Binds         []string          `json:"Binds"`
	NetworkMode   string            `json:"NetworkMode"`
	PortBindings  map[string][]PortBinding `json:"PortBindings"`
	RestartPolicy RestartPolicy     `json:"RestartPolicy"`
	Memory        int64             `json:"Memory"`
	MemorySwap    int64             `json:"MemorySwap"`
	CPUShares     int64             `json:"CpuShares"`
	CPUQuota      int64             `json:"CpuQuota"`
	CPUPeriod     int64             `json:"CpuPeriod"`
}

type PortBinding struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type RestartPolicy struct {
	Name              string `json:"Name"`
	MaximumRetryCount int    `json:"MaximumRetryCount"`
}

type NetworkSettingsInspect struct {
	Bridge      string                   `json:"Bridge"`
	Ports       map[string][]PortBinding `json:"Ports"`
	Networks    map[string]NetworkEndpoint `json:"Networks"`
}

type NetworkEndpoint struct {
	EndpointID  string   `json:"EndpointID"`
	Gateway     string   `json:"Gateway"`
	IPAddress   string   `json:"IPAddress"`
	MacAddress  string   `json:"MacAddress"`
	NetworkID   string   `json:"NetworkID"`
	Aliases     []string `json:"Aliases"`
}

type ContainerStats struct {
	ContainerID string    `json:"container_id"`
	Name        string    `json:"name"`
	Read        time.Time `json:"read"`
	PreRead     time.Time `json:"preread"`

	CPUStats    CPUStats    `json:"cpu_stats"`
	PreCPUStats CPUStats    `json:"precpu_stats"`
	MemoryStats MemoryStats `json:"memory_stats"`
	NetworkStats map[string]NetworkStats `json:"networks,omitempty"`

	PIDs        PIDStats `json:"pids_stats"`
}

type CPUStats struct {
	CPUUsage       CPUUsage `json:"cpu_usage"`
	SystemCPUUsage uint64   `json:"system_cpu_usage"`
	OnlineCPUs     uint32   `json:"online_cpus"`
}

type CPUUsage struct {
	TotalUsage        uint64   `json:"total_usage"`
	PercpuUsage       []uint64 `json:"percpu_usage"`
	UsageInKernelmode uint64   `json:"usage_in_kernelmode"`
	UsageInUsermode   uint64   `json:"usage_in_usermode"`
}

type MemoryStats struct {
	Usage    uint64 `json:"usage"`
	MaxUsage uint64 `json:"max_usage"`
	Limit    uint64 `json:"limit"`
	Stats    map[string]uint64 `json:"stats"`
}

type NetworkStats struct {
	RxBytes   uint64 `json:"rx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxPackets uint64 `json:"tx_packets"`
	TxErrors  uint64 `json:"tx_errors"`
	TxDropped uint64 `json:"tx_dropped"`
}

type PIDStats struct {
	Current uint64 `json:"current"`
	Limit   uint64 `json:"limit"`
}

type Event struct {
	Type   string            `json:"Type"`
	Action string            `json:"Action"`
	Actor  EventActor        `json:"Actor"`
	Status string            `json:"status,omitempty"`
	Time   int64             `json:"time"`
	TimeNano int64           `json:"timeNano"`
}

type EventActor struct {
	ID         string            `json:"ID"`
	Attributes map[string]string `json:"Attributes"`
}
