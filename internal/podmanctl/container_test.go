package podmanctl

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

func TestBuildRunArgsFullSpecDeterministic(t *testing.T) {
	spec := ContainerSpec{
		Name:       "dev-web",
		Image:      "nginx:alpine",
		Command:    []string{"nginx", "-g", "daemon off;"},
		Entrypoint: []string{"/entrypoint.sh"},
		WorkingDir: "/app",
		Env: map[string]workspace.EnvValue{
			"ZED":   workspace.StringEnvValue("last"),
			"ALPHA": workspace.NumberEnvValue("1"),
		},
		Ports: []PortSpec{{Container: 80, Published: 8080, Protocol: "tcp"}, {Container: 443, Published: 8443, HostIP: "127.0.0.1"}},
		Volumes: []VolumeSpec{{Source: "/z", Target: "/z", ReadOnly: true}, {Source: "/a", Target: "/a"}},
		Labels: map[string]string{"z": "last", "a": "first"},
		Network: "dev-net",
		RestartPolicy: "unless-stopped",
		Health: &workspace.Health{Test: workspace.StringList{"curl", "-f", "http://localhost"}, Interval: "10s", Timeout: "2s", Retries: 3, StartPeriod: "5s"},
	}
	want := []string{"run", "--detach", "--replace", "--name", "dev-web", "--workdir", "/app", "--entrypoint", "/entrypoint.sh", "--env", "ALPHA=1", "--env", "ZED=last", "--publish", "127.0.0.1:8443:443/tcp", "--publish", "8080:80/tcp", "--volume", "/a:/a", "--volume", "/z:/z:ro", "--label", "a=first", "--label", "z=last", "--network", "dev-net", "--restart", "unless-stopped", "--health-cmd", "curl -f http://localhost", "--health-interval", "10s", "--health-timeout", "2s", "--health-retries", "3", "--health-start-period", "5s", "nginx:alpine", "nginx", "-g", "daemon off;"}
	if got := BuildRunArgs(spec); !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildRunArgs = %#v, want %#v", got, want)
	}
}

func TestBuildRunArgsMinimalSpec(t *testing.T) {
	want := []string{"run", "--detach", "--replace", "alpine"}
	if got := BuildRunArgs(ContainerSpec{Image: "alpine"}); !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildRunArgs = %#v, want %#v", got, want)
	}
}

func TestApplyContainerRunsBuiltArgs(t *testing.T) {
	runner := &fakeRunner{}
	err := ApplyContainer(context.Background(), runner, ContainerSpec{Name: "dev", Image: "alpine"})
	if err != nil {
		t.Fatalf("ApplyContainer returned error: %v", err)
	}
	want := []call{{command: "podman", args: []string{"run", "--detach", "--replace", "--name", "dev", "alpine"}}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestApplyContainerRequiresImage(t *testing.T) {
	if err := ApplyContainer(context.Background(), &fakeRunner{}, ContainerSpec{Name: "dev"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoveContainerMissingDoesNotFail(t *testing.T) {
	runner := &fakeRunner{outs: [][]byte{[]byte("container not found")}, errs: []error{errors.New("exit status 1")}}
	if err := RemoveContainer(context.Background(), runner, "dev"); err != nil {
		t.Fatalf("RemoveContainer returned error: %v", err)
	}
}

func TestRestartContainer(t *testing.T) {
	runner := &fakeRunner{}
	if err := RestartContainer(context.Background(), runner, "dev"); err != nil {
		t.Fatalf("RestartContainer returned error: %v", err)
	}
	want := []call{{command: "podman", args: []string{"restart", "dev"}}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestContainerErrorPropagation(t *testing.T) {
	runner := &fakeRunner{errs: []error{errors.New("daemon down")}}
	if err := RestartContainer(context.Background(), runner, "dev"); err == nil {
		t.Fatal("expected error")
	}
}
