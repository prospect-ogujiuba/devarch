package appsvc

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/apply"
	cachepkg "github.com/prospect-ogujiuba/devarch/internal/cache"
	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	contractspkg "github.com/prospect-ogujiuba/devarch/internal/contracts"
	"github.com/prospect-ogujiuba/devarch/internal/events"
	"github.com/prospect-ogujiuba/devarch/internal/importv1"
	planpkg "github.com/prospect-ogujiuba/devarch/internal/plan"
	"github.com/prospect-ogujiuba/devarch/internal/projectscan"
	resolvepkg "github.com/prospect-ogujiuba/devarch/internal/resolve"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	dockeradapter "github.com/prospect-ogujiuba/devarch/internal/runtime/docker"
	podmanadapter "github.com/prospect-ogujiuba/devarch/internal/runtime/podman"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
	"gopkg.in/yaml.v3"
)

// Config wires the shared Phase 4 service boundary without exposing transport
// concerns.
type Config struct {
	WorkspaceRoots []string
	CatalogRoots   []string
	Adapters       map[string]runtimepkg.Adapter
	EventBus       *events.Bus
	Cache          cachepkg.Store
	LookPath       func(string) (string, error)
}

// Service is the narrow shared seam consumed by Phase 4 transports.
type Service struct {
	workspaceRoots []string
	catalogRoots   []string
	adapters       map[string]runtimepkg.Adapter
	bus            *events.Bus
	cache          cachepkg.Store
	lookPath       func(string) (string, error)
}

type workspaceState struct {
	Workspace *workspace.Workspace
	Graph     *resolvepkg.Graph
	Contracts *contractspkg.Result
	Desired   *runtimepkg.DesiredWorkspace
	Adapter   runtimepkg.Adapter
}

func New(config Config) (*Service, error) {
	service := &Service{
		workspaceRoots: append([]string(nil), config.WorkspaceRoots...),
		catalogRoots:   append([]string(nil), config.CatalogRoots...),
		adapters:       cloneAdapters(config.Adapters),
		bus:            config.EventBus,
		cache:          config.Cache,
		lookPath:       config.LookPath,
	}
	if len(service.adapters) == 0 {
		service.adapters = defaultAdapters()
	}
	if service.bus == nil {
		service.bus = events.NewBus()
	}
	if service.lookPath == nil {
		service.lookPath = exec.LookPath
	}

	if _, err := DiscoverWorkspaces(service.workspaceRoots); err != nil {
		return nil, err
	}
	if _, err := LoadCatalogIndex(service.catalogRoots); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *Service) CatalogTemplates(context.Context) ([]TemplateSummary, error) {
	index, err := LoadCatalogIndex(s.catalogRoots)
	if err != nil {
		return nil, err
	}
	templates := index.Templates()
	summaries := make([]TemplateSummary, 0, len(templates))
	for _, template := range templates {
		if template == nil {
			continue
		}
		summaries = append(summaries, TemplateSummary{
			Name:        template.Metadata.Name,
			Description: template.Metadata.Description,
			Tags:        append([]string(nil), template.Metadata.Tags...),
		})
	}
	return summaries, nil
}

func (s *Service) Workspaces(context.Context) ([]WorkspaceSummary, error) {
	workspaces, err := DiscoverWorkspaces(s.workspaceRoots)
	if err != nil {
		return nil, err
	}
	summaries := make([]WorkspaceSummary, 0, len(workspaces))
	for _, ws := range workspaces {
		provider, capabilities := s.describeProvider(ws.Runtime.Provider)
		summaries = append(summaries, WorkspaceSummary{
			Name:          ws.Metadata.Name,
			DisplayName:   ws.Metadata.DisplayName,
			Description:   ws.Metadata.Description,
			Provider:      provider,
			Capabilities:  capabilities,
			ResourceCount: len(ws.Resources),
		})
	}
	return summaries, nil
}

func (s *Service) WorkspaceManifest(_ context.Context, name string) (*workspace.Workspace, error) {
	ws, err := s.loadWorkspace(name)
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *Service) WorkspaceGraph(_ context.Context, name string) (*WorkspaceGraphView, error) {
	state, err := s.loadWorkspaceState(name)
	if err != nil {
		return nil, err
	}
	return &WorkspaceGraphView{Graph: state.Graph, Contracts: state.Contracts}, nil
}

