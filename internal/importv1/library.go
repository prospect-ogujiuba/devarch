package importv1

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/prospect-ogujiuba/devarch/internal/spec"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
	"gopkg.in/yaml.v3"
)

type templateDocument struct {
	APIVersion string           `yaml:"apiVersion"`
	Kind       string           `yaml:"kind"`
	Metadata   templateMetadata `yaml:"metadata"`
	Spec       templateSpec     `yaml:"spec"`
}

type templateMetadata struct {
	Name        string   `yaml:"name"`
	Tags        []string `yaml:"tags,omitempty"`
	Description string   `yaml:"description,omitempty"`
}

type templateSpec struct {
	Runtime map[string]any   `yaml:"runtime"`
	Env     map[string]any   `yaml:"env,omitempty"`
	Ports   []workspace.Port `yaml:"ports,omitempty"`
	Volumes []workspace.Volume `yaml:"volumes,omitempty"`
	Imports []workspace.Import `yaml:"imports,omitempty"`
	Exports []workspace.Export `yaml:"exports,omitempty"`
	Health  map[string]any   `yaml:"health,omitempty"`
	Develop map[string]any   `yaml:"develop,omitempty"`
}

func importLibrary(root string) (*Result, error) {
	result := newResult(ModeV1Library, root)

	composeFiles, err := discoverComposeFiles(root)
	if err != nil {
		return nil, err
	}
	if len(composeFiles) == 0 {
		result.Diagnostics = append(result.Diagnostics, failure("library-no-compose-files", fmt.Sprintf("directory %s does not contain any compose.yml files", root), root, ""))
		return finalizeResult(result), nil
	}

	for _, composePath := range composeFiles {
		services, err := parseComposeServices(composePath)
		if err != nil {
			name := filepath.Base(filepath.Dir(composePath))
			artifact := Artifact{
				Kind:          ArtifactKindTemplate,
				Name:          name,
				SourcePath:    composePath,
				SuggestedPath: slashPath(filepath.Join("catalog", "imported", categoryFromComposePath(composePath), name, "template.yaml")),
				Status:        StatusRejected,
				Diagnostics: []Diagnostic{
					failure("library-compose-parse-failed", err.Error(), composePath, ""),
				},
			}
			result.Artifacts = append(result.Artifacts, artifact)
			continue
		}
		for _, service := range services {
			result.Artifacts = append(result.Artifacts, importLibraryService(service))
		}
	}

	return finalizeResult(result), nil
}

func importLibraryService(service parsedComposeService) Artifact {
	artifact := Artifact{
		Kind:          ArtifactKindTemplate,
		Name:          service.Name,
		SourcePath:    service.SourcePath,
		SuggestedPath: suggestedTemplatePath(service.Category, service.Name),
	}

	doc, diagnostics, ok := buildTemplateDocument(service)
	artifact.Diagnostics = append(artifact.Diagnostics, diagnostics...)
	if !ok {
		artifact.Status = StatusRejected
		return artifact
	}

	data, err := yaml.Marshal(doc)
	if err != nil {
		artifact.Status = StatusRejected
		artifact.Diagnostics = append(artifact.Diagnostics, failure("template-marshal-failed", err.Error(), service.SourcePath, ""))
		return artifact
	}
	if err := spec.ValidateTemplateBytes(data); err != nil {
		artifact.Status = StatusRejected
		artifact.Diagnostics = append(artifact.Diagnostics, validationDiagnostics("template-validation-failed", err, service.SourcePath)...)
		return artifact
	}

	artifact.Document = string(data)
	artifact.Status = deriveArtifactStatus(artifact.Document, artifact.Diagnostics)
	return artifact
}

