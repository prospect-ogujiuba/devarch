package resolve

import "github.com/prospect-ogujiuba/devarch/internal/workspace"

// Graph is the deterministic Phase 2 effective graph shared by resolve,
// contracts, and later planning work. It intentionally omits runtime snapshot,
// apply state, and compose-materialized data. raw-compose resources remain
// pass-through source references in this phase.
type Graph struct {
	Workspace Workspace   `json:"workspace"`
	Resources []*Resource `json:"resources"`
}

// Workspace captures the stable workspace metadata needed by later phases while
// keeping manifest-local absolute paths out of serialized output.
type Workspace struct {
	Name           string                       `json:"name"`
	DisplayName    string                       `json:"displayName,omitempty"`
	Description    string                       `json:"description,omitempty"`
	Runtime        workspace.RuntimePreferences `json:"runtime,omitempty"`
	Policies       workspace.Policies           `json:"policies,omitempty"`
	CatalogSources []string                     `json:"catalogSources,omitempty"`

	ManifestPath string `json:"-"`
	ManifestDir  string `json:"-"`
}

// Resource is one resolved workspace resource in deterministic key order.
type Resource struct {
	Key       string              `json:"key"`
	Enabled   bool                `json:"enabled"`
	Host      string              `json:"host"`
	Template  *TemplateRef        `json:"template,omitempty"`
	Source    *SourceRef          `json:"source,omitempty"`
	Runtime   *Runtime            `json:"runtime,omitempty"`
	Env       map[string]EnvValue `json:"env,omitempty"`
	Ports     []Port              `json:"ports,omitempty"`
	Volumes   []Volume            `json:"volumes,omitempty"`
	DependsOn []string            `json:"dependsOn,omitempty"`
	Imports   []Import            `json:"imports,omitempty"`
	Exports   []Export            `json:"exports,omitempty"`
	Health    *Health             `json:"health,omitempty"`
	Domains   []string            `json:"domains,omitempty"`
	Develop   map[string]any      `json:"develop,omitempty"`
	Overrides map[string]any      `json:"overrides,omitempty"`
}

type TemplateRef struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`

	ResolvedPath string `json:"-"`
}

type SourceRef struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	Service string `json:"service,omitempty"`

	ResolvedPath string `json:"-"`
}

type Runtime struct {
	Image      string     `json:"image,omitempty"`
	Build      *Build     `json:"build,omitempty"`
	Command    StringList `json:"command,omitempty"`
	Entrypoint StringList `json:"entrypoint,omitempty"`
	WorkingDir string     `json:"workingDir,omitempty"`
}

type Build struct {
	Context    string              `json:"context"`
	Dockerfile string              `json:"dockerfile,omitempty"`
	Target     string              `json:"target,omitempty"`
	Args       map[string]EnvValue `json:"args,omitempty"`

	ResolvedContext    string `json:"-"`
	ResolvedDockerfile string `json:"-"`
}

type StringList = workspace.StringList

type EnvValue = workspace.EnvValue

type Port = workspace.Port

type Volume = workspace.Volume

type Import = workspace.Import

type Export = workspace.Export

type Health = workspace.Health

func (g *Graph) Resource(key string) *Resource {
	if g == nil {
		return nil
	}
	for _, resource := range g.Resources {
		if resource != nil && resource.Key == key {
			return resource
		}
	}
	return nil
}
