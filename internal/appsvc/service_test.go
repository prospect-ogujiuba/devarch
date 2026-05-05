package appsvc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	stdruntime "runtime"
	"strings"
	"testing"

	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/internal/events"
	planpkg "github.com/prospect-ogujiuba/devarch/internal/plan"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
	"github.com/prospect-ogujiuba/devarch/internal/workflows"
)

func TestDiscoverWorkspacesSortsByNameAndRejectsDuplicates(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceCopy(t, filepath.Join(repoRoot(t), "examples", "v2", "workspaces", "shop-local", "devarch.workspace.yaml"), filepath.Join(root, "zeta", "devarch.workspace.yaml"), "zeta-local", "Zeta Local")
	writeWorkspaceCopy(t, filepath.Join(repoRoot(t), "examples", "v2", "workspaces", "laravel-local", "devarch.workspace.yaml"), filepath.Join(root, "alpha", "devarch.workspace.yaml"), "alpha-local", "Alpha Local")

	workspaces, err := DiscoverWorkspaces([]string{root})
	if err != nil {
		t.Fatalf("DiscoverWorkspaces returned error: %v", err)
	}
	if got, want := []string{workspaces[0].Metadata.Name, workspaces[1].Metadata.Name}, []string{"alpha-local", "zeta-local"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("workspace order = %v, want %v", got, want)
	}

	dupRoot := t.TempDir()
	writeWorkspaceCopy(t, filepath.Join(repoRoot(t), "examples", "v2", "workspaces", "shop-local", "devarch.workspace.yaml"), filepath.Join(dupRoot, "one", "devarch.workspace.yaml"), "duplicate-local", "Duplicate Local")
	writeWorkspaceCopy(t, filepath.Join(repoRoot(t), "examples", "v2", "workspaces", "laravel-local", "devarch.workspace.yaml"), filepath.Join(dupRoot, "two", "devarch.workspace.yaml"), "duplicate-local", "Duplicate Local")

	_, err = DiscoverWorkspaces([]string{dupRoot})
	var duplicateErr *DuplicateWorkspaceNameError
	if !errors.As(err, &duplicateErr) {
		t.Fatalf("DiscoverWorkspaces duplicate error = %v, want DuplicateWorkspaceNameError", err)
	}
}

func TestServiceDescribeProviderUsesDeterministicAutoOrder(t *testing.T) {
	service := newTestService(t, Config{
		WorkspaceRoots: exampleWorkspaceRoots(t),
		CatalogRoots:   exampleCatalogRoots(t),
		Adapters: map[string]runtimepkg.Adapter{
			runtimepkg.ProviderDocker: &fakeAdapter{provider: runtimepkg.ProviderDocker},
			runtimepkg.ProviderPodman: &fakeAdapter{provider: runtimepkg.ProviderPodman},
		},
		LookPath: func(file string) (string, error) {
			return "/usr/bin/" + file, nil
		},
	})

	provider, capabilities := service.describeProvider(runtimepkg.ProviderAuto)
	if got, want := provider, runtimepkg.ProviderDocker; got != want {
		t.Fatalf("describeProvider(auto) = %q, want %q", got, want)
	}
	if capabilities != (runtimepkg.AdapterCapabilities{}) {
		t.Fatalf("describeProvider(auto) capabilities = %#v, want zero-value default fake capabilities", capabilities)
	}

	service.lookPath = func(file string) (string, error) {
		if file == runtimepkg.ProviderPodman {
			return "/usr/bin/podman", nil
		}
		return "", fmt.Errorf("missing %s", file)
	}
	provider, _ = service.describeProvider(runtimepkg.ProviderAuto)
	if got, want := provider, runtimepkg.ProviderPodman; got != want {
		t.Fatalf("describeProvider(auto podman fallback) = %q, want %q", got, want)
	}
}

