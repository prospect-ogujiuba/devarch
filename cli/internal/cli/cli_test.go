package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
	"github.com/prospect-ogujiuba/devarch/cli/internal/testutil"
)

type harness struct {
	fake     *runner.Fake
	out, err strings.Builder
	repo     string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	repo := t.TempDir()
	testutil.WriteFiles(t, filepath.Join(repo, "services-library"), map[string]string{
		"database/postgres/compose.yml": testutil.Compose("postgres", "x-devarch:\n  title: PostgreSQL\n  tags: [database]\n  urls: [https://pg.test]\n  version: {var: POSTGRES_VERSION, default: \"18.2\"}\n"),
		"project/glpi/compose.yml":      testutil.Compose("glpi", "x-devarch:\n  tags: [project-management]\n"),
		"project/redmine/compose.yml":   testutil.Compose("redmine", "x-devarch:\n  tags: [project-management]\n"),
	})
	testutil.WriteFiles(t, repo, map[string]string{
		"recipes/demo/recipe.yml": "name: demo\ntitle: Demo\nentry: scripts/demo.sh\nargs:\n  - {name: name, positional: true, pattern: '^[a-z]+$'}\n",
		"scripts/demo.sh":         "#!/bin/sh\necho demo \"$@\"\n",
	})
	cfg := t.TempDir()
	testutil.WriteFiles(t, cfg, map[string]string{"versions.env": "POSTGRES_VERSION=17\n"})
	hostsFile := filepath.Join(t.TempDir(), "hosts")
	testutil.WriteFiles(t, filepath.Dir(hostsFile), map[string]string{"hosts": "127.0.0.1 localhost\n"})
	t.Setenv("HOSTS_FILE", hostsFile)
	t.Setenv("DEVARCH_HOSTS_PLATFORM", "unix")
	t.Setenv("DEVARCH_ROOT", repo)
	t.Setenv("DEVARCH_CONFIG_HOME", cfg)
	return &harness{fake: &runner.Fake{Responses: map[string]runner.Response{"podman ps": {Out: []byte("[]")}}}, repo: repo}
}

func (h *harness) run(t *testing.T, args ...string) error {
	t.Helper()
	h.out.Reset()
	h.err.Reset()
	a := &app{in: strings.NewReader(""), out: &h.out, err: &h.err, newRunner: func(io.Writer) runner.Runner { return h.fake }}
	cmd := newRoot(a)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestLsJSON(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "ls", "--json", "--tag", "database"); err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(h.out.String()), &rows); err != nil {
		t.Fatal(err, h.out.String())
	}
	if len(rows) != 1 || rows[0]["id"] != "database/postgres" || rows[0]["version"] != "17" || rows[0]["state"] != "absent" {
		t.Fatalf("rows = %v", rows)
	}
}

func TestLsTable(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "ls"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.out.String(), "database/postgres  17       -      PostgreSQL  https://pg.test") {
		t.Fatalf("table:\n%s", h.out.String())
	}
}

func TestUpByTagAppliesVersionEnv(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "up", "postgres", "--tag", "project-management"); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(h.fake.Lines(), "\n")
	for _, want := range []string{
		"postgres && POSTGRES_VERSION=17 podman compose up -d)",
		"glpi && podman compose up -d)",
		"redmine && podman compose up -d)",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in\n%s", want, joined)
		}
	}
	if !strings.Contains(h.out.String(), "https://pg.test") {
		t.Errorf("urls not printed: %s", h.out.String())
	}
}

func TestUpRequiresSelection(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "up"); err == nil || !strings.Contains(err.Error(), "at least one service") {
		t.Fatalf("got %v", err)
	}
}

func TestDryRunPrintsWithoutRunning(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "--dry-run", "down", "postgres", "--volumes"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.out.String(), "would run: (cd ") || !strings.Contains(h.out.String(), "podman compose down --volumes)") {
		t.Fatalf("out = %s", h.out.String())
	}
	for _, l := range h.fake.Lines() {
		if strings.Contains(l, "down") {
			t.Fatalf("dry run executed %s", l)
		}
	}
}

