package apply

import (
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/plan"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

type Payload struct {
	Workspace   string                  `json:"workspace"`
	Provider    string                  `json:"provider,omitempty"`
	Network     *NetworkPayload         `json:"network,omitempty"`
	Resources   []*ResourcePayload      `json:"resources,omitempty"`
	Diagnostics []runtimepkg.Diagnostic `json:"diagnostics,omitempty"`
	Blocked     bool                    `json:"blocked,omitempty"`
}

type NetworkPayload struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels,omitempty"`
}

type ResourcePayload struct {
	Key           string                        `json:"key"`
	LogicalHost   string                        `json:"logicalHost"`
	RuntimeName   string                        `json:"runtimeName"`
	Source        *runtimepkg.SourceRef         `json:"source,omitempty"`
	Image         string                        `json:"image,omitempty"`
	Build         *BuildPayload                 `json:"build,omitempty"`
	Command       []string                      `json:"command,omitempty"`
	Entrypoint    []string                      `json:"entrypoint,omitempty"`
	WorkingDir    string                        `json:"workingDir,omitempty"`
	DeclaredEnv   map[string]workspace.EnvValue `json:"declaredEnv,omitempty"`
	InjectedEnv   map[string]workspace.EnvValue `json:"injectedEnv,omitempty"`
	Env           map[string]workspace.EnvValue `json:"env,omitempty"`
	Ports         []PortPayload                 `json:"ports,omitempty"`
	Volumes       []VolumePayload               `json:"volumes,omitempty"`
	Health        *workspace.Health             `json:"health,omitempty"`
	ProjectSource *runtimepkg.ProjectSource     `json:"projectSource,omitempty"`
	DevelopWatch  []runtimepkg.WatchRule        `json:"developWatch,omitempty"`
	Labels        map[string]string             `json:"labels,omitempty"`
}

type BuildPayload struct {
	Context    string                        `json:"context"`
	Dockerfile string                        `json:"dockerfile,omitempty"`
	Target     string                        `json:"target,omitempty"`
	Args       map[string]workspace.EnvValue `json:"args,omitempty"`
}

type PortPayload struct {
	Container int    `json:"container"`
	Published int    `json:"published,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	HostIP    string `json:"hostIP,omitempty"`
}

type VolumePayload struct {
	Source   string `json:"source,omitempty"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"readOnly,omitempty"`
	Kind     string `json:"kind,omitempty"`
	Type     string `json:"type,omitempty"`
}

type Result struct {
	Workspace  string               `json:"workspace"`
	Provider   string               `json:"provider,omitempty"`
	StartedAt  time.Time            `json:"startedAt"`
	FinishedAt time.Time            `json:"finishedAt"`
	Operations []Operation          `json:"operations,omitempty"`
	Snapshot   *runtimepkg.Snapshot `json:"snapshot,omitempty"`
}

type Operation struct {
	Scope       plan.ActionScope `json:"scope"`
	Target      string           `json:"target"`
	RuntimeName string           `json:"runtimeName,omitempty"`
	Kind        plan.ActionKind  `json:"kind"`
	Status      string           `json:"status"`
	Message     string           `json:"message,omitempty"`
}

func (p *Payload) Resource(key string) *ResourcePayload {
	if p == nil {
		return nil
	}
	for _, resource := range p.Resources {
		if resource != nil && resource.Key == key {
			return resource
		}
	}
	return nil
}
