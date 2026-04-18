package runtime

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/prospect-ogujiuba/devarch/internal/contracts"
	"github.com/prospect-ogujiuba/devarch/internal/resolve"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

// BuildDesiredWorkspace derives the Phase 3 runtime boundary from the stable
// Phase 2 resolve graph and contract result without mutating either input.
func BuildDesiredWorkspace(graph *resolve.Graph, result *contracts.Result) (*DesiredWorkspace, error) {
	if graph == nil {
		return nil, fmt.Errorf("build desired workspace: nil graph")
	}

	desired := &DesiredWorkspace{
		Name:           graph.Workspace.Name,
		DisplayName:    graph.Workspace.DisplayName,
		Description:    graph.Workspace.Description,
		Provider:       normalizedProvider(graph.Workspace.Runtime.Provider),
		NamingStrategy: normalizedNamingStrategy(graph.Workspace.Runtime.NamingStrategy),
		ManifestPath:   graph.Workspace.ManifestPath,
		ManifestDir:    graph.Workspace.ManifestDir,
		Resources:      make([]*DesiredResource, 0, len(graph.Resources)),
		Diagnostics:    convertContractDiagnostics(graph.Workspace.Name, result),
	}

	if graph.Workspace.Runtime.IsolatedNetwork {
		networkName := WorkspaceNetworkName(desired.Name, desired.NamingStrategy)
		desired.Network = &DesiredNetwork{
			Name:   networkName,
			Labels: WorkspaceLabels(desired.Name),
		}
	}

	injectedEnv := mapInjectedEnv(result)
	for _, resource := range graph.Resources {
		if resource == nil {
			continue
		}

		item := &DesiredResource{
			Key:          resource.Key,
			Enabled:      resource.Enabled,
			LogicalHost:  resource.Host,
			RuntimeName:  ResourceRuntimeName(desired.Name, resource.Key, desired.NamingStrategy),
			DeclaredEnv:  cloneEnvMap(resource.Env),
			InjectedEnv:  cloneEnvMap(injectedEnv[resource.Key]),
			DependsOn:    cloneStringSlice(resource.DependsOn),
			Domains:      cloneStringSlice(resource.Domains),
			Diagnostics:  nil,
			TemplateName: "",
		}
		if resource.Template != nil {
			item.TemplateName = resource.Template.Name
		}
		if resource.Source != nil {
			item.Source = &SourceRef{
				Type:         resource.Source.Type,
				Path:         resource.Source.Path,
				Service:      resource.Source.Service,
				ResolvedPath: resource.Source.ResolvedPath,
			}
		}

		if item.Source != nil && item.Source.Type == "raw-compose" {
			item.Diagnostics = append(item.Diagnostics, UnsupportedSourceDiagnostic(desired.Name, resource.Key, item.Source.Type))
		}

		overrideLabels, diagnostics := extractLabels(desired.Name, resource.Key, resource.Overrides)
		item.OverrideLabels = overrideLabels
		item.Diagnostics = append(item.Diagnostics, diagnostics...)

		watchRules, diagnostics := extractWatchRules(desired.Name, desired.ManifestDir, item.Source, resource.Key, resource.Develop)
		item.Diagnostics = append(item.Diagnostics, diagnostics...)

		item.Spec = ResourceSpec{
			Image:         imageFromResolve(resource.Runtime),
			Build:         buildFromResolve(resource.Runtime),
			Command:       commandFromResolve(resource.Runtime),
			Entrypoint:    entrypointFromResolve(resource.Runtime),
			WorkingDir:    workingDirFromResolve(resource.Runtime),
			Env:           mergeEnv(item.InjectedEnv, item.DeclaredEnv),
			Ports:         portsFromResolve(resource.Ports),
			Volumes:       volumesFromResolve(resource.Volumes),
			Health:        cloneHealth(resource.Health),
			ProjectSource: projectSourceFromResolve(item.Source, resource.Runtime, watchRules),
			DevelopWatch:  watchRules,
			Labels:        mergeLabels(ResourceLabels(desired.Name, resource.Key, resource.Host, networkName(desired)), item.OverrideLabels),
		}

		desired.Resources = append(desired.Resources, item)
	}

	return desired, nil
}

func normalizedProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", ProviderAuto:
		return ProviderAuto
	case ProviderDocker:
		return ProviderDocker
	case ProviderPodman:
		return ProviderPodman
	default:
		return strings.ToLower(strings.TrimSpace(provider))
	}
}

func normalizedNamingStrategy(strategy string) string {
	if strings.TrimSpace(strategy) == "" {
		return NamingStrategyWorkspaceResource
	}
	return strings.TrimSpace(strategy)
}

