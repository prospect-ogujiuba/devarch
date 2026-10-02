package catalog_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/testutil"
)

func fixture(t *testing.T) (builtin, user string) {
	t.Helper()
	base := t.TempDir()
	builtin = filepath.Join(base, "services-library")
	user = filepath.Join(base, "user")
	testutil.WriteFiles(t, builtin, map[string]string{
		"database/postgres/compose.yml": testutil.Compose("postgres", `x-devarch:
  title: PostgreSQL
  tags: [database, sql]
  version: {var: POSTGRES_VERSION, default: "18.2", choices: ["18.2", "17"], data: same-major}
`),
		"database/redis/compose.yml":        testutil.Compose("redis", "x-devarch:\n  tags: [cache]\n"),
		"exporters/redis/compose.yml":       testutil.Compose("redis", ""),
		"dbms/adminer/compose.yml":          testutil.Compose("adminer", "x-devarch:\n  requires: [postgres]\n"),
		"broken/nofile/README.md":           "no compose here",
		"broken/empty-services/compose.yml": "services: {}\n",
		"Bad_Category/x/compose.yml":        testutil.Compose("x", ""),
	})
	testutil.WriteFiles(t, user, map[string]string{
		"database/postgres/compose.yml": testutil.Compose("postgres", "x-devarch:\n  title: My Postgres\n"),
		"custom/thing/compose.yml":      testutil.Compose("thing", ""),
	})
	return builtin, user
}

func load(t *testing.T) (*catalog.Catalog, []catalog.Problem) {
	builtin, user := fixture(t)
	return catalog.Load([]catalog.Path{{Dir: user, Source: "user"}, {Dir: builtin, Source: "builtin"}})
}

func TestLoadOverlayAndProblems(t *testing.T) {
	c, problems := load(t)
	var ids []string
	for _, s := range c.Services {
		ids = append(ids, s.ID)
	}
	want := "custom/thing database/postgres database/redis dbms/adminer exporters/redis"
	if got := strings.Join(ids, " "); got != want {
		t.Fatalf("ids = %q, want %q", got, want)
	}
	pg, _ := c.Get("database/postgres")
	if pg.Source != "user" || !pg.Overrides || pg.Meta.Title != "My Postgres" {
		t.Fatalf("user overlay did not win: %+v", pg)
	}
	if len(problems) != 3 {
		t.Fatalf("want 3 problems (missing compose, empty services, bad category), got %v", problems)
	}
}

func TestMetaParsing(t *testing.T) {
	builtin, _ := fixture(t)
	c, _ := catalog.Load([]catalog.Path{{Dir: builtin, Source: "builtin"}})
	pg, ok := c.Get("database/postgres")
	if !ok {
		t.Fatal("postgres missing")
	}
	v := pg.Meta.Version
	if v == nil || v.Var != "POSTGRES_VERSION" || v.Default != "18.2" || v.Data != "same-major" || len(v.Choices) != 2 {
		t.Fatalf("version meta = %+v", v)
	}
	if pg.Primary().ContainerName != "postgres" || strings.Join(pg.Volumes, " ") != "postgres_postgres_data" {
		t.Fatalf("containers/volumes = %+v %v", pg.Containers, pg.Volumes)
	}
}

func TestResolve(t *testing.T) {
	c, _ := load(t)
	if s, err := c.Resolve("postgres"); err != nil || s.ID != "database/postgres" {
		t.Fatalf("short name: %v %v", s.ID, err)
	}
	if s, err := c.Resolve("exporters/redis"); err != nil || s.ID != "exporters/redis" {
		t.Fatalf("exact id: %v %v", s.ID, err)
	}
	var amb *catalog.AmbiguousError
	if _, err := c.Resolve("redis"); !errors.As(err, &amb) || len(amb.Candidates) != 2 {
		t.Fatalf("want ambiguous error, got %v", err)
	}
	for _, q := range []string{"", "nope", "a/b/c", "database/nope", "Bad Name"} {
		if _, err := c.Resolve(q); err == nil {
			t.Errorf("Resolve(%q) succeeded", q)
		}
	}
}

func TestSearch(t *testing.T) {
	c, _ := load(t)
	// The user overlay replaces postgres metadata, so its sql tag no longer matches.
	if got := c.Search("sql", ""); len(got) != 0 {
		t.Fatalf("search sql = %v", got)
	}
	if got := c.Search("", "cache"); len(got) != 1 || got[0].ID != "database/redis" {
		t.Fatalf("tag search = %v", got)
	}
	if got := c.Search("REDIS", ""); len(got) != 2 {
		t.Fatalf("case-insensitive search = %v", got)
	}
}

func TestWithRequires(t *testing.T) {
	c, _ := load(t)
	adminer, _ := c.Get("dbms/adminer")
	redis, _ := c.Get("database/redis")
	got, err := c.WithRequires([]catalog.Service{adminer, redis, adminer})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, s := range got {
		ids = append(ids, s.ID)
	}
	if strings.Join(ids, " ") != "database/postgres dbms/adminer database/redis" {
		t.Fatalf("order = %v", ids)
	}
}

func TestWithRequiresCycle(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFiles(t, dir, map[string]string{
		"a/one/compose.yml": testutil.Compose("one", "x-devarch:\n  requires: [two]\n"),
		"a/two/compose.yml": testutil.Compose("two", "x-devarch:\n  requires: [a/one]\n"),
	})
	c, _ := catalog.Load([]catalog.Path{{Dir: dir}})
	one, _ := c.Get("a/one")
	if _, err := c.WithRequires([]catalog.Service{one}); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("want cycle error, got %v", err)
	}
}

func TestSymlinkedServiceIsRejected(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFiles(t, dir, map[string]string{"real/svc/compose.yml": testutil.Compose("svc", "")})
	if err := os.Symlink(filepath.Join(dir, "real", "svc"), filepath.Join(dir, "real", "link")); err != nil {
		t.Skip(err)
	}
	c, problems := catalog.Load([]catalog.Path{{Dir: dir}})
	if _, ok := c.Get("real/link"); ok || len(problems) != 1 {
		t.Fatalf("symlink accepted: problems=%v", problems)
	}
}

func TestVolumeNames(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFiles(t, dir, map[string]string{"a/svc/compose.yml": `services:
  svc: {image: x}
volumes:
  plain: null
  named: {name: custom_name}
  shared: {external: true}
`})
	c, _ := catalog.Load([]catalog.Path{{Dir: dir}})
	s, _ := c.Get("a/svc")
	if got := strings.Join(s.Volumes, " "); got != "custom_name svc_plain" {
		t.Fatalf("volumes = %q", got)
	}
}

func TestMissingSearchPathIsSkipped(t *testing.T) {
	c, problems := catalog.Load([]catalog.Path{{Dir: filepath.Join(t.TempDir(), "absent")}})
	if len(c.Services) != 0 || len(problems) != 0 {
		t.Fatalf("got %v %v", c.Services, problems)
	}
}
