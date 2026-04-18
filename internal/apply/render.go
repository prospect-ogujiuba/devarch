package apply

import (
	"fmt"

	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

func Render(desired *runtimepkg.DesiredWorkspace) (*Payload, error) {
	if desired == nil {
		return nil, fmt.Errorf("render payload: nil desired workspace")
	}

	payload := &Payload{
		Workspace:   desired.Name,
		Provider:    desired.Provider,
		Diagnostics: append([]runtimepkg.Diagnostic(nil), desired.Diagnostics...),
		Blocked:     desired.Blocked(),
	}
	if desired.Network != nil {
		payload.Network = &NetworkPayload{Name: desired.Network.Name, Labels: cloneStringMap(desired.Network.Labels)}
	}

	resources := make([]*ResourcePayload, 0, len(desired.Resources))
	for _, resource := range desired.Resources {
		if resource == nil || !resource.Enabled {
			continue
		}
		if len(resource.Diagnostics) > 0 {
			payload.Diagnostics = append(payload.Diagnostics, resource.Diagnostics...)
		}
		if resource.Blocked() {
			continue
		}
		resources = append(resources, &ResourcePayload{
			Key:           resource.Key,
			LogicalHost:   resource.LogicalHost,
			RuntimeName:   resource.RuntimeName,
			Source:        cloneSource(resource.Source),
			Image:         resource.Spec.Image,
			Build:         buildPayload(resource.Spec.Build),
			Command:       cloneStringSlice(resource.Spec.Command),
			Entrypoint:    cloneStringSlice(resource.Spec.Entrypoint),
			WorkingDir:    resource.Spec.WorkingDir,
			DeclaredEnv:   cloneEnvMap(resource.DeclaredEnv),
			InjectedEnv:   cloneEnvMap(resource.InjectedEnv),
			Env:           cloneEnvMap(resource.Spec.Env),
			Ports:         portPayloads(resource.Spec.Ports),
			Volumes:       volumePayloads(resource.Spec.Volumes),
			Health:        cloneHealth(resource.Spec.Health),
			ProjectSource: cloneProjectSource(resource.Spec.ProjectSource),
			DevelopWatch:  cloneWatchRules(resource.Spec.DevelopWatch),
			Labels:        cloneStringMap(resource.Spec.Labels),
		})
	}
	payload.Resources = resources
	payload.Blocked = payloadHasBlockingDiagnostics(payload)
	return payload, nil
}

func payloadHasBlockingDiagnostics(payload *Payload) bool {
	if payload == nil {
		return false
	}
	for _, diagnostic := range payload.Diagnostics {
		if diagnostic.BlocksApply() {
			return true
		}
	}
	return payload.Blocked
}

func cloneSource(source *runtimepkg.SourceRef) *runtimepkg.SourceRef {
	if source == nil {
		return nil
	}
	cloned := *source
	return &cloned
}

func buildPayload(build *runtimepkg.BuildSpec) *BuildPayload {
	if build == nil {
		return nil
	}
	return &BuildPayload{
		Context:    build.Context,
		Dockerfile: build.Dockerfile,
		Target:     build.Target,
		Args:       cloneEnvMap(build.Args),
	}
}

func portPayloads(values []runtimepkg.PortSpec) []PortPayload {
	if len(values) == 0 {
		return nil
	}
	ports := make([]PortPayload, len(values))
	for i := range values {
		ports[i] = PortPayload(values[i])
	}
	return ports
}

func volumePayloads(values []runtimepkg.VolumeSpec) []VolumePayload {
	if len(values) == 0 {
		return nil
	}
	volumes := make([]VolumePayload, len(values))
	for i := range values {
		volumes[i] = VolumePayload(values[i])
	}
	return volumes
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

func cloneProjectSource(source *runtimepkg.ProjectSource) *runtimepkg.ProjectSource {
	if source == nil {
		return nil
	}
	cloned := *source
	return &cloned
}

func cloneWatchRules(values []runtimepkg.WatchRule) []runtimepkg.WatchRule {
	if len(values) == 0 {
		return nil
	}
	return append([]runtimepkg.WatchRule(nil), values...)
}
