package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"runtime"

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
		OpenDir: openDir,
		Events:  podmanEvents,
	})
}

// podmanEvents signals container lifecycle changes and health transitions
// while the TUI is open. Healthcheck runs that do not change health are
// ignored, so an idle stack causes no refreshes.
func podmanEvents(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)
	args := []string{"events", "--format", "json", "--filter", "type=container"}
	for _, e := range []string{"create", "start", "stop", "died", "remove", "restart", "pause", "unpause", "health_status"} {
		args = append(args, "--filter", "event="+e)
	}
	cmd := exec.CommandContext(ctx, "podman", args...)
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
			var ev struct {
				Name         string `json:"Name"`
				Status       string `json:"Status"`
				HealthStatus string `json:"HealthStatus"`
			}
			if json.Unmarshal(sc.Bytes(), &ev) != nil {
				continue
			}
			if ev.Status == "health_status" {
				if health[ev.Name] == ev.HealthStatus {
					continue
				}
				health[ev.Name] = ev.HealthStatus
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

func openDir(dir string) error {
	code, err := exec.LookPath("code")
	if err != nil {
		return errors.New("VS Code's `code` command is not on PATH")
	}
	return startDetached(exec.Command(code, dir))
}

func startDetached(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}
