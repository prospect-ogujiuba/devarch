// Package replace keeps project provisioning recoverable. Before a bootstrap
// touches anything it takes a guard: a record at
// <apps>/.devarch-recovery/<app> that also blocks a second run. Replacing an
// existing project moves it to <apps>/.devarch-backups/<app>-<stamp>, and
// every database created for the project is added to the record. On success
// the guard is released; on failure Recover quarantines the partial project
// in <apps>/.devarch-failed, drops the new databases, restores replaced ones,
// and moves the old project back. Each completed recovery step is saved, so
// a recovery that stops part-way can be rerun.
package replace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
)

const (
	recoveryDir = ".devarch-recovery"
	backupDir   = ".devarch-backups"
	failedDir   = ".devarch-failed"
)

// States of a record.
const (
	InProgress = "in progress"
	Incomplete = "recovery incomplete"
)

// Database is a database created for the project, and whether recovery has
// undone it.
type Database struct {
	engine.DBResult
	Undone bool `json:"undone,omitempty"`
}

// Record is the guard file's content.
type Record struct {
	App    string `json:"app"`
	State  string `json:"state"`
	Target string `json:"target"`
	Stamp  string `json:"stamp"`
	// Backup is where the replaced project was moved; Restored is set once it
	// is back in place.
	Backup    string     `json:"backup,omitempty"`
	Restored  bool       `json:"restored,omitempty"`
	Failed    string     `json:"failed,omitempty"`
	Databases []Database `json:"databases,omitempty"`
	// Problems lists what the last recovery could not do.
	Problems []string `json:"problems,omitempty"`

	path string
}

// DB is what recovery needs from the database layer.
type DB interface {
	DropDB(ctx context.Context, engine, name, user string) error
	RestoreDB(ctx context.Context, engine, name, user, dump, prevUser string) error
}

var appRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// Paths for one app.
type Paths struct {
	AppsDir string
	App     string
	// Now is replaceable in tests.
	Now func() time.Time
}

func (p Paths) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

func (p Paths) validate() error {
	if !appRE.MatchString(p.App) {
		return fmt.Errorf("app name %q must be a lowercase DNS label", p.App)
	}
	return nil
}

// RecordPath is the guard file.
func (p Paths) RecordPath() string { return filepath.Join(p.AppsDir, recoveryDir, p.App) }

// Target is the project directory.
func (p Paths) Target() string { return filepath.Join(p.AppsDir, p.App) }

// DumpDir is where replaced databases of this run are saved.
func (r *Record) DumpDir() string {
	return filepath.Join(filepath.Dir(filepath.Dir(r.path)), backupDir, r.App+"-"+r.Stamp+"-databases")
}

// Guard takes the guard for an app. An existing project is refused unless
// replace is set, in which case it is moved to the backups directory.
func Guard(p Paths, replace bool) (*Record, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	path := p.RecordPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return nil, fmt.Errorf("%s is guarded by %s from an earlier run; run `devarch app recover %s` (or inspect and delete the file)", p.App, path, p.App)
	}
	if err != nil {
		return nil, err
	}
	f.Close()
	r := &Record{App: p.App, State: InProgress, Target: p.Target(), Stamp: p.now().Format("20060102-150405"), path: path}
	if _, err := os.Lstat(r.Target); err == nil {
		if !replace {
			os.Remove(path)
			return nil, fmt.Errorf("%s already exists; use --force to back it up and replace it", r.Target)
		}
		r.Backup = unique(filepath.Join(p.AppsDir, backupDir, p.App+"-"+r.Stamp))
		if err := os.MkdirAll(filepath.Dir(r.Backup), 0o755); err != nil {
			os.Remove(path)
			return nil, err
		}
		// Record the destination first: a crash after the move is recoverable.
		if err := r.save(); err != nil {
			os.Remove(path)
			return nil, err
		}
		if err := os.Rename(r.Target, r.Backup); err != nil {
			os.Remove(path)
			return nil, fmt.Errorf("move %s to %s: %w", r.Target, r.Backup, err)
		}
	}
	if err := r.save(); err != nil {
		return nil, err
	}
	return r, nil
}

