package cache

import (
	"context"
	"time"

	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

type Store interface {
	SaveSnapshot(ctx context.Context, record SnapshotRecord) error
	LatestSnapshot(ctx context.Context, workspace string) (*SnapshotRecord, error)
	SaveApply(ctx context.Context, record ApplyRecord) error
	ApplyHistory(ctx context.Context, workspace string, limit int) ([]ApplyRecord, error)
	Close() error
}

type SnapshotRecord struct {
	Workspace  string               `json:"workspace"`
	CapturedAt time.Time            `json:"capturedAt"`
	Snapshot   *runtimepkg.Snapshot `json:"snapshot"`
}

type ApplyRecord struct {
	Workspace  string            `json:"workspace"`
	Provider   string            `json:"provider,omitempty"`
	StartedAt  time.Time         `json:"startedAt"`
	FinishedAt time.Time         `json:"finishedAt"`
	Succeeded  bool              `json:"succeeded"`
	Operations []OperationRecord `json:"operations,omitempty"`
}

type OperationRecord struct {
	Scope       string `json:"scope"`
	Target      string `json:"target"`
	RuntimeName string `json:"runtimeName,omitempty"`
	Kind        string `json:"kind"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
}

type NopStore struct{}

func Normalize(store Store) Store {
	if store == nil {
		return NopStore{}
	}
	return store
}

func (NopStore) SaveSnapshot(context.Context, SnapshotRecord) error { return nil }

func (NopStore) LatestSnapshot(context.Context, string) (*SnapshotRecord, error) { return nil, nil }

func (NopStore) SaveApply(context.Context, ApplyRecord) error { return nil }

func (NopStore) ApplyHistory(context.Context, string, int) ([]ApplyRecord, error) { return nil, nil }

func (NopStore) Close() error { return nil }
