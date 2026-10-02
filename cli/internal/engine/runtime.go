package engine

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Runtime is everything that differs between Podman and Docker. The rest of
// the CLI surface (compose subcommands, exec, logs, restart, inspect) is the
// same, so DevArch runs the configured binary with the same arguments.
type Runtime struct {
	Name string
	// NetworkExists and VolumeExists exit non-zero when the object is absent.
	NetworkExists func(name string) []string
	VolumeExists  func(name string) []string
	// RunHealthcheck: Podman runs a container's healthcheck on demand because
	// rootless Podman without systemd timers never runs it by itself; Docker's
	// daemon runs it and inspect reports State.Health.Status.
	RunHealthcheck bool
	// ContainerUser is the uid:gid bootstraps exec as so files they create in
	// bind mounts belong to you: root inside a rootless Podman container maps
	// to your user, while Docker needs your own uid:gid.
	ContainerUser func() string
	// Rootless: doctor checks rootless mode, unprivileged ports 80/443 and
	// lingering. Otherwise it checks that you can reach the Docker daemon.
	Rootless bool
	// PS is the `ps` command and ParsePS reads its output.
	PS      []string
	ParsePS func([]byte) ([]psEntry, error)
	// Events streams container events as JSON lines; EventNames are the
	// lifecycle events that change what the TUI shows.
	Events     []string
	EventNames []string
	// ReplicaSep joins project, service and replica in default container
	// names: podman-compose uses php_php_1, Docker Compose php-php-1.
	ReplicaSep string
	// InstallHint and ComposeHint are doctor fixes when either is missing.
	InstallHint, ComposeHint string
}

var runtimes = map[string]Runtime{
	"podman": {
		Name:           "podman",
		NetworkExists:  func(n string) []string { return []string{"network", "exists", n} },
		VolumeExists:   func(v string) []string { return []string{"volume", "exists", v} },
		RunHealthcheck: true,
		ContainerUser:  func() string { return "0:0" },
		Rootless:       true,
		PS:             []string{"ps", "-a", "--format", "json"},
		ParsePS:        parsePodmanPS,
		Events:         []string{"events", "--format", "json", "--filter", "type=container"},
		EventNames:     []string{"create", "start", "stop", "died", "remove", "restart", "pause", "unpause", "health_status"},
		ReplicaSep:     "_",
		InstallHint:    "install Podman: https://podman.io/docs/installation",
		ComposeHint:    "install podman-compose or docker-compose",
	},
	"docker": {
		Name:          "docker",
		NetworkExists: func(n string) []string { return []string{"network", "inspect", "--format", "{{.Name}}", n} },
		VolumeExists:  func(v string) []string { return []string{"volume", "inspect", "--format", "{{.Name}}", v} },
		ContainerUser: func() string { return fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()) },
		PS:            []string{"ps", "-a", "--no-trunc", "--format", "{{json .}}"},
		ParsePS:       parseDockerPS,
		Events:        []string{"events", "--format", "{{json .}}", "--filter", "type=container"},
		EventNames:    []string{"create", "start", "stop", "die", "destroy", "restart", "pause", "unpause", "health_status"},
		ReplicaSep:    "-",
		InstallHint:   "install Docker Engine: https://docs.docker.com/engine/install/",
		ComposeHint:   "install the Docker Compose plugin: https://docs.docker.com/compose/install/linux/",
	},
}

// Runtime returns the configured runtime's traits.
func (e *Engine) Runtime() Runtime {
	if rt, ok := runtimes[e.Settings.Runtime]; ok {
		return rt
	}
	return runtimes["podman"]
}

// EventsCmd is the command whose JSON lines signal container changes.
func (e *Engine) EventsCmd() []string {
	rt := e.Runtime()
	args := append([]string{}, rt.Events...)
	for _, ev := range rt.EventNames {
		args = append(args, "--filter", "event="+ev)
	}
	return args
}

// Event is one container event from either runtime.
type Event struct {
	Name, Action, Health string
}

// ParseEvent reads a Podman or Docker JSON event line.
func ParseEvent(line []byte) (Event, bool) {
	var raw struct {
		Name         string `json:"Name"`
		Status       string `json:"Status"`
		HealthStatus string `json:"HealthStatus"`
		Action       string `json:"Action"`
		Actor        struct {
			Attributes map[string]string `json:"Attributes"`
		} `json:"Actor"`
	}
	if json.Unmarshal(line, &raw) != nil {
		return Event{}, false
	}
	ev := Event{Name: raw.Name, Action: raw.Status, Health: raw.HealthStatus}
	if raw.Action != "" { // Docker: "health_status: healthy"
		ev.Name = raw.Actor.Attributes["name"]
		action, health, _ := strings.Cut(raw.Action, ":")
		ev.Action, ev.Health = action, strings.TrimSpace(health)
	}
	return ev, true
}

func parsePodmanPS(out []byte) ([]psEntry, error) {
	var entries []psEntry
	if len(bytes.TrimSpace(out)) == 0 {
		return nil, nil
	}
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("parse podman ps: %w", err)
	}
	return entries, nil
}

// parseDockerPS reads `docker ps --format '{{json .}}'`: one object per line
// whose Names, Labels and Ports are strings.
func parseDockerPS(out []byte) ([]psEntry, error) {
	var entries []psEntry
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var d struct {
			Names, Image, State, Status, Labels, Ports string
		}
		if err := json.Unmarshal(line, &d); err != nil {
			return nil, fmt.Errorf("parse docker ps: %w", err)
		}
		entries = append(entries, psEntry{
			Names: strings.Split(d.Names, ","), Image: d.Image, State: d.State, Status: d.Status,
			Labels: dockerLabels(d.Labels), Ports: dockerPorts(d.Ports),
		})
	}
	return entries, sc.Err()
}

// dockerLabels splits "a=1,b=2". A segment without "=" continues the
// previous value, which keeps commas inside values intact.
func dockerLabels(s string) map[string]string {
	labels := map[string]string{}
	last := ""
	for _, part := range strings.Split(s, ",") {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			if last != "" {
				labels[last] += "," + part
			}
			continue
		}
		labels[k], last = v, k
	}
	return labels
}

// dockerPorts parses "127.0.0.1:8080->80/tcp, :::80->80/tcp, 3000/tcp".
func dockerPorts(s string) []Port {
	var ports []Port
	for _, p := range strings.Split(s, ", ") {
		host, target, ok := strings.Cut(strings.TrimSpace(p), "->")
		if !ok {
			continue // exposed but not published
		}
		i := strings.LastIndexByte(host, ':')
		if i < 0 {
			continue
		}
		hostPort, err := strconv.Atoi(host[i+1:])
		if err != nil {
			continue
		}
		cport, proto, _ := strings.Cut(target, "/")
		containerPort, _ := strconv.Atoi(cport)
		ports = append(ports, Port{HostIP: host[:i], HostPort: hostPort, ContainerPort: containerPort, Protocol: proto})
	}
	return ports
}
