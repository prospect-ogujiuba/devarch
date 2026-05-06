package catalog

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestLoadIndexBuildsDeterministicLookups(t *testing.T) {
	root := t.TempDir()
	postgresPath := writeCatalogFixture(t, filepath.Join(root, "database", "postgres", TemplateFilename), `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: postgres
  tags:
    - database
    - sql
spec:
  runtime:
    image: postgres:16
  exports:
    - postgres
`)
	nodeAPIPath := writeCatalogFixture(t, filepath.Join(root, "backend", "node-api", TemplateFilename), `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: node-api
  tags:
    - backend
    - node
spec:
  runtime:
    image: node:22-alpine
  imports:
    - contract: postgres
  exports:
    - contract: http
      env:
        API_URL: "http://${resource.host}:${resource.port.3000}"
`)
	nginxPath := writeCatalogFixture(t, filepath.Join(root, "proxy", "nginx", TemplateFilename), `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: nginx
  tags:
    - proxy
spec:
  runtime:
    image: nginx:1.27-alpine
  imports:
    - contract: http
  exports:
    - contract: http
      env:
        PROXY_URL: "http://${resource.host}:${resource.port.80}"
`)

	index, err := LoadIndex([]string{nginxPath, postgresPath, nodeAPIPath, postgresPath})
	if err != nil {
		t.Fatalf("LoadIndex returned error: %v", err)
	}

	if got, want := templateNames(index.Templates()), []string{"nginx", "node-api", "postgres"}; !slices.Equal(got, want) {
		t.Fatalf("Templates() names = %v, want %v", got, want)
	}

	postgres, ok := index.ByName("postgres")
	if !ok {
		t.Fatal("expected ByName(postgres) to succeed")
	}
	if postgres.Path != postgresPath {
		t.Fatalf("postgres.Path = %s, want %s", postgres.Path, postgresPath)
	}
	if got, want := postgres.Spec.Exports[0].Contract, "postgres"; got != want {
		t.Fatalf("postgres export contract = %q, want %q", got, want)
	}
}

func TestLoadIndexRejectsDuplicateTemplateNames(t *testing.T) {
	root := t.TempDir()
	firstPath := writeCatalogFixture(t, filepath.Join(root, "database", "postgres", TemplateFilename), `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: postgres
spec:
  runtime:
    image: postgres:16
`)
	secondPath := writeCatalogFixture(t, filepath.Join(root, "custom", "postgres-alt", TemplateFilename), `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: postgres
spec:
  runtime:
    image: postgres:17
`)

	_, err := LoadIndex([]string{secondPath, firstPath})
	if err == nil {
		t.Fatal("expected duplicate template name error, got nil")
	}

	var duplicateErr *DuplicateTemplateNameError
	if !errors.As(err, &duplicateErr) {
		t.Fatalf("expected DuplicateTemplateNameError, got %T (%v)", err, err)
	}
	if duplicateErr.Name != "postgres" {
		t.Fatalf("duplicateErr.Name = %q, want postgres", duplicateErr.Name)
	}
	paths := []string{duplicateErr.FirstPath, duplicateErr.SecondPath}
	slices.Sort(paths)
	wantPaths := []string{firstPath, secondPath}
	slices.Sort(wantPaths)
	if !slices.Equal(paths, wantPaths) {
		t.Fatalf("duplicateErr paths = %v, want %v", paths, wantPaths)
	}
	if !strings.Contains(err.Error(), firstPath) || !strings.Contains(err.Error(), secondPath) {
		t.Fatalf("expected duplicate error to include both paths, got %v", err)
	}
}

func TestLoadIndexRejectsInvalidTemplateDocuments(t *testing.T) {
	root := t.TempDir()
	invalidPath := writeCatalogFixture(t, filepath.Join(root, "backend", "broken", TemplateFilename), `apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: broken
spec:
  exports:
    - contract: http
`)

	_, err := LoadIndex([]string{invalidPath})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "validate template") {
		t.Fatalf("expected validation error prefix, got %v", err)
	}
	if !strings.Contains(err.Error(), invalidPath) {
		t.Fatalf("expected validation error to mention %s, got %v", invalidPath, err)
	}
}

func templateNames(templates []*Template) []string {
	names := make([]string, 0, len(templates))
	for _, template := range templates {
		names = append(names, template.Metadata.Name)
	}
	return names
}
