package tui

import (
	"context"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

type (
	logLineMsg string
	logEndMsg  struct{ err error }
)

const maxLogLines = 5000

// logsModel streams a native logs command into a scrollable viewport.
type logsModel struct {
	title     string
	command   string
	lines     []string
	vp        viewport.Model
	filter    textinput.Model
	filtering bool
	follow    bool
	ended     string
	cancel    context.CancelFunc
	ch        chan string
	done      chan error
}

func (m *Model) openLogs(title string, c runner.Cmd) tea.Cmd {
	ctx, cancel := context.WithCancel(m.ctx)
	l := &logsModel{title: title, command: c.String(), follow: true, cancel: cancel,
		ch: make(chan string, 1024), done: make(chan error, 1)}
	l.filter = textinput.New()
	l.filter.Prompt = "/"
	l.resize(m.width, m.height)

	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	cmd.Env = append(append(cmd.Environ(), c.Env...), streamEnv...)
	w := &lineWriter{ch: l.ch}
	cmd.Stdout, cmd.Stderr = w, w
	if err := cmd.Start(); err != nil {
		cancel()
		m.setStatus(false, "logs: %v", err)
		return nil
	}
	go func() { l.done <- cmd.Wait() }()
	m.logs, m.mode = l, modeLogs
	return l.wait()
}

func (l *logsModel) wait() tea.Cmd {
	return func() tea.Msg {
		select {
		case line := <-l.ch:
			return logLineMsg(line)
		case err := <-l.done:
			// Drain lines written before the process exited.
			for {
				select {
				case line := <-l.ch:
					l.lines = append(l.lines, line)
				default:
					return logEndMsg{err}
				}
			}
		}
	}
}

func (l *logsModel) stop() { l.cancel() }

func (l *logsModel) resize(w, h int) {
	if w == 0 || h == 0 {
		w, h = 80, 24
	}
	l.vp.Width, l.vp.Height = w, max(h-3, 1)
	l.render()
}

func (l *logsModel) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case logLineMsg:
		l.lines = append(l.lines, string(msg))
		if len(l.lines) > maxLogLines {
			l.lines = l.lines[len(l.lines)-maxLogLines:]
		}
		l.render()
		return l.wait()
	case logEndMsg:
		l.ended = "stream ended"
		if msg.err != nil && !strings.Contains(msg.err.Error(), "killed") {
			l.ended = "stream ended: " + msg.err.Error()
		}
		l.render()
	}
	return nil
}

func (l *logsModel) render() {
	lines := l.lines
	if q := strings.ToLower(l.filter.Value()); q != "" {
		lines = nil
		for _, line := range l.lines {
			if strings.Contains(strings.ToLower(line), q) {
				lines = append(lines, line)
			}
		}
	}
	l.vp.SetContent(strings.Join(lines, "\n"))
	if l.follow {
		l.vp.GotoBottom()
	}
}

// handleKey returns true when the view should close.
func (l *logsModel) handleKey(k tea.KeyMsg) bool {
	if l.filtering {
		switch k.String() {
		case "enter", "esc":
			if k.String() == "esc" {
				l.filter.SetValue("")
			}
			l.filtering = false
			l.filter.Blur()
		default:
			l.filter, _ = l.filter.Update(k)
		}
		l.render()
		return false
	}
	switch k.String() {
	case "esc", "q":
		return true
	case "/":
		l.filtering = true
		l.filter.Focus()
	case "G", "end":
		l.follow = true
		l.vp.GotoBottom()
	case "g", "home":
		l.follow = false
		l.vp.GotoTop()
	default:
		l.vp, _ = l.vp.Update(k)
		l.follow = l.vp.AtBottom()
	}
	return false
}

func (l *logsModel) view(width int) string {
	header := titleStyle.Render(l.title) + "  " + dimStyle.Render(truncate("$ "+l.command, max(width-lipgloss.Width(l.title)-2, 10)))
	footer := dimStyle.Render("esc back · / filter · g/G top/bottom · ↑↓ pgup pgdn scroll")
	if l.filtering || l.filter.Value() != "" {
		footer = l.filter.View() + "  " + footer
	}
	if l.ended != "" {
		footer = warnStyle.Render(l.ended) + "  " + footer
	}
	return header + "\n" + l.vp.View() + "\n" + footer
}
