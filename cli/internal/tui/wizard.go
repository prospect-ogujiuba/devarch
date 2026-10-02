package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/prospect-ogujiuba/devarch/cli/internal/recipe"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

type (
	progressMsg   recipe.Event
	wizardDoneMsg struct{ err error }
)

type wizardStep int

const (
	stepRecipe wizardStep = iota
	stepForm
	stepConfirm
	stepRun
	stepDone
)

type field struct {
	arg   recipe.Arg
	input textinput.Model
	value string
}

func (f field) kind() string {
	switch {
	case f.arg.Type == "bool":
		return "bool"
	case len(f.arg.Choices) > 0 || f.arg.ChoicesFrom != "":
		return "choice"
	}
	return "text"
}

type stepState struct{ name, state, message string }

type wizardModel struct {
	step    wizardStep
	recipes []recipe.Recipe
	rcursor int
	r       recipe.Recipe
	fields  []field
	fcursor int
	err     string
	argv    []string
	steps   []stepState
	runErr  error
	prog    chan recipe.Event
	cancel  context.CancelFunc
}

func (m *Model) openWizard() tea.Cmd {
	if m.deps.Recipes == nil {
		return nil
	}
	rs, err := m.deps.Recipes()
	if err != nil {
		m.setStatus(false, "recipes: %v", err)
		return nil
	}
	if len(rs) == 0 {
		m.setStatus(false, "no recipes found under recipes/")
		return nil
	}
	m.wizard, m.mode = &wizardModel{recipes: rs}, modeWizard
	return nil
}

func (w *wizardModel) values() map[string]string {
	v := map[string]string{}
	for _, f := range w.fields {
		if f.value != "" {
			v[f.arg.Name] = f.value
		}
	}
	return v
}

func (w *wizardModel) choose(r recipe.Recipe) tea.Cmd {
	w.r, w.step, w.fcursor, w.err = r, stepForm, 0, ""
	w.fields = nil
	for _, a := range r.Args {
		f := field{arg: a, value: a.Default}
		if f.kind() == "text" {
			f.input = textinput.New()
			f.input.Prompt = ""
			f.input.Placeholder = a.Help
			f.input.SetValue(a.Default)
		}
		w.fields = append(w.fields, f)
	}
	w.syncChoices()
	return w.focus()
}

// syncChoices clears choice values that are no longer valid after a value
// they depend on changed (a JavaScript profile depends on its framework).
func (w *wizardModel) syncChoices() {
	vals := w.values()
	for i := range w.fields {
		f := &w.fields[i]
		if f.kind() != "choice" || f.value == "" {
			continue
		}
		if cs := w.r.ChoicesFor(f.arg, vals); !slices.Contains(cs, f.value) {
			f.value = ""
			vals = w.values()
		}
	}
}

func (w *wizardModel) focus() tea.Cmd {
	var cmd tea.Cmd
	for i := range w.fields {
		if w.fields[i].kind() != "text" {
			continue
		}
		if i == w.fcursor {
			cmd = w.fields[i].input.Focus()
		} else {
			w.fields[i].input.Blur()
		}
	}
	return cmd
}

func (w *wizardModel) validate() error {
	vals := w.values()
	for _, f := range w.fields {
		if (f.arg.Required || f.arg.Positional) && vals[f.arg.Name] == "" {
			return fmt.Errorf("%s is required", f.arg.Name)
		}
	}
	return w.r.Validate(vals)
}

func (w *wizardModel) confirmations() []string {
	var out []string
	for _, f := range w.fields {
		if f.arg.Confirm != "" && f.value == "true" {
			out = append(out, f.arg.Confirm)
		}
	}
	return out
}

