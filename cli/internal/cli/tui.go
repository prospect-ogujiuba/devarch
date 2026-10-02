package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/prospect-ogujiuba/devarch/cli/internal/doctor"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/hosts"
	"github.com/prospect-ogujiuba/devarch/cli/internal/lint"
	"github.com/prospect-ogujiuba/devarch/cli/internal/tui"
)

func (a *app) runTUI() error {
	if err := a.load(); err != nil {
		return err
	}
	exe, _ := os.Executable()
	return tui.Run(tui.Deps{
		Engine:    a.eng,
		Recipes:   a.recipes,
		RecipeEnv: a.recipeEnv(),
		Doctor: func(ctx context.Context, e *engine.Engine) []doctor.Check {
			return doctor.Run(ctx, doctor.Env{Engine: e, RootHow: string(a.rootHow), Problems: len(a.problems)})
		},
		Lint: func(ctx context.Context) (int, int) {
			return lint.Count(lint.Run(ctx, a.eng.Catalog, a.problems, lint.Options{}))
		},
		Exe:     exe,
		OpenURL: openURL,
		OpenDir: func(dir string) error { return openDir(a.eng.Settings.Editor, dir) },
		Events:  func(ctx context.Context) <-chan struct{} { return containerEvents(ctx, a.eng) },
	})
}

// containerEvents signals container lifecycle changes and health transitions
// while the TUI is open. Healthcheck runs that do not change health are
// ignored, so an idle stack causes no refreshes.
func containerEvents(ctx context.Context, e *engine.Engine) <-chan struct{} {
	ch := make(chan struct{}, 1)
	cmd := exec.CommandContext(ctx, e.Settings.Runtime, e.EventsCmd()...)
	out, err := cmd.StdoutPipe()
	if err != nil || cmd.Start() != nil {
		close(ch)
		return ch
	}
	go func() {
		defer close(ch)
		defer cmd.Wait()
		health := map[string]string{}
		sc := bufio.NewScanner(out)
		for sc.Scan() {
			ev, ok := engine.ParseEvent(sc.Bytes())
			if !ok {
				continue
			}
			if ev.Action == "health_status" {
				if health[ev.Name] == ev.Health {
					continue
				}
				health[ev.Name] = ev.Health
			}
			select {
			case ch <- struct{}{}:
			default: // a refresh is already pending
			}
		}
	}()
	return ch
}

func openURL(url string) error {
	var cmd *exec.Cmd
	switch {
	case runtime.GOOS == "darwin":
		cmd = exec.Command("open", url)
	case runtime.GOOS == "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case hosts.IsWSL():
		if p, err := exec.LookPath("wslview"); err == nil {
			cmd = exec.Command(p, url)
			break
		}
		cmdExe := "/mnt/c/Windows/System32/cmd.exe"
		if _, err := os.Stat("/init"); err == nil {
			// The interop host keeps Wine or another binfmt handler out of the way.
			cmd = exec.Command("/init", cmdExe, "/c", "start", "", url)
		} else {
			cmd = exec.Command(cmdExe, "/c", "start", "", url)
		}
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return startDetached(cmd)
}

// openDir runs the configured editor command with dir as its last argument.
func openDir(editor, dir string) error {
	argv := strings.Fields(editor)
	if len(argv) == 0 {
		return errors.New("no editor is configured; run: devarch config set editor code")
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return fmt.Errorf("editor %q is not on PATH; change it with: devarch config set editor <command>", argv[0])
	}
	return startDetached(exec.Command(path, append(argv[1:], dir)...))
}

func startDetached(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}
