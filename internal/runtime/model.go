package runtime

import (
	"strings"
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

const (
	ProviderAuto   = "auto"
	ProviderDocker = "docker"
	ProviderPodman = "podman"

	SeverityWarning = "warning"
	SeverityError   = "error"
)

// DesiredWorkspace is the Phase 3 runtime-owned desired-state boundary derived
// from the stable Phase 2 resolve/contracts output.
type DesiredWorkspace struct {
	Name           string              `json:"name"`
	DisplayName    string              `json:"displayName,omitempty"`
	Description    string              `json:"description,omitempty"`
	Provider       string              `json:"provider,omitempty"`
	NamingStrategy string              `json:"namingStrategy,omitempty"`
	ManifestPath   string              `json:"-"`
	ManifestDir    string              `json:"-"`
	Network        *DesiredNetwork     `json:"network,omitempty"`
	Resources      []*DesiredResource  `json:"resources,omitempty"`
	Diagnostics    []Diagnostic        `json:"diagnostics,omitempty"`
	Capabilities   AdapterCapabilities `json:"capabilities,omitempty"`
}

type DesiredNetwork struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels,omitempty"`
}

type DesiredResource struct {
	Key            string                        `json:"key"`
	Enabled        bool                          `json:"enabled"`
	LogicalHost    string                        `json:"logicalHost"`
	RuntimeName    string                        `json:"runtimeName"`
	TemplateName   string                        `json:"templateName,omitempty"`
	Source         *SourceRef                    `json:"source,omitempty"`
	DeclaredEnv    map[string]workspace.EnvValue `json:"declaredEnv,omitempty"`
	InjectedEnv    map[string]workspace.EnvValue `json:"injectedEnv,omitempty"`
	DependsOn      []string                      `json:"dependsOn,omitempty"`
	Domains        []string                      `json:"domains,omitempty"`
	OverrideLabels map[string]string             `json:"overrideLabels,omitempty"`
	Diagnostics    []Diagnostic                  `json:"diagnostics,omitempty"`
	Spec           ResourceSpec                  `json:"spec"`
}

type SourceRef struct {
	Type         string `json:"type"`
	Path         string `json:"path"`
	Service      string `json:"service,omitempty"`
	ResolvedPath string `json:"-"`
}

type ResourceSpec struct {
	Image         string                        `json:"image,omitempty"`
	Build         *BuildSpec                    `json:"build,omitempty"`
	Command       []string                      `json:"command,omitempty"`
	Entrypoint    []string                      `json:"entrypoint,omitempty"`
	WorkingDir    string                        `json:"workingDir,omitempty"`
	Env           map[string]workspace.EnvValue `json:"env,omitempty"`
	Ports         []PortSpec                    `json:"ports,omitempty"`
	Volumes       []VolumeSpec                  `json:"volumes,omitempty"`
	Health        *workspace.Health             `json:"health,omitempty"`
	ProjectSource *ProjectSource                `json:"projectSource,omitempty"`
	DevelopWatch  []WatchRule                   `json:"developWatch,omitempty"`
	Labels        map[string]string             `json:"labels,omitempty"`
}

type BuildSpec struct {
	Context            string                        `json:"context"`
	Dockerfile         string                        `json:"dockerfile,omitempty"`
	Target             string                        `json:"target,omitempty"`
	Args               map[string]workspace.EnvValue `json:"args,omitempty"`
	ResolvedContext    string                        `json:"-"`
	ResolvedDockerfile string                        `json:"-"`
}

type ProjectSource struct {
	HostPath      string `json:"hostPath"`
	ContainerPath string `json:"containerPath"`
}

