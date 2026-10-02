package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/prospect-ogujiuba/devarch/cli/internal/doctor"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
)

const (
	detailLines   = 5
	activityLines = 5
)

func (m *Model) View() string {
	w, h := m.width, m.height
	if w == 0 || h == 0 {
		w, h = 100, 30
	}
	switch m.mode {
	case modeLogs:
		return m.logs.view(w)
	case modeHelp:
		return helpView(m.eng.Settings.Runtime)
	}

	header := m.headerView(w)
	footer := m.footerView(w)
	activity := m.activityView(w)
	var body, detail string
	if m.mode == modeWizard {
		body = m.wizard.view(m)
	} else {
		detail = m.detailView(w)
	}
	listHeight := h - lipgloss.Height(header) - lipgloss.Height(footer) - activityLines - 1
	if detail != "" {
		listHeight -= detailLines + 1
	}
	listHeight = max(listHeight, 3)
	if body == "" {
		body = m.listView(w, listHeight)
	}
	body = fitHeight(body, listHeight)

	parts := []string{header, body}
	if detail != "" {
		parts = append(parts, dimStyle.Render(strings.Repeat("─", w)), fitHeight(detail, detailLines))
	}
	parts = append(parts, dimStyle.Render(strings.Repeat("─", w)), activity, footer)
	return strings.Join(parts, "\n")
}