func TestDownVolumesRefusesWithoutTTY(t *testing.T) {
	h := newHarness(t)
	err := h.run(t, "down", "postgres", "--volumes")
	if err == nil || !strings.Contains(err.Error(), "postgres_postgres_data") || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("got %v", err)
	}
	if err := h.run(t, "down", "postgres", "--volumes", "--yes"); err != nil {
		t.Fatal(err)
	}
}

func TestComposePassthrough(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "compose", "postgres", "--", "exec", "postgres", "psql", "-U", "postgres"); err != nil {
		t.Fatal(err)
	}
	lines := h.fake.Lines()
	if last := lines[len(lines)-1]; !strings.HasSuffix(last, "POSTGRES_VERSION=17 podman compose exec postgres psql -U postgres)") {
		t.Fatalf("got %s", last)
	}
}

func TestLogs(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "logs", "postgres", "-f", "-n", "50"); err != nil {
		t.Fatal(err)
	}
	lines := h.fake.Lines()
	if last := lines[len(lines)-1]; !strings.HasSuffix(last, "podman compose logs -f --tail 50)") {
		t.Fatalf("got %s", last)
	}
}

func TestUpRegistersMissingHostsOnce(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "up", "postgres"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(os.Getenv("HOSTS_FILE"))
	for _, want := range []string{"127.0.0.1 localhost\n\n# BEGIN DEVARCH HOSTS", "pg.test", "postgres.test", "glpi.test"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("missing %q in hosts:\n%s", want, data)
		}
	}
	if !strings.Contains(h.err.String(), "registering") {
		t.Fatalf("no registration message: %s", h.err.String())
	}
	if err := h.run(t, "up", "postgres"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(h.err.String(), "registering") {
		t.Fatal("registered again although every hostname is mapped")
	}
}

func TestUpNoHosts(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "up", "postgres", "--no-hosts"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(os.Getenv("HOSTS_FILE"))
	if strings.Contains(string(data), "DEVARCH") {
		t.Fatal("--no-hosts wrote the hosts file")
	}
}

func TestHostsCommands(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "--dry-run", "hosts", "sync"); err != nil || !strings.Contains(h.out.String(), "127.0.0.1 glpi.test pg.test postgres.test redmine.test") {
		t.Fatalf("dry run: %v\n%s", err, h.out.String())
	}
	if err := h.run(t, "hosts", "list"); err == nil {
		t.Fatal("list without a block should fail")
	}
	if err := h.run(t, "hosts", "sync"); err != nil || !strings.Contains(h.out.String(), "synchronized 5 domains") {
		t.Fatalf("sync: %v %s", err, h.out.String())
	}
	if err := h.run(t, "hosts", "list"); err != nil || strings.Contains(h.err.String(), "out of date") {
		t.Fatalf("list: %v %s", err, h.err.String())
	}
	if err := h.run(t, "hosts", "add", "demo.test"); err != nil || !strings.Contains(h.out.String(), "registered 127.0.0.1 demo.test") {
		t.Fatalf("add: %v %s", err, h.out.String())
	}
	if err := h.run(t, "hosts", "remove", "demo.test"); err != nil || !strings.Contains(h.out.String(), "removed demo.test") {
		t.Fatalf("remove: %v %s", err, h.out.String())
	}
	if err := h.run(t, "hosts", "add", "Bad Name"); err == nil {
		t.Fatal("accepted invalid hostname")
	}
}

func TestNew(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "new"); err != nil || !strings.Contains(h.out.String(), "demo    Demo") {
		t.Fatalf("list: %v\n%s", err, h.out.String())
	}
	if err := h.run(t, "new", "demo", "Bad"); err == nil {
		t.Fatal("invalid name accepted")
	}
	if err := h.run(t, "new", "nope", "x"); err == nil || !strings.Contains(err.Error(), "available: demo") {
		t.Fatalf("unknown recipe: %v", err)
	}
	if err := h.run(t, "new", "demo", "shop", "--extra", "flag"); err != nil {
		t.Fatal(err)
	}
	last := h.fake.Cmds[len(h.fake.Cmds)-1]
	if !strings.HasSuffix(last.Name, "scripts/demo.sh") || strings.Join(last.Args, " ") != "shop --extra flag" {
		t.Fatalf("exec = %+v", last)
	}
	env := strings.Join(last.Env, " ")
	if !strings.Contains(env, "DEVARCH_ROOT="+h.repo) || !strings.Contains(env, "DEVARCH_BIN=") {
		t.Fatalf("env = %s", env)
	}
	if err := h.run(t, "new", "--dry-run", "demo", "shop"); err != nil {
		t.Fatal(err)
	}
	last = h.fake.Cmds[len(h.fake.Cmds)-1]
	if strings.Join(last.Args, " ") != "shop --dry-run" {
		t.Fatalf("dry run should be forwarded to the script: %v", last.Args)
	}
}

