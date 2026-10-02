package replace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
)

type fakeDB struct {
	calls   []string
	failing map[string]int // call prefix -> remaining failures
}

func (f *fakeDB) do(call string) error {
	f.calls = append(f.calls, call)
	for prefix, n := range f.failing {
		if strings.HasPrefix(call, prefix) && n > 0 {
			f.failing[prefix] = n - 1
			return errors.New("injected " + prefix + " failure")
		}
	}
	return nil
}

func (f *fakeDB) DropDB(_ context.Context, engine, name, user string) error {
	return f.do("drop " + engine + " " + name + "/" + user)
}

func (f *fakeDB) RestoreDB(_ context.Context, engine, name, user, dump, prevUser string) error {
	return f.do("restore " + engine + " " + name + "/" + user + " " + filepath.Base(dump) + " " + prevUser)
}

var now = time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)

func paths(t *testing.T) Paths {
	return Paths{AppsDir: t.TempDir(), App: "demo", Now: func() time.Time { return now }}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(path string) string {
	data, _ := os.ReadFile(path)
	return string(data)
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func TestGuardBlocksASecondRun(t *testing.T) {
	p := paths(t)
	r, err := Guard(p, false)
	if err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Stat(p.RecordPath()); info.Mode().Perm() != 0o600 {
		t.Fatalf("record mode %v", info.Mode())
	}
	if _, err := Guard(p, false); err == nil || !strings.Contains(err.Error(), "devarch app recover demo") {
		t.Fatalf("second guard: %v", err)
	}
	if err := r.Release(); err != nil || exists(p.RecordPath()) {
		t.Fatalf("release: %v", err)
	}
	if _, err := Guard(p, false); err != nil {
		t.Fatalf("guard after release: %v", err)
	}
}

func TestGuardRefusesAnExistingProjectWithoutReplace(t *testing.T) {
	p := paths(t)
	write(t, filepath.Join(p.Target(), "keep.txt"), "original")
	if _, err := Guard(p, false); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("got %v", err)
	}
	if exists(p.RecordPath()) || read(filepath.Join(p.Target(), "keep.txt")) != "original" {
		t.Fatal("refusal left a guard or touched the project")
	}
}

func TestReplaceAndRelease(t *testing.T) {
	p := paths(t)
	write(t, filepath.Join(p.Target(), "keep.txt"), "original")
	r, err := Guard(p, true)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(p.AppsDir, ".devarch-backups", "demo-20261002-120000")
	if r.Backup != want || read(filepath.Join(want, "keep.txt")) != "original" || exists(p.Target()) {
		t.Fatalf("backup %s", r.Backup)
	}
	write(t, filepath.Join(p.Target(), "new.txt"), "new")
	if err := r.Release(); err != nil {
		t.Fatal(err)
	}
	if !exists(want) || exists(p.RecordPath()) {
		t.Fatal("release must keep the backup and drop the guard")
	}
	// A second replacement in the same second gets a distinct backup path.
	r2, err := Guard(p, true)
	if err != nil || r2.Backup != want+"-1" {
		t.Fatalf("second backup %v %v", r2, err)
	}
}

func TestRecoverRestoresEverything(t *testing.T) {
	p := paths(t)
	write(t, filepath.Join(p.Target(), "keep.txt"), "original")
	r, _ := Guard(p, true)
	r.AddDatabase(engine.DBResult{Engine: "mariadb", Name: "laravel_demo", User: "lv_demo", CreatedDB: true, CreatedUser: true})
	r.AddDatabase(engine.DBResult{Engine: "mariadb", Name: "wp_demo", User: "wp_demo", CreatedDB: true, CreatedUser: true, Dump: "/b/mariadb-wp_demo.sql", PrevUser: "CREATE USER x;"})
	write(t, filepath.Join(p.Target(), "partial.txt"), "half-built")

	loaded, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	db := &fakeDB{}
	if err := loaded.Recover(context.Background(), db, now); err != nil {
		t.Fatal(err)
	}
	if read(filepath.Join(p.Target(), "keep.txt")) != "original" || exists(filepath.Join(p.Target(), "partial.txt")) {
		t.Fatal("original not restored")
	}
	if read(filepath.Join(loaded.Failed, "partial.txt")) != "half-built" || !strings.Contains(loaded.Failed, ".devarch-failed/demo-") {
		t.Fatalf("partial project not quarantined: %s", loaded.Failed)
	}
	want := []string{
		"drop mariadb wp_demo/wp_demo",
		"restore mariadb wp_demo/wp_demo mariadb-wp_demo.sql CREATE USER x;",
		"drop mariadb laravel_demo/lv_demo",
	}
	if strings.Join(db.calls, "\n") != strings.Join(want, "\n") {
		t.Fatalf("db calls:\n%s", strings.Join(db.calls, "\n"))
	}
	if exists(p.RecordPath()) {
		t.Fatal("a complete recovery must release the guard")
	}
}

func TestRecoverOnlyDropsWhatTheRunCreated(t *testing.T) {
	p := paths(t)
	r, _ := Guard(p, false)
	r.AddDatabase(engine.DBResult{Engine: "mariadb", Name: "wp_demo", User: "wp_demo"}) // reused
	db := &fakeDB{}
	if err := r.Recover(context.Background(), db, now); err != nil {
		t.Fatal(err)
	}
	if want := "drop mariadb /"; len(db.calls) != 1 || db.calls[0] != want {
		t.Fatalf("calls %v", db.calls)
	}
}

func TestIncompleteRecoveryKeepsTheGuardAndResumes(t *testing.T) {
	for _, tc := range []struct {
		name    string
		failing map[string]int
		block   bool // occupy the target so the restore cannot happen
		problem string
	}{
		{name: "database cleanup", failing: map[string]int{"drop": 1}, problem: "mariadb database laravel_demo"},
		{name: "database restore", failing: map[string]int{"restore": 1}, problem: "injected restore failure"},
		{name: "occupied target", block: true, problem: "is occupied"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := paths(t)
			write(t, filepath.Join(p.Target(), "keep.txt"), "original")
			r, _ := Guard(p, true)
			r.AddDatabase(engine.DBResult{Engine: "mariadb", Name: "laravel_demo", User: "lv_demo", CreatedDB: true, CreatedUser: true, Dump: "/b/d.sql"})
			db := &fakeDB{failing: tc.failing}
			err := r.Recover(context.Background(), &blockingDB{fakeDB: db, target: p.Target(), block: tc.block}, now)
			if err == nil {
				t.Fatal("recovery reported success")
			}
			loaded, lerr := Load(p)
			if lerr != nil || loaded.State != Incomplete || !strings.Contains(strings.Join(loaded.Problems, "; "), tc.problem) {
				t.Fatalf("record %+v %v", loaded, lerr)
			}
			// Rerun after the cause is fixed: finished steps are not repeated.
			if tc.block {
				os.RemoveAll(p.Target())
			}
			db2 := &fakeDB{}
			if err := loaded.Recover(context.Background(), db2, now); err != nil {
				t.Fatalf("rerun: %v (%v)", err, loaded.Problems)
			}
			if tc.block && len(db2.calls) != 0 {
				t.Fatalf("rerun repeated database work: %v", db2.calls)
			}
			if read(filepath.Join(p.Target(), "keep.txt")) != "original" || exists(p.RecordPath()) {
				t.Fatal("rerun did not finish the recovery")
			}
		})
	}
}

