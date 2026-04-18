package resolve

import (
	"fmt"
	"path/filepath"

	"github.com/prospect-ogujiuba/devarch/internal/workspace"
	"gopkg.in/yaml.v3"
)

type runtimeDocument struct {
	Image      string               `yaml:"image"`
	Build      *buildDocument       `yaml:"build,omitempty"`
	Command    workspace.StringList `yaml:"command,omitempty"`
	Entrypoint workspace.StringList `yaml:"entrypoint,omitempty"`
	WorkingDir string               `yaml:"workingDir,omitempty"`
}

type buildDocument struct {
	Context    string                        `yaml:"context"`
	Dockerfile string                        `yaml:"dockerfile,omitempty"`
	Target     string                        `yaml:"target,omitempty"`
	Args       map[string]workspace.EnvValue `yaml:"args,omitempty"`
}

func decodeRuntime(raw map[string]any, templatePath string) (*Runtime, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	data, err := yaml.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshal runtime block: %w", err)
	}

	var document runtimeDocument
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("decode runtime block: %w", err)
	}

	runtime := &Runtime{
		Image:      document.Image,
		Command:    cloneStringList(document.Command),
		Entrypoint: cloneStringList(document.Entrypoint),
		WorkingDir: document.WorkingDir,
	}
	if document.Build != nil {
		runtime.Build = &Build{
			Context:    normalizeDisplayPath(document.Build.Context),
			Dockerfile: normalizeDisplayPath(document.Build.Dockerfile),
			Target:     document.Build.Target,
			Args:       cloneEnvMap(document.Build.Args),
		}

		templateDir := filepath.Dir(templatePath)
		runtime.Build.ResolvedContext = resolvePath(templateDir, runtime.Build.Context)
		dockerfileBase := runtime.Build.ResolvedContext
		if dockerfileBase == "" {
			dockerfileBase = templateDir
		}
		runtime.Build.ResolvedDockerfile = resolvePath(dockerfileBase, runtime.Build.Dockerfile)
	}

	return runtime, nil
}

func cloneStringList(values workspace.StringList) workspace.StringList {
	if len(values) == 0 {
		return nil
	}
	return append(workspace.StringList(nil), values...)
}
