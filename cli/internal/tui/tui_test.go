package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/doctor"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/recipe"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
	"github.com/prospect-ogujiuba/devarch/cli/internal/state"
	"github.com/prospect-ogujiuba/devarch/cli/internal/testutil"
)

type harness struct {
	m     *Model
	fake  *runner.Fake
	saved state.Versions
	root  string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	root := t.TempDir()
	lib := filepath.Join(root, "services-library")
	testutil.WriteFiles(t, lib, map[string]string{
		"database/postgres/compose.yml": testutil.Compose("postgres", "x-devarch:\n  title: PostgreSQL\n  urls: [https://pg.test]\n  version: {var: POSTGRES_VERSION, default: \"18\", choices: [\"18\", \"17\"]}\n"),
		"database/redis/compose.yml":    testutil.Compose("redis", "x-devarch:\n  title: Redis\n  tags: [cache]\n"),
	})
	testutil.WriteFiles(t, root, map[string]string{
		"apps/shop/wp-config.php": "",
		"apps/web/package.json":   `{"dependencies": {"next": "15"}}`,
		"recipes/demo/recipe.yml": "name: demo\ntitle: Demo\nentry: scripts/demo.sh\nno_hosts_flag: [--no-hosts]\nargs:\n  - {name: name, positional: true, pattern: '^[a-z]+$'}\n  - {name: force, flag: --force, type: bool, confirm: 'Replace it?'}\n",
		"scripts/demo.sh":         "#!/bin/sh\nprintf '{\"step\":\"scaffold\",\"state\":\"start\"}\\n' >&3\necho \"demo $*\"\nprintf '{\"step\":\"scaffold\",\"state\":\"done\"}\\n' >&3\n",
	})
	os.Chmod(filepath.Join(root, "scripts/demo.sh"), 0o755)
	c, _ := catalog.Load([]catalog.Path{{Dir: lib, Source: "builtin"}})
	ps := `[{"Names":["postgres"],"Image":"postgres:18","State":"running","Status":"Up 1 hour (healthy)",
	  "Labels":{"com.docker.compose.project":"postgres","com.docker.compose.project.working_dir":"` + lib + `/database/postgres"},
	  "Ports":[{"host_ip":"127.0.0.1","host_port":8502,"container_port":5432}]}]`
	f := &runner.Fake{Responses: map[string]runner.Response{"podman ps": {Out: []byte(ps)}}, RunErrors: map[string][]error{}}
	h := &harness{fake: f, root: root, saved: state.Versions{}}
	eng := &engine.Engine{Root: root, Catalog: c, Runner: f, Versions: state.Versions{},
		SaveVersions: func(v state.Versions) error {
			for k, val := range v {
				h.saved[k] = val
			}
			return nil
		}}
	h.m = New(Deps{
		Engine:  eng,
		Recipes: func() ([]recipe.Recipe, error) { return recipe.Load(root, filepath.Join(root, "recipes")) },
		Doctor: func(context.Context, *engine.Engine) []doctor.Check {
			return []doctor.Check{{Name: "podman", Status: doctor.OK, Detail: "podman version 6"}, {Name: "hosts", Status: doctor.Warn, Detail: "out of date", Fix: "devarch hosts sync"}}
		},
		Lint: func(context.Context) (int, int) { return 0, 2 },
	})
	// The TUI's engine shares the fake runner so tests see every command.
	h.m.eng.Runner = f
	h.m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	h.m.Update(h.m.loadContainers()())
	return h
}

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// press sends keys and returns the last command.
func (h *harness) press(keys ...string) tea.Cmd {
	var cmd tea.Cmd
	for _, k := range keys {
		_, cmd = h.m.Update(key(k))
	}
	return cmd
}

// finish runs an action command and feeds its result back.
func (h *harness) finish(t *testing.T, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected a command")
	}
	msg := cmd()
	if _, ok := msg.(actionDoneMsg); !ok {
		t.Fatalf("expected actionDoneMsg, got %T", msg)
	}
	h.m.Update(msg)
}

