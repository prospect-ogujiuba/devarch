// Package tui is the interactive DevArch screen: the catalog with live state,
// apps, running containers, doctor, logs, version switching, and the new
// project wizard. Every action calls the same engine functions as the CLI and
// shows the native command it runs; the TUI never edits compose files.
package tui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/doctor"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/hosts"
	"github.com/prospect-ogujiuba/devarch/cli/internal/recipe"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

// Deps is what the TUI needs from the CLI.
type Deps struct {
	Engine    *engine.Engine
	Recipes   func() ([]recipe.Recipe, error)
	RecipeEnv []string
	Doctor    func(ctx context.Context, e *engine.Engine) []doctor.Check
	// Lint returns catalog lint error and warning counts for the Doctor view.
	Lint func(ctx context.Context) (int, int)
	// Exe is the devarch binary, used for interactive `hosts sync`.
	Exe string
	// OpenURL opens a URL in the browser; OpenDir opens a directory in an editor.
	OpenURL func(string) error
	OpenDir func(string) error
	// Events streams container events; nil disables live updates.
	Events func(ctx context.Context) <-chan struct{}
}

type view int

const (
	viewServices view = iota
	viewApps
	viewRunning
	viewDoctor
)

var viewNames = []string{"Services", "Apps", "Running", "Doctor"}

type mode int

const (
	modeNormal mode = iota
	modeFilter
	modeVersion
	modeConfirm
	modeLogs
	modeWizard
	modeHelp
)

type (
	containersMsg struct {
		cs  []engine.Container
		err error
	}
	actionDoneMsg struct {
		label string
		err   error
		after func(*Model) tea.Cmd
	}
	containerEventMsg struct{}
	eventsClosedMsg   struct{}
	refreshTickMsg    struct{}
	doctorMsg         struct {
		checks           []doctor.Check
		lintErr, lintWrn int
	}
	hostsSyncedMsg struct{ err error }
)

// Model is the root Bubble Tea model.
type Model struct {
	deps   Deps
	eng    *engine.Engine
	ctx    context.Context
	cancel context.CancelFunc
	out    chan string
	events <-chan struct{}

	width, height int
	view          view
	mode          mode
	cursor        [4]int

	filter     textinput.Model
	services   []catalog.Service
	containers []engine.Container
	states     map[string]engine.ServiceState
	psErr      error
	apps       []appInfo

	activity []string
	busy     string
	spin     spinner.Model
	status   string
	statusOK bool

	version struct {
		svc     catalog.Service
		choices []string
		cursor  int
	}
	confirm struct {
		prompt  string
		yes, no func() tea.Cmd
	}
	checks         []doctor.Check
	lintErr        int
	lintWrn        int
	doctorLoaded   bool
	refreshPending bool
	loading        bool
	reloadQueued   bool

	logs   *logsModel
	wizard *wizardModel
}

// New builds the model. The engine is copied so command output is captured
// by the TUI instead of written to the terminal.
func New(deps Deps) *Model {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan string, 4096)
	w := &lineWriter{ch: out}
	eng := *deps.Engine
	eng.Runner = captureRunner{inner: runner.System{}, out: w}
	eng.Log = w
	eng.DryRun = false

	f := textinput.New()
	f.Prompt = "/"
	f.Placeholder = "filter by name, title, tag"
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot

	m := &Model{deps: deps, eng: &eng, ctx: ctx, cancel: cancel, out: out, filter: f, spin: sp,
		states: map[string]engine.ServiceState{}}
	m.services = eng.Catalog.Services
	m.apps = discoverApps(filepath.Join(eng.Root, "apps"))
	if deps.Events != nil {
		m.events = deps.Events(ctx)
	}
	return m
}

// Run starts the TUI on the alternate screen.
func Run(deps Deps) error {
	m := New(deps)
	defer m.cancel()
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.loadContainers(), waitOutput(m.out), m.spin.Tick}
	if m.events != nil {
		cmds = append(cmds, waitEvent(m.events))
	}
	return tea.Batch(cmds...)
}

