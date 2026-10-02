package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

// A JavaScript app runs in its own node-<app> container, started from the
// node service's app.compose.yml as compose project devarch-node-<app>. The
// shared node router forwards <app>.test to it.
const (
	nodeService  = "backend/node"
	proxyService = "proxy/nginx-proxy-manager"
	appCompose   = "app.compose.yml"
)

// The node- prefix must still fit a 63-character DNS label.
var (
	appNameRE   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,56}[a-z0-9])?$`)
	appScriptRE = regexp.MustCompile(`^[A-Za-z0-9:_-]+$`)
)

// AppOptions controls StartApp.
type AppOptions struct {
	// Script is the package.json script to run (default devarch).
	Script string
	// PackageManager is auto, npm, pnpm or yarn.
	PackageManager string
	NoHosts        bool
}

// AppPlan is what StartApp resolved before running anything.
type AppPlan struct {
	App, Dir, Container, Script, PackageManager string
}

// AppContainer is the container name of a JavaScript app.
func AppContainer(app string) string { return "node-" + app }

// PlanApp validates a JavaScript app and resolves its options.
func (e *Engine) PlanApp(app string, opts AppOptions) (AppPlan, error) {
	p := AppPlan{App: app, Container: AppContainer(app), Script: opts.Script, PackageManager: opts.PackageManager}
	if p.Script == "" {
		p.Script = "devarch"
	}
	if p.PackageManager == "" {
		p.PackageManager = "auto"
	}
	if !appNameRE.MatchString(app) {
		return p, fmt.Errorf("app name %q must be a lowercase DNS label of at most 58 characters", app)
	}
	if !appScriptRE.MatchString(p.Script) {
		return p, fmt.Errorf("package script %q contains unsupported characters", p.Script)
	}
	p.Dir = filepath.Join(e.Settings.AppsDir, app)
	data, err := os.ReadFile(filepath.Join(p.Dir, "package.json"))
	if errors.Is(err, os.ErrNotExist) {
		if _, derr := os.Stat(p.Dir); derr != nil {
			return p, fmt.Errorf("no app at %s", p.Dir)
		}
		return p, fmt.Errorf("%s has no package.json", p.Dir)
	}
	if err != nil {
		return p, err
	}
	var pkg struct {
		Scripts map[string]any `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return p, fmt.Errorf("%s/package.json: %w", p.Dir, err)
	}
	if s, ok := pkg.Scripts[p.Script].(string); !ok || strings.TrimSpace(s) == "" {
		return p, fmt.Errorf("package.json does not define scripts.%s; the app's server must listen on 0.0.0.0:3000", p.Script)
	}
	switch p.PackageManager {
	case "auto":
		p.PackageManager = "npm"
		for _, lock := range [][2]string{{"pnpm-lock.yaml", "pnpm"}, {"yarn.lock", "yarn"}} {
			if _, err := os.Stat(filepath.Join(p.Dir, lock[0])); err == nil {
				p.PackageManager = lock[1]
				break
			}
		}
	case "npm", "pnpm", "yarn":
	default:
		return p, fmt.Errorf("package manager must be auto, npm, pnpm or yarn, not %q", p.PackageManager)
	}
	return p, nil
}

func (e *Engine) appService() (catalog.Service, error) {
	svc, ok := e.Catalog.Get(nodeService)
	if !ok {
		return svc, fmt.Errorf("%s is not in the catalog", nodeService)
	}
	return svc, nil
}

// appCompose runs compose for one app's project from the node service directory.
func (e *Engine) appCompose(svc catalog.Service, p AppPlan, args ...string) runner.Cmd {
	c := e.Compose(svc, append([]string{"-p", "devarch-node-" + p.App, "-f", appCompose}, args...)...)
	c.Env = append(c.Env,
		"DEVARCH_NODE_APP_NAME="+p.App,
		"DEVARCH_NODE_SCRIPT="+p.Script,
		"DEVARCH_NODE_PACKAGE_MANAGER="+p.PackageManager,
		"DEVARCH_NODE_CONTAINER_USER="+e.Runtime().ContainerUser(),
	)
	return c
}

// StartApp starts the shared router and proxy, (re)creates the app's
// container so environment changes apply, reloads the proxy, and registers
// <app>.test.
func (e *Engine) StartApp(ctx context.Context, p AppPlan, opts AppOptions) error {
	svc, err := e.appService()
	if err != nil {
		return err
	}
	var shared []catalog.Service
	for _, id := range []string{nodeService, proxyService} {
		if s, ok := e.Catalog.Get(id); ok {
			shared = append(shared, s)
		}
	}
	if err := e.Up(ctx, shared, UpOptions{Build: true, NoHosts: true}); err != nil {
		return err
	}
	e.logf("starting %s for %s (%s run %s)", p.Container, p.Dir, p.PackageManager, p.Script)
	if err := e.Runner.Run(ctx, e.appCompose(svc, p, "up", "-d", "--build", "--force-recreate")); err != nil {
		return fmt.Errorf("start %s: %w", p.Container, err)
	}
	if proxy, ok := e.Catalog.Get(proxyService); ok {
		name := proxy.Primary().ContainerName
		for _, args := range [][]string{{"exec", "-T", name, "nginx", "-t"}, {"exec", "-T", name, "nginx", "-s", "reload"}} {
			if err := e.Runner.Run(ctx, e.Compose(proxy, args...)); err != nil {
				return fmt.Errorf("reload the proxy: %w", err)
			}
		}
	}
	if !opts.NoHosts && e.Hosts != nil && e.Settings.HostsManage {
		host := p.App + ".test"
		if e.DryRun {
			e.logf("would register %s %s in %s", e.Settings.HostsAddress, host, e.Hosts.Path())
		} else if _, err := e.Hosts.Add(ctx, host, e.Settings.HostsAddress); err != nil {
			e.logf("warning: could not register %s: %v (run `devarch hosts add %s`)", host, err, host)
		}
	}
	return nil
}

// StopApp removes an app's container; volumes also removes its cached
// node_modules volume.
func (e *Engine) StopApp(ctx context.Context, app string, volumes bool) error {
	if !appNameRE.MatchString(app) {
		return fmt.Errorf("app name %q must be a lowercase DNS label of at most 58 characters", app)
	}
	svc, err := e.appService()
	if err != nil {
		return err
	}
	args := []string{"down"}
	if volumes {
		args = append(args, "--volumes")
	}
	e.logf("stopping %s", AppContainer(app))
	return e.Runner.Run(ctx, e.appCompose(svc, AppPlan{App: app, Script: "devarch", PackageManager: "npm"}, args...))
}
