package appsvc

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	podmanadapter "github.com/prospect-ogujiuba/devarch/internal/runtime/podman"
)

func TestPodmanSmoke(t *testing.T) {
	if os.Getenv("DEVARCH_INTEGRATION") != "podman" {
		t.Skip("set DEVARCH_INTEGRATION=podman and ensure podman can pull/run redis:7-alpine")
	}

	probeNetwork := "devarch-podman-smoke-probe"
	_ = exec.Command("podman", "network", "rm", probeNetwork).Run()
	if output, err := exec.Command("podman", "network", "create", probeNetwork).CombinedOutput(); err != nil {
		t.Skipf("podman network create unavailable; host must support rootless podman networks: %v: %s", err, output)
	}
	if output, err := exec.Command("podman", "network", "inspect", probeNetwork).CombinedOutput(); err != nil {
		_ = exec.Command("podman", "network", "rm", probeNetwork).Run()
		t.Skipf("podman network inspect unavailable after create; host must support rootless podman networks: %v: %s", err, output)
	}
	_ = exec.Command("podman", "network", "rm", probeNetwork).Run()

	root := repoRoot(t)
	workspaceRoot := t.TempDir()
	fixture := filepath.Join(root, "examples", "v2", "workspaces", "podman-smoke", "devarch.yaml")
	data, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatalf("read smoke fixture: %v", err)
	}
	data = []byte(strings.ReplaceAll(string(data), "../../../../catalog/builtin", filepath.Join(root, "catalog", "builtin")))
	manifestPath := filepath.Join(workspaceRoot, "devarch.workspace.yaml")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatalf("write temp smoke manifest: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	adapter := podmanadapter.New(nil)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		_ = adapter.RemoveResource(cleanupCtx, runtimepkg.ResourceRef{Workspace: "podman-smoke", Key: "redis", RuntimeName: "devarch-podman-smoke-redis"})
		_ = adapter.RemoveNetwork(cleanupCtx, &runtimepkg.DesiredNetwork{Name: "devarch-podman-smoke-net"})
	})

	service := newTestService(t, Config{
		WorkspaceRoots: []string{workspaceRoot},
		CatalogRoots:   exampleCatalogRoots(t),
		Adapters:       nil,
		LookPath:       func(file string) (string, error) { return "/usr/bin/" + file, nil },
	})

	plan, err := service.WorkspacePlan(ctx, "podman-smoke")
	if err != nil {
		t.Fatalf("WorkspacePlan returned error: %v", err)
	}
	if len(plan.Actions) == 0 {
		t.Fatal("WorkspacePlan returned no actions for empty Podman runtime")
	}
	t.Logf("podman smoke plan actions: %#v", plan.Actions)

	result, err := service.ApplyWorkspace(ctx, "podman-smoke")
	if err != nil {
		t.Fatalf("ApplyWorkspace returned error: %v", err)
	}
	if len(result.Operations) == 0 {
		t.Fatal("ApplyWorkspace returned no operations")
	}

	status, err := service.WorkspaceStatus(ctx, "podman-smoke")
	if err != nil {
		t.Fatalf("WorkspaceStatus returned error: %v", err)
	}
	if status.Snapshot == nil || status.Snapshot.Resource("redis") == nil {
		t.Fatalf("WorkspaceStatus missing redis snapshot: %#v", status.Snapshot)
	}
}