// Load reads an app's guard.
func Load(p Paths) (*Record, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	path := p.RecordPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%s is not guarded: no %s", p.App, path)
		}
		return nil, err
	}
	r := &Record{path: path}
	if err := json.Unmarshal(data, r); err != nil {
		return nil, fmt.Errorf("%s was written by an older bootstrap (%q); inspect %s and %s by hand, then delete the file",
			path, firstLine(data), filepath.Join(p.AppsDir, backupDir), filepath.Join(p.AppsDir, failedDir))
	}
	return r, nil
}

func firstLine(data []byte) string {
	for i, b := range data {
		if b == '\n' {
			return string(data[:i])
		}
	}
	return string(data)
}

// AddDatabase records a database created for the project.
func (r *Record) AddDatabase(res engine.DBResult) error {
	r.Databases = append(r.Databases, Database{DBResult: res})
	return r.save()
}

// Release ends a successful run: the guard is removed and the backup kept.
func (r *Record) Release() error {
	return os.Remove(r.path)
}

// Recover undoes the run. It returns an error listing what it could not do;
// the guard then stays, marked incomplete, and Recover can be run again.
func (r *Record) Recover(ctx context.Context, db DB, now time.Time) error {
	r.Problems = nil
	problem := func(format string, args ...any) { r.Problems = append(r.Problems, fmt.Sprintf(format, args...)) }

	// A guard can stop between recording the backup path and moving the
	// project there; then the target is still the untouched original.
	if r.Backup != "" && !r.Restored {
		if _, err := os.Lstat(r.Backup); errors.Is(err, fs.ErrNotExist) {
			r.Restored = true
		}
	}
	// 1. Quarantine whatever the failed run left at the target, unless that
	//    is the restored original.
	if !r.Restored {
		if _, err := os.Lstat(r.Target); err == nil {
			dest := unique(filepath.Join(filepath.Dir(r.Target), failedDir, r.App+"-"+now.Format("20060102-150405")))
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				problem("quarantine %s: %v", r.Target, err)
			} else if err := os.Rename(r.Target, dest); err != nil {
				problem("quarantine %s: %v", r.Target, err)
			} else {
				r.Failed = dest
			}
		}
	}
	// 2. Undo databases, newest first.
	for i := len(r.Databases) - 1; i >= 0; i-- {
		d := &r.Databases[i]
		if d.Undone {
			continue
		}
		var name, user string
		if d.CreatedDB || d.Dump != "" {
			name = d.Name
		}
		if d.CreatedUser || d.PrevUser != "" {
			user = d.User
		}
		err := db.DropDB(ctx, d.Engine, name, user)
		if err == nil && (d.Dump != "" || d.PrevUser != "") {
			err = db.RestoreDB(ctx, d.Engine, d.Name, d.User, d.Dump, d.PrevUser)
		}
		if err != nil {
			problem("%s database %s: %v", d.Engine, d.Name, err)
			continue
		}
		d.Undone = true
		r.save()
	}
	// 3. Put the original project back.
	if r.Backup != "" && !r.Restored {
		if _, err := os.Lstat(r.Target); err == nil {
			problem("cannot restore %s: %s is occupied", r.Backup, r.Target)
		} else if err := os.Rename(r.Backup, r.Target); err != nil {
			problem("restore %s: %v", r.Backup, err)
		} else {
			r.Restored = true
		}
	}
	if len(r.Problems) > 0 {
		r.State = Incomplete
		if err := r.save(); err != nil {
			problem("update %s: %v", r.path, err)
		}
		return fmt.Errorf("recovery of %s is incomplete; the guard stays at %s", r.App, r.path)
	}
	return os.Remove(r.path)
}

// save writes the record atomically.
func (r *Record) save() error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}

// unique returns base, or base-1, base-2, … when it exists.
func unique(base string) string {
	candidate := base
	for i := 1; ; i++ {
		if _, err := os.Lstat(candidate); errors.Is(err, fs.ErrNotExist) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
}

// BackupPath returns a fresh path for a copy of the app in the backups directory.
func BackupPath(p Paths) string {
	return unique(filepath.Join(p.AppsDir, backupDir, p.App+"-"+p.now().Format("20060102-150405")))
}
