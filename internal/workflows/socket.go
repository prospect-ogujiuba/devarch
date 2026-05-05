package workflows

import "context"

// SocketStatus describes Podman socket state.
type SocketStatusReport struct {
	Status WorkflowStatus `json:"status"`
	Check  CheckResult    `json:"check"`
}

func SocketStatus(ctx context.Context, runner Runner) *SocketStatusReport {
	if runner == nil {
		runner = ExecRunner{}
	}
	result := runner.Run(ctx, "podman", "system", "connection", "list")
	if result.Status == StatusPass {
		check := CheckResult{ID: "podman.socket", Name: "Podman socket", Status: StatusPass, Message: "podman connection list available"}
		return &SocketStatusReport{Status: StatusPass, Check: check}
	}
	check := CheckResult{ID: "podman.socket", Name: "Podman socket", Status: StatusWarn, Message: "podman socket unavailable", Diagnostics: []Diagnostic{{ID: "podman.socket.unavailable", Severity: StatusWarn, Message: result.StderrSummary, Detail: result.Error}}}
	return &SocketStatusReport{Status: StatusWarn, Check: check}
}

func SocketStart(ctx context.Context, runner Runner) (*CommandResult, error) {
	return socketCommand(ctx, runner, "start")
}

func SocketStop(ctx context.Context, runner Runner) (*CommandResult, error) {
	return socketCommand(ctx, runner, "stop")
}

func socketCommand(ctx context.Context, runner Runner, action string) (*CommandResult, error) {
	if runner == nil {
		runner = ExecRunner{}
	}
	result := runner.Run(ctx, "systemctl", "--user", action, "podman.socket")
	return &result, nil
}
