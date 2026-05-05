package podman

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

func TestPodmanAdapterContractInspectLogsAndExec(t *testing.T) {
	runner := &fakeRunner{responses: map[string]fakeResponse{
		"podman ps -aq --filter label=devarch.workspace=shop-local --filter label=devarch.managed-by=devarch-v2": {
			stdout: []byte("container-api\n"),
		},
		"podman inspect container-api": {
			stdout: []byte(`[
  {
    "Id": "container-api",
    "Name": "/devarch-shop-local-api",
    "Config": {
      "Image": "node:22-alpine",
      "Env": ["PORT=3000"],
      "Cmd": ["npm", "run", "dev"],
      "WorkingDir": "/workspace",
      "Labels": {
        "devarch.managed-by": "devarch-v2",
        "devarch.workspace": "shop-local",
        "devarch.resource": "api",
        "devarch.host": "api",
        "devarch.network": "devarch-shop-local-net"
      }
    },
    "State": {"Status": "running", "Running": true},
    "NetworkSettings": {"Ports": {"3000/tcp": [{"HostIp": "127.0.0.1", "HostPort": "8200"}]}}
  }
]`),
		},
		"podman network inspect devarch-shop-local-net": {
			stdout: []byte(`[{"Name":"devarch-shop-local-net","Id":"network-1","Driver":"bridge"}]`),
		},
		"podman logs --timestamps --tail 10 devarch-shop-local-api": {
			stdout: []byte("2026-04-17T12:00:00Z started\n2026-04-17T12:00:01Z listening\n"),
		},
		"podman exec devarch-shop-local-api php artisan about": {
			stdout: []byte("Application\n"),
		},
	}}
	adapter := New(runner)
	if got, want := adapter.Capabilities(), (runtimepkg.AdapterCapabilities{Inspect: true, Apply: true, Logs: true, Exec: true, Network: true}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Capabilities() = %#v, want %#v", got, want)
	}

	desired := &runtimepkg.DesiredWorkspace{
		Name:           "shop-local",
		NamingStrategy: runtimepkg.NamingStrategyWorkspaceResource,
		Network:        &runtimepkg.DesiredNetwork{Name: "devarch-shop-local-net"},
		Resources:      []*runtimepkg.DesiredResource{{Key: "api", RuntimeName: "devarch-shop-local-api"}},
	}
	snapshot, err := adapter.InspectWorkspace(context.Background(), desired)
	if err != nil {
		t.Fatalf("InspectWorkspace returned error: %v", err)
	}
	if got, want := snapshot.Resource("api").Spec.Ports[0].Published, 8200; got != want {
		t.Fatalf("published port = %d, want %d", got, want)
	}

	var lines []string
	err = adapter.StreamLogs(context.Background(), runtimepkg.ResourceRef{Workspace: "shop-local", Key: "api", RuntimeName: "devarch-shop-local-api"}, runtimepkg.LogsRequest{Tail: 10}, func(chunk runtimepkg.LogChunk) error {
		lines = append(lines, chunk.Line)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamLogs returned error: %v", err)
	}
	if got, want := lines, []string{"started", "listening"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("log lines = %v, want %v", got, want)
	}

	result, err := adapter.Exec(context.Background(), runtimepkg.ResourceRef{Workspace: "shop-local", Key: "api", RuntimeName: "devarch-shop-local-api"}, runtimepkg.ExecRequest{Command: []string{"php", "artisan", "about"}})
	if err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}
	if got, want := result.Stdout, "Application\n"; got != want {
		t.Fatalf("Exec stdout = %q, want %q", got, want)
	}
}

