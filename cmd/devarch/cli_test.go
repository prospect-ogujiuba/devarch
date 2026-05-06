package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	stdruntime "runtime"
	"strings"
	"testing"

	"github.com/prospect-ogujiuba/devarch/internal/apply"
	"github.com/prospect-ogujiuba/devarch/internal/appsvc"
	planpkg "github.com/prospect-ogujiuba/devarch/internal/plan"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	"github.com/prospect-ogujiuba/devarch/internal/workflows"
)

func TestRunJSONWorkflowCommands(t *testing.T) {
	args := append(baseCLIArgs(t), "--json", "doctor")
	stdout, stderr, err := runCLI(args, newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI doctor returned error: %v\nstderr:\n%s", err, stderr)
	}
	var doctor appsvc.DoctorReport
	if err := json.Unmarshal([]byte(stdout), &doctor); err != nil {
		t.Fatalf("json.Unmarshal doctor returned error: %v\nstdout:\n%s", err, stdout)
	}
	if doctor.Status == "" || len(doctor.Checks) == 0 {
		t.Fatalf("doctor = %#v, want status and checks", doctor)
	}

	args = append(baseCLIArgs(t), "--json", "runtime", "status")
	stdout, stderr, err = runCLI(args, newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI runtime status returned error: %v\nstderr:\n%s", err, stderr)
	}
	var runtimeStatus appsvc.RuntimeStatusReport
	if err := json.Unmarshal([]byte(stdout), &runtimeStatus); err != nil {
		t.Fatalf("json.Unmarshal runtime status returned error: %v\nstdout:\n%s", err, stdout)
	}
	if runtimeStatus.Status == "" || len(runtimeStatus.Checks) == 0 {
		t.Fatalf("runtimeStatus = %#v, want status and checks", runtimeStatus)
	}

	args = append(baseCLIArgs(t), "--json", "socket", "status")
	stdout, stderr, err = runCLI(args, newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI socket status returned error: %v\nstderr:\n%s", err, stderr)
	}
	var socketStatus appsvc.SocketStatusReport
	if err := json.Unmarshal([]byte(stdout), &socketStatus); err != nil {
		t.Fatalf("json.Unmarshal socket status returned error: %v\nstdout:\n%s", err, stdout)
	}
	if socketStatus.Check.ID != "podman.socket" {
		t.Fatalf("socketStatus.Check.ID = %q, want podman.socket", socketStatus.Check.ID)
	}
}

func TestRunInvalidWorkflowArgs(t *testing.T) {
	_, _, err := runCLI(append(baseCLIArgs(t), "runtime"), newTestServiceFactory(t))
	if err == nil || !strings.Contains(err.Error(), "runtime status") {
		t.Fatalf("runtime error = %v, want runtime status usage error", err)
	}
	_, _, err = runCLI(append(baseCLIArgs(t), "socket", "bounce"), newTestServiceFactory(t))
	if err == nil || !strings.Contains(err.Error(), "unknown socket subcommand") {
		t.Fatalf("socket error = %v, want unknown socket subcommand", err)
	}
}

func TestRunJSONWorkspaceCommands(t *testing.T) {
	args := append(baseCLIArgs(t), "--json", "workspace", "plan", "shop-local")
	stdout, stderr, err := runCLI(args, newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI plan returned error: %v\nstderr:\n%s", err, stderr)
	}
	var plan planpkg.Result
	if err := json.Unmarshal([]byte(stdout), &plan); err != nil {
		t.Fatalf("json.Unmarshal plan returned error: %v\nstdout:\n%s", err, stdout)
	}
	if got, want := plan.Workspace, "shop-local"; got != want {
		t.Fatalf("plan.Workspace = %q, want %q", got, want)
	}
	if len(plan.Actions) == 0 {
		t.Fatal("expected plan actions")
	}

	args = append(baseCLIArgs(t), "--json", "workspace", "status", "shop-local")
	stdout, stderr, err = runCLI(args, newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI status returned error: %v\nstderr:\n%s", err, stderr)
	}
	var status map[string]any
	if err := json.Unmarshal([]byte(stdout), &status); err != nil {
		t.Fatalf("json.Unmarshal status returned error: %v\nstdout:\n%s", err, stdout)
	}
	if _, ok := status["desired"]; !ok {
		t.Fatalf("status = %#v, want desired key", status)
	}
	if _, ok := status["snapshot"]; !ok {
		t.Fatalf("status = %#v, want snapshot key", status)
	}

	args = append(baseCLIArgs(t), "--json", "workspace", "apply", "shop-local")
	stdout, stderr, err = runCLI(args, newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI apply returned error: %v\nstderr:\n%s", err, stderr)
	}
	var result apply.Result
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("json.Unmarshal apply returned error: %v\nstdout:\n%s", err, stdout)
	}
	if got, want := result.Workspace, "shop-local"; got != want {
		t.Fatalf("apply.Workspace = %q, want %q", got, want)
	}
	if len(result.Operations) == 0 {
		t.Fatal("expected apply operations")
	}
}

