package engine_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/hosts"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
	"github.com/prospect-ogujiuba/devarch/cli/internal/state"
	"github.com/prospect-ogujiuba/devarch/cli/internal/testutil"
)

// Ported from scripts/node/bootstrap.test.sh, which this replaced.

func appSetup(t *testing.T) (*engine.Engine, *runner.Fake, string, string) {
	t.Helper()
	root := t.TempDir()
	lib := filepath.Join(root, "services-library")
	testutil.WriteFiles(t, lib, map[string]string{
		"backend/node/compose.yml":              testutil.Compose("node", ""),
		"backend/node/app.compose.yml":          "services: {}\n",
		"proxy/nginx-proxy-manager/compose.yml": testutil.Compose("nginx-proxy-manager", ""),
	})
	testutil.WriteFiles(t, root, map[string]string{
		"apps/demo/package.json":      `{"scripts":{"devarch":"next dev --hostname 0.0.0.0 --port 3000"}}`,
		"apps/demo/package-lock.json": "",
	})
	c, _ := catalog.Load([]catalog.Path{{Dir: lib}})
	f := &runner.Fake{Responses: map[string]runner.Response{}, RunErrors: map[string][]error{}}
	hostsFile := filepath.Join(t.TempDir(), "hosts")
	os.WriteFile(hostsFile, []byte("127.0.0.1 localhost\n"), 0o644)
	e := &engine.Engine{Root: root, Catalog: c, Runner: f, Settings: state.Defaults(root),
		Hosts: &hosts.Manager{Platform: hosts.Unix, File: hostsFile, Runner: f}}
	return e, f, root, hostsFile
}

func TestPlanAppValidation(t *testing.T) {
	e, _, root, _ := appSetup(t)
	max := strings.Repeat("a", 58)
	testutil.WriteFiles(t, root, map[string]string{
		"apps/" + max + "/package.json":  `{"scripts":{"devarch":"x"}}`,
		"apps/" + max + "a/package.json": `{"scripts":{"devarch":"x"}}`,
		"apps/nopkg/README":              "",
	})
	if _, err := e.PlanApp(max, engine.AppOptions{}); err != nil {
		t.Errorf("58-character names fit the node- prefix: %v", err)
	}
	for name, opts := range map[string]engine.AppOptions{
		"Unsafe/name": {},
		max + "a":     {}, // 59 characters break the DNS label
		"missing":     {},
		"nopkg":       {},
		"demo":        {Script: "missing"},
		"demo ":       {},
	} {
		if _, err := e.PlanApp(name, opts); err == nil {
			t.Errorf("%q %+v accepted", name, opts)
		}
	}
	if _, err := e.PlanApp("demo", engine.AppOptions{Script: "rm -rf"}); err == nil {
		t.Error("unsafe script name accepted")
	}
	if _, err := e.PlanApp("demo", engine.AppOptions{PackageManager: "bun"}); err == nil {
		t.Error("unknown package manager accepted")
	}
}

func TestPlanAppDetectsPackageManager(t *testing.T) {
	e, _, root, _ := appSetup(t)
	if p, _ := e.PlanApp("demo", engine.AppOptions{}); p.PackageManager != "npm" || p.Script != "devarch" || p.Container != "node-demo" {
		t.Fatalf("plan %+v", p)
	}
	testutil.WriteFiles(t, root, map[string]string{"apps/demo/yarn.lock": ""})
	if p, _ := e.PlanApp("demo", engine.AppOptions{}); p.PackageManager != "yarn" {
		t.Fatalf("yarn: %+v", p)
	}
	testutil.WriteFiles(t, root, map[string]string{"apps/demo/pnpm-lock.yaml": ""})
	if p, _ := e.PlanApp("demo", engine.AppOptions{}); p.PackageManager != "pnpm" {
		t.Fatalf("pnpm wins over yarn: %+v", p)
	}
	if p, _ := e.PlanApp("demo", engine.AppOptions{PackageManager: "npm"}); p.PackageManager != "npm" {
		t.Fatalf("explicit choice: %+v", p)
	}
}

