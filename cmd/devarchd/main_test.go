package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	stdruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/appsvc"
)

func TestServeStartsAndShutsDown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- serve(ctx, config{
			workspaceRoots: []string{filepath.Join(repoRoot(t), "examples", "v2", "workspaces")},
			catalogRoots:   []string{filepath.Join(repoRoot(t), "catalog", "builtin")},
			listen:         "127.0.0.1:0",
			cachePath:      filepath.Join(t.TempDir(), "devarchd.db"),
		}, ready)
	}()

	addr := <-ready
	resp, err := http.Get("http://" + addr + "/api/workspaces")
	if err != nil {
		t.Fatalf("http.Get returned error: %v", err)
	}
	defer resp.Body.Close()
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("GET /api/workspaces status = %d, want %d", got, want)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("serve returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for serve shutdown")
	}
}

func TestServeFailsClearlyOnDuplicateWorkspaceNames(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceCopy(t, filepath.Join(repoRoot(t), "examples", "v2", "workspaces", "shop-local", "devarch.workspace.yaml"), filepath.Join(root, "one", "devarch.workspace.yaml"), "duplicate-local")
	writeWorkspaceCopy(t, filepath.Join(repoRoot(t), "examples", "v2", "workspaces", "laravel-local", "devarch.workspace.yaml"), filepath.Join(root, "two", "devarch.workspace.yaml"), "duplicate-local")

	err := serve(context.Background(), config{workspaceRoots: []string{root}, listen: "127.0.0.1:0"}, nil)
	var duplicateErr *appsvc.DuplicateWorkspaceNameError
	if !errors.As(err, &duplicateErr) {
		t.Fatalf("serve duplicate error = %v, want DuplicateWorkspaceNameError", err)
	}
}

func TestParseConfigRequiresWorkspaceRoots(t *testing.T) {
	_, err := parseConfig(nil)
	if err == nil {
		t.Fatal("expected parseConfig error")
	}
}

func writeWorkspaceCopy(t *testing.T, sourcePath, targetPath, name string) {
	t.Helper()
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("os.ReadFile(%s): %v", sourcePath, err)
	}
	content := strings.ReplaceAll(string(data), "shop-local", name)
	content = strings.ReplaceAll(content, "laravel-local", name)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(targetPath), err)
	}
	if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%s): %v", targetPath, err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := stdruntime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
