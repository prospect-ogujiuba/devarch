package importv1

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	contractspkg "github.com/prospect-ogujiuba/devarch/internal/contracts"
	"github.com/prospect-ogujiuba/devarch/internal/spec"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
	"github.com/prospect-ogujiuba/devarch/internal/resolve"
	"gopkg.in/yaml.v3"
)

func TestImportLibraryFixtureEmitsSchemaValidArtifactsAndDiagnostics(t *testing.T) {
	root := filepath.Join(repoRoot(t), "examples", "v1", "library")
	result, err := ImportLibrary(root)
	if err != nil {
		t.Fatalf("ImportLibrary returned error: %v", err)
	}
	if got, want := result.Mode, ModeV1Library; got != want {
		t.Fatalf("Mode = %q, want %q", got, want)
	}
	if got, want := result.Status, StatusPartial; got != want {
		t.Fatalf("Status = %q, want %q", got, want)
	}
	if got, want := result.Summary.Total, 2; got != want {
		t.Fatalf("Summary.Total = %d, want %d", got, want)
	}

	postgres := findArtifact(t, result, ArtifactKindTemplate, "postgres")
	if postgres.Document == "" {
		t.Fatal("postgres template document is empty")
	}
	if err := spec.ValidateTemplateBytes([]byte(postgres.Document)); err != nil {
		t.Fatalf("spec.ValidateTemplateBytes(postgres) returned error: %v", err)
	}
	var postgresDoc map[string]any
	if err := yaml.Unmarshal([]byte(postgres.Document), &postgresDoc); err != nil {
		t.Fatalf("yaml.Unmarshal(postgres) returned error: %v", err)
	}
	assertNestedEqual(t, postgresDoc, "metadata.name", "postgres")
	assertNestedEqual(t, postgresDoc, "spec.runtime.image", "postgres:16")
	assertNestedEqual(t, postgresDoc, "spec.ports.0.container", 5432)

	php := findArtifact(t, result, ArtifactKindTemplate, "php")
	if php.Status != StatusPartial {
		t.Fatalf("php.Status = %q, want %q", php.Status, StatusPartial)
	}
	if err := spec.ValidateTemplateBytes([]byte(php.Document)); err != nil {
		t.Fatalf("spec.ValidateTemplateBytes(php) returned error: %v", err)
	}
	var phpDoc map[string]any
	if err := yaml.Unmarshal([]byte(php.Document), &phpDoc); err != nil {
		t.Fatalf("yaml.Unmarshal(php) returned error: %v", err)
	}
	assertNestedEqual(t, phpDoc, "spec.runtime.build.context", "./config")
	assertNestedEqual(t, phpDoc, "spec.runtime.workingDir", "/var/www/html")
	assertNestedEqual(t, phpDoc, "spec.volumes.0.source", "php_vendor")
	compat := nestedMap(t, phpDoc, "spec", "develop", "importv1")
	configFiles := nestedMapValue(t, compat, "configFiles")
	configPHP := nestedMapValue(t, configFiles, "config/php.ini")
	if got, want := configPHP["fileMode"], any("0644"); got != want {
		t.Fatalf("config/php.ini fileMode = %#v, want %#v", got, want)
	}
	labels := nestedMapValue(t, compat, "labels")
	if got, want := labels["com.example.role"], any("backend"); got != want {
		t.Fatalf("compat label = %#v, want %#v", got, want)
	}
	volumes := nestedSlice(t, compat, "volumes")
	if got, want := nestedMapFromSlice(t, volumes, 0)["source"], any("../../../apps"); got != want {
		t.Fatalf("compat volume source = %#v, want %#v", got, want)
	}
	assertHasDiagnostic(t, php.Diagnostics, "template-compat-fields-preserved")
	assertHasDiagnostic(t, php.Diagnostics, "template-contracts-unavailable")
}

