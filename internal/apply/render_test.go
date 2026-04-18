package apply_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	stdruntime "runtime"
	"testing"

	"github.com/prospect-ogujiuba/devarch/internal/apply"
	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	contractspkg "github.com/prospect-ogujiuba/devarch/internal/contracts"
	resolvepkg "github.com/prospect-ogujiuba/devarch/internal/resolve"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	workspacepkg "github.com/prospect-ogujiuba/devarch/internal/workspace"
)

func TestRenderLaravelLocalGolden(t *testing.T) {
	desired := loadDesiredWorkspace(t, "laravel-local")
	payload, err := apply.Render(desired)
	if err != nil {
		t.Fatalf("apply.Render returned error: %v", err)
	}
	if payload.Blocked {
		t.Fatalf("payload unexpectedly blocked: %#v", payload.Diagnostics)
	}
	app := payload.Resource("app")
	if app == nil {
		t.Fatal("expected app payload")
	}
	if app.ProjectSource == nil {
		t.Fatal("expected project source payload")
	}
	if got, want := app.ProjectSource.ContainerPath, "/var/www/html"; got != want {
		t.Fatalf("app.ProjectSource.ContainerPath = %q, want %q", got, want)
	}
	if got, want := app.DevelopWatch[0].Target, "/var/www/html"; got != want {
		t.Fatalf("app.DevelopWatch[0].Target = %q, want %q", got, want)
	}
	if got, want := app.Health.Interval, "30s"; got != want {
		t.Fatalf("app.Health.Interval = %q, want %q", got, want)
	}
	if got, want := app.Labels["devarch.example"], "laravel-local"; got != want {
		t.Fatalf("app.Labels[devarch.example] = %q, want %q", got, want)
	}

	actual := marshalJSON(t, payload)
	goldenPath := filepath.Join(repoRoot(t), "testdata", "goldens", "phase3", "laravel-local.render.golden.json")
	if updateGoldens() {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(goldenPath), err)
		}
		if err := os.WriteFile(goldenPath, actual, 0o644); err != nil {
			t.Fatalf("os.WriteFile(%s): %v", goldenPath, err)
		}
	}
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%s): %v", goldenPath, err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatalf("golden mismatch\n--- actual ---\n%s\n--- expected ---\n%s", actual, expected)
	}
}

func loadDesiredWorkspace(t *testing.T, name string) *runtimepkg.DesiredWorkspace {
	t.Helper()
	manifestPath := filepath.Join(repoRoot(t), "examples", "v2", "workspaces", name, "devarch.workspace.yaml")
	ws, err := workspacepkg.Load(manifestPath)
	if err != nil {
		t.Fatalf("workspace.Load(%s): %v", manifestPath, err)
	}
	paths, err := catalog.DiscoverTemplateFiles(ws.ResolvedCatalogSources())
	if err != nil {
		t.Fatalf("catalog.DiscoverTemplateFiles returned error: %v", err)
	}
	index, err := catalog.LoadIndex(paths)
	if err != nil {
		t.Fatalf("catalog.LoadIndex returned error: %v", err)
	}
	graph, err := resolvepkg.Resolve(ws, index)
	if err != nil {
		t.Fatalf("resolve.Resolve returned error: %v", err)
	}
	contracts := contractspkg.Resolve(graph)
	desired, err := runtimepkg.BuildDesiredWorkspace(graph, contracts)
	if err != nil {
		t.Fatalf("runtime.BuildDesiredWorkspace returned error: %v", err)
	}
	return desired
}

func marshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.MarshalIndent(value, "", "  ")
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
	_, file, _, ok := stdruntime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