func (h *harness) commands() string { return strings.Join(h.fake.Lines(), "\n") }

func TestServicesViewShowsStateAndDetail(t *testing.T) {
	h := newHarness(t)
	v := h.m.View()
	for _, want := range []string{"1 Services", "database/postgres", "running", "https://pg.test", "PostgreSQL", "version 18 (POSTGRES_VERSION; choices 18 17)", "containers: postgres"} {
		if !strings.Contains(v, want) {
			t.Errorf("view missing %q:\n%s", want, v)
		}
	}
}

func TestFilter(t *testing.T) {
	h := newHarness(t)
	h.press("/", "c", "a", "c", "h", "e", "enter")
	if len(h.m.services) != 1 || h.m.services[0].ID != "database/redis" {
		t.Fatalf("filtered = %v", h.m.services)
	}
	h.press("esc")
	if len(h.m.services) != 2 {
		t.Fatal("esc should clear the filter")
	}
}

func TestUpDownRestartRunNativeCommands(t *testing.T) {
	h := newHarness(t)
	h.press("down") // redis
	h.fake.Responses["podman inspect"] = runner.Response{Out: []byte(`[{"Name":"redis","State":{"Status":"running"},"Config":{}}]`)}
	h.fake.Responses["podman ps"] = runner.Response{Out: []byte(`[{"Names":["redis"],"State":"running","Status":"Up","Labels":{"com.docker.compose.project":"redis","com.docker.compose.project.working_dir":"` + filepath.Join(h.root, "services-library/database/redis") + `"}}]`)}
	h.finish(t, h.press("u"))
	if h.m.busy != "" || !strings.Contains(h.m.status, "start redis: done") {
		t.Fatalf("busy=%q status=%q", h.m.busy, h.m.status)
	}
	h.finish(t, h.press("r"))
	h.finish(t, h.press("d"))
	cmds := h.commands()
	for _, want := range []string{"redis && podman compose up -d)", "redis && podman compose restart)", "redis && podman compose down)"} {
		if !strings.Contains(cmds, want) {
			t.Errorf("missing %q in\n%s", want, cmds)
		}
	}
}

func TestOneActionAtATime(t *testing.T) {
	h := newHarness(t)
	if h.press("u") == nil {
		t.Fatal("first action should start")
	}
	if h.press("d") != nil || !strings.Contains(h.m.status, "busy") {
		t.Fatalf("second action should be refused while busy: %q", h.m.status)
	}
}

func TestPasteDoesNotTriggerCommands(t *testing.T) {
	h := newHarness(t)
	if cmd := h.press("udr"); cmd != nil || h.m.busy != "" {
		t.Fatal("a multi-character burst triggered a command")
	}
}

func TestVersionPickerRecreatesRunningServiceAfterConfirm(t *testing.T) {
	h := newHarness(t) // postgres is running
	h.press("v")
	if h.m.mode != modeVersion || !strings.Contains(h.m.View(), "PostgreSQL version") {
		t.Fatal("version picker did not open")
	}
	h.press("down", "enter")
	if h.m.mode != modeConfirm || !strings.Contains(h.m.View(), "Recreate postgres with 17 now?") {
		t.Fatalf("expected a recreate confirmation:\n%s", h.m.View())
	}
	h.finish(t, h.press("y"))
	if h.saved["POSTGRES_VERSION"] != "17" || !strings.Contains(h.commands(), "POSTGRES_VERSION=17 podman compose up -d --force-recreate") {
		t.Fatalf("saved=%v\n%s", h.saved, h.commands())
	}
}

func TestVersionPickerWithoutRecreate(t *testing.T) {
	h := newHarness(t)
	h.press("v", "down", "enter")
	h.finish(t, h.press("n"))
	if h.saved["POSTGRES_VERSION"] != "17" || strings.Contains(h.commands(), "--force-recreate") {
		t.Fatalf("saved=%v\n%s", h.saved, h.commands())
	}
}