func waitEvent(ch <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		if _, ok := <-ch; !ok {
			return eventsClosedMsg{}
		}
		return containerEventMsg{}
	}
}

// loadContainers refreshes container state. podman ps can take seconds, so
// at most one refresh runs and at most one more is queued behind it.
func (m *Model) loadContainers() tea.Cmd {
	if m.loading {
		m.reloadQueued = true
		return nil
	}
	m.loading = true
	return func() tea.Msg {
		cs, err := m.eng.Containers(m.ctx)
		return containersMsg{cs, err}
	}
}

func (m *Model) addActivity(line string) {
	m.activity = append(m.activity, line)
	if len(m.activity) > 1000 {
		m.activity = m.activity[len(m.activity)-1000:]
	}
}

func (m *Model) setStatus(ok bool, format string, args ...any) {
	m.status, m.statusOK = fmt.Sprintf(format, args...), ok
}

// runAction runs fn in the background; only one action runs at a time.
func (m *Model) runAction(label string, fn func(ctx context.Context) error, after func(*Model) tea.Cmd) tea.Cmd {
	if m.busy != "" {
		m.setStatus(false, "busy: %s", m.busy)
		return nil
	}
	m.busy = label
	m.status = ""
	m.addActivity("· " + label)
	return func() tea.Msg {
		return actionDoneMsg{label: label, err: fn(m.ctx), after: after}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.logs != nil {
			m.logs.resize(m.width, m.height)
		}
		return m, nil
	case outputMsg:
		m.addActivity(string(msg))
		return m, waitOutput(m.out)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	case containersMsg:
		m.loading = false
		m.containers, m.psErr = msg.cs, msg.err
		if msg.err == nil {
			m.states = engine.States(m.eng.Catalog.Services, msg.cs)
		}
		if m.reloadQueued {
			m.reloadQueued = false
			return m, m.loadContainers()
		}
		return m, nil
	case containerEventMsg:
		cmds := []tea.Cmd{waitEvent(m.events)}
		if !m.refreshPending {
			m.refreshPending = true
			cmds = append(cmds, tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg { return refreshTickMsg{} }))
		}
		return m, tea.Batch(cmds...)
	case eventsClosedMsg:
		m.events = nil
		return m, nil
	case refreshTickMsg:
		m.refreshPending = false
		return m, m.loadContainers()
	case actionDoneMsg:
		m.busy = ""
		if msg.err != nil {
			m.setStatus(false, "%s failed: %v", msg.label, msg.err)
			m.addActivity("✗ " + msg.err.Error())
		} else {
			m.setStatus(true, "%s: done", msg.label)
			m.addActivity("✓ " + msg.label)
		}
		cmds := []tea.Cmd{m.loadContainers()}
		if msg.after != nil && msg.err == nil {
			cmds = append(cmds, msg.after(m))
		}
		return m, tea.Batch(cmds...)
	case doctorMsg:
		m.checks, m.lintErr, m.lintWrn, m.doctorLoaded = msg.checks, msg.lintErr, msg.lintWrn, true
		m.busy = ""
		return m, nil
	case hostsSyncedMsg:
		if msg.err != nil {
			m.setStatus(false, "hosts sync failed: %v", msg.err)
		} else {
			m.setStatus(true, "hosts synchronized")
		}
		return m, nil
	case logLineMsg, logEndMsg:
		if m.logs != nil {
			return m, m.logs.update(msg)
		}
		return m, nil
	case progressMsg, wizardDoneMsg:
		if m.wizard != nil {
			return m, m.wizard.update(m, msg)
		}
		return m, nil
	case tea.KeyMsg:
		return m, m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(k tea.KeyMsg) tea.Cmd {
	key := k.String()
	if key == "ctrl+c" {
		m.cancel()
		return tea.Quit
	}
	switch m.mode {
	case modeLogs:
		if m.logs.handleKey(k) {
			m.logs.stop()
			m.logs, m.mode = nil, modeNormal
		}
		return nil
	case modeWizard:
		cmd, closed := m.wizard.handleKey(m, k)
		if closed {
			m.wizard, m.mode = nil, modeNormal
		}
		return cmd
	case modeHelp:
		m.mode = modeNormal
		return nil
	case modeFilter:
		switch key {
		case "esc":
			m.filter.SetValue("")
			m.filter.Blur()
			m.mode = modeNormal
		case "enter":
			m.filter.Blur()
			m.mode = modeNormal
		default:
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(k)
			m.applyFilter()
			return cmd
		}
		m.applyFilter()
		return nil
	case modeVersion:
		return m.handleVersionKey(key)
	case modeConfirm:
		switch key {
		case "y", "Y":
			m.mode = modeNormal
			return m.confirm.yes()
		case "n", "N", "esc":
			m.mode = modeNormal
			if m.confirm.no != nil {
				return m.confirm.no()
			}
		}
		return nil
	}

	// A burst of several characters is pasted text, not a command.
	if k.Type == tea.KeyRunes && len(k.Runes) > 1 {
		return nil
	}
	switch key {
	case "q":
		m.cancel()
		return tea.Quit
	case "?":
		m.mode = modeHelp
		return nil
	case "tab", "right":
		return m.switchView((m.view + 1) % 4)
	case "shift+tab", "left":
		return m.switchView((m.view + 3) % 4)
	case "1", "2", "3", "4":
		return m.switchView(view(key[0] - '1'))
	case "up", "k":
		m.move(-1)
		return nil
	case "down", "j":
		m.move(1)
		return nil
	case "pgup":
		m.move(-10)
		return nil
	case "pgdown":
		m.move(10)
		return nil
	case "ctrl+r":
		m.apps = discoverApps(filepath.Join(m.eng.Root, "apps"))
		return m.loadContainers()
	case "n":
		return m.openWizard()
	case "H":
		return m.syncHosts()
	}
	switch m.view {
	case viewServices:
		return m.handleServicesKey(key)
	case viewApps:
		return m.handleAppsKey(key)
	case viewRunning:
		return m.handleRunningKey(key)
	case viewDoctor:
		if key == "r" {
			return m.runDoctor()
		}
	}
	return nil
}

func (m *Model) switchView(v view) tea.Cmd {
	m.view = v
	if v == viewDoctor && !m.doctorLoaded {
		return m.runDoctor()
	}
	return nil
}

func (m *Model) listLen() int {
	switch m.view {
	case viewServices:
		return len(m.services)
	case viewApps:
		return len(m.apps)
	case viewRunning:
		return len(m.running())
	case viewDoctor:
		return len(m.checks)
	}
	return 0
}

func (m *Model) move(delta int) {
	n := m.listLen()
	c := m.cursor[m.view] + delta
	if c >= n {
		c = n - 1
	}
	if c < 0 {
		c = 0
	}
	m.cursor[m.view] = c
}

func (m *Model) applyFilter() {
	m.services = m.eng.Catalog.Search(strings.TrimSpace(m.filter.Value()), "")
	m.move(0)
}

func (m *Model) selectedService() (catalog.Service, bool) {
	if c := m.cursor[viewServices]; c < len(m.services) {
		return m.services[c], true
	}
	return catalog.Service{}, false
}

func (m *Model) handleServicesKey(key string) tea.Cmd {
	if key == "/" {
		m.mode = modeFilter
		return m.filter.Focus()
	}
	if key == "esc" && m.filter.Value() != "" {
		m.filter.SetValue("")
		m.applyFilter()
		return nil
	}
	svc, ok := m.selectedService()
	if !ok {
		return nil
	}
	switch key {
	case "u":
		return m.runAction("start "+svc.Name, func(ctx context.Context) error {
			return m.eng.Up(ctx, []catalog.Service{svc}, engine.UpOptions{Wait: true, Timeout: 2 * time.Minute, NoHosts: true})
		}, func(m *Model) tea.Cmd { m.noteMissingHosts(svc); return nil })
	case "d":
		return m.runAction("stop "+svc.Name, func(ctx context.Context) error {
			return m.eng.Down(ctx, []catalog.Service{svc}, false)
		}, nil)
	case "r":
		return m.runAction("restart "+svc.Name, func(ctx context.Context) error {
			return m.eng.Restart(ctx, []catalog.Service{svc})
		}, nil)
	case "v":
		v := svc.Meta.Version
		if v == nil {
			m.setStatus(false, "%s has no switchable version", svc.ID)
			return nil
		}
		m.version.svc, m.version.choices, m.version.cursor = svc, v.Choices, 0
		if len(m.version.choices) == 0 {
			m.version.choices = []string{v.Default}
		}
		current := m.eng.SelectedVersion(svc)
		for i, c := range m.version.choices {
			if c == current {
				m.version.cursor = i
			}
		}
		m.mode = modeVersion
	case "l", "enter":
		if key == "enter" && m.states[svc.ID].State == "absent" {
			return nil
		}
		return m.openLogs("Logs: "+svc.ID, m.eng.LogsCmd(svc, true, "200"))
	case "o":
		if len(svc.Meta.URLs) == 0 {
			m.setStatus(false, "%s has no URL in its x-devarch metadata", svc.ID)
			return nil
		}
		return m.open(svc.Meta.URLs[0])
	}
	return nil
}

func (m *Model) handleVersionKey(key string) tea.Cmd {
	switch key {
	case "esc", "q":
		m.mode = modeNormal
	case "up", "k":
		if m.version.cursor > 0 {
			m.version.cursor--
		}
	case "down", "j":
		if m.version.cursor < len(m.version.choices)-1 {
			m.version.cursor++
		}
	case "enter":
		svc, choice := m.version.svc, m.version.choices[m.version.cursor]
		m.mode = modeNormal
		if choice == m.eng.SelectedVersion(svc) {
			m.setStatus(true, "%s already uses %s", svc.Name, choice)
			return nil
		}
		use := func(noRecreate bool) func() tea.Cmd {
			return func() tea.Cmd {
				return m.runAction(fmt.Sprintf("switch %s to %s", svc.Name, choice), func(ctx context.Context) error {
					_, err := m.eng.Use(ctx, svc, choice, engine.UseOptions{NoRecreate: noRecreate})
					return err
				}, nil)
			}
		}
		if st := m.states[svc.ID].State; st == "running" || st == "partial" {
			m.askConfirm(fmt.Sprintf("Recreate %s with %s now? (y/n)", svc.Name, choice), use(false), use(true))
			return nil
		}
		return use(false)()
	}
	return nil
}

func (m *Model) askConfirm(prompt string, yes, no func() tea.Cmd) {
	m.confirm.prompt, m.confirm.yes, m.confirm.no = prompt, yes, no
	m.mode = modeConfirm
}

// noteMissingHosts tells the user when a started service's hostnames are not
// mapped yet. Registration may need sudo or UAC, so it runs interactively (H).
func (m *Model) noteMissingHosts(svc catalog.Service) {
	if m.eng.Hosts == nil {
		return
	}
	content, err := m.eng.Hosts.Read()
	if err != nil {
		return
	}
	mapped := hosts.Mapped(content)
	var missing []string
	for _, l := range engine.HostLabels(svc) {
		if _, ok := mapped[l+".test"]; !ok {
			missing = append(missing, l+".test")
		}
	}
	if len(missing) > 0 {
		m.setStatus(false, "%s not in the hosts file; press H to register", strings.Join(missing, ", "))
	}
}

func (m *Model) syncHosts() tea.Cmd {
	if m.deps.Exe == "" {
		return nil
	}
	m.addActivity("$ devarch hosts sync")
	return tea.ExecProcess(exec.Command(m.deps.Exe, "hosts", "sync"), func(err error) tea.Msg { return hostsSyncedMsg{err} })
}

func (m *Model) open(url string) tea.Cmd {
	if m.deps.OpenURL == nil {
		return nil
	}
	if err := m.deps.OpenURL(url); err != nil {
		m.setStatus(false, "open %s: %v", url, err)
	} else {
		m.setStatus(true, "opened %s", url)
	}
	return nil
}

func (m *Model) running() []engine.Container {
	var out []engine.Container
	for _, c := range m.containers {
		if c.State == "running" {
			out = append(out, c)
		}
	}
	return out
}

func (m *Model) handleRunningKey(key string) tea.Cmd {
	rs := m.running()
	c := m.cursor[viewRunning]
	if c >= len(rs) {
		return nil
	}
	ct := rs[c]
	switch key {
	case "l", "enter":
		return m.openLogs("Logs: "+ct.Name, runner.Cmd{Name: "podman", Args: []string{"logs", "-f", "--tail", "200", ct.Name}})
	case "r":
		return m.runAction("restart "+ct.Name, func(ctx context.Context) error {
			return m.eng.Runner.Run(ctx, runner.Cmd{Name: "podman", Args: []string{"restart", ct.Name}})
		}, nil)
	}
	return nil
}

func (m *Model) runDoctor() tea.Cmd {
	if m.deps.Doctor == nil {
		return nil
	}
	m.busy = "doctor"
	return func() tea.Msg {
		checks := m.deps.Doctor(m.ctx, m.eng)
		var e, w int
		if m.deps.Lint != nil {
			e, w = m.deps.Lint(m.ctx)
		}
		return doctorMsg{checks, e, w}
	}
}

// appInfo is one apps/ directory.
type appInfo struct {
	Name, Kind, Dir string
	Routable        bool
}

func discoverApps(dir string) []appInfo {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []appInfo
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		d := filepath.Join(dir, e.Name())
		out = append(out, appInfo{Name: e.Name(), Dir: d, Kind: detectKind(d), Routable: hosts.Routable(d)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func detectKind(dir string) string {
	exists := func(rel string) bool { _, err := os.Stat(filepath.Join(dir, rel)); return err == nil }
	switch {
	case exists("wp-config.php") || exists("wp-load.php"):
		return "WordPress"
	case exists("artisan"):
		return "Laravel"
	case exists("package.json"):
		return packageFramework(filepath.Join(dir, "package.json"))
	case exists("index.php") || exists("public/index.php"):
		return "PHP"
	case exists("public/index.html") || exists("index.html"):
		return "static"
	}
	return "-"
}

func packageFramework(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "JavaScript"
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		for _, fw := range []struct{ dep, name string }{
			{`"next"`, "Next.js"}, {`"nuxt"`, "Nuxt"}, {`"@sveltejs/kit"`, "SvelteKit"}, {`"astro"`, "Astro"},
			{`"@angular/core"`, "Angular"}, {`"@builder.io/qwik"`, "Qwik"}, {`"react-router"`, "React Router"}, {`"vite"`, "Vite"},
		} {
			if strings.Contains(line, fw.dep) {
				return fw.name
			}
		}
	}
	return "JavaScript"
}

func (m *Model) handleAppsKey(key string) tea.Cmd {
	c := m.cursor[viewApps]
	if c >= len(m.apps) {
		return nil
	}
	app := m.apps[c]
	switch key {
	case "o", "enter":
		if !app.Routable {
			m.setStatus(false, "%s is not routable (no index.php, public/, or package.json)", app.Name)
			return nil
		}
		return m.open("https://" + app.Name + ".test")
	case "e":
		if m.deps.OpenDir == nil {
			return nil
		}
		if err := m.deps.OpenDir(app.Dir); err != nil {
			m.setStatus(false, "open editor: %v", err)
		} else {
			m.setStatus(true, "opened %s in the editor", app.Name)
		}
	}
	return nil
}
