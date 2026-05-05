package podmanctl

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestEnsureNetworkSkipsCreateWhenExists(t *testing.T) {
	runner := &fakeRunner{}
	if err := EnsureNetwork(context.Background(), runner, "dev", nil); err != nil {
		t.Fatalf("EnsureNetwork returned error: %v", err)
	}
	want := []call{{command: "podman", args: []string{"network", "exists", "dev"}}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestEnsureNetworkCreatesWithSortedLabels(t *testing.T) {
	runner := &fakeRunner{outs: [][]byte{[]byte("network not found")}, errs: []error{errors.New("network not found")}}
	err := EnsureNetwork(context.Background(), runner, "dev", map[string]string{"z": "last", "a": "first"})
	if err != nil {
		t.Fatalf("EnsureNetwork returned error: %v", err)
	}
	want := []call{
		{command: "podman", args: []string{"network", "exists", "dev"}},
		{command: "podman", args: []string{"network", "create", "--label", "a=first", "--label", "z=last", "dev"}},
	}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestNetworkExistsUnexpectedError(t *testing.T) {
	runner := &fakeRunner{errs: []error{errors.New("permission denied")}}
	_, err := NetworkExists(context.Background(), runner, "dev")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoveNetworkExisting(t *testing.T) {
	runner := &fakeRunner{}
	if err := RemoveNetwork(context.Background(), runner, "dev"); err != nil {
		t.Fatalf("RemoveNetwork returned error: %v", err)
	}
	want := []call{{command: "podman", args: []string{"network", "rm", "dev"}}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestRemoveNetworkMissingDoesNotFail(t *testing.T) {
	runner := &fakeRunner{outs: [][]byte{[]byte("no such network")}, errs: []error{errors.New("exit status 1")}}
	if err := RemoveNetwork(context.Background(), runner, "dev"); err != nil {
		t.Fatalf("RemoveNetwork returned error: %v", err)
	}
}

func TestRemoveNetworkUnexpectedError(t *testing.T) {
	runner := &fakeRunner{errs: []error{errors.New("daemon unavailable")}}
	if err := RemoveNetwork(context.Background(), runner, "dev"); err == nil {
		t.Fatal("expected error")
	}
}