func (s *Service) WorkspaceStatus(ctx context.Context, name string) (*WorkspaceStatusView, error) {
	state, err := s.loadRuntimeState(name, "status")
	if err != nil {
		return nil, err
	}
	if !state.Desired.Capabilities.Inspect {
		return nil, unsupportedCapability(name, "", state.Desired.Provider, "status", "inspect", "selected runtime does not support workspace inspection")
	}
	snapshot, err := state.Adapter.InspectWorkspace(ctx, state.Desired)
	if err != nil {
		return nil, err
	}
	s.saveSnapshot(ctx, state.Desired.Name, snapshot)
	return &WorkspaceStatusView{Desired: state.Desired, Snapshot: snapshot}, nil
}

func (s *Service) WorkspacePlan(ctx context.Context, name string) (*planpkg.Result, error) {
	state, err := s.loadWorkspaceState(name)
	if err != nil {
		return nil, err
	}
	adapter, provider, capabilities := s.planProvider(state.Desired.Provider)
	state.Adapter = adapter
	state.Desired.Provider = provider
	state.Desired.Capabilities = capabilities

	var snapshot *runtimepkg.Snapshot
	var warning *runtimepkg.Diagnostic
	if adapter == nil {
		warning = inspectWarning(name, provider, fmt.Sprintf("runtime provider %q is unavailable; planning against an empty runtime snapshot", provider))
	} else if !capabilities.Inspect {
		warning = inspectWarning(name, provider, fmt.Sprintf("runtime provider %q does not support inspection; planning against an empty runtime snapshot", provider))
	} else {
		snapshot, err = adapter.InspectWorkspace(ctx, state.Desired)
		if err != nil {
			warning = inspectWarning(name, provider, fmt.Sprintf("runtime inspection failed; planning against an empty runtime snapshot: %v", err))
			snapshot = nil
		} else {
			s.saveSnapshot(ctx, state.Desired.Name, snapshot)
		}
	}

	result, err := planpkg.Diff(state.Desired, snapshot)
	if err != nil {
		return nil, err
	}
	if warning != nil {
		result.Diagnostics = append(result.Diagnostics, *warning)
	}
	return result, nil
}

func (s *Service) ApplyWorkspace(ctx context.Context, name string) (*apply.Result, error) {
	state, err := s.loadRuntimeState(name, "apply")
	if err != nil {
		return nil, err
	}
	if !state.Desired.Capabilities.Inspect {
		return nil, unsupportedCapability(name, "", state.Desired.Provider, "apply", "inspect", "selected runtime does not support workspace inspection")
	}
	snapshot, err := state.Adapter.InspectWorkspace(ctx, state.Desired)
	if err != nil {
		return nil, err
	}
	diff, err := planpkg.Diff(state.Desired, snapshot)
	if err != nil {
		return nil, err
	}
	if err := ensureApplyCapabilities(name, state.Desired.Provider, state.Desired.Capabilities, diff); err != nil {
		return nil, err
	}
	payload, err := apply.Render(state.Desired)
	if err != nil {
		return nil, err
	}
	executor := &apply.Executor{Adapter: state.Adapter, Cache: s.cache, Publisher: s.bus}
	return executor.Execute(ctx, diff, payload)
}

