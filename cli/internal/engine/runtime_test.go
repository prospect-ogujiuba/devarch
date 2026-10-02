package engine_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

func dockerSetup(t *testing.T) (*engine.Engine, *runner.Fake, string) {
	e, f, lib := setup(t)
	e.Settings.Runtime = "docker"
	return e, f, lib
}

func TestDockerUpUsesDockerCommands(t *testing.T) {
	e, f, lib := dockerSetup(t)
	f.Responses["docker network inspect"] = runner.Response{Err: &runner.ExitError{Code: 1}}
	if err := e.Up(context.Background(), []catalog.Service{get(t, e, "backend/php")}, engine.UpOptions{}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"docker network inspect --format '{{.Name}}' microservices-net",
		"docker network create microservices-net",
		"(cd " + filepath.Join(lib, "backend/php") + " && PHP_VERSION=8.3 docker compose up -d)",
	}
	if got := f.Lines(); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("commands:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

const dockerPS = `{"Names":"php","Image":"php-php","State":"running","Status":"Up 2 hours (healthy)","Labels":"com.docker.compose.project=php,com.docker.compose.service=php,com.docker.compose.project.working_dir=%LIB%/backend/php,com.docker.compose.project.config_files=%LIB%/backend/php/compose.yml,extra.yml","Ports":"127.0.0.1:8100->8000/tcp, 9000/tcp"}
{"Names":"npm","Image":"npm","State":"running","Status":"Up 1 minute (health: starting)","Labels":"com.docker.compose.project=npm,com.docker.compose.service=npm,com.docker.compose.project.working_dir=%LIB%/proxy/npm","Ports":":::443->443/tcp, 0.0.0.0:443->443/tcp"}
{"Names":"stray","Image":"z","State":"exited","Status":"Exited (0) 2 days ago","Labels":"","Ports":""}
`

func TestDockerContainers(t *testing.T) {
	e, f, lib := dockerSetup(t)
	f.Responses["docker ps"] = runner.Response{Out: []byte(strings.ReplaceAll(dockerPS, "%LIB%", lib))}
	cs, err := e.Containers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 3 {
		t.Fatalf("got %+v", cs)
	}
	byName := map[string]engine.Container{}
	for _, c := range cs {
		byName[c.Name] = c
	}
	php := byName["php"]
	if php.Service != "backend/php" || php.Compose != "php" || php.Health != "healthy" ||
		len(php.Ports) != 1 || php.Ports[0] != (engine.Port{HostIP: "127.0.0.1", HostPort: 8100, ContainerPort: 8000, Protocol: "tcp"}) {
		t.Fatalf("php = %+v", php)
	}
	if npm := byName["npm"]; npm.Service != "proxy/npm" || npm.Health != "starting" || len(npm.Ports) != 2 || npm.Ports[0].HostIP != "::" {
		t.Fatalf("npm = %+v", npm)
	}
	if byName["stray"].Service != "" {
		t.Fatal("stray mapped")
	}
}

func TestDockerWaitReadsHealthFromInspect(t *testing.T) {
	e, f, lib := dockerSetup(t)
	f.Responses["docker ps"] = runner.Response{Out: []byte(strings.ReplaceAll(dockerPS, "%LIB%", lib))}
	f.Responses["docker inspect"] = runner.Response{Out: []byte(`[{"Name":"/npm","State":{"Status":"running","Health":{"Status":"healthy"}},"Config":{"Healthcheck":{"Test":["CMD","true"]}}}]`)}
	if err := e.Wait(context.Background(), get(t, e, "proxy/npm"), time.Second); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.Lines(), "\n")
	if strings.Contains(joined, "healthcheck run") || !strings.Contains(joined, "docker exec npm curl -f localhost") {
		t.Fatalf("commands:\n%s", joined)
	}
	f.Responses["docker inspect"] = runner.Response{Out: []byte(`[{"Name":"/npm","State":{"Status":"running","Health":{"Status":"starting"}},"Config":{"Healthcheck":{"Test":["CMD","true"]}}}]`)}
	if err := e.Wait(context.Background(), get(t, e, "proxy/npm"), time.Nanosecond); err == nil || !strings.Contains(err.Error(), "not healthy yet") {
		t.Fatalf("got %v", err)
	}
}

func TestDockerVolumeCheck(t *testing.T) {
	e, f, _, _ := useSetup(t)
	e.Settings.Runtime = "docker"
	f.Responses["docker ps"] = runner.Response{}
	pg, _ := e.Catalog.Get("database/postgres")
	if _, err := e.Use(context.Background(), pg, "17", engine.UseOptions{}); err == nil {
		t.Fatal("downgrade allowed while the volume exists")
	}
	if !strings.Contains(strings.Join(f.Lines(), "\n"), "docker volume inspect --format '{{.Name}}' postgres_postgres_data") {
		t.Fatalf("commands:\n%s", strings.Join(f.Lines(), "\n"))
	}
}

func TestRuntimeTraits(t *testing.T) {
	e, _, _ := setup(t)
	if u := e.Runtime().ContainerUser(); u != "0:0" {
		t.Fatalf("podman user %s", u)
	}
	if !strings.Contains(strings.Join(e.EventsCmd(), " "), "event=died") {
		t.Fatal(e.EventsCmd())
	}
	e.Settings.Runtime = "docker"
	if u := e.Runtime().ContainerUser(); u != fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()) {
		t.Fatalf("docker user %s", u)
	}
	if args := strings.Join(e.EventsCmd(), " "); !strings.Contains(args, "event=die ") || !strings.Contains(args, "{{json .}}") {
		t.Fatal(args)
	}
}

func TestParseEvent(t *testing.T) {
	for line, want := range map[string]engine.Event{
		`{"Name":"php","Status":"health_status","HealthStatus":"healthy"}`:                                                {Name: "php", Action: "health_status", Health: "healthy"},
		`{"Name":"php","Status":"start"}`:                                                                                 {Name: "php", Action: "start"},
		`{"status":"health_status: unhealthy","Action":"health_status: unhealthy","Actor":{"Attributes":{"name":"npm"}}}`: {Name: "npm", Action: "health_status", Health: "unhealthy"},
		`{"Action":"die","Actor":{"Attributes":{"name":"npm"}}}`:                                                          {Name: "npm", Action: "die"},
	} {
		if got, ok := engine.ParseEvent([]byte(line)); !ok || got != want {
			t.Errorf("%s: got %+v", line, got)
		}
	}
	if _, ok := engine.ParseEvent([]byte("not json")); ok {
		t.Error("parsed garbage")
	}
}
