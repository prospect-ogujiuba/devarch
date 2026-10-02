package lint

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/testutil"
)

func run(t *testing.T, files map[string]string) []Finding {
	t.Helper()
	dir := t.TempDir()
	testutil.WriteFiles(t, dir, files)
	c, problems := catalog.Load([]catalog.Path{{Dir: dir}})
	return Run(context.Background(), c, problems, Options{})
}

func has(fs []Finding, sev Severity, svc, substr string) bool {
	for _, f := range fs {
		if f.Severity == sev && strings.Contains(f.Service, svc) && strings.Contains(f.Message, substr) {
			return true
		}
	}
	return false
}

func TestCleanCatalog(t *testing.T) {
	fs := run(t, map[string]string{
		"db/pg/compose.yml": `x-devarch:
  version: {var: PG_VERSION, default: "18", choices: ["18", "17"], data: same-major}
  urls: [https://pg.test]
  requires: [db/cache]
services:
  pg:
    image: postgres:${PG_VERSION:-18}
    container_name: pg
    ports: ["127.0.0.1:5432:5432"]
`,
		"db/cache/compose.yml":  "services:\n  cache:\n    image: redis:8\n    ports: [\"127.0.0.1:6379:6379\"]\n",
		"app/built/compose.yml": "services:\n  built:\n    build: ./config\n",
	})
	if len(fs) != 0 {
		t.Fatalf("unexpected findings: %+v", fs)
	}
}

func TestFindings(t *testing.T) {
	fs := run(t, map[string]string{
		"a/one/compose.yml": `x-devarch:
  version: {var: lower, default: "1"}
  urls: [not-a-url]
  requires: [missing]
  ready: {exec: []}
  tags: [Bad_Tag]
services:
  one:
    image: thing
    container_name: Bad_Name
    ports: ["127.0.0.1:8000:80"]
`,
		"a/two/compose.yml": `x-devarch:
  version: {var: TWO_VERSION, default: "2", choices: ["3"], data: sometimes}
services:
  two:
    image: two:latest
    container_name: shared
    ports: ["8000:80"]
  twin:
    image: two:latest
    container_name: shared
`,
		"a/three/compose.yml": "services:\n  three:\n    image: x:1\n    container_name: shared\n    ports: [\"127.0.0.2:8000:80\"]\n",
	})
	for _, want := range []struct {
		sev     Severity
		svc, in string
	}{
		{Error, "a/one", "version.var"},
		{Error, "a/one", "absolute http(s) URL"},
		{Error, "a/one", "requires missing"},
		{Error, "a/one", "ready.exec"},
		{Warning, "a/one", "kebab-case"},
		{Error, "a/one", "not a valid DNS label"},
		{Warning, "a/one", "image thing is not pinned"},
		{Error, "a/two", "must reference ${TWO_VERSION:-2}"},
		{Error, "a/two", "not in version.choices"},
		{Error, "a/two", "version.data"},
		{Error, "a/two", "host port 8000/tcp is also published by a/one"},
		{Error, "a/two", "host port 8000/tcp is also published by a/three"},
		{Warning, "a/two", "is used by more than one service"},
	} {
		if !has(fs, want.sev, want.svc, want.in) {
			t.Errorf("missing %s %s %q in\n%+v", want.sev, want.svc, want.in, fs)
		}
	}
	if has(fs, Error, "a/three", "also published by a/one") {
		t.Error("127.0.0.1 and 127.0.0.2 do not collide")
	}
	pinned := 0
	for _, f := range fs {
		if f.Service == "a/two" && strings.Contains(f.Message, "two:latest") {
			pinned++
		}
	}
	if pinned != 1 {
		t.Errorf("duplicate findings were not compacted: %d", pinned)
	}
}

func TestProblemsAreErrors(t *testing.T) {
	fs := run(t, map[string]string{"x/broken/README.md": ""})
	if e, _ := Count(fs); e != 1 || !strings.Contains(fs[0].Service, filepath.Join("x", "broken")) {
		t.Fatalf("%+v", fs)
	}
}

func TestSharedNetworkMustBeNamed(t *testing.T) {
	net := "networks:\n  microservices-net:\n    external: true\n"
	fs := run(t, map[string]string{
		"a/bare/compose.yml":  "services:\n  bare:\n    image: x:1\n" + net,
		"a/named/compose.yml": "services:\n  named:\n    image: x:1\n" + net + "    name: ${DEVARCH_NETWORK:-microservices-net}\n",
	})
	if !has(fs, Error, "a/bare", "DEVARCH_NETWORK") || has(fs, Error, "a/named", "DEVARCH_NETWORK") {
		t.Fatalf("findings: %+v", fs)
	}
}
