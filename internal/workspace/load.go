package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/prospect-ogujiuba/devarch/internal/spec"
	"gopkg.in/yaml.v3"
)

// SemanticError reports a post-schema validation error.
type SemanticError struct {
	Field   string
	Message string
}

func (e *SemanticError) Error() string {
	if e == nil || e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Load reads, validates, decodes, and normalizes a workspace manifest. The
// input path may point at either a manifest file or a directory containing the
// canonical manifest filename.
func Load(path string) (*Workspace, error) {
	manifestPath, err := resolveManifestPath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read workspace manifest %s: %w", manifestPath, err)
	}
	if err := spec.ValidateWorkspaceBytes(data); err != nil {
		return nil, fmt.Errorf("validate workspace manifest %s: %w", manifestPath, err)
	}

	var ws Workspace
	if err := yaml.Unmarshal(data, &ws); err != nil {
		return nil, fmt.Errorf("decode workspace manifest %s: %w", manifestPath, err)
	}

	ws.ManifestPath = manifestPath
	ws.ManifestDir = filepath.Dir(manifestPath)

	if err := validateSemantics(&ws); err != nil {
		return nil, err
	}
	if err := Normalize(&ws); err != nil {
		return nil, err
	}

	return &ws, nil
}

func resolveManifestPath(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", fmt.Errorf("stat workspace path %s: %w", cleanPath, err)
	}

	manifestPath := cleanPath
	if info.IsDir() {
		manifestPath = filepath.Join(cleanPath, spec.ManifestFilename)
	}

	absolutePath, err := filepath.Abs(manifestPath)
	if err != nil {
		return "", fmt.Errorf("resolve workspace manifest %s: %w", manifestPath, err)
	}
	absolutePath = filepath.Clean(absolutePath)

	manifestInfo, err := os.Stat(absolutePath)
	if err != nil {
		return "", fmt.Errorf("stat workspace manifest %s: %w", absolutePath, err)
	}
	if manifestInfo.IsDir() {
		return "", fmt.Errorf("workspace manifest %s: path is a directory", absolutePath)
	}

	return absolutePath, nil
}

func validateSemantics(ws *Workspace) error {
	for resourceKey, resource := range ws.Resources {
		if resource == nil || resource.Source == nil {
			continue
		}
		if resource.Source.Type != "raw-compose" {
			continue
		}
		if resource.Source.Service == "" {
			return &SemanticError{
				Field:   fmt.Sprintf("resources.%s.source.service", resourceKey),
				Message: "is required when source.type=raw-compose",
			}
		}
	}
	return nil
}