func TestViewsAppsRunningDoctor(t *testing.T) {
	h := newHarness(t)
	h.press("2")
	v := h.m.View()
	if !strings.Contains(v, "shop") || !strings.Contains(v, "WordPress") || !strings.Contains(v, "Next.js") || !strings.Contains(v, "https://web.test") {
		t.Fatalf("apps view:\n%s", v)
	}
	h.press("3")
	if v := h.m.View(); !strings.Contains(v, "postgres") || !strings.Contains(v, "8502→5432") {
		t.Fatalf("running view:\n%s", v)
	}
	cmd := h.press("4")
	h.m.Update(cmd())
	v = h.m.View()
	for _, want := range []string{"podman version 6", "→ devarch hosts sync", "0 errors, 2 warnings"} {
		if !strings.Contains(v, want) {
			t.Errorf("doctor view missing %q:\n%s", want, v)
		}
	}
}

func TestWizardRunsRecipeWithProgress(t *testing.T) {
	h := newHarness(t)
	h.press("n")
	if h.m.mode != modeWizard {
		t.Fatalf("wizard did not open: %s", h.m.status)
	}
	h.press("enter")                         // choose demo
	h.press("B", "a", "d", "enter", "enter") // name, then past force
	if !strings.Contains(h.m.View(), "must match") {
		t.Fatalf("invalid name should be reported:\n%s", h.m.View())
	}
	h.press("up")
	for range 3 {
		h.m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	h.press("s", "h", "o", "p", "down", "space", "enter")
	v := h.m.View()
	if !strings.Contains(v, "scripts/demo.sh shop --force --no-hosts") || !strings.Contains(v, "Replace it?") {
		t.Fatalf("confirm view:\n%s", v)
	}
	if h.press("enter"); !strings.Contains(h.m.View(), "press y to confirm") {
		t.Fatal("a destructive flag must need y")
	}
	cmd := h.press("y")
	if cmd == nil || h.m.wizard.step != stepRun {
		t.Fatal("run did not start")
	}
	// Drive the batch: the run command and the progress reader.
	batch := cmd().(tea.BatchMsg)
	done := make(chan tea.Msg, 1)
	go func() { done <- batch[0]() }()
	readProgress := batch[1]
	for {
		msg := readProgress()
		if msg == nil {
			break
		}
		_, next := h.m.Update(msg)
		readProgress = next
	}
	h.m.Update(<-done)
	v = h.m.View()
	if !strings.Contains(v, "✓ scaffold") || !strings.Contains(v, "Ready: https://shop.test") {
		t.Fatalf("done view:\n%s", v)
	}
	if h.m.busy != "" {
		t.Fatal("busy flag not cleared")
	}
	h.press("esc")
	if h.m.mode != modeNormal || h.m.wizard != nil {
		t.Fatal("wizard did not close")
	}
}

func TestLineWriterSplitsLines(t *testing.T) {
	ch := make(chan string, 10)
	w := &lineWriter{ch: ch}
	w.Write([]byte("one\ntw"))
	w.Write([]byte("o\r\nthree"))
	if got := []string{<-ch, <-ch}; got[0] != "one" || got[1] != "two" || len(ch) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestHelpAndQuit(t *testing.T) {
	h := newHarness(t)
	h.press("?")
	if !strings.Contains(h.m.View(), "DevArch keys") {
		t.Fatal("help not shown")
	}
	h.press("x")
	if h.m.mode != modeNormal {
		t.Fatal("any key should close help")
	}
	if cmd := h.press("q"); cmd == nil || cmd() != tea.Quit() {
		t.Fatal("q should quit")
	}
}

func TestRefreshesAreCoalesced(t *testing.T) {
	h := newHarness(t)
	first := h.m.loadContainers()
	if first == nil || h.m.loadContainers() != nil || h.m.loadContainers() != nil {
		t.Fatal("only one refresh may run at a time")
	}
	if _, next := h.m.Update(first()); next == nil {
		t.Fatal("a queued refresh should run once the first finishes")
	} else if _, after := h.m.Update(next()); after != nil {
		t.Fatal("only one refresh should have been queued")
	}
}
