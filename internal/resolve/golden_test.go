package resolve_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	contractspkg "github.com/prospect-ogujiuba/devarch/internal/contracts"
	resolvepkg "github.com/prospect-ogujiuba/devarch/internal/resolve"
	workspacepkg "github.com/prospect-ogujiuba/devarch/internal/workspace"
)

func TestPhase2Goldens(t *testing.T) {
	tests := []struct {
		name         string
		manifestPath string
		goldenPath   string
	}{
		{
			name:         "shop-local",
			manifestPath: filepath.Join(repoRoot(t), "examples", "v2", "workspaces", "shop-local", "devarch.workspace.yaml"),
			goldenPath:   filepath.Join(repoRoot(t), "testdata", "goldens", "phase2", "shop-local.resolved.golden.json"),
		},
		{
			name:         "laravel-local",
			manifestPath: filepath.Join(repoRoot(t), "examples", "v2", "workspaces", "laravel-local", "devarch.workspace.yaml"),
			goldenPath:   filepath.Join(repoRoot(t), "testdata", "goldens", "phase2", "laravel-local.resolved.golden.json"),
		},
		{
			name:         "ambiguous-http",
			manifestPath: filepath.Join(repoRoot(t), "testdata", "goldens", "phase2", "fixtures", "ambiguous-http", "devarch.workspace.yaml"),
			goldenPath:   filepath.Join(repoRoot(t), "testdata", "goldens", "phase2", "ambiguous-http.resolved.golden.json"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := renderGoldenOutput(t, tt.manifestPath)
			if updateGoldens() {
				if err := os.WriteFile(tt.goldenPath, actual, 0o644); err != nil {
					t.Fatalf("os.WriteFile(%s): %v", tt.goldenPath, err)
				}
			}

			expected, err := os.ReadFile(tt.goldenPath)
			if err != nil {
				t.Fatalf("os.ReadFile(%s): %v", tt.goldenPath, err)
			}
			if !bytes.Equal(actual, expected) {
				t.Fatalf("golden mismatch for %s\n--- actual ---\n%s\n--- expected ---\n%s", tt.name, actual, expected)
			}
		})
	}
}

type goldenOutput struct {
	Workspace   resolvepkg.Workspace      `json:"workspace"`
	Resources   []*resolvepkg.Resource    `json:"resources"`
	Links       []contractspkg.Link       `json:"links,omitempty"`
	Diagnostics []contractspkg.Diagnostic `json:"diagnostics,omitempty"`
}

func renderGoldenOutput(t *testing.T, manifestPath string) []byte {
	t.Helper()

	ws, err := workspacepkg.Load(manifestPath)
	if err != nil {
		t.Fatalf("workspace.Load(%s) returned error: %v", manifestPath, err)
	}
	paths, err := catalog.DiscoverTemplateFiles(ws.ResolvedCatalogSources())
	if err != nil {
		t.Fatalf("catalog.DiscoverTemplateFiles(%v) returned error: %v", ws.ResolvedCatalogSources(), err)
	}
	index, err := catalog.LoadIndex(paths)
	if err != nil {
		t.Fatalf("catalog.LoadIndex(%v) returned error: %v", paths, err)
	}
	graph, err := resolvepkg.Resolve(ws, index)
	if err != nil {
		t.Fatalf("resolve.Resolve returned error: %v", err)
	}
	contracts := contractspkg.Resolve(graph)

	output := goldenOutput{
		Workspace:   graph.Workspace,
		Resources:   graph.Resources,
		Links:       contracts.Links,
		Diagnostics: contracts.Diagnostics,
	}
	encoded, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent returned error: %v", err)
	}
	return append(encoded, '\n')
}

func updateGoldens() bool {
	return os.Getenv("DEVARCH_UPDATE_GOLDENS") == "1"
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
