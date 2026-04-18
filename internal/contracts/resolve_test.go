package contracts

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	resolvepkg "github.com/prospect-ogujiuba/devarch/internal/resolve"
	workspacepkg "github.com/prospect-ogujiuba/devarch/internal/workspace"
)

func TestResolveExplicitLinksAndPreservesSecretRefs(t *testing.T) {
	graph := loadExampleGraph(t, "shop-local")

	result := Resolve(graph)

	if got, want := len(result.Links), 3; got != want {
		t.Fatalf("len(result.Links) = %d, want %d", got, want)
	}

	postgresLink := findLink(t, result, "api", "postgres")
	if got, want := postgresLink.Source, "explicit"; got != want {
		t.Fatalf("postgres link source = %q, want %q", got, want)
	}
	if got, want := postgresLink.Provider, "postgres"; got != want {
		t.Fatalf("postgres link provider = %q, want %q", got, want)
	}
	if got, want := postgresLink.InjectedEnv["DB_HOST"].Text(), "postgres"; got != want {
		t.Fatalf("DB_HOST = %q, want %q", got, want)
	}
	if got, want := postgresLink.InjectedEnv["DB_PORT"].Text(), "5432"; got != want {
		t.Fatalf("DB_PORT = %q, want %q", got, want)
	}
	if got, want := postgresLink.InjectedEnv["DB_NAME"].Text(), "shop"; got != want {
		t.Fatalf("DB_NAME = %q, want %q", got, want)
	}
	if got, want := postgresLink.InjectedEnv["DB_USER"].Text(), "shop"; got != want {
		t.Fatalf("DB_USER = %q, want %q", got, want)
	}
	if got, ok := postgresLink.InjectedEnv["DB_PASSWORD"].SecretRef(); !ok || got != "postgres_password" {
		t.Fatalf("DB_PASSWORD secretRef = (%q, %v), want (%q, true)", got, ok, "postgres_password")
	}
	if _, ok := postgresLink.InjectedEnv["DATABASE_URL"]; ok {
		t.Fatal("DATABASE_URL should be omitted when interpolation would flatten a secretRef")
	}

	webLink := findLink(t, result, "web", "http")
	if got, want := webLink.Provider, "api"; got != want {
		t.Fatalf("web http provider = %q, want %q", got, want)
	}
	if got, want := webLink.InjectedEnv["API_URL"].Text(), "http://api:3000"; got != want {
		t.Fatalf("API_URL = %q, want %q", got, want)
	}
	if got, want := webLink.InjectedEnv["PORT"].Text(), "3000"; got != want {
		t.Fatalf("PORT = %q, want %q", got, want)
	}

	diagnostic := findDiagnostic(t, result, "api", "postgres", "secret-flatten")
	if got, want := diagnostic.Provider, "postgres"; got != want {
		t.Fatalf("secret-flatten provider = %q, want %q", got, want)
	}
	if got, want := diagnostic.EnvKey, "DATABASE_URL"; got != want {
		t.Fatalf("secret-flatten envKey = %q, want %q", got, want)
	}
}

func TestResolveAutoLinksSingleProvider(t *testing.T) {
	manifestPath := writeContractsWorkspaceFixture(t, filepath.Join(t.TempDir(), "devarch.workspace.yaml"), `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: auto-link
catalog:
  sources:
    - `+filepath.ToSlash(filepath.Join(repoRoot(t), "catalog", "builtin"))+`
resources:
  postgres:
    template: postgres
    exports:
      - postgres
  redis:
    template: redis
    exports:
      - redis
  api:
    template: node-api
    imports:
      - contract: postgres
        from: postgres
      - contract: redis
        from: redis
    exports:
      - http
  web:
    template: vite-web
`)

	graph := loadGraphFromManifest(t, manifestPath)
	result := Resolve(graph)

	link := findLink(t, result, "web", "http")
	if got, want := link.Source, "auto"; got != want {
		t.Fatalf("auto link source = %q, want %q", got, want)
	}
	if got, want := link.Provider, "api"; got != want {
		t.Fatalf("auto link provider = %q, want %q", got, want)
	}
	if got, want := link.InjectedEnv["API_URL"].Text(), "http://api:3000"; got != want {
		t.Fatalf("auto link API_URL = %q, want %q", got, want)
	}
}