func (s *Service) WorkspaceLogs(ctx context.Context, name, resource string, request runtimepkg.LogsRequest) ([]runtimepkg.LogChunk, error) {
	resource = strings.TrimSpace(resource)
	if resource == "" {
		return nil, fmt.Errorf("resource is required")
	}
	state, err := s.loadRuntimeState(name, "logs")
	if err != nil {
		return nil, err
	}
	item := state.Desired.Resource(resource)
	if item == nil {
		return nil, &NotFoundError{Kind: "resource", Name: resource, Workspace: name}
	}
	if !state.Desired.Capabilities.Logs {
		return nil, unsupportedCapability(name, resource, state.Desired.Provider, "logs", "logs", "selected runtime does not support log streaming")
	}
	ref := runtimepkg.ResourceRef{Workspace: state.Desired.Name, Key: item.Key, RuntimeName: item.RuntimeName}
	if s.bus != nil {
		if _, err := s.bus.Publish(events.LogsStarted(ref.Workspace, ref.Key, request.Tail, request.Follow)); err != nil {
			return nil, err
		}
	}
	chunks := make([]runtimepkg.LogChunk, 0)
	err = state.Adapter.StreamLogs(ctx, ref, request, func(chunk runtimepkg.LogChunk) error {
		chunks = append(chunks, chunk)
		if s.bus != nil {
			_, err := s.bus.Publish(events.LogsChunk(ref.Workspace, ref.Key, chunk.Stream, chunk.Line, chunk.Timestamp))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.bus != nil {
		if _, err := s.bus.Publish(events.LogsCompleted(ref.Workspace, ref.Key, request.Tail, request.Follow)); err != nil {
			return nil, err
		}
	}
	return chunks, nil
}

func (s *Service) ExecWorkspace(ctx context.Context, name, resource string, request runtimepkg.ExecRequest) (*runtimepkg.ExecResult, error) {
	resource = strings.TrimSpace(resource)
	if resource == "" {
		return nil, fmt.Errorf("resource is required")
	}
	if request.Interactive || request.TTY {
		return nil, unsupportedCapability(name, resource, "", "exec", "interactive", "interactive and tty exec are not supported in Phase 4")
	}
	state, err := s.loadRuntimeState(name, "exec")
	if err != nil {
		return nil, err
	}
	item := state.Desired.Resource(resource)
	if item == nil {
		return nil, &NotFoundError{Kind: "resource", Name: resource, Workspace: name}
	}
	if !state.Desired.Capabilities.Exec {
		return nil, unsupportedCapability(name, resource, state.Desired.Provider, "exec", "exec", "selected runtime does not support exec")
	}
	ref := runtimepkg.ResourceRef{Workspace: state.Desired.Name, Key: item.Key, RuntimeName: item.RuntimeName}
	return runtimepkg.ExecWithEvents(ctx, state.Adapter, s.bus, ref, request)
}

func (s *Service) SubscribeWorkspaceEvents(ctx context.Context, name string, buffer int) (<-chan events.Envelope, func(), error) {
	if _, err := s.loadWorkspace(name); err != nil {
		return nil, nil, err
	}
	if buffer <= 0 {
		buffer = 1
	}
	source, unsubscribe := s.bus.Subscribe(buffer)
	filtered := make(chan events.Envelope, buffer)
	stop := make(chan struct{})
	var once sync.Once
	cancel := func() {
		once.Do(func() {
			close(stop)
			unsubscribe()
		})
	}
	go func() {
		defer close(filtered)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case envelope, ok := <-source:
				if !ok {
					return
				}
				if envelope.Workspace != name {
					continue
				}
				select {
				case filtered <- envelope:
				case <-ctx.Done():
					return
				case <-stop:
					return
				}
			}
		}
	}()
	return filtered, cancel, nil
}

func (s *Service) CatalogTemplate(_ context.Context, name string) (*TemplateDetail, error) {
	index, err := LoadCatalogIndex(s.catalogRoots)
	if err != nil {
		return nil, err
	}
	template, ok := index.ByName(name)
	if !ok {
		return nil, &NotFoundError{Kind: "template", Name: name}
	}
	return templateDetailFromCatalog(template)
}

func (s *Service) ImportV1Stack(_ context.Context, path string) (*ImportPreview, error) {
	return importv1.PrepareStackImport(path)
}

func (s *Service) ImportV1Library(_ context.Context, path string) (*ImportPreview, error) {
	return importv1.PrepareLibraryImport(path)
}

func (s *Service) ScanProject(_ context.Context, path string) (*ProjectScanView, error) {
	return projectscan.Scan(path)
}

func (s *Service) Workspace(_ context.Context, name string) (*WorkspaceDetail, error) {
	ws, err := s.loadWorkspace(name)
	if err != nil {
		return nil, err
	}
	provider, capabilities := s.describeProvider(ws.Runtime.Provider)
	return &WorkspaceDetail{
		Name:          ws.Metadata.Name,
		DisplayName:   ws.Metadata.DisplayName,
		Description:   ws.Metadata.Description,
		Provider:      provider,
		Capabilities:  capabilities,
		ResourceCount: len(ws.Resources),
		ManifestPath:  ws.ManifestPath,
		ResourceKeys:  ws.SortedResourceKeys(),
	}, nil
}

func (s *Service) loadWorkspace(name string) (*workspace.Workspace, error) {
	workspaces, err := DiscoverWorkspaces(s.workspaceRoots)
	if err != nil {
		return nil, err
	}
	for _, ws := range workspaces {
		if ws != nil && ws.Metadata.Name == name {
			return ws, nil
		}
	}
	return nil, &NotFoundError{Kind: "workspace", Name: name}
}

func (s *Service) loadWorkspaceState(name string) (*workspaceState, error) {
	ws, err := s.loadWorkspace(name)
	if err != nil {
		return nil, err
	}
	paths, err := catalog.DiscoverTemplateFiles(ws.ResolvedCatalogSources())
	if err != nil {
		return nil, err
	}
	index, err := catalog.LoadIndex(paths)
	if err != nil {
		return nil, err
	}
	graph, err := resolvepkg.Resolve(ws, index)
	if err != nil {
		return nil, err
	}
	contractResult := contractspkg.Resolve(graph)
	desired, err := runtimepkg.BuildDesiredWorkspace(graph, contractResult)
	if err != nil {
		return nil, err
	}
	return &workspaceState{Workspace: ws, Graph: graph, Contracts: contractResult, Desired: desired}, nil
}

func (s *Service) loadRuntimeState(name, operation string) (*workspaceState, error) {
	state, err := s.loadWorkspaceState(name)
	if err != nil {
		return nil, err
	}
	adapter, provider, capabilities, err := s.requireProvider(state.Desired.Provider, name, operation)
	if err != nil {
		return nil, err
	}
	state.Adapter = adapter
	state.Desired.Provider = provider
	state.Desired.Capabilities = capabilities
	return state, nil
}

func (s *Service) describeProvider(provider string) (string, runtimepkg.AdapterCapabilities) {
	adapter, resolvedProvider, capabilities, err := s.resolveProvider(normalizeProvider(provider), false)
	if err != nil || adapter == nil {
		return normalizeProvider(provider), runtimepkg.AdapterCapabilities{}
	}
	return resolvedProvider, capabilities
}

func (s *Service) planProvider(provider string) (runtimepkg.Adapter, string, runtimepkg.AdapterCapabilities) {
	adapter, resolvedProvider, capabilities, err := s.resolveProvider(normalizeProvider(provider), false)
	if err != nil {
		return nil, normalizeProvider(provider), runtimepkg.AdapterCapabilities{}
	}
	if adapter == nil {
		return nil, normalizeProvider(provider), runtimepkg.AdapterCapabilities{}
	}
	return adapter, resolvedProvider, capabilities
}

func (s *Service) requireProvider(provider, workspaceName, operation string) (runtimepkg.Adapter, string, runtimepkg.AdapterCapabilities, error) {
	return s.resolveProvider(normalizeProvider(provider), true, workspaceName, operation)
}

func (s *Service) resolveProvider(provider string, strict bool, details ...string) (runtimepkg.Adapter, string, runtimepkg.AdapterCapabilities, error) {
	switch provider {
	case "", runtimepkg.ProviderAuto:
		for _, candidate := range []string{runtimepkg.ProviderDocker, runtimepkg.ProviderPodman} {
			adapter, ok := s.adapters[candidate]
			if !ok || adapter == nil {
				continue
			}
			if s.adapterAvailable(candidate) {
				return adapter, adapter.Provider(), adapter.Capabilities(), nil
			}
		}
		if !strict {
			return nil, runtimepkg.ProviderAuto, runtimepkg.AdapterCapabilities{}, nil
		}
		workspaceName, operation := detailPair(details)
		return nil, runtimepkg.ProviderAuto, runtimepkg.AdapterCapabilities{}, unsupportedCapability(workspaceName, "", runtimepkg.ProviderAuto, operation, "provider", "no available runtime found for auto provider (order: docker, podman)")
	case runtimepkg.ProviderDocker, runtimepkg.ProviderPodman:
		adapter, ok := s.adapters[provider]
		if !ok || adapter == nil {
			if strict {
				workspaceName, operation := detailPair(details)
				return nil, provider, runtimepkg.AdapterCapabilities{}, unsupportedCapability(workspaceName, "", provider, operation, "provider", "runtime adapter is not configured")
			}
			return nil, provider, runtimepkg.AdapterCapabilities{}, nil
		}
		if strict && !s.adapterAvailable(provider) {
			workspaceName, operation := detailPair(details)
			return nil, provider, runtimepkg.AdapterCapabilities{}, unsupportedCapability(workspaceName, "", provider, operation, "provider", "runtime binary is not available on PATH")
		}
		if !strict && !s.adapterAvailable(provider) {
			return nil, provider, runtimepkg.AdapterCapabilities{}, nil
		}
		return adapter, adapter.Provider(), adapter.Capabilities(), nil
	default:
		if strict {
			workspaceName, operation := detailPair(details)
			return nil, provider, runtimepkg.AdapterCapabilities{}, unsupportedCapability(workspaceName, "", provider, operation, "provider", "unknown runtime provider")
		}
		return nil, provider, runtimepkg.AdapterCapabilities{}, nil
	}
}

func (s *Service) adapterAvailable(provider string) bool {
	adapter := s.adapters[provider]
	if adapter == nil {
		return false
	}
	if s.lookPath == nil {
		return true
	}
	_, err := s.lookPath(provider)
	return err == nil
}

func (s *Service) saveSnapshot(ctx context.Context, workspaceName string, snapshot *runtimepkg.Snapshot) {
	if snapshot == nil || s.cache == nil {
		return
	}
	_ = s.cache.SaveSnapshot(ctx, cachepkg.SnapshotRecord{
		Workspace:  workspaceName,
		CapturedAt: time.Now(),
		Snapshot:   snapshot,
	})
}

func defaultAdapters() map[string]runtimepkg.Adapter {
	return map[string]runtimepkg.Adapter{
		runtimepkg.ProviderDocker: dockeradapter.New(nil),
		runtimepkg.ProviderPodman: podmanadapter.New(nil),
	}
}

func cloneAdapters(adapters map[string]runtimepkg.Adapter) map[string]runtimepkg.Adapter {
	if len(adapters) == 0 {
		return nil
	}
	cloned := make(map[string]runtimepkg.Adapter, len(adapters))
	for key, adapter := range adapters {
		cloned[strings.ToLower(strings.TrimSpace(key))] = adapter
	}
	return cloned
}

func templateDetailFromCatalog(template *catalog.Template) (*TemplateDetail, error) {
	if template == nil {
		return nil, fmt.Errorf("nil catalog template")
	}
	env, err := templateEnv(template.Spec.Env)
	if err != nil {
		return nil, err
	}
	health, err := templateHealth(template.Spec.Health)
	if err != nil {
		return nil, err
	}
	return &TemplateDetail{
		APIVersion:  template.APIVersion,
		Kind:        template.Kind,
		Name:        template.Metadata.Name,
		Description: template.Metadata.Description,
		Tags:        append([]string(nil), template.Metadata.Tags...),
		Runtime:     cloneMap(template.Spec.Runtime),
		Env:         env,
		Ports:       templatePorts(template.Spec.Ports),
		Volumes:     templateVolumes(template.Spec.Volumes),
		Imports:     templateImports(template.Spec.Imports),
		Exports:     templateExports(template.Spec.Exports),
		Health:      health,
		Develop:     cloneMap(template.Spec.Develop),
	}, nil
}

func templateEnv(values map[string]any) (map[string]workspace.EnvValue, error) {
	if len(values) == 0 {
		return nil, nil
	}
	env := make(map[string]workspace.EnvValue, len(values))
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value, err := workspace.EnvValueFromAny(values[key])
		if err != nil {
			return nil, fmt.Errorf("template env %s: %w", key, err)
		}
		env[key] = value
	}
	return env, nil
}