func convertContractDiagnostics(workspaceName string, result *contracts.Result) []Diagnostic {
	if result == nil || len(result.Diagnostics) == 0 {
		return nil
	}
	diagnostics := make([]Diagnostic, 0, len(result.Diagnostics))
	for _, diagnostic := range result.Diagnostics {
		diagnostics = append(diagnostics, Diagnostic{
			Severity:  diagnostic.Severity,
			Code:      diagnostic.Code,
			Workspace: workspaceName,
			Resource:  diagnostic.Consumer,
			Contract:  diagnostic.Contract,
			Provider:  diagnostic.Provider,
			Providers: append([]string(nil), diagnostic.Providers...),
			EnvKey:    diagnostic.EnvKey,
			Message:   diagnostic.Message,
		})
	}
	return diagnostics
}

func mapInjectedEnv(result *contracts.Result) map[string]map[string]workspace.EnvValue {
	if result == nil || len(result.Links) == 0 {
		return nil
	}
	mapped := make(map[string]map[string]workspace.EnvValue)
	for _, link := range result.Links {
		if len(link.InjectedEnv) == 0 {
			continue
		}
		current := mapped[link.Consumer]
		if current == nil {
			current = make(map[string]workspace.EnvValue, len(link.InjectedEnv))
			mapped[link.Consumer] = current
		}
		keys := make([]string, 0, len(link.InjectedEnv))
		for key := range link.InjectedEnv {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			current[key] = link.InjectedEnv[key].Clone()
		}
	}
	return mapped
}

func extractLabels(workspaceName, resourceKey string, overrides map[string]any) (map[string]string, []Diagnostic) {
	if len(overrides) == 0 {
		return nil, nil
	}
	labels := make(map[string]string)
	diagnostics := make([]Diagnostic, 0)
	keys := make([]string, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if key != "labels" {
			diagnostics = append(diagnostics, UnsupportedFieldDiagnostic(workspaceName, resourceKey, "unsupported-override", fmt.Sprintf("resource %q override %q is unsupported in Phase 3", resourceKey, key)))
			continue
		}
		typed, ok := toStringMap(overrides[key])
		if !ok {
			diagnostics = append(diagnostics, UnsupportedFieldDiagnostic(workspaceName, resourceKey, "unsupported-labels", fmt.Sprintf("resource %q overrides.labels must be a string map", resourceKey)))
			continue
		}
		for labelKey, value := range typed {
			labels[labelKey] = value
		}
	}
	if len(labels) == 0 {
		labels = nil
	}
	if len(diagnostics) == 0 {
		diagnostics = nil
	}
	return labels, diagnostics
}

func extractWatchRules(workspaceName, manifestDir string, source *SourceRef, resourceKey string, develop map[string]any) ([]WatchRule, []Diagnostic) {
	if len(develop) == 0 {
		return nil, nil
	}
	rules := make([]WatchRule, 0)
	diagnostics := make([]Diagnostic, 0)
	keys := make([]string, 0, len(develop))
	for key := range develop {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if key != "watch" {
			diagnostics = append(diagnostics, UnsupportedFieldDiagnostic(workspaceName, resourceKey, "unsupported-develop", fmt.Sprintf("resource %q develop.%s is unsupported in Phase 3", resourceKey, key)))
			continue
		}
		entries, ok := develop[key].([]any)
		if !ok {
			diagnostics = append(diagnostics, UnsupportedFieldDiagnostic(workspaceName, resourceKey, "unsupported-develop-watch", fmt.Sprintf("resource %q develop.watch must be a list", resourceKey)))
			continue
		}
		for _, entry := range entries {
			typed, ok := entry.(map[string]any)
			if !ok {
				diagnostics = append(diagnostics, UnsupportedFieldDiagnostic(workspaceName, resourceKey, "unsupported-develop-watch", fmt.Sprintf("resource %q develop.watch entries must be objects", resourceKey)))
				continue
			}
			pathValue, _ := typed["path"].(string)
			targetValue, _ := typed["target"].(string)
			actionValue, _ := typed["action"].(string)
			if strings.TrimSpace(pathValue) == "" || strings.TrimSpace(targetValue) == "" {
				diagnostics = append(diagnostics, UnsupportedFieldDiagnostic(workspaceName, resourceKey, "unsupported-develop-watch", fmt.Sprintf("resource %q develop.watch entries must include path and target", resourceKey)))
				continue
			}
			rules = append(rules, WatchRule{
				Path:         filepath.ToSlash(filepath.Clean(pathValue)),
				ResolvedPath: resolveWatchPath(manifestDir, source, pathValue),
				Target:       targetValue,
				Action:       actionValue,
			})
		}
	}
	if len(rules) == 0 {
		rules = nil
	}
	if len(diagnostics) == 0 {
		diagnostics = nil
	}
	return rules, diagnostics
}