func TestStartApp(t *testing.T) {
	e, f, root, hostsFile := appSetup(t)
	p, err := e.PlanApp("demo", engine.AppOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.StartApp(context.Background(), p, engine.AppOptions{}); err != nil {
		t.Fatal(err)
	}
	node := filepath.Join(root, "services-library/backend/node")
	npm := filepath.Join(root, "services-library/proxy/nginx-proxy-manager")
	want := []string{
		"podman network exists microservices-net",
		"(cd " + node + " && podman compose up -d --build)",
		"(cd " + npm + " && podman compose up -d --build)",
		"(cd " + node + " && DEVARCH_NODE_APP_NAME=demo DEVARCH_NODE_SCRIPT=devarch DEVARCH_NODE_PACKAGE_MANAGER=npm DEVARCH_NODE_CONTAINER_USER=0:0 podman compose -p devarch-node-demo -f app.compose.yml up -d --build --force-recreate)",
		"(cd " + npm + " && podman compose exec -T nginx-proxy-manager nginx -t)",
		"(cd " + npm + " && podman compose exec -T nginx-proxy-manager nginx -s reload)",
	}
	if got := f.Lines(); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("commands:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if data, _ := os.ReadFile(hostsFile); !strings.Contains(string(data), "127.0.0.1\tdemo.test") {
		t.Fatalf("hosts:\n%s", data)
	}
}

func TestStartAppNoHostsDockerAndFailure(t *testing.T) {
	e, f, _, hostsFile := appSetup(t)
	e.Settings.Runtime = "docker"
	f.Responses["docker network inspect"] = runner.Response{}
	p, _ := e.PlanApp("demo", engine.AppOptions{Script: "devarch"})
	if err := e.StartApp(context.Background(), p, engine.AppOptions{NoHosts: true}); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.Lines(), "\n")
	if !strings.Contains(joined, fmt.Sprintf("DEVARCH_NODE_CONTAINER_USER=%d:%d docker compose -p devarch-node-demo", os.Getuid(), os.Getgid())) {
		t.Fatalf("docker user:\n%s", joined)
	}
	if data, _ := os.ReadFile(hostsFile); strings.Contains(string(data), "demo.test") {
		t.Fatal("--no-hosts registered the host")
	}
	f.RunErrors["docker compose -p devarch-node-demo"] = []error{fmt.Errorf("build failed")}
	if err := e.StartApp(context.Background(), p, engine.AppOptions{NoHosts: true}); err == nil || !strings.Contains(err.Error(), "node-demo") {
		t.Fatalf("got %v", err)
	}
}

func TestStopApp(t *testing.T) {
	e, f, root, _ := appSetup(t)
	if err := e.StopApp(context.Background(), "demo", true); err != nil {
		t.Fatal(err)
	}
	want := "(cd " + filepath.Join(root, "services-library/backend/node") + " && DEVARCH_NODE_APP_NAME=demo DEVARCH_NODE_SCRIPT=devarch DEVARCH_NODE_PACKAGE_MANAGER=npm DEVARCH_NODE_CONTAINER_USER=0:0 podman compose -p devarch-node-demo -f app.compose.yml down --volumes)"
	if got := f.Lines(); len(got) != 1 || got[0] != want {
		t.Fatalf("got %v", got)
	}
	if err := e.StopApp(context.Background(), "../x", false); err == nil {
		t.Fatal("accepted an unsafe name")
	}
}

func TestRepositoryAppCompose(t *testing.T) {
	data, err := os.ReadFile("../../../services-library/backend/node/app.compose.yml")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"__VITE_ADDITIONAL_SERVER_ALLOWED_HOSTS", // Vite must trust its proxied hostname
		`container_name: "node-${DEVARCH_NODE_APP_NAME`,
		"DEVARCH_NODE_CONTAINER_USER",
		"${DEVARCH_APPS_DIR:-../../../apps}/${DEVARCH_NODE_APP_NAME",
	} {
		if !strings.Contains(string(data), want) {
			t.Errorf("app.compose.yml lacks %s", want)
		}
	}
}
