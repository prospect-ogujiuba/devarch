package resolve

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	workspacepkg "github.com/prospect-ogujiuba/devarch/internal/workspace"
)

func TestBuildRejectsMissingTemplate(t *testing.T) {
	manifestPath := writeResolveWorkspaceFixture(t, filepath.Join(t.TempDir(), "devarch.workspace.yaml"), `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: missing-template
catalog:
  sources:
    - `+filepath.ToSlash(filepath.Join(repoRoot(t), "catalog", "builtin"))+`
resources:
  api:
    template: does-not-exist
`)

	ws, err := workspacepkg.Load(manifestPath)
	if err != nil {
		t.Fatalf("workspace.Load(%s) returned error: %v", manifestPath, err)
	}
	index := loadCatalogIndex(t, ws.ResolvedCatalogSources())

	_, err = Resolve(ws, index)
	if err == nil {
		t.Fatal("expected missing template error, got nil")
	}

	var missingErr *MissingTemplateError
	if !errors.As(err, &missingErr) {
		t.Fatalf("expected MissingTemplateError, got %T (%v)", err, err)
	}
	if got, want := missingErr.ResourceKey, "api"; got != want {
		t.Fatalf("missingErr.ResourceKey = %q, want %q", got, want)
	}
	if got, want := missingErr.TemplateName, "does-not-exist"; got != want {
		t.Fatalf("missingErr.TemplateName = %q, want %q", got, want)
	}
}

func TestBuildMergesTemplateDefaultsWithWorkspaceOverrides(t *testing.T) {
	ws, index := loadExampleGraphInputs(t, "shop-local")

	graph, err := Resolve(ws, index)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	postgres := graph.Resource("postgres")
	if postgres == nil {
		t.Fatal("expected postgres resource")
	}
	if got, want := postgres.Env["POSTGRES_DB"].Text(), "shop"; got != want {
		t.Fatalf("postgres POSTGRES_DB = %q, want %q", got, want)
	}
	if got, want := postgres.Env["POSTGRES_USER"].Text(), "shop"; got != want {
		t.Fatalf("postgres POSTGRES_USER = %q, want %q", got, want)
	}
	if got, ok := postgres.Env["POSTGRES_PASSWORD"].SecretRef(); !ok || got != "postgres_password" {
		t.Fatalf("postgres POSTGRES_PASSWORD secretRef = (%q, %v), want (%q, true)", got, ok, "postgres_password")
	}
	if got, want := postgres.Exports[0].Env["DB_HOST"], "${resource.host}"; got != want {
		t.Fatalf("postgres export DB_HOST template = %q, want %q", got, want)
	}

	api := graph.Resource("api")
	if api == nil {
		t.Fatal("expected api resource")
	}
	if got, want := api.Env["NODE_ENV"].Text(), "development"; got != want {
		t.Fatalf("api NODE_ENV = %q, want %q", got, want)
	}
	if got, want := api.Env["PORT"].Text(), "3000"; got != want {
		t.Fatalf("api PORT = %q, want %q", got, want)
	}

	redis := graph.Resource("redis")
	if redis == nil {
		t.Fatal("expected redis resource")
	}
	if got, want := redis.Exports[0].Env["REDIS_URL"], "redis://:devarch@redis:6379/0"; got != want {
		t.Fatalf("redis export REDIS_URL = %q, want %q", got, want)
	}
	if got, want := redis.Exports[0].Env["REDIS_HOST"], "${resource.host}"; got != want {
		t.Fatalf("redis export REDIS_HOST = %q, want %q", got, want)
	}
}

func TestBuildMergesPortsVolumesHealthAndDevelop(t *testing.T) {
	manifestPath := writeResolveWorkspaceFixture(t, filepath.Join(t.TempDir(), "devarch.workspace.yaml"), `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: merge-check
catalog:
  sources:
    - `+filepath.ToSlash(filepath.Join(repoRoot(t), "catalog", "builtin"))+`
resources:
  api:
    template: node-api
    ports:
      - host: 8200
        container: 3000
    volumes:
      - source: workspace.node-modules
        target: /workspace/node_modules
        kind: cache
        readOnly: true
      - source: workspace.logs
        target: /workspace/logs
        kind: data
    health:
      test:
        - CMD-SHELL
        - test -f package.json
      interval: 5s
      timeout: 2s
      retries: 2
      startPeriod: 1s
    develop:
      watch:
        - path: ./app
          target: /workspace
          action: sync
`)

	ws, err := workspacepkg.Load(manifestPath)
	if err != nil {
		t.Fatalf("workspace.Load(%s) returned error: %v", manifestPath, err)
	}
	index := loadCatalogIndex(t, ws.ResolvedCatalogSources())

	graph, err := Resolve(ws, index)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	api := graph.Resource("api")
	if api == nil {
		t.Fatal("expected api resource")
	}
	if got, want := api.Ports, []Port{{Host: 8200, Container: 3000, Protocol: "tcp"}}; !slices.Equal(got, want) {
		t.Fatalf("api.Ports = %#v, want %#v", got, want)
	}
	if got, want := api.Volumes, []Volume{
		{Source: "workspace.logs", Target: "/workspace/logs", Kind: "data"},
		{Source: "workspace.node-modules", Target: "/workspace/node_modules", ReadOnly: true, Kind: "cache"},
	}; !slices.Equal(got, want) {
		t.Fatalf("api.Volumes = %#v, want %#v", got, want)
	}
	if api.Health == nil {
		t.Fatal("api.Health should be present")
	}
	if got, want := []string(api.Health.Test), []string{"CMD-SHELL", "test -f package.json"}; !slices.Equal(got, want) {
		t.Fatalf("api.Health.Test = %v, want %v", got, want)
	}
	watch := api.Develop["watch"].([]any)
	if got, want := len(watch), 1; got != want {
		t.Fatalf("len(api.Develop.watch) = %d, want %d", got, want)
	}
	entry := watch[0].(map[string]any)
	if got, want := entry["path"], "./app"; got != want {
		t.Fatalf("api.Develop.watch[0].path = %v, want %q", got, want)
	}
}

