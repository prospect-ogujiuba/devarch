package workflows

import "context"

// RuntimeStatusReport reports Podman-first runtime health. Docker is only
// compatibility/status information.
type RuntimeStatusReport struct {
	Status WorkflowStatus `json:"status"`
	Checks []CheckResult  `json:"checks"`
}

func RuntimeStatus(ctx context.Context, runner Runner) *RuntimeStatusReport {
	if runner == nil {
		runner = ExecRunner{}
	}
	podman := runner.Run(ctx, "podman", "--version")
	checks := []CheckResult{}
	if podman.Status == StatusPass {
		checks = append(checks, CheckResult{ID: "runtime.podman", Name: "Podman runtime", Status: StatusPass, Message: firstNonEmpty(podman.StdoutSummary, "podman available")})
	} else {
		checks = append(checks, CheckResult{ID: "runtime.podman", Name: "Podman runtime", Status: StatusFail, Message: "podman unavailable", Diagnostics: []Diagnostic{{ID: "runtime.podman.unavailable", Severity: StatusFail, Message: podman.StderrSummary, Detail: podman.Error}}})
	}
	docker := runner.Run(ctx, "docker", "--version")
	if docker.Status == StatusPass {
		checks = append(checks, CheckResult{ID: "runtime.docker.compat", Name: "Docker compatibility", Status: StatusWarn, Message: "docker available for compatibility only"})
	} else {
		checks = append(checks, CheckResult{ID: "runtime.docker.compat", Name: "Docker compatibility", Status: StatusUnavailable, Message: "docker not available; podman remains canonical"})
	}
	return &RuntimeStatusReport{Status: ReportStatus(checks), Checks: checks}
}