func TestServiceReadFlowReturnsLockedWorkspaceAndCatalogShapes(t *testing.T) {
	service := newTestService(t, Config{
		WorkspaceRoots: exampleWorkspaceRoots(t),
		CatalogRoots:   exampleCatalogRoots(t),
		Adapters: map[string]runtimepkg.Adapter{
			runtimepkg.ProviderDocker: &fakeAdapter{provider: runtimepkg.ProviderDocker, capabilities: runtimepkg.AdapterCapabilities{Inspect: true, Logs: true, Exec: true}},
		},
		LookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
	})

	templates, err := service.CatalogTemplates(context.Background())
	if err != nil {
		t.Fatalf("CatalogTemplates returned error: %v", err)
	}
	if len(templates) == 0 {
		t.Fatal("expected catalog templates")
	}
	if got, want := templates[0].Name, "laravel-app"; got != want {
		t.Fatalf("templates[0].Name = %q, want %q", got, want)
	}

	detail, err := service.CatalogTemplate(context.Background(), "postgres")
	if err != nil {
		t.Fatalf("CatalogTemplate returned error: %v", err)
	}
	if got, want := detail.Name, "postgres"; got != want {
		t.Fatalf("detail.Name = %q, want %q", got, want)
	}
	if detail.Runtime == nil || detail.Runtime["image"] == nil {
		t.Fatalf("detail.Runtime missing image: %#v", detail.Runtime)
	}

	workspaces, err := service.Workspaces(context.Background())
	if err != nil {
		t.Fatalf("Workspaces returned error: %v", err)
	}
	if len(workspaces) < 3 {
		t.Fatalf("len(workspaces) = %d, want at least 3", len(workspaces))
	}
	if got, want := workspaces[0].Name, "compat-local"; got != want {
		t.Fatalf("workspaces[0].Name = %q, want %q", got, want)
	}

	workspaceDetail, err := service.Workspace(context.Background(), "shop-local")
	if err != nil {
		t.Fatalf("Workspace returned error: %v", err)
	}
	if got, want := workspaceDetail.Provider, runtimepkg.ProviderDocker; got != want {
		t.Fatalf("workspaceDetail.Provider = %q, want %q", got, want)
	}
	if !strings.HasSuffix(filepath.ToSlash(workspaceDetail.ManifestPath), "/examples/v2/workspaces/shop-local/devarch.workspace.yaml") {
		t.Fatalf("workspaceDetail.ManifestPath = %q, want shop-local manifest path", workspaceDetail.ManifestPath)
	}
	if got, want := workspaceDetail.ResourceKeys, []string{"api", "postgres", "redis", "web"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("workspaceDetail.ResourceKeys = %v, want %v", got, want)
	}

	graph, err := service.WorkspaceGraph(context.Background(), "shop-local")
	if err != nil {
		t.Fatalf("WorkspaceGraph returned error: %v", err)
	}
	if graph.Graph == nil || graph.Contracts == nil {
		t.Fatalf("WorkspaceGraph view missing graph/contracts: %#v", graph)
	}
}

func TestServiceStatusAndPlanUseSelectedRuntime(t *testing.T) {
	adapter := &fakeAdapter{
		provider:     runtimepkg.ProviderDocker,
		capabilities: runtimepkg.AdapterCapabilities{Inspect: true, Logs: true, Exec: true},
		snapshot:     &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: "shop-local", Provider: runtimepkg.ProviderDocker}},
	}
	service := newTestService(t, Config{
		WorkspaceRoots: exampleWorkspaceRoots(t),
		CatalogRoots:   exampleCatalogRoots(t),
		Adapters: map[string]runtimepkg.Adapter{
			runtimepkg.ProviderDocker: adapter,
		},
		LookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
	})

	status, err := service.WorkspaceStatus(context.Background(), "shop-local")
	if err != nil {
		t.Fatalf("WorkspaceStatus returned error: %v", err)
	}
	if got, want := status.Desired.Provider, runtimepkg.ProviderDocker; got != want {
		t.Fatalf("status.Desired.Provider = %q, want %q", got, want)
	}
	if status.Snapshot == nil {
		t.Fatal("expected status snapshot")
	}

	plan, err := service.WorkspacePlan(context.Background(), "shop-local")
	if err != nil {
		t.Fatalf("WorkspacePlan returned error: %v", err)
	}
	if got, want := plan.Workspace, "shop-local"; got != want {
		t.Fatalf("plan.Workspace = %q, want %q", got, want)
	}
	if len(plan.Actions) == 0 {
		t.Fatal("expected plan actions")
	}
	if got, want := adapter.inspectCalls, 2; got != want {
		t.Fatalf("adapter.inspectCalls = %d, want %d", got, want)
	}
}

func TestServiceApplyCapabilityGateReturnsTypedError(t *testing.T) {
	service := newTestService(t, Config{
		WorkspaceRoots: exampleWorkspaceRoots(t),
		CatalogRoots:   exampleCatalogRoots(t),
		Adapters: map[string]runtimepkg.Adapter{
			runtimepkg.ProviderDocker: &fakeAdapter{
				provider:     runtimepkg.ProviderDocker,
				capabilities: runtimepkg.AdapterCapabilities{Inspect: true},
				snapshot:     &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: "shop-local", Provider: runtimepkg.ProviderDocker}},
			},
		},
		LookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
	})

	_, err := service.ApplyWorkspace(context.Background(), "shop-local")
	var capabilityErr *UnsupportedCapabilityError
	if !errors.As(err, &capabilityErr) {
		t.Fatalf("ApplyWorkspace error = %v, want UnsupportedCapabilityError", err)
	}
	if got, want := capabilityErr.Operation, "apply"; got != want {
		t.Fatalf("capabilityErr.Operation = %q, want %q", got, want)
	}
	if got, want := capabilityErr.Capability, "network"; got != want {
		t.Fatalf("capabilityErr.Capability = %q, want %q", got, want)
	}
}