func TestResolveAmbiguousImportProducesSortedDiagnostic(t *testing.T) {
	manifestPath := writeContractsWorkspaceFixture(t, filepath.Join(t.TempDir(), "devarch.workspace.yaml"), `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: ambiguous-http
catalog:
  sources:
    - `+filepath.ToSlash(filepath.Join(repoRoot(t), "catalog", "builtin"))+`
resources:
  postgres:
    template: postgres
    exports:
      - postgres
  redis:
    template: redis
    exports:
      - redis
  api:
    template: node-api
    imports:
      - contract: postgres
        from: postgres
      - contract: redis
        from: redis
    exports:
      - http
  web:
    template: vite-web
    imports:
      - contract: http
        from: api
  proxy:
    template: nginx
`)

	graph := loadGraphFromManifest(t, manifestPath)
	result := Resolve(graph)

	diagnostic := findDiagnostic(t, result, "proxy", "http", "ambiguous-import")
	if got, want := diagnostic.Providers, []string{"api", "web"}; !slices.Equal(got, want) {
		t.Fatalf("ambiguous providers = %v, want %v", got, want)
	}
	if link := maybeFindLink(result, "proxy", "http"); link != nil {
		t.Fatalf("expected no proxy/http link when ambiguous, got %#v", link)
	}
}

func TestResolveUnresolvedImportProducesDiagnostic(t *testing.T) {
	manifestPath := writeContractsWorkspaceFixture(t, filepath.Join(t.TempDir(), "devarch.workspace.yaml"), `apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: unresolved-http
catalog:
  sources:
    - `+filepath.ToSlash(filepath.Join(repoRoot(t), "catalog", "builtin"))+`
resources:
  proxy:
    template: nginx
`)

	graph := loadGraphFromManifest(t, manifestPath)
	result := Resolve(graph)

	diagnostic := findDiagnostic(t, result, "proxy", "http", "unresolved-import")
	if got, want := diagnostic.Message, `no enabled providers export contract "http"`; got != want {
		t.Fatalf("unresolved message = %q, want %q", got, want)
	}
	if got, want := len(result.Links), 0; got != want {
		t.Fatalf("len(result.Links) = %d, want %d", got, want)
	}
}

func loadExampleGraph(t *testing.T, name string) *resolvepkg.Graph {
	t.Helper()
	manifestPath := filepath.Join(repoRoot(t), "examples", "v2", "workspaces", name, "devarch.workspace.yaml")
	return loadGraphFromManifest(t, manifestPath)
}

func loadGraphFromManifest(t *testing.T, manifestPath string) *resolvepkg.Graph {
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
	return graph
}

func findLink(t *testing.T, result *Result, consumer, contract string) Link {
	t.Helper()

	link := maybeFindLink(result, consumer, contract)
	if link == nil {
		t.Fatalf("expected link for consumer=%q contract=%q", consumer, contract)
	}
	return *link
}

func maybeFindLink(result *Result, consumer, contract string) *Link {
	for i := range result.Links {
		if result.Links[i].Consumer == consumer && result.Links[i].Contract == contract {
			return &result.Links[i]
		}
	}
	return nil
}

func findDiagnostic(t *testing.T, result *Result, consumer, contract, code string) Diagnostic {
	t.Helper()

	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Consumer == consumer && diagnostic.Contract == contract && diagnostic.Code == code {
			return diagnostic
		}
	}
	t.Fatalf("expected diagnostic consumer=%q contract=%q code=%q, got %#v", consumer, contract, code, result.Diagnostics)
	return Diagnostic{}
}

func writeContractsWorkspaceFixture(t *testing.T, path, content string) string {
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