func buildTemplateDocument(service parsedComposeService) (*templateDocument, []Diagnostic, bool) {
	diagnostics := make([]Diagnostic, 0)
	if len(service.Runtime) == 0 {
		diagnostics = append(diagnostics, failure("template-runtime-missing", fmt.Sprintf("service %q does not define an image or safe build context", service.Name), service.SourcePath, "spec.runtime"))
		return nil, diagnostics, false
	}

	tags := compactStrings([]string{service.Category})
	sort.Strings(tags)
	volumes, volumeCompat, volumeDiagnostics := templateVolumesForService(service)
	diagnostics = append(diagnostics, volumeDiagnostics...)

	doc := &templateDocument{
		APIVersion: "devarch.io/v2alpha1",
		Kind:       "Template",
		Metadata: templateMetadata{
			Name: service.Name,
			Tags: tags,
		},
		Spec: templateSpec{
			Runtime: cloneAnyMap(service.Runtime),
			Env:     cloneAnyMap(service.Env),
			Ports:   append([]workspace.Port(nil), service.Ports...),
			Volumes: volumes,
			Health:  cloneAnyMap(service.Health),
		},
	}

	compat := make(map[string]any)
	compat["sourceComposePath"] = slashPath(service.SourcePath)
	compat["contracts"] = map[string]any{
		"imports": "unavailable-from-v1-compose",
		"exports": "unavailable-from-v1-compose",
	}
	diagnostics = append(diagnostics, warning("template-contracts-unavailable", "V1 compose library inputs do not encode contract metadata, so imports/exports were emitted empty.", service.SourcePath, "spec"))

	compatFields := make([]string, 0)
	if service.ContainerName != "" {
		compat["containerName"] = service.ContainerName
		compatFields = append(compatFields, "containerName")
	}
	if service.RestartPolicy != "" {
		compat["restartPolicy"] = service.RestartPolicy
		compatFields = append(compatFields, "restartPolicy")
	}
	if service.User != "" {
		compat["user"] = service.User
		compatFields = append(compatFields, "user")
	}
	if len(service.EnvFiles) > 0 {
		compat["envFiles"] = append([]string(nil), service.EnvFiles...)
		compatFields = append(compatFields, "envFiles")
	}
	if len(service.Dependencies) > 0 {
		compat["dependsOn"] = append([]string(nil), service.Dependencies...)
		compatFields = append(compatFields, "dependsOn")
	}
	if len(service.Networks) > 0 {
		compat["networks"] = append([]string(nil), service.Networks...)
		compatFields = append(compatFields, "networks")
	}
	if len(service.Labels) > 0 {
		compat["labels"] = cloneStringMap(service.Labels)
		compatFields = append(compatFields, "labels")
	}
	if len(volumeCompat) > 0 {
		compat["volumes"] = volumeCompat
		compatFields = append(compatFields, "volumes")
	}
	if len(service.ConfigMounts) > 0 {
		compat["configMounts"] = configMountsForCompat(service.ConfigMounts)
		compatFields = append(compatFields, "configMounts")
	}
	if len(service.ConfigFiles) > 0 {
		compat["configFiles"] = configFilesForCompat(service.ConfigFiles)
		compatFields = append(compatFields, "configFiles")
	}
	if len(compatFields) > 0 {
		sort.Strings(compatFields)
		diagnostics = append(diagnostics, warning("template-compat-fields-preserved", fmt.Sprintf("Preserved non-schema-native V1 fields under spec.develop.importv1: %s.", strings.Join(compatFields, ", ")), service.SourcePath, "spec.develop.importv1"))
	}
	if len(compat) > 0 {
		doc.Spec.Develop = map[string]any{"importv1": compat}
	}

	return doc, diagnostics, true
}

func templateVolumesForService(service parsedComposeService) ([]workspace.Volume, []map[string]any, []Diagnostic) {
	volumes := make([]workspace.Volume, 0, len(service.Volumes))
	compat := make([]map[string]any, 0)
	diagnostics := make([]Diagnostic, 0)
	compatSources := make([]string, 0)
	for _, volume := range service.Volumes {
		switch volume.Type {
		case "named":
			kind := "named"
			if volume.External {
				kind = "external"
			}
			volumes = append(volumes, workspace.Volume{Source: volume.Source, Target: volume.Target, ReadOnly: volume.ReadOnly, Kind: kind})
		case "bind":
			rebased, safe := rebasePathWithinService(service.ServiceDir, volume.Source)
			if !safe {
				compatSources = append(compatSources, volume.Source)
				compat = append(compat, map[string]any{
					"source":   slashPath(volume.Source),
					"target":   volume.Target,
					"readOnly": volume.ReadOnly,
					"kind":     "bind",
				})
				continue
			}
			volumes = append(volumes, workspace.Volume{Source: rebased, Target: volume.Target, ReadOnly: volume.ReadOnly, Kind: "bind"})
		}
	}
	if len(compatSources) > 0 {
		sort.Strings(compatSources)
		diagnostics = append(diagnostics, warning("template-bind-volume-preserved-compat", fmt.Sprintf("Bind volume sources %s were preserved in compatibility data instead of first-class template volumes because their paths escape the service entry.", strings.Join(compatSources, ", ")), service.SourcePath, "spec.develop.importv1"))
	}
	return volumes, compat, diagnostics
}

func suggestedTemplatePath(category, name string) string {
	category = strings.TrimSpace(category)
	if category == "" {
		category = "uncategorized"
	}
	return slashPath(filepath.Join("catalog", "imported", category, name, "template.yaml"))
}

func configMountsForCompat(values []parsedConfigMount) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		result = append(result, map[string]any{
			"sourcePath": slashPath(value.SourcePath),
			"targetPath": value.TargetPath,
			"readOnly":   value.ReadOnly,
		})
	}
	return result
}

func configFilesForCompat(values map[string]importedConfigFile) map[string]any {
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
		result[key] = map[string]any{
			"content":  values[key].Content,
			"fileMode": values[key].FileMode,
		}
	}
	return result
}

func cloneAnyMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = cloneAnyValue(value)
	}
	return cloned
}

func cloneAnyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneAnyMap(typed)
	case []string:
		return append([]string(nil), typed...)
	case []any:
		cloned := make([]any, len(typed))
		for i := range typed {
			cloned[i] = cloneAnyValue(typed[i])
		}
		return cloned
	default:
		return typed
	}
}

func validationDiagnostics(code string, err error, path string) []Diagnostic {
	validationErr, ok := err.(*spec.ValidationErrors)
	if !ok {
		return []Diagnostic{failure(code, err.Error(), path, "")}
	}
	diagnostics := make([]Diagnostic, 0, len(validationErr.Errors))
	for _, detail := range validationErr.Errors {
		diagnostics = append(diagnostics, failure(code, detail.Message, path, detail.Field))
	}
	return diagnostics
}