func TestRunJSONCatalogImportAndScanCommands(t *testing.T) {
	args := append(baseCLIArgs(t), "--json", "catalog", "show", "postgres")
	stdout, stderr, err := runCLI(args, newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI catalog show returned error: %v\nstderr:\n%s", err, stderr)
	}
	var template map[string]any
	if err := json.Unmarshal([]byte(stdout), &template); err != nil {
		t.Fatalf("json.Unmarshal template returned error: %v\nstdout:\n%s", err, stdout)
	}
	if got, want := template["name"], any("postgres"); got != want {
		t.Fatalf("template[name] = %#v, want %#v", got, want)
	}

	stackPath := filepath.Join(repoRoot(t), "examples", "v1", "stacks", "shop-export.yaml")
	args = append(baseCLIArgs(t), "--json", "import", "v1-stack", stackPath)
	stdout, stderr, err = runCLI(args, newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI import returned error: %v\nstderr:\n%s", err, stderr)
	}
	var preview appsvc.ImportPreview
	if err := json.Unmarshal([]byte(stdout), &preview); err != nil {
		t.Fatalf("json.Unmarshal preview returned error: %v\nstdout:\n%s", err, stdout)
	}
	if got, want := preview.Status, "partial"; got != want {
		t.Fatalf("preview.Status = %q, want %q", got, want)
	}
	if got, want := preview.Summary.Total, 1; got != want {
		t.Fatalf("preview.Summary.Total = %d, want %d", got, want)
	}

	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "package.json"), `{"name":"shop-api","dependencies":{"express":"^4.19.0"}}`)
	writeFile(t, filepath.Join(projectRoot, "compose.yml"), `services:
  db:
    image: postgres:16
`)
	args = append(baseCLIArgs(t), "--json", "scan", "project", projectRoot)
	stdout, stderr, err = runCLI(args, newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI scan returned error: %v\nstderr:\n%s", err, stderr)
	}
	var scan appsvc.ProjectScanView
	if err := json.Unmarshal([]byte(stdout), &scan); err != nil {
		t.Fatalf("json.Unmarshal scan returned error: %v\nstdout:\n%s", err, stdout)
	}
	if got, want := scan.ProjectType, "node"; got != want {
		t.Fatalf("scan.ProjectType = %q, want %q", got, want)
	}
	if got, want := scan.SuggestedTemplates, []string{"node-api", "postgres"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("scan.SuggestedTemplates = %v, want %v", got, want)
	}
}

func TestRunHumanWorkspaceCommands(t *testing.T) {
	stdout, stderr, err := runCLI(append(baseCLIArgs(t), "workspace", "list"), newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI workspace list returned error: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "shop-local") {
		t.Fatalf("workspace list stdout = %q, want shop-local", stdout)
	}

	stdout, stderr, err = runCLI(append(baseCLIArgs(t), "workspace", "open", "shop-local"), newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI workspace open returned error: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "Manifest:") || !strings.Contains(stdout, "api, postgres, redis, web") {
		t.Fatalf("workspace open stdout = %q, want manifest and resources", stdout)
	}

	stdout, stderr, err = runCLI(append(baseCLIArgs(t), "workspace", "logs", "--tail", "5", "shop-local", "api"), newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI workspace logs returned error: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "ready") {
		t.Fatalf("workspace logs stdout = %q, want ready", stdout)
	}

	stdout, stderr, err = runCLI(append(baseCLIArgs(t), "workspace", "exec", "shop-local", "api", "--", "echo", "ok"), newTestServiceFactory(t))
	if err != nil {
		t.Fatalf("runCLI workspace exec returned error: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "ok") {
		t.Fatalf("workspace exec stdout = %q, want ok", stdout)
	}
}

func TestRunRequiresWorkspaceRootForWorkspaceCommands(t *testing.T) {
	_, _, err := runCLI([]string{"workspace", "list"}, newTestServiceFactory(t))
	if err == nil || !strings.Contains(err.Error(), "--workspace-root") {
		t.Fatalf("runCLI error = %v, want missing workspace-root error", err)
	}
}

func runCLI(args []string, factory serviceFactory) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), args, &stdout, &stderr, factory)
	return stdout.String(), stderr.String(), err
}

