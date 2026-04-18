package importv1

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/prospect-ogujiuba/devarch/internal/spec"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
	"gopkg.in/yaml.v3"
)

type v1StackFile struct {
	Version   int                       `yaml:"version"`
	Stack     v1StackConfig             `yaml:"stack"`
	Instances map[string]v1InstanceDef  `yaml:"instances"`
	Wires     []v1WireDef               `yaml:"wires,omitempty"`
}

type v1StackConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	NetworkName string `yaml:"network_name,omitempty"`
}

type v1InstanceDef struct {
	Template     string                          `yaml:"template"`
	Enabled      bool                            `yaml:"enabled"`
	Image        string                          `yaml:"image,omitempty"`
	Ports        []v1PortDef                     `yaml:"ports,omitempty"`
	Volumes      []v1VolumeDef                   `yaml:"volumes,omitempty"`
	Environment  map[string]string               `yaml:"environment,omitempty"`
	EnvFiles     []string                        `yaml:"env_files,omitempty"`
	Networks     []string                        `yaml:"networks,omitempty"`
	Labels       map[string]string               `yaml:"labels,omitempty"`
	Domains      []v1DomainDef                   `yaml:"domains,omitempty"`
	Healthcheck  *v1HealthcheckDef               `yaml:"healthcheck,omitempty"`
	Dependencies []string                        `yaml:"dependencies,omitempty"`
	ConfigFiles  map[string]v1ConfigFileDef      `yaml:"config_files,omitempty"`
	ConfigMounts []v1ConfigMountDef              `yaml:"config_mounts,omitempty"`
	Command      string                          `yaml:"command,omitempty"`
}

type v1PortDef struct {
	HostIP        string `yaml:"host_ip"`
	HostPort      int    `yaml:"host_port"`
	ContainerPort int    `yaml:"container_port"`
	Protocol      string `yaml:"protocol"`
}

type v1VolumeDef struct {
	Source   string `yaml:"source"`
	Target   string `yaml:"target"`
	ReadOnly bool   `yaml:"read_only,omitempty"`
}

type v1DomainDef struct {
	Domain    string `yaml:"domain"`
	ProxyPort int    `yaml:"proxy_port"`
}

type v1HealthcheckDef struct {
	Test        string `yaml:"test"`
	Interval    string `yaml:"interval,omitempty"`
	Timeout     string `yaml:"timeout,omitempty"`
	Retries     int    `yaml:"retries,omitempty"`
	StartPeriod string `yaml:"start_period,omitempty"`
}

type v1ConfigFileDef struct {
	Content  string `yaml:"content"`
	FileMode string `yaml:"file_mode"`
}

type v1ConfigMountDef struct {
	SourcePath     string `yaml:"source_path"`
	TargetPath     string `yaml:"target_path"`
	ReadOnly       bool   `yaml:"read_only,omitempty"`
	ConfigFilePath string `yaml:"config_file_path,omitempty"`
}

type v1WireDef struct {
	ConsumerInstance string `yaml:"consumer_instance"`
	ProviderInstance string `yaml:"provider_instance"`
	ImportContract   string `yaml:"import_contract"`
	ExportContract   string `yaml:"export_contract"`
	Source           string `yaml:"source"`
}

var secretPlaceholderPattern = regexp.MustCompile(`^\$\{SECRET:([^}]+)\}$`)

