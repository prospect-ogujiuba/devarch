package spec

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestValidateWorkspaceExamples(t *testing.T) {
	examples := []string{"shop-local", "laravel-local", "compat-local"}

	for _, example := range examples {
		example := example
		t.Run(example, func(t *testing.T) {
			manifestPath := filepath.Join(repoRoot(t), "examples", "v2", "workspaces", example, ManifestFilename)
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				t.Fatalf("os.ReadFile(%s): %v", manifestPath, err)
			}
			if err := ValidateWorkspaceBytes(data); err != nil {
				t.Fatalf("ValidateWorkspaceBytes(%s) returned error: %v", manifestPath, err)
			}
		})
	}
}

func TestValidateWorkspaceRejectsMalformedDocuments(t *testing.T) {
	tests := []struct {
		name      string
		document  string
		wantField string
		wantText  string
	}{
		{
			name: "missing metadata name",
			document: `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  displayName: Broken
resources:
  api:
    template: node-api
`,
			wantField: "metadata",
			wantText:  "name is required",
		},
		{
			name: "malformed import entry",
			document: `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: broken-import
resources:
  api:
    template: node-api
    imports:
      - from: postgres
`,
			wantField: "resources.api.imports.0",
			wantText:  "contract is required",
		},
		{
			name: "malformed export entry",
			document: `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: broken-export
resources:
  api:
    template: node-api
    exports:
      - env:
          APP_URL: http://localhost
`,
			wantField: "resources.api.exports.0",
			wantText:  "Must validate one and only one schema",
		},
		{
			name: "invalid source type",
			document: `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: broken-source
resources:
  api:
    source:
      type: archive
      path: ./app
`,
			wantField: "resources.api.source.type",
			wantText:  "must be one of the following",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkspaceBytes([]byte(tt.document))
			assertValidationErrorContains(t, err, tt.wantField, tt.wantText)
		})
	}
}

func TestValidateTemplateDocuments(t *testing.T) {
	validTemplate := `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: postgres
  tags: [database, sql]
  description: PostgreSQL database template
spec:
  runtime:
    image: postgres:16
  env:
    POSTGRES_DB: app
    POSTGRES_USER: app
    POSTGRES_PASSWORD: devarch
  ports:
    - container: 5432
  volumes:
    - target: /var/lib/postgresql/data
      kind: data
  exports:
    - contract: postgres
      env:
        DATABASE_URL: postgres://${env.POSTGRES_USER}:${env.POSTGRES_PASSWORD}@${resource.host}:${resource.port.5432}/${env.POSTGRES_DB}
  health:
    test:
      - CMD-SHELL
      - pg_isready -U ${env.POSTGRES_USER}
`
	if err := ValidateTemplateBytes([]byte(validTemplate)); err != nil {
		t.Fatalf("ValidateTemplateBytes(validTemplate) returned error: %v", err)
	}

	tests := []struct {
		name      string
		document  string
		wantField string
		wantText  string
	}{
		{
			name: "missing metadata name",
			document: `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  description: Broken
spec:
  runtime:
    image: postgres:16
`,
			wantField: "metadata",
			wantText:  "name is required",
		},
		{
			name: "malformed import entry",
			document: `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: broken-import
spec:
  runtime:
    image: node:22-alpine
  imports:
    - from: postgres
`,
			wantField: "spec.imports.0",
			wantText:  "contract is required",
		},
		{
			name: "malformed export entry",
			document: `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: broken-export
spec:
  runtime:
    image: node:22-alpine
  exports:
    - env:
        APP_URL: http://localhost
`,
			wantField: "spec.exports.0",
			wantText:  "Must validate one and only one schema",
		},
		{
			name: "reject contracts block in favor of imports and exports",
			document: `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: broken-contracts
spec:
  runtime:
    image: node:22-alpine
  contracts:
    provided:
      - http
`,
			wantField: "spec",
			wantText:  "Additional property contracts is not allowed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplateBytes([]byte(tt.document))
			assertValidationErrorContains(t, err, tt.wantField, tt.wantText)
		})
	}
}

func TestValidateBuiltinTemplates(t *testing.T) {
	templatePaths := builtinTemplatePaths(t)
	if len(templatePaths) == 0 {
		t.Fatal("expected builtin template corpus, found no template.yaml files")
	}

	relativePaths := make(map[string]struct{}, len(templatePaths))
	for _, path := range templatePaths {
		relativePath, err := filepath.Rel(repoRoot(t), path)
		if err != nil {
			t.Fatalf("filepath.Rel(%s): %v", path, err)
		}
		relativePaths[filepath.Clean(relativePath)] = struct{}{}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("os.ReadFile(%s): %v", path, err)
		}
		if err := ValidateTemplateBytes(data); err != nil {
			t.Fatalf("ValidateTemplateBytes(%s) returned error: %v", path, err)
		}
	}

	required := []string{
		filepath.Join("catalog", "builtin", "database", "postgres", "template.yaml"),
		filepath.Join("catalog", "builtin", "cache", "redis", "template.yaml"),
		filepath.Join("catalog", "builtin", "backend", "node-api", "template.yaml"),
		filepath.Join("catalog", "builtin", "backend", "laravel-app", "template.yaml"),
		filepath.Join("catalog", "builtin", "frontend", "vite-web", "template.yaml"),
		filepath.Join("catalog", "builtin", "proxy", "nginx", "template.yaml"),
	}

	for _, requiredPath := range required {
		if _, ok := relativePaths[filepath.Clean(requiredPath)]; !ok {
			t.Fatalf("expected builtin template path %s to exist", requiredPath)
		}
	}
}

func assertValidationErrorContains(t *testing.T, err error, wantField, wantText string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErrs *ValidationErrors
	if !errors.As(err, &validationErrs) {
		t.Fatalf("expected *ValidationErrors, got %T (%v)", err, err)
	}

	for _, validationErr := range validationErrs.Errors {
		if validationErr.Field == wantField && strings.Contains(validationErr.Message, wantText) {
			return
		}
	}

	t.Fatalf("expected validation error containing field %q and text %q, got %#v", wantField, wantText, validationErrs.Errors)
}

func builtinTemplatePaths(t *testing.T) []string {
	t.Helper()

	root := filepath.Join(repoRoot(t), "catalog", "builtin")
	paths := make([]string, 0)
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == "template.yaml" {
			paths = append(paths, filepath.Clean(path))
		}
		return nil
	}); err != nil {
		t.Fatalf("filepath.WalkDir(%s): %v", root, err)
	}

	sort.Strings(paths)
	return paths
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
