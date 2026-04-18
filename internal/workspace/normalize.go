package workspace

import (
	"fmt"
	"path/filepath"
	"sort"
)

// Normalize applies deterministic defaults and path resolution to a loaded
// workspace manifest.
func Normalize(ws *Workspace) error {
	if ws == nil {
		return fmt.Errorf("normalize workspace: nil workspace")
	}

	ws.Catalog.Sources, ws.Catalog.ResolvedSources = normalizeCatalogSources(ws.ManifestDir, ws.Catalog.Sources)
	ws.Secrets = cloneRawMap(ws.Secrets)
	ws.Profiles = cloneRawMap(ws.Profiles)

	for _, key := range ws.SortedResourceKeys() {
		resource := ws.Resources[key]
		if resource == nil {
			continue
		}

		if resource.Enabled == nil {
			resource.SetEnabled(true)
		}
		resource.Env = cloneEnvMap(resource.Env)
		resource.Ports = normalizePorts(resource.Ports)
		resource.Volumes = normalizeVolumes(resource.Volumes)
		resource.DependsOn = normalizeStringSlice(resource.DependsOn)
		resource.Imports = normalizeImports(resource.Imports)
		resource.Exports = normalizeExports(resource.Exports)
		resource.Domains = normalizeStringSlice(resource.Domains)
		resource.Develop = cloneRawMap(resource.Develop)
		resource.Overrides = cloneRawMap(resource.Overrides)
		resource.Health = cloneHealth(resource.Health)

		if resource.Source != nil {
			resource.Source.Path = normalizeDisplayPath(resource.Source.Path)
			resource.Source.ResolvedPath = resolveManifestRelativePath(ws.ManifestDir, resource.Source.Path)
		}
	}

	return nil
}

func normalizeCatalogSources(baseDir string, sources []string) ([]string, []string) {
	if len(sources) == 0 {
		return nil, nil
	}

	type sourcePair struct {
		display  string
		resolved string
	}

	seen := make(map[string]sourcePair, len(sources))
	for _, source := range sources {
		display := normalizeDisplayPath(source)
		resolved := resolveManifestRelativePath(baseDir, display)
		if _, ok := seen[resolved]; ok {
			continue
		}
		seen[resolved] = sourcePair{display: display, resolved: resolved}
	}

	pairs := make([]sourcePair, 0, len(seen))
	for _, pair := range seen {
		pairs = append(pairs, pair)
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].resolved != pairs[j].resolved {
			return pairs[i].resolved < pairs[j].resolved
		}
		return pairs[i].display < pairs[j].display
	})

	display := make([]string, 0, len(pairs))
	resolved := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		display = append(display, pair.display)
		resolved = append(resolved, pair.resolved)
	}
	return display, resolved
}

func normalizePorts(ports []Port) []Port {
	if len(ports) == 0 {
		return nil
	}

	normalized := make([]Port, len(ports))
	copy(normalized, ports)
	for i := range normalized {
		normalized[i].Protocol = normalizeProtocol(normalized[i].Protocol)
	}

	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].Container != normalized[j].Container {
			return normalized[i].Container < normalized[j].Container
		}
		if normalized[i].Protocol != normalized[j].Protocol {
			return normalized[i].Protocol < normalized[j].Protocol
		}
		if normalized[i].Host != normalized[j].Host {
			return normalized[i].Host < normalized[j].Host
		}
		return normalized[i].HostIP < normalized[j].HostIP
	})

	return normalized
}

func normalizeVolumes(volumes []Volume) []Volume {
	if len(volumes) == 0 {
		return nil
	}

	normalized := make([]Volume, len(volumes))
	copy(normalized, volumes)
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].Target != normalized[j].Target {
			return normalized[i].Target < normalized[j].Target
		}
		if normalized[i].Source != normalized[j].Source {
			return normalized[i].Source < normalized[j].Source
		}
		if normalized[i].Kind != normalized[j].Kind {
			return normalized[i].Kind < normalized[j].Kind
		}
		if normalized[i].ReadOnly != normalized[j].ReadOnly {
			return !normalized[i].ReadOnly && normalized[j].ReadOnly
		}
		return false
	})
	return normalized
}

func normalizeImports(imports []Import) []Import {
	if len(imports) == 0 {
		return nil
	}

	normalized := make([]Import, len(imports))
	copy(normalized, imports)
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].Contract != normalized[j].Contract {
			return normalized[i].Contract < normalized[j].Contract
		}
		if normalized[i].From != normalized[j].From {
			return normalized[i].From < normalized[j].From
		}
		return normalized[i].Alias < normalized[j].Alias
	})
	return normalized
}

func normalizeExports(exports []Export) []Export {
	if len(exports) == 0 {
		return nil
	}

	normalized := make([]Export, len(exports))
	for i := range exports {
		normalized[i] = Export{
			Contract: exports[i].Contract,
			Env:      cloneStringMap(exports[i].Env),
		}
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].Contract < normalized[j].Contract
	})
	return normalized
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

func normalizeProtocol(protocol string) string {
	if protocol == "" {
		return "tcp"
	}
	return protocol
}

func resolveManifestRelativePath(baseDir, rawPath string) string {
	if rawPath == "" {
		return ""
	}
	if filepath.IsAbs(rawPath) {
		return filepath.Clean(rawPath)
	}
	return filepath.Clean(filepath.Join(baseDir, rawPath))
}

func normalizeDisplayPath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(path))
}
