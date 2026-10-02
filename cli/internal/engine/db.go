package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// A dbServer is one of the shared database services apps get their own
// database and user on. SQL always travels on stdin and the admin password
// expands inside the container, so no secret appears in an argument list.
type dbServer struct {
	service          string // catalog ID
	port             int
	maxName, maxUser int
	// admin runs SQL from stdin; query prints rows without headers.
	admin, query string
	// dump prints a database as SQL; load reads one into an existing database.
	dump, load func(db string) string

	dbExists, userExists func(name string) string
	createDB             func(db, user string) string
	createUser           func(user, password string) string
	resetUser            func(user, password string) string
	grant                func(db, user string) string
	dropDB, dropUser     func(name string) string
	// userDefinition returns SQL that, run through query, prints statements
	// recreating the user with its current password hash and grants.
	userDefinition func(user string) string
}

const mariadbRoot = `-uroot -p"$MARIADB_ROOT_PASSWORD"`

var dbServers = map[string]dbServer{
	"mariadb": {
		service: "database/mariadb", port: 3306, maxName: 64, maxUser: 32,
		admin: "mariadb " + mariadbRoot,
		query: "mariadb -N -B " + mariadbRoot,
		dump: func(db string) string {
			return "mariadb-dump " + mariadbRoot + " --single-transaction --routines --triggers --events " + db
		},
		load: func(db string) string { return "mariadb " + mariadbRoot + " " + db },
		dbExists: func(n string) string {
			return "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME='" + n + "';"
		},
		userExists: func(n string) string { return "SELECT User FROM mysql.user WHERE User='" + n + "' LIMIT 1;" },
		createDB: func(db, _ string) string {
			return "CREATE DATABASE `" + db + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
		},
		createUser: func(u, p string) string { return "CREATE USER '" + u + "'@'%' IDENTIFIED BY '" + p + "';" },
		resetUser:  func(u, p string) string { return "ALTER USER '" + u + "'@'%' IDENTIFIED BY '" + p + "';" },
		grant: func(db, u string) string {
			return "GRANT ALL PRIVILEGES ON `" + db + "`.* TO '" + u + "'@'%'; FLUSH PRIVILEGES;"
		},
		dropDB:         func(n string) string { return "DROP DATABASE IF EXISTS `" + n + "`;" },
		dropUser:       func(n string) string { return "DROP USER IF EXISTS '" + n + "'@'%';" },
		userDefinition: func(u string) string { return "SHOW CREATE USER '" + u + "'@'%'; SHOW GRANTS FOR '" + u + "'@'%';" },
	},
	"postgres": {
		service: "database/postgres", port: 5432, maxName: 63, maxUser: 63,
		admin: `psql -v ON_ERROR_STOP=1 -X -q -U "$POSTGRES_USER" -d postgres`,
		query: `psql -v ON_ERROR_STOP=1 -X -q -t -A -U "$POSTGRES_USER" -d postgres`,
		dump:  func(db string) string { return `pg_dump -U "$POSTGRES_USER" ` + db },
		load:  func(db string) string { return `psql -v ON_ERROR_STOP=1 -X -q -U "$POSTGRES_USER" -d ` + db },
		dbExists: func(n string) string {
			return "SELECT datname FROM pg_database WHERE datname='" + n + "';"
		},
		userExists: func(n string) string { return "SELECT rolname FROM pg_roles WHERE rolname='" + n + "';" },
		createDB:   func(db, u string) string { return `CREATE DATABASE "` + db + `" OWNER "` + u + `";` },
		createUser: func(u, p string) string { return `CREATE ROLE "` + u + `" LOGIN PASSWORD '` + p + `';` },
		resetUser:  func(u, p string) string { return `ALTER ROLE "` + u + `" LOGIN PASSWORD '` + p + `';` },
		grant:      func(db, u string) string { return `GRANT ALL PRIVILEGES ON DATABASE "` + db + `" TO "` + u + `";` },
		dropDB:     func(n string) string { return `DROP DATABASE IF EXISTS "` + n + `" WITH (FORCE);` },
		dropUser:   func(n string) string { return `DROP ROLE IF EXISTS "` + n + `";` },
		userDefinition: func(u string) string {
			return `SELECT format('CREATE ROLE %I LOGIN PASSWORD %L;', rolname, rolpassword) FROM pg_authid WHERE rolname='` + u + `';`
		},
	},
}

