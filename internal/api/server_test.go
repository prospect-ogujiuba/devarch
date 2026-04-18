package api

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	stdruntime "runtime"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/prospect-ogujiuba/devarch/internal/appsvc"
	"github.com/prospect-ogujiuba/devarch/internal/events"
	planpkg "github.com/prospect-ogujiuba/devarch/internal/plan"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

func TestReadRoutes(t *testing.T) {
	server, _, _, _ := newTestHTTPServer(t)
	defer server.Close()

	var templates []appsvc.TemplateSummary
	getJSON(t, server.URL+"/api/catalog/templates", http.StatusOK, &templates)
	if len(templates) == 0 {
		t.Fatal("expected catalog templates")
	}

	var templateDetail map[string]any
	getJSON(t, server.URL+"/api/catalog/templates/postgres", http.StatusOK, &templateDetail)
	if _, ok := templateDetail["path"]; ok {
		t.Fatalf("catalog template detail unexpectedly exposed path: %#v", templateDetail)
	}
	if got, want := templateDetail["name"], any("postgres"); got != want {
		t.Fatalf("templateDetail[name] = %#v, want %#v", got, want)
	}

	var missing errorEnvelope
	getJSON(t, server.URL+"/api/catalog/templates/missing", http.StatusNotFound, &missing)
	if got, want := missing.Error.Code, "not_found"; got != want {
		t.Fatalf("missing template error code = %q, want %q", got, want)
	}

	var workspaces []appsvc.WorkspaceSummary
	getJSON(t, server.URL+"/api/workspaces", http.StatusOK, &workspaces)
	if len(workspaces) < 3 {
		t.Fatalf("len(workspaces) = %d, want at least 3", len(workspaces))
	}
	if got, want := workspaces[0].Name, "compat-local"; got != want {
		t.Fatalf("workspaces[0].Name = %q, want %q", got, want)
	}

	var workspaceDetail appsvc.WorkspaceDetail
	getJSON(t, server.URL+"/api/workspaces/shop-local", http.StatusOK, &workspaceDetail)
	if got, want := workspaceDetail.Provider, runtimepkg.ProviderDocker; got != want {
		t.Fatalf("workspaceDetail.Provider = %q, want %q", got, want)
	}
	if len(workspaceDetail.ResourceKeys) != 4 {
		t.Fatalf("len(workspaceDetail.ResourceKeys) = %d, want 4", len(workspaceDetail.ResourceKeys))
	}

	var manifest map[string]any
	getJSON(t, server.URL+"/api/workspaces/shop-local/manifest", http.StatusOK, &manifest)
	if got, want := manifest["kind"], any("Workspace"); got != want {
		t.Fatalf("manifest[kind] = %#v, want %#v", got, want)
	}

	var graph map[string]any
	getJSON(t, server.URL+"/api/workspaces/shop-local/graph", http.StatusOK, &graph)
	if _, ok := graph["graph"]; !ok {
		t.Fatalf("graph response missing graph key: %#v", graph)
	}
	if _, ok := graph["contracts"]; !ok {
		t.Fatalf("graph response missing contracts key: %#v", graph)
	}

	var status struct {
		Desired  map[string]any `json:"desired"`
		Snapshot map[string]any `json:"snapshot"`
	}
	getJSON(t, server.URL+"/api/workspaces/shop-local/status", http.StatusOK, &status)
	if got, want := status.Desired["provider"], any(runtimepkg.ProviderDocker); got != want {
		t.Fatalf("status.desired.provider = %#v, want %#v", got, want)
	}
	if status.Snapshot == nil {
		t.Fatal("expected status snapshot")
	}

	var plan planpkg.Result
	getJSON(t, server.URL+"/api/workspaces/shop-local/plan", http.StatusOK, &plan)
	if got, want := plan.Workspace, "shop-local"; got != want {
		t.Fatalf("plan.Workspace = %q, want %q", got, want)
	}
	if len(plan.Actions) == 0 {
		t.Fatal("expected plan actions")
	}
}

