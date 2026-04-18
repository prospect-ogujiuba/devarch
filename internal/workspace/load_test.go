package workspace

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLoadAcceptsManifestFileOrDirectory(t *testing.T) {
	dirPath := filepath.Join(repoRoot(t), "examples", "v2", "workspaces", "shop-local")
	filePath := filepath.Join(dirPath, "devarch.workspace.yaml")

	fromDir, err := Load(dirPath)
	if err != nil {
		t.Fatalf("Load(%s) returned error: %v", dirPath, err)
	}
	fromFile, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load(%s) returned error: %v", filePath, err)
	}

	if got, want := fromDir.Metadata.Name, "shop-local"; got != want {
		t.Fatalf("fromDir.Metadata.Name = %q, want %q", got, want)
	}
	if got, want := fromFile.Metadata.Name, "shop-local"; got != want {
		t.Fatalf("fromFile.Metadata.Name = %q, want %q", got, want)
	}
	if fromDir.ManifestPath != fromFile.ManifestPath {
		t.Fatalf("ManifestPath mismatch: dir=%s file=%s", fromDir.ManifestPath, fromFile.ManifestPath)
	}
	if got, want := filepath.Base(fromDir.ManifestPath), "devarch.workspace.yaml"; got != want {
		t.Fatalf("ManifestPath base = %q, want %q", got, want)
	}
	if got, want := len(fromDir.ResolvedCatalogSources()), 1; got != want {
		t.Fatalf("len(ResolvedCatalogSources()) = %d, want %d", got, want)
	}
}

func TestLoadRejectsMissingManifestInDirectory(t *testing.T) {
	root := t.TempDir()

	_, err := Load(root)
	if err == nil {
		t.Fatal("expected missing manifest error, got nil")
	}
	if !strings.Contains(err.Error(), "devarch.workspace.yaml") {
		t.Fatalf("expected missing manifest error to mention canonical filename, got %v", err)
	}
}

func TestLoadRejectsInvalidWorkspaceDocument(t *testing.T) {
	manifestPath := writeWorkspaceFixture(t, filepath.Join(t.TempDir(), "devarch.workspace.yaml"), `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: invalid
resources:
  api:
    exports:
      - contract: http
`)

	_, err := Load(manifestPath)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "validate workspace manifest") {
		t.Fatalf("expected validation error prefix, got %v", err)
	}
}

func TestLoadRejectsRawComposeWithoutService(t *testing.T) {
	manifestPath := writeWorkspaceFixture(t, filepath.Join(t.TempDir(), "devarch.workspace.yaml"), `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: compat
resources:
  postgres:
    source:
      type: raw-compose
      path: ./compose.yml
`)

	_, err := Load(manifestPath)
	if err == nil {
		t.Fatal("expected semantic validation error, got nil")
	}
	if !strings.Contains(err.Error(), "resources.postgres.source.service") {
		t.Fatalf("expected raw-compose semantic error, got %v", err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
