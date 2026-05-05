// Package podmanctl contains low-level, testable Podman command helpers.
package podmanctl

import (
	"context"
	"os/exec"
)

// Runner executes a command and returns combined stdout/stderr output.
type Runner interface {
	Run(ctx context.Context, command string, args ...string) ([]byte, error)
}

// ExecRunner executes commands using os/exec.
type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	return cmd.CombinedOutput()
}

// Podman invokes the podman binary through runner.
func Podman(ctx context.Context, runner Runner, args ...string) ([]byte, error) {
	return runner.Run(ctx, "podman", args...)
}
