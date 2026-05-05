package podmanctl

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type call struct {
	command string
	args    []string
}

type fakeRunner struct {
	calls []call
	outs  [][]byte
	errs  []error
}

func (f *fakeRunner) Run(ctx context.Context, command string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, call{command: command, args: append([]string(nil), args...)})
	idx := len(f.calls) - 1
	var out []byte
	if idx < len(f.outs) {
		out = f.outs[idx]
	}
	var err error
	if idx < len(f.errs) {
		err = f.errs[idx]
	}
	return out, err
}

func TestPodmanForwardsCommandAndArgs(t *testing.T) {
	runner := &fakeRunner{outs: [][]byte{[]byte("ok")}}
	out, err := Podman(context.Background(), runner, "network", "exists", "dev")
	if err != nil {
		t.Fatalf("Podman returned error: %v", err)
	}
	if string(out) != "ok" {
		t.Fatalf("output = %q", out)
	}
	want := []call{{command: "podman", args: []string{"network", "exists", "dev"}}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestPodmanPropagatesRunnerError(t *testing.T) {
	boom := errors.New("boom")
	runner := &fakeRunner{errs: []error{boom}}
	_, err := Podman(context.Background(), runner, "ps")
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want %v", err, boom)
	}
}
