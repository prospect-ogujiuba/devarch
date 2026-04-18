package appsvc

import (
	"fmt"

	"github.com/prospect-ogujiuba/devarch/internal/contracts"
	"github.com/prospect-ogujiuba/devarch/internal/importv1"
	"github.com/prospect-ogujiuba/devarch/internal/projectscan"
	"github.com/prospect-ogujiuba/devarch/internal/resolve"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

// TemplateSummary is the API-safe catalog list shape used by Phase 4 surfaces.
type TemplateSummary struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// TemplateDetail is the API-safe catalog detail shape. It intentionally omits
// internal file paths and uses stable JSON field names instead of the raw
// catalog package struct layout.
type TemplateDetail struct {
	APIVersion  string                        `json:"apiVersion,omitempty"`
	Kind        string                        `json:"kind,omitempty"`
	Name        string                        `json:"name"`
	Description string                        `json:"description,omitempty"`
	Tags        []string                      `json:"tags,omitempty"`
	Runtime     map[string]any                `json:"runtime,omitempty"`
	Env         map[string]workspace.EnvValue `json:"env,omitempty"`
	Ports       []workspace.Port              `json:"ports,omitempty"`
	Volumes     []workspace.Volume            `json:"volumes,omitempty"`
	Imports     []workspace.Import            `json:"imports,omitempty"`
	Exports     []workspace.Export            `json:"exports,omitempty"`
	Health      *workspace.Health             `json:"health,omitempty"`
	Develop     map[string]any                `json:"develop,omitempty"`
}

// WorkspaceSummary is the locked Phase 4 list shape for /api/workspaces.
type WorkspaceSummary struct {
	Name          string                         `json:"name"`
	DisplayName   string                         `json:"displayName,omitempty"`
	Description   string                         `json:"description,omitempty"`
	Provider      string                         `json:"provider,omitempty"`
	Capabilities  runtimepkg.AdapterCapabilities `json:"capabilities,omitempty"`
	ResourceCount int                            `json:"resourceCount"`
}

// WorkspaceDetail is the locked Phase 4 detail shape for /api/workspaces/{name}.
type WorkspaceDetail struct {
	Name          string                         `json:"name"`
	DisplayName   string                         `json:"displayName,omitempty"`
	Description   string                         `json:"description,omitempty"`
	Provider      string                         `json:"provider,omitempty"`
	Capabilities  runtimepkg.AdapterCapabilities `json:"capabilities,omitempty"`
	ResourceCount int                            `json:"resourceCount"`
	ManifestPath  string                         `json:"manifestPath"`
	ResourceKeys  []string                       `json:"resourceKeys,omitempty"`
}

// WorkspaceGraphView keeps the graph endpoint transport-thin while still
// returning contract links and diagnostics needed by the UI.
type WorkspaceGraphView struct {
	Graph     *resolve.Graph    `json:"graph"`
	Contracts *contracts.Result `json:"contracts,omitempty"`
}

// WorkspaceStatusView carries the desired runtime boundary alongside the latest
// inspected snapshot for /api/workspaces/{name}/status.
type WorkspaceStatusView struct {
	Desired  *runtimepkg.DesiredWorkspace `json:"desired"`
	Snapshot *runtimepkg.Snapshot         `json:"snapshot,omitempty"`
}

// ProjectScanView is the transport-safe project scan result returned by the
// shared service boundary.
type ProjectScanView = projectscan.Result

// ImportPreview is the transport-safe V1 import result returned by the shared
// service boundary.
type ImportPreview = importv1.Preview

// NotFoundError reports a typed missing service object.
type NotFoundError struct {
	Kind      string
	Name      string
	Workspace string
}

func (e *NotFoundError) Error() string {
	if e == nil {
		return "not found"
	}
	if e.Workspace != "" {
		return fmt.Sprintf("%s %q not found in workspace %q", e.Kind, e.Name, e.Workspace)
	}
	return fmt.Sprintf("%s %q not found", e.Kind, e.Name)
}

// DuplicateWorkspaceNameError reports two discovered workspace manifests with
// the same metadata.name.
type DuplicateWorkspaceNameError struct {
	Name       string
	FirstPath  string
	SecondPath string
}

func (e *DuplicateWorkspaceNameError) Error() string {
	if e == nil {
		return "duplicate workspace name"
	}
	return fmt.Sprintf("duplicate workspace name %q in %s and %s", e.Name, e.FirstPath, e.SecondPath)
}

// UnsupportedCapabilityError reports an operation gated by the selected runtime
// capability surface.
type UnsupportedCapabilityError struct {
	Workspace  string `json:"workspace,omitempty"`
	Resource   string `json:"resource,omitempty"`
	Provider   string `json:"provider,omitempty"`
	Operation  string `json:"operation"`
	Capability string `json:"capability"`
	Reason     string `json:"reason,omitempty"`
}

func (e *UnsupportedCapabilityError) Error() string {
	if e == nil {
		return "unsupported capability"
	}
	prefix := ""
	if e.Workspace != "" {
		prefix = fmt.Sprintf("workspace %q", e.Workspace)
		if e.Resource != "" {
			prefix += fmt.Sprintf(" resource %q", e.Resource)
		}
		prefix += ": "
	}
	if e.Reason == "" {
		return fmt.Sprintf("%sprovider %q does not support capability %q for %s", prefix, e.Provider, e.Capability, e.Operation)
	}
	return fmt.Sprintf("%sprovider %q does not support capability %q for %s: %s", prefix, e.Provider, e.Capability, e.Operation, e.Reason)
}
