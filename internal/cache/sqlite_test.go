package cache_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	cachepkg "github.com/prospect-ogujiuba/devarch/internal/cache"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

func TestSQLiteStorePersistsSnapshotsAndApplyHistory(t *testing.T) {
	store, err := cachepkg.NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("cache.NewSQLite returned error: %v", err)
	}
	defer store.Close()

	snapshotTime := time.Date(2026, 4, 17, 15, 0, 0, 0, time.UTC)
	snapshot := &runtimepkg.Snapshot{
		Workspace: runtimepkg.SnapshotWorkspace{Name: "shop-local", Provider: runtimepkg.ProviderDocker},
		Resources: []*runtimepkg.SnapshotResource{{Key: "api", RuntimeName: "devarch-shop-local-api", State: runtimepkg.ResourceState{Running: true, Status: "running"}}},
	}
	if err := store.SaveSnapshot(context.Background(), cachepkg.SnapshotRecord{Workspace: "shop-local", CapturedAt: snapshotTime, Snapshot: snapshot}); err != nil {
		t.Fatalf("SaveSnapshot returned error: %v", err)
	}
	loadedSnapshot, err := store.LatestSnapshot(context.Background(), "shop-local")
	if err != nil {
		t.Fatalf("LatestSnapshot returned error: %v", err)
	}
	if got, want := loadedSnapshot.CapturedAt, snapshotTime; !got.Equal(want) {
		t.Fatalf("CapturedAt = %v, want %v", got, want)
	}
	if got, want := loadedSnapshot.Snapshot.Resource("api").RuntimeName, "devarch-shop-local-api"; got != want {
		t.Fatalf("RuntimeName = %q, want %q", got, want)
	}

	firstRecord := cachepkg.ApplyRecord{
		Workspace:  "shop-local",
		Provider:   runtimepkg.ProviderDocker,
		StartedAt:  snapshotTime,
		FinishedAt: snapshotTime.Add(2 * time.Minute),
		Succeeded:  true,
		Operations: []cachepkg.OperationRecord{{Scope: "workspace", Target: "network", Kind: "add", Status: "success"}},
	}
	secondRecord := cachepkg.ApplyRecord{
		Workspace:  "shop-local",
		Provider:   runtimepkg.ProviderDocker,
		StartedAt:  snapshotTime.Add(5 * time.Minute),
		FinishedAt: snapshotTime.Add(7 * time.Minute),
		Succeeded:  false,
		Operations: []cachepkg.OperationRecord{{Scope: "resource", Target: "api", Kind: "modify", Status: "failed"}},
	}
	if err := store.SaveApply(context.Background(), firstRecord); err != nil {
		t.Fatalf("SaveApply(first) returned error: %v", err)
	}
	if err := store.SaveApply(context.Background(), secondRecord); err != nil {
		t.Fatalf("SaveApply(second) returned error: %v", err)
	}
	loadedHistory, err := store.ApplyHistory(context.Background(), "shop-local", 10)
	if err != nil {
		t.Fatalf("ApplyHistory returned error: %v", err)
	}
	if got, want := len(loadedHistory), 2; got != want {
		t.Fatalf("len(history) = %d, want %d", got, want)
	}
	if got, want := loadedHistory[0].Succeeded, false; got != want {
		t.Fatalf("history[0].Succeeded = %v, want %v", got, want)
	}
	if got, want := loadedHistory[1].Succeeded, true; got != want {
		t.Fatalf("history[1].Succeeded = %v, want %v", got, want)
	}
	if got, want := loadedHistory[0].Operations, secondRecord.Operations; !reflect.DeepEqual(got, want) {
		t.Fatalf("history[0].Operations = %#v, want %#v", got, want)
	}
}

func TestNormalizeReturnsNopStoreForNil(t *testing.T) {
	store := cachepkg.Normalize(nil)
	if err := store.SaveApply(context.Background(), cachepkg.ApplyRecord{}); err != nil {
		t.Fatalf("SaveApply on normalized nil store returned error: %v", err)
	}
	if err := store.SaveSnapshot(context.Background(), cachepkg.SnapshotRecord{}); err != nil {
		t.Fatalf("SaveSnapshot on normalized nil store returned error: %v", err)
	}
}