// handleKey returns the next command and whether the wizard closed.
func (w *wizardModel) handleKey(m *Model, k tea.KeyMsg) (tea.Cmd, bool) {
	key := k.String()
	switch w.step {
	case stepRecipe:
		switch key {
		case "esc", "q":
			return nil, true
		case "up", "k":
			w.rcursor = max(w.rcursor-1, 0)
		case "down", "j":
			w.rcursor = min(w.rcursor+1, len(w.recipes)-1)
		case "enter":
			return w.choose(w.recipes[w.rcursor]), false
		}
	case stepForm:
		f := &w.fields[w.fcursor]
		switch key {
		case "esc":
			w.step = stepRecipe
			return nil, false
		case "up", "shift+tab":
			w.fcursor = max(w.fcursor-1, 0)
			return w.focus(), false
		case "down", "tab":
			w.fcursor = min(w.fcursor+1, len(w.fields)-1)
			return w.focus(), false
		case "enter":
			if w.fcursor < len(w.fields)-1 {
				w.fcursor++
				return w.focus(), false
			}
			if err := w.validate(); err != nil {
				w.err = err.Error()
				return nil, false
			}
			w.err = ""
			w.argv = append(w.r.Argv(w.values()), w.r.NoHostsFlag...)
			w.step = stepConfirm
			return nil, false
		}
		switch f.kind() {
		case "bool":
			if key == " " || key == "x" {
				f.value = map[bool]string{true: "", false: "true"}[f.value == "true"]
			}
		case "choice":
			cs := w.r.ChoicesFor(f.arg, w.values())
			if len(cs) == 0 || (key != "left" && key != "right" && key != " " && key != "h" && key != "l") {
				return nil, false
			}
			i := slices.Index(cs, f.value)
			if key == "left" || key == "h" {
				i = (i - 1 + len(cs)) % len(cs)
			} else {
				i = (i + 1) % len(cs)
			}
			f.value = cs[i]
			w.syncChoices()
		default:
			var cmd tea.Cmd
			f.input, cmd = f.input.Update(k)
			f.value = strings.TrimSpace(f.input.Value())
			w.syncChoices()
			return cmd, false
		}
	case stepConfirm:
		switch key {
		case "esc":
			w.step = stepForm
		case "enter", "y", "Y":
			if len(w.confirmations()) > 0 && key == "enter" {
				w.err = "press y to confirm"
				return nil, false
			}
			return w.start(m), false
		}
	case stepRun:
		if key == "esc" {
			m.setStatus(false, "the recipe keeps running; its output is in the activity pane")
			return nil, true
		}
	case stepDone:
		switch key {
		case "esc", "q", "enter":
			return nil, true
		case "H":
			return m.syncHosts(), false
		case "o":
			if url := w.url(); url != "" {
				return m.open(url), false
			}
		}
	}
	return nil, false
}

func (w *wizardModel) url() string {
	if name := w.values()["name"]; name != "" {
		return "https://" + name + ".test"
	}
	return ""
}

func (w *wizardModel) start(m *Model) tea.Cmd {
	if m.busy != "" {
		w.err = "busy: " + m.busy
		return nil
	}
	ctx, cancel := context.WithCancel(m.ctx)
	w.cancel, w.step, w.err = cancel, stepRun, ""
	w.prog = make(chan recipe.Event, 256)
	label := fmt.Sprintf("new %s %s", w.r.Name, w.values()["name"])
	m.busy = label
	m.addActivity("$ " + runner.Cmd{Name: w.r.Entry, Args: w.argv}.String())
	out := &lineWriter{ch: m.out}
	run := func() tea.Msg {
		err := w.r.RunWithProgress(ctx, w.argv, m.deps.RecipeEnv, out, func(e recipe.Event) { w.prog <- e })
		close(w.prog)
		return wizardDoneMsg{err}
	}
	return tea.Batch(run, w.waitProgress())
}

func (w *wizardModel) waitProgress() tea.Cmd {
	return func() tea.Msg {
		e, ok := <-w.prog
		if !ok {
			return nil
		}
		return progressMsg(e)
	}
}

func (w *wizardModel) update(m *Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case progressMsg:
		e := recipe.Event(msg)
		found := false
		for i := range w.steps {
			if w.steps[i].name == e.Step {
				w.steps[i].state, w.steps[i].message, found = e.State, e.Message, true
			}
		}
		if !found {
			w.steps = append(w.steps, stepState{e.Step, e.State, e.Message})
		}
		return w.waitProgress()
	case wizardDoneMsg:
		m.busy = ""
		w.runErr, w.step = msg.err, stepDone
		if msg.err != nil {
			m.setStatus(false, "new %s failed: %v", w.r.Name, msg.err)
		} else {
			m.setStatus(true, "created %s", w.url())
		}
		m.apps = discoverApps(filepath.Join(m.eng.Root, "apps"))
		return m.loadContainers()
	}
	return nil
}

