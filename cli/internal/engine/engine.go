// Package engine holds the operations shared by the CLI and the TUI. Every
// action resolves to native podman / podman compose commands via the runner.
package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/hosts"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
	"github.com/prospect-ogujiuba/devarch/cli/internal/state"
)

// Network is the shared external network every catalog service joins.
const Network = "microservices-net"

// Engine is the shared operation layer.
type Engine struct {
	Root     string
	Catalog  *catalog.Catalog
	Runner   runner.Runner
	Versions state.Versions
	// Log receives progress messages (not command output).
	Log io.Writer
	// Sleep is replaceable in tests.
	Sleep func(time.Duration)
	// DryRun suppresses state-file writes (the runner suppresses commands).
	DryRun bool
	// SaveVersions persists Versions; nil skips persistence.
	SaveVersions func(state.Versions) error
	// Hosts manages the hosts file; nil disables hostname registration.
	Hosts *hosts.Manager
}

func (e *Engine) logf(format string, args ...any) {
	if e.Log != nil {
		fmt.Fprintf(e.Log, format+"\n", args...)
	}
}

func (e *Engine) sleep(d time.Duration) {
	if e.Sleep != nil {
		e.Sleep(d)
		return
	}
	time.Sleep(d)
}

// Env returns the environment DevArch adds when running compose for svc: the
// selected version, if one is recorded in versions.env.
func (e *Engine) Env(svc catalog.Service) []string {
	v := svc.Meta.Version
	if v == nil || v.Var == "" {
		return nil
	}
	if val, ok := e.Versions[v.Var]; ok && val != "" {
		return []string{v.Var + "=" + val}
	}
	return nil
}

// SelectedVersion returns the version compose will use for svc.
func (e *Engine) SelectedVersion(svc catalog.Service) string {
	v := svc.Meta.Version
	if v == nil {
		return ""
	}
	if val := e.Versions[v.Var]; val != "" {
		return val
	}
	return v.Default
}

// Compose builds `podman compose <args>` run from the service directory,
// exactly as the README documents running it by hand.
func (e *Engine) Compose(svc catalog.Service, args ...string) runner.Cmd {
	return runner.Cmd{
		Name: "podman",
		Args: append([]string{"compose"}, args...),
		Dir:  svc.Dir,
		Env:  e.Env(svc),
	}
}

// EnsureNetwork creates the shared network when it does not exist.
func (e *Engine) EnsureNetwork(ctx context.Context) error {
	_, err := e.Runner.Output(ctx, runner.Cmd{Name: "podman", Args: []string{"network", "exists", Network}})
	if err == nil {
		return nil
	}
	var exit *runner.ExitError
	if !errors.As(err, &exit) {
		return fmt.Errorf("podman is unavailable: %w", err)
	}
	e.logf("creating network %s", Network)
	return e.Runner.Run(ctx, runner.Cmd{Name: "podman", Args: []string{"network", "create", Network}})
}

// UpOptions controls Up.
type UpOptions struct {
	Wait    bool
	Timeout time.Duration
	// NoRequires starts only the named services.
	NoRequires bool
	// NoHosts skips hostname registration.
	NoHosts bool
	// Build passes --build to compose up (rebuild Dockerfile-based services).
	Build bool
}

// Up starts services (dependencies first), retrying a failed start once:
// podman-compose occasionally fails the first start of multi-container stacks.
func (e *Engine) Up(ctx context.Context, svcs []catalog.Service, opts UpOptions) error {
	if !opts.NoRequires {
		var err error
		if svcs, err = e.Catalog.WithRequires(svcs); err != nil {
			return err
		}
	}
	if err := e.EnsureNetwork(ctx); err != nil {
		return err
	}
	args := []string{"up", "-d"}
	if opts.Build {
		args = append(args, "--build")
	}
	for _, svc := range svcs {
		e.logf("starting %s", svc.ID)
		if err := e.Runner.Run(ctx, e.Compose(svc, args...)); err != nil {
			e.logf("first start of %s was incomplete; retrying once", svc.ID)
			e.sleep(2 * time.Second)
			if err := e.Runner.Run(ctx, e.Compose(svc, args...)); err != nil {
				return fmt.Errorf("start %s: %w", svc.ID, err)
			}
		}
		if opts.Wait {
			if err := e.Wait(ctx, svc, opts.Timeout); err != nil {
				return err
			}
		}
	}
	if !opts.NoHosts {
		e.EnsureHosts(ctx, svcs)
	}
	return nil
}

