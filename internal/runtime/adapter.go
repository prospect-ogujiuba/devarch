package runtime

import (
	"context"
	"time"
)

// AdapterCapabilities reports which Phase 3 runtime surfaces a provider can
// satisfy without widening scope into live integration requirements.
type AdapterCapabilities struct {
	Inspect bool `json:"inspect,omitempty"`
	Apply   bool `json:"apply,omitempty"`
	Logs    bool `json:"logs,omitempty"`
	Exec    bool `json:"exec,omitempty"`
	Network bool `json:"network,omitempty"`
}

type ResourceRef struct {
	Workspace   string `json:"workspace"`
	Key         string `json:"key"`
	RuntimeName string `json:"runtimeName"`
}

type ApplyResourceRequest struct {
	Workspace   string          `json:"workspace"`
	NetworkName string          `json:"networkName,omitempty"`
	Resource    AppliedResource `json:"resource"`
}

type AppliedResource struct {
	Key         string       `json:"key"`
	LogicalHost string       `json:"logicalHost,omitempty"`
	RuntimeName string       `json:"runtimeName"`
	Spec        ResourceSpec `json:"spec"`
}

type LogsRequest struct {
	Tail   int        `json:"tail,omitempty"`
	Follow bool       `json:"follow,omitempty"`
	Since  *time.Time `json:"since,omitempty"`
}

type LogChunk struct {
	Timestamp *time.Time `json:"timestamp,omitempty"`
	Stream    string     `json:"stream,omitempty"`
	Line      string     `json:"line"`
}

type LogsConsumer func(LogChunk) error

type ExecRequest struct {
	Command     []string `json:"command"`
	Interactive bool     `json:"interactive,omitempty"`
	TTY         bool     `json:"tty,omitempty"`
}

type ExecResult struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
}

// Adapter is the common Phase 3 runtime seam for desired/snapshot inspection,
// apply primitives, logs, and exec.
type Adapter interface {
	Provider() string
	Capabilities() AdapterCapabilities
	InspectWorkspace(ctx context.Context, desired *DesiredWorkspace) (*Snapshot, error)
	EnsureNetwork(ctx context.Context, network *DesiredNetwork) error
	RemoveNetwork(ctx context.Context, network *DesiredNetwork) error
	ApplyResource(ctx context.Context, request ApplyResourceRequest) error
	RemoveResource(ctx context.Context, resource ResourceRef) error
	RestartResource(ctx context.Context, resource ResourceRef) error
	StreamLogs(ctx context.Context, resource ResourceRef, request LogsRequest, consume LogsConsumer) error
	Exec(ctx context.Context, resource ResourceRef, request ExecRequest) (*ExecResult, error)
}

// CommandRunner allows Docker and Podman adapters to be tested deterministically
// without requiring a live daemon.
type CommandRunner interface {
	Run(ctx context.Context, command string, args ...string) ([]byte, error)
}