func importStack(path string) (*Result, error) {
	result := newResult(ModeV1Stack, path)
	artifact := Artifact{
		Kind:          ArtifactKindWorkspace,
		Name:          filepath.Base(filepath.Dir(path)),
		SourcePath:    path,
		SuggestedPath: suggestedWorkspacePath(filepath.Base(filepath.Dir(path))),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read V1 stack export %s: %w", path, err)
	}

	var file v1StackFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		artifact.Status = StatusRejected
		artifact.Diagnostics = append(artifact.Diagnostics, failure("stack-export-decode-failed", err.Error(), path, ""))
		result.Artifacts = append(result.Artifacts, artifact)
		return finalizeResult(result), nil
	}

	if strings.TrimSpace(file.Stack.Name) != "" {
		artifact.Name = file.Stack.Name
		artifact.SuggestedPath = suggestedWorkspacePath(file.Stack.Name)
	}

	if file.Version != 1 {
		artifact.Status = StatusRejected
		artifact.Diagnostics = append(artifact.Diagnostics, failure("stack-export-version-unsupported", fmt.Sprintf("unsupported V1 export version %d", file.Version), path, "version"))
		result.Artifacts = append(result.Artifacts, artifact)
		return finalizeResult(result), nil
	}

	workspaceDoc, diagnostics, ok := buildWorkspaceDocument(&file, path)
	artifact.Diagnostics = append(artifact.Diagnostics, diagnostics...)
	if !ok {
		artifact.Status = StatusRejected
		result.Artifacts = append(result.Artifacts, artifact)
		return finalizeResult(result), nil
	}

	artifact.Name = workspaceDoc.Metadata.Name
	artifact.SuggestedPath = suggestedWorkspacePath(workspaceDoc.Metadata.Name)
	workspaceBytes, err := yaml.Marshal(workspaceDoc)
	if err != nil {
		artifact.Status = StatusRejected
		artifact.Diagnostics = append(artifact.Diagnostics, failure("workspace-marshal-failed", err.Error(), path, ""))
		result.Artifacts = append(result.Artifacts, artifact)
		return finalizeResult(result), nil
	}
	if err := spec.ValidateWorkspaceBytes(workspaceBytes); err != nil {
		artifact.Status = StatusRejected
		artifact.Diagnostics = append(artifact.Diagnostics, validationDiagnostics("workspace-validation-failed", err, path)...)
		result.Artifacts = append(result.Artifacts, artifact)
		return finalizeResult(result), nil
	}

	artifact.Document = string(workspaceBytes)
	artifact.Status = deriveArtifactStatus(artifact.Document, artifact.Diagnostics)
	result.Artifacts = append(result.Artifacts, artifact)
	return finalizeResult(result), nil
}

func buildWorkspaceDocument(file *v1StackFile, sourcePath string) (*workspace.Workspace, []Diagnostic, bool) {
	diagnostics := make([]Diagnostic, 0)
	if file == nil {
		diagnostics = append(diagnostics, failure("stack-export-missing", "stack export is nil", sourcePath, ""))
		return nil, diagnostics, false
	}
	if strings.TrimSpace(file.Stack.Name) == "" {
		diagnostics = append(diagnostics, failure("stack-name-missing", "stack.name is required", sourcePath, "stack.name"))
		return nil, diagnostics, false
	}
	if len(file.Instances) == 0 {
		diagnostics = append(diagnostics, failure("stack-instances-missing", "stack export must contain at least one instance", sourcePath, "instances"))
		return nil, diagnostics, false
	}

	ws := &workspace.Workspace{
		APIVersion: "devarch.io/v2alpha1",
		Kind:       "Workspace",
		Metadata: workspace.Metadata{
			Name:        file.Stack.Name,
			Description: file.Stack.Description,
		},
		Runtime: workspace.RuntimePreferences{
			Provider:        "auto",
			IsolatedNetwork: true,
			NamingStrategy:  "workspace-resource",
		},
		Policies: workspace.Policies{AutoWire: false},
		Resources: make(map[string]*workspace.Resource, len(file.Instances)),
	}

	if file.Stack.NetworkName != "" && file.Stack.NetworkName != fmt.Sprintf("devarch-%s-net", file.Stack.Name) {
		diagnostics = append(diagnostics, warning("stack-network-name-lossy", fmt.Sprintf("V1 stack network %q is not first-class in the V2 workspace schema and was omitted in favor of runtime.namingStrategy=workspace-resource.", file.Stack.NetworkName), sourcePath, "stack.network_name"))
	}

	providerExports := make(map[string][]workspace.Export)
	consumerImports := make(map[string][]workspace.Import)
	wireCompat := make(map[string][]map[string]any)
	for _, wire := range file.Wires {
		if strings.TrimSpace(wire.ConsumerInstance) == "" || strings.TrimSpace(wire.ProviderInstance) == "" || strings.TrimSpace(wire.ImportContract) == "" || strings.TrimSpace(wire.ExportContract) == "" {
			diagnostics = append(diagnostics, warning("wire-skipped-incomplete", fmt.Sprintf("skipped incomplete wire %+v", wire), sourcePath, "wires"))
			continue
		}
		if _, ok := file.Instances[wire.ConsumerInstance]; !ok {
			diagnostics = append(diagnostics, warning("wire-consumer-missing", fmt.Sprintf("wire consumer %q is not present in instances and was skipped", wire.ConsumerInstance), sourcePath, "wires.consumer_instance"))
			continue
		}
		if _, ok := file.Instances[wire.ProviderInstance]; !ok {
			diagnostics = append(diagnostics, warning("wire-provider-missing", fmt.Sprintf("wire provider %q is not present in instances and was skipped", wire.ProviderInstance), sourcePath, "wires.provider_instance"))
			continue
		}
		consumerImports[wire.ConsumerInstance] = append(consumerImports[wire.ConsumerInstance], workspace.Import{Contract: wire.ImportContract, From: wire.ProviderInstance})
		providerExports[wire.ProviderInstance] = append(providerExports[wire.ProviderInstance], workspace.Export{Contract: wire.ExportContract})
		if wire.Source != "" {
			wireCompat[wire.ConsumerInstance] = append(wireCompat[wire.ConsumerInstance], map[string]any{
				"provider":       wire.ProviderInstance,
				"importContract": wire.ImportContract,
				"exportContract": wire.ExportContract,
				"source":         wire.Source,
			})
		}
	}

	keys := make([]string, 0, len(file.Instances))
	for key := range file.Instances {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		resource, resourceDiagnostics, ok := buildWorkspaceResource(key, file.Instances[key], consumerImports[key], providerExports[key], wireCompat[key], sourcePath)
		diagnostics = append(diagnostics, resourceDiagnostics...)
		if !ok {
			return nil, diagnostics, false
		}
		ws.Resources[key] = resource
	}

	return ws, diagnostics, true
}

