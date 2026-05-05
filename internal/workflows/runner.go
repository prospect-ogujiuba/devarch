package workflows

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

// Runner is the host-command boundary for workflows that cannot avoid probing
// local tools such as podman or systemctl.
type Runner interface {
	Run(ctx context.Context, command string, args ...string) CommandResult
}

// ExecRunner executes commands on the host.
type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, command string, args ...string) CommandResult {
	cmd := exec.CommandContext(ctx, command, args...)
	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	result := CommandResult{Command: command, Args: append([]string(nil), args...)}
	err := cmd.Run()
	result.StdoutSummary = summarize(stdout.String())
	result.StderrSummary = summarize(stderr.String())
	if err == nil {
		result.Status = StatusPass
		return result
	}
	result.Status = StatusFail
	result.Error = err.Error()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
	} else {
		result.ExitCode = -1
	}
	return result
}

func summarize(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 512 {
		return value
	}
	return value[:512]
}

// FakeRunner is a test helper for deterministic workflow tests.
type FakeRunner struct {
	Results []CommandResult
	Calls   []CommandResult
}

func (f *FakeRunner) Run(ctx context.Context, command string, args ...string) CommandResult {
	_ = ctx
	call := CommandResult{Command: command, Args: append([]string(nil), args...)}
	f.Calls = append(f.Calls, call)
	if len(f.Results) == 0 {
		return CommandResult{Command: command, Args: append([]string(nil), args...), Status: StatusFail, ExitCode: -1, Error: "fake result missing"}
	}
	result := f.Results[0]
	f.Results = f.Results[1:]
	if result.Command == "" {
		result.Command = command
	}
	if result.Args == nil {
		result.Args = append([]string(nil), args...)
	}
	return result
}
