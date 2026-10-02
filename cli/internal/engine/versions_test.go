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

func useSetup(t *testing.T) (*engine.Engine, *runner.Fake, *state.Versions, string) {
	t.Helper()
	lib := filepath.Join(t.TempDir(), "services-library")
	testutil.WriteFiles(t, lib, map[string]string{
		"backend/php/compose.yml":       testutil.Compose("php", "x-devarch:\n  version: {var: PHP_VERSION, default: \"8.4\", choices: [\"8.4\", \"8.3\"], rebuild: true}\n"),
		"database/postgres/compose.yml": testutil.Compose("postgres", "x-devarch:\n  version: {var: POSTGRES_VERSION, default: \"18.2\", choices: [\"18.2\", \"18\", \"17\"], data: same-major}\n"),
		"database/mariadb/compose.yml":  testutil.Compose("mariadb", "x-devarch:\n  version: {var: MARIADB_VERSION, default: \"12.3\", choices: [\"12.3\", \"11.8\"], data: upgrade-only}\n"),
		"misc/plain/compose.yml":        testutil.Compose("plain", ""),
	})
	c, _ := catalog.Load([]catalog.Path{{Dir: lib}})
	f := &runner.Fake{Responses: map[string]runner.Response{"podman ps": {Out: []byte("[]")}}, RunErrors: map[string][]error{}}
	saved := &state.Versions{}
	e := &engine.Engine{Catalog: c, Runner: f, Versions: state.Versions{}, Sleep: func(time.Duration) {},
		SaveVersions: func(v state.Versions) error {
			cp := state.Versions{}
			for k, val := range v {
				cp[k] = val
			}
			*saved = cp
			return nil
		}}
	return e, f, saved, lib
}

func TestUseRebuildsAndRecreatesRunningService(t *testing.T) {
	e, f, saved, lib := useSetup(t)
	f.Responses["podman ps"] = runner.Response{Out: []byte(`[{"Names":["php"],"State":"running","Status":"Up","Labels":{"com.docker.compose.project":"php","com.docker.compose.project.working_dir":"` + lib + `/backend/php"}}]`)}
	php, _ := e.Catalog.Get("backend/php")
	res, err := e.Use(context.Background(), php, "8.3", engine.UseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if (*saved)["PHP_VERSION"] != "8.3" || !res.Rebuilt || !res.Recreated || res.From != "8.4" {
		t.Fatalf("res=%+v saved=%v", res, *saved)
	}
	joined := strings.Join(f.Lines(), "\n")
	for _, want := range []string{"PHP_VERSION=8.3 podman compose build)", "PHP_VERSION=8.3 podman compose up -d --force-recreate)"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in\n%s", want, joined)
		}
	}
}

func TestUseDefaultClearsEntryAndSkipsRecreateWhenStopped(t *testing.T) {
	e, f, saved, _ := useSetup(t)
	e.Versions["PHP_VERSION"] = "8.3"
	php, _ := e.Catalog.Get("backend/php")
	res, err := e.Use(context.Background(), php, "8.4", engine.UseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := (*saved)["PHP_VERSION"]; ok || res.Recreated {
		t.Fatalf("saved=%v res=%+v", *saved, res)
	}
	for _, l := range f.Lines() {
		if strings.Contains(l, "--force-recreate") {
			t.Fatalf("recreated a stopped service: %s", l)
		}
	}
}

func TestUseRejectsUnlistedUnlessForced(t *testing.T) {
	e, _, _, _ := useSetup(t)
	php, _ := e.Catalog.Get("backend/php")
	if _, err := e.Use(context.Background(), php, "7.4", engine.UseOptions{}); err == nil || !strings.Contains(err.Error(), "choices: 8.4, 8.3") {
		t.Fatalf("got %v", err)
	}
	if _, err := e.Use(context.Background(), php, "7.4", engine.UseOptions{Force: true, NoRecreate: true}); err != nil {
		t.Fatal(err)
	}
}

func TestUseDataRules(t *testing.T) {
	cases := []struct {
		id, to      string
		volume      bool
		wantRefusal bool
	}{
		{"database/postgres", "17", true, true},    // major change with data
		{"database/postgres", "17", false, false},  // major change, no data yet
		{"database/postgres", "18", true, false},   // same major
		{"database/mariadb", "11.8", true, true},   // downgrade with data
		{"database/mariadb", "11.8", false, false}, // downgrade, no data
	}
	for _, tc := range cases {
		e, f, _, _ := useSetup(t)
		if !tc.volume {
			f.Responses["podman volume exists"] = runner.Response{Err: &runner.ExitError{Code: 1}}
		}
		svc, _ := e.Catalog.Get(tc.id)
		_, err := e.Use(context.Background(), svc, tc.to, engine.UseOptions{NoRecreate: true})
		var conflict *engine.DataConflictError
		if got := errors.As(err, &conflict); got != tc.wantRefusal {
			t.Errorf("%s -> %s (volume=%v): err=%v", tc.id, tc.to, tc.volume, err)
		}
	}
	e, _, _, _ := useSetup(t)
	pg, _ := e.Catalog.Get("database/postgres")
	if _, err := e.Use(context.Background(), pg, "17", engine.UseOptions{Force: true, NoRecreate: true}); err != nil {
		t.Fatalf("--force should skip the data check: %v", err)
	}
}

func TestUseDryRunDoesNotSave(t *testing.T) {
	e, _, saved, _ := useSetup(t)
	e.DryRun = true
	php, _ := e.Catalog.Get("backend/php")
	if _, err := e.Use(context.Background(), php, "8.3", engine.UseOptions{}); err != nil {
		t.Fatal(err)
	}
	if len(*saved) != 0 {
		t.Fatalf("dry run saved %v", *saved)
	}
}

func TestUseWithoutVersionMeta(t *testing.T) {
	e, _, _, _ := useSetup(t)
	plain, _ := e.Catalog.Get("misc/plain")
	if _, err := e.Use(context.Background(), plain, "1", engine.UseOptions{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestCompareVersions(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"8.3", "8.4", -1}, {"12.3", "11.8", 1}, {"12.3", "12.3.2", 0}, {"v1.29", "v1.22", 1}, {"18", "18.2", 0}, {"10.11", "10.9", 1},
	} {
		if got := engine.CompareVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("Compare(%s,%s)=%d want %d", tc.a, tc.b, got, tc.want)
		}
	}
	if engine.Major("v1.22") != "1" || engine.Major("18.2") != "18" {
		t.Fatal("Major")
	}
}