func TestUserServiceOverlay(t *testing.T) {
	h := newHarness(t)
	testutil.WriteFiles(t, os.Getenv("DEVARCH_CONFIG_HOME"), map[string]string{
		"services/custom/tool/compose.yml":       testutil.Compose("tool", "x-devarch:\n  title: My Tool\n"),
		"services/database/postgres/compose.yml": testutil.Compose("postgres", "x-devarch:\n  title: My Postgres\n"),
	})
	if err := h.run(t, "ls", "--json"); err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	json.Unmarshal([]byte(h.out.String()), &rows)
	got := map[string]string{}
	for _, r := range rows {
		got[fmt.Sprint(r["id"])] = fmt.Sprint(r["source"]) + ":" + fmt.Sprint(r["meta"].(map[string]any)["title"])
	}
	if got["custom/tool"] != "user:My Tool" || got["database/postgres"] != "user:My Postgres" {
		t.Fatalf("overlay rows = %v", got)
	}
}

func (h *harness) config(t *testing.T, yml string) {
	t.Helper()
	testutil.WriteFiles(t, os.Getenv("DEVARCH_CONFIG_HOME"), map[string]string{"config.yml": yml})
}

func TestConfigShowGetSet(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "config"); err != nil {
		t.Fatal(err)
	}
	fields := func() string { return strings.Join(strings.Fields(h.out.String()), " ") }
	for _, want := range []string{"root " + h.repo + " DEVARCH_ROOT", "runtime podman default", "apps_dir " + filepath.Join(h.repo, "apps") + " default", "hosts.manage true default"} {
		if !strings.Contains(fields(), want) {
			t.Fatalf("missing %q in:\n%s", want, h.out.String())
		}
	}
	if err := h.run(t, "config", "set", "network", "dev-net"); err != nil {
		t.Fatal(err)
	}
	if err := h.run(t, "config", "set", "colour", "blue"); err == nil || !strings.Contains(err.Error(), "settable keys") {
		t.Fatalf("unknown key: %v", err)
	}
	if err := h.run(t, "config", "set", "runtime", "lxc"); err == nil {
		t.Fatal("accepted runtime lxc")
	}
	if err := h.run(t, "config", "get", "network"); err != nil || h.out.String() != "dev-net\n" {
		t.Fatalf("get: %q %v", h.out.String(), err)
	}
	if err := h.run(t, "config"); err != nil || !strings.Contains(fields(), "network dev-net config.yml") {
		t.Fatalf("show after set:\n%s", h.out.String())
	}
}

func TestConfigReachesComposeAndRecipes(t *testing.T) {
	h := newHarness(t)
	h.config(t, "network: dev-net\napps_dir: sites\n")
	if err := h.run(t, "up", "postgres", "--no-hosts"); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(h.fake.Lines(), "\n")
	for _, want := range []string{
		"podman network exists dev-net",
		"POSTGRES_VERSION=17 DEVARCH_NETWORK=dev-net DEVARCH_APPS_DIR=" + filepath.Join(h.repo, "sites") + " podman compose up -d",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in\n%s", want, joined)
		}
	}
	if err := h.run(t, "new", "demo", "shop"); err != nil {
		t.Fatal(err)
	}
	last := h.fake.Lines()[len(h.fake.Lines())-1]
	if !strings.Contains(last, "DEVARCH_APPS_DIR="+filepath.Join(h.repo, "sites")) || !strings.Contains(last, "DEVARCH_NETWORK=dev-net") {
		t.Fatalf("recipe env: %s", last)
	}
}

func TestDefaultConfigAddsNoEnvironment(t *testing.T) {
	h := newHarness(t)
	if err := h.run(t, "up", "glpi", "--no-hosts"); err != nil {
		t.Fatal(err)
	}
	if joined := strings.Join(h.fake.Lines(), "\n"); strings.Contains(joined, "DEVARCH_") {
		t.Fatalf("defaults leaked into compose:\n%s", joined)
	}
}

