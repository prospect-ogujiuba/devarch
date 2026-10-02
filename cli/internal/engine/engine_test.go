package engine_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
	"github.com/prospect-ogujiuba/devarch/cli/internal/state"
	"github.com/prospect-ogujiuba/devarch/cli/internal/testutil"
)

func setup(t *testing.T) (*engine.Engine, *runner.Fake, string) {
	t.Helper()
	lib := filepath.Join(t.TempDir(), "services-library")
	testutil.WriteFiles(t, lib, map[string]string{
		"backend/php/compose.yml": testutil.Compose("php", `x-devarch:
  version: {var: PHP_VERSION, default: "8.4", rebuild: true}
`),
		"proxy/npm/compose.yml": testutil.Compose("npm", "x-devarch:\n  requires: [php]\n  ready: {exec: [curl, -f, localhost]}\n"),
	})
	c, problems := catalog.Load([]catalog.Path{{Dir: lib}})
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	f := &runner.Fake{Responses: map[string]runner.Response{}, RunErrors: map[string][]error{}}
	e := &engine.Engine{Catalog: c, Runner: f, Versions: state.Versions{"PHP_VERSION": "8.3"}, Sleep: func(time.Duration) {}}
	return e, f, lib
}

func get(t *testing.T, e *engine.Engine, id string) catalog.Service {
	s, ok := e.Catalog.Get(id)
	if !ok {
		t.Fatalf("%s missing", id)
	}
	return s
}

func TestUpCreatesNetworkAndStartsRequiresFirst(t *testing.T) {
	e, f, lib := setup(t)
	f.Responses["podman network exists"] = runner.Response{Err: &runner.ExitError{Code: 1}}
	if err := e.Up(context.Background(), []catalog.Service{get(t, e, "proxy/npm")}, engine.UpOptions{}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"podman network exists microservices-net",
		"podman network create microservices-net",
		"(cd " + filepath.Join(lib, "backend/php") + " && PHP_VERSION=8.3 podman compose up -d)",
		"(cd " + filepath.Join(lib, "proxy/npm") + " && podman compose up -d)",
	}
	if got := f.Lines(); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("commands:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestUpRetriesOnce(t *testing.T) {
	e, f, _ := setup(t)
	f.RunErrors["podman compose up"] = []error{errors.New("flaky"), errors.New("still broken")}
	err := e.Up(context.Background(), []catalog.Service{get(t, e, "backend/php")}, engine.UpOptions{})
	if err == nil || !strings.Contains(err.Error(), "still broken") {
		t.Fatalf("want second failure surfaced, got %v", err)
	}
	ups := 0
	for _, l := range f.Lines() {
		if strings.HasSuffix(l, "compose up -d)") {
			ups++
		}
	}
	if ups != 2 {
		t.Fatalf("want exactly 2 attempts, got %d", ups)
	}
}

func TestUpFailsWhenPodmanMissing(t *testing.T) {
	e, f, _ := setup(t)
	f.Responses["podman network exists"] = runner.Response{Err: errors.New("executable file not found")}
	if err := e.Up(context.Background(), []catalog.Service{get(t, e, "backend/php")}, engine.UpOptions{}); err == nil || !strings.Contains(err.Error(), "podman is unavailable") {
		t.Fatalf("got %v", err)
	}
}

func TestDownReverseOrderWithVolumes(t *testing.T) {
	e, f, _ := setup(t)
	svcs := []catalog.Service{get(t, e, "backend/php"), get(t, e, "proxy/npm")}
	if err := e.Down(context.Background(), svcs, true); err != nil {
		t.Fatal(err)
	}
	lines := f.Lines()
	if len(lines) != 2 || !strings.Contains(lines[0], "proxy/npm") || !strings.HasSuffix(lines[1], "compose down --volumes)") {
		t.Fatalf("got %v", lines)
	}
}

const psJSON = `[
 {"Names":["php"],"Image":"localhost/php_php","State":"running","Status":"Up 2 hours (healthy)",
  "Labels":{"com.docker.compose.project":"php","com.docker.compose.service":"php","com.docker.compose.project.working_dir":"%LIB%/backend/php"},
  "Ports":[{"host_ip":"127.0.0.1","host_port":8100,"container_port":8000,"protocol":"tcp"}]},
 {"Names":["php-app"],"Image":"x","State":"exited","Status":"Exited (0)",
  "Labels":{"com.docker.compose.project":"devarch-node-app","com.docker.compose.project.working_dir":"%LIB%/backend/php"}},
 {"Names":["npm"],"Image":"y","State":"running","Status":"Up 1 minute (starting)",
  "Labels":{"com.docker.compose.project":"npm","com.docker.compose.service":"npm","com.docker.compose.project.working_dir":"%LIB%/proxy/npm"}},
 {"Names":["stray"],"Image":"z","State":"running","Status":"Up"}
]`

func TestContainersAndStates(t *testing.T) {
	e, f, lib := setup(t)
	f.Responses["podman ps"] = runner.Response{Out: []byte(strings.ReplaceAll(psJSON, "%LIB%", lib))}
	cs, err := e.Containers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	mapped := map[string]string{}
	for _, c := range cs {
		mapped[c.Name] = c.Service
	}
	if mapped["php"] != "backend/php" || mapped["php-app"] != "" || mapped["npm"] != "proxy/npm" || mapped["stray"] != "" {
		t.Fatalf("mapping = %v", mapped)
	}
	states := engine.States(e.Catalog.Services, cs)
	if s := states["backend/php"]; s.State != "running" || s.Health != "healthy" {
		t.Fatalf("php state = %+v", s)
	}
	if s := states["proxy/npm"]; s.Health != "starting" {
		t.Fatalf("npm state = %+v", s)
	}
}

func TestWaitRunsHealthcheckAndReadyProbe(t *testing.T) {
	e, f, lib := setup(t)
	f.Responses["podman ps"] = runner.Response{Out: []byte(strings.ReplaceAll(psJSON, "%LIB%", lib))}
	f.Responses["podman inspect"] = runner.Response{Out: []byte(`[{"Name":"npm","State":{"Status":"running"},"Config":{"Healthcheck":{"Test":["CMD","true"]}}}]`)}
	if err := e.Wait(context.Background(), get(t, e, "proxy/npm"), time.Second); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.Lines(), "\n")
	for _, want := range []string{"podman healthcheck run npm", "podman exec npm curl -f localhost"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}
}

func TestWaitTimesOut(t *testing.T) {
	e, f, lib := setup(t)
	f.Responses["podman ps"] = runner.Response{Out: []byte(strings.ReplaceAll(psJSON, "%LIB%", lib))}
	f.Responses["podman inspect"] = runner.Response{Out: []byte(`[{"Name":"php","State":{"Status":"exited"},"Config":{}}]`)}
	err := e.Wait(context.Background(), get(t, e, "backend/php"), time.Nanosecond)
	if err == nil || !strings.Contains(err.Error(), "php is exited") {
		t.Fatalf("got %v", err)
	}
}

func TestSelectedVersion(t *testing.T) {
	e, _, _ := setup(t)
	php := get(t, e, "backend/php")
	if v := e.SelectedVersion(php); v != "8.3" {
		t.Fatalf("selected = %s", v)
	}
	e.Versions = state.Versions{}
	if v := e.SelectedVersion(php); v != "8.4" {
		t.Fatalf("default = %s", v)
	}
	if env := e.Env(php); env != nil {
		t.Fatalf("no env expected without a selection, got %v", env)
	}
}