func baseCLIArgs(t *testing.T) []string {
	t.Helper()
	return []string{
		"--workspace-root", filepath.Join(repoRoot(t), "examples", "v2", "workspaces"),
		"--catalog-root", filepath.Join(repoRoot(t), "catalog", "builtin"),
	}
}

func newTestServiceFactory(t *testing.T) serviceFactory {
	t.Helper()
	return func(cfg cliConfig) (serviceAPI, error) {
		return appsvc.New(appsvc.Config{
			WorkspaceRoots: cfg.workspaceRoots,
			CatalogRoots:   cfg.catalogRoots,
			Adapters: map[string]runtimepkg.Adapter{
				runtimepkg.ProviderDocker: &fakeAdapter{
					provider: runtimepkg.ProviderDocker,
					capabilities: runtimepkg.AdapterCapabilities{
						Inspect: true,
						Apply:   true,
						Logs:    true,
						Exec:    true,
						Network: true,
					},
					snapshot:   &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: "shop-local", Provider: runtimepkg.ProviderDocker}},
					logChunks:  []runtimepkg.LogChunk{{Line: "ready", Stream: "combined"}},
					execResult: &runtimepkg.ExecResult{ExitCode: 0, Stdout: "ok\n"},
				},
			},
			LookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
			WorkflowRunner: &fakeWorkflowCommandRunner{Results: []workflows.CommandResult{
				{Status: workflows.StatusPass, StdoutSummary: "podman version 5.0"},
				{Status: workflows.StatusPass, StdoutSummary: "socket ready"},
				{Status: workflows.StatusPass, StdoutSummary: "package ok"},
				{Status: workflows.StatusPass, StdoutSummary: "podman version 5.0"},
				{Status: workflows.StatusFail, StderrSummary: "docker unavailable", Error: "not found", ExitCode: -1},
				{Status: workflows.StatusPass, StdoutSummary: "socket ready"},
			}},
		})
	}
}

type fakeWorkflowCommandRunner struct {
	Results []workflows.CommandResult
	Calls   []workflows.CommandResult
}

func (f *fakeWorkflowCommandRunner) Run(ctx context.Context, command string, args ...string) workflows.CommandResult {
	_ = ctx
	call := workflows.CommandResult{Command: command, Args: append([]string(nil), args...)}
	f.Calls = append(f.Calls, call)
	if len(f.Results) == 0 {
		return workflows.CommandResult{Command: command, Args: append([]string(nil), args...), Status: workflows.StatusFail, ExitCode: -1, Error: "fake result missing"}
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

type fakeAdapter struct {
	provider     string
	capabilities runtimepkg.AdapterCapabilities
	snapshot     *runtimepkg.Snapshot
	logChunks    []runtimepkg.LogChunk
	execResult   *runtimepkg.ExecResult
}

func (f *fakeAdapter) Provider() string { return f.provider }

func (f *fakeAdapter) Capabilities() runtimepkg.AdapterCapabilities { return f.capabilities }

func (f *fakeAdapter) InspectWorkspace(context.Context, *runtimepkg.DesiredWorkspace) (*runtimepkg.Snapshot, error) {
	if f.snapshot == nil {
		return &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: "shop-local", Provider: f.provider}}, nil
	}
	return f.snapshot, nil
}

func (f *fakeAdapter) EnsureNetwork(context.Context, *runtimepkg.DesiredNetwork) error { return nil }
func (f *fakeAdapter) RemoveNetwork(context.Context, *runtimepkg.DesiredNetwork) error { return nil }
func (f *fakeAdapter) ApplyResource(context.Context, runtimepkg.ApplyResourceRequest) error {
	return nil
}
func (f *fakeAdapter) RemoveResource(context.Context, runtimepkg.ResourceRef) error  { return nil }
func (f *fakeAdapter) RestartResource(context.Context, runtimepkg.ResourceRef) error { return nil }

func (f *fakeAdapter) StreamLogs(_ context.Context, _ runtimepkg.ResourceRef, _ runtimepkg.LogsRequest, consume runtimepkg.LogsConsumer) error {
	for _, chunk := range f.logChunks {
		if err := consume(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeAdapter) Exec(context.Context, runtimepkg.ResourceRef, runtimepkg.ExecRequest) (*runtimepkg.ExecResult, error) {
	if f.execResult == nil {
		return &runtimepkg.ExecResult{ExitCode: 0}, nil
	}
	return f.execResult, nil
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := stdruntime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%s): %v", path, err)
	}
}

var _ runtimepkg.Adapter = (*fakeAdapter)(nil)
