package runtime_test

import (
	"path/filepath"
	stdruntime "runtime"
	"testing"
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	contractspkg "github.com/prospect-ogujiuba/devarch/internal/contracts"
	resolvepkg "github.com/prospect-ogujiuba/devarch/internal/resolve"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	workspacepkg "github.com/prospect-ogujiuba/devarch/internal/workspace"
)

func TestBuildDesiredWorkspaceSeparatesLogicalHostFromRuntimeNames(t *testing.T) {
	desired := loadDesiredWorkspace(t, "shop-local")
	if desired.Blocked() {
		t.Fatalf("desired workspace unexpectedly blocked: %#v", desired.BlockingDiagnostics())
	}
	if got, want := desired.Network.Name, "devarch-shop-local-net"; got != want {
		t.Fatalf("desired.Network.Name = %q, want %q", got, want)
	}

	api := desired.Resource("api")
	if api == nil {
		t.Fatal("expected api resource")
	}
	if got, want := api.LogicalHost, "api"; got != want {
		t.Fatalf("api.LogicalHost = %q, want %q", got, want)
	}
	if got, want := api.RuntimeName, "devarch-shop-local-api"; got != want {
		t.Fatalf("api.RuntimeName = %q, want %q", got, want)
	}
	if got, want := api.Spec.Ports[0].Container, 3000; got != want {
		t.Fatalf("api.Spec.Ports[0].Container = %d, want %d", got, want)
	}
	if got, want := api.Spec.Ports[0].Published, 8200; got != want {
		t.Fatalf("api.Spec.Ports[0].Published = %d, want %d", got, want)
	}
	if got, want := api.InjectedEnv["DB_HOST"].Text(), "postgres"; got != want {
		t.Fatalf("api.InjectedEnv[DB_HOST] = %q, want %q", got, want)
	}
	if got, want := api.Spec.Labels[runtimepkg.LabelHostAlias], "api"; got != want {
		t.Fatalf("api host label = %q, want %q", got, want)
	}
}

func TestBuildDesiredWorkspaceRejectsRawComposeCompatResources(t *testing.T) {
	desired := loadDesiredWorkspace(t, "compat-local")
	if !desired.Blocked() {
		t.Fatal("expected compat-local desired workspace to be blocked")
	}
	if got, want := len(desired.Resources), 2; got != want {
		t.Fatalf("len(desired.Resources) = %d, want %d", got, want)
	}
	for _, resource := range desired.Resources {
		if resource == nil {
			continue
		}
		if !resource.Blocked() {
			t.Fatalf("resource %q expected blocking diagnostic", resource.Key)
		}
		if got, want := resource.Diagnostics[0].Code, "unsupported-source-type"; got != want {
			t.Fatalf("resource %q diagnostic code = %q, want %q", resource.Key, got, want)
		}
	}
}

func TestNormalizeInspectSnapshotPreservesPublishedPortsAndHealth(t *testing.T) {
	desired := loadDesiredWorkspace(t, "shop-local")
	snapshot, err := runtimepkg.NormalizeInspectSnapshot(runtimepkg.ProviderDocker, desired, []byte(`[
  {
    "Id": "container-api",
    "Name": "/devarch-shop-local-api",
    "RestartCount": 2,
    "Config": {
      "Image": "node:22-alpine",
      "Env": ["NODE_ENV=development", "PORT=3000"],
      "Cmd": ["sh", "-c", "npm install && npm run dev"],
      "WorkingDir": "/workspace",
      "Labels": {
        "devarch.managed-by": "devarch-v2",
        "devarch.workspace": "shop-local",
        "devarch.resource": "api",
        "devarch.host": "api",
        "devarch.network": "devarch-shop-local-net"
      },
      "Healthcheck": {
        "Test": ["CMD-SHELL", "wget --spider -q http://localhost:3000/ || exit 1"],
        "Interval": 30000000000,
        "Timeout": 10000000000,
        "StartPeriod": 40000000000,
        "Retries": 3
      }
    },
    "State": {
      "Status": "running",
      "Running": true,
      "ExitCode": 0,
      "StartedAt": "2026-04-17T12:00:00Z",
      "FinishedAt": "0001-01-01T00:00:00Z",
      "Health": {"Status": "healthy"}
    },
    "NetworkSettings": {
      "Ports": {
        "3000/tcp": [{"HostIp": "127.0.0.1", "HostPort": "8200"}],
        "8080/tcp": null
      },
      "Networks": {
        "devarch-shop-local-net": {"Aliases": ["api"]}
      }
    },
    "Mounts": [
      {"Type": "volume", "Source": "shop-node-modules", "Destination": "/workspace/node_modules", "RW": true}
    ]
  }
]`), []byte(`[
  {
    "Name": "devarch-shop-local-net",
    "Id": "network-123",
    "Driver": "bridge",
    "Labels": {
      "devarch.managed-by": "devarch-v2",
      "devarch.workspace": "shop-local"
    }
  }
]`))
	if err != nil {
		t.Fatalf("NormalizeInspectSnapshot returned error: %v", err)
	}
	if snapshot.Workspace.Network == nil {
		t.Fatal("expected workspace network snapshot")
	}
	if got, want := snapshot.Workspace.Network.Name, "devarch-shop-local-net"; got != want {
		t.Fatalf("snapshot.Workspace.Network.Name = %q, want %q", got, want)
	}
	api := snapshot.Resource("api")
	if api == nil {
		t.Fatal("expected api snapshot resource")
	}
	if got, want := api.RuntimeName, "devarch-shop-local-api"; got != want {
		t.Fatalf("api.RuntimeName = %q, want %q", got, want)
	}
	if got, want := api.State.Health, "healthy"; got != want {
		t.Fatalf("api.State.Health = %q, want %q", got, want)
	}
	if got, want := api.Spec.Ports[0].Container, 3000; got != want {
		t.Fatalf("api.Spec.Ports[0].Container = %d, want %d", got, want)
	}
	if got, want := api.Spec.Ports[0].Published, 8200; got != want {
		t.Fatalf("api.Spec.Ports[0].Published = %d, want %d", got, want)
	}
	if got, want := api.Spec.Ports[1].Container, 8080; got != want {
		t.Fatalf("api.Spec.Ports[1].Container = %d, want %d", got, want)
	}
	if api.Spec.Ports[1].Published != 0 {
		t.Fatalf("api.Spec.Ports[1].Published = %d, want 0", api.Spec.Ports[1].Published)
	}
	if got, want := api.Spec.Health.Interval, "30s"; got != want {
		t.Fatalf("api.Spec.Health.Interval = %q, want %q", got, want)
	}
	if got, want := api.State.StartedAt.UTC().Format(time.RFC3339), "2026-04-17T12:00:00Z"; got != want {
		t.Fatalf("api.State.StartedAt = %q, want %q", got, want)
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

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := stdruntime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
