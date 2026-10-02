package tui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

// Colors use ANSI indexes so the terminal theme decides light or dark.
var (
	green  = lipgloss.Color("2")
	yellow = lipgloss.Color("3")
	red    = lipgloss.Color("1")
	blue   = lipgloss.Color("4")
	dim    = lipgloss.Color("8")

	titleStyle  = lipgloss.NewStyle().Bold(true)
	tabStyle    = lipgloss.NewStyle().Padding(0, 1).Foreground(dim)
	tabActive   = lipgloss.NewStyle().Padding(0, 1).Bold(true).Underline(true)
	dimStyle    = lipgloss.NewStyle().Foreground(dim)
	cursorStyle = lipgloss.NewStyle().Reverse(true)
	okStyle     = lipgloss.NewStyle().Foreground(green)
	warnStyle   = lipgloss.NewStyle().Foreground(yellow)
	errStyle    = lipgloss.NewStyle().Foreground(red)
	cmdStyle    = lipgloss.NewStyle().Foreground(blue)
	boxStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(dim).Padding(0, 1)
)

// outputMsg is one line of command output or a progress note.
type outputMsg string

// lineWriter turns writes into outputMsg lines on a channel.
type lineWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
	ch  chan<- string
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Write(p)
	for {
		line, err := w.buf.ReadString('\n')
		if err != nil {
			w.buf.Reset()
			w.buf.WriteString(line) // keep the partial line
			return len(p), nil
		}
		w.ch <- strings.TrimRight(line, "\r\n")
	}
}

// waitOutput delivers the next output line to the model.
func waitOutput(ch <-chan string) tea.Cmd {
	return func() tea.Msg { return outputMsg(<-ch) }
}

// captureRunner sends mutating command output to the TUI instead of the
// terminal. Queries run unchanged.
type captureRunner struct {
	inner runner.Runner
	out   io.Writer
}

// streamEnv makes podman-compose (Python) write through a pipe as it goes and
// drops podman's external-provider banner, which is noise in a TUI pane. It is
// added after the command is echoed so the displayed command stays native.
var streamEnv = []string{"PYTHONUNBUFFERED=1", "PODMAN_COMPOSE_WARNING_LOGS=false"}

func (c captureRunner) Run(ctx context.Context, cmd runner.Cmd) error {
	fmt.Fprintln(c.out, "$ "+cmd.String())
	cmd.Env = append(append([]string{}, cmd.Env...), streamEnv...)
	if cmd.Stdout == nil {
		cmd.Stdout = c.out
	}
	if cmd.Stderr == nil {
		cmd.Stderr = c.out
	}
	if cmd.Stdin == nil {
		cmd.Stdin = strings.NewReader("")
	}
	return c.inner.Run(ctx, cmd)
}

func (c captureRunner) Output(ctx context.Context, cmd runner.Cmd) ([]byte, error) {
	return c.inner.Output(ctx, cmd)
}

func (c captureRunner) Exec(cmd runner.Cmd) error { return c.Run(context.Background(), cmd) }

// truncate cuts s to width display cells.
func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > width {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

// pad right-pads s to width display cells.
func pad(s string, width int) string {
	if w := lipgloss.Width(s); w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}
