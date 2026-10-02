// Package doctor checks the local environment DevArch depends on.
package doctor

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/hosts"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

type Status string

const (
	OK   Status = "ok"
	Warn Status = "warn"
	Fail Status = "fail"
)

// Check is one doctor result. Fix is a command or instruction, when known.
type Check struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Detail string `json:"detail"`
	Fix    string `json:"fix,omitempty"`
}

// Env holds what checks need beyond the engine; fields are replaceable in tests.
type Env struct {
	Engine   *engine.Engine
	RootHow  string
	Problems int
	// PortStartFile is /proc/sys/net/ipv4/ip_unprivileged_port_start.
	PortStartFile string
	// PortFree reports whether a host TCP port can be bound.
	PortFree func(ip string, port int) bool
	// DockerGroup reports whether the docker group exists and whether the
	// current user is in it.
	DockerGroup func() (exists, member bool)
	Now         func() time.Time
}

// Run executes every check in order.
func Run(ctx context.Context, env Env) []Check {
	if env.PortStartFile == "" {
		env.PortStartFile = "/proc/sys/net/ipv4/ip_unprivileged_port_start"
	}
	if env.PortFree == nil {
		env.PortFree = portFree
	}
	if env.Now == nil {
		env.Now = time.Now
	}
	if env.DockerGroup == nil {
		env.DockerGroup = dockerGroup
	}
	e := env.Engine
	r := e.Runner
	var out []Check
	add := func(c Check) { out = append(out, c) }

	add(Check{Name: "checkout", Status: OK, Detail: fmt.Sprintf("%s (from %s)", e.Root, env.RootHow)})
	if env.Problems > 0 {
		add(Check{Name: "catalog", Status: Warn, Detail: fmt.Sprintf("%d services, %d entries skipped", len(e.Catalog.Services), env.Problems), Fix: "devarch lint"})
	} else {
		add(Check{Name: "catalog", Status: OK, Detail: fmt.Sprintf("%d services", len(e.Catalog.Services))})
	}

	rt := e.Runtime()
	version, err := r.Output(ctx, e.Cmd("--version"))
	if err != nil {
		add(Check{Name: rt.Name, Status: Fail, Detail: err.Error(), Fix: rt.InstallHint})
		return out // everything else needs the runtime
	}
	add(Check{Name: rt.Name, Status: OK, Detail: strings.TrimSpace(string(version))})

	if _, err := r.Output(ctx, e.Cmd("compose", "version")); err != nil {
		add(Check{Name: "compose provider", Status: Fail, Detail: err.Error(), Fix: rt.ComposeHint})
	} else {
		add(Check{Name: "compose provider", Status: OK, Detail: rt.Name + " compose works"})
	}

	rootless := false
	if rt.Rootless {
		if out, err := r.Output(ctx, e.Cmd("info", "--format", "{{.Host.Security.Rootless}}")); err == nil {
			rootless = strings.TrimSpace(string(out)) == "true"
			if rootless {
				add(Check{Name: "rootless", Status: OK, Detail: "running as " + currentUser()})
			} else {
				add(Check{Name: "rootless", Status: Warn, Detail: "podman is running rootful; containers are not visible to your rootless user",
					Fix: "run devarch as the same non-root user that owns the stack"})
			}
		}
	} else {
		add(dockerAccess(ctx, env))
	}

	network := e.Settings.Network
	if _, err := r.Output(ctx, e.Cmd(rt.NetworkExists(network)...)); err != nil {
		add(Check{Name: "network", Status: Warn, Detail: network + " does not exist", Fix: "devarch up <service> creates it, or: " + e.Cmd("network", "create", network).String()})
	} else {
		add(Check{Name: "network", Status: OK, Detail: network})
	}

	if runtime.GOOS == "linux" && rootless {
		add(unprivilegedPorts(env.PortStartFile))
	}

	containers, cerr := e.Containers(ctx)
	if cerr == nil {
		states := engine.States(e.Catalog.Services, containers)
		if st := states["proxy/nginx-proxy-manager"]; st.State == "running" {
			add(Check{Name: "proxy", Status: OK, Detail: "nginx-proxy-manager is running"})
		} else if _, ok := e.Catalog.Get("proxy/nginx-proxy-manager"); ok {
			add(Check{Name: "proxy", Status: Warn, Detail: "nginx-proxy-manager is not running, so .test sites will not load", Fix: "devarch up nginx-proxy-manager"})
		}
		add(portConflicts(env, states))
	}

	add(certificate(e.Root, env.Now()))
	add(hostsCheck(e))
	if runtime.GOOS == "linux" && rt.Rootless {
		add(linger(ctx, r))
	}
	return out
}

