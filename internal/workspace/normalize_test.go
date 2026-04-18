package workspace

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadNormalizesExamples(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T, ws *Workspace)
	}{
		{
			name: "shop-local",
			check: func(t *testing.T, ws *Workspace) {
				t.Helper()

				if got, want := ws.SortedResourceKeys(), []string{"api", "postgres", "redis", "web"}; !slices.Equal(got, want) {
					t.Fatalf("SortedResourceKeys() = %v, want %v", got, want)
				}

				postgres := ws.Resources["postgres"]
				if !postgres.EnabledValue() {
					t.Fatal("postgres should remain enabled")
				}
				if got, want := len(postgres.Exports), 1; got != want {
					t.Fatalf("len(postgres.Exports) = %d, want %d", got, want)
				}
				if got, want := postgres.Exports[0].Contract, "postgres"; got != want {
					t.Fatalf("postgres export contract = %q, want %q", got, want)
				}
				if postgres.Exports[0].Env != nil {
					t.Fatalf("scalar postgres export should normalize to nil env, got %#v", postgres.Exports[0].Env)
				}

				redis := ws.Resources["redis"]
				if got, want := redis.Exports[0].Env["REDIS_URL"], "redis://:devarch@redis:6379/0"; got != want {
					t.Fatalf("redis export env REDIS_URL = %q, want %q", got, want)
				}
			},
		},
		{
			name: "laravel-local",
			check: func(t *testing.T, ws *Workspace) {
				t.Helper()

				for _, key := range []string{"app", "postgres", "redis"} {
					if !ws.Resources[key].EnabledValue() {
						t.Fatalf("resource %s should default enabled=true", key)
					}
				}

				app := ws.Resources["app"]
				if app.Source == nil {
					t.Fatal("app source should be present")
				}
				if got, want := app.Source.Path, "app"; got != want {
					t.Fatalf("app.Source.Path = %q, want %q", got, want)
				}
				if got, want := app.Source.ResolvedPath, filepath.Join(filepath.Dir(ws.ManifestPath), "app"); got != want {
					t.Fatalf("app.Source.ResolvedPath = %q, want %q", got, want)
				}
				if got, want := app.Domains, []string{"laravel.local.test"}; !slices.Equal(got, want) {
					t.Fatalf("app.Domains = %v, want %v", got, want)
				}
				labels, ok := app.Overrides["labels"].(map[string]any)
				if !ok {
					t.Fatalf("app.Overrides.labels type = %T, want map[string]any", app.Overrides["labels"])
				}
				if got, want := labels["devarch.example"], "laravel-local"; got != want {
					t.Fatalf("override label = %v, want %q", got, want)
				}
				watch, ok := app.Develop["watch"].([]any)
				if !ok || len(watch) != 1 {
					t.Fatalf("app.Develop.watch = %#v, want one entry", app.Develop["watch"])
				}
			},
		},
		{
			name: "compat-local",
			check: func(t *testing.T, ws *Workspace) {
				t.Helper()

				if got, want := ws.SortedResourceKeys(), []string{"postgres", "redis"}; !slices.Equal(got, want) {
					t.Fatalf("SortedResourceKeys() = %v, want %v", got, want)
				}

				postgres := ws.Resources["postgres"]
				if postgres.Source == nil {
					t.Fatal("postgres source should be present")
				}
				if got, want := postgres.Source.Type, "raw-compose"; got != want {
					t.Fatalf("postgres.Source.Type = %q, want %q", got, want)
				}
				if got, want := postgres.Source.Path, "compose.yml"; got != want {
					t.Fatalf("postgres.Source.Path = %q, want %q", got, want)
				}
				if got, want := postgres.Source.ResolvedPath, filepath.Join(filepath.Dir(ws.ManifestPath), "compose.yml"); got != want {
					t.Fatalf("postgres.Source.ResolvedPath = %q, want %q", got, want)
				}
				if got, want := postgres.Source.Service, "postgres"; got != want {
					t.Fatalf("postgres.Source.Service = %q, want %q", got, want)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			manifestPath := filepath.Join(repoRoot(t), "examples", "v2", "workspaces", tt.name, "devarch.workspace.yaml")
			ws, err := Load(manifestPath)
			if err != nil {
				t.Fatalf("Load(%s) returned error: %v", manifestPath, err)
			}
			tt.check(t, ws)
		})
	}
}

func TestNormalizeCatalogSourcesDedupesAndSorts(t *testing.T) {
	root := t.TempDir()
	manifestPath := writeWorkspaceFixture(t, filepath.Join(root, "devarch.workspace.yaml"), `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: catalog-order
catalog:
  sources:
    - ./zeta
    - ./alpha
    - alpha/../alpha
resources:
  api:
    template: node-api
`)

	ws, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("Load(%s) returned error: %v", manifestPath, err)
	}

	if got, want := ws.Catalog.Sources, []string{"alpha", "zeta"}; !slices.Equal(got, want) {
		t.Fatalf("Catalog.Sources = %v, want %v", got, want)
	}
	if got, want := ws.ResolvedCatalogSources(), []string{filepath.Join(root, "alpha"), filepath.Join(root, "zeta")}; !slices.Equal(got, want) {
		t.Fatalf("ResolvedCatalogSources() = %v, want %v", got, want)
	}
}

func writeWorkspaceFixture(t *testing.T, path, content string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%s): %v", path, err)
	}
	return filepath.Clean(path)
}