func buildWorkspaceResource(name string, instance v1InstanceDef, imports []workspace.Import, exports []workspace.Export, wireCompat []map[string]any, sourcePath string) (*workspace.Resource, []Diagnostic, bool) {
	diagnostics := make([]Diagnostic, 0)
	if strings.TrimSpace(instance.Template) == "" {
		diagnostics = append(diagnostics, failure("instance-template-missing", fmt.Sprintf("instance %q is missing template", name), sourcePath, fmt.Sprintf("instances.%s.template", name)))
		return nil, diagnostics, false
	}

	resource := &workspace.Resource{Template: instance.Template}
	resource.SetEnabled(instance.Enabled)
	resource.Env = convertInstanceEnv(instance.Environment)
	resource.Ports = convertInstancePorts(instance.Ports)
	resource.Volumes = convertInstanceVolumes(instance.Volumes)
	resource.DependsOn = append([]string(nil), compactStrings(instance.Dependencies)...)
	resource.Imports = dedupeImports(imports)
	resource.Exports = dedupeExports(exports)
	resource.Health = convertInstanceHealth(instance.Healthcheck)
	resource.Domains = convertInstanceDomains(instance.Domains)

	compat := make(map[string]any)
	compatFields := make([]string, 0)
	if instance.Image != "" {
		compat["image"] = instance.Image
		compatFields = append(compatFields, "image")
	}
	if instance.Command != "" {
		compat["command"] = instance.Command
		compatFields = append(compatFields, "command")
	}
	if len(instance.EnvFiles) > 0 {
		compat["envFiles"] = append([]string(nil), instance.EnvFiles...)
		compatFields = append(compatFields, "envFiles")
	}
	if len(instance.Networks) > 0 {
		compat["networks"] = append([]string(nil), compactStrings(instance.Networks)...)
		compatFields = append(compatFields, "networks")
	}
	if len(instance.Labels) > 0 {
		compat["labels"] = cloneStringMap(instance.Labels)
		compatFields = append(compatFields, "labels")
	}
	if len(instance.ConfigFiles) > 0 {
		compat["configFiles"] = configFilesForInstance(instance.ConfigFiles)
		compatFields = append(compatFields, "configFiles")
	}
	if len(instance.ConfigMounts) > 0 {
		compat["configMounts"] = configMountsForInstance(instance.ConfigMounts)
		compatFields = append(compatFields, "configMounts")
	}
	if len(instance.Domains) > 0 {
		compat["domains"] = domainsForCompat(instance.Domains)
		compatFields = append(compatFields, "domains")
	}
	if len(wireCompat) > 0 {
		compat["wires"] = wireCompat
		compatFields = append(compatFields, "wires")
	}
	if len(compat) > 0 {
		resource.Overrides = map[string]any{"importv1": compat}
		sort.Strings(compatFields)
		diagnostics = append(diagnostics, warning("instance-compat-fields-preserved", fmt.Sprintf("Instance %q preserved non-schema-native V1 fields under resources.%s.overrides.importv1: %s.", name, name, strings.Join(compatFields, ", ")), sourcePath, fmt.Sprintf("resources.%s.overrides.importv1", name)))
	}
	if len(imports) > 0 || len(exports) > 0 {
		diagnostics = append(diagnostics, info("instance-wires-mapped", fmt.Sprintf("Mapped %d import(s) and %d export(s) for instance %q from V1 wires.", len(resource.Imports), len(resource.Exports), name), sourcePath, fmt.Sprintf("resources.%s", name)))
	}

	return resource, diagnostics, true
}