// dockerAccess checks that the Docker daemon answers. The usual cause when it
// does not is a user outside the docker group.
func dockerAccess(ctx context.Context, env Env) Check {
	e := env.Engine
	_, err := e.Runner.Output(ctx, e.Cmd("info", "--format", "{{.ServerVersion}}"))
	exists, member := env.DockerGroup()
	switch {
	case err == nil && member:
		return Check{Name: "docker access", Status: OK, Detail: currentUser() + " is in the docker group"}
	case err == nil:
		return Check{Name: "docker access", Status: OK, Detail: "the Docker daemon answers"}
	case exists && !member:
		return Check{Name: "docker access", Status: Fail, Detail: currentUser() + " is not in the docker group, so the daemon refuses connections",
			Fix: "sudo usermod -aG docker " + currentUser() + ", then log out and back in"}
	}
	fix := "start it (sudo systemctl start docker, or Docker Desktop) and check docker info"
	if member {
		fix = "if you were just added to the docker group, log out and back in; otherwise " + fix
	}
	return Check{Name: "docker access", Status: Fail, Detail: "the Docker daemon does not answer: " + err.Error(), Fix: fix}
}

func dockerGroup() (exists, member bool) {
	g, err := user.LookupGroup("docker")
	if err != nil {
		return false, false
	}
	u, err := user.Current()
	if err != nil {
		return true, false
	}
	ids, _ := u.GroupIds()
	for _, id := range ids {
		if id == g.Gid {
			return true, true
		}
	}
	return true, u.Uid == "0"
}

func currentUser() string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return "current user"
}

func unprivilegedPorts(file string) Check {
	data, err := os.ReadFile(file)
	if err != nil {
		return Check{Name: "ports 80/443", Status: Warn, Detail: "cannot read " + file}
	}
	start, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || start <= 80 {
		return Check{Name: "ports 80/443", Status: OK, Detail: "rootless containers may bind ports from " + strings.TrimSpace(string(data))}
	}
	return Check{Name: "ports 80/443", Status: Fail,
		Detail: fmt.Sprintf("rootless containers cannot bind ports below %d, so the proxy cannot serve 80/443", start),
		Fix:    "echo 'net.ipv4.ip_unprivileged_port_start=80' | sudo tee /etc/sysctl.d/99-devarch.conf && sudo sysctl --system"}
}

// portConflicts checks host ports of stopped x-devarch services: something
// else already listening there will make `devarch up` fail.
func portConflicts(env Env, states map[string]engine.ServiceState) Check {
	var busy []string
	for _, s := range env.Engine.Catalog.Services {
		if s.Meta.Title == "" && s.Meta.Version == nil {
			continue // only services that opted into metadata, to stay fast and relevant
		}
		if st := states[s.ID].State; st == "running" || st == "partial" {
			continue
		}
		for _, c := range s.Containers {
			for _, p := range c.Ports {
				if p.Published > 0 && p.Protocol == "tcp" && !env.PortFree(p.HostIP, p.Published) {
					busy = append(busy, fmt.Sprintf("%d (%s)", p.Published, s.Name))
				}
			}
		}
	}
	if len(busy) == 0 {
		return Check{Name: "host ports", Status: OK, Detail: "no conflicts for stopped core services"}
	}
	return Check{Name: "host ports", Status: Warn, Detail: "already in use: " + strings.Join(busy, ", "),
		Fix: "stop whatever listens there (ss -ltnp), or change the service's published port"}
}

