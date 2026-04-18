package catalog

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestDiscoverTemplateFilesReturnsSortedCanonicalPaths(t *testing.T) {
	root := t.TempDir()
	rootA := filepath.Join(root, "catalog-a")
	rootB := filepath.Join(root, "catalog-b")

	postgresPath := writeCatalogFixture(t, filepath.Join(rootA, "database", "postgres", TemplateFilename), "apiVersion: devarch.io/v2alpha1\n")
	nodeAPIPath := writeCatalogFixture(t, filepath.Join(rootA, "backend", "node-api", TemplateFilename), "apiVersion: devarch.io/v2alpha1\n")
	vitePath := writeCatalogFixture(t, filepath.Join(rootB, "frontend", "vite-web", TemplateFilename), "apiVersion: devarch.io/v2alpha1\n")

	writeCatalogFixture(t, filepath.Join(rootA, "backend", "node-api", "template.yml"), "ignored\n")
	writeCatalogFixture(t, filepath.Join(rootA, "README.md"), "ignored\n")

	got, err := DiscoverTemplateFiles([]string{rootB, filepath.Join(rootA, "backend"), rootA})
	if err != nil {
		t.Fatalf("DiscoverTemplateFiles returned error: %v", err)
	}

	want := []string{postgresPath, nodeAPIPath, vitePath}
	sort.Strings(want)
	if !slices.Equal(got, want) {
		t.Fatalf("DiscoverTemplateFiles returned %v, want %v", got, want)
	}
}

func TestDiscoverTemplateFilesRejectsNonDirectoryRoot(t *testing.T) {
	root := t.TempDir()
	filePath := writeCatalogFixture(t, filepath.Join(root, "not-a-directory.txt"), "not a directory\n")

	_, err := DiscoverTemplateFiles([]string{filePath})
	if err == nil {
		t.Fatal("expected error for non-directory root, got nil")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("expected non-directory error, got %v", err)
	}
}

func writeCatalogFixture(t *testing.T, path, content string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%s): %v", path, err)
	}
	return filepath.Clean(path)
}
