package plan_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	stdruntime "runtime"
	"testing"

	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	contractspkg "github.com/prospect-ogujiuba/devarch/internal/contracts"
	planpkg "github.com/prospect-ogujiuba/devarch/internal/plan"
	resolvepkg "github.com/prospect-ogujiuba/devarch/internal/resolve"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	workspacepkg "github.com/prospect-ogujiuba/devarch/internal/workspace"
)

func TestPhase3PlanGoldens(t *testing.T) {
	tests := []struct {
		name       string
		workspace  string
		goldenPath string
		snapshot   func(*runtimepkg.DesiredWorkspace) *runtimepkg.Snapshot
	}{
		{
			name:       "shop-local-empty",
			workspace:  "shop-local",
			goldenPath: filepath.Join(repoRoot(t), "testdata", "goldens", "phase3", "shop-local.plan.golden.json"),
			snapshot: func(desired *runtimepkg.DesiredWorkspace) *runtimepkg.Snapshot {
				return &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: desired.Name, Provider: desired.Provider}}
			},
		},
		{
			name:       "laravel-local-empty",
			workspace:  "laravel-local",
			goldenPath: filepath.Join(repoRoot(t), "testdata", "goldens", "phase3", "laravel-local.plan.golden.json"),
			snapshot: func(desired *runtimepkg.DesiredWorkspace) *runtimepkg.Snapshot {
				return &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: desired.Name, Provider: desired.Provider}}
			},
		},
		{
			name:       "compat-local-empty",
			workspace:  "compat-local",
			goldenPath: filepath.Join(repoRoot(t), "testdata", "goldens", "phase3", "compat-local.plan.golden.json"),
			snapshot: func(desired *runtimepkg.DesiredWorkspace) *runtimepkg.Snapshot {
				return &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: desired.Name, Provider: desired.Provider}}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			desired := loadDesiredWorkspace(t, tt.workspace)
			result, err := planpkg.Diff(desired, tt.snapshot(desired))
			if err != nil {
				t.Fatalf("plan.Diff returned error: %v", err)
			}
			actual := marshalJSON(t, result)
			if updateGoldens() {
				if err := os.MkdirAll(filepath.Dir(tt.goldenPath), 0o755); err != nil {
					t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(tt.goldenPath), err)
				}
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

func TestDiffComputesModifyRestartRemoveAndNoopReasons(t *testing.T) {
	desired := &runtimepkg.DesiredWorkspace{
		Name:     "shop-local",
		Provider: runtimepkg.ProviderDocker,
		Resources: []*runtimepkg.DesiredResource{
			{
				Key:         "api",
				Enabled:     true,
				RuntimeName: "devarch-shop-local-api",
				Spec: runtimepkg.ResourceSpec{
					Image: "node:22-alpine",
					Env: map[string]workspacepkg.EnvValue{
						"PORT": workspacepkg.StringEnvValue("3000"),
					},
				},
			},
		},
	}
	result, err := planpkg.Diff(desired, &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: desired.Name}, Resources: []*runtimepkg.SnapshotResource{
		{
			Key:         "api",
			RuntimeName: "devarch-shop-local-api",
			State:       runtimepkg.ResourceState{Running: false, Status: "exited"},
			Spec:        runtimepkg.ResourceSpec{Image: "node:22-alpine", Env: map[string]workspacepkg.EnvValue{"PORT": workspacepkg.StringEnvValue("3000")}},
		},
		{
			Key:         "redis",
			RuntimeName: "devarch-shop-local-redis",
			State:       runtimepkg.ResourceState{Running: true, Status: "running"},
			Spec:        runtimepkg.ResourceSpec{Image: "redis:7-alpine"},
		},
	}})
	if err != nil {
		t.Fatalf("plan.Diff returned error: %v", err)
	}
	if got, want := result.Actions[0].Kind, planpkg.ActionRestart; got != want {
		t.Fatalf("api action kind = %q, want %q", got, want)
	}
	if got, want := result.Actions[1].Kind, planpkg.ActionRemove; got != want {
		t.Fatalf("redis action kind = %q, want %q", got, want)
	}

	modifiedSnapshot := &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: desired.Name}, Resources: []*runtimepkg.SnapshotResource{{
		Key:         "api",
		RuntimeName: "devarch-shop-local-api",
		State:       runtimepkg.ResourceState{Running: true, Status: "running"},
		Spec: runtimepkg.ResourceSpec{Image: "node:22-alpine", Env: map[string]workspacepkg.EnvValue{
			"PORT": workspacepkg.StringEnvValue("3100"),
		}},
	}}}
	result, err = planpkg.Diff(desired, modifiedSnapshot)
	if err != nil {
		t.Fatalf("plan.Diff returned error: %v", err)
	}
	if got, want := result.Actions[0].Kind, planpkg.ActionModify; got != want {
		t.Fatalf("modify action kind = %q, want %q", got, want)
	}
	if got, want := result.Actions[0].Reasons, []string{"environment changed"}; !bytes.Equal(marshalJSON(t, got), marshalJSON(t, want)) {
		t.Fatalf("modify reasons = %v, want %v", got, want)
	}

	matchingSnapshot := &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: desired.Name}, Resources: []*runtimepkg.SnapshotResource{{
		Key:         "api",
		RuntimeName: "devarch-shop-local-api",
		State:       runtimepkg.ResourceState{Running: true, Status: "running", Health: "healthy"},
		Spec: runtimepkg.ResourceSpec{Image: "node:22-alpine", Env: map[string]workspacepkg.EnvValue{
			"PORT": workspacepkg.StringEnvValue("3000"),
		}},
	}}}
	result, err = planpkg.Diff(desired, matchingSnapshot)
	if err != nil {
		t.Fatalf("plan.Diff returned error: %v", err)
	}
	if got, want := result.Actions[0].Kind, planpkg.ActionNoop; got != want {
		t.Fatalf("noop action kind = %q, want %q", got, want)
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