func TestServiceLogsExecAndEventsUseSharedBoundary(t *testing.T) {
	bus := events.NewBus()
	adapter := &fakeAdapter{
		provider:     runtimepkg.ProviderDocker,
		capabilities: runtimepkg.AdapterCapabilities{Inspect: true, Logs: true, Exec: true},
		snapshot:     &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: "shop-local", Provider: runtimepkg.ProviderDocker}},
		logChunks:    []runtimepkg.LogChunk{{Line: "ready", Stream: "combined"}},
		execResult:   &runtimepkg.ExecResult{ExitCode: 0, Stdout: "ok\n"},
	}
	service := newTestService(t, Config{
		WorkspaceRoots: exampleWorkspaceRoots(t),
		CatalogRoots:   exampleCatalogRoots(t),
		EventBus:       bus,
		Adapters: map[string]runtimepkg.Adapter{
			runtimepkg.ProviderDocker: adapter,
		},
		LookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
	})

	stream, cancel, err := service.SubscribeWorkspaceEvents(context.Background(), "shop-local", 8)
	if err != nil {
		t.Fatalf("SubscribeWorkspaceEvents returned error: %v", err)
	}
	defer cancel()

	chunks, err := service.WorkspaceLogs(context.Background(), "shop-local", "api", runtimepkg.LogsRequest{Tail: 10})
	if err != nil {
		t.Fatalf("WorkspaceLogs returned error: %v", err)
	}
	if got, want := len(chunks), 1; got != want {
		t.Fatalf("len(chunks) = %d, want %d", got, want)
	}

	result, err := service.ExecWorkspace(context.Background(), "shop-local", "api", runtimepkg.ExecRequest{Command: []string{"echo", "ok"}})
	if err != nil {
		t.Fatalf("ExecWorkspace returned error: %v", err)
	}
	if got, want := result.ExitCode, 0; got != want {
		t.Fatalf("result.ExitCode = %d, want %d", got, want)
	}

	var kinds []events.Kind
	for i := 0; i < 5; i++ {
		kinds = append(kinds, (<-stream).Kind)
	}
	if got, want := kinds, []events.Kind{events.KindLogsStarted, events.KindLogsChunk, events.KindLogsCompleted, events.KindExecStarted, events.KindExecCompleted}; !reflect.DeepEqual(got, want) {
		t.Fatalf("event kinds = %v, want %v", got, want)
	}
}

func TestServiceImportPreviewAndProjectScanUseSharedBoundary(t *testing.T) {
	service := newTestService(t, Config{})

	stackPath := filepath.Join(repoRoot(t), "examples", "v1", "stacks", "shop-export.yaml")
	preview, err := service.ImportV1Stack(context.Background(), stackPath)
	if err != nil {
		t.Fatalf("ImportV1Stack returned error: %v", err)
	}
	if got, want := preview.Status, "partial"; got != want {
		t.Fatalf("preview.Status = %q, want %q", got, want)
	}
	if got, want := len(preview.Artifacts), 1; got != want {
		t.Fatalf("len(preview.Artifacts) = %d, want %d", got, want)
	}

	libraryPath := filepath.Join(repoRoot(t), "examples", "v1", "library")
	preview, err = service.ImportV1Library(context.Background(), libraryPath)
	if err != nil {
		t.Fatalf("ImportV1Library returned error: %v", err)
	}
	if got, want := preview.Mode, "v1-library"; got != want {
		t.Fatalf("preview.Mode = %q, want %q", got, want)
	}
	if got, want := preview.Summary.Total, 2; got != want {
		t.Fatalf("preview.Summary.Total = %d, want %d", got, want)
	}

	projectRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectRoot, "package.json"), []byte(`{"name":"shop-api","dependencies":{"express":"^4.19.0"}}`), 0o644); err != nil {
		t.Fatalf("os.WriteFile(package.json): %v", err)
	}
	scan, err := service.ScanProject(context.Background(), projectRoot)
	if err != nil {
		t.Fatalf("ScanProject returned error: %v", err)
	}
	if got, want := scan.ProjectType, "node"; got != want {
		t.Fatalf("scan.ProjectType = %q, want %q", got, want)
	}
	if got, want := scan.SuggestedTemplates, []string{"node-api"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("scan.SuggestedTemplates = %v, want %v", got, want)
	}
}

