package resolve

import (
	"fmt"
	"path/filepath"

	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
	"gopkg.in/yaml.v3"
)

// MissingTemplateError reports a resource referencing a catalog template that is
// absent from the loaded index.
type MissingTemplateError struct {
	ResourceKey  string
	TemplateName string
}

func (e *MissingTemplateError) Error() string {
	return fmt.Sprintf("resource %q references unknown template %q", e.ResourceKey, e.TemplateName)
}

// Resolve resolves a loaded workspace against a deterministic catalog index.
func Resolve(ws *workspace.Workspace, index *catalog.Index) (*Graph, error) {
	if ws == nil {
		return nil, fmt.Errorf("resolve workspace: nil workspace")
	}

	graph := &Graph{
		Workspace: Workspace{
			Name:           ws.Metadata.Name,
			DisplayName:    ws.Metadata.DisplayName,
			Description:    ws.Metadata.Description,
			Runtime:        ws.Runtime,
			Policies:       ws.Policies,
			CatalogSources: append([]string(nil), ws.Catalog.Sources...),
			ManifestPath:   ws.ManifestPath,
			ManifestDir:    ws.ManifestDir,
		},
		Resources: make([]*Resource, 0, len(ws.Resources)),
	}

	for _, key := range ws.SortedResourceKeys() {
		resource, err := buildResource(ws, index, key, ws.Resources[key])
		if err != nil {
			return nil, err
		}
		graph.Resources = append(graph.Resources, resource)
	}

	return graph, nil
}

func buildResource(ws *workspace.Workspace, index *catalog.Index, key string, resource *workspace.Resource) (*Resource, error) {
	resolved := &Resource{
		Key:       key,
		Enabled:   resource.EnabledValue(),
		Host:      key,
		Env:       cloneEnvMap(resource.Env),
		Ports:     append([]Port(nil), resource.Ports...),
		Volumes:   append([]Volume(nil), resource.Volumes...),
		DependsOn: normalizeStringSlice(resource.DependsOn),
		Imports:   append([]Import(nil), resource.Imports...),
		Exports:   append([]Export(nil), resource.Exports...),
		Health:    cloneHealth(resource.Health),
		Domains:   normalizeStringSlice(resource.Domains),
		Develop:   cloneRawMap(resource.Develop),
		Overrides: cloneRawMap(resource.Overrides),
	}

	if resource.Source != nil {
		resolved.Source = &SourceRef{
			Type:         resource.Source.Type,
			Path:         normalizeDisplayPath(resource.Source.Path),
			Service:      resource.Source.Service,
			ResolvedPath: resource.Source.ResolvedPath,
		}
	}

	if resource.Template == "" {
		resolved.Ports = mergePorts(nil, resolved.Ports)
		resolved.Volumes = mergeVolumes(nil, resolved.Volumes)
		resolved.Imports = mergeImports(nil, resolved.Imports)
		resolved.Exports = mergeExports(nil, resolved.Exports)
		resolved.Health = selectHealth(nil, resolved.Health)
		resolved.Develop = selectRawMap(nil, resolved.Develop)
		return resolved, nil
	}

	template, ok := index.ByName(resource.Template)
	if !ok {
		return nil, &MissingTemplateError{ResourceKey: key, TemplateName: resource.Template}
	}

	resolved.Template = &TemplateRef{
		Name:         template.Metadata.Name,
		Path:         displayTemplatePath(ws.ManifestDir, template.Path),
		ResolvedPath: template.Path,
	}

	templateRuntime, err := decodeRuntime(template.Spec.Runtime, template.Path)
	if err != nil {
		return nil, fmt.Errorf("decode runtime for resource %s template %s: %w", key, template.Metadata.Name, err)
	}
	templateEnv, err := decodeEnvMap(template.Spec.Env)
	if err != nil {
		return nil, fmt.Errorf("decode env for resource %s template %s: %w", key, template.Metadata.Name, err)
	}
	templateHealth, err := decodeHealth(template.Spec.Health)
	if err != nil {
		return nil, fmt.Errorf("decode health for resource %s template %s: %w", key, template.Metadata.Name, err)
	}

	resolved.Runtime = templateRuntime
	resolved.Env = mergeEnv(templateEnv, resource.Env)
	resolved.Ports = mergePorts(convertPorts(template.Spec.Ports), resource.Ports)
	resolved.Volumes = mergeVolumes(convertVolumes(template.Spec.Volumes), resource.Volumes)
	resolved.Imports = mergeImports(convertImports(template.Spec.Imports), resource.Imports)
	resolved.Exports = mergeExports(convertExports(template.Spec.Exports), resource.Exports)
	resolved.Health = selectHealth(templateHealth, resource.Health)
	resolved.Develop = selectRawMap(template.Spec.Develop, resource.Develop)

	return resolved, nil
}

func decodeEnvMap(raw map[string]any) (map[string]EnvValue, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	decoded := make(map[string]EnvValue, len(raw))
	for key, value := range raw {
		converted, err := workspace.EnvValueFromAny(value)
		if err != nil {
			return nil, fmt.Errorf("env %s: %w", key, err)
		}
		decoded[key] = converted
	}
	return decoded, nil
}

func decodeHealth(raw map[string]any) (*Health, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	data, err := yaml.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshal health block: %w", err)
	}

	var health workspace.Health
	if err := yaml.Unmarshal(data, &health); err != nil {
		return nil, fmt.Errorf("decode health block: %w", err)
	}
	return &health, nil
}

func convertPorts(ports []catalog.TemplatePort) []Port {
	if len(ports) == 0 {
		return nil
	}

	converted := make([]Port, len(ports))
	for i := range ports {
		converted[i] = Port{
			Host:      ports[i].Host,
			Container: ports[i].Container,
			Protocol:  normalizeProtocol(ports[i].Protocol),
			HostIP:    ports[i].HostIP,
		}
	}
	return converted
}

func convertVolumes(volumes []catalog.TemplateVolume) []Volume {
	if len(volumes) == 0 {
		return nil
	}

	converted := make([]Volume, len(volumes))
	for i := range volumes {
		converted[i] = Volume{
			Source:   volumes[i].Source,
			Target:   volumes[i].Target,
			ReadOnly: volumes[i].ReadOnly,
			Kind:     volumes[i].Kind,
		}
	}
	return converted
}

func convertImports(imports []catalog.TemplateImport) []Import {
	if len(imports) == 0 {
		return nil
	}

	converted := make([]Import, len(imports))
	for i := range imports {
		converted[i] = Import{
			Contract: imports[i].Contract,
			From:     imports[i].From,
			Alias:    imports[i].Alias,
		}
	}
	return converted
}

func convertExports(exports []catalog.TemplateExport) []Export {
	if len(exports) == 0 {
		return nil
	}

	converted := make([]Export, len(exports))
	for i := range exports {
		converted[i] = Export{
			Contract: exports[i].Contract,
			Env:      cloneStringMap(exports[i].Env),
		}
	}
	return converted
}

func displayTemplatePath(workspaceDir, templatePath string) string {
	if templatePath == "" {
		return ""
	}
	if !filepath.IsAbs(templatePath) {
		return normalizeDisplayPath(templatePath)
	}
	relativePath, err := filepath.Rel(workspaceDir, templatePath)
	if err != nil {
		return ""
	}
	return normalizeDisplayPath(relativePath)
}

func resolvePath(baseDir, rawPath string) string {
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