func resolveWatchPath(manifestDir string, source *SourceRef, raw string) string {
	clean := filepath.Clean(raw)
	if filepath.IsAbs(clean) {
		return clean
	}
	if source != nil && source.Type == "project" && (clean == "." || clean == "./" || clean == filepath.Clean(source.Path)) {
		if source.ResolvedPath != "" {
			return source.ResolvedPath
		}
	}
	return filepath.Clean(filepath.Join(manifestDir, clean))
}

func imageFromResolve(runtime *resolve.Runtime) string {
	if runtime == nil {
		return ""
	}
	return runtime.Image
}

func buildFromResolve(runtime *resolve.Runtime) *BuildSpec {
	if runtime == nil || runtime.Build == nil {
		return nil
	}
	return &BuildSpec{
		Context:            runtime.Build.Context,
		Dockerfile:         runtime.Build.Dockerfile,
		Target:             runtime.Build.Target,
		Args:               cloneEnvMap(runtime.Build.Args),
		ResolvedContext:    runtime.Build.ResolvedContext,
		ResolvedDockerfile: runtime.Build.ResolvedDockerfile,
	}
}

func commandFromResolve(runtime *resolve.Runtime) []string {
	if runtime == nil {
		return nil
	}
	return cloneStringSlice(runtime.Command)
}

func entrypointFromResolve(runtime *resolve.Runtime) []string {
	if runtime == nil {
		return nil
	}
	return cloneStringSlice(runtime.Entrypoint)
}

func workingDirFromResolve(runtime *resolve.Runtime) string {
	if runtime == nil {
		return ""
	}
	return runtime.WorkingDir
}

func portsFromResolve(ports []resolve.Port) []PortSpec {
	if len(ports) == 0 {
		return nil
	}
	converted := make([]PortSpec, len(ports))
	for i := range ports {
		converted[i] = PortSpec{
			Container: ports[i].Container,
			Published: ports[i].Host,
			Protocol:  ports[i].Protocol,
			HostIP:    ports[i].HostIP,
		}
	}
	return converted
}

func volumesFromResolve(volumes []resolve.Volume) []VolumeSpec {
	if len(volumes) == 0 {
		return nil
	}
	converted := make([]VolumeSpec, len(volumes))
	for i := range volumes {
		converted[i] = VolumeSpec{
			Source:   volumes[i].Source,
			Target:   volumes[i].Target,
			ReadOnly: volumes[i].ReadOnly,
			Kind:     volumes[i].Kind,
		}
	}
	return converted
}

func projectSourceFromResolve(source *SourceRef, runtime *resolve.Runtime, watchRules []WatchRule) *ProjectSource {
	if source == nil || source.Type != "project" || source.ResolvedPath == "" {
		return nil
	}
	containerPath := ""
	if runtime != nil {
		containerPath = runtime.WorkingDir
	}
	if containerPath == "" && len(watchRules) > 0 {
		containerPath = watchRules[0].Target
	}
	if containerPath == "" {
		containerPath = "/workspace"
	}
	return &ProjectSource{HostPath: source.ResolvedPath, ContainerPath: containerPath}
}

func mergeEnv(base, overlay map[string]workspace.EnvValue) map[string]workspace.EnvValue {
	merged := cloneEnvMap(base)
	if merged == nil {
		merged = make(map[string]workspace.EnvValue, len(overlay))
	}
	for key, value := range overlay {
		merged[key] = value.Clone()
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func mergeLabels(base, overlay map[string]string) map[string]string {
	merged := cloneStringMap(base)
	if merged == nil {
		merged = make(map[string]string, len(overlay))
	}
	for key, value := range overlay {
		merged[key] = value
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func networkName(desired *DesiredWorkspace) string {
	if desired == nil || desired.Network == nil {
		return ""
	}
	return desired.Network.Name
}

func toStringMap(value any) (map[string]string, bool) {
	switch typed := value.(type) {
	case map[string]string:
		return cloneStringMap(typed), true
	case map[string]any:
		converted := make(map[string]string, len(typed))
		for key, raw := range typed {
			stringValue, ok := raw.(string)
			if !ok {
				return nil, false
			}
			converted[key] = stringValue
		}
		return converted, true
	case map[any]any:
		converted := make(map[string]string, len(typed))
		for key, raw := range typed {
			stringValue, ok := raw.(string)
			if !ok {
				return nil, false
			}
			converted[fmt.Sprint(key)] = stringValue
		}
		return converted, true
	default:
		return nil, false
	}
}
