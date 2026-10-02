package engine_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
	"github.com/prospect-ogujiuba/devarch/cli/internal/state"
	"github.com/prospect-ogujiuba/devarch/cli/internal/testutil"
)

func dbSetup(t *testing.T) (*engine.Engine, *runner.Fake) {
	t.Helper()
	lib := filepath.Join(t.TempDir(), "services-library")
	testutil.WriteFiles(t, lib, map[string]string{
		"database/mariadb/compose.yml":  testutil.Compose("mariadb", ""),
		"database/postgres/compose.yml": testutil.Compose("pg", ""),
	})
	c, _ := catalog.Load([]catalog.Path{{Dir: lib}})
	f := &runner.Fake{Responses: map[string]runner.Response{}, RunErrors: map[string][]error{}}
	return &engine.Engine{Catalog: c, Runner: f, Settings: state.Defaults("")}, f
}

// sqlSent returns the stdin of every command whose line contains substr.
func sqlSent(f *runner.Fake, substr string) string {
	var b strings.Builder
	for i, c := range f.Cmds {
		if strings.Contains(c.String(), substr) {
			b.WriteString(f.Stdin[i])
		}
	}
	return b.String()
}

const mariadbQuery = "podman exec -i mariadb sh -c 'mariadb -N -B"

func TestCreateDBMariaDB(t *testing.T) {
	e, f := dbSetup(t)
	res, err := e.CreateDB(context.Background(), engine.DBSpec{Engine: "mariadb", Name: "laravel_shop", User: "lv_shop"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Host != "mariadb" || res.Port != 3306 || len(res.Password) != 48 || !res.CreatedDB || !res.CreatedUser {
		t.Fatalf("result %+v", res)
	}
	for _, l := range f.Lines() {
		if strings.Contains(l, res.Password) {
			t.Fatalf("password in argv: %s", l)
		}
	}
	admin := sqlSent(f, `-uroot -p"$MARIADB_ROOT_PASSWORD"'`)
	for _, want := range []string{
		"CREATE USER 'lv_shop'@'%' IDENTIFIED BY '" + res.Password + "';",
		"CREATE DATABASE `laravel_shop` CHARACTER SET utf8mb4",
		"GRANT ALL PRIVILEGES ON `laravel_shop`.* TO 'lv_shop'@'%'",
	} {
		if !strings.Contains(admin, want) {
			t.Errorf("missing %q in\n%s", want, admin)
		}
	}
	if env := strings.Join(res.Env(), "\n"); !strings.Contains(env, "DB_HOST=mariadb\nDB_PORT=3306\nDB_NAME=laravel_shop\nDB_USER=lv_shop\nDB_PASSWORD=") {
		t.Fatal(env)
	}
}

func TestCreateDBRefusesExisting(t *testing.T) {
	e, f := dbSetup(t)
	f.Responses[mariadbQuery] = runner.Response{Out: []byte("shop\n")}
	_, err := e.CreateDB(context.Background(), engine.DBSpec{Engine: "mariadb", Name: "shop"})
	if err == nil || !strings.Contains(err.Error(), "database shop already exists") {
		t.Fatalf("got %v", err)
	}
	for _, c := range f.Cmds {
		if !strings.Contains(c.String(), "-N -B") {
			t.Fatalf("mutated: %s", c.String())
		}
	}
}

func TestCreateDBReuseResetsPassword(t *testing.T) {
	e, f := dbSetup(t)
	f.Responses[mariadbQuery] = runner.Response{Out: []byte("wp_blog\n")}
	res, err := e.CreateDB(context.Background(), engine.DBSpec{Engine: "mariadb", Name: "wp_blog", Existing: engine.ExistingReuse})
	if err != nil {
		t.Fatal(err)
	}
	if res.CreatedDB || res.CreatedUser {
		t.Fatalf("reuse created: %+v", res)
	}
	admin := sqlSent(f, `-uroot -p"$MARIADB_ROOT_PASSWORD"'`)
	if !strings.Contains(admin, "ALTER USER 'wp_blog'@'%'") || strings.Contains(admin, "CREATE DATABASE") || strings.Contains(admin, "DROP") {
		t.Fatalf("sql:\n%s", admin)
	}
}

func TestCreateDBReplaceSavesDatabaseAndUser(t *testing.T) {
	e, f := dbSetup(t)
	backup := filepath.Join(t.TempDir(), "b")
	f.Responses[mariadbQuery] = runner.Response{Out: []byte("CREATE USER `wp_blog`@`%` IDENTIFIED BY PASSWORD '*ABC'\n")}
	f.Responses["podman exec mariadb sh -c 'mariadb-dump"] = runner.Response{Out: []byte("-- dump of wp_blog\n")}
	res, err := e.CreateDB(context.Background(), engine.DBSpec{Engine: "mariadb", Name: "wp_blog", Existing: engine.ExistingReplace, BackupDir: backup})
	if err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(res.Dump); string(data) != "-- dump of wp_blog\n" {
		t.Fatalf("dump %s: %q", res.Dump, data)
	}
	if info, _ := os.Stat(res.Dump); info.Mode().Perm() != 0o600 {
		t.Fatalf("dump mode %v", info.Mode())
	}
	if !strings.Contains(res.PrevUser, "IDENTIFIED BY PASSWORD '*ABC'") {
		t.Fatalf("previous user %q", res.PrevUser)
	}
	admin := sqlSent(f, `-uroot -p"$MARIADB_ROOT_PASSWORD"'`)
	drop := strings.Index(admin, "DROP DATABASE IF EXISTS `wp_blog`")
	create := strings.Index(admin, "CREATE DATABASE `wp_blog`")
	if drop < 0 || create < drop || !res.CreatedDB || !res.CreatedUser {
		t.Fatalf("sql:\n%s", admin)
	}
}

func TestCreateDBReplaceNeedsBackupDir(t *testing.T) {
	e, f := dbSetup(t)
	f.Responses[mariadbQuery] = runner.Response{Out: []byte("x\n")}
	if _, err := e.CreateDB(context.Background(), engine.DBSpec{Engine: "mariadb", Name: "x", Existing: engine.ExistingReplace}); err == nil {
		t.Fatal("replaced without a backup location")
	}
}

func TestCreateDBUndoesPartialWork(t *testing.T) {
	e, f := dbSetup(t)
	f.RunErrors["podman exec -i mariadb sh -c 'mariadb -uroot"] = []error{nil, errors.New("disk full")}
	_, err := e.CreateDB(context.Background(), engine.DBSpec{Engine: "mariadb", Name: "shop"})
	if err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("got %v", err)
	}
	if admin := sqlSent(f, `-uroot -p"$MARIADB_ROOT_PASSWORD"'`); !strings.Contains(admin, "DROP USER IF EXISTS 'shop'@'%'") {
		t.Fatalf("created user was not dropped:\n%s", admin)
	}
}

func TestCreateDBPostgres(t *testing.T) {
	e, f := dbSetup(t)
	res, err := e.CreateDB(context.Background(), engine.DBSpec{Engine: "postgres", Name: "api"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Host != "pg" || res.Port != 5432 {
		t.Fatalf("result %+v", res)
	}
	admin := sqlSent(f, "-d postgres'")
	role, db := strings.Index(admin, `CREATE ROLE "api" LOGIN PASSWORD`), strings.Index(admin, `CREATE DATABASE "api" OWNER "api"`)
	if role < 0 || db < role {
		t.Fatalf("sql:\n%s", admin)
	}
}

func TestDBValidation(t *testing.T) {
	e, f := dbSetup(t)
	for _, spec := range []engine.DBSpec{
		{Engine: "oracle", Name: "x"},
		{Engine: "mariadb", Name: "Bad-Name"},
		{Engine: "mariadb", Name: "x'; DROP"},
		{Engine: "mariadb", Name: "x", User: strings.Repeat("u", 33)},
		{Engine: "mariadb", Name: "x", Existing: "maybe"},
	} {
		if _, err := e.CreateDB(context.Background(), spec); err == nil {
			t.Errorf("accepted %+v", spec)
		}
	}
	if len(f.Cmds) != 0 {
		t.Fatalf("ran %v", f.Lines())
	}
	if res, err := e.CreateDB(context.Background(), engine.DBSpec{Engine: "mariadb", Name: strings.Repeat("n", 40)}); err != nil || len(res.User) != 32 {
		t.Fatalf("default user %q %v", res.User, err)
	}
}

func TestDropAndRestoreDB(t *testing.T) {
	e, f := dbSetup(t)
	if err := e.DropDB(context.Background(), "mariadb", "shop", "shop"); err != nil {
		t.Fatal(err)
	}
	dump := filepath.Join(t.TempDir(), "d.sql")
	os.WriteFile(dump, []byte("INSERT 1;\n"), 0o600)
	if err := e.RestoreDB(context.Background(), "mariadb", "shop", "shop", dump, "CREATE USER `shop`@`%` IDENTIFIED BY PASSWORD '*A'"); err != nil {
		t.Fatal(err)
	}
	all := strings.Join(f.Stdin, "")
	for _, want := range []string{"DROP DATABASE IF EXISTS `shop`;\nDROP USER IF EXISTS 'shop'@'%';", "IDENTIFIED BY PASSWORD '*A'", "CREATE DATABASE `shop`", "INSERT 1;"} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %q in\n%s", want, all)
		}
	}
	if last := f.Lines()[len(f.Lines())-1]; !strings.HasSuffix(last, `mariadb -uroot -p"$MARIADB_ROOT_PASSWORD" shop'`) {
		t.Fatalf("load command %s", last)
	}
}
