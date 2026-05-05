package podmanctl

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

func NetworkExists(ctx context.Context, runner Runner, name string) (bool, error) {
	output, err := Podman(ctx, runner, "network", "exists", name)
	if err == nil {
		return true, nil
	}
	if isNotFound(output, err) {
		return false, nil
	}
	return false, fmt.Errorf("podman network exists %q: %w", name, err)
}

func EnsureNetwork(ctx context.Context, runner Runner, name string, labels map[string]string) error {
	exists, err := NetworkExists(ctx, runner, name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	args := []string{"network", "create"}
	for _, key := range sortedKeys(labels) {
		args = append(args, "--label", key+"="+labels[key])
	}
	args = append(args, name)
	if _, err := Podman(ctx, runner, args...); err != nil {
		return fmt.Errorf("podman network create %q: %w", name, err)
	}
	return nil
}

func RemoveNetwork(ctx context.Context, runner Runner, name string) error {
	output, err := Podman(ctx, runner, "network", "rm", name)
	if err == nil || isNotFound(output, err) {
		return nil
	}
	return fmt.Errorf("podman network rm %q: %w", name, err)
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func isNotFound(output []byte, err error) bool {
	text := strings.ToLower(strings.TrimSpace(string(output) + " " + errString(err)))
	return strings.Contains(text, "not found") || strings.Contains(text, "no such") || strings.Contains(text, "does not exist")
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
