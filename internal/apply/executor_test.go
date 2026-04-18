package apply_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/apply"
	"github.com/prospect-ogujiuba/devarch/internal/events"
	planpkg "github.com/prospect-ogujiuba/devarch/internal/plan"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

func TestExecutorShopLocalGolden(t *testing.T) {
	desired := loadDesiredWorkspace(t, "shop-local")
	diff, err := planpkg.Diff(desired, &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: desired.Name, Provider: desired.Provider}})
	if err != nil {
		t.Fatalf("plan.Diff returned error: %v", err)
	}
	payload, err := apply.Render(desired)
	if err != nil {
		t.Fatalf("apply.Render returned error: %v", err)
	}

	adapter := &mockAdapter{snapshot: &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: desired.Name, Provider: desired.Provider, Network: &runtimepkg.SnapshotNetwork{Name: desired.Network.Name}}, Resources: []*runtimepkg.SnapshotResource{
		{Key: "api", RuntimeName: "devarch-shop-local-api", State: runtimepkg.ResourceState{Running: true, Status: "running"}},
		{Key: "postgres", RuntimeName: "devarch-shop-local-postgres", State: runtimepkg.ResourceState{Running: true, Status: "running"}},
		{Key: "redis", RuntimeName: "devarch-shop-local-redis", State: runtimepkg.ResourceState{Running: true, Status: "running"}},
		{Key: "web", RuntimeName: "devarch-shop-local-web", State: runtimepkg.ResourceState{Running: true, Status: "running"}},
	}}}
	bus := events.NewBus()
	bus.SetNow(func() time.Time { return time.Date(2026, 4, 17, 14, 0, 0, 0, time.UTC) })
	executor := &apply.Executor{Adapter: adapter, Publisher: bus, Now: func() time.Time { return time.Date(2026, 4, 17, 14, 0, 0, 0, time.UTC) }}
	result, err := executor.Execute(context.Background(), diff, payload)
	if err != nil {
		t.Fatalf("Executor.Execute returned error: %v", err)
	}

	if got, want := adapter.calls, []string{
		"ensure-network:devarch-shop-local-net",
		"apply-resource:api",
		"apply-resource:postgres",
		"apply-resource:redis",
		"apply-resource:web",
		"inspect-workspace:shop-local",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("adapter calls = %v, want %v", got, want)
	}

	actual := marshalJSON(t, result)
	goldenPath := filepath.Join(repoRoot(t), "testdata", "goldens", "phase3", "shop-local.apply.golden.json")
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

func TestExecutorCompatLocalBlocksBeforeSideEffects(t *testing.T) {
	desired := loadDesiredWorkspace(t, "compat-local")
	diff, err := planpkg.Diff(desired, &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: desired.Name, Provider: desired.Provider}})
	if err != nil {
		t.Fatalf("plan.Diff returned error: %v", err)
	}
	payload, err := apply.Render(desired)
	if err != nil {
		t.Fatalf("apply.Render returned error: %v", err)
	}
	adapter := &mockAdapter{}
	executor := &apply.Executor{Adapter: adapter, Now: func() time.Time { return time.Date(2026, 4, 17, 14, 30, 0, 0, time.UTC) }}
	_, err = executor.Execute(context.Background(), diff, payload)
	if !errors.Is(err, apply.ErrBlocked) {
		t.Fatalf("Executor.Execute error = %v, want ErrBlocked", err)
	}
	if len(adapter.calls) != 0 {
		t.Fatalf("expected no adapter calls, got %v", adapter.calls)
	}
}

type mockAdapter struct {
	calls    []string
	snapshot *runtimepkg.Snapshot
}

func (m *mockAdapter) Provider() string { return runtimepkg.ProviderDocker }

func (m *mockAdapter) Capabilities() runtimepkg.AdapterCapabilities {
	return runtimepkg.AdapterCapabilities{Inspect: true, Apply: true, Network: true}
}

func (m *mockAdapter) InspectWorkspace(context.Context, *runtimepkg.DesiredWorkspace) (*runtimepkg.Snapshot, error) {
	m.calls = append(m.calls, "inspect-workspace:shop-local")
	return m.snapshot, nil
}

func (m *mockAdapter) EnsureNetwork(_ context.Context, network *runtimepkg.DesiredNetwork) error {
	m.calls = append(m.calls, "ensure-network:"+network.Name)
	return nil
}

func (m *mockAdapter) RemoveNetwork(context.Context, *runtimepkg.DesiredNetwork) error {
	m.calls = append(m.calls, "remove-network")
	return nil
}

func (m *mockAdapter) ApplyResource(_ context.Context, request runtimepkg.ApplyResourceRequest) error {
	m.calls = append(m.calls, "apply-resource:"+request.Resource.Key)
	return nil
}

func (m *mockAdapter) RemoveResource(_ context.Context, resource runtimepkg.ResourceRef) error {
	m.calls = append(m.calls, "remove-resource:"+resource.Key)
	return nil
}

func (m *mockAdapter) RestartResource(_ context.Context, resource runtimepkg.ResourceRef) error {
	m.calls = append(m.calls, "restart-resource:"+resource.Key)
	return nil
}

func (m *mockAdapter) StreamLogs(context.Context, runtimepkg.ResourceRef, runtimepkg.LogsRequest, runtimepkg.LogsConsumer) error {
	return nil
}

func (m *mockAdapter) Exec(context.Context, runtimepkg.ResourceRef, runtimepkg.ExecRequest) (*runtimepkg.ExecResult, error) {
	return &runtimepkg.ExecResult{ExitCode: 0}, nil
}
