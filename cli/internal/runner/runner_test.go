package runner

import (
	"context"
	"strings"
	"testing"
)

func TestQuote(t *testing.T) {
	for in, want := range map[string]string{
		"plain":        "plain",
		"":             "''",
		"has space":    "'has space'",
		"it's":         `'it'\''s'`,
		"a=b/c:d@e%f+": "a=b/c:d@e%f+",
		"$HOME":        "'$HOME'",
	} {
		if got := Quote(in); got != want {
			t.Errorf("Quote(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestCmdString(t *testing.T) {
	c := Cmd{Name: "podman", Args: []string{"compose", "up", "-d"}, Dir: "/abs/dir", Env: []string{"PHP_VERSION=8.3"}}
	if got := c.String(); got != "(cd /abs/dir && PHP_VERSION=8.3 podman compose up -d)" {
		t.Fatal(got)
	}
}

func TestDryRunSkipsMutationsButRunsQueries(t *testing.T) {
	inner := &Fake{Responses: map[string]Response{"podman ps": {Out: []byte("[]")}}}
	var log strings.Builder
	d := DryRun{Inner: inner, Log: &log}
	if err := d.Run(context.Background(), Cmd{Name: "podman", Args: []string{"rm", "x"}}); err != nil {
		t.Fatal(err)
	}
	out, _ := d.Output(context.Background(), Cmd{Name: "podman", Args: []string{"ps"}})
	if string(out) != "[]" || len(inner.Cmds) != 1 || log.String() != "would run: podman rm x\n" {
		t.Fatalf("out=%q cmds=%v log=%q", out, inner.Cmds, log.String())
	}
}

func TestSystemExitError(t *testing.T) {
	_, err := System{}.Output(context.Background(), Cmd{Name: "sh", Args: []string{"-c", "echo boom >&2; exit 3"}})
	e, ok := err.(*ExitError)
	if !ok || e.Code != 3 || strings.TrimSpace(e.Stderr) != "boom" {
		t.Fatalf("got %#v", err)
	}
}
