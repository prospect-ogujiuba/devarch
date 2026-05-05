package workflows

import (
	"context"
	"testing"
)

func TestSocketStatusWarnsWhenUnavailable(t *testing.T) {
	runner := &FakeRunner{Results: []CommandResult{{Status: StatusFail, ExitCode: 1, StderrSummary: "inactive"}}}
	report := SocketStatus(context.Background(), runner)
	if report.Status != StatusWarn {
		t.Fatalf("status = %s, want warn", report.Status)
	}
}

func TestSocketStartStopUseSystemdUserBoundary(t *testing.T) {
	runner := &FakeRunner{Results: []CommandResult{{Status: StatusPass}, {Status: StatusFail, ExitCode: 1, Error: "failed"}}}
	if _, err := SocketStart(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	if _, err := SocketStop(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	if runner.Calls[0].Command != "systemctl" || runner.Calls[0].Args[1] != "start" {
		t.Fatalf("start call = %#v", runner.Calls[0])
	}
	if runner.Calls[1].Args[1] != "stop" {
		t.Fatalf("stop call = %#v", runner.Calls[1])
	}
}