func TestImportStackFixtureResolvesAgainstBuiltinCatalog(t *testing.T) {
	path := filepath.Join(repoRoot(t), "examples", "v1", "stacks", "shop-export.yaml")
	result, err := ImportStack(path)
	if err != nil {
		t.Fatalf("ImportStack returned error: %v", err)
	}
	if got, want := result.Status, StatusPartial; got != want {
		t.Fatalf("Status = %q, want %q", got, want)
	}
	artifact := findArtifact(t, result, ArtifactKindWorkspace, "shop-import")
	if artifact.Document == "" {
		t.Fatal("workspace artifact document is empty")
	}
	if err := spec.ValidateWorkspaceBytes([]byte(artifact.Document)); err != nil {
		t.Fatalf("spec.ValidateWorkspaceBytes returned error: %v", err)
	}

	workspaceDir := filepath.Join(t.TempDir(), "workspaces", "shop-import")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s): %v", workspaceDir, err)
	}
	manifestPath := filepath.Join(workspaceDir, spec.ManifestFilename)
	if err := os.WriteFile(manifestPath, []byte(artifact.Document), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%s): %v", manifestPath, err)
	}

	ws, err := workspace.Load(manifestPath)
	if err != nil {
		t.Fatalf("workspace.Load returned error: %v", err)
	}
	if got, want := ws.Policies.AutoWire, false; got != want {
		t.Fatalf("workspace.Policies.AutoWire = %t, want %t", got, want)
	}
	secretRef, ok := ws.Resources["postgres"].Env["POSTGRES_PASSWORD"].SecretRef()
	if !ok || secretRef != "POSTGRES_PASSWORD" {
		t.Fatalf("postgres secretRef = (%q, %t), want (POSTGRES_PASSWORD, true)", secretRef, ok)
	}
	if ws.Resources["api"].Overrides == nil {
		t.Fatal("expected api overrides.importv1 compatibility bucket")
	}

	catalogPaths, err := catalog.DiscoverTemplateFiles([]string{filepath.Join(repoRoot(t), "catalog", "builtin")})
	if err != nil {
		t.Fatalf("catalog.DiscoverTemplateFiles returned error: %v", err)
	}
	index, err := catalog.LoadIndex(catalogPaths)
	if err != nil {
		t.Fatalf("catalog.LoadIndex returned error: %v", err)
	}
	graph, err := resolve.Resolve(ws, index)
	if err != nil {
		t.Fatalf("resolve.Resolve returned error: %v", err)
	}
	contracts := contractspkg.Resolve(graph)
	if got, want := len(contracts.Links), 2; got != want {
		t.Fatalf("len(contracts.Links) = %d, want %d", got, want)
	}
	assertHasDiagnostic(t, artifact.Diagnostics, "stack-network-name-lossy")
	assertHasDiagnostic(t, artifact.Diagnostics, "instance-compat-fields-preserved")
}

func TestImportStackRejectsMissingTemplate(t *testing.T) {
	path := filepath.Join(repoRoot(t), "examples", "v1", "stacks", "rejected-missing-template.yaml")
	result, err := ImportStack(path)
	if err != nil {
		t.Fatalf("ImportStack returned error: %v", err)
	}
	if got, want := result.Status, StatusRejected; got != want {
		t.Fatalf("Status = %q, want %q", got, want)
	}
	artifact := findArtifact(t, result, ArtifactKindWorkspace, "broken-import")
	if artifact.Status != StatusRejected {
		t.Fatalf("artifact.Status = %q, want %q", artifact.Status, StatusRejected)
	}
	assertHasDiagnostic(t, artifact.Diagnostics, "instance-template-missing")
}

func findArtifact(t *testing.T, result *Result, kind, name string) Artifact {
	t.Helper()
	for _, artifact := range result.Artifacts {
		if artifact.Kind == kind && artifact.Name == name {
			return artifact
		}
	}
	t.Fatalf("artifact %s/%s not found in %#v", kind, name, result.Artifacts)
	return Artifact{}
}

func assertHasDiagnostic(t *testing.T, diagnostics []Diagnostic, code string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return
		}
	}
	t.Fatalf("diagnostic %q not found in %#v", code, diagnostics)
}

func assertNestedEqual(t *testing.T, value map[string]any, path string, want any) {
	t.Helper()
	got, ok := lookupNested(value, path)
	if !ok {
		t.Fatalf("lookupNested(%s) failed in %#v", path, value)
	}
	if got != want {
		t.Fatalf("lookupNested(%s) = %#v, want %#v", path, got, want)
	}
}

func lookupNested(value any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	current := value
	for _, part := range parts {
		switch typed := current.(type) {
		case map[string]any:
			next, ok := typed[part]
			if !ok {
				return nil, false
			}
			current = next
		case []any:
			index := int(part[0] - '0')
			if index < 0 || index >= len(typed) {
				return nil, false
			}
			current = typed[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func nestedMap(t *testing.T, root map[string]any, keys ...string) map[string]any {
	t.Helper()
	current := root
	for _, key := range keys {
		next, ok := current[key].(map[string]any)
		if !ok {
			t.Fatalf("nested map %q missing in %#v", key, current)
		}
		current = next
	}
	return current
}

func nestedMapValue(t *testing.T, root map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := root[key].(map[string]any)
	if !ok {
		t.Fatalf("map value %q missing in %#v", key, root)
	}
	return value
}

func nestedSlice(t *testing.T, root map[string]any, key string) []any {
	t.Helper()
	value, ok := root[key].([]any)
	if !ok {
		t.Fatalf("slice value %q missing in %#v", key, root)
	}
	return value
}

func nestedMapFromSlice(t *testing.T, root []any, index int) map[string]any {
	t.Helper()
	if index < 0 || index >= len(root) {
		t.Fatalf("slice index %d out of range in %#v", index, root)
	}
	value, ok := root[index].(map[string]any)
	if !ok {
		t.Fatalf("slice value %d is not a map in %#v", index, root)
	}
	return value
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
