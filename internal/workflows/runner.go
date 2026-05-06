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
