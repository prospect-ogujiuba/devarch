// Package runner is the only place DevArch starts processes.
//
// Convention: Run and Exec are for commands that change something and are
// echoed (and skipped under --dry-run). Output is for read-only queries and
// always executes, so dry runs can still inspect real state.
package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Cmd is a native command invocation.
type Cmd struct {
	Name string
	Args []string
	// Dir is the working directory; empty means the current directory.
	Dir string
	// Env holds extra KEY=VALUE entries layered over the current environment.
	Env    []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// String renders the command as a copy-pasteable shell line.
func (c Cmd) String() string {
	var b strings.Builder
	for _, kv := range c.Env {
		k, v, _ := strings.Cut(kv, "=")
		b.WriteString(k + "=" + Quote(v) + " ")
	}
	b.WriteString(Quote(c.Name))
	for _, a := range c.Args {
		b.WriteString(" " + Quote(a))
	}
	if c.Dir == "" {
		return b.String()
	}
	return "(cd " + Quote(displayDir(c.Dir)) + " && " + b.String() + ")"
}

func displayDir(dir string) string {
	wd, err := os.Getwd()
	if err != nil {
		return dir
	}
	if rel, err := filepath.Rel(wd, dir); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return dir
}

// Quote shell-quotes s only when needed.
func Quote(s string) string {
	if s == "" {
		return "''"
	}
	safe := true
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("-_./:=@%+,", r)) {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// ExitError reports a command that ran and exited non-zero.
type ExitError struct {
	Cmd    string
	Code   int
	Stderr string
}

func (e *ExitError) Error() string {
	msg := fmt.Sprintf("%s: exit status %d", e.Cmd, e.Code)
	if s := strings.TrimSpace(e.Stderr); s != "" {
		msg += ": " + s
	}
	return msg
}

// Runner executes commands.
type Runner interface {
	// Run echoes and executes a mutating command, streaming its output.
	Run(ctx context.Context, c Cmd) error
	// Output executes a read-only query and returns its stdout.
	Output(ctx context.Context, c Cmd) ([]byte, error)
	// Exec hands the terminal to the command (logs, shells, passthrough).
	Exec(c Cmd) error
}

// System runs real processes and echoes mutating commands to Log.
type System struct {
	Log io.Writer
}

func (s System) echo(c Cmd) {
	if s.Log != nil {
		fmt.Fprintln(s.Log, "$ "+c.String())
	}
}

func build(ctx context.Context, c Cmd) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	if len(c.Env) > 0 {
		cmd.Env = append(os.Environ(), c.Env...)
	}
	cmd.Stdin = c.Stdin
	return cmd
}

func (s System) Run(ctx context.Context, c Cmd) error {
	s.echo(c)
	cmd := build(ctx, c)
	cmd.Stdout = writer(c.Stdout, os.Stdout)
	cmd.Stderr = writer(c.Stderr, os.Stderr)
	return wrap(c, cmd.Run(), "")
}

func (s System) Output(ctx context.Context, c Cmd) ([]byte, error) {
	cmd := build(ctx, c)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	return out, wrap(c, err, stderr.String())
}

func (s System) Exec(c Cmd) error {
	s.echo(c)
	cmd := build(context.Background(), c)
	if c.Stdin == nil {
		cmd.Stdin = os.Stdin
	}
	cmd.Stdout = writer(c.Stdout, os.Stdout)
	cmd.Stderr = writer(c.Stderr, os.Stderr)
	return wrap(c, cmd.Run(), "")
}

func writer(w io.Writer, def io.Writer) io.Writer {
	if w == nil {
		return def
	}
	return w
}

func wrap(c Cmd, err error, stderr string) error {
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return &ExitError{Cmd: c.String(), Code: exitErr.ExitCode(), Stderr: stderr}
	}
	return fmt.Errorf("%s: %w", c.String(), err)
}

// DryRun prints mutating commands instead of running them. Queries still run.
type DryRun struct {
	Inner Runner
	Log   io.Writer
}

func (d DryRun) Run(_ context.Context, c Cmd) error {
	fmt.Fprintln(d.Log, "would run: "+c.String())
	return nil
}

func (d DryRun) Output(ctx context.Context, c Cmd) ([]byte, error) {
	return d.Inner.Output(ctx, c)
}

func (d DryRun) Exec(c Cmd) error {
	fmt.Fprintln(d.Log, "would run: "+c.String())
	return nil
}