func fitHeight(s string, h int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func (m *Model) headerView(w int) string {
	var tabs []string
	for i, name := range viewNames {
		label := fmt.Sprintf("%d %s", i+1, name)
		if view(i) == m.view && m.mode != modeWizard {
			tabs = append(tabs, tabActive.Render(label))
		} else {
			tabs = append(tabs, tabStyle.Render(label))
		}
	}
	left := titleStyle.Render("DevArch") + " " + strings.Join(tabs, "")
	right := ""
	if m.mode == modeFilter || m.filter.Value() != "" {
		right = m.filter.View()
	}
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return truncate(left, w)
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m *Model) footerView(w int) string {
	var line string
	switch {
	case m.mode == modeConfirm:
		line = warnStyle.Render(m.confirm.prompt)
	case m.mode == modeVersion:
		line = dimStyle.Render("↑↓ choose · enter switch · esc cancel")
	case m.busy != "":
		line = m.spin.View() + " " + m.busy + "…"
	case m.status != "":
		if m.statusOK {
			line = okStyle.Render(m.status)
		} else {
			line = warnStyle.Render(m.status)
		}
	default:
		keys := map[view]string{
			viewServices: "u up · d down · r restart · v version · l logs · o open · / filter",
			viewApps:     "o open · e editor",
			viewRunning:  "l logs · r restart",
			viewDoctor:   "r rerun",
		}[m.view]
		line = dimStyle.Render(keys + " · n new · H hosts · tab view · ? help · q quit")
	}
	return truncate(line, w)
}

func (m *Model) activityView(w int) string {
	start := max(len(m.activity)-activityLines, 0)
	lines := make([]string, 0, activityLines)
	for _, l := range m.activity[start:] {
		style := dimStyle
		switch {
		case strings.HasPrefix(l, "$ "):
			style = cmdStyle
		case strings.HasPrefix(l, "✗"):
			style = errStyle
		case strings.HasPrefix(l, "✓"):
			style = okStyle
		}
		lines = append(lines, style.Render(truncate(l, w)))
	}
	return fitHeight(strings.Join(lines, "\n"), activityLines)
}

// window returns the visible slice bounds keeping the cursor in view.
func window(cursor, n, height int) (int, int) {
	start := 0
	if cursor >= height {
		start = cursor - height + 1
	}
	return start, min(start+height, n)
}

func (m *Model) listView(w, h int) string {
	var rows []string
	cursor := m.cursor[m.view]
	switch m.view {
	case viewServices:
		if len(m.services) == 0 {
			return dimStyle.Render("no services match the filter")
		}
		idW := 10
		for _, s := range m.services {
			idW = max(idW, len(s.ID))
		}
		idW = min(idW, w/2)
		start, end := window(cursor, len(m.services), h)
		for i := start; i < end; i++ {
			s := m.services[i]
			st := m.states[s.ID]
			dot, label := stateDot(st), stateLabel(st)
			if m.psErr != nil {
				dot, label = dimStyle.Render("?"), "unknown"
			}
			url := ""
			if len(s.Meta.URLs) > 0 {
				url = s.Meta.URLs[0]
			}
			line := fmt.Sprintf("%s %s  %s  %s  %s", dot, pad(truncate(s.ID, idW), idW), pad(dashIf(m.eng.SelectedVersion(s)), 6), pad(label, 18), url)
			rows = append(rows, m.row(line, i == cursor, w))
		}
	case viewApps:
		if len(m.apps) == 0 {
			return dimStyle.Render("no apps yet; press n to create one")
		}
		start, end := window(cursor, len(m.apps), h)
		for i := start; i < end; i++ {
			a := m.apps[i]
			url := dimStyle.Render("not routable")
			if a.Routable {
				url = "https://" + a.Name + ".test"
			}
			rows = append(rows, m.row(fmt.Sprintf("%s  %s  %s", pad(truncate(a.Name, 34), 34), pad(a.Kind, 12), url), i == cursor, w))
		}
	case viewRunning:
		rs := m.running()
		if len(rs) == 0 {
			if m.psErr != nil {
				return errStyle.Render(m.eng.Settings.Runtime + ": " + m.psErr.Error())
			}
			return dimStyle.Render("no running containers")
		}
		start, end := window(cursor, len(rs), h)
		for i := start; i < end; i++ {
			c := rs[i]
			rows = append(rows, m.row(fmt.Sprintf("%s  %s  %s  %s", pad(truncate(c.Name, 28), 28), pad(dashIf(c.Service), 26), pad(c.Status, 24), ports(c.Ports)), i == cursor, w))
		}
	case viewDoctor:
		if !m.doctorLoaded {
			return m.spin.View() + " running checks…"
		}
		for i, c := range m.checks {
			mark := map[doctor.Status]string{doctor.OK: okStyle.Render("✓"), doctor.Warn: warnStyle.Render("!"), doctor.Fail: errStyle.Render("✗")}[c.Status]
			rows = append(rows, m.row(fmt.Sprintf("%s %s %s", mark, pad(c.Name, 17), c.Detail), i == cursor, w))
			if c.Fix != "" && c.Status != doctor.OK {
				rows = append(rows, dimStyle.Render(truncate("    → "+c.Fix, w)))
			}
		}
		lint := okStyle.Render("✓")
		if m.lintErr > 0 {
			lint = errStyle.Render("✗")
		} else if m.lintWrn > 0 {
			lint = warnStyle.Render("!")
		}
		rows = append(rows, fmt.Sprintf("%s %s %d errors, %d warnings (devarch lint for details)", lint, pad("catalog lint", 17), m.lintErr, m.lintWrn))
		if len(rows) > h {
			rows = rows[:h]
		}
	}
	return strings.Join(rows, "\n")
}

func (m *Model) row(line string, selected bool, w int) string {
	line = truncate(line, w)
	if selected {
		return cursorStyle.Render(pad(line, w))
	}
	return line
}

func dashIf(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func stateDot(st engine.ServiceState) string {
	switch {
	case st.State == "running" && (st.Health == "" || st.Health == "healthy"):
		return okStyle.Render("●")
	case st.State == "running" || st.State == "partial":
		return warnStyle.Render("◐")
	case st.State == "stopped":
		return dimStyle.Render("○")
	}
	return dimStyle.Render("·")
}

func stateLabel(st engine.ServiceState) string {
	switch {
	case st.State == "" || st.State == "absent":
		return "-"
	case st.Health != "" && st.Health != "healthy":
		return st.State + " (" + st.Health + ")"
	}
	return st.State
}

func ports(ps []engine.Port) string {
	var out []string
	for _, p := range ps {
		if p.HostPort > 0 {
			out = append(out, fmt.Sprintf("%d→%d", p.HostPort, p.ContainerPort))
		}
	}
	return strings.Join(out, " ")
}

func (m *Model) detailView(w int) string {
	if m.mode == modeVersion {
		return m.versionView()
	}
	switch m.view {
	case viewServices:
		s, ok := m.selectedService()
		if !ok {
			return ""
		}
		var lines []string
		head := titleStyle.Render(s.Title())
		if s.Meta.Description != "" {
			head += " · " + s.Meta.Description
		}
		lines = append(lines, truncate(head, w))
		var facts []string
		if len(s.Meta.Tags) > 0 {
			facts = append(facts, strings.Join(s.Meta.Tags, ", "))
		}
		if v := s.Meta.Version; v != nil {
			facts = append(facts, fmt.Sprintf("version %s (%s; choices %s)", m.eng.SelectedVersion(s), v.Var, strings.Join(v.Choices, " ")))
		}
		if len(s.Meta.Requires) > 0 {
			facts = append(facts, "requires "+strings.Join(s.Meta.Requires, ", "))
		}
		if len(facts) > 0 {
			lines = append(lines, dimStyle.Render(truncate(strings.Join(facts, " · "), w)))
		}
		rel, err := filepath.Rel(m.eng.Root, s.ComposeFile)
		if err != nil {
			rel = s.ComposeFile
		}
		lines = append(lines, dimStyle.Render(truncate(rel+"  ("+s.Source+")", w)))
		var cs []string
		for _, c := range s.Containers {
			name := c.ContainerName
			if name == "" {
				name = c.Service
			}
			for _, p := range c.Ports {
				if p.Published > 0 {
					name += fmt.Sprintf(" :%d", p.Published)
				}
			}
			cs = append(cs, name)
		}
		lines = append(lines, truncate("containers: "+strings.Join(cs, ", "), w))
		if s.Meta.Notes != "" {
			lines = append(lines, warnStyle.Render(truncate(s.Meta.Notes, w)))
		} else if len(s.Meta.URLs) > 1 {
			lines = append(lines, truncate(strings.Join(s.Meta.URLs, "  "), w))
		}
		return strings.Join(lines, "\n")
	case viewApps:
		if c := m.cursor[viewApps]; c < len(m.apps) {
			a := m.apps[c]
			return titleStyle.Render(a.Name) + " · " + a.Kind + "\n" + dimStyle.Render(a.Dir)
		}
	case viewRunning:
		if rs := m.running(); m.cursor[viewRunning] < len(rs) {
			c := rs[m.cursor[viewRunning]]
			return titleStyle.Render(c.Name) + "\n" + dimStyle.Render(c.Image) + "\n" + c.Status
		}
	}
	return ""
}

func (m *Model) versionView() string {
	svc := m.version.svc
	current := m.eng.SelectedVersion(svc)
	var b strings.Builder
	b.WriteString(titleStyle.Render(svc.Title()+" version") + "\n")
	for i, c := range m.version.choices {
		mark := "  "
		if c == current {
			mark = "* "
		}
		label := mark + c
		if c == svc.Meta.Version.Default {
			label += dimStyle.Render("  default")
		}
		if i == m.version.cursor {
			label = cursorStyle.Render(label)
		}
		b.WriteString(label + "\n")
	}
	return b.String()
}

func helpView(runtime string) string {
	return boxStyle.Render(strings.Join([]string{
		titleStyle.Render("DevArch keys"),
		"",
		"tab / ← →  switch view (or 1-4)      ↑↓ / j k  move",
		"/          filter services            ctrl+r    refresh",
		"u d r      up, down, restart          v         switch version",
		"l / enter  follow logs                o         open URL",
		"n          new project                H         sync .test hosts",
		"e          open app in editor         q         quit",
		"",
		"Every action runs native " + runtime + " commands; they appear",
		"in the activity pane. Nothing keeps running after you quit.",
		"",
		dimStyle.Render("press any key"),
	}, "\n"))
}
