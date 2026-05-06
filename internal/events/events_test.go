package events_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	stdruntime "runtime"
	"testing"
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/events"
)

func TestBusSequencingAndCodecGolden(t *testing.T) {
	bus := events.NewBus()
	current := time.Date(2026, 4, 17, 16, 0, 0, 0, time.UTC)
	bus.SetNow(func() time.Time {
		current = current.Add(time.Second)
		return current
	})

	stream, cancel := bus.Subscribe(16)
	defer cancel()

	publish(t, bus, events.ApplyStarted("shop-local", 5))
	publish(t, bus, events.ApplyProgress("shop-local", "api", "resource", "api", "devarch-shop-local-api", "add", "success", "resource is not present in runtime snapshot"))
	publish(t, bus, events.ApplyCompleted("shop-local", true, 5))
	publish(t, bus, events.LogsStarted("shop-local", "api", 20, false))
	logTimestamp := time.Date(2026, 4, 17, 16, 0, 10, 0, time.UTC)
	publish(t, bus, events.LogsChunk("shop-local", "api", "combined", "server ready", &logTimestamp))
	publish(t, bus, events.LogsCompleted("shop-local", "api", 20, false))
	publish(t, bus, events.ExecStarted("shop-local", "api", []string{"php", "artisan", "about"}))
	publish(t, bus, events.ExecCompleted("shop-local", "api", 0))

	captured := make([]events.Envelope, 0, 8)
	for i := 0; i < 8; i++ {
		captured = append(captured, <-stream)
	}
	for i := range captured {
		if got, want := captured[i].Sequence, uint64(i+1); got != want {
			t.Fatalf("envelope[%d].Sequence = %d, want %d", i, got, want)
		}
		encoded, err := json.Marshal(captured[i])
		if err != nil {
			t.Fatalf("json.Marshal(envelope) returned error: %v", err)
		}
		var decoded events.Envelope
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("json.Unmarshal(envelope) returned error: %v", err)
		}
		if got, want := decoded.Kind, captured[i].Kind; got != want {
			t.Fatalf("decoded.Kind = %q, want %q", got, want)
		}
	}

	var progress events.ApplyProgressPayload
	if err := json.Unmarshal(captured[1].Payload, &progress); err != nil {
		t.Fatalf("json.Unmarshal(ApplyProgressPayload) returned error: %v", err)
	}
	if got, want := progress.Target, "api"; got != want {
		t.Fatalf("progress.Target = %q, want %q", got, want)
	}
	var logsChunk events.LogsChunkPayload
	if err := json.Unmarshal(captured[4].Payload, &logsChunk); err != nil {
		t.Fatalf("json.Unmarshal(LogsChunkPayload) returned error: %v", err)
	}
	if got, want := logsChunk.Line, "server ready"; got != want {
		t.Fatalf("logs chunk line = %q, want %q", got, want)
	}

	actual := marshalJSON(t, captured)
	goldenPath := filepath.Join(repoRoot(t), "testdata", "goldens", "phase3", "runtime-events.golden.json")
	if updateGoldens() {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(goldenPath), err)
		}
		if err := os.WriteFile(goldenPath, actual, 0o644); err != nil {
			t.Fatalf("os.WriteFile(%s): %v", goldenPath, err)
		}
	}
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%s): %v", goldenPath, err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatalf("golden mismatch\n--- actual ---\n%s\n--- expected ---\n%s", actual, expected)
	}
}

func publish(t *testing.T, bus *events.Bus, spec events.Spec) {
	t.Helper()
	if _, err := bus.Publish(spec); err != nil {
		t.Fatalf("bus.Publish returned error: %v", err)
	}
}

func marshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent returned error: %v", err)
	}
	return append(encoded, '\n')
}

func updateGoldens() bool {
	return os.Getenv("DEVARCH_UPDATE_GOLDENS") == "1"
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := stdruntime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