type fakeAdapter struct {
	provider     string
	capabilities runtimepkg.AdapterCapabilities
	snapshot     *runtimepkg.Snapshot
	logChunks    []runtimepkg.LogChunk
	execResult   *runtimepkg.ExecResult
	inspectCalls int
	restartCalls int
}

func (f *fakeAdapter) Provider() string { return f.provider }

func (f *fakeAdapter) Capabilities() runtimepkg.AdapterCapabilities { return f.capabilities }

func (f *fakeAdapter) InspectWorkspace(context.Context, *runtimepkg.DesiredWorkspace) (*runtimepkg.Snapshot, error) {
	f.inspectCalls++
	if f.snapshot == nil {
		return &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Provider: f.provider}}, nil
	}
	return f.snapshot, nil
}

func (f *fakeAdapter) EnsureNetwork(context.Context, *runtimepkg.DesiredNetwork) error { return nil }

func (f *fakeAdapter) RemoveNetwork(context.Context, *runtimepkg.DesiredNetwork) error { return nil }

func (f *fakeAdapter) ApplyResource(context.Context, runtimepkg.ApplyResourceRequest) error {
	return nil
}

func (f *fakeAdapter) RemoveResource(context.Context, runtimepkg.ResourceRef) error { return nil }

func (f *fakeAdapter) RestartResource(context.Context, runtimepkg.ResourceRef) error {
	f.restartCalls++
	return nil
}

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

func newTestService(t *testing.T, config Config) *Service {
	t.Helper()
	service, err := New(config)
	if err != nil {
		t.Fatalf("appsvc.New returned error: %v", err)
	}
	return service
}

func writeWorkspaceCopy(t *testing.T, sourcePath, targetPath, name, displayName string) {
	t.Helper()
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("os.ReadFile(%s): %v", sourcePath, err)
	}
	content := strings.ReplaceAll(string(data), "shop-local", name)
	content = strings.ReplaceAll(content, "Shop Local", displayName)
	content = strings.ReplaceAll(content, "laravel-local", name)
	content = strings.ReplaceAll(content, "Laravel Local", displayName)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(targetPath), err)
	}
	if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%s): %v", targetPath, err)
	}
}

func exampleWorkspaceRoots(t *testing.T) []string {
	t.Helper()
	return []string{filepath.Join(repoRoot(t), "examples", "v2", "workspaces")}
}

func exampleCatalogRoots(t *testing.T) []string {
	t.Helper()
	return []string{filepath.Join(repoRoot(t), "catalog", "builtin")}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := stdruntime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestServiceWorkflowMethodsDelegate(t *testing.T) {
	runner := &fakeWorkflowRunner{results: []workflows.CommandResult{{Status: workflows.StatusPass}, {Status: workflows.StatusPass}, {Status: workflows.StatusPass}, {Status: workflows.StatusPass}, {Status: workflows.StatusPass}, {Status: workflows.StatusPass}, {Status: workflows.StatusPass}}}
	service := newTestService(t, Config{WorkspaceRoots: exampleWorkspaceRoots(t), CatalogRoots: exampleCatalogRoots(t), WorkflowRunner: runner})
	if _, err := service.Doctor(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RuntimeStatus(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SocketStatus(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SocketStart(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SocketStop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) == 0 {
		t.Fatal("workflow runner was not used")
	}
}

func TestRestartWorkspaceResourceDelegatesToRuntimeAdapter(t *testing.T) {
	adapter := &fakeAdapter{provider: runtimepkg.ProviderPodman, capabilities: runtimepkg.AdapterCapabilities{Inspect: true, Apply: true}}
	service := newTestService(t, Config{WorkspaceRoots: exampleWorkspaceRoots(t), CatalogRoots: exampleCatalogRoots(t), Adapters: map[string]runtimepkg.Adapter{runtimepkg.ProviderPodman: adapter}})
	if err := service.RestartWorkspaceResource(context.Background(), "shop-local", "postgres"); err != nil {
		t.Fatal(err)
	}
	if adapter.restartCalls != 1 {
		t.Fatalf("restartCalls = %d, want 1", adapter.restartCalls)
	}
}

type fakeWorkflowRunner struct {
	results []workflows.CommandResult
	calls   []workflows.CommandResult
}

func (f *fakeWorkflowRunner) Run(ctx context.Context, command string, args ...string) workflows.CommandResult {
	_ = ctx
	f.calls = append(f.calls, workflows.CommandResult{Command: command, Args: args})
	if len(f.results) == 0 {
		return workflows.CommandResult{Command: command, Args: args, Status: workflows.StatusFail, ExitCode: -1}
	}
	result := f.results[0]
	f.results = f.results[1:]
	return result
}

var _ runtimepkg.Adapter = (*fakeAdapter)(nil)
var _ = catalog.Template{}
var _ = planpkg.Result{}
