package cli

import (
	"encoding/json"
	"io"
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
	cfg := t.TempDir()
	testutil.WriteFiles(t, cfg, map[string]string{"versions.env": "POSTGRES_VERSION=17\n"})
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