func TestPodmanAdapterMutationsCallPodmanctl(t *testing.T) {
	runner := &fakeRunner{responses: map[string]fakeResponse{
		"podman network exists devarch-shop-local-net": {},
		"podman network rm devarch-shop-local-net":     {},
		"podman run --detach --replace --name devarch-shop-local-api --workdir /app --env APP_ENV=local --publish 127.0.0.1:8080:80/tcp --volume ./app:/app:ro --label devarch.host=api --label devarch.managed-by=devarch-v2 --label devarch.network=devarch-shop-local-net --label devarch.resource=api --label devarch.workspace=shop-local --label tier=web --network devarch-shop-local-net --restart unless-stopped --health-cmd curl -f http://localhost/health --health-interval 10s nginx:alpine nginx -g daemon off;": {},
		"podman rm --force devarch-shop-local-api": {},
		"podman restart devarch-shop-local-api":    {},
	}}
	adapter := New(runner)

	if err := adapter.EnsureNetwork(context.Background(), &runtimepkg.DesiredNetwork{Name: "devarch-shop-local-net"}); err != nil {
		t.Fatalf("EnsureNetwork returned error: %v", err)
	}
	if err := adapter.RemoveNetwork(context.Background(), &runtimepkg.DesiredNetwork{Name: "devarch-shop-local-net"}); err != nil {
		t.Fatalf("RemoveNetwork returned error: %v", err)
	}
	request := runtimepkg.ApplyResourceRequest{
		Workspace:   "shop-local",
		NetworkName: "devarch-shop-local-net",
		Resource: runtimepkg.AppliedResource{
			Key:         "api",
			LogicalHost: "api",
			RuntimeName: "devarch-shop-local-api",
			Spec: runtimepkg.ResourceSpec{
				Image:      "nginx:alpine",
				Command:    []string{"nginx", "-g", "daemon off;"},
				WorkingDir: "/app",
				Env:        map[string]workspace.EnvValue{"APP_ENV": workspace.StringEnvValue("local")},
				Ports:      []runtimepkg.PortSpec{{HostIP: "127.0.0.1", Published: 8080, Container: 80, Protocol: "tcp"}},
				Volumes:    []runtimepkg.VolumeSpec{{Source: "./app", Target: "/app", ReadOnly: true}},
				Labels:     map[string]string{runtimepkg.LabelManagedBy: runtimepkg.ManagedByValue, "tier": "web"},
				Health:     &workspace.Health{Test: []string{"curl", "-f", "http://localhost/health"}, Interval: "10s"},
			},
		},
	}
	if err := adapter.ApplyResource(context.Background(), request); err != nil {
		t.Fatalf("ApplyResource returned error: %v", err)
	}
	ref := runtimepkg.ResourceRef{RuntimeName: "devarch-shop-local-api"}
	if err := adapter.RemoveResource(context.Background(), ref); err != nil {
		t.Fatalf("RemoveResource returned error: %v", err)
	}
	if err := adapter.RestartResource(context.Background(), ref); err != nil {
		t.Fatalf("RestartResource returned error: %v", err)
	}
}

func TestPodmanAdapterMutationValidation(t *testing.T) {
	adapter := New(&fakeRunner{responses: map[string]fakeResponse{}})
	if err := adapter.ApplyResource(context.Background(), runtimepkg.ApplyResourceRequest{}); err == nil || !strings.Contains(err.Error(), "runtime name is required") {
		t.Fatalf("ApplyResource missing runtime name error = %v", err)
	}
	request := runtimepkg.ApplyResourceRequest{Resource: runtimepkg.AppliedResource{RuntimeName: "devarch-shop-local-api", Spec: runtimepkg.ResourceSpec{Build: &runtimepkg.BuildSpec{Context: "."}}}}
	if err := adapter.ApplyResource(context.Background(), request); err == nil || !strings.Contains(err.Error(), "apply-resource-build") {
		t.Fatalf("ApplyResource build error = %v", err)
	}
	if err := adapter.RemoveResource(context.Background(), runtimepkg.ResourceRef{}); err == nil || !strings.Contains(err.Error(), "runtime name is required") {
		t.Fatalf("RemoveResource missing runtime name error = %v", err)
	}
}

type fakeRunner struct {
	responses map[string]fakeResponse
}

type fakeResponse struct {
	stdout []byte
	err    error
}

func (f *fakeRunner) Run(_ context.Context, command string, args ...string) ([]byte, error) {
	key := strings.TrimSpace(command + " " + strings.Join(args, " "))
	response, ok := f.responses[key]
	if !ok {
		return nil, fmt.Errorf("unexpected command %q", key)
	}
	return response.stdout, response.err
}