type PortSpec struct {
	Container int    `json:"container"`
	Published int    `json:"published,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	HostIP    string `json:"hostIP,omitempty"`
}

type VolumeSpec struct {
	Source   string `json:"source,omitempty"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"readOnly,omitempty"`
	Kind     string `json:"kind,omitempty"`
	Type     string `json:"type,omitempty"`
}

type WatchRule struct {
	Path         string `json:"path"`
	ResolvedPath string `json:"resolvedPath,omitempty"`
	Target       string `json:"target"`
	Action       string `json:"action,omitempty"`
}

type Diagnostic struct {
	Severity  string   `json:"severity"`
	Code      string   `json:"code"`
	Workspace string   `json:"workspace,omitempty"`
	Resource  string   `json:"resource,omitempty"`
	Contract  string   `json:"contract,omitempty"`
	Provider  string   `json:"provider,omitempty"`
	Providers []string `json:"providers,omitempty"`
	EnvKey    string   `json:"envKey,omitempty"`
	Message   string   `json:"message"`
}

// Snapshot is the runtime-owned observed-state boundary consumed by the planner.
type Snapshot struct {
	Workspace SnapshotWorkspace   `json:"workspace"`
	Resources []*SnapshotResource `json:"resources,omitempty"`
}

type SnapshotWorkspace struct {
	Name     string           `json:"name"`
	Provider string           `json:"provider,omitempty"`
	Network  *SnapshotNetwork `json:"network,omitempty"`
}

type SnapshotNetwork struct {
	Name   string            `json:"name"`
	ID     string            `json:"id,omitempty"`
	Driver string            `json:"driver,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

type SnapshotResource struct {
	Key         string        `json:"key"`
	RuntimeName string        `json:"runtimeName"`
	LogicalHost string        `json:"logicalHost,omitempty"`
	ID          string        `json:"id,omitempty"`
	State       ResourceState `json:"state,omitempty"`
	Spec        ResourceSpec  `json:"spec"`
}

type ResourceState struct {
	Status       string     `json:"status,omitempty"`
	Running      bool       `json:"running,omitempty"`
	Health       string     `json:"health,omitempty"`
	ExitCode     int        `json:"exitCode,omitempty"`
	RestartCount int        `json:"restartCount,omitempty"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
	Error        string     `json:"error,omitempty"`
}

func (w *DesiredWorkspace) Blocked() bool {
	for _, diagnostic := range w.Diagnostics {
		if diagnostic.BlocksApply() {
			return true
		}
	}
	for _, resource := range w.Resources {
		if resource != nil && resource.Blocked() {
			return true
		}
	}
	return false
}

func (w *DesiredWorkspace) Resource(key string) *DesiredResource {
	if w == nil {
		return nil
	}
	for _, resource := range w.Resources {
		if resource != nil && resource.Key == key {
			return resource
		}
	}
	return nil
}

func (w *DesiredWorkspace) BlockingDiagnostics() []Diagnostic {
	if w == nil {
		return nil
	}
	blocked := make([]Diagnostic, 0)
	for _, diagnostic := range w.Diagnostics {
		if diagnostic.Blocking() {
			blocked = append(blocked, diagnostic)
		}
	}
	for _, resource := range w.Resources {
		if resource == nil {
			continue
		}
		for _, diagnostic := range resource.Diagnostics {
			if diagnostic.Blocking() {
				blocked = append(blocked, diagnostic)
			}
		}
	}
	if len(blocked) == 0 {
		return nil
	}
	return blocked
}

func (r *DesiredResource) Blocked() bool {
	if r == nil {
		return false
	}
	for _, diagnostic := range r.Diagnostics {
		if diagnostic.BlocksApply() {
			return true
		}
	}
	return false
}

func (r *DesiredResource) EffectiveEnv() map[string]workspace.EnvValue {
	if r == nil {
		return nil
	}
	merged := cloneEnvMap(r.InjectedEnv)
	if merged == nil {
		merged = make(map[string]workspace.EnvValue, len(r.DeclaredEnv))
	}
	for key, value := range r.DeclaredEnv {
		merged[key] = value.Clone()
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func (r *DesiredResource) EffectiveLabels() map[string]string {
	if r == nil {
		return nil
	}
	return cloneStringMap(r.Spec.Labels)
}

func (s *Snapshot) Resource(key string) *SnapshotResource {
	if s == nil {
		return nil
	}
	for _, resource := range s.Resources {
		if resource != nil && resource.Key == key {
			return resource
		}
	}
	return nil
}

func (d Diagnostic) Blocking() bool {
	return strings.EqualFold(d.Severity, SeverityError)
}

// BlocksApply is intentionally narrower than severity alone. Phase 3 must carry
// contract diagnostics forward, but not every error-level contract diagnostic is
// a hard stop for the current runtime/apply surface. Unsupported runtime-owned
// diagnostics still block side effects explicitly.
func (d Diagnostic) BlocksApply() bool {
	if !d.Blocking() {
		return false
	}
	switch d.Code {
	case "secret-flatten":
		return false
	default:
		return true
	}
}

func cloneEnvMap(values map[string]workspace.EnvValue) map[string]workspace.EnvValue {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]workspace.EnvValue, len(values))
	for key, value := range values {
		cloned[key] = value.Clone()
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func cloneHealth(health *workspace.Health) *workspace.Health {
	if health == nil {
		return nil
	}
	cloned := *health
	if len(health.Test) > 0 {
		cloned.Test = append(workspace.StringList(nil), health.Test...)
	}
	return &cloned
}

func clonePorts(values []PortSpec) []PortSpec {
	if len(values) == 0 {
		return nil
	}
	return append([]PortSpec(nil), values...)
}

func cloneVolumes(values []VolumeSpec) []VolumeSpec {
	if len(values) == 0 {
		return nil
	}
	return append([]VolumeSpec(nil), values...)
}

func cloneWatchRules(values []WatchRule) []WatchRule {
	if len(values) == 0 {
		return nil
	}
	return append([]WatchRule(nil), values...)
}

func cloneProjectSource(source *ProjectSource) *ProjectSource {
	if source == nil {
		return nil
	}
	cloned := *source
	return &cloned
}

func cloneBuildSpec(build *BuildSpec) *BuildSpec {
	if build == nil {
		return nil
	}
	cloned := *build
	cloned.Args = cloneEnvMap(build.Args)
	return &cloned
}

func (s ResourceSpec) Clone() ResourceSpec {
	return ResourceSpec{
		Image:         s.Image,
		Build:         cloneBuildSpec(s.Build),
		Command:       cloneStringSlice(s.Command),
		Entrypoint:    cloneStringSlice(s.Entrypoint),
		WorkingDir:    s.WorkingDir,
		Env:           cloneEnvMap(s.Env),
		Ports:         clonePorts(s.Ports),
		Volumes:       cloneVolumes(s.Volumes),
		Health:        cloneHealth(s.Health),
		ProjectSource: cloneProjectSource(s.ProjectSource),
		DevelopWatch:  cloneWatchRules(s.DevelopWatch),
		Labels:        cloneStringMap(s.Labels),
	}
}