func TestOperationalRoutes(t *testing.T) {
	server, bus, adapter, _ := newTestHTTPServer(t)
	defer server.Close()

	var applyErr errorEnvelope
	getJSONMethod(t, http.MethodPost, server.URL+"/api/workspaces/shop-local/apply", http.StatusConflict, &applyErr)
	if got, want := applyErr.Error.Code, "unsupported_capability"; got != want {
		t.Fatalf("apply error code = %q, want %q", got, want)
	}
	if got, want := adapter.inspectCalls > 0, true; got != want {
		t.Fatalf("adapter.inspectCalls > 0 = %v, want %v", got, want)
	}

	var logsErr errorEnvelope
	getJSON(t, server.URL+"/api/workspaces/shop-local/logs", http.StatusBadRequest, &logsErr)
	if got, want := logsErr.Error.Code, "bad_request"; got != want {
		t.Fatalf("logs error code = %q, want %q", got, want)
	}

	var chunks []runtimepkg.LogChunk
	getJSON(t, server.URL+"/api/workspaces/shop-local/logs?resource=api&tail=5", http.StatusOK, &chunks)
	if got, want := len(chunks), 1; got != want {
		t.Fatalf("len(chunks) = %d, want %d", got, want)
	}
	if got, want := chunks[0].Line, "ready"; got != want {
		t.Fatalf("chunks[0].Line = %q, want %q", got, want)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/workspaces/shop-local/events", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("events request returned error: %v", err)
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("ReadString connected comment returned error: %v", err)
	}
	if !strings.HasPrefix(line, ": connected") {
		t.Fatalf("first SSE line = %q, want connected comment", line)
	}
	if _, err := reader.ReadString('\n'); err != nil {
		t.Fatalf("ReadString comment separator returned error: %v", err)
	}
	if _, err := bus.Publish(events.ApplyStarted("shop-local", 1)); err != nil {
		t.Fatalf("bus.Publish returned error: %v", err)
	}
	var dataLine string
	for dataLine == "" {
		line, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("ReadString SSE payload returned error: %v", err)
		}
		if strings.HasPrefix(line, "data: ") {
			dataLine = strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		}
	}
	cancel()
	var envelope events.Envelope
	if err := json.Unmarshal([]byte(dataLine), &envelope); err != nil {
		t.Fatalf("json.Unmarshal SSE envelope returned error: %v", err)
	}
	if got, want := envelope.Kind, events.KindApplyStarted; got != want {
		t.Fatalf("envelope.Kind = %q, want %q", got, want)
	}

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/api/workspaces/shop-local/resources/api/exec"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial returned error: %v", err)
	}
	if err := conn.WriteJSON(map[string]any{"type": "request", "command": []string{"echo", "ok"}}); err != nil {
		t.Fatalf("conn.WriteJSON returned error: %v", err)
	}
	var execResult map[string]any
	if err := conn.ReadJSON(&execResult); err != nil {
		t.Fatalf("conn.ReadJSON result returned error: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("conn.Close returned error: %v", err)
	}
	if got, want := execResult["type"], any("result"); got != want {
		t.Fatalf("exec result type = %#v, want %#v", got, want)
	}

	conn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial interactive returned error: %v", err)
	}
	if err := conn.WriteJSON(map[string]any{"type": "request", "command": []string{"echo", "ok"}, "interactive": true}); err != nil {
		t.Fatalf("conn.WriteJSON interactive returned error: %v", err)
	}
	var execError map[string]any
	if err := conn.ReadJSON(&execError); err != nil {
		t.Fatalf("conn.ReadJSON error frame returned error: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("conn.Close interactive returned error: %v", err)
	}
	if got, want := execError["type"], any("error"); got != want {
		t.Fatalf("exec error type = %#v, want %#v", got, want)
	}
}

type errorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

type fakeAdapter struct {
	provider     string
	capabilities runtimepkg.AdapterCapabilities
	snapshot     *runtimepkg.Snapshot
	logChunks    []runtimepkg.LogChunk
	execResult   *runtimepkg.ExecResult
	inspectCalls int
}

func (f *fakeAdapter) Provider() string { return f.provider }

func (f *fakeAdapter) Capabilities() runtimepkg.AdapterCapabilities { return f.capabilities }

func (f *fakeAdapter) InspectWorkspace(context.Context, *runtimepkg.DesiredWorkspace) (*runtimepkg.Snapshot, error) {
	f.inspectCalls++
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

func newTestHTTPServer(t *testing.T) (*httptest.Server, *events.Bus, *fakeAdapter, *appsvc.Service) {
	t.Helper()
	bus := events.NewBus()
	adapter := &fakeAdapter{
		provider:     runtimepkg.ProviderDocker,
		capabilities: runtimepkg.AdapterCapabilities{Inspect: true, Logs: true, Exec: true},
		snapshot:     &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: "shop-local", Provider: runtimepkg.ProviderDocker}},
		logChunks:    []runtimepkg.LogChunk{{Line: "ready", Stream: "combined"}},
		execResult:   &runtimepkg.ExecResult{ExitCode: 0, Stdout: "ok\n"},
	}
	service, err := appsvc.New(appsvc.Config{
		WorkspaceRoots: []string{filepath.Join(repoRoot(t), "examples", "v2", "workspaces")},
		CatalogRoots:   []string{filepath.Join(repoRoot(t), "catalog", "builtin")},
		EventBus:       bus,
		Adapters: map[string]runtimepkg.Adapter{
			runtimepkg.ProviderDocker: adapter,
		},
		LookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
	})
	if err != nil {
		t.Fatalf("appsvc.New returned error: %v", err)
	}
	server := httptest.NewServer(NewServer(service))
	return server, bus, adapter, service
}

func getJSON(t *testing.T, url string, wantStatus int, target any) {
	t.Helper()
	getJSONMethod(t, http.MethodGet, url, wantStatus, target)
}

func getJSONMethod(t *testing.T, method, url string, wantStatus int, target any) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("http.NewRequest returned error: %v", err)
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do(%s) returned error: %v", url, err)
	}
	defer resp.Body.Close()
	if got, want := resp.StatusCode, wantStatus; got != want {
		t.Fatalf("%s %s status = %d, want %d", method, url, got, want)
	}
	if target == nil {
		return
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("json decode for %s returned error: %v", url, err)
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

var _ runtimepkg.Adapter = (*fakeAdapter)(nil)