func portFree(ip string, port int) bool {
	if ip == "" {
		ip = "0.0.0.0"
	}
	l, err := net.Listen("tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
	if err != nil {
		return false
	}
	l.Close()
	return true
}

func certificate(root string, now time.Time) Check {
	path := filepath.Join(root, "services-library", "proxy", "nginx-proxy-manager", "config", "certs", "local.crt")
	data, err := os.ReadFile(path)
	if err != nil {
		return Check{Name: "certificate", Status: Fail, Detail: "local HTTPS certificate is missing: " + path,
			Fix: "create one for *.test, e.g. mkcert -cert-file local.crt -key-file local.key '*.test' localhost 127.0.0.1"}
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return Check{Name: "certificate", Status: Fail, Detail: "local.crt is not PEM encoded"}
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return Check{Name: "certificate", Status: Fail, Detail: "cannot parse local.crt: " + err.Error()}
	}
	if !covers(cert.DNSNames, "example.test") {
		return Check{Name: "certificate", Status: Warn, Detail: "local.crt does not cover *.test"}
	}
	left := cert.NotAfter.Sub(now)
	switch {
	case left <= 0:
		return Check{Name: "certificate", Status: Fail, Detail: "local.crt expired on " + cert.NotAfter.Format("2006-01-02")}
	case left < 30*24*time.Hour:
		return Check{Name: "certificate", Status: Warn, Detail: "local.crt expires on " + cert.NotAfter.Format("2006-01-02")}
	}
	return Check{Name: "certificate", Status: OK, Detail: "*.test valid until " + cert.NotAfter.Format("2006-01-02")}
}

func covers(names []string, host string) bool {
	for _, n := range names {
		if n == host || strings.HasPrefix(n, "*.") && strings.Count(host, ".") == strings.Count(n, ".") && strings.HasSuffix(host, n[1:]) {
			return true
		}
	}
	return false
}

func hostsCheck(e *engine.Engine) Check {
	if !e.Settings.HostsManage {
		return Check{Name: "hosts", Status: OK, Detail: "not managed (hosts.manage: false in config.yml)"}
	}
	if e.Hosts == nil {
		return Check{Name: "hosts", Status: Warn, Detail: "hosts management unavailable"}
	}
	content, err := e.Hosts.Read()
	if err != nil {
		return Check{Name: "hosts", Status: Warn, Detail: "cannot read " + e.Hosts.Path()}
	}
	current := hosts.CurrentBlock(content)
	if current == "" {
		return Check{Name: "hosts", Status: Warn, Detail: "no DevArch block in " + e.Hosts.Path(), Fix: "devarch hosts sync"}
	}
	want, _, err := e.HostsBlock(e.Settings.HostsAddress)
	if err != nil {
		return Check{Name: "hosts", Status: Warn, Detail: err.Error()}
	}
	if strings.TrimSpace(want) != strings.TrimSpace(current) {
		return Check{Name: "hosts", Status: Warn, Detail: "the DevArch block in " + e.Hosts.Path() + " is out of date", Fix: "devarch hosts sync"}
	}
	return Check{Name: "hosts", Status: OK, Detail: "DevArch block is current in " + e.Hosts.Path()}
}

func linger(ctx context.Context, r runner.Runner) Check {
	out, err := r.Output(ctx, runner.Cmd{Name: "loginctl", Args: []string{"show-user", currentUser(), "--property=Linger"}})
	if err != nil {
		return Check{Name: "lingering", Status: Warn, Detail: "cannot query loginctl; persistent services need user lingering"}
	}
	if strings.TrimSpace(string(out)) == "Linger=yes" {
		return Check{Name: "lingering", Status: OK, Detail: "user services survive logout"}
	}
	return Check{Name: "lingering", Status: Warn, Detail: "user services stop when you log out", Fix: "loginctl enable-linger " + currentUser()}
}