func (w *wizardModel) view(m *Model) string {
	var b strings.Builder
	title := "New project"
	if w.step > stepRecipe {
		title += " · " + w.r.Title
	}
	b.WriteString(titleStyle.Render(title) + "\n\n")
	switch w.step {
	case stepRecipe:
		for i, r := range w.recipes {
			line := fmt.Sprintf("%-12s %s", r.Name, r.Title)
			if i == w.rcursor {
				line = cursorStyle.Render(line)
			}
			b.WriteString("  " + line + "\n")
		}
		if r := w.recipes[w.rcursor]; r.Description != "" {
			b.WriteString("\n" + dimStyle.Render(r.Description) + "\n")
			if len(r.Requires) > 0 {
				b.WriteString(dimStyle.Render("Starts: "+strings.Join(r.Requires, ", ")) + "\n")
			}
		}
		b.WriteString("\n" + dimStyle.Render("enter choose · esc close"))
	case stepForm:
		vals := w.values()
		for i, f := range w.fields {
			label := pad(f.arg.Name, 12)
			var val string
			switch f.kind() {
			case "bool":
				val = map[bool]string{true: "[x]", false: "[ ]"}[f.value == "true"] + " " + dimStyle.Render(f.arg.Help)
			case "choice":
				cs := w.r.ChoicesFor(f.arg, vals)
				switch {
				case len(cs) == 0:
					val = dimStyle.Render("(set the fields above first)")
				case f.value == "":
					val = dimStyle.Render("‹ choose ›  " + strings.Join(cs, " "))
				default:
					val = "‹ " + f.value + " ›  " + dimStyle.Render(fmt.Sprintf("%d options", len(cs)))
				}
			default:
				val = f.input.View()
			}
			if i == w.fcursor {
				label = cursorStyle.Render(label)
			}
			b.WriteString("  " + label + "  " + val + "\n")
		}
		if w.err != "" {
			b.WriteString("\n" + errStyle.Render(w.err) + "\n")
		}
		b.WriteString("\n" + dimStyle.Render("↑↓ field · ←→ choose · space toggle · enter next/continue · esc back"))
	case stepConfirm:
		b.WriteString("This runs:\n\n  " + cmdStyle.Render(runner.Cmd{Name: w.r.Entry, Args: w.argv}.String()) + "\n\n")
		if len(w.r.Requires) > 0 {
			b.WriteString(dimStyle.Render("It starts "+strings.Join(w.r.Requires, ", ")+" if they are not running.") + "\n")
		}
		if len(w.r.NoHostsFlag) > 0 {
			b.WriteString(dimStyle.Render("Hostname registration needs sudo or UAC, so it runs afterwards with H.") + "\n")
		}
		for _, c := range w.confirmations() {
			b.WriteString("\n" + warnStyle.Render("! "+c) + "\n")
		}
		if w.err != "" {
			b.WriteString("\n" + errStyle.Render(w.err) + "\n")
		}
		hint := "enter run · esc back"
		if len(w.confirmations()) > 0 {
			hint = "y run · esc back"
		}
		b.WriteString("\n" + dimStyle.Render(hint))
	case stepRun, stepDone:
		for _, s := range w.steps {
			icon := m.spin.View()
			switch s.state {
			case "done":
				icon = okStyle.Render("✓")
			case "warn":
				icon = warnStyle.Render("!")
			case "fail":
				icon = errStyle.Render("✗")
			}
			line := icon + " " + s.name
			if s.message != "" {
				line += "  " + dimStyle.Render(s.message)
			}
			b.WriteString("  " + line + "\n")
		}
		if len(w.steps) == 0 && w.step == stepRun {
			b.WriteString("  " + m.spin.View() + " starting…\n")
		}
		b.WriteString("\n")
		switch {
		case w.step == stepRun:
			b.WriteString(dimStyle.Render("Output streams into the activity pane below. esc hides this view."))
		case w.runErr != nil:
			b.WriteString(errStyle.Render("Failed: "+w.runErr.Error()) + "\n\n" + dimStyle.Render("esc close"))
		default:
			b.WriteString(okStyle.Render("Ready: "+w.url()) + "\n\n" + dimStyle.Render("H register hostname · o open · esc close"))
		}
	}
	return b.String()
}
