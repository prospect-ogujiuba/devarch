package workflows

import (
	"context"
	"testing"
)

func TestRuntimeStatusPodmanFirst(t *testing.T) {
	runner := &FakeRunner{Results: []CommandResult{{Status: StatusPass, StdoutSummary: "podman version 5"}, {Status: StatusFail, Error: "docker missing"}}}
	report := RuntimeStatus(context.Background(), runner)
	if report.Checks[0].ID != "runtime.podman" || report.Checks[0].Status != StatusPass {
		t.Fatalf("podman check = %#v", report.Checks[0])
	}
	if report.Checks[1].Status != StatusUnavailable {
		t.Fatalf("docker should be compatibility only: %#v", report.Checks[1])
	}
}