// Down stops services in reverse order. Volumes are removed only when asked.
func (e *Engine) Down(ctx context.Context, svcs []catalog.Service, volumes bool) error {
	for i := len(svcs) - 1; i >= 0; i-- {
		svc := svcs[i]
		args := []string{"down"}
		if volumes {
			args = append(args, "--volumes")
		}
		e.logf("stopping %s", svc.ID)
		if err := e.Runner.Run(ctx, e.Compose(svc, args...)); err != nil {
			return fmt.Errorf("stop %s: %w", svc.ID, err)
		}
	}
	return nil
}

// Restart restarts services in order.
func (e *Engine) Restart(ctx context.Context, svcs []catalog.Service) error {
	for _, svc := range svcs {
		if err := e.Runner.Run(ctx, e.Compose(svc, "restart")); err != nil {
			return fmt.Errorf("restart %s: %w", svc.ID, err)
		}
	}
	return nil
}

// Port is a published port.
type Port struct {
	HostIP        string `json:"host_ip"`
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"`
}

// Container is a podman container, matched to a catalog service when possible.
type Container struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	State   string `json:"state"`
	Status  string `json:"status"`
	Health  string `json:"health,omitempty"`
	Ports   []Port `json:"ports,omitempty"`
	Service string `json:"service,omitempty"`
	// Compose is the compose service name inside the catalog entry.
	Compose string `json:"compose_service,omitempty"`
}

type psEntry struct {
	Names  []string          `json:"Names"`
	Image  string            `json:"Image"`
	State  string            `json:"State"`
	Status string            `json:"Status"`
	Labels map[string]string `json:"Labels"`
	Ports  []Port            `json:"Ports"`
}

// Containers lists podman containers (all states) and maps each to a catalog
// service through the compose working-directory label.
func (e *Engine) Containers(ctx context.Context) ([]Container, error) {
	out, err := e.Runner.Output(ctx, runner.Cmd{Name: "podman", Args: []string{"ps", "-a", "--format", "json"}})
	if err != nil {
		return nil, err
	}
	var entries []psEntry
	if len(strings.TrimSpace(string(out))) > 0 {
		if err := json.Unmarshal(out, &entries); err != nil {
			return nil, fmt.Errorf("parse podman ps: %w", err)
		}
	}
	byDir := map[string]catalog.Service{}
	for _, s := range e.Catalog.Services {
		byDir[filepath.Clean(s.Dir)] = s
	}
	cs := make([]Container, 0, len(entries))
	for _, p := range entries {
		c := Container{Image: p.Image, State: p.State, Status: p.Status, Ports: p.Ports, Health: healthFromStatus(p.Status)}
		if len(p.Names) > 0 {
			c.Name = p.Names[0]
		}
		// Match the default project of a service directory only: other projects
		// started from the same directory (per-app node runtimes use -p) are not it.
		if dir := p.Labels["com.docker.compose.project.working_dir"]; dir != "" {
			if svc, ok := byDir[filepath.Clean(dir)]; ok && p.Labels["com.docker.compose.project"] == svc.Name {
				c.Service = svc.ID
			}
		}
		c.Compose = p.Labels["com.docker.compose.service"]
		cs = append(cs, c)
	}
	sort.Slice(cs, func(i, j int) bool { return cs[i].Name < cs[j].Name })
	return cs, nil
}

func healthFromStatus(status string) string {
	for _, h := range []string{"unhealthy", "healthy", "starting"} {
		if strings.Contains(status, "("+h+")") {
			return h
		}
	}
	return ""
}

// ServiceState summarizes a catalog service's containers.
type ServiceState struct {
	State  string `json:"state"` // running, partial, stopped, absent
	Health string `json:"health,omitempty"`
}