func TestBuildAttachesProjectSourceAndTemplateRuntime(t *testing.T) {
	ws, index := loadExampleGraphInputs(t, "laravel-local")

	graph, err := Resolve(ws, index)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	app := graph.Resource("app")
	if app == nil {
		t.Fatal("expected app resource")
	}
	if app.Template == nil || app.Template.Name != "laravel-app" {
		t.Fatalf("app.Template = %#v, want laravel-app", app.Template)
	}
	if app.Source == nil {
		t.Fatal("app.Source should be present")
	}
	if got, want := app.Source.Type, "project"; got != want {
		t.Fatalf("app.Source.Type = %q, want %q", got, want)
	}
	if got, want := app.Source.Path, "app"; got != want {
		t.Fatalf("app.Source.Path = %q, want %q", got, want)
	}
	if got, want := app.Source.ResolvedPath, filepath.Join(filepath.Dir(ws.ManifestPath), "app"); got != want {
		t.Fatalf("app.Source.ResolvedPath = %q, want %q", got, want)
	}
	if app.Runtime == nil {
		t.Fatal("app.Runtime should be present")
	}
	if got, want := app.Runtime.Image, "php:8.3-cli"; got != want {
		t.Fatalf("app.Runtime.Image = %q, want %q", got, want)
	}
	if got, want := []string(app.Runtime.Command), []string{"sh", "-c", "php artisan serve --host=0.0.0.0 --port=8000"}; !slices.Equal(got, want) {
		t.Fatalf("app.Runtime.Command = %v, want %v", got, want)
	}
	watch := app.Develop["watch"].([]any)
	entry := watch[0].(map[string]any)
	if got, want := entry["path"], "./app"; got != want {
		t.Fatalf("app.Develop.watch[0].path = %v, want %q", got, want)
	}
}

func TestBuildKeepsSourceOnlyRawComposePassThrough(t *testing.T) {
	ws, index := loadExampleGraphInputs(t, "compat-local")

	graph, err := Resolve(ws, index)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	postgres := graph.Resource("postgres")
	if postgres == nil {
		t.Fatal("expected postgres resource")
	}
	if postgres.Template != nil {
		t.Fatalf("postgres.Template = %#v, want nil", postgres.Template)
	}
	if postgres.Source == nil {
		t.Fatal("postgres.Source should be present")
	}
	if got, want := postgres.Source.Type, "raw-compose"; got != want {
		t.Fatalf("postgres.Source.Type = %q, want %q", got, want)
	}
	if got, want := postgres.Source.Path, "compose.yml"; got != want {
		t.Fatalf("postgres.Source.Path = %q, want %q", got, want)
	}
	if postgres.Runtime != nil {
		t.Fatalf("postgres.Runtime = %#v, want nil", postgres.Runtime)
	}
	if got, want := len(postgres.Exports), 1; got != want {
		t.Fatalf("len(postgres.Exports) = %d, want %d", got, want)
	}
	if got, want := postgres.Exports[0].Contract, "postgres"; got != want {
		t.Fatalf("postgres.Exports[0].Contract = %q, want %q", got, want)
	}
	if postgres.Exports[0].Env != nil {
		t.Fatalf("postgres.Exports[0].Env = %#v, want nil", postgres.Exports[0].Env)
	}
}

func loadExampleGraphInputs(t *testing.T, name string) (*workspacepkg.Workspace, *catalog.Index) {
	t.Helper()

	manifestPath := filepath.Join(repoRoot(t), "examples", "v2", "workspaces", name, "devarch.workspace.yaml")
	ws, err := workspacepkg.Load(manifestPath)
	if err != nil {
		t.Fatalf("workspace.Load(%s) returned error: %v", manifestPath, err)
	}
	return ws, loadCatalogIndex(t, ws.ResolvedCatalogSources())
}

func loadCatalogIndex(t *testing.T, roots []string) *catalog.Index {
	t.Helper()

	paths, err := catalog.DiscoverTemplateFiles(roots)
	if err != nil {
		t.Fatalf("catalog.DiscoverTemplateFiles(%v) returned error: %v", roots, err)
	}
	index, err := catalog.LoadIndex(paths)
	if err != nil {
		t.Fatalf("catalog.LoadIndex(%v) returned error: %v", paths, err)
	}
	return index
}

func writeResolveWorkspaceFixture(t *testing.T, path, content string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%s): %v", path, err)
	}
	return filepath.Clean(path)
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
