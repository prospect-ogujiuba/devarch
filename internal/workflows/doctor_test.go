package workflows

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDoctorReportsFailingPodman(t *testing.T) {
	runner := &FakeRunner{Results: []CommandResult{
		{Status: StatusFail, ExitCode: -1, Error: "not found"},
		{Status: StatusFail, ExitCode: 1, StderrSummary: "socket missing"},
		{Status: StatusPass},
	}}
	report, err := Doctor(context.Background(), runner, DoctorOptions{WorkspaceRoots: []string{"."}, CatalogRoots: []string{"."}})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %s, want fail", report.Status)
	}
	if _, err := json.Marshal(report); err != nil {
		t.Fatalf("report must serialize: %v", err)
	}
}

func TestDoctorPassesWithReadableRoots(t *testing.T) {
	runner := &FakeRunner{Results: []CommandResult{{Status: StatusPass, StdoutSummary: "podman version 5"}, {Status: StatusPass}, {Status: StatusPass}}}
	report, err := Doctor(context.Background(), runner, DoctorOptions{WorkspaceRoots: []string{"."}, CatalogRoots: []string{"."}})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %s, want pass", report.Status)
	}
}