func TestHostsManageFalse(t *testing.T) {
	h := newHarness(t)
	h.config(t, "hosts:\n  manage: false\n")
	before, _ := os.ReadFile(os.Getenv("HOSTS_FILE"))
	for _, args := range [][]string{{"up", "postgres"}, {"hosts", "sync"}, {"hosts", "add", "demo.test"}, {"hosts", "remove", "demo.test"}} {
		if err := h.run(t, args...); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}
	if after, _ := os.ReadFile(os.Getenv("HOSTS_FILE")); string(after) != string(before) {
		t.Fatalf("hosts file changed:\n%s", after)
	}
	if !strings.Contains(h.err.String(), "hosts.manage is false") {
		t.Fatalf("no explanation: %s", h.err.String())
	}
}

func TestHostsAddressSetting(t *testing.T) {
	h := newHarness(t)
	h.config(t, "hosts:\n  address: 127.0.0.2\n")
	if err := h.run(t, "hosts", "add", "demo.test"); err != nil || !strings.Contains(h.out.String(), "registered 127.0.0.2 demo.test") {
		t.Fatalf("add: %v %s", err, h.out.String())
	}
	if err := h.run(t, "--dry-run", "hosts", "sync"); err != nil || !strings.Contains(h.out.String(), "127.0.0.2 glpi.test") {
		t.Fatalf("sync: %v %s", err, h.out.String())
	}
}

func TestInvalidConfigIsReported(t *testing.T) {
	h := newHarness(t)
	h.config(t, "runtime: lxc\n")
	if err := h.run(t, "ls"); err == nil || !strings.Contains(err.Error(), "config.yml") {
		t.Fatalf("got %v", err)
	}
}

func TestDockerRuntimeReachesCommandsAndRecipes(t *testing.T) {
	h := newHarness(t)
	h.config(t, "runtime: docker\n")
	h.fake.Responses["docker ps"] = runner.Response{}
	if err := h.run(t, "up", "postgres", "--no-hosts"); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(h.fake.Lines(), "\n")
	if !strings.Contains(joined, "POSTGRES_VERSION=17 docker compose up -d") || strings.Contains(joined, "podman") {
		t.Fatalf("commands:\n%s", joined)
	}
	if err := h.run(t, "new", "demo", "shop"); err != nil {
		t.Fatal(err)
	}
	last := h.fake.Lines()[len(h.fake.Lines())-1]
	user := fmt.Sprintf("DEVARCH_CONTAINER_USER=%d:%d", os.Getuid(), os.Getgid())
	if !strings.Contains(last, "DEVARCH_RUNTIME=docker") || !strings.Contains(last, user) {
		t.Fatalf("recipe env: %s", last)
	}
	if err := h.run(t, "config", "--env"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"DEVARCH_RUNTIME=docker\n", user + "\n", "DEVARCH_APPS_DIR=" + filepath.Join(h.repo, "apps") + "\n", "DEVARCH_NETWORK=microservices-net\n"} {
		if !strings.Contains(h.out.String(), want) {
			t.Errorf("config --env missing %q:\n%s", want, h.out.String())
		}
	}
}

func TestDBCommands(t *testing.T) {
	h := newHarness(t)
	testutil.WriteFiles(t, filepath.Join(h.repo, "services-library"), map[string]string{
		"database/mariadb/compose.yml": testutil.Compose("mariadb", ""),
	})
	if err := h.run(t, "db", "create", "shop", "--env"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.out.String(), "DB_HOST=mariadb\nDB_PORT=3306\nDB_NAME=shop\nDB_USER=shop\nDB_PASSWORD=") {
		t.Fatalf("env output:\n%s", h.out.String())
	}
	if err := h.run(t, "db", "drop", "shop"); err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("drop without a terminal: %v", err)
	}
	if err := h.run(t, "--dry-run", "db", "drop", "shop", "--user", "shop"); err != nil || !strings.Contains(h.out.String(), "would run: podman exec -i mariadb") {
		t.Fatalf("dry-run drop: %v %s", err, h.out.String())
	}
}