// States summarizes container state per catalog service ID.
func States(svcs []catalog.Service, cs []Container) map[string]ServiceState {
	byService := map[string][]Container{}
	for _, c := range cs {
		if c.Service != "" {
			byService[c.Service] = append(byService[c.Service], c)
		}
	}
	out := map[string]ServiceState{}
	for _, s := range svcs {
		list := byService[s.ID]
		if len(list) == 0 {
			out[s.ID] = ServiceState{State: "absent"}
			continue
		}
		running, health := 0, "healthy"
		for _, c := range list {
			if c.State == "running" {
				running++
			}
			switch {
			case c.Health == "unhealthy":
				health = "unhealthy"
			case c.Health == "starting" && health != "unhealthy":
				health = "starting"
			}
		}
		st := ServiceState{State: "stopped"}
		switch {
		case running == len(list):
			st.State = "running"
		case running > 0:
			st.State = "partial"
		}
		if running > 0 {
			st.Health = health
		}
		out[s.ID] = st
	}
	return out
}

type inspectEntry struct {
	Name  string `json:"Name"`
	State struct {
		Status string `json:"Status"`
	} `json:"State"`
	Config struct {
		Healthcheck *json.RawMessage `json:"Healthcheck"`
	} `json:"Config"`
}

// Wait blocks until every container of svc is running and passes its
// healthcheck, then runs the x-devarch ready probe on the primary container.
func (e *Engine) Wait(ctx context.Context, svc catalog.Service, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	deadline := time.Now().Add(timeout)
	e.logf("waiting for %s", svc.ID)
	var last error
	for {
		if last = e.ready(ctx, svc); last == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s was not ready within %s: %w", svc.ID, timeout, last)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		e.sleep(time.Second)
	}
}

func (e *Engine) ready(ctx context.Context, svc catalog.Service) error {
	cs, err := e.Containers(ctx)
	if err != nil {
		return err
	}
	var names []string
	for _, c := range cs {
		if c.Service == svc.ID {
			names = append(names, c.Name)
		}
	}
	if len(names) == 0 {
		return errors.New("no containers found")
	}
	out, err := e.Runner.Output(ctx, runner.Cmd{Name: "podman", Args: append([]string{"inspect", "--format", "json"}, names...)})
	if err != nil {
		return err
	}
	var entries []inspectEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return fmt.Errorf("parse podman inspect: %w", err)
	}
	for _, in := range entries {
		name := strings.TrimPrefix(in.Name, "/")
		if in.State.Status != "running" {
			return fmt.Errorf("%s is %s", name, in.State.Status)
		}
		if in.Config.Healthcheck == nil || string(*in.Config.Healthcheck) == "null" {
			continue
		}
		if _, err := e.Runner.Output(ctx, runner.Cmd{Name: "podman", Args: []string{"healthcheck", "run", name}}); err != nil {
			return fmt.Errorf("%s is not healthy yet", name)
		}
	}
	if r := svc.Meta.Ready; r != nil && len(r.Exec) > 0 {
		primary := e.containerName(svc, cs)
		if _, err := e.Runner.Output(ctx, runner.Cmd{Name: "podman", Args: append([]string{"exec", primary}, r.Exec...)}); err != nil {
			return fmt.Errorf("readiness probe %q failed in %s", strings.Join(r.Exec, " "), primary)
		}
	}
	return nil
}

// containerName returns the running name of svc's primary container.
func (e *Engine) containerName(svc catalog.Service, cs []Container) string {
	p := svc.Primary()
	if p.ContainerName != "" {
		return p.ContainerName
	}
	for _, c := range cs {
		if c.Service == svc.ID && c.Compose == p.Service {
			return c.Name
		}
	}
	return svc.Name + "_" + p.Service + "_1"
}

// LogsCmd returns the native logs command for svc.
func (e *Engine) LogsCmd(svc catalog.Service, follow bool, tail string, extra ...string) runner.Cmd {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail != "" {
		args = append(args, "--tail", tail)
	}
	return e.Compose(svc, append(args, extra...)...)
}