// blockingDB recreates the target during database cleanup when block is set,
// as a container bind mount can.
type blockingDB struct {
	*fakeDB
	target string
	block  bool
}

func (b *blockingDB) DropDB(ctx context.Context, engine, name, user string) error {
	if b.block {
		os.MkdirAll(b.target, 0o755)
	}
	return b.fakeDB.DropDB(ctx, engine, name, user)
}

func TestCrashBeforeTheMoveLeavesTheOriginalAlone(t *testing.T) {
	p := paths(t)
	write(t, filepath.Join(p.Target(), "keep.txt"), "original")
	r := &Record{App: "demo", State: InProgress, Target: p.Target(), Stamp: "x",
		Backup: filepath.Join(p.AppsDir, ".devarch-backups", "demo-x"), path: p.RecordPath()}
	os.MkdirAll(filepath.Dir(r.path), 0o755)
	r.save()
	if err := r.Recover(context.Background(), &fakeDB{}, now); err != nil {
		t.Fatal(err)
	}
	if read(filepath.Join(p.Target(), "keep.txt")) != "original" || r.Failed != "" {
		t.Fatalf("original moved: failed=%s", r.Failed)
	}
}

func TestLegacyMarkerAndValidation(t *testing.T) {
	p := paths(t)
	write(t, p.RecordPath(), "provisioning in progress\n")
	if _, err := Load(p); err == nil || !strings.Contains(err.Error(), "older bootstrap") {
		t.Fatalf("legacy marker: %v", err)
	}
	if _, err := Guard(p, false); err == nil {
		t.Fatal("guard ignored a legacy marker")
	}
	if _, err := Guard(Paths{AppsDir: p.AppsDir, App: "../etc"}, false); err == nil {
		t.Fatal("accepted a path as an app name")
	}
	if _, err := Load(Paths{AppsDir: p.AppsDir, App: "none"}); err == nil || !strings.Contains(err.Error(), "not guarded") {
		t.Fatalf("missing guard: %v", err)
	}
}
