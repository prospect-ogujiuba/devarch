package resolve

import (
	"fmt"
	"sort"
	"strings"
)

func mergeEnv(templateEnv, workspaceEnv map[string]EnvValue) map[string]EnvValue {
	if len(templateEnv) == 0 && len(workspaceEnv) == 0 {
		return nil
	}

	merged := make(map[string]EnvValue, len(templateEnv)+len(workspaceEnv))
	for key, value := range templateEnv {
		merged[key] = value.Clone()
	}
	for key, value := range workspaceEnv {
		merged[key] = value.Clone()
	}
	return merged
}

func mergePorts(templatePorts, workspacePorts []Port) []Port {
	if len(templatePorts) == 0 && len(workspacePorts) == 0 {
		return nil
	}

	merged := make(map[string]Port, len(templatePorts)+len(workspacePorts))
	for _, port := range templatePorts {
		normalized := normalizePort(port)
		merged[portMergeKey(normalized)] = normalized
	}
	for _, port := range workspacePorts {
		normalized := normalizePort(port)
		merged[portMergeKey(normalized)] = normalized
	}

	ports := make([]Port, 0, len(merged))
	for _, port := range merged {
		ports = append(ports, port)
	}
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Container != ports[j].Container {
			return ports[i].Container < ports[j].Container
		}
		if ports[i].Protocol != ports[j].Protocol {
			return ports[i].Protocol < ports[j].Protocol
		}
		if ports[i].Host != ports[j].Host {
			return ports[i].Host < ports[j].Host
		}
		return ports[i].HostIP < ports[j].HostIP
	})
	return ports
}

func mergeVolumes(templateVolumes, workspaceVolumes []Volume) []Volume {
	if len(templateVolumes) == 0 && len(workspaceVolumes) == 0 {
		return nil
	}

	merged := make(map[string]Volume, len(templateVolumes)+len(workspaceVolumes))
	for _, volume := range templateVolumes {
		merged[volume.Target] = volume
	}
	for _, volume := range workspaceVolumes {
		merged[volume.Target] = volume
	}

	volumes := make([]Volume, 0, len(merged))
	for _, volume := range merged {
		volumes = append(volumes, volume)
	}
	sort.Slice(volumes, func(i, j int) bool {
		if volumes[i].Target != volumes[j].Target {
			return volumes[i].Target < volumes[j].Target
		}
		if volumes[i].Source != volumes[j].Source {
			return volumes[i].Source < volumes[j].Source
		}
		if volumes[i].Kind != volumes[j].Kind {
			return volumes[i].Kind < volumes[j].Kind
		}
		if volumes[i].ReadOnly != volumes[j].ReadOnly {
			return !volumes[i].ReadOnly && volumes[j].ReadOnly
		}
		return false
	})
	return volumes
}

func mergeImports(templateImports, workspaceImports []Import) []Import {
	if len(templateImports) == 0 && len(workspaceImports) == 0 {
		return nil
	}

	merged := make(map[string]Import, len(templateImports)+len(workspaceImports))
	for _, imp := range templateImports {
		merged[importMergeKey(imp)] = imp
	}
	for _, imp := range workspaceImports {
		key := importMergeKey(imp)
		if existing, ok := merged[key]; ok {
			if imp.From == "" {
				imp.From = existing.From
			}
			if imp.Alias == "" {
				imp.Alias = existing.Alias
			}
		}
		merged[key] = imp
	}

	imports := make([]Import, 0, len(merged))
	for _, imp := range merged {
		imports = append(imports, imp)
	}
	sort.Slice(imports, func(i, j int) bool {
		if imports[i].Contract != imports[j].Contract {
			return imports[i].Contract < imports[j].Contract
		}
		if imports[i].From != imports[j].From {
			return imports[i].From < imports[j].From
		}
		return imports[i].Alias < imports[j].Alias
	})
	return imports
}

func mergeExports(templateExports, workspaceExports []Export) []Export {
	if len(templateExports) == 0 && len(workspaceExports) == 0 {
		return nil
	}

	merged := make(map[string]Export, len(templateExports)+len(workspaceExports))
	for _, export := range templateExports {
		merged[export.Contract] = Export{
			Contract: export.Contract,
			Env:      cloneStringMap(export.Env),
		}
	}
	for _, export := range workspaceExports {
		existing, ok := merged[export.Contract]
		if !ok {
			merged[export.Contract] = Export{
				Contract: export.Contract,
				Env:      cloneStringMap(export.Env),
			}
			continue
		}
		if len(export.Env) == 0 {
			merged[export.Contract] = existing
			continue
		}
		mergedEnv := cloneStringMap(existing.Env)
		if mergedEnv == nil {
			mergedEnv = make(map[string]string, len(export.Env))
		}
		for key, value := range export.Env {
			mergedEnv[key] = value
		}
		existing.Env = mergedEnv
		merged[export.Contract] = existing
	}

	exports := make([]Export, 0, len(merged))
	for _, export := range merged {
		exports = append(exports, export)
	}
	sort.Slice(exports, func(i, j int) bool {
		return exports[i].Contract < exports[j].Contract
	})
	return exports
}

func selectHealth(templateHealth, workspaceHealth *Health) *Health {
	if workspaceHealth != nil {
		return cloneHealth(workspaceHealth)
	}
	return cloneHealth(templateHealth)
}

func selectRawMap(templateValue, workspaceValue map[string]any) map[string]any {
	if workspaceValue != nil {
		return cloneRawMap(workspaceValue)
	}
	return cloneRawMap(templateValue)
}

func normalizePort(port Port) Port {
	port.Protocol = normalizeProtocol(port.Protocol)
	return port
}

func normalizeProtocol(protocol string) string {
	if protocol == "" {
		return "tcp"
	}
	return protocol
}

func portMergeKey(port Port) string {
	return fmt.Sprintf("%d/%s", port.Container, port.Protocol)
}

func importMergeKey(imp Import) string {
	key := imp.Contract
	if imp.Alias != "" {
		key = key + "\x00" + imp.Alias
	}
	return key
}

func cloneEnvMap(values map[string]EnvValue) map[string]EnvValue {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]EnvValue, len(values))
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

func cloneHealth(health *Health) *Health {
	if health == nil {
		return nil
	}

	cloned := *health
	if len(health.Test) > 0 {
		cloned.Test = append(StringList(nil), health.Test...)
	}
	return &cloned
}

func cloneRawMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = cloneRawValue(value)
	}
	return cloned
}

func cloneRawValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneRawMap(typed)
	case map[any]any:
		converted := make(map[string]any, len(typed))
		for key, nested := range typed {
			converted[fmt.Sprint(key)] = cloneRawValue(nested)
		}
		return converted
	case []any:
		cloned := make([]any, len(typed))
		for i := range typed {
			cloned[i] = cloneRawValue(typed[i])
		}
		return cloned
	case []string:
		return append([]string(nil), typed...)
	default:
		return typed
	}
}

func normalizeStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return nil
	}
	sort.Strings(normalized)
	return normalized
}