func templateHealth(values map[string]any) (*workspace.Health, error) {
	if len(values) == 0 {
		return nil, nil
	}
	data, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("marshal template health: %w", err)
	}
	var health workspace.Health
	if err := yaml.Unmarshal(data, &health); err != nil {
		return nil, fmt.Errorf("decode template health: %w", err)
	}
	return &health, nil
}

func templatePorts(values []catalog.TemplatePort) []workspace.Port {
	if len(values) == 0 {
		return nil
	}
	ports := make([]workspace.Port, len(values))
	for i := range values {
		ports[i] = workspace.Port{Host: values[i].Host, Container: values[i].Container, Protocol: values[i].Protocol, HostIP: values[i].HostIP}
	}
	return ports
}

func templateVolumes(values []catalog.TemplateVolume) []workspace.Volume {
	if len(values) == 0 {
		return nil
	}
	volumes := make([]workspace.Volume, len(values))
	for i := range values {
		volumes[i] = workspace.Volume{Source: values[i].Source, Target: values[i].Target, ReadOnly: values[i].ReadOnly, Kind: values[i].Kind}
	}
	return volumes
}

func templateImports(values []catalog.TemplateImport) []workspace.Import {
	if len(values) == 0 {
		return nil
	}
	imports := make([]workspace.Import, len(values))
	for i := range values {
		imports[i] = workspace.Import{Contract: values[i].Contract, From: values[i].From, Alias: values[i].Alias}
	}
	return imports
}

