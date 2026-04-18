package runtime_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/events"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

func TestStreamLogsWithEventsPublishesLifecycleAndChunks(t *testing.T) {
	bus := events.NewBus()
	bus.SetNow(func() time.Time { return time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC) })
	stream, cancel := bus.Subscribe(8)
	defer cancel()

	adapter := &fakeAdapter{
		logChunks: []runtimepkg.LogChunk{
			{Line: "first", Stream: "combined"},
			{Line: "second", Stream: "combined"},
		},
	}
	ref := runtimepkg.ResourceRef{Workspace: "shop-local", Key: "api", RuntimeName: "devarch-shop-local-api"}
	if err := runtimepkg.StreamLogsWithEvents(context.Background(), adapter, bus, ref, runtimepkg.LogsRequest{Tail: 20}); err != nil {
		t.Fatalf("StreamLogsWithEvents returned error: %v", err)
	}

	var kinds []events.Kind
	for i := 0; i < 4; i++ {
		envelope := <-stream
		kinds = append(kinds, envelope.Kind)
	}
	if want := []events.Kind{events.KindLogsStarted, events.KindLogsChunk, events.KindLogsChunk, events.KindLogsCompleted}; !reflect.DeepEqual(kinds, want) {
		t.Fatalf("event kinds = %v, want %v", kinds, want)
	}
}

func TestExecWithEventsPublishesLifecycleAndReturnsResult(t *testing.T) {
	bus := events.NewBus()
	bus.SetNow(func() time.Time { return time.Date(2026, 4, 17, 12, 30, 0, 0, time.UTC) })
	stream, cancel := bus.Subscribe(4)
	defer cancel()

	adapter := &fakeAdapter{execResult: &runtimepkg.ExecResult{ExitCode: 0, Stdout: "ok\n"}}
	ref := runtimepkg.ResourceRef{Workspace: "shop-local", Key: "api", RuntimeName: "devarch-shop-local-api"}
	result, err := runtimepkg.ExecWithEvents(context.Background(), adapter, bus, ref, runtimepkg.ExecRequest{Command: []string{"php", "artisan", "about"}})
	if err != nil {
		t.Fatalf("ExecWithEvents returned error: %v", err)
	}
	if got, want := result.ExitCode, 0; got != want {
		t.Fatalf("result.ExitCode = %d, want %d", got, want)
	}

	started := <-stream
	completed := <-stream
	if got, want := []events.Kind{started.Kind, completed.Kind}, []events.Kind{events.KindExecStarted, events.KindExecCompleted}; !reflect.DeepEqual(got, want) {
		t.Fatalf("event kinds = %v, want %v", got, want)
	}
}

type fakeAdapter struct {
	logChunks  []runtimepkg.LogChunk
	execResult *runtimepkg.ExecResult
	execErr    error
}

func (f *fakeAdapter) Provider() string { return runtimepkg.ProviderDocker }

func (f *fakeAdapter) Capabilities() runtimepkg.AdapterCapabilities {
	return runtimepkg.AdapterCapabilities{Logs: true, Exec: true}
}

func (f *fakeAdapter) InspectWorkspace(context.Context, *runtimepkg.DesiredWorkspace) (*runtimepkg.Snapshot, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeAdapter) EnsureNetwork(context.Context, *runtimepkg.DesiredNetwork) error {
	return errors.New("not implemented")
}

func (f *fakeAdapter) RemoveNetwork(context.Context, *runtimepkg.DesiredNetwork) error {
	return errors.New("not implemented")
}

func (f *fakeAdapter) ApplyResource(context.Context, runtimepkg.ApplyResourceRequest) error {
	return errors.New("not implemented")
}

func (f *fakeAdapter) RemoveResource(context.Context, runtimepkg.ResourceRef) error {
	return errors.New("not implemented")
}

func (f *fakeAdapter) RestartResource(context.Context, runtimepkg.ResourceRef) error {
	return errors.New("not implemented")
}

func (f *fakeAdapter) StreamLogs(_ context.Context, _ runtimepkg.ResourceRef, _ runtimepkg.LogsRequest, consume runtimepkg.LogsConsumer) error {
	for _, chunk := range f.logChunks {
		if err := consume(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeAdapter) Exec(context.Context, runtimepkg.ResourceRef, runtimepkg.ExecRequest) (*runtimepkg.ExecResult, error) {
	if f.execErr != nil {
		return nil, f.execErr
	}
	return f.execResult, nil
}