// DBEngines lists the supported database engines.
var DBEngines = []string{"mariadb", "postgres"}

// What CreateDB does when the database or user already exists.
const (
	ExistingFail    = "fail"    // refuse (the default)
	ExistingReuse   = "reuse"   // keep the database, reset the user's password
	ExistingReplace = "replace" // dump and drop both, then create fresh
)

// DBSpec describes a database to create.
type DBSpec struct {
	Engine string
	Name   string
	// User defaults to Name, shortened to the engine's limit.
	User     string
	Existing string
	// BackupDir receives the dump of a replaced database.
	BackupDir string
}

// DBResult reports what CreateDB did, with the connection details an app
// container uses.
type DBResult struct {
	Engine   string `json:"engine"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	// CreatedDB and CreatedUser are false when reuse kept existing ones.
	CreatedDB   bool `json:"created_database"`
	CreatedUser bool `json:"created_user"`
	// Dump holds the replaced database; PrevUser recreates the replaced user.
	Dump     string `json:"dump,omitempty"`
	PrevUser string `json:"previous_user,omitempty"`
}

var dbIdentRE = regexp.MustCompile(`^[a-z0-9_]+$`)

func (e *Engine) dbServer(engine string) (dbServer, string, error) {
	srv, ok := dbServers[engine]
	if !ok {
		return srv, "", fmt.Errorf("database engine must be %s, not %q", strings.Join(DBEngines, " or "), engine)
	}
	container := engine
	if svc, ok := e.Catalog.Get(srv.service); ok && svc.Primary().ContainerName != "" {
		container = svc.Primary().ContainerName
	}
	return srv, container, nil
}

func (spec *DBSpec) validate(srv dbServer) error {
	if spec.User == "" {
		spec.User = spec.Name[:min(len(spec.Name), srv.maxUser)]
	}
	if spec.Existing == "" {
		spec.Existing = ExistingFail
	}
	switch {
	case !dbIdentRE.MatchString(spec.Name) || len(spec.Name) > srv.maxName:
		return fmt.Errorf("database name %q must be lowercase letters, digits and underscores, at most %d characters", spec.Name, srv.maxName)
	case !dbIdentRE.MatchString(spec.User) || len(spec.User) > srv.maxUser:
		return fmt.Errorf("user name %q must be lowercase letters, digits and underscores, at most %d characters", spec.User, srv.maxUser)
	}
	switch spec.Existing {
	case ExistingFail, ExistingReuse, ExistingReplace:
	default:
		return fmt.Errorf("existing must be fail, reuse or replace, not %q", spec.Existing)
	}
	return nil
}

// sql runs statements on the server. Output goes to the log so it never
// mixes with what a command prints for scripts.
func (e *Engine) sql(ctx context.Context, container, shell, stmts string) error {
	c := e.Cmd("exec", "-i", container, "sh", "-c", shell)
	c.Stdin = strings.NewReader(stmts + "\n")
	c.Stdout = e.Log
	return e.Runner.Run(ctx, c)
}

func (e *Engine) sqlQuery(ctx context.Context, container, shell, stmts string) (string, error) {
	c := e.Cmd("exec", "-i", container, "sh", "-c", shell)
	c.Stdin = strings.NewReader(stmts + "\n")
	out, err := e.Runner.Output(ctx, c)
	return strings.TrimSpace(string(out)), err
}

// DBExists reports whether the database and the user exist.
func (e *Engine) DBExists(ctx context.Context, engine, name, user string) (db, usr bool, err error) {
	srv, container, err := e.dbServer(engine)
	if err != nil {
		return false, false, err
	}
	out, err := e.sqlQuery(ctx, container, srv.query, srv.dbExists(name))
	if err != nil {
		return false, false, fmt.Errorf("cannot query %s; is it running? (devarch up %s): %w", container, engine, err)
	}
	db = out != ""
	if user != "" {
		if out, err = e.sqlQuery(ctx, container, srv.query, srv.userExists(user)); err != nil {
			return db, false, err
		}
		usr = out != ""
	}
	return db, usr, nil
}

// CreateDB creates a database and a user that owns it, with a generated
// password. On failure it undoes its own changes, including restoring a
// database and user it replaced.
func (e *Engine) CreateDB(ctx context.Context, spec DBSpec) (DBResult, error) {
	srv, container, err := e.dbServer(spec.Engine)
	if err != nil {
		return DBResult{}, err
	}
	if err := spec.validate(srv); err != nil {
		return DBResult{}, err
	}
	res := DBResult{Engine: spec.Engine, Host: container, Port: srv.port, Name: spec.Name, User: spec.User}
	dbFound, userFound, err := e.DBExists(ctx, spec.Engine, spec.Name, spec.User)
	if err != nil {
		if !e.DryRun {
			return res, err
		}
		e.logf("cannot check %s (%v); assuming nothing exists", container, err)
	}
	if spec.Existing == ExistingFail && (dbFound || userFound) {
		what := "database " + spec.Name
		if !dbFound {
			what = "user " + spec.User
		}
		return res, fmt.Errorf("%s already exists on %s; pass --existing reuse or --existing replace", what, spec.Engine)
	}
	if res.Password, err = generatePassword(); err != nil {
		return res, err
	}

	fail := func(err error) (DBResult, error) {
		if uerr := e.undoCreate(ctx, res); uerr != nil {
			return res, fmt.Errorf("%w; undoing it also failed: %v", err, uerr)
		}
		return res, err
	}
	if spec.Existing == ExistingReplace {
		if dbFound {
			if res.Dump, err = e.dumpDB(ctx, spec.Engine, spec.Name, spec.BackupDir); err != nil {
				return res, err
			}
		}
		if userFound {
			def, err := e.sqlQuery(ctx, container, srv.query, srv.userDefinition(spec.User))
			if err != nil {
				return res, fmt.Errorf("cannot save the definition of user %s: %w", spec.User, err)
			}
			res.PrevUser = statements(def)

		}
		e.logf("replacing %s database %s and user %s (previous database saved to %s)", spec.Engine, spec.Name, spec.User, orNone(res.Dump))
		if err := e.sql(ctx, container, srv.admin, srv.dropDB(spec.Name)+"\n"+srv.dropUser(spec.User)); err != nil {
			return fail(err)
		}
		dbFound, userFound = false, false
	}

	e.logf("creating %s database %s owned by %s", spec.Engine, spec.Name, spec.User)
	if userFound {
		if err := e.sql(ctx, container, srv.admin, srv.resetUser(spec.User, res.Password)); err != nil {
			return fail(err)
		}
	} else {
		if err := e.sql(ctx, container, srv.admin, srv.createUser(spec.User, res.Password)); err != nil {
			return fail(err)
		}
		res.CreatedUser = true
	}
	if !dbFound {
		if err := e.sql(ctx, container, srv.admin, srv.createDB(spec.Name, spec.User)); err != nil {
			return fail(err)
		}
		res.CreatedDB = true
	}
	if err := e.sql(ctx, container, srv.admin, srv.grant(spec.Name, spec.User)); err != nil {
		return fail(err)
	}
	return res, nil
}

func orNone(s string) string {
	if s == "" {
		return "nowhere: it did not exist"
	}
	return s
}

// undoCreate removes what CreateDB created and restores what it replaced.
func (e *Engine) undoCreate(ctx context.Context, res DBResult) error {
	var errs []error
	if res.CreatedDB || res.Dump != "" {
		errs = append(errs, e.DropDB(ctx, res.Engine, res.Name, ""))
	}
	if res.CreatedUser || res.PrevUser != "" {
		errs = append(errs, e.DropDB(ctx, res.Engine, "", res.User))
	}
	if res.PrevUser != "" || res.Dump != "" {
		errs = append(errs, e.RestoreDB(ctx, res.Engine, res.Name, res.User, res.Dump, res.PrevUser))
	}
	return errors.Join(errs...)
}

// DropDB drops a database and/or a user; an empty name skips it.
func (e *Engine) DropDB(ctx context.Context, engine, name, user string) error {
	srv, container, err := e.dbServer(engine)
	if err != nil {
		return err
	}
	var stmts []string
	if name != "" {
		if !dbIdentRE.MatchString(name) {
			return fmt.Errorf("invalid database name %q", name)
		}
		stmts = append(stmts, srv.dropDB(name))
	}
	if user != "" {
		if !dbIdentRE.MatchString(user) {
			return fmt.Errorf("invalid user name %q", user)
		}
		stmts = append(stmts, srv.dropUser(user))
	}
	if len(stmts) == 0 {
		return nil
	}
	e.logf("dropping %s %s", engine, strings.TrimSpace(name+" "+user))
	return e.sql(ctx, container, srv.admin, strings.Join(stmts, "\n"))
}

// dumpDB writes a database to <dir>/<engine>-<name>.sql (mode 0600).
func (e *Engine) dumpDB(ctx context.Context, engine, name, dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("replacing database %s needs a backup location for its dump", name)
	}
	srv, container, _ := e.dbServer(engine)
	path := filepath.Join(dir, engine+"-"+name+".sql")
	if e.DryRun {
		e.logf("would save database %s to %s", name, path)
		return path, nil
	}
	e.logf("saving database %s to %s", name, path)
	out, err := e.Runner.Output(ctx, e.Cmd("exec", container, "sh", "-c", srv.dump(name)))
	if err != nil {
		return "", fmt.Errorf("cannot dump database %s: %w", name, err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// RestoreDB recreates a replaced user from its saved definition and a
// database from its dump. Either may be empty.
func (e *Engine) RestoreDB(ctx context.Context, engine, name, user, dump, prevUser string) error {
	srv, container, err := e.dbServer(engine)
	if err != nil {
		return err
	}
	owner := user
	if prevUser != "" {
		e.logf("restoring %s user %s", engine, user)
		if err := e.sql(ctx, container, srv.admin, prevUser); err != nil {
			return fmt.Errorf("restore user %s: %w", user, err)
		}
	} else if engine == "postgres" {
		owner = "postgres" // the replaced database had no saved owner role
	}
	if dump == "" {
		return nil
	}
	e.logf("restoring %s database %s from %s", engine, name, dump)
	if err := e.sql(ctx, container, srv.admin, srv.createDB(name, owner)); err != nil {
		return fmt.Errorf("recreate database %s: %w", name, err)
	}
	f, err := os.Open(dump)
	if err != nil {
		return err
	}
	defer f.Close()
	c := e.Cmd("exec", "-i", container, "sh", "-c", srv.load(name))
	c.Stdin, c.Stdout = f, e.Log
	if err := e.Runner.Run(ctx, c); err != nil {
		return fmt.Errorf("load %s into %s: %w", dump, name, err)
	}
	return nil
}

// statements ends every line of SHOW output with a semicolon so the lines
// replay as separate statements.
func statements(out string) string {
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, strings.TrimSuffix(l, ";")+";")
		}
	}
	return strings.Join(lines, "\n")
}

func generatePassword() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Env renders a result as NAME=value lines for scripts.
func (r DBResult) Env() []string {
	return []string{
		"DB_ENGINE=" + r.Engine,
		"DB_HOST=" + r.Host,
		fmt.Sprintf("DB_PORT=%d", r.Port),
		"DB_NAME=" + r.Name,
		"DB_USER=" + r.User,
		"DB_PASSWORD=" + r.Password,
	}
}
