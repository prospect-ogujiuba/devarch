package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLite(path string) (*SQLiteStore, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite cache: empty path")
	}
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("sqlite cache mkdir: %w", err)
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	store := &SQLiteStore{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) init() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite cache: nil store")
	}
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS snapshots (
  workspace TEXT PRIMARY KEY,
  captured_at TEXT NOT NULL,
  snapshot_json TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS apply_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workspace TEXT NOT NULL,
  provider TEXT,
  started_at TEXT NOT NULL,
  finished_at TEXT NOT NULL,
  succeeded INTEGER NOT NULL,
  operations_json TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_apply_history_workspace_id ON apply_history(workspace, id DESC);
`)
	if err != nil {
		return fmt.Errorf("sqlite cache init: %w", err)
	}
	return nil
}

func (s *SQLiteStore) SaveSnapshot(ctx context.Context, record SnapshotRecord) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite cache: nil store")
	}
	payload, err := json.Marshal(record.Snapshot)
	if err != nil {
		return fmt.Errorf("sqlite cache marshal snapshot: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO snapshots(workspace, captured_at, snapshot_json)
VALUES(?, ?, ?)
ON CONFLICT(workspace) DO UPDATE SET captured_at=excluded.captured_at, snapshot_json=excluded.snapshot_json
`, record.Workspace, record.CapturedAt.Format(time.RFC3339Nano), string(payload))
	if err != nil {
		return fmt.Errorf("sqlite cache save snapshot: %w", err)
	}
	return nil
}

func (s *SQLiteStore) LatestSnapshot(ctx context.Context, workspace string) (*SnapshotRecord, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("sqlite cache: nil store")
	}
	row := s.db.QueryRowContext(ctx, `SELECT captured_at, snapshot_json FROM snapshots WHERE workspace = ?`, workspace)
	var capturedAt string
	var payload string
	if err := row.Scan(&capturedAt, &payload); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("sqlite cache latest snapshot: %w", err)
	}
	parsedTime, err := time.Parse(time.RFC3339Nano, capturedAt)
	if err != nil {
		return nil, fmt.Errorf("sqlite cache parse snapshot time: %w", err)
	}
	var snapshot runtimepkg.Snapshot
	if err := json.Unmarshal([]byte(payload), &snapshot); err != nil {
		return nil, fmt.Errorf("sqlite cache decode snapshot: %w", err)
	}
	return &SnapshotRecord{Workspace: workspace, CapturedAt: parsedTime, Snapshot: &snapshot}, nil
}

func (s *SQLiteStore) SaveApply(ctx context.Context, record ApplyRecord) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite cache: nil store")
	}
	payload, err := json.Marshal(record.Operations)
	if err != nil {
		return fmt.Errorf("sqlite cache marshal apply: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO apply_history(workspace, provider, started_at, finished_at, succeeded, operations_json)
VALUES(?, ?, ?, ?, ?, ?)
`, record.Workspace, record.Provider, record.StartedAt.Format(time.RFC3339Nano), record.FinishedAt.Format(time.RFC3339Nano), boolToInt(record.Succeeded), string(payload))
	if err != nil {
		return fmt.Errorf("sqlite cache save apply: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ApplyHistory(ctx context.Context, workspace string, limit int) ([]ApplyRecord, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("sqlite cache: nil store")
	}
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT provider, started_at, finished_at, succeeded, operations_json
FROM apply_history
WHERE workspace = ?
ORDER BY id DESC
LIMIT ?
`, workspace, limit)
	if err != nil {
		return nil, fmt.Errorf("sqlite cache query history: %w", err)
	}
	defer rows.Close()

	var history []ApplyRecord
	for rows.Next() {
		var record ApplyRecord
		var startedAt string
		var finishedAt string
		var succeeded int
		var operations string
		record.Workspace = workspace
		if err := rows.Scan(&record.Provider, &startedAt, &finishedAt, &succeeded, &operations); err != nil {
			return nil, fmt.Errorf("sqlite cache scan history: %w", err)
		}
		parsedStarted, err := time.Parse(time.RFC3339Nano, startedAt)
		if err != nil {
			return nil, fmt.Errorf("sqlite cache parse started_at: %w", err)
		}
		parsedFinished, err := time.Parse(time.RFC3339Nano, finishedAt)
		if err != nil {
			return nil, fmt.Errorf("sqlite cache parse finished_at: %w", err)
		}
		record.StartedAt = parsedStarted
		record.FinishedAt = parsedFinished
		record.Succeeded = succeeded == 1
		if err := json.Unmarshal([]byte(operations), &record.Operations); err != nil {
			return nil, fmt.Errorf("sqlite cache decode operations: %w", err)
		}
		history = append(history, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite cache iterate history: %w", err)
	}
	if len(history) == 0 {
		return nil, nil
	}
	return history, nil
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