func templateExports(values []catalog.TemplateExport) []workspace.Export {
	if len(values) == 0 {
		return nil
	}
	exports := make([]workspace.Export, len(values))
	for i := range values {
		exports[i] = workspace.Export{Contract: values[i].Contract, Env: cloneStringMap(values[i].Env)}
	}
	return exports
}

func cloneMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = cloneValue(value)
	}
	return cloned
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMap(typed)
	case map[any]any:
		converted := make(map[string]any, len(typed))
		for key, nested := range typed {
			converted[fmt.Sprint(key)] = cloneValue(nested)
		}
		return converted
	case []any:
		cloned := make([]any, len(typed))
		for i := range typed {
			cloned[i] = cloneValue(typed[i])
		}
		return cloned
	case []string:
		return append([]string(nil), typed...)
	default:
		return typed
	}
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

func normalizeProvider(provider string) string {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if normalized == "" {
		return runtimepkg.ProviderAuto
	}
	return normalized
}

func ensureApplyCapabilities(workspaceName, provider string, capabilities runtimepkg.AdapterCapabilities, diff *planpkg.Result) error {
	if diff == nil {
		return nil
	}
	for _, action := range diff.Actions {
		if action.Kind == planpkg.ActionNoop {
			continue
		}
		switch action.Scope {
		case planpkg.ScopeWorkspace:
			if !capabilities.Network {
				return unsupportedCapability(workspaceName, "", provider, "apply", "network", "selected runtime does not implement workspace network mutations in Phase 4")
			}
		case planpkg.ScopeResource:
			if !capabilities.Apply {
				return unsupportedCapability(workspaceName, action.Target, provider, "apply", "apply", "selected runtime does not implement resource mutations in Phase 4")
			}
		}
	}
	return nil
}

func unsupportedCapability(workspaceName, resourceName, provider, operation, capability, reason string) error {
	return &UnsupportedCapabilityError{
		Workspace:  workspaceName,
		Resource:   resourceName,
		Provider:   provider,
		Operation:  operation,
		Capability: capability,
		Reason:     reason,
	}
}

func inspectWarning(workspaceName, provider, message string) *runtimepkg.Diagnostic {
	return &runtimepkg.Diagnostic{
		Severity:  runtimepkg.SeverityWarning,
		Code:      "inspect-unavailable",
		Workspace: workspaceName,
		Provider:  provider,
		Message:   message,
	}
}

func detailPair(values []string) (string, string) {
	workspaceName := ""
	operation := ""
	if len(values) > 0 {
		workspaceName = values[0]
	}
	if len(values) > 1 {
		operation = values[1]
	}
	return workspaceName, operation
}