func convertInstanceEnv(values map[string]string) map[string]workspace.EnvValue {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make(map[string]workspace.EnvValue, len(keys))
	for _, key := range keys {
		if matches := secretPlaceholderPattern.FindStringSubmatch(values[key]); len(matches) == 2 {
			result[key] = workspace.SecretRefEnvValue(matches[1])
			continue
		}
		result[key] = workspace.StringEnvValue(values[key])
	}
	return result
}

func convertInstancePorts(values []v1PortDef) []workspace.Port {
	result := make([]workspace.Port, 0, len(values))
	for _, value := range values {
		protocol := value.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		result = append(result, workspace.Port{Host: value.HostPort, Container: value.ContainerPort, Protocol: protocol, HostIP: value.HostIP})
	}
	return result
}

func convertInstanceVolumes(values []v1VolumeDef) []workspace.Volume {
	result := make([]workspace.Volume, 0, len(values))
	for _, value := range values {
		result = append(result, workspace.Volume{Source: value.Source, Target: value.Target, ReadOnly: value.ReadOnly})
	}
	return result
}

func convertInstanceHealth(value *v1HealthcheckDef) *workspace.Health {
	if value == nil || strings.TrimSpace(value.Test) == "" {
		return nil
	}
	return &workspace.Health{Test: workspace.StringList{value.Test}, Interval: value.Interval, Timeout: value.Timeout, Retries: value.Retries, StartPeriod: value.StartPeriod}
}

func convertInstanceDomains(values []v1DomainDef) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value.Domain) == "" {
			continue
		}
		result = append(result, value.Domain)
	}
	return compactStrings(result)
}

func dedupeImports(values []workspace.Import) []workspace.Import {
	seen := make(map[string]workspace.Import)
	for _, value := range values {
		key := value.Contract + "::" + value.From + "::" + value.Alias
		seen[key] = value
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]workspace.Import, 0, len(keys))
	for _, key := range keys {
		result = append(result, seen[key])
	}
	return result
}

func dedupeExports(values []workspace.Export) []workspace.Export {
	seen := make(map[string]workspace.Export)
	for _, value := range values {
		seen[value.Contract] = value
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]workspace.Export, 0, len(keys))
	for _, key := range keys {
		result = append(result, seen[key])
	}
	return result
}

func configFilesForInstance(values map[string]v1ConfigFileDef) map[string]any {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make(map[string]any, len(keys))
	for _, key := range keys {
		result[key] = map[string]any{"content": values[key].Content, "fileMode": values[key].FileMode}
	}
	return result
}

func configMountsForInstance(values []v1ConfigMountDef) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		result = append(result, map[string]any{
			"sourcePath":     slashPath(value.SourcePath),
			"targetPath":     value.TargetPath,
			"readOnly":       value.ReadOnly,
			"configFilePath": value.ConfigFilePath,
		})
	}
	return result
}

func domainsForCompat(values []v1DomainDef) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		result = append(result, map[string]any{"domain": value.Domain, "proxyPort": value.ProxyPort})
	}
	return result
}

func suggestedWorkspacePath(name string) string {
	if strings.TrimSpace(name) == "" {
		name = "imported-workspace"
	}
	return slashPath(filepath.Join("workspaces", name, spec.ManifestFilename))
}
